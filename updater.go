package main

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

	UpdateTargetBySourceFunc func(context.Context, TargetID, Source) error
}

func (u *Updater) Update(ctx context.Context, srcs []Source, tgts []Target) {
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
			LogStage(ctx, "[%s] Processing %s...", u.Prefix, statusStr)
		}

		// Show progress (overwrites previous line)
		LogProgress(ctx, processedCount, len(srcs), statusStr)

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

		u.updateSourceByTargets(ctx, src, tgtsByID)
	}
}

func (u *Updater) updateSourceByTargets(ctx context.Context, src Source, tgts map[TargetID]Target) {
	tgtID := src.GetTargetID()

	if !(*forceSync) {
		// Use strategy chain to find target
		tgt, err := u.StrategyChain.FindTarget(ctx, src, tgts, u.Prefix)
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

	if *dryRun {
		LogInfo(ctx, "[%s] Dry run: Skipping update for %s", u.Prefix, src.GetTitle())
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

	if err := u.UpdateTargetBySourceFunc(ctx, id, src); err != nil {
		u.Statistics.RecordError(UpdateResult{
			Title:  src.GetTitle(),
			Status: src.GetStatusString(),
			Error:  err,
		})
		LogDebug(ctx, "[%s] Error updating %s: %v", u.Prefix, src.GetTitle(), err)
		return
	}

	// Generate concise update message
	detail := generateUpdateDetail(src)

	u.Statistics.RecordUpdate(UpdateResult{
		Title:  src.GetTitle(),
		Detail: detail,
		Status: src.GetStatusString(),
	})

	// Single-line success message (replaces 3-4 lines)
	LogInfoUpdate(ctx, src.GetTitle(), detail)
}

// generateUpdateDetail generates a concise update detail string
func generateUpdateDetail(src Source) string {
	return fmt.Sprintf("(ID: %d)", src.GetTargetID())
}

// DPrintf is deprecated - use LogDebug with context instead
func DPrintf(_ string, _ ...any) {
	// Deprecated: use LogDebug(ctx, ...) instead
	// This function is kept for backward compatibility but does nothing
}
