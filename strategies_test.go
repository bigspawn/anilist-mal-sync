package main

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"
)

// TestIDStrategy_FindsExistingTarget tests that IDStrategy finds targets by ID when they exist
func TestIDStrategy_FindsExistingTarget(t *testing.T) {
	ctx := context.Background()
	strategy := IDStrategy{}

	source := Anime{
		IDMal:       123,
		TitleEN:     "Test Anime",
		NumEpisodes: 12,
	}

	existingTargets := map[TargetID]Target{
		123: Anime{
			IDMal:       123,
			TitleEN:     "Test Anime",
			NumEpisodes: 12,
		},
	}

	target, found, err := strategy.FindTarget(ctx, source, existingTargets, "[Test]")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !found {
		t.Error("Expected to find target by ID, but didn't")
	}

	if target == nil {
		t.Error("Expected non-nil target")
	}
}

// TestIDStrategy_NotFoundInUserList tests that IDStrategy returns false when ID not in user's list
func TestIDStrategy_NotFoundInUserList(t *testing.T) {
	ctx := context.Background()
	strategy := IDStrategy{}

	source := Anime{
		IDMal:       44983, // DanMachi OVA - not in user's list
		TitleEN:     "Is It Wrong to Try to Pick Up Girls in a Dungeon? III OVA",
		NumEpisodes: 1,
	}

	existingTargets := map[TargetID]Target{
		28121: Anime{
			IDMal:       28121, // Main series - in user's list
			TitleEN:     "Is It Wrong to Try to Pick Up Girls in a Dungeon?",
			NumEpisodes: 13,
		},
	}

	target, found, err := strategy.FindTarget(ctx, source, existingTargets, "[Test]")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if found {
		t.Error("Expected NOT to find target (ID not in list), but did")
	}

	if target != nil {
		t.Errorf("Expected nil target when not found, got %v", target)
	}
}

// TestTitleStrategy_ShouldRejectMismatchedMALIDs tests the bug where TitleStrategy
// matches entries with different MAL IDs, causing repeated updates
func TestTitleStrategy_ShouldRejectMismatchedMALIDs(t *testing.T) {
	ctx := context.Background()
	strategy := TitleStrategy{}

	tests := []struct {
		name           string
		source         Anime
		existingTarget Anime
		shouldMatch    bool
		description    string
	}{
		{
			name: "DanMachi OVA wrongly matched to main series",
			source: Anime{
				IDMal: 44983, // OVA
				TitleEN: "Is It Wrong to Try to Pick Up Girls in a Dungeon? III: " +
					"Is It Wrong to Try to Find a Hot Spring in Orario? -Bath God Forever-",
				TitleJP:     "ダンジョンに出会いを求めるのは間違っているだろうかⅢ OVA",
				NumEpisodes: 1,
			},
			existingTarget: Anime{
				IDMal:       28121, // Main series
				TitleEN:     "Is It Wrong to Try to Pick Up Girls in a Dungeon?",
				TitleJP:     "ダンジョンに出会いを求めるのは間違っているだろうか",
				NumEpisodes: 13,
			},
			shouldMatch: false, // Should NOT match - different MAL IDs!
			description: "OVA (ID 44983) should not match main series (ID 28121) even though title is similar",
		},
		{
			name: "Girls Band Cry Movie wrongly matched to TV series",
			source: Anime{
				IDMal:       62550, // Movie
				TitleJP:     "ガールズバンドクライ (新作映画)",
				NumEpisodes: 1,
			},
			existingTarget: Anime{
				IDMal:       55102, // TV series
				TitleEN:     "Girls Band Cry",
				TitleJP:     "ガールズバンドクライ",
				NumEpisodes: 13,
			},
			shouldMatch: false, // Should NOT match - different MAL IDs!
			description: "Movie (ID 62550) should not match TV series (ID 55102) even after title normalization",
		},
		{
			name: "Same MAL ID should match",
			source: Anime{
				IDMal:       12345,
				TitleEN:     "Test Anime",
				NumEpisodes: 12,
			},
			existingTarget: Anime{
				IDMal:       12345,
				TitleEN:     "Test Anime",
				NumEpisodes: 12,
			},
			shouldMatch: true, // Should match - same MAL ID
			description: "Entries with same MAL ID should match",
		},
		{
			name: "Source without MAL ID can match by title",
			source: Anime{
				IDMal:       0, // No MAL ID
				TitleEN:     "Test Anime",
				NumEpisodes: 12,
			},
			existingTarget: Anime{
				IDMal:       12345,
				TitleEN:     "Test Anime",
				NumEpisodes: 12,
			},
			shouldMatch: true, // Can match - source has no MAL ID
			description: "Source without MAL ID should be able to match by title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingTargets := map[TargetID]Target{
				TargetID(tt.existingTarget.IDMal): tt.existingTarget,
			}

			target, found, err := strategy.FindTarget(ctx, tt.source, existingTargets, "[Test]")
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.shouldMatch {
				if !found {
					t.Errorf("Expected to find match (%s), but didn't", tt.description)
				}
				if target == nil {
					t.Error("Expected non-nil target when match should be found")
				}
			} else if found {
				t.Errorf("Expected NOT to find match (%s), but did! This is the BUG.", tt.description)
				t.Errorf("Source MAL ID: %d, Target MAL ID: %d", tt.source.IDMal, target.GetTargetID())
				t.Errorf("This will cause repeated updates on every sync!")
			}
		})
	}
}

// TestTitleStrategy_ShouldRejectLargeEpisodeCountDifference tests that TitleStrategy
// should reject matches when episode counts differ significantly
func TestTitleStrategy_ShouldRejectLargeEpisodeCountDifference(t *testing.T) {
	ctx := context.Background()
	strategy := TitleStrategy{}

	tests := []struct {
		name           string
		source         Anime
		existingTarget Anime
		shouldMatch    bool
		description    string
	}{
		{
			name: "1 episode vs 13 episodes - should reject",
			source: Anime{
				IDMal:       0, // No MAL ID, so title matching is allowed
				TitleJP:     "テストアニメ (新作映画)",
				NumEpisodes: 1,
			},
			existingTarget: Anime{
				IDMal:       12345,
				TitleJP:     "テストアニメ",
				NumEpisodes: 13,
			},
			shouldMatch: false,
			description: "Movie (1 ep) should not match TV series (13 eps) - 1200% difference",
		},
		{
			name: "12 episodes vs 13 episodes - should accept",
			source: Anime{
				IDMal:       0,
				TitleEN:     "Test Anime",
				NumEpisodes: 12,
			},
			existingTarget: Anime{
				IDMal:       12345,
				TitleEN:     "Test Anime",
				NumEpisodes: 13,
			},
			shouldMatch: true,
			description: "12 vs 13 episodes - only 8% difference, acceptable",
		},
		{
			name: "24 episodes vs 25 episodes - should accept",
			source: Anime{
				IDMal:       0,
				TitleEN:     "Test Anime Season 2",
				NumEpisodes: 24,
			},
			existingTarget: Anime{
				IDMal:       12346,
				TitleEN:     "Test Anime Season 2",
				NumEpisodes: 25,
			},
			shouldMatch: true,
			description: "24 vs 25 episodes - only 4% difference, acceptable",
		},
		{
			name: "Both have 0 episodes (unknown) - should accept",
			source: Anime{
				IDMal:       0,
				TitleEN:     "Upcoming Anime",
				NumEpisodes: 0,
			},
			existingTarget: Anime{
				IDMal:       12347,
				TitleEN:     "Upcoming Anime",
				NumEpisodes: 0,
			},
			shouldMatch: true,
			description: "Both have unknown episode count - should match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingTargets := map[TargetID]Target{
				TargetID(tt.existingTarget.IDMal): tt.existingTarget,
			}

			target, found, err := strategy.FindTarget(ctx, tt.source, existingTargets, "[Test]")
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.shouldMatch {
				if !found {
					t.Errorf("Expected to find match (%s), but didn't", tt.description)
				}
			} else if found {
				t.Errorf("Expected NOT to find match (%s), but did!", tt.description)
				targetAnime, ok := target.(Anime)
				if ok {
					t.Errorf("Source episodes: %d, Target episodes: %d",
						tt.source.NumEpisodes, targetAnime.NumEpisodes)
				}
			}
		})
	}
}

// TestStrategyChain_Integration tests the full strategy chain behavior
func TestStrategyChain_Integration(t *testing.T) {
	ctx := context.Background()

	// Real-world scenario: DanMachi OVA not in user's MAL list
	source := Anime{
		IDMal: 44983, // OVA - not in list
		TitleEN: "Is It Wrong to Try to Pick Up Girls in a Dungeon? III: " +
			"Is It Wrong to Try to Find a Hot Spring in Orario? -Bath God Forever-",
		TitleJP:     "ダンジョンに出会いを求めるのは間違っているだろうかⅢ OVA",
		NumEpisodes: 1,
		Status:      StatusCompleted,
		Score:       6,
		Progress:    1,
	}

	// User has the main series in their list
	existingTargets := map[TargetID]Target{
		28121: Anime{
			IDMal:       28121, // Main series
			TitleEN:     "Is It Wrong to Try to Pick Up Girls in a Dungeon?",
			TitleJP:     "ダンジョンに出会いを求めるのは間違っているだろうか",
			NumEpisodes: 13,
			Status:      StatusCompleted,
			Score:       7,
			Progress:    13,
		},
	}

	chain := NewStrategyChain(
		IDStrategy{},
		TitleStrategy{},
	)

	target, err := chain.FindTarget(ctx, source, existingTargets, "[Test]")

	// Expected behavior after fix: should return error (no target found)
	// Current buggy behavior: returns main series (wrong match)
	if err == nil {
		t.Error("Expected error (no target found), but got a target")
		if target != nil {
			t.Errorf("BUG: Found wrong target! Source MAL ID: %d, Target MAL ID: %d",
				source.IDMal, target.GetTargetID())
			t.Error("This causes repeated updates on every sync!")
		}
	}
}

// TestMALIDStrategy_FindsTargetByMALID tests that MALIDStrategy finds targets by MAL ID
func TestMALIDStrategy_FindsTargetByMALID(t *testing.T) {
	// Set reverse direction for this test (MAL -> AniList sync)
	origReverse := reverseDirection
	defer func() { reverseDirection = origReverse }()
	trueVal := true
	reverseDirection = &trueVal

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Source from MAL with IDMal set
	source := Anime{
		IDMal:       37341, // Laid-Back Camp Specials on MAL
		IDAnilist:   0,
		TitleEN:     "Laid-Back Camp Specials",
		TitleJP:     "ゆるキャン△ スペシャル",
		NumEpisodes: 3,
	}

	// Target from AniList with different title but same MAL ID
	apiTarget := Anime{
		IDMal:       37341,
		IDAnilist:   101206,
		TitleEN:     "",
		TitleJP:     "ゆるキャン△ OVA", // Different title on AniList
		NumEpisodes: 3,
	}

	existingTargets := map[TargetID]Target{
		101206: apiTarget,
	}

	mockService := NewMockMediaServiceWithMALID(ctrl)
	mockService.EXPECT().GetByMALID(ctx, 37341, "[Test]").Return(apiTarget, nil)

	strategy := MALIDStrategy{Service: mockService}

	target, found, err := strategy.FindTarget(ctx, source, existingTargets, "[Test]")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !found {
		t.Error("Expected to find target by MAL ID, but didn't")
	}

	if target == nil {
		t.Error("Expected non-nil target")
	}

	targetAnime, ok := target.(Anime)
	if !ok {
		t.Error("Expected target to be Anime type")
	}

	if targetAnime.IDMal != 37341 {
		t.Errorf("Expected MAL ID 37341, got %d", targetAnime.IDMal)
	}

	if targetAnime.IDAnilist != 101206 {
		t.Errorf("Expected AniList ID 101206, got %d", targetAnime.IDAnilist)
	}
}

// TestMALIDStrategy_ReturnsExistingUserTarget tests that MALIDStrategy returns existing target from user's list
func TestMALIDStrategy_ReturnsExistingUserTarget(t *testing.T) {
	// Set reverse direction for this test (MAL -> AniList sync)
	origReverse := reverseDirection
	defer func() { reverseDirection = origReverse }()
	trueVal := true
	reverseDirection = &trueVal

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	source := Anime{
		IDMal:       37341,
		IDAnilist:   0,
		TitleEN:     "Laid-Back Camp Specials",
		NumEpisodes: 3,
	}

	// API returns this target
	apiTarget := Anime{
		IDMal:       37341,
		IDAnilist:   101206,
		TitleJP:     "ゆるキャン△ OVA",
		NumEpisodes: 3,
	}

	// But user has this in their list (with different status)
	userTarget := Anime{
		IDMal:       37341,
		IDAnilist:   101206,
		TitleJP:     "ゆるキャン△ OVA",
		NumEpisodes: 3,
		Status:      StatusCompleted,
		Progress:    3,
		Score:       8,
	}

	existingTargets := map[TargetID]Target{
		101206: userTarget,
	}

	mockService := NewMockMediaServiceWithMALID(ctrl)
	mockService.EXPECT().GetByMALID(ctx, 37341, "[Test]").Return(apiTarget, nil)

	strategy := MALIDStrategy{Service: mockService}

	target, found, err := strategy.FindTarget(ctx, source, existingTargets, "[Test]")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !found {
		t.Error("Expected to find target")
	}

	// Should return the user's target (with status), not the API target
	targetAnime, ok := target.(Anime)
	if !ok {
		t.Error("Expected target to be Anime type")
	}

	if targetAnime.Status != StatusCompleted {
		t.Errorf("Expected user's target status (completed), got %v", targetAnime.Status)
	}

	if targetAnime.Progress != 3 {
		t.Errorf("Expected user's target progress (3), got %d", targetAnime.Progress)
	}
}

// TestMALIDStrategy_SkipsZeroMALID tests that MALIDStrategy skips when source has no MAL ID
func TestMALIDStrategy_SkipsZeroMALID(t *testing.T) {
	// Set reverse direction for this test (MAL -> AniList sync)
	origReverse := reverseDirection
	defer func() { reverseDirection = origReverse }()
	trueVal := true
	reverseDirection = &trueVal

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	source := Anime{
		IDMal:       0, // No MAL ID
		IDAnilist:   0,
		TitleEN:     "Test Anime",
		NumEpisodes: 12,
	}

	existingTargets := map[TargetID]Target{}

	mockService := NewMockMediaServiceWithMALID(ctrl)
	// GetByMALID should not be called when source ID is 0
	mockService.EXPECT().GetByMALID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	strategy := MALIDStrategy{Service: mockService}

	target, found, err := strategy.FindTarget(ctx, source, existingTargets, "[Test]")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if found {
		t.Error("Expected not to find target (source has no MAL ID)")
	}

	if target != nil {
		t.Errorf("Expected nil target, got %v", target)
	}
}

// TestMALIDStrategy_ContextCancellation tests that MALIDStrategy respects context cancellation
func TestMALIDStrategy_ContextCancellation(t *testing.T) {
	// Set reverse direction for this test (MAL -> AniList sync)
	origReverse := reverseDirection
	defer func() { reverseDirection = origReverse }()
	trueVal := true
	reverseDirection = &trueVal

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	source := Anime{
		IDMal:       12345,
		IDAnilist:   0,
		TitleEN:     "Test Anime",
		NumEpisodes: 12,
	}

	existingTargets := map[TargetID]Target{}

	mockService := NewMockMediaServiceWithMALID(ctrl)
	// GetByMALID should not be called when context is cancelled
	mockService.EXPECT().GetByMALID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	strategy := MALIDStrategy{Service: mockService}

	target, found, err := strategy.FindTarget(ctx, source, existingTargets, "[Test]")
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}

	if found {
		t.Error("Expected not to find target when context is cancelled")
	}

	if target != nil {
		t.Errorf("Expected nil target, got %v", target)
	}
}

// TestMALIDStrategy_ErrorHandling tests that MALIDStrategy properly handles API errors
func TestMALIDStrategy_ErrorHandling(t *testing.T) {
	// Set reverse direction for this test (MAL -> AniList sync)
	origReverse := reverseDirection
	defer func() { reverseDirection = origReverse }()
	trueVal := true
	reverseDirection = &trueVal

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	source := Anime{
		IDMal:       99999, // Non-existent MAL ID
		IDAnilist:   0,
		TitleEN:     "Non-existent Anime",
		NumEpisodes: 12,
	}

	existingTargets := map[TargetID]Target{}

	mockService := NewMockMediaServiceWithMALID(ctrl)
	mockService.EXPECT().GetByMALID(ctx, 99999, "[Test]").Return(nil, fmt.Errorf("no anime found with MAL ID %d", 99999))

	strategy := MALIDStrategy{Service: mockService}

	target, found, err := strategy.FindTarget(ctx, source, existingTargets, "[Test]")
	if err == nil {
		t.Error("Expected error from API, got nil")
	}

	if found {
		t.Error("Expected not to find target when API returns error")
	}

	if target != nil {
		t.Errorf("Expected nil target, got %v", target)
	}
}

// TestAnime_GetSourceID tests that GetSourceID returns correct source ID based on sync direction
func TestAnime_GetSourceID(t *testing.T) {
	origReverse := reverseDirection
	defer func() { reverseDirection = origReverse }()

	falseVal := false
	trueVal := true

	source := Anime{
		IDMal:     12345,
		IDAnilist: 67890,
		TitleEN:   "Test Anime",
	}

	// Normal sync: source is AniList, so source ID is IDAnilist
	reverseDirection = &falseVal
	if got := source.GetSourceID(); got != 67890 {
		t.Errorf("Normal sync: expected source ID 67890 (AniList), got %d", got)
	}

	// Reverse sync: source is MAL, so source ID is IDMal
	reverseDirection = &trueVal
	if got := source.GetSourceID(); got != 12345 {
		t.Errorf("Reverse sync: expected source ID 12345 (MAL), got %d", got)
	}
}

// TestManga_GetSourceID tests that GetSourceID returns correct source ID based on sync direction
func TestManga_GetSourceID(t *testing.T) {
	origReverse := reverseDirection
	defer func() { reverseDirection = origReverse }()

	falseVal := false
	trueVal := true

	source := Manga{
		IDMal:     11111,
		IDAnilist: 22222,
		TitleEN:   "Test Manga",
	}

	// Normal sync: source is AniList, so source ID is IDAnilist
	reverseDirection = &falseVal
	if got := source.GetSourceID(); got != 22222 {
		t.Errorf("Normal sync: expected source ID 22222 (AniList), got %d", got)
	}

	// Reverse sync: source is MAL, so source ID is IDMal
	reverseDirection = &trueVal
	if got := source.GetSourceID(); got != 11111 {
		t.Errorf("Reverse sync: expected source ID 11111 (MAL), got %d", got)
	}
}
