package main

import (
	"fmt"
	"testing"
)

type testCase struct {
	name                string
	input               string
	expectedOutput      any
	expectedInputLength int
	err                 error
}

func TestDecodeBencode(t *testing.T) {
	testCases := []*testCase{
		{
			name:                "String",
			input:               "5:hello",
			expectedOutput:      "hello",
			expectedInputLength: 7,
			err:                 nil,
		},
		{
			name:                "String with numbers",
			input:               "10:hello12345",
			expectedOutput:      "hello12345",
			expectedInputLength: 13,
			err:                 nil,
		},
		{
			name:                "Integer",
			input:               "i123e",
			expectedOutput:      123,
			expectedInputLength: 5,
			err:                 nil,
		},
		{
			name:                "Negative Integer",
			input:               "i-123e",
			expectedOutput:      -123,
			expectedInputLength: 6,
			err:                 nil,
		},
		{
			name:                "Invalid Integer",
			input:               "i-0e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrNegativeZero,
		},
		{
			name:                "Invalid Integer 2",
			input:               "i01e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrZeroPrefixedInteger,
		},
		{
			name:                "Invalid Integer 3",
			input:               "i-01e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrZeroPrefixedInteger,
		},
		{
			name:                "List",
			input:               "li-22e5:helloe",
			expectedOutput:      List{-22, "hello"},
			expectedInputLength: 14,
			err:                 nil,
		},
		{
			name:                "Missing list terminator",
			input:               "l",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrUnterminatedList,
		},
		{
			name:                "List with elements but missing terminator",
			input:               "li-22e5:helloi5e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrUnterminatedList,
		},
		{
			name:                "Missing dictionary terminator",
			input:               "d",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrUnterminatedDictionary,
		},
		{
			name:                "Parses dictionary with key-value pair",
			input:               "d3:keyi23ee",
			expectedOutput:      map[string]any{"key": 23},
			expectedInputLength: 11,
			err:                 nil,
		},
		{
			name:                "Parses dictionary with multiple key-value pairs",
			input:               "d3:keyi23e4:key2i-12ee",
			expectedOutput:      map[string]any{"key": 23, "key2": -12},
			expectedInputLength: 22,
			err:                 nil,
		},
		{
			name:                "Fails to parse dictionary than ends in key",
			input:               "d3:keye",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrUnterminatedDictionary,
		},
		{
			name:                "Missing dictionary terminator after key-value pairs",
			input:               "d3:keyi23e",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrUnterminatedDictionary,
		},
		{
			name:                "Dictionary keys must be strings",
			input:               "d3:keyi23ei10ei-12ee",
			expectedOutput:      "",
			expectedInputLength: 0,
			err:                 ErrInvalidDictionaryKey,
		},
		{
			name:                "Decodes maps as values of maps",
			input:               "d3:keyi23e6:mapkeyd9:insidekeyi987eee",
			expectedOutput:      map[string]any{"key": 23, "mapkey": map[string]any{"insidekey": 987}},
			expectedInputLength: 37,
			err:                 nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := DecodeBencode(tc.input)
			if tc.err != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err != tc.err {
					t.Fatalf("expected different error. expected %q, found %q", tc.err.Error(), err.Error())
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
				if expectedMap, isMapExpected := tc.expectedOutput.(map[string]any); isMapExpected {
					compareMaps(t, actual.Output.(map[string]any), expectedMap)
					if actual.InputLength != tc.expectedInputLength {
						t.Fatalf("expected detected input length %d, got %d", tc.expectedInputLength, actual.InputLength)
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

func compareMaps(t *testing.T, actualMap map[string]any, expectedMap map[string]any) {
	t.Helper()

	fmt.Println("here")

	if len(actualMap) != len(expectedMap) {
		t.Fatalf("different length maps. expected output %v, got %v", expectedMap, actualMap)
	}
	for k, v := range expectedMap {
		// if v itself is also a map
		if mapValue, valueIsMap := v.(map[string]any); valueIsMap {
			compareMaps(t, actualMap[k].(map[string]any), mapValue)
			continue
		}
		// if it's another type of value
		// FIXME: this will break for list values too. need to generalise these comparisons
		if v != actualMap[k] {
			t.Fatalf("while parsing map, expected value %v for key %q, but got value %v instead", v, k, actualMap[k])
		}
	}
}
