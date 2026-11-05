# Review — S14-33 A bridge between two handles files when it is made

**Verdict:** (round 2 pending) — round 1: Fits with changes — Not ready
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- D-1 / D-2: filing runs inside `connect.CreateCitedBridge` through `identityclaims.AppendFiling` on the new bridge's id. It does not run on the two endpoint ids, so a connect write never files older unfiled bridges. Those stay with Promote batch Done.
- D-3: every bridge kind files on create, not only relationships. The working rule says so; the dogfood note named relationship only.
- D-4 / D-5: a wait or refusal is silent. There is no toast, no proto field, and no Swift or FakeStore change; membership shows through the graph reload the effects already trigger.
- Out: no back-fill of older catalogs (Promote Done on the graph files them); no change to Unpromote; no filing when a bridge's Observations are edited later.

## Round 1

Executability findings are summarised; maintainability findings are summarised with their evidence.

### X-1 · major · Re-seed helper looks up a type id it already has
`subjects.Subject.SubjectTypeID` is on the value. The plan also did not name the `writes.Call` return type or the imports.
**Answer:** Fixed. STEP-2 uses `s.SubjectTypeID` and returns `[]byte`; superseded by M-2, so the helper is now `promotetest.ClaimWithoutFiling`.

### X-2 · blocker · T-6 has no revision source in `bridgeWorld`
**Answer:** Fixed. M-6 moved T-6 to `batch_test.go` as `TestSaveBatchKeepsABridgeFiledOnCreate`, which uses `batchFixture`'s `save`, `revision` and `bridgeClaims`. No new revision helper is needed.

### X-3 · blocker · STEP-4 / T-7 relies on relationship helpers that do not exist
**Answer:** Fixed. T-7 uses `connectFixture`'s participation request, which needs no new helpers (same as M-5). STEP-4 now says how `dir` is captured by a closure, how both ends are promoted (`promoteInDir`), and how the bridge's claim is read in `after` (`withProjectCatalog` + `identityclaims.AcceptedEntityForSubject`).

### X-4 · major · T-7 assertion is weak
**Answer:** Fixed. `effects.handles` must contain `database.UUIDString` of the bridge's own claim entity id.

### X-5 · minor · STEP-5 rewrites only part of the §9.1 "When:" bullet
**Answer:** Fixed. Only the first sentence changes; the bullet's other sentences stay.

### X-6 · major · The "only two tests fail" check could not be re-run
The writer ran the full `go test -tags fts5 ./core/... ./api/...` with the hook applied at `c8a8bdb`. Only the two batch tests failed.
**Answer:** Fixed. STEP-1's check is now the full suite. The Stop condition says to report any other failure rather than re-seed it.

### X-7 · minor · Xcode step cannot run off macOS
**Answer:** Fixed. The DoD check-pr line says to report `xcode-test` as not run off macOS (this PR changes no Swift).

### X-8 · minor · Map row does not link the plan
**Answer:** Fixed in the wiring step. The map row's Plan column links this plan, and STEP-7 sets status `Landed in #N`.

### X-9 · nit · Smaller choices
**Answer:** Logged. T-2 now names `w.term(w.prop("place_relationship_type"), "part_of")`.

### M-1 · major · Two domain docs and the package doc are made false
`docs/conclusion-layer-data-model.md` §5.4 ("happens on Done … or later, when a claim makes its second end a member"; "never fails the claim that triggered it"), `docs/audit-revision-history.md` `create_cited_bridge` row, `core/database/connect/connect.go` package doc.
**Answer:** Fixed. STEP-5 gives the exact edits for both docs. §6 Go core rewrites the package doc. The DoD lists all three.

### M-2 · major · `claimWithoutFiling` re-spells an existing claim step; put it in `promotetest`
`promote.fileSubject`; `core/database/effects/conclusion_handles_test.go` `claim` closure and the inline copy; `promotetest` "so tests do not each spell the writes.Call block".
**Answer:** Fixed. `promotetest.ClaimWithoutFiling` *(new)* (D-6, STEP-2, Codebase fit). Moving the effects test's copies onto it is an Out line.

### M-3 · minor · Filing doc comments do not describe the bridge-id input
**Answer:** Fixed. `AppendFiling` and `FileBridgesTx` doc comments gain the sentence (§6 Go core, STEP-1 Files).

### M-4 · minor · The import-cycle claim is false
`go list -deps ./core/database/promote` has no `core/database/connect`; only `promote_test` files import it.
**Answer:** Fixed. §4 "Tests — connect" and D-7 now give the real reason: `bridgeWorld` and `TestBridgeFiling` live in `promote`.

### M-5 · minor · T-7 builds a relationship fixture the handler tests lack
**Answer:** Fixed with X-3. AC-1.4 now reads "Same as AC-1.2".

### M-6 · nit · T-6 re-spells a batch call `batchFixture` already has
**Answer:** Fixed with X-2.
