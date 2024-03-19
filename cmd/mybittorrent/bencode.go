package main

import (
	// Uncomment this line to pass the first stage

	"fmt"
	"strconv"
	"strings"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func DecodeBencode(bencodedString string) (any, error) {
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
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else if bencodedString[0] == 'i' { // integer
		integerEndIndex := strings.IndexRune(bencodedString, 'e')

		possibleIntegerStr := bencodedString[1:integerEndIndex]

		if possibleIntegerStr[0] == '-' && possibleIntegerStr[1] == '0' {
			return "", fmt.Errorf("invalid integer, cannot have negative zero or zero-prefixed integers")
		}

		if possibleIntegerStr[0] == '0' && len(possibleIntegerStr) > 1 {
			return "", fmt.Errorf("invalid integer, cannot have zero-prefixed integers")
		}

		integer, err := strconv.Atoi(possibleIntegerStr)
		if err != nil {
			return "", fmt.Errorf("failed to parse integer value %q: %s", possibleIntegerStr, err.Error())
		}

		return integer, nil
	} else {
		return "", fmt.Errorf("only strings and integers are supported at the moment")
	}
}
