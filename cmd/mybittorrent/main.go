package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
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
	} else if bencodedString[0] == 'i' && bencodedString[bencodedLength-1] == 'e' && bencodedLength > 2 {
		possibleIntegerStr := bencodedString[1 : bencodedLength-1]

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

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
