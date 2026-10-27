# Review — S12-15 Graphs in the omnibar

**Verdict:** pending round 2 — round 1: Fits with changes — Ready with changes
**Rounds:** 1 (round 2 running). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **M-4 — where a graph hit's Source comes from.** The writer recommended reading it at search time (no schema change). The user chose projecting it onto the search document (2026-10-27): "a more robust long-term solution". The plan now adds migration `000056` with two parent display columns (D-13), fans a Source rename out to its graphs (D-3), and drops the post-cut fill. Answered.

## Writer's own calls

- D-7: every graph is indexed, the Default graph of a single-graph Source included, so `default` lists one hit per Source.
- D-8: a graph hit opens the canvas even when the graph links no Artifact; the canvas shows its existing "no artifact" callout.
- AC-1.4: graph hits match their own name, description and ref only, never their Source's title.
- D-13: the parent columns are generic (`display_parent_id`, `display_parent_ref`), not Source-named, and the parent title rides in `display_subtitle`.
- X-1 declined: the implementer runs on the Mac (`ship-pr`), so proto regeneration and Xcode checks stay in the steps; their absence is a stop condition.

## Round 1 — maintainability

### M-1 · blocker · A second Go builder of the graph-canvas location, with a different Ref/Title rule
**Plan:** D-6 builds the graph location with the Source's ref and title; the Out row sends `deleteimpact.projectEvidenceGraph` (graph ref and name) to the dogfood log.
**Codebase:** `core/database/deleteimpact/projectors.go` `projectEvidenceGraph`; Swift callers use the Source title.
**Better shape:** Align `projectEvidenceGraph` to the Source title and ref in this PR, with a test pinning both producers.
**Answer:** Declined with evidence. There is no single rule to align to today: `projectSubject` in the same file also builds a graph-surface location whose `Ref` / `Title` are the Subject's. Changing only `projectEvidenceGraph` makes a third variant, and both live in `deleteimpact`, which this PR does not otherwise touch. The canvas never reads `Title` except as the breadcrumb's fallback behind the session's Source title. Out row now homes it to a suggested follow-up task that settles all of `deleteimpact`'s graph-surface projectors in one pass.

### M-2 · blocker · Graph → Source read written in `core/search` with hand-rolled `IN (…)`
**Plan:** D-6's `fillEvidenceGraphHits` query in `core/search`.
**Codebase:** `evidencegraphs.SourceOf`, `core/database/sqlbatch` `Query`; working rule "One graph list".
**Answer:** Fixed by the M-4 reshape. The fill is gone: the projector (`ReprojectEvidenceGraph`, D-4) reads the graph's Source in one row, the way `ReprojectSource` reads its own tables, and no by-id IN-list is added. STEP-7 rewrites the "One graph list" rule so it no longer says the omnibar location reads the list.

### M-3 · major · Three new kind-id literals when the kind already has a declared owner
**Answer:** Fixed — D-1: all three constants are `string(catalogmodel.KindEvidenceGraph)`; Codebase fit row updated.

### M-4 · major · D-2 rejects projecting the Source onto the document for a reason the registry does not bear out
**Answer:** To user → the user chose projection (above). Fixed — D-2, D-3, D-4, D-6, D-13; Schema section; STEP-1 (new); T-4 (orphan drop) removed; the rationale no longer calls the fan-out costly.

### M-5 · minor · "Next adds" names a future the skill forbids
**Answer:** Fixed — Next adds: "None named in the map", plus how a parent-scoped kind would use the new columns.

### M-6 · minor · The seventh search-kind string literal across the Swift switches
**Answer:** Declined for this PR (the six existing kinds share it). Out row homes a typed `CatalogSearchKind` enum pinned to `search.Registry` to a suggested follow-up task.

### M-7 · nit · FakeStore rebuilds the list button's location by hand
**Answer:** Fixed — D-11: `SourcesListNavigation.canvasLocation(for:graph:)` *(new)*, ungated, called by `graphLocation` and FakeStore; T-7 compares against it.

## Round 1 — executability

### X-1 · blocker · STEP-3/4/5 checks cannot be run here (no protoc, no Xcode)
**Answer:** Declined with evidence: `scripts/check-pr.sh` says xcode-test is macOS only and fails elsewhere by design, and `ship-pr` runs the implementer on the Mac. Added two stop conditions: `generate-proto.sh` cannot run or regenerates more than comments; no macOS toolchain.

### X-2 · major · T-1/T-2 fixture cannot set the graph description
**Answer:** Fixed — §8 fixture creates Jones through `evidencegraphs.Create` with `Description`, names `artifactstest.Create` for "Page 2", and names the `writes.Run` actions.

### X-3 · major · FakeStore `refBoost` leaves the score floor and skip rule open
**Answer:** Fixed — D-11 states the callers' floor, reason and skip rules, and where the graph loop sits.

### X-4 · major · `add-searchable-kind` has two priority lines
**Answer:** Fixed — STEP-7 edits both; the location-table row uses the raw `sources` section.

### X-5 · minor · `fillEvidenceGraphHits` details
**Answer:** Fixed, then superseded by M-4 (the function is gone; D-4 formats the Source id with `database.UUIDString`).

### X-6 · minor · T-3/T-4 setup unspecified
**Answer:** Fixed — T-3 uses `db, _ := c.DB()`; T-4 was removed with the fill.

### X-7 · minor · T-7/T-9 assertion form
**Answer:** Fixed — the VoiceOver test (now T-6) compares against `ListFormatter`; T-7 seeds the fake by assignment. Back / Forward and relaunch stay manual (AC-2.2, AC-2.3), since the history code is shared with the list button and unchanged.

### X-8 · nit · Doc and DoD wording
**Answer:** Fixed — manual check order; STEP-7 states the map row text.
