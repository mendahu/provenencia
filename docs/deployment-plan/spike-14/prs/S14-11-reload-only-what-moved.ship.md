# Ship log — S14-11 Reload only what moved, only when it is on screen

**Plan:** [`S14-11-reload-only-what-moved.md`](S14-11-reload-only-what-moved.md) · **Branch:** `s14-11-reload-only-what-moved` · **PR:** —
**State:** implementing

## Round 0 — implement
Base `main` @ `fdd37f19`.
Stopped on the base check (implement-pr round 0 step 2), before any code change; no branch, no commit, no PR. `main` moved 42 commits past the Planned `85a5ce2`; S13-01 (#655) changed:
- `CatalogQueryDescriptor.refresh: CatalogQueryRefresh?` → `refreshes: [CatalogQueryRefresh]` (verified: `CatalogQueryRegistry.swift:42`); `CatalogQueryRefresh` is now `{ follows, action: .revalidate | .run }`.
- `applyHiddenPolicy(follows:id:)` → `applyHiddenPolicy(follows: CatalogQueryFollows, ids:)`, plus `followsValue:` and `applyHiddenPolicyOnAny` variants.
- `CatalogQueryFollows` gained `.touchedHandles` and `.evidenceGraphs`.
- `WorkspaceSession.apply` no longer calls `apply(invalidations:)`; it dedupes `invalidated + revalidatedKeys(for:)` and runs `.run` refreshes in their own loop. The registry no longer emits per-graph `.evidenceGraph` keys; that arm now revalidates via `refreshes: [.applyHiddenPolicy(follows: .evidenceGraphs)]`.
- Smaller: `.personClaims` key kind; `"sources"` effect now reports `EvidenceGraph` on title/source_type change; `listPersonClaims` on the stores.

Affected: §4 rows (patch/refresh slots, Follows, Session, progress refresh, graph list arm), D-5–D-8, D-10, D-11, STEPs 4–7, T-12, T-15. D-1–D-4, D-9 look unaffected. Next: plan refresh with `plan-pr` against `fdd37f19`, then resume this log.

Plan refreshed with `plan-pr` against `fdd37f19` (2026-11-01); `review-plan` rounds 3–4 in the review log, no open blocker. Resumed: round 0 again.
