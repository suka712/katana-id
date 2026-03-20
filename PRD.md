# KatanaID — PRD

## User Flow

1. User lands on homepage, sees a single prompt input
2. User types anything — a name, a description, or both
   - `"ruffle"`                              (→ user has a name)           → check that name + suggest names
   - `"tinder for dog lovers called ruffle"` (→ user has a name)           → check that name + suggest names
   - `"tinder for dog lovers"`               (→ user doesn't have a name)  → suggest names only, no direct check yet
3. User hits "Check" → **redirected to sign in if not authenticated**
4. After auth, user is returned to the results page for their query
5. Results page loads with skeleton table while backend streams results via SSE
6. Table fills in cell-by-cell as each goroutine check resolves:
   - Rows = names (user-provided + AI-suggested)
   - Columns = `.com` domain, GitHub org, npm package, Twitter/X handle
7. User browses results, picks a name they like

## Auth

- **Checking requires authentication** (sign in before any check runs)
- Saving / favoriting names also requires sign in
- Anonymous users: can use the landing page input, but are redirected to sign in before results load
- After sign in, query is preserved and check starts immediately

## Backend Architecture

### Endpoint

```
POST /check        — auth-gated, accepts { query: string }
GET  /check/:id    — SSE stream; emits CheckResult events as goroutines resolve
```

### Goroutine Fan-out

```
POST /check
  → parse input (extract name if present, else AI generates suggestions)
  → for each name: fan out goroutines per checker
      ├── checkDomain(ctx, name)   — DNS / WHOIS lookup
      ├── checkGitHub(ctx, name)   — api.github.com/orgs/{name}
      ├── checkNpm(ctx, name)      — registry.npmjs.org/{name}
      └── checkTwitter(ctx, name)  — scrape or Rapid API
  → collect results via channel, stream each via SSE as it lands
```

Each checker runs independently; results stream to the client as soon as they resolve — not after all are done.

### CheckResult shape

```go
type CheckResult struct {
    Name      string
    Platform  string // "domain.com" | "github" | "npm" | "twitter"
    Available bool
    Error     string
}
```

### Streaming (SSE)

- `GET /check/:id` opens an SSE connection
- Each resolved goroutine writes one `data: <json>` event to the stream
- Connection closes when all goroutines for a session are done
- Frontend skeleton cells flip to ✓ / ✗ as events arrive

## Checks

| Platform   | Method                        | Endpoint / API                        |
|------------|-------------------------------|---------------------------------------|
| `.com`     | DNS lookup / WHOIS            | net.LookupHost or WHOIS API           |
| GitHub org | REST API (no auth needed)     | `api.github.com/orgs/{name}`          |
| npm        | Registry fetch                | `registry.npmjs.org/{name}`           |
| Twitter/X  | Scrape or third-party API     | Rapid API or similar                  |

## Frontend Pages

| Route         | Page            | Notes                                      |
|---------------|-----------------|--------------------------------------------|
| `/`           | Landing         | Input + "Check" button                     |
| `/signin`     | Sign In         | Email OTP; after auth, redirect back       |
| `/result`    | Results         | SSE-driven table; requires auth            |

## Future / Out of Scope (MVP)

- Search presence (Google, App Store, Play Store)
- Shareable summary reports
- Favoriting / saving names
- OAuth providers (Google, GitHub, etc.) — routes exist, not implemented
