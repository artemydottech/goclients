package repository

import (
	"reflect"
	"testing"

	"github.com/artemydottech/goclients/internal/models"
)

func TestPaginate(t *testing.T) {
	tests := []struct {
		name      string
		page      models.Page
		args      []any
		wantQuery string
		wantArgs  []any
	}{
		{"no page", models.Page{}, []any{1}, "SELECT 1", []any{1}},
		{"limit", models.Page{Limit: 10}, nil, "SELECT 1 LIMIT ? OFFSET ?", []any{10, 0}},
		{"limit and offset keep earlier args", models.Page{Limit: 10, Offset: 30}, []any{7}, "SELECT 1 LIMIT ? OFFSET ?", []any{7, 10, 30}},
		{"offset only", models.Page{Offset: 5}, nil, "SELECT 1 LIMIT -1 OFFSET ?", []any{5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := paginate("SELECT 1", tt.args, tt.page)

			if query != tt.wantQuery {
				t.Errorf("query = %q, want %q", query, tt.wantQuery)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) && !(len(args) == 0 && len(tt.wantArgs) == 0) {
				t.Errorf("args = %v, want %v", args, tt.wantArgs)
			}
		})
	}
}
