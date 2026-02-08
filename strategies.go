package main

//go:generate mockgen -destination mock_strategy_test.go -package main -source=strategies.go

import (
	"context"
	"fmt"
	"sort"
	"strings"
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

// IDStrategy finds targets by TargetID in existing targets map
type IDStrategy struct{}

func (s IDStrategy) Name() string {
	return "IDStrategy"
}

func (s IDStrategy) FindTarget(
	ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string, _ *SyncReport,
) (Target, bool, error) {
	srcID := src.GetTargetID()
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
	return "TitleStrategy"
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
		if src.SameTitleWithTarget(target) && src.SameTypeWithTarget(target) {
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

	srcID := src.GetSourceID()
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
			LogDebugDecision(ctx, "[%s] Found target by API name search in user's list: %s", prefix, tgt.GetTitle())
			return existingTarget, true, nil
		}

		if src.SameTypeWithTarget(tgt) {
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

	targetID := TargetID(targetServiceID)
	if target, exists := existingTargets[targetID]; exists {
		LogDebugDecision(ctx, "[%s] Found target by ARM API: ID %d -> %s",
			prefix, targetServiceID, target.GetTitle())
		return target, true, nil
	}

	LogDebugDecision(ctx, "[%s] ARM API mapped to ID %d but not in user's list",
		prefix, targetServiceID)
	return nil, false, nil
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
			return s.checkTarget(ctx, targetServiceID, existingTargets, prefix)
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
			return s.checkTarget(ctx, targetServiceID, existingTargets, prefix)
		}
		return nil, false, nil
	}

	return nil, false, nil
}

//nolint:dupl // Similar to lookupIDManga but with different media type
func (s HatoAPIStrategy) lookupIDAnime(ctx context.Context, src Anime) (int, bool, error) {
	// Try MAL ID → AniList ID lookup
	if src.IDMal > 0 {
		LogDebug(ctx, "[HATO API] Looking up AniList ID for MAL ID: %d (anime)", src.IDMal)
		id, found, err := s.Client.GetAniListID(ctx, src.IDMal, "anime")
		if err != nil {
			return 0, false, err
		}
		if found {
			LogDebug(ctx, "[HATO API] Found: MAL %d -> AniList %d (anime)", src.IDMal, id)
			return id, true, nil
		}
		LogDebug(ctx, "[HATO API] Not found: MAL %d -> (no mapping) (anime)", src.IDMal)
	}

	// Try AniList ID → MAL ID lookup
	if src.IDAnilist > 0 {
		LogDebug(ctx, "[HATO API] Looking up MAL ID for AniList ID: %d (anime)", src.IDAnilist)
		id, found, err := s.Client.GetMALID(ctx, src.IDAnilist, "anime")
		if err != nil {
			return 0, false, err
		}
		if found {
			LogDebug(ctx, "[HATO API] Found: AniList %d -> MAL %d (anime)", src.IDAnilist, id)
			return id, true, nil
		}
		LogDebug(ctx, "[HATO API] Not found: AniList %d -> (no mapping) (anime)", src.IDAnilist)
	}

	return 0, false, nil
}

//nolint:dupl // Similar to lookupIDAnime but with different media type
func (s HatoAPIStrategy) lookupIDManga(ctx context.Context, src Manga) (int, bool, error) {
	// Try MAL ID → AniList ID lookup
	if src.IDMal > 0 {
		LogDebug(ctx, "[HATO API] Looking up AniList ID for MAL ID: %d (manga)", src.IDMal)
		id, found, err := s.Client.GetAniListID(ctx, src.IDMal, "manga")
		if err != nil {
			return 0, false, err
		}
		if found {
			LogDebug(ctx, "[HATO API] Found: MAL %d -> AniList %d (manga)", src.IDMal, id)
			return id, true, nil
		}
		LogDebug(ctx, "[HATO API] Not found: MAL %d -> (no mapping) (manga)", src.IDMal)
	}

	// Try AniList ID → MAL ID lookup
	if src.IDAnilist > 0 {
		LogDebug(ctx, "[HATO API] Looking up MAL ID for AniList ID: %d (manga)", src.IDAnilist)
		id, found, err := s.Client.GetMALID(ctx, src.IDAnilist, "manga")
		if err != nil {
			return 0, false, err
		}
		if found {
			LogDebug(ctx, "[HATO API] Found: AniList %d -> MAL %d (manga)", src.IDAnilist, id)
			return id, true, nil
		}
		LogDebug(ctx, "[HATO API] Not found: AniList %d -> (no mapping) (manga)", src.IDAnilist)
	}

	return 0, false, nil
}

func (s HatoAPIStrategy) checkTarget(
	ctx context.Context,
	targetServiceID int,
	existingTargets map[TargetID]Target,
	prefix string,
) (Target, bool, error) {
	targetID := TargetID(targetServiceID)
	if target, exists := existingTargets[targetID]; exists {
		LogDebugDecision(ctx, "[%s] Found target by Hato API: ID %d -> %s",
			prefix, targetServiceID, target.GetTitle())
		return target, true, nil
	}

	LogDebugDecision(ctx, "[%s] Hato API mapped to ID %d but not in user's list",
		prefix, targetServiceID)
	return nil, false, nil
}
