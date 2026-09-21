package models

import "time"

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
	for network := range s {
		switch network {
		case SocialVK, SocialTelegram, SocialWhatsApp, SocialViber:
		default:
			return Invalid("unsupported social network %s", network)
		}
	}
	return nil
}
