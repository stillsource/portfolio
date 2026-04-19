package usecase_test

import (
	"context"
	"errors"
	"io"
	"kdrive-sync/pkg/domain"
	"kdrive-sync/pkg/service/servicefakes"
	"kdrive-sync/pkg/usecase"
	"log/slog"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

// silentLogger returns a slog.Logger that discards all output so the Ginkgo
// console stays readable.
func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// dir builds a folder DriveFile.
func dir(id, name string, createdAt time.Time) domain.DriveFile {
	return domain.DriveFile{ID: id, Name: name, Type: domain.DriveFileTypeDir, CreatedAt: createdAt}
}

// file builds a plain file DriveFile.
func file(id, name string) domain.DriveFile {
	return domain.DriveFile{ID: id, Name: name, Type: domain.DriveFileTypeFile}
}

var _ = Describe("SyncRolls.Execute", func() {
	var (
		lister   *servicefakes.FileListerFake
		dl       *servicefakes.FileDownloaderFake
		pub      *servicefakes.SharePublisherFake
		analyzer *servicefakes.ImageAnalyzerFake
		aggr     *servicefakes.PaletteAggregatorFake
		poet     *servicefakes.PoetryParserFake
		roller   *servicefakes.RollWriterFake
		idx      *servicefakes.SearchIndexWriterFake
		uc       *usecase.SyncRolls
		ctx      context.Context
	)

	BeforeEach(func() {
		lister = &servicefakes.FileListerFake{ListFilesResults: map[string]servicefakes.ListFilesResult{}}
		dl = &servicefakes.FileDownloaderFake{
			DefaultDownload: servicefakes.DownloadResult{Data: []byte{0xff, 0xd8}},
		}
		pub = &servicefakes.SharePublisherFake{}
		analyzer = &servicefakes.ImageAnalyzerFake{
			DefaultResult: servicefakes.AnalyzeResult{
				Analysis: domain.ImageAnalysis{
					Palette:       []string{"#111"},
					DominantColor: "#111",
				},
			},
		}
		aggr = &servicefakes.PaletteAggregatorFake{DefaultPalette: []string{"#111", "#222", "#333"}}
		poet = &servicefakes.PoetryParserFake{}
		roller = &servicefakes.RollWriterFake{}
		idx = &servicefakes.SearchIndexWriterFake{}

		uc = usecase.NewSyncRolls(usecase.SyncRollsDeps{
			Logger: silentLogger(), Lister: lister, Downloader: dl, Publisher: pub,
			Analyzer: analyzer, Aggregator: aggr, Poetry: poet,
			RollWriter: roller, IndexWriter: idx,
			Concurrency: 2, PaletteSize: 5,
		})
		ctx = context.Background()
	})

	// -------------------------------------------------------------------------
	// Scenario 1 — happy path: one roll with three images.
	// -------------------------------------------------------------------------
	Context("given a root folder with one roll that contains three images", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Nuit à Tokyo", time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{file("i1", "a.jpg"), file("i2", "b.jpg"), file("i3", "c.jpg")},
			}
		})

		It("writes one markdown file and one index entry", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(roller.WrittenSlugs()).To(ConsistOf("nuit-a-tokyo"))
			Expect(idx.LastIndex()).To(HaveLen(1))
			Expect(idx.LastIndex()[0].ID).To(Equal("nuit-a-tokyo"))
			Expect(idx.LastIndex()[0].Title).To(Equal("Nuit à Tokyo"))
			Expect(idx.LastIndex()[0].Date).To(Equal("2024-03-15"))
		})

		It("preserves the drive listing order in the output images", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.Images).To(HaveLen(3))
			Expect(roll.Images[0].URL).To(Equal("https://share/i1"))
			Expect(roll.Images[1].URL).To(Equal("https://share/i2"))
			Expect(roll.Images[2].URL).To(Equal("https://share/i3"))
		})

		It("uses the first image as the index cover", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(idx.LastIndex()[0].Cover).To(Equal("https://share/i1"))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 2 — two rolls.
	// -------------------------------------------------------------------------
	Context("given two rolls", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{
					dir("100", "Matin brumeux", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					dir("200", "Nuit", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
				},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
			lister.ListFilesResults["200"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("b", "y.jpg")}}
		})

		It("writes two markdown files and two index entries", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(roller.WrittenSlugs()).To(ConsistOf("matin-brumeux", "nuit"))
			Expect(idx.LastIndex()).To(HaveLen(2))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 3 — empty root folder.
	// -------------------------------------------------------------------------
	Context("given an empty root folder", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{}}
		})

		It("writes an empty index and never calls RollWriter", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(roller.WrittenSlugs()).To(BeEmpty())
			Expect(idx.CallCount()).To(Equal(1))
			Expect(idx.LastIndex()).To(BeEmpty())
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 4 — ListFiles(root) fails.
	// -------------------------------------------------------------------------
	Context("when ListFiles on the root folder fails", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{Err: errors.New("boom")}
		})

		It("returns an error wrapped with 'list root folder'", func() {
			err := uc.Execute(ctx, "root")
			Expect(err).To(MatchError(ContainSubstring("list root folder")))
			Expect(err).To(MatchError(ContainSubstring("boom")))
			Expect(idx.CallCount()).To(Equal(0))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 5 — ListFiles(rollID) fails, other rolls still processed.
	// -------------------------------------------------------------------------
	Context("when ListFiles on a roll fails", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{
					dir("100", "Bad", time.Now()),
					dir("200", "Good", time.Now()),
				},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Err: errors.New("nope")}
			lister.ListFilesResults["200"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
		})

		It("skips the failing roll and processes the others", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(roller.WrittenSlugs()).To(ConsistOf("good"))
			Expect(idx.LastIndex()).To(HaveLen(1))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 6 — roll with no images.
	// -------------------------------------------------------------------------
	Context("given a roll with no images", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Vide", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{file("m", "only.md")},
			}
		})

		It("skips the roll", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(roller.WrittenSlugs()).To(BeEmpty())
			Expect(idx.LastIndex()).To(BeEmpty())
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 7 — image download fails.
	// -------------------------------------------------------------------------
	Context("when downloading an image fails", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{file("i1", "a.jpg"), file("i2", "b.jpg"), file("i3", "c.jpg")},
			}
			dl.DownloadResults = map[string]servicefakes.DownloadResult{
				"i1": {Data: []byte{0xff}},
				"i2": {Err: errors.New("download failed")},
				"i3": {Data: []byte{0xff}},
			}
		})

		It("skips the failing image and keeps the others in order", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.Images).To(HaveLen(2))
			Expect(roll.Images[0].URL).To(Equal("https://share/i1"))
			Expect(roll.Images[1].URL).To(Equal("https://share/i3"))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 8 — analyzer returns a non-nil error.
	// -------------------------------------------------------------------------
	Context("when image analysis fails but returns a partial result", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{file("i1", "a.jpg")},
			}
			analyzer.AnalyzeStub = func(_ []byte) (domain.ImageAnalysis, error) {
				return domain.ImageAnalysis{}, errors.New("analyze failure")
			}
		})

		It("still writes the image with a nil Exif", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.Images).To(HaveLen(1))
			Expect(roll.Images[0].Exif).To(BeNil())
			Expect(roll.Images[0].Alt).To(ContainSubstring("Photographie du roll Roll"))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 9 — PublishShare fails for an image.
	// -------------------------------------------------------------------------
	Context("when PublishShare fails for an image", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{file("i1", "a.jpg"), file("i2", "b.jpg")},
			}
			pub.ShareResults = map[string]servicefakes.ShareResult{
				"i1": {URL: "https://share/i1"},
				"i2": {Err: errors.New("share ko")},
			}
		})

		It("skips the affected image", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.Images).To(HaveLen(1))
			Expect(roll.Images[0].URL).To(Equal("https://share/i1"))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 10 — PublishShare fails for audio/video.
	// -------------------------------------------------------------------------
	Context("when PublishShare fails for the roll's audio and video", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{
					file("i1", "a.jpg"),
					file("au", "ambiance.mp3"),
					file("vi", "clip.mp4"),
				},
			}
			pub.ShareResults = map[string]servicefakes.ShareResult{
				"i1": {URL: "https://share/i1"},
				"au": {Err: errors.New("audio ko")},
				"vi": {Err: errors.New("video ko")},
			}
		})

		It("writes the roll with empty AudioURL and VideoURL", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.AudioURL).To(BeEmpty())
			Expect(roll.VideoURL).To(BeEmpty())
			Expect(roll.Images).To(HaveLen(1))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 11 — poem.md is present.
	// -------------------------------------------------------------------------
	Context("given a roll that includes a poem.md file", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{file("i1", "a.jpg"), file("p", "poem.md")},
			}
			poet.ParseStub = func(_ []byte) (domain.Poetry, error) {
				return domain.Poetry{
					GlobalPoem: "un souffle",
					PhotoPoems: map[string]string{"a.jpg": "verse dédié"},
				}, nil
			}
		})

		It("wires GlobalPoem onto the roll and PhotoPoems onto each matching image", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.Poem).To(Equal("un souffle"))
			Expect(roll.Images[0].Poem).To(Equal("verse dédié"))
			Expect(idx.LastIndex()[0].Poem).To(Equal("un souffle"))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 12 — poetry parsing fails.
	// -------------------------------------------------------------------------
	Context("when poetry parsing fails", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{file("i1", "a.jpg"), file("p", "poem.md")},
			}
			poet.ParseStub = func(_ []byte) (domain.Poetry, error) {
				return domain.Poetry{}, errors.New("bad yaml")
			}
		})

		It("writes the roll without any poem", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.Poem).To(BeEmpty())
			Expect(roll.Images[0].Poem).To(BeEmpty())
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 13 — ctx cancelled between two rolls.
	// -------------------------------------------------------------------------
	Context("when the context is cancelled between the first and second roll", func() {
		It("returns ctx.Err() and never calls WriteIndex", func() {
			cctx, cancel := context.WithCancel(context.Background())
			lister.ListFilesStub = func(_ context.Context, id string) ([]domain.DriveFile, error) {
				switch id {
				case "root":
					return []domain.DriveFile{
						dir("100", "First", time.Now()),
						dir("200", "Second", time.Now()),
					}, nil
				case "100":
					return []domain.DriveFile{file("a", "x.jpg")}, nil
				case "200":
					return []domain.DriveFile{file("b", "y.jpg")}, nil
				}
				return nil, nil
			}
			// Cancel after we start processing the first roll.
			pub.PublishShareStub = func(_ context.Context, id string) (string, error) {
				cancel()
				return "https://share/" + id, nil
			}

			err := uc.Execute(cctx, "root")
			Expect(err).To(MatchError(context.Canceled))
			Expect(idx.CallCount()).To(Equal(0))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 14 — WriteRoll fails.
	// -------------------------------------------------------------------------
	Context("when WriteRoll fails for one roll", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{
					dir("100", "KO Roll", time.Now()),
					dir("200", "OK Roll", time.Now()),
				},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
			lister.ListFilesResults["200"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("b", "y.jpg")}}
			roller.WriteResults = map[string]error{"ko-roll": errors.New("disk full")}
		})

		It("excludes the broken roll from the index and keeps going", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			// WriteRoll is still invoked (and then fails) — the slug "ko-roll" is
			// recorded in Written but not in the index.
			Expect(roller.WrittenSlugs()).To(ContainElement("ko-roll"))
			slugs := []string{}
			for _, e := range idx.LastIndex() {
				slugs = append(slugs, e.ID)
			}
			Expect(slugs).To(ConsistOf("ok-roll"))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 15 — WriteIndex fails.
	// -------------------------------------------------------------------------
	Context("when WriteIndex fails", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
			idx.WriteErr = errors.New("fs ro")
		})

		It("returns an error wrapped with 'write search index'", func() {
			err := uc.Execute(ctx, "root")
			Expect(err).To(MatchError(ContainSubstring("write search index")))
			Expect(err).To(MatchError(ContainSubstring("fs ro")))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 16 — Concurrency <= 0 falls back to 4 (no deadlock).
	// -------------------------------------------------------------------------
	Context("given a Concurrency value of 0 (fallback 4)", func() {
		It("processes every image without deadlocking", func() {
			fallbackUC := usecase.NewSyncRolls(usecase.SyncRollsDeps{
				Logger: silentLogger(), Lister: lister, Downloader: dl, Publisher: pub,
				Analyzer: analyzer, Aggregator: aggr, Poetry: poet,
				RollWriter: roller, IndexWriter: idx,
				Concurrency: 0, PaletteSize: 5,
			})
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			imgs := make([]domain.DriveFile, 8)
			for i := range imgs {
				imgs[i] = file("i"+string(rune('a'+i)), "img.jpg")
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: imgs}

			Expect(fallbackUC.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.Images).To(HaveLen(8))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 17 — tags are deduplicated and sorted.
	// -------------------------------------------------------------------------
	Context("when several images share tags", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{file("i1", "a.jpg"), file("i2", "b.jpg")},
			}
			analyzer.AnalyzeStub = func(_ []byte) (domain.ImageAnalysis, error) {
				return domain.ImageAnalysis{
					Palette: []string{"#111"},
					Tags:    []string{"zebra", "alpha", "alpha"},
				}, nil
			}
		})

		It("produces a sorted, deduplicated tag list at the roll level", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.Tags).To(Equal([]string{"alpha", "zebra"}))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 18 — DominantColor mirrors rollPalette[0], empty when palette is empty.
	// -------------------------------------------------------------------------
	Context("when the aggregator returns an empty palette", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
			aggr.DefaultPalette = []string{}
		})

		It("leaves DominantColor empty", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			roll := roller.LastRoll()
			Expect(roll.DominantColor).To(BeEmpty())
		})
	})

	Context("when the aggregator returns a non-empty palette", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
			aggr.DefaultPalette = []string{"#abcdef", "#123456"}
		})

		It("uses the first color as the dominant color", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(roller.LastRoll().DominantColor).To(Equal("#abcdef"))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 19 — Exif IsZero() maps to img.Exif == nil.
	// -------------------------------------------------------------------------
	Context("given an analysis without EXIF data", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
			analyzer.DefaultResult = servicefakes.AnalyzeResult{
				Analysis: domain.ImageAnalysis{Palette: []string{"#111"}},
			}
		})

		It("leaves Image.Exif nil", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(roller.LastRoll().Images[0].Exif).To(BeNil())
		})
	})

	Context("given an analysis with non-empty EXIF", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{dir("100", "Roll", time.Now())},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
			analyzer.DefaultResult = servicefakes.AnalyzeResult{
				Analysis: domain.ImageAnalysis{
					Palette: []string{"#111"},
					Exif:    domain.ExifData{Body: "Leica M11"},
				},
			}
		})

		It("populates Image.Exif", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			img := roller.LastRoll().Images[0]
			Expect(img.Exif).NotTo(BeNil())
			Expect(img.Exif.Body).To(Equal("Leica M11"))
		})
	})

	// -------------------------------------------------------------------------
	// Scenario 20 — root-level files are ignored.
	// -------------------------------------------------------------------------
	Context("given a root folder that mixes files and roll directories", func() {
		BeforeEach(func() {
			lister.ListFilesResults["root"] = servicefakes.ListFilesResult{
				Files: []domain.DriveFile{
					file("f1", "readme.md"),
					dir("100", "Vrai roll", time.Now()),
					file("f2", "logo.jpg"),
				},
			}
			lister.ListFilesResults["100"] = servicefakes.ListFilesResult{Files: []domain.DriveFile{file("a", "x.jpg")}}
		})

		It("ignores root-level files and only processes directories", func() {
			Expect(uc.Execute(ctx, "root")).To(Succeed())
			Expect(roller.WrittenSlugs()).To(ConsistOf("vrai-roll"))
			// Only "root" and "100" should be listed.
			Expect(lister.Calls()).To(ConsistOf("root", "100"))
		})
	})
})
