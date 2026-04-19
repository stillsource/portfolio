// Package extractpalette implements the "extract-palette" subcommand.
// It downloads an image from a public URL and prints its CIELAB k-means
// palette as JSON, ready to paste into a roll markdown frontmatter.
package extractpalette

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kdrive-sync/pkg/infrastructure/imageanalyzer"
	"net/http"
	"os"
	"time"
)

// Result is the JSON output shape.
type Result struct {
	Palette       []string `json:"palette"`
	DominantColor string   `json:"dominantColor"`
}

// Run downloads the image at url, extracts a 5-color CIELAB palette and
// writes the result as JSON to stdout.
func Run(ctx context.Context, url string) error {
	if url == "" {
		return fmt.Errorf("usage: extract-palette <image-url>")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	// Some hosts require a User-Agent to serve the image.
	req.Header.Set("User-Agent", "kdrive-sync/palette-extractor")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch image: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	analysis, err := imageanalyzer.NewExifKMeans().Analyze(data)
	if err != nil {
		return fmt.Errorf("analyze image: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(Result{
		Palette:       analysis.Palette,
		DominantColor: analysis.DominantColor,
	})
}
