package main

import (
	"context"
	"testing"

	"github.com/rl404/verniy"
)

// Test constructors
func TestNewMALAnimeService(t *testing.T) {
	var client *MyAnimeListClient
	service := NewMALAnimeService(client)

	if service.client != client {
		t.Error("client should be set")
	}
}

func TestNewMALMangaService(t *testing.T) {
	var client *MyAnimeListClient
	service := NewMALMangaService(client)

	if service.client != client {
		t.Error("client should be set")
	}
}

func TestNewAniListAnimeService(t *testing.T) {
	var client *AnilistClient
	scoreFormat := verniy.ScoreFormatPoint10
	service := NewAniListAnimeService(client, scoreFormat, false)

	if service.client != client {
		t.Error("client should be set")
	}
	if service.scoreFormat != scoreFormat {
		t.Error("scoreFormat should be set")
	}
}

func TestNewAniListMangaService(t *testing.T) {
	var client *AnilistClient
	scoreFormat := verniy.ScoreFormatPoint10
	service := NewAniListMangaService(client, scoreFormat, false)

	if service.client != client {
		t.Error("client should be set")
	}
	if service.scoreFormat != scoreFormat {
		t.Error("scoreFormat should be set")
	}
}

func TestNewStrategyChain(t *testing.T) {
	strategy1 := IDStrategy{}
	strategy2 := TitleStrategy{}

	chain := NewStrategyChain(strategy1, strategy2)

	if len(chain.strategies) != 2 {
		t.Errorf("Expected 2 strategies, got %d", len(chain.strategies))
	}
}

func TestNewStatistics(t *testing.T) {
	stats := NewStatistics()

	if stats.StatusCounts == nil {
		t.Error("StatusCounts should be initialized")
	}
	if stats.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
}

// Test service client assignment
func TestMALAnimeService_Client(t *testing.T) {
	client := &MyAnimeListClient{}
	service := NewMALAnimeService(client)

	if service.client != client {
		t.Error("client should be assigned correctly")
	}
}

func TestMALMangaService_Client(t *testing.T) {
	client := &MyAnimeListClient{}
	service := NewMALMangaService(client)

	if service.client != client {
		t.Error("client should be assigned correctly")
	}
}

func TestAniListAnimeService_Client(t *testing.T) {
	client := &AnilistClient{}
	service := NewAniListAnimeService(client, verniy.ScoreFormatPoint10, false)

	if service.client != client {
		t.Error("client should be assigned correctly")
	}
}

func TestAniListMangaService_Client(t *testing.T) {
	client := &AnilistClient{}
	service := NewAniListMangaService(client, verniy.ScoreFormatPoint10, false)

	if service.client != client {
		t.Error("client should be assigned correctly")
	}
}

// Test service score format handling
func TestAniListAnimeService_ScoreFormat(t *testing.T) {
	tests := []struct {
		name        string
		scoreFormat verniy.ScoreFormat
	}{
		{"POINT_10", verniy.ScoreFormatPoint10},
		{"POINT_100", verniy.ScoreFormatPoint100},
		{"POINT_5", verniy.ScoreFormatPoint5},
		{"POINT_3", verniy.ScoreFormatPoint3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAniListAnimeService(nil, tt.scoreFormat, false)

			if service.scoreFormat != tt.scoreFormat {
				t.Error("scoreFormat should be set correctly")
			}
		})
	}
}

func TestAniListMangaService_ScoreFormat(t *testing.T) {
	tests := []struct {
		name        string
		scoreFormat verniy.ScoreFormat
	}{
		{"POINT_10", verniy.ScoreFormatPoint10},
		{"POINT_100", verniy.ScoreFormatPoint100},
		{"POINT_5", verniy.ScoreFormatPoint5},
		{"POINT_3", verniy.ScoreFormatPoint3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAniListMangaService(nil, tt.scoreFormat, false)

			if service.scoreFormat != tt.scoreFormat {
				t.Error("scoreFormat should be set correctly")
			}
		})
	}
}

// Test service Update methods with invalid source type (type assertion)
func TestMALAnimeService_Update_InvalidSource(t *testing.T) {
	service := NewMALAnimeService(nil)
	ctx := context.Background()

	// Pass Manga as source to Anime service - should fail type assertion
	mangaSource := Manga{IDMal: 12345, TitleEN: "Test Manga"}
	err := service.Update(ctx, 12345, mangaSource, "[Test]")

	if err == nil {
		t.Error("Expected error when source is not Anime")
	}
	// The error message should mention "not an anime"
	if err != nil && !containsString(err.Error(), "not an anime") {
		t.Errorf("Expected 'not an anime' error, got: %v", err)
	}
}

func TestMALMangaService_Update_InvalidSource(t *testing.T) {
	service := NewMALMangaService(nil)
	ctx := context.Background()

	// Pass Anime as source to Manga service - should fail type assertion
	animeSource := Anime{IDMal: 12345, TitleEN: "Test Anime"}
	err := service.Update(ctx, 12345, animeSource, "[Test]")

	if err == nil {
		t.Error("Expected error when source is not Manga")
	}
	// The error message should mention "not a manga"
	if err != nil && !containsString(err.Error(), "not a manga") {
		t.Errorf("Expected 'not a manga' error, got: %v", err)
	}
}

func TestAniListAnimeService_Update_InvalidSource(t *testing.T) {
	service := NewAniListAnimeService(nil, verniy.ScoreFormatPoint10, false)
	ctx := context.Background()

	// Pass Manga as source to Anime service
	mangaSource := Manga{IDMal: 12345, TitleEN: "Test Manga"}
	err := service.Update(ctx, 12345, mangaSource, "[Test]")

	if err == nil {
		t.Error("Expected error when source is not Anime")
	}
	if err != nil && !containsString(err.Error(), "not an anime") {
		t.Errorf("Expected 'not an anime' error, got: %v", err)
	}
}

func TestAniListMangaService_Update_InvalidSource(t *testing.T) {
	service := NewAniListMangaService(nil, verniy.ScoreFormatPoint10, false)
	ctx := context.Background()

	// Pass Anime as source to Manga service
	animeSource := Anime{IDMal: 12345, TitleEN: "Test Anime"}
	err := service.Update(ctx, 12345, animeSource, "[Test]")

	if err == nil {
		t.Error("Expected error when source is not Manga")
	}
	if err != nil && !containsString(err.Error(), "not a manga") {
		t.Errorf("Expected 'not a manga' error, got: %v", err)
	}
}

// Helper function for string contains check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
