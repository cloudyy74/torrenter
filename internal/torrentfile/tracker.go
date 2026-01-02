package torrentfile

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cloudyy74/torrenter/internal/peer"
	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (t TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", fmt.Errorf("parse announce: %w", err)
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (t TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peer.Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, fmt.Errorf("build tracker url: %w", err)
	}

	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get from tracker: %w", err)
	}
	defer resp.Body.Close()

	var trackerResp bencodeTrackerResp
	if err := bencode.Unmarshal(resp.Body, &trackerResp); err != nil {
		return nil, fmt.Errorf("bencode unmarshal response: %w", err)
	}

	return peer.Unmarshal([]byte(trackerResp.Peers))
}
