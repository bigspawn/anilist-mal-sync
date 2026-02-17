package main

import (
	"strings"
	"testing"
	"time"
)

func TestManga_SameTypeWithTarget(t *testing.T) {
	tests := []struct {
		name   string
		source Manga
		target Target
		want   bool
	}{
		{
			name: "same MAL IDs match",
			source: Manga{
				IDMal:     12345,
				IDAnilist: 0, // MAL source doesn't know AniList ID
				TitleEN:   "Different Title",
			},
			target: Manga{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Another Different Title",
			},
			want: true,
		},
		{
			name: "same AniList IDs match",
			source: Manga{
				IDMal:     0,
				IDAnilist: 98765,
				TitleEN:   "Different Title",
			},
			target: Manga{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Another Different Title",
			},
			want: true,
		},
		{
			name: "different IDs but same title matches",
			source: Manga{
				IDMal:     0,
				IDAnilist: 0,
				TitleEN:   "Same Title",
			},
			target: Manga{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Same Title",
			},
			want: true,
		},
		{
			name: "different IDs and different titles no match",
			source: Manga{
				IDMal:     111,
				IDAnilist: 0,
				TitleEN:   "Title A",
			},
			target: Manga{
				IDMal:     222,
				IDAnilist: 333,
				TitleEN:   "Title B",
			},
			want: false,
		},
		{
			name: "zero MAL IDs should not match",
			source: Manga{
				IDMal:     0,
				IDAnilist: 0,
				TitleEN:   "Title A",
			},
			target: Manga{
				IDMal:     0,
				IDAnilist: 98765,
				TitleEN:   "Title B",
			},
			want: false,
		},
		{
			name: "zero AniList IDs should not match",
			source: Manga{
				IDMal:     12345,
				IDAnilist: 0,
				TitleEN:   "Title A",
			},
			target: Manga{
				IDMal:     67890,
				IDAnilist: 0,
				TitleEN:   "Title B",
			},
			want: false,
		},
		{
			name: "target is not Manga returns false",
			source: Manga{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Some Title",
			},
			target: Anime{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Some Title",
			},
			want: false,
		},
		{
			name: "reverse sync scenario - MAL source with zero AniList ID finds AniList target by MAL ID",
			source: Manga{
				IDMal:     30013, // Oshi no Ko MAL ID
				IDAnilist: 0,     // MAL source doesn't know AniList ID
				TitleEN:   "[Oshi No Ko]",
				Chapters:  167,
			},
			target: Manga{
				IDMal:     30013, // Same MAL ID
				IDAnilist: 117195,
				TitleEN:   "Oshi no Ko", // Slightly different title on AniList
				Chapters:  167,
			},
			want: true,
		},
		{
			name: "same chapters and volumes match as fallback",
			source: Manga{
				IDMal:     111,
				IDAnilist: 0,
				TitleEN:   "Title A",
				Chapters:  50,
				Volumes:   5,
			},
			target: Manga{
				IDMal:     222,
				IDAnilist: 333,
				TitleEN:   "Title B",
				Chapters:  50,
				Volumes:   5,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.SameTypeWithTarget(tt.target)
			if got != tt.want {
				t.Errorf("Manga.SameTypeWithTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManga_GetTargetID(t *testing.T) {
	tests := []struct {
		name         string
		manga        Manga
		reverse      bool
		wantTargetID TargetID
	}{
		{
			name: "normal sync returns MAL ID",
			manga: Manga{
				IDMal:     12345,
				IDAnilist: 67890,
			},
			reverse:      false,
			wantTargetID: 12345,
		},
		{
			name: "reverse sync returns AniList ID",
			manga: Manga{
				IDMal:     12345,
				IDAnilist: 67890,
			},
			reverse:      true,
			wantTargetID: 67890,
		},
		{
			name: "zero MAL ID in normal mode",
			manga: Manga{
				IDMal:     0,
				IDAnilist: 67890,
			},
			reverse:      false,
			wantTargetID: 0,
		},
		{
			name: "zero AniList ID in reverse mode",
			manga: Manga{
				IDMal:     12345,
				IDAnilist: 0,
			},
			reverse:      true,
			wantTargetID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origReverse := reverseDirection
			defer func() { reverseDirection = origReverse }()

			reverseDirection = &tt.reverse
			got := tt.manga.GetTargetID()
			if got != tt.wantTargetID {
				t.Errorf("GetTargetID() = %v, want %v", got, tt.wantTargetID)
			}
		})
	}
}

func TestManga_GetAniListID(t *testing.T) {
	tests := []struct {
		name  string
		manga Manga
		want  TargetID
	}{
		{
			name: "returns AniList ID",
			manga: Manga{
				IDMal:     12345,
				IDAnilist: 67890,
			},
			want: 67890,
		},
		{
			name: "zero AniList ID",
			manga: Manga{
				IDMal:     12345,
				IDAnilist: 0,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.manga.GetAniListID()
			if got != tt.want {
				t.Errorf("GetAniListID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManga_GetMALID(t *testing.T) {
	tests := []struct {
		name  string
		manga Manga
		want  TargetID
	}{
		{
			name: "returns MAL ID",
			manga: Manga{
				IDMal:     12345,
				IDAnilist: 67890,
			},
			want: 12345,
		},
		{
			name: "zero MAL ID",
			manga: Manga{
				IDMal:     0,
				IDAnilist: 67890,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.manga.GetMALID()
			if got != tt.want {
				t.Errorf("GetMALID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManga_GetStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status MangaStatus
		want   string
	}{
		{
			name:   "reading",
			status: MangaStatusReading,
			want:   "reading",
		},
		{
			name:   "completed",
			status: MangaStatusCompleted,
			want:   "completed",
		},
		{
			name:   "on_hold",
			status: MangaStatusOnHold,
			want:   "on_hold",
		},
		{
			name:   "dropped",
			status: MangaStatusDropped,
			want:   "dropped",
		},
		{
			name:   "plan_to_read",
			status: MangaStatusPlanToRead,
			want:   "plan_to_read",
		},
		{
			name:   "unknown",
			status: MangaStatusUnknown,
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := Manga{Status: tt.status}
			got := manga.GetStatusString()
			if got != tt.want {
				t.Errorf("GetStatusString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManga_SameProgressWithTarget(t *testing.T) {
	tests := []struct {
		name   string
		source Manga
		target Target
		want   bool
	}{
		{
			name: "identical progress",
			source: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			target: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			want: true,
		},
		{
			name: "different status",
			source: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			target: Manga{
				Status:          MangaStatusCompleted,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			want: false,
		},
		{
			name: "different score",
			source: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			target: Manga{
				Status:          MangaStatusReading,
				Score:           9,
				Progress:        10,
				ProgressVolumes: 2,
			},
			want: false,
		},
		{
			name: "different progress",
			source: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			target: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        12,
				ProgressVolumes: 2,
			},
			want: false,
		},
		{
			name: "different progress volumes",
			source: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			target: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 3,
			},
			want: false,
		},
		{
			name: "target is not Manga",
			source: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			target: Anime{
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.SameProgressWithTarget(tt.target)
			if got != tt.want {
				t.Errorf("SameProgressWithTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManga_SameTitleWithTarget(t *testing.T) {
	tests := []struct {
		name   string
		source Manga
		target Target
		want   bool
	}{
		{
			name: "exact English title match",
			source: Manga{
				TitleEN: "Test Manga",
			},
			target: Manga{
				TitleEN: "Test Manga",
			},
			want: true,
		},
		{
			name: "exact Japanese title match",
			source: Manga{
				TitleJP: "テストマンガ",
			},
			target: Manga{
				TitleJP: "テストマンガ",
			},
			want: true,
		},
		{
			name: "exact Romaji title match",
			source: Manga{
				TitleRomaji: "Test Manga",
			},
			target: Manga{
				TitleRomaji: "Test Manga",
			},
			want: true,
		},
		{
			name: "different titles no match",
			source: Manga{
				TitleEN: "Manga A",
			},
			target: Manga{
				TitleEN: "Manga B",
			},
			want: false,
		},
		{
			name: "target is not Manga",
			source: Manga{
				TitleEN: "Test Manga",
			},
			target: Anime{
				TitleEN: "Test Manga",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.SameTitleWithTarget(tt.target)
			if got != tt.want {
				t.Errorf("SameTitleWithTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManga_GetTitle(t *testing.T) {
	tests := []struct {
		name  string
		manga Manga
		want  string
	}{
		{
			name: "returns English title when available",
			manga: Manga{
				TitleEN:     "English Title",
				TitleJP:     "日本語タイトル",
				TitleRomaji: "Romaji Title",
			},
			want: "English Title",
		},
		{
			name: "returns Japanese title when English is empty",
			manga: Manga{
				TitleEN:     "",
				TitleJP:     "日本語タイトル",
				TitleRomaji: "Romaji Title",
			},
			want: "日本語タイトル",
		},
		{
			name: "returns Romaji title when EN and JP are empty",
			manga: Manga{
				TitleEN:     "",
				TitleJP:     "",
				TitleRomaji: "Romaji Title",
			},
			want: "Romaji Title",
		},
		{
			name: "returns empty string when all titles are empty",
			manga: Manga{
				TitleEN:     "",
				TitleJP:     "",
				TitleRomaji: "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.manga.GetTitle()
			if got != tt.want {
				t.Errorf("GetTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManga_String(t *testing.T) {
	tests := []struct {
		name  string
		manga Manga
	}{
		{
			name: "full manga string representation",
			manga: Manga{
				IDAnilist:       12345,
				IDMal:           67890,
				TitleEN:         "Test Manga",
				TitleJP:         "テストマンガ",
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
				Chapters:        50,
				Volumes:         5,
			},
		},
		{
			name: "minimal manga string representation",
			manga: Manga{
				IDAnilist: 0,
				IDMal:     0,
				TitleEN:   "",
				Status:    MangaStatusUnknown,
				Score:     0,
				Progress:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.manga.String()
			if got == "" {
				t.Error("String() returned empty string")
			}
			// Check that it contains expected format
			if !strings.Contains(got, "Manga{") || !strings.Contains(got, "}") {
				t.Errorf("String() has unexpected format: %s", got)
			}
		})
	}
}

func TestManga_SameProgressWithTarget_Dates(t *testing.T) {
	jun1 := timePtr(time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC))
	jun15 := timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC))
	jul1 := timePtr(time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC))
	jul15 := timePtr(time.Date(2023, 7, 15, 0, 0, 0, 0, time.UTC))

	base := Manga{
		Status:          MangaStatusCompleted,
		Score:           8,
		Progress:        50,
		ProgressVolumes: 5,
	}

	tests := []struct {
		name   string
		source Manga
		target Target
		want   bool
	}{
		{
			name: "same everything same dates",
			source: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
				StartedAt: jun1, FinishedAt: jun15,
			},
			want: true,
		},
		{
			name: "same everything different startedAt",
			source: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
				StartedAt: jul1, FinishedAt: jun15,
			},
			want: false,
		},
		{
			name: "same everything different finishedAt",
			source: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
				StartedAt: jun1, FinishedAt: jul15,
			},
			want: false,
		},
		{
			name: "source dates nil target dates set",
			source: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
			},
			target: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
				StartedAt: jun1, FinishedAt: jun15,
			},
			want: true,
		},
		{
			name: "source dates set target dates nil",
			source: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
			},
			want: false,
		},
		{
			name: "both dates nil",
			source: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
			},
			target: Manga{
				Status: base.Status, Score: base.Score, Progress: base.Progress, ProgressVolumes: base.ProgressVolumes,
			},
			want: true,
		},
		{
			name: "non-completed different finishedAt ignored",
			source: Manga{
				Status: MangaStatusPlanToRead, Score: 0, Progress: 0, ProgressVolumes: 0,
				FinishedAt: jun15,
			},
			target: Manga{
				Status: MangaStatusPlanToRead, Score: 0, Progress: 0, ProgressVolumes: 0,
			},
			want: true,
		},
		{
			name: "reading different finishedAt ignored",
			source: Manga{
				Status: MangaStatusReading, Score: 8, Progress: 10, ProgressVolumes: 2,
				FinishedAt: jun15,
			},
			target: Manga{
				Status: MangaStatusReading, Score: 8, Progress: 10, ProgressVolumes: 2,
				FinishedAt: jul15,
			},
			want: true,
		},
		{
			name: "completed different finishedAt triggers update",
			source: Manga{
				Status: MangaStatusCompleted, Score: 8, Progress: 50, ProgressVolumes: 5,
				FinishedAt: jun15,
			},
			target: Manga{
				Status: MangaStatusCompleted, Score: 8, Progress: 50, ProgressVolumes: 5,
				FinishedAt: jul15,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.SameProgressWithTarget(tt.target)
			if got != tt.want {
				t.Errorf("SameProgressWithTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManga_GetStringDiffWithTarget_Dates(t *testing.T) {
	jun1 := timePtr(time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC))
	jul1 := timePtr(time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name       string
		source     Manga
		target     Target
		wantField  string
		shouldShow bool
	}{
		{
			name: "dates differ shows StartedAt in diff",
			source: Manga{
				Status: MangaStatusCompleted, Score: 8, Progress: 50, ProgressVolumes: 5,
				StartedAt: jun1,
			},
			target: Manga{
				Status: MangaStatusCompleted, Score: 8, Progress: 50, ProgressVolumes: 5,
				StartedAt: jul1,
			},
			wantField:  "StartedAt",
			shouldShow: true,
		},
		{
			name: "dates same does not show StartedAt in diff",
			source: Manga{
				Status: MangaStatusCompleted, Score: 8, Progress: 50, ProgressVolumes: 5,
				StartedAt: jun1,
			},
			target: Manga{
				Status: MangaStatusCompleted, Score: 8, Progress: 50, ProgressVolumes: 5,
				StartedAt: jun1,
			},
			wantField:  "StartedAt",
			shouldShow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.GetStringDiffWithTarget(tt.target)
			contains := strings.Contains(got, tt.wantField)
			if contains != tt.shouldShow {
				t.Errorf("GetStringDiffWithTarget() = %q, wantField %q shouldShow=%v", got, tt.wantField, tt.shouldShow)
			}
		})
	}
}

func TestManga_GetUpdateOptions(t *testing.T) {
	date1 := time.Date(2024, 12, 18, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 12, 19, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		manga    Manga
		wantOpts int
	}{
		{
			name: "nil dates - 6 options",
			manga: Manga{
				Status:          MangaStatusCompleted,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
			},
			wantOpts: 6, // status, score, progress, progressVolumes, startDate, finishDate
		},
		{
			name: "with started at - 6 options",
			manga: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
				StartedAt:       &date1,
			},
			wantOpts: 6,
		},
		{
			name: "completed with both dates - 6 options",
			manga: Manga{
				Status:          MangaStatusCompleted,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
				StartedAt:       &date1,
				FinishedAt:      &date2,
			},
			wantOpts: 6,
		},
		{
			name: "reading with both dates - 6 options",
			manga: Manga{
				Status:          MangaStatusReading,
				Score:           8,
				Progress:        10,
				ProgressVolumes: 2,
				StartedAt:       &date1,
				FinishedAt:      &date2,
			},
			wantOpts: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.manga.GetUpdateOptions()
			if len(opts) != tt.wantOpts {
				t.Errorf("GetUpdateOptions() returned %d options, want %d", len(opts), tt.wantOpts)
			}
		})
	}
}
