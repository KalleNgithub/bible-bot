package bible

import (
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	tr := New()
	tr.SetBook(43, "John")
	tr.Add(Ref{43, 11, 25}, "Jesus said")
	tr.Add(Ref{43, 11, 26}, "and whoever lives")
	var b bytes.Buffer
	if err := tr.Write(&b); err != nil {
		t.Fatal(err)
	}
	got, err := Load(&b)
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := got.Text(Ref{43, 11, 26}); !ok || s != "and whoever lives" {
		t.Fatalf("got %q %v", s, ok)
	}
	if got.Books[43] != "John" || got.LastVerse(43, 11) != 26 || got.LastVerse(43, 12) != 0 {
		t.Fatal("books or chapter bounds wrong")
	}
	if _, ok := got.Text(Ref{43, 11, 27}); ok {
		t.Fatal("verse 27 should not exist")
	}
}
