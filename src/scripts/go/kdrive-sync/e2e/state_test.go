package e2e

import (
	"github.com/cucumber/godog"
	"github.com/go-rod/rod"
)

// state holds per-scenario browser state shared across step definitions.
type state struct {
	page      *rod.Page
	noteColor string // noted CSS value for comparison across steps
}

func newState() *state { return &state{} }

// registerNavigationSteps wires shared navigation steps used across features.
func registerNavigationSteps(ctx *godog.ScenarioContext, s *state) {
	ctx.Step(`^I am on the homepage$`, func() error {
		return s.page.Navigate(baseURL + "/")
	})

	ctx.Step(`^I am on roll page "([^"]+)"$`, func(slug string) error {
		return s.page.Navigate(baseURL + "/roll/" + slug)
	})

	ctx.Step(`^I am on roll page "([^"]+)" with poem animation "([^"]+)"$`, func(slug, _ string) error {
		// Navigate — the animation mode is set in the roll's MD frontmatter, not via URL
		return s.page.Navigate(baseURL + "/roll/" + slug)
	})

	ctx.Step(`^I am on a roll page with a pair layout$`, func() error {
		return s.page.Navigate(baseURL + "/roll/matin-brumeux")
	})

	ctx.Step(`^I am on a roll page with images of different sizes$`, func() error {
		return s.page.Navigate(baseURL + "/roll/matin-brumeux")
	})

	ctx.Step(`^I navigate to roll "([^"]+)"$`, func(slug string) error {
		return s.page.Navigate(baseURL + "/roll/" + slug)
	})

	ctx.Step(`^the viewport is (\d+)x(\d+)$`, func(w, h int) error {
		s.page.MustSetViewport(w, h, 1, false)
		return nil
	})
}
