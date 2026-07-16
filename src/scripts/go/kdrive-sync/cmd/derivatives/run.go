// Package derivatives implements the "derivatives" subcommand: it reads JPEGs
// from a local folder and writes downsized, metadata-stripped copies to another
// local folder. Standalone tool: it does not touch the sync pipeline or kDrive.
package derivatives

import (
	"context"
	"fmt"
	"kdrive-sync/pkg/infrastructure/imageresizer"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxLongEdge = 1920
	jpegQuality = 85
)

// Run derives every .jpg/.jpeg in inputDir into outputDir. It is fail-soft:
// a bad file is logged and skipped, not fatal.
func Run(ctx context.Context, inputDir, outputDir string) error {
	if inputDir == "" || outputDir == "" {
		return fmt.Errorf("usage: derivatives <input-dir> <output-dir>")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("read input dir: %w", err)
	}

	r := imageresizer.New(maxLongEdge, jpegQuality)
	var processed, skipped int
	var bytesIn, bytesOut int64

	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return err //nolint:wrapcheck // passthrough: ctx cancellation is not our error to wrap
		}
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".jpg" && ext != ".jpeg" {
			slog.Warn("skipping non-jpeg file", slog.String("file", e.Name()))
			skipped++
			continue
		}
		data, err := os.ReadFile(filepath.Join(inputDir, e.Name()))
		if err != nil {
			slog.Warn("read failed", slog.String("file", e.Name()), slog.String("err", err.Error()))
			skipped++
			continue
		}
		out, err := r.Derive(data)
		if err != nil {
			slog.Warn("derive failed", slog.String("file", e.Name()), slog.String("err", err.Error()))
			skipped++
			continue
		}
		if err := os.WriteFile(filepath.Join(outputDir, e.Name()), out, 0o644); err != nil {
			slog.Warn("write failed", slog.String("file", e.Name()), slog.String("err", err.Error()))
			skipped++
			continue
		}
		processed++
		bytesIn += int64(len(data))
		bytesOut += int64(len(out))
	}

	slog.Info("derivatives complete",
		slog.Int("processed", processed),
		slog.Int("skipped", skipped),
		slog.Int64("bytes_in", bytesIn),
		slog.Int64("bytes_out", bytesOut),
	)
	return nil
}
