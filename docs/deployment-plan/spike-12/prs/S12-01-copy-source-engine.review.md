# Review — S12-01 Copy Source engine

**Verdict:** Fits with changes — Ready with changes (round 1)
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **M-9.** After this PR, Add Source (`CreateSourceResponse`, no `WriteEffects`, `SourcesModel.create()` patches the list by hand) and Copy Source (effects on the response) report a new Source two ways. Recommendation: add a small PR to the Spike 12 map after S12-02 that gives `CreateSourceResponse` `WriteEffects` and moves `create()` onto the same path, so both creates use one cache-update mechanism.

## Writer's own calls

- No new error code: every copy refusal is `sources.invalid` (D-8).
- Metadata copies verbatim through a new `sourcemetadata.CreateAll`, not through `Set` / `Reorder` / `DismissSuggestion` (D-4).
- `CopySourceResponse` carries `WriteEffects` (D-13).
- Swift store method uses flat parameters (D-14); `FakeStore.copySource` behaves like the engine (D-15).
- Method number 102.
- No version bump, as in Spike 11.

## Round 1

### M-1 · blocker · `ListLayoutTx` adds a second layout scan and type
**Answer:** Fixed — D-5: `LayoutRow` replaces the private `layout`; one `listLayoutRows` with `ORDER BY sort_order, field_id`; `listLayout` and `getLayoutTx` build on it.

### M-2 · blocker · `CreateAll` re-writes `Set`'s create branch
**Answer:** Fixed — D-6: `insertValueTx` extracted from `Set`; `Set` and `CreateAll` both call it. STEP-2 does the extraction first.

### M-3 · major · Draft made of row types with ignored fields
**Answer:** Fixed — D-9 and §6: `Draft` holds `sources.CreateInput`, `[]sourcemetadata.Input`, `[]LayoutRow`, `*sourcecredibility.UpsertInput`, `[]DraftArtifact{FromID, In artifacts.CreateInput}`. `CreateAll` takes `[]Input` and refuses a foreign `SourceID`.

### M-4 · major · New read pattern not written into `add-catalog-query`
**Answer:** Fixed — STEP-6 adds **Reads inside a write** to the skill; Scope In row updated.

### M-5 · minor · `GetTx` drops the length check; unnamed querier
**Answer:** Fixed — D-3: `GetTx` wraps `getTx` with the length check like `subjects.GetTx`. D-2: one unexported `querier` per package. Lower-case because nothing outside the package needs it.

### M-6 · minor · Wrong error code spellings
**Answer:** Fixed — §6 Errors uses `sourcemetadata.invalid`, `sourcecredibility.invalid`.

### M-7 · minor · `Copy` signature differs from `connect`
**Answer:** Fixed — D-1, D-7: `Copy(tx, userID, in CopyInput)` with `SourceID` inside.

### M-8 · minor · `getFieldTx` / `lookupField` duplicate in the same file
**Answer:** Fixed — D-2: folded into `getField(q querier, …)` in STEP-1.

### M-9 · minor · Create and Copy report writes differently
**Answer:** To user (above). Out row added.

### X-1 · blocker · Proto generation and Xcode cannot run on Linux
**Answer:** Fixed — D-16, §4 Toolchain row, and a Stop condition: STEP-4 / STEP-5 need the macOS toolchain; never hand-edit generated files.

### X-2 · blocker · Test fixtures not specified
**Answer:** Fixed — STEP-3 names `seedObituary` and every call it makes (types, fields with data types, suggestions, grades, Files, cover PNG, note, Subject). STEP-4 seeds T-9 through RPCs. AC-1.1 points at the STEP-3 fixture.

### X-3 · major · Missing-original mapping unassigned
**Answer:** Fixed — D-8: `ReadDraft` maps `ErrNoRows`; credibility `ErrNoRows` → none. D-2: Tx twins length-check.

### X-4 · major · `CreateAll` internals open
**Answer:** Fixed — D-4: empty input no-op, `sqlHasAny`, `sqlLayoutInsert`, unknown field → `ErrInvalid`, valueless layout rows allowed, empty value → `ErrInvalid`, change order.

### X-5 · major · Shared-function shapes half specified
**Answer:** Fixed — D-2 names the `querier` type and its methods; D-5 adds the `ORDER BY` to the one layout query. `sources` keeps `getTx` (D-3).

### X-6 · major · Input edge behavior undecided
**Answer:** Fixed — D-1 (nil tx, user check, returned Source), D-7 (title untrimmed to `Create`, bad type, credibility on with none).

### X-7 · major · FakeStore fields ambiguous
**Answer:** Fixed — D-15: Artifact ref `ART-FAKE1`, `thumbnailRelPath`, `hasArtifact`, cover defaults, stale `suggested` accepted.

### X-8 · major · STEP-6 under-specified
**Answer:** Fixed — STEP-6 names the map status, the plan status, and where the `completed.md` paragraph goes. The map link is added when the plan is Planned.

### X-9 · minor · Handler details
**Answer:** Fixed — §6 FFI: parse errors return before the catalog opens. T-9 separates a malformed id from a foreign one.

### X-10 · minor · Test helpers not pinned
**Answer:** Fixed — STEP-3 names `users.Upsert`, `sources.Count`, `sources.ListNotes`, `c.DB()` reads outside the write; T-1 names the `workspaceKeys` helper.

## Round 2

### X-11 · blocker · A third `GenealogyStore` conformer, `ThrowingStore`, is not in STEP-5
**Answer:** Fixed — verified `macos/ProvenenciaTests/OnboardingModelTests.swift:342`. Scope In, §4 Test store row, D-14, STEP-5 Files and §8 "Most likely to break" now name it; its body is `throw StoreBoom.boom`.

### X-12 · major · T-9 does not say it needs two Sources
**Answer:** Fixed — STEP-4: Sources A and B, one fileless Artifact each; the foreign case copies A with B's Artifact.

### X-13 · minor · T-1 and the cover case
**Answer:** Fixed — STEP-3: Artifact 1's File is the PNG and is the original's cover in every case.

### X-14 · minor · `getLayoutTx` and `FieldID`
**Answer:** Fixed — D-5: `FieldID` set from the argument.

### X-15 · minor · Test and doc details
**Answer:** Fixed — `runCreateAll` and `runCopy` wrappers named (STEP-2, STEP-3); the skill section goes between `## Functions` and `## Tests` (STEP-6). Status flips to Planned at the end of the review.
