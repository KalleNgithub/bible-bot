package topic

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/KalleNgithub/bible-bot/internal/azure"
	"github.com/KalleNgithub/bible-bot/internal/bible"
	"github.com/KalleNgithub/bible-bot/internal/vec"
)

type fake struct {
	tr      azure.Translation
	trErr   error
	emb     []float32
	embErr  error
	gotText string
}

func (f *fake) Translate(context.Context, string) (azure.Translation, error) { return f.tr, f.trErr }

func (f *fake) Embed(_ context.Context, texts []string, _ int) ([][]float32, error) {
	f.gotText = texts[0]
	if f.embErr != nil {
		return nil, f.embErr
	}
	return [][]float32{append([]float32(nil), f.emb...)}, nil
}

// Verse 1 has cosine 0.9 with (1,0); verse 2 has 0.1; verse 3 has 0.05.
func fixture() (*vec.Store, *bible.Translation) {
	s := &vec.Store{Dims: 2, Keys: []uint32{1001001, 1001002, 1001003}, Data: []float32{0.9, 0.436, 0.1, 0.995, 0.05, 0.9987}}
	en := bible.New()
	en.SetBook(1, "Genesis")
	en.Add(bible.Ref{Book: 1, Chapter: 1, Verse: 1}, "God made light.")
	en.Add(bible.Ref{Book: 1, Chapter: 1, Verse: 2}, "Man baked bread.")
	en.Add(bible.Ref{Book: 1, Chapter: 1, Verse: 3}, "The sea was wide.")
	return s, en
}

func finder(f *fake) *Finder {
	s, en := fixture()
	return New(f, s, en, rand.New(rand.NewPCG(1, 2)))
}

func TestPrimaryTier(t *testing.T) {
	f := &fake{tr: azure.Translation{English: "light", Lang: "fi", Answer: "God lights"}, emb: []float32{1, 0}}
	r := finder(f).Find(context.Background(), "valo")
	if r.Tier != Primary || r.Ref.Verse != 1 || r.Lang != "fi" {
		t.Fatalf("got %+v", r)
	}
	if f.gotText != "light. God lights" {
		t.Fatalf("embedded %q", f.gotText)
	}
}

func TestLenientTierWhenNothingQualifies(t *testing.T) {
	f := &fake{tr: azure.Translation{English: "x", Lang: "en"}, emb: []float32{0, 1}}
	fd := finder(f)
	fd.Threshold = 0.9999
	seen := map[int]bool{}
	for range 200 {
		r := fd.Find(context.Background(), "x")
		if r.Tier != Lenient {
			t.Fatalf("tier = %d", r.Tier)
		}
		seen[r.Ref.Verse] = true
	}
	if len(seen) < 2 {
		t.Fatalf("lenient tier should vary, got %v", seen)
	}
}

func TestChatFailureStillSearches(t *testing.T) {
	f := &fake{trErr: errors.New("down"), emb: []float32{1, 0}}
	r := finder(f).Find(context.Background(), "raw words")
	if r.Tier != Primary || r.Lang != "" || f.gotText != "raw words" {
		t.Fatalf("got %+v embedded %q", r, f.gotText)
	}
}

func TestEmbedFailureUsesWordMatch(t *testing.T) {
	f := &fake{tr: azure.Translation{English: "how do I bake bread"}, embErr: errors.New("down")}
	r := finder(f).Find(context.Background(), "x")
	if r.Tier != WordOnly || r.Ref.Verse != 2 {
		t.Fatalf("got %+v", r)
	}
}

func TestNoWordMatchReturnsSomeVerse(t *testing.T) {
	f := &fake{trErr: errors.New("down"), embErr: errors.New("down")}
	r := finder(f).Find(context.Background(), "zzzz qqqq")
	if r.Tier != WordOnly || r.Ref.Book != 1 || r.Ref.Verse < 1 || r.Ref.Verse > 3 {
		t.Fatalf("got %+v", r)
	}
}
