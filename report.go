package main

import "sync"

// SyncReport accumulates warnings for deferred printing
type SyncReport struct {
	Warnings []Warning
	mu       sync.Mutex
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
		Warnings: []Warning{},
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
