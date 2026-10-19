// Package azure is a minimal Azure OpenAI client.
package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiVersion = "2024-10-21"

type Client struct {
	Endpoint        string
	Key             string
	EmbedDeployment string
	ChatDeployment  string
	// ReasoningEffort is sent to the chat model when set (reasoning models only).
	ReasoningEffort string
	HTTP            *http.Client
	// Backoff returns the wait before retry attempt n (starting at 1).
	Backoff func(n int) time.Duration
}

// FromEnv builds a client from AZURE_OPENAI_* environment variables.
func FromEnv() (*Client, error) {
	c := &Client{
		Endpoint:        os.Getenv("AZURE_OPENAI_ENDPOINT"),
		Key:             os.Getenv("AZURE_OPENAI_KEY"),
		EmbedDeployment: os.Getenv("AZURE_OPENAI_EMBED_DEPLOYMENT"),
		ChatDeployment:  os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT"),
		ReasoningEffort: os.Getenv("AZURE_OPENAI_REASONING_EFFORT"),
	}
	for name, v := range map[string]string{
		"AZURE_OPENAI_ENDPOINT":         c.Endpoint,
		"AZURE_OPENAI_KEY":              c.Key,
		"AZURE_OPENAI_EMBED_DEPLOYMENT": c.EmbedDeployment,
	} {
		if v == "" {
			return nil, fmt.Errorf("%s is not set", name)
		}
	}
	return c, nil
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (c *Client) backoff(n int) time.Duration {
	if c.Backoff != nil {
		return c.Backoff(n)
	}
	return time.Duration(1<<n) * time.Second
}

const maxAttempts = 6

// post sends a JSON request, retrying on 429 and 5xx, and decodes the response into out.
func (c *Client) post(ctx context.Context, url string, body, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-key", c.Key)
		resp, err := c.httpClient().Do(req)
		wait := c.backoff(attempt)
		if err != nil {
			lastErr = err
		} else {
			data, rerr := io.ReadAll(resp.Body)
			resp.Body.Close()
			switch {
			case rerr != nil:
				lastErr = rerr
			case resp.StatusCode == 200:
				return json.Unmarshal(data, out)
			case resp.StatusCode == 429 || resp.StatusCode >= 500:
				lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, snippet(data))
				if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil {
					wait = time.Duration(s) * time.Second
				}
			default:
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, snippet(data))
			}
		}
		if attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return fmt.Errorf("azure openai: giving up after %d attempts: %w", maxAttempts, lastErr)
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 300 {
		s = s[:300]
	}
	return s
}

// Embed returns one vector per input text, requested at the given dimensions.
func (c *Client) Embed(ctx context.Context, texts []string, dims int) ([][]float32, error) {
	url := fmt.Sprintf("%s/openai/deployments/%s/embeddings?api-version=%s",
		strings.TrimRight(c.Endpoint, "/"), c.EmbedDeployment, apiVersion)
	var out struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	req := map[string]any{"input": texts, "dimensions": dims}
	if err := c.post(ctx, url, req, &out); err != nil {
		return nil, err
	}
	if len(out.Data) != len(texts) {
		return nil, fmt.Errorf("azure openai: got %d embeddings for %d inputs", len(out.Data), len(texts))
	}
	res := make([][]float32, len(texts))
	for _, d := range out.Data {
		if d.Index < 0 || d.Index >= len(texts) || res[d.Index] != nil {
			return nil, fmt.Errorf("azure openai: bad embedding index %d", d.Index)
		}
		if len(d.Embedding) != dims {
			return nil, fmt.Errorf("azure openai: got %d dimensions, want %d", len(d.Embedding), dims)
		}
		res[d.Index] = d.Embedding
	}
	return res, nil
}
