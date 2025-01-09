package json

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/log"
)

func Test_value_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   log.Value
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{"String", log.StringValue("string"), `"string"`, assert.NoError},
		{"Int", log.IntValue(1), `1`, assert.NoError},
		{"-Int", log.IntValue(-1), `-1`, assert.NoError},
		{"Int64", log.Int64Value(1), `1`, assert.NoError},
		{"-Int64", log.Int64Value(-1), `-1`, assert.NoError},
		{"Bool", log.BoolValue(true), `true`, assert.NoError},
		{"Bytes", log.BytesValue([]byte("bytes")), `"Ynl0ZXM="`, assert.NoError},
		{"Map", log.MapValue(log.String("key", "value")), `{"key":"value"}`, assert.NoError},
		{"Map of map", log.MapValue(log.Map("key1", log.String("key2", "value"))), `{"key1":{"key2":"value"}}`, assert.NoError},
		{"Slice", log.SliceValue(log.StringValue("value"), log.IntValue(1234)), `["value",1234]`, assert.NoError},
		{"Slice of slice", log.SliceValue(log.SliceValue(log.StringValue("value")), log.IntValue(1234)), `[["value"],1234]`, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := value{
				Value: tt.value,
			}
			got, err := v.MarshalJSON()
			if !tt.wantErr(t, err, fmt.Sprintf("MarshalJSON()")) {
				return
			}
			assert.Equalf(t, tt.want, string(got), "MarshalJSON()")
		})
	}
}
