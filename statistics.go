package main

import (
	"context"
	"fmt"
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
	IsDryRun   bool // True if this is a dry run "update"
}

// Statistics tracks sync operation metrics
type Statistics struct {
	StartTime time.Time
	EndTime   time.Time

	TotalCount   int
	UpdatedCount int
	SkippedCount int
	ErrorCount   int
	DryRunCount  int

	// Detailed tracking
	UpdatedItems []UpdateResult
	SkippedItems []UpdateResult
	ErrorItems   []UpdateResult
	DryRunItems  []UpdateResult

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

// RecordDryRun records a dry run item
func (s *Statistics) RecordDryRun(result UpdateResult) {
	result.IsDryRun = true
	s.DryRunCount++
	s.DryRunItems = append(s.DryRunItems, result)
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
	s.DryRunCount = 0
	s.UpdatedItems = nil
	s.SkippedItems = nil
	s.ErrorItems = nil
	s.DryRunItems = nil
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

	s.printStatusBreakdown(logger)
	s.printSkipReasons(logger)
	s.printErrors(logger)

	logger.Info("")
}

func (s *Statistics) printStatusBreakdown(logger *Logger) {
	if len(s.StatusCounts) == 0 {
		return
	}

	logger.Info("")
	logger.Info("Status breakdown:")

	statuses := make([]string, 0, len(s.StatusCounts))
	for status := range s.StatusCounts {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)

	for _, status := range statuses {
		logger.Info("  %s: %d", status, s.StatusCounts[status])
	}
}

func (s *Statistics) printSkipReasons(logger *Logger) {
	if len(s.SkippedItems) == 0 {
		return
	}

	byReason := groupSkipReasons(s.SkippedItems)

	logger.Info("")
	logger.Info("Skipped by reason:")
	for _, reason := range sortedKeys(byReason) {
		logger.Info("  %s: %d", reason, byReason[reason])
	}

	// Detailed list in verbose mode only
	if logger.level >= LogLevelDebug {
		s.printSkippedItemsDetail(logger)
	}
}

func (s *Statistics) printSkippedItemsDetail(logger *Logger) {
	logger.Debug("")
	logger.Debug("Skipped items detail:")
	for i, item := range s.SkippedItems {
		reason := item.SkipReason
		if reason == "" {
			reason = "unspecified"
		}
		logger.Debug("  %d. %s: %s", i+1, item.Title, reason)
	}
}

func (s *Statistics) printErrors(logger *Logger) {
	if len(s.ErrorItems) == 0 {
		return
	}

	logger.Info("")
	logger.Error("Failed updates:")
	for i, item := range s.ErrorItems {
		logger.Error("  %d. %s: %v", i+1, item.Title, item.Error)
	}
}

func groupSkipReasons(items []UpdateResult) map[string]int {
	byReason := make(map[string]int)
	for _, item := range items {
		reason := item.SkipReason
		if reason == "" {
			reason = "unspecified"
		}
		byReason[reason]++
	}
	return byReason
}

// PrintGlobalSummary prints combined statistics for multiple sync operations
func PrintGlobalSummary(ctx context.Context, stats []*Statistics, report *SyncReport, totalDuration time.Duration, reverse bool) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}

	totals := aggregateStats(stats)

	// Header
	logger.Info("")
	logger.Stage("=== Sync Complete ===")
	logger.Info("Duration: %v", totalDuration.Round(time.Millisecond))
	logger.Info("")
	logger.Info("Total: %d | Updated: %d | Skipped: %d | Errors: %d",
		totals.items, totals.updated, totals.skipped, totals.errors)

	printGlobalSkipReasons(logger, totals.skipReasons)
	printGlobalUpdates(logger, totals.updatedItems)
	printGlobalDryRunUpdates(logger, totals.dryRunItems)
	printGlobalUnmapped(logger, report)
	printGlobalWarnings(logger, report)
	printGlobalDuplicateConflicts(logger, report)
	printGlobalFavorites(logger, report, reverse)
	printGlobalErrors(logger, totals.errorItems)
}

type aggregatedTotals struct {
	items, updated, skipped, errors, dryRun int
	skipReasons                             map[string]int
	errorItems                              []UpdateResult
	updatedItems                            []UpdateResult
	dryRunItems                             []UpdateResult
}

func aggregateStats(stats []*Statistics) aggregatedTotals {
	totals := aggregatedTotals{
		skipReasons: make(map[string]int),
	}

	for _, s := range stats {
		if s == nil {
			continue
		}
		totals.items += s.TotalCount
		totals.updated += s.UpdatedCount
		totals.skipped += s.SkippedCount
		totals.errors += s.ErrorCount
		totals.dryRun += s.DryRunCount

		for reason, count := range groupSkipReasons(s.SkippedItems) {
			totals.skipReasons[reason] += count
		}
		totals.errorItems = append(totals.errorItems, s.ErrorItems...)
		totals.updatedItems = append(totals.updatedItems, s.UpdatedItems...)
		totals.dryRunItems = append(totals.dryRunItems, s.DryRunItems...)
	}

	return totals
}

func printGlobalSkipReasons(logger *Logger, skipReasons map[string]int) {
	if len(skipReasons) == 0 {
		return
	}

	logger.Info("")
	logger.Info("Skipped by reason:")
	for _, reason := range sortedKeys(skipReasons) {
		logger.Info("  %s: %d", reason, skipReasons[reason])
	}
}

func printGlobalUnmapped(logger *Logger, report *SyncReport) {
	if report == nil || !report.HasUnmappedItems() {
		return
	}

	logger.Info("")
	logger.Warn("Unmapped entries (%d):", len(report.UnmappedItems))
	for i, item := range report.UnmappedItems {
		mediaLabel := capitalizeFirst(item.MediaType)
		line := formatUnmappedLine(i+1, item, mediaLabel)
		logger.Warn("%s", line)
	}
	logger.Info("")
	logger.Info("Hint: run 'anilist-mal-sync unmapped --fix' to manage these entries")
}

func printGlobalWarnings(logger *Logger, report *SyncReport) {
	if report == nil || !report.HasWarnings() {
		return
	}

	logger.Info("")
	logger.Warn("Warnings (%d):", len(report.Warnings))
	for _, w := range report.Warnings {
		if w.Detail != "" {
			logger.Warn("  %q - %s %s", w.Title, w.Reason, w.Detail)
		} else {
			logger.Warn("  %q - %s", w.Title, w.Reason)
		}
	}
}

func printGlobalDuplicateConflicts(logger *Logger, report *SyncReport) {
	if report == nil || !report.HasDuplicateConflicts() {
		return
	}

	logger.Info("")
	logger.Warn("Duplicate target conflicts (%d):", len(report.DuplicateConflicts))
	for _, c := range report.DuplicateConflicts {
		mediaLabel := capitalizeFirst(c.MediaType)
		logger.Warn("  %q -> target %q [%s]", c.LoserTitle, c.TargetTitle, mediaLabel)
		logger.Warn("    Kept: %q via %s", c.WinnerTitle, c.WinnerStrat)
	}
}

func printGlobalUpdates(logger *Logger, updatedItems []UpdateResult) {
	if len(updatedItems) == 0 {
		return
	}

	logger.Info("")
	logger.InfoSuccess("Updates (%d):", len(updatedItems))
	for _, item := range updatedItems {
		if item.Detail != "" {
			logger.InfoSuccess("  %s %s", item.Title, item.Detail)
		} else {
			logger.InfoSuccess("  %s", item.Title)
		}
	}
}

func printGlobalDryRunUpdates(logger *Logger, dryRunItems []UpdateResult) {
	if len(dryRunItems) == 0 {
		return
	}

	logger.Info("")
	logger.InfoDryRun("Would update (%d):", len(dryRunItems))
	for _, item := range dryRunItems {
		if item.Detail != "" {
			logger.Info("  %s %s", item.Title, item.Detail)
		} else {
			logger.Info("  %s", item.Title)
		}
	}
}

func printGlobalErrors(logger *Logger, errorItems []UpdateResult) {
	if len(errorItems) == 0 {
		return
	}

	logger.Info("")
	logger.Error("Errors:")
	for i, item := range errorItems {
		logger.Error("  %d. %s: %v", i+1, item.Title, item.Error)
	}
}

// sortedKeys returns sorted keys from a map
func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatUnmappedLine(num int, item UnmappedEntry, mediaLabel string) string {
	var base string
	switch {
	case item.AniListID > 0:
		base = fmt.Sprintf("  %d. %q (AniList: %d) [%s]", num, item.Title, item.AniListID, mediaLabel)
	case item.MALID > 0:
		base = fmt.Sprintf("  %d. %q (MAL: %d) [%s]", num, item.Title, item.MALID, mediaLabel)
	default:
		base = fmt.Sprintf("  %d. %q [%s]", num, item.Title, mediaLabel)
	}
	if item.Reason != "" {
		return base + " - " + item.Reason
	}
	return base
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func printGlobalFavorites(logger *Logger, report *SyncReport, _ bool) {
	if report == nil {
		return
	}

	// Print favorites summary if any were added or mismatches exist.
	// Mismatches are only populated by ReportMismatches() which runs in
	// forward (AniList→MAL) direction only, so direction label is always AniList→MAL.
	if report.FavoritesAdded == 0 && !report.HasFavoritesMismatches() {
		return
	}

	logger.Info("")
	if report.FavoritesAdded > 0 {
		logger.InfoSuccess("Favorites: +%d added on AniList", report.FavoritesAdded)
	}

	if report.HasFavoritesMismatches() {
		logger.Info("Favorites: %d mismatches (AniList→MAL, report only)", len(report.FavoritesMismatches))
	}
}
