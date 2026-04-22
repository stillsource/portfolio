// Package di wires every concrete infrastructure component to the abstract
// services consumed by the usecase layer.
//
// Each dependency lives in its own file and is exposed through a lazy getter
// that memoizes the singleton so the container stays cheap to construct.
package di

import (
	"log/slog"
	"net/http"

	"github.com/stillsource/kdrive-fuse/kdrive"

	"kdrive-sync/pkg/infrastructure/imageanalyzer"
	"kdrive-sync/pkg/infrastructure/kdriveadapter"
	"kdrive-sync/pkg/infrastructure/paletteaggregator"
	"kdrive-sync/pkg/infrastructure/poetryparser"
	"kdrive-sync/pkg/infrastructure/rollwriter"
	"kdrive-sync/pkg/infrastructure/searchindexwriter"
	"kdrive-sync/pkg/usecase"
)

// Config bundles the runtime parameters needed to build a Container.
type Config struct {
	DriveID       string
	APIToken      string
	OutDir        string
	IndexFile     string
	Concurrency   int
	PaletteSize   int
	HTTPTimeout   int // seconds
	KDriveBaseURL string
}

// Container owns all singletons and exposes the entry-point usecase.
type Container struct {
	cfg Config

	logger         *slog.Logger
	httpClient     *http.Client
	kdriveClient   *kdrive.Client
	fileLister     *kdriveadapter.FileLister
	fileDownloader *kdriveadapter.FileDownloader
	sharePublisher *kdriveadapter.SharePublisher
	imageAnalyzer  *imageanalyzer.ExifKMeans
	paletteAgg     *paletteaggregator.CIELAB
	poetryParser   *poetryparser.Frontmatter
	rollWriter     *rollwriter.Markdown
	indexWriter    *searchindexwriter.JSONFile

	syncRolls *usecase.SyncRolls
}

// NewContainer returns a Container parameterised by cfg. No heavy work runs
// until a getter is called.
func NewContainer(cfg Config) *Container {
	return &Container{cfg: cfg}
}
