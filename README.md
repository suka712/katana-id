# KatanaID

Naming a thing is the hardest part of starting it. You have an idea — "a budgeting
app for freelancers," "a place where dog owners find playdates" — but the moment
you pick a name, the real work begins. Is the `.com` taken? What about the GitHub
org, the npm package, the Reddit community, the handle you'll want on half a dozen
platforms you haven't thought of yet? People open twenty browser tabs to find out,
give up halfway, and settle for something worse than what they started with.

KatanaID is the tool that does those twenty tabs for you, and does them before you
even have a name.

You describe your idea in a sentence. Gemini reads it and comes back with a whole
brand concept — a handful of coined, brandable names, a tagline, a mission line, a
color palette, a few keywords to capture the vibe. Then, the instant those names
exist, the backend goes to work: for every candidate it fans out across more than
twenty services at once — domain registrars, GitHub, GitLab, npm, PyPI, crates.io,
RubyGems, Docker Hub, Homebrew, Reddit, dev.to, Keybase — and reports back which
corners of the internet are still yours to claim. The answers stream in live, one
cell at a time, so the page fills itself in front of you instead of spinning on a
loading screen. When it settles, you can walk away with a PDF brand report.

**It's live at [katanaid.com](https://katanaid.com).**

## How it holds together

The interesting part is the fan-out. A naive version would check each platform in
sequence and take the better part of a minute; KatanaID launches every check as its
own goroutine and lets the results race back through a channel, streaming each one to
the browser over server-sent events the moment it resolves. A single generation ends
up orchestrating dozens of external API calls in the time the slowest one takes to
answer.

Wrapped around that is a trust engine. Before anything runs, the browser quietly
fingerprints itself — screen, timezone, a canvas signature, hardware hints — and the
server folds that together with what it can see for itself (the user-agent, whether
you're signed in, where the request came from) into a 0–100 score. It's transparent
on purpose: every point comes with a reason, so the number can be read and tuned
rather than trusted blindly.

Underneath, Postgres is driven through Ent. The schema is written once as typed Go
and everything else — the query builders, the migrations — falls out of it, so the
database can't drift away from the code. Adding a field is a schema edit and a
regenerate, not a hand-written migration hoping to match a hand-written struct.

There's a companion `server-depr/` in the repo — an earlier cut of the backend, kept
around for reference rather than deleted.

## Running it yourself

You'll need Go 1.25+, Node with pnpm, and a local Postgres.

The backend lives in `server/`. Copy `.env.example` to `.env`, fill in a database
URL, a Resend key for the sign-in emails, a Gemini key, and your OAuth credentials,
then start it:

```bash
cd server
go run ./cmd/server
```

Ent migrates the schema on boot, so there's no separate setup step. If you don't have
a Gemini key handy, it falls back to a local name generator and everything else keeps
working. After changing anything under `internal/db/ent/schema`, regenerate the client
with `go generate ./internal/db/ent/...`.

The frontend lives in `client/`. Point `VITE_API_URL` at the backend and run it:

```bash
cd client
pnpm install
pnpm dev
```

The app comes up on `localhost:5173`, the API on `localhost:8080`, and from there
you just describe something and watch it fill in.
