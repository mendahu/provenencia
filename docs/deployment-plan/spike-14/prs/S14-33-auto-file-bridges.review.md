# Review — S14-33 A bridge between two handles files when it is made

**Verdict:** Fits — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

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

## Round 2

No new blockers from either reviewer. Maintainability: Fits. Executability: Ready with changes. Both confirmed every round-1 answer.

### M-7 · major · The new §9.1 sentence says a connect files every bridge whose ends are handles
That would teach the shape D-2 rejects.
**Answer:** Fixed. STEP-5 now says that bridge is filed.

### M-8 · major · The Out row names a move the helper cannot take
The effects test copies run inside an open `*sql.Tx`.
**Answer:** Fixed. The Out row now names the real follow-up: an in-transaction step shared by the helper and those copies. Not done here, because this PR does not touch `effects` tests.

### M-9 · minor · STEP-3 re-spells `bridgeWorld.filed`
**Answer:** Fixed. STEP-3 uses `w.filed(s)` (`unpromote_test.go`).

### M-10 · minor · A second top-level test beside `TestBridgeFiling` (`go-tests.mdc`)
**Answer:** Declined. "Drawn after the ends are handles" is its own behavior with its own setup, and D-7 says so. `TestBridgeFiling` covers "drawn before".

### X-10 · minor · §9.1's "Refused filings don't fail the claim" title is left stale (F-3)
**Answer:** Fixed. STEP-5 retitles it "…the claim or the connect".

### X-11 · minor · STEP-4 imports are not named (F-4)
**Answer:** Fixed. STEP-4 lists `slices`, `uuid`, `database` and `identityclaims`.

### X-12 · minor · T-2 / T-5 grid positions are not given (F-5, F-6)
**Answer:** Fixed. STEP-3 says any distinct positions with the 0 / 2 / 4 spacing, and that the bridge is the `connect.Result.Subject` the helper returns.

### X-13 · minor/nit · Full-suite check unverifiable read-only; delete paths on filed bridges (F-1, F-2, F-7, F-8)
**Answer:** Logged. The writer's prototype ran the full suite at `c8a8bdb`, and only the two named tests failed. STEP-1's Stop line covers any other failure.
