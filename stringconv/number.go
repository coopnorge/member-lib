package stringconv

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// WholeNumber type.
type WholeNumber interface {
	uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64
}

// ToWholeNumber safely converting string to given whole number.md with positive or negative handling.
func ToWholeNumber[T WholeNumber](str string) (T, error) {
	const errTmpl string = "failed to convert string to WholeNumber, error: %v"
	var result T
	isWithErr := true

	maxBitSize, maxBitSizeErr := defineMaxIntType(T(0))
	if maxBitSizeErr != nil {
		return 0, fmt.Errorf(errTmpl, "unable define bit size for number parsing")
	}

	result = parseInt[T](&str, &maxBitSize, &isWithErr)
	if result == 0 && isWithErr {
		result = parseUint[T](&str, &maxBitSize, &isWithErr)
	}

	if isWithErr {
		return 0, fmt.Errorf(errTmpl, "string value not matching any whole number premitive type")
	}

	return result, nil
}

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
		err = fmt.Errorf("unable to define type of (%v) it's not part of WholeNumber", number)
	}

	return
}
