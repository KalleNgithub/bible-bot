package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/KalleNgithub/bible-bot/internal/bible"
	"github.com/KalleNgithub/bible-bot/internal/expand"
	"github.com/KalleNgithub/bible-bot/internal/quote"
	"github.com/KalleNgithub/bible-bot/internal/search"
)

const listTemp = 0.03

// writeList writes each query with its top 5 combo-mode results, the chance the bot
// picks each one, and the final quote after context expansion.
func writeList(path string, qs []*query, tops []map[string][]search.Hit, en *bible.Translation, thr float64) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Threshold %.2f, T %.2f. P is the chance the bot returns that verse.\n", thr, listTemp)
	for i, q := range qs {
		hits := tops[i]["combo"]
		fmt.Fprintf(&b, "\n## %s\n", q.text)
		fmt.Fprintf(&b, "english: %s\ndescription: %s\n", q.english, q.answer)

		var sum float64
		best := float64(hits[0].Score)
		for _, h := range hits {
			if float64(h.Score) >= thr {
				sum += math.Exp((float64(h.Score) - best) / listTemp)
			}
		}
		if sum == 0 {
			b.WriteString("=> no suitable verse found\n")
			fmt.Fprintf(&b, "(best was %.3f)\n", hits[0].Score)
			continue
		}
		for _, h := range hits[:5] {
			p := 0.0
			if float64(h.Score) >= thr {
				p = math.Exp((float64(h.Score)-best)/listTemp) / sum
			}
			book, ch, v := int(h.Key/1000000), int(h.Key/1000%1000), int(h.Key%1000)
			from, to := expand.Range(en, book, ch, v, v)
			text, _ := quote.Format(en, book, ch, from, to)
			fmt.Fprintf(&b, "- %.3f P=%2.0f%%  %s\n", h.Score, p*100, text)
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
