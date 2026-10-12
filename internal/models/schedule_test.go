package models

import (
	"errors"
	"testing"
)

func TestParseDayTime(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"00:00", 0, false},
		{"09:30", 570, false},
		{"23:59", 1439, false},
		{"9:05", 545, false},
		{"24:00", 0, true},
		{"12:60", 0, true},
		{"-1:00", 0, true},
		{"noon", 0, true},
		{"10", 0, true},
		{"10:00:00", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseDayTime(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFormatDayTimeRoundTrip(t *testing.T) {
	for _, minutes := range []int{0, 5, 60, 570, 1439} {
		parsed, err := ParseDayTime(FormatDayTime(minutes))
		if err != nil {
			t.Fatalf("ParseDayTime(%d): %v", minutes, err)
		}
		if parsed != minutes {
			t.Errorf("round trip %d -> %d", minutes, parsed)
		}
	}
}

func TestFormatDayTimePadsWithZeros(t *testing.T) {
	if got := FormatDayTime(65); got != "01:05" {
		t.Errorf("got %q, want 01:05", got)
	}
}

func TestWorkingDayHasBreak(t *testing.T) {
	if (WorkingDay{}).HasBreak() {
		t.Error("an empty day has no break")
	}
	if !(WorkingDay{BreakStartsAt: "13:00"}).HasBreak() {
		t.Error("a day with a break start has a break")
	}
}

func TestWorkingDayValidate(t *testing.T) {
	valid := WorkingDay{Weekday: 1, StartsAt: "09:00", EndsAt: "18:00"}

	tests := []struct {
		name    string
		mutate  func(d *WorkingDay)
		wantErr bool
	}{
		{"plain day", func(d *WorkingDay) {}, false},
		{"sunday", func(d *WorkingDay) { d.Weekday = 0 }, false},
		{"with lunch", func(d *WorkingDay) { d.BreakStartsAt, d.BreakEndsAt = "13:00", "14:00" }, false},
		{"weekday too small", func(d *WorkingDay) { d.Weekday = -1 }, true},
		{"weekday too large", func(d *WorkingDay) { d.Weekday = 7 }, true},
		{"bad start", func(d *WorkingDay) { d.StartsAt = "nine" }, true},
		{"bad end", func(d *WorkingDay) { d.EndsAt = "25:00" }, true},
		{"ends before start", func(d *WorkingDay) { d.StartsAt, d.EndsAt = "18:00", "09:00" }, true},
		{"zero length", func(d *WorkingDay) { d.StartsAt, d.EndsAt = "09:00", "09:00" }, true},
		{"break without end", func(d *WorkingDay) { d.BreakStartsAt = "13:00" }, true},
		{"break without start", func(d *WorkingDay) { d.BreakEndsAt = "14:00" }, true},
		{"bad break start", func(d *WorkingDay) { d.BreakStartsAt, d.BreakEndsAt = "lunch", "14:00" }, true},
		{"bad break end", func(d *WorkingDay) { d.BreakStartsAt, d.BreakEndsAt = "13:00", "later" }, true},
		{"break ends before start", func(d *WorkingDay) { d.BreakStartsAt, d.BreakEndsAt = "14:00", "13:00" }, true},
		{"break before the day", func(d *WorkingDay) { d.BreakStartsAt, d.BreakEndsAt = "08:00", "09:30" }, true},
		{"break after the day", func(d *WorkingDay) { d.BreakStartsAt, d.BreakEndsAt = "17:30", "19:00" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day := valid
			tt.mutate(&day)

			err := day.Validate()

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
