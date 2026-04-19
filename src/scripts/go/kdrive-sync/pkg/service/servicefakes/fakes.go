// Package servicefakes provides thread-safe test doubles for every interface
// declared in kdrive-sync/pkg/service.
//
// Each fake exposes three interchangeable ways to script behavior:
//
//  1. A Stub function that, if set, handles every call.
//  2. A Results map keyed by the first argument (usually an ID).
//  3. A default zero-value return.
//
// Every call is recorded in Calls (a mutex-guarded slice) so assertions can
// verify what was requested and in what order.
package servicefakes

import (
	"context"
	"kdrive-sync/pkg/domain"
	"sync"
)

// ---------------------------------------------------------------------------
// FileLister
// ---------------------------------------------------------------------------

// ListFilesResult is a scripted result for FileListerFake.
type ListFilesResult struct {
	Files []domain.DriveFile
	Err   error
}

// FileListerFake fakes service.FileLister.
type FileListerFake struct {
	mu               sync.Mutex
	calls            []string
	ListFilesStub    func(ctx context.Context, folderID string) ([]domain.DriveFile, error)
	ListFilesResults map[string]ListFilesResult
}

// ListFiles records the folderID and returns the scripted result.
func (f *FileListerFake) ListFiles(ctx context.Context, folderID string) ([]domain.DriveFile, error) {
	f.mu.Lock()
	f.calls = append(f.calls, folderID)
	f.mu.Unlock()
	if f.ListFilesStub != nil {
		return f.ListFilesStub(ctx, folderID)
	}
	if f.ListFilesResults == nil {
		return nil, nil
	}
	r := f.ListFilesResults[folderID]
	return r.Files, r.Err
}

// Calls returns a snapshot of the folderIDs requested so far.
func (f *FileListerFake) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	copy(out, f.calls)
	return out
}

// CallCount returns the number of times ListFiles was invoked.
func (f *FileListerFake) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

// ---------------------------------------------------------------------------
// FileDownloader
// ---------------------------------------------------------------------------

// DownloadResult is a scripted result for FileDownloaderFake.
type DownloadResult struct {
	Data []byte
	Err  error
}

// FileDownloaderFake fakes service.FileDownloader.
type FileDownloaderFake struct {
	mu               sync.Mutex
	calls            []string
	DownloadFileStub func(ctx context.Context, fileID string) ([]byte, error)
	DownloadResults  map[string]DownloadResult
	DefaultDownload  DownloadResult
}

// DownloadFile records the fileID and returns the scripted result.
func (f *FileDownloaderFake) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	f.mu.Lock()
	f.calls = append(f.calls, fileID)
	f.mu.Unlock()
	if f.DownloadFileStub != nil {
		return f.DownloadFileStub(ctx, fileID)
	}
	if r, ok := f.DownloadResults[fileID]; ok {
		return r.Data, r.Err
	}
	return f.DefaultDownload.Data, f.DefaultDownload.Err
}

// Calls returns the fileIDs requested so far.
func (f *FileDownloaderFake) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	copy(out, f.calls)
	return out
}

// ---------------------------------------------------------------------------
// SharePublisher
// ---------------------------------------------------------------------------

// ShareResult is a scripted result for SharePublisherFake.
type ShareResult struct {
	URL string
	Err error
}

// SharePublisherFake fakes service.SharePublisher.
type SharePublisherFake struct {
	mu               sync.Mutex
	calls            []string
	PublishShareStub func(ctx context.Context, fileID string) (string, error)
	ShareResults     map[string]ShareResult
	DefaultShare     ShareResult
}

// PublishShare records the fileID and returns the scripted URL.
func (f *SharePublisherFake) PublishShare(ctx context.Context, fileID string) (string, error) {
	f.mu.Lock()
	f.calls = append(f.calls, fileID)
	f.mu.Unlock()
	if f.PublishShareStub != nil {
		return f.PublishShareStub(ctx, fileID)
	}
	if r, ok := f.ShareResults[fileID]; ok {
		return r.URL, r.Err
	}
	if f.DefaultShare.URL == "" && f.DefaultShare.Err == nil {
		return "https://share/" + fileID, nil
	}
	return f.DefaultShare.URL, f.DefaultShare.Err
}

// Calls returns the fileIDs requested so far.
func (f *SharePublisherFake) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	copy(out, f.calls)
	return out
}

// ---------------------------------------------------------------------------
// ImageAnalyzer
// ---------------------------------------------------------------------------

// AnalyzeResult is a scripted result for ImageAnalyzerFake.
type AnalyzeResult struct {
	Analysis domain.ImageAnalysis
	Err      error
}

// ImageAnalyzerFake fakes service.ImageAnalyzer.
type ImageAnalyzerFake struct {
	mu            sync.Mutex
	callCount     int
	AnalyzeStub   func(data []byte) (domain.ImageAnalysis, error)
	DefaultResult AnalyzeResult
}

// Analyze returns the scripted analysis result.
func (f *ImageAnalyzerFake) Analyze(data []byte) (domain.ImageAnalysis, error) {
	f.mu.Lock()
	f.callCount++
	f.mu.Unlock()
	if f.AnalyzeStub != nil {
		return f.AnalyzeStub(data)
	}
	return f.DefaultResult.Analysis, f.DefaultResult.Err
}

// CallCount returns the number of times Analyze was invoked.
func (f *ImageAnalyzerFake) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount
}

// ---------------------------------------------------------------------------
// PaletteAggregator
// ---------------------------------------------------------------------------

// PaletteAggregatorFake fakes service.PaletteAggregator.
type PaletteAggregatorFake struct {
	mu             sync.Mutex
	LastPalettes   [][]string
	LastSize       int
	AggregateStub  func(palettes [][]string, size int) []string
	DefaultPalette []string
}

// Aggregate records the inputs and returns the scripted palette.
func (f *PaletteAggregatorFake) Aggregate(palettes [][]string, size int) []string {
	f.mu.Lock()
	f.LastPalettes = palettes
	f.LastSize = size
	f.mu.Unlock()
	if f.AggregateStub != nil {
		return f.AggregateStub(palettes, size)
	}
	return f.DefaultPalette
}

// ---------------------------------------------------------------------------
// PoetryParser
// ---------------------------------------------------------------------------

// PoetryResult is a scripted result for PoetryParserFake.
type PoetryResult struct {
	Poetry domain.Poetry
	Err    error
}

// PoetryParserFake fakes service.PoetryParser.
type PoetryParserFake struct {
	mu            sync.Mutex
	callCount     int
	ParseStub     func(data []byte) (domain.Poetry, error)
	DefaultResult PoetryResult
}

// Parse returns the scripted poetry result.
func (f *PoetryParserFake) Parse(data []byte) (domain.Poetry, error) {
	f.mu.Lock()
	f.callCount++
	f.mu.Unlock()
	if f.ParseStub != nil {
		return f.ParseStub(data)
	}
	return f.DefaultResult.Poetry, f.DefaultResult.Err
}

// CallCount returns how many times Parse was invoked.
func (f *PoetryParserFake) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount
}

// ---------------------------------------------------------------------------
// RollWriter
// ---------------------------------------------------------------------------

// WrittenRoll captures a single WriteRoll invocation.
type WrittenRoll struct {
	Slug string
	Roll *domain.Roll
}

// RollWriterFake fakes service.RollWriter.
type RollWriterFake struct {
	mu            sync.Mutex
	written       []WrittenRoll
	WriteRollStub func(slug string, roll *domain.Roll) error
	WriteResults  map[string]error
}

// WriteRoll records the invocation and returns the scripted error.
func (f *RollWriterFake) WriteRoll(slug string, roll *domain.Roll) error {
	f.mu.Lock()
	f.written = append(f.written, WrittenRoll{Slug: slug, Roll: roll})
	f.mu.Unlock()
	if f.WriteRollStub != nil {
		return f.WriteRollStub(slug, roll)
	}
	if f.WriteResults != nil {
		return f.WriteResults[slug]
	}
	return nil
}

// WrittenSlugs returns the slugs of every WriteRoll call, in order.
func (f *RollWriterFake) WrittenSlugs() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.written))
	for i, w := range f.written {
		out[i] = w.Slug
	}
	return out
}

// Written returns a snapshot of every captured WriteRoll call.
func (f *RollWriterFake) Written() []WrittenRoll {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]WrittenRoll, len(f.written))
	copy(out, f.written)
	return out
}

// LastRoll returns the Roll struct of the most recent call, or nil.
func (f *RollWriterFake) LastRoll() *domain.Roll {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.written) == 0 {
		return nil
	}
	return f.written[len(f.written)-1].Roll
}

// ---------------------------------------------------------------------------
// SearchIndexWriter
// ---------------------------------------------------------------------------

// SearchIndexWriterFake fakes service.SearchIndexWriter.
type SearchIndexWriterFake struct {
	mu             sync.Mutex
	callCount      int
	lastIndex      []domain.SearchIndexItem
	WriteIndexStub func(items []domain.SearchIndexItem) error
	WriteErr       error
}

// WriteIndex captures the items and returns the scripted error.
func (f *SearchIndexWriterFake) WriteIndex(items []domain.SearchIndexItem) error {
	f.mu.Lock()
	f.callCount++
	copyItems := make([]domain.SearchIndexItem, len(items))
	copy(copyItems, items)
	f.lastIndex = copyItems
	f.mu.Unlock()
	if f.WriteIndexStub != nil {
		return f.WriteIndexStub(items)
	}
	return f.WriteErr
}

// CallCount returns how many times WriteIndex was called.
func (f *SearchIndexWriterFake) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount
}

// LastIndex returns the items captured on the last invocation.
func (f *SearchIndexWriterFake) LastIndex() []domain.SearchIndexItem {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domain.SearchIndexItem, len(f.lastIndex))
	copy(out, f.lastIndex)
	return out
}
