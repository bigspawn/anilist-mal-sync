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
	GetTitle() string
	GetStringDiffWithTarget(Target) string
	SameProgressWithTarget(Target) bool
	SameTypeWithTarget(Target) bool
	String() string
}

type Target interface {
	GetTargetID() TargetID
	String() string
}

type Updater struct {
	Prefix       string
	Statistics   *Statistics
	IgnoreTitles map[string]struct{}

	GetTargetByIDFunc        func(context.Context, TargetID) (Target, error)
	GetTargetsByNameFunc     func(context.Context, string) ([]Target, error)
	UpdateTargetBySourceFunc func(context.Context, TargetID, Source) error
}

func (u *Updater) Update(ctx context.Context, srcs []Source, tgts []Target) {
	tgtsByID := make(map[TargetID]Target, len(tgts))
	for _, tgt := range tgts {
		tgtsByID[tgt.GetTargetID()] = tgt
	}

	var statusStr string
	for _, src := range srcs {
		if src.GetStatusString() == "" {
			continue
		}

		u.Statistics.TotalCount++

		if statusStr != src.GetStatusString() {
			statusStr = src.GetStatusString()
			log.Printf("[%s] Processing for status: %s", u.Prefix, statusStr)
		}

		DPrintf("[%s] Processing for: %s", u.Prefix, src.String())

		if _, ok := u.IgnoreTitles[strings.ToLower(src.GetTitle())]; ok {
			log.Printf("[%s] Ignoring anime: %s", u.Prefix, src.GetTitle())
			u.Statistics.SkippedCount++
			continue
		}

		u.updateSourceByTargets(ctx, src, tgtsByID)
	}
}

func (u *Updater) updateSourceByTargets(ctx context.Context, src Source, tgts map[TargetID]Target) {
	tgtID := src.GetTargetID()

	if !(*forceSync) { // filter sources by different progress with targets
		tgt, ok := tgts[src.GetTargetID()]
		if !ok {
			var err error
			tgt, err = u.findTarget(ctx, src)
			if err != nil {
				log.Printf("[%s] Error processing target anime: %v", u.Prefix, err)
				u.Statistics.SkippedCount++
				return
			}
		}

		DPrintf("[%s] Target: %s", u.Prefix, tgt.String())

		if src.SameProgressWithTarget(tgt) {
			u.Statistics.SkippedCount++
			return
		}

		log.Printf("[%s] Title: %s", u.Prefix, src.GetTitle())
		log.Printf("[%s] Progress is not same, need to update: %s", u.Prefix, src.GetStringDiffWithTarget(tgt))

		tgtID = tgt.GetTargetID()
	}

	if *dryRun { // skip update if dry run
		log.Printf("[%s] Dry run: Skipping update for %s", u.Prefix, src.GetTitle())
		return
	}

	u.updateTarget(ctx, tgtID, src)
}

func (u *Updater) findTarget(ctx context.Context, src Source) (Target, error) {
	tgtID := src.GetTargetID()

	if tgtID > 0 {
		DPrintf("[%s] Finding target by id: %d", u.Prefix, tgtID)

		tgt, err := u.GetTargetByIDFunc(ctx, tgtID)
		if err != nil {
			return nil, fmt.Errorf("error getting mal anime by id: %s: %w", src.GetTitle(), err)
		}
		return tgt, nil
	}

	DPrintf("[%s] Finding target by name: %s", u.Prefix, src.GetTitle())

	tgts, err := u.GetTargetsByNameFunc(ctx, src.GetTitle())
	if err != nil {
		return nil, fmt.Errorf("error getting targets by source name: %s: %w", src.GetTitle(), err)
	}

	for _, tgt := range tgts {
		if src.SameTypeWithTarget(tgt) {
			DPrintf("[%s] Found target by name: %s", u.Prefix, src.GetTitle())
			return tgt, nil
		}
		DPrintf("[%s] Ignoring target by name: %s", u.Prefix, tgt.String())
	}

	return nil, fmt.Errorf("no target found for source: %s", src.GetTitle())
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
