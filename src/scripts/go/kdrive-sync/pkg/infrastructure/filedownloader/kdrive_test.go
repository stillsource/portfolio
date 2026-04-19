package filedownloader_test

import (
	"context"
	"io"
	"kdrive-sync/pkg/infrastructure/filedownloader"
	"kdrive-sync/pkg/infrastructure/kdriveapi"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

var _ = Describe("KDrive.DownloadFile", func() {
	It("reads the raw bytes served by the /files/<id>/download endpoint", func() {
		var gotPath string
		payload := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(payload)
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		dl := filedownloader.NewKDrive(client)

		got, err := dl.DownloadFile(context.Background(), "i1")
		Expect(err).NotTo(HaveOccurred())
		Expect(gotPath).To(Equal("/42/files/i1/download"))
		Expect(got).To(Equal(payload))
	})

	It("wraps the error with 'download <id>' on 404", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		dl := filedownloader.NewKDrive(client)

		_, err := dl.DownloadFile(context.Background(), "missing")
		Expect(err).To(MatchError(ContainSubstring("download missing")))
	})
})
