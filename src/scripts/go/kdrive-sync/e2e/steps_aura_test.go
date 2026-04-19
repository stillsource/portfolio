package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func registerAuraSteps(ctx *godog.ScenarioContext, s *state) {
	ctx.Step(`^I note the current background color$`, func() error {
		result, err := s.page.Eval(`() => getComputedStyle(document.documentElement).getPropertyValue('--p1').trim()`)
		if err != nil {
			return err
		}
		s.noteColor = result.Value.String()
		return nil
	})

	ctx.Step(`^the background color should have changed$`, func() error {
		result, err := s.page.Eval(`() => getComputedStyle(document.documentElement).getPropertyValue('--p1').trim()`)
		if err != nil {
			return err
		}
		current := result.Value.String()
		if current == s.noteColor {
			return fmt.Errorf("background color did not change, still: %q", current)
		}
		return nil
	})

	ctx.Step(`^I scroll to the photo "([^"]+)"$`, func(altText string) error {
		script := fmt.Sprintf(`() => {
			const imgs = document.querySelectorAll('.styled-image');
			for (const img of imgs) {
				if (img.alt && img.alt.includes(%s)) {
					const rect = img.getBoundingClientRect();
					window.scrollTo({
						top: rect.top + window.scrollY - window.innerHeight / 2 + rect.height / 2,
						behavior: 'instant'
					});
					return true;
				}
			}
			return false;
		}`, jsonString(altText))
		result, err := s.page.Eval(script)
		if err != nil {
			return err
		}
		if !result.Value.Bool() {
			return fmt.Errorf("photo with alt text containing %q not found", altText)
		}
		time.Sleep(600 * time.Millisecond) // let IntersectionObserver color-trigger fire
		return nil
	})

	ctx.Step(`^the CSS variable "([^"]+)" should not be a warm orange color$`, func(variable string) error {
		// Resolve the CSS variable to an rgb() string via a temporary DOM element
		script := fmt.Sprintf(`() => {
			const val = getComputedStyle(document.documentElement).getPropertyValue(%s).trim();
			if (!val) return '';
			const tmp = document.createElement('span');
			tmp.style.color = val;
			document.body.appendChild(tmp);
			const rgb = getComputedStyle(tmp).color;
			document.body.removeChild(tmp);
			return rgb;
		}`, jsonString(variable))
		result, err := s.page.Eval(script)
		if err != nil {
			return err
		}
		rgbStr := result.Value.String()
		if rgbStr == "" {
			return fmt.Errorf("CSS variable %q resolved to empty string", variable)
		}
		r, g, b, err := parseRGB(rgbStr)
		if err != nil {
			return fmt.Errorf("cannot parse color %q for variable %q: %w", rgbStr, variable, err)
		}
		h, s2, _ := rgbToHSL(r, g, b)
		// Warm orange: hue in [15..45] degrees with significant saturation
		if h >= 15 && h <= 45 && s2 > 0.35 {
			return fmt.Errorf("color %q for %q is warm orange (hue=%.0f°, sat=%.2f) — expected blue/violet palette", rgbStr, variable, h, s2)
		}
		return nil
	})
}

// jsonString encodes s as a JSON string literal safe for inline JS use.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// parseRGB parses "rgb(r, g, b)" or "rgba(r, g, b, a)" into 0-255 integers.
func parseRGB(rgb string) (r, g, b int, err error) {
	rgb = strings.TrimSpace(rgb)
	rgb = strings.TrimPrefix(rgb, "rgba(")
	rgb = strings.TrimPrefix(rgb, "rgb(")
	rgb = strings.TrimSuffix(rgb, ")")
	parts := strings.Split(rgb, ",")
	if len(parts) < 3 {
		return 0, 0, 0, fmt.Errorf("unexpected rgb format: %q", rgb)
	}
	parse := func(s string) (int, error) {
		v, e := strconv.Atoi(strings.TrimSpace(s))
		return v, e
	}
	if r, err = parse(parts[0]); err != nil {
		return
	}
	if g, err = parse(parts[1]); err != nil {
		return
	}
	b, err = parse(parts[2])
	return
}

// rgbToHSL converts 0-255 RGB to HSL (hue 0-360, saturation 0-1, lightness 0-1).
func rgbToHSL(r, g, b int) (h, s, l float64) {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	l = (max + min) / 2

	if max == min {
		return 0, 0, l
	}

	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	case bf:
		h = (rf-gf)/d + 4
	}
	h *= 60
	return
}
