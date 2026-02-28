package main

import (
	"testing"

	"github.com/nstratos/go-myanimelist/mal"
)

func TestStatus_GetMalStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		status    Status
		want      mal.AnimeStatus
		wantError bool
	}{
		{
			name:      "watching",
			status:    StatusWatching,
			want:      mal.AnimeStatusWatching,
			wantError: false,
		},
		{
			name:      "completed",
			status:    StatusCompleted,
			want:      mal.AnimeStatusCompleted,
			wantError: false,
		},
		{
			name:      "on_hold",
			status:    StatusOnHold,
			want:      mal.AnimeStatusOnHold,
			wantError: false,
		},
		{
			name:      "dropped",
			status:    StatusDropped,
			want:      mal.AnimeStatusDropped,
			wantError: false,
		},
		{
			name:      "plan_to_watch",
			status:    StatusPlanToWatch,
			want:      mal.AnimeStatusPlanToWatch,
			wantError: false,
		},
		{
			name:      "unknown",
			status:    StatusUnknown,
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid status",
			status:    Status("invalid"),
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.status.GetMalStatus()
			if tt.wantError {
				if err == nil {
					t.Error("GetMalStatus() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetMalStatus() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetMalStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_GetAnilistStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{
			name:   "watching",
			status: StatusWatching,
			want:   "CURRENT",
		},
		{
			name:   "completed",
			status: StatusCompleted,
			want:   "COMPLETED",
		},
		{
			name:   "on_hold",
			status: StatusOnHold,
			want:   "PAUSED",
		},
		{
			name:   "dropped",
			status: StatusDropped,
			want:   "DROPPED",
		},
		{
			name:   "plan_to_watch",
			status: StatusPlanToWatch,
			want:   "PLANNING",
		},
		{
			name:   "unknown",
			status: StatusUnknown,
			want:   "",
		},
		{
			name:   "invalid status",
			status: Status("invalid"),
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.GetAnilistStatus()
			if got != tt.want {
				t.Errorf("GetAnilistStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMangaStatus_GetMalStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		status    MangaStatus
		want      mal.MangaStatus
		wantError bool
	}{
		{
			name:      "reading",
			status:    MangaStatusReading,
			want:      mal.MangaStatusReading,
			wantError: false,
		},
		{
			name:      "completed",
			status:    MangaStatusCompleted,
			want:      mal.MangaStatusCompleted,
			wantError: false,
		},
		{
			name:      "on_hold",
			status:    MangaStatusOnHold,
			want:      mal.MangaStatusOnHold,
			wantError: false,
		},
		{
			name:      "dropped",
			status:    MangaStatusDropped,
			want:      mal.MangaStatusDropped,
			wantError: false,
		},
		{
			name:      "plan_to_read",
			status:    MangaStatusPlanToRead,
			want:      mal.MangaStatusPlanToRead,
			wantError: false,
		},
		{
			name:      "unknown",
			status:    MangaStatusUnknown,
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid status",
			status:    MangaStatus("invalid"),
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.status.GetMalStatus()
			if tt.wantError {
				if err == nil {
					t.Error("GetMalStatus() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetMalStatus() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetMalStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMangaStatus_GetAnilistStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status MangaStatus
		want   string
	}{
		{
			name:   "reading",
			status: MangaStatusReading,
			want:   "CURRENT",
		},
		{
			name:   "completed",
			status: MangaStatusCompleted,
			want:   "COMPLETED",
		},
		{
			name:   "on_hold",
			status: MangaStatusOnHold,
			want:   "PAUSED",
		},
		{
			name:   "dropped",
			status: MangaStatusDropped,
			want:   "DROPPED",
		},
		{
			name:   "plan_to_read",
			status: MangaStatusPlanToRead,
			want:   "PLANNING",
		},
		{
			name:   "unknown",
			status: MangaStatusUnknown,
			want:   "",
		},
		{
			name:   "invalid status",
			status: MangaStatus("invalid"),
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.GetAnilistStatus()
			if got != tt.want {
				t.Errorf("GetAnilistStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
