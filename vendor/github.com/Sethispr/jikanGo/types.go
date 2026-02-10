package jikan

import "strconv"

type ID int

func (id ID) String() string { return strconv.Itoa(int(id)) }

type Resource struct {
	MalID ID     `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

type ImageURL struct {
	Large  string `json:"large_image_url"`
	Medium string `json:"image_url"`
	Small  string `json:"small_image_url"`
}

type ImageSet struct {
	JPG  ImageURL `json:"jpg"`
	WebP ImageURL `json:"webp"`
}

type Title struct {
	Language string `json:"type"`
	Title    string `json:"title"`
}

type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Pagination struct {
	LastPage    int  `json:"last_visible_page"`
	CurrentPage int  `json:"current_page"`
	HasNext     bool `json:"has_next_page"`
	Total       int  `json:"-"`
}

type Statistics struct {
	Days      float64 `json:"days_watched"`
	MeanScore float64 `json:"mean_score"`
	Completed int     `json:"completed"`
	Total     int     `json:"total_entries"`
}
