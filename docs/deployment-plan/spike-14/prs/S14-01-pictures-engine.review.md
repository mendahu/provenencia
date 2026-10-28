# Review — S14-01 Pictures engine

**Verdict:** Fits / Ready
**Rounds:** 2. **Reviewers:** maintainability (claude-opus-5-5-high), executability (inherit)

## To user

- Spike 12 **S12-24** (Planned) edits `core/database/likenesses/likenesses_test.go` and `conclusion-layer-data-model.md` §6. S14-01 renames that package to `gallery` and rewrites the same §6 gallery/pointer prose. Recommendation: land S12-24 first when practical; the plan does **not** hard-Depends on it (no behavior dependency). S14-01 Risks/Stop preserve S12-24's preferred-delete assertions if they are already on `main`. Optionally add a one-line cross-link on the S12-24 plan that S14-01 moves `likenesses/`.

## Writer's own calls

- Unified header lead field name `PictureImageFileID` / `picture_image_rel_path` on Person, Event and Place (D-7, D-8), not Person-only `likeness_*` plus Event/Place `picture_*`.
- Keep `SetPreferredLikeness` RPC and column name (Out).
- Exact seed descriptions mirror `likeness` with kind noun (D-2).
- Declined hard-Depends on S12-24; Fixed via Risks/Stop/STEP notes instead (M-7).
- Declined seeded-depiction ≡ `GalleryKeys()` equality (round-2 note): Signature may be a depiction Property outside the gallery ([ideas/depictions-and-likenesses.md](../../../ideas/depictions-and-likenesses.md)).

## Round 1

### M-1 · blocker · STEP-3 copies the Person header tests per kind
**Answer:** Fixed — D-12; table-drive renamed lead tests over person/event/place rows.

### M-2 · major · Person header gallery-lead block becomes three copies
**Answer:** Fixed — D-7 `leadPictures` helper; all three compose paths call it; `placeShell` stays shell-only.

### M-3 · major · Handler path resolution becomes three copies
**Answer:** Superseded by M-12 in round 2 — pass-through `pictureRelPaths` dropped (D-11 rewritten).

### M-4 · major · T-2 second seed-vs-vocabkey test
**Answer:** Fixed — D-13 extends `TestVocabKeysMatchSeed`; separate `TestGalleryKeys` for order (also X-3). Round 2: stale §6 `TestGalleryKeysAreSeededDepictions` row removed.

### M-5 · major · `depiction-values.mdc` glob still points at likenesses
**Answer:** Fixed — STEP-7 and DoD update `.cursor/rules/depiction-values.mdc` glob to `gallery`.

### M-6 · major · Handler-test fixture hard-coded to Person
**Answer:** Fixed — D-14 `depictionFixture`; T-7 uses it.

### M-7 · major · S12-24 cross-spike package edit unnamed
**Answer:** Fixed (coordination) / Declined (hard Depends) — Depends line + Risks + Stop + STEP-2/7 notes; To user for optional S12-24 cross-link. Not a behavior dependency.

### M-8 · minor · D-5 missing SQLInPlaceholders
**Answer:** Fixed — D-5 uses `database.SQLInPlaceholders`.

### M-9 · minor · Place lead site + Scale search path
**Answer:** Fixed — D-7 forbids lead in `placeShell`; Scale names search re-index gallery cost.

### M-10 · minor · Domain doc still describes cache image for gallery
**Answer:** Fixed — STEP-7 SEC-4 clause in `conclusion-reconciliation.md` and skill.

### M-11 · nit · Leftover likeness names in renamed code
**Answer:** Logged — STEP-2 rewords `observationstest` package doc; other comment polish stays with the rename.

### X-1 · major · STEP-2 “tree compiles” claim false
**Answer:** Fixed — STEP-2 renames detail fields + handler field reads; does not claim `./api/ffi/...` green beyond type-check; STEP-5 owns FFI behavior tests.

### X-2 · major · Event/Place fixtures unspecified
**Answer:** Fixed in round 1 partially; completed by X-7 / D-12 rewrite (`eventSubject` / `placeSubject`, gallery `subjectOn`).

### X-3 · major · AC-1.3 order assert unnamed
**Answer:** Fixed — D-13 / T-2 `TestGalleryKeys`.

### X-4 · major · Host cannot run Xcode / generate-proto / go 1.27
**Answer:** Declined as plan blocker — same CI/Mac gate as other engine PRs; STEP-6 records Linux vs Mac evidence split.

### X-5 · minor · ConclusionDetailPage in STEP-6
**Answer:** Fixed — dropped; S14-02 owns that view layer.

### X-6 · nit · File basename after move
**Answer:** Fixed — STEP-2 names `gallery.go` / `gallery_test.go`.

## Round 2

### M-12 · major · `pictureRelPaths` is a pass-through (reopens M-3)
**Answer:** Fixed — D-11 drops the wrapper; three list handlers keep short collect+`files.RelPaths`+assign loops; `attachPictures` one RelPaths for lead+tiles. Codebase fit row removed.

### X-7 · blocker · D-12 event/place rows cannot call `SaveDepiction` via `f.event` / `f.place`
**Answer:** Fixed — D-12 / STEP-3 / T-5: `eventSubject` / `placeSubject` *(new)* return `(Subject, handle)`; existing `f.event` / `f.place` signatures unchanged.

### Round-2 notes (non-blocking)
- Gallery `subjectOn(src, kind)` — Fixed in STEP-2.
- Shadowing local `gallery` — Fixed in STEP-2 (`tiles`).
- Test names `TestHeaderLeadsWith…` — Fixed in D-12 / T-5.
- Seeded depiction ≡ `GalleryKeys()` — Declined (Signature may sit outside the gallery).
