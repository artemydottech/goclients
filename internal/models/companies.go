package models

import (
	"time"
	"unicode/utf8"
)

type Company struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	Geolocation string  `json:"geolocation"`
	Schedule    string  `json:"schedule"`
	Logo        string  `json:"logo"`
	Site        string  `json:"site"`
	Socials     Socials `json:"socials,omitempty"`
	Timezone    string  `json:"timezone"`
}

func (c Company) Location() (*time.Location, error) {
	if c.Timezone == "" {
		return time.UTC, nil
	}

	return time.LoadLocation(c.Timezone)
}

func (s Socials) Validate() error {
	for network, value := range s {
		if utf8.RuneCountInString(value) > 500 {
			return Invalid("social contact %s cannot exceed 500 characters", network)
		}
		switch network {
		case SocialVK, SocialTelegram, SocialWhatsApp, SocialViber:
		default:
			return Invalid("unsupported social network %s", network)
		}
	}
	return nil
}
