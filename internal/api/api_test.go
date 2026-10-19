package api

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/KalleNgithub/bible-bot/data"
	"github.com/KalleNgithub/bible-bot/internal/bible"
	"github.com/KalleNgithub/bible-bot/internal/ref"
	"github.com/KalleNgithub/bible-bot/internal/topic"
)

type stub struct {
	res   topic.Result
	calls int
}

func (s *stub) Find(context.Context, string) topic.Result { s.calls++; return s.res }

func setup(t *testing.T) (*Handler, *stub) {
	en, fi, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	s := &stub{res: topic.Result{Ref: bible.Ref{Book: 19, Chapter: 34, Verse: 18}, Tier: topic.Primary, Lang: "en"}}
	return &Handler{EN: en, FI: fi, Parser: ref.NewParser(en, fi), Topic: s}, s
}

func get(h http.Handler, params url.Values) (int, string, string) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/?"+params.Encode(), nil))
	return rec.Code, rec.Header().Get("Content-Type"), rec.Body.String()
}

func TestValidation(t *testing.T) {
	h, _ := setup(t)
	for name, p := range map[string]url.Values{
		"missing":  {},
		"empty":    {"q": {""}},
		"long":     {"q": {strings.Repeat("a", 301)}},
		"bad lang": {"q": {"x"}, "lang": {"sv"}},
		"bad mode": {"q": {"x"}, "mode": {"foo"}},
	} {
		code, ct, body := get(h, p)
		if code != 400 || !strings.HasPrefix(ct, "text/plain") || body == "" {
			t.Errorf("%s: %d %q %q", name, code, ct, body)
		}
	}
}

func TestVerseRouting(t *testing.T) {
	h, s := setup(t)
	code, ct, body := get(h, url.Values{"q": {"Jhon 11:25 to 26"}})
	if code != 200 || ct != "text/plain; charset=utf-8" || !strings.HasPrefix(body, "John 11:25-26: ") || s.calls != 0 {
		t.Fatalf("%d %q %q calls=%d", code, ct, body, s.calls)
	}
	_, _, fi := get(h, url.Values{"q": {"John 11:25"}, "lang": {"fi"}})
	if strings.HasPrefix(fi, "John") || fi == body {
		t.Fatalf("expected Finnish: %q", fi)
	}
	_, _, fiName := get(h, url.Values{"q": {"Joh 3:16"}})
	_, _, fiBook := get(h, url.Values{"q": {"Johannes 3:16"}})
	if !strings.Contains(fiBook, "3:16") || fiName == "" {
		t.Fatalf("%q", fiBook)
	}
}

func TestNotFound(t *testing.T) {
	h, s := setup(t)
	code, _, body := get(h, url.Values{"q": {"John 99:1"}})
	if code != 404 || body == "" || s.calls != 0 {
		t.Fatalf("%d %q", code, body)
	}
	if code, _, _ = get(h, url.Values{"q": {"hello there"}, "mode": {"verse"}}); code != 404 {
		t.Fatalf("verse mode on text: %d", code)
	}
}

func TestTopicRouting(t *testing.T) {
	h, s := setup(t)
	_, _, body := get(h, url.Values{"q": {"what happens to people who believe in jesus"}})
	if !strings.HasPrefix(body, "Psalms 34:18") && !strings.HasPrefix(body, "Psalm 34:18") || s.calls != 1 {
		t.Fatalf("%q calls=%d", body, s.calls)
	}
	get(h, url.Values{"q": {"John 11:25"}, "mode": {"topic"}})
	if s.calls != 2 {
		t.Fatalf("explicit topic mode must skip the parser, calls=%d", s.calls)
	}
}

func TestLanguageDefaults(t *testing.T) {
	h, s := setup(t)
	_, _, en := get(h, url.Values{"q": {"free text"}})
	s.res.Lang = "fi"
	_, _, fi := get(h, url.Values{"q": {"vapaa teksti"}})
	_, _, forced := get(h, url.Values{"q": {"vapaa teksti"}, "lang": {"en"}})
	if en == fi || forced != en {
		t.Fatalf("en=%q fi=%q forced=%q", en, fi, forced)
	}
	s.res.Lang = "sv"
	if _, _, sv := get(h, url.Values{"q": {"hej"}}); sv != en {
		t.Fatalf("other language should answer in English: %q", sv)
	}
}

func TestQueryNotLogged(t *testing.T) {
	h, _ := setup(t)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	get(h, url.Values{"q": {"secret grief text"}})
	get(h, url.Values{"q": {"John 99:1 secret"}})
	if strings.Contains(buf.String(), "secret") {
		t.Fatalf("query logged: %q", buf.String())
	}
}
