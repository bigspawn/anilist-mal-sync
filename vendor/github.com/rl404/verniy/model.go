package verniy

// Media is main anime & manga model.
type Media struct {
	ID                      int                       `json:"id"`
	IDMAL                   *int                      `json:"idMal"`
	Title                   *MediaTitle               `json:"title"`
	Type                    *MediaType                `json:"type"`
	Format                  *MediaFormat              `json:"format"`
	Status                  *MediaStatus              `json:"status"`
	Description             *string                   `json:"description"`
	StartDate               *FuzzyDate                `json:"startDate"`
	EndDate                 *FuzzyDate                `json:"endDate"`
	Season                  *MediaSeason              `json:"season"`
	SeasonYear              *int                      `json:"seasonYear"`
	SeasonInt               *int                      `json:"seasonInt"`
	Episodes                *int                      `json:"episodes"`
	Duration                *int                      `json:"duration"` // in minutes
	Chapters                *int                      `json:"chapters"`
	Volumes                 *int                      `json:"volumes"`
	CountryOfOrigin         *string                   `json:"countryOfOrigin"`
	IsLicensed              *bool                     `json:"isLicensed"`
	Source                  *MediaSource              `json:"source"`
	HashTag                 *string                   `json:"hashTag"`
	Trailer                 *MediaTrailer             `json:"trailer"`
	UpdatedAt               *int                      `json:"updatedAt"`
	CoverImage              *MediaCoverImage          `json:"coverImage"`
	BannerImage             *string                   `json:"bannerImage"`
	Genres                  []string                  `json:"genres"`
	Synonyms                []string                  `json:"synonyms"`
	AverageScore            *int                      `json:"averageScore"`
	MeanScore               *int                      `json:"meanScore"`
	Popularity              *int                      `json:"popularity"`
	IsLocked                *bool                     `json:"isLocked"`
	Trending                *int                      `json:"trending"`
	Favourites              *int                      `json:"favourites"`
	Tags                    []MediaTag                `json:"tags"`
	Relations               *MediaConnection          `json:"relations"`
	Characters              *CharacterConnection      `json:"characters"`
	Staff                   *StaffConnection          `json:"staff"`
	Studios                 *StudioConnection         `json:"studios"`
	IsFavourite             *bool                     `json:"isFavourite"`
	IsAdult                 *bool                     `json:"isAdult"`
	NextAiringEpisode       *AiringSchedule           `json:"nextAiringEpisode"`
	AiringSchedule          *AiringScheduleConnection `json:"airingSchedule"`
	Trends                  *MediaTrendConnection     `json:"trends"`
	ExternalLinks           []MediaExternalLink       `json:"externalLinks"`
	StreamingEpisodes       []MediaStreamingEpisode   `json:"streamingEpisodes"`
	Rankings                []MediaRank               `json:"rankings"`
	MediaListEntry          *MediaList                `json:"mediaListEntry"`
	Reviews                 *ReviewConnection         `json:"reviews"`
	Recommendations         *RecommendationConnection `json:"recommendations"`
	Stats                   *MediaStats               `json:"stats"`
	SiteURL                 *string                   `json:"siteUrl"`
	AutoCreateForumThread   *bool                     `json:"autoCreateForumThread"`
	IsRecommendationBlocked *bool                     `json:"isRecommendationBlocked"`
	ModNotes                *string                   `json:"modNotes"`
}

// MediaTitle is anime & manga titles.
type MediaTitle struct {
	Romaji        *string `json:"romaji"`
	English       *string `json:"english"`
	Native        *string `json:"native"`
	UserPreferred *string `json:"userPreferred"`
}

// FuzzyDateInt is 8 digit long date integer (YYYYMMDD).
type FuzzyDateInt int

// FuzzyDate is common date format.
type FuzzyDate struct {
	Year  *int `json:"year"`
	Month *int `json:"month"`
	Day   *int `json:"day"`
}

// MediaTrailer is anime & manga trailer data.
type MediaTrailer struct {
	ID        *string `json:"id"`
	Site      *string `json:"site"`
	Thumbnail *string `json:"thumbnail"`
}

// MediaCoverImage is anime & manga cover image model.
type MediaCoverImage struct {
	ExtraLarge *string `json:"extraLarge"`
	Large      *string `json:"large"`
	Medium     *string `json:"medium"`
	Color      *string `json:"color"`
}

// MediaTag is anime & manga tag model.
type MediaTag struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Description      *string `json:"description"`
	Category         *string `json:"category"`
	Rank             *int    `json:"rank"`
	IsGeneralSpoiler *bool   `json:"isGeneralSpoiler"`
	IsMediaSpoiler   *bool   `json:"isMediaSpoiler"`
	IsAdult          *bool   `json:"isAdult"`
}

// MediaConnection is anime & manga related model.
type MediaConnection struct {
	Edges    []MediaEdge `json:"edges"`
	Nodes    []Media     `json:"nodes"`
	PageInfo *PageInfo   `json:"pageInfo"`
}

// MediaEdge is anime & manga detail related model.
type MediaEdge struct {
	Node            *Media          `json:"node"`
	ID              *int            `json:"id"`
	RelationType    *MediaRelation  `json:"relationType"`
	IsMainStudio    bool            `json:"isMainStudio"`
	Characters      []Character     `json:"characters"`
	CharacterRole   *CharacterRole  `json:"characterRole"`
	CharacterName   *string         `json:"characterName"`
	RoleNotes       *string         `json:"roleNotes"`
	DubGroup        *string         `json:"dubGroup"`
	StaffRole       *string         `json:"staffRole"`
	VoiceActors     []Staff         `json:"voiceActors"`
	VoiceActorRoles []StaffRoleType `json:"voiceActorRoles"`
	FavouriteOrder  *int            `json:"favouriteOrder"`
}

// StaffRoleType is type of staff role.
type StaffRoleType struct {
	VoiceActor *Staff  `json:"voiceActor"`
	RoleNotes  *string `json:"roleNotes"`
	DubGroup   *string `json:"dubGroup"`
}

// PageInfo is common pagination model.
type PageInfo struct {
	Total       *int  `json:"total"`
	PerPage     *int  `json:"perPage"`
	CurrentPage *int  `json:"currentPage"`
	LastPage    *int  `json:"lastPage"`
	HasNextPage *bool `json:"hasNextPage"`
}

// CharacterConnection is character related model.
type CharacterConnection struct {
	Edges    []CharacterEdge `json:"edges"`
	Nodes    []Character     `json:"nodes"`
	PageInfo *PageInfo       `json:"pageInfo"`
}

// CharacterEdge is detail character related model.
type CharacterEdge struct {
	Node            *Character      `json:"node"`
	ID              *int            `json:"id"`
	Role            *CharacterRole  `json:"role"`
	Name            *string         `json:"name"`
	VoiceActors     []Staff         `json:"voiceActors"`
	VoiceActorRoles []StaffRoleType `json:"voiceActorRoles"`
	Media           []Media         `json:"media"`
	FavouriteOrder  *int            `json:"favouriteOrder"`
}

// Character is main character model.
type Character struct {
	ID                 int              `json:"id"`
	Name               *CharacterName   `json:"name"`
	Image              *CharacterImage  `json:"image"`
	Description        *string          `json:"description"`
	Gender             *string          `json:"gender"`
	DateOfBirth        *FuzzyDate       `json:"dateOfBirth"`
	Age                *string          `json:"age"`
	IsFavourite        *bool            `json:"isFavourite"`
	IsFavouriteBlocked *bool            `json:"isFavouriteBlocked"`
	SiteURL            *string          `json:"siteURL"`
	Media              *MediaConnection `json:"media"`
	Favourites         *int             `json:"favourites"`
	ModNotes           *string          `json:"modNotes"`
}

// CharacterName is character names.
type CharacterName struct {
	First              *string  `json:"first"`
	Middle             *string  `json:"middle"`
	Last               *string  `json:"last"`
	Full               *string  `json:"full"`
	Native             *string  `json:"native"`
	Alternative        []string `json:"alternative"`
	AlternativeSpoiler []string `json:"alternativeSpoiler"`
	UserPreferred      *string  `json:"userPreferred"`
}

// CharacterImage is character image.
type CharacterImage struct {
	Large  *string `json:"large"`
	Medium *string `json:"medium"`
}

// StaffConnection is staff related model.
type StaffConnection struct {
	Edges    []StaffEdge `json:"edges"`
	Nodes    []Staff     `json:"nodes"`
	PageInfo *PageInfo   `json:"pageInfo"`
}

// StaffEdge is detail staff related model.
type StaffEdge struct {
	Node           *Staff  `json:"node"`
	ID             *int    `json:"id"`
	Role           *string `json:"role"`
	FavouriteOrder *int    `json:"favouriteOrder"`
}

// Staff is main staff model.
type Staff struct {
	ID                 int                  `json:"id"`
	Name               *StaffName           `json:"name"`
	LanguageV2         *string              `json:"languageV2"`
	Image              *StaffImage          `json:"image"`
	Description        *string              `json:"description"`
	PrimaryOccupations []string             `json:"primaryOccupation"`
	Gender             *string              `json:"gender"`
	DateOfBirth        *FuzzyDate           `json:"dateOfBirth"`
	DateOfDeath        *FuzzyDate           `json:"dateOfDeath"`
	Age                *int                 `json:"age"`
	YearsActive        []int                `json:"yearsActive"`
	HomeTown           *string              `json:"homeTown"`
	IsFavourite        bool                 `json:"isFavourite"`
	IsFavouriteBlocked bool                 `json:"isFavouriteBlocked"`
	SiteURL            *string              `json:"siteURL"`
	StaffMedia         *MediaConnection     `json:"staffMedia"`
	Characters         *CharacterConnection `json:"characters"`
	CharacterMedia     *MediaConnection     `json:"characterMedia"`
	Staff              *Staff               `json:"staff"`
	Submitter          *User                `json:"submitter"`
	SubmissionStatus   *int                 `json:"submissionStatus"`
	SubmissionNotes    *string              `json:"submissionNotes"`
	Favourites         *int                 `json:"favourites"`
	ModNotes           *string              `json:"modNotes"`
}

// StaffName is staff names.
type StaffName struct {
	First         *string  `json:"first"`
	Middle        *string  `json:"middle"`
	Last          *string  `json:"last"`
	Full          *string  `json:"full"`
	Native        *string  `json:"native"`
	Alternative   []string `json:"alternative"`
	UserPreferred *string  `json:"userPreferred"`
}

// StaffImage is staff image.
type StaffImage struct {
	Large  *string `json:"large"`
	Medium *string `json:"medium"`
}

// StudioConnection is studio related model.
type StudioConnection struct {
	Edges    []StudioEdge `json:"edges"`
	Nodes    []Studio     `json:"nodes"`
	PageInfo *PageInfo    `json:"pageInfo"`
}

// StudioEdge is detail studio related model.
type StudioEdge struct {
	Node           *Studio `json:"node"`
	ID             *int    `json:"id"`
	IsMain         bool    `json:"isMain"`
	FavouriteOrder *int    `json:"favouriteOrder"`
}

// Studio is main studio model.
type Studio struct {
	ID                int              `json:"id"`
	Name              string           `json:"name"`
	IsAnimationStudio bool             `json:"isAnimationStudio"`
	Media             *MediaConnection `json:"media"`
	SiteURL           *string          `json:"siteURL"`
	IsFavourite       bool             `json:"isFavourite"`
	Favourites        *int             `json:"favourites"`
}

// AiringSchedule is anime airing schedule model.
type AiringSchedule struct {
	ID              int    `json:"id"`
	AiringAt        int    `json:"airingAt"`
	TimeUntilAiring int    `json:"timeUntilAiring"`
	Episode         int    `json:"episode"`
	MediaID         int    `json:"mediaId"`
	Media           *Media `json:"media"`
}

// AiringScheduleConnection is anime related airing schedule model.
type AiringScheduleConnection struct {
	Edges    []AiringScheduleEdge `json:"edges"`
	Nodes    []AiringSchedule     `json:"nodes"`
	PageInfo *PageInfo            `json:"pageInfo"`
}

// AiringScheduleEdge is airing schedule edge model.
type AiringScheduleEdge struct {
	Node *AiringSchedule `json:"node"`
	ID   *int            `json:"id"`
}

// MediaTrendConnection is media trend connection model.
type MediaTrendConnection struct {
	Edges    []MediaTrendEdge `json:"edges"`
	Nodes    []MediaTrend     `json:"nodes"`
	PageInfo *PageInfo        `json:"pageInfo"`
}

// MediaTrendEdge is media trend edge model.
type MediaTrendEdge struct {
	Node *MediaTrend `json:"node"`
}

// MediaTrend is media trend model.
type MediaTrend struct {
	MediaID      int    `json:"mediaId"`
	Date         int    `json:"date"`
	Trending     int    `json:"trending"`
	AverageScore *int   `json:"averageScore"`
	Popularity   *int   `json:"popularity"`
	InProgress   *int   `json:"inProgress"`
	Releasing    bool   `json:"releasing"`
	Episode      *int   `json:"episode"`
	Media        *Media `json:"media"`
}

// MediaExternalLink is anime & manga site links.
type MediaExternalLink struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Site string `json:"site"`
}

// MediaStreamingEpisode is anime streaming site links.
type MediaStreamingEpisode struct {
	Title     *string `json:"title"`
	Thumbnail *string `json:"thumbnail"`
	URL       *string `json:"url"`
	Site      *string `json:"site"`
}

// MediaRank is anime & manga ranking model.
type MediaRank struct {
	ID      int           `json:"id"`
	Rank    int           `json:"rank"`
	Type    MediaRankType `json:"type"`
	Format  MediaFormat   `json:"format"`
	Year    *int          `json:"year"`
	Season  *MediaSeason  `json:"season"`
	AllTime *bool         `json:"allTime"`
	Context string        `json:"context"`
}

// MediaList is user's anime & manga model.
type MediaList struct {
	ID                    int              `json:"id"`
	UserID                int              `json:"userId"`
	MediaID               int              `json:"mediaId"`
	Status                *MediaListStatus `json:"status"`
	Score                 *float64         `json:"score"`
	Progress              *int             `json:"progress"`
	ProgressVolumes       *int             `json:"progressVolumes"`
	Repeat                *int             `json:"repeat"`
	Priority              *int             `json:"priority"`
	Private               *bool            `json:"private"`
	Notes                 *string          `json:"notes"`
	HiddenFromStatusLists *bool            `json:"hiddenFromStatusLists"`
	CustomLists           *string          `json:"customLists"`    // json
	AdvancedScores        *string          `json:"advancedScores"` // json
	StartedAt             *FuzzyDate       `json:"startedAt"`
	CompletedAt           *FuzzyDate       `json:"completedAt"`
	UpdatedAt             *int             `json:"updatedAt"`
	CreatedAt             *int             `json:"createdAt"`
	Media                 *Media           `json:"media"`
	User                  *User            `json:"user"`
}

// User is main user model.
type User struct {
	ID                      int                 `json:"id"`
	Name                    string              `json:"name"`
	About                   *string             `json:"about"`
	Avatar                  *UserAvatar         `json:"avatar"`
	BannerImage             *string             `json:"bannerImage"`
	IsFollowing             *bool               `json:"isFollowing"`
	IsFollower              *bool               `json:"isFollower"`
	IsBlocked               *bool               `json:"isBlocked"`
	Bans                    *string             `json:"bans"` // json
	Options                 *UserOptions        `json:"options"`
	MediaListOptions        *MediaListOptions   `json:"mediaListOptions"`
	Favourites              *Favourites         `json:"favourites"`
	Statistics              *UserStatisticTypes `json:"statistics"`
	UnreadNotificationCount *int                `json:"unreadNotificationCount"`
	SiteURL                 *string             `json:"siteUrl"`
	DonatorTier             *int                `json:"donatorTier"`
	DonatorBadge            *string             `json:"donatorBadge"`
	ModeratorRoles          []ModRole           `json:"moderatorRoles"`
	CreatedAt               *int                `json:"createdAt"`
	UpdatedAt               *int                `json:"updatedAt"`
}

// UserAvatar is user avatar image.
type UserAvatar struct {
	Large  *string `json:"large"`
	Medium *string `json:"medium"`
}

// UserOptions is user option model.
type UserOptions struct {
	TitleLanguage       *UserTitleLanguage     `json:"titleLanguage"`
	DisplayAdultContent *bool                  `json:"displayAdultContent"`
	AiringNotifications *bool                  `json:"airingNotification"`
	ProfileColor        *string                `json:"profileColor"`
	NotificationOptions []NotificationOption   `json:"notificationOptions"`
	Timezone            *string                `json:"timezone"`
	ActivityMergeTime   *int                   `json:"activityMergeTime"`
	StaffNameLanguage   *UserStaffNameLanguage `json:"staffNameLanguage"`
}

// NotificationOption is user notification option model.
type NotificationOption struct {
	Type    *NotificationType `json:"type"`
	Enabled *bool             `json:"enabled"`
}

// MediaListOptions is user anime & manga option model.
type MediaListOptions struct {
	ScoreFormat *ScoreFormat          `json:"scoreFormat"`
	RowOrder    *string               `json:"rowOrder"`
	AnimeList   *MediaListTypeOptions `json:"animeList"`
	MangaList   *MediaListTypeOptions `json:"mangaList"`
}

// MediaListTypeOptions is media list type options model.
type MediaListTypeOptions struct {
	SectionOrder                  []string `json:"sectionOrder"`
	SplitCompletedSectionByFormat *bool    `json:"splitCompletedSectionByFormat"`
	CustomLists                   []string `json:"customLists"`
	AdvancedScoring               []string `json:"advancedScoring"`
	AdvancedScoringEnabled        *bool    `json:"advancedScoringEnabled"`
}

// Favourites is user's favourite model.
type Favourites struct {
	Anime      *MediaConnection     `json:"anime"`
	Manga      *MediaConnection     `json:"manga"`
	Characters *CharacterConnection `json:"characters"`
	Staff      *StaffConnection     `json:"staff"`
	Studios    *StudioConnection    `json:"studios"`
}

// UserStatisticTypes is user statistic data.
type UserStatisticTypes struct {
	Anime *UserStatistics `json:"anime"`
	Manga *UserStatistics `json:"manga"`
}

// UserStatistics is detail user statistic data.
type UserStatistics struct {
	Count             int                        `json:"count"`
	MeanScore         float64                    `json:"meanScore"`
	StandardDeviation float64                    `json:"standardDeviation"`
	MinutesWatched    int                        `json:"minutesWatched"`
	EpisodesWatched   int                        `json:"episodesWatched"`
	ChaptersRead      int                        `json:"chaptersRead"`
	VolumesRead       int                        `json:"volumesRead"`
	Formats           []UserFormatStatistic      `json:"formats"`
	Statuses          []UserStatusStatistic      `json:"statuses"`
	Scores            []UserScoreStatistic       `json:"scores"`
	Lengths           []UserLengthStatistic      `json:"lengths"`
	ReleaseYears      []UserReleaseYearStatistic `json:"releaseYears"`
	StartYears        []UserStartYearStatistic   `json:"startYears"`
	Genres            []UserGenreStatistic       `json:"genres"`
	Tags              []UserTagStatistic         `json:"tags"`
	Countries         []UserCountryStatistic     `json:"countries"`
	VoiceActors       []UserVoiceActorStatistic  `json:"voiceActors"`
	Staff             []UserStaffStatistic       `json:"staff"`
	Studios           []UserStudioStatistic      `json:"studios"`
}

// UserFormatStatistic is user format stats.
type UserFormatStatistic struct {
	Count          int          `json:"count"`
	MeanScore      float64      `json:"meanScore"`
	MinutesWatched int          `json:"minutesWatched"`
	ChapterRead    int          `json:"chapterRead"`
	MediaIDs       []int        `json:"mediaIDs"`
	Format         *MediaFormat `json:"format"`
}

// UserStatusStatistic is user status stats.
type UserStatusStatistic struct {
	Count          int              `json:"count"`
	MeanScore      float64          `json:"meanScore"`
	MinutesWatched int              `json:"minutesWatched"`
	ChaptersRead   int              `json:"chaptersRead"`
	MediaIDs       []int            `json:"mediaIDs"`
	Status         *MediaListStatus `json:"status"`
}

// UserScoreStatistic is user score stats.
type UserScoreStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	Score          *int    `json:"score"`
}

// UserLengthStatistic is user watch/read duration stats.
type UserLengthStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	Length         *string `json:"length"`
}

// UserReleaseYearStatistic is user anime & manga release year stats.
type UserReleaseYearStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	ReleaseYear    *int    `json:"releaseYear"`
}

// UserStartYearStatistic is user start year stats.
type UserStartYearStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	StartYear      *int    `json:"startYear"`
}

// UserGenreStatistic is user genre stats.
type UserGenreStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	Genre          *string `json:"genre"`
}

// UserTagStatistic is user tag stats.
type UserTagStatistic struct {
	Count          int       `json:"count"`
	MeanScore      float64   `json:"meanScore"`
	MinutesWatched int       `json:"minutesWatched"`
	ChaptersRead   int       `json:"chaptersRead"`
	MediaIDs       []int     `json:"mediaIDs"`
	Tag            *MediaTag `json:"tag"`
}

// UserCountryStatistic is user anime & manga country stats.
type UserCountryStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	Country        *string `json:"country"`
}

// UserVoiceActorStatistic is user voice actor stats.
type UserVoiceActorStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	VoiceActor     *Staff  `json:"voiceActor"`
	CharacterIDs   []int   `json:"characterIDs"`
}

// UserStaffStatistic is user staff stats.
type UserStaffStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	Staff          *Staff  `json:"staff"`
}

// UserStudioStatistic is user studio stats.
type UserStudioStatistic struct {
	Count          int     `json:"count"`
	MeanScore      float64 `json:"meanScore"`
	MinutesWatched int     `json:"minutesWatched"`
	ChaptersRead   int     `json:"chaptersRead"`
	MediaIDs       []int   `json:"mediaIDs"`
	Studio         *Studio `json:"studio"`
}

// ReviewConnection is anime & manga related review.
type ReviewConnection struct {
	Edges    []ReviewEdge `json:"edges"`
	Nodes    []Review     `json:"nodes"`
	PageInfo *PageInfo    `json:"pageInfo"`
}

// ReviewEdge is detail anime & manga related review.
type ReviewEdge struct {
	Node *Review `json:"node"`
}

// Review is main review model.
type Review struct {
	ID           int           `json:"id"`
	UserID       int           `json:"userId"`
	MediaID      int           `json:"mediaId"`
	MediaType    *MediaType    `json:"mediaType"`
	Summary      *string       `json:"summary"`
	Body         *string       `json:"body"`
	Rating       *int          `json:"rating"`
	RatingAmount *int          `json:"ratingAmount"`
	UserRating   *ReviewRating `json:"userRating"`
	Score        *int          `json:"score"`
	Private      *bool         `json:"private"`
	SiteURL      *string       `json:"siteUrl"`
	CreatedAt    int           `json:"createdAt"`
	UpdatedAt    int           `json:"updatedAt"`
	User         *User         `json:"user"`
	Media        *Media        `json:"media"`
}

// RecommendationConnection is anime & manga related recommendation.
type RecommendationConnection struct {
	Edges    []RecommendationEdge `json:"edges"`
	Nodes    []Recommendation     `json:"nodes"`
	PageInfo *PageInfo            `json:"pageInfo"`
}

// RecommendationEdge is detail anime & manga related recommendation.
type RecommendationEdge struct {
	Node *Recommendation `json:"node"`
}

// Recommendation is main recommendation model.
type Recommendation struct {
	ID                  int                   `json:"id"`
	Rating              *int                  `json:"rating"`
	UserRating          *RecommendationRating `json:"userRating"`
	Media               *Media                `json:"media"`
	MediaRecommendation *Media                `json:"mediaRecommendation"`
	User                *User                 `json:"user"`
}

// MediaStats is anime & manga stats.
type MediaStats struct {
	ScoreDistribution  []ScoreDistribution  `json:"scoreDistribution"`
	StatusDistribution []StatusDistribution `json:"statusDistribution"`
}

// ScoreDistribution is detail anime & manga score stats.
type ScoreDistribution struct {
	Score  *int `json:"score"`
	Amount *int `json:"amount"`
}

// StatusDistribution is detail anime & manga status stats.
type StatusDistribution struct {
	Status *MediaListStatus `json:"status"`
	Amount *int             `json:"amount"`
}

// MediaListCollection is collection of media list.
type MediaListCollection struct {
	Lists        []MediaListGroup `json:"lists"`
	User         *User            `json:"user"`
	HasNextChunk *bool            `json:"hasNextChunk"`
}

// MediaListGroup is group of media list.
type MediaListGroup struct {
	Entries              []MediaList      `json:"entries"`
	Name                 *string          `json:"name"`
	IsCustomList         *bool            `json:"isCustomList"`
	IsSplitCompletedList *bool            `json:"isSplitCompletedList"`
	Status               *MediaListStatus `json:"status"`
}
