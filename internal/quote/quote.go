// Package quote renders verse ranges from a translation as plain text.
package quote

import (
	"fmt"
	"strings"

	"github.com/KalleNgithub/bible-bot/internal/bible"
)

// Format renders "Book C:V[-V]: text" for a verse range in one chapter.
// Verses missing from the translation (versification differences) are skipped;
// ok is false when none of them exist.
func Format(t *bible.Translation, book, chapter, from, to int) (s string, ok bool) {
	var texts []string
	for v := from; v <= to; v++ {
		if txt, found := t.Text(bible.Ref{Book: book, Chapter: chapter, Verse: v}); found {
			texts = append(texts, txt)
		}
	}
	if len(texts) == 0 {
		return "", false
	}
	label := fmt.Sprintf("%s %d:%d", t.Books[book], chapter, from)
	if to > from {
		label += fmt.Sprintf("-%d", to)
	}
	return label + ": " + strings.Join(texts, " "), true
}
