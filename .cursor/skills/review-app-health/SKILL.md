---
name: review-app-health
description: >-
  Runs a periodic whole-app implementation and maintainability review of
  Provenencia (Go core, FFI, macOS SwiftUI) against a fixed checklist: code
  cleanliness, security, performance, project structure, separation of concerns,
  idiomatic patterns, test coverage, error UX, internationalization, docs/skills/rules
  drift, accessibility, and UI component organization (orphans, near-duplicates,
  generic vs domain-specific). Use when the user asks for an app health review,
  tech-debt review, maintainability audit, bloat check, architecture hygiene
  pass, or to periodically review the application for debt creep—especially
  after AI-assisted development.
---

# Review app health

Whole-app **snapshot** review of Provenencia as it exists now—not a PR/diff
review. Goal: catch bloat, tech-debt creep, and structural drift early.

Do **not** fix findings unless the user explicitly asks. Do **not** bump
`VERSION`. Do **not** invent work outside the checklist.

Authoritative stack notes: [`docs/application-stack.md`](../../../docs/application-stack.md),
[`docs/macos-client-patterns.md`](../../../docs/macos-client-patterns.md).
Detailed criteria: [`criteria.md`](criteria.md).

## When invoked

1. Confirm scope if unclear (default: **whole app** — `core/`, `api/`, `macos/`).
   Accept focus slices (`macos` only, `core/search`, one feature) when asked.
2. Copy the progress checklist below and work it top to bottom.
3. Explore with Grep/Glob/Read and light tooling (see Toolset). Prefer evidence
   over vibes: cite `path` (and line ranges when useful).
4. Deliver a **chat report** in the Output format—primary deliverable is the
   numbered action-item list. Write `docs/reviews/YYYY-MM-DD-app-health.md`
   only if the user asks for a file.
5. Ask which action items to take next—do not start fixing unprompted.

## Progress checklist

```
App health review:
- [ ] 1. Code cleanliness
- [ ] 2. Security
- [ ] 3. Performance
- [ ] 4. Project structure & naming
- [ ] 5. Separation of concerns
- [ ] 6. Idiomatic patterns
- [ ] 7. Test coverage
- [ ] 8. Error handling & messaging UX
- [ ] 9. Internationalization
- [ ] 10. Docs / skills / rules drift
- [ ] 11. Accessibility
- [ ] 12. UI component organization
- [ ] Report delivered
```

## Toolset (use what fits)

| Need | Approach |
| --- | --- |
| Map layout | `Glob` / directory listing of `core/`, `api/`, `macos/App/` |
| Dead / vague names | Grep for `common`, `shared`, `util`, `utils`, `helpers`, `misc`, `temp` |
| Hard-coded UI copy | Grep SwiftUI `Text("…")`, `Button("…")`, raw English in `macos/App` (exclude Generated, previews if clearly fixture) |
| Error surface | Trace `apperr` → FFI protobuf codes → `L10n.Errors` → toast/inline UI |
| Go tests | `CGO_ENABLED=1 go test -tags fts5 ./...` (or focused packages); note skips/failures |
| Swift tests | Note ProvenenciaTests coverage gaps; run `xcodebuild test` only if practical / user wants |
| Imports / cycles | Inspect Go import graph for cross-domain edges; Swift feature→feature coupling |
| Security (recent delta) | Optional: `/review-security` / security-review subagent for **branch/uncommitted** diffs—not a substitute for whole-app §2 |
| Perf hotspots | Catalog session misuse, N+1 queries, SwiftUI `body` work, sync FFI on UI path, image/derivative work, search/FTS |
| Docs / skills drift | Spot-check `.cursor/skills/`, `.cursor/rules/`, and key `docs/*.md` against the live tree; flag broken paths and contradicted “blessed” patterns |
| Accessibility | Grep `.accessibilityIdentifier` / labels; sample interactive controls (esp. icon-only); compare to `docs/macos-client-patterns.md` §5 |
| Component reuse | Sample `DesignSystem/Components/**` + hot feature chrome; flag orphans/near-duplicates. Deep dive on one named type → [`evaluate-ui-component`](../evaluate-ui-component/SKILL.md) (do not expand §12 into full per-type plans here) |

Keep commands read-only unless the user asked to fix. Prefer sampling deeply in hot paths over exhaustively listing every file.

## Dimension summary

Read [`criteria.md`](criteria.md) for the full rubric. In short:

1. **Cleanliness** — unused imports, dead code, redundant wrappers, outdated comments, low-hanging duplication.
2. **Security** — FFI/path/SQL/ingest/identity/entitlements; no secrets; least privilege; safe defaults.
3. **Performance** — bottlenecks and concrete speedups (catalog session, queries, UI main-thread, derivatives, search).
4. **Structure** — semantic folders/files; reject vague `common` / `shared` / `project` dumps; Mac: `Features/<Name>/`, `Platform/`, `DesignSystem/`.
5. **SoC** — clear dependency tree; no cross-domain cycles; shared code hoisted upward; extract reusable utilities when justified.
6. **Idiomatic** — go with Go / SwiftUI / SQLite / project skills grain; replace bespoke band-aids with blessed patterns.
7. **Tests** — Go owns domain/SQLite; Mac models via FakeStore; gaps and brittle English asserts.
8. **Errors** — specific, context-aware, placed where the user acts (inline / toast / floated)—not generic or silent swallow.
9. **i18n** — no hard-coded user-facing strings; `L10n` + catalogs; FFI codes mapped under `L10n.Errors`.
10. **Docs/skills/rules drift** — authoritative guidance still matches code; no orphan skills, stale rules, or contradicted docs that would mis-train the next agent.
11. **Accessibility** — VoiceOver/keyboard-ready controls; stable dotted `accessibilityIdentifier`s; labels on icon-only actions; no UI-testing by localized title.
12. **UI component organization** — orphans, near-duplicates, and wrong Frost layer (DS vs recipe vs snowflake); sample and flag; for a pointed deep-dive + compose plan use [`evaluate-ui-component`](../evaluate-ui-component/SKILL.md); for new chrome use [`add-ui-component`](../add-ui-component/SKILL.md).

## Provenencia invariants (flag violations)

- Genealogy / SQLite / identity live in **Go**, not Swift views.
- Mac UI: thin SwiftUI → model → `GenealogyStore` → FFI → core.
- Catalog access: held session + queue (`.cursor/skills/use-catalog-session`), not open-per-RPC.
- User copy: typed `L10n` + String Catalogs (`.cursor/skills/add-localized-string`).
- Errors: stable `apperr` / protobuf codes → `L10n.Errors.message`—do not match English Go strings in Swift.
- Prefer product-named packages/folders over grab-bags (`Shared/`, `Common/`, `Utils/`).
- Skills/rules/docs that agents follow must stay truthful; fix or archive drift, don’t leave lying guidance.
- Interactive Mac controls that matter for tests or AT get stable `.accessibilityIdentifier("dotted.name")`; don’t query by localized title.
- DesignSystem `PV*` is content-agnostic; recipes are product-specific reused maps onto `PV*`; snowflakes stay feature-private—not a second copy of the same floating menu / list / chip / dialog chrome.

## Output format

The review’s main product is a **numbered list of action items**: each item is the
**smallest meaningful change** that still delivers value—roughly the smallest PR
worth opening. Prefer splitting a large cleanup into several numbered items over
one mega-item.

Number items **globally** across priority groups (do not restart at 1 in each
section) so the user can say “do 3 and 7.”

Each action item **must** include:

- **Problem** — what is wrong, with evidence (`path` / lines)
- **Consequence** — what happens if we leave it
- **Benefit** — what we gain by fixing it (and a one-line sketch of the change)

```markdown
# App health review — YYYY-MM-DD

## Verdict
[2–4 sentences: overall health, biggest risks, whether debt is creeping]

## Action items

### High priority
1. **Short title**
   - **Problem:** …
   - **Consequence:** …
   - **Benefit:** …

### Medium priority
2. **Short title**
   - **Problem:** …
   - **Consequence:** …
   - **Benefit:** …

### Low priority
3. **Short title**
   - **Problem:** …
   - **Consequence:** …
   - **Benefit:** …

## Dimension notes
| Dimension | Status | Notes |
| --- | --- | --- |
| Cleanliness | OK / mixed / weak | … |
| Security | … | … |
| Performance | … | … |
| Structure | … | … |
| Separation of concerns | … | … |
| Idiomatic patterns | … | … |
| Test coverage | … | … |
| Error UX | … | … |
| Internationalization | … | … |
| Docs / skills / rules drift | … | … |
| Accessibility | … | … |
| UI component organization | … | … |
```

Omit an empty priority section rather than writing “None.”

Rules for action items:

- Prefer **few high-signal** items over exhaustive nit lists.
- One item ≈ one small PR: single concern, reviewable diff, clear done state.
- Do **not** bundle unrelated fixes; do **not** invent work outside the checklist.
- Cite evidence in **Problem**; keep **Benefit** concrete (what lands in the PR).
- For UI organization items that name a specific type, **Benefit** may say “run evaluate-ui-component on X” rather than embedding a full compose plan in this report.
- Call out **false alarms** you considered and dismissed when useful.
- If a dimension looks healthy, say so briefly in the table—do not invent issues.
- Priority mapping: see [`criteria.md`](criteria.md) § Priority guide.

## Related skills (fix only, do not run unless asked)

- Pointed UI layering audit / compose-down plan: [`evaluate-ui-component`](../evaluate-ui-component/SKILL.md) — prefer this when an action item names a specific `PV*` or feature sheet
- New UI classification: [`add-ui-component`](../add-ui-component/SKILL.md)
- Diff security: Cursor `review-security` / security-review subagent
- L10n fixes: `add-localized-string`
- Catalog session: `use-catalog-session`
- Tests: `add-swift-test`; Go tests per `.cursor/rules/go-tests.mdc`
- Structure/nav: `add-workspace-location`, `add-ffi-handler`, `add-catalog-query`
