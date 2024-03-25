package main

import (
	// Uncomment this line to pass the first stage

	"crypto/sha1"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var downloadOutputPath string

func init() {
	downloadCmd.Flags().StringVarP(&downloadOutputPath, "output", "o", "", "--output path/to/output_file")
	downloadCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:  "download path/to/torrent_file",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log := log.Level(zerolog.DebugLevel)
		filename := args[0]

		outputPath := downloadOutputPath

		torrent, err := ParseTorrent(filename)
		if err != nil {
			fmt.Println("parse torrent: ", err.Error())
			return
		}

		log.Debug().Msgf("%d", torrent.Info.Length)
		log.Debug().Msgf(strings.Join(torrent.Info.PieceHashes, ","))
		log.Debug().Msgf("%d", torrent.Info.PieceLength)

		fileLength := torrent.Info.Length
		// open file for writing
		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("failed to open file for writing: %s\n", err.Error())
			return
		}
		defer file.Close()

		pieceLengths := []int{}
		sumPieceLenghts := 0
		for i := 0; i < torrent.Info.NumPieces; i++ {
			pieceLength := torrent.Info.PieceLength
			if pieceLength+sumPieceLenghts > fileLength {
				pieceLength = fileLength - sumPieceLenghts
			}
			pieceLengths = append(pieceLengths, pieceLength)
			sumPieceLenghts += pieceLength
		}

		log.Debug().Msgf("piece lengths: %s", pretty.Sprint(pieceLengths))

		// reserve space for entire file
		err = file.Truncate(int64(fileLength))
		if err != nil {
			fmt.Printf("failed to truncate file: %s\n", err.Error())
			return
		}

		trackerInfo, err := GetTrackerInfo(torrent)
		if err != nil {
			fmt.Println("get tracker info: ", err.Error())
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
			fmt.Println("net dial: ", err.Error())
			return
		}

		err = SendHandshake(tcpConn, torrent.Info.Sha1Sum())
		if err != nil {
			fmt.Println("send handshake: ", err.Error())
			return
		}

		remotePeerIDBytes, err := ReadHandshakeAck(tcpConn, torrent.Info.Sha1Sum())
		if err != nil {
			fmt.Println("read handshake ack: ", err.Error())
			return
		}

		log.Debug().Msgf("Chosen peer id: %x", remotePeerIDBytes)

		messageLengthBytes := make([]byte, 4)
		messageIdBytes := make([]byte, 1)
		var messageId byte

		numBlocksPerPiece := []int{}
		totalNumBlocks := 0
		for _, pieceLength := range pieceLengths {
			numBlocks := int(math.Ceil(float64(pieceLength) / 16384))
			numBlocksPerPiece = append(numBlocksPerPiece, numBlocks)
			totalNumBlocks += int(numBlocks)
		}

		numBlocksWritten := 0
		blocksPerPiece := make([][][]byte, torrent.Info.NumPieces)
		for i := 0; i < torrent.Info.NumPieces; i++ {
			blocksPerPiece[i] = make([][]byte, numBlocksPerPiece[i])
		}
		for numBlocksWritten < totalNumBlocks {
			_, err = io.ReadAtLeast(tcpConn, messageLengthBytes, 4)
			if err != nil {
				fmt.Println("read message length: ", err.Error())
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
				fmt.Println("read message id: ", err.Error())
				return
			}

			messageId = messageIdBytes[0]
			if messageId == BitfieldMessageID {
				log.Debug().Msg("found bitfield message")

				payload := make([]byte, payloadLength)
				_, err := io.ReadAtLeast(tcpConn, payload, payloadLength)
				if err != nil {
					fmt.Println("read message payload: ", err.Error())
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

				for pieceNumber := 0; pieceNumber < torrent.Info.NumPieces; pieceNumber++ {
					remainingBlockLength := pieceLengths[pieceNumber]
					numBlocks := numBlocksPerPiece[pieceNumber]
					for blockNumber := 0; blockNumber < numBlocks; blockNumber++ {
						// min(x,y) is only go >1.21, and codecrafters are running this on 1.19 it seems
						blockSize := remainingBlockLength
						if remainingBlockLength > 16384 {
							blockSize = 16384
						}
						log.Debug().Msgf("requesting block %d for piece %d, size %d", blockNumber, pieceNumber, blockSize)
						err := RequestPieceBlock(tcpConn, pieceNumber, blockNumber, blockSize)
						if err != nil {
							fmt.Printf("failed to request block %d for piece %d: %q", blockNumber, pieceNumber, err.Error())
							return
						}
						remainingBlockLength -= 16384
						log.Debug().Msgf("requested block %d for piece %d", blockNumber, pieceNumber)
					}
				}
			} else if messageId == PieceMessageID {
				headers := make([]byte, 8)
				blockLength := payloadLength - 8
				if blockLength > 16384 || blockLength < 10 { // 10 should be minimum: 1-byte message_id + 8-byte index/begin + 1 byte block
					fmt.Printf("wrong block length received: %d\n", blockLength)
				}

				_, err := io.ReadAtLeast(tcpConn, headers, 8)
				if err != nil {
					fmt.Println("read piece headers: ", err.Error())
					return
				}

				pieceIndex := byteSliceToInt(headers[0:4])
				log.Debug().Msgf("received block for piece %d", pieceIndex)
				beginOffset := byteSliceToInt(headers[4:8])
				if beginOffset > pieceLengths[pieceIndex] {
					fmt.Printf("begin offset greater than piece length: %d\n", beginOffset)
					return
				}
				log.Debug().Msgf("block begins at %d", beginOffset)
				log.Debug().Msgf("block length is %d", blockLength)

				block := make([]byte, blockLength)
				_, err = io.ReadAtLeast(tcpConn, block, blockLength)
				if err != nil {
					fmt.Println("read block", err.Error())
					return
				}

				filePos := pieceIndex*torrent.Info.PieceLength + beginOffset
				_, err = file.Seek(int64(filePos), 0)
				if err != nil {
					fmt.Println("output file seek: ", err.Error())
					return
				}

				n, err := file.Write(block)
				if err != nil {
					fmt.Println("file write: ", err.Error())
					return
				}
				if n != blockLength {
					fmt.Printf("wrote %d bytes but block length was %d\n", n, blockLength)
					return
				}

				blockNumber := beginOffset / 16384
				blocksPerPiece[pieceIndex][blockNumber] = block
				numBlocksWritten++
				log.Debug().Msgf("wrote block %d for piece %d", blockNumber, pieceIndex)
			} else {
				log.Debug().Msgf("found message id %d\n", messageId)
				break
			}
		}

		for i := 0; i < torrent.Info.NumPieces; i++ {
			var allPieceBytes []byte
			for _, b := range blocksPerPiece[i] {
				allPieceBytes = append(allPieceBytes, b...)
			}

			hasher := sha1.New()
			hasher.Write([]byte(allPieceBytes))
			pieceHash := hasher.Sum(nil)

			if fmt.Sprintf("%x", pieceHash) != torrent.Info.PieceHashes[i] {
				fmt.Printf("hash mismatch, wanted %q, obtained %q\n", torrent.Info.PieceHashes[i], fmt.Sprintf("%x", pieceHash))
				return
			}
			log.Debug().Msgf("Piece %d downloaded to %s\n", i, outputPath)
		}
	},
}
