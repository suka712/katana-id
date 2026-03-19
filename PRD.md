# KatanaID — PRD

## User Flow

1. User lands on homepage, sees a single prompt input
2. User types anything — a name, a description, or both
   - `"ruffle"` (→ user has a name) → check that name + suggest variants
   - `"tinder for dog lovers"` (→ user doesn't have a name) → generate names, check all
   - `"tinder for dog lovers called ruffle"` → check ruffle first, suggest alternatives
3. User hits "Check" → navigates to `/generate?q=<prompt>`
4. Results page loads, skeletons show while backend works
5. Names appear as cards, each with availability badges:
   - `.com` domain
   - GitHub org
   - npm package
   - Twitter/X handle
6. User browses results, picks a name they like

## Auth Gate (later)
- Saving / favoriting names requires sign in
