package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestUpdater_Update(t *testing.T) {
	tests := []struct {
		name          string
		sources       []Source
		targets       []Target
		ignoreTitles  map[string]struct{}
		forceSync     bool
		dryRun        bool
		expectUpdate  bool
		expectSkipped bool
		cancelContext bool
	}{
		{
			name: "force sync - successful update",
			sources: []Source{
				Anime{
					IDMal:       12345,
					IDAnilist:   67890,
					TitleEN:     "Test Anime",
					Status:      StatusWatching,
					Score:       8,
					Progress:    10,
					NumEpisodes: 12,
				},
			},
			targets: []Target{
				Anime{
					IDMal:       54321, // Different ID - doesn't matter in force sync
					IDAnilist:   99999,
					TitleEN:     "Different Anime",
					Status:      StatusCompleted,
					Score:       9,
					Progress:    12,
					NumEpisodes: 12,
				},
			},
			forceSync:     true,
			dryRun:        false,
			expectUpdate:  true,
			expectSkipped: false,
		},
		{
			name: "force sync - dry run mode",
			sources: []Source{
				Anime{
					IDMal:       12345,
					IDAnilist:   67890,
					TitleEN:     "Test Anime",
					Status:      StatusWatching,
					Score:       8,
					Progress:    10,
					NumEpisodes: 12,
				},
			},
			targets: []Target{
				Anime{
					IDMal:       54321,
					IDAnilist:   99999,
					TitleEN:     "Different Anime",
					Status:      StatusCompleted,
					Score:       9,
					Progress:    12,
					NumEpisodes: 12,
				},
			},
			forceSync:     true,
			dryRun:        true,
			expectUpdate:  true,
			expectSkipped: false,
		},
		{
			name: "skip when in ignore list",
			sources: []Source{
				Anime{
					IDMal:       12345,
					IDAnilist:   67890,
					TitleEN:     "Ignored Anime",
					Status:      StatusWatching,
					Score:       8,
					Progress:    10,
					NumEpisodes: 12,
				},
			},
			targets: []Target{
				Anime{
					IDMal:       12345,
					IDAnilist:   67890,
					TitleEN:     "Ignored Anime",
					Status:      StatusCompleted,
					Score:       9,
					Progress:    12,
					NumEpisodes: 12,
				},
			},
			ignoreTitles: map[string]struct{}{
				strings.ToLower("ignored anime"): {},
			},
			forceSync:     true,
			dryRun:        false,
			expectUpdate:  false,
			expectSkipped: true,
		},
		{
			name: "skip source with empty status",
			sources: []Source{
				Anime{
					IDMal:       12345,
					IDAnilist:   67890,
					TitleEN:     "Test Anime",
					Status:      "",
					Score:       8,
					Progress:    10,
					NumEpisodes: 12,
				},
			},
			targets:       []Target{},
			forceSync:     true,
			dryRun:        false,
			expectUpdate:  false,
			expectSkipped: false,
		},
		{
			name: "context cancellation stops processing",
			sources: []Source{
				Anime{
					IDMal:       12345,
					IDAnilist:   67890,
					TitleEN:     "Test Anime",
					Status:      StatusWatching,
					Score:       8,
					Progress:    10,
					NumEpisodes: 12,
				},
			},
			targets: []Target{
				Anime{
					IDMal:       54321,
					IDAnilist:   99999,
					TitleEN:     "Different Anime",
					Status:      StatusCompleted,
					Score:       9,
					Progress:    12,
					NumEpisodes: 12,
				},
			},
			forceSync:     true,
			dryRun:        false,
			cancelContext: true,
			expectUpdate:  false,
			expectSkipped: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			if tt.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			mockService := NewMockMediaService(ctrl)
			updater := &Updater{
				Prefix:       "[Test]",
				Statistics:   NewStatistics(),
				IgnoreTitles: tt.ignoreTitles,
				Service:      mockService,
				ForceSync:    tt.forceSync,
				DryRun:       tt.dryRun,
			}

			// Set up expectations for service update
			if tt.expectUpdate && !tt.dryRun && !tt.cancelContext {
				mockService.EXPECT().Update(ctx, gomock.Any(), gomock.Any(), "[Test]").Return(nil)
			}

			report := NewSyncReport()
			updater.Update(ctx, tt.sources, tt.targets, report)

			if tt.expectUpdate {
				if tt.dryRun {
					if updater.Statistics.DryRunCount == 0 {
						t.Error("Expected dry run, but got none")
					}
				} else {
					if updater.Statistics.UpdatedCount == 0 {
						t.Error("Expected update, but got none")
					}
				}
			}

			if tt.expectSkipped {
				if updater.Statistics.SkippedCount == 0 {
					t.Error("Expected skip, but got none")
				}
			}
		})
	}
}

func TestUpdater_updateTarget(t *testing.T) {
	tests := []struct {
		name         string
		src          Source
		id           TargetID
		serviceErr   error
		expectUpdate bool
		expectError  bool
	}{
		{
			name: "successful update",
			src: Anime{
				IDMal:    12345,
				TitleEN:  "Test Anime",
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			id:           12345,
			serviceErr:   nil,
			expectUpdate: true,
			expectError:  false,
		},
		{
			name: "service update error",
			src: Anime{
				IDMal:    12345,
				TitleEN:  "Test Anime",
				Status:   StatusWatching,
				Score:    8,
				Progress: 10,
			},
			id:           12345,
			serviceErr:   errors.New("update failed"),
			expectUpdate: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			mockService := NewMockMediaService(ctrl)

			updater := &Updater{
				Prefix:     "[Test]",
				Statistics: NewStatistics(),
				Service:    mockService,
			}

			if tt.serviceErr != nil {
				mockService.EXPECT().Update(ctx, tt.id, tt.src, "[Test]").Return(tt.serviceErr)
			} else {
				mockService.EXPECT().Update(ctx, tt.id, tt.src, "[Test]").Return(nil)
			}

			updater.updateTarget(ctx, tt.id, tt.src)

			if tt.expectUpdate {
				if updater.Statistics.UpdatedCount == 0 {
					t.Error("Expected update, but got none")
				}
			}

			if tt.expectError {
				if updater.Statistics.ErrorCount == 0 {
					t.Error("Expected error, but got none")
				}
			}
		})
	}
}

func TestUpdater_updateSourceByTargets(t *testing.T) {
	tests := []struct {
		name         string
		src          Source
		targets      map[TargetID]Target
		forceSync    bool
		dryRun       bool
		expectUpdate bool
		expectSkip   bool
	}{
		{
			name: "force sync with different progress",
			src: Anime{
				IDMal:       12345,
				TitleEN:     "Test Anime",
				Status:      StatusWatching,
				Score:       8,
				Progress:    10,
				NumEpisodes: 12,
			},
			targets: map[TargetID]Target{
				12345: Anime{
					IDMal:       12345,
					TitleEN:     "Test Anime",
					Status:      StatusCompleted,
					Score:       9,
					Progress:    12,
					NumEpisodes: 12,
				},
			},
			forceSync:    true,
			dryRun:       false,
			expectUpdate: true,
		},
		{
			name: "dry run - update without service call",
			src: Anime{
				IDMal:       12345,
				TitleEN:     "Test Anime",
				Status:      StatusWatching,
				Score:       8,
				Progress:    10,
				NumEpisodes: 12,
			},
			targets: map[TargetID]Target{
				12345: Anime{
					IDMal:       12345,
					TitleEN:     "Test Anime",
					Status:      StatusCompleted,
					Score:       9,
					Progress:    12,
					NumEpisodes: 12,
				},
			},
			forceSync:    true,
			dryRun:       true,
			expectUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			mockService := NewMockMediaService(ctrl)

			updater := &Updater{
				Prefix:     "[Test]",
				Statistics: NewStatistics(),
				Service:    mockService,
				ForceSync:  tt.forceSync,
				DryRun:     tt.dryRun,
			}

			// Set up expectations for service update
			if tt.expectUpdate && !tt.dryRun {
				mockService.EXPECT().Update(ctx, gomock.Any(), tt.src, "[Test]").Return(nil)
			}

			report := NewSyncReport()
			updater.updateSourceByTargets(ctx, tt.src, tt.targets, report)

			if tt.expectUpdate {
				if tt.dryRun {
					if updater.Statistics.DryRunCount == 0 {
						t.Error("Expected dry run, but got none")
					}
				} else {
					if updater.Statistics.UpdatedCount == 0 {
						t.Error("Expected update, but got none")
					}
				}
			}

			if tt.expectSkip {
				if updater.Statistics.SkippedCount == 0 {
					t.Error("Expected skip, but got none")
				}
			}
		})
	}
}

func TestUpdater_DryRunRecordsInDryRunItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockService := NewMockMediaService(ctrl)

	updater := &Updater{
		Prefix:     "[Test]",
		Statistics: NewStatistics(),
		Service:    mockService,
		ForceSync:  true,
		DryRun:     true,
	}

	src := Anime{
		IDMal:       12345,
		TitleEN:     "Test Anime",
		Status:      StatusWatching,
		Score:       8,
		Progress:    10,
		NumEpisodes: 12,
	}

	report := NewSyncReport()
	updater.updateSourceByTargets(ctx, src, map[TargetID]Target{}, report)

	if updater.Statistics.DryRunCount == 0 {
		t.Error("Expected dry run count > 0")
	}
	if updater.Statistics.UpdatedCount != 0 {
		t.Error("Expected updated count to be 0 in dry run mode")
	}
	if len(updater.Statistics.DryRunItems) == 0 {
		t.Error("Expected dry run items to be recorded")
	}
	if !updater.Statistics.DryRunItems[0].IsDryRun {
		t.Error("Expected IsDryRun to be true")
	}
}
