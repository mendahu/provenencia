# Review — S10-44 Unpromote a filed card

**Verdict:** Fits with changes — Ready with changes once Q-1 is answered (round 2 fixes applied)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- **Q-1 (from X-1).** Which SF Symbol is the unlink icon? The planner could not check SF Symbols from a Linux container. Recommendation: confirm on a Mac which slashed-link symbol resolves, and write it into D-15. If none exists, a custom glyph is a design call. The plan stays Draft until this is answered.

## Writer's own calls
- No preview RPC: the confirm's bridge list comes from the client snapshot (D-7), and the engine recomputes the set in the write (D-3).
- A second click on a card that is already unpromoted is a silent no-op, not an error (D-4).
- The end rule ignores polarity, to match what bridge filing reads (D-3).
- "Cards stay, handle stays" goes in the message, so the detail can be empty without bridges (D-16).
- The VoiceOver action is plain "Unpromote", without the name (D-14).
- Canvas tooltips come from a new optional `help` on `GraphCanvasActionTarget`, shown by the document view (D-13), because the card views never see the mouse.
- The footer split is a 36pt trailing square and a 1pt hairline (D-10).

## Round 1

### M-1 · blocker · Refresh through `.promotedSubject` reloads every Conclusion list and page, though this write reports its effects
**Answer:** Fixed — D-9 is now `session.apply(effects)` with no mutation. I checked `CatalogQueryRegistry.invalidations(for:project:)`: `effects.sources` invalidates the Source graph, `effects.touchedHandles` invalidates the lists and each page, and `WorkspaceSession.apply` refreshes progress per `effects.sources`. D-17 makes `FakeStore` return those effects, so T-10 exercises the path.

### M-2 · major · The confirm error uses `localizedDescription`
**Answer:** Fixed — D-8 and T-11 use `L10n.Errors.message(for:)`.

### M-3 · major · Unfiling bridges belongs beside filing bridges in `identityclaims`
**Answer:** Fixed — D-3 is now `identityclaims.UnfileBridgesTx`. `FiledBridge` and the exported query are gone. `deleteimpact` imports no database package, so `identityclaims` can import it.

### M-4 · major · The promoted footer split is computed in up to four places
**Answer:** Fixed — `EvidenceSubjectCard.promotedFooterParts(footer:)` (D-10) feeds the regions, the paint, and both hit branches (D-11). The two duplicated footer blocks in `subject` become `recordFooter`. `handleFrame` and `unpromoteFrame` are gone.

### M-5 · minor · `ReleaseClaimTx` breaks `deleteimpact`'s naming
**Answer:** Fixed — renamed to `deleteimpact.ReleaseClaim`.

### M-6 · minor · The delete docs will still say a claim is removed only by Subject delete
**Answer:** Fixed — STEP-9 and the DoD add one line each to `docs/catalog-deletes.md` item 9 and to `core/database/deleteimpact/README.md`.

### M-7 · minor · Nest the pending type and fold the error into the item
**Answer:** Fixed — the type is now `EvidenceGraphModel.PendingUnpromote` with `var error`. `unpromoteError` and the clearing rule are gone.

### M-8 · nit · Codebase fit omits the precedent for D-7
**Answer:** Fixed — named in §4, D-7, and Builds on. Round 2 corrected the type to `PromoteFlow.BridgeFact` (`PromoteFlow.swift`), built in `PromoteModel.swift`.

### X-1 · blocker · Q-1 is open, so STEP-5 and D-15 cannot be done
**Answer:** To user — see To user. Status stays Draft.

### X-2 · blocker · STEP-6 cannot compile before STEP-7
**Answer:** Fixed — the VoiceOver work (D-14, T-7) moved to STEP-7. STEP-6 says the `unpromote` target falls through `performCardAction`'s `default` until STEP-7.

### X-3 · blocker · The Swift household fixture cannot be built as described
**Answer:** Fixed — D-17 adds `householdModel()`, with the Subjects, end Observations, connect rules (the fake's default `productMatrix`), and memberships set directly. It also adds `graphRows(memberships:)` and the fake's end-finding rule (`observationsBySource` `valueSubjectID`, matching the engine).

### X-4 · blocker · T-3 and T-1 need Go fixtures that do not exist
**Answer:** Fixed — D-18: T-1 and T-2 reuse `newBridgeWorld`, and `onSecondSource()` *(new)* gives the second Source. T-3 uses `scopeSeed` with `handleRevision`, `latest`, and `sourceRevisions` unchanged. AC-3.7 now checks the person handle and the Source; association scopes come from the same registry entry.

### X-5 · major · Footer, hit, and tooltip choices left open
**Answer:** Fixed — D-10 fixes the x positions (ref and name x unchanged; name width `handle.width - 120`), region registration in `recordFooter`, and `fillRound`. D-13 moves the tooltip to the canvas document view: it owns the mouse (`GraphCanvasDocumentView.hitTest` returns itself), so a card tool-tip rect would never fire. T-6 says which footer assertions change. `LocalizedStringResource` conforms to `PVCopy` (`PVCopy.swift`); §4 says so.

### X-6 · major · T-7 cannot reach the private action builders
**Answer:** Fixed — D-14 adds `EvidenceSubjectCard.accessibilityActionNames(for:orders:)`, which T-7 tests.

### X-7 · major · `touchesEnd` under-specified
**Answer:** Fixed — D-3: `touchesEnd` is only the EXISTS arm, with one `?`. `bridgesTouching` keeps its `s.id = ? OR` and its argument list.

### X-8 · major · Name `AcceptedEntityForSubjectTx`
**Answer:** Fixed — D-4 names it and its `sql.ErrNoRows` meaning.

### X-9 · minor · The stale comment and `sourceID`
**Answer:** Fixed — D-8 deletes line 13, documents `pendingUnpromote`, and gives `sourceID` no comment.

### X-10 · minor · Long-name fixture
**Answer:** Fixed — T-6 adds the case `longHandleName+promoted` with the AC-1.2 name.

### X-11 · minor · Checks that need macOS or protoc
**Answer:** Fixed — §7 says which steps need which tools, and says to stop rather than skip. STEP-3's check now greps the generated files for the new types.

### X-12 · minor · Paths in STEP-9; §5.2 vs §5.3
**Answer:** Fixed (paths are now repo-relative). Declined (section): §5.2, "When a comparison subject leaves", is the case Unpromote creates, and the map's docs table names §5.2 for S10-44.

### X-13 · minor · Temptations
**Answer:** Fixed — new Out lines: help on other canvas targets, and teaching `FakeStore.promoteSubject` to file bridges.

## Round 2

### M-9 · blocker · The new `help` field is dropped in `publishHitTargets`
**Answer:** Fixed — D-13 adds `GraphCanvasActionTarget.mapFrame(_:)`, which `EvidenceGraphView.publishHitTargets()` uses at both rebuild sites. `EvidenceGraphView.swift` is in STEP-6, and T-13 asserts `mapFrame` keeps `help`.

### M-10 · blocker · `hoveredHelp` stored beside `hoveredCardAction`
**Answer:** Fixed — `hoveredHelp` is computed from `hoveredCardAction` and `hitTargets` (D-13). It was also raised as X-23.

### M-11 · blocker · A names-only action list is a third copy that drives nothing
**Answer:** Fixed — D-14 lifts the existing `named` list into `EvidenceCanvasCardContent.namedActions(for:)`, adds Unpromote there once, and T-7 reads that list. The SwiftUI twin chain gets the action by hand, the reviewer's fallback, because replacing that chain is wider than this PR.

### Not raised: the twin's dead `.help(promoteHelp(kind:))`
**Answer:** Logged — it is the obvious first user of the `help` slot, but Out here ("Help text on any canvas target other than Unpromote").

### X-14 · blocker · Q-1 unresolved
**Answer:** To user (same as X-1).

### X-15 · blocker · A cross-Source bridge cannot be built through the product
**Answer:** Fixed — `connect` refuses it, so AC-3.3 and T-1 (d) now call `UnfileBridgesTx` with another Source id directly (D-18). The Source filter is a guard.

### X-16 · major · `ThrowingStore` conforms to `GenealogyStore`
**Answer:** Fixed — `OnboardingModelTests.swift` is in STEP-4, with a throwing stub and no protocol default.

### X-17 · major · `bridgeWorld.join` takes no pairs
**Answer:** Fixed — T-2 copies `TestSavePins`'s first subtest (`newPinFixture`, `promote.Save` with `Pairs`). AC-3.2 is reworded to that fixture.

### X-18 · major · `StoreBoom` is private
**Answer:** Fixed — T-11 throws `CoreInvokeError.coded(status: 1, code: "internal.unknown", …)`, the file's own error.

### X-19 · major · The `pvConfirm` title is not a `String`
**Answer:** Fixed — the title is wrapped in `L10n.string(…)`.

### X-20 · major · No way to set the handle name in the layout fixture
**Answer:** Fixed — `placed(...)` gains `handleName:` (`CatalogNameValue(form:)`). T-6 says what it proves for AC-1.2.

### X-21 · major · The action-names helper lacks inputs
**Answer:** Fixed — superseded by M-11. `namedActions(for:)` takes the whole `EvidenceCanvasCard`.

### X-22 · minor · `recordFooter` signature
**Answer:** Fixed — the full signature is in D-10.

### X-23 · minor · Stale `hoveredHelp`
**Answer:** Fixed — computed (M-10).

### X-24 · minor · Fixture ordering
**Answer:** Fixed — D-17 fixes the bridge ids so son-of sorts first.

### X-25 · minor · Twin hairline background
**Answer:** Fixed — `.background(style.band)` is on the whole `HStack`.
