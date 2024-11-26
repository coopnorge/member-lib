package configloader

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"regexp"
	"time"
)

var defaultTypeHandlers = []Option{
	WithTypeHandler(func(val string) (time.Duration, error) {
		return time.ParseDuration(val)
	}),

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
}
