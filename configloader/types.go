//nolint:gocritic // Explcitiness is good.
package configloader

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"fortio.org/safecast"
)

var defaultTypeHandlers = []Option{
	WithTypeHandler(func(val string) (time.Duration, error) {
		return time.ParseDuration(val)
	}),

	// nolint:unlambda // Explcitiness is good.
	WithTypeHandler(func(val string) (net.IP, error) {
		ip := net.ParseIP(val)
		if ip == nil {
			return nil, errors.New("invalid ip")
		}
		return ip, nil
	}),

	WithTypeHandler(func(val string) (json.RawMessage, error) {
		return json.RawMessage(val), nil
	}),

	WithTypeHandler(func(val string) (time.Time, error) {
		return time.Parse(time.RFC3339, val)
	}),

	WithTypeHandler(func(val string) ([]byte, error) {
		return base64.StdEncoding.DecodeString(val)
	}),

	WithTypeHandler(func(val string) (*url.URL, error) {
		return url.Parse(val)
	}),
	WithTypeHandler(func(val string) (*regexp.Regexp, error) {
		return regexp.Compile(val)
	}),
	WithTypeHandler(func(val string) (*net.IPNet, error) {
		_, network, err := net.ParseCIDR(val)
		return network, err
	}),

	WithTypeHandler(parseInt(safecast.Convert[int64])),
	WithTypeHandler(parseInt(safecast.Convert[int32])),
	WithTypeHandler(parseInt(safecast.Convert[int16])),
	WithTypeHandler(parseInt(safecast.Convert[int8])),
	WithTypeHandler(parseInt(safecast.Convert[int])),

	WithTypeHandler(parseUint(safecast.Convert[uint64])),
	WithTypeHandler(parseUint(safecast.Convert[uint32])),
	WithTypeHandler(parseUint(safecast.Convert[uint16])),
	WithTypeHandler(parseUint(safecast.Convert[uint8])),
	WithTypeHandler(parseUint(safecast.Convert[uint])),

	WithTypeHandler(parseFloat(safecast.Convert[float64])),
	WithTypeHandler(parseFloat(safecast.Convert[float32])),

	WithTypeHandler(strconv.ParseBool),
	WithTypeHandler(func(val string) (string, error) {
		return val, nil
	}),
}

func parseInt[T any](cast func(int64) (T, error)) func(string) (T, error) {
	return func(val string) (res T, err error) {
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return res, err
		}
		return cast(parsed)
	}
}

func parseUint[T any](cast func(uint64) (T, error)) func(string) (T, error) {
	return func(val string) (res T, err error) {
		parsed, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return res, err
		}
		return cast(parsed)
	}
}

func parseFloat[T any](cast func(float64) (T, error)) func(string) (T, error) {
	return func(val string) (res T, err error) {
		parsed, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return res, err
		}
		return cast(parsed)
	}
}
