package dd

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func Test_formatBoundValue(t *testing.T) {
	type args struct {
		val float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"positive inf", args{math.Inf(1)}, "inf"},
		{"negative inf", args{math.Inf(-1)}, "-inf"},

		{"whole number zero", args{0.0}, "0.0"},
		{"whole number positive", args{5.0}, "5.0"},
		{"whole number negative", args{-5.0}, "-5.0"},

		{"decimal number", args{5.5}, "5.5"},
		{"small decimal", args{0.1}, "0.1"},
		{"negative decimal", args{-5.5}, "-5.5"},

		{"repeating third", args{1.0 / 3.0}, "0.3333333333333333"},
		{"repeating third negative", args{-1.0 / 3.0}, "-0.3333333333333333"},
		{"repeating sixth", args{1.0 / 6.0}, "0.16666666666666666"},

		{"very small positive number", args{0.000001}, "0.000001"},
		{"very small negative number", args{-0.000001}, "-0.000001"},

		{"very large number", args{1000000.5}, "1000000.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, formatBoundValue(tt.args.val), "formatBoundValue(%v)", tt.args.val)
		})
	}
}
