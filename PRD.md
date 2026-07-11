# KatanaID — PRD

**AI Branding Toolkit.** Turn a product idea into a brand concept and instantly
check whether the name is available everywhere that matters.

## User Flow

1. User lands on the homepage and sees a single prompt input.
2. User describes their product — e.g. `"an app that helps dog owners find playdates"`.
3. User hits **Generate** → redirected to sign in if not authenticated.
4. After auth, the query is preserved and generation starts immediately.
5. Gemini returns a brand concept: candidate names, tagline, mission, palette, keywords.
6. The results page renders the concept, then fills an availability matrix
   cell-by-cell as each checker goroutine resolves (streamed over SSE).
7. User reviews availability, then downloads a **PDF brand report**.

## Auth

- Generation requires authentication (sign in before a generation runs).
- OTP email sign-in **and** Google / GitHub OAuth are supported.
- Anonymous users can type on the landing page but are redirected to sign in
  before results load; the prompt is preserved across sign-in.

## Backend Architecture

### Endpoints

```
POST /generate            — auth-gated; { prompt, fingerprint, components }
                            → { id, concept, names, total, trust }
GET  /generate/{id}/stream — SSE; emits one Availability event per resolved check
GET  /kits/{id}/pdf        — auth-gated; PDF brand report for a saved kit
```

### Generation pipeline

```
POST /generate
  → trust engine scores the request (fingerprint + user-agent + auth state)
  → Gemini generates a BrandConcept (fallback: local generator)
  → persist a BrandKit (Ent)
  → for each candidate name: fan out one goroutine per checker
      ├── domains (DNS): .com .io .dev .co .net .org .app .ai .xyz .me
      ├── code/registries: GitHub, GitLab, npm, PyPI, crates.io,
      │                    RubyGems, Docker Hub, Homebrew
      └── community: Reddit, dev.to, Keybase  (+ X, search when keyed)
  → each result streams over SSE as it lands; the aggregate is saved on the kit
```

Each checker runs independently; results stream to the client as soon as they
resolve — not after all are done. With the defaults, a single generation
orchestrates 21 external integrations per name.

### Data model (Ent)

`User`, `Provider` (OAuth links), `OTP`, `Session`, and `BrandKit` (prompt,
concept JSON, aggregated results JSON, trust score, fingerprint). Ent's typed
schema is the single source of truth and auto-migrates on boot, eliminating
schema drift between code and database.

### Trust engine

A transparent, tunable 0–100 score combining a client browser fingerprint
(screen, timezone, canvas, hardware) with server-visible signals (user-agent
heuristics for automation, authentication state, network). Surfaced in the UI
and stored on each kit.

## Frontend Pages

| Route      | Page     | Notes                                            |
|------------|----------|--------------------------------------------------|
| `/`        | Landing  | Prompt input + "Generate"                        |
| `/signin`  | Sign In  | OTP email + Google/GitHub OAuth; redirects back  |
| `/result`  | Results  | Concept + SSE availability matrix + PDF export   |

## Future / Out of Scope

- Saved kit history & favoriting
- Instagram / TikTok handle checks (require authenticated scraping)
- Shareable public report links
