package data

import (
	"strings"
	"testing"

	"github.com/KalleNgithub/bible-bot/internal/bible"
)

func TestLoad(t *testing.T) {
	en, fi, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if n := len(en.Refs()); n < 31000 || n > 31200 {
		t.Fatalf("en verse count %d", n)
	}
	if n := len(fi.Refs()); n < 31000 || n > 31200 {
		t.Fatalf("fi verse count %d", n)
	}
	for _, r := range []bible.Ref{{43, 11, 25}, {43, 11, 26}, {19, 34, 18}} {
		if _, ok := en.Text(r); !ok {
			t.Errorf("en missing %v", r)
		}
		if _, ok := fi.Text(r); !ok {
			t.Errorf("fi missing %v", r)
		}
	}
	if s, _ := fi.Text(bible.Ref{43, 11, 26}); !strings.Contains(s, "Uskotko sen") {
		t.Errorf("fi John 11:26 = %q", s)
	}
	if en.Books[43] != "John" || fi.Books[43] != "Johanneksen evankeliumi" {
		t.Errorf("book names: %q %q", en.Books[43], fi.Books[43])
	}
}
