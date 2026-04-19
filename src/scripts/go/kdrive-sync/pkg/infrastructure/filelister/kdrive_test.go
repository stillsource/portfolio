package filelister_test

import (
	"context"
	"io"
	"kdrive-sync/pkg/domain"
	"kdrive-sync/pkg/infrastructure/filelister"
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

var _ = Describe("KDrive.ListFiles", func() {
	It("decodes drive entries and converts id/created_at", func() {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[
				{"id":12,"name":"Nuit","type":"dir","created_at":1710000000},
				{"id":99,"name":"photo.jpg","type":"file","created_at":1710000100}
			]}`))
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		lister := filelister.NewKDrive(client)

		got, err := lister.ListFiles(context.Background(), "root")
		Expect(err).NotTo(HaveOccurred())
		Expect(gotPath).To(Equal("/42/files/root/files"))
		Expect(got).To(HaveLen(2))

		Expect(got[0]).To(Equal(domain.DriveFile{
			ID:        "12",
			Name:      "Nuit",
			Type:      domain.DriveFileTypeDir,
			CreatedAt: time.Unix(1710000000, 0),
		}))
		Expect(got[1].ID).To(Equal("99"))
		Expect(got[1].Type).To(Equal(domain.DriveFileTypeFile))
	})

	It("returns an empty slice when data is empty", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		lister := filelister.NewKDrive(client)

		got, err := lister.ListFiles(context.Background(), "empty")
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	It("wraps the HTTP error with 'list files <id>'", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		client := kdriveapi.NewClient(srv.Client(), silentLogger(), "42", "tok",
			kdriveapi.Options{BaseURL: srv.URL, MaxRetries: 0, Backoff: time.Millisecond})
		lister := filelister.NewKDrive(client)

		_, err := lister.ListFiles(context.Background(), "bad")
		Expect(err).To(MatchError(ContainSubstring("list files bad")))
	})
})
