package main

import (
	"testing"
)

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"000", "001"},
		{"009", "010"},
		{"099", "100"},
		{"100", "101"},
		{"999", "1000"},
		{},
	}

	for _, test := range tests {
		result := IncrementVersion(test.input)
		if result != test.expected {
			t.Errorf("IncrementVersion(%s) = %s; expected %s", test.input, result, test.expected)
		}
	}
}
