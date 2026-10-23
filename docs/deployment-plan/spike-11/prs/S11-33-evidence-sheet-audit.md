# S11-33 — Evidence sheet audit follow-up

**Status:** Landed in [#525](https://github.com/mendahu/provenencia/pull/525) (2026-10-23)

**Landed 2026-10-23.** Differences from the plan, by step. STEP-1: the chrome calls `PVSheetDismiss.action` directly and the presenters pass the running and dirty Bools (no `PVSheet` type, policy or per-presenter helper), so T-2 is folded into T-1; Escape is a hidden cancel-action button in the chrome and was not tried by hand (the argument box's two-Escape stop condition is unchecked). STEP-2: T-4 is a mapping test over `claimEvidenceProto` with a catalog's vocabulary, and the real fill of the reading is pinned in `promotealign` (`TestEvidenceReadsEachLinkFromTheNeighbor`, `TestEvidenceValuesTerm`); the handler tests that read the removed pin fields now read the member Observation out of the pin id. STEP-3: `BridgeFact.kind` and `EvidenceBridgeEdgeSummary.productTermKey` / `eventTypeKey` / `eventTypeLabel` (with their test) are deleted too, since nothing reads them now; `PromoteFlow.bridge(between:and:)` is the lookup `reasonText` uses. STEP-4: `L10n.PropertyTerm.resource` stays as a one-line accessor for `PropertyTermDisplay`; the tests that read a term's name use `.label`; `ConnectionSentenceRegistryTests` keeps no term lists. STEP-5: `StyledFormat.attributed` gained `literal:`, the attributes of text outside an argument, so "joins" stays muted inside the one `titleHandle` format. STEP-7: `EvidenceBoard.build(evidence:neighbor:labels:worded:)` takes the evidence (which carries the Source titles) and a lookup for the page's connected rows, not a row; its cards and a relationship's groups are one `Group` type; a line view takes the line's own checked pins (`checked`) where D-9 says `check`, so the member boxes of an open line need no second input; the shared section body, the pin box and `isDepiction` landed here rather than in STEP-8; `PromoteModel.nameFormatChanged()` is called from `PromoteView` when the session's name order changes. STEP-8: names are set by `.stronglyEmphasized`, which a semibold display face does not render heavier, so a `PVSectionHeader` given an `AttributedString` title draws it at regular weight and its emphasized runs carry the weight (a `String` or localized title stays semibold; `TextEmphasisRenderingTests` pins the body face, the regular display face and the header); `rg horizontalPadding macos/App/Features` still prints two `WorkspaceSidebar` comments. STEP-9: the translator comments of STEP-11 for `recordCount`, `spokenRelationship` and `untitledRecord` were written here. The S11-20 plan's Next-adds line was fixed in STEP-12. The §10 manual check was not run: it needs the example project and the app.
**Spike / slice:** Spike 11 · Slice 5 — Claim evidence
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the sheet keeps the S11-D2 board's layout; this PR changes wiring, copy plumbing, keyboard and close behavior)
**Board:** —
**Depends on:** S11-20 (same branch, PR [#525](https://github.com/mendahu/provenencia/pull/525), not merged)
**Blocks:** —
**Branch:** `s11-20-claim-evidence-sheet` (the user chose to land this on S11-20's branch and PR)
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-client.mdc` (seeded keys spelled once in `ConclusionVocabulary.swift`, pinned by `api/ffi/swift_vocabulary_test.go`), `.cursor/rules/go-tests.mdc`
**Context:** [`claim-evidence.md`](../claim-evidence.md); the S11-20 ship log [`S11-20-claim-evidence-sheet.ship.md`](S11-20-claim-evidence-sheet.ship.md) (Round 16 is the sheet's current spec; rounds 17 – 20 refine it); working rules "The pin policy has one owner, in Go", "One wire vocabulary"
**Planned:** 2026-10-23 against `s11-20-claim-evidence-sheet` @ `a860a37c`
**Review:** [`S11-33-evidence-sheet-audit.review.md`](S11-33-evidence-sheet-audit.review.md) — Fits with changes / Ready with changes after 2 rounds; every finding fixed

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

An area audit of the Promote evidence sheet, after twenty rounds of user-driven changes on S11-20, found:
- stale docs and copies of the same code;
- wiring that grew one field at a time;
- a sheet that rebuilds every line on each keystroke;
- a relationship header named from the page instead of from the engine;
- word-splicing copy that a second language cannot translate;
- dialogs that close in three different ways.

The user chose audit items 1 – 10 and A, B, C, F, G, H, I for this PR, on S11-20's branch. D and E are declined by design.

**Why now.** The sheet is about to merge; each item is cheaper now than after other features build on it.

**What stays.**
- The sheet's layout and copy, as the S11-D2 board and S11-20 rounds 16 – 20 left them.
- The matching rules.
- The pin policy: every line can be pinned, and Go owns it.
- The sheet still opens only from a row with a matched record. Audit item D is declined: New and Skip rows need no evidence sheet.

## 2. User stories

**US-1.** As a researcher, I close any dialog the same way: Escape or a click outside it. A dialog with unsaved edits asks first either way, and its Cancel button still discards at once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The evidence sheet is open; press Escape | The sheet closes |
| AC-1.2 | A `pvConfirm` dialog (e.g. "Delete note?") is open; click outside it | It closes as Cancel does |
| AC-1.3 | A "can't delete" notice (`PVDeleteImpactNotice`) is open; press Escape, or click outside | It closes |
| AC-1.4 | A form dialog with unsaved edits; press Escape | The discard question shows (as an outside click already does); it does not close |
| AC-1.5 | Any sheet while its action runs (`isRunning`); Escape or an outside click | Nothing happens |
| AC-1.6 | A form dialog with unsaved edits; click its Cancel button | It closes without asking (unchanged) |

**US-2.** As a researcher, the evidence sheet stays responsive while I type an argument or check boxes on a record with many sources.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A sheet with 200 lines; type one character in the argument box | Every evidence line's input value is unchanged, so no line body is re-evaluated (T-6) |
| AC-2.2 | Check one line's box | Only that line's checked state differs; other lines' inputs are equal to before (T-6) |

**US-3.** As a researcher, each relationship card names the connection from the neighbor's side, as the engine reports it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | John (row) and Mary are spouses, and also cousins, both linked on the page | Two cards: "…they share a spouse (…)" and "…they share a cousin (…)" |
| AC-3.2 | Mary is John's parent (recorded "Mary is the parent of John") | "…they share a parent (…)" |
| AC-3.3 | The neighbor holds a researcher's directed term "Godparent" with no inverse, and the row is the godparent | "…they share a connection (Godparent) (…)" |
| AC-3.4 | The neighbor is John's birth event | "…they share a birth event (…)" |
| AC-3.5 | The row is Leeds, the neighbor Yorkshire, "Leeds is part of Yorkshire" | "…they share a parent place (…)"; from Yorkshire's sheet, Leeds reads "a place within it" |

**US-4.** As a translator in a future language, every sentence on the sheet is one translatable phrase; no sentence has a translated word spliced into it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Subtitle and evidence heading for an event row | One key per kind and titled/untitled (`promote.sheet.subtitle.event`, …); no `kindWord` argument |
| AC-4.2 | A similar relationship to a pibling | The per-term similar phrase "a similar aunt or uncle" comes whole from the term table; `evidenceGraph.share.similarTerm` is gone |
| AC-4.3 | The sheet title | One format with "joins" inside it (`promote.sheet.titleHandle`); `promote.sheet.titleJoins` is gone |

**US-5.** As a contributor, adding a seeded term means one entry in one wording table, and a test fails if I forget it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | A seeded `role` term exists in Go with no Swift constant, or a constant with no `L10n.PropertyTerm` wording | `TestSwiftVocabularyMatchesGo` fails, naming it |
| AC-5.2 | `add-seeded-vocabulary` | Lists the Swift step: the term's `Seeded…Term` constant and its `L10n.PropertyTerm` wording entry |

**US-6.** As a researcher, VoiceOver tells me what the screen shows, and a failed load can be retried.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-6.1 | A relationship bundling 3 records | Its checkbox speaks "…, 3 records." (one record: no count) |
| AC-6.2 | A bundled likeness line of 2 records | It speaks "…, 2 records." |
| AC-6.3 | A member under an untitled Source | Spoken "a record: John Smith." (as shown) |
| AC-6.4 | Evidence loading | VoiceOver reads "Loading" once for the skeleton |
| AC-6.5 | The evidence read fails | The danger callout has "Try again", which reloads the evidence |
| AC-6.6 | A researcher's own property `my_prop` labelled "Occupation note" | The card header reads "…they have a similarity in occupation note" |

## 3. Scope

| In | Covered by |
| --- | --- |
| Audit 1 — docs: `claim-evidence.md`, map S11-20 entry, `docs/deployment-plan/spike-11/completed.md`, `docs/promote-graph-alignment.md:135`, S11-20 plan superseded-AC note | STEP-12 |
| Audit 2 — `add-seeded-vocabulary` Swift step | STEP-4, AC-5.2 |
| Audit 3 — stale comments, previews, misplaced doc, `PinOffers` wrap, delete `connectionLink` | STEP-11 |
| Audit 4 — README rows for `PVTextArea(fillsHeight:)` and `PVSectionHeader` | STEP-11 |
| Audit 5 — spoken counts, untitled Source spoken as shown, loading spoken | STEP-9, AC-6.1 – AC-6.4 |
| Audit 6 — one sheet presenter with one Escape route and one outside-click rule; failed-load retry | STEP-1, STEP-9, AC-1.1 – AC-1.6, AC-6.5 |
| Audit 7 — shared section body, pin box, `isDepiction`, Updated caption; fold `LikenessTile` | STEP-8 |
| Audit 8 — layout numbers from tokens where a token equals them; disclosure alignment on `PVButton` | STEP-8 |
| Audit 9 — one untitled-Source rule | STEP-7, STEP-9 |
| Audit 10 — names styled by weight, drop `titleWeight`; no ellipsis surgery; split the sheet file | STEP-5, STEP-8, STEP-10 |
| A — sheet performance | STEP-7, AC-2.1, AC-2.2 |
| B — relationship named from the engine's neighbor-side reading | STEP-2, STEP-3, AC-3.1 – AC-3.5 |
| C — one seeded-term wording table, checked both ways | STEP-4, AC-5.1 |
| F — whole-phrase copy for every sheet sentence; the rule in `add-localized-string` | STEP-5, AC-4.1 – AC-4.3 |
| G — unread evidence fields off the wire and out of Swift | STEP-2 |
| H — `repropose` stops rebuilding bridges; no O(n²) bridge lookup | STEP-6 |
| I — property label for a researcher's property | STEP-9, AC-6.6 |

| Out | Home |
| --- | --- |
| Audit D — a sheet for New / Skip / unset rows | not planned (user: by design) |
| Audit E — "similarity" wording on all-conflict cards | not planned (user declined) |
| Menu and popup dismissal (`pvContextMenu`, `HistoryJumpMenuHost`, `OmnibarResultsHost`, `PVComboBox` each dismiss their own way) | `docs/ideas/shared-list-helpers.md` gains a "one popup dismissal" item (STEP-12) |
| Word-splicing outside the sheet: `EventTitle` formats, `evidenceGraph.bridge.summary.*`, `CitationComposer` term formats, `sources.sortedBy`, `conclusions.outcomeOutvotedBy` | `docs/ideas/international-names.md` gains a "whole-phrase copy app-wide" item (STEP-12) |
| Callers of `EvidenceBridgeEdgeSummary.sentence` in the graph and composer, beyond the O(n²) fix | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Sheet | `macos/App/Features/Promote/PromoteEvidenceSheet.swift` (754 lines, 13 types) | Presented by `PromoteView.swift:68` `.sheet(item: sheetBinding)`, the only raw `.sheet(` under `Features/`. `sheetBinding.get` reads `model.flow.row(model.flow.sheetSubjectID)`. Content is `ScrollView { VStack }`. `cards(_:)` builds `EvidenceContext(row:evidence:model:)` (:253, the per-render wording layer: `worded`, `propertyLabel`, `sentence`, `member`, `thisSourcePlacement`), passed to every card and line, so any `flow` change re-runs every line body |
| Close controls | same file :43 – :54 | Footer "Back to all rows" has a11y id `promote.sheet.close`. The title X `PVIconButton(.dismiss, label: L10n.Promote.closeSheet)` has a11y id `promote.sheet.dismiss`; ids are crossed with the L10n key names. `.pvOutsideClick { model.closeSheet() }`. No Escape. No UI test references either id (`rg "promote.sheet.close\|promote.sheet.dismiss" macos` finds only this file) |
| Kit dismissal | `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift` | `PVFormDialogDismiss.action(isRunning:isDirty:)` (:42). Cancel has `.keyboardShortcut(.cancelAction)` (:117) and closes at once (doc :172 – :173). `.pvOutsideClick(isEnabled: !confirmingDiscard)` (:216) asks first when dirty |
| Kit dismissal | `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift` | Cancel has `.cancelAction` (:150). No outside click |
| Kit dismissal | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` | `PVDeleteImpactNotice` (:176, raw `.sheet` :359). "Done" has `.defaultAction` only, so Escape does nothing |
| Kit dismissal | `macos/App/DesignSystem/Components/Panel/PVOutsideClick.swift` | `.pvOutsideClick(isEnabled:perform:)` |
| Display helpers | `macos/App/Features/Promote/PromoteEvidenceDisplay.swift` | `bridge(for:row:in:)` :42 and `bridge(between:and:kind:in:)` :28; `sharePhrase` :189; `spokenHeader` :209 (strips "…"); `claim` :140 (`kindWord`); `similarityHeader` :162; `Member` :61; `recordSource` :85 |
| Registry | `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift` | `ShareInput` :37; `roleShares`/`eventShares`/`relationshipShares` :80 – :117; place switch on `(termKey, neighborIsA)` :184; `termInLowercase` :69 |
| Bridge facts | `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift` | `reasonRole`, `productTermKey`, `eventTypeKey`, `eventTypeLabel`, `sentence` (:71 linear `snapshot.bridges.first`). Four lookups per bridge in `PromoteModel.facts()` |
| Model | `macos/App/Features/Promote/PromoteModel.swift` | `load()` :138; `repropose` :298 – :325 (recomputes bridge facts and drops them); `facts()` :361 – :415; `propertyLabel(key:origin:)` :288; `loadEvidence` :201; `setSheetSubject` :193; `closeSheet` :185 |
| Flow | `macos/App/Features/Promote/PromoteFlow.swift` | `Row.evidence` :159; `setEvidence` :407 – :416 (sets pins from `offeredPinIDs`/`checkedPinIDs`); default pins :479; `BridgeFact` :228 – :251 (`term`, `termKey`, `eventTypeKey`, `eventType`); `selectionReadout` doc misplaced above `readingTitle` :179 – :187 |
| Term copy | `macos/App/Platform/L10n.swift` | `PropertyTerm.resource` :2668 (32 cases; none for gender or sex at birth); `EvidenceGraph.ShareTerm` :2270 (30 statics, 60 keys); `shareSimilarTerm` :2279; `Promote.kindWord` :1462; `titleJoins` :1456; `connectionLink` :1498; `similarityIn` :1471 and the share headers :1474 – :1480 each start with "…" |
| Readout | macos/App/Features/Promote/PromoteReadoutBadges.swift (renamed `PromoteReadout.swift` by this PR) | The band badges shared by row and sheet. The "Updated" caption is copied at `PromoteView.swift:522` and `PromoteEvidenceSheet.swift:323` |
| Vocabulary | `macos/App/Platform/ConclusionVocabulary.swift` | `Seeded…Term` enums :302 – :376 (`SeededSexAtBirthTerm` lacks intersex, indeterminate, unknown); `SeededRelationshipTypeTerm.inverse` :357; `SegmentCause` :227 |
| Go test | `api/ffi/swift_vocabulary_test.go` | "Seeded term keys are seeded" checks Swift ⊆ seed (:374). The inverse parity subtest is at :399 – :432. `SegmentCause` entries are at :297 and :349 |
| Go key | `core/segmentcompare/key.go` | `Key(l Link, terms Terms)` reads a link from one end. A term and its inverse read as whichever sorts first, with End and Other swapped, so `End` is the reading end in the key's term, not always the stored end. A symmetric same-type link has `End = ""`. Tests: `segmentcompare_test.go:57 – :67` |
| Go evidence | `core/claimevidence/claimevidence.go` `Relationship` :76 (`Key vocabkey.BridgeKey`); `core/database/promotealign/evidence.go:103 – :111` (row-side key via `b.From(row.SubjectID)`); `core/database/promotealign/terms.go` `loadTerms` (only `Directed`, `Inverse`, only disambiguation properties); `core/database/graphcache/vocab.go:87` `VocabTerm` (has `Label`, `Origin`, `InverseKey`); `core/connectrules` `Bridges()` endpoints | The neighbor's reading and term labels are not on the wire |
| Wire | `api/proto/engine.proto` :1531 – :1611; `api/ffi/handlers/promote_evidence.go` `claimEvidenceProto` :67 | Sent but never read in Swift: relationship `term`/`end`/`event_type`/`cause`/`points`/`checked`; line `points`/`checked`; pin `member_subject_id`/`member_observation_id`; bundle `member_subject_id`/`source_id`. `ConclusionTerm` (:1418) already carries `id, key, label, origin` |
| Likeness | macos/App/Features/Conclusions/LikenessTile.swift (deleted by this PR) | One caller (`ConclusionDetailPage.swift:269`). The sheet reads only `LikenessTile<EmptyView>.crop` (96) |
| Tests | `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`, `ConnectionSentenceRegistryTests.swift` (term lists :272 – :287), `PVPanelFormDialogTests.swift`, `PromoteModelTests.swift`, `core/claimevidence/*_test.go`, `core/database/promotealign/*_test.go`, `api/ffi/handlers/promote_evidence_test.go` | — |

**Precedent.**
- **One shared dismissal.** `9b48c31f` / [#371](https://github.com/mendahu/provenencia/pull/371) introduced the outside-click watcher with `PVFormDialogDismiss`, and S11-20 round 14 extracted it to `PVOutsideClick` (`d591470d`). Extend that; do not add a second mechanism.
- **Removing wire fields.** S11-20 round 11 reserved `lifespan`, and round 16 reserved `pinnable` (`0818810b`), each as `reserved N; reserved "name";`.
- **A term on the wire.** `EventHeader.event_type` is a `ConclusionTerm`, filled at `api/ffi/handlers/conclusion_headers.go:119`.
- **Lazy lists.** `macos/App/DesignSystem/Components/List/PVList.swift` uses `LazyVStack`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **One kit sheet presenter** (detailed below the table). | One rule for every sheet (user, item 6). Escape and outside click cannot disagree, and a future rule (running, dirty) reaches both. | A second Escape route (`.onExitCommand` beside `.cancelAction`), which the audit found drifting |
| D-2 | **Confirms and the notice also close on an outside click**, as Cancel / Done. The evidence sheet closes on Escape. | Standard behavior (user) | Keeping confirms closed to outside clicks |
| D-3 | **The evidence sheet keeps both close controls** (board E01). The title X closes with a11y id `promote.sheet.close`; the footer "Back to all rows" closes with id `promote.sheet.back` *(renamed)*. Neither is `.defaultAction`. | The ids matched neither control's key; no test reads them | Removing one control, which is the board's call |
| D-4 | **The neighbor's reading lives in `segmentcompare`** (detailed below the table). The sheet stops reading `SeededRelationshipTypeTerm.inverse`. The map and its parity subtest **stay**: `CitationComposerAuthoring.swift:706 – :713` `inverseTerm(of:property:vocabulary:)` still reads it (Out). It goes when `docs/ideas/term-management.md` item 2 carries `inverse_key` on the wire term. | The orientation rule stays in the package that owns `Key`. Place and role links keep their wording, and a researcher's term pair or a future `part_of` inverse needs no sheet change. One Swift inverse map remains until term-management replaces it everywhere at once | Computing it in `claimevidence` (a second copy of `Key`'s rule); deleting the map while the composer reads it |
| D-5 | **Wire.** `PromoteClaimEvidenceRelationship` gains `neighbor_end = 14` (string), `neighbor_term = 15` (`ConclusionTerm`, the reading's term), `neighbor_term_names_neighbor = 16` (bool) and `event_type = 17` (`ConclusionTerm`). Remove and reserve relationship `term` (4), `end` (5), `event_type` (6, replaced by 17), `cause` (8), `points` (9) and `checked` (11); line `points` (6) and `checked` (8); pin `member_subject_id` (2) and `member_observation_id` (3); bundle `member_subject_id` (2) and `source_id` (3). Go types keep `Points` and `Checked`, which Go tests and `PinOffers` read. Delete Swift `SegmentCause` from `ConclusionVocabulary.swift:227`, `GenealogyStore.swift:388` and `GoStore.swift:1062`, and its `swift_vocabulary_test.go` entries at :297 and :349. | "One wire vocabulary". A term stays one message (`ConclusionTerm`) | Flat label and origin strings |
| D-6 | **The wording table lives in `L10n.PropertyTerm`.** `L10n.PropertyTerm.wording(propertyKey:termKey:) -> Wording?` *(new)*, with `struct Wording { label; share: LocalizedStringResource?; shareSimilar: LocalizedStringResource? }` and a `switch` over `Seeded…Term` constants, the shape `resource` has today. `resource` becomes `wording(…)?.label`, so `PropertyTermDisplay` is unchanged. Every seeded term of `event_type`, `role`, `relationship_type`, `place_relationship_type`, `gender` and `sex_at_birth` has an entry (40). The registry's share lookup reads `share` / `shareSimilar`. `ShareTerm`, its 60 keys, the three registry tables and `shareSimilarTerm` go. `SeededSexAtBirthTerm` gains `intersex`, `indeterminate` and `unknown`. | Copy stays in `L10n.swift`, as `add-localized-string` says; one place per term | A new `SeededTermWording.swift` file |
| D-7 | **Both-direction checks in Go.** `TestSwiftVocabularyMatchesGo` adds two subtests. "every seeded term is in Swift" checks seed ⊆ the `Seeded…Term` enums. "every seeded term has wording" reads `L10n.swift`'s `PropertyTerm.wording` body and checks every enum constant appears as `Seeded…Term.name` there (the `swiftEnumBody` approach). | AC-5.1: a forgotten term fails one test with no hand list | A Swift test with its own constant list |
| D-8 | **Whole-phrase rule.** A format's argument is data (a name, a title, a count) or a complete noun phrase owned by one table entry (a share phrase with its article). It never takes a bare translated word that the sentence must agree with. The subtitle and heading become twelve keys: `promote.sheet.subtitle.{person,event,place}`, `promote.sheet.subtitleUntitled.{…}`, `promote.sheet.evidenceHeading.{…}` and `promote.sheet.evidenceHeadingUntitled.{…}`. `promote.sheet.titleHandle` holds "joins" ("%1$@ joins %2$@ %3$@"). Similar share phrases come whole from D-6. `kindWord`, `titleJoins` and `shareSimilarTerm` go. `add-localized-string` gains this rule. | User, item F: ready for a second language | Inflection helpers, with no second language to test against |
| D-9 | **Evidence board** (detailed below the table). | AC-2.1, AC-2.2: work once per load, one copy of each fact, one place that words a line | Keeping `Row.evidence` beside a board; memoising inside views |
| D-10 | **The argument box keeps its direct binding** to `model.setArgument`. | With `Equatable` line inputs (D-9) a keystroke no longer re-runs line bodies; a local draft would need flushing on every close path | A debounced local draft |
| D-11 | **`repropose` builds subjects only.** `facts()` splits into `subjectFacts()` and `bridgeFacts(_ snapshot:)`, a static pure function. Only `load()` builds bridges. `EvidenceBridgeEdgeSummary.sentence` resolves endpoints from the snapshot only when the passed bridge lacks them, through an id index built once like `subjects(of:)`. | Audit H | — |
| D-12 | **Property label fallback.** `PromoteModel.propertyLabel(key:)` *(changed signature)* looks up the product property first, then any property with that key. Callers that pass an origin keep `propertyLabel(key:origin:)`. | Audit I, with no wire change for a label | Property origin on the wire |
| D-13 | **Names by weight.** `EvidenceTextStyle` paints `.name` runs with `inlinePresentationIntent = .stronglyEmphasized` and the display color, inheriting size from the container. Delete `PVSectionHeader.titleWeight` and the `name:` font parameter. | Audit 10; one sentence style | A font per call |
| D-14 | **One ellipsis wrapper.** Card-header formats (`similarityIn`, the share headers) carry no "…". `promote.sheet.cardHeaderLead` *(new)* "…%@" wraps them for display, and VoiceOver reads the unwrapped text. Delete `spokenHeader`. | Audit 10 / M-11 | A display copy of each header format |
| D-15 | **Layout from tokens where one equals the value.** `cropGap` = `PVSpacing.space4`; `relationshipIndent` = `PVSpacing.space6 + PVSpacing.space5`; `inset` and `gap` stay `PVSpacing.space4`. The board's grid numbers stay as named sheet constants with a one-line source comment: `box` 18 (board 18 px column), `badgeColumn` 116, `valueColumn` 200, `sourceColumn` 200, `memberValueColumn` 170, `memberIndent` 44. On screen nothing moves. `PVButton` gains `flushLeading: Bool = false` *(new)*, which offsets the label by `size.horizontalPadding` so a ghost button's text starts on its column edge. `PVDisclosureButton` forwards it, and the sheet uses it in place of `.padding(.leading, -…)`. | User: use existing tokens; the padding belongs to `PVButton` | New column tokens; deriving 44 from tokens (it would move rows 12 pt) |
| D-16 | **`LikenessTile` folds back** into `ConclusionDetailPage` as a private subview, `LikenessTileView`. `CachedThumbnail.likenessCrop: CGFloat = 96` *(new static)* is read by the gallery and the sheet. Delete `LikenessTile.swift` and its unused `Overlay == EmptyView` init. Fix the S11-20 plan's Next-adds line. | Audit 7; one caller; no domain word in the kit | A kit `PVThumbnailSize` |
| D-17 | **File split.** After STEP-8, `PromoteEvidenceSheet.swift` keeps the panel, title, claim fields and bindings. `PromoteEvidenceCards.swift` *(new)* holds layout, cards, lines, members and style. | Audit 10 | — |
| D-18 | **One PR despite size.** 12 steps (over the guide's 10). The user chose one PR on S11-20's branch. | User decision | Splitting |

**D-1 — the sheet presenter, in full.**
- **The modifier.** `.pvSheet(item:isRunning:isDirty:content:)` *(new)*, plus an `isPresented:` overload, in `macos/App/DesignSystem/Components/Panel/PVSheet.swift` *(new)*. It presents the sheet and owns both close routes:
  - a zero-size hidden `Button` with `.keyboardShortcut(.cancelAction)`;
  - `.pvOutsideClick`.
- **One action.** Both routes call `PVSheetDismiss.action(isRunning:isDirty:isConfirmingDiscard:)`, `PVFormDialogDismiss` renamed and moved to `macos/App/DesignSystem/Components/Panel/PVSheetDismiss.swift` *(new)*:
  - `.ignore` does nothing (running);
  - `.keepEditing` hides the discard question (it is showing);
  - `.confirmDiscard` shows the discard question (dirty);
  - `.close` clears the binding.
- **The presenter owns the dirty guard.** `.pvSheet` holds `@State confirmingDiscard` and renders the discard question (`PVConfirmContent`, today in `PVFormDialogModifier` at `PVFormDialog.swift:211 – :240`) in place of the content while it shows. Its "Discard" closes; its "Keep editing" sets `confirmingDiscard = false`. The outside-click watcher is off while it shows. `PVFormDialogModifier` keeps only the form.
- **Cancel buttons.** The kit's visible Cancel buttons drop their own `.keyboardShortcut(.cancelAction)` and close at once (AC-1.6).
- **Who uses it.** `pvFormDialog`, `pvConfirm`, `PVDeleteImpactNotice` and `PromoteView`'s evidence sheet all present through `.pvSheet`. Confirm and notice pass `isRunning` (confirm only) and `isDirty: false`.
- **The discard question.** While it shows, Escape returns `.keepEditing`, back to the form, as its "Keep editing" does. Outside clicks stay off, as today.

**D-4 — the neighbor's reading, in full.**
- **What `k` means.** `k` is the row-side key from `Key` (`segmentcompare_test.go:60 – :63` pins it). `Key` swaps a term to its inverse when the inverse sorts first, and swaps the ends with it. So `k.End` is the row's endpoint **in the key's reading**: "`first` is the `k.Term` of `second`", with the row at `k.End`. This is not always the endpoint the row is stored at. `k.End == ""` means a symmetric same-type link.
- **The function.** `segmentcompare.ReadFrom(k vocabkey.BridgeKey, first, second vocabkey.PropertyKey, terms Terms) (neighborEnd vocabkey.PropertyKey, term vocabkey.BridgeTerm, namesNeighbor bool)` *(new)*, beside `Key`. `first` and `second` are the bridge type's endpoint properties in `connectrules` order. The caller looks them up from `k.Type` in `connectrules.Bridges()`.
- **The rule.**
  - **Symmetric** (`k.End == ""`): `("", k.Term, true)`.
  - **Otherwise** `neighborEnd` is the endpoint that is not `k.End`.
  - **`neighborEnd == first`:** `(first, k.Term, true)`. The neighbor is the `k.Term` of the row.
  - **`neighborEnd == second`, and the term has an `Inverse` in `terms`:** `(second, inverse, true)`.
  - **Otherwise:** `(second, k.Term, false)`. The term names the row's role, not the neighbor's.
- **Worked examples**, written the way `segmentcompare_test.go` writes links (stored term, the end the row reads from):

  | Stored link, row's end | `k` | Result | Phrase |
  | --- | --- | --- | --- |
  | `rel("parent", related_to, person)` (row John, the child) | `child·person` | `(related_to, parent, true)` | "a parent" |
  | `rel("parent", person, related_to)` (row Mary, the parent) | `child·related_to` | `(person, child, true)` | "a child" |
  | `rel("spouse", person, related_to)` | `spouse` (End `""`) | `("", spouse, true)` | "a spouse" |
  | part_of from `from` (row Leeds) | `part_of·from` | `(to, part_of, false)` | "a parent place" |
  | part_of from `to` (row Yorkshire) | `part_of·to` | `(from, part_of, true)` | "a place within it" |
  | participation subject from `person` (row John), event type birth | `subject·person·birth` | `(event, subject, false)` | Worded by the event type: "a birth event" |
  | participation subject from `event` (row the birth) | `subject·event·birth` | `(person, subject, true)` | "a subject, John Smith" |
  | Researcher "godparent", no inverse, from `person` (row the godparent) | `godparent·person` | `(related_to, godparent, false)` | "a connection (Godparent)" |

- **Who fills it.** `claimevidence.Relationship` gains `NeighborEnd`, `NeighborTerm` and `NeighborTermNamesNeighbor` *(new)*. `core/database/promotealign/evidence.go` sets them after `claimevidence.Build`, for each relationship: it looks up `first` and `second` for `rel.Key.Type` in `connectrules.Bridges()`, calls `ReadFrom` with the `terms` it already loads (`loadTerms`), and stores the result. No `Link` needs to be carried through.
- **Labels and origins.** `core/database/promotealign/values.go` `EvidenceValues` (built from the same `vocab.Terms` in `NewEvidenceValues`, already passed to `claimEvidenceProto`) stores label and origin per term, keeps researcher-origin terms as a fallback after the product term, and gains `Term(property, key string) (label, origin string)` *(new)*. The handler fills each `ConclusionTerm` from it. Line labels keep their product-only behavior by checking the origin. `promotealign.Evidence` keeps its signature.
- **The Swift side.** `ShareInput` holds `neighborEnd`, `neighborTerm` (key, label, origin), `namesNeighbor`, `eventType` (key, label, origin) and `neighborName`.
  - **Labels** go through `PropertyTermDisplay.name`.
  - **Product keys** go through the `CatalogPropertyTerm.productKey` rule.
  - **Each kind's `share` closure** reads `(term, namesNeighbor, neighborEnd)`:
    - relationship: `namesNeighbor` gives the term's share; otherwise "a connection (Label)";
    - place: `(part_of, false)` gives "a parent place", `(part_of, true)` "a place within it", and likewise for `succeeded_by`;
    - participation: the neighbor at `event` is worded by the event type; at `person`, by role and name;
    - location: the neighbor at `place` gives "a place, Leeds".

**D-9 — the evidence board, in full.**
- **Where it's built.** `PromoteModel.loadEvidence` builds `EvidenceBoard`, a value type in `macos/App/Features/Promote/PromoteEvidenceBoard.swift` *(new)*, through `EvidenceBoard.build(evidence:row:labels:worded:sourceTitles:) -> EvidenceBoard`, a static pure function.
  - `labels` is a `(String) -> String` property-label lookup.
  - `worded` is a `(CatalogConclusionValue) -> String` with the session's name format.
- **Where it's stored.** It passes the board to `PromoteFlow.setEvidence(subjectID:handleID:board:)`. `Row.board: EvidenceBoard?` replaces `Row.evidence`.
- **What it carries:**
  - `offeredPinIDs` and `checkedPinIDs`, for the pin rules at `PromoteFlow.swift:411 – :416` and the default pins at `:479`;
  - `cards: [Card]`, each with `header` and `lines: [Line]`;
  - `relationships: [RelationshipCard]` (header display and spoken strings, pin ids, outcome, groups).
- **Each line** is an `Equatable` value: id, pin ids, value text or crops, sentence, sub, source title (normalized: blank → `untitledRecord`), outcome, spoken label, and members.
- **Name format.** When `flow.nameOrders` changes (`PromoteModel.swift:133`, `:153`) while a sheet is open, the model reloads its evidence (`setSheetSubject(flow.sheetSubjectID)`) so the board is reworded.
- **The views.** `EvidenceContext` is deleted. Line views take `(line: Line, check: CheckState, isOpen: Bool)`, and the sheet's content is a `LazyVStack`.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `.pvSheet`, `PVSheetDismiss` *(new)* | kit presenter | `PVFormDialogDismiss`, `.pvOutsideClick` | Extends them: renamed, generalised (D-1) |
| `L10n.PropertyTerm.wording` *(new)* | table | `L10n.PropertyTerm.resource`, `ShareTerm`, registry tables | Extends `PropertyTerm`, replaces the other two (D-6) |
| `segmentcompare.ReadFrom` *(new)* | Go func | `segmentcompare.Key` | Beside it; the orientation rule in one package (D-4) |
| `EvidenceValues.Term` *(new method)* | Go method | `core/database/promotealign/values.go` `EvidenceValues.labels` | Extends the one term-label lookup the package has |
| `EvidenceBoard` *(new)* | value type | `EvidenceContext`, `PromoteEvidenceDisplay` | Replaces `EvidenceContext`; calls the display helpers once |
| `PromoteEvidenceCards.swift` *(new)* | file | `PromoteEvidenceSheet.swift` | Split (D-17) |
| `PVButton.flushLeading`, `CachedThumbnail.likenessCrop`, `promote.sheet.cardHeaderLead` *(new)* | option / constant / key | the sheet's negative padding, `LikenessTile.crop`, per-header "…" | Moves each to its owner (D-14 – D-16) |
| `PromoteUpdatedCaption` *(new view)* | view | `PromoteView.swift:522`, `PromoteEvidenceSheet.swift:323` | One copy, in `macos/App/Features/Promote/PromoteReadout.swift` *(new)* (renamed from `PromoteReadoutBadges.swift`) |

**Builds on.** `PVOutsideClick` (S11-20 round 14); `StyledFormat`; `PropertyTermDisplay`; `ConclusionTerm`; `segmentcompare.Key` and `Terms`; the vocabulary drift test.
**Deletes.**
- Swift: `PVFormDialogDismiss` (renamed), `PromoteEvidenceDisplay.bridge(for:row:in:)` and `bridge(between:and:kind:in:)`, `BridgeFact.termKey`/`eventTypeKey`/`eventType`, `EvidenceContext`, `Row.evidence`, `ShareTerm` and its 60 keys, the three registry tables, `shareSimilarTerm`, `kindWord`, `titleJoins`, `connectionLink` with its key and test, `spokenHeader`, `spokenMemberUntitled`, `PVSectionHeader.titleWeight`, `LikenessTile.swift`, Swift `SegmentCause`, the unread Swift evidence fields.
- Proto: the reserved fields (D-5).

**Next adds.** A seeded term is one `L10n.PropertyTerm.wording` case. A new sheet gets Escape, outside click and the dirty guard by presenting with `.pvSheet`. A future `part_of` inverse is a seed change only. `docs/ideas/term-management.md` item 2 gains: the evidence sheet now reads the inverse through `neighbor_term`; only the composer still reads the Swift map.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/segmentcompare` | `ReadFrom` *(new)* (D-4) |
| `core/claimevidence` | `Relationship.NeighborEnd`, `NeighborTerm`, `NeighborTermNamesNeighbor` *(new)* |
| `core/database/promotealign` | `evidence.go` fills them; `values.go` `EvidenceValues.Term` *(new)*; `pins.go` doc wrap |

### FFI

`api/proto/engine.proto` per D-5. Handler `api/ffi/handlers/promote_evidence.go` fills `neighbor_end`, `neighbor_term`, `neighbor_term_names_neighbor` and `event_type` (`ConclusionTerm` with key, label and origin from `EvidenceValues.Term`). Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Sheet presenter | `.pvSheet` | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` *(new)* | D-1 |
| Board | `EvidenceBoard` | `macos/App/Features/Promote/PromoteEvidenceBoard.swift` *(new)* | D-9 |
| Cards | split | `macos/App/Features/Promote/PromoteEvidenceCards.swift` *(new)* | D-17 |
| Retry | `PVCallout(actions:)` + `PVButton` | sheet failed state | `PromoteModel.reloadEvidence()` *(new)* calls `setSheetSubject(flow.sheetSubjectID)`, the existing load path |
| Loading | skeleton stack | sheet | one accessibility element, label `L10n.DesignSystem.listLoading` |

### Strings

**New:**
- the twelve subtitle and heading keys (D-8);
- `promote.sheet.titleHandle` changed to "%1$@ joins %2$@ %3$@";
- `promote.sheet.retry` "Try again";
- `promote.sheet.cardHeaderLead` "…%@";
- `promote.sheet.spokenRelationshipCount` (plural, "%1$@. %2$@, %3$lld records.");
- `promote.sheet.spokenLikenessCount` (plural, "Likeness from this source and %1$@. %2$@, %3$lld records.");
- `propertyTerm.gender.{man,woman,non_binary}`;
- `propertyTerm.sex_at_birth.{female,male,intersex,indeterminate,unknown}`;
- `evidenceGraph.share.{role,event,relationship}.<term>.similar` (one similar phrase per term, replacing the noun keys).

**Removed:**
- `promote.sheet.kind.{person,event,place}`;
- `promote.sheet.titleJoins`;
- `promote.sheet.subtitle` and `promote.sheet.subtitleUntitled` (the old shared forms);
- `promote.sheet.evidenceHeading` and `promote.sheet.evidenceHeadingUntitled` (the old shared forms);
- `evidenceGraph.share.similarTerm`;
- `evidenceGraph.share.{roleNoun,eventNoun,relationshipNoun}.*`;
- `promote.connection.link`;
- `promote.sheet.spokenMemberUntitled`.

## 7. Steps

**STEP-1 — One sheet presenter**
Files: `macos/App/DesignSystem/Components/Panel/PVSheet.swift` *(new)*, `macos/App/DesignSystem/Components/Panel/PVSheetDismiss.swift` *(new)*, `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`, `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Features/Promote/PromoteEvidenceSheet.swift`, `macos/App/DesignSystem/README.md`, `.agents/skills/add-ui-component/SKILL.md`, `macos/ProvenenciaTests/PVPanelFormDialogTests.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-ui-component`.
Do: D-1, D-2, D-3.
Failing tests first: T-1, T-2.
Check: T-1, T-2; `rg "sheet\((item|isPresented)" macos/App/Features macos/App/DesignSystem` prints only `PVSheet.swift`; scripts/check-pr.sh PASS.

**STEP-2 — Go reads the link from the neighbor's end; unread fields go**
Files:
- Go: `core/segmentcompare/key.go` and `core/segmentcompare/segmentcompare_test.go`, `core/claimevidence/claimevidence.go`, `core/database/promotealign/{evidence,values,pins}.go` and tests.
- Wire: `api/proto/engine.proto`, generated code, `api/ffi/handlers/promote_evidence.go` and test.
- Swift: `macos/App/Platform/{GenealogyStore,GoStore,FakeStore,ConclusionVocabulary}.swift`, `api/ffi/swift_vocabulary_test.go`.

Skills: `add-ffi-handler`, `add-match-rule`.
Do: D-4 (Go and wire), D-5.
Failing tests first: T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` PASS.

**STEP-3 — The card reads the engine's reading**
Files: `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`, `macos/App/Features/Promote/{PromoteEvidenceDisplay,PromoteFlow,PromoteModel}.swift`, `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`, `macos/App/Platform/ConclusionVocabulary.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/{PromoteEvidenceDisplayTests,ConnectionSentenceRegistryTests}.swift`.
Do:
- D-4 (Swift side).
- Delete `bridge(for:row:in:)` and `bridge(between:and:kind:in:)` (move its one other caller, `PromoteModel.reasonText`, to a lookup on `BridgeFact` ends). Remove the registry's use of `SeededRelationshipTypeTerm.inverse` (`ConnectionSentenceRegistry.swift:164`); the map and its parity subtest stay for the composer (D-4).
- Delete `BridgeFact.termKey`, `eventTypeKey` and `eventType`. `BridgeFact.term` stays: the row's reason reads it.

Failing tests first: T-5.
Check: T-5; `rg "\.inverse\b" macos/App` prints only `ConclusionVocabulary.swift` and `CitationComposerAuthoring.swift`; check-pr PASS.

**STEP-4 — One seeded-term wording table**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift`, `.agents/skills/add-seeded-vocabulary/SKILL.md`. Skills: `add-seeded-vocabulary`, `add-localized-string`.
Do: D-6, D-7. `ConnectionSentenceRegistryTests` drops its hand-kept term lists; spot checks of a few phrases stay.
Failing tests first: T-7. "every seeded term is in Swift" fails on the three missing `sex_at_birth` terms before the fix.
Check: T-7, T-8; check-pr PASS.

**STEP-5 — Whole-phrase copy on the sheet**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Promote/PromoteEvidenceDisplay.swift`, `macos/App/Features/Promote/PromoteEvidenceSheet.swift`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`, `.agents/skills/add-localized-string/SKILL.md`.
Do: D-8, D-14.
Failing tests first: T-9.
Check: `rg "kindWord|titleJoins|shareSimilarTerm|spokenHeader" macos/App` prints nothing; check-pr PASS.

**STEP-6 — Facts once**
Files: `macos/App/Features/Promote/PromoteModel.swift`, `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`.
Do: D-11.
Check: T-10; check-pr PASS.

**STEP-7 — Evidence board and a lazy sheet**
Files: `macos/App/Features/Promote/PromoteEvidenceBoard.swift` *(new)*, `macos/App/Features/Promote/{PromoteFlow,PromoteModel,PromoteEvidenceSheet,PromoteEvidenceDisplay}.swift`, `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift` *(new)*, `macos/ProvenenciaTests/PromoteFlowTests.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`.
Do: D-9, D-10. Normalise the untitled Source once in the board.
Failing tests first: T-6.
Check: T-6; `rg "EvidenceContext|\.evidence\b" macos/App/Features/Promote` prints only `sheetEvidence` state uses; check-pr PASS.

**STEP-8 — One body, one box, tokens, likeness**
Files:
- Promote: `macos/App/Features/Promote/PromoteEvidenceSheet.swift`; macos/App/Features/Promote/PromoteReadoutBadges.swift → `PromoteReadout.swift` *(renamed, + `PromoteUpdatedCaption`)*; `macos/App/Features/Promote/PromoteView.swift`.
- Kit: `macos/App/DesignSystem/Components/Button/PVButton.swift`, `macos/App/DesignSystem/Components/Disclosure/PVDisclosureButton.swift`, `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift`, `macos/App/DesignSystem/README.md`.
- Likeness: `macos/App/Platform/CachedThumbnail.swift`, `macos/App/Features/Conclusions/{LikenessTile,ConclusionDetailPage}.swift`.
- Other: `macos/App/Platform/GenealogyStore.swift` (`CatalogConclusionValue.isDepiction` *(new)*), `macos/Provenencia.xcodeproj/project.pbxproj`.

Do:
- **Audit 7:**
  - `EvidenceSectionLines(section:style:expanded:)` *(new view)* replaces the property-group body written twice, in `EvidenceSectionCard` and inside `EvidenceRelationshipCard`.
  - `EvidencePinBox(ids:spoken:identifier:)` *(new view)* replaces the checkbox block repeated for lines and relationship headers (`binding(ids)`, `isIndeterminate`, `.frame(width: box)`).
  - `isDepiction` replaces the four `if case .depiction` tests.
  - `PromoteUpdatedCaption` replaces the two copies of the caption.
- D-13, D-15, D-16.

Check: `rg "LikenessTile|titleWeight|horizontalPadding" macos/App/Features` prints nothing; check-pr PASS; no new warning in changed files (fresh derived data).

**STEP-9 — VoiceOver, untitled Source, retry, labels**
Files: `macos/App/Features/Promote/{PromoteEvidenceBoard,PromoteEvidenceSheet,PromoteModel}.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/{PromoteEvidenceBoardTests,PromoteModelTests}.swift`.
Do:
- The board's spoken labels use the count keys when a relationship or likeness bundles more than one record.
- An untitled member is spoken with `untitledRecord`; delete `spokenMemberUntitled`.
- The skeleton is one element labelled "Loading".
- `reloadEvidence()` and the "Try again" action.
- D-12.

Check: T-11, T-12; check-pr PASS.

**STEP-10 — Split the sheet file**
Files: `macos/App/Features/Promote/PromoteEvidenceSheet.swift`, `macos/App/Features/Promote/PromoteEvidenceCards.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`.
Do: D-17. Move only; no behavior change.
Check: check-pr PASS.

**STEP-11 — Comments, previews, README, dead code**
Files and edits:
- `macos/App/Features/Promote/PromoteFlow.swift` :179 – :187: move the `selectionReadout` doc line back above `selectionReadout`.
- `macos/App/Platform/L10n.swift` translator comments:
  - `recordCount`: "After a relationship that bundles records".
  - `spokenRelationship`: "2 = the outcome word".
  - `untitledRecord`: "A record whose Source has no title, on any line or member".
- `macos/App/DesignSystem/Components/Panel/PVPanel.swift` preview title: "John Smith joins John Smith PER-7QH3D".
- `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` preview: "John Smith from Leeds Mercury obituary belongs in the concluded person John Smith because…".
- `core/database/promotealign/pins.go`: re-wrap the `PinOffers` doc.
- `core/matchregistry/registry.go:29 – :31`: drop "or can be pinned".
- `core/claimevidence/claimevidence.go:2 – :4`: "…which lines start checked; every line can be pinned".
- `macos/App/DesignSystem/README.md`: a TextArea row naming `fillsHeight`, and a SectionHeader row.
- Delete `L10n.Promote.connectionLink`, its key and its test.
- `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`: drop the `// MARK: - T-…` labels; rename `checkedCountIsOfferedPinsOnly` to `checkedCountCountsCheckedOffers`.

Check: `rg "connectionLink|promote.connection.link" macos` prints nothing; check-pr PASS.

**STEP-12 — Docs**
Files and edits:
- `docs/deployment-plan/spike-11/claim-evidence.md`:
  - Line 7: Status.
  - Line 30: a bundled line shows "Show N records" in the Source column.
  - Line 53: the relationship header is "…they share <phrase> (<refs>)".
  - Lines 55 – 65: point at `L10n.Promote.lineSentence`, with no English copy.
- `docs/deployment-plan/spike-11/deployment-plan.md`: S11-20's Ships and Lifespan paragraphs; the S11-33 row status.
- `docs/deployment-plan/spike-11/completed.md`: the S11-20 paragraph describes the final sheet; add an S11-33 paragraph.
- `docs/promote-graph-alignment.md:135`: "A line starts pinned only when value compare says the values are the same; the researcher can pin any other line."
- `docs/deployment-plan/spike-11/prs/S11-20-claim-evidence-sheet.md`, one line under Landed: "§2 ACs describe the round-0 build; the final sheet is the ship log's Round 16 spec and S11-33."
- `docs/ideas/shared-list-helpers.md`: one popup dismissal.
- `docs/ideas/international-names.md`: whole-phrase copy app-wide.
- `docs/ideas/term-management.md` item 2: the evidence sheet reads the inverse through `neighbor_term`; only the composer still reads `SeededRelationshipTypeTerm.inverse`.
- This plan: status and Landed note.

Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-33-evidence-sheet-audit.md docs/deployment-plan/spike-11/prs/S11-20-claim-evidence-sheet.md` reports 0 errors.

Commits: one per step. The subject is one full imperative sentence that says what changed for the researcher, ending in a period.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PVPanelFormDialogTests.swift` `sheetDismissActionTable` *(new)* | `PVSheetDismiss.action`: (running, any, any) → ignore; (idle, any, confirming) → keepEditing; (idle, dirty, not confirming) → confirmDiscard; (idle, clean, not confirming) → close | AC-1.4, AC-1.5, D-1 |
| T-2 | Swift | `PVPanelFormDialogTests.swift` `everyKitSheetPresentsThroughPVSheet` *(new)* | `PVSheet.dismissRoute(isRunning:isDirty:isConfirmingDiscard:)` *(new, the pure function both routes call)* returns the same action as T-1 for each case; and the form, confirm and notice presenters pass `isRunning` and `isDirty` (confirm and notice dirty false) | AC-1.1, AC-1.2, AC-1.3, AC-1.6, D-1 |
| T-3 | Go | `core/segmentcompare/segmentcompare_test.go` `TestReadFrom` *(new)* | the eight D-4 worked examples, built with the file's `rel` helper and `Key` → their `(neighborEnd, term, namesNeighbor)` | AC-3.2, AC-3.3, AC-3.5, D-4 |
| T-4 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestClaimEvidenceNeighborReading` *(new)* | spouse + cousin to one neighbor → two relationships with `neighbor_term` spouse / cousin; a birth → `event_type.key` birth with label and origin; removed fields are reserved in the proto | AC-3.1, AC-3.4, D-5 |
| T-5 | Swift | `PromoteEvidenceDisplayTests.swift` `relationshipShareFromTheEngine` *(new)* | `(parent, true)` → "a parent"; `(godparent, false)` with label "Godparent" → "a connection (Godparent)"; `(part_of, false)` → "a parent place"; `(part_of, true)` → "a place within it"; event neighbor birth → "a birth event" | AC-3.2, AC-3.3, AC-3.4, AC-3.5 |
| T-6 | Swift | `PromoteEvidenceBoardTests.swift` *(new)* `aKeystrokeLeavesLinesEqual`, `aToggleChangesOneLine` | `EvidenceBoard.build` over a fixture of 200 lines; set the argument → `board` and every `Line` equal; toggle one pin → only that line's `CheckState` differs | AC-2.1, AC-2.2, D-9, D-10 |
| T-7 | Go | `api/ffi/swift_vocabulary_test.go` "every seeded term is in Swift", "every seeded term has wording" *(new subtests)* | seed ⊆ enums; every enum constant appears in `PropertyTerm.wording` | AC-5.1, D-7 |
| T-8 | Swift | `ConnectionSentenceRegistryTests.swift` `aFewSharesReadWhole` *(new)* | `wording(relationship_type, pibling)` → share "an aunt or uncle", similar "a similar aunt or uncle"; `(role, informant)` → "an informant" | AC-4.2, D-6 |
| T-9 | Swift | `PromoteEvidenceDisplayTests.swift` `subtitleAndHeadingPerKind`, `titleHoldsJoins` *(new)* | event row titled → `promote.sheet.subtitle.event`; untitled place → `promote.sheet.evidenceHeadingUntitled.place`; title → "John Smith joins John Smith PER-7QH3D" | AC-4.1, AC-4.3, D-8 |
| T-10 | Swift | `PromoteModelTests.swift` `reproposeKeepsBridges` *(new)* | load; change the FakeStore graph's bridges; decide a row (repropose) → `flow.bridges` still equals the loaded ones | D-11 |
| T-11 | Swift | `PromoteEvidenceBoardTests.swift` `spokenCountsAndUntitled` *(new)* | relationship of 3 bundles → spoken "…, 3 records."; 1 bundle → no count; bundled likeness of 2 → "…, 2 records."; untitled member → "a record: John Smith." | AC-6.1, AC-6.2, AC-6.3 |
| T-12 | Swift | `PromoteModelTests.swift` `researcherPropertyLabel`, `reloadEvidenceRetries` *(new)* | key `my_prop` with only a user property "Occupation note" → that label; a failed read, then `reloadEvidence()` → `.loading` | AC-6.5, AC-6.6, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PVPanelFormDialogTests`, `PromoteEvidenceDisplayTests`, `ConnectionSentenceRegistryTests`, `TestSwiftVocabularyMatchesGo`, `promotealign` evidence tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] A clean build-for-testing (fresh -derivedDataPath) shows no warning in a file this PR changed
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the S11-D2 board as S11-20 rounds 16 – 20 left it; no new kit primitive beyond §6
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #525 (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S11-33 row; docs/deployment-plan/spike-11/completed.md: S11-33 paragraph
- [ ] Domain doc updated: claim-evidence.md, promote-graph-alignment.md
- [ ] PR #525 description: an S11-33 section with what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open Promote on the Leeds Mercury obituary (S11-D2 §5.2 project) and open John Smith's sheet. Press Escape: it closes. Reopen it and click outside: it closes. (AC-1.1)
2. Open a form dialog (Sources → Add source) and type a title. Press Escape: the discard question shows (AC-1.4). Press Escape again: back to the form. Click Cancel: it closes without asking (AC-1.6).
3. Delete a note (Source page): the confirm shows. Click outside: it closes and nothing is deleted. (AC-1.2)
4. Try to delete a source type that is in use: the "can't delete" notice shows. Press Escape: it closes. Reopen it and click outside: it closes. (AC-1.3)
5. Reopen John's sheet and type a paragraph in the argument box: no lag. Check and uncheck several lines.
6. A pair linked as spouses and cousins shows two cards with two phrases.
7. With VoiceOver, a bundled relationship speaks its count, and the loading state speaks "Loading". (AC-6.4)
8. The names in the subtitle and heading read heavier than the sentence around them (D-13).
9. Read `.agents/skills/add-seeded-vocabulary/SKILL.md`: it lists the Swift step, the `Seeded…Term` constant and its `L10n.PropertyTerm` wording entry. (AC-5.2)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The hidden cancel-action button does not receive Escape while the argument box's `NSTextView` has focus | Escape in a text view first ends completion; a second Escape reaches the key equivalent. If it never does, stop (below) |
| `EvidenceBoard` changes what a line shows | T-6, T-11 and the existing display tests pin the strings |

**Stop and report if:**

- A removed wire field (D-5) has a reader outside the `GoStore`/`FakeStore` mapping (`rg` it first).
- `connectrules` has a bridge type whose endpoint order disagrees with the reading in D-4's table.
- With focus in the argument box, two presses of Escape do not close the evidence sheet.
- `inlinePresentationIntent = .stronglyEmphasized` renders regular weight in `Text` (D-13).
- An outside click closes a sheet when the click lands in a sheet presented from it (the Name form's nested split confirm).

## 12. Open questions

None.

## 13. PR

**Title:** S11-20 — Claim evidence sheet (with S11-33 audit follow-up)
**Body:** an S11-33 section: objective; what shipped per US; AC evidence; deviations; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-33-evidence-sheet-audit.md`.
