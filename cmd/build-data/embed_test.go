package main

import (
	"context"
	"testing"
)

type fakeEmbedder struct{ batches []int }

func (f *fakeEmbedder) Embed(_ context.Context, texts []string, dims int) ([][]float32, error) {
	f.batches = append(f.batches, len(texts))
	out := make([][]float32, len(texts))
	for i := range texts {
		v := make([]float32, dims)
		v[0] = 3
		v[1] = 4
		out[i] = v
	}
	return out, nil
}

func TestEmbedAllBatchesAndNormalizes(t *testing.T) {
	f := &fakeEmbedder{}
	texts := make([]string, embedBatch+5)
	got, err := embedAll(context.Background(), f, texts, 4, func(int) {})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(texts)*4 {
		t.Fatalf("len = %d", len(got))
	}
	if len(f.batches) != 2 || f.batches[0] != embedBatch || f.batches[1] != 5 {
		t.Fatalf("batches = %v", f.batches)
	}
	if got[0] < 0.59 || got[0] > 0.61 || got[1] < 0.79 || got[1] > 0.81 {
		t.Fatalf("not normalized: %v", got[:4])
	}
}
