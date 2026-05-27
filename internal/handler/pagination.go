package handler

import (
	"net/http"
	"strconv"
)

// ParsePagination extracts and validates limit/offset query parameters.
// If values are missing or invalid, defaults are applied.
// limit is clamped to [1, maxLimit]; offset is clamped to >= 0.
func ParsePagination(r *http.Request, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	offset = 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
			if limit > maxLimit {
				limit = maxLimit
			}
			if limit < 1 {
				limit = defaultLimit
			}
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}
