// Package kdriveapi exposes a thin, transport-layer client for the
// Infomaniak kDrive REST API.
//
// It handles authentication, timeouts, and exponential-backoff retries on
// transient failures (5xx and 429). Business logic stays in the infrastructure
// packages that depend on it.
package kdriveapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// DefaultBaseURL is the production endpoint of the kDrive v2 API.
const DefaultBaseURL = "https://api.infomaniak.com/2/drive"

const (
	defaultMaxRetries      = 3
	defaultInitialBackoff  = time.Second
	defaultRequestDeadline = 30 * time.Second
)

// Client is a reusable kDrive HTTP client scoped to a single drive.
//
// It is safe for concurrent use by multiple goroutines.
type Client struct {
	httpClient *http.Client
	logger     *slog.Logger
	baseURL    string
	driveID    string
	token      string
	maxRetries int
	backoff    time.Duration
}

// Options configures a Client at construction time.
type Options struct {
	BaseURL    string
	MaxRetries int
	Backoff    time.Duration
}

// NewClient returns a Client ready to talk to the drive identified by driveID.
//
// httpClient is injected so callers control transport-level tuning
// (timeouts, pooling). logger must be non-nil.
func NewClient(httpClient *http.Client, logger *slog.Logger, driveID, token string, opts Options) *Client {
	if opts.BaseURL == "" {
		opts.BaseURL = DefaultBaseURL
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = defaultMaxRetries
	}
	if opts.Backoff == 0 {
		opts.Backoff = defaultInitialBackoff
	}
	return &Client{
		httpClient: httpClient,
		logger:     logger,
		baseURL:    opts.BaseURL,
		driveID:    driveID,
		token:      token,
		maxRetries: opts.MaxRetries,
		backoff:    opts.Backoff,
	}
}

// Do executes an authenticated request against the drive-scoped endpoint.
//
// endpoint is the path relative to "/drive/{driveID}" (e.g. "/files/42/files").
// body may be nil. The returned response's Body must be closed by the caller.
func (c *Client) Do(ctx context.Context, method, endpoint string, body []byte) (*http.Response, error) {
	url := c.baseURL + "/" + c.driveID + endpoint

	var lastErr error
	backoff := c.backoff
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, backoff); err != nil {
				return nil, err
			}
			backoff *= 2
		}

		req, err := c.buildRequest(ctx, method, url, body)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logger.Warn("kdrive request failed",
				slog.String("method", method),
				slog.String("url", url),
				slog.Int("attempt", attempt+1),
				slog.String("err", err.Error()),
			)
			continue
		}

		if shouldRetry(resp.StatusCode) && attempt < c.maxRetries {
			drainAndClose(resp.Body)
			lastErr = fmt.Errorf("kdrive: transient status %d", resp.StatusCode)
			c.logger.Warn("kdrive transient status",
				slog.String("url", url),
				slog.Int("status", resp.StatusCode),
				slog.Int("attempt", attempt+1),
			)
			continue
		}

		if resp.StatusCode >= 400 {
			snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			drainAndClose(resp.Body)
			return nil, fmt.Errorf("kdrive: %s %s -> %d: %s",
				method, endpoint, resp.StatusCode, bytes.TrimSpace(snippet))
		}

		return resp, nil
	}

	if lastErr == nil {
		lastErr = errors.New("kdrive: exhausted retries")
	}
	return nil, lastErr
}

// DecodeJSON executes a request and decodes the JSON response into out.
func (c *Client) DecodeJSON(ctx context.Context, method, endpoint string, body []byte, out any) error {
	resp, err := c.Do(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) buildRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func shouldRetry(status int) bool {
	return status >= 500 || status == http.StatusTooManyRequests
}

func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck // passthrough: ctx cancellation is not our error to wrap
	case <-t.C:
		return nil
	}
}
