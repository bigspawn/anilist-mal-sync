package main

import (
	"context"
	"log"
	"strings"
)

type TargetID int

type Source interface {
	GetStatusString() string
	GetTargetID() TargetID
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
	for _, src := range srcs {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			log.Printf("[%s] Context cancelled, stopping sync", u.Prefix)
			return
		default:
		}

		if src.GetStatusString() == "" {
			DPrintf("[%s] Skipping source with empty status: %s", u.Prefix, src.String())
			continue
		}

		u.Statistics.TotalCount++

		if statusStr != src.GetStatusString() {
			statusStr = src.GetStatusString()
			log.Printf("[%s] Processing for status: %s", u.Prefix, statusStr)
		}

		DPrintf("[%s] Processing for: %s", u.Prefix, src.String())

		if _, ok := u.IgnoreTitles[strings.ToLower(src.GetTitle())]; ok {
			log.Printf("[%s] Ignoring entry: %s", u.Prefix, src.GetTitle())
			u.Statistics.SkippedCount++
			continue
		}

		u.updateSourceByTargets(ctx, src, tgtsByID)
	}
}

func (u *Updater) updateSourceByTargets(ctx context.Context, src Source, tgts map[TargetID]Target) {
	tgtID := src.GetTargetID()

	if !(*forceSync) { // filter sources by different progress with targets
		// Use strategy chain to find target
		tgt, err := u.StrategyChain.FindTarget(ctx, src, tgts, u.Prefix)
		if err != nil {
			log.Printf("[%s] Error finding target: %v", u.Prefix, err)
			u.Statistics.SkippedCount++
			return
		}

		DPrintf("[%s] Target: %s", u.Prefix, tgt.String())

		if src.SameProgressWithTarget(tgt) {
			u.Statistics.SkippedCount++
			return
		}

		log.Printf("[%s] Src title: %s", u.Prefix, src.String() /*GetTitle()*/)
		log.Printf("[%s] Tgt title: %s", u.Prefix, tgt.String() /*GetTitle()*/)
		log.Printf("[%s] Progress is not same, need to update: %s", u.Prefix, src.GetStringDiffWithTarget(tgt))

		tgtID = tgt.GetTargetID()
	}

	if *dryRun { // skip update if dry run
		log.Printf("[%s] Dry run: Skipping update for %s", u.Prefix, src.GetTitle())
		return
	}

	// Check for context cancellation before update operation
	select {
	case <-ctx.Done():
		log.Printf("[%s] Context cancelled before update", u.Prefix)
		return
	default:
	}

	u.updateTarget(ctx, tgtID, src)
}

func (u *Updater) updateTarget(ctx context.Context, id TargetID, src Source) {
	DPrintf("[%s] Updating %s", u.Prefix, src.GetTitle())

	if err := u.UpdateTargetBySourceFunc(ctx, id, src); err != nil {
		log.Printf("[%s] Error updating target: %s: %v", u.Prefix, src.GetTitle(), err)
		return
	}

	log.Printf("[%s] Updated %s", u.Prefix, src.GetTitle())

	u.Statistics.UpdatedCount++
}

func DPrintf(format string, v ...any) {
	if !(*verbose) {
		return
	}
	log.Printf(format, v...)
}
