package main

import (
	// Uncomment this line to pass the first stage

	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const PeerID = "99887766554433221100"

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

			// pieceLength, err := GetIntValue(infoMap, "piece length")
			// if err != nil {
			// 	fmt.Println(err.Error())
			// 	return
			// }

			piecesString, err := GetStringValue(infoMap, "pieces")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			piecesBytes := []byte(piecesString)
			piecesFullLength := len(piecesBytes)
			if piecesFullLength%20 != 0 {
				fmt.Printf("Invalid piece hashes length: %d\n", piecesFullLength)
				return
			}

			piecesHashes := []string{}
			for i := 0; i < piecesFullLength/20; i++ {
				piecesHashes = append(piecesHashes, fmt.Sprintf("%x", piecesBytes[i*20:(i+1)*20]))
			}

			encodedInfo, err := EncodeBencode(infoMap)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			// fmt.Printf("Tracker URL: %s\n", announce)
			// fmt.Printf("Length: %d\n", infoFileLength)

			hasher := sha1.New()
			hasher.Write([]byte(encodedInfo))
			infoSha1Sum := hasher.Sum(nil)

			// fmt.Printf("Info Hash: %s\n", fmt.Sprintf("%x", infoSha1Sum))
			// fmt.Printf("Piece Length: %d\n", pieceLength)
			// fmt.Printf("Piece Hashes:\n%s\n", strings.Join(piecesHashes, "\n"))

			httpClient := &http.Client{}

			trackerURL, err := url.Parse(announce)
			if err != nil {
				fmt.Println("invalid tracker URL")
				return
			}
			queryParams := url.Values{}
			queryParams.Add("info_hash", string(infoSha1Sum))
			queryParams.Add("peer_id", PeerID)
			queryParams.Add("port", "6881")
			queryParams.Add("uploaded", "0")
			queryParams.Add("downloaded", "0")
			queryParams.Add("left", strconv.Itoa(infoFileLength))
			queryParams.Add("compact", "1")
			trackerURL.RawQuery = queryParams.Encode()

			httpResponse, err := httpClient.Get(trackerURL.String())
			if err != nil {
				fmt.Printf("http error calling tracker: %s\n", err.Error())
				return
			}
			defer httpResponse.Body.Close()
			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				fmt.Printf("http error reading tracker response body: %s\n", err.Error())
				return
			}

			decodedBody, err := DecodeBencode(string(body))
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			responseMap, isMap := decodedBody.Output.(BencodeMap)
			if !isMap {
				fmt.Println("failed to obtain map from tracker response")
				return
			}

			peerListString, err := GetStringValue(responseMap, "peers")
			if err != nil {
				fmt.Println("failed to read peer list as string")
				return
			}

			peerListBytes := []byte(peerListString)
			if len(peerListBytes)%6 != 0 {
				fmt.Println("Invalid compact peer list length")
				return
			}

			numPeers := len(peerListBytes) / 6
			peerURLs := []string{}
			for i := 0; i < numPeers; i++ {
				ipParts := []string{}
				for b := i * 6; b < (i*6)+4; b++ {
					ipParts = append(ipParts, strconv.Itoa(int(peerListBytes[b])))
				}
				port := 256*int(peerListBytes[i*6+4]) + int(peerListBytes[i*6+5])
				peerURLs = append(peerURLs, strings.Join(ipParts, ".")+":"+strconv.Itoa(port))
			}

			fmt.Printf("%s\n", strings.Join(peerURLs, "\n"))
		} else {
			fmt.Println("expected top level dict")
			return
		}

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
