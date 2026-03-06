package models

import "fmt"

type Company struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	Geolocation string  `json:"geolocation"`
	Schedule    string  `json:"schedule"`
	Logo        string  `json:"logo"`
	Site        string  `json:"site"`
	Socials     Socials `json:"socials,omitempty"`
}

func (s Socials) Validate() error {
	for network := range s {
		switch network {
		case SocialVK, SocialTelegram, SocialWhatsApp, SocialViber:
		default:
			return fmt.Errorf("Неподдерживаемая соц. сеть %s", network)
		}
	}
	return nil
}
