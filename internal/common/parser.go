package common

import (
	"net/http"
	"strconv"
	"strings"
)

func ParseFilter(r *http.Request) Filter {
	filter := Filter{
		Page:   1,
		Limit:  10,
		Search: "",
	}

	query := r.URL.Query()
	if query.Get("page") != "" {
		if pageInt, err := strconv.Atoi(query.Get("page")); err == nil && pageInt > 0 {
			filter.Page = pageInt
		}
	}

	if query.Get("limit") != "" {
		if limitInt, err := strconv.Atoi(query.Get("limit")); err == nil && limitInt > 0 {
			filter.Limit = limitInt
		}
	}

	if query.Get("search") != "" {
		filter.Search = strings.ToLower(query.Get("search"))
	}
	return filter
}
