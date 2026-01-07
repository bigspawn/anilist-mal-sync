package verniy

import "context"

type genreResponse struct {
	Data struct {
		Genres []string `json:"genreCollection"`
	} `json:"data"`
}

// GetGenres to get all genre list.
func (c *Client) GetGenres() ([]string, error) {
	return c.GetGenresWithContext(context.Background())
}

// GetGenresWithContext to get all genre list with context.
func (c *Client) GetGenresWithContext(ctx context.Context) ([]string, error) {
	query := FieldObject("query", nil, "GenreCollection")

	var d genreResponse
	if err := c.post(ctx, query, nil, &d); err != nil {
		return nil, err
	}

	return d.Data.Genres, nil
}
