package di_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kdrive-sync/pkg/infrastructure/di"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestContainer_Wiring_EndToEnd exercises every getter on the DI container
// by running a full Execute() against an in-memory httptest.Server that
// mimics the three kDrive endpoints the usecase calls: folder listing,
// download, and share creation.
//
// Asserting on the side-effects (one .md per roll, one JSON index) is enough
// to prove the graph is correctly wired — the individual adapters already
// have their own unit coverage.
func TestContainer_Wiring_EndToEnd(t *testing.T) {
	// Load the committed fixture that carries EXIF + palette-worthy pixels.
	imgBytes, err := os.ReadFile(filepath.Join("..", "imageanalyzer", "testdata", "with_exif.jpg"))
	if err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}

	const (
		driveID   = "42"
		rootID    = "root"
		folderID  = "100"
		imageID   = "200"
		poetryID  = "300"
		audioID   = "400"
		shareURL  = "https://share.example/abc"
		createdAt = int64(1_700_000_000)
	)

	// makeListEntry produces the JSON payload the listing endpoint returns.
	type entry struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		CreatedAt int64  `json:"created_at"`
	}

	handler := http.NewServeMux()

	// Root listing: one folder representing the single roll.
	handler.HandleFunc(fmt.Sprintf("/%s/files/%s/files", driveID, rootID),
		func(w http.ResponseWriter, _ *http.Request) {
			resp := map[string]any{
				"data": []entry{
					{ID: 100, Name: "Nuit à Tokyo", Type: "dir", CreatedAt: createdAt},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		},
	)

	// Roll folder listing: one image, one poetry markdown, one audio file.
	handler.HandleFunc(fmt.Sprintf("/%s/files/%s/files", driveID, folderID),
		func(w http.ResponseWriter, _ *http.Request) {
			resp := map[string]any{
				"data": []entry{
					{ID: 200, Name: "IMG_0001.jpg", Type: "file", CreatedAt: createdAt},
					{ID: 300, Name: "poem.md", Type: "file", CreatedAt: createdAt},
					{ID: 400, Name: "track.mp3", Type: "file", CreatedAt: createdAt},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		},
	)

	// Downloads: JPEG for the image, markdown for the poetry.
	handler.HandleFunc(fmt.Sprintf("/%s/files/%s/download", driveID, imageID),
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(imgBytes)
		},
	)
	handler.HandleFunc(fmt.Sprintf("/%s/files/%s/download", driveID, poetryID),
		func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w,
				"---\nphotos:\n  IMG_0001.jpg: \"Le néon tremble.\"\n---\nPoème global.")
		},
	)

	// Shares: GET returns empty list so the code path falls through to POST,
	// which returns the final share URL.
	sharesHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = io.WriteString(w, `{"data":[]}`)
		case http.MethodPost:
			resp := map[string]any{"data": map[string]any{"share_url": shareURL}}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
	handler.HandleFunc(fmt.Sprintf("/%s/files/%s/shares", driveID, imageID), sharesHandler)
	handler.HandleFunc(fmt.Sprintf("/%s/files/%s/shares", driveID, audioID), sharesHandler)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	outDir := t.TempDir()
	indexPath := filepath.Join(t.TempDir(), "search-index.json")

	cfg := di.Config{
		DriveID:       driveID,
		APIToken:      "test-token",
		OutDir:        outDir,
		IndexFile:     indexPath,
		Concurrency:   2,
		PaletteSize:   5,
		HTTPTimeout:   10,
		KDriveBaseURL: srv.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := di.NewContainer(cfg)
	uc := c.GetSyncRolls()
	if uc == nil {
		t.Fatal("GetSyncRolls returned nil")
	}

	// Calling GetSyncRolls again must return the memoised instance — this
	// also covers the short-circuit branch of every lazy getter.
	if c.GetSyncRolls() != uc {
		t.Error("GetSyncRolls did not memoise the singleton")
	}

	// GetHTTPClient and GetLogger are exported helpers. Call each twice to
	// exercise both the construction and the cached-return branches.
	if c.GetHTTPClient() == nil {
		t.Error("GetHTTPClient returned nil on first call")
	}
	if c.GetHTTPClient() == nil {
		t.Error("GetHTTPClient returned nil on cached call")
	}
	if c.GetLogger() == nil {
		t.Error("GetLogger returned nil on first call")
	}
	if c.GetLogger() == nil {
		t.Error("GetLogger returned nil on cached call")
	}

	if err := uc.Execute(ctx, rootID); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Exactly one .md file should have been produced for the single roll.
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("read outDir: %v", err)
	}
	var mds []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			mds = append(mds, e.Name())
		}
	}
	if len(mds) != 1 {
		t.Fatalf("expected 1 .md file, got %d: %v", len(mds), mds)
	}

	mdContent, err := os.ReadFile(filepath.Join(outDir, mds[0]))
	if err != nil {
		t.Fatalf("read roll md: %v", err)
	}
	// Share URL should be embedded as the image URL and the audio URL.
	if !strings.Contains(string(mdContent), shareURL) {
		t.Errorf("expected share URL %q in markdown output:\n%s", shareURL, mdContent)
	}

	// Search index file must exist and be valid JSON with one entry.
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	var index []map[string]any
	if err := json.Unmarshal(indexData, &index); err != nil {
		t.Fatalf("index is not valid JSON: %v\n%s", err, indexData)
	}
	if len(index) != 1 {
		t.Errorf("expected 1 index entry, got %d", len(index))
	}
}

// TestContainer_HTTPTimeoutDefault covers the branch in GetHTTPClient where
// HTTPTimeout <= 0 falls back to the 60-second default.
func TestContainer_HTTPTimeoutDefault(t *testing.T) {
	c := di.NewContainer(di.Config{HTTPTimeout: 0})
	client := c.GetHTTPClient()
	if client == nil {
		t.Fatal("GetHTTPClient returned nil")
	}
	if client.Timeout <= 0 {
		t.Errorf("timeout = %v, want > 0 (default fallback)", client.Timeout)
	}
}
