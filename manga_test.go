package main

import (
	"testing"
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
				IDAnilist: 0,    // MAL source doesn't know AniList ID
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
