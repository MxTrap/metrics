package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePointer(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		validate func(t *testing.T, ptr interface{})
	}{
		{
			name:  "Integer",
			input: 42,
			validate: func(t *testing.T, ptr interface{}) {
				p, ok := ptr.(*int)
				assert.True(t, ok, "Should return *int")
				assert.Equal(t, 42, *p, "Pointer should point to input value")
			},
		},
		{
			name:  "Float64",
			input: 3.14,
			validate: func(t *testing.T, ptr interface{}) {
				p, ok := ptr.(*float64)
				assert.True(t, ok, "Should return *float64")
				assert.Equal(t, 3.14, *p, "Pointer should point to input value")
			},
		},
		{
			name:  "String",
			input: "test",
			validate: func(t *testing.T, ptr interface{}) {
				p, ok := ptr.(*string)
				assert.True(t, ok, "Should return *string")
				assert.Equal(t, "test", *p, "Pointer should point to input value")
			},
		},
		{
			name:  "Zero value int",
			input: 0,
			validate: func(t *testing.T, ptr interface{}) {
				p, ok := ptr.(*int)
				assert.True(t, ok, "Should return *int")
				assert.Equal(t, 0, *p, "Pointer should point to zero value")
			},
		},
		{
			name: "Struct",
			input: struct {
				Name string
				Age  int
			}{Name: "Alice", Age: 30},
			validate: func(t *testing.T, ptr interface{}) {
				p, ok := ptr.(*struct {
					Name string
					Age  int
				})
				assert.True(t, ok, "Should return pointer to struct")
				assert.Equal(t, "Alice", p.Name, "Struct field Name should match")
				assert.Equal(t, 30, p.Age, "Struct field Age should match")
			},
		},
		{
			name:  "Nil-like value (empty string)",
			input: "",
			validate: func(t *testing.T, ptr interface{}) {
				p, ok := ptr.(*string)
				assert.True(t, ok, "Should return *string")
				assert.Equal(t, "", *p, "Pointer should point to empty string")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.input.(type) {
			case int:
				ptr := MakePointer(v)
				tt.validate(t, ptr)
			case float64:
				ptr := MakePointer(v)
				tt.validate(t, ptr)
			case string:
				ptr := MakePointer(v)
				tt.validate(t, ptr)
			case struct {
				Name string
				Age  int
			}:
				ptr := MakePointer(v)
				tt.validate(t, ptr)
			default:
				t.Fatal("Unsupported type")
			}
		})
	}
}
