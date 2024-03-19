package main

import (
	"testing"
)

func TestDecodeBencode(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected any
		err      bool
	}{
		{
			name:     "String",
			input:    "5:hello",
			expected: "hello",
			err:      false,
		},
		{
			name:     "String with numbers",
			input:    "10:hello12345",
			expected: "hello12345",
			err:      false,
		},
		{
			name:     "Integer",
			input:    "i123e",
			expected: 123,
			err:      false,
		},
		{
			name:     "Negative Integer",
			input:    "i-123e",
			expected: -123,
			err:      false,
		},
		{
			name:     "Invalid Integer",
			input:    "i-0e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 2",
			input:    "i01e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 3",
			input:    "i-01e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 4",
			input:    "i01e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 5",
			input:    "i-0e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 6",
			input:    "i-01e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 7",
			input:    "i-01e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 8",
			input:    "i-01e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 9",
			input:    "i-01e",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid Integer 10",
			input:    "i-01e",
			expected: "",
			err:      true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := DecodeBencode(tc.input)
			if tc.err {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if actual != tc.expected {
					t.Fatalf("expected %v, got %v", tc.expected, actual)
				}
			}
		})
	}
}
