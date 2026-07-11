// Package model holds plain data types shared across the data layer, the
// generation pipeline, and the PDF/report code. It imports nothing internal so
// the Ent schema can embed these as JSON columns without an import cycle.
package model

// Color is a single swatch in a generated brand palette.
type Color struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

// BrandConcept is what Gemini produces for a single prompt: a set of candidate
// names plus supporting brand assets.
type BrandConcept struct {
	Names    []string `json:"names"`
	Tagline  string   `json:"tagline"`
	Mission  string   `json:"mission"`
	Palette  []Color  `json:"palette"`
	Keywords []string `json:"keywords"`
}

// Availability is one platform's verdict for one candidate name, produced by a
// checker goroutine in the fan-out.
type Availability struct {
	Name      string            `json:"name"`
	Platform  string            `json:"platform"`
	Available bool              `json:"available"`
	Meta      map[string]string `json:"meta,omitempty"`
	Err       string            `json:"err,omitempty"`
}
