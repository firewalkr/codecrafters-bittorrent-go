package main

import (
	// Uncomment this line to pass the first stage

	"fmt"
	"strconv"
	"strings"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

type List []any

type DecodedToken struct {
	Output      any
	InputLength int
}

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func DecodeBencode(bencodedString string) (*DecodedToken, error) {
	bencodedLength := len(bencodedString)

	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < bencodedLength; i++ {
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
			return nil, fmt.Errorf("invalid integer, cannot have negative zero or zero-prefixed integers")
		}

		if possibleIntegerStr[0] == '0' && len(possibleIntegerStr) > 1 {
			return nil, fmt.Errorf("invalid integer, cannot have zero-prefixed integers")
		}

		integer, err := strconv.Atoi(possibleIntegerStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer value %q: %s", possibleIntegerStr, err.Error())
		}

		return &DecodedToken{
			Output:      integer,
			InputLength: integerEndIndex + 1,
		}, nil
	} else if bencodedString[0] == 'l' {
		list := List{}
		if bencodedLength < 2 {
			return nil, fmt.Errorf("found incomplete list")
		}

		totalInputLength := 1
		nextString := bencodedString[1:]
		for {
			if len(nextString) == 0 {
				return nil, fmt.Errorf("list with elements but missing terminator")
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
	} else {
		return nil, fmt.Errorf("only strings and integers are supported at the moment")
	}
}
