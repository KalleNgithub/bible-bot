package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/KalleNgithub/bible-bot/internal/bible"
)

type getBible struct {
	Books []struct {
		Nr       int    `json:"nr"`
		Name     string `json:"name"`
		Chapters []struct {
			Chapter int `json:"chapter"`
			Verses  []struct {
				Chapter int    `json:"chapter"`
				Verse   int    `json:"verse"`
				Text    string `json:"text"`
			} `json:"verses"`
		} `json:"chapters"`
	} `json:"books"`
}

var spaceFix = strings.NewReplacer("\t", " ", "\r", " ", "\n", " ")

func download(id, outPath string) error {
	resp, err := http.Get("https://api.getbible.net/v2/" + id + ".json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: HTTP %d", id, resp.StatusCode)
	}
	var g getBible
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return err
	}
	t := bible.New()
	for _, b := range g.Books {
		if b.Nr < 1 || b.Nr > 66 {
			continue
		}
		t.SetBook(b.Nr, strings.TrimSpace(spaceFix.Replace(b.Name)))
		for _, c := range b.Chapters {
			for _, v := range c.Verses {
				t.Add(bible.Ref{Book: b.Nr, Chapter: c.Chapter, Verse: v.Verse}, strings.TrimSpace(spaceFix.Replace(v.Text)))
			}
		}
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := t.Write(f); err != nil {
		return err
	}
	fmt.Printf("%s: %d verses -> %s\n", id, len(t.Refs()), outPath)
	return nil
}
