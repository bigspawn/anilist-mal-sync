package verniy

import (
	"context"
)

type userResponse struct {
	Data struct {
		User User `json:"user"`
	} `json:"data"`
}

func (c *Client) userQuery(params QueryParam, fields ...UserField) string {
	p := make([]string, len(fields))
	for i := range fields {
		p[i] = string(fields[i])
	}
	return FieldObject("User", params, p...)
}

// GetUser to get user data.
func (c *Client) GetUser(username string, fields ...UserField) (*User, error) {
	return c.GetUserWithContext(context.Background(), username, fields...)
}

// GetUserWithContext to get user data with context.
func (c *Client) GetUserWithContext(ctx context.Context, username string, fields ...UserField) (*User, error) {
	if len(fields) == 0 {
		fields = []UserField{
			UserFieldID,
			UserFieldName,
			UserFieldAbout,
			UserFieldAvatar(UserAvatarFieldLarge),
			UserFieldBannerImage,
			UserFieldStatistics(
				UserStatisticTypesFieldAnime(
					UserStatisticsFieldCount,
					UserStatisticsFieldMeanScore,
					UserStatisticsFieldStandardDeviation,
					UserStatisticsFieldMinutesWatched,
					UserStatisticsFieldEpisodesWatched),
				UserStatisticTypesFieldManga(
					UserStatisticsFieldCount,
					UserStatisticsFieldMeanScore,
					UserStatisticsFieldStandardDeviation,
					UserStatisticsFieldChaptersRead,
					UserStatisticsFieldVolumesRead)),
		}
	}

	query := FieldObject("query", QueryParam{
		"$name": "String",
	}, c.userQuery(QueryParam{
		"name": "$name",
	}, fields...))

	var d userResponse
	err := c.post(ctx, query, queryVariable{
		"name": username,
	}, &d)
	if err != nil {
		return nil, err
	}

	return &d.Data.User, nil
}

// GetUserFavouriteAnime to get user's favourite anime.
func (c *Client) GetUserFavouriteAnime(username string, page int, perPage int) (*User, error) {
	return c.GetUserFavouriteAnimeWithContext(context.Background(), username, page, perPage)
}

// GetUserFavouriteAnimeWithContext to get user's favourite anime with context.
func (c *Client) GetUserFavouriteAnimeWithContext(ctx context.Context, username string, page int, perPage int) (*User, error) {
	return c.GetUserWithContext(ctx, username, UserFieldFavourites(UserParamFavourites{
		Page: page,
	}, FavouritesFieldAnime(FavouritesParamAnime{
		Page:    page,
		PerPage: perPage,
	},
		MediaConnectionFieldPageInfo(
			PageInfoFieldTotal,
			PageInfoFieldPerPage,
			PageInfoFieldCurrentPage,
			PageInfoFieldLastPage,
			PageInfoFieldHasNextPage),
		MediaConnectionFieldEdges(
			MediaEdgeFieldFavouriteOrder,
			MediaEdgeFieldNode(
				MediaFieldID,
				MediaFieldIsAdult,
				MediaFieldTitle(
					MediaTitleFieldRomaji,
					MediaTitleFieldEnglish,
					MediaTitleFieldNative),
				MediaFieldCoverImage(
					MediaCoverImageFieldLarge,
					MediaCoverImageFieldLarge))))))
}

// GetUserFavouriteManga to get user's favourite manga.
func (c *Client) GetUserFavouriteManga(username string, page int, perPage int) (*User, error) {
	return c.GetUserFavouriteMangaWithContext(context.Background(), username, page, perPage)
}

// GetUserFavouriteMangaWithContext to get user's favourite manga with context.
func (c *Client) GetUserFavouriteMangaWithContext(ctx context.Context, username string, page int, perPage int) (*User, error) {
	return c.GetUserWithContext(ctx, username, UserFieldFavourites(UserParamFavourites{
		Page: page,
	}, FavouritesFieldManga(FavouritesParamManga{
		Page:    page,
		PerPage: perPage,
	},
		MediaConnectionFieldPageInfo(
			PageInfoFieldTotal,
			PageInfoFieldPerPage,
			PageInfoFieldCurrentPage,
			PageInfoFieldLastPage,
			PageInfoFieldHasNextPage),
		MediaConnectionFieldEdges(
			MediaEdgeFieldFavouriteOrder,
			MediaEdgeFieldNode(
				MediaFieldID,
				MediaFieldIsAdult,
				MediaFieldTitle(
					MediaTitleFieldRomaji,
					MediaTitleFieldEnglish,
					MediaTitleFieldNative),
				MediaFieldCoverImage(
					MediaCoverImageFieldLarge,
					MediaCoverImageFieldLarge))))))
}

// GetUserFavouriteCharacters to get user's favourite characters.
func (c *Client) GetUserFavouriteCharacters(username string, page int, perPage int) (*User, error) {
	return c.GetUserFavouriteCharactersWithContext(context.Background(), username, page, perPage)
}

// GetUserFavouriteCharactersWithContext to get user's favourite characters with context.
func (c *Client) GetUserFavouriteCharactersWithContext(ctx context.Context, username string, page int, perPage int) (*User, error) {
	return c.GetUserWithContext(ctx, username, UserFieldFavourites(UserParamFavourites{
		Page: page,
	}, FavouritesFieldCharacters(FavouritesParamCharacters{
		Page:    page,
		PerPage: perPage,
	},
		CharacterConnectionFieldPageInfo(
			PageInfoFieldTotal,
			PageInfoFieldPerPage,
			PageInfoFieldCurrentPage,
			PageInfoFieldLastPage,
			PageInfoFieldHasNextPage),
		CharacterConnectionFieldEdges(
			CharacterEdgeFieldFavouriteOrder,
			CharacterEdgeFieldNode(
				CharacterFieldID,
				CharacterFieldName(
					CharacterNameFieldFull,
					CharacterNameFieldNative),
				CharacterFieldImage(CharacterImageFieldLarge))))))
}

// GetUserFavouriteStaff to get user's favourite staff.
func (c *Client) GetUserFavouriteStaff(username string, page int, perPage int) (*User, error) {
	return c.GetUserFavouriteStaffWithContext(context.Background(), username, page, perPage)
}

// GetUserFavouriteStaffWithContext to get user's favourite staff with context.
func (c *Client) GetUserFavouriteStaffWithContext(ctx context.Context, username string, page int, perPage int) (*User, error) {
	return c.GetUserWithContext(ctx, username, UserFieldFavourites(UserParamFavourites{
		Page: page,
	}, FavouritesFieldStaff(FavouritesParamStaff{
		Page:    page,
		PerPage: perPage,
	},
		StaffConnectionFieldPageInfo(
			PageInfoFieldTotal,
			PageInfoFieldPerPage,
			PageInfoFieldCurrentPage,
			PageInfoFieldLastPage,
			PageInfoFieldHasNextPage),
		StaffConnectionFieldEdges(
			StaffEdgeFieldFavouriteOrder,
			StaffEdgeFieldNode(
				StaffFieldID,
				StaffFieldName(
					StaffNameFieldFull,
					StaffNameFieldNative),
				StaffFieldImage(StaffImageFieldLarge))))))
}

// GetUserFavouriteStudios to get user's favourite studios.
func (c *Client) GetUserFavouriteStudios(username string, page int, perPage int) (*User, error) {
	return c.GetUserFavouriteStudiosWithContext(context.Background(), username, page, perPage)
}

// GetUserFavouriteStudiosWithContext to get user's favourite studios with context.
func (c *Client) GetUserFavouriteStudiosWithContext(ctx context.Context, username string, page int, perPage int) (*User, error) {
	return c.GetUserWithContext(ctx, username, UserFieldFavourites(UserParamFavourites{
		Page: page,
	}, FavouritesFieldStudios(FavouritesParamStudios{
		Page:    page,
		PerPage: perPage,
	},
		StudioConnectionFieldPageInfo(
			PageInfoFieldTotal,
			PageInfoFieldPerPage,
			PageInfoFieldCurrentPage,
			PageInfoFieldLastPage,
			PageInfoFieldHasNextPage),
		StudioConnectionFieldEdges(
			StudioEdgeFieldFavouriteOrder,
			StudioEdgeFieldNode(
				StudioFieldID,
				StudioFieldName)))))
}

type userAnimeListResponse struct {
	Data struct {
		MediaListCollection *MediaListCollection `json:"mediaListCollection"`
	} `json:"data"`
}

func (c *Client) mediaListCollectionQuery(params QueryParam, fields ...MediaListCollectionField) string {
	p := make([]string, len(fields))
	for i := range fields {
		p[i] = string(fields[i])
	}
	return FieldObject("MediaListCollection", params, p...)
}

func (c *Client) getMediaListCollection(ctx context.Context, username string, mediaType MediaType, fields ...MediaListGroupField) ([]MediaListGroup, error) {
	query := FieldObject("query", QueryParam{
		"$username": "String",
		"$type":     "MediaType",
	}, c.mediaListCollectionQuery(QueryParam{
		"userName": "$username",
		"type":     "$type",
	}, MediaListCollectionFieldLists(MediaListGroupFieldStatus, fields...)))

	var d userAnimeListResponse
	err := c.post(ctx, query, queryVariable{
		"username": username,
		"type":     mediaType,
	}, &d)
	if err != nil {
		return nil, err
	}

	return d.Data.MediaListCollection.Lists, nil
}

// GetUserAnimeList to get user's anime list.
func (c *Client) GetUserAnimeList(username string, fields ...MediaListGroupField) ([]MediaListGroup, error) {
	return c.GetUserAnimeListWithContext(context.Background(), username, fields...)
}

// GetUserAnimeListWithContext to get user's anime list with context.
func (c *Client) GetUserAnimeListWithContext(ctx context.Context, username string, fields ...MediaListGroupField) ([]MediaListGroup, error) {
	if len(fields) == 0 {
		fields = []MediaListGroupField{
			MediaListGroupFieldName,
			MediaListGroupFieldStatus,
			MediaListGroupFieldEntries(
				MediaListFieldID,
				MediaListFieldStatus,
				MediaListFieldScore,
				MediaListFieldProgress,
				MediaListFieldNotes,
				MediaListFieldStartedAt,
				MediaListFieldCompletedAt,
				MediaListFieldMedia(
					MediaFieldID,
					MediaFieldTitle(
						MediaTitleFieldRomaji,
						MediaTitleFieldEnglish,
						MediaTitleFieldNative),
					MediaFieldType,
					MediaFieldFormat,
					MediaFieldStatusV2,
					MediaFieldCoverImage(MediaCoverImageFieldLarge),
					MediaFieldAverageScore,
					MediaFieldPopularity,
					MediaFieldIsAdult,
					MediaFieldEpisodes)),
		}
	}
	return c.getMediaListCollection(ctx, username, MediaTypeAnime, fields...)
}

// GetUserMangaList to get user's manga list.
func (c *Client) GetUserMangaList(username string, fields ...MediaListGroupField) ([]MediaListGroup, error) {
	return c.GetUserMangaListWithContext(context.Background(), username, fields...)
}

// GetUserMangaListWithContext to get user's manga list with context.
func (c *Client) GetUserMangaListWithContext(ctx context.Context, username string, fields ...MediaListGroupField) ([]MediaListGroup, error) {
	if len(fields) == 0 {
		fields = []MediaListGroupField{
			MediaListGroupFieldName,
			MediaListGroupFieldStatus,
			MediaListGroupFieldEntries(
				MediaListFieldID,
				MediaListFieldStatus,
				MediaListFieldScore,
				MediaListFieldProgress,
				MediaListFieldProgressVolumes,
				MediaListFieldNotes,
				MediaListFieldStartedAt,
				MediaListFieldCompletedAt,
				MediaListFieldMedia(
					MediaFieldID,
					MediaFieldTitle(
						MediaTitleFieldRomaji,
						MediaTitleFieldEnglish,
						MediaTitleFieldNative),
					MediaFieldType,
					MediaFieldFormat,
					MediaFieldStatusV2,
					MediaFieldCoverImage(MediaCoverImageFieldLarge),
					MediaFieldAverageScore,
					MediaFieldPopularity,
					MediaFieldIsAdult,
					MediaFieldChapters,
					MediaFieldVolumes)),
		}
	}
	return c.getMediaListCollection(ctx, username, MediaTypeManga, fields...)
}
