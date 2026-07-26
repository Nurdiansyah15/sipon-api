package persistence

import (
	"strings"

	"sipon-api/internal/app/dto"
)

// resolvePaginationParams menormalisasi dto.PaginationParams menjadi limit/offset/sort
// siap pakai untuk query listing — dipakai oleh semua query model persistence.
func resolvePaginationParams(params dto.PaginationParams, defaultLimit, maxLimit int, sortable map[string]string, defaultSortColumn, defaultSortType string) (limit int, offset int, currentPage int64, sortColumn string, sortType string) {
	page := 1
	if params.Page != nil && *params.Page > 0 {
		page = *params.Page
	}

	limit = defaultLimit
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset = (page - 1) * limit
	currentPage = int64(page)

	sortColumn = defaultSortColumn
	if params.SortBy != nil {
		if mapped, ok := sortable[strings.ToLower(strings.TrimSpace(*params.SortBy))]; ok {
			sortColumn = mapped
		}
	}

	sortType = strings.ToUpper(strings.TrimSpace(defaultSortType))
	if sortType != "ASC" && sortType != "DESC" {
		sortType = "DESC"
	}
	if params.SortType != nil {
		if strings.EqualFold(strings.TrimSpace(*params.SortType), "asc") {
			sortType = "ASC"
		} else if strings.EqualFold(strings.TrimSpace(*params.SortType), "desc") {
			sortType = "DESC"
		}
	}

	return limit, offset, currentPage, sortColumn, sortType
}
