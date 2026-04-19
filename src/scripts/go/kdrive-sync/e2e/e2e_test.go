// Package e2e contains BDD end-to-end tests for the portfolio frontend.
//
// Uses godog (Cucumber/Gherkin) + go-rod (Chrome automation).
//
// Run: go test -C . ./e2e/... -v
//
// Requires: npm run dev already running on :4321, OR set E2E_BASE_URL env var.
// For CI, set E2E_START_DEV_SERVER=1 to auto-start the dev server.
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const defaultBaseURL = "http://localhost:4321"

var (
	baseURL   string
	devServer *exec.Cmd
	browser   *rod.Browser
)

func TestMain(m *testing.M) {
	baseURL = os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// Optionally start the dev server
	if os.Getenv("E2E_START_DEV_SERVER") == "1" {
		devServer = exec.Command("npm", "run", "dev")
		devServer.Dir = "../../../../.." // repo root from kdrive-sync dir
		devServer.Stdout = os.Stdout
		devServer.Stderr = os.Stderr
		if err := devServer.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: start dev server: %v\n", err)
			os.Exit(1)
		}
		if err := waitForHTTP(baseURL, 30*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: dev server not ready: %v\n", err)
			devServer.Process.Kill()
			os.Exit(1)
		}
	}

	// Launch headless browser
	u := launcher.New().Headless(true).MustLaunch()
	browser = rod.New().ControlURL(u).MustConnect()

	code := m.Run()

	browser.MustClose()
	if devServer != nil {
		devServer.Process.Kill()
	}

	os.Exit(code)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run BDD tests")
	}
}

// InitializeScenario wires step definitions for all features.
func InitializeScenario(ctx *godog.ScenarioContext) {
	s := newState()

	// Browser lifecycle: fresh page per scenario
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		s.page = browser.MustPage("")
		s.page.MustSetViewport(1440, 900, 1, false)
		return c, nil
	})
	ctx.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		s.page.MustClose()
		return c, nil
	})

	registerNavigationSteps(ctx, s)
	registerTimerSteps(ctx, s)
	registerCursorSteps(ctx, s)
	registerPoemSteps(ctx, s)
	registerAuraSteps(ctx, s)
	registerLayoutSteps(ctx, s)
}

// waitForHTTP polls url until it responds 200 or timeout.
func waitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url) //nolint:noctx
		if err == nil && resp.StatusCode < 500 {
			resp.Body.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", url)
}
