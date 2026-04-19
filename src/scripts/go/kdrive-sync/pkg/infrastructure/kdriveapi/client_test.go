package kdriveapi_test

import (
	"context"
	"errors"
	"io"
	"kdrive-sync/pkg/infrastructure/kdriveapi"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// fastOpts returns Options with a millisecond backoff so the suite stays fast.
func fastOpts(baseURL string, maxRetries int) kdriveapi.Options {
	return kdriveapi.Options{BaseURL: baseURL, MaxRetries: maxRetries, Backoff: time.Millisecond}
}

var _ = Describe("Client.Do", func() {
	Context("authentication", func() {
		It("injects Bearer <token> into the Authorization header", func() {
			var gotAuth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth = r.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok", fastOpts(srv.URL, 3))
			resp, err := c.Do(context.Background(), "GET", "/files", nil)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()
			Expect(gotAuth).To(Equal("Bearer tok"))
		})

		It("adds Content-Type application/json when body is non-nil", func() {
			var gotCT string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCT = r.Header.Get("Content-Type")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok", fastOpts(srv.URL, 0))
			resp, err := c.Do(context.Background(), "POST", "/x", []byte(`{"k":"v"}`))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()
			Expect(gotCT).To(Equal("application/json"))
		})
	})

	Context("retry & backoff", func() {
		It("retries on 503 and eventually succeeds", func() {
			var hits int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				n := atomic.AddInt32(&hits, 1)
				if n < 3 {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`ok`))
			}))
			defer srv.Close()

			c := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok", fastOpts(srv.URL, 3))
			resp, err := c.Do(context.Background(), "GET", "/x", nil)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(atomic.LoadInt32(&hits)).To(Equal(int32(3)))
		})

		It("retries on 429 (Too Many Requests)", func() {
			var hits int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				n := atomic.AddInt32(&hits, 1)
				if n < 2 {
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`ok`))
			}))
			defer srv.Close()

			c := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok", fastOpts(srv.URL, 3))
			resp, err := c.Do(context.Background(), "GET", "/x", nil)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()
			Expect(atomic.LoadInt32(&hits)).To(Equal(int32(2)))
		})

		It("gives up after MaxRetries and returns the last error", func() {
			var hits int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&hits, 1)
				w.WriteHeader(http.StatusServiceUnavailable)
			}))
			defer srv.Close()

			c := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok", fastOpts(srv.URL, 2))
			_, err := c.Do(context.Background(), "GET", "/x", nil)
			Expect(err).To(HaveOccurred())
			// MaxRetries=2 means: 1 initial attempt + 2 retries = 3 hits.
			Expect(atomic.LoadInt32(&hits)).To(Equal(int32(3)))
		})

		It("does not retry on 4xx (except 429) and surfaces the body snippet", func() {
			var hits int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&hits, 1)
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("invalid payload"))
			}))
			defer srv.Close()

			c := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok", fastOpts(srv.URL, 3))
			_, err := c.Do(context.Background(), "GET", "/x", nil)
			Expect(err).To(HaveOccurred())
			Expect(atomic.LoadInt32(&hits)).To(Equal(int32(1)))
			Expect(err).To(MatchError(ContainSubstring("400")))
			Expect(err).To(MatchError(ContainSubstring("invalid payload")))
		})

		It("honors context cancellation while waiting on backoff", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			}))
			defer srv.Close()

			// 500 ms backoff, context cancelled almost immediately — the error
			// must surface as ctx.DeadlineExceeded or ctx.Canceled.
			c := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
				kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 3, Backoff: 500 * time.Millisecond})

			cctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			_, err := c.Do(cctx, "GET", "/x", nil)
			Expect(err).To(HaveOccurred())
			Expect(errorsContainsCtx(err)).To(BeTrue(), "err should wrap ctx.Err(): %v", err)
		})
	})

	Context("URL construction", func() {
		It("concatenates BaseURL + /driveID + endpoint", func() {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok", fastOpts(srv.URL, 0))
			resp, err := c.Do(context.Background(), "GET", "/files/99/files", nil)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()
			Expect(gotPath).To(Equal("/42/files/99/files"))
		})
	})
})

var _ = Describe("Client.Do buildRequest failure", func() {
	It("returns an error when the HTTP method is invalid", func() {
		// An invalid method (containing whitespace) makes
		// http.NewRequestWithContext fail before the request is dispatched,
		// exercising the buildRequest error path of Do.
		c := kdriveapi.NewClient(http.DefaultClient, silentLogger(), "42", "tok",
			fastOpts("http://example.invalid", 0))
		_, err := c.Do(context.Background(), "BAD METHOD", "/x", nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("build request"))
	})
})

var _ = Describe("Client.DecodeJSON", func() {
	It("decodes the JSON body into the target struct", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"name":"x","id":42}}`))
		}))
		defer srv.Close()

		c := kdriveapi.NewClient(srv.Client(), silentLogger(), "1", "tok", fastOpts(srv.URL, 0))
		var out struct {
			Data struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"data"`
		}
		Expect(c.DecodeJSON(context.Background(), "GET", "/x", nil, &out)).To(Succeed())
		Expect(out.Data.Name).To(Equal("x"))
		Expect(out.Data.ID).To(Equal(42))
	})

	It("returns an error wrapped with 'decode response' when JSON is invalid", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`not json`))
		}))
		defer srv.Close()

		c := kdriveapi.NewClient(srv.Client(), silentLogger(), "1", "tok", fastOpts(srv.URL, 0))
		var out struct{}
		err := c.DecodeJSON(context.Background(), "GET", "/x", nil, &out)
		Expect(err).To(MatchError(ContainSubstring("decode response")))
	})

	It("propagates the transport-level error from Do", func() {
		// Target a closed server so Do fails at the TCP stage for every
		// attempt; DecodeJSON must surface that error without wrapping as
		// 'decode response'.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		srv.Close() // shut it down immediately

		c := kdriveapi.NewClient(srv.Client(), silentLogger(), "1", "tok", fastOpts(srv.URL, 0))
		var out struct{}
		err := c.DecodeJSON(context.Background(), "GET", "/x", nil, &out)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).NotTo(ContainSubstring("decode response"))
	})
})

// errorsContainsCtx reports whether err matches context.Canceled or DeadlineExceeded.
func errorsContainsCtx(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
