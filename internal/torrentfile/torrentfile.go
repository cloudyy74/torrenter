package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/cloudyy74/torrenter/internal/p2p"
	"github.com/jackpal/bencode-go"
)

// TorrentFile encodes the metadata from a .torrent file
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

// DownloadToFile downloads a torrent and writes it to a file
func (t TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return fmt.Errorf("generate peerID: %w", err)
	}

	peers, err := t.requestPeers(peerID, 1111)
	if err != nil {
		return fmt.Errorf("request peers: %w", err)
	}

	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}

	buf, err := torrent.Download()
	if err != nil {
		return fmt.Errorf("torrent download: %w", err)
	}

	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()

	if _, err := outFile.Write(buf); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	return nil
}

// Open parses a torrent file
func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	bto := bencodeTorrent{}
	if err := bencode.Unmarshal(file, &bto); err != nil {
		return TorrentFile{}, fmt.Errorf("decode file: %w", err)
	}
	return bto.toTorrentFile()
}

func (bi bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, bi); err != nil {
		return [20]byte{}, fmt.Errorf("marshal info dictionary: %w", err)
	}

	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (bi bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // length of SHA-1 hash
	buf := []byte(bi.Pieces)
	if len(buf)%hashLen != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(buf))
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := range numHashes {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, fmt.Errorf("compute info hash: %w", err)
	}

	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, fmt.Errorf("parse piece hashes: %w", err)
	}

	t := TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return t, nil
}
