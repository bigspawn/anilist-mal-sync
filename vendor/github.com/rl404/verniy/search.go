package verniy

import "context"

// SearchAnime to search anime.
func (c *Client) SearchAnime(query PageParamMedia, page int, perPage int, fields ...MediaField) (*Page, error) {
	return c.SearchAnimeWithContext(context.Background(), query, page, perPage, fields...)
}

// SearchAnimeWithContext to search anime with context.
func (c *Client) SearchAnimeWithContext(ctx context.Context, query PageParamMedia, page int, perPage int, fields ...MediaField) (*Page, error) {
	query.Type = MediaTypeAnime
	if len(fields) == 0 {
		isMain := true
		fields = []MediaField{
			MediaFieldID,
			MediaFieldTitle(
				MediaTitleFieldEnglish,
				MediaTitleFieldRomaji,
				MediaTitleFieldNative),
			MediaFieldSynonyms,
			MediaFieldCoverImage(
				MediaCoverImageFieldLarge,
				MediaCoverImageFieldColor),
			MediaFieldStartDate,
			MediaFieldEndDate,
			MediaFieldSeason,
			MediaFieldSeasonYear,
			MediaFieldDescription,
			MediaFieldType,
			MediaFieldFormat,
			MediaFieldStatusV2,
			MediaFieldEpisodes,
			MediaFieldDuration,
			MediaFieldGenres,
			MediaFieldIsAdult,
			MediaFieldAverageScore,
			MediaFieldPopularity,
			MediaFieldStudios(
				MediaParamStudios{IsMain: &isMain},
				StudioConnectionFieldEdges(
					StudioEdgeFieldIsMain,
					StudioEdgeFieldNode(
						StudioFieldID,
						StudioFieldName,
						StudioFieldIsAnimationStudio))),
		}
	}
	if len(query.Sort) == 0 {
		query.Sort = []MediaSort{MediaSortPopularityDesc, MediaSortScoreDesc}
	}
	return c.page(ctx, page, perPage, PageFieldMedia(query, "", fields...))
}

// SearchManga to search manga.
func (c *Client) SearchManga(query PageParamMedia, page int, perPage int, fields ...MediaField) (*Page, error) {
	return c.SearchMangaWithContext(context.Background(), query, page, perPage, fields...)
}

// SearchMangaWithContext to search manga with context.
func (c *Client) SearchMangaWithContext(ctx context.Context, query PageParamMedia, page int, perPage int, fields ...MediaField) (*Page, error) {
	query.Type = MediaTypeManga
	if len(fields) == 0 {
		fields = []MediaField{
			MediaFieldID,
			MediaFieldTitle(
				MediaTitleFieldEnglish,
				MediaTitleFieldRomaji,
				MediaTitleFieldNative),
			MediaFieldCoverImage(
				MediaCoverImageFieldLarge,
				MediaCoverImageFieldColor),
			MediaFieldStartDate,
			MediaFieldEndDate,
			MediaFieldDescription,
			MediaFieldType,
			MediaFieldFormat,
			MediaFieldStatusV2,
			MediaFieldChapters,
			MediaFieldVolumes,
			MediaFieldGenres,
			MediaFieldIsAdult,
			MediaFieldAverageScore,
			MediaFieldPopularity,
		}
	}
	if len(query.Sort) == 0 {
		query.Sort = []MediaSort{MediaSortPopularityDesc, MediaSortScoreDesc}
	}
	return c.page(ctx, page, perPage, PageFieldMedia(query, "", fields...))
}

// SearchCharacter to search character.
func (c *Client) SearchCharacter(query PageParamCharacters, page int, perPage int, fields ...CharacterField) (*Page, error) {
	return c.SearchCharacterWithContext(context.Background(), query, page, perPage, fields...)
}

// SearchCharacterWithContext to search character with context.
func (c *Client) SearchCharacterWithContext(ctx context.Context, query PageParamCharacters, page int, perPage int, fields ...CharacterField) (*Page, error) {
	if len(fields) == 0 {
		fields = []CharacterField{
			CharacterFieldID,
			CharacterFieldName(CharacterNameFieldFull),
			CharacterFieldImage(CharacterImageFieldLarge),
		}
	}
	if len(query.Sort) == 0 {
		query.Sort = []CharacterSort{CharacterSortSearchMatch, CharacterSortFavouritesDesc}
	}
	return c.page(ctx, page, perPage, PageFieldCharacters(query, "", fields...))
}

// SearchStaff to search staff.
func (c *Client) SearchStaff(query PageParamStaff, page int, perPage int, fields ...StaffField) (*Page, error) {
	return c.SearchStaffWithContext(context.Background(), query, page, perPage, fields...)
}

// SearchStaffWithContext to search staff with context.
func (c *Client) SearchStaffWithContext(ctx context.Context, query PageParamStaff, page int, perPage int, fields ...StaffField) (*Page, error) {
	if len(fields) == 0 {
		fields = []StaffField{
			StaffFieldID,
			StaffFieldName(StaffNameFieldFull),
			StaffFieldImage(StaffImageFieldLarge),
		}
	}
	if len(query.Sort) == 0 {
		query.Sort = []StaffSort{StaffSortSearchMatch, StaffSortFavouritesDesc}
	}
	return c.page(ctx, page, perPage, PageFieldStaff(query, "", fields...))
}
