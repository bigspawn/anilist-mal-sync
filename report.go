package main

// SyncReport accumulates warnings for deferred printing
type SyncReport struct {
	Warnings []Warning
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
	return &SyncReport{}
}

// AddWarning adds a warning to the report
func (r *SyncReport) AddWarning(title, reason, detail, mediaType string) {
	r.Warnings = append(r.Warnings, Warning{
		Title:     title,
		Reason:    reason,
		Detail:    detail,
		MediaType: mediaType,
	})
}

// HasWarnings returns true if there are warnings
func (r *SyncReport) HasWarnings() bool {
	return len(r.Warnings) > 0
}
