// Package vec reads and writes the flat verse vector file.
//
// Layout (little endian): magic "BBV1", uint32 dims, uint32 count, then count
// uint32 verse keys, then count*dims float32 values (unit length).
package vec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

const magic = "BBV1"

type Store struct {
	Dims int
	Keys []uint32
	Data []float32
}

// Normalize scales v to unit length in place.
func Normalize(v []float32) {
	var s float64
	for _, x := range v {
		s += float64(x) * float64(x)
	}
	if s == 0 {
		return
	}
	n := float32(1 / math.Sqrt(s))
	for i := range v {
		v[i] *= n
	}
}

func (s *Store) Write(w io.Writer) error {
	if len(s.Data) != len(s.Keys)*s.Dims {
		return fmt.Errorf("vec: %d values for %d keys at %d dims", len(s.Data), len(s.Keys), s.Dims)
	}
	if _, err := io.WriteString(w, magic); err != nil {
		return err
	}
	hdr := []uint32{uint32(s.Dims), uint32(len(s.Keys))}
	for _, x := range [][]uint32{hdr, s.Keys} {
		if err := binary.Write(w, binary.LittleEndian, x); err != nil {
			return err
		}
	}
	return binary.Write(w, binary.LittleEndian, s.Data)
}

func Read(r io.Reader) (*Store, error) {
	m := make([]byte, len(magic))
	if _, err := io.ReadFull(r, m); err != nil {
		return nil, err
	}
	if string(m) != magic {
		return nil, errors.New("vec: bad magic")
	}
	hdr := make([]uint32, 2)
	if err := binary.Read(r, binary.LittleEndian, hdr); err != nil {
		return nil, err
	}
	s := &Store{Dims: int(hdr[0]), Keys: make([]uint32, hdr[1])}
	if err := binary.Read(r, binary.LittleEndian, s.Keys); err != nil {
		return nil, err
	}
	s.Data = make([]float32, len(s.Keys)*s.Dims)
	if err := binary.Read(r, binary.LittleEndian, s.Data); err != nil {
		return nil, err
	}
	return s, nil
}
