package main

//go:generate mockgen -destination mock_updater_test.go -package main -source=updater.go

import (
	"context"
	"fmt"
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
	SameTypeWithTarget(Target) bool
	SameTitleWithTarget(Target) bool
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
	StrategyChain *StrategyChain
	Service       MediaService // Replaces callback
	ForceSync     bool         // Skip matching logic, force sync all
	DryRun        bool         // Skip actual updates
}

func (u *Updater) Update(ctx context.Context, srcs []Source, tgts []Target, report *SyncReport) {
	tgtsByID := make(map[TargetID]Target, len(tgts))
	for _, tgt := range tgts {
		tgtsByID[tgt.GetTargetID()] = tgt
	}

	var statusStr string
	processedCount := 0

	for _, src := range srcs {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			LogWarn(ctx, "[%s] Context cancelled, stopping sync", u.Prefix)
			return
		default:
		}

		if src.GetStatusString() == "" {
			LogDebug(ctx, "[%s] Skipping source with empty status: %s", u.Prefix, src.String())
			continue
		}

		u.Statistics.IncrementTotal()
		processedCount++

		// Show status transitions
		if statusStr != src.GetStatusString() {
			statusStr = src.GetStatusString()
			// Shorten prefix for cleaner output
			shortPrefix := strings.TrimPrefix(u.Prefix, "AniList to MAL ")
			shortPrefix = strings.TrimPrefix(shortPrefix, "MAL to AniList ")
			LogStage(ctx, "[%s] Processing %s...", shortPrefix, statusStr)
		}

		// Show progress (overwrites previous line)
		LogProgress(ctx, processedCount, len(srcs), statusStr, src.GetTitle())

		LogDebug(ctx, "[%s] Processing: %s", u.Prefix, src.String())

		// Check ignore list
		if _, ok := u.IgnoreTitles[strings.ToLower(src.GetTitle())]; ok {
			LogDebug(ctx, "[%s] Ignoring entry: %s", u.Prefix, src.GetTitle())
			u.Statistics.RecordSkip(UpdateResult{
				Title:      src.GetTitle(),
				Status:     src.GetStatusString(),
				Skipped:    true,
				SkipReason: "in ignore list",
			})
			continue
		}

		u.updateSourceByTargets(ctx, src, tgtsByID, report)
	}
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
				SkipReason: "target not found",
			})
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
		u.Statistics.RecordUpdate(UpdateResult{
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
	detail := generateUpdateDetail(src, id)

	u.Statistics.RecordUpdate(UpdateResult{
		Title:  src.GetTitle(),
		Detail: detail,
		Status: src.GetStatusString(),
	})

	// Single-line success message (replaces 3-4 lines)
	LogInfoUpdate(ctx, src.GetTitle(), detail)
}

// generateUpdateDetail generates a concise update detail string
func generateUpdateDetail(src Source, tgtID TargetID) string {
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
	if *reverseDirection {
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

// DPrintf is deprecated - use LogDebug with context instead
func DPrintf(_ string, _ ...any) {
	// Deprecated: use LogDebug(ctx, ...) instead
	// This function is kept for backward compatibility but does nothing
}
