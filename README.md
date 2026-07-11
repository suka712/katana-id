# KatanaID

**AI Branding Toolkit.** Describe a product idea and KatanaID generates a brand
concept — names, tagline, palette, keywords — with Google Gemini, then fans out
goroutines across 20+ external APIs to check name availability across domains,
package registries, and social platforms in real time. Results stream in live
over SSE and can be exported as a PDF brand report.

**Live:** [katanaid.com](https://katanaid.com)

## How it works

```
prompt ──▶ Gemini ──▶ brand concept (names, tagline, palette, keywords)
                          │
                          ▼
              goroutine fan-out (per name × per platform)
     ┌──────────────┬───────────────┬────────────────────────┐
     │ domains (DNS)│ code/registries│ community handles       │
     │ .com .io .ai…│ GitHub GitLab  │ Reddit dev.to Keybase   │
     │              │ npm PyPI crates│                         │
     │              │ RubyGems Docker│                         │
     │              │ Homebrew       │                         │
     └──────────────┴───────────────┴────────────────────────┘
                          │  results stream via SSE, cell by cell
                          ▼
                 availability matrix ──▶ PDF brand report
```

Every request is scored by a **trust engine** that combines a client-side
browser fingerprint with server-visible signals (user-agent, auth state).

## Tech Stack

- **Frontend:** React + Vite + TypeScript, Tailwind, Framer Motion
- **Backend:** Go (chi router, goroutine fan-out, SSE streaming)
- **AI:** Google Gemini (`google.golang.org/genai`)
- **Database:** PostgreSQL via **Ent** (type-safe schema + auto-migration)
- **PDF:** `github.com/go-pdf/fpdf`

## Development

### Prerequisites

- Node.js 20+ (pnpm)
- Go 1.25+
- PostgreSQL 14+

### Backend

```bash
cd server
cp .env.example .env          # fill in DB_URL, RESEND_API_KEY, GEMINI_API_KEY, OAuth
go run ./cmd/server           # Ent auto-migrates on boot — no manual migration step
```

Regenerate the Ent client after changing anything in `internal/db/ent/schema`:

```bash
go generate ./internal/db/ent/...
```

### Frontend

```bash
cd client
cp .env.example .env          # VITE_API_URL=http://localhost:8080
pnpm install
pnpm dev
```

The frontend runs on `http://localhost:5173` and the backend on
`http://localhost:8080`.

## API

| Method | Route                     | Notes                                        |
|--------|---------------------------|----------------------------------------------|
| `POST` | `/generate`               | auth-gated; `{prompt, fingerprint, ...}` → concept + kit id |
| `GET`  | `/generate/{id}/stream`   | SSE; one availability event per resolved check |
| `GET`  | `/kits/{id}/pdf`          | auth-gated; PDF brand report                 |
| `*`    | `/auth/*`                 | OTP email + Google/GitHub OAuth              |

> `server-depr/` is the previous Ent-based server, kept for reference.
