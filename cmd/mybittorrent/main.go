package main

import (
	// Uncomment this line to pass the first stage

	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := DecodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded.Output)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		filename := os.Args[2]
		fileBytes, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println(err)
			return
		}

		encoded := string(fileBytes)
		decoded, err := DecodeBencode(encoded)
		if err != nil {
			fmt.Println(err)
			return
		}

		if decodedMap, ok := decoded.Output.(BencodeMap); ok {
			announce, err := GetStringValue(decodedMap, "announce")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			infoMap, err := GetMapValue(decodedMap, "info")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			infoFileLength, err := GetIntValue(infoMap, "length")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			encodedInfo, err := EncodeBencode(infoMap)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("Tracker URL: %s\n", announce)
			fmt.Printf("Length: %d\n", infoFileLength)

			hasher := sha1.New()
			hasher.Write([]byte(encodedInfo))
			sum := hasher.Sum(nil)

			fmt.Printf("Info Hash: %s\n", fmt.Sprintf("%x", sum))

		} else {
			fmt.Println("expected top level dict")
			return
		}

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
