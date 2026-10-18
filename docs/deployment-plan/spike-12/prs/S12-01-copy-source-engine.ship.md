# Ship log — S12-01 Copy Source engine

**Plan:** [`S12-01-copy-source-engine.md`](S12-01-copy-source-engine.md) · **Branch:** `s12-01-copy-source-engine` · **PR:** #478
**State:** fixing 2

Worktree: `/Users/jakerobins/Repos/provenencia-s12-01`, from `main` @ `998306f9`.

## Round 0 — implement
Head `141890cf`. Gate PASS (orchestrator run). PR #478 (draft). Report deviations: `call_test.go` unknown-method probe moved 102→103 (102 is now `METHOD_COPY_SOURCE`); `sourcecredibility/assessments.go` gofmt'd (pre-existing on main, whitespace only, own commit); `TestCopy/all_parts` compares 2 visible workspace entries plus 3 raw layout rows; fixture helper `must` panics instead of `t.Fatal`.

## Round 1 — review
Base `998306f9`, head `141890cf`. Five lenses. PRF and SEC clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | A new metadata value is trimmed and normalized twice; `Set` and `insertValueTx` both own the rule | Fix now → F-1 | verified: `sourcemetadata.go` `Set` normalizes before `if creating`, then `insertValueTx:170` normalizes again. D-6 puts trim + `normalizeValue` in `insertValueTx`, so the fix keeps D-6 |
| CON-1 | CON | minor | `FakeStore.copySource` reports no evidence effect when Artifacts are copied; the engine does | To user → Fix now → F-3 (user, 2026-10-18) | verified: `effects/registry.go:184-189` maps `artifacts` to `Source` + `CitationSource`; `FakeStore.swift:412` `bumpSource(copyID, evidence: false)`. Fix contradicts D-15's literal text, so it goes to the user. Recommend `evidence: !chosen.isEmpty` |
| CON-2 | CON | minor | `TestCopy/other_type` counts metadata values but does not compare them (AC-2.4) | Fix now → F-2 | read the test: only `len(values) != 2`. Rides along: AC evidence in a test file this PR adds, a few lines |
| MNT-3 | MNT | minor | `must[T]` panics instead of `t.Fatal`; comment misdescribes it | Fix now → F-2 | read `sourcecopy_test.go:49-54`. Same new test file as CON-2 |
| MNT-4 | MNT | minor | `_ = i` dead binding in `seedObituary` | Fix now → F-2 | same file |
| CON-3 | CON | nit | `TestCreateAll` never hits the layout-only branch of `sqlHasAny` | Fix now → F-1 | rides with F-1 in `sourcemetadata` (one table row) |
| MNT-2 | MNT | minor | Third copy of the repeated-id parse loop in handlers | Follow-up | fix plan does not touch handlers; `parseIDs` helper over three callers → idea or next handler PR |
| MNT-5 | MNT | minor | Another private `writeTinyPNG` copy | Follow-up | plan STEP-3 asked for the copy; shared catalog test-fixture package → `docs/ideas/shared-list-helpers.md` |
| UX-1 | UX | minor | Stale form (original or Artifact deleted in another window) fails as generic "Invalid source." | Follow-up | D-8 binds no new code here; S12-02 plan should handle stale-form refresh or copy |
| MNT-6 | MNT | nit | Skill paragraph cites `ListLayoutTx` as having a Catalog twin; it has none | Logged | |
| UX-note | UX | nit | Fake silently drops unknown/foreign Artifact ids and accepts a blank title, where the engine throws | Logged | matches D-15 as written; S12-02 model tests cannot exercise the copy error path through the fake |

### Fix plan 1

**F-1 — One owner for trim + normalize of a new metadata value** (from MNT-1, CON-3)
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`.
Failing test first: none for MNT-1 (structural, no behavior change; existing `Set` and `TestCreateAll` tests guard it). CON-3: add a `TestCreateAll` row that seeds the Source with `runDismissSuggestion` only (no value), then expects `ErrInvalid` from `runCreateAll`.
Do: in `Set`, move the `normalizeValue(field.DataType, in.ValueText)` call so it runs only on the update branch (after the `if creating { … }` block, before `prev.ValueText == in.ValueText`). The create branch passes the trimmed `in.ValueText` to `insertValueTx`, which stays the one place that trims and normalizes a new value (D-6). Keep `Set`'s top-of-function `TrimSpace`. Make `insertValueTx`'s doc comment true as written.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcemetadata/...` passes; scripts/check-pr.sh PASS.

**F-2 — Tidy the `sourcecopy` tests** (from CON-2, MNT-3, MNT-4)
Files: `core/database/sourcecopy/sourcecopy_test.go`.
Failing test first: none (test-only).
Do: (a) In `TestCopy/other_type`, map both Sources' `sourcemetadata.ListBySource` rows to `{FieldID, ValueText}` and `reflect.DeepEqual` them, like `artifactKeys` does for Artifacts. (b) Replace the panicking `must[T]` with checks that fail the test through `t` (`t.Helper(); t.Fatal(err)`), e.g. a `must[T any](t *testing.T, v T, err error) T`, and update its callers; fix its comment. (c) In `seedObituary`, `for _, spec := range` and drop `_ = i`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcecopy/...` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `b3dd61c6`. Gate PASS (orchestrator run). F-1 done (`35a6de92`), F-2 done (`b3dd61c6`). Deviation: F-2(b) removed `must` and inlined `t.Fatal` checks, since Go cannot pass a multi-value call plus `t` to one helper. CON-3's new row passes at once (guards existing behavior; not shown red).

## Round 2 — review
Scope `141890cf...b3dd61c6`. Lenses: MNT (only lens with a Fix-now finding; CON-2/CON-3 were riding minors).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | — | Resolved | — | `normalizeValue` call sites: `Set` update branch and `insertValueTx` only |
| MNT-3, MNT-4, CON-2, CON-3 | MNT | — | Resolved | — | read by the round-2 reviewer |

No new findings. No Fix plan 2.

## To user — answered
CON-1: the user said: "Yes: update D-15 and change FakeStore.copySource to bumpSource(copyID, evidence: !chosen.isEmpty) in this PR."

### Fix plan 2 (user-approved, not a review round)
**F-3 — The fake reports the evidence effect when it copies Artifacts** (from CON-1)
Files: `macos/App/Platform/FakeStore.swift`, `docs/deployment-plan/spike-12/prs/S12-01-copy-source-engine.md` (D-15 row only).
Failing test first: none (FakeStore unit tests are out of scope for this PR per the plan; S12-02 covers the model).
Do: in `FakeStore.copySource`, `bumpSource(copyID, evidence: false)` → `bumpSource(copyID, evidence: !chosen.isEmpty)`, the same rule `createArtifact` uses. In the plan's D-15 row, replace "returns `bumpSource(copyID, evidence: false)`" with "returns `bumpSource(copyID, evidence: !chosen.isEmpty)` (the engine's `artifacts` effects entry reports `sources` and `citation_sources`; changed after review, CON-1)". Add one line to the plan's Landed note naming this change.
Check: scripts/check-pr.sh PASS.

## Result
Gate PASS · CI pending · open: none blocker/major · follow-ups: MNT-2, MNT-5, UX-1 · to user: CON-1 (fake evidence effect vs D-15) · logged: MNT-6, UX-note
