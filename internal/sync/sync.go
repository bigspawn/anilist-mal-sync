// Package syncer implements the core sync logic including updaters, strategies, and statistics.
package syncer

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Package-level flags controlled by NewUpdater
var (
	Verbose   = false
	DryRun    = false
	ForceSync = false
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

type Statistics struct {
	UpdatedCount int
	SkippedCount int
	TotalCount   int
}

func (s Statistics) Print(prefix string) {
	log.Printf("[%s] Updated %d out of %d\n", prefix, s.UpdatedCount, s.TotalCount)
	log.Printf("[%s] Skipped %d\n", prefix, s.SkippedCount)
}

func DPrintf(format string, v ...any) {
	if !Verbose {
		return
	}
	log.Printf(format, v...)
}

type Updater struct {
	Prefix        string
	Statistics    *Statistics
	IgnoreTitles  map[string]struct{}
	StrategyChain *StrategyChain

	UpdateTargetBySourceFunc func(context.Context, TargetID, Source) error
}

func NewUpdater(prefix string, stats *Statistics, ignore map[string]struct{}, sc *StrategyChain, opts ...func()) *Updater {
	// apply options
	for _, o := range opts {
		o()
	}
	return &Updater{Prefix: prefix, Statistics: stats, IgnoreTitles: ignore, StrategyChain: sc}
}

func (u *Updater) Update(ctx context.Context, srcs []Source, tgts []Target) {
	tgtsByID := make(map[TargetID]Target, len(tgts))
	for _, tgt := range tgts {
		tgtsByID[tgt.GetTargetID()] = tgt
	}

	var statusStr string
	for _, src := range srcs {
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

	if !ForceSync {
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

		log.Printf("[%s] Src title: %s", u.Prefix, src.GetTitle())
		log.Printf("[%s] Tgt title: %s", u.Prefix, tgt.GetTitle())
		log.Printf("[%s] Progress is not same, need to update: %s", u.Prefix, src.GetStringDiffWithTarget(tgt))

		tgtID = tgt.GetTargetID()
	}

	if DryRun {
		log.Printf("[%s] Dry run: Skipping update for %s", u.Prefix, src.GetTitle())
		u.Statistics.SkippedCount++
		return
	}

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

// StrategyChain and strategies
type TargetFindStrategy interface {
	FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error)
	Name() string
}

type StrategyChain struct {
	strategies []TargetFindStrategy
}

func NewStrategyChain(strategies ...TargetFindStrategy) *StrategyChain {
	return &StrategyChain{strategies: strategies}
}

func (sc *StrategyChain) FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, error) {
	for _, strategy := range sc.strategies {
		target, found, err := strategy.FindTarget(ctx, src, existingTargets, prefix)
		if err != nil {
			return nil, fmt.Errorf("strategy %s failed: %w", strategy.Name(), err)
		}
		if found {
			DPrintf("[%s] Found target using strategy: %s", prefix, strategy.Name())
			return target, nil
		}
	}
	return nil, fmt.Errorf("no target found for source: %s", src.GetTitle())
}

type IDStrategy struct{}

func (s IDStrategy) Name() string { return "IDStrategy" }

func (s IDStrategy) FindTarget(_ context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error) {
	target, found := existingTargets[src.GetTargetID()]
	if found {
		DPrintf("[%s] Found target by ID: %d", prefix, src.GetTargetID())
	}
	return target, found, nil
}

type TitleStrategy struct{}

func (s TitleStrategy) Name() string { return "TitleStrategy" }

func (s TitleStrategy) FindTarget(_ context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error) {
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
		DPrintf("[%s] Found target by title: %s", prefix, srcTitle)
		return target, true, nil
	}

	for _, target := range targetSlice {
		if src.SameTitleWithTarget(target) && src.SameTypeWithTarget(target) {
			DPrintf("[%s] Found target by title comparison: %s", prefix, srcTitle)
			return target, true, nil
		}
	}

	DPrintf("[%s] No target found by title comparison: %s", prefix, srcTitle)
	return nil, false, nil
}

type APISearchStrategy struct {
	GetTargetByIDFunc    func(context.Context, TargetID) (Target, error)
	GetTargetsByNameFunc func(context.Context, string) ([]Target, error)
}

func (s APISearchStrategy) Name() string { return "APISearchStrategy" }

func (s APISearchStrategy) FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error) {
	select {
	case <-ctx.Done():
		return nil, false, fmt.Errorf("context cancelled during API search")
	default:
	}

	tgtID := src.GetTargetID()

	if tgtID > 0 {
		DPrintf("[%s] Finding target by API ID: %d", prefix, tgtID)
		target, err := s.GetTargetByIDFunc(ctx, tgtID)
		if err != nil {
			return nil, false, fmt.Errorf("error getting target by ID %d: %w", tgtID, err)
		}

		if existingTarget, exists := existingTargets[tgtID]; exists {
			DPrintf("[%s] Found existing user target by ID: %d", prefix, tgtID)
			return existingTarget, true, nil
		}

		DPrintf("[%s] Found target by API ID: %d", prefix, tgtID)
		return target, true, nil
	}

	DPrintf("[%s] Finding target by API name: %s", prefix, src.GetTitle())
	targets, err := s.GetTargetsByNameFunc(ctx, src.GetTitle())
	if err != nil {
		return nil, false, fmt.Errorf("error getting targets by name %s: %w", src.GetTitle(), err)
	}

	for i, tgt := range targets {
		if existingTarget, exists := existingTargets[tgt.GetTargetID()]; exists {
			DPrintf("[%s] Found existing user target by API name: %s: %d", prefix, tgt.GetTitle(), i+1)
			tgt = existingTarget
		}

		if src.SameTypeWithTarget(tgt) {
			DPrintf("[%s] Found target by API name: %s: %d", prefix, tgt.String(), i+1)
			return tgt, true, nil
		}
		DPrintf("[%s] Ignoring target by API name: %s: %d (type mismatch)", prefix, tgt.GetTitle(), i+1)
	}

	return nil, false, nil
}

// Backoff helper
func createBackoffPolicy() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Second
	b.MaxInterval = 30 * time.Second
	b.MaxElapsedTime = 2 * time.Minute
	b.Multiplier = 2.0
	b.RandomizationFactor = 0.1
	return b
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "too many requests") || strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "429")
}

func retryWithBackoff(ctx context.Context, operation func() error, operationName string, prefix ...string) error {
	b := createBackoffPolicy()

	retryableOperation := func() error {
		err := operation()
		if err != nil && !isRateLimitError(err) {
			return backoff.Permanent(err)
		}
		return err
	}

	return backoff.RetryNotify(retryableOperation, backoff.WithContext(b, ctx), func(err error, duration time.Duration) {
		if isRateLimitError(err) {
			if len(prefix) > 0 {
				log.Printf("[%s] Rate limit hit for %s, retrying in %v: %v", prefix[0], operationName, duration, err)
			} else {
				log.Printf("Rate limit hit for %s, retrying in %v: %v", operationName, duration, err)
			}
		}
	})
}
