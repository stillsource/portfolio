package e2e

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/godog"
)

func registerPoemSteps(ctx *godog.ScenarioContext, s *state) {
	ctx.Step(`^I scroll to image (\d+)$`, func(n int) error {
		_, err := s.page.Eval(fmt.Sprintf(`() => {
			const wrappers = [...document.querySelectorAll('.image-wrapper')];
			const el = wrappers[%d];
			if (!el) return;
			const rect = el.getBoundingClientRect();
			window.scrollTo({
				top: rect.top + window.scrollY - window.innerHeight / 2 + rect.height / 2,
				behavior: 'instant'
			});
		}`, n-1))
		if err != nil {
			return err
		}
		time.Sleep(400 * time.Millisecond) // let scroll handler and IntersectionObserver fire
		return nil
	})

	ctx.Step(`^I scroll past image (\d+)$`, func(n int) error {
		_, err := s.page.Eval(fmt.Sprintf(`() => {
			const wrappers = [...document.querySelectorAll('.image-wrapper')];
			const el = wrappers[%d];
			if (!el) return;
			const rect = el.getBoundingClientRect();
			window.scrollTo({
				top: rect.top + window.scrollY + window.innerHeight * 2,
				behavior: 'instant'
			});
		}`, n-1))
		if err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond)
		return nil
	})

	ctx.Step(`^the poem for image (\d+) should have opacity greater than "([^"]+)"$`, func(n int, expected string) error {
		return assertPoemOpacity(s, n, expected, true)
	})

	ctx.Step(`^the poem for image (\d+) should have opacity less than "([^"]+)"$`, func(n int, expected string) error {
		return assertPoemOpacity(s, n, expected, false)
	})

	ctx.Step(`^the poem for image (\d+) should contain visible text$`, func(n int) error {
		result, err := s.page.Eval(fmt.Sprintf(`() => {
			const thoughts = [...document.querySelectorAll('.thought-fragment')];
			const el = thoughts[%d];
			if (!el) return '';
			return (el.textContent || '').trim();
		}`, n-1))
		if err != nil {
			return err
		}
		text := result.Value.String()
		if text == "" {
			return fmt.Errorf("expected poem for image %d to contain visible text, got empty string", n)
		}
		return nil
	})
}

func assertPoemOpacity(s *state, n int, threshold string, greaterThan bool) error {
	result, err := s.page.Eval(fmt.Sprintf(`() => {
		const thoughts = [...document.querySelectorAll('.thought-fragment')];
		const el = thoughts[%d];
		if (!el) return '-1';
		return window.getComputedStyle(el).opacity;
	}`, n-1))
	if err != nil {
		return err
	}
	opacityStr := result.Value.String()
	if opacityStr == "-1" {
		return fmt.Errorf("poem element for image %d not found", n)
	}
	opacity, err := strconv.ParseFloat(opacityStr, 64)
	if err != nil {
		return fmt.Errorf("cannot parse opacity %q: %w", opacityStr, err)
	}
	thresh, err := strconv.ParseFloat(threshold, 64)
	if err != nil {
		return fmt.Errorf("cannot parse threshold %q: %w", threshold, err)
	}
	if greaterThan && opacity <= thresh {
		return fmt.Errorf("expected poem opacity > %s, got %s", threshold, opacityStr)
	}
	if !greaterThan && opacity >= thresh {
		return fmt.Errorf("expected poem opacity < %s, got %s", threshold, opacityStr)
	}
	return nil
}
