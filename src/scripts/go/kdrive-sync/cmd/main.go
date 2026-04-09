// Command kdrive-sync mirrors a kDrive folder tree into Astro-compatible
// markdown files. It is a drop-in Go replacement for the legacy
// fetch-kdrive.ts script.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"kdrive-sync/cmd/config"
	"kdrive-sync/pkg/infrastructure/di"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "kdrive-sync: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	env, err := config.Load(ctx)
	if err != nil {
		return err
	}

	container := di.NewContainer(di.Config{
		DriveID:       env.DriveID,
		APIToken:      env.APIToken,
		OutDir:        env.OutDir,
		IndexFile:     env.IndexFile,
		Concurrency:   env.Concurrency,
		PaletteSize:   env.PaletteSize,
		HTTPTimeout:   env.HTTPTimeout,
		KDriveBaseURL: env.KDriveBaseURL,
	})

	return container.GetSyncRolls().Execute(ctx, env.FolderID)
}
