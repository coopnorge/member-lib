package stringconv

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type wholeNumberTestCase struct {
	name      string
	input     string
	want      interface{}
	wantError bool
}

// TestToWholeNumber tests function with various inputs that converts string to WholeNumber.
func TestToWholeNumberInt(t *testing.T) {
	// Define test cases
	testCases := []wholeNumberTestCase{
		// Test cases with Int
		{"PositiveInt8", "100", int8(100), false},
		{"PositiveInt16", "32767", int16(32767), false},
		{"NegativeInt32", "-67890", int32(-67890), false},
		{"MaxInt64", "9223372036854775807", int64(9223372036854775807), false},
		{"MinInt64", "-9223372036854775808", int64(-9223372036854775808), false},
		// Test cases for Uint
		{"MaxByteAliasToUint8", "5", byte(5), false},
		{"MaxUint8", "255", uint8(255), false},
		{"MaxUint16", "65535", uint16(65535), false},
		{"MaxUint32", "4294967295", uint32(4294967295), false},
		{"MaxUint64", "18446744073709551615", uint64(18446744073709551615), false},

		// General test cases
		{"BeyondMaxUint64", "18446744073709551616", nil, true},
		{"Zero", "0", int64(0), false},
		{"NegativeZero", "-0", int64(0), false},
		{"SmallNegativeInt", "-1", int64(-1), false},
		{"InvalidNumber", "not_a_number", nil, true},
		{"Hexadecimal", "0x1F", nil, true},
		{"PartialInvalid", "123abc", nil, true},
		{"LargeUint", "18446744073709551615", uint64(18446744073709551615), false},
		{"EmptyString", "", nil, true},
		{"PositiveOverflowInt64", "19223372036854775807", nil, true},
		{"PositiveOverflowUint64", "28446744073709551615", nil, true},
		{"NegativeOverflowInt64", "-10223372036854775809", nil, true},
		{"FloatNumber", "42.2", nil, true},
	}

	// Verify int
	t.Run(testCases[0].name, func(t *testing.T) { runTestCase[int8](testCases[0], t) })
	t.Run(testCases[1].name, func(t *testing.T) { runTestCase[int16](testCases[1], t) })
	t.Run(testCases[2].name, func(t *testing.T) { runTestCase[int32](testCases[2], t) })
	t.Run(testCases[3].name, func(t *testing.T) { runTestCase[int64](testCases[4], t) })
	t.Run(testCases[4].name, func(t *testing.T) { runTestCase[int64](testCases[4], t) })

	// Verify uint
	t.Run(testCases[5].name, func(t *testing.T) { runTestCase[byte](testCases[5], t) })
	t.Run(testCases[6].name, func(t *testing.T) { runTestCase[uint8](testCases[6], t) })
	t.Run(testCases[7].name, func(t *testing.T) { runTestCase[uint16](testCases[7], t) })
	t.Run(testCases[8].name, func(t *testing.T) { runTestCase[uint32](testCases[8], t) })
	t.Run(testCases[9].name, func(t *testing.T) { runTestCase[uint64](testCases[9], t) })

	// Verify general test cases
	t.Run(testCases[10].name, func(t *testing.T) { runTestCase[uint64](testCases[10], t) })
	t.Run(testCases[11].name, func(t *testing.T) { runTestCase[int64](testCases[11], t) })
	t.Run(testCases[12].name, func(t *testing.T) { runTestCase[int64](testCases[12], t) })
	t.Run(testCases[13].name, func(t *testing.T) { runTestCase[int64](testCases[13], t) })
	t.Run(testCases[14].name, func(t *testing.T) { runTestCase[uint64](testCases[14], t) })
	t.Run(testCases[15].name, func(t *testing.T) { runTestCase[uint64](testCases[15], t) })
	t.Run(testCases[16].name, func(t *testing.T) { runTestCase[uint64](testCases[16], t) })
	t.Run(testCases[17].name, func(t *testing.T) { runTestCase[uint64](testCases[17], t) })
	t.Run(testCases[18].name, func(t *testing.T) { runTestCase[uint64](testCases[18], t) })
	t.Run(testCases[19].name, func(t *testing.T) { runTestCase[int64](testCases[19], t) })
	t.Run(testCases[20].name, func(t *testing.T) { runTestCase[uint64](testCases[20], t) })
	t.Run(testCases[21].name, func(t *testing.T) { runTestCase[int64](testCases[21], t) })
	t.Run(testCases[22].name, func(t *testing.T) { runTestCase[int](testCases[22], t) })
}

func runTestCase[T WholeNumber](tc wholeNumberTestCase, t *testing.T) {
	got, err := ToWholeNumber[T](tc.input)

	// Check for error expectation
	if tc.wantError {
		assert.NotNil(t, err, "expected an error but got none")
	} else if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !tc.wantError {
		// Check if the result is what we expect
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, got)
		}
	}
}

func BenchmarkProcessData(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"ShortString", "123"},
		{"MediumString", "12345"},
		{"LongString", "1234567890"},
		{"NegativeNumber", "-12345"},
		{"LeadingZeros", "000012345"},
		{"MaxInt", "2147483647"},
		{"BeyondMaxInt", "2147483648"},
		{"NegativeInt", "-552183648"},
		{"LargeUint64", "18446744073709551615"},
		{"Zero", "0"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ToWholeNumber[uint64](tc.input)
				if err != nil {
					b.Fatal("Benchmark failed with error:", err)
				}
			}
		})
	}
}
