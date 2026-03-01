package main

//go:generate mockgen -destination mock_strategy_test.go -package main -source=strategies.go

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// MatchResult holds a matched target with metadata about how it was found.
type MatchResult struct {
	Target       Target
	StrategyName string
	StrategyIdx  int // position in chain, lower = higher priority
}

// Strategy name constants used by Name() methods.
const (
	StrategyNameID    = "IDStrategy"
	StrategyNameTitle = "TitleStrategy"
)

// TargetFindStrategy defines a strategy for finding targets
type TargetFindStrategy interface {
	FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string, report *SyncReport) (Target, bool, error)
	Name() string
}

// StrategyChain manages a chain of target finding strategies
type StrategyChain struct {
	strategies []TargetFindStrategy
}

// NewStrategyChain creates a new strategy chain
func NewStrategyChain(strategies ...TargetFindStrategy) *StrategyChain {
	return &StrategyChain{strategies: strategies}
}

// FindTarget executes strategies in order until one succeeds
func (sc *StrategyChain) FindTarget(
	ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string, report *SyncReport,
) (Target, error) {
	for _, strategy := range sc.strategies {
		target, found, err := strategy.FindTarget(ctx, src, existingTargets, prefix, report)
		if err != nil {
			return nil, fmt.Errorf("strategy %s failed: %w", strategy.Name(), err)
		}
		if found {
			LogDebugDecision(ctx, "[%s] Found target using strategy: %s", prefix, strategy.Name())
			return target, nil
		}
	}
	return nil, fmt.Errorf("no target found for source: %s", src.GetTitle())
}

// FindTargetWithMeta executes strategies in order and returns match metadata.
func (sc *StrategyChain) FindTargetWithMeta(
	ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string, report *SyncReport,
) (*MatchResult, error) {
	for idx, strategy := range sc.strategies {
		target, found, err := strategy.FindTarget(ctx, src, existingTargets, prefix, report)
		if err != nil {
			return nil, fmt.Errorf("strategy %s failed: %w", strategy.Name(), err)
		}
		if found {
			LogDebugDecision(ctx, "[%s] Found target using strategy: %s", prefix, strategy.Name())
			return &MatchResult{
				Target:       target,
				StrategyName: strategy.Name(),
				StrategyIdx:  idx,
			}, nil
		}
	}
	return nil, fmt.Errorf("no target found for source: %s", src.GetTitle())
}

// ManualMappingStrategy finds targets using user-defined manual mappings.
// This should be the first strategy in the chain.
type ManualMappingStrategy struct {
	Mappings *MappingsConfig
	Reverse  bool // true for MAL→AniList direction
}

func (s ManualMappingStrategy) Name() string {
	return "ManualMappingStrategy"
}

func (s ManualMappingStrategy) FindTarget(
	ctx context.Context,
	src Source,
	existingTargets map[TargetID]Target,
	prefix string,
	_ *SyncReport,
) (Target, bool, error) {
	if s.Mappings == nil {
		return nil, false, nil
	}

	targetID, found := s.lookupManualMapping(ctx, src)
	if !found {
		return nil, false, nil
	}

	tgtID := TargetID(targetID)
	if target, exists := existingTargets[tgtID]; exists {
		LogDebugDecision(ctx, "[%s] Found target by manual mapping: ID %d -> %s", prefix, targetID, target.GetTitle())
		return target, true, nil
	}

	LogDebugDecision(ctx, "[%s] Manual mapping found ID %d but not in user's list", prefix, targetID)
	return nil, false, nil
}

func (s ManualMappingStrategy) lookupManualMapping(ctx context.Context, src Source) (int, bool) {
	switch v := src.(type) {
	case Anime:
		return s.lookupByIDs(ctx, v.IDAnilist, v.IDMal)
	case Manga:
		return s.lookupByIDs(ctx, v.IDAnilist, v.IDMal)
	}
	return 0, false
}

func (s ManualMappingStrategy) lookupByIDs(_ context.Context, anilistID, malID int) (int, bool) {
	if s.Reverse {
		return s.lookupReverse(anilistID, malID)
	}
	return s.lookupForward(anilistID, malID)
}

func (s ManualMappingStrategy) lookupForward(anilistID, malID int) (int, bool) {
	// AniList→MAL: source has AniList ID, need MAL ID
	if anilistID > 0 {
		if id, ok := s.Mappings.GetManualMALID(anilistID); ok {
			return id, true
		}
	}
	if malID > 0 {
		if id, ok := s.Mappings.GetManualMALID(malID); ok {
			return id, true
		}
	}
	return 0, false
}

func (s ManualMappingStrategy) lookupReverse(_, malID int) (int, bool) {
	// MAL→AniList: source is a MAL entry with IDAnilist=0 (set by newAnimeFromMalAnime
	// when reverse=true). Only MAL ID is meaningful here.
	if malID > 0 {
		if id, ok := s.Mappings.GetManualAniListID(malID); ok {
			return id, true
		}
	}
	return 0, false
}

// IDStrategy finds targets by TargetID in existing targets map
type IDStrategy struct{}

func (s IDStrategy) Name() string {
	return StrategyNameID
}

func (s IDStrategy) FindTarget(
	ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string, _ *SyncReport,
) (Target, bool, error) {
	direction := DirectionFromContext(ctx)
	srcID := GetTargetIDWithDirection(src, direction)
	target, found := existingTargets[srcID]
	if found {
		LogDebugDecision(ctx, "[%s] Found target by ID %d (direct lookup in user's list): %s", prefix, srcID, target.GetTitle())
	} else if srcID > 0 {
		LogDebugDecision(ctx, "[%s] Target ID %d not found in user's list (will try other strategies)", prefix, srcID)
	}
	return target, found, nil
}

// TitleStrategy finds targets by title comparison with existing targets
type TitleStrategy struct{}

func (s TitleStrategy) Name() string {
	return StrategyNameTitle
}

// shouldRejectMatch checks if a potential match should be rejected
// Returns true if the match should be rejected with appropriate logging
func shouldRejectMatch(ctx context.Context, src Source, target Target, prefix string, report *SyncReport) bool {
	// Check MAL ID mismatch
	srcID := src.GetTargetID()
	tgtID := target.GetTargetID()

	if srcID > 0 && tgtID > 0 && srcID != tgtID {
		LogDebugDecision(ctx, "[%s] Rejecting title match due to MAL ID mismatch: Source MAL ID: %d, Target MAL ID: %d",
			prefix, srcID, tgtID)
		LogDebugDecision(ctx, "[%s]   Source: %s", prefix, src.String())
		LogDebugDecision(ctx, "[%s]   Target: %s", prefix, target.String())
		return true
	}

	// Check for potentially incorrect matches (special vs series)
	srcAnime, ok := src.(Anime)
	if !ok {
		return false
	}

	if srcAnime.IsPotentiallyIncorrectMatch(target) {
		tgtAnime, _ := target.(Anime)

		// Determine the specific reason for rejection
		reason := "unknown reason"
		if srcAnime.IDMal == 0 && tgtAnime.IDMal > 0 && !srcAnime.IdenticalTitleMatch(tgtAnime) {
			reason = "different titles (source has no MAL ID, target has different MAL ID)"
		} else if (srcAnime.NumEpisodes == 0 || srcAnime.NumEpisodes == 1) && tgtAnime.NumEpisodes > 4 {
			reason = "episode count mismatch (special vs series)"
		}

		// Accumulate warning for deferred output instead of logging immediately
		if report != nil {
			mediaType := strings.TrimPrefix(prefix, "AniList to MAL ")
			mediaType = strings.TrimPrefix(mediaType, "MAL to AniList ")
			report.AddWarning(
				srcAnime.GetTitle(),
				reason,
				fmt.Sprintf("(%d vs %d)", srcAnime.NumEpisodes, tgtAnime.NumEpisodes),
				mediaType,
			)
		}

		LogDebugDecision(ctx, "[%s] Rejecting potentially incorrect match: %q - %s (%d vs %d)",
			prefix, srcAnime.GetTitle(), reason, srcAnime.NumEpisodes, tgtAnime.NumEpisodes)
		return true
	}

	return false
}

func (s TitleStrategy) FindTarget(
	ctx context.Context,
	src Source,
	existingTargets map[TargetID]Target,
	prefix string,
	report *SyncReport,
) (Target, bool, error) {
	srcTitle := src.GetTitle()

	targetSlice := make([]Target, 0, len(existingTargets))
	for _, target := range existingTargets {
		targetSlice = append(targetSlice, target)
	}

	sort.Slice(targetSlice, func(i, j int) bool {
		return targetSlice[i].GetTitle() < targetSlice[j].GetTitle()
	})

	byTitle := map[string]Target{}
	for _, target := range targetSlice {
		byTitle[target.GetTitle()] = target
	}

	if target, ok := byTitle[srcTitle]; ok {
		LogDebugDecision(ctx, "[%s] Found target by exact title match: %s", prefix, srcTitle)
		return target, true, nil
	}

	for _, target := range targetSlice {
		if src.SameTitleWithTarget(ctx, target) && src.SameTypeWithTarget(ctx, target) {
			// Check for potential mismatches and reject if needed
			if shouldRejectMatch(ctx, src, target, prefix, report) {
				continue
			}

			LogDebugDecision(ctx, "[%s] Found target by title comparison (fuzzy match): '%s' -> '%s'",
				prefix, srcTitle, target.GetTitle())
			return target, true, nil
		}
	}

	LogDebugDecision(ctx, "[%s] No target found by title comparison: %s", prefix, srcTitle)
	return nil, false, nil
}

// MALIDStrategy finds targets by searching AniList using source MAL ID
type MALIDStrategy struct {
	Service MediaServiceWithMALID
}

func (s MALIDStrategy) Name() string {
	return "MALIDStrategy"
}

func (s MALIDStrategy) FindTarget(
	ctx context.Context,
	src Source,
	existingTargets map[TargetID]Target,
	prefix string,
	_ *SyncReport,
) (Target, bool, error) {
	// Check for context cancellation before potentially long-running search
	select {
	case <-ctx.Done():
		return nil, false, fmt.Errorf("context cancelled during MAL ID search")
	default:
	}

	direction := DirectionFromContext(ctx)
	srcID := GetSourceIDWithDirection(src, direction)
	if srcID <= 0 {
		return nil, false, nil
	}

	LogDebugDecision(ctx, "[%s] Finding target by MAL ID (title match failed): %d", prefix, srcID)
	target, err := s.Service.GetByMALID(ctx, srcID, prefix)
	if err != nil {
		return nil, false, fmt.Errorf("error getting target by MAL ID %d: %w", srcID, err)
	}

	if target == nil {
		return nil, false, nil
	}

	// Log if titles differ (this is why MAL ID search is useful)
	if target.GetTitle() != src.GetTitle() {
		LogDebugDecision(ctx, "[%s] MAL ID search matched different titles: '%s' (source) -> '%s' (target)",
			prefix, src.GetTitle(), target.GetTitle())
	}

	tgtID := target.GetTargetID()
	if existingTarget, exists := existingTargets[tgtID]; exists {
		LogDebugDecision(ctx, "[%s] Found existing user target by MAL ID %d: %s", prefix, srcID, target.GetTitle())
		return existingTarget, true, nil
	}

	LogDebugDecision(ctx, "[%s] Found target by MAL ID %d: %s (using MAL ID lookup instead of title match)",
		prefix, srcID, target.GetTitle())
	return target, true, nil
}

// APISearchStrategy finds targets by making API calls
type APISearchStrategy struct {
	Service MediaService
}

func (s APISearchStrategy) Name() string {
	return "APISearchStrategy"
}

func (s APISearchStrategy) FindTarget(
	ctx context.Context,
	src Source,
	existingTargets map[TargetID]Target,
	prefix string,
	report *SyncReport,
) (Target, bool, error) {
	// Check for context cancellation before potentially long-running search
	select {
	case <-ctx.Done():
		return nil, false, fmt.Errorf("context cancelled during API search")
	default:
	}

	tgtID := src.GetTargetID()

	if tgtID > 0 {
		LogDebugDecision(ctx, "[%s] Finding target by API ID (not in user's list): %d", prefix, tgtID)
		target, err := s.Service.GetByID(ctx, tgtID, prefix)
		if err != nil {
			return nil, false, fmt.Errorf("error getting target by ID %d: %w", tgtID, err)
		}

		if existingTarget, exists := existingTargets[tgtID]; exists {
			LogDebugDecision(ctx, "[%s] Found target by API ID lookup in user's list: %s", prefix, existingTarget.GetTitle())
			return existingTarget, true, nil
		}

		LogDebugDecision(ctx, "[%s] Found target by API ID lookup (not in user's list): %s", prefix, target.GetTitle())
		return target, true, nil
	}

	LogDebugDecision(ctx, "[%s] Finding target by API name search (ID lookup failed): %s", prefix, src.GetTitle())
	targets, err := s.Service.GetByName(ctx, src.GetTitle(), prefix)
	if err != nil {
		return nil, false, fmt.Errorf("error getting targets by name %s: %w", src.GetTitle(), err)
	}

	for _, tgt := range targets {
		if existingTarget, exists := existingTargets[tgt.GetTargetID()]; exists {
			// Check for potential mismatches before accepting API search result
			if shouldRejectMatch(ctx, src, existingTarget, prefix, report) {
				continue
			}
			// Verify title similarity to avoid matching different entries
			// (e.g., multiple AniList volumes matching a single MAL umbrella series)
			if !src.SameTitleWithTarget(ctx, existingTarget) {
				LogDebugDecision(ctx, "[%s] Rejecting API name search match due to title mismatch: %q vs %q",
					prefix, src.GetTitle(), existingTarget.GetTitle())
				continue
			}
			LogDebugDecision(ctx, "[%s] Found target by API name search in user's list: %s", prefix, tgt.GetTitle())
			return existingTarget, true, nil
		}

		if src.SameTypeWithTarget(ctx, tgt) {
			LogDebugDecision(ctx, "[%s] Found target by API name search (not in user's list): %s", prefix, tgt.GetTitle())
			return tgt, true, nil
		}
		LogDebugDecision(ctx, "[%s] Ignoring target by API name: %s (type mismatch)", prefix, tgt.GetTitle())
	}

	return nil, false, nil
}

// OfflineDatabaseStrategy finds targets using the anime-offline-database ID mappings.
// Only works for anime (not manga).
type OfflineDatabaseStrategy struct {
	Database *OfflineDatabase
}

func (s OfflineDatabaseStrategy) Name() string {
	return "OfflineDatabaseStrategy"
}

func (s OfflineDatabaseStrategy) FindTarget(
	ctx context.Context,
	src Source,
	existingTargets map[TargetID]Target,
	prefix string,
	_ *SyncReport,
) (Target, bool, error) {
	if s.Database == nil {
		return nil, false, nil
	}

	srcAnime, ok := src.(Anime)
	if !ok {
		return nil, false, nil
	}

	targetServiceID, found := s.lookupID(srcAnime)
	if !found {
		return nil, false, nil
	}

	targetID := TargetID(targetServiceID)
	if target, exists := existingTargets[targetID]; exists {
		LogDebugDecision(ctx, "[%s] Found target by offline database: ID %d -> %s",
			prefix, targetServiceID, target.GetTitle())
		return target, true, nil
	}

	LogDebugDecision(ctx, "[%s] Offline database mapped to ID %d but not in user's list",
		prefix, targetServiceID)
	return nil, false, nil
}

func (s OfflineDatabaseStrategy) lookupID(src Anime) (int, bool) {
	if src.IDMal > 0 {
		if id, ok := s.Database.GetAniListID(src.IDMal); ok {
			return id, true
		}
	}
	if src.IDAnilist > 0 {
		if id, ok := s.Database.GetMALID(src.IDAnilist); ok {
			return id, true
		}
	}
	return 0, false
}

// ARMAPIStrategy finds targets using the ARM API for ID mapping.
// Only works for anime (not manga).
type ARMAPIStrategy struct {
	Client *ARMClient
}

func (s ARMAPIStrategy) Name() string {
	return "ARMAPIStrategy"
}

func (s ARMAPIStrategy) FindTarget(
	ctx context.Context,
	src Source,
	existingTargets map[TargetID]Target,
	prefix string,
	_ *SyncReport,
) (Target, bool, error) {
	if s.Client == nil {
		return nil, false, nil
	}

	srcAnime, ok := src.(Anime)
	if !ok {
		return nil, false, nil
	}

	targetServiceID, found, err := s.lookupID(ctx, srcAnime)
	if err != nil {
		LogDebug(ctx, "[%s] ARM API error: %v", prefix, err)
		return nil, false, nil
	}
	if !found {
		return nil, false, nil
	}

	return checkExistingTarget(ctx, existingTargets, targetServiceID, prefix, "ARM API")
}

func (s ARMAPIStrategy) lookupID(ctx context.Context, src Anime) (int, bool, error) {
	if src.IDMal > 0 {
		LogDebug(ctx, "[ARM API] Looking up AniList ID for MAL ID: %d", src.IDMal)
		id, found, err := s.Client.GetAniListID(ctx, src.IDMal)
		if err != nil {
			return 0, false, err
		}
		if found {
			LogDebug(ctx, "[ARM API] Found: MAL %d -> AniList %d", src.IDMal, id)
			return id, true, nil
		}
		LogDebug(ctx, "[ARM API] Not found: MAL %d -> (no mapping)", src.IDMal)
	}
	if src.IDAnilist > 0 {
		LogDebug(ctx, "[ARM API] Looking up MAL ID for AniList ID: %d", src.IDAnilist)
		id, found, err := s.Client.GetMALID(ctx, src.IDAnilist)
		if err != nil {
			return 0, false, err
		}
		if found {
			LogDebug(ctx, "[ARM API] Found: AniList %d -> MAL %d", src.IDAnilist, id)
			return id, true, nil
		}
		LogDebug(ctx, "[ARM API] Not found: AniList %d -> (no mapping)", src.IDAnilist)
	}
	return 0, false, nil
}

// HatoAPIStrategy finds targets using the Hato API for ID mapping.
// Works for both anime and manga.
type HatoAPIStrategy struct {
	Client *HatoClient
}

func (s HatoAPIStrategy) Name() string {
	return "HatoAPIStrategy"
}

func (s HatoAPIStrategy) FindTarget(
	ctx context.Context,
	src Source,
	existingTargets map[TargetID]Target,
	prefix string,
	_ *SyncReport,
) (Target, bool, error) {
	// Early return if client is disabled
	if s.Client == nil {
		return nil, false, nil
	}

	// Try Anime type
	if srcAnime, ok := src.(Anime); ok {
		targetServiceID, found, err := s.lookupIDAnime(ctx, srcAnime)
		if err != nil {
			LogDebug(ctx, "[%s] Hato API error: %v", prefix, err)
			return nil, false, nil
		}
		if found {
			return checkExistingTarget(ctx, existingTargets, targetServiceID, prefix, "Hato API")
		}
		return nil, false, nil
	}

	// Try Manga type
	if srcManga, ok := src.(Manga); ok {
		targetServiceID, found, err := s.lookupIDManga(ctx, srcManga)
		if err != nil {
			LogDebug(ctx, "[%s] Hato API error: %v", prefix, err)
			return nil, false, nil
		}
		if found {
			return checkExistingTarget(ctx, existingTargets, targetServiceID, prefix, "Hato API")
		}
		return nil, false, nil
	}

	return nil, false, nil
}

// lookupID performs bidirectional ID lookup using the Hato API.
// First tries MAL ID → AniList ID, then AniList ID → MAL ID.
func (s HatoAPIStrategy) lookupID(ctx context.Context, malID, anilistID int, mediaType string) (int, bool, error) {
	// Try MAL ID → AniList ID lookup
	if malID > 0 {
		LogDebug(ctx, "[HATO API] Looking up AniList ID for MAL ID: %d (%s)", malID, mediaType)
		id, found, err := s.Client.GetAniListID(ctx, malID, mediaType)
		if err != nil {
			return 0, false, err
		}
		if found {
			LogDebug(ctx, "[HATO API] Found: MAL %d -> AniList %d (%s)", malID, id, mediaType)
			return id, true, nil
		}
		LogDebug(ctx, "[HATO API] Not found: MAL %d -> (no mapping) (%s)", malID, mediaType)
	}

	// Try AniList ID → MAL ID lookup
	if anilistID > 0 {
		LogDebug(ctx, "[HATO API] Looking up MAL ID for AniList ID: %d (%s)", anilistID, mediaType)
		id, found, err := s.Client.GetMALID(ctx, anilistID, mediaType)
		if err != nil {
			return 0, false, err
		}
		if found {
			LogDebug(ctx, "[HATO API] Found: AniList %d -> MAL %d (%s)", anilistID, id, mediaType)
			return id, true, nil
		}
		LogDebug(ctx, "[HATO API] Not found: AniList %d -> (no mapping) (%s)", anilistID, mediaType)
	}

	return 0, false, nil
}

func (s HatoAPIStrategy) lookupIDAnime(ctx context.Context, src Anime) (int, bool, error) {
	return s.lookupID(ctx, src.IDMal, src.IDAnilist, "anime")
}

func (s HatoAPIStrategy) lookupIDManga(ctx context.Context, src Manga) (int, bool, error) {
	return s.lookupID(ctx, src.IDMal, src.IDAnilist, "manga")
}

// JikanAPIStrategy finds targets using the Jikan API for manga ID mapping.
// Only works for manga (not anime).
type JikanAPIStrategy struct {
	Client *JikanClient
}

func (s JikanAPIStrategy) Name() string {
	return "JikanAPIStrategy"
}

func (s JikanAPIStrategy) FindTarget(
	ctx context.Context,
	src Source,
	existingTargets map[TargetID]Target,
	prefix string,
	_ *SyncReport,
) (Target, bool, error) {
	if s.Client == nil {
		return nil, false, nil
	}

	srcManga, ok := src.(Manga)
	if !ok {
		return nil, false, nil
	}

	// AniList→MAL direction: source has AniList ID, needs MAL ID
	if srcManga.IDMal == 0 && srcManga.IDAnilist > 0 {
		return s.findMALTarget(ctx, srcManga, existingTargets, prefix)
	}

	// MAL→AniList direction: source has MAL ID, needs AniList ID
	if srcManga.IDAnilist == 0 && srcManga.IDMal > 0 {
		return s.findAniListTarget(ctx, srcManga, existingTargets, prefix)
	}

	return nil, false, nil
}

// findMALTarget searches Jikan by title to find the MAL ID, then looks up the target.
func (s JikanAPIStrategy) findMALTarget(
	ctx context.Context,
	src Manga,
	existingTargets map[TargetID]Target,
	prefix string,
) (Target, bool, error) {
	titles := searchTitlesForJikan(src.TitleEN, src.TitleJP, src.TitleRomaji)
	if len(titles) == 0 {
		return nil, false, nil
	}

	for _, query := range titles {
		results := s.Client.SearchManga(ctx, query)

		malID := findBestJikanMatch(ctx, results, src.TitleEN, src.TitleJP, src.TitleRomaji)
		if malID <= 0 {
			continue
		}

		LogDebug(ctx, "[%s] Jikan: matched %q -> MAL ID %d", prefix, src.GetTitle(), malID)
		return checkExistingTarget(ctx, existingTargets, malID, prefix, "Jikan API")
	}

	LogDebug(ctx, "[%s] Jikan: no match found for %q", prefix, src.GetTitle())
	return nil, false, nil
}

// findAniListTarget gets manga from Jikan by MAL ID, then uses enriched titles
// to match against existing AniList targets.
func (s JikanAPIStrategy) findAniListTarget(
	ctx context.Context,
	src Manga,
	existingTargets map[TargetID]Target,
	prefix string,
) (Target, bool, error) {
	jikanData, found := s.Client.GetMangaByMALID(ctx, src.IDMal)
	if !found || jikanData == nil {
		return nil, false, nil
	}

	// Use enriched title data from Jikan to find match in existing targets
	for _, target := range existingTargets {
		tgtManga, ok := target.(Manga)
		if !ok {
			continue
		}

		if matchJikanMangaToSource(ctx, jikanData, tgtManga.TitleEN, tgtManga.TitleJP, tgtManga.TitleRomaji) {
			LogDebugDecision(ctx, "[%s] Found target by Jikan API: MAL %d -> %s (AniList %d)",
				prefix, src.IDMal, tgtManga.GetTitle(), tgtManga.IDAnilist)
			return target, true, nil
		}
	}

	LogDebug(ctx, "[%s] Jikan: MAL %d (%s) not matched to any existing target",
		prefix, src.IDMal, jikanData.Title)
	return nil, false, nil
}

// checkExistingTarget checks if a target ID exists in the user's list and returns appropriate result.
// This is a shared helper used by API-based strategies (ARM, Hato, Jikan).
func checkExistingTarget(
	ctx context.Context,
	existingTargets map[TargetID]Target,
	targetServiceID int,
	prefix string,
	apiName string,
) (Target, bool, error) {
	targetID := TargetID(targetServiceID)
	if target, exists := existingTargets[targetID]; exists {
		LogDebugDecision(ctx, "[%s] Found target by %s: ID %d -> %s",
			prefix, apiName, targetServiceID, target.GetTitle())
		return target, true, nil
	}

	LogDebugDecision(ctx, "[%s] %s mapped to ID %d but not in user's list",
		prefix, apiName, targetServiceID)
	return nil, false, nil
}
