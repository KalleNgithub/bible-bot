package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/KalleNgithub/bible-bot/data"
	"github.com/KalleNgithub/bible-bot/internal/azure"
	"github.com/KalleNgithub/bible-bot/internal/vec"
)

const embedBatch = 128

type embedder interface {
	Embed(ctx context.Context, texts []string, dims int) ([][]float32, error)
}

func embedCmd(args []string) error {
	fs := flag.NewFlagSet("embed", flag.ExitOnError)
	limit := fs.Int("limit", 0, "embed only the first N verses (0 = all)")
	out := fs.String("out", "data/vectors.bin", "output file")
	dims := fs.Int("dims", 512, "embedding dimensions")
	fs.Parse(args)

	c, err := azure.FromEnv()
	if err != nil {
		return err
	}
	en, _, err := data.Load()
	if err != nil {
		return err
	}
	refs := en.Refs()
	if *limit > 0 && *limit < len(refs) {
		refs = refs[:*limit]
	}
	s := &vec.Store{Dims: *dims}
	texts := make([]string, len(refs))
	for i, r := range refs {
		texts[i], _ = en.Text(r)
		s.Keys = append(s.Keys, uint32(r.Book*1000000+r.Chapter*1000+r.Verse))
	}
	s.Data, err = embedAll(context.Background(), c, texts, *dims, func(done int) {
		fmt.Fprintf(os.Stderr, "\r%d/%d", done, len(texts))
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr)
	f, err := os.Create(*out)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := s.Write(f); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %d vectors to %s\n", len(s.Keys), *out)
	return f.Close()
}

// embedAll embeds texts in batches and returns the unit-length vectors, flattened.
func embedAll(ctx context.Context, e embedder, texts []string, dims int, progress func(int)) ([]float32, error) {
	all := make([]float32, 0, len(texts)*dims)
	for i := 0; i < len(texts); i += embedBatch {
		end := min(i+embedBatch, len(texts))
		vs, err := e.Embed(ctx, texts[i:end], dims)
		if err != nil {
			return nil, fmt.Errorf("batch at verse %d: %w", i, err)
		}
		for _, v := range vs {
			vec.Normalize(v)
			all = append(all, v...)
		}
		progress(end)
	}
	return all, nil
}
