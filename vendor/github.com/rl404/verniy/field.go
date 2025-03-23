package verniy

// QueryParam is field params when request to graphql api.
type QueryParam map[string]interface{}

// MediaField is fields for media.
type MediaField string

// Fields for MediaField.
const (
	MediaFieldID                      MediaField = "id"
	MediaFieldIDMAL                   MediaField = "idMal"
	MediaFieldType                    MediaField = "type"
	MediaFieldFormat                  MediaField = "format"
	MediaFieldStatus                  MediaField = "status"
	MediaFieldStatusV2                MediaField = "status(version:2)"
	MediaFieldDescription             MediaField = "description"
	MediaFieldDescriptionHTML         MediaField = "description(asHtml:true)"
	MediaFieldStartDate               MediaField = "startDate{year month day}"
	MediaFieldEndDate                 MediaField = "endDate{year month day}"
	MediaFieldSeason                  MediaField = "season"
	MediaFieldSeasonYear              MediaField = "seasonYear"
	MediaFieldSeasonInt               MediaField = "seasonInt"
	MediaFieldEpisodes                MediaField = "episodes"
	MediaFieldDuration                MediaField = "duration"
	MediaFieldChapters                MediaField = "chapters"
	MediaFieldVolumes                 MediaField = "volumes"
	MediaFieldCountryOfOrigin         MediaField = "countryOfOrigin" // ISO 3166-1 alpha-2
	MediaFieldIsLicensed              MediaField = "isLicensed"
	MediaFieldSource                  MediaField = "source"
	MediaFieldSourceV2                MediaField = "source(version:2)"
	MediaFieldHashtag                 MediaField = "hashtag"
	MediaFieldUpdatedAt               MediaField = "updatedAt"
	MediaFieldBannerImage             MediaField = "bannerImage"
	MediaFieldGenres                  MediaField = "genres"
	MediaFieldSynonyms                MediaField = "synonyms"
	MediaFieldAverageScore            MediaField = "averageScore"
	MediaFieldMeanScore               MediaField = "meanScore"
	MediaFieldPopularity              MediaField = "popularity"
	MediaFieldIsLocked                MediaField = "isLocked"
	MediaFieldTrending                MediaField = "trending"
	MediaFieldFavourites              MediaField = "favourites"
	MediaFieldIsFavourite             MediaField = "isFavourite"
	MediaFieldIsAdult                 MediaField = "isAdult"
	MediaFieldSiteURL                 MediaField = "siteUrl"
	MediaFieldAutoCreateForumThread   MediaField = "autoCreateForumThread"
	MediaFieldIsRecommendationBlocked MediaField = "isRecommendationBlocked"
	MediaFieldModNotes                MediaField = "modNotes"
)

// MediaTitleField is fields for media title.
type MediaTitleField string

// Fields for MediaTitleField.
const (
	MediaTitleFieldRomaji        MediaTitleField = "romaji"
	MediaTitleFieldEnglish       MediaTitleField = "english"
	MediaTitleFieldNative        MediaTitleField = "native"
	MediaTitleFieldUserPreferred MediaTitleField = "userPreferred"
)

// MediaFieldTitle to generate media title query.
func MediaFieldTitle(field MediaTitleField, fields ...MediaTitleField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("title", nil, str...))
}

// MediaTrailerField is fields for media trailer.
type MediaTrailerField string

// Fields for MediaTrailerField.
const (
	MediaTrailerFieldID        MediaTrailerField = "id"
	MediaTrailerFieldSite      MediaTrailerField = "site"
	MediaTrailerFieldThumbnail MediaTrailerField = "thumbnail"
)

// MediaFieldTrailer to generate media trailer query.
func MediaFieldTrailer(field MediaTrailerField, fields ...MediaTrailerField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("trailer", nil, str...))
}

// MediaCoverImageField is fields for media cover image.
type MediaCoverImageField string

// Fields for MediaCoverImageField.
const (
	MediaCoverImageFieldExtraLarge MediaCoverImageField = "extraLarge"
	MediaCoverImageFieldLarge      MediaCoverImageField = "large"
	MediaCoverImageFieldMedium     MediaCoverImageField = "medium"
	MediaCoverImageFieldColor      MediaCoverImageField = "color"
)

// MediaFieldCoverImage to generate media cover image query.
func MediaFieldCoverImage(field MediaCoverImageField, fields ...MediaCoverImageField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("coverImage", nil, str...))
}

// MediaTagField is fields for media tag.
type MediaTagField string

// Fields for MediaTagField.
const (
	MediaTagFieldID               MediaTagField = "id"
	MediaTagFieldName             MediaTagField = "name"
	MediaTagFieldDescription      MediaTagField = "description"
	MediaTagFieldCategory         MediaTagField = "category"
	MediaTagFieldRank             MediaTagField = "rank"
	MediaTagFieldIsGeneralSpoiler MediaTagField = "isGeneralSpoiler"
	MediaTagFieldIsMediaSpoiler   MediaTagField = "isMediaSpoiler"
	MediaTagFieldIsAdult          MediaTagField = "isAdult"
)

// MediaFieldTags to generate media tags query.
func MediaFieldTags(field MediaTagField, fields ...MediaTagField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("tags", nil, str...))
}

// MediaConnectionField is fields for media connection.
type MediaConnectionField string

// MediaFieldRelations to generate media relations query.
func MediaFieldRelations(field MediaConnectionField, fields ...MediaConnectionField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("relations", nil, str...))
}

// MediaEdgeField is fields for media edge.
type MediaEdgeField string

// Fields for MediaEdgeField.
const (
	MediaEdgeFieldID             MediaEdgeField = "id"
	MediaEdgeFieldRelationType   MediaEdgeField = "relationType"
	MediaEdgeFieldRelationTypeV2 MediaEdgeField = "relationType(version:2)"
	MediaEdgeFieldIsMainStudio   MediaEdgeField = "isMainStudio"
	MediaEdgeFieldCharacterRole  MediaEdgeField = "characterRole"
	MediaEdgeFieldCharacterName  MediaEdgeField = "characterName"
	MediaEdgeFieldRoleNotes      MediaEdgeField = "roleNotes"
	MediaEdgeFieldDubGroup       MediaEdgeField = "dubGroup"
	MediaEdgeFieldStaffRole      MediaEdgeField = "staffRole"
	MediaEdgeFieldFavouriteOrder MediaEdgeField = "favouriteOrder"
)

// MediaConnectionFieldEdges to generate media connection edges query.
func MediaConnectionFieldEdges(field MediaEdgeField, fields ...MediaEdgeField) MediaConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaConnectionField(FieldObject("edges", nil, str...))
}

// MediaEdgeFieldNode to generate media edge node query.
func MediaEdgeFieldNode(field MediaField, fields ...MediaField) MediaEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaEdgeField(FieldObject("node", nil, str...))
}

// CharacterField is fields for character.
type CharacterField string

// Fields for CharacterField.
const (
	CharacterFieldID                 CharacterField = "id"
	CharacterFieldDescription        CharacterField = "description"
	CharacterFieldDescriptionAsHTML  CharacterField = "description(asHtml:true)"
	CharacterFieldGender             CharacterField = "gender"
	CharacterFieldDateOfBirth        CharacterField = "dateOfBirth{year month day}"
	CharacterFieldAge                CharacterField = "age"
	CharacterFieldIsFavourite        CharacterField = "isFavourite"
	CharacterFieldIsFavouriteBlocked CharacterField = "isFavouriteBlocked"
	CharacterFieldSiteURL            CharacterField = "siteUrl"
	CharacterFieldFavourite          CharacterField = "favourites"
	CharacterFieldModNotes           CharacterField = "modNotes"
)

// MediaEdgeFieldCharacters to generate media edge characters query.
func MediaEdgeFieldCharacters(field CharacterField, fields ...CharacterField) MediaEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaEdgeField(FieldObject("characters", nil, str...))
}

// CharacterNameField is fields for character name.
type CharacterNameField string

// Fields for CharacterNameField.
const (
	CharacterNameFieldFirst              CharacterNameField = "first"
	CharacterNameFieldMiddle             CharacterNameField = "middle"
	CharacterNameFieldLast               CharacterNameField = "last"
	CharacterNameFieldFull               CharacterNameField = "full"
	CharacterNameFieldNative             CharacterNameField = "native"
	CharacterNameFieldAlternative        CharacterNameField = "alternative"
	CharacterNameFieldAlternativeSpoiler CharacterNameField = "alternativeSpoiler"
	CharacterNameFieldUserPreferred      CharacterNameField = "userPreferred"
)

// CharacterFieldName to generate character name query.
func CharacterFieldName(field CharacterNameField, fields ...CharacterNameField) CharacterField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterField(FieldObject("name", nil, str...))
}

// CharacterImageField is fields for character image.
type CharacterImageField string

// Fields for CharacterImageField.
const (
	CharacterImageFieldLarge  CharacterImageField = "large"
	CharacterImageFieldMedium CharacterImageField = "medium"
)

// CharacterFieldImage to generate character image query.
func CharacterFieldImage(field CharacterImageField, fields ...CharacterImageField) CharacterField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterField(FieldObject("image", nil, str...))
}

// CharacterFieldMedia to generate character media query.
func CharacterFieldMedia(param CharacterParamMedia, field MediaConnectionField, fields ...MediaConnectionField) CharacterField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterField(FieldObject("media", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
		"type":    param.Type,
		"onList":  param.OnList,
		"sort":    param.Sort,
	}, str...))
}

// StaffField is fields for staff.
type StaffField string

// Fields for StaffField.
const (
	StaffFieldID                 StaffField = "id"
	StaffFieldLanguage           StaffField = "languageV2"
	StaffFieldDescription        StaffField = "description"
	StaffFieldDescriptionAsHTML  StaffField = "description(asHtml:true)"
	StaffFieldGender             StaffField = "gender"
	StaffFieldDateOfBirth        StaffField = "dateOfBirth{year month day}"
	StaffFieldDateOfDeath        StaffField = "dateOfDeath{year month day}"
	StaffFieldAge                StaffField = "age"
	StaffFieldYearsActive        StaffField = "yearsActive"
	StaffFieldHomeTown           StaffField = "homeTown"
	StaffFieldIsFavourite        StaffField = "isFavourite"
	StaffFieldIsFavouriteBlocked StaffField = "isFavouriteBlocked"
	StaffFieldSiteURL            StaffField = "siteUrl"
	StaffFieldSubmissionStatus   StaffField = "submissionStatus"
	StaffFieldSubmissionNotes    StaffField = "submissionNotes"
	StaffFieldFavourites         StaffField = "favourites"
	StaffFieldModNotes           StaffField = "ModNotes"
)

// StaffNameField is fields for staff name.
type StaffNameField string

// Fields for StaffNameField.
const (
	StaffNameFieldFirst         StaffNameField = "first"
	StaffNameFieldMiddle        StaffNameField = "middle"
	StaffNameFieldLast          StaffNameField = "last"
	StaffNameFieldFull          StaffNameField = "full"
	StaffNameFieldNative        StaffNameField = "native"
	StaffNameFieldAlternative   StaffNameField = "alternative"
	StaffNameFieldUserPreferred StaffNameField = "userPreferred"
)

// StaffFieldName to generate staff name query.
func StaffFieldName(field StaffNameField, fields ...StaffNameField) StaffField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffField(FieldObject("name", nil, str...))
}

// StaffImageField is fields for staff image.
type StaffImageField string

// Fields for StaffImageField.
const (
	StaffImageFieldLarge  StaffImageField = "large"
	StaffImageFieldMedium StaffImageField = "medium"
)

// StaffFieldImage to generate staff image query.
func StaffFieldImage(field StaffImageField, fields ...StaffImageField) StaffField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffField(FieldObject("image", nil, str...))
}

// MediaEdgeFieldVoiceActors to generate media edge voice actors query.
func MediaEdgeFieldVoiceActors(param MediaEdgeParamVoiceActors, field StaffField, fields ...StaffField) MediaEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaEdgeField(FieldObject("voiceActors", QueryParam{
		"language": param.Language,
		"sort":     param.Sort,
	}, str...))
}

// StaffFieldStaffMedia to generate staff's staff media query.
func StaffFieldStaffMedia(param StaffParamStaffMedia, field MediaConnectionField, fields ...MediaConnectionField) StaffField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffField(FieldObject("staffMedia", QueryParam{
		"type":    param.Type,
		"onList":  param.OnList,
		"page":    param.Page,
		"perPage": param.PerPage,
		"sort":    param.Sort,
	}, str...))
}

// StaffFieldCharacters to generate staff characters query.
func StaffFieldCharacters(param StaffParamCharacters, field CharacterConnectionField, fields ...CharacterConnectionField) StaffField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffField(FieldObject("characters", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
		"sort":    param.Sort,
	}, str...))
}

// StaffFieldCharacterMedia to generate staff character media query.
func StaffFieldCharacterMedia(param StaffParamCharacterMedia, field MediaConnectionField, fields ...MediaConnectionField) StaffField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffField(FieldObject("characterMedia", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
		"sort":    param.Sort,
		"onList":  param.OnList,
	}, str...))
}

// StaffFieldStaff to generate staff's staff query.
func StaffFieldStaff(field StaffField, fields ...StaffField) StaffField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffField(FieldObject("staff", nil, str...))
}

// UserField is fields for user.
type UserField string

// Fields for UserField.
const (
	UserFieldID                      UserField = "id"
	UserFieldName                    UserField = "name"
	UserFieldAbout                   UserField = "about"
	UserFieldAboutAsHTML             UserField = "about(asHtml:true)"
	UserFieldBannerImage             UserField = "bannerImage"
	UserFieldIsFollowing             UserField = "isFollowing"
	UserFieldIsFollower              UserField = "isFollower"
	UserFieldIsBlocked               UserField = "isBlocked"
	UserFieldBans                    UserField = "bans"
	UserFieldUnreadNotificationCount UserField = "unreadNotificationCount"
	UserFieldSiteURL                 UserField = "siteUrl"
	UserFieldDonatorTier             UserField = "donatorTier"
	UserFieldDonatorBadge            UserField = "donatorBadge"
	UserFieldModeratorRoles          UserField = "moderatorRoles"
	UserFieldCreatedAt               UserField = "createdAt"
	UserFieldUpdatedAt               UserField = "updatedAt"
)

// UserAvatarField is fields for user avatar.
type UserAvatarField string

// Fields for UserAvatarField.
const (
	UserAvatarFieldLarge  UserAvatarField = "large"
	UserAvatarFieldMedium UserAvatarField = "medium"
)

// UserFieldAvatar to generate user avatar query.
func UserFieldAvatar(field UserAvatarField, fields ...UserAvatarField) UserField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserField(FieldObject("avatar", nil, str...))
}

// UserPreviousNameField is fields for user previous name.
type UserPreviousNameField string

// Fields for UserPreviousNameField.
const (
	UserPreviousNameFieldName      UserPreviousNameField = "name"
	UserPreviousNameFieldCreatedAt UserPreviousNameField = "createdAt"
	UserPreviousNameFieldUpdatedAt UserPreviousNameField = "udpatedAt"
)

// UserFieldPreviousNames to generate user previous names query.
func UserFieldPreviousNames(field UserPreviousNameField, fields ...UserPreviousNameField) UserField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserField(FieldObject("previousNames", nil, str...))
}

// UserStatisticTypesField is fields for user statistic types.
type UserStatisticTypesField string

// UserStatisticsField is fields for user statistics.
type UserStatisticsField string

// Fields for UserStatisticsField.
const (
	UserStatisticsFieldCount             UserStatisticsField = "count"
	UserStatisticsFieldMeanScore         UserStatisticsField = "meanScore"
	UserStatisticsFieldStandardDeviation UserStatisticsField = "standardDeviation"
	UserStatisticsFieldMinutesWatched    UserStatisticsField = "minutesWatched"
	UserStatisticsFieldEpisodesWatched   UserStatisticsField = "episodesWatched"
	UserStatisticsFieldChaptersRead      UserStatisticsField = "chaptersRead"
	UserStatisticsFieldVolumesRead       UserStatisticsField = "volumesRead"
)

// UserFormatStatisticField is fields for user format statistic.
type UserFormatStatisticField string

// Fields for UserFormatStatisticField.
const (
	UserFormatStatisticFieldCount          UserFormatStatisticField = "count"
	UserFormatStatisticFieldMeanScore      UserFormatStatisticField = "meanScore"
	UserFormatStatisticFieldMinutesWatched UserFormatStatisticField = "minutesWatched"
	UserFormatStatisticFieldChaptersRead   UserFormatStatisticField = "chaptersRead"
	UserFormatStatisticFieldMediaIDs       UserFormatStatisticField = "mediaIds"
	UserFormatStatisticFieldFormat         UserFormatStatisticField = "format"
)

// UserStatisticsFieldFormats to generate user statistics formats query.
func UserStatisticsFieldFormats(param UserStatisticsParamFormats, field UserFormatStatisticField, fields ...UserFormatStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("formats", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserStatusStatisticField is fields for user status statistic.
type UserStatusStatisticField string

// Fields for UserStatusStatisticField.
const (
	UserStatusStatisticFieldCount          UserStatusStatisticField = "count"
	UserStatusStatisticFieldMeanScore      UserStatusStatisticField = "meanScore"
	UserStatusStatisticFieldMinutesWatched UserStatusStatisticField = "minutesWatched"
	UserStatusStatisticFieldChaptersRead   UserStatusStatisticField = "chaptersRead"
	UserStatusStatisticFieldMediaIDs       UserStatusStatisticField = "mediaIds"
	UserStatusStatisticFieldStatus         UserStatusStatisticField = "status"
)

// UserStatisticsFieldStatuses to generate user statistics statuses query.
func UserStatisticsFieldStatuses(param UserStatisticsParamStatuses, field UserStatusStatisticField, fields ...UserStatusStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("statuses", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserScoreStatisticField is fields for user score statistic.
type UserScoreStatisticField string

// Fields for UserScoreStatisticField.
const (
	UserScoreStatisticFieldCount          UserScoreStatisticField = "count"
	UserScoreStatisticFieldMeanScore      UserScoreStatisticField = "meanScore"
	UserScoreStatisticFieldMinutesWatched UserScoreStatisticField = "minutesWatched"
	UserScoreStatisticFieldChaptersRead   UserScoreStatisticField = "chaptersRead"
	UserScoreStatisticFieldMediaIDs       UserScoreStatisticField = "mediaIds"
	UserScoreStatisticFieldScore          UserScoreStatisticField = "score"
)

// UserLengthStatisticField is fields for user length statistic.
type UserLengthStatisticField string

// Fields for UserLengthStatisticField.
const (
	UserLengthStatisticFieldCount          UserLengthStatisticField = "count"
	UserLengthStatisticFieldMeanScore      UserLengthStatisticField = "meanScore"
	UserLengthStatisticFieldMinutesWatched UserLengthStatisticField = "minutesWatched"
	UserLengthStatisticFieldChaptersRead   UserLengthStatisticField = "chaptersRead"
	UserLengthStatisticFieldMediaIDs       UserLengthStatisticField = "mediaIds"
	UserLengthStatisticFieldLength         UserLengthStatisticField = "length"
)

// UserStatisticsFieldLength to generate user statistics length query.
func UserStatisticsFieldLength(param UserStatisticsParamLengths, field UserLengthStatisticField, fields ...UserLengthStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("lengths", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserReleaseYearStatisticField is fields for user release year statistic.
type UserReleaseYearStatisticField string

// Fields for UserReleaseYearStatisticField.
const (
	UserReleaseYearStatisticFieldCount          UserReleaseYearStatisticField = "count"
	UserReleaseYearStatisticFieldMeanScore      UserReleaseYearStatisticField = "meanScore"
	UserReleaseYearStatisticFieldMinutesWatched UserReleaseYearStatisticField = "minutesWatched"
	UserReleaseYearStatisticFieldChaptersRead   UserReleaseYearStatisticField = "chaptersRead"
	UserReleaseYearStatisticFieldMediaIDs       UserReleaseYearStatisticField = "mediaIds"
	UserReleaseYearStatisticFieldReleaseYear    UserReleaseYearStatisticField = "releaseYear"
)

// UserStatisticsFieldReleaseYears to generate user statistic release years query.
func UserStatisticsFieldReleaseYears(param UserStatisticsParamReleaseYears, field UserReleaseYearStatisticField, fields ...UserReleaseYearStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("releaseYears", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserStartYearStatisticField is fields for user start year.
type UserStartYearStatisticField string

// Fields for UserStartYearStatisticField.
const (
	UserStartYearStatisticFieldCount          UserStartYearStatisticField = "count"
	UserStartYearStatisticFieldMeanScore      UserStartYearStatisticField = "meanScore"
	UserStartYearStatisticFieldMinutesWatched UserStartYearStatisticField = "minutesWatched"
	UserStartYearStatisticFieldChaptersRead   UserStartYearStatisticField = "chaptersRead"
	UserStartYearStatisticFieldMediaIDs       UserStartYearStatisticField = "mediaIds"
	UserStartYearStatisticFieldStartYear      UserStartYearStatisticField = "startYear"
)

// UserStatisticsFieldStartYears to generate user statistics start years.
func UserStatisticsFieldStartYears(param UserStatisticsParamStartYears, field UserStartYearStatisticField, fields ...UserStartYearStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("startYears", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserGenreStatisticField is fields for user genre statistic.
type UserGenreStatisticField string

// Fields for UserGenreStatisticField.
const (
	UserGenreStatisticFieldCount          UserGenreStatisticField = "count"
	UserGenreStatisticFieldMeanScore      UserGenreStatisticField = "meanScore"
	UserGenreStatisticFieldMinutesWatched UserGenreStatisticField = "minutesWatched"
	UserGenreStatisticFieldChaptersRead   UserGenreStatisticField = "chaptersRead"
	UserGenreStatisticFieldMediaIDs       UserGenreStatisticField = "mediaIds"
	UserGenreStatisticFieldGenre          UserGenreStatisticField = "genre"
)

// UserStatisticsFieldGenre to generate user statistics genre query.
func UserStatisticsFieldGenre(param UserStatisticsParamGenres, field UserGenreStatisticField, fields ...UserGenreStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("genres", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserTagStatisticField is fields for user tag statistic.
type UserTagStatisticField string

// Fields for UserTagStatisticField.
const (
	UserTagStatisticFieldCount          UserTagStatisticField = "count"
	UserTagStatisticFieldMeanScore      UserTagStatisticField = "meanScore"
	UserTagStatisticFieldMinutesWatched UserTagStatisticField = "minutesWatched"
	UserTagStatisticFieldChaptersRead   UserTagStatisticField = "chaptersRead"
	UserTagStatisticFieldMediaIDs       UserTagStatisticField = "mediaIds"
	UserTagStatisticFieldTag            UserTagStatisticField = "tag"
)

// UserStatisticsFieldTags to generate user statistics tags query.
func UserStatisticsFieldTags(param UserStatisticsParamTags, field UserTagStatisticField, fields ...UserTagStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("tags", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserCountryStatisticField is fields for user country statistic.
type UserCountryStatisticField string

// Fields for UserCountryStatisticField.
const (
	UserCountryStatisticFieldCount          UserCountryStatisticField = "count"
	UserCountryStatisticFieldMeanScore      UserCountryStatisticField = "meanScore"
	UserCountryStatisticFieldMinutesWatched UserCountryStatisticField = "minutesWatched"
	UserCountryStatisticFieldChaptersRead   UserCountryStatisticField = "chaptersRead"
	UserCountryStatisticFieldMediaIDs       UserCountryStatisticField = "mediaIds"
	UserCountryStatisticFieldCountry        UserCountryStatisticField = "country"
)

// UserStatisticsFieldCountries to generate user statistics countries query.
func UserStatisticsFieldCountries(param UserStatisticsParamCountries, field UserCountryStatisticField, fields ...UserCountryStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("countries", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserVoiceActorStatisticField is fields for user voice actor statistic.
type UserVoiceActorStatisticField string

// Fields for UserVoiceActorStatisticField.
const (
	UserVoiceActorStatisticFieldCount          UserVoiceActorStatisticField = "count"
	UserVoiceActorStatisticFieldMeanScore      UserVoiceActorStatisticField = "meanScore"
	UserVoiceActorStatisticFieldMinutesWatched UserVoiceActorStatisticField = "minutesWatched"
	UserVoiceActorStatisticFieldChaptersRead   UserVoiceActorStatisticField = "chaptersRead"
	UserVoiceActorStatisticFieldMediaIDs       UserVoiceActorStatisticField = "mediaIds"
	UserVoiceActorStatisticFieldCharacterIDs   UserVoiceActorStatisticField = "characterIds"
)

// UserStatisticsFieldVoiceActors to generate user statistics voice actors query.
func UserStatisticsFieldVoiceActors(param UserStatisticsParamVoiceActors, field UserVoiceActorStatisticField, fields ...UserVoiceActorStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("voiceActors", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserStaffStatisticField is fields for user staff statistic.
type UserStaffStatisticField string

// Fields for UserStaffStatisticField.
const (
	UserStaffStatisticFieldCount          UserStaffStatisticField = "count"
	UserStaffStatisticFieldMeanScore      UserStaffStatisticField = "meanScore"
	UserStaffStatisticFieldMinutesWatched UserStaffStatisticField = "minutesWatched"
	UserStaffStatisticFieldChaptersRead   UserStaffStatisticField = "chaptersRead"
	UserStaffStatisticFieldMediaIDs       UserStaffStatisticField = "mediaIds"
)

// UserStaffStatisticFieldStaff to generate user staff statistic staff query.
func UserStaffStatisticFieldStaff(field StaffField, fields ...StaffField) UserStaffStatisticField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStaffStatisticField(FieldObject("staff", nil, str...))
}

// UserStatisticsFieldStaff to generate user statistics staff query.
func UserStatisticsFieldStaff(param UserStatisticsParamStaff, field UserStaffStatisticField, fields ...UserStaffStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("staff", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserStudioStatisticField is fields for user studio statistic.
type UserStudioStatisticField string

// Fields for UserStudioStatisticField.
const (
	UserStudioStatisticFieldCount          UserStudioStatisticField = "count"
	UserStudioStatisticFieldMeanScore      UserStudioStatisticField = "meanScore"
	UserStudioStatisticFieldMinutesWatched UserStudioStatisticField = "minutesWatched"
	UserStudioStatisticFieldChaptersRead   UserStudioStatisticField = "chaptersRead"
	UserStudioStatisticFieldMediaIDs       UserStudioStatisticField = "mediaIds"
)

// UserStudioStatisticFieldStudio to generate user studio statistic studio query.
func UserStudioStatisticFieldStudio(field StudioField, fields ...StudioField) UserStudioStatisticField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStudioStatisticField(FieldObject("studio", nil, str...))
}

// UserStatisticsFieldStudios to generate user statistics studios query.
func UserStatisticsFieldStudios(param UserStatisticsParamStudios, field UserStudioStatisticField, fields ...UserStudioStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("studios", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserVoiceActorStatisticFieldVoiceActor to generate user voice actor statistic voice actor query.
func UserVoiceActorStatisticFieldVoiceActor(field StaffField, fields ...StaffField) UserVoiceActorStatisticField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserVoiceActorStatisticField(FieldObject("voiceActor", nil, str...))
}

// UserStatisticsFieldScores to generate user statistics scores query.
func UserStatisticsFieldScores(param UserStatisticsParamScores, field UserScoreStatisticField, fields ...UserScoreStatisticField) UserStatisticsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticsField(FieldObject("scores", QueryParam{
		"limit": param.Limit,
		"sort":  param.Sort,
	}, str...))
}

// UserStatisticTypesFieldAnime to generate user statistic type anime query.
func UserStatisticTypesFieldAnime(field UserStatisticsField, fields ...UserStatisticsField) UserStatisticTypesField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticTypesField(FieldObject("anime", nil, str...))
}

// UserStatisticTypesFieldManga to generate user statistic type manga query.
func UserStatisticTypesFieldManga(field UserStatisticsField, fields ...UserStatisticsField) UserStatisticTypesField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserStatisticTypesField(FieldObject("manga", nil, str...))
}

// UserFieldStatistics to generate user statistics query.
func UserFieldStatistics(field UserStatisticTypesField, fields ...UserStatisticTypesField) UserField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserField(FieldObject("statistics", nil, str...))
}

// FavouritesField is fields for favourites.
type FavouritesField string

// FavouritesFieldAnime to generate favourites anime query.
func FavouritesFieldAnime(param FavouritesParamAnime, field MediaConnectionField, fields ...MediaConnectionField) FavouritesField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return FavouritesField(FieldObject("anime", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
	}, str...))
}

// FavouritesFieldManga to generate favourites manga query.
func FavouritesFieldManga(param FavouritesParamManga, field MediaConnectionField, fields ...MediaConnectionField) FavouritesField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return FavouritesField(FieldObject("manga", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
	}, str...))
}

// FavouritesFieldCharacters to generate favourites characters query.
func FavouritesFieldCharacters(param FavouritesParamCharacters, field CharacterConnectionField, fields ...CharacterConnectionField) FavouritesField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return FavouritesField(FieldObject("characters", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
	}, str...))
}

// FavouritesFieldStaff to generate favourites staff query.
func FavouritesFieldStaff(param FavouritesParamStaff, field StaffConnectionField, fields ...StaffConnectionField) FavouritesField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return FavouritesField(FieldObject("staff", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
	}, str...))
}

// FavouritesFieldStudios to generate favourites studio query.
func FavouritesFieldStudios(param FavouritesParamStudios, field StudioConnectionField, fields ...StudioConnectionField) FavouritesField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return FavouritesField(FieldObject("studios", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
	}, str...))
}

// UserFieldFavourites to generate user favourites query.
func UserFieldFavourites(param UserParamFavourites, field FavouritesField, fields ...FavouritesField) UserField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserField(FieldObject("favourites", QueryParam{
		"page": param.Page,
	}, str...))
}

// UserOptionsField is fields for user options.
type UserOptionsField string

// FIelds for UserOptionsField.
const (
	UserOptionsFieldTitleLanguage       UserOptionsField = "titleLanguage"
	UserOptionsFieldDisplayAdultContent UserOptionsField = "displayAdultContent"
	UserOptionsFieldAiringNotifications UserOptionsField = "airingNotifications"
	UserOptionsFieldProfileColor        UserOptionsField = "profileColor"
	UserOptionsFieldTimezone            UserOptionsField = "timezone"
	UserOptionsFieldActivityMergeTime   UserOptionsField = "activityMergeTime"
	UserOptionsFieldStaffNameLanguage   UserOptionsField = "staffNameLanguage"
)

// UserFieldOptions to generate user options query.
func UserFieldOptions(field UserOptionsField, fields ...UserOptionsField) UserField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserField(FieldObject("options", nil, str...))
}

// MediaListOptionsField is fields for media list options.
type MediaListOptionsField string

// Fields for MediaListOptionsField.
const (
	MediaListOptionsFieldScoreFormat MediaListOptionsField = "scoreFormat"
	MediaListOptionsFieldRowOrder    MediaListOptionsField = "rowOrder"
)

// MediaListTypeOptionsField is fields for media list type options.
type MediaListTypeOptionsField string

// Fields for MediaListTypeOptionsField.
const (
	MediaListTypeOptionsFieldSectionOrder                  MediaListTypeOptionsField = "sectionOrder"
	MediaListTypeOptionsFieldSplitCompletedSectionByFormat MediaListTypeOptionsField = "splitCompletedSectionByFormat"
	MediaListTypeOptionsFieldCustomLists                   MediaListTypeOptionsField = "customLists"
	MediaListTypeOptionsFieldAdvancedScoring               MediaListTypeOptionsField = "advancedScoring"
	MediaListTypeOptionsFieldAdvancedScoringEnabled        MediaListTypeOptionsField = "advancedScoringEnabled"
)

// MediaListOptionsFieldAnimeList to generate media list options animelist query.
func MediaListOptionsFieldAnimeList(field MediaListTypeOptionsField, fields ...MediaListTypeOptionsField) MediaListOptionsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaListOptionsField(FieldObject("animeList", nil, str...))
}

// MediaListOptionsFieldMangaList to generate media list options mangalist query.
func MediaListOptionsFieldMangaList(field MediaListTypeOptionsField, fields ...MediaListTypeOptionsField) MediaListOptionsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaListOptionsField(FieldObject("mangaList", nil, str...))
}

// UserFieldMediaListOptions to generate user media list options query.
func UserFieldMediaListOptions(field MediaListOptionsField, fields ...MediaListOptionsField) UserField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserField(FieldObject("mediaListOptions", nil, str...))
}

// NotificationOptionField is fields for notification option.
type NotificationOptionField string

// Fields for NotificationOptionField.
const (
	NotificationOptionFieldType    NotificationOptionField = "type"
	NotificationOptionFieldEnabled NotificationOptionField = "enabled"
)

// UserOptionsFieldNotificationOptions to generate user options notification options query.
func UserOptionsFieldNotificationOptions(field NotificationOptionField, fields ...NotificationOptionField) UserOptionsField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return UserOptionsField(FieldObject("notificationOptions", nil, str...))
}

// StaffFieldSubmitter to generate staff submitter query.
func StaffFieldSubmitter(field UserField, fields ...UserField) StaffField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffField(FieldObject("user", nil, str...))
}

// StaffRoleTypeField is fields for staff role type.
type StaffRoleTypeField string

// Fields for StaffRoleTypeField.
const (
	StaffRoleTypeFieldRoleNotes StaffRoleTypeField = "roleNotes"
	StaffRoleTypeFieldDubGroup  StaffRoleTypeField = "dubGroup"
)

// MediaEdgeFieldVoiceActorRoles to generate media edge voice actor roles query.
func MediaEdgeFieldVoiceActorRoles(param MediaEdgeParamVoiceActorRoles, field StaffRoleTypeField, fields ...StaffRoleTypeField) MediaEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaEdgeField(FieldObject("voiceActorRoles", QueryParam{
		"language": param.Language,
		"sort":     param.Sort,
	}, str...))
}

// StaffRoleTypeFieldVoiceActor to generate staff role type voice actor query.
func StaffRoleTypeFieldVoiceActor(field StaffField, fields ...StaffField) StaffRoleTypeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffRoleTypeField(FieldObject("voiceActor", nil, str...))
}

// MediaConnectionFieldNodes to generate media connection nodes query.
func MediaConnectionFieldNodes(field MediaField, fields ...MediaField) MediaConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaConnectionField(FieldObject("nodes", nil, str...))
}

// PageInfoField is fields for page info.
type PageInfoField string

// Fields for PageInfoField.
const (
	PageInfoFieldTotal       PageInfoField = "total"
	PageInfoFieldPerPage     PageInfoField = "perPage"
	PageInfoFieldCurrentPage PageInfoField = "currentPage"
	PageInfoFieldLastPage    PageInfoField = "lastPage"
	PageInfoFieldHasNextPage PageInfoField = "hasNextPage"
)

// MediaConnectionFieldPageInfo to generate media connection page info query.
func MediaConnectionFieldPageInfo(field PageInfoField, fields ...PageInfoField) MediaConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaConnectionField(FieldObject("pageInfo", nil, str...))
}

// CharacterConnectionField is fields for character connection.
type CharacterConnectionField string

// MediaFieldCharacters to generate media characters query.
func MediaFieldCharacters(param MediaParamCharacters, field CharacterConnectionField, fields ...CharacterConnectionField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("characters", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
		"role":    param.Role,
		"sort":    param.Sort,
	}, str...))
}

// CharacterEdgeField is fields for character edge.
type CharacterEdgeField string

// Fields for CharacterEdgeField.
const (
	CharacterEdgeFieldID             CharacterEdgeField = "id"
	CharacterEdgeFieldRole           CharacterEdgeField = "role"
	CharacterEdgeFieldName           CharacterEdgeField = "name"
	CharacterEdgeFieldFavouriteOrder CharacterEdgeField = "favouriteOrder"
)

// CharacterConnectionFieldEdges to generate character connection edges query.
func CharacterConnectionFieldEdges(field CharacterEdgeField, fields ...CharacterEdgeField) CharacterConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterConnectionField(FieldObject("edges", nil, str...))
}

// CharacterEdgeFieldNode to generate character edge node query.
func CharacterEdgeFieldNode(field CharacterField, fields ...CharacterField) CharacterEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterEdgeField(FieldObject("node", nil, str...))
}

// CharacterConnectionFieldNodes to generate character connection node query.
func CharacterConnectionFieldNodes(field CharacterField, fields ...CharacterField) CharacterConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterConnectionField(FieldObject("nodes", nil, str...))
}

// CharacterEdgeFieldVoiceActors to generate character edge voice actors query.
func CharacterEdgeFieldVoiceActors(param CharacterEdgeParamVoiceActors, field StaffField, fields ...StaffField) CharacterEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterEdgeField(FieldObject("voiceActors", QueryParam{
		"language": param.Language,
		"sort":     param.Sort,
	}, str...))
}

// CharacterEdgeFieldVoiceActorRoles to generate character edge voice actor roles query.
func CharacterEdgeFieldVoiceActorRoles(param CharacterEdgeParamVoiceActorRoles, field StaffRoleTypeField, fields ...StaffRoleTypeField) CharacterEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterEdgeField(FieldObject("voiceActorRoles", QueryParam{
		"language": param.Language,
		"sort":     param.Sort,
	}, str...))
}

// CharacterEdgeFieldMedia to generate character edge media query.
func CharacterEdgeFieldMedia(field MediaField, fields ...MediaField) CharacterEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterEdgeField(FieldObject("media", nil, str...))
}

// CharacterConnectionFieldPageInfo to generate character connection page info query.
func CharacterConnectionFieldPageInfo(field PageInfoField, fields ...PageInfoField) CharacterConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return CharacterConnectionField(FieldObject("pageInfo", nil, str...))
}

// StaffConnectionField is fields for staff connection.
type StaffConnectionField string

// MediaFieldStaff to generate media staff query.
func MediaFieldStaff(param MediaParamStaff, field StaffConnectionField, fields ...StaffConnectionField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("staff", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
		"sort":    param.Sort,
	}, str...))
}

// StaffEdgeField is fields for staff edge query.
type StaffEdgeField string

// Fields for StaffEdgeField.
const (
	StaffEdgeFieldID             StaffEdgeField = "id"
	StaffEdgeFieldRole           StaffEdgeField = "role"
	StaffEdgeFieldFavouriteOrder StaffEdgeField = "favouriteOrder"
)

// StaffConnectionFieldEdges to generate staff connection edges query.
func StaffConnectionFieldEdges(field StaffEdgeField, fields ...StaffEdgeField) StaffConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffConnectionField(FieldObject("edges", nil, str...))
}

// StaffEdgeFieldNode to generate staff edge node query.
func StaffEdgeFieldNode(field StaffField, fields ...StaffField) StaffEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffEdgeField(FieldObject("node", nil, str...))
}

// StaffConnectionFieldNode to generate staff connection node.
func StaffConnectionFieldNode(field StaffField, fields ...StaffField) StaffConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffConnectionField(FieldObject("nodes", nil, str...))
}

// StaffConnectionFieldPageInfo to generate staff connection page info query.
func StaffConnectionFieldPageInfo(field PageInfoField, fields ...PageInfoField) StaffConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StaffConnectionField(FieldObject("pageInfo", nil, str...))
}

// StudioConnectionField is fields for studio connection.
type StudioConnectionField string

// MediaFieldStudios to generate media studios query.
func MediaFieldStudios(param MediaParamStudios, field StudioConnectionField, fields ...StudioConnectionField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("studios", QueryParam{
		"isMain": param.IsMain,
		"sort":   param.Sort,
	}, str...))
}

// StudioEdgeField is fields for studio edge.
type StudioEdgeField string

// Fields for StudioEdgeField.
// Fields for StudioEdgeField.
const (
	StudioEdgeFieldID             StudioEdgeField = "id"
	StudioEdgeFieldIsMain         StudioEdgeField = "isMain"
	StudioEdgeFieldFavouriteOrder StudioEdgeField = "favouriteOrder"
)

// StudioConnectionFieldEdges to generate studio connection edges query.
func StudioConnectionFieldEdges(field StudioEdgeField, fields ...StudioEdgeField) StudioConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StudioConnectionField(FieldObject("edges", nil, str...))
}

// StudioEdgeFieldNode to generate studio edge node query.
func StudioEdgeFieldNode(field StudioField, fields ...StudioField) StudioEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StudioEdgeField(FieldObject("node", nil, str...))
}

// StudioField is fields for studio.
type StudioField string

// Fields for StudioField.
const (
	StudioFieldID                StudioField = "id"
	StudioFieldName              StudioField = "name"
	StudioFieldIsAnimationStudio StudioField = "isAnimationStudio"
	StudioFieldSiteURL           StudioField = "siteUrl"
	StudioFieldIsFavourite       StudioField = "isFavourite"
	StudioFieldFavourites        StudioField = "favourites"
)

// StudioConnectionFieldNodes to generate studio connection nodes query.
func StudioConnectionFieldNodes(field StudioField, fields ...StudioField) StudioConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StudioConnectionField(FieldObject("nodes", nil, str...))
}

// StudioFieldMedia to generate studio media query.
func StudioFieldMedia(param StudioParamMedia, field MediaConnectionField, fields ...MediaConnectionField) StudioField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StudioField(FieldObject("media", QueryParam{
		"isMain":  param.IsMain,
		"onList":  param.OnList,
		"page":    param.Page,
		"perPage": param.PerPage,
		"sort":    param.Sort,
	}, str...))
}

// StudioConnectionFieldPageInfo to generate studio connection page info query.
func StudioConnectionFieldPageInfo(field PageInfoField, fields ...PageInfoField) StudioConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return StudioConnectionField(FieldObject("pageInfo", nil, str...))
}

// AiringScheduleField is fields for airing schedule.
type AiringScheduleField string

// Fields for AiringScheduleField.
const (
	AiringScheduleFieldID              AiringScheduleField = "id"
	AiringScheduleFieldAiringAt        AiringScheduleField = "airingAt"
	AiringScheduleFieldTimeUntilAiring AiringScheduleField = "timeUntilAiring"
	AiringScheduleFieldEpisode         AiringScheduleField = "episode"
	AiringScheduleFieldMediaID         AiringScheduleField = "mediaId"
)

// MediaFieldNextAiringEpisode to generate media next airing episode query.
func MediaFieldNextAiringEpisode(field AiringScheduleField, fields ...AiringScheduleField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("nextAiringEpisode", nil, str...))
}

// AiringScheduleFieldMedia to generate airing schedule media query.
func AiringScheduleFieldMedia(field MediaField, fields ...MediaField) AiringScheduleField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return AiringScheduleField(FieldObject("media", nil, str...))
}

// MediaFieldAiringSchedule to generate media airing schedule query.
func MediaFieldAiringSchedule(param MediaParamAiringSchedule, field AiringScheduleConnectionField, fields ...AiringScheduleConnectionField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("airingSchedule", QueryParam{
		"notYetAired": param.NotYetAired,
		"page":        param.Page,
		"perPage":     param.PerPage,
	}, str...))
}

// AiringScheduleConnectionField is fields for airing schedule connection.
type AiringScheduleConnectionField string

// AiringScheduleEdgeField is fields for airing schedule edge.
type AiringScheduleEdgeField string

// Fields for AiringScheduleEdgeField.
const (
	AiringScheduleEdgeFieldID AiringScheduleEdgeField = "id"
)

// AiringScheduleConnectionFieldEdges to generate airing schedule connection edge query.
func AiringScheduleConnectionFieldEdges(field AiringScheduleEdgeField, fields ...AiringScheduleEdgeField) AiringScheduleConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return AiringScheduleConnectionField(FieldObject("edges", nil, str...))
}

// AiringScheduleEdgeFieldNode to generate airing schedule edge node query.
func AiringScheduleEdgeFieldNode(field AiringScheduleField, fields ...AiringScheduleField) AiringScheduleEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return AiringScheduleEdgeField(FieldObject("node", nil, str...))
}

// AiringScheduleConnectionFieldNodes to generate airing schedule connection nodes query.
func AiringScheduleConnectionFieldNodes(field AiringScheduleField, fields ...AiringScheduleField) AiringScheduleConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return AiringScheduleConnectionField(FieldObject("nodes", nil, str...))
}

// AiringScheduleConnectionFieldPageInfo to generate airing schedule connection page info query.
func AiringScheduleConnectionFieldPageInfo(field PageInfoField, fields ...PageInfoField) AiringScheduleConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return AiringScheduleConnectionField(FieldObject("pageInfo", nil, str...))
}

// MediaTrendConnectionField is fields for media trend connection.
type MediaTrendConnectionField string

// MediaFieldTrends to generate media trends query.
func MediaFieldTrends(param MediaParamTrends, field MediaTrendConnectionField, fields ...MediaTrendConnectionField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("trends", QueryParam{
		"page":      param.Page,
		"perPage":   param.PerPage,
		"releasing": param.Releasing,
		"sort":      param.Sort,
	}, str...))
}

// MediaTrendEdgeField is fields for media trend edge.
type MediaTrendEdgeField string

// MediaTrendConnectionFieldEdges to generate media trend connection edges query.
func MediaTrendConnectionFieldEdges(field MediaTrendEdgeField, fields ...MediaTrendEdgeField) MediaTrendConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaTrendConnectionField(FieldObject("edges", nil, str...))
}

// MediaTrendEdgeFieldNode to generate media trend edge node query.
func MediaTrendEdgeFieldNode(field MediaTrendField, fields ...MediaTrendField) MediaTrendEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaTrendEdgeField(FieldObject("node", nil, str...))
}

// MediaTrendField is fields for media trend.
type MediaTrendField string

// Fields for MediaTrendField.
const (
	MediaTrendFieldMediaID      MediaTrendField = "mediaID"
	MediaTrendFieldDate         MediaTrendField = "date"
	MediaTrendFieldTrending     MediaTrendField = "trending"
	MediaTrendFieldAverageScore MediaTrendField = "averageScore"
	MediaTrendFieldPopularity   MediaTrendField = "popularity"
	MediaTrendFieldInProgress   MediaTrendField = "inProgress"
	MediaTrendFieldReleasing    MediaTrendField = "releasing"
	MediaTrendFieldEpisode      MediaTrendField = "episode"
)

// MediaTrendConnectionFieldNodes to generate media trend connection nodes query.
func MediaTrendConnectionFieldNodes(field MediaTrendField, fields ...MediaTrendField) MediaTrendConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaTrendConnectionField(FieldObject("nodes", nil, str...))
}

// MediaTrendFieldMedia to generate media trend media query.
func MediaTrendFieldMedia(field MediaField, fields ...MediaField) MediaTrendField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaTrendField(FieldObject("media", nil, str...))
}

// MediaTrendConnectionFieldPageInfo to generate media trend connection page info query.
func MediaTrendConnectionFieldPageInfo(field PageInfoField, fields ...PageInfoField) MediaTrendConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaTrendConnectionField(FieldObject("pageInfo", nil, str...))
}

// MediaExternalLinkField is fields for media external link.
type MediaExternalLinkField string

// Fields for MediaExternalLinkField.
const (
	MediaExternalLinkFieldID   MediaExternalLinkField = "id"
	MediaExternalLinkFieldURL  MediaExternalLinkField = "url"
	MediaExternalLinkFieldSite MediaExternalLinkField = "site"
)

// MediaFieldExternalLinks to generate media external links query.
func MediaFieldExternalLinks(field MediaExternalLinkField, fields ...MediaExternalLinkField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("externalLinks", nil, str...))
}

// MediaStreamingEpisodeField is fields media streaming episode.
type MediaStreamingEpisodeField string

// Fields for MediaStreamingEpisodeField.
const (
	MediaStreamingEpisodeFieldTitle     MediaStreamingEpisodeField = "title"
	MediaStreamingEpisodeFieldThumbnail MediaStreamingEpisodeField = "thumbnail"
	MediaStreamingEpisodeFieldURL       MediaStreamingEpisodeField = "url"
	MediaStreamingEpisodeFieldSite      MediaStreamingEpisodeField = "site"
)

// MediaFieldStreamingEpisodes to generate media streaming episodes query.
func MediaFieldStreamingEpisodes(field MediaStreamingEpisodeField, fields ...MediaStreamingEpisodeField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("streamingEpisodes", nil, str...))
}

// MediaRankField is fields for media rank.
type MediaRankField string

// Fields for MediaRankField.
const (
	MediaRankFieldID      MediaRankField = "id"
	MediaRankFieldRank    MediaRankField = "rank"
	MediaRankFieldType    MediaRankField = "type"
	MediaRankFieldFormat  MediaRankField = "format"
	MediaRankFieldYear    MediaRankField = "year"
	MediaRankFieldSeason  MediaRankField = "season"
	MediaRankFieldAllTime MediaRankField = "allTime"
	MediaRankFieldContext MediaRankField = "context"
)

// MediaFieldRankings to generate media rankings query.
func MediaFieldRankings(field MediaRankField, fields ...MediaRankField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("rankings", nil, str...))
}

// MediaListField is fields for media list.
type MediaListField string

// Fields for MediaListField.
const (
	MediaListFieldID                    MediaListField = "id"
	MediaListFieldUserID                MediaListField = "userId"
	MediaListFieldMediaID               MediaListField = "mediaId"
	MediaListFieldStatus                MediaListField = "status"
	MediaListFieldScore                 MediaListField = "score" // TODO: with param
	MediaListFieldProgress              MediaListField = "progress"
	MediaListFieldProgressVolumes       MediaListField = "progressVolumes"
	MediaListFieldRepeat                MediaListField = "repeat"
	MediaListFieldPriority              MediaListField = "priority"
	MediaListFieldPrivate               MediaListField = "private"
	MediaListFieldNotes                 MediaListField = "notes"
	MediaListFieldHiddenFromStatusLists MediaListField = "hiddenFromStatusLists"
	MediaListFieldCustomLists           MediaListField = "customLists"
	MediaListFieldCustomListsAsArray    MediaListField = "customLists(asArray:true)"
	MediaListFieldAdvancedScores        MediaListField = "advancedScores"
	MediaListFieldStartedAt             MediaListField = "startedAt{year month day}"
	MediaListFieldCompletedAt           MediaListField = "completedAt{year month day}"
	MediaListFieldUpdatedAt             MediaListField = "updatedAt"
	MediaListFieldCreatedAt             MediaListField = "createdAt"
)

// MediaFieldMediaListEntry to generate media's media list entry query.
func MediaFieldMediaListEntry(field MediaListField, fields ...MediaListField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("mediaListEntry", nil, str...))
}

// MediaListFieldMedia to generate media list media query.
func MediaListFieldMedia(field MediaField, fields ...MediaField) MediaListField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaListField(FieldObject("media", nil, str...))
}

// MediaListFieldUser to generate media list user query.
func MediaListFieldUser(field UserField, fields ...UserField) MediaListField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaListField(FieldObject("user", nil, str...))
}

// ReviewConnectionField is fields for review connection.
type ReviewConnectionField string

// MediaFieldReviews to generate media reviews query.
func MediaFieldReviews(param MediaParamReviews, field ReviewConnectionField, fields ...ReviewConnectionField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("reviews", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
		"limit":   param.Limit,
		"sort":    param.Sort,
	}, str...))
}

// ReviewEdgeField is fields for review edge.
type ReviewEdgeField string

// ReviewConnectionFieldEdges to generate review connection edges query.
func ReviewConnectionFieldEdges(field ReviewEdgeField, fields ...ReviewEdgeField) ReviewConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return ReviewConnectionField(FieldObject("edges", nil, str...))
}

// ReviewEdgeFieldNode to generate review edge node query.
func ReviewEdgeFieldNode(field ReviewField, fields ...ReviewField) ReviewEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return ReviewEdgeField(FieldObject("node", nil, str...))
}

// ReviewField is fields for review.
type ReviewField string

// Fields for ReviewField.
const (
	ReviewFieldID           ReviewField = "id"
	ReviewFieldUserID       ReviewField = "userId"
	ReviewFieldMediaID      ReviewField = "mediaId"
	ReviewFieldMediaType    ReviewField = "mediaType"
	ReviewFieldSummary      ReviewField = "summary"
	ReviewFieldBody         ReviewField = "body"
	ReviewFieldBodyAsHTML   ReviewField = "body(asHtml:true)"
	ReviewFieldRating       ReviewField = "rating"
	ReviewFieldRatingAmount ReviewField = "ratingAmount"
	ReviewFieldUserRating   ReviewField = "userRating"
	ReviewFieldScore        ReviewField = "score"
	ReviewFieldPrivate      ReviewField = "private"
	ReviewFieldSiteURL      ReviewField = "siteUrl"
	ReviewFieldCreatedAt    ReviewField = "createdAt"
	ReviewFieldUpdatedAt    ReviewField = "updatedAt"
)

// ReviewConnectionFieldNodes to generate review connection nodes query.
func ReviewConnectionFieldNodes(field ReviewField, fields ...ReviewField) ReviewConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return ReviewConnectionField(FieldObject("nodes", nil, str...))
}

// ReviewConnectionFieldPageInfo to generate review connection page info query.
func ReviewConnectionFieldPageInfo(field PageInfoField, fields ...PageInfoField) ReviewConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return ReviewConnectionField(FieldObject("pageInfo", nil, str...))
}

// ReviewFieldUser to generate review user query.
func ReviewFieldUser(field UserField, fields ...UserField) ReviewField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return ReviewField(FieldObject("user", nil, str...))
}

// ReviewFieldMedia to generate review media query.
func ReviewFieldMedia(field MediaField, fields ...MediaField) ReviewField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return ReviewField(FieldObject("media", nil, str...))
}

// RecommendationConnectionField is fields for recommendation connection.
type RecommendationConnectionField string

// MediaFieldRecommendations to generate media recommendations query.
func MediaFieldRecommendations(param MediaParamRecommendations, field RecommendationConnectionField, fields ...RecommendationConnectionField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("recommendations", QueryParam{
		"page":    param.Page,
		"perPage": param.PerPage,
		"sort":    param.Sort,
	}, str...))
}

// RecommendationEdgeField is fields for recommendatino edge.
type RecommendationEdgeField string

// RecommendationConnectionFieldEdges to generate recommendation connection edges query.
func RecommendationConnectionFieldEdges(field RecommendationEdgeField, fields ...RecommendationEdgeField) RecommendationConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return RecommendationConnectionField(FieldObject("edges", nil, str...))
}

// RecommendationEdgeFieldNode to generate recommendation edge node query.
func RecommendationEdgeFieldNode(field RecommendationField, fields ...RecommendationField) RecommendationEdgeField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return RecommendationEdgeField(FieldObject("node", nil, str...))
}

// RecommendationField is fields for recommendation.
type RecommendationField string

// Fields for RecommendationField.
const (
	RecommendationFieldID         RecommendationField = "id"
	RecommendationFieldRating     RecommendationField = "rating"
	RecommendationFieldUserRating RecommendationField = "userRating"
)

// RecommendationConnectionFieldNodes to generate recommendation connection nodes query.
func RecommendationConnectionFieldNodes(field RecommendationField, fields ...RecommendationField) RecommendationConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return RecommendationConnectionField(FieldObject("nodes", nil, str...))
}

// RecommendationFieldMedia to generate recommendation media query.
func RecommendationFieldMedia(field MediaField, fields ...MediaField) RecommendationField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return RecommendationField(FieldObject("media", nil, str...))
}

// RecommendationFieldMediaRecommendation to generate recommendation media recommendation query.
func RecommendationFieldMediaRecommendation(field MediaField, fields ...MediaField) RecommendationField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return RecommendationField(FieldObject("mediaRecommendation", nil, str...))
}

// RecommendationFieldUser to generate recommendation user query.
func RecommendationFieldUser(field UserField, fields ...UserField) RecommendationField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return RecommendationField(FieldObject("user", nil, str...))
}

// RecommendationConnectionFieldPageInfo to generate recommendation connection page info query.
func RecommendationConnectionFieldPageInfo(field PageInfoField, fields ...PageInfoField) RecommendationConnectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return RecommendationConnectionField(FieldObject("pageInfo", nil, str...))
}

// MediaStatsField is fields for media stats.
type MediaStatsField string

// Fields for MediaStatsField.
const (
	MediaStatsFieldScoreDistribution  MediaStatsField = "scoreDistribution{score amount}"
	MediaStatsFieldStatusDistribution MediaStatsField = "statusDistribution{status amount}"
)

// MediaFieldStats to generate media stats query.
func MediaFieldStats(field MediaStatsField, fields ...MediaStatsField) MediaField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaField(FieldObject("stats", nil, str...))
}

// PageField is fields for page.
type PageField string

// PageFieldPageInfo to generate page's page info query.
func PageFieldPageInfo(field PageInfoField, fields ...PageInfoField) PageField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return PageField(FieldObject("pageInfo", nil, str...))
}

// PageFieldStudios to generate page studios query.
func PageFieldStudios(param PageParamStudios, field StudioField, fields ...StudioField) PageField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return PageField(FieldObject("studios", QueryParam{
		"search":    toQueryString(param.Search),
		"id":        param.ID,
		"id_not":    param.IDNot,
		"id_in":     param.IDIn,
		"id_not_in": param.IDNotIn,
		"sort":      param.Sort,
	}, str...))
}

// PageFieldMedia to generate page media query.
func PageFieldMedia(param PageParamMedia, field MediaField, fields ...MediaField) PageField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return PageField(FieldObject("media", QueryParam{
		"id":                   param.ID,
		"idMal":                param.IDMAL,
		"startDate":            param.StartDate,
		"endDate":              param.EndDate,
		"season":               param.Season,
		"seasonYear":           param.SeasonYear,
		"type":                 param.Type,
		"format":               param.Format,
		"status":               param.Status,
		"episodes":             param.Episodes,
		"duration":             param.Duration,
		"chapters":             param.Chapters,
		"volumes":              param.Volumes,
		"isAdult":              param.IsAdult,
		"genre":                toQueryString(param.Genre),
		"tag":                  toQueryString(param.Tag),
		"minimumTagRank":       param.MinimumTagRank,
		"tagCategory":          toQueryString(param.TagCategory),
		"licensedBy":           toQueryString(param.LicensedBy),
		"averageScore":         param.AverageScore,
		"popularity":           param.Popularity,
		"source":               param.Source,
		"countryOfOrigin":      toQueryString(param.CountryOfOrigin),
		"search":               toQueryString(param.Search),
		"id_not":               param.IDNot,
		"id_in":                param.IDIn,
		"id_not_in":            param.IDNotIn,
		"id_mal_not":           param.IDMALNot,
		"id_mal_in":            param.IDMALIn,
		"id_mal_not_in":        param.IDMALNotIn,
		"startDate_greater":    param.StartDateGreater,
		"startDate_lesser":     param.StartDateLesser,
		"startDate_like":       toQueryString(param.StartDateLike),
		"endDate_greater":      param.EndDateGreater,
		"endDate_lesser":       param.EndDateLesser,
		"endDate_like":         toQueryString(param.EndDateLike),
		"format_in":            param.FormatIn,
		"format_not":           param.FormatNot,
		"format_not_in":        param.FormatNotIn,
		"status_in":            param.StatusIn,
		"status_not":           param.StatusNot,
		"status_not_in":        param.StatusNotIn,
		"episodes_greater":     param.EpisodesGreater,
		"episodes_lesser":      param.EpisodesLesser,
		"duration_greater":     param.DurationGreater,
		"duration_lesser":      param.DurationLesser,
		"chapters_greater":     param.ChaptersGreater,
		"chapters_lesser":      param.ChaptersLesser,
		"volumes_greater":      param.VolumesGreater,
		"volumes_lesser":       param.VolumesLesser,
		"genre_in":             toQueryStringArray(param.GenreIn),
		"genre_not_in":         toQueryStringArray(param.GenreNotIn),
		"tag_in":               toQueryStringArray(param.TagIn),
		"tag_not_in":           toQueryStringArray(param.TagNotIn),
		"tagCategory_in":       toQueryStringArray(param.TagCategoryIn),
		"tagCategory_not_in":   toQueryStringArray(param.TagCategoryNotIn),
		"licensedBy_in":        toQueryStringArray(param.LicensedByIn),
		"averageScore_not":     param.AverageScoreNot,
		"averageScore_greater": param.AverageScoreGreater,
		"averageScore_lesser":  param.AverageScoreLesser,
		"popularity_not":       param.PopularityNot,
		"popularity_greater":   param.PopularityGreater,
		"popularity_lesser":    param.PopularityLesser,
		"source_in":            param.SourceIn,
		"sort":                 param.Sort,
		"onList":               param.OnList,
	}, str...))
}

// PageFieldCharacters to generate page character query.
func PageFieldCharacters(param PageParamCharacters, field CharacterField, fields ...CharacterField) PageField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return PageField(FieldObject("characters", QueryParam{
		"search":     toQueryString(param.Search),
		"id":         param.ID,
		"isBirthday": param.IsBirthday,
		"id_not":     param.IDNot,
		"id_in":      param.IDIn,
		"id_not_in":  param.IDNotIn,
		"sort":       param.Sort,
	}, str...))
}

// PageFieldStaff to generate page staff query.
func PageFieldStaff(param PageParamStaff, field StaffField, fields ...StaffField) PageField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return PageField(FieldObject("staff", QueryParam{
		"search":     toQueryString(param.Search),
		"id":         param.ID,
		"isBirthday": param.IsBirthday,
		"id_not":     param.IDNot,
		"id_in":      param.IDIn,
		"id_not_in":  param.IDNotIn,
		"sort":       param.Sort,
	}, str...))
}

// MediaListCollectionField is fields for media list collection.
type MediaListCollectionField string

// Fields for MediaListCollectionField.
const (
	MediaListCollectionFieldHasNextChunk MediaListCollectionField = "hasNextChunk"
)

// MediaListGroupField is fields for media list group.
type MediaListGroupField string

// Fields for MediaListGroupField.
const (
	MediaListGroupFieldName                 MediaListGroupField = "name"
	MediaListGroupFieldIsCustomList         MediaListGroupField = "isCustomList"
	MediaListGroupFieldIsSplitCompletedList MediaListGroupField = "isSplitCompletedList"
	MediaListGroupFieldStatus               MediaListGroupField = "status"
)

// MediaListGroupFieldEntries to generate media list group entries query.
func MediaListGroupFieldEntries(field MediaListField, fields ...MediaListField) MediaListGroupField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaListGroupField(FieldObject("entries", nil, str...))
}

// MediaListCollectionFieldLists to generate media list collection lists query.
func MediaListCollectionFieldLists(field MediaListGroupField, fields ...MediaListGroupField) MediaListCollectionField {
	str := []string{string(field)}
	for _, f := range fields {
		str = append(str, string(f))
	}
	return MediaListCollectionField(FieldObject("lists", nil, str...))
}
