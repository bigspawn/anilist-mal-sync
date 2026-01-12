package main

import (
	"testing"
	"time"
)

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
