package e2e

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

func registerLayoutSteps(ctx *godog.ScenarioContext, s *state) {
	ctx.Step(`^the two paired images should appear in the same horizontal row$`, func() error {
		result, err := s.page.Eval(`() => {
			const items = document.querySelectorAll('.pair-item');
			if (items.length < 2) return JSON.stringify({ok: false, reason: 'fewer than 2 .pair-item elements found'});
			const r0 = items[0].getBoundingClientRect();
			const r1 = items[1].getBoundingClientRect();
			const diff = Math.abs(r0.top - r1.top);
			return JSON.stringify({ok: diff < 50, diff: diff, top0: r0.top, top1: r1.top});
		}`)
		if err != nil {
			return err
		}
		return checkJSONOk(result.Value.String(), "paired images not in same row")
	})

	ctx.Step(`^each image should occupy roughly half the viewport width$`, func() error {
		result, err := s.page.Eval(`() => {
			const items = document.querySelectorAll('.pair-item');
			const vw = window.innerWidth;
			const widths = Array.from(items).map(p => p.getBoundingClientRect().width);
			const ok = widths.length >= 2 && widths.every(w => w > vw * 0.3 && w < vw * 0.7);
			return JSON.stringify({ok: ok, widths: widths, vw: vw});
		}`)
		if err != nil {
			return err
		}
		return checkJSONOk(result.Value.String(), "images do not occupy ~half viewport width each")
	})

	ctx.Step(`^the two paired images should appear stacked vertically$`, func() error {
		result, err := s.page.Eval(`() => {
			const items = document.querySelectorAll('.pair-item');
			if (items.length < 2) return JSON.stringify({ok: false, reason: 'fewer than 2 .pair-item elements'});
			const r0 = items[0].getBoundingClientRect();
			const r1 = items[1].getBoundingClientRect();
			// Stacked: second item starts below the bottom of the first (with 10px tolerance)
			const ok = r1.top > r0.bottom - 10;
			return JSON.stringify({ok: ok, top0: r0.top, bottom0: r0.bottom, top1: r1.top});
		}`)
		if err != nil {
			return err
		}
		return checkJSONOk(result.Value.String(), "images are not stacked vertically on mobile")
	})

	ctx.Step(`^the "([^"]+)" image should be narrower than the "([^"]+)" image$`, func(smallSize, largeSize string) error {
		smallSel := fmt.Sprintf(".image-wrapper.size-%s", smallSize)
		var largeSel string
		if largeSize == "full" {
			largeSel = `.image-wrapper:not([class*="size-"])`
		} else {
			largeSel = fmt.Sprintf(".image-wrapper.size-%s", largeSize)
		}
		script := fmt.Sprintf(`() => {
			const small = document.querySelector(%s);
			const large = document.querySelector(%s);
			if (!small) return JSON.stringify({ok: false, reason: 'small element not found'});
			if (!large) return JSON.stringify({ok: false, reason: 'large element not found'});
			const sw = small.getBoundingClientRect().width;
			const lw = large.getBoundingClientRect().width;
			return JSON.stringify({ok: sw < lw, smallWidth: sw, largeWidth: lw});
		}`, jsonString(smallSel), jsonString(largeSel))
		result, err := s.page.Eval(script)
		if err != nil {
			return err
		}
		return checkJSONOk(result.Value.String(), fmt.Sprintf("'%s' image is not narrower than '%s' image", smallSize, largeSize))
	})
}

// checkJSONOk parses {"ok": bool, ...} from raw JSON and returns an error if ok=false.
func checkJSONOk(raw string, fallbackMsg string) error {
	var payload struct {
		Ok     bool   `json:"ok"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return fmt.Errorf("unexpected response %q: %w", raw, err)
	}
	if !payload.Ok {
		msg := fallbackMsg
		if payload.Reason != "" {
			msg = payload.Reason
		}
		return fmt.Errorf("%s (raw: %s)", msg, raw)
	}
	return nil
}
