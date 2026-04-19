// Command kdrive-sync mirrors a kDrive folder tree into Astro-compatible
// markdown files. It is a drop-in Go replacement for the legacy
// fetch-kdrive.ts script.
//
// Subcommands:
//
//	(default)         sync kDrive → markdown files
//	extract-palette   download an image URL and print its CIELAB palette as JSON
package main

import (
	"context"
	"fmt"
	"kdrive-sync/cmd/config"
	"kdrive-sync/cmd/extractpalette"
	"kdrive-sync/pkg/infrastructure/di"
	"os"
	"os/signal"
	"syscall"
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

	if len(os.Args) > 1 && os.Args[1] == "extract-palette" {
		url := ""
		if len(os.Args) > 2 {
			url = os.Args[2]
		}
		return extractpalette.Run(ctx, url)
	}

	env, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
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

	if err := container.GetSyncRolls().Execute(ctx, env.FolderID); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	return nil
}
