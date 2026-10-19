package expand

import (
	"fmt"
	"testing"

	"github.com/KalleNgithub/bible-bot/data"
	"github.com/KalleNgithub/bible-bot/internal/bible"
)

func chapter(verses ...string) *bible.Translation {
	t := bible.New()
	for i, v := range verses {
		t.Add(bible.Ref{Book: 1, Chapter: 1, Verse: i + 1}, v)
	}
	return t
}

func TestRange(t *testing.T) {
	cases := []struct {
		name         string
		verses       []string
		from, to     int
		wantF, wantT int
	}{
		{"self-contained", []string{"One.", "Two.", "Three."}, 2, 2, 2, 2},
		{"unfinished end", []string{"One.", "Two,", "three."}, 2, 2, 2, 3},
		{"no punctuation", []string{"One.", "Two", "three."}, 2, 2, 2, 3},
		{"lowercase start", []string{"One.", "two.", "Three."}, 2, 2, 1, 2},
		{"connective start", []string{"One.", "That two.", "Three."}, 2, 2, 1, 2},
		{"and start stands alone", []string{"One.", "And two.", "Three."}, 2, 2, 2, 2},
		{"for start stands alone", []string{"One.", "For two.", "Three."}, 2, 2, 2, 2},
		{"unclosed quote does not extend", []string{"One.", "He said, “Two.", "Three,” and left."}, 2, 2, 2, 2},
		{"closing quote only stays", []string{"He said, “One.", "Two.”", "Three."}, 2, 2, 2, 2},
		{"first verse continuation", []string{"and one.", "Two."}, 1, 1, 1, 1},
		{"last verse unfinished", []string{"One.", "Two,"}, 2, 2, 2, 2},
		{"cap", []string{"a,", "b,", "c,", "d,", "e,", "f."}, 1, 1, 1, MaxVerses},
		{"cap keeps hit", []string{"a.", "that b,", "that c,", "that d,", "that e,", "that f."}, 3, 3, 1, 4},
	}
	for _, c := range cases {
		f, to := Range(chapter(c.verses...), 1, 1, c.from, c.to)
		if f != c.wantF || to != c.wantT {
			t.Errorf("%s: got %d-%d want %d-%d", c.name, f, to, c.wantF, c.wantT)
		}
	}
}

func TestJohn11(t *testing.T) {
	en, _, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ hit, wantF, wantT int }{
		{25, 25, 25}, // ends with a full stop; the unclosed quote no longer extends it
		{26, 26, 26}, // starts a sentence and ends one: stands alone
	} {
		f, to := Range(en, 43, 11, c.hit, c.hit)
		if f != c.wantF || to != c.wantT {
			t.Errorf("John 11:%d expanded to %d-%d, want %d-%d", c.hit, f, to, c.wantF, c.wantT)
		}
	}
	// Informational: how the rules treat a few well known verses.
	for _, r := range []bible.Ref{{Book: 19, Chapter: 34, Verse: 18}, {Book: 43, Chapter: 3, Verse: 16}, {Book: 24, Chapter: 29, Verse: 11}} {
		f, to := Range(en, r.Book, r.Chapter, r.Verse, r.Verse)
		t.Log(fmt.Sprintf("%s %d:%d -> %d-%d", en.Books[r.Book], r.Chapter, r.Verse, f, to))
	}
}
