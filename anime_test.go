package main

import (
	"strings"
	"testing"
	"time"
)

func TestTimeToFuzzyDateInput(t *testing.T) {
	tests := []struct {
		name     string
		input    *time.Time
		expected map[string]int
	}{
		{
			name:     "nil time returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "valid date",
			input:    timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
			expected: map[string]int{"year": 2023, "month": 6, "day": 15},
		},
		{
			name:     "new year edge",
			input:    timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			expected: map[string]int{"year": 2024, "month": 1, "day": 1},
		},
		{
			name:     "end of year",
			input:    timePtr(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)),
			expected: map[string]int{"year": 2023, "month": 12, "day": 31},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeToFuzzyDateInput(tt.input)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("timeToFuzzyDateInput() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("timeToFuzzyDateInput() = nil, want %v", tt.expected)
			}
			if got["year"] != tt.expected["year"] || got["month"] != tt.expected["month"] || got["day"] != tt.expected["day"] {
				t.Errorf("timeToFuzzyDateInput() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSameDates(t *testing.T) {
	tests := []struct {
		name   string
		a      *time.Time
		b      *time.Time
		expect bool
	}{
		{
			name:   "both nil",
			a:      nil,
			b:      nil,
			expect: true,
		},
		{
			name:   "source nil target set",
			a:      nil,
			b:      timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
			expect: true,
		},
		{
			name:   "source set target nil",
			a:      timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
			b:      nil,
			expect: false,
		},
		{
			name:   "same date",
			a:      timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
			b:      timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
			expect: true,
		},
		{
			name:   "different year",
			a:      timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
			b:      timePtr(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)),
			expect: false,
		},
		{
			name:   "different month",
			a:      timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
			b:      timePtr(time.Date(2023, 7, 15, 0, 0, 0, 0, time.UTC)),
			expect: false,
		},
		{
			name:   "different day",
			a:      timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
			b:      timePtr(time.Date(2023, 6, 16, 0, 0, 0, 0, time.UTC)),
			expect: false,
		},
		{
			name:   "same date different time of day",
			a:      timePtr(time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)),
			b:      timePtr(time.Date(2023, 6, 15, 20, 0, 0, 0, time.UTC)),
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sameDates(tt.a, tt.b)
			if got != tt.expect {
				t.Errorf("sameDates() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestAnime_SameTypeWithTarget(t *testing.T) {
	tests := []struct {
		name   string
		source Anime
		target Target
		want   bool
	}{
		{
			name: "same MAL IDs match",
			source: Anime{
				IDMal:     12345,
				IDAnilist: 0, // MAL source doesn't know AniList ID
				TitleEN:   "Different Title",
			},
			target: Anime{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Another Different Title",
			},
			want: true,
		},
		{
			name: "same AniList IDs match",
			source: Anime{
				IDMal:     0,
				IDAnilist: 98765,
				TitleEN:   "Different Title",
			},
			target: Anime{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Another Different Title",
			},
			want: true,
		},
		{
			name: "different IDs but same title matches",
			source: Anime{
				IDMal:     0,
				IDAnilist: 0,
				TitleEN:   "Same Title",
			},
			target: Anime{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Same Title",
			},
			want: true,
		},
		{
			name: "different IDs and different titles no match",
			source: Anime{
				IDMal:     111,
				IDAnilist: 0,
				TitleEN:   "Title A",
			},
			target: Anime{
				IDMal:     222,
				IDAnilist: 333,
				TitleEN:   "Title B",
			},
			want: false,
		},
		{
			name: "zero MAL IDs should not match",
			source: Anime{
				IDMal:     0,
				IDAnilist: 0,
				TitleEN:   "Title A",
			},
			target: Anime{
				IDMal:     0,
				IDAnilist: 98765,
				TitleEN:   "Title B",
			},
			want: false,
		},
		{
			name: "zero AniList IDs should not match",
			source: Anime{
				IDMal:     12345,
				IDAnilist: 0,
				TitleEN:   "Title A",
			},
			target: Anime{
				IDMal:     67890,
				IDAnilist: 0,
				TitleEN:   "Title B",
			},
			want: false,
		},
		{
			name: "target is not Anime returns false",
			source: Anime{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Some Title",
			},
			target: Manga{
				IDMal:     12345,
				IDAnilist: 98765,
				TitleEN:   "Some Title",
			},
			want: false,
		},
		{
			name: "reverse sync scenario - MAL source with zero AniList ID finds AniList target by MAL ID",
			source: Anime{
				IDMal:       38680, // Fruits Basket 1st Season MAL ID
				IDAnilist:   0,     // MAL source doesn't know AniList ID
				TitleEN:     "Fruits Basket 1st Season",
				NumEpisodes: 25,
			},
			target: Anime{
				IDMal:       38680, // Same MAL ID
				IDAnilist:   105334,
				TitleEN:     "Fruits Basket (2019)", // Different title on AniList
				NumEpisodes: 25,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.SameTypeWithTarget(tt.target)
			if got != tt.want {
				t.Errorf("Anime.SameTypeWithTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnime_IsPotentiallyIncorrectMatch(t *testing.T) {
	tests := []struct {
		name     string
		source   Anime
		target   Anime
		expected bool // true = should reject
	}{
		{
			name: "Special (0 eps) vs TV series (13 eps) - different titles",
			source: Anime{
				TitleJP:     "ガールズバンドクライ なぁ、未来。",
				NumEpisodes: 0,
				IDMal:       0,
			},
			target: Anime{
				TitleJP:     "ガールズバンドクライ",
				NumEpisodes: 13,
				IDMal:       55102,
			},
			expected: true, // Should reject
		},
		{
			name: "Special (1 ep) vs TV series (13 eps) - different titles",
			source: Anime{
				TitleJP:     "ガールズバンドクライ なぁ、未来。",
				NumEpisodes: 1,
				IDMal:       0,
			},
			target: Anime{
				TitleJP:     "ガールズバンドクライ",
				NumEpisodes: 13,
				IDMal:       55102,
			},
			expected: true, // Should reject
		},
		{
			name: "Same MAL ID - should not reject",
			source: Anime{
				TitleJP:     "Girls Band Cry",
				NumEpisodes: 0,
				IDMal:       55102,
			},
			target: Anime{
				TitleJP:     "Girls Band Cry",
				NumEpisodes: 13,
				IDMal:       55102,
			},
			expected: false, // Should NOT reject (valid MAL ID match)
		},
		{
			name: "Identical titles - should not reject",
			source: Anime{
				TitleJP:     "Girls Band Cry",
				NumEpisodes: 0,
				IDMal:       0,
			},
			target: Anime{
				TitleJP:     "Girls Band Cry",
				NumEpisodes: 13,
				IDMal:       55102,
			},
			expected: false, // Should NOT reject (exact title match)
		},
		{
			name: "Both have few episodes - should not reject",
			source: Anime{
				TitleJP:     "Special Episode",
				NumEpisodes: 1,
				IDMal:       0,
			},
			target: Anime{
				TitleJP:     "Special Episode",
				NumEpisodes: 2,
				IDMal:       0,
			},
			expected: false, // Should NOT reject (both are specials)
		},
		{
			name: "Source has 2 episodes - should not reject",
			source: Anime{
				TitleJP:     "Short OVA",
				NumEpisodes: 2,
				IDMal:       0,
			},
			target: Anime{
				TitleJP:     "Short OVA",
				NumEpisodes: 13,
				IDMal:       12345,
			},
			expected: false, // Should NOT reject (source has > 1 episode)
		},
		{
			name: "Target has 4 episodes or fewer - should not reject",
			source: Anime{
				TitleJP:     "Short Series",
				NumEpisodes: 0,
				IDMal:       0,
			},
			target: Anime{
				TitleJP:     "Short Series",
				NumEpisodes: 4,
				IDMal:       12345,
			},
			expected: false, // Should NOT reject (target has <= 4 episodes)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.source.IsPotentiallyIncorrectMatch(tt.target)
			if result != tt.expected {
				t.Errorf("IsPotentiallyIncorrectMatch() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAnime_GetUpdateOptions(t *testing.T) {
	date1 := time.Date(2024, 12, 18, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 12, 19, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		anime    Anime
		wantOpts int
	}{
		{
			name: "nil dates - only 3 options",
			anime: Anime{
				Status:   StatusCompleted,
				Score:    8,
				Progress: 13,
			},
			wantOpts: 3, // status, score, progress only
		},
		{
			name: "with started at - 4 options",
			anime: Anime{
				Status:    StatusCompleted,
				Score:     8,
				Progress:  13,
				StartedAt: &date1,
			},
			wantOpts: 4,
		},
		{
			name: "completed with both dates - 5 options",
			anime: Anime{
				Status:     StatusCompleted,
				Score:      8,
				Progress:   13,
				StartedAt:  &date1,
				FinishedAt: &date2,
			},
			wantOpts: 5,
		},
		{
			name: "watching with dates - only started included",
			anime: Anime{
				Status:     StatusWatching,
				Score:      8,
				Progress:   13,
				StartedAt:  &date1,
				FinishedAt: &date2,
			},
			wantOpts: 4, // finish date ignored for non-completed
		},
		{
			name: "plan to watch with dates - only started included",
			anime: Anime{
				Status:     StatusPlanToWatch,
				Score:      0,
				Progress:   0,
				StartedAt:  &date1,
				FinishedAt: &date2,
			},
			wantOpts: 4, // finish date ignored for non-completed
		},
		{
			name: "on hold with both dates - only started included",
			anime: Anime{
				Status:     StatusOnHold,
				Score:      7,
				Progress:   5,
				StartedAt:  &date1,
				FinishedAt: &date2,
			},
			wantOpts: 4, // finish date ignored for non-completed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.anime.GetUpdateOptions()
			if len(opts) != tt.wantOpts {
				t.Errorf("GetUpdateOptions() returned %d options, want %d", len(opts), tt.wantOpts)
			}
		})
	}
}

func TestAnime_GetTargetID(t *testing.T) {
	tests := []struct {
		name         string
		anime        Anime
		reverse      bool
		wantTargetID TargetID
	}{
		{
			name: "normal sync returns MAL ID",
			anime: Anime{
				IDMal:     12345,
				IDAnilist: 67890,
			},
			reverse:      false,
			wantTargetID: 12345,
		},
		{
			name: "reverse sync returns AniList ID",
			anime: Anime{
				IDMal:     12345,
				IDAnilist: 67890,
			},
			reverse:      true,
			wantTargetID: 67890,
		},
		{
			name: "zero MAL ID in normal mode",
			anime: Anime{
				IDMal:     0,
				IDAnilist: 67890,
			},
			reverse:      false,
			wantTargetID: 0,
		},
		{
			name: "zero AniList ID in reverse mode",
			anime: Anime{
				IDMal:     12345,
				IDAnilist: 0,
			},
			reverse:      true,
			wantTargetID: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := tt.anime
			a.isReverse = tt.reverse
			got := a.GetTargetID()
			if got != tt.wantTargetID {
				t.Errorf("GetTargetID() = %v, want %v", got, tt.wantTargetID)
			}
		})
	}
}

func TestAnime_GetAniListID(t *testing.T) {
	tests := []struct {
		name  string
		anime Anime
		want  TargetID
	}{
		{
			name: "returns AniList ID",
			anime: Anime{
				IDMal:     12345,
				IDAnilist: 67890,
			},
			want: 67890,
		},
		{
			name: "zero AniList ID",
			anime: Anime{
				IDMal:     12345,
				IDAnilist: 0,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.anime.GetAniListID()
			if got != tt.want {
				t.Errorf("GetAniListID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnime_GetMALID(t *testing.T) {
	tests := []struct {
		name  string
		anime Anime
		want  TargetID
	}{
		{
			name: "returns MAL ID",
			anime: Anime{
				IDMal:     12345,
				IDAnilist: 67890,
			},
			want: 12345,
		},
		{
			name: "zero MAL ID",
			anime: Anime{
				IDMal:     0,
				IDAnilist: 67890,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.anime.GetMALID()
			if got != tt.want {
				t.Errorf("GetMALID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnime_GetStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{
			name:   "watching",
			status: StatusWatching,
			want:   "watching",
		},
		{
			name:   "completed",
			status: StatusCompleted,
			want:   "completed",
		},
		{
			name:   "on_hold",
			status: StatusOnHold,
			want:   "on_hold",
		},
		{
			name:   "dropped",
			status: StatusDropped,
			want:   "dropped",
		},
		{
			name:   "plan_to_watch",
			status: StatusPlanToWatch,
			want:   "plan_to_watch",
		},
		{
			name:   "unknown",
			status: StatusUnknown,
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anime := Anime{Status: tt.status}
			got := anime.GetStatusString()
			if got != tt.want {
				t.Errorf("GetStatusString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnime_SameProgressWithTarget(t *testing.T) {
	tests := []struct {
		name   string
		source Anime
		target Target
		want   bool
	}{
		{
			name: "identical progress",
			source: Anime{
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			target: Anime{
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			want: true,
		},
		{
			name: "different status",
			source: Anime{
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			target: Anime{
				Status:   StatusCompleted,
				Score:    8,
				Progress: 10,
			},
			want: false,
		},
		{
			name: "different score",
			source: Anime{
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			target: Anime{
				Status:   StatusWatching,
				Score:    9,
				Progress: 10,
			},
			want: false,
		},
		{
			name: "different progress",
			source: Anime{
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			target: Anime{
				Status:   StatusWatching,
				Score:    8,
				Progress: 12,
			},
			want: false,
		},
		{
			name: "same progress with different NumEpisodes",
			source: Anime{
				Status:      StatusWatching,
				Score:       8,
				Progress:    10,
				NumEpisodes: 12,
			},
			target: Anime{
				Status:      StatusWatching,
				Score:       8,
				Progress:    10,
				NumEpisodes: 13,
			},
			want: true,
		},
		{
			name: "zero NumEpisodes both",
			source: Anime{
				Status:      StatusWatching,
				Score:       8,
				Progress:    10,
				NumEpisodes: 0,
			},
			target: Anime{
				Status:      StatusWatching,
				Score:       8,
				Progress:    10,
				NumEpisodes: 0,
			},
			want: true,
		},
		{
			name: "relative progress same (10/12 vs 8/10)",
			source: Anime{
				Status:      StatusWatching,
				Score:       8,
				Progress:    10,
				NumEpisodes: 12,
			},
			target: Anime{
				Status:      StatusWatching,
				Score:       8,
				Progress:    8,
				NumEpisodes: 10,
			},
			want: true,
		},
		{
			name: "target is not Anime",
			source: Anime{
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			target: Manga{
				Status:   "watching",
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

func TestAnime_SameTitleWithTarget(t *testing.T) {
	tests := []struct {
		name   string
		source Anime
		target Target
		want   bool
	}{
		{
			name: "exact English title match",
			source: Anime{
				TitleEN: "Test Anime",
			},
			target: Anime{
				TitleEN: "Test Anime",
			},
			want: true,
		},
		{
			name: "exact Japanese title match",
			source: Anime{
				TitleJP: "テストアニメ",
			},
			target: Anime{
				TitleJP: "テストアニメ",
			},
			want: true,
		},
		{
			name: "exact Romaji title match",
			source: Anime{
				TitleRomaji: "Test Anime",
			},
			target: Anime{
				TitleRomaji: "Test Anime",
			},
			want: true,
		},
		{
			name: "different titles no match",
			source: Anime{
				TitleEN: "Anime A",
			},
			target: Anime{
				TitleEN: "Anime B",
			},
			want: false,
		},
		{
			name: "20% episode difference accepted",
			source: Anime{
				TitleEN:     "Test Anime",
				NumEpisodes: 10,
			},
			target: Anime{
				TitleEN:     "Test Anime",
				NumEpisodes: 12,
			},
			want: true,
		},
		{
			name: "more than 20% episode difference rejected",
			source: Anime{
				TitleEN:     "Test Anime",
				NumEpisodes: 10,
			},
			target: Anime{
				TitleEN:     "Test Anime",
				NumEpisodes: 13,
			},
			want: false,
		},
		{
			name: "zero NumEpisodes both accepted",
			source: Anime{
				TitleEN:     "Test Anime",
				NumEpisodes: 0,
			},
			target: Anime{
				TitleEN:     "Test Anime",
				NumEpisodes: 0,
			},
			want: true,
		},
		{
			name: "target is not Anime",
			source: Anime{
				TitleEN: "Test Anime",
			},
			target: Manga{
				TitleEN: "Test Anime",
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

func TestAnime_IdenticalTitleMatch(t *testing.T) {
	tests := []struct {
		name   string
		source Anime
		target Anime
		want   bool
	}{
		{
			name: "exact English title match",
			source: Anime{
				TitleEN: "Test Anime",
			},
			target: Anime{
				TitleEN: "Test Anime",
			},
			want: true,
		},
		{
			name: "exact Japanese title match",
			source: Anime{
				TitleJP: "テストアニメ",
			},
			target: Anime{
				TitleJP: "テストアニメ",
			},
			want: true,
		},
		{
			name: "exact Romaji title match",
			source: Anime{
				TitleRomaji: "Test Anime",
			},
			target: Anime{
				TitleRomaji: "Test Anime",
			},
			want: true,
		},
		{
			name: "different English titles",
			source: Anime{
				TitleEN: "Anime A",
			},
			target: Anime{
				TitleEN: "Anime B",
			},
			want: false,
		},
		{
			name: "empty titles no match",
			source: Anime{
				TitleEN:     "",
				TitleJP:     "",
				TitleRomaji: "",
			},
			target: Anime{
				TitleEN:     "",
				TitleJP:     "",
				TitleRomaji: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.IdenticalTitleMatch(tt.target)
			if got != tt.want {
				t.Errorf("IdenticalTitleMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnime_GetTitle(t *testing.T) {
	tests := []struct {
		name  string
		anime Anime
		want  string
	}{
		{
			name: "returns English title when available",
			anime: Anime{
				TitleEN:     "English Title",
				TitleJP:     "日本語タイトル",
				TitleRomaji: "Romaji Title",
			},
			want: "English Title",
		},
		{
			name: "returns Japanese title when English is empty",
			anime: Anime{
				TitleEN:     "",
				TitleJP:     "日本語タイトル",
				TitleRomaji: "Romaji Title",
			},
			want: "日本語タイトル",
		},
		{
			name: "returns Romaji title when EN and JP are empty",
			anime: Anime{
				TitleEN:     "",
				TitleJP:     "",
				TitleRomaji: "Romaji Title",
			},
			want: "Romaji Title",
		},
		{
			name: "returns empty string when all titles are empty",
			anime: Anime{
				TitleEN:     "",
				TitleJP:     "",
				TitleRomaji: "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.anime.GetTitle()
			if got != tt.want {
				t.Errorf("GetTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnime_String(t *testing.T) {
	tests := []struct {
		name  string
		anime Anime
	}{
		{
			name: "full anime string representation",
			anime: Anime{
				IDAnilist:   12345,
				IDMal:       67890,
				TitleEN:     "Test Anime",
				TitleJP:     "テストアニメ",
				Status:      StatusWatching,
				Score:       8,
				Progress:    10,
				NumEpisodes: 12,
				SeasonYear:  2024,
			},
		},
		{
			name: "minimal anime string representation",
			anime: Anime{
				IDAnilist: 0,
				IDMal:     0,
				TitleEN:   "",
				Status:    StatusUnknown,
				Score:     0,
				Progress:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.anime.String()
			if got == "" {
				t.Error("String() returned empty string")
			}
			// Check that it contains expected format
			if !strings.Contains(got, "Anime{") || !strings.Contains(got, "}") {
				t.Errorf("String() has unexpected format: %s", got)
			}
		})
	}
}

func TestAnime_SameProgressWithTarget_Dates(t *testing.T) {
	jun1 := timePtr(time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC))
	jun15 := timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC))
	jul1 := timePtr(time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC))
	jul15 := timePtr(time.Date(2023, 7, 15, 0, 0, 0, 0, time.UTC))

	base := Anime{
		Status:   StatusCompleted,
		Score:    8,
		Progress: 12,
	}

	tests := []struct {
		name   string
		source Anime
		target Target
		want   bool
	}{
		{
			name: "same everything same dates",
			source: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jun15,
			},
			want: true,
		},
		{
			name: "same everything different startedAt",
			source: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
				StartedAt: jul1, FinishedAt: jun15,
			},
			want: false,
		},
		{
			name: "same everything different finishedAt",
			source: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jul15,
			},
			want: false,
		},
		{
			name: "source dates nil target dates set",
			source: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
			},
			target: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jun15,
			},
			want: true,
		},
		{
			name: "source dates set target dates nil",
			source: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
			},
			want: false,
		},
		{
			name: "both dates nil",
			source: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
			},
			target: Anime{
				Status: base.Status, Score: base.Score, Progress: base.Progress,
			},
			want: true,
		},
		{
			name: "different status dates irrelevant",
			source: Anime{
				Status: StatusWatching, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Anime{
				Status: StatusCompleted, Score: base.Score, Progress: base.Progress,
				StartedAt: jun1, FinishedAt: jun15,
			},
			want: false,
		},
		{
			name: "non-completed different finishedAt ignored",
			source: Anime{
				Status: StatusPlanToWatch, Score: 0, Progress: 0,
				FinishedAt: jun15,
			},
			target: Anime{
				Status: StatusPlanToWatch, Score: 0, Progress: 0,
			},
			want: true,
		},
		{
			name: "plan_to_watch with finishedAt set on source only",
			source: Anime{
				Status: StatusPlanToWatch, Score: 0, Progress: 0,
				StartedAt: jun1, FinishedAt: jun15,
			},
			target: Anime{
				Status: StatusPlanToWatch, Score: 0, Progress: 0,
				StartedAt: jun1,
			},
			want: true,
		},
		{
			name: "watching different finishedAt ignored",
			source: Anime{
				Status: StatusWatching, Score: 8, Progress: 5,
				FinishedAt: jun15,
			},
			target: Anime{
				Status: StatusWatching, Score: 8, Progress: 5,
				FinishedAt: jul15,
			},
			want: true,
		},
		{
			name: "completed different finishedAt triggers update",
			source: Anime{
				Status: StatusCompleted, Score: 8, Progress: 12,
				FinishedAt: jun15,
			},
			target: Anime{
				Status: StatusCompleted, Score: 8, Progress: 12,
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

func TestAnime_GetStringDiffWithTarget_Dates(t *testing.T) {
	jun1 := timePtr(time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC))
	jul1 := timePtr(time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name       string
		source     Anime
		target     Target
		wantField  string
		shouldShow bool
	}{
		{
			name: "dates differ shows StartedAt in diff",
			source: Anime{
				Status: StatusCompleted, Score: 8, Progress: 12,
				StartedAt: jun1,
			},
			target: Anime{
				Status: StatusCompleted, Score: 8, Progress: 12,
				StartedAt: jul1,
			},
			wantField:  "StartedAt",
			shouldShow: true,
		},
		{
			name: "dates same does not show StartedAt in diff",
			source: Anime{
				Status: StatusCompleted, Score: 8, Progress: 12,
				StartedAt: jun1,
			},
			target: Anime{
				Status: StatusCompleted, Score: 8, Progress: 12,
				StartedAt: jun1,
			},
			wantField:  "StartedAt",
			shouldShow: false,
		},
		{
			name: "source date nil target date set does not show",
			source: Anime{
				Status: StatusCompleted, Score: 8, Progress: 12,
			},
			target: Anime{
				Status: StatusCompleted, Score: 8, Progress: 12,
				StartedAt: jun1,
			},
			wantField:  "StartedAt",
			shouldShow: true,
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
