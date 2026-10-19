// Package api is the HTTP GET interface: it routes a query to verse lookup or topic search.
package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/KalleNgithub/bible-bot/internal/bible"
	"github.com/KalleNgithub/bible-bot/internal/expand"
	"github.com/KalleNgithub/bible-bot/internal/quote"
	"github.com/KalleNgithub/bible-bot/internal/ref"
	"github.com/KalleNgithub/bible-bot/internal/topic"
)

const MaxQueryChars = 300

// TopicTimeout bounds the search service calls; on expiry the word-match tier answers.
const TopicTimeout = 10 * time.Second

// TopicFinder is satisfied by *topic.Finder.
type TopicFinder interface {
	Find(ctx context.Context, query string) topic.Result
}

type Handler struct {
	EN, FI *bible.Translation
	Parser *ref.Parser
	Topic  TopicFinder
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		fail(w, http.StatusMethodNotAllowed, "Use GET.")
		return
	}
	q := r.URL.Query()
	query, lang, mode := q.Get("q"), q.Get("lang"), q.Get("mode")
	switch {
	case query == "":
		fail(w, http.StatusBadRequest, "Missing parameter q.")
		return
	case utf8.RuneCountInString(query) > MaxQueryChars:
		fail(w, http.StatusBadRequest, "Parameter q is too long (max 300 characters).")
		return
	case lang != "" && lang != "fi" && lang != "en":
		fail(w, http.StatusBadRequest, "Parameter lang must be fi or en.")
		return
	case mode != "" && mode != "verse" && mode != "topic":
		fail(w, http.StatusBadRequest, "Parameter mode must be verse or topic.")
		return
	}

	if mode != "topic" {
		res, err := h.Parser.Parse(query)
		var nf *ref.NotFoundError
		switch {
		case err == nil:
			h.verse(w, res, lang)
			return
		case errors.As(err, &nf):
			fail(w, http.StatusNotFound, "Reference not found: "+nf.Msg)
			return
		case mode == "verse":
			fail(w, http.StatusNotFound, "Reference not found.")
			return
		}
	}
	h.topic(w, r.Context(), query, lang)
}

func (h *Handler) verse(w http.ResponseWriter, res ref.Result, lang string) {
	if lang == "" {
		lang = "en"
		if res.Finnish {
			lang = "fi"
		}
	}
	h.write(w, res.Book, res.Chapter, res.From, res.To, lang, "verse")
}

func (h *Handler) topic(w http.ResponseWriter, ctx context.Context, query, lang string) {
	ctx, cancel := context.WithTimeout(ctx, TopicTimeout)
	defer cancel()
	t := h.Topic.Find(ctx, query)
	if lang == "" {
		lang = "en"
		if t.Lang == "fi" {
			lang = "fi"
		}
	}
	from, to := expand.Range(h.EN, t.Ref.Book, t.Ref.Chapter, t.Ref.Verse, t.Ref.Verse)
	h.write(w, t.Ref.Book, t.Ref.Chapter, from, to, lang, "topic")
}

func (h *Handler) write(w http.ResponseWriter, book, chapter, from, to int, lang, kind string) {
	tr := h.EN
	if lang == "fi" {
		tr = h.FI
	}
	text, ok := quote.Format(tr, book, chapter, from, to)
	if !ok {
		text, ok = quote.Format(h.EN, book, chapter, from, to)
	}
	if !ok {
		fail(w, http.StatusNotFound, "Reference not found.")
		return
	}
	// Query text is never logged.
	log.Printf("ok mode=%s lang=%s", kind, lang)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(text + "\n"))
}

func fail(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(msg + "\n"))
}
