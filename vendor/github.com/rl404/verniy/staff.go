package verniy

import "context"

type staffResponse struct {
	Data struct {
		Staff Staff `json:"staff"`
	} `json:"data"`
}

func (c *Client) staffQuery(params QueryParam, fields ...StaffField) string {
	p := make([]string, len(fields))
	for i := range fields {
		p[i] = string(fields[i])
	}
	return FieldObject("Staff", params, p...)
}

// GetStaff to get staff & voice actor data.
func (c *Client) GetStaff(id int, fields ...StaffField) (*Staff, error) {
	return c.GetStaffWithContext(context.Background(), id, fields...)
}

// GetStaffWithContext to get staff & voice actor data with context.
func (c *Client) GetStaffWithContext(ctx context.Context, id int, fields ...StaffField) (*Staff, error) {
	if len(fields) == 0 {
		fields = []StaffField{
			StaffFieldID,
			StaffFieldName(
				StaffNameFieldFirst,
				StaffNameFieldMiddle,
				StaffNameFieldLast,
				StaffNameFieldFull,
				StaffNameFieldNative,
				StaffNameFieldAlternative),
			StaffFieldImage(StaffImageFieldLarge),
			StaffFieldDescription,
			StaffFieldFavourites,
			StaffFieldAge,
			StaffFieldHomeTown,
			StaffFieldDateOfBirth,
			StaffFieldDateOfDeath,
		}
	}

	query := FieldObject("query", QueryParam{
		"$id": "Int",
	}, c.staffQuery(QueryParam{
		"id": "$id",
	}, fields...))

	var d staffResponse
	if err := c.post(ctx, query, queryVariable{
		"id": id,
	}, &d); err != nil {
		return nil, err
	}

	return &d.Data.Staff, nil
}

// GetStaffCharacters to get list of character the staff play.
func (c *Client) GetStaffCharacters(id int, page int, perPage int) (*Staff, error) {
	return c.GetStaffCharactersWithContext(context.Background(), id, page, perPage)
}

// GetStaffCharactersWithContext to get list of character the staff play with context.
func (c *Client) GetStaffCharactersWithContext(ctx context.Context, id int, page int, perPage int) (*Staff, error) {
	return c.GetStaffWithContext(ctx, id, StaffFieldCharacterMedia(StaffParamCharacterMedia{
		Page:    page,
		PerPage: perPage,
		Sort:    []MediaSort{MediaSortStartDateDesc},
	},
		MediaConnectionFieldPageInfo(
			PageInfoFieldTotal,
			PageInfoFieldPerPage,
			PageInfoFieldCurrentPage,
			PageInfoFieldLastPage,
			PageInfoFieldHasNextPage),
		MediaConnectionFieldEdges(
			MediaEdgeFieldCharacterRole,
			MediaEdgeFieldNode(
				MediaFieldID,
				MediaFieldType,
				MediaFieldTitle(
					MediaTitleFieldEnglish,
					MediaTitleFieldNative,
					MediaTitleFieldRomaji),
				MediaFieldCoverImage(MediaCoverImageFieldMedium)),
			MediaEdgeFieldCharacters(
				CharacterFieldID,
				CharacterFieldName(CharacterNameFieldFull),
				CharacterFieldImage(CharacterImageFieldMedium)))))
}

func (c *Client) getStaffMedia(ctx context.Context, id int, mediaType MediaType, page int, perPage int) (*Staff, error) {
	return c.GetStaffWithContext(ctx, id, StaffFieldStaffMedia(StaffParamStaffMedia{
		Page:    page,
		PerPage: perPage,
		Type:    mediaType,
		Sort:    []MediaSort{MediaSortStartDateDesc},
	},
		MediaConnectionFieldPageInfo(
			PageInfoFieldTotal,
			PageInfoFieldPerPage,
			PageInfoFieldCurrentPage,
			PageInfoFieldLastPage,
			PageInfoFieldHasNextPage),
		MediaConnectionFieldEdges(
			MediaEdgeFieldStaffRole,
			MediaEdgeFieldNode(
				MediaFieldID,
				MediaFieldType,
				MediaFieldTitle(
					MediaTitleFieldEnglish,
					MediaTitleFieldNative,
					MediaTitleFieldRomaji),
				MediaFieldCoverImage(MediaCoverImageFieldMedium)))))
}

// GetStaffAnime to get staff anime list.
func (c *Client) GetStaffAnime(id int, page int, perPage int) (*Staff, error) {
	return c.GetStaffAnimeWithContext(context.Background(), id, page, perPage)
}

// GetStaffAnimeWithContext to get staff anime list with context.
func (c *Client) GetStaffAnimeWithContext(ctx context.Context, id int, page int, perPage int) (*Staff, error) {
	return c.getStaffMedia(ctx, id, MediaTypeAnime, page, perPage)
}

// GetStaffManga to get staff manga list.
func (c *Client) GetStaffManga(id int, page int, perPage int) (*Staff, error) {
	return c.GetStaffMangaWithContext(context.Background(), id, page, perPage)
}

// GetStaffMangaWithContext to get staff manga list with context.
func (c *Client) GetStaffMangaWithContext(ctx context.Context, id int, page int, perPage int) (*Staff, error) {
	return c.getStaffMedia(ctx, id, MediaTypeManga, page, perPage)
}
