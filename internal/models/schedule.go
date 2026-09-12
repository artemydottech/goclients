package models

import (
	"fmt"
	"strconv"
	"strings"
)

const minutesInDay = 24 * 60

// WorkingDay — рабочий день мастера в недельном графике. Время хранится как
// «10:00»: так его читает и человек, и запрос календаря.
type WorkingDay struct {
	EmployeeID int    `json:"employee_id"`
	Weekday    int    `json:"weekday"`
	StartsAt   string `json:"starts_at"`
	EndsAt     string `json:"ends_at"`
}

// ParseDayTime переводит «10:30» в минуты от полуночи.
func ParseDayTime(value string) (int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("ожидался формат ЧЧ:ММ, получено %q", value)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 || hours > 23 {
		return 0, fmt.Errorf("некорректные часы в %q", value)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("некорректные минуты в %q", value)
	}

	return hours*60 + minutes, nil
}

func FormatDayTime(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

func (d WorkingDay) Validate() error {
	if d.Weekday < 0 || d.Weekday > 6 {
		return Invalid("День недели должен быть от 0 (воскресенье) до 6!")
	}

	start, err := ParseDayTime(d.StartsAt)
	if err != nil {
		return Invalid("Начало дня: %v", err)
	}

	end, err := ParseDayTime(d.EndsAt)
	if err != nil {
		return Invalid("Конец дня: %v", err)
	}

	if end <= start {
		return Invalid("Конец рабочего дня должен быть позже начала!")
	}

	if end > minutesInDay {
		return Invalid("Рабочий день не может выходить за сутки!")
	}

	return nil
}
