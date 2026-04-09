// Package usecase contains the orchestration layer. A Usecase composes
// service ports to deliver a single business capability.
package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"kdrive-sync/pkg/domain"
	"kdrive-sync/pkg/service"
)

// SyncRolls mirrors the fetch-kdrive.ts pipeline: list rolls, enrich each
// one with metadata, palette and poetry, then persist markdown files and the
// global search index.
type SyncRolls struct {
	logger      *slog.Logger
	lister      service.FileLister
	downloader  service.FileDownloader
	publisher   service.SharePublisher
	analyzer    service.ImageAnalyzer
	aggregator  service.PaletteAggregator
	poetry      service.PoetryParser
	rollWriter  service.RollWriter
	indexWriter service.SearchIndexWriter
	concurrency int
	paletteSize int
}

// SyncRollsDeps groups every collaborator required by SyncRolls.
type SyncRollsDeps struct {
	Logger      *slog.Logger
	Lister      service.FileLister
	Downloader  service.FileDownloader
	Publisher   service.SharePublisher
	Analyzer    service.ImageAnalyzer
	Aggregator  service.PaletteAggregator
	Poetry      service.PoetryParser
	RollWriter  service.RollWriter
	IndexWriter service.SearchIndexWriter
	Concurrency int
	PaletteSize int
}

// NewSyncRolls assembles a SyncRolls usecase from its dependencies.
func NewSyncRolls(deps SyncRollsDeps) *SyncRolls {
	if deps.Concurrency <= 0 {
		deps.Concurrency = 4
	}
	if deps.PaletteSize <= 0 {
		deps.PaletteSize = 5
	}
	return &SyncRolls{
		logger:      deps.Logger,
		lister:      deps.Lister,
		downloader:  deps.Downloader,
		publisher:   deps.Publisher,
		analyzer:    deps.Analyzer,
		aggregator:  deps.Aggregator,
		poetry:      deps.Poetry,
		rollWriter:  deps.RollWriter,
		indexWriter: deps.IndexWriter,
		concurrency: deps.Concurrency,
		paletteSize: deps.PaletteSize,
	}
}

// Execute runs the synchronization for every Roll found under rootFolderID.
//
// The function is fail-soft at the roll level: a single broken roll logs
// a warning and does not abort the whole sync. An error is only returned
// when the drive listing itself fails or the search-index write fails.
func (uc *SyncRolls) Execute(ctx context.Context, rootFolderID string) error {
	uc.logger.Info("sync starting", slog.String("root", rootFolderID))

	entries, err := uc.lister.ListFiles(ctx, rootFolderID)
	if err != nil {
		return fmt.Errorf("list root folder: %w", err)
	}

	folders := filterDirs(entries)
	uc.logger.Info("rolls discovered", slog.Int("count", len(folders)))

	index := make([]domain.SearchIndexItem, 0, len(folders))
	for _, folder := range folders {
		if err := ctx.Err(); err != nil {
			return err
		}

		item, ok := uc.processRoll(ctx, folder)
		if !ok {
			continue
		}
		index = append(index, item)
	}

	if err := uc.indexWriter.WriteIndex(index); err != nil {
		return fmt.Errorf("write search index: %w", err)
	}
	uc.logger.Info("sync complete",
		slog.Int("rolls", len(index)),
	)
	return nil
}

// processRoll fetches one folder and produces the corresponding search-index
// entry. It returns false when the roll was skipped.
func (uc *SyncRolls) processRoll(ctx context.Context, folder domain.DriveFile) (domain.SearchIndexItem, bool) {
	rollLogger := uc.logger.With(slog.String("roll", folder.Name))
	rollLogger.Info("processing roll")

	files, err := uc.lister.ListFiles(ctx, folder.ID)
	if err != nil {
		rollLogger.Warn("list roll folder failed", slog.String("err", err.Error()))
		return domain.SearchIndexItem{}, false
	}

	classified := classifyFiles(files)
	if len(classified.images) == 0 {
		rollLogger.Warn("no images found, skipping")
		return domain.SearchIndexItem{}, false
	}

	poems := uc.loadPoetry(ctx, classified.poetry, rollLogger)
	audioURL := uc.loadAudio(ctx, classified.audio, rollLogger)

	images, palettes, tags := uc.analyzeImages(ctx, folder.Name, classified.images, poems.PhotoPoems, rollLogger)
	if len(images) == 0 {
		rollLogger.Warn("no images analyzed, skipping")
		return domain.SearchIndexItem{}, false
	}

	rollPalette := uc.aggregator.Aggregate(palettes, uc.paletteSize)
	dominant := ""
	if len(rollPalette) > 0 {
		dominant = rollPalette[0]
	}

	slug := domain.Slugify(folder.Name)
	date := folder.CreatedAt.Format("2006-01-02")

	roll := &domain.Roll{
		Title:         folder.Name,
		Date:          date,
		Tags:          tags,
		Poem:          poems.GlobalPoem,
		Palette:       rollPalette,
		DominantColor: dominant,
		AudioURL:      audioURL,
		Images:        images,
	}

	if err := uc.rollWriter.WriteRoll(slug, roll); err != nil {
		rollLogger.Error("write roll failed", slog.String("err", err.Error()))
		return domain.SearchIndexItem{}, false
	}
	rollLogger.Info("roll written", slog.String("slug", slug), slog.Int("images", len(images)))

	cover := ""
	if len(images) > 0 {
		cover = images[0].URL
	}
	return domain.SearchIndexItem{
		ID:      slug,
		Title:   folder.Name,
		Date:    date,
		Tags:    roll.Tags,
		Poem:    poems.GlobalPoem,
		Cover:   cover,
		Palette: rollPalette,
	}, true
}

// -----------------------------------------------------------------------------
// Poetry / audio loading
// -----------------------------------------------------------------------------

func (uc *SyncRolls) loadPoetry(ctx context.Context, file *domain.DriveFile, logger *slog.Logger) domain.Poetry {
	empty := domain.Poetry{PhotoPoems: map[string]string{}}
	if file == nil {
		return empty
	}
	logger.Info("poetry found", slog.String("file", file.Name))

	data, err := uc.downloader.DownloadFile(ctx, file.ID)
	if err != nil {
		logger.Warn("download poetry failed", slog.String("err", err.Error()))
		return empty
	}
	poems, err := uc.poetry.Parse(data)
	if err != nil {
		logger.Warn("parse poetry failed", slog.String("err", err.Error()))
		return empty
	}
	if poems.PhotoPoems == nil {
		poems.PhotoPoems = map[string]string{}
	}
	return poems
}

func (uc *SyncRolls) loadAudio(ctx context.Context, file *domain.DriveFile, logger *slog.Logger) string {
	if file == nil {
		return ""
	}
	logger.Info("audio found", slog.String("file", file.Name))

	url, err := uc.publisher.PublishShare(ctx, file.ID)
	if err != nil {
		logger.Warn("publish audio failed", slog.String("err", err.Error()))
		return ""
	}
	return url
}

// -----------------------------------------------------------------------------
// Concurrent image processing
// -----------------------------------------------------------------------------

type imageResult struct {
	order   int
	image   domain.Image
	palette []string
	tags    []string
}

func (uc *SyncRolls) analyzeImages(
	ctx context.Context,
	rollName string,
	photos []domain.DriveFile,
	photoPoems map[string]string,
	logger *slog.Logger,
) ([]domain.Image, [][]string, []string) {
	results := make([]imageResult, 0, len(photos))
	var mu sync.Mutex

	sem := make(chan struct{}, uc.concurrency)
	var wg sync.WaitGroup

	for i, photo := range photos {
		if err := ctx.Err(); err != nil {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(order int, file domain.DriveFile) {
			defer wg.Done()
			defer func() { <-sem }()

			img, palette, tags, ok := uc.processImage(ctx, rollName, file, photoPoems, logger)
			if !ok {
				return
			}
			mu.Lock()
			results = append(results, imageResult{order: order, image: img, palette: palette, tags: tags})
			mu.Unlock()
		}(i, photo)
	}
	wg.Wait()

	// Restore deterministic order to match the drive listing.
	sortByOrder(results)

	images := make([]domain.Image, 0, len(results))
	palettes := make([][]string, 0, len(results))
	tagSet := make(map[string]struct{})
	for _, r := range results {
		images = append(images, r.image)
		if len(r.palette) > 0 {
			palettes = append(palettes, r.palette)
		}
		for _, t := range r.tags {
			tagSet[t] = struct{}{}
		}
	}
	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	return images, palettes, tags
}

func (uc *SyncRolls) processImage(
	ctx context.Context,
	rollName string,
	file domain.DriveFile,
	photoPoems map[string]string,
	logger *slog.Logger,
) (domain.Image, []string, []string, bool) {
	imgLogger := logger.With(slog.String("photo", file.Name))

	url, err := uc.publisher.PublishShare(ctx, file.ID)
	if err != nil {
		imgLogger.Warn("publish share failed", slog.String("err", err.Error()))
		return domain.Image{}, nil, nil, false
	}

	data, err := uc.downloader.DownloadFile(ctx, file.ID)
	if err != nil {
		imgLogger.Warn("download image failed", slog.String("err", err.Error()))
		return domain.Image{}, nil, nil, false
	}

	analysis, err := uc.analyzer.Analyze(data)
	if err != nil {
		imgLogger.Warn("analyze image failed", slog.String("err", err.Error()))
	}

	img := domain.Image{
		URL:           url,
		Alt:           buildAltText(rollName, analysis.Exif),
		Poem:          photoPoems[file.Name],
		Palette:       analysis.Palette,
		DominantColor: analysis.DominantColor,
	}
	if !analysis.Exif.IsZero() {
		exif := analysis.Exif
		img.Exif = &exif
	}
	return img, analysis.Palette, analysis.Tags, true
}

// -----------------------------------------------------------------------------
// File classification helpers
// -----------------------------------------------------------------------------

type classifiedFiles struct {
	images []domain.DriveFile
	poetry *domain.DriveFile
	audio  *domain.DriveFile
}

func classifyFiles(files []domain.DriveFile) classifiedFiles {
	var out classifiedFiles
	for _, f := range files {
		if f.Type != domain.DriveFileTypeFile {
			continue
		}
		switch strings.ToLower(filepath.Ext(f.Name)) {
		case ".jpg", ".jpeg":
			out.images = append(out.images, f)
		case ".md":
			if out.poetry == nil {
				file := f
				out.poetry = &file
			}
		case ".mp3":
			if out.audio == nil {
				file := f
				out.audio = &file
			}
		}
	}
	return out
}

func filterDirs(entries []domain.DriveFile) []domain.DriveFile {
	dirs := make([]domain.DriveFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		}
	}
	return dirs
}

func buildAltText(rollName string, exif domain.ExifData) string {
	if exif.Body == "" {
		return fmt.Sprintf("Photographie du roll %s", rollName)
	}
	parts := []string{exif.Body}
	if exif.FocalLength != "" {
		parts = append(parts, exif.FocalLength)
	}
	if exif.Aperture != "" {
		parts = append(parts, exif.Aperture)
	}
	if exif.Shutter != "" {
		parts = append(parts, exif.Shutter)
	}
	return "Photographie prise avec " + strings.Join(parts, " • ")
}

func sortByOrder(results []imageResult) {
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].order < results[j].order
	})
}
