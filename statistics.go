package main

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"
)

// UpdateResult tracks the outcome of an update attempt
type UpdateResult struct {
	Title      string
	Detail     string // What changed (e.g., "ep 10→15")
	Status     string // watching, completed, etc.
	Error      error
	Skipped    bool
	SkipReason string
}

// Statistics tracks sync operation metrics
type Statistics struct {
	StartTime time.Time
	EndTime   time.Time

	TotalCount   int
	UpdatedCount int
	SkippedCount int
	ErrorCount   int

	// Detailed tracking
	UpdatedItems []UpdateResult
	SkippedItems []UpdateResult
	ErrorItems   []UpdateResult

	// Status breakdown
	StatusCounts map[string]int
}

// NewStatistics creates a new statistics tracker
func NewStatistics() *Statistics {
	return &Statistics{
		StartTime:    time.Now(),
		StatusCounts: make(map[string]int),
	}
}

// RecordUpdate records a successful update
func (s *Statistics) RecordUpdate(result UpdateResult) {
	s.UpdatedCount++
	s.UpdatedItems = append(s.UpdatedItems, result)
	s.StatusCounts[result.Status]++
}

// RecordSkip records a skipped item
func (s *Statistics) RecordSkip(result UpdateResult) {
	s.SkippedCount++
	s.SkippedItems = append(s.SkippedItems, result)
	s.StatusCounts[result.Status]++
}

// RecordError records an error
func (s *Statistics) RecordError(result UpdateResult) {
	s.ErrorCount++
	s.ErrorItems = append(s.ErrorItems, result)
}

// IncrementTotal increments total count
func (s *Statistics) IncrementTotal() {
	s.TotalCount++
}

// Reset resets statistics
func (s *Statistics) Reset() {
	s.StartTime = time.Now()
	s.EndTime = time.Time{}
	s.TotalCount = 0
	s.UpdatedCount = 0
	s.SkippedCount = 0
	s.ErrorCount = 0
	s.UpdatedItems = nil
	s.SkippedItems = nil
	s.ErrorItems = nil
	s.StatusCounts = make(map[string]int)
}

// Print outputs comprehensive statistics
func (s *Statistics) Print(ctx context.Context, prefix string) {
	s.EndTime = time.Now()
	duration := s.EndTime.Sub(s.StartTime)

	logger := LoggerFromContext(ctx)
	if logger == nil {
		// Fallback to old behavior for backward compatibility
		log.Printf("[%s] Updated %d out of %d\n", prefix, s.UpdatedCount, s.TotalCount)
		log.Printf("[%s] Skipped %d\n", prefix, s.SkippedCount)
		return
	}

	// Header
	logger.Info("")
	// Shorten prefix for cleaner output
	shortPrefix := strings.TrimPrefix(prefix, "AniList to MAL ")
	shortPrefix = strings.TrimPrefix(shortPrefix, "MAL to AniList ")
	logger.Stage("=== %s: Sync Complete ===", shortPrefix)

	// Summary
	logger.Info("Duration: %v", duration.Round(time.Millisecond))
	logger.Info("Total items: %d", s.TotalCount)
	logger.InfoSuccess("Updated: %d", s.UpdatedCount)

	if s.SkippedCount > 0 {
		logger.Info("Skipped: %d (no changes needed)", s.SkippedCount)
	}

	if s.ErrorCount > 0 {
		logger.Error("Errors: %d", s.ErrorCount)
	}

	// Status breakdown
	if len(s.StatusCounts) > 0 {
		logger.Info("")
		logger.Info("Status breakdown:")

		statuses := make([]string, 0, len(s.StatusCounts))
		for status := range s.StatusCounts {
			statuses = append(statuses, status)
		}
		sort.Strings(statuses)

		for _, status := range statuses {
			count := s.StatusCounts[status]
			logger.Info("  %s: %d", status, count)
		}
	}

	// Error details
	if len(s.ErrorItems) > 0 {
		logger.Info("")
		for i, item := range s.ErrorItems {
			logger.Error("Failed updates:")
			logger.Error("  %d. %s: %v", i+1, item.Title, item.Error)
		}
	}

	// Skipped details (verbose only)
	if len(s.SkippedItems) > 0 && logger.level >= LogLevelDebug {
		logger.Info("")
		logger.Debug("Skipped items:")
		for i, item := range s.SkippedItems {
			if item.SkipReason != "" {
				logger.Debug("  %d. %s: %s", i+1, item.Title, item.SkipReason)
			} else {
				logger.Debug("  %d. %s", i+1, item.Title)
			}
		}
	}

	logger.Info("")
}
