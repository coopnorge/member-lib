package phone

import "regexp"

// Format of phone number validation.
type Format byte

const (
	// FormatE164 https://en.wikipedia.org/wiki/E.164 standard.
	FormatE164 Format = iota
)

const (
	patternE164 = `^\+[1-9]\d{1,14}$`
)

// Validate validates if the given phone number in E.164 format.
func Validate(f Format, p string) bool {
	if f == FormatE164 {
		return regexp.MustCompile(patternE164).MatchString(p)
	}

	return false
}
