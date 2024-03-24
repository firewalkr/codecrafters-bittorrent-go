package main

import (
	// Uncomment this line to pass the first stage

	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:  "info",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		torrent, err := ParseTorrent(filename)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("Tracker URL: %s\n", torrent.Announce)
		fmt.Printf("Length: %d\n", torrent.Info.Length)

		infoSha1Sum := torrent.Info.Sha1Sum()

		fmt.Printf("Info Hash: %s\n", fmt.Sprintf("%x", infoSha1Sum))
		fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)

		piecesHashes := torrent.Info.PieceHashes()
		fmt.Printf("Piece Hashes:\n%s\n", strings.Join(piecesHashes, "\n"))
	},
}
