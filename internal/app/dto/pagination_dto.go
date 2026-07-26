package dto

type PaginationParams struct {
	Page     *int    `form:"page"`
	Limit    *int    `form:"limit"`
	SortBy   *string `form:"sort_by"`
	SortType *string `form:"sort_type"`
}

type Meta struct {
	CurrentPage int64 `json:"current_page"`
	PerPage     int64 `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int64 `json:"total_pages"`
}
