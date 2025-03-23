package verniy

// MediaType is main media type.
type MediaType string

// Options for MediaType.
const (
	MediaTypeAnime MediaType = "ANIME"
	MediaTypeManga MediaType = "MANGA"
)

// MediaFormat is anime & manga format.
type MediaFormat string

// Options for MediaFormat.
const (
	MediaFormatTv      MediaFormat = "TV"
	MediaFormatTvShort MediaFormat = "TV_SHORT"
	MediaFormatMovie   MediaFormat = "MOVIE"
	MediaFormatSpecial MediaFormat = "SPECIAL"
	MediaFormatOVA     MediaFormat = "OVA"
	MediaFormatONA     MediaFormat = "ONA"
	MediaFormatMusic   MediaFormat = "MUSIC"
	MediaFormatManga   MediaFormat = "MANGA"
	MediaFormatNovel   MediaFormat = "NOVEL"
	MediaFormatOneShot MediaFormat = "ONE_SHOT"
)

// MediaStatus is anime & manga status.
type MediaStatus string

// Options for MediaStatus.
const (
	MediaStatusFinished       MediaStatus = "FINISHED"
	MediaStatusReleasing      MediaStatus = "RELEASING"
	MediaStatusNotYetReleased MediaStatus = "NOT_YET_RELEASED"
	MediaStatusCancelled      MediaStatus = "CANCELLED"
	MediaStatusHiatus         MediaStatus = "HIATUS"
)

// MediaSeason is anime season.
type MediaSeason string

// Options for MediaSeason.
const (
	MediaSeasonWinter MediaSeason = "WINTER" // 12-2
	MediaSeasonSpring MediaSeason = "SPRING" // 3-5
	MediaSeasonSummer MediaSeason = "SUMMER" // 6-8
	MediaSeasonFall   MediaSeason = "FALL"   // 9-11
)

// MediaSource is anime & manga source.
type MediaSource string

// Options for MediaSource.
const (
	MediaSourceOriginal    MediaSource = "ORIGINAL"
	MediaSourceManga       MediaSource = "MANGA"
	MediaSourceLightNovel  MediaSource = "LIGHT_NOVEL"
	MediaSourceVisualNovel MediaSource = "VISUAL_NOVEL"
	MediaSourceVideoGame   MediaSource = "VIDEO_GAME"
	MediaSourceOther       MediaSource = "OTHER"
	MediaSourceNovel       MediaSource = "NOVEL"
	MediaSourceDoujinshi   MediaSource = "DOUJINSHI"
	MediaSourceAnime       MediaSource = "ANIME"
)

// MediaRankType is type of ranking.
type MediaRankType string

// Options for MediaRankType
const (
	MediaRankTypeRated   MediaRankType = "RATED"
	MediaRankTypePopular MediaRankType = "POPULAR"
)

// MediaListStatus is user's anime & manga status.
type MediaListStatus string

// Options for MediaListStatus.
const (
	MediaListStatusCurrent   MediaListStatus = "CURRENT"
	MediaListStatusPlanning  MediaListStatus = "PLANNING"
	MediaListStatusCompleted MediaListStatus = "COMPLETED"
	MediaListStatusDropped   MediaListStatus = "DROPPED"
	MediaListStatusPaused    MediaListStatus = "PAUSED"
	MediaListStatusRepeating MediaListStatus = "REPEATING"
)

// MediaRelation is anime & manga relation.
type MediaRelation string

// Options for MediaRelation.
const (
	MediaRelationAdaptation  MediaRelation = "ADAPTATION"
	MediaRelationPrequel     MediaRelation = "PREQUEL"
	MediaRelationSequel      MediaRelation = "SEQUEL"
	MediaRelationParent      MediaRelation = "PARENT"
	MediaRelationSideStory   MediaRelation = "SIDE_STORY"
	MediaRelationCharacter   MediaRelation = "CHARACTER"
	MediaRelationSummary     MediaRelation = "SUMMARY"
	MediaRelationAlternative MediaRelation = "ALTERNATIVE"
	MediaRelationSpinOff     MediaRelation = "SPIN_OFF"
	MediaRelationOther       MediaRelation = "OTHER"
	MediaRelationSource      MediaRelation = "SOURCE"
	MediaRelationCompilation MediaRelation = "COMPILATION"
	MediaRelationContains    MediaRelation = "CONTAINS"
)

// CharacterRole is type of character role.
type CharacterRole string

// Options for CharacterRole.
const (
	CharacterRoleMain       CharacterRole = "MAIN"
	CharacterRoleSupporting CharacterRole = "SUPPORTING"
	CharacterRoleBackground CharacterRole = "BACKGROUND"
)

// UserTitleLanguage is default user anime & manga title language.
type UserTitleLanguage string

// Options for UserTitleLanguage.
const (
	UserTitleLanguageRomaji          UserTitleLanguage = "ROMAJI"
	UserTitleLanguageEnglish         UserTitleLanguage = "ENGLISH"
	UserTitleLanguageNative          UserTitleLanguage = "NATIVE"
	UserTitleLanguageRomajiStylised  UserTitleLanguage = "ROMAJI_STYLISED"
	UserTitleLanguageEnglishStylised UserTitleLanguage = "ENGLISH_STYLISED"
	UserTitleLanguageNativeStylised  UserTitleLanguage = "NATIVE_STYLISED"
)

// NotificationType is user notification type.
type NotificationType string

// Options for NotificationType.
const (
	NotificationTypeActivityMessage         NotificationType = "ACTIVITY_MESSAGE"
	NotificationTypeActivityReply           NotificationType = "ACTIVITY_REPLY"
	NotificationTypeActivityFollowing       NotificationType = "FOLLOWING"
	NotificationTypeActivityMention         NotificationType = "ACTIVITY_MENTION"
	NotificationTypeThreadCommentMention    NotificationType = "THREAD_COMMENT_MENTION"
	NotificationTypeThreadSubscribed        NotificationType = "THREAD_SUBSCRIBED"
	NotificationTypeThreadCommentReply      NotificationType = "THREAD_COMMENT_REPLY"
	NotificationTypeAiring                  NotificationType = "AIRING"
	NotificationTypeActivityLike            NotificationType = "ACTIVITY_LIKE"
	NotificationTypeActivityReplyLike       NotificationType = "ACTIVITY_REPLY_LIKE"
	NotificationTypeThreadLike              NotificationType = "THREAD_LIKE"
	NotificationTypeThreadCommentLike       NotificationType = "THREAD_COMMENT_LIKE"
	NotificationTypeActivityReplySubscribed NotificationType = "ACTIVITY_REPLY_SUBSCRIBED"
	NotificationTypeRelatedMediaAddition    NotificationType = "RELATED_MEDIA_ADDITION"
)

// UserStaffNameLanguage is default user staff naming language.
type UserStaffNameLanguage string

// Options for UserStaffNameLanguage
const (
	UserStaffNameLanguageRomajiWestern UserStaffNameLanguage = "ROMAJI_WESTERN"
	UserStaffNameLanguageRomaji        UserStaffNameLanguage = "ROMAJI"
	UserStaffNameLanguageNative        UserStaffNameLanguage = "NATIVE"
)

// ReviewRating is type of review rating.
type ReviewRating string

// Options for ReviewRating.
const (
	ReviewRatingNoVote   ReviewRating = "NO_VOTE"
	ReviewRatingUpVote   ReviewRating = "UP_VOTE"
	ReviewRatingDownVote ReviewRating = "DOWN_VOTE"
)

// RecommendationRating is type of recommendation rating.
type RecommendationRating string

// Options for RecommendationRating.
const (
	RecommendationRatingNoRating RecommendationRating = "NO_RATING"
	RecommendationRatingRateUp   RecommendationRating = "RATE_UP"
	RecommendationRatingRateDown RecommendationRating = "RATE_DOWN"
)

// ScoreFormat is scoring format.
type ScoreFormat string

// Options for ScoreFormat.
const (
	ScoreFormatPoint100        ScoreFormat = "POINT_100"
	ScoreFormatPoint100Decimal ScoreFormat = "POINT_10_DECIMAL"
	ScoreFormatPoint10         ScoreFormat = "POINT_10"
	ScoreFormatPoint5          ScoreFormat = "POINT_5"
	ScoreFormatPoint3          ScoreFormat = "POINT_3"
)

// ModRole is mod role.
type ModRole string

// Options for ModRole.
const (
	ModRoleAdmin            ModRole = "ADMIN"
	ModRoleLeadDeveloper    ModRole = "LEAD_DEVELOPER"
	ModRoleDeveloper        ModRole = "DEVELOPER"
	ModRoleLeadCommunity    ModRole = "LEAD_COMMUNITY"
	ModRoleCommunity        ModRole = "COMMUNITY"
	ModRoleDiscordCommunity ModRole = "DISCORD_COMMUNITY"
	ModRoleLeadAnimeData    ModRole = "LEAD_ANIME_DATA"
	ModRoleAnimeData        ModRole = "ANIME_DATA"
	ModRoleLeadMangaData    ModRole = "LEAD_MANGA_DATA"
	ModRoleMangaData        ModRole = "MANGA_DATA"
	ModRoleLeadSocialMedia  ModRole = "LEAD_SOCIAL_MEDIA"
	ModRoleSocialMedia      ModRole = "SOCIAL_MEDIA"
	ModRoleRetired          ModRole = "RETIRED"
)

// CharacterSort is sorting option for character list.
type CharacterSort string

// Options for CharacterSort.
const (
	CharacterSortID             CharacterSort = "ID"
	CharacterSortIDDesc         CharacterSort = "ID_DESC"
	CharacterSortRole           CharacterSort = "ROLE"
	CharacterSortRoleDesc       CharacterSort = "ROLE_DESC"
	CharacterSortSearchMatch    CharacterSort = "SEARCH_MATCH"
	CharacterSortFavourites     CharacterSort = "FAVOURITES"
	CharacterSortFavouritesDesc CharacterSort = "FAVOURITES_DESC"
	CharacterSortRelevance      CharacterSort = "RELEVANCE"
)

// StaffSort is sorting option for staff list.
type StaffSort string

// Options for StaffSort.
const (
	StaffSortID             StaffSort = "ID"
	StaffSortIDDesc         StaffSort = "ID_DESC"
	StaffSortRole           StaffSort = "ROLE"
	StaffSortRoleDesc       StaffSort = "ROLE_DESC"
	StaffSortLanguage       StaffSort = "LANGUAGE"
	StaffSortLanguageDesc   StaffSort = "LANGUAGE_DESC"
	StaffSortSearchMatch    StaffSort = "SEARCH_MATCH"
	StaffSortFavourites     StaffSort = "FAVOURITES"
	StaffSortFavouritesDesc StaffSort = "FAVOURITES_DESC"
	StaffSortRelevance      StaffSort = "RELEVANCE"
)

// MediaSort is sorting option for anime & manga list.
type MediaSort string

// Options for MediaSort.
const (
	MediaSortID               MediaSort = "ID"
	MediaSortIDDesc           MediaSort = "ID_DESC"
	MediaSortTitleRomaji      MediaSort = "TITLE_ROMAJI"
	MediaSortTitleRomajiDesc  MediaSort = "TITLE_ROMAJI_DESC"
	MediaSortTitleEnglish     MediaSort = "TITLE_ENGLISH"
	MediaSortTitleEnglishDesc MediaSort = "TITLE_ENGLISH_DESC"
	MediaSortTitleNative      MediaSort = "TITLE_NATIVE"
	MediaSortTitleNativeDesc  MediaSort = "TITLE_NATIVE_DESC"
	MediaSortType             MediaSort = "TYPE"
	MediaSortTypeDesc         MediaSort = "TYPE_DESC"
	MediaSortFormat           MediaSort = "FORMAT"
	MediaSortFormatDesc       MediaSort = "FORMAT_DESC"
	MediaSortStartDate        MediaSort = "START_DATE"
	MediaSortStartDateDesc    MediaSort = "START_DATE_DESC"
	MediaSortEndDate          MediaSort = "END_DATE"
	MediaSortEndDateDesc      MediaSort = "END_DATE_DESC"
	MediaSortScore            MediaSort = "SCORE"
	MediaSortScoreDesc        MediaSort = "SCORE_DESC"
	MediaSortPopularity       MediaSort = "POPULARITY"
	MediaSortPopularityDesc   MediaSort = "POPULARITY_DESC"
	MediaSortTrending         MediaSort = "TRENDING"
	MediaSortTrendingDesc     MediaSort = "TRENDING_DESC"
	MediaSortEpisodes         MediaSort = "EPISODES"
	MediaSortEpisodesDesc     MediaSort = "EPISODES_DESC"
	MediaSortDuration         MediaSort = "DURATION"
	MediaSortDurationDesc     MediaSort = "DURATION_DESC"
	MediaSortStatus           MediaSort = "STATUS"
	MediaSortStatusDesc       MediaSort = "STATUS_DESC"
	MediaSortChapters         MediaSort = "CHAPTERS"
	MediaSortChaptersDesc     MediaSort = "CHAPTERS_DESC"
	MediaSortVolumes          MediaSort = "VOLUMES"
	MediaSortVolumesDesc      MediaSort = "VOLUMES_DESC"
	MediaSortUpdatedAt        MediaSort = "UPDATED_AT"
	MediaSortUpdatedAtDesc    MediaSort = "UPDATED_AT_DESC"
	MediaSortSearchMatch      MediaSort = "SEARCH_MATCH"
	MediaSortFavourites       MediaSort = "FAVOURITES"
	MediaSortFavouritesDesc   MediaSort = "FAVOURITES_DESC"
)

// StaffLanguage is staff language.
type StaffLanguage string

// Options for StaffLanguage.
const (
	StaffLanguageJapanese   StaffLanguage = "JAPANESE"
	StaffLanguageEnglish    StaffLanguage = "ENGLISH"
	StaffLanguageKorean     StaffLanguage = "KOREAN"
	StaffLanguageItalian    StaffLanguage = "ITALIAN"
	StaffLanguageSpanish    StaffLanguage = "SPANISH"
	StaffLanguagePortuguese StaffLanguage = "PORTUGUESE"
	StaffLanguageFrench     StaffLanguage = "FRENCH"
	StaffLanguageGerman     StaffLanguage = "GERMAN"
	StaffLanguageHebrew     StaffLanguage = "HEBREW"
	StaffLanguageHungarian  StaffLanguage = "HUNGARIAN"
)

// StudioSort is sorting option for studio list.
type StudioSort string

// Options for StudioSort.
const (
	StudioSortID             StudioSort = "ID"
	StudioSortIDDesc         StudioSort = "ID_DESC"
	StudioSortName           StudioSort = "NAME"
	StudioSortNameDesc       StudioSort = "NAME_DESC"
	StudioSortSearchMatch    StudioSort = "SEARCH_MATCH"
	StudioSortFavourites     StudioSort = "FAVOURITES"
	StudioSortFavouritesDesc StudioSort = "FAVOURITES_DESC"
)

// MediaTrendSort is sorting option for media trend list.
type MediaTrendSort string

// Options for MediaTrendSort.
const (
	MediaTrendSortID             MediaTrendSort = "ID"
	MediaTrendSortIDDesc         MediaTrendSort = "ID_DESC"
	MediaTrendSortMediaID        MediaTrendSort = "MEDIA_ID"
	MediaTrendSortMediaIDDesc    MediaTrendSort = "MEDIA_ID_DESC"
	MediaTrendSortDate           MediaTrendSort = "DATE"
	MediaTrendSortDateDesc       MediaTrendSort = "DATE_DESC"
	MediaTrendSortScore          MediaTrendSort = "SCORE"
	MediaTrendSortScoreDesc      MediaTrendSort = "SCORE_DESC"
	MediaTrendSortPopularity     MediaTrendSort = "POPULARITY"
	MediaTrendSortPopularityDesc MediaTrendSort = "POPULARITY_DESC"
	MediaTrendSortTrending       MediaTrendSort = "TRENDING"
	MediaTrendSortTrendingDesc   MediaTrendSort = "TRENDING_DESC"
	MediaTrendSortEpisode        MediaTrendSort = "EPISODE"
	MediaTrendSortEpisodeDesc    MediaTrendSort = "EPISODE_DESC"
)

// ReviewSort is sorting option for review list.
type ReviewSort string

// Options for ReviewSort.
const (
	ReviewSortID            ReviewSort = "ID"
	ReviewSortIDDesc        ReviewSort = "ID_DESC"
	ReviewSortScore         ReviewSort = "SCORE"
	ReviewSortScoreDesc     ReviewSort = "SCORE_DESC"
	ReviewSortRating        ReviewSort = "RATING"
	ReviewSortRatingDesc    ReviewSort = "RATING_DESC"
	ReviewSortCreatedAt     ReviewSort = "CREATED_AT"
	ReviewSortCreatedAtDesc ReviewSort = "CREATED_AT_DESC"
	ReviewSortUpdatedAt     ReviewSort = "UPDATED_AT"
	ReviewSortUpdatedAtDesc ReviewSort = "UPDATED_AT_DESC"
)

// RecommendationSort is sorting option for recommendation list.
type RecommendationSort string

// Options for RecommendationSort.
const (
	RecommendationSortID         RecommendationSort = "ID"
	RecommendationSortIDDesc     RecommendationSort = "ID_DESC"
	RecommendationSortRating     RecommendationSort = "RATING"
	RecommendationSortRatingDesc RecommendationSort = "RATING_DESC"
)

// UserStatisticsSort is sorting option for user statistics list.
type UserStatisticsSort string

// Options for UserStatisticsSort.
const (
	UserStatisticsSortID            UserStatisticsSort = "ID"
	UserStatisticsSortIDDesc        UserStatisticsSort = "ID_DESC"
	UserStatisticsSortCount         UserStatisticsSort = "COUNT"
	UserStatisticsSortCountDesc     UserStatisticsSort = "COUNT_DESC"
	UserStatisticsSortProgress      UserStatisticsSort = "PROGRESS"
	UserStatisticsSortProgessDesc   UserStatisticsSort = "PROGRESS_DESC"
	UserStatisticsSortMeanScore     UserStatisticsSort = "MEAN_SCORE"
	UserStatisticsSortMeanScoreDesc UserStatisticsSort = "MEAN_SCORE_DESC"
)

// MediaListSort is sorting option for media list.
type MediaListSort string

// Options for MediaListSort.
const (
	MediaListSortMediaID               MediaListSort = "MEDIA_ID"
	MediaListSortMediaIDDesc           MediaListSort = "MEDIA_ID_DESC"
	MediaListSortScore                 MediaListSort = "SCORE"
	MediaListSortScoreDesc             MediaListSort = "SCORE_DESC"
	MediaListSortStatus                MediaListSort = "STATUS"
	MediaListSortStatusDesc            MediaListSort = "STATUS_DESC"
	MediaListSortProgress              MediaListSort = "PROGRESS"
	MediaListSortProgressDesc          MediaListSort = "PROGRESS_DESC"
	MediaListSortProgressVolumes       MediaListSort = "PROGRESS_VOLUMES"
	MediaListSortProgressVolumesDesc   MediaListSort = "PROGRESS_VOLUMES_DESC"
	MediaListSortRepeat                MediaListSort = "REPEAT"
	MediaListSortRepeatDesc            MediaListSort = "REPEAT_DESC"
	MediaListSortPriority              MediaListSort = "PRIORITY"
	MediaListSortPriorityDesc          MediaListSort = "PRIORITY_DESC"
	MediaListSortStartedOn             MediaListSort = "STARTED_ON"
	MediaListSortStartedOnDesc         MediaListSort = "STARTED_ON_DESC"
	MediaListSortFinishedOn            MediaListSort = "FINISHED_ON"
	MediaListSortFinishedOnDesc        MediaListSort = "FINISHED_ON_DESC"
	MediaListSortAddedTime             MediaListSort = "ADDED_TIME"
	MediaListSortAddedTimeDesc         MediaListSort = "ADDED_TIME_DESC"
	MediaListSortUpdatedTime           MediaListSort = "UPDATED_TIME"
	MediaListSortUpdatedTimeDesc       MediaListSort = "UPDATED_TIME_DESC"
	MediaListSortMediaTitleRomaji      MediaListSort = "MEDIA_TITLE_ROMAJI"
	MediaListSortMediaTitleRomajiDesc  MediaListSort = "MEDIA_TITLE_ROMAJI_DESC"
	MediaListSortMediaTitleEnglish     MediaListSort = "MEDIA_TITLE_ENGLISH"
	MediaListSortMediaTitleEnglishDesc MediaListSort = "MEDIA_TITLE_ENGLISH_DESC"
	MediaListSortMediaTitleNative      MediaListSort = "MEDIA_TITLE_NATIVE"
	MediaListSortMediaTitleNativeDesc  MediaListSort = "MEDIA_TITLE_NATIVE_DESC"
	MediaListSortMediaPopularity       MediaListSort = "MEDIA_POPULARITY"
	MediaListSortMediaPopularityDesc   MediaListSort = "MEDIA_POPULARITY_DESC"
)
