package search

import (
	"math/rand/v2"
	"strings"
	"unicode"

	"github.com/KalleNgithub/bible-bot/internal/bible"
)

var stopwords = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`the and for are but not you your yours with what how why who when where
		does did can could would should will shall have has had was were been being this that these those
		from into about they them their our out all any its his her him she his one get got let may`) {
		stopwords[w] = true
	}
}

func words(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) })
}

// WordMatch returns a random verse among those containing the most distinct words of the
// query (ignoring stopwords and words shorter than 3 letters), or any random verse if none match.
func WordMatch(en *bible.Translation, query string, rng *rand.Rand) bible.Ref {
	want := map[string]bool{}
	for _, w := range words(query) {
		if len(w) >= 3 && !stopwords[w] {
			want[w] = true
		}
	}
	refs := en.Refs()
	var best []bible.Ref
	bestN := 0
	for _, r := range refs {
		text, _ := en.Text(r)
		seen := map[string]bool{}
		for _, w := range words(text) {
			if want[w] {
				seen[w] = true
			}
		}
		switch n := len(seen); {
		case n == 0 || n < bestN:
		case n > bestN:
			bestN, best = n, append(best[:0], r)
		default:
			best = append(best, r)
		}
	}
	if bestN == 0 {
		return refs[rng.IntN(len(refs))]
	}
	return best[rng.IntN(len(best))]
}
