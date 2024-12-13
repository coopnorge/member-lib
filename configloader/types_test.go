package configloader

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function for running type handler tests
func runTypeHandlerTest[T any](t *testing.T, tests []struct {
	name      string
	input     string
	expected  T
	expectErr assert.ErrorAssertionFunc
},
) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := struct {
				Val      T
				ValPtr   *T
				ValSlice []T
			}{}
			sliceVal := strings.Join([]string{tt.input, tt.input}, ",")
			require.NoError(t, os.Setenv("VAL", tt.input))
			require.NoError(t, os.Setenv("VAL_PTR", tt.input))
			require.NoError(t, os.Setenv("VAL_SLICE", sliceVal))
			defer os.Clearenv()

			err := Load(&val)
			tt.expectErr(t, err)
			assert.EqualValues(t, tt.expected, val.Val)
			assert.EqualValues(t, tt.expected, *val.ValPtr)
			assert.EqualValues(t, []T{tt.expected, tt.expected}, val.ValSlice)
		})
	}
}

// Helper function to create IPNet for testing
func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
	}
	return network
}

func TestBooleanHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  bool
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "True boolean",
			input:     "true",
			expected:  true,
			expectErr: assert.NoError,
		},
		{
			name:      "False boolean",
			input:     "false",
			expected:  false,
			expectErr: assert.NoError,
		},
		{
			name:      "False boolean",
			input:     "invalid",
			expected:  false,
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestFloatHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  float64
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Positive float",
			input:     "10.01",
			expected:  10.01,
			expectErr: assert.NoError,
		},
		{
			name:      "Negative float",
			input:     "-10.01",
			expected:  -10.01,
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid float",
			input:     "invalid",
			expected:  0,
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestIntegerHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  int
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Positive integer",
			input:     "10",
			expected:  10,
			expectErr: assert.NoError,
		},
		{
			name:      "Negative integer",
			input:     "-10",
			expected:  -10,
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid integer",
			input:     "invalid",
			expected:  0,
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestDurationHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  time.Duration
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Valid Duration",
			input:     "5s",
			expected:  5 * time.Second,
			expectErr: assert.NoError,
		},
		{
			name:      "Zero duration",
			input:     "0",
			expected:  time.Duration(0),
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid duration",
			input:     "invalid",
			expected:  time.Duration(0),
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestIPHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  net.IP
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Valid IP",
			input:     "192.168.1.1",
			expected:  net.ParseIP("192.168.1.1"),
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid IP",
			input:     "invalid-ip",
			expected:  nil,
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func mustParseUrl(val string) *url.URL {
	v, err := url.Parse(val)
	if err != nil {
		panic(err)
	}
	return v
}

func TestURLHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  *url.URL
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Valid URL",
			input:     "https://example.com",
			expected:  mustParseUrl("https://example.com"),
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid URL",
			input:     "://invalid",
			expected:  nil,
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestRegexpHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  *regexp.Regexp
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Valid Regexp",
			input:     "^test$",
			expected:  regexp.MustCompile("^test$"),
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid Regexp",
			input:     "[invalid",
			expected:  (*regexp.Regexp)(nil),
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestJSONHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  json.RawMessage
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Valid JSON",
			input:     `{"key":"value"}`,
			expected:  json.RawMessage(`{"key":"value"}`),
			expectErr: assert.NoError,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestTimeHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  time.Time
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Valid Time",
			input:     "2023-01-02T15:04:05Z",
			expected:  time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid Time",
			input:     "invalid-time",
			expected:  time.Time{},
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestBase64Handler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  []byte
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Valid Base64",
			input:     "SGVsbG8=",
			expected:  []byte("Hello"),
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid Base64",
			input:     "invalid-base64!",
			expected:  []byte(nil),
			expectErr: assert.Error,
		},
	}

	runTypeHandlerTest(t, tests)
}

func TestCIDRHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  *net.IPNet
		expectErr assert.ErrorAssertionFunc
	}{
		{
			name:      "Valid CIDR",
			input:     "192.168.1.0/24",
			expected:  mustParseCIDR("192.168.1.0/24"),
			expectErr: assert.NoError,
		},
		{
			name:      "Invalid CIDR",
			input:     "invalid-cidr",
			expected:  (*net.IPNet)(nil),
			expectErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}

	runTypeHandlerTest(t, tests)
}
