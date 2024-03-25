package main

import (
	// Uncomment this line to pass the first stage

	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
)

const PeerID = "99887766554433221100"

var HandshakeHeader = append([]byte{19}, []byte("BitTorrent protocol")...)

var rootCmd = &cobra.Command{}

func main() {

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func SendHandshake(tcpConn net.Conn, torrentSha1Sum []byte) error {
	handshake := make([]byte, 68)
	copy(handshake[:20], HandshakeHeader)
	copy(handshake[20:28], []byte{0, 0, 0, 0, 0, 0, 0, 0})
	copy(handshake[28:48], torrentSha1Sum)
	copy(handshake[48:68], []byte(PeerID))

	numBytesWritten, err := tcpConn.Write(handshake)
	if err != nil {
		return err
	}

	if numBytesWritten != 68 {
		return fmt.Errorf("didn't write full ack. num bytes written: %d", numBytesWritten)
	}

	return nil
}

func ReadHandshakeAck(tcpConn net.Conn, torrentSha1Sum []byte) ([]byte, error) {
	ack := make([]byte, 68)

	// tcpConn.SetDeadline(time.Now().Add(5 * time.Second))
	// numBytesRead, err := tcpConn.Read(ack)
	_, err := io.ReadAtLeast(tcpConn, ack, 68)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if !bytes.Equal(HandshakeHeader, ack[0:20]) {
		return nil, fmt.Errorf("invalid handshake ack header: %v", ack[0:20])
	}

	if !bytes.Equal(ack[28:48], torrentSha1Sum) {
		return nil, fmt.Errorf("invalid info hash in handshake ack")
	}

	remotePeerID := ack[48:68]

	return remotePeerID, nil
}

func ParseTorrent(filename string) (*TorrentFile, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	encoded := string(fileBytes)
	decoded, err := DecodeBencode(encoded)
	if err != nil {
		return nil, err
	}

	if decodedMap, ok := decoded.Output.(BencodeMap); ok {
		announce, err := GetStringValue(decodedMap, "announce")
		if err != nil {
			return nil, err
		}

		infoMap, err := GetMapValue(decodedMap, "info")
		if err != nil {
			return nil, err
		}

		infoFileLength, err := GetIntValue(infoMap, "length")
		if err != nil {
			return nil, err
		}

		pieceLength, err := GetIntValue(infoMap, "piece length")
		if err != nil {
			return nil, err
		}

		piecesString, err := GetStringValue(infoMap, "pieces")
		if err != nil {
			return nil, err
		}

		piecesBytes := []byte(piecesString)
		piecesFullLength := len(piecesBytes)
		if piecesFullLength%20 != 0 {
			return nil, fmt.Errorf("invalid piece hashes length: %d", piecesFullLength)
		}

		info := &TorrentInfo{
			decodedMap:   infoMap,
			Length:       infoFileLength,
			PieceLength:  pieceLength,
			piecesString: piecesString,
		}
		info.PieceHashes = info.parsePieceHashes()
		info.NumPieces = len(info.PieceHashes)

		return &TorrentFile{
			Announce: announce,
			Info:     info,
		}, nil
	}

	return nil, errors.New("expected top level dict")
}

type TorrentFile struct {
	Announce string
	Info     *TorrentInfo
}

type TorrentInfo struct {
	decodedMap   BencodeMap
	Length       int
	PieceLength  int
	PieceHashes  []string
	NumPieces    int
	piecesString string
}

func (info *TorrentInfo) parsePieceHashes() []string {
	piecesBytes := []byte(info.piecesString)
	piecesFullLength := len(piecesBytes)

	piecesHashes := []string{}
	for i := 0; i < piecesFullLength/20; i++ {
		piecesHashes = append(piecesHashes, fmt.Sprintf("%x", piecesBytes[i*20:(i+1)*20]))
	}
	return piecesHashes
}

func (info *TorrentInfo) Sha1Sum() []byte {
	encodedInfo, err := EncodeBencode(info.decodedMap)
	if err != nil {
		panic(err)
	}

	hasher := sha1.New()
	hasher.Write([]byte(encodedInfo))
	return hasher.Sum(nil)
}
