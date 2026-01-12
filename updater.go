package main

import (
	"context"
	"fmt"
	"log"
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
	Logger        *Logger

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
			u.Logger.Warn("[%s] Context cancelled, stopping sync", u.Prefix)
			return
		default:
		}

		if src.GetStatusString() == "" {
			u.Logger.Debug("[%s] Skipping source with empty status: %s", u.Prefix, src.String())
			continue
		}

		u.Statistics.IncrementTotal()
		processedCount++

		// Show status transitions
		if statusStr != src.GetStatusString() {
			statusStr = src.GetStatusString()
			u.Logger.Stage("[%s] Processing %s...", u.Prefix, statusStr)
		}

		// Show progress (overwrites previous line)
		u.Logger.Progress(processedCount, len(srcs), statusStr)

		u.Logger.Debug("[%s] Processing: %s", u.Prefix, src.String())

		// Check ignore list
		if _, ok := u.IgnoreTitles[strings.ToLower(src.GetTitle())]; ok {
			u.Logger.Debug("[%s] Ignoring entry: %s", u.Prefix, src.GetTitle())
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
			u.Logger.Debug("[%s] Error finding target: %v", u.Prefix, err)
			u.Statistics.RecordSkip(UpdateResult{
				Title:      src.GetTitle(),
				Status:     src.GetStatusString(),
				Skipped:    true,
				SkipReason: "target not found",
			})
			return
		}

		u.Logger.Debug("[%s] Target: %s", u.Prefix, tgt.String())

		if src.SameProgressWithTarget(tgt) {
			u.Logger.Debug("[%s] No changes needed, skipping", u.Prefix)
			u.Statistics.RecordSkip(UpdateResult{
				Title:      src.GetTitle(),
				Status:     src.GetStatusString(),
				Skipped:    true,
				SkipReason: "no changes",
			})
			return
		}

		// Debug logging for verbose mode - details of source and target
		u.Logger.Debug("[%s] Source: %s", u.Prefix, src.GetTitle())
		u.Logger.Debug("[%s] Target: %s", u.Prefix, tgt.GetTitle())
		u.Logger.Debug("[%s] Diff: %s", u.Prefix, src.GetStringDiffWithTarget(tgt))

		tgtID = tgt.GetTargetID()
	}

	if *dryRun {
		u.Logger.Info("[%s] Dry run: Skipping update for %s", u.Prefix, src.GetTitle())
		return
	}

	// Check for context cancellation before update operation
	select {
	case <-ctx.Done():
		u.Logger.Warn("[%s] Context cancelled before update", u.Prefix)
		return
	default:
	}

	u.updateTarget(ctx, tgtID, src)
}

func (u *Updater) updateTarget(ctx context.Context, id TargetID, src Source) {
	u.Logger.Debug("[%s] Updating %s", u.Prefix, src.GetTitle())

	if err := u.UpdateTargetBySourceFunc(ctx, id, src); err != nil {
		u.Statistics.RecordError(UpdateResult{
			Title:  src.GetTitle(),
			Status: src.GetStatusString(),
			Error:  err,
		})
		u.Logger.Debug("[%s] Error updating %s: %v", u.Prefix, src.GetTitle(), err)
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
	u.Logger.InfoUpdate(src.GetTitle(), detail)
}

// generateUpdateDetail generates a concise update detail string
func generateUpdateDetail(src Source) string {
	return fmt.Sprintf("(ID: %d)", src.GetTargetID())
}

func DPrintf(format string, v ...any) {
	// Keep for backward compatibility - will be replaced with logger
	if !(*verbose) {
		return
	}
	log.Printf(format, v...)
}
