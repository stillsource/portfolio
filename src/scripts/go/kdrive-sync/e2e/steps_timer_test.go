package e2e

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/godog"
)

func registerTimerSteps(ctx *godog.ScenarioContext, s *state) {
	ctx.Step(`^I scroll to the bottom of the page$`, func() error {
		_, err := s.page.Eval(`() => window.scrollTo(0, document.body.scrollHeight)`)
		if err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond) // let IntersectionObserver fire
		return nil
	})

	ctx.Step(`^I scroll back to the top$`, func() error {
		_, err := s.page.Eval(`() => window.scrollTo(0, 0)`)
		if err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond)
		return nil
	})

	ctx.Step(`^the navigation timer should start$`, func() error {
		el := s.page.MustElement("#next-roll-trigger")
		classes, err := el.Attribute("class")
		if err != nil {
			return err
		}
		if classes == nil || !containsClass(*classes, "is-loading") {
			return fmt.Errorf("expected #next-roll-trigger to have class 'is-loading', got: %v", classes)
		}
		return nil
	})

	ctx.Step(`^the progress bar value should be "([^"]+)"$`, func(expected string) error {
		bar := s.page.MustElement(".next-progress-bar")
		val, err := bar.Attribute("aria-valuenow")
		if err != nil {
			return err
		}
		if val == nil || *val != expected {
			return fmt.Errorf("expected aria-valuenow=%q, got %v", expected, val)
		}
		return nil
	})

	ctx.Step(`^the navigation timer should be cancelled$`, func() error {
		el := s.page.MustElement("#next-roll-trigger")
		classes, err := el.Attribute("class")
		if err != nil {
			return err
		}
		if classes != nil && containsClass(*classes, "is-loading") {
			return fmt.Errorf("expected 'is-loading' to be removed from #next-roll-trigger")
		}
		return nil
	})

	ctx.Step(`^after (\d+) milliseconds the page should have navigated$`, func(ms string) error {
		d, err := strconv.Atoi(ms)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(d) * time.Millisecond)
		url := s.page.MustInfo().URL
		if url == baseURL+"/roll/matin-brumeux" {
			return fmt.Errorf("page did not navigate after %dms, still on %s", d, url)
		}
		return nil
	})
}

func containsClass(classes, target string) bool {
	for _, c := range splitClasses(classes) {
		if c == target {
			return true
		}
	}
	return false
}

func splitClasses(classes string) []string {
	var parts []string
	start := -1
	for i, ch := range classes {
		if ch == ' ' || ch == '\t' {
			if start >= 0 {
				parts = append(parts, classes[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		parts = append(parts, classes[start:])
	}
	return parts
}
