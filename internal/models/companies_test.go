package models

import (
	"testing"
	"time"
)

func TestSocialsValidate(t *testing.T) {
	tests := []struct {
		name    string
		socials Socials
		wantErr bool
	}{
		{"nil map", nil, false},
		{"empty map", Socials{}, false},
		{"all supported", Socials{SocialVK: "a", SocialTelegram: "b", SocialWhatsApp: "c", SocialViber: "d"}, false},
		{"unknown network", Socials{"myspace": "tom"}, true},
		{"mixed", Socials{SocialVK: "a", "icq": "123"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.socials.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompanyLocationDefaultsToUTC(t *testing.T) {
	loc, err := Company{}.Location()
	if err != nil {
		t.Fatalf("Location: %v", err)
	}

	if loc != time.UTC {
		t.Errorf("got %v, want UTC", loc)
	}
}

func TestCompanyLocationLoadsIANAName(t *testing.T) {
	loc, err := Company{Timezone: "Asia/Yekaterinburg"}.Location()
	if err != nil {
		t.Fatalf("Location: %v", err)
	}

	_, offset := time.Date(2026, 10, 20, 12, 0, 0, 0, loc).Zone()
	if offset != 5*3600 {
		t.Errorf("offset = %d, want 18000", offset)
	}
}

func TestCompanyLocationRejectsUnknownName(t *testing.T) {
	if _, err := (Company{Timezone: "Mars/Olympus"}).Location(); err == nil {
		t.Fatal("expected an error for an unknown timezone")
	}
}
