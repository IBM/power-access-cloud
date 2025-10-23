package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExpectedType_AllKinds(t *testing.T) {
	type customStruct struct {
		Foo string
	}

	tests := []struct {
		name     string
		typ      reflect.Type
		expected string
	}{
		{
			name:     "int",
			typ:      reflect.TypeOf(int(0)),
			expected: "number",
		},
		{
			name:     "uint64",
			typ:      reflect.TypeOf(uint64(0)),
			expected: "number",
		},
		{
			name:     "float64",
			typ:      reflect.TypeOf(float64(0)),
			expected: "number",
		},
		{
			name:     "bool",
			typ:      reflect.TypeOf(true),
			expected: "boolean",
		},
		{
			name:     "string",
			typ:      reflect.TypeOf("hello"),
			expected: "string",
		},
		{
			name:     "slice",
			typ:      reflect.TypeOf([]int{1, 2, 3}),
			expected: "array",
		},
		{
			name:     "array",
			typ:      reflect.TypeOf([3]int{1, 2, 3}),
			expected: "array",
		},
		{
			name:     "map",
			typ:      reflect.TypeOf(map[string]int{}),
			expected: "object",
		},
		{
			name:     "struct",
			typ:      reflect.TypeOf(customStruct{}),
			expected: "object",
		},
		{
			name:     "chan (default case)",
			typ:      reflect.TypeOf(make(chan int)),
			expected: "value",
		},
		{
			name:     "pointer (default case)",
			typ:      reflect.TypeOf(&customStruct{}),
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getExpectedType(tt.typ)
			assert.Equal(t, tt.expected, got)
		})
	}
}
