package brand

import (
	"strings"

	"github.com/trnahnh/katana-id/internal/gemini"
	"github.com/trnahnh/katana-id/internal/model"
)

// defaultPalette is a neutral, on-brand fallback palette.
var defaultPalette = []model.Color{
	{Name: "Katana Orange", Hex: "#FF6B35"},
	{Name: "Ink", Hex: "#1A1A1A"},
	{Name: "Slate", Hex: "#3D4451"},
	{Name: "Mist", Hex: "#E8ECEF"},
	{Name: "Signal", Hex: "#2EC4B6"},
}

var nameAffixes = []struct{ prefix, suffix string }{
	{"", "ly"}, {"", "ify"}, {"get", ""}, {"", "hq"},
	{"", "labs"}, {"try", ""}, {"", "kit"}, {"go", ""},
	{"", "flow"}, {"", "io"},
}

// fallbackConcept produces a deterministic brand concept from the prompt without
// calling any external model. It keeps the app fully functional when Gemini is
// unconfigured or unreachable.
func fallbackConcept(prompt string) model.BrandConcept {
	base := baseWord(prompt)

	candidates := []string{base}
	for _, a := range nameAffixes {
		candidates = append(candidates, a.prefix+base+a.suffix)
	}
	names := gemini.SanitizeNames(candidates)
	if len(names) > 8 {
		names = names[:8]
	}

	return model.BrandConcept{
		Names:    names,
		Tagline:  "Own your name everywhere.",
		Mission:  "A brand identity for " + strings.TrimSpace(prompt) + ".",
		Palette:  defaultPalette,
		Keywords: keywords(prompt),
	}
}

// baseWord picks a short, usable seed word from the prompt.
func baseWord(prompt string) string {
	for _, w := range strings.Fields(strings.ToLower(prompt)) {
		w = clean(w)
		if len(w) >= 3 && !stopWords[w] {
			return w
		}
	}
	return "katana"
}

func keywords(prompt string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, w := range strings.Fields(strings.ToLower(prompt)) {
		w = clean(w)
		if len(w) < 3 || stopWords[w] {
			continue
		}
		if _, dup := seen[w]; dup {
			continue
		}
		seen[w] = struct{}{}
		out = append(out, w)
		if len(out) == 6 {
			break
		}
	}
	if len(out) == 0 {
		out = []string{"modern", "bold", "trustworthy"}
	}
	return out
}

func clean(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var stopWords = map[string]bool{
	"the": true, "for": true, "and": true, "app": true, "with": true,
	"that": true, "this": true, "called": true, "named": true, "your": true,
	"you": true, "our": true, "are": true, "was": true, "has": true,
}
