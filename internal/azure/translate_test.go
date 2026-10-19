package azure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func chatServer(t *testing.T, content string, got *map[string]any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openai/v1/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(got)
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": content}}}})
	}))
}

func TestTranslate(t *testing.T) {
	var req map[string]any
	srv := chatServer(t, `{"english":" my mom died ","lang":"FI","answer":" God comforts the grieving. "}`, &req)
	defer srv.Close()
	c := testClient(srv.URL)
	c.ChatDeployment = "chat"
	c.ReasoningEffort = "low"
	tr, err := c.Translate(context.Background(), "äitini kuoli")
	if err != nil {
		t.Fatal(err)
	}
	if tr.English != "my mom died" || tr.Lang != "fi" || tr.Answer != "God comforts the grieving." {
		t.Fatalf("got %+v", tr)
	}
	if req["model"] != "chat" || req["reasoning_effort"] != "low" || req["max_completion_tokens"] == nil {
		t.Fatalf("request = %v", req)
	}
}

func TestTranslateOmitsReasoningEffortWhenUnset(t *testing.T) {
	var req map[string]any
	srv := chatServer(t, `{"english":"hi","lang":"en"}`, &req)
	defer srv.Close()
	c := testClient(srv.URL)
	c.ChatDeployment = "chat"
	if _, err := c.Translate(context.Background(), "hi"); err != nil {
		t.Fatal(err)
	}
	if _, ok := req["reasoning_effort"]; ok {
		t.Fatalf("unexpected reasoning_effort: %v", req)
	}
}

func TestTranslateRejectsBadOutput(t *testing.T) {
	for _, content := range []string{`not json`, `{"english":"","lang":"en"}`} {
		var req map[string]any
		srv := chatServer(t, content, &req)
		c := testClient(srv.URL)
		c.ChatDeployment = "chat"
		if _, err := c.Translate(context.Background(), "x"); err == nil {
			t.Errorf("expected error for %q", content)
		}
		srv.Close()
	}
}

func TestTranslateNeedsDeployment(t *testing.T) {
	if _, err := testClient("http://x").Translate(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}
