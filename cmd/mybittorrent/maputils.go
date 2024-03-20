package main

import "errors"

var ErrMissingMapKey = errors.New("requested map key is missing")
var ErrMapValueIsNotString = errors.New("value obtained is not a string")
var ErrMapValueIsNotMap = errors.New("value obtained is not a map")
var ErrMapValueIsNotInt = errors.New("value obtained is not an int")

func GetStringValue(m map[string]any, key string) (string, error) {
	val, keyExists := m[key]
	if !keyExists {
		return "", ErrMissingMapKey
	}

	stringValue, isValueString := val.(string)
	if !isValueString {
		return "", ErrMapValueIsNotString
	}

	return stringValue, nil
}

func GetMapValue(m map[string]any, key string) (map[string]any, error) {
	val, keyExists := m[key]
	if !keyExists {
		return nil, ErrMissingMapKey
	}

	mapValue, isValueMap := val.(map[string]any)
	if !isValueMap {
		return nil, ErrMapValueIsNotMap
	}

	return mapValue, nil
}

func GetIntValue(m map[string]any, key string) (int, error) {
	val, keyExists := m[key]
	if !keyExists {
		return 0, ErrMissingMapKey
	}

	intValue, isValueMap := val.(int)
	if !isValueMap {
		return 0, ErrMapValueIsNotInt
	}

	return intValue, nil
}
