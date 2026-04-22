package kdriveadapter_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stillsource/kdrive-fuse/kdrive"
	"github.com/stillsource/kdrive-fuse/kdrive/kdrivefakes"

	"kdrive-sync/pkg/domain"
	"kdrive-sync/pkg/infrastructure/kdriveadapter"
)

var _ = Describe("FileLister", func() {
	var (
		fake *kdrivefakes.FilesFake
		a    *kdriveadapter.FileLister
		ctx  context.Context
	)

	BeforeEach(func() {
		fake = &kdrivefakes.FilesFake{}
		a = kdriveadapter.NewFileLister(fake)
		ctx = context.Background()
	})

	It("converts lib FileInfo to domain.DriveFile", func() {
		fake.ListResults = map[int64]kdrivefakes.ListResult{
			42: {Files: []kdrive.FileInfo{
				{ID: 10, Name: "a.txt", Type: kdrive.FileTypeFile, CreatedAt: 1_700_000_000},
				{ID: 20, Name: "sub", Type: kdrive.FileTypeDir, CreatedAt: 1_700_000_100},
			}},
		}
		files, err := a.ListFiles(ctx, "42")
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(2))
		Expect(files[0].ID).To(Equal("10"))
		Expect(files[0].Name).To(Equal("a.txt"))
		Expect(files[0].Type).To(Equal(domain.DriveFileType("file")))
		Expect(files[1].IsDir()).To(BeTrue())
	})

	It("rejects non-numeric folder IDs", func() {
		_, err := a.ListFiles(ctx, "abc")
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, kdrive.ErrValidation)).To(BeTrue())
	})

	It("propagates lib errors with context", func() {
		fake.ListResults = map[int64]kdrivefakes.ListResult{
			1: {Err: kdrive.ErrNotFound},
		}
		_, err := a.ListFiles(ctx, "1")
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, kdrive.ErrNotFound)).To(BeTrue())
	})
})

var _ = Describe("FileDownloader", func() {
	It("returns raw bytes for a valid id", func() {
		fake := &kdrivefakes.FilesFake{
			DownloadResults: map[int64]kdrivefakes.DownloadResult{7: {Data: []byte("hello")}},
		}
		a := kdriveadapter.NewFileDownloader(fake)
		b, err := a.DownloadFile(context.Background(), "7")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("hello"))
	})

	It("rejects non-numeric ids", func() {
		a := kdriveadapter.NewFileDownloader(&kdrivefakes.FilesFake{})
		_, err := a.DownloadFile(context.Background(), "oops")
		Expect(errors.Is(err, kdrive.ErrValidation)).To(BeTrue())
	})

	It("propagates lib errors", func() {
		fake := &kdrivefakes.FilesFake{
			DownloadResults: map[int64]kdrivefakes.DownloadResult{1: {Err: kdrive.ErrNotFound}},
		}
		a := kdriveadapter.NewFileDownloader(fake)
		_, err := a.DownloadFile(context.Background(), "1")
		Expect(errors.Is(err, kdrive.ErrNotFound)).To(BeTrue())
	})
})

var _ = Describe("SharePublisher", func() {
	It("returns the share URL from the lib", func() {
		fake := &kdrivefakes.SharesFake{
			PublishResults: map[int64]kdrivefakes.PublishResult{
				9: {Info: kdrive.ShareInfo{ShareURL: "https://s/9"}},
			},
		}
		a := kdriveadapter.NewSharePublisher(fake)
		got, err := a.PublishShare(context.Background(), "9")
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal("https://s/9"))
	})

	It("rejects non-numeric ids", func() {
		a := kdriveadapter.NewSharePublisher(&kdrivefakes.SharesFake{})
		_, err := a.PublishShare(context.Background(), "bad")
		Expect(errors.Is(err, kdrive.ErrValidation)).To(BeTrue())
	})

	It("propagates lib errors", func() {
		fake := &kdrivefakes.SharesFake{
			PublishResults: map[int64]kdrivefakes.PublishResult{1: {Err: kdrive.ErrAuth}},
		}
		a := kdriveadapter.NewSharePublisher(fake)
		_, err := a.PublishShare(context.Background(), "1")
		Expect(errors.Is(err, kdrive.ErrAuth)).To(BeTrue())
	})
})
