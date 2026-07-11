// Package gemini wraps Google's Gemini API to turn a free-form product idea
// into a structured brand concept (candidate names, tagline, palette, ...).
package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/trnahnh/katana-id/internal/model"
)

// ErrNoKey is returned by New when no API key is configured. Callers may treat
// this as "AI disabled" and fall back to a local generator.
var ErrNoKey = errors.New("gemini: missing API key")

type Client struct {
	genai *genai.Client
	model string
}

// New builds a Gemini client. modelName falls back to a fast default when empty.
func New(ctx context.Context, apiKey, modelName string) (*Client, error) {
	if apiKey == "" {
		return nil, ErrNoKey
	}
	if modelName == "" {
		modelName = "gemini-flash-latest"
	}
	gc, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	return &Client{genai: gc, model: modelName}, nil
}

const promptTemplate = `You are a senior brand strategist. Turn the product idea below into a brandable identity.

Rules:
- "names": exactly 8 short, memorable, brandable names. Lowercase, no spaces or punctuation, each 3-14 characters, safe to use as a domain/handle. Prefer coined/invented words over generic dictionary words.
- "tagline": one punchy tagline, max 8 words.
- "mission": one sentence describing what the product does.
- "palette": exactly 5 colors, each with a human "name" and a "hex" like "#1A2B3C".
- "keywords": 6 lowercase keywords describing the brand's vibe.

Respond with ONLY a JSON object of this exact shape, no prose, no markdown:
{"names":["..."],"tagline":"...","mission":"...","palette":[{"name":"...","hex":"#RRGGBB"}],"keywords":["..."]}

Product idea: %q`

// GenerateConcept asks Gemini for a brand concept and parses the JSON response.
func (c *Client) GenerateConcept(ctx context.Context, prompt string) (model.BrandConcept, error) {
	var concept model.BrandConcept

	temp := float32(1.1)
	cfg := &genai.GenerateContentConfig{
		Temperature: &temp,
	}

	resp, err := c.genai.Models.GenerateContent(
		ctx, c.model,
		genai.Text(fmt.Sprintf(promptTemplate, prompt)),
		cfg,
	)
	if err != nil {
		return concept, err
	}

	raw := extractJSON(resp.Text())
	if raw == "" {
		return concept, errors.New("gemini: empty response")
	}
	if err := json.Unmarshal([]byte(raw), &concept); err != nil {
		return concept, fmt.Errorf("gemini: parse concept: %w", err)
	}

	concept.Names = SanitizeNames(concept.Names)
	if len(concept.Names) == 0 {
		return concept, errors.New("gemini: no usable names returned")
	}
	return concept, nil
}

// extractJSON pulls the JSON object out of a model response, tolerating
// ```json fences, leading prose, and trailing commentary by slicing from the
// first "{" to the last "}".
func extractJSON(s string) string {
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start < 0 || end < start {
		return ""
	}
	return s[start : end+1]
}

// SanitizeNames lowercases, strips unsafe characters, and de-duplicates the
// candidate names so they are valid as domains/handles.
func SanitizeNames(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, n := range in {
		n = sanitize(n)
		if len(n) < 2 || len(n) > 30 {
			continue
		}
		if _, dup := seen[n]; dup {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
