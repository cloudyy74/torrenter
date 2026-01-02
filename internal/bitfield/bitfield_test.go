package bitfield

import (
	"reflect"
	"testing"
)

type hasPieceTest struct {
	input Bitfield
	want  []bool // want[i] is a result for input.HasPiece(i)
}

type setPieceTest struct {
	input Bitfield
	index int
	want  Bitfield
}

func TestHasPiece(t *testing.T) {
	tests := []hasPieceTest{
		{
			input: Bitfield{},
			want:  []bool{false, false, false},
		},
		{
			input: Bitfield{0b10100110},
			want:  []bool{true, false, true, false, false, true, true, false, false, false},
		},
		{
			input: Bitfield{0b01010100, 0b01010100},
			want:  []bool{false, true, false, true, false, true, false, false, false, true, false, true, false, true, false, false, false, false, false, false},
		},
	}

	for _, tt := range tests {
		resultLen := len(tt.want)
		result := make([]bool, resultLen)
		for i := range resultLen {
			result[i] = tt.input.HasPiece(i)
		}
		if !reflect.DeepEqual(result, tt.want) {
			t.Errorf("bitfield: %v, wanted: %v, got: %v", tt.input, tt.want, result)
		}
	}
}

func TestSetPiece(t *testing.T) {
	tests := []setPieceTest{
		{
			input: Bitfield{0b01010100, 0b01010100},
			index: 4, //                             v (set)
			want:  Bitfield{0b01011100, 0b01010100},
		},
		{
			input: Bitfield{0b01010100, 0b01010100},
			index: 9, //                             v (noop)
			want:  Bitfield{0b01010100, 0b01010100},
		},
		{
			input: Bitfield{0b01010100, 0b01010100},
			index: 15, //                            v (set)
			want:  Bitfield{0b01010100, 0b01010101},
		},
		{
			input: Bitfield{0b01010100, 0b01010100},
			index: 19, //                            v (noop)
			want:  Bitfield{0b01010100, 0b01010100},
		},
	}

	for _, tt := range tests {
		bf := append(Bitfield{}, tt.input...)
		bf.SetPiece(tt.index)
		if !reflect.DeepEqual(bf, tt.want) {
			t.Errorf("bitfield: %v, index: %d, wanted: %v, got: %v", tt.input, tt.index, tt.want, bf)
		}
	}
}
