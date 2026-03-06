package models

type SocialNetwork string

const (
	SocialVK       SocialNetwork = "vk"
	SocialTelegram SocialNetwork = "telegram"
	SocialWhatsApp SocialNetwork = "whatsapp"
	SocialViber    SocialNetwork = "viber"
)

type Socials map[SocialNetwork]string
