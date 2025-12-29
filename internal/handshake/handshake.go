package handshake

import (
	"errors"
	"fmt"
	"io"
)

// Handshake is a special message that a peer uses to identify itself
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// New creates a new handshake with the standard pstr
func New(infoHash, peerID [20]byte) Handshake {
	h := Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
	return h
}

// Serialize serializes the handshake to a buffer
func (h Handshake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	cur := 1
	cur += copy(buf[cur:], h.Pstr)
	cur += copy(buf[cur:], make([]byte, 8)) // 8 reserved bytes for extensions (not supported)
	cur += copy(buf[cur:], h.InfoHash[:])
	cur += copy(buf[cur:], h.PeerID[:])
	return buf
}

// Read parses a handshake from a stream
func Read(r io.Reader) (Handshake, error) {
	lengthBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, lengthBuf); err != nil {
		return Handshake{}, fmt.Errorf("read pstr length: %w", err)
	}

	pstrLength := int(lengthBuf[0])
	if pstrLength == 0 {
		return Handshake{}, errors.New("pstr length cannot be 0")
	}

	handshakeBuf := make([]byte, pstrLength+48)
	if _, err := io.ReadFull(r, handshakeBuf); err != nil {
		return Handshake{}, fmt.Errorf("read handshake: %w", err)
	}

	pstr := string(handshakeBuf[:pstrLength])

	var infoHash, peerID [20]byte
	copy(infoHash[:], handshakeBuf[pstrLength+8:pstrLength+8+20])
	copy(peerID[:], handshakeBuf[pstrLength+8+20:])

	h := Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerID:   peerID,
	}
	return h, nil
}
