package azure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testClient(url string) *Client {
	return &Client{Endpoint: url, Key: "k", EmbedDeployment: "emb", Backoff: func(int) time.Duration { return time.Millisecond }}
}

func embedHandler(dims int, shuffle bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Input      []string `json:"input"`
			Dimensions int      `json:"dimensions"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		type item struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		}
		items := make([]item, len(req.Input))
		for i := range req.Input {
			e := make([]float32, dims)
			e[0] = float32(i + 1)
			items[i] = item{i, e}
		}
		if shuffle && len(items) > 1 {
			items[0], items[1] = items[1], items[0]
		}
		json.NewEncoder(w).Encode(map[string]any{"data": items})
	}
}

func TestEmbedRequestAndOrdering(t *testing.T) {
	var path, key string
	h := embedHandler(4, true)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, key = r.URL.String(), r.Header.Get("api-key")
		h(w, r)
	}))
	defer srv.Close()
	got, err := testClient(srv.URL).Embed(context.Background(), []string{"a", "b"}, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, "/openai/deployments/emb/embeddings?api-version=") || key != "k" {
		t.Fatalf("path=%q key=%q", path, key)
	}
	if got[0][0] != 1 || got[1][0] != 2 {
		t.Fatalf("results not in input order: %v", got)
	}
}

func TestEmbedRetriesOn429(t *testing.T) {
	calls := 0
	h := embedHandler(2, false)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		h(w, r)
	}))
	defer srv.Close()
	if _, err := testClient(srv.URL).Embed(context.Background(), []string{"a"}, 2); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestEmbedDoesNotRetryClientErrors(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(404)
	}))
	defer srv.Close()
	if _, err := testClient(srv.URL).Embed(context.Background(), []string{"a"}, 2); err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestEmbedGivesUp(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(503)
	}))
	defer srv.Close()
	if _, err := testClient(srv.URL).Embed(context.Background(), []string{"a"}, 2); err == nil {
		t.Fatal("expected error")
	}
	if calls != maxAttempts {
		t.Fatalf("calls = %d", calls)
	}
}

func TestEmbedRejectsWrongDimensions(t *testing.T) {
	srv := httptest.NewServer(embedHandler(3, false))
	defer srv.Close()
	if _, err := testClient(srv.URL).Embed(context.Background(), []string{"a"}, 4); err == nil {
		t.Fatal("expected dimension error")
	}
}
