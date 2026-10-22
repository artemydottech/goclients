package handlers

import (
	"net/http"
	"strconv"

	"github.com/artemydottech/goclients/internal/models"
)

func parsePage(w http.ResponseWriter, r *http.Request) (models.Page, bool) {
	var page models.Page
	query := r.URL.Query()

	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "limit must be a number")
			return models.Page{}, false
		}
		page.Limit = limit
	}

	if raw := query.Get("offset"); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "offset must be a number")
			return models.Page{}, false
		}
		page.Offset = offset
	}

	if err := page.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return models.Page{}, false
	}

	return page, true
}
