package main

import "sync"

// SyncReport accumulates warnings and unmapped entries for deferred printing
type SyncReport struct {
	Warnings           []Warning
	UnmappedItems      []UnmappedEntry
	DuplicateConflicts []DuplicateConflict
	mu                 sync.Mutex
}

// DuplicateConflict records when multiple sources map to the same target.
type DuplicateConflict struct {
	LoserTitle  string
	WinnerTitle string
	TargetTitle string
	LoserStrat  string
	WinnerStrat string
	MediaType   string
}

// Warning represents a warning about a potentially problematic match
type Warning struct {
	Title     string
	Reason    string
	Detail    string
	MediaType string // "Anime" or "Manga"
}

// NewSyncReport creates a new sync report
func NewSyncReport() *SyncReport {
	return &SyncReport{
		Warnings:           []Warning{},
		DuplicateConflicts: []DuplicateConflict{},
	}
}

// AddWarning adds a warning to the report (thread-safe)
func (r *SyncReport) AddWarning(title, reason, detail, mediaType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Warnings = append(r.Warnings, Warning{
		Title:     title,
		Reason:    reason,
		Detail:    detail,
		MediaType: mediaType,
	})
}

// HasWarnings returns true if there are warnings (thread-safe)
func (r *SyncReport) HasWarnings() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Warnings) > 0
}

// AddUnmappedItems adds unmapped entries to the report (thread-safe)
func (r *SyncReport) AddUnmappedItems(items []UnmappedEntry) {
	if len(items) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.UnmappedItems = append(r.UnmappedItems, items...)
}

// HasUnmappedItems returns true if there are unmapped entries (thread-safe)
func (r *SyncReport) HasUnmappedItems() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.UnmappedItems) > 0
}

// AddDuplicateConflict adds a duplicate conflict to the report (thread-safe)
func (r *SyncReport) AddDuplicateConflict(
	loserTitle, winnerTitle, targetTitle, loserStrat, winnerStrat, mediaType string,
) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DuplicateConflicts = append(r.DuplicateConflicts, DuplicateConflict{
		LoserTitle:  loserTitle,
		WinnerTitle: winnerTitle,
		TargetTitle: targetTitle,
		LoserStrat:  loserStrat,
		WinnerStrat: winnerStrat,
		MediaType:   mediaType,
	})
}

// HasDuplicateConflicts returns true if there are duplicate conflicts (thread-safe)
func (r *SyncReport) HasDuplicateConflicts() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.DuplicateConflicts) > 0
}
