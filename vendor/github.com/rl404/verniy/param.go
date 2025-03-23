package verniy

// MediaEdgeParamVoiceActors is media edge param for voice actors.
type MediaEdgeParamVoiceActors struct {
	Language StaffLanguage
	Sort     []StaffSort
}

// CharacterEdgeParamVoiceActors is character edge param for voice actors.
type CharacterEdgeParamVoiceActors struct {
	Language StaffLanguage
	Sort     []StaffSort
}

// MediaEdgeParamVoiceActorRoles is media edge param for voice actor roles.
type MediaEdgeParamVoiceActorRoles struct {
	Language StaffLanguage
	Sort     []StaffSort
}

// CharacterEdgeParamVoiceActorRoles is character edge param for voice actor roles.
type CharacterEdgeParamVoiceActorRoles struct {
	Language StaffLanguage
	Sort     []StaffSort
}

// MediaParamCharacters is media param for characters.
type MediaParamCharacters struct {
	Page    int
	PerPage int
	Role    CharacterRole
	Sort    []CharacterSort
}

// CharacterParamMedia is character param for media.
type CharacterParamMedia struct {
	Type    MediaType
	OnList  *bool
	Page    int
	PerPage int
	Sort    []MediaSort
}

// StaffParamStaffMedia is staff param for staff media.
type StaffParamStaffMedia struct {
	Type    MediaType
	OnList  *bool
	Page    int
	PerPage int
	Sort    []MediaSort
}

// StaffParamCharacters is staff param for characters.
type StaffParamCharacters struct {
	Page    int
	PerPage int
	Sort    []CharacterSort
}

// StaffParamCharacterMedia is staff param for character media.
type StaffParamCharacterMedia struct {
	OnList  *bool
	Page    int
	PerPage int
	Sort    []MediaSort
}

// MediaParamStaff is media param for staff.
type MediaParamStaff struct {
	Page    int
	PerPage int
	Sort    []StaffSort
}

// MediaParamStudios is media param for studios.
type MediaParamStudios struct {
	IsMain *bool
	Sort   []StudioSort
}

// StudioParamMedia is studio param for media.
type StudioParamMedia struct {
	IsMain  *bool
	OnList  *bool
	Page    int
	PerPage int
	Sort    []MediaSort
}

// MediaParamAiringSchedule is media param for airing schedule.
type MediaParamAiringSchedule struct {
	NotYetAired *bool
	Page        int
	PerPage     int
}

// MediaParamTrends is media param for trends.
type MediaParamTrends struct {
	Page      int
	PerPage   int
	Releasing *bool
	Sort      []MediaTrendSort
}

// MediaParamReviews is media param for reviews.
type MediaParamReviews struct {
	Page    int
	PerPage int
	Sort    []ReviewSort
	Limit   int
}

// MediaParamRecommendations is media param for recommendation
type MediaParamRecommendations struct {
	Page    int
	PerPage int
	Sort    []RecommendationSort
}

// PageParamStudios is page param for studios.
type PageParamStudios struct {
	Search  string
	ID      int
	IDNot   int
	IDIn    []int
	IDNotIn []int
	Sort    []StudioSort
}

// PageParamMedia is page param for media.
type PageParamMedia struct {
	ID                  int
	IDMAL               int
	StartDate           int
	EndDate             int
	Season              MediaSeason
	SeasonYear          int
	Type                MediaType
	Format              MediaFormat
	Status              MediaStatus
	Episodes            int
	Duration            int
	Chapters            int
	Volumes             int
	IsAdult             *bool
	Genre               string
	Tag                 string
	MinimumTagRank      int
	TagCategory         string
	OnList              *bool
	LicensedBy          string
	AverageScore        int
	Popularity          int
	Source              MediaSource
	CountryOfOrigin     string
	Search              string
	IDNot               int
	IDIn                []int
	IDNotIn             []int
	IDMALNot            int
	IDMALIn             []int
	IDMALNotIn          []int
	StartDateGreater    int
	StartDateLesser     int
	StartDateLike       string
	EndDateGreater      int
	EndDateLesser       int
	EndDateLike         string
	FormatIn            []MediaFormat
	FormatNot           MediaFormat
	FormatNotIn         []MediaFormat
	StatusIn            []MediaStatus
	StatusNot           MediaStatus
	StatusNotIn         []MediaStatus
	EpisodesGreater     int
	EpisodesLesser      int
	DurationGreater     int
	DurationLesser      int
	ChaptersGreater     int
	ChaptersLesser      int
	VolumesGreater      int
	VolumesLesser       int
	GenreIn             []string
	GenreNotIn          []string
	TagIn               []string
	TagNotIn            []string
	TagCategoryIn       []string
	TagCategoryNotIn    []string
	LicensedByIn        []string
	AverageScoreNot     int
	AverageScoreGreater int
	AverageScoreLesser  int
	PopularityNot       int
	PopularityGreater   int
	PopularityLesser    int
	SourceIn            []MediaSource
	Sort                []MediaSort
}

// PageParamCharacters is page param for characters.
type PageParamCharacters struct {
	ID         int
	IsBirthday *bool
	Search     string
	IDNot      int
	IDIn       []int
	IDNotIn    []int
	Sort       []CharacterSort
}

// PageParamStaff is page param for staff.
type PageParamStaff struct {
	ID         int
	IsBirthday *bool
	Search     string
	IDNot      int
	IDIn       []int
	IDNotIn    []int
	Sort       []StaffSort
}

// UserParamFavourites is user param for favourites.
type UserParamFavourites struct {
	Page int
}

// FavouritesParamAnime is favourites param for anime.
type FavouritesParamAnime struct {
	Page    int
	PerPage int
}

// FavouritesParamManga is favourites param for manga.
type FavouritesParamManga struct {
	Page    int
	PerPage int
}

// FavouritesParamCharacters is favourites param for characters.
type FavouritesParamCharacters struct {
	Page    int
	PerPage int
}

// FavouritesParamStaff is favourites param for staff.
type FavouritesParamStaff struct {
	Page    int
	PerPage int
}

// FavouritesParamStudios is favourites param for studios.
type FavouritesParamStudios struct {
	Page    int
	PerPage int
}

// UserStatisticsParamFormats is user statistics param for formats.
type UserStatisticsParamFormats struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamStatuses is user statistics param for statuses.
type UserStatisticsParamStatuses struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamScores is user statistics param for scores.
type UserStatisticsParamScores struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamLengths is user statistics param for lengths.
type UserStatisticsParamLengths struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamReleaseYears is user statistics param for release years.
type UserStatisticsParamReleaseYears struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamStartYears is user statistics param for start years.
type UserStatisticsParamStartYears struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamGenres is user statistics param for genres.
type UserStatisticsParamGenres struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamTags is user statistics param for tags.
type UserStatisticsParamTags struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamCountries is user statistics param for countries.
type UserStatisticsParamCountries struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamVoiceActors is user statistics param for voice actors.
type UserStatisticsParamVoiceActors struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamStaff is user statistics param for staff.
type UserStatisticsParamStaff struct {
	Limit int
	Sort  []UserStatisticsSort
}

// UserStatisticsParamStudios is user statistics param for studios.
type UserStatisticsParamStudios struct {
	Limit int
	Sort  []UserStatisticsSort
}

// MediaListCollectionParam is media list collection param.
type MediaListCollectionParam struct {
	UserID                   int
	Username                 string
	Type                     MediaType
	Status                   MediaListStatus
	Notes                    string
	StartedAt                FuzzyDateInt
	CompletedAt              FuzzyDateInt
	ForceSingleCompletedList *bool
	Chunk                    int
	PerChunk                 int
	StatusIn                 []MediaListStatus
	StatusNotIn              []MediaListStatus
	StatusNot                MediaListStatus
	NotesLike                string
	StartedAtGreater         FuzzyDateInt
	StartedAtLesser          FuzzyDateInt
	StartedAtLike            string
	CompletedAtGreater       FuzzyDateInt
	CompletedAtLesser        FuzzyDateInt
	CompletedAtLike          string
	Sort                     []MediaListSort
}
