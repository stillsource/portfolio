package domain

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Simple Name", "simple-name"},
		{"Nuit à Tokyo", "nuit-a-tokyo"},
		{"Été en Île-de-France", "ete-en-ile-de-france"},
		{"Brüssel & Köln", "brussel-koln"},
		{"  leading trailing  ", "leading-trailing"},
		{"multiple   spaces", "multiple-spaces"},
		{"spécial!@#chars", "special-chars"},
		{"Paris 2024", "paris-2024"},
		{"--already-slug--", "already-slug"},
		{"UPPERCASE", "uppercase"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := Slugify(tt.in)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
