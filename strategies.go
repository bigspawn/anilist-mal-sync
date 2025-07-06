package main

import (
	"context"
	"fmt"
)

// TargetFindStrategy defines a strategy for finding targets
type TargetFindStrategy interface {
	FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error)
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

// IDStrategy finds targets by TargetID in existing targets map
type IDStrategy struct{}

func (s IDStrategy) Name() string {
	return "IDStrategy"
}

func (s IDStrategy) FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error) {
	target, found := existingTargets[src.GetTargetID()]
	if found {
		DPrintf("[%s] Found target by ID: %d", prefix, src.GetTargetID())
	}
	return target, found, nil
}

// TitleStrategy finds targets by title comparison with existing targets
type TitleStrategy struct{}

func (s TitleStrategy) Name() string {
	return "TitleStrategy"
}

func (s TitleStrategy) FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error) {
	srcTitle := src.GetTitle()
	
	for _, target := range existingTargets {
		if src.SameTitleWithTarget(target) && src.SameTypeWithTarget(target) {
			DPrintf("[%s] Found target by title comparison: %s", prefix, srcTitle)
			return target, true, nil
		}
	}
	
	DPrintf("[%s] No target found by title comparison: %s", prefix, srcTitle)
	return nil, false, nil
}

// APISearchStrategy finds targets by making API calls
type APISearchStrategy struct {
	GetTargetByIDFunc   func(context.Context, TargetID) (Target, error)
	GetTargetsByNameFunc func(context.Context, string) ([]Target, error)
}

func (s APISearchStrategy) Name() string {
	return "APISearchStrategy"
}

func (s APISearchStrategy) FindTarget(ctx context.Context, src Source, existingTargets map[TargetID]Target, prefix string) (Target, bool, error) {
	// Check for context cancellation before potentially long-running search
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
		return target, true, nil
	}

	DPrintf("[%s] Finding target by API name: %s", prefix, src.GetTitle())
	targets, err := s.GetTargetsByNameFunc(ctx, src.GetTitle())
	if err != nil {
		return nil, false, fmt.Errorf("error getting targets by name %s: %w", src.GetTitle(), err)
	}

	for _, target := range targets {
		if src.SameTypeWithTarget(target) {
			DPrintf("[%s] Found target by API name: %s", prefix, src.GetTitle())
			return target, true, nil
		}
		DPrintf("[%s] Ignoring target by API name: %s", prefix, target.String())
	}

	return nil, false, nil
}