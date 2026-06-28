package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Interface is the single outbound HTTP gateway. Every third-party repository
// (CoinGecko, Yahoo Finance, MetalPrice, ExchangeRate, Telegram) calls through
// it instead of touching net/http directly, so timeouts, JSON handling, and
// error mapping live in one place. The decoded JSON body is returned as the
// result value.
type Interface interface {
	GetJSON(ctx context.Context, req Request) (map[string]any, error)
	GetJSONArray(ctx context.Context, req Request) ([]map[string]any, error)
	PostJSON(ctx context.Context, req Request) (map[string]any, error)
}

// Request describes a single outbound call. Body is JSON-encoded when non-nil.
type Request struct {
	URL     string
	Headers map[string]string
	Query   map[string]string
	Body    any
}

// StatusError is returned when the response status is not 2xx. It carries the
// status code and a (truncated) body so callers can branch on it via errors.As.
type StatusError struct {
	StatusCode int
	Body       string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected status %d: %s", e.StatusCode, e.Body)
}

type client struct {
	httpClient *http.Client
}

// Init builds a client with the given per-request timeout.
func Init(timeout time.Duration) Interface {
	return &client{httpClient: &http.Client{Timeout: timeout}}
}

func (c *client) GetJSON(ctx context.Context, req Request) (map[string]any, error) {
	respBody, err := c.do(ctx, http.MethodGet, req)
	if err != nil || len(respBody) == 0 {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *client) GetJSONArray(ctx context.Context, req Request) ([]map[string]any, error) {
	respBody, err := c.do(ctx, http.MethodGet, req)
	if err != nil || len(respBody) == 0 {
		return nil, err
	}

	var result []map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *client) PostJSON(ctx context.Context, req Request) (map[string]any, error) {
	respBody, err := c.do(ctx, http.MethodPost, req)
	if err != nil || len(respBody) == 0 {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *client) do(ctx context.Context, method string, req Request) ([]byte, error) {
	fullURL, err := buildURL(req.URL, req.Query)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if req.Body != nil {
		encoded, err := json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode, Body: truncate(string(respBody), 512)}
	}

	return respBody, nil
}

func buildURL(rawURL string, query map[string]string) (string, error) {
	if len(query) == 0 {
		return rawURL, nil
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := parsed.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
