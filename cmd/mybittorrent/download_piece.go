package main

import (
	// Uncomment this line to pass the first stage

	"fmt"

	"github.com/spf13/cobra"
)

var downloadOutputPath string

func init() {
	downloadPieceCmd.Flags().StringVarP(&downloadOutputPath, "output", "o", "", "--output path/to/output_file")
	downloadPieceCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(downloadPieceCmd)
}

var downloadPieceCmd = &cobra.Command{
	Use:  "download_piece path/to/torrent_file piece_index",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		
		// pieceIndex, err := strconv.Atoi(args[1])
		// if err != nil {
		// 	return err
		// }
		// outputPath := downloadOutputPath

		_, err := ParseTorrent(filename)
		if err != nil {
			fmt.Println(err)
			return err
		}

		return nil
	},
}
