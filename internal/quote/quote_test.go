package quote

import (
	"strings"
	"testing"

	"github.com/KalleNgithub/bible-bot/data"
)

func TestFormat(t *testing.T) {
	en, fi, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}

	s, ok := Format(en, 43, 11, 25, 26)
	if !ok || !strings.HasPrefix(s, "John 11:25-26: Jesus said to her") || !strings.HasSuffix(s, "Do you believe this?”") {
		t.Errorf("en range: %q", s)
	}

	s, ok = Format(fi, 43, 11, 25, 26)
	if !ok || !strings.HasPrefix(s, "Johanneksen evankeliumi 11:25-26: Jeesus sanoi") || !strings.Contains(s, "Uskotko sen?") {
		t.Errorf("fi range: %q", s)
	}

	s, ok = Format(en, 19, 34, 18, 18)
	if !ok || !strings.HasPrefix(s, "Psalms 34:18: ") {
		t.Errorf("single verse: %q", s)
	}

	if _, ok := Format(fi, 43, 11, 500, 501); ok {
		t.Error("nonexistent verses should not be ok")
	}
}
