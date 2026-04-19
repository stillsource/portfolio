package kdriveapi

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

// TestNewClient_AppliesDefaults verifies that passing a zero-value Options
// fills in the package defaults for BaseURL, MaxRetries and Backoff. This
// exercises the `if opts.Field == ""` / `== 0` branches in NewClient that are
// otherwise skipped when callers supply explicit values.
func TestNewClient_AppliesDefaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := NewClient(http.DefaultClient, logger, "42", "tok", Options{})

	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.maxRetries != defaultMaxRetries {
		t.Errorf("maxRetries = %d, want %d", c.maxRetries, defaultMaxRetries)
	}
	if c.backoff != defaultInitialBackoff {
		t.Errorf("backoff = %v, want %v", c.backoff, defaultInitialBackoff)
	}
}

// TestNewClient_ExplicitOptionsOverrideDefaults complements the above by
// asserting the non-default branches are also honoured.
func TestNewClient_ExplicitOptionsOverrideDefaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	opts := Options{
		BaseURL:    "https://example.invalid",
		MaxRetries: 7,
		Backoff:    42 * time.Millisecond,
	}
	c := NewClient(http.DefaultClient, logger, "1", "t", opts)

	if c.baseURL != opts.BaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, opts.BaseURL)
	}
	if c.maxRetries != opts.MaxRetries {
		t.Errorf("maxRetries = %d, want %d", c.maxRetries, opts.MaxRetries)
	}
	if c.backoff != opts.Backoff {
		t.Errorf("backoff = %v, want %v", c.backoff, opts.Backoff)
	}
}
