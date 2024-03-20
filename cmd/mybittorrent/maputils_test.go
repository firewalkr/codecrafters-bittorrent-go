package main

import (
	"testing"
)

type getStringValueTestCase struct {
	name          string
	inputMap      map[string]any
	inputKey      string
	expectedValue string
	err           error
}

func TestGetStringValue(t *testing.T) {
	testCases := []*getStringValueTestCase{
		{
			name:          "missing map key",
			inputMap:      map[string]any{},
			inputKey:      "testkey",
			expectedValue: "",
			err:           ErrMissingMapKey,
		},
		{
			name:          "returns value for provided key",
			inputMap:      map[string]any{"testkey": "testvalue"},
			inputKey:      "testkey",
			expectedValue: "testvalue",
			err:           nil,
		},
		{
			name:          "errors out if value is not string",
			inputMap:      map[string]any{"testkey": 123},
			inputKey:      "testkey",
			expectedValue: "",
			err:           ErrMapValueIsNotString,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := GetStringValue(tc.inputMap, tc.inputKey)
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
				if actual != tc.expectedValue {
					t.Fatalf("expected output %v, got %q", tc.expectedValue, actual)
				}
			}
		})
	}
}

type getMapValueTestCase struct {
	name          string
	inputMap      map[string]any
	inputKey      string
	expectedValue map[string]any
	err           error
}

func TestGetMapValue(t *testing.T) {
	testCases := []*getMapValueTestCase{
		{
			name:          "missing map key",
			inputMap:      map[string]any{},
			inputKey:      "testkey",
			expectedValue: nil,
			err:           ErrMissingMapKey,
		},
		{
			name:          "returns value for provided key",
			inputMap:      map[string]any{"testkey": map[string]any{"innerkey": 123}},
			inputKey:      "testkey",
			expectedValue: map[string]any{"innerkey": 123},
			err:           nil,
		},
		{
			name:          "errors out if value is not map",
			inputMap:      map[string]any{"testkey": 123},
			inputKey:      "testkey",
			expectedValue: nil,
			err:           ErrMapValueIsNotMap,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, actualErr := GetMapValue(testCase.inputMap, testCase.inputKey)
			if testCase.err != nil {
				if actualErr == nil {
					t.Fatalf("expected error, got nil")
				}
				if actualErr != testCase.err {
					t.Fatalf("expected different error. expected %q, found %q", testCase.err.Error(), actualErr.Error())
				}
			} else { // testCase.err is nil
				if actualErr != nil {
					t.Fatalf("unexpected error: %s", actualErr)
				}
				compareMaps(t, actual, testCase.expectedValue)
			}
		})
	}
}

type getIntValueTestCase struct {
	name          string
	inputMap      map[string]any
	inputKey      string
	expectedValue int
	err           error
}

func TestGetIntValue(t *testing.T) {
	testCases := []*getIntValueTestCase{
		{
			name:          "missing map key",
			inputMap:      map[string]any{},
			inputKey:      "testkey",
			expectedValue: 0,
			err:           ErrMissingMapKey,
		},
		{
			name:          "returns value for provided key",
			inputMap:      map[string]any{"testkey": 123},
			inputKey:      "testkey",
			expectedValue: 123,
			err:           nil,
		},
		{
			name:          "errors out if value is not int",
			inputMap:      map[string]any{"testkey": "not an int"},
			inputKey:      "testkey",
			expectedValue: 0,
			err:           ErrMapValueIsNotInt,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, actualErr := GetIntValue(testCase.inputMap, testCase.inputKey)
			if testCase.err != nil {
				t.Logf("hey")
				if actualErr == nil {
					t.Fatalf("expected error, got nil")
				}
				if actualErr != testCase.err {
					t.Fatalf("expected different error. expected %q, found %q", testCase.err.Error(), actualErr.Error())
				}
			} else { // testCase.err is nil
				if actualErr != nil {
					t.Fatalf("unexpected error: %s", actualErr)
				}
				if actual != testCase.expectedValue {
					t.Fatalf("expected output %v, got %q", testCase.expectedValue, actual)
				}
			}
		})
	}
}
