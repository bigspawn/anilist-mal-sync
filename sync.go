package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
)

// Statistics tracks sync operation results
type Statistics struct {
	UpdatedCount  int // Successfully updated
	SkippedCount  int // Already in sync (no update needed)
	NotFoundCount int // Could not match in target service
	ErrorCount    int // Failed to update due to errors
	TotalCount    int
}

// Reset resets all statistics counters to zero
func (s *Statistics) Reset() {
	s.UpdatedCount = 0
	s.SkippedCount = 0
	s.NotFoundCount = 0
	s.ErrorCount = 0
	s.TotalCount = 0
}

// Print prints statistics with the given prefix for logging context
func (s Statistics) Print(prefix string) {
	if s.TotalCount > 0 {
		updatedPct := float64(s.UpdatedCount) / float64(s.TotalCount) * PercentMultiplier
		log.Printf("[%s] Updated %d out of %d (%.1f%%)", prefix, s.UpdatedCount, s.TotalCount, updatedPct)
	} else {
		log.Printf("[%s] Updated %d out of %d", prefix, s.UpdatedCount, s.TotalCount)
	}
	log.Printf("[%s] Skipped %d (already in sync)", prefix, s.SkippedCount)
	if s.NotFoundCount > 0 {
		log.Printf("[%s] Not found %d (could not match in target service)", prefix, s.NotFoundCount)
	}
	if s.ErrorCount > 0 {
		log.Printf("[%s] Errors %d (failed to update)", prefix, s.ErrorCount)
	}
}

// Updater handles syncing sources to targets using a strategy chain
type Updater struct {
	Prefix        string
	Statistics    *Statistics
	IgnoreTitles  map[string]struct{}
	StrategyChain *StrategyChain

	Verbose   bool
	DryRun    bool
	ForceSync bool

	UpdateTargetBySourceFunc func(context.Context, TargetID, Source) error
}

// NewUpdater creates a new Updater with the specified configuration
func NewUpdater(prefix string, stats *Statistics, ignore map[string]struct{}, sc *StrategyChain, verbose, dryRun, forceSync bool) *Updater {
	return &Updater{
		Prefix:        prefix,
		Statistics:    stats,
		IgnoreTitles:  ignore,
		StrategyChain: sc,
		Verbose:       verbose,
		DryRun:        dryRun,
		ForceSync:     forceSync,
	}
}

// DPrintf prints a debug message only if verbose mode is enabled
func (u *Updater) DPrintf(format string, v ...any) {
	if !u.Verbose {
		return
	}
	log.Printf(format, v...)
}

func (u *Updater) Update(ctx context.Context, srcs []Source, tgts []Target) {
	if u.Statistics == nil {
		log.Printf("[%s] Error: Statistics is not set for updater", u.Prefix)
		return
	}

	tgtsByID := make(map[TargetID]Target, len(tgts))
	for _, tgt := range tgts {
		if isNil(tgt) {
			u.DPrintf("[%s] Warning: Skipping nil target", u.Prefix)
			continue
		}
		tgtID := tgt.GetTargetID()
		if existing, exists := tgtsByID[tgtID]; exists {
			// Log warning if duplicate IDs detected (shouldn't happen in practice)
			u.DPrintf("[%s] Warning: Duplicate target ID %d detected. Overwriting: %s -> %s", u.Prefix, tgtID, existing.GetTitle(), tgt.GetTitle())
		}
		tgtsByID[tgtID] = tgt
	}

	var statusStr string
	for _, src := range srcs {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Context cancelled, stopping sync", u.Prefix)
			return
		default:
		}

		if isNil(src) {
			u.DPrintf("[%s] Warning: Skipping nil source", u.Prefix)
			continue
		}

		if src.GetStatusString() == "" {
			u.DPrintf("[%s] Skipping source with empty status: %s", u.Prefix, src.String())
			continue
		}

		u.Statistics.TotalCount++

		if statusStr != src.GetStatusString() {
			statusStr = src.GetStatusString()
			log.Printf("[%s] Processing for status: %s", u.Prefix, statusStr)
		}

		u.DPrintf("[%s] Processing for: %s", u.Prefix, src.String())

		// Check if title should be ignored (nil map check is unnecessary - map access returns zero value)
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

	if !u.ForceSync {
		if u.StrategyChain == nil {
			log.Printf("[%s] Error: StrategyChain is not set for updater", u.Prefix)
			u.Statistics.ErrorCount++
			return
		}
		tgt, err := u.StrategyChain.FindTarget(ctx, src, tgts, u.Prefix)
		if err != nil {
			log.Printf("[%s] Error finding target: %v", u.Prefix, err)
			u.Statistics.NotFoundCount++
			return
		}

		if isNil(tgt) {
			log.Printf("[%s] Error: FindTarget returned nil target for source: %s", u.Prefix, src.GetTitle())
			u.Statistics.NotFoundCount++
			return
		}

		u.DPrintf("[%s] Target: %s", u.Prefix, tgt.String())

		if src.SameProgressWithTarget(tgt) {
			u.Statistics.SkippedCount++
			return
		}

		log.Printf("[%s] Src title: %s", u.Prefix, src.GetTitle())
		log.Printf("[%s] Tgt title: %s", u.Prefix, tgt.GetTitle())
		log.Printf("[%s] Progress is not same, need to update: %s", u.Prefix, src.GetStringDiffWithTarget(tgt))

		tgtID = tgt.GetTargetID()
	}

	if u.DryRun {
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
	u.DPrintf("[%s] Updating %s", u.Prefix, src.GetTitle())

	if u.UpdateTargetBySourceFunc == nil {
		log.Printf("[%s] Error: UpdateTargetBySourceFunc is not set for updater", u.Prefix)
		u.Statistics.ErrorCount++
		return
	}

	if err := u.UpdateTargetBySourceFunc(ctx, id, src); err != nil {
		log.Printf("[%s] Error updating target: %s: %v", u.Prefix, src.GetTitle(), err)
		u.Statistics.ErrorCount++
		return
	}

	log.Printf("[%s] Updated %s", u.Prefix, src.GetTitle())

	u.Statistics.UpdatedCount++
}

// TargetFindStrategy defines a strategy for finding matching targets
type TargetFindStrategy interface {
	FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error)
	Name() string
}

// StrategyChain tries multiple strategies in order until one succeeds
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
		if found && !isNil(target) {
			return target, nil
		}
	}
	return nil, fmt.Errorf("no target found for source: %s", src.GetTitle())
}

// IDStrategy finds targets by matching IDs
type IDStrategy struct{}

func (s IDStrategy) Name() string { return "IDStrategy" }

func (s IDStrategy) FindTarget(_ context.Context, src Source, existingTargets map[TargetID]Target, _ string) (Target, bool, error) {
	target, found := existingTargets[src.GetTargetID()]
	// Only return found=true if target is not nil
	return target, found && !isNil(target), nil
}

// TitleStrategy finds targets by matching titles
type TitleStrategy struct{}

func (s TitleStrategy) Name() string { return "TitleStrategy" }

func (s TitleStrategy) FindTarget(_ context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error) {
	srcTitle := src.GetTitle()
	if srcTitle == "" {
		// Early return if source title is empty - no point searching
		return nil, false, nil
	}

	targetSlice := make([]Target, 0, len(existingTargets))
	for _, target := range existingTargets {
		if !isNil(target) {
			targetSlice = append(targetSlice, target)
		}
	}

	if len(targetSlice) == 0 {
		return nil, false, nil
	}

	sort.Slice(targetSlice, func(i, j int) bool {
		return targetSlice[i].GetTitle() < targetSlice[j].GetTitle()
	})

	// Pre-allocate map with capacity estimate to reduce reallocations
	byTitle := make(map[string]Target, len(targetSlice))
	for _, target := range targetSlice {
		title := target.GetTitle()
		if existing, exists := byTitle[title]; exists {
			// Log warning if duplicate titles detected (will fall back to fuzzy matching)
			// This is less critical than duplicate IDs since we have fallback logic
			// Note: Using log.Printf instead of DPrintf since TitleStrategy doesn't have access to Updater
			log.Printf("[%s] Warning: Duplicate title '%s' detected. Exact match may be ambiguous: %s -> %s", prefix, title, existing.GetTitle(), target.GetTitle())
		}
		byTitle[title] = target
	}

	if target, ok := byTitle[srcTitle]; ok {
		return target, true, nil
	}

	for _, target := range targetSlice {
		if src.SameTitleWithTarget(target) && src.SameTypeWithTarget(target) {
			return target, true, nil
		}
	}

	return nil, false, nil
}

// APISearchStrategy finds targets by querying the target service API
type APISearchStrategy struct {
	GetTargetByIDFunc     func(context.Context, TargetID) (Target, error)
	GetTargetsByNameFunc  func(context.Context, string) ([]Target, error)
	GetTargetsByMALIDFunc func(context.Context, int) ([]Target, error) // Optional: for reverse sync
}

func (s APISearchStrategy) Name() string { return "APISearchStrategy" }

func (s APISearchStrategy) FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, _ string) (Target, bool, error) {
	select {
	case <-ctx.Done():
		return nil, false, fmt.Errorf("context cancelled during API search")
	default:
	}

	tgtID := src.GetTargetID()

	// Try 1: Direct ID lookup
	if tgtID > 0 && s.GetTargetByIDFunc != nil {
		target, err := s.GetTargetByIDFunc(ctx, tgtID)
		if err != nil {
			return nil, false, fmt.Errorf("error getting target by ID %d: %w", tgtID, err)
		}

		if isNil(target) {
			// Function returned nil target - treat as not found
			return nil, false, nil
		}

		if existingTarget, exists := existingTargets[tgtID]; exists {
			return existingTarget, true, nil
		}

		return target, true, nil
	}

	// Try 2: Search by MAL ID (for reverse sync: MAL -> AniList)
	// Forward sync already uses MAL IDs via GetTargetID() above
	if s.GetTargetsByMALIDFunc != nil {
		if malID := extractMALIDFromSource(src); malID > 0 {
			targets, err := s.GetTargetsByMALIDFunc(ctx, malID)
			if err == nil && len(targets) > 0 {
				// Match by title to ensure correctness
				for _, target := range targets {
					if !isNil(target) && src.SameTitleWithTarget(target) && src.SameTypeWithTarget(target) {
						return target, true, nil
					}
				}
				// Single result from MAL ID search is usually reliable
				if len(targets) == 1 && !isNil(targets[0]) {
					return targets[0], true, nil
				}
			}
		}
	}

	// Try 3: Search by name (fallback)
	if s.GetTargetsByNameFunc == nil {
		return nil, false, fmt.Errorf("GetTargetsByNameFunc is not set")
	}
	srcTitle := src.GetTitle()
	if srcTitle == "" {
		// Skip search if title is empty - would be wasteful and unlikely to find matches
		return nil, false, nil
	}
	targets, err := s.GetTargetsByNameFunc(ctx, srcTitle)
	if err != nil {
		return nil, false, fmt.Errorf("error searching targets by name '%s': %w", srcTitle, err)
	}

	if len(targets) == 0 {
		return nil, false, nil
	}

	for _, tgt := range targets {
		if isNil(tgt) {
			continue
		}

		// Prefer existing target if available (has more complete data)
		if existingTarget, exists := existingTargets[tgt.GetTargetID()]; exists {
			tgt = existingTarget
		}

		if src.SameTitleWithTarget(tgt) && src.SameTypeWithTarget(tgt) {
			return tgt, true, nil
		}
	}

	return nil, false, nil
}

// extractMALIDFromSource extracts the MAL ID from a Source for reverse sync
func extractMALIDFromSource(src Source) int {
	if unwrapped, ok := safeUnwrapSourceAdapter(src); ok {
		if anime, ok := unwrapped.(Anime); ok {
			return anime.IDMal
		}
		if manga, ok := unwrapped.(Manga); ok {
			return manga.IDMal
		}
	}
	return 0
}
