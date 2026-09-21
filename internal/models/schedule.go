package models

import (
	"fmt"
	"strconv"
	"strings"
)

const minutesInDay = 24 * 60

type WorkingDay struct {
	EmployeeID int    `json:"employee_id"`
	Weekday    int    `json:"weekday"`
	StartsAt   string `json:"starts_at"`
	EndsAt     string `json:"ends_at"`

	BreakStartsAt string `json:"break_starts_at,omitempty"`
	BreakEndsAt   string `json:"break_ends_at,omitempty"`
}

func (d WorkingDay) HasBreak() bool {
	return d.BreakStartsAt != "" || d.BreakEndsAt != ""
}

func ParseDayTime(value string) (int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("expected HH:MM format, got %q", value)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 || hours > 23 {
		return 0, fmt.Errorf("invalid hours in %q", value)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("invalid minutes in %q", value)
	}

	return hours*60 + minutes, nil
}

func FormatDayTime(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

func (d WorkingDay) Validate() error {
	if d.Weekday < 0 || d.Weekday > 6 {
		return Invalid("weekday must be between 0 (sunday) and 6")
	}

	start, err := ParseDayTime(d.StartsAt)
	if err != nil {
		return Invalid("day start: %v", err)
	}

	end, err := ParseDayTime(d.EndsAt)
	if err != nil {
		return Invalid("day end: %v", err)
	}

	if end <= start {
		return Invalid("working day must end after it starts")
	}

	if end > minutesInDay {
		return Invalid("working day cannot span past midnight")
	}

	if !d.HasBreak() {
		return nil
	}

	if d.BreakStartsAt == "" || d.BreakEndsAt == "" {
		return Invalid("break needs both start and end")
	}

	breakStart, err := ParseDayTime(d.BreakStartsAt)
	if err != nil {
		return Invalid("break start: %v", err)
	}

	breakEnd, err := ParseDayTime(d.BreakEndsAt)
	if err != nil {
		return Invalid("break end: %v", err)
	}

	if breakEnd <= breakStart {
		return Invalid("break must end after it starts")
	}

	if breakStart < start || breakEnd > end {
		return Invalid("break must fall inside the working day")
	}

	return nil
}
