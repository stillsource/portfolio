package e2e

import (
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"github.com/go-rod/rod/lib/proto"
)

func registerCursorSteps(ctx *godog.ScenarioContext, s *state) {
	ctx.Step(`^I move the mouse to position (\d+),(\d+)$`, func(x, y int) error {
		err := s.page.Mouse.MoveTo(proto.Point{X: float64(x), Y: float64(y)})
		if err != nil {
			return err
		}
		time.Sleep(150 * time.Millisecond)
		return nil
	})

	ctx.Step(`^the cursor should be visible$`, func() error {
		return assertCursorOpacity(s, false)
	})

	ctx.Step(`^the cursor should still be visible$`, func() error {
		return assertCursorOpacity(s, false)
	})

	ctx.Step(`^the mouse leaves the browser window$`, func() error {
		_, err := s.page.Eval(`() => window.dispatchEvent(new MouseEvent('mouseleave'))`)
		if err != nil {
			return err
		}
		time.Sleep(150 * time.Millisecond)
		return nil
	})

	ctx.Step(`^the cursor should not be visible$`, func() error {
		return assertCursorOpacity(s, true)
	})
}

// assertCursorOpacity checks cursor opacity. If expectHidden=true, expects "0"; otherwise expects non-"0".
func assertCursorOpacity(s *state, expectHidden bool) error {
	result, err := s.page.Eval(`() => {
		const el = document.getElementById('custom-cursor');
		if (!el) return 'missing';
		return window.getComputedStyle(el).opacity;
	}`)
	if err != nil {
		return err
	}
	opacity := result.Value.String()
	if opacity == "missing" {
		return fmt.Errorf("cursor element #custom-cursor not found in DOM")
	}
	if expectHidden && opacity != "0" {
		return fmt.Errorf("expected cursor to be hidden (opacity=0), got %s", opacity)
	}
	if !expectHidden && opacity == "0" {
		return fmt.Errorf("expected cursor to be visible (opacity!=0), got 0")
	}
	return nil
}
