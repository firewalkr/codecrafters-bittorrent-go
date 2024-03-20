package main

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

var ErrUnsupported = errors.New("only strings and integers are supported at the moment")
var ErrUnterminatedList = errors.New("found incomplete list")
var ErrNegativeZero = errors.New("cannot have negative zero")
var ErrZeroPrefixedInteger = errors.New("invalid integer, cannot have zero-prefixed integers")
var ErrInvalidInteger = errors.New("unparseable integer")
var ErrUnterminatedDictionary = errors.New("found incomplete dictionary")
var ErrInvalidDictionaryKey = errors.New("invalid dictionary key, must be string")
var ErrUnsupportedType = errors.New("type is unsupported by bencode")

type BencodeList []any
type BencodeMap map[string]any

type DecodedToken struct {
	Output      any
	InputLength int
}

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func DecodeBencode(bencodedString string) (*DecodedToken, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, err
		}

		return &DecodedToken{
			Output:      bencodedString[firstColonIndex+1 : firstColonIndex+1+length],
			InputLength: firstColonIndex + 1 + length,
		}, nil
	} else if bencodedString[0] == 'i' { // integer
		integerEndIndex := strings.IndexRune(bencodedString, 'e')

		possibleIntegerStr := bencodedString[1:integerEndIndex]

		if possibleIntegerStr[0] == '-' && possibleIntegerStr[1] == '0' {
			if len(possibleIntegerStr) > 2 && unicode.IsDigit(rune(possibleIntegerStr[2])) {
				return nil, ErrZeroPrefixedInteger
			}
			return nil, ErrNegativeZero
		}

		if possibleIntegerStr[0] == '0' && len(possibleIntegerStr) > 1 {
			return nil, ErrZeroPrefixedInteger
		}

		integer, err := strconv.Atoi(possibleIntegerStr)
		if err != nil {
			return nil, ErrInvalidInteger
		}

		return &DecodedToken{
			Output:      integer,
			InputLength: integerEndIndex + 1,
		}, nil
	} else if bencodedString[0] == 'l' {
		list := BencodeList{}
		if len(bencodedString) < 2 {
			return nil, ErrUnterminatedList
		}

		totalInputLength := 1
		nextString := bencodedString[1:]
		for {
			if len(nextString) == 0 {
				return nil, ErrUnterminatedList
			}
			if nextString[0] == 'e' {
				totalInputLength += 1
				break
			}
			decodedToken, err := DecodeBencode(nextString)
			if err != nil {
				return nil, err
			}
			list = append(list, decodedToken.Output)
			totalInputLength += decodedToken.InputLength
			nextString = nextString[decodedToken.InputLength:]
		}

		return &DecodedToken{
			Output:      list,
			InputLength: totalInputLength,
		}, nil
	} else if bencodedString[0] == 'd' {
		resultMap := BencodeMap{}
		if len(bencodedString) < 2 {
			return nil, ErrUnterminatedDictionary
		}
		expectingKeyOrTerminator := true
		currentKey := ""
		nextString := bencodedString[1:]
		totalInputLength := 1
		for {
			if len(nextString) == 0 {
				return nil, ErrUnterminatedDictionary
			}
			if !expectingKeyOrTerminator && nextString[0] == 'e' {
				return nil, ErrUnterminatedDictionary
			}
			if nextString[0] == 'e' {
				totalInputLength++
				break
			}
			decodedToken, err := DecodeBencode(nextString)
			if err != nil {
				return nil, err
			}
			if expectingKeyOrTerminator {
				maybeKey, isString := decodedToken.Output.(string)
				if !isString {
					return nil, ErrInvalidDictionaryKey
				}
				currentKey = maybeKey
				expectingKeyOrTerminator = false
			} else {
				resultMap[currentKey] = decodedToken.Output
				expectingKeyOrTerminator = true
			}
			totalInputLength += decodedToken.InputLength
			nextString = nextString[decodedToken.InputLength:]
		}
		return &DecodedToken{Output: resultMap, InputLength: totalInputLength}, nil
	} else {
		return nil, ErrUnsupported
	}
}

func EncodeBencode(input any) (string, error) {
	if str, isString := input.(string); isString {
		if str == "" {
			return "", nil
		}
		return fmt.Sprintf("%d:%s", len(str), str), nil
	}

	if intNumber, isInt := input.(int); isInt {
		return fmt.Sprintf("i%de", intNumber), nil
	}

	if list, isList := input.(BencodeList); isList {
		b := strings.Builder{}
		b.WriteRune('l')
		for _, elem := range list {
			encodedElem, err := EncodeBencode(elem)
			if err != nil {
				return "", err
			}
			b.WriteString(encodedElem)
		}
		b.WriteRune('e')
		return b.String(), nil
	}

	if aMap, isMap := input.(BencodeMap); isMap {
		b := strings.Builder{}
		b.WriteRune('d')

		keys := []string{}
		for k := range aMap {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, k := range keys {
			encodedKey, err := EncodeBencode(k)
			if err != nil {
				return "", err
			}
			b.WriteString(encodedKey)
			encodedValue, err := EncodeBencode(aMap[k])
			if err != nil {
				return "", err
			}
			b.WriteString(encodedValue)
		}
		b.WriteRune('e')
		return b.String(), nil
	}

	return "", ErrUnsupportedType
}
