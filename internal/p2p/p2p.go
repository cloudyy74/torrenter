package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/cloudyy74/torrenter/internal/client"
	"github.com/cloudyy74/torrenter/internal/message"
	"github.com/cloudyy74/torrenter/internal/peer"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers       []peer.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

// Download downloads the torrent. This stores the entire file in memory.
func (t Torrent) Download() ([]byte, error) {
	log.Println("Starting download for", t.Name)

	// create channels for workers to retrieve work and send results
	workCh := make(chan pieceWork, len(t.PieceHashes))
	resultCh := make(chan pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workCh <- pieceWork{index, hash, length}
	}

	// start workers
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workCh, resultCh)
	}

	// collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-resultCh
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		log.Printf("(%0.2f%%) Downloaded piece #%d", percent, res.index)
	}
	close(workCh)

	return buf, nil
}

func (t Torrent) calculateBoundsForPiece(index int) (begin, end int) {
	begin = index * t.PieceLength
	end = min(begin+t.PieceLength, t.Length)
	return begin, end
}

func (t Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func checkIntegrity(pw pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Piece #%d failed integrity check", pw.index)
	}
	return nil
}

func attemptDownloadPiece(c *client.Client, pw pieceWork) ([]byte, error) {
	buf := make([]byte, pw.length)
	downloaded := 0
	requested := 0
	backlog := 0

	// setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for downloaded < pw.length {
		// if unchoked, send requests until we have unfulfilled requests
		if !c.Choked {
			for backlog < MaxBacklog && requested < pw.length {
				// last block might be shorter than the typical block
				blockSize := min(pw.length-requested, MaxBlockSize)

				if err := c.SendRequest(pw.index, requested, blockSize); err != nil {
					return nil, fmt.Errorf("client send request: %w", err)
				}
				backlog++
				requested += blockSize
			}
		}

		msg, err := c.Read()
		if err != nil {
			return nil, fmt.Errorf("client message read: %w", err)
		}

		if msg == nil { // keep-alive
			continue
		}

		switch msg.ID {
		case message.MsgUnchoke:
			c.Choked = false
		case message.MsgChoke:
			c.Choked = true
		case message.MsgHave:
			index, err := message.ParseHave(msg)
			if err != nil {
				return nil, fmt.Errorf("parse have message: %w", err)
			}
			c.Bitfield.SetPiece(index)
		case message.MsgPiece:
			n, err := message.ParsePiece(pw.index, buf, msg)
			if err != nil {
				return nil, fmt.Errorf("parse piece message: %w", err)
			}
			downloaded += n
			backlog--
		}
	}

	return buf, nil
}

func (t Torrent) startDownloadWorker(peer peer.Peer, workCh chan pieceWork, resultCh chan<- pieceResult) {
	c, err := client.New(peer, t.InfoHash, t.PeerID)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workCh {
		if !c.Bitfield.HasPiece(pw.index) {
			workCh <- pw // put piece back to the channel
		}

		// download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("exiting", err)
			workCh <- pw // put piece back to the channel
			return
		}

		if err := checkIntegrity(pw, buf); err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workCh <- pw // put piece back to the channel
		}

		c.SendHave(pw.index)
		resultCh <- pieceResult{pw.index, buf}
	}
}
