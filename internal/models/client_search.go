package models

import (
	"strings"
	"unicode/utf8"
)

type ClientSearch struct {
	CompanyID int
	Tag       string
	Query     string
	Phone     string
	Page      Page
}

func (s ClientSearch) Validate() error {
	if s.CompanyID <= 0 {
		return Invalid("company_id must be positive")
	}
	if utf8.RuneCountInString(s.Query) > 200 {
		return Invalid("query cannot exceed 200 characters")
	}
	if s.Phone != "" && (len(s.Phone) < 10 || len(s.Phone) > 15 || strings.Trim(s.Phone, "0123456789") != "") {
		return Invalid("phone must contain 10 to 15 ASCII digits")
	}
	if s.Tag != "" {
		if _, err := NormalizeClientTags([]string{s.Tag}); err != nil {
			return err
		}
	}
	return s.Page.Validate()
}
