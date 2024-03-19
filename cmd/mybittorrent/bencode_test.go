package main

import (
	"testing"
)

func TestDecodeBencode(t *testing.T) {
	testCases := []struct {
		name                string
		input               string
		expectedOutput      any
		expectedInputLength int
		err                 bool
	}{
		{
			name:                "String",
			input:               "5:hello",
			expectedOutput:      "hello",
			expectedInputLength: 7,
			err:                 false,
		},
		{
			name:                "String with numbers",
			input:               "10:hello12345",
			expectedOutput:      "hello12345",
			expectedInputLength: 13,
			err:                 false,
		},
		{
			name:                "Integer",
			input:               "i123e",
			expectedOutput:      123,
			expectedInputLength: 5,
			err:                 false,
		},
		{
			name:                "Negative Integer",
			input:               "i-123e",
			expectedOutput:      -123,
			expectedInputLength: 6,
			err:                 false,
		},
		{
			name:                "Invalid Integer",
			input:               "i-0e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 true,
		},
		{
			name:                "Invalid Integer 2",
			input:               "i01e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 true,
		},
		{
			name:                "Invalid Integer 3",
			input:               "i-01e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 true,
		},
		{
			name:                "List",
			input:               "li-22e5:helloe",
			expectedOutput:      List{-22, "hello"},
			expectedInputLength: 14,
			err:                 false,
		},
		{
			name:                "Missing terminator",
			input:               "l",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 true,
		},
		{
			name:                "List with elements but missing terminator",
			input:               "li-22e5:helloi5e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 true,
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
				if list, ok := actual.Output.(List); ok {
					if len(list) != len(tc.expectedOutput.(List)) {
						t.Fatalf("expected output %v, got %v", tc.expectedOutput, actual.Output)
					}
					for i := 0; i < len(list); i++ {
						if list[i] != tc.expectedOutput.(List)[i] {
							t.Fatalf("expected output %v, got %v", tc.expectedOutput, actual.Output)
						}
					}
					return
				}
				if actual.Output != tc.expectedOutput {
					t.Fatalf("expected output %v, got %v", tc.expectedOutput, actual.Output)
				}
				if actual.InputLength != tc.expectedInputLength {
					t.Fatalf("expected detected input length %d, got %d", tc.expectedInputLength, actual.InputLength)
				}
			}
		})
	}
}
