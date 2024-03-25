package main

import (
	// Uncomment this line to pass the first stage

	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var downloadOutputPath string

const (
	ChokeMessageID byte = iota
	UnchokeMessageID
	InterestedMessageID
	NotInterestedMessageID
	HaveMessageID
	BitfieldMessageID
	RequestMessageID
	PieceMessageID
	CancelMessageID
)

func init() {
	downloadPieceCmd.Flags().StringVarP(&downloadOutputPath, "output", "o", "", "--output path/to/output_file")
	downloadPieceCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(downloadPieceCmd)
}

var downloadPieceCmd = &cobra.Command{
	Use:  "download_piece path/to/torrent_file piece_index",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		log := log.Level(zerolog.DebugLevel)
		filename := args[0]
		pieceIndexStr := args[1]

		requestedPieceIndex, err := strconv.Atoi(pieceIndexStr)
		if err != nil {
			fmt.Printf("failed to convert index to int: %q\n", pieceIndexStr)
			return
		}
		outputPath := downloadOutputPath

		torrent, err := ParseTorrent(filename)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		log.Debug().Msgf("%d", torrent.Info.Length)
		log.Debug().Msgf(strings.Join(torrent.Info.PieceHashes(), ","))
		log.Debug().Msgf("%d", torrent.Info.PieceLength)

		// fileLength := torrent.Info.Length
		// open file for writing
		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("failed to open file for writing: %s\n", err.Error())
			return
		}
		defer file.Close()

		// reserve space for 1 piece length
		err = file.Truncate(int64(torrent.Info.PieceLength))
		if err != nil {
			fmt.Printf("failed to truncate file: %s\n", err.Error())
			return
		}

		trackerInfo, err := GetTrackerInfo(torrent)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		numPeers := len(trackerInfo.Peers)
		if numPeers == 0 {
			fmt.Println("no peers found")
			return
		}

		randomPeerAddress := trackerInfo.Peers[rand.Intn(numPeers)]

		log.Debug().Msgf("Chosen peer: %s", randomPeerAddress)

		tcpConn, err := net.Dial("tcp", randomPeerAddress)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		err = SendHandshake(tcpConn, torrent.Info.Sha1Sum())
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		remotePeerIDBytes, err := ReadHandshakeAck(tcpConn, torrent.Info.Sha1Sum())
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		log.Debug().Msgf("Chosen peer id: %x", remotePeerIDBytes)

		messageLengthBytes := make([]byte, 4)
		messageIdBytes := make([]byte, 1)
		var messageId byte

		pieceLength := torrent.Info.PieceLength
		numBlocks := pieceLength / 16384
		blocks := make([][]byte, numBlocks)
		numBlocksWritten := 0

		for numBlocksWritten < numBlocks {
			_, err = io.ReadAtLeast(tcpConn, messageLengthBytes, 4)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			messageLength := byteSliceToInt(messageLengthBytes)
			log.Debug().Msgf("message length: %d", messageLength)

			if messageLength == 0 {
				fmt.Println("got keepalive")
				continue
			}

			payloadLength := messageLength - 1

			_, err = io.ReadAtLeast(tcpConn, messageIdBytes, 1)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			messageId = messageIdBytes[0]
			if messageId == BitfieldMessageID {
				log.Debug().Msg("found bitfield message")

				payload := make([]byte, payloadLength)
				_, err := io.ReadAtLeast(tcpConn, payload, payloadLength)
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				log.Debug().Msg(pretty.Sprint(payload))

				err = SendInterested(tcpConn)
				if err != nil {
					fmt.Printf("failed to send Interested message: %s\n", err.Error())
					return
				}
			} else if messageId == ChokeMessageID {
				log.Debug().Msgf("found choke message")
				time.Sleep(3 * time.Second)
			} else if messageId == UnchokeMessageID {
				log.Debug().Msgf("got unchoke message")

				remainingLength := pieceLength
				for blockNumber := 0; blockNumber < numBlocks; blockNumber++ {
					log.Debug().Msgf("requesting block %d for piece %d", blockNumber, requestedPieceIndex)
					// min(x,y) is only go >1.21, and codecrafters are running this on 1.19 it seems
					blockSize := remainingLength
					if remainingLength > 16384 {
						blockSize = 16384
					}
					err := RequestPieceBlock(tcpConn, requestedPieceIndex, blockNumber, blockSize)
					if err != nil {
						fmt.Printf("failed to request block %d for piece %d: %q\n", blockNumber, requestedPieceIndex, err.Error())
						return
					}
					remainingLength -= 16384
					log.Debug().Msgf("requested block %d for piece %d\n", blockNumber, requestedPieceIndex)
				}
			} else if messageId == PieceMessageID {
				headers := make([]byte, 8)
				blockLength := payloadLength - 8
				if blockLength > 16384 || blockLength < 10 { // 10 should be minimum: 1-byte message_id + 8-byte index/begin + 1 byte block
					fmt.Printf("wrong block length received: %d\n", blockLength)
				}

				_, err := io.ReadAtLeast(tcpConn, headers, 8)
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				pieceIndex := byteSliceToInt(headers[0:4])
				if pieceIndex != requestedPieceIndex {
					fmt.Printf("requested piece index %d, but got a block for piece index %d\n", requestedPieceIndex, pieceIndex)
					return
				}
				beginOffset := byteSliceToInt(headers[4:8])
				if beginOffset > pieceLength {
					fmt.Printf("begin offset greater than piece length: %d\n", beginOffset)
					return
				}

				block := make([]byte, blockLength)
				_, err = io.ReadAtLeast(tcpConn, block, blockLength)
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				_, err = file.Seek(int64(beginOffset), 0)
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				n, err := file.Write(block)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				if n != blockLength {
					fmt.Printf("wrote %d bytes but block length was %d\n", n, blockLength)
					return
				}

				blockNumber := beginOffset / blockLength
				blocks[blockNumber] = block
				numBlocksWritten++
				log.Debug().Msgf("wrote block %d", blockNumber)
			} else {
				log.Debug().Msgf("found message id %d\n", messageId)
				break
			}
		}

		fmt.Printf("Piece %d downloaded to %s\n", requestedPieceIndex, outputPath)
	},
}

func byteSliceToInt(b []byte) int {
	bLen := len(b)
	res := 0
	power := float64(bLen - 1)
	for i := 0; i < bLen; i++ {
		res += int(math.Pow(256, power)) * int(b[i])
		power--
	}
	return res
}

func intToByteSlice(i int) []byte {
	res := make([]byte, 4)
	for b := 3; b >= 0; b-- {
		res[b] = byte(i & 255)
		i = i >> 8
	}
	return res
}

func SendInterested(tcpConn net.Conn) error {
	numBytesWritten, err := tcpConn.Write([]byte{0, 0, 0, 1, InterestedMessageID})
	if err != nil {
		return err
	}

	if numBytesWritten != 5 {
		return fmt.Errorf("didn't write full ack. num bytes written: %d", numBytesWritten)
	}

	return nil
}

func RequestPieceBlock(tcpConn net.Conn, pieceIndex int, blockIndex int, blockSize int) error {
	// length 13 = message id + 3 32-bit integers (index, begin, length)
	indexBytes := intToByteSlice(pieceIndex)
	beginBytes := intToByteSlice(blockIndex * 16384)
	blockBytes := intToByteSlice(blockSize)

	payload := []byte{0, 0, 0, 13, RequestMessageID}
	payload = append(payload, indexBytes...)
	payload = append(payload, beginBytes...)
	payload = append(payload, blockBytes...)
	numBytesWritten, err := tcpConn.Write(payload)
	if err != nil {
		return err
	}

	if numBytesWritten != 17 {
		return fmt.Errorf("didn't write full ack. num bytes written: %d", numBytesWritten)
	}

	return nil
}
