package main

import "errors"

var ErrMissingMapKey = errors.New("requested map key is missing")
var ErrMapValueIsNotString = errors.New("value obtained is not a string")
var ErrMapValueIsNotMap = errors.New("value obtained is not a map")
var ErrMapValueIsNotInt = errors.New("value obtained is not an int")

func GetStringValue(m BencodeMap, key string) (string, error) {
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

func GetMapValue(m BencodeMap, key string) (BencodeMap, error) {
	val, keyExists := m[key]
	if !keyExists {
		return nil, ErrMissingMapKey
	}

	mapValue, isValueMap := val.(BencodeMap)
	if !isValueMap {
		return nil, ErrMapValueIsNotMap
	}

	return mapValue, nil
}

func GetIntValue(m BencodeMap, key string) (int, error) {
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
