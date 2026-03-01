package main

//go:generate mockgen -destination mock_updater_test.go -package main -source=updater.go

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type TargetID int

type Source interface {
	GetStatusString() string
	GetTargetID() TargetID
	GetSourceID() int
	GetTitle() string
	GetStringDiffWithTarget(Target) string
	SameProgressWithTarget(Target) bool
	SameTypeWithTarget(ctx context.Context, t Target) bool
	SameTitleWithTarget(ctx context.Context, t Target) bool
	String() string
}

type Target interface {
	GetTargetID() TargetID
	GetTitle() string
	String() string
}

type Updater struct {
	Prefix        string
	Statistics    *Statistics
	IgnoreTitles  map[string]struct{}
	IgnoreIDs     map[int]struct{}
	StrategyChain *StrategyChain
	Service       MediaService // Replaces callback
	ForceSync     bool         // Skip matching logic, force sync all
	DryRun        bool         // Skip actual updates
	Reverse       bool         // true for MAL→AniList direction
	MediaType     string       // "anime" or "manga" — used for unmapped tracking
	UnmappedList  []UnmappedEntry
	ResolvedMappings []resolvedMapping // Saved for favorites sync
}

// resolvedMapping holds a source→target mapping with strategy metadata.
type resolvedMapping struct {
	src          Source
	target       Target
	strategyName string
	strategyIdx  int
}

// duplicateConflict records when multiple sources map to the same target.
type duplicateConflict struct {
	loserSrc    Source
	winnerSrc   Source
	target      Target
	loserStrat  string
	winnerStrat string
}

func (u *Updater) Update(ctx context.Context, srcs []Source, tgts []Target, report *SyncReport) {
	tgtsByID := buildTargetMap(ctx, tgts)
	srcs = u.sortSources(srcs)
	filtered := u.filterSources(ctx, srcs)

	if u.ForceSync {
		u.processForceSyncSources(ctx, filtered, tgtsByID, report)
		return
	}

	// Phase 1: Resolve all source→target mappings
	resolved, unmapped := u.resolveAllMappings(ctx, filtered, tgtsByID, report)

	// Phase 2: Deduplicate (detect N:1 conflicts)
	kept, conflicts := u.deduplicateMappings(ctx, resolved)

	// Save resolved mappings for favorites sync (before processing to preserve target state)
	u.ResolvedMappings = kept

	// Phase 3: Process
	u.processResolvedMappings(ctx, kept, report)
	u.recordConflicts(ctx, conflicts, report)
	u.recordUnmapped(ctx, unmapped)
}

func buildTargetMap(ctx context.Context, tgts []Target) map[TargetID]Target {
	tgtsByID := make(map[TargetID]Target, len(tgts))
	for _, tgt := range tgts {
		tgtsByID[tgt.GetTargetID()] = tgt
	}
	return tgtsByID
}

func (u *Updater) sortSources(srcs []Source) []Source {
	sort.Slice(srcs, func(i, j int) bool {
		if srcs[i].GetStatusString() != srcs[j].GetStatusString() {
			return srcs[i].GetStatusString() < srcs[j].GetStatusString()
		}
		return srcs[i].GetTitle() < srcs[j].GetTitle()
	})
	return srcs
}

func (u *Updater) filterSources(ctx context.Context, srcs []Source) []Source {
	filtered := make([]Source, 0, len(srcs))
	for _, src := range srcs {
		if src.GetStatusString() == "" {
			LogDebug(ctx, "[%s] Skipping source with empty status: %s", u.Prefix, src.String())
			continue
		}
		// Count all valid entries in Total (including ignored ones).
		// This ensures Total == Updated + Skipped + Errors in the summary.
		u.Statistics.IncrementTotal()
		if u.isIgnored(ctx, src) {
			continue
		}
		filtered = append(filtered, src)
	}
	return filtered
}

func (u *Updater) isIgnored(ctx context.Context, src Source) bool {
	if _, ok := u.IgnoreTitles[strings.ToLower(src.GetTitle())]; ok {
		LogDebug(ctx, "[%s] Ignoring entry: %s", u.Prefix, src.GetTitle())
		u.Statistics.RecordSkip(UpdateResult{
			Title:      src.GetTitle(),
			Status:     src.GetStatusString(),
			Skipped:    true,
			SkipReason: "in ignore list",
		})
		return true
	}
	if len(u.IgnoreIDs) > 0 {
		srcID := src.GetSourceID()
		if _, ok := u.IgnoreIDs[srcID]; ok {
			LogDebug(ctx, "[%s] Ignoring entry by ID: %s (ID: %d)",
				u.Prefix, src.GetTitle(), srcID)
			u.Statistics.RecordSkip(UpdateResult{
				Title:      src.GetTitle(),
				Status:     src.GetStatusString(),
				Skipped:    true,
				SkipReason: "in ignore list",
			})
			return true
		}
	}
	return false
}

func (u *Updater) logStatusTransition(ctx context.Context, prev *string, src Source) {
	if *prev != src.GetStatusString() {
		*prev = src.GetStatusString()
		shortPrefix := strings.TrimPrefix(u.Prefix, "AniList to MAL ")
		shortPrefix = strings.TrimPrefix(shortPrefix, "MAL to AniList ")
		LogStage(ctx, "[%s] Processing %s...", shortPrefix, *prev)
	}
}

func (u *Updater) processForceSyncSources(
	ctx context.Context, srcs []Source, tgtsByID map[TargetID]Target, report *SyncReport,
) {
	var statusStr string
	for i, src := range srcs {
		select {
		case <-ctx.Done():
			LogWarn(ctx, "[%s] Context cancelled, stopping sync", u.Prefix)
			return
		default:
		}
		u.logStatusTransition(ctx, &statusStr, src)
		LogProgress(ctx, i+1, len(srcs), statusStr, src.GetTitle())
		LogDebug(ctx, "[%s] Processing: %s", u.Prefix, src.String())
		u.updateSourceByTargets(ctx, src, tgtsByID, report)
	}
}

// Phase 1: resolveAllMappings finds target for each source using strategy chain.
func (u *Updater) resolveAllMappings(
	ctx context.Context, srcs []Source, tgtsByID map[TargetID]Target, report *SyncReport,
) ([]resolvedMapping, []Source) {
	var (
		resolved  = make([]resolvedMapping, 0, len(srcs))
		unmapped  []Source
		statusStr string
	)

	for i, src := range srcs {
		select {
		case <-ctx.Done():
			LogWarn(ctx, "[%s] Context cancelled, stopping sync", u.Prefix)
			return resolved, unmapped
		default:
		}

		u.logStatusTransition(ctx, &statusStr, src)
		LogProgress(ctx, i+1, len(srcs), statusStr, src.GetTitle())
		LogDebug(ctx, "[%s] Processing: %s", u.Prefix, src.String())

		match, err := u.StrategyChain.FindTargetWithMeta(ctx, src, tgtsByID, u.Prefix, report)
		if err != nil {
			LogDebug(ctx, "[%s] Error finding target: %v", u.Prefix, err)
			u.Statistics.RecordSkip(UpdateResult{
				Title:      src.GetTitle(),
				Status:     src.GetStatusString(),
				Skipped:    true,
				SkipReason: "unmapped",
			})
			unmapped = append(unmapped, src)
			continue
		}

		resolved = append(resolved, resolvedMapping{
			src:          src,
			target:       match.Target,
			strategyName: match.StrategyName,
			strategyIdx:  match.StrategyIdx,
		})
	}

	return resolved, unmapped
}

// Phase 2: deduplicateMappings detects N:1 target conflicts and resolves them.
func (u *Updater) deduplicateMappings(
	ctx context.Context, mappings []resolvedMapping,
) ([]resolvedMapping, []duplicateConflict) {
	groups := make(map[TargetID][]resolvedMapping)
	for _, m := range mappings {
		tid := m.target.GetTargetID()
		groups[tid] = append(groups[tid], m)
	}

	kept := make([]resolvedMapping, 0, len(mappings))
	var conflicts []duplicateConflict

	for _, group := range groups {
		if len(group) == 1 {
			kept = append(kept, group[0])
			continue
		}
		winner, losers := u.resolveConflictGroup(ctx, group)
		kept = append(kept, winner)
		conflicts = append(conflicts, losers...)
	}

	return kept, conflicts
}

// resolveConflictGroup picks the best match from a group sharing the same target.
func (u *Updater) resolveConflictGroup(
	ctx context.Context, group []resolvedMapping,
) (resolvedMapping, []duplicateConflict) {
	// Sort: lowest strategyIdx wins; on tie, SameTitleWithTarget wins
	sort.SliceStable(group, func(i, j int) bool {
		if group[i].strategyIdx != group[j].strategyIdx {
			return group[i].strategyIdx < group[j].strategyIdx
		}
		iTitle := group[i].src.SameTitleWithTarget(ctx, group[i].target)
		jTitle := group[j].src.SameTitleWithTarget(ctx, group[j].target)
		return iTitle && !jTitle
	})

	winner := group[0]
	LogDebug(ctx, "[%s] Duplicate target %q claimed by %d sources, keeping %q via %s",
		u.Prefix, winner.target.GetTitle(), len(group), winner.src.GetTitle(), winner.strategyName)

	conflicts := make([]duplicateConflict, 0, len(group)-1)
	for _, loser := range group[1:] {
		conflicts = append(conflicts, duplicateConflict{
			loserSrc:    loser.src,
			winnerSrc:   winner.src,
			target:      winner.target,
			loserStrat:  loser.strategyName,
			winnerStrat: winner.strategyName,
		})
	}
	return winner, conflicts
}

// Phase 3: processResolvedMappings checks progress and updates/dry-runs kept mappings.
func (u *Updater) processResolvedMappings(
	ctx context.Context, kept []resolvedMapping, _ *SyncReport,
) {
	for _, m := range kept {
		select {
		case <-ctx.Done():
			LogWarn(ctx, "[%s] Context cancelled before update", u.Prefix)
			return
		default:
		}

		LogDebug(ctx, "[%s] Target: %s", u.Prefix, m.target.String())

		if m.src.SameProgressWithTarget(m.target) {
			LogDebug(ctx, "[%s] No changes needed, skipping", u.Prefix)
			u.Statistics.RecordSkip(UpdateResult{
				Title:      m.src.GetTitle(),
				Status:     m.src.GetStatusString(),
				Skipped:    true,
				SkipReason: "no changes",
			})
			continue
		}

		LogDebug(ctx, "[%s] Source: %s", u.Prefix, m.src.GetTitle())
		LogDebug(ctx, "[%s] Target: %s", u.Prefix, m.target.GetTitle())
		LogDebug(ctx, "[%s] Diff: %s", u.Prefix, m.src.GetStringDiffWithTarget(m.target))

		tgtID := m.target.GetTargetID()

		if u.DryRun {
			u.Statistics.RecordDryRun(UpdateResult{
				Title:  m.src.GetTitle(),
				Status: m.src.GetStatusString(),
				Detail: "dry run",
			})
			continue
		}

		u.updateTarget(ctx, tgtID, m.src)
	}
}

func (u *Updater) recordConflicts(
	ctx context.Context, conflicts []duplicateConflict, report *SyncReport,
) {
	for _, c := range conflicts {
		reason := fmt.Sprintf("duplicate: same target already matched by %q via %s",
			c.winnerSrc.GetTitle(), c.winnerStrat)
		LogDebug(ctx, "[%s] Duplicate conflict: %q lost to %q for target %q",
			u.Prefix, c.loserSrc.GetTitle(), c.winnerSrc.GetTitle(), c.target.GetTitle())

		u.Statistics.RecordSkip(UpdateResult{
			Title:      c.loserSrc.GetTitle(),
			Status:     c.loserSrc.GetStatusString(),
			Skipped:    true,
			SkipReason: reason,
		})
		u.trackUnmapped(c.loserSrc, reason)

		if report != nil {
			report.AddDuplicateConflict(
				c.loserSrc.GetTitle(),
				c.winnerSrc.GetTitle(),
				c.target.GetTitle(),
				c.loserStrat,
				c.winnerStrat,
				u.MediaType,
			)
		}
	}
}

func (u *Updater) recordUnmapped(ctx context.Context, unmapped []Source) {
	for _, src := range unmapped {
		u.trackUnmapped(src, "no matching entry found on target service")
	}
}

// GetResolvedMappings returns the source→target mappings from the last Update call.
// This is used for favorites sync to get mapped AniList IDs for MAL entries.
func (u *Updater) GetResolvedMappings() []resolvedMapping {
	return u.ResolvedMappings
}

func (u *Updater) updateSourceByTargets(ctx context.Context, src Source, tgts map[TargetID]Target, report *SyncReport) {
	tgtID := src.GetTargetID()

	if !u.ForceSync { // filter sources by different progress with targets
		// Use strategy chain to find target
		tgt, err := u.StrategyChain.FindTarget(ctx, src, tgts, u.Prefix, report)
		if err != nil {
			LogDebug(ctx, "[%s] Error finding target: %v", u.Prefix, err)
			u.Statistics.RecordSkip(UpdateResult{
				Title:      src.GetTitle(),
				Status:     src.GetStatusString(),
				Skipped:    true,
				SkipReason: "unmapped",
			})
			u.trackUnmapped(src, "no matching entry found on target service")
			return
		}

		LogDebug(ctx, "[%s] Target: %s", u.Prefix, tgt.String())

		if src.SameProgressWithTarget(tgt) {
			LogDebug(ctx, "[%s] No changes needed, skipping", u.Prefix)
			u.Statistics.RecordSkip(UpdateResult{
				Title:      src.GetTitle(),
				Status:     src.GetStatusString(),
				Skipped:    true,
				SkipReason: "no changes",
			})
			return
		}

		// Debug logging for verbose mode - details of source and target
		LogDebug(ctx, "[%s] Source: %s", u.Prefix, src.GetTitle())
		LogDebug(ctx, "[%s] Target: %s", u.Prefix, tgt.GetTitle())
		LogDebug(ctx, "[%s] Diff: %s", u.Prefix, src.GetStringDiffWithTarget(tgt))

		tgtID = tgt.GetTargetID()
	}

	if u.DryRun { // skip update if dry run
		// Record in statistics for summary, don't log each item individually
		u.Statistics.RecordDryRun(UpdateResult{
			Title:  src.GetTitle(),
			Status: src.GetStatusString(),
			Detail: "dry run",
		})
		return
	}

	// Check for context cancellation before update operation
	select {
	case <-ctx.Done():
		LogWarn(ctx, "[%s] Context cancelled before update", u.Prefix)
		return
	default:
	}

	u.updateTarget(ctx, tgtID, src)
}

func (u *Updater) updateTarget(ctx context.Context, id TargetID, src Source) {
	LogDebug(ctx, "[%s] Updating %s", u.Prefix, src.GetTitle())

	if err := u.Service.Update(ctx, id, src, u.Prefix); err != nil {
		u.Statistics.RecordError(UpdateResult{
			Title:  src.GetTitle(),
			Status: src.GetStatusString(),
			Error:  err,
		})
		LogDebug(ctx, "[%s] Error updating %s: %v", u.Prefix, src.GetTitle(), err)
		return
	}

	// Generate concise update message
	detail := generateUpdateDetail(src, id, u.Reverse)

	u.Statistics.RecordUpdate(UpdateResult{
		Title:  src.GetTitle(),
		Detail: detail,
		Status: src.GetStatusString(),
	})
}

// generateUpdateDetail generates a concise update detail string.
// reverse=true means MAL→AniList direction (tgtID is an AniList ID).
func generateUpdateDetail(src Source, tgtID TargetID, reverse bool) string {
	// Try to get both MAL and AniList IDs from source
	var malID, anilistID TargetID

	// Use type assertion to get both IDs if available
	if anime, ok := src.(Anime); ok {
		malID = TargetID(anime.IDMal)
		anilistID = TargetID(anime.IDAnilist)
	} else if manga, ok := src.(Manga); ok {
		malID = TargetID(manga.IDMal)
		anilistID = TargetID(manga.IDAnilist)
	}

	// In reverse sync (MAL -> AniList), tgtID is the AniList ID
	// In forward sync (AniList -> MAL), tgtID is the MAL ID
	if reverse {
		anilistID = tgtID // Use the found AniList ID
	} else {
		malID = tgtID // Use the found MAL ID
	}

	// Show both IDs if available and different
	switch {
	case malID > 0 && anilistID > 0 && malID != anilistID:
		return fmt.Sprintf("(MAL: %d, AniList: %d)", malID, anilistID)
	case anilistID > 0:
		return fmt.Sprintf("(AniList: %d)", anilistID)
	case malID > 0:
		return fmt.Sprintf("(MAL: %d)", malID)
	}
	return fmt.Sprintf("(ID: %d)", tgtID)
}

func (u *Updater) trackUnmapped(src Source, reason string) {
	direction := SyncDirectionForward
	if u.Reverse {
		direction = SyncDirectionReverse
	}
	entry := UnmappedEntry{
		Title:     src.GetTitle(),
		MediaType: u.MediaType,
		Direction: direction.String(),
		Reason:    reason,
	}
	switch v := src.(type) {
	case Anime:
		entry.AniListID = v.IDAnilist
		entry.MALID = v.IDMal
	case Manga:
		entry.AniListID = v.IDAnilist
		entry.MALID = v.IDMal
	}
	u.UnmappedList = append(u.UnmappedList, entry)
}


