package search

import (
	"math/rand/v2"
	"testing"

	"github.com/KalleNgithub/bible-bot/internal/vec"
)

// store with 2D unit vectors at known cosines to the query (1,0).
func store(cos ...float32) *vec.Store {
	s := &vec.Store{Dims: 2}
	for i, c := range cos {
		s.Keys = append(s.Keys, uint32(i+1))
		sin := float32(0)
		if c < 1 {
			sin = float32(sqrt(1 - float64(c*c)))
		}
		s.Data = append(s.Data, c, sin)
	}
	return s
}

func sqrt(x float64) float64 {
	z := x
	for range 30 {
		z = (z + x/z) / 2
	}
	return z
}

var q = []float32{1, 0}

func TestTopOrdersAndCaps(t *testing.T) {
	hits := Top(store(0.5, 0.9, 0.7, 0.6), q, 2, 0.4)
	if len(hits) != 2 || hits[0].Key != 2 || hits[1].Key != 3 {
		t.Fatalf("got %+v", hits)
	}
}

func TestTopFewMatches(t *testing.T) {
	hits := Top(store(0.9, 0.2, 0.1), q, 10, 0.5)
	if len(hits) != 1 || hits[0].Key != 1 {
		t.Fatalf("got %+v", hits)
	}
}

func TestTopNoMatches(t *testing.T) {
	if hits := Top(store(0.2, 0.1), q, 10, 0.5); len(hits) != 0 {
		t.Fatalf("got %+v", hits)
	}
}

func counts(hits []Hit, temp float64, n int) map[uint32]int {
	rng := rand.New(rand.NewPCG(1, 2))
	c := map[uint32]int{}
	for range n {
		c[Pick(hits, temp, rng).Key]++
	}
	return c
}

func TestPickDominantScoreNearlyAlwaysWins(t *testing.T) {
	hits := []Hit{{1, 0.73}, {2, 0.60}, {3, 0.59}}
	c := counts(hits, 0.03, 10000)
	if c[1] < 9700 {
		t.Fatalf("dominant verse won only %d/10000: %v", c[1], c)
	}
}

func TestPickSimilarScoresVaryWithHigherMoreOften(t *testing.T) {
	hits := []Hit{{1, 0.50}, {2, 0.48}, {3, 0.46}}
	c := counts(hits, 0.03, 10000)
	if c[1] == 0 || c[2] == 0 || c[3] == 0 {
		t.Fatalf("expected variety: %v", c)
	}
	if !(c[1] > c[2] && c[2] > c[3]) {
		t.Fatalf("higher scores should win more often: %v", c)
	}
}

func TestPickSingleHit(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	if got := Pick([]Hit{{7, 0.5}}, 0.03, rng); got.Key != 7 {
		t.Fatalf("got %+v", got)
	}
}
