# Review — S12-24 Audit history: indexed note authors, audited pointer resets

**Verdict:** Fits — Ready
**Rounds:** 2. **Reviewers:** maintainability (Claude Sonnet), executability (Composer)

## To user

None.

## Writer's own calls

- Kept index + pointer releases in one PR (map K-20 nit / scope entry).
- Chose `optionalPointer` beside `rowFacet` after round-1 M-2 (map said "two facetRelease entries"; the helper yields exactly two entries).
- Left `FK.Audited` CASCADE-only and keyed honesty on `BucketOptional` (round-1 M-3).
- Round-2 M-7: clear only the FK column and re-read side-effect columns so the `000015` trigger stays the owner of cover_mode → `type_icon` (no third hard-coded literal).

## Round 1

### M-1 · blocker · T-1 / D-2 is a migration test with a hard-coded format number
**Answer:** Fixed — D-2 / T-1 keep only `dbtest.QueryPlan` failing on `SCAN c`; AC-1.2 and format asserts removed; Out names the forbidden migration-presence test.

### M-2 · blocker · Cover and portrait clear changes built by hand; two releases copy each other
**Answer:** Fixed — D-4 `optionalPointer` beside `rowFacet`; two call sites; D-5 Fields parity against `SetCover` / `SetPreferredLikeness`.

### M-3 · major · `Audited: true` on `BucketOptional` cannot vary
**Answer:** Fixed — D-3 / D-6: optional FKs stay not Audited; honesty keys on `BucketOptional` `SET NULL`.

### M-4 · major · STEP-3 and sweep miss CASCADE-only docs
**Answer:** Fixed — STEP-3 lists catalog-deletes, both skills, model comment, facets comments, conclusion model. Widened again in round 2 (M-9).

### M-5 · minor · T-2 / T-3 placement left open
**Answer:** Fixed — D-9: extend existing artifacts cover-erase and likenesses preferred-delete subtests.

### M-6 · nit · Two note queries
**Answer:** Logged — Out leaves the fold; T-1 guards both SQL constants.

### X-1 · minor · Note attribution call chain
**Answer:** Fixed — Current state distinguishes `ListNotes` → workspace from `GetNote` single-note RPC.

### X-2 · minor · T-2 placement
**Answer:** Fixed — with M-5 / D-9.

### X-3 · minor · T-3 package
**Answer:** Fixed — likenesses only (D-9).

### X-4 · minor · Sibling order among KindArtifact releases
**Answer:** Fixed — D-4 places cover release immediately after Artifact `evidence_graph_artifacts` `rowFacet`.

### X-5 · minor · Plan-doc map link warning
**Answer:** Fixed — STEP-3 / DoD wires the map row (done at Planned).

## Round 2

### M-7 · major · `pointerReset{col, new}` copies the cover rule a third time
**Answer:** Fixed — D-4 clears only the FK, re-reads `sideEffectCols` (trigger owns cover_mode); signature is `sideEffectCols ...string`, no `pointerReset`.

### M-8 · minor · D-6 "narrow" wording wrong for current honesty
**Answer:** Fixed — D-6 only adds the `BucketOptional` branch; Risks row updated; no rewrite of `!Audited && has`.

### M-9 · minor · STEP-3 misses three CASCADE-only lines
**Answer:** Fixed — STEP-3 and DoD name catalog-deletes policy item 7, checklist item 5, and add-catalog-model L51.

### M-10 · nit · Next adds honesty for narrative pointer
**Answer:** Fixed — "when its parent becomes deletable".

### X-1 · minor · soft word `e.g.`
**Answer:** Fixed — Next adds sentence rewritten without `e.g.`.

### X-2 · minor · Observation optionalPointer list position
**Answer:** Fixed — D-4: immediately after `observation_notes` `rowFacet`.

### X-3 · minor · twin-row fixture shape
**Answer:** Fixed — D-5: two Sources / two handles in the same catalog.

### X-4 · minor · D-9 writes.Run wording
**Answer:** Fixed — D-9 uses `writes.Call` like the graph-links subtest.

### X-5 · nit · likenesses path shorthand
**Answer:** Fixed — full `core/database/likenesses/likenesses_test.go`.
