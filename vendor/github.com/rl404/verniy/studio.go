package verniy

import "context"

type studioResponse struct {
	Data struct {
		Studio Studio `json:"studio"`
	} `json:"data"`
}

func (c *Client) studioQuery(params QueryParam, fields ...StudioField) string {
	p := make([]string, len(fields))
	for i := range fields {
		p[i] = string(fields[i])
	}
	return FieldObject("Studio", params, p...)
}

// GetStudio to get anime list produced by the studio.
func (c *Client) GetStudio(id int, page int, perPage int, fields ...StudioField) (*Studio, error) {
	return c.GetStudioWithContext(context.Background(), id, page, perPage, fields...)
}

// GetStudioWithContext to get anime list produced by the studio with context.
func (c *Client) GetStudioWithContext(ctx context.Context, id int, page int, perPage int, fields ...StudioField) (*Studio, error) {
	if len(fields) == 0 {
		fields = []StudioField{
			StudioFieldID,
			StudioFieldName,
			StudioFieldIsAnimationStudio,
			StudioFieldFavourites,
			StudioFieldMedia(StudioParamMedia{
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
					MediaEdgeFieldIsMainStudio,
					MediaEdgeFieldNode(
						MediaFieldID,
						MediaFieldTitle(
							MediaTitleFieldEnglish,
							MediaTitleFieldNative,
							MediaTitleFieldRomaji),
						MediaFieldCoverImage(MediaCoverImageFieldMedium),
						MediaFieldStartDate,
						MediaFieldEndDate,
						MediaFieldDescription,
						MediaFieldSeason,
						MediaFieldSeasonYear,
						MediaFieldType,
						MediaFieldFormat,
						MediaFieldStatusV2,
						MediaFieldGenres,
						MediaFieldIsAdult,
						MediaFieldAverageScore,
						MediaFieldPopularity))),
		}
	}

	query := FieldObject("query", QueryParam{
		"$id": "Int",
	}, c.studioQuery(QueryParam{
		"id": "$id",
	}, fields...))

	var d studioResponse
	if err := c.post(ctx, query, QueryParam{
		"id": id,
	}, &d); err != nil {
		return nil, err
	}

	return &d.Data.Studio, nil
}

// GetStudios to get list of studios.
func (c *Client) GetStudios(page int, perPage int, fields ...StudioField) (*Page, error) {
	return c.GetStudiosWithContext(context.Background(), page, perPage, fields...)
}

// GetStudiosWithContext to get list of studios with context.
func (c *Client) GetStudiosWithContext(ctx context.Context, page int, perPage int, fields ...StudioField) (*Page, error) {
	if len(fields) == 0 {
		fields = []StudioField{
			StudioFieldID,
			StudioFieldName,
			StudioFieldIsAnimationStudio,
			StudioFieldFavourites,
		}
	}
	pageFields := PageFieldStudios(PageParamStudios{
		Sort: []StudioSort{StudioSortName},
	}, "", fields...)
	return c.page(ctx, page, perPage, pageFields)
}
