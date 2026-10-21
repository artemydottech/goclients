package models

import (
	"errors"
	"testing"
)

func TestPageValidate(t *testing.T) {
	tests := []struct {
		name    string
		page    Page
		wantErr bool
	}{
		{"zero value means everything", Page{}, false},
		{"limit only", Page{Limit: 20}, false},
		{"limit and offset", Page{Limit: 20, Offset: 40}, false},
		{"offset only", Page{Offset: 10}, false},
		{"largest limit", Page{Limit: MaxPageLimit}, false},
		{"limit above the maximum", Page{Limit: MaxPageLimit + 1}, true},
		{"negative limit", Page{Limit: -1}, true},
		{"negative offset", Page{Limit: 10, Offset: -5}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.page.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				var validationErr ValidationError
				if !errors.As(err, &validationErr) {
					t.Errorf("error %T is not a ValidationError", err)
				}
			}
		})
	}
}

func TestPageUnlimited(t *testing.T) {
	if !(Page{}).Unlimited() {
		t.Error("the zero page must be unlimited")
	}
	if (Page{Limit: 1}).Unlimited() || (Page{Offset: 1}).Unlimited() {
		t.Error("a page with a limit or an offset is not unlimited")
	}
}
