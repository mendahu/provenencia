# Review — S12-19 One typed spelling of the omnibar's hit kinds

**Verdict:** Fits with changes / Not ready, then fixed. **Rounds:** 1. **Reviewers:** maintainability (fresh agent), executability (fresh agent).

## To user

None.

## Writer's own calls

- `CatalogSearchHit.kind` stays a `String`; the typed view is `searchKind` (D-2). The request side (`searchCatalog(kinds:)`) stays `[String]` (Out).
- The plan and the code ship on the session's assigned branch in one pass, at the user's request; round 1 ran while the first steps were in the tree.
- No round 2: no finding was declined, and the fixes follow the reviewers' own "Better shape".

## Round 1

### M-1 · major · Link the handle cases to `EvidencePrimaryKind`
Go defines handle kinds as `vocabkey` primary kinds; D-5 added a third hand-written list.
**Answer:** Fixed, in part. D-5 adds `CatalogSearchKind.init(_: EvidencePrimaryKind)`; `primaryKind` was dropped once the lead switched per case (it had no reader); FakeStore derives `handleSearchKinds` from `EvidencePrimaryKind.allCases` and switches decoded primary kinds exhaustively; the lead reads `markKey`.

### M-2 / X-3 · minor / major · Two lead switches
**Answer:** Fixed in D-3. `leadThumbnail` returns `PVThumbnail.Content?`, nil for `.source`; the row is `if let` with no switch of its own.

### M-3 · minor · The rule's wording is already false
**Answer:** Fixed. STEP-5 rewords `macos-client.mdc` to name `swift_vocabulary_test.go` as the index of pinned enums, and the maintainability lens Q2 row follows.

### M-4 · minor · "Next adds" overclaims
**Answer:** Fixed. D-7 makes the fake priority an exhaustive switch; "Next adds" says fake hits are a checklist line, not a compiler check; the skill gets that line.

### M-5 / X-1 · nit / blocker · STEP-1 and STEP-2 already in the tree
**Answer:** Fixed. The header says the plan and code ship together on the assigned branch. The Go constant is `swiftSearchKind` in the plan, and `searchKinds()` sits above the `sortedRelations` comment.

### X-2 · blocker · STEP-3's check catches match-field codes
**Answer:** Fixed. The check is scoped to `kindLabel`, `leadThumbnail` and the row; the match-field codes are named as a different vocabulary.

### X-4 · major · FakeStore types and the unknown-kind path
**Answer:** Fixed in D-5 and D-6: parameter types, exhaustive handle switches, skip on an undecodable membership kind, and the all-unknown request returning nothing.

### X-5 · major · Swift checks need macOS
**Answer:** Fixed. STEP list says Swift evidence off macOS is the `macOS tests` workflow on the non-draft PR.

### X-6 · minor · Test names
**Answer:** Fixed. T-2–T-4 name their structs and tests; the §4 Tests row is corrected; the `widget` line moves into T-2.

### X-7 · minor · Skill section names
**Answer:** Fixed. STEP-5 names `## Checklist` and `### 4. FFI + Mac store`.

### X-8 · minor · Plan fails its own lint
**Answer:** Fixed. This log is linked; status is In progress.

### X-9 · nit · Extension's file
**Answer:** Fixed in D-2.
