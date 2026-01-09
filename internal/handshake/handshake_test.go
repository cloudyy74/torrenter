package handshake

import (
	"bytes"
	"testing"
)

type serializeTest struct {
	input Handshake
	want  []byte
}

type readTest struct {
	input []byte
	want  Handshake
	fails bool
}

func TestNew(t *testing.T) {
	infoHash := [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116}
	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	got := New(infoHash, peerID)
	want := Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	if got != want {
		t.Errorf("got %+v, wanted %+v", got, want)
	}
}

func TestSerialize(t *testing.T) {
	tests := []serializeTest{
		{
			input: Handshake{
				Pstr:     "BitTorrent protocol",
				InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
				PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			want: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
		{
			input: Handshake{
				Pstr:     "BitTorrent protocol, but cooler?",
				InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
				PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			want: []byte{32, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 44, 32, 98, 117, 116, 32, 99, 111, 111, 108, 101, 114, 63, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
	}

	for _, tt := range tests {
		got := tt.input.Serialize()
		if !bytes.Equal(got, tt.want) {
			t.Errorf("got %+v, wanted %+v", got, tt.want)
		}
	}
}

func TestRead(t *testing.T) {
	tests := []readTest{
		{
			input: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			want: Handshake{
				Pstr:     "BitTorrent protocol",
				InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
				PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			fails: false,
		},
		{
			input: []byte{},
			want:  Handshake{},
			fails: true,
		},
		{
			input:  []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111},
			want: Handshake{},
			fails:  true,
		},
		{
			input:  []byte{0, 0, 0},
			want: Handshake{},
			fails:  true,
		},
	}

	for _, tt := range tests {
		r := bytes.NewReader(tt.input)
		h, err := Read(r)
		if err != nil && !tt.fails {
			t.Errorf("not expected error: %s", err)
		}
		if err == nil && tt.fails {
			t.Errorf("expected error, got nil")
		}
		if h != tt.want {
			t.Errorf("got %+v, want %+v", h, tt.want)
		}
	}
}
