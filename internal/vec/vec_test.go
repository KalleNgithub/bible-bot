package vec

import (
	"bytes"
	"math"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	s := &Store{Dims: 3, Keys: []uint32{1001001, 1001002}, Data: []float32{1, 2, 3, 4, 5, 6}}
	var b bytes.Buffer
	if err := s.Write(&b); err != nil {
		t.Fatal(err)
	}
	got, err := Read(&b)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dims != 3 || len(got.Keys) != 2 || got.Keys[1] != 1001002 || got.Data[5] != 6 {
		t.Fatalf("mismatch: %+v", got)
	}
}

func TestReadRejectsBadMagic(t *testing.T) {
	if _, err := Read(bytes.NewReader([]byte("XXXX0000000000000"))); err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteRejectsShapeMismatch(t *testing.T) {
	s := &Store{Dims: 2, Keys: []uint32{1}, Data: []float32{1}}
	if err := s.Write(&bytes.Buffer{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalize(t *testing.T) {
	v := []float32{3, 4}
	Normalize(v)
	if math.Abs(float64(v[0])-0.6) > 1e-6 || math.Abs(float64(v[1])-0.8) > 1e-6 {
		t.Fatalf("got %v", v)
	}
}
