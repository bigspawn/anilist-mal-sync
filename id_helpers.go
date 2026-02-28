package main

// GetTargetIDWithDirection returns the appropriate target ID based on direction parameter
func GetTargetIDWithDirection(source Source, direction SyncDirection) TargetID {
	if direction == SyncDirectionReverse {
		// Reverse: MAL → AniList, target is AniList ID
		switch v := source.(type) {
		case Anime:
			return TargetID(v.IDAnilist)
		case Manga:
			return TargetID(v.IDAnilist)
		}
	}
	// Forward: AniList → MAL, target is MAL ID
	switch v := source.(type) {
	case Anime:
		return TargetID(v.IDMal)
	case Manga:
		return TargetID(v.IDMal)
	}
	return 0
}

// GetSourceIDWithDirection returns the appropriate source ID based on direction parameter
func GetSourceIDWithDirection(source Source, direction SyncDirection) int {
	if direction == SyncDirectionReverse {
		// Reverse: source is MAL, return MAL ID
		switch v := source.(type) {
		case Anime:
			return v.IDMal
		case Manga:
			return v.IDMal
		}
	}
	// Forward: source is AniList, return AniList ID
	switch v := source.(type) {
	case Anime:
		return v.IDAnilist
	case Manga:
		return v.IDAnilist
	}
	return 0
}
