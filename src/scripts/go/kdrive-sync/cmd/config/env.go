// Package config loads the kdrive-sync runtime configuration from
// environment variables.
package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

// Env is the typed view of the process environment used by main.
//
// Tags follow the go-envconfig syntax. Defaults mirror the behaviour of the
// legacy fetch-kdrive.ts script.
type Env struct {
	DriveID       string `env:"KDRIVE_DRIVE_ID, required"`
	FolderID      string `env:"KDRIVE_FOLDER_ID, required"`
	APIToken      string `env:"KDRIVE_API_TOKEN, required"`
	OutDir        string `env:"KDRIVE_OUT_DIR, default=src/content/rolls/synced"`
	IndexFile     string `env:"KDRIVE_INDEX_FILE, default=public/search-index.json"`
	Concurrency   int    `env:"KDRIVE_CONCURRENCY, default=4"`
	PaletteSize   int    `env:"KDRIVE_PALETTE_SIZE, default=5"`
	HTTPTimeout   int    `env:"KDRIVE_HTTP_TIMEOUT, default=60"`
	KDriveBaseURL string `env:"KDRIVE_BASE_URL, default="`
}

// Load materialises an Env from the current process environment. Missing
// required variables produce a wrapped error.
func Load(ctx context.Context) (Env, error) {
	var e Env
	if err := envconfig.Process(ctx, &e); err != nil {
		return Env{}, fmt.Errorf("load config: %w", err)
	}
	return e, nil
}
