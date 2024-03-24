package main

import (
	// Uncomment this line to pass the first stage

	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(peersCmd)
}

type TrackerInfo struct {
	Complete    int
	Incomplete  int
	Interval    int
	MinInterval int
	Peers       []string
}

var peersCmd = &cobra.Command{
	Use:  "peers",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		torrent, err := ParseTorrent(filename)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		trackerInfo, err := GetTrackerInfo(torrent)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println(strings.Join(trackerInfo.Peers, "\n"))
	},
}

func GetTrackerInfo(torrent *TorrentFile) (*TrackerInfo, error) {
	httpClient := &http.Client{}

	trackerURL, err := url.Parse(torrent.Announce)
	if err != nil {
		return nil, fmt.Errorf("invalid tracker URL")
	}
	queryParams := url.Values{}
	queryParams.Add("info_hash", string(torrent.Info.Sha1Sum()))
	queryParams.Add("peer_id", PeerID)
	queryParams.Add("port", "6881")
	queryParams.Add("uploaded", "0")
	queryParams.Add("downloaded", "0")
	queryParams.Add("left", strconv.Itoa(torrent.Info.Length))
	queryParams.Add("compact", "1")
	trackerURL.RawQuery = queryParams.Encode()

	httpResponse, err := httpClient.Get(trackerURL.String())
	if err != nil {
		return nil, fmt.Errorf("http error calling tracker: %s", err.Error())
	}
	defer httpResponse.Body.Close()
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("http error reading tracker response body: %s", err.Error())
	}

	decodedBody, err := DecodeBencode(string(body))
	if err != nil {
		return nil, err
	}

	responseMap, isMap := decodedBody.Output.(BencodeMap)
	if !isMap {
		return nil, fmt.Errorf("failed to obtain map from tracker response")
	}

	peerURLs, err := parsePeers(responseMap)
	if err != nil {
		return nil, err
	}

	complete, err := GetIntValue(responseMap, "complete")
	if err != nil {
		return nil, err
	}

	incomplete, err := GetIntValue(responseMap, "incomplete")
	if err != nil {
		return nil, err
	}

	interval, err := GetIntValue(responseMap, "interval")
	if err != nil {
		return nil, err
	}

	minInterval, err := GetIntValue(responseMap, "min interval")
	if err != nil {
		return nil, err
	}

	return &TrackerInfo{
		Complete:    complete,
		Incomplete:  incomplete,
		Interval:    interval,
		MinInterval: minInterval,
		Peers:       peerURLs,
	}, nil
}

func parsePeers(trackerResponseMap BencodeMap) ([]string, error) {
	peerListString, err := GetStringValue(trackerResponseMap, "peers")
	if err != nil {
		return nil, fmt.Errorf("failed to read peer list as string")
	}

	peerListBytes := []byte(peerListString)
	if len(peerListBytes)%6 != 0 {
		return nil, fmt.Errorf("invalid compact peer list length")
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

	return peerURLs, nil
}
