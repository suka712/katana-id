// Package report renders a brand kit into a downloadable PDF summary.
package report

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/trnahnh/katana-id/internal/model"
)

// Kit is everything the report needs about a single generated brand.
type Kit struct {
	Prompt     string
	Concept    model.BrandConcept
	Results    []model.Availability
	TrustScore float64
	CreatedAt  time.Time
}

const (
	ink   = 0x1A
	muted = 0x6B
)

// Render produces a one-or-more page PDF report and returns the raw bytes.
func Render(k Kit) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("KatanaID Brand Report", true)
	pdf.SetMargins(18, 18, 18)
	pdf.AddPage()

	primary := "Your Brand"
	if len(k.Concept.Names) > 0 {
		primary = k.Concept.Names[0]
	}

	// ── Header ────────────────────────────────────────────────────────────
	pdf.SetTextColor(muted, muted, muted)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(0, 6, "KATANAID  //  BRAND REPORT", "", 1, "L", false, 0, "")

	pdf.SetTextColor(ink, ink, ink)
	pdf.SetFont("Helvetica", "B", 30)
	pdf.CellFormat(0, 16, ascii(titleCase(primary)), "", 1, "L", false, 0, "")

	if k.Concept.Tagline != "" {
		pdf.SetFont("Helvetica", "I", 13)
		pdf.SetTextColor(muted, muted, muted)
		pdf.CellFormat(0, 8, ascii(k.Concept.Tagline), "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	// ── Prompt + meta ────────────────────────────────────────────────────
	pdf.SetTextColor(ink, ink, ink)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "Brief", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.MultiCell(0, 6, ascii(k.Prompt), "", "L", false)
	pdf.Ln(1)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(muted, muted, muted)
	pdf.CellFormat(0, 6, fmt.Sprintf("Trust score: %d/100   ·   Generated %s",
		int(k.TrustScore), k.CreatedAt.Format("Jan 2, 2006 15:04")), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	if k.Concept.Mission != "" {
		section(pdf, "Mission")
		pdf.SetFont("Helvetica", "", 11)
		pdf.SetTextColor(ink, ink, ink)
		pdf.MultiCell(0, 6, ascii(k.Concept.Mission), "", "L", false)
		pdf.Ln(2)
	}

	// ── Palette ──────────────────────────────────────────────────────────
	if len(k.Concept.Palette) > 0 {
		section(pdf, "Palette")
		x0 := pdf.GetX()
		y := pdf.GetY()
		sw := 32.0
		for i, c := range k.Concept.Palette {
			r, g, b := hexToRGB(c.Hex)
			x := x0 + float64(i)*(sw+4)
			pdf.SetFillColor(r, g, b)
			pdf.Rect(x, y, sw, 18, "F")
			pdf.SetXY(x, y+19)
			pdf.SetFont("Helvetica", "B", 8)
			pdf.SetTextColor(ink, ink, ink)
			pdf.CellFormat(sw, 4, ascii(c.Name), "", 2, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 8)
			pdf.SetTextColor(muted, muted, muted)
			pdf.CellFormat(sw, 4, strings.ToUpper(c.Hex), "", 0, "L", false, 0, "")
		}
		pdf.SetXY(x0, y+30)
		pdf.Ln(2)
	}

	// ── Keywords ─────────────────────────────────────────────────────────
	if len(k.Concept.Keywords) > 0 {
		section(pdf, "Keywords")
		pdf.SetFont("Helvetica", "", 11)
		pdf.SetTextColor(ink, ink, ink)
		pdf.MultiCell(0, 6, ascii(strings.Join(k.Concept.Keywords, "  ·  ")), "", "L", false)
		pdf.Ln(2)
	}

	// ── Availability matrix ──────────────────────────────────────────────
	section(pdf, "Availability")
	renderMatrix(pdf, k)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// renderMatrix draws a name × platform table of availability results.
func renderMatrix(pdf *fpdf.Fpdf, k Kit) {
	// Group results by name, collecting the set of platforms.
	byName := map[string]map[string]model.Availability{}
	platformSet := map[string]struct{}{}
	for _, r := range k.Results {
		if byName[r.Name] == nil {
			byName[r.Name] = map[string]model.Availability{}
		}
		byName[r.Name][r.Platform] = r
		platformSet[r.Platform] = struct{}{}
	}

	names := k.Concept.Names
	if len(names) == 0 {
		for n := range byName {
			names = append(names, n)
		}
		sort.Strings(names)
	}

	platforms := make([]string, 0, len(platformSet))
	for p := range platformSet {
		platforms = append(platforms, p)
	}
	sort.Strings(platforms)

	if len(platforms) == 0 {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.SetTextColor(muted, muted, muted)
		pdf.CellFormat(0, 6, "No availability data.", "", 1, "L", false, 0, "")
		return
	}

	pdf.SetFont("Helvetica", "", 8)
	for _, name := range names {
		row := byName[name]
		if row == nil {
			continue
		}
		var available, taken int
		for _, p := range platforms {
			r, ok := row[p]
			if !ok || r.Err != "" {
				continue
			}
			if r.Available {
				available++
			} else {
				taken++
			}
		}
		pdf.SetTextColor(ink, ink, ink)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(60, 6, ascii(name), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(muted, muted, muted)
		pdf.CellFormat(0, 6, fmt.Sprintf("%s available · %s taken",
			strconv.Itoa(available), strconv.Itoa(taken)), "", 1, "L", false, 0, "")
	}
}

func section(pdf *fpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(ink, ink, ink)
	pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")
}

// hexToRGB parses "#RRGGBB" (or "RRGGBB"); it falls back to mid-grey on error.
func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return 0x88, 0x88, 0x88
	}
	v, err := strconv.ParseInt(hex, 16, 32)
	if err != nil {
		return 0x88, 0x88, 0x88
	}
	return int(v>>16) & 0xFF, int(v>>8) & 0xFF, int(v) & 0xFF
}

// titleCase upper-cases the first rune of s.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// ascii strips non-latin1 runes so the core PDF fonts render cleanly.
func ascii(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 256 {
			b.WriteRune(r)
		}
	}
	return b.String()
}
