// Package search ranks verse vectors against a query and picks a result.
package search

import (
	"math"
	"math/rand/v2"
	"sort"

	"github.com/KalleNgithub/bible-bot/internal/vec"
)

type Hit struct {
	Key   uint32
	Score float32
}

// Top returns up to k hits with score >= threshold, best first. Vectors must be unit length.
func Top(s *vec.Store, q []float32, k int, threshold float32) []Hit {
	var hits []Hit
	for i, key := range s.Keys {
		row := s.Data[i*s.Dims : (i+1)*s.Dims]
		var d float32
		for j, x := range q {
			d += x * row[j]
		}
		if d >= threshold {
			hits = append(hits, Hit{key, d})
		}
	}
	sort.Slice(hits, func(a, b int) bool { return hits[a].Score > hits[b].Score })
	if len(hits) > k {
		hits = hits[:k]
	}
	return hits
}

// Pick chooses one hit with probability proportional to exp(score/temp). hits must not be empty.
func Pick(hits []Hit, temp float64, rng *rand.Rand) Hit {
	best := float64(hits[0].Score)
	for _, h := range hits {
		best = math.Max(best, float64(h.Score))
	}
	w := make([]float64, len(hits))
	var sum float64
	for i, h := range hits {
		w[i] = math.Exp((float64(h.Score) - best) / temp)
		sum += w[i]
	}
	r := rng.Float64() * sum
	for i, x := range w {
		r -= x
		if r < 0 {
			return hits[i]
		}
	}
	return hits[len(hits)-1]
}
