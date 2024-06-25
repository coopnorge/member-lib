// Package phone provides functions for working with string number related information.
package phone

import "regexp"

// Format represents the phone number format.
type Format byte

const (
	// FormatE164 represents the E.164 phone number format.
	// For more details about the E.164 format, see https://en.wikipedia.org/wiki/E.164.
	FormatE164 Format = iota
)

const (
	patternE164 = `^\+[1-9]\d{1,14}$`
)

// Validate checks if the given phone number p is in the specified format f.
// Currently, it supports the E.164 format only. If the phone number does not match the format,
// the function returns false. The format is not case-sensitive.
//
// The function uses the regular expression match to validate the phone number.
//
// Example:
//
//	isValid := phone.Validate(phone.FormatE164, "+1234567890")
//	fmt.Println(isValid)  // Output: true
func Validate(f Format, p string) bool {
	if f == FormatE164 {
		return regexp.MustCompile(patternE164).MatchString(p)
	}

	return false
}
