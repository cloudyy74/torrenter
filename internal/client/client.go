package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/cloudyy74/torrenter/internal/bitfield"
	"github.com/cloudyy74/torrenter/internal/handshake"
	"github.com/cloudyy74/torrenter/internal/message"
	"github.com/cloudyy74/torrenter/internal/peer"
)

// Client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peer.Peer
	infoHash [20]byte
	peerID   [20]byte
}

// New connects with a peer, completes a handshake, and receives a handshake
// returns an err if any of those fail.
func New(peer peer.Peer, infoHash, peerID [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to peer %s: %w", peer, err)
	}

	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake with peer %s: %w", peer, err)
	}

	bf, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("receive bitfield from peer %s: %w", peer, err)
	}

	c := &Client{
		Conn: conn,
		Choked: true,
		Bitfield: bf,
		peer: peer,
		infoHash: infoHash,
		peerID: peerID,
	}
	return c, nil
}

func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

func completeHandshake(conn net.Conn, infoHash, peerID [20]byte) (handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := handshake.New(infoHash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return handshake.Handshake{}, fmt.Errorf("write handshake request: %w", err)
	}

	resp, err := handshake.Read(conn)
	if err != nil {
		return handshake.Handshake{}, fmt.Errorf("read handshake response: %w", err)
	}
	if !bytes.Equal(resp.InfoHash[:], infoHash[:]) {
		return handshake.Handshake{}, fmt.Errorf("expected infohash %x but got %x", resp.InfoHash, infoHash)
	}

	return resp, nil
}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	msg, err := message.Read(conn)
	if err != nil {
		return nil, fmt.Errorf("read bitfield message: %w", err)
	}
	if msg == nil {
		return nil, fmt.Errorf("expected bitfield but got %s", msg)
	}
	if msg.ID != message.MsgBitfield {
		return nil, fmt.Errorf("expected bitfield (ID = 5) but got ID %d", msg.ID)
	}

	return msg.Payload, nil
}