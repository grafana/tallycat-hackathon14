package api

import (
	"net/http"
	"strconv"

	"github.com/tallycat/tallycat/internal/repository/query"
)

func ParseListQueryParams(r *http.Request) query.ListQueryParams {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	return query.ListQueryParams{
		FilterType: q.Get("type"),
		Search:     q.Get("search"),
		Page:       page,
		PageSize:   pageSize,
	}
}
