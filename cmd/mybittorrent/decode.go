package main

import (
	// Uncomment this line to pass the first stage

	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(decodeCmd)
}

var decodeCmd = &cobra.Command{
	Use:  "decode",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bencodedValue := args[0]

		decoded, err := DecodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded.Output)
		fmt.Println(string(jsonOutput))
	},
}
