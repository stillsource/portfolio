package paletteaggregator

import (
	"strings"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

func TestCIELAB_Aggregate(t *testing.T) {
	agg := NewCIELAB()

	t.Run("empty palettes returns nil", func(t *testing.T) {
		got := agg.Aggregate(nil, 5)
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("size zero returns nil", func(t *testing.T) {
		got := agg.Aggregate([][]string{{"#ff0000"}}, 0)
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("negative size returns nil", func(t *testing.T) {
		got := agg.Aggregate([][]string{{"#ff0000"}}, -1)
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("single palette passthrough", func(t *testing.T) {
		palette := [][]string{{"#ffffff", "#808080", "#000000"}}
		got := agg.Aggregate(palette, 3)
		if len(got) != 3 {
			t.Fatalf("got len %d, want 3", len(got))
		}
		for _, hex := range got {
			if !strings.HasPrefix(hex, "#") || len(hex) != 7 {
				t.Errorf("invalid hex %q", hex)
			}
		}
	})

	t.Run("output sorted bright to dark", func(t *testing.T) {
		// Mix of a very light and a very dark color — average should still be ordered.
		palettes := [][]string{
			{"#ffffff", "#000000"},
			{"#eeeeee", "#111111"},
		}
		got := agg.Aggregate(palettes, 2)
		if len(got) != 2 {
			t.Fatalf("got len %d, want 2", len(got))
		}
		c0, err0 := colorful.Hex(got[0])
		c1, err1 := colorful.Hex(got[1])
		if err0 != nil || err1 != nil {
			t.Fatalf("invalid hex in result: %v %v", got[0], got[1])
		}
		l0, _, _ := c0.Lab()
		l1, _, _ := c1.Lab()
		if l0 < l1 {
			t.Errorf("result not sorted bright→dark: L*[0]=%f < L*[1]=%f (colors: %v)", l0, l1, got)
		}
	})

	t.Run("invalid hex entries are skipped", func(t *testing.T) {
		palettes := [][]string{
			{"#ff0000", "not-a-color", "#0000ff"},
		}
		got := agg.Aggregate(palettes, 3)
		// Invalid entry at index 1 produces an empty bucket → output has 2 colors
		if len(got) == 0 {
			t.Error("expected at least one valid color in output")
		}
		for _, hex := range got {
			if _, err := colorful.Hex(hex); err != nil {
				t.Errorf("invalid hex in output: %q", hex)
			}
		}
	})

	t.Run("size limits colors taken from each palette", func(t *testing.T) {
		palettes := [][]string{
			{"#ff0000", "#00ff00", "#0000ff", "#ffff00", "#ff00ff"},
			{"#ff0000", "#00ff00", "#0000ff", "#ffff00", "#ff00ff"},
		}
		got := agg.Aggregate(palettes, 3)
		if len(got) != 3 {
			t.Errorf("got len %d, want 3", len(got))
		}
	})

	t.Run("all palettes contribute to average", func(t *testing.T) {
		// Two identical palettes → aggregate should equal the input (roundtrip through LAB)
		hex := "#7f7f7f"
		palettes := [][]string{{hex}, {hex}}
		got := agg.Aggregate(palettes, 1)
		if len(got) != 1 {
			t.Fatalf("got len %d, want 1", len(got))
		}
		// Allow ±1 per channel due to LAB roundtrip rounding
		assertHexClose(t, got[0], hex, 2)
	})
}

// assertHexClose checks that two hex colors are within maxDelta per RGB channel.
func assertHexClose(t *testing.T, got, want string, maxDelta float64) {
	t.Helper()
	cGot, err := colorful.Hex(got)
	if err != nil {
		t.Fatalf("invalid got hex %q: %v", got, err)
	}
	cWant, err := colorful.Hex(want)
	if err != nil {
		t.Fatalf("invalid want hex %q: %v", want, err)
	}
	dist := cGot.DistanceLab(cWant)
	if dist > maxDelta {
		t.Errorf("color %q too far from expected %q (LAB distance %.4f > %.4f)", got, want, dist, maxDelta)
	}
}
