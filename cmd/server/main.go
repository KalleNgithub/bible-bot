// Command server serves the bible-bot HTTP GET API.
package main

import (
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/KalleNgithub/bible-bot/data"
	"github.com/KalleNgithub/bible-bot/internal/api"
	"github.com/KalleNgithub/bible-bot/internal/azure"
	"github.com/KalleNgithub/bible-bot/internal/ref"
	"github.com/KalleNgithub/bible-bot/internal/topic"
	"github.com/KalleNgithub/bible-bot/internal/vec"
)

func main() {
	en, fi, err := data.Load()
	if err != nil {
		log.Fatal(err)
	}
	store, err := vec.Read(data.Vectors())
	if err != nil {
		log.Fatal(err)
	}
	client, err := azure.FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	h := &api.Handler{
		EN: en, FI: fi,
		Parser: ref.NewParser(en, fi),
		Topic:  topic.New(client, store, en, rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))),
	}
	addr := ":" + envOr("PORT", "8080")
	log.Printf("listening on %s", addr)
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 60 * time.Second}
	log.Fatal(srv.ListenAndServe())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
