// Command eval scores topic search against eval/queries.tsv and reports hit rates,
// threshold behaviour and the effect of the softmax temperature. It makes live Azure calls.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/KalleNgithub/bible-bot/data"
	"github.com/KalleNgithub/bible-bot/internal/azure"
	"github.com/KalleNgithub/bible-bot/internal/bible"
	"github.com/KalleNgithub/bible-bot/internal/ref"
	"github.com/KalleNgithub/bible-bot/internal/search"
	"github.com/KalleNgithub/bible-bot/internal/vec"
)

type query struct {
	lang, text string
	ok         map[uint32]bool
	topK       []search.Hit
	english    string
	answer     string
}

func key(r bible.Ref) uint32 { return uint32(r.Book*1000000 + r.Chapter*1000 + r.Verse) }

func main() {
	file := flag.String("f", "eval/queries.tsv", "query file")
	vectors := flag.String("v", "data/vectors.bin", "vector file")
	thr := flag.Float64("thr", 0.50, "score threshold used for the temperature analysis")
	show := flag.Bool("show", false, "print the top 3 verses for queries whose top hit is not acceptable")
	list := flag.String("list", "", "write a reviewable query and answer list (combo mode) to this file")
	nruns := flag.Int("runs", 0, "if > 0, only run the repeated-run score report")
	flag.Parse()

	c, err := azure.FromEnv()
	check(err)
	en, fi, err := data.Load()
	check(err)
	p := ref.NewParser(en, fi)
	vf, err := os.Open(*vectors)
	check(err)
	store, err := vec.Read(vf)
	check(err)

	qs := load(*file, p, en)
	if *nruns > 0 {
		check(runs(c, store, qs, *nruns, *thr))
		return
	}
	modes := []string{"raw", "answer", "combo"}
	tops := make([]map[string][]search.Hit, len(qs))
	for i, q := range qs {
		t, err := c.Translate(context.Background(), q.text)
		check(err)
		inputs := map[string]string{"raw": t.English, "answer": t.Answer, "combo": t.English + ". " + t.Answer}
		tops[i] = map[string][]search.Hit{}
		for _, m := range modes {
			v, err := c.Embed(context.Background(), []string{inputs[m]}, store.Dims)
			check(err)
			vec.Normalize(v[0])
			tops[i][m] = search.Top(store, v[0], 10, -1)
		}
		if *show {
			fmt.Printf("[%s] %s => %s\n", q.text, t.English, t.Answer)
		}
		q.english, q.answer = t.English, t.Answer
	}

	if *list != "" {
		check(writeList(*list, qs, tops, en, *thr))
	}

	for _, m := range modes {
		for i, q := range qs {
			q.topK = tops[i][m]
		}
		fmt.Printf("\n######## mode: %s\n", m)
		report(qs, en, *thr, *show)
	}
}

func load(path string, p *ref.Parser, en *bible.Translation) []*query {
	f, err := os.Open(path)
	check(err)
	defer f.Close()
	var qs []*query
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), "\t")
		if len(parts) != 3 {
			continue
		}
		q := &query{lang: parts[0], text: parts[1], ok: map[uint32]bool{}}
		if parts[2] != "-" {
			for _, s := range strings.Split(parts[2], ";") {
				r, err := p.Parse(strings.TrimSpace(s))
				if err != nil {
					check(fmt.Errorf("query %q ref %q: %w", q.text, s, err))
				}
				for v := r.From; v <= r.To; v++ {
					q.ok[key(bible.Ref{Book: r.Book, Chapter: r.Chapter, Verse: v})] = true
				}
			}
		}
		qs = append(qs, q)
	}
	return qs
}

func report(qs []*query, en *bible.Translation, thr float64, show bool) {
	fmt.Println("== per query (top1 score, rank of first acceptable verse, 0 = none in top 10)")
	var on, off []*query
	for _, q := range qs {
		rank := 0
		for i, h := range q.topK {
			if q.ok[h.Key] {
				rank = i + 1
				break
			}
		}
		kind := "on "
		if len(q.ok) == 0 {
			kind = "off"
			off = append(off, q)
		} else {
			on = append(on, q)
		}
		fmt.Printf("%s %.3f rank=%d  %s\n", kind, q.topK[0].Score, rank, q.text)
		if show && rank != 1 {
			for _, h := range q.topK[:3] {
				r := bible.Ref{Book: int(h.Key / 1000000), Chapter: int(h.Key / 1000 % 1000), Verse: int(h.Key % 1000)}
				t, _ := en.Text(r)
				if len(t) > 90 {
					t = t[:90]
				}
				mark := " "
				if q.ok[h.Key] {
					mark = "*"
				}
				fmt.Printf("     %s%.3f %s %d:%d %s\n", mark, h.Score, en.Books[r.Book], r.Chapter, r.Verse, t)
			}
		}
	}

	hit1, hit10 := 0, 0
	for _, q := range on {
		if q.ok[q.topK[0].Key] {
			hit1++
		}
		for _, h := range q.topK {
			if q.ok[h.Key] {
				hit10++
				break
			}
		}
	}
	fmt.Printf("\n== on-topic hit@1 %d/%d, hit@10 %d/%d\n", hit1, len(on), hit10, len(on))

	fmt.Println("\n== threshold: on-topic answered / off-topic answered (should be 0)")
	for t := 0.25; t <= 0.651; t += 0.05 {
		a, b := 0, 0
		for _, q := range on {
			if float64(q.topK[0].Score) >= t {
				a++
			}
		}
		for _, q := range off {
			if float64(q.topK[0].Score) >= t {
				b++
			}
		}
		fmt.Printf("thr %.2f  on %d/%d  off %d/%d\n", t, a, len(on), b, len(off))
	}

	fmt.Printf("\n== temperature at threshold %.2f: expected chance a pick is acceptable\n", thr)
	fmt.Println("T      all on-topic   single-verse quotes")
	for _, T := range []float64{0.005, 0.01, 0.02, 0.03, 0.05, 0.08} {
		var allSum, allN, oneSum, oneN float64
		for _, q := range on {
			p, ok := expectedOK(q, T, thr)
			if !ok {
				continue
			}
			allSum += p
			allN++
			if len(q.ok) == 1 {
				oneSum += p
				oneN++
			}
		}
		fmt.Printf("%.3f  %.2f (n=%d)     %.2f (n=%d)\n", T, allSum/math.Max(allN, 1), int(allN), oneSum/math.Max(oneN, 1), int(oneN))
	}
}

// expectedOK is the softmax probability mass on acceptable verses among hits >= thr.
func expectedOK(q *query, T, thr float64) (float64, bool) {
	var hits []search.Hit
	for _, h := range q.topK {
		if float64(h.Score) >= thr {
			hits = append(hits, h)
		}
	}
	if len(hits) == 0 {
		return 0, false
	}
	best := float64(hits[0].Score)
	var sum, good float64
	for _, h := range hits {
		w := math.Exp((float64(h.Score) - best) / T)
		sum += w
		if q.ok[h.Key] {
			good += w
		}
	}
	return good / sum, true
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
