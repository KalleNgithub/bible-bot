// Package topic finds a verse for a free-text query using tiered fallbacks, so every
// non-empty query returns something.
package topic

import (
	"context"
	"math/rand/v2"
	"sync"

	"github.com/KalleNgithub/bible-bot/internal/azure"
	"github.com/KalleNgithub/bible-bot/internal/bible"
	"github.com/KalleNgithub/bible-bot/internal/search"
	"github.com/KalleNgithub/bible-bot/internal/vec"
)

type Backend interface {
	Translate(ctx context.Context, query string) (azure.Translation, error)
	Embed(ctx context.Context, texts []string, dims int) ([][]float32, error)
}

// Tier says how a result was found.
type Tier int

const (
	Primary  Tier = 1 // top matches above the threshold
	Lenient  Tier = 2 // nearest matches regardless of score, more random
	WordOnly Tier = 3 // search service unavailable: word match or random verse
)

type Result struct {
	Ref  bible.Ref
	Tier Tier
	// Lang is the detected query language, empty if translation failed.
	Lang string
}

type Finder struct {
	Backend   Backend
	Store     *vec.Store
	EN        *bible.Translation
	Threshold float32
	TopK      int
	Temp      float64
	// LenientPool and LenientTemp control tier 2.
	LenientPool int
	LenientTemp float64

	mu  sync.Mutex
	rng *rand.Rand
}

// New returns a Finder with the tuned defaults.
func New(b Backend, s *vec.Store, en *bible.Translation, seed *rand.Rand) *Finder {
	return &Finder{
		Backend: b, Store: s, EN: en, rng: seed,
		Threshold: 0.50, TopK: 10, Temp: 0.03,
		LenientPool: 50, LenientTemp: 0.1,
	}
}

// Find returns a verse for query. It only fails if the verse data is empty.
func (f *Finder) Find(ctx context.Context, query string) Result {
	res := Result{}
	text, english := query, query
	if t, err := f.Backend.Translate(ctx, query); err == nil {
		res.Lang, english = t.Lang, t.English
		text = t.English
		if t.Answer != "" {
			text += ". " + t.Answer
		}
	}

	vs, err := f.Backend.Embed(ctx, []string{text}, f.Store.Dims)
	if err != nil || len(vs) != 1 {
		f.mu.Lock()
		defer f.mu.Unlock()
		res.Ref, res.Tier = search.WordMatch(f.EN, english, f.rng), WordOnly
		return res
	}
	vec.Normalize(vs[0])

	pool := max(f.TopK, f.LenientPool)
	hits := search.Top(f.Store, vs[0], pool, -1)

	f.mu.Lock()
	defer f.mu.Unlock()
	var primary []search.Hit
	for _, h := range hits {
		if h.Score >= f.Threshold && len(primary) < f.TopK {
			primary = append(primary, h)
		}
	}
	if len(primary) > 0 {
		res.Ref, res.Tier = bible.RefFromKey(search.Pick(primary, f.Temp, f.rng).Key), Primary
		return res
	}
	hits = hits[:min(len(hits), f.LenientPool)]
	if len(hits) == 0 {
		res.Ref, res.Tier = search.WordMatch(f.EN, english, f.rng), WordOnly
		return res
	}
	res.Ref, res.Tier = bible.RefFromKey(search.Pick(hits, f.LenientTemp, f.rng).Key), Lenient
	return res
}
