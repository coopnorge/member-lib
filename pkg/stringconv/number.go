// Package stringconv provides functions for string conversions to various number types.
// It simplifies the conversion (casting) between strings and various numeric
// types in Go, such as int, int64, uint, etc., in a secure manner that gracefully handles type overflows.
package stringconv

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// WholeNumber is a type constraint that includes all integer types.
type WholeNumber interface {
	uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64
}

// ToWholeNumber safely converts a string to a given whole number type (e.g., int, uint) with positive or negative handling.
// It returns the converted number and an error if the conversion fails.
func ToWholeNumber[T WholeNumber](str string) (T, error) {
	const errTmpl = "failed to convert string to WholeNumber, error: %v"
	var result T
	isWithErr := true

	maxBitSize, maxBitSizeErr := defineMaxIntType(T(0))
	if maxBitSizeErr != nil {
		return result, fmt.Errorf(errTmpl, "unable to define bit size for number parsing")
	}

	result = parseInt[T](&str, &maxBitSize, &isWithErr)
	if result == 0 && isWithErr {
		result = parseUint[T](&str, &maxBitSize, &isWithErr)
	}

	if isWithErr {
		return result, fmt.Errorf(errTmpl, "string value not matching any whole number primitive type")
	}

	return result, nil
}

// parseInt attempts to parse a string as a signed integer of the specified type.
// It takes into account the maximum bit size and whether the string represents a negative number.
func parseInt[T WholeNumber](str *string, maxBitSize *int, isWithErr *bool) (result T) {
	isNegativeInt := strings.HasPrefix(*str, "-")
	maxIntStr := strconv.FormatInt(math.MaxInt64, 10)
	if isNegativeInt || len(*str) <= len(maxIntStr) {
		if pInt, pIntErr := strconv.ParseInt(*str, 10, *maxBitSize); pIntErr == nil {
			result = T(pInt)
			*isWithErr = false
		}
	}

	return result
}

// parseUint attempts to parse a string as an unsigned integer of the specified type.
// It takes into account the maximum bit size.
func parseUint[T WholeNumber](str *string, maxBitSize *int, isWithErr *bool) (result T) {
	maxUintStr := strconv.FormatUint(math.MaxUint64, 10)
	if len(*str) <= len(maxUintStr) {
		if pUint, pUintErr := strconv.ParseUint(*str, 10, *maxBitSize); pUintErr == nil {
			result = T(pUint)
			*isWithErr = false
		}
	}

	return result
}

// defineMaxIntType determines the maximum bit size for a given integer type.
// It returns the bit size and an error if the type is not recognized.
func defineMaxIntType(number any) (maxBitSize int, err error) {
	switch reflect.TypeOf(number).Kind() {
	case reflect.Int8, reflect.Uint8:
		maxBitSize = 8
	case reflect.Int16, reflect.Uint16:
		maxBitSize = 16
	case reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32:
		maxBitSize = 32
	case reflect.Int64, reflect.Uint64:
		maxBitSize = 64
	default:
		err = fmt.Errorf("unsupported given type: %v", number)
	}

	return maxBitSize, err
}
