package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func Test_attrSliceToMap(t *testing.T) {
	type args struct {
		attributes []attribute.KeyValue
	}
	tests := []struct {
		name string
		args args
		want *map[string]any
	}{
		{
			name: "string attribute",
			args: args{attributes: []attribute.KeyValue{
				attribute.String("key", "value"),
			}},
			want: &map[string]any{
				"key": "value",
			},
		},
		{
			name: "int64 attribute",
			args: args{attributes: []attribute.KeyValue{
				attribute.Int64("key", 123),
			}},
			want: &map[string]any{
				"key": int64(123),
			},
		},
		{
			name: "bool attribute",
			args: args{attributes: []attribute.KeyValue{
				attribute.Bool("key", true),
			}},
			want: &map[string]any{
				"key": true,
			},
		},
		{
			name: "float64 attribute",
			args: args{attributes: []attribute.KeyValue{
				attribute.Float64("key", 123.456),
			}},
			want: &map[string]any{
				"key": 123.456,
			},
		},
		{
			name: "multiple attributes",
			args: args{attributes: []attribute.KeyValue{
				attribute.String("str", "value"),
				attribute.Int64("int", 123),
				attribute.Bool("bool", true),
				attribute.Float64("float", 123.456),
			}},
			want: &map[string]any{
				"str":   "value",
				"int":   int64(123),
				"bool":  true,
				"float": 123.456,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValuesf(t, tt.want, attrSliceToMap(tt.args.attributes), "attrSliceToMap(%v)", tt.args.attributes)
		})
	}
}

func Test_attrSetToMap(t *testing.T) {
	type args struct {
		attributes attribute.Set
	}
	tests := []struct {
		name string
		args args
		want *map[string]any
	}{
		{
			name: "set with single attribute",
			args: args{attributes: attribute.NewSet(
				attribute.String("key", "value"),
			)},
			want: &map[string]any{
				"key": "value",
			},
		},
		{
			name: "set with multiple attributes",
			args: args{attributes: attribute.NewSet(
				attribute.String("str", "value"),
				attribute.Int64("int", 123),
				attribute.Bool("bool", true),
				attribute.Float64("float", 123.456),
			)},
			want: &map[string]any{
				"str":   "value",
				"int":   int64(123),
				"bool":  true,
				"float": 123.456,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, attrSetToMap(tt.args.attributes), "attrSetToMap(%v)", tt.args.attributes)
		})
	}
}
