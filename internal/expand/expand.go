// Package expand widens a matched verse range so it does not start or end mid-sentence.
package expand

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/KalleNgithub/bible-bot/internal/bible"
)

// MaxVerses is the hard cap on the size of an expanded range.
const MaxVerses = 4

var connectives = map[string]bool{
	"that": true, "who": true, "which": true, "because": true, "or": true, "nor": true,
}

// Range expands [from, to] within one chapter. It never shrinks the range,
// never leaves the chapter, and stops at MaxVerses.
func Range(t *bible.Translation, book, chapter, from, to int) (int, int) {
	last := t.LastVerse(book, chapter)
	text := func(v int) string {
		s, _ := t.Text(bible.Ref{Book: book, Chapter: chapter, Verse: v})
		return s
	}
	for to-from+1 < MaxVerses {
		changed := false
		if from > 1 && continuesPrevious(text(from)) {
			from--
			changed = true
		}
		if to-from+1 < MaxVerses && to < last && unfinished(rangeText(text, from, to)) {
			to++
			changed = true
		}
		if !changed {
			break
		}
	}
	return from, to
}

func rangeText(text func(int) string, from, to int) string {
	var parts []string
	for v := from; v <= to; v++ {
		parts = append(parts, text(v))
	}
	return strings.Join(parts, " ")
}

// continuesPrevious is true when the verse starts lowercase or with a connective.
func continuesPrevious(s string) bool {
	s = strings.TrimLeft(s, "“‘\"'([ ")
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return false
	}
	if unicode.IsLower(r) {
		return true
	}
	word := s
	if i := strings.IndexFunc(s, func(r rune) bool { return !unicode.IsLetter(r) }); i >= 0 {
		word = s[:i]
	}
	return connectives[strings.ToLower(word)]
}

// unfinished is true when the text lacks terminal punctuation.
func unfinished(s string) bool {
	s = strings.TrimRight(s, "”’\"')] ")
	return !(strings.HasSuffix(s, ".") || strings.HasSuffix(s, "!") || strings.HasSuffix(s, "?"))
}
