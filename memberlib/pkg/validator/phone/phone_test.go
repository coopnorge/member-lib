package phone

import "testing"

func TestValidateE164Format(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{"Valid E.164", "+12345678901", true},
		{"Valid E.164 with max length", "+123456789012345", true},
		{"Missing plus sign", "12345678901", false},
		{"Invalid country code 0", "+0123456789", false},
		{"Too short", "+1", false},
		{"Too long", "+1234567890123456", false},
		{"Contains letters", "+12345A7890", false},
		{"Only plus sign", "+", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(FormatE164, tt.number)
			if result != tt.expected {
				t.Errorf("validateE164Format(%s) got %v, want %v", tt.number, result, tt.expected)
			}
		})
	}
}
