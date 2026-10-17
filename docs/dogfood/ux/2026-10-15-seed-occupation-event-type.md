# Seed an "Occupation" event type

- **Date:** 2026-10-15
- **Type:** Feature request / seed data
- **Where:** Event types
- **Annoyance:** Occupation is one of the most common facts a census, directory, or civil record gives, and the Properties page copy already steers researchers toward recording it as an event ("Facts that change over time, like occupation or residence, usually belong in events"). The seeded event types stop at vital events plus census / residence / migration / travel, so every project has to mint its own `Occupation` term first.
- **Wanted:** Ship `occupation` as a product-seeded `event_type` term alongside `residence`, so a new project can record an occupation event without adding vocabulary first.

## Code notes (observed, not implemented)

- `core/database/subjectvocab/registry.go` — `seedTerms` holds the product `event_type` terms (`birth` … `travel`); an `Occupation` entry would sit next to `residence`. `Install` upserts them with `propertyterms.OriginProvenencia`.
- `docs/seeded-vocabulary.md` §3.4 — the horizon list of `event_type` keys; needs the same addition. Title composition ("Occupation of {A}") follows the generic subject-count rule, so no title work is implied.
- `core/onboarding/ready.go` — `subjectvocab.Install` runs only at catalog create time (per `.cursor/rules/seeded-vocabulary.mdc`, open does not heal vocab), so existing projects would not pick the term up; researchers there keep using an `origin=user` term unless a separate backfill is decided.
- Follow `.agents/skills/add-seeded-vocabulary/SKILL.md` when implementing.
