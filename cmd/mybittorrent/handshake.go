package main

import (
	// Uncomment this line to pass the first stage

	"fmt"
	"net"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(handshakeCmd)
}

var handshakeCmd = &cobra.Command{
	Use:  "handshake",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		peerAddress := args[1]
		// peerParts := strings.Split(peerAddress, ":")
		// peerIP, peerPort := peerParts[0], peerParts[1]

		torrentFile, err := ParseTorrent(filename)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		tcpConn, err := net.Dial("tcp", peerAddress)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer tcpConn.Close()

		torrentSha1Sum := torrentFile.Info.Sha1Sum()

		if err := SendHandshake(tcpConn, torrentSha1Sum); err != nil {
			fmt.Println(err.Error())
			return
		}

		trackerPeerID, err := ReadHandshakeAck(tcpConn, torrentSha1Sum)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("Peer ID: %x\n", trackerPeerID)
	},
}
