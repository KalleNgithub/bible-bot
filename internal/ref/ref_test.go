package ref

import (
	"errors"
	"strings"
	"testing"

	"github.com/KalleNgithub/bible-bot/data"
)

func newParser(t *testing.T) *Parser {
	t.Helper()
	en, fi, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	return NewParser(en, fi)
}

func TestAllBooksResolveByCanonicalName(t *testing.T) {
	en, fi, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	p := NewParser(en, fi)
	for nr := 1; nr <= 66; nr++ {
		for _, name := range []string{en.Books[nr], fi.Books[nr]} {
			r, err := p.Parse(name + " 1:1")
			if err != nil || r.Book != nr || r.Chapter != 1 || r.From != 1 {
				t.Errorf("%q: got %+v, %v; want book %d", name, r, err, nr)
			}
		}
	}
}

func TestParse(t *testing.T) {
	p := newParser(t)
	cases := []struct {
		in           string
		book, ch, v1 int
		v2           int
		fi           bool
	}{
		{"John 11:25-26", 43, 11, 25, 26, false},
		{"Jhon 11:25 to 26", 43, 11, 25, 26, false},
		{"Jhn chapter11 verses 25 and 26", 43, 11, 25, 26, false},
		{"john 11 25", 43, 11, 25, 25, false},
		{"Joh 11:25", 43, 11, 25, 25, false},
		{"Johanneksen evankeliumi 11:25-26", 43, 11, 25, 26, true},
		{"Joh. 11:25–26", 43, 11, 25, 26, false},
		{"psalm 34:18", 19, 34, 18, 18, false},
		{"Psalmit 34:18", 19, 34, 18, 18, true},
		{"pslam 34:18", 19, 34, 18, 18, false},
		{"1 John 3:16", 62, 3, 16, 16, false},
		{"First John 3:16", 62, 3, 16, 16, false},
		{"I John 3:16", 62, 3, 16, 16, false},
		{"2 Joh 1:3", 63, 1, 3, 3, false},
		{"1. Moos. 1:1", 1, 1, 1, 1, true},
		{"1 Moos 1:1", 1, 1, 1, 1, true},
		{"Jude 5", 65, 1, 5, 5, false},
		{"Jude 5-7", 65, 1, 5, 7, false},
		{"Jude 1:5", 65, 1, 5, 5, false},
		{"Matt 5:3", 40, 5, 3, 3, false},
		{"Matteus 5:3", 40, 5, 3, 3, true},
		{"Romans 8:28", 45, 8, 28, 28, false},
		{"Room 8:28", 45, 8, 28, 28, true},
		{"Johanneksen evankeliumi luku 11 jae 25", 43, 11, 25, 25, true},
	}
	for _, c := range cases {
		r, err := p.Parse(c.in)
		if err != nil {
			t.Errorf("%q: %v", c.in, err)
			continue
		}
		if r.Book != c.book || r.Chapter != c.ch || r.From != c.v1 || r.To != c.v2 || r.Finnish != c.fi {
			t.Errorf("%q: got %+v want book=%d ch=%d %d-%d fi=%v", c.in, r, c.book, c.ch, c.v1, c.v2, c.fi)
		}
	}
}

func TestNotFound(t *testing.T) {
	p := newParser(t)
	for _, in := range []string{"John 99:1", "John 11:99", "John 3", "John 11:26-25", "Obadiah 2:1"} {
		_, err := p.Parse(in)
		var nf *NotFoundError
		if !errors.As(err, &nf) {
			t.Errorf("%q: got %v, want NotFoundError", in, err)
		}
	}
}

func TestNotReference(t *testing.T) {
	p := newParser(t)
	for _, in := range []string{
		"my mom died and I require consolation",
		"what happens to people who believe in jesus",
		"I need help 5 times a day",
		"john",
		"2 days ago my dog died",
		"äitini kuoli ja kaipaan lohtua",
		"John 11:25 and 30",
		"Xyzzy 3:16",
		"",
	} {
		if _, err := p.Parse(in); !errors.Is(err, ErrNotReference) {
			t.Errorf("%q: got %v, want ErrNotReference", in, err)
		}
	}
}

func TestFuzzyDoesNotCrossBooks(t *testing.T) {
	p := newParser(t)
	// "jud" is ambiguous between Judges and Jude.
	if _, err := p.Parse("Jud 1:1"); err == nil || !strings.Contains(err.Error(), "reference") {
		t.Errorf("want ErrNotReference for ambiguous abbreviation, got %v", err)
	}
}
