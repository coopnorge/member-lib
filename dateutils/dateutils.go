package dateutils

import (
	"fmt"
	"time"
)

// validateYearLayout ensures that the year layout is either '2006' or '06'
func validateYearLayout(layout string) error {
	if layout != "06" && layout != "2006" {
		return fmt.Errorf("invalid year layout '%s'", layout)
	}

	return nil
}

// validateMonthLayout ensures that the month layout is 'Jan', 'January', '01' or '1'
func validateMonthLayout(layout string) error {
	if layout != "Jan" && layout != "January" && layout != "1" && layout != "01" {
		return fmt.Errorf("invalid month layout '%s'", layout)
	}

	return nil
}

// validateDayOfTheMonthLayout ensures that the day layout is '2', '_2' or '02'
func validateDayOfTheMonthLayout(layout string) error {
	if layout != "2" && layout != "_2" && layout != "02" {
		return fmt.Errorf("invalid day layout '%s'", layout)
	}

	return nil
}

// ParseDateComponent parses year, month or day strings and returns a time object
func ParseDateComponent(layout, dc string) (time.Time, error) {
	date, parseErr := time.Parse(layout, dc)
	if parseErr != nil {
		return time.Time{}, fmt.Errorf("error parsing date: %v", parseErr)
	}

	return date, parseErr
}

// ParseYear parses year string
func ParseYear(year, layout string) (uint16, error) {
	if err := validateYearLayout(layout); err != nil {
		return 0, err
	}

	date, parseDateComponentErr := ParseDateComponent(layout, year)
	if parseDateComponentErr != nil {
		return 0, fmt.Errorf("error parsing year: %v", parseDateComponentErr)
	}

	return uint16(date.Year()), nil
}

// ParseMonth parses month string
func ParseMonth(month, layout string) (uint8, error) {
	if err := validateMonthLayout(layout); err != nil {
		return 0, err
	}

	date, parseDateComponentErr := ParseDateComponent(layout, month)
	if parseDateComponentErr != nil {
		return 0, fmt.Errorf("error parsing month: %v", parseDateComponentErr)
	}

	return uint8(date.Month()), nil
}

// ParseDayOfTheMonth parses day string
func ParseDayOfTheMonth(day, layout string) (uint8, error) {
	if err := validateDayOfTheMonthLayout(layout); err != nil {
		return 0, err
	}

	date, parseDateComponentErr := ParseDateComponent(layout, day)
	if parseDateComponentErr != nil {
		return 0, fmt.Errorf("error parsing day: %v", parseDateComponentErr)
	}

	return uint8(date.Day()), nil
}

// IsLeapYear determines whether a specific year is a leap year
func IsLeapYear(year, layout string) (bool, error) {
	y, parseYearErr := ParseYear(year, layout)
	if parseYearErr != nil {
		return false, parseYearErr
	}

	condition := (y%400 == 0) || (y%4 == 0 && y%100 != 0)

	return condition, nil
}
