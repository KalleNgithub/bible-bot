package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/KalleNgithub/bible-bot/internal/azure"
	"github.com/KalleNgithub/bible-bot/internal/search"
	"github.com/KalleNgithub/bible-bot/internal/vec"
)

// runs repeats the combo query n times per query (the description varies per run)
// and reports top scores against the tier 1 threshold.
func runs(c *azure.Client, store *vec.Store, qs []*query, n int, thr float64) error {
	var onMin, offMax = 2.0, 0.0
	var onBelow, onTotal, offAbove, offTotal, onHit1 int
	fmt.Printf("\n######## %d runs, combo mode, thr %.2f\n", n, thr)
	for _, q := range qs {
		var scores []float64
		hit1 := 0
		for i := 0; i < n; i++ {
			t, err := c.Translate(context.Background(), q.text)
			if err != nil {
				return err
			}
			v, err := c.Embed(context.Background(), []string{t.English + ". " + t.Answer}, store.Dims)
			if err != nil {
				return err
			}
			vec.Normalize(v[0])
			h := search.Top(store, v[0], 10, -1)
			scores = append(scores, float64(h[0].Score))
			if q.ok[h[0].Key] {
				hit1++
			}
		}
		sort.Float64s(scores)
		lo, hi := scores[0], scores[len(scores)-1]
		if len(q.ok) == 0 {
			offMax = max(offMax, hi)
			for _, s := range scores {
				offTotal++
				if s >= thr {
					offAbove++
				}
			}
		} else {
			onMin = min(onMin, lo)
			onHit1 += hit1
			for _, s := range scores {
				onTotal++
				if s < thr {
					onBelow++
				}
			}
		}
		kind := "on "
		if len(q.ok) == 0 {
			kind = "off"
		}
		fmt.Printf("%s min %.3f max %.3f top1ok %d/%d  %s\n", kind, lo, hi, hit1, n, q.text)
	}
	fmt.Printf("\non-topic runs below thr (tier 2): %d/%d, top1 acceptable %d/%d, lowest on-topic top1 %.3f\n", onBelow, onTotal, onHit1, onTotal, onMin)
	fmt.Printf("off-topic runs at or above thr (tier 1): %d/%d, highest off-topic top1 %.3f\n", offAbove, offTotal, offMax)
	return nil
}
