package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientDefaults(t *testing.T) {
	c := NewClient(Options{Uri: "https://example.invalid"})
	if c.client.RetryWaitMin != time.Second || c.client.RetryWaitMax != 30*time.Second {
		t.Errorf("backoff = %v..%v, want 1s..30s", c.client.RetryWaitMin, c.client.RetryWaitMax)
	}
	if c.client.HTTPClient.Timeout != 180*time.Second {
		t.Errorf("timeout = %v, want 180s", c.client.HTTPClient.Timeout)
	}
	c = NewClient(Options{Uri: "https://example.invalid", RequestTimeout: 7, MinBackoff: 2, MaxBackoff: 9})
	if c.client.HTTPClient.Timeout != 7*time.Second || c.client.RetryWaitMin != 2*time.Second || c.client.RetryWaitMax != 9*time.Second {
		t.Errorf("overrides not applied: %v %v %v", c.client.HTTPClient.Timeout, c.client.RetryWaitMin, c.client.RetryWaitMax)
	}
}

// A firewall that stops answering must fail the call after one timeout,
// not hang and not be retried.
func TestRequestTimeoutIsNotRetried(t *testing.T) {
	var hits atomic.Int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)

	c := NewClient(Options{Uri: srv.URL, RequestTimeout: 1, MinBackoff: 1, MaxBackoff: 1, MaxRetries: 3})
	start := time.Now()
	var out map[string]any
	err := c.doRequest(context.Background(), "POST", "/wireguard/server/addServer", map[string]any{}, &out)
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("took %v, want about 1s", d)
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("server saw %d attempts, want 1", n)
	}
}

// Server errors are still retried, now with a real backoff.
func TestServerErrorsStillRetried(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"result":"saved"}`))
	}))
	defer srv.Close()

	c := NewClient(Options{Uri: srv.URL, MinBackoff: 1, MaxBackoff: 1})
	start := time.Now()
	var out map[string]any
	if err := c.doRequest(context.Background(), "POST", "/core/x/set", map[string]any{}, &out); err != nil {
		t.Fatalf("doRequest: %v", err)
	}
	if n := hits.Load(); n != 3 {
		t.Errorf("attempts = %d, want 3", n)
	}
	if d := time.Since(start); d < 2*time.Second {
		t.Errorf("two retries took %v; the backoff is not being applied", d)
	}
}
