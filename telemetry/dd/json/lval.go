package json

import (
	"encoding/json"
	"errors"

	"go.opentelemetry.io/otel/log"
)

type value struct {
	log.Value
}

func newValue(v log.Value) value {
	return value{Value: v}
}

// MarshalJSON implements a custom marshal function to encode log.Value.
func (v value) MarshalJSON() ([]byte, error) {
	var val any
	switch v.Kind() {
	case log.KindString:
		val = v.AsString()
	case log.KindInt64:
		val = v.AsInt64()
	case log.KindFloat64:
		val = v.AsFloat64()
	case log.KindBool:
		val = v.AsBool()
	case log.KindBytes:
		val = v.AsBytes()
	case log.KindMap:
		m := v.AsMap()
		values := make(map[string]json.RawMessage, len(m))
		for _, kv := range m {
			data, err := newValue(kv.Value).MarshalJSON()
			if err != nil {
				return nil, err
			}
			values[kv.Key] = data
		}
		val = values
	case log.KindSlice:
		s := v.AsSlice()
		values := make([]json.RawMessage, 0, len(s))
		for _, e := range s {
			data, err := newValue(e).MarshalJSON()
			if err != nil {
				return nil, err
			}
			values = append(values, data)
		}
		val = values
	case log.KindEmpty:
		val = nil
	default:
		return nil, errors.New("invalid Kind")
	}
	return json.Marshal(val)
}
