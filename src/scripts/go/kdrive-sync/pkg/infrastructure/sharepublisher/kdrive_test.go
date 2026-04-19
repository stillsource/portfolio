package sharepublisher_test

import (
	"context"
	"encoding/json"
	"io"
	"kdrive-sync/pkg/infrastructure/kdriveapi"
	"kdrive-sync/pkg/infrastructure/sharepublisher"
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

var _ = Describe("KDrive.PublishShare", func() {
	It("reuses an existing share without POSTing", func() {
		var posts int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":[{"share_url":"https://share/abc"}]}`))
			case http.MethodPost:
				atomic.AddInt32(&posts, 1)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"share_url":"https://share/new"}}`))
			}
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		pub := sharepublisher.NewKDrive(client)

		url, err := pub.PublishShare(context.Background(), "i1")
		Expect(err).NotTo(HaveOccurred())
		Expect(url).To(Equal("https://share/abc"))
		Expect(atomic.LoadInt32(&posts)).To(BeZero())
	})

	It("creates a public, non-protected share when none exists", func() {
		var gotBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":[]}`))
			case http.MethodPost:
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"share_url":"https://share/created"}}`))
			}
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		pub := sharepublisher.NewKDrive(client)

		url, err := pub.PublishShare(context.Background(), "i1")
		Expect(err).NotTo(HaveOccurred())
		Expect(url).To(Equal("https://share/created"))
		Expect(gotBody).To(HaveKeyWithValue("type", "public"))
		Expect(gotBody).To(HaveKeyWithValue("password_protected", false))
	})

	It("returns an 'empty share url' error when POST returns a blank share_url", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":[]}`))
			case http.MethodPost:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"share_url":""}}`))
			}
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		pub := sharepublisher.NewKDrive(client)

		_, err := pub.PublishShare(context.Background(), "i1")
		Expect(err).To(MatchError(ContainSubstring("empty share url")))
	})

	It("wraps the error with 'create share <id>' when the POST fails", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":[]}`))
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("nope"))
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		pub := sharepublisher.NewKDrive(client)

		_, err := pub.PublishShare(context.Background(), "i1")
		Expect(err).To(MatchError(ContainSubstring("create share i1")))
	})
})
