package main

// TargetID represents an ID in the target service (MAL or AniList)
type TargetID int

// Source represents an item from a source service (AniList or MAL)
type Source interface {
	GetStatusString() string
	GetTargetID() TargetID
	GetTitle() string
	GetStringDiffWithTarget(Target) string
	SameProgressWithTarget(Target) bool
	SameTypeWithTarget(Target) bool
	SameTitleWithTarget(Target) bool
	String() string
}

// Target represents an item in the target service (MAL or AniList)
type Target interface {
	GetTargetID() TargetID
	GetTitle() string
	String() string
}
