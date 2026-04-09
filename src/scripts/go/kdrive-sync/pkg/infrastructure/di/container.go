// Package di wires every concrete infrastructure component to the abstract
// services consumed by the usecase layer.
//
// Each dependency lives in its own file and is exposed through a lazy getter
// that memoizes the singleton so the container stays cheap to construct.
package di

import (
	"kdrive-sync/pkg/infrastructure/filedownloader"
	"kdrive-sync/pkg/infrastructure/filelister"
	"kdrive-sync/pkg/infrastructure/imageanalyzer"
	"kdrive-sync/pkg/infrastructure/kdriveapi"
	"kdrive-sync/pkg/infrastructure/paletteaggregator"
	"kdrive-sync/pkg/infrastructure/poetryparser"
	"kdrive-sync/pkg/infrastructure/rollwriter"
	"kdrive-sync/pkg/infrastructure/searchindexwriter"
	"kdrive-sync/pkg/infrastructure/sharepublisher"
	"kdrive-sync/pkg/usecase"
	"log/slog"
	"net/http"
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
	apiClient      *kdriveapi.Client
	fileLister     *filelister.KDrive
	fileDownloader *filedownloader.KDrive
	sharePublisher *sharepublisher.KDrive
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
