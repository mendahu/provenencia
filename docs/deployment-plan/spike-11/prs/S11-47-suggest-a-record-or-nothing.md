# S11-47 — Promote suggests a record or nothing

**Status:** Landed in #598 (2026-10-27)
**Spike / slice:** Spike 11 · Slice 10 — Name candidates
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the row, its menu, its Updated badge and caption already exist; this PR changes which target a row opens on and one caption sentence)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s11-47-suggest-a-record-or-nothing`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-client.mdc` (wire vocabulary as typed enums, D-13), `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [name candidates § Promote suggests a record or nothing](../name-candidates.md#promote-suggests-a-record-or-nothing), [`promote-graph-alignment.md` § 3](../../../promote-graph-alignment.md#3-the-flow-one-page), map scope entry [S11-47](../deployment-plan.md#s11-47--promote-suggests-a-record-or-nothing); working rules: "Promote suggests a record or nothing", "High tolerance for churn"
**Planned:** 2026-10-27 against `main` @ `a70932e`
**Review:** [`S11-47-suggest-a-record-or-nothing.review.md`](S11-47-suggest-a-record-or-nothing.review.md) — Fits with changes / Ready with changes, 1 round (every finding answered)

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

Promote proposes a record for a row, or nothing. Today Go gives every row it does not place a target the researcher never chose: Skip (reason `weak`) when a candidate fell short, New (reason `no_match`) when there was none. Swift mostly ignores them and opens the row unset, but a medium or strong row that was not placed and not taken still falls through to Skip, and the row reasons word a suggestion Promote did not make. After this PR a proposal row carries a handle or no target, with its alternatives as today. New and Skip are only ever the researcher's choice. When a suggested row loses its record on a re-proposal, its caption reads "Updated · No record suggested". Swift reads the row target as a typed wire code, as it already reads the row reason.

**Why now.** The user's call of 2026-10-24, answering S11-43's open question ([detail](../name-candidates.md#promote-suggests-a-record-or-nothing)); the caption wording is the user's call of 2026-10-27 (D-4).

**What stays.** Which candidates exist, every band and score, each alternative's own band and reason (`ReasonWeak` on an alternative stays), the weak-band draft onto the first alternative, a taken row opening unset with its duplicate note, the `empty` and `taken` row reasons, and decided New and Skip: the page still sends them as fixed pairs (`PromoteGraphAlignmentFixed`) and Go still holds them with `TargetNew` / `TargetSkip`.

## 2. User stories

**US-1.** As a researcher, when Promote finds no record for a row, or finds one that falls short, the row opens with nothing chosen so that New and Skip are only ever my decision.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A person Subject `Ann Ames` and no candidate (`graphalign.Align`, nothing fixed) | Row `Target` is empty, `Reason` is `ReasonUnplaced` (`""`) |
| AC-1.2 | A person Subject `Ann Ames` and one handle `Ames` (surname only, a candidate under the weak bar) | Row `Target` is empty, `Reason` `ReasonUnplaced`; the handle is still an alternative (asserted today by `TestOnlyCandidatesOrNominatedHandlesAreCompared` and `TestBirthIsNeverProposedOnADeath`) |
| AC-1.3 | A person Subject with no values | Row `Target` is empty, `Reason` `ReasonEmpty` |
| AC-1.4 | Two `John Smith` Subjects and one handle `John Smith` | The losing row has `Target` empty, `Reason` `ReasonTaken`, and is a possible duplicate of the placed row (as today, with an empty target instead of Skip) |
| AC-1.5 | A row fixed New, a row fixed Skip | Their rows keep `Target` `TargetNew` / `TargetSkip` and `Reason` `ReasonDecided` (unchanged) |
| AC-1.6 | Swift `PromoteFlow.load` with a proposal row whose target is `.unplaced` and `assessment` is `"medium"` or `"strong"`, reason not `taken` | The row's target is `.unset`, never `.skip` or `.newKind` |
| AC-1.7 | Swift, a proposal row with target `.unplaced`, `assessment` `"weak"` and one alternative `e9` | The row drafts `.handle(e9)`, as today (`aWeakCandidateBelowTheBarStartsSelected`) |
| AC-1.8 | A real project: a Source with a person who has no record and a person whose only candidate falls short | Both rows open unset ("Select where this files"); choosing New or Skip and pressing Done files a New and leaves the Skip off |

**US-2.** As a researcher, two look-alike rows that would both become new records still warn me once I set both to New.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Two `John Smith` Subjects, no handles | Both rows are unplaced, `Flags.PossibleDuplicate`, each `DuplicateOf` the other (today this is found through `TargetNew`) |
| AC-2.2 | Two `John Smith` Subjects, one handle `Smith` (surname only, a candidate under the weak bar for both) | Both rows are unplaced, and each is `DuplicateOf` the other (new: today these rows are Skip and are not checked) |
| AC-2.3 | Two `John Smith` Subjects, no handles, `s-a` fixed New | `s-a` keeps `TargetNew`; both rows name each other |
| AC-2.4 | Two `John Smith` Subjects, no handles, `s-a` fixed Skip | Neither row is flagged |
| AC-2.5 | A real project: a Source with two persons entered identically, no record | Both rows open unset; after setting both to New each row shows the "Also new" note |
| AC-2.6 | Swift: rows `c` and `d` arrive unplaced, each a possible duplicate of the other; the researcher sets both to New | Each shows `.alsoNew` naming the other |

**US-3.** As a researcher, when a suggested row loses its record on a re-proposal, its caption says no record is suggested.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Swift: a row loaded on handle `e1`, then merged with a proposal where that row has target `.unplaced`, reason `.unplaced`, `assessment` `"none"` | The row is `.unset`, `updated` is true, `captionReason` is `.unplaced` |
| AC-3.2 | `PromoteModel.reasonText(.unplaced, on:)` | `L10n.string(L10n.Promote.reasonUnplaced)` = "No record suggested", so the caption reads "Updated · No record suggested" |
| AC-3.3 | Wire vocabulary: `PromoteReasonCode(wire: "")`, `PromoteReasonCode(wire: "no_match")`, `PromoteTargetCode(wire: "")`, `PromoteTargetCode(wire: "handle")` | `.unplaced`; `.unrecognized("no_match")`; `.unplaced`; `.handle` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/graphalign/rows.go`: a row Align does not place and holds no decision for has an empty `Target`; its row reason is `ReasonUnplaced` (`""`), `ReasonEmpty` or `ReasonTaken`; alternatives, score and band unchanged | STEP-1, AC-1.1 – AC-1.5 |
| `ReasonNoMatch` deleted; `ReasonUnplaced` *(new)* = `""`; `ReasonWeak` documented as an alternative's reason only | STEP-1, D-2, D-3 |
| The possible-duplicate check (`flagAlikeNewRows`) checks rows that would be new records (`wouldBeNew`: unplaced with `ReasonUnplaced`, or held New) instead of `TargetNew` rows | STEP-1, AC-2.1 – AC-2.4, D-10 |
| Every Go test asserting `TargetNew` / `TargetSkip` / `ReasonNoMatch` / row `ReasonWeak` on an unplaced proposal row asserts an unplaced row | STEP-2, STEP-3 |
| `PromoteGraphAlignmentRow.target` and `.reason` comments in `engine.proto`, regenerated | STEP-3, D-6 |
| Swift: `PromoteReasonCode` and `PromoteFlow.Reason` lose `noMatch` and gain `unplaced`; the row's default reason is `.unplaced`; `swift_vocabulary_test.go` follows | STEP-4, AC-3.3, D-5 |
| Swift: `PromoteTargetCode` *(new)* types `CatalogPromoteGraphAlignmentRow.target`, decoded once in `GoStore`, with a guard row | STEP-4, AC-3.3, D-13 |
| L10n: `reasonNoMatch` removed, `reasonUnplaced` *(new)* "No record suggested"; `reasonWeak`'s comment names menu records | STEP-4, AC-3.2 |
| `PromoteFlow.draftTarget` drafts the proposed handle, the weak-band first alternative, or `.unset`; never `.skip` or `.newKind` | STEP-5, AC-1.6, AC-1.7 |
| `FakeStore.buildPromoteProposal` sends typed targets; its unmatched row sends `.unplaced` and reason `.unplaced` | STEP-5, D-8 |
| Comments this PR makes false (D-14) and the docs that say Promote proposes New or Skip (D-9) | STEP-1, STEP-5, STEP-6 |
| `add-match-rule` says how a row reason travels to the page (D-15) | STEP-6 |

| Out | Home |
| --- | --- |
| Decided New and Skip as fixed pairs, `TargetNew` / `TargetSkip` in `Fixed`, the batch and Done paths | Not planned: unchanged by design |
| Typing the send side: `CatalogPromoteGraphAlignmentFixed.target` and the batch row's `target` stay strings, spelled by `PromoteFlow.Target.token` and `batchRows` | Not planned here: an encode-side wire code needs a wire accessor `swift_vocabulary_test.go` does not read (its regex reads decode cases); the next PR that touches the batch rows takes it |
| Which record a weak band drafts (the first alternative) | Not planned: the map's scope entry keeps it |
| The `empty` and `taken` row reasons and their copy | Not planned: they still say something true about an unplaced row |
| `PromoteFlow.Reason.init(code:)`'s fallback for an unknown code (`.empty`) | Not planned: D-5 |
| `promote-find-another-record.md` line 9 ("A row with no candidate opens New (reason `no_match`)") | Not planned: it is that feature's problem statement as of S11-35; the spike folder is distilled at S11-99 |
| A Go constant for an empty row target | Not planned: D-1 |
| Changing the handler code in `api/ffi/handlers/promote_align.go` | Not planned: it copies `Target` and `Reason` as strings and needs no change (§4) |

## 4. Current state

Verified at `a70932e`.

| Area | Where | Today |
| --- | --- | --- |
| Unplaced row | `core/graphalign/rows.go` `proposal`, the final `else` branch (lines 58–84) | `row.Assessment = AssessNone`; with a best candidate: `Target = TargetSkip`, `Score`, `Alternatives = st.alts(sk, nil)`, `Reason = ReasonWeak`, band from the best score when at or above `Bands.Weak`; with none and `subjectHasValues`: `TargetNew`, `ReasonNoMatch`; with no values: `TargetSkip`, `ReasonEmpty`. Then `lostTo` overrides with `ReasonTaken` and the duplicate flags. Comment line 64: "Skip, never force a merge" |
| Held decision | `core/graphalign/rows.go` `proposal`, `st.held` branch | `Target = decided` (New or Skip), `AssessNone`, `ReasonDecided`. Unchanged |
| Alike-New check | `core/graphalign/rows.go` `flagAlikeNewRows` (lines 125–147), `alikeNew` (doc line 151: "two New subjects") | Visits rows with `Target == TargetNew` and no `PossibleDuplicate`, pairs them with other `TargetNew` rows of the kind through `alikeNew`, which adds neighbor-segment evidence (`newPairSegments`) to the pair's own. Today that set is unplaced no-candidate rows and held New rows; Skip rows (weak, empty, taken) are never checked |
| Types | `core/graphalign/types.go` | `TargetHandle/New/Skip`; `Reason` consts with `ReasonWeak // the best candidate is below the accept bar`, `ReasonNoMatch // no candidate at all: New`, `ReasonEmpty // nothing to match on: Skip`; `Row.Target` has no doc comment; `Alternative.Reason // via, agrees, or weak`; `Fixed`'s comment: an empty target on a `Fixed` means handle |
| Alternatives | `core/graphalign/rows.go` `alternative` | Always sets `Reason`: `ReasonWeak` by default, `ReasonVia` or `ReasonAgrees` when they apply. Never `""` |
| Align doc | `core/graphalign/align.go` `Align` | "Align proposes a handle, New, or Skip for every primary Subject on layer" |
| Fixed validation | `core/database/promotealign/promotealign.go` `checkFixed` | Accepts `TargetNew` / `TargetSkip` with no handle. Unchanged |
| Handler | `api/ffi/handlers/promote_align.go` `proposalProto` | Copies `string(r.Target)` and `string(r.Reason)`; reads `TargetHandle` only to attach headers. No code change needed |
| Proto | `api/proto/engine.proto` `PromoteGraphAlignmentRow` | Line 1669 `string target = 3; // handle, new, skip`; lines 1691–1692 field 20's two-line comment listing "via, decided, agrees, weak, taken, / no_match, empty". Generated: `api/proto/engine/engine.pb.go` line 13372 (target, trailing comment) and 13393–13394 (reason, two lines above the field); `macos/App/Platform/Generated/engine.pb.swift` line 5569 (`/// handle, new, skip` above `target`) and 5663–5664 (reason). `PromoteGraphAlignmentFixed.target` (pb.go 13771, pb.swift 5759) keeps "handle, new, skip". `protoc` may be missing in the implementer's environment |
| Swift wire vocabulary | `macos/App/Platform/ConclusionVocabulary.swift` `PromoteReasonCode` | Cases via, decided, agrees, weak, taken, noMatch, empty, three move codes, `unrecognized`; `""` decodes as `.unrecognized("")`. No enum for the row target |
| Vocabulary guard | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`, "promote reasons" row (lines 323–328) | Requires Swift's `case "…": self =` wire set to equal the Go reason consts, `ReasonNoMatch` included. Regex `case "([a-z_]*)": self = ` matches `case "":`; rows already list `""` literally (`sorted("", …)` at lines 336, 349) |
| Row model | `macos/App/Platform/GenealogyStore.swift` `CatalogPromoteGraphAlignmentRow` (line 320) | `var target: String`; `var reason: PromoteReasonCode = .empty` (line 329). `CatalogPromoteGraphAlignmentAlternative.reason` defaults `.empty` (313) |
| Decode | `macos/App/Platform/GoStore.swift` (line 986) | `target: r.target` raw; `reason: PromoteReasonCode(wire: r.reason)` |
| Swift reasons | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.Reason` | Has `weak`, `noMatch`, `empty`; `init(code:…)` maps `.unrecognized` to `.empty`. So a row reason `""` would read "Nothing to match on" |
| Swift raw target reads | `PromoteFlow.swift` lines 694 (`proposal.target == "handle"` in `makeRow`), 733, 748, 756 (`draftTarget`, `menu(for:)`) | String compares |
| Swift draft | `PromoteFlow.swift` `draftTarget` (line 727) | `.none` → `.unset`; proposed handle → handle; `taken` → `.unset`; `.weak` with a menu → first item; `.strong` and target `new` → `.newKind`; else `.skip` |
| Swift comments | `PromoteFlow.swift` line 10 (`Target.unset`: "A no-match row starts here so Skip and New are explicit."), line 321 ("An unselected no-match menu blocks Done.") | Speak of "no-match" rows |
| Swift caption | `PromoteFlow.Row.captionReason` (`move ?? reason`); drawn by `PromoteView` and `PromoteEvidenceSheet` through `PromoteModel.reasonText` | `reasonText` words `.weak` → `reasonWeak`, `.noMatch` → `reasonNoMatch`, `.empty` → `reasonEmpty` |
| Alternative readout | `PromoteFlow.Row.selectionReadout` | Reads the selected menu item's own `reason` (an alternative's `weak` stays worded by `reasonWeak`) |
| Duplicate notes | `PromoteFlow.noteDuplicates` | Runs after load, merge and every `select`; `.alsoNew` needs both rows `.newKind` |
| L10n | `macos/App/Platform/L10n.swift` lines 1344–1347; `macos/App/Resources/Localizable.xcstrings` keys `promote.reason.weak`, `promote.reason.noMatch` (line 17676), `promote.reason.empty` | `reasonNoMatch` "No matching record" is read only by `reasonText` |
| FakeStore | `macos/App/Platform/FakeStore.swift` `buildPromoteProposal` (lines 1799, 1816, 1830) | `target: "handle"` twice; the unmatched row `target: "skip"`, `assessment: "none"`, `reason: .empty` |
| Skill | `.agents/skills/add-match-rule/SKILL.md` line 78 | "A move reason is layer 6's: a `Reason` constant in `core/graphalign/types.go`." Nothing on row reasons or their Swift path |
| Go tests | `core/graphalign/align_test.go`: `TestGracieObituaryGolden` (543–548), `TestIsolatedSubjectScoresOnTheWalkScale` (630), `TestCandidateLoserDoesNotNominate` (1053), `TestPossibleDuplicates` (1218–1255), `TestRowReasons` (1328–1410), `TestNewDuplicateCheckComparesDirectedNeighborsByEnd` (1693), `TestBirthIsNeverProposedOnADeath` (1811, 1835, 1851), `TestLifespanRulesOutAPerson` (1882), `TestOnlyCandidatesOrNominatedHandlesAreCompared` (2088, 2127), `TestContestedRow` (2235); `core/database/promotealign/promotealign_test.go`: `TestSpellingOnlyToponymIsNotShortlisted` (593), `TestProposeRelationshipFromEitherEnd` (710), `TestProposeLifespanFromLifeEvents` (1009), `TestProposeSeesAHandlePromotedAfterTheIndexWasBuilt` (1161), `TestProposeDoesNotCompareAnUnshortlistedHandle` (1270); `api/ffi/handlers/promote_align_test.go`: `TestProposeSendsContested` (101), `TestProposalProtoSendsMove` (185–186) | Assert `TargetNew` / `TargetSkip` / `ReasonNoMatch` / row `ReasonWeak` on unplaced rows. The held-decision tests (`TestHeldNewAndSkipDecisions`, `TestRowReasons` "a held row is decided", `TestContestedRow` row `k`, `TestKinshipShapesCountOnce`, `TestProposeHoldsNewAndSkipDecisions`, `core/database/promotealign/evidence_test.go`, `core/database/promotealign/pins_test.go`, `core/database/promote/batch_test.go`) use fixed New/Skip and stay |
| Swift tests | `macos/ProvenenciaTests/PromoteFlowTests.swift`: helper `row(id:target:handleID:assessment:)` (line 10, `target: String = "handle"`), and fixture rows with `target: "new"` / `"skip"` at lines 173, 193, 257, 341, 356, 358, 361, 676, 830–832, 904, 907, 910, 947; `reasonCodesParseThroughTheWireVocabulary` (436–462); `macos/ProvenenciaTests/PromoteModelTests.swift` helper `row(_:handleID:assessment:)` (line 40, `target: "handle"`); `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift` lines 50–55 and 248–253 (`target: "handle"`) | String targets. `evidenceNeighborsAreTheRowsDoneFiles` (822–845) and `duplicateNotesNameTheOtherRowAndItsRef` (903–929) rely on Swift drafting `.newKind` / `.skip` from a strong New / Skip row |

**Precedent.** `9cc42ba4` ("Leave a taken row open…") changed `draftTarget` so a taken row opens unset whatever its band, with a `PromoteFlowTests` test that loops over bands. Copy its shape for the draft change and T-4. `edb419f1` ("promote: typed reasons…") is how a reason code travels: `graphalign.Reason` const → proto comment → `PromoteReasonCode` wire case → `PromoteFlow.Reason` → `PromoteModel.reasonText` → L10n. Copy that path for `unplaced`, and copy `PromoteReasonCode` (decoded once in `GoStore`, a guard row) for `PromoteTargetCode`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **An unplaced row's `Target` is the zero value `""`.** No new Go `Target` const. `Row.Target` gets the doc comment `// Target is TargetHandle when Align places the Subject; a held decision keeps its TargetNew or TargetSkip; otherwise empty: Align suggests nothing.` Go tests compare with `""` | The map's wording ("handle or empty"); `Fixed` already reads an empty target as "handle", so a named empty `Target` const would mean two things. The Swift guard lists `""` literally, as the cause rows do | `TargetNone Target = ""` — a name whose meaning flips between `Row` and `Fixed` |
| D-2 | **`ReasonNoMatch` is deleted. `ReasonUnplaced Reason = ""` *(new)***, comment `// Align placed nothing and nothing else applies; the page says no record is suggested`, listed after `ReasonTaken`. `ReasonEmpty`'s comment becomes `// nothing to match on`; `ReasonWeak`'s becomes `// an alternative below the weak bar: too little agrees to name` | The map: "no `weak` / `no_match` reason". The wire carries `""`, as the map says; a named const gives `wouldBeNew` (D-10) and the tests a name, as `autoreconcile.StateEmpty = ""` does. `ReasonWeak` stays because `alternative` still sets it | Leaving the reason unset with no const — `wouldBeNew` would test `Reason == ""`; keeping `ReasonWeak` on rows — the user's call says no |
| D-3 | **The unplaced branch** in `proposal` becomes, in this order: `row.Assessment = AssessNone`; with `best := st.bestNonAssigned(sk, nil)` non-nil: `row.Reason = ReasonUnplaced`, `row.Score = best.score`, `row.Alternatives = st.alts(sk, nil)`, `reported = best`, band from `best.score` when at or above `Bands.Weak` (unchanged); else if `st.subjectHasValues(s)`: `row.Reason = ReasonUnplaced`; else `row.Reason = ReasonEmpty`. The `lostTo` override is unchanged. No branch sets `row.Target`. The comment above the first branch reads `// A candidate existed but was not accepted (below the bar or lost one-to-one): suggest nothing, never force a merge. New and Skip are the researcher's.` | The smallest edit that removes both targets and both reasons and keeps score, band, alternatives and contested as they are | Folding the two `ReasonUnplaced` branches — `subjectHasValues` still picks `ReasonEmpty`, so the shape stays three-way |
| D-4 | **Caption copy** (user, 2026-10-27): an unplaced row's reason is worded "No record suggested". New `L10n.Promote.reasonUnplaced = LocalizedStringResource("promote.reason.unplaced", defaultValue: "No record suggested", comment: "Reason on a Promote row that Promote suggests no record for")`. `reasonNoMatch` and key `promote.reason.noMatch` are deleted. `reasonWeak`'s comment becomes "Reason on a record in a row's menu whose score is below the weak bar" (value unchanged) | The user chose one new line over keeping the two reasons or a bare badge | "Keep the two reasons" and "Badge only" — the user's answer |
| D-5 | **Swift names the reason case `unplaced`** in `PromoteReasonCode` (`case "": self = .unplaced`, first in `init(wire:)`) and `PromoteFlow.Reason` (doc `/// Promote suggests no record for this row.`). `Reason.init(code:…)` maps `.unplaced` → `.unplaced`; `.noMatch` is deleted from both enums and from `reasonText`. `CatalogPromoteGraphAlignmentRow.reason` defaults to `.unplaced` (the wire's zero value). `CatalogPromoteGraphAlignmentAlternative.reason` keeps its `.empty` default and `Reason.init`'s fallback for unknown codes stays `.empty`: Go's `alternative` always sends via, agrees or weak, so an alternative never carries `""`, and an unknown code's wording is not this PR's | `.none` would collide with `Optional.none` on `Reason?` (`move`) and `PromoteReasonCode?` (`moveReason`). The row default follows the wire so fixtures and FakeStore get the common case without an override | `.none` — ambiguous with optionals; changing the unknown-code fallback — a separate wording call |
| D-6 | **Proto: comments only.** `engine.proto` line 1669 becomes `  string target = 3; // handle, new or skip; empty when Promote suggests no record`. Lines 1691–1692 become `  // Why the row reads as it does: via, decided, agrees, taken, empty, or` and `  // nothing when Promote suggests no record. The page words it. With agrees, the agreeing Property.` Regenerate with `./scripts/generate-proto.sh`. If `protoc` is not installed, hand-edit the generated files to the bytes the generators emit: `engine.pb.go` line 13372's trailing comment `// handle, new or skip; empty when Promote suggests no record`, and lines 13393–13394 as `	// Why the row reads as it does: via, decided, agrees, taken, empty, or` / `	// nothing when Promote suggests no record. The page words it. With agrees, the agreeing Property.` (tab-indented); `engine.pb.swift` line 5569 `  /// handle, new or skip; empty when Promote suggests no record` and lines 5663–5664 `  /// Why the row reads as it does: via, decided, agrees, taken, empty, or` / `  /// nothing when Promote suggests no record. The page words it. With agrees, the agreeing Property.`. Say in the PR body which way was used. `PromoteGraphAlignmentFixed.target` is untouched | The wire shape is unchanged; generators copy field comments line for line, so the hand edit matches them. A held decision still sends new or skip on a row, so the comment keeps them | A new field or enum — nothing new travels |
| D-7 | **`draftTarget`** after this PR, in order: `assessment == .none` → `.unset`; `proposal.target == .handle` and the proposed handle first in the menu → `.handle`; `Reason(proposal) == .taken` → `.unset`; `assessment == .weak` with a menu item → first item; otherwise `.unset`. The two lines that return `.newKind` and `.skip` go. The first comment becomes `// No suggestion: leave the choice open so Skip and New are the researcher's.` | The map's scope entry; a medium or strong row that was not placed and not taken (a settle or a later registry) now opens unset instead of Skip | Drafting the first alternative for medium and strong — the map keeps "the weak-band first alternative as today" only |
| D-8 | **FakeStore**: lines 1799 and 1816 `target: .handle`; the unmatched row (1830) `target: .unplaced`, `reason: .unplaced`, everything else as today | The fake mirrors Go; its rows have no values to say `empty` about | Keeping `.empty` — it would word "Nothing to match on" for a fake person with a name |
| D-9 | **Docs sweep**, word for word. `docs/promote-graph-alignment.md`: line 33 "such a row comes back **Skip** (New only when it has no candidate at all). The page starts a row with no match **unset**, so Done waits for the researcher's Skip or New, and drafts a weak one on its best match for review." → "such a row comes back with no target, its candidates as alternatives. The page starts it **unset**, so Done waits for the researcher's Skip or New, and drafts a weak one on its best match for review."; line 52 "or two New rows whose properties clear the accept bar" → "or two rows that would both be new records (unplaced or decided New) whose properties clear the accept bar"; line 73 "Align proposes a handle, New or Skip for every" → "Align proposes a handle or nothing for every"; line 108 "A candidate that cleared no bar comes back Skip with the candidate as an alternative (New only when there is no candidate at all);" → "A row with no accepted candidate comes back with no target and its candidates as alternatives;"; line 125 "a match below the bar comes back Skip, never New, and a row with no match starts unset on the page." → "a match below the bar comes back as an alternative with no target, and the row starts unset on the page.". `docs/ideas/coverage-aware-similarity.md` line 7 "so the row is Skip `weak` with the handle as an unbanded alternative" → "so the row has no target, with the handle as an unbanded alternative". `docs/deployment-plan/spike-11/name-candidates.md` line 11 "a Subject with values opens New (`no_match`) and an empty menu" → "a row opens unset with an empty menu"; line 69 "Go today gives every row it does not place a target:" → "Until S11-47, Go gave every row it did not place a target:" and "Swift drafts neither." → "Swift drafted neither."; line 73 "finds its rows by `TargetNew` today; S11-47 finds them as unplaced rows instead" → "found its rows by `TargetNew` before S11-47; it now checks unplaced rows and decided New rows". `core/graphalign/align.go` `Align`'s comment: "proposes a handle, New, or Skip for every primary Subject" → "proposes a handle or nothing for every primary Subject" | Each PR updates the docs it makes false | Rewording `promote-find-another-record.md` — Out |
| D-10 | **One predicate, `wouldBeNew(r Row) bool`** *(new, private, in `rows.go` above `flagAlikeNewRows`)*: `return r.Target == TargetNew \|\| (r.Target == "" && r.Reason == ReasonUnplaced)`, doc `// wouldBeNew reports whether filing the row New is the open question: a decided New, or a row Promote suggests no record for. A taken row already carries its duplicate, an empty row has nothing to compare, and a Skip files nothing.` `flagAlikeNewRows` visits `i` when `wouldBeNew(rows[i]) && !rows[i].Flags.PossibleDuplicate` and pairs it with `j` when `wouldBeNew(rows[j])` and the kinds match (both copies of the old `TargetNew` test go). Its doc's first sentence becomes `// flagAlikeNewRows marks two rows of one kind that would both be new records (wouldBeNew) and would clear the accept bar against each other: filing both New mints two handles for what may be one entity.` `alikeNew`'s doc "two New subjects" → "two subjects that would both be new records" | The map: "found among unplaced rows". Held New rows stay in, because Swift sends both rows New before it shows "also new". Taken and empty rows stay out, as they were (both were Skip), so the set grows only by the below-bar-candidate rows the map names. `alikeNew` adds neighbor evidence, so an empty row is excluded by name, not by its score | A negative test `Target != TargetHandle && Target != TargetSkip` written twice — pulls in taken and empty rows silently |
| D-11 | **Go test edits are mechanical**: each assertion of `TargetNew` + `ReasonNoMatch` or `TargetSkip` + `ReasonWeak` on an unplaced row becomes `Target == ""` + `Reason == graphalign.ReasonUnplaced`; every other part of the assertion keeps its meaning, and messages say "unplaced" or "no target" for "New"/"Skip". The exact list is in STEP-2 and STEP-3 | AC-1.x are exactly the old assertions with the new target and reason | Deleting tests whose subject was the target — they also pin bands and alternatives |
| D-12 | **`TestPossibleDuplicates`** table gains `fixed []graphalign.Fixed` and `taken string` (the subject that must read `ReasonTaken`, empty for none). The loop's `r.Target == graphalign.TargetSkip && flagged && r.Reason != graphalign.ReasonTaken` check becomes `string(r.SubjectID) == tt.taken && r.Reason != graphalign.ReasonTaken`. Case 1 sets `taken: "s-b"`. `TestRowReasons` gains a `wantTarget graphalign.Target` column, asserted beside the reason (D-16) | The old check keyed "taken" on Skip, which no longer exists | Keying on `Target == ""` — the new below-bar case is unplaced and flagged but not taken |
| D-13 | **`PromoteTargetCode`** *(new)* in `macos/App/Platform/ConclusionVocabulary.swift`, after `PromoteReasonCode`: doc `/// What a Promote proposal row targets (\`graphalign.Target\`): a record, a held decision, or nothing.`; cases `unplaced` (doc `/// Promote suggests no record.`), `handle`, `new`, `skip`, `unrecognized(String)`; `init(wire:)` with `case "": self = .unplaced`, `case "handle": self = .handle`, `case "new": self = .new`, `case "skip": self = .skip`, `default: self = .unrecognized(wire)`. `CatalogPromoteGraphAlignmentRow.target` becomes `PromoteTargetCode` (no default); `GoStore` line 986 decodes `PromoteTargetCode(wire: r.target)`; `PromoteFlow` lines 694, 733 and 756 compare `== .handle`. Guard row in `TestSwiftVocabularyMatchesGo` after "promote reasons": `{name: "promote row targets", path: swiftVocabulary, enum: "PromoteTargetCode", re: wireCase, want: sorted("", string(graphalign.TargetHandle), string(graphalign.TargetNew), string(graphalign.TargetSkip))}`. Test helpers take `target: PromoteTargetCode = .handle` | `.cursor/rules/macos-client.mdc`: "Go wire vocabulary is spelled once in Swift, as a typed enum" and "every such enum is a row there". This PR changes the meaning of the row target, the one Promote wire field still read as a string; STEP-5 rewrites every fixture that spells it anyway | Leaving it a string — the new `""` meaning and the removed New/Skip meanings would be guarded by nothing; typing the send side too — Out |
| D-14 | **Comments this PR makes false**: `PromoteFlow.Target.unset` doc → `/// No choice yet. Every row Promote suggests no record for starts here, so Skip and New are explicit.`; `targetsChosen`'s doc → `/// Every visible row has a target. An unset row blocks Done.`; `PromoteFlowTests.swift` line 191 "only a no-match stays empty" → "a row with no band, or a medium or strong one Promote did not place, stays empty"; `rows.go` line 64 per D-3; `alikeNew` per D-10 | Comments stay true for the next reader | — |
| D-15 | **`add-match-rule` line 78** becomes: "A row or move reason is layer 6's: a `Reason` constant in `core/graphalign/types.go`, its `PromoteReasonCode` wire case with its row in `TestSwiftVocabularyMatchesGo`, its `PromoteFlow.Reason` case, and its `PromoteModel.reasonText` arm with an L10n line." | The next reason follows the skill, not an old commit | — |
| D-16 | **Tests extend the existing tables** instead of adding parallel ones: `TestRowReasons` takes the target column and a "taken" row (T-1); T-4 covers only the medium and strong bands, the new behavior (the weak and none bands are `aWeakCandidateBelowTheBarStartsSelected` and `anUnmatchedRowStartsUnselected`, edited in STEP-5) | One table per concern | A new `TestUnplacedRowsCarryNoTarget` over the same fixtures |

**Rules this plan bends.** None. D-13 follows `.cursor/rules/macos-client.mdc` for the field this PR changes; the send side's strings predate this PR and are Out with a home.

## 6. Design

### Codebase fit

Searched: `rg -n "ReasonNoMatch|ReasonWeak|TargetNew|TargetSkip" core api`, `rg -n "noMatch|reasonNoMatch|case \"\": self" macos/App`, `rg -n 'target: "|target == "' macos`, `rg -n "promote.reason" macos/App/Resources/Localizable.xcstrings`, `rg -n "draftTarget|captionReason|noteDuplicates" macos`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphalign.ReasonUnplaced` *(new)* | const, `core/graphalign/types.go` | `ReasonNoMatch` (deleted), `ReasonEmpty` | `no_match` meant "New"; `empty` means no values (D-2) |
| `wouldBeNew` *(new)* | private func, `core/graphalign/rows.go` | the two inline `Target == TargetNew` tests in `flagAlikeNewRows` | Replaces both with one named set (D-10) |
| `PromoteTargetCode` *(new)* | enum, `ConclusionVocabulary.swift` | `PromoteReasonCode` | Same shape for the other row field; required by `macos-client.mdc` (D-13) |
| `PromoteReasonCode.unplaced` *(new)*, `PromoteFlow.Reason.unplaced` *(new)* | enum cases | `.noMatch` (deleted), `.empty` | Replaces `.noMatch` with the new meaning; `.empty` is worded "Nothing to match on" (D-5) |
| `L10n.Promote.reasonUnplaced` *(new)*, key `promote.reason.unplaced` *(new)* | L10n key | `reasonNoMatch` (deleted) | New wording chosen by the user (D-4) |
| `anUnplacedRowNeverDraftsNewOrSkip` *(new)*, `aRowThatLosesItsRecordSaysNoRecordIsSuggested` *(new)* | Swift tests, `PromoteFlowTests.swift` | `aRowWhoseRecordWasTakenStartsUnset`, `decidedRowsSurviveAReproposal` | Same shapes (loop over bands; merge a re-proposal), new assertions |
| `anUnplacedRowIsWordedNoRecordSuggested` *(new)* | Swift test, `PromoteModelTests.swift` | `theSheetCarriesTheRowsCaption` | One `reasonText` assertion on the new case |

**Builds on.** `graphalign`'s row building in `rows.go` (S11-16), the typed reason path of `edb419f1`, the taken-row draft of `9cc42ba4`, the wire-vocabulary guard `swift_vocabulary_test.go`.
**Deletes.** `graphalign.ReasonNoMatch`; the `TargetNew` / `TargetSkip` and `ReasonWeak` / `ReasonNoMatch` assignments in `proposal`; the two inline `TargetNew` tests in `flagAlikeNewRows`; `PromoteReasonCode.noMatch`; `PromoteFlow.Reason.noMatch`; `reasonText`'s `.noMatch` arm; `L10n.Promote.reasonNoMatch` and `promote.reason.noMatch` in `Localizable.xcstrings`; `draftTarget`'s `.newKind` and `.skip` returns; every raw `"handle"` / `"new"` / `"skip"` string that reads a proposal row's target in `PromoteFlow`, `FakeStore` and the tests. `ReasonWeak`, `.weak` and `reasonWeak` stay: alternatives read them.
**Next adds.** S11-45 (prefix names) adds candidates that open unset with no target. Any later row reason follows `add-match-rule` (D-15). A later target value is one `PromoteTargetCode` case and a guard entry.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A proposal row with target `.unplaced` and a band of medium or strong (not taken) | `PromoteFlow.draftTarget` → `.unset`; `PromoteView` target menu | `.unset` is already a row state (no band, placeholder "Select where this files", `targetsChosen` false); `selectionReadout` is nil for it, so no badge |
| A row reason `""` | `PromoteReasonCode(wire:)` → `.unplaced` → `PromoteFlow.Reason.unplaced` → `reasonText` | New case handled in every `switch` (the compiler enforces it); shown only in the Updated caption |
| A target code Swift does not know | `PromoteTargetCode.unrecognized` | Not `.handle`, so `draftTarget` opens it unset, as an unknown string did |
| An unplaced row with a below-bar candidate is now checked by `flagAlikeNewRows` | `PromoteFlow.noteDuplicates` (`proposedDuplicateOf`) | The note shows only when the other row has a handle (`sharesHandle`) or both rows are `.newKind` (`alsoNew`); an unset pair shows nothing until the researcher chooses |

### Scale

At the map's Scale (a 12-generation tree, the large-source budgets): `flagAlikeNewRows` now also visits rows with a below-bar candidate. It is pairwise over `wouldBeNew` rows of one kind, as it was over New rows; on a large Source most unplaced rows had no candidate and were already visited. `TestLargeSourceBudgets` (check-pr) bounds it; going over budget is a stop condition.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/graphalign` | `types.go`: `ReasonUnplaced` *(new)*, `ReasonNoMatch` deleted, comments per D-1, D-2. `rows.go`: unplaced branch per D-3, `wouldBeNew` and `flagAlikeNewRows` per D-10. `align.go`: `Align` comment per D-9 |

### FFI

`api/proto/engine.proto`: comments on `PromoteGraphAlignmentRow` fields 3 and 20 (D-6). Regenerate with `./scripts/generate-proto.sh`. Handler `api/ffi/handlers/promote_align.go` unchanged. `api/ffi/swift_vocabulary_test.go`: the "promote reasons" row and the new "promote row targets" row (D-13).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Target wire code | `PromoteTargetCode` | `macos/App/Platform/ConclusionVocabulary.swift` | D-13 |
| Reason wire code | `PromoteReasonCode.unplaced` | `macos/App/Platform/ConclusionVocabulary.swift` | D-5 |
| Row model | `CatalogPromoteGraphAlignmentRow` | `macos/App/Platform/GenealogyStore.swift` | `target: PromoteTargetCode`; `reason` default `.unplaced` |
| Decode | `GoStore` proposal mapping | `macos/App/Platform/GoStore.swift` | D-13 |
| Row draft | `PromoteFlow.draftTarget` | `macos/App/Features/Promote/PromoteFlow.swift` | D-7 |
| Row reason | `PromoteFlow.Reason.unplaced` | `macos/App/Features/Promote/PromoteFlow.swift` | D-5 |
| Caption wording | `PromoteModel.reasonText` | `macos/App/Features/Promote/PromoteModel.swift` | `.unplaced` → `L10n.Promote.reasonUnplaced` |
| Fake proposal | `FakeStore.buildPromoteProposal` | `macos/App/Platform/FakeStore.swift` | D-8 |

States: an unplaced row is the existing `.unset` state. Accessibility: the caption is a `Text`; no new control.

### Strings

L10n keys *(new)*: `L10n.Promote.reasonUnplaced` = "No record suggested" (`promote.reason.unplaced`). Removed: `L10n.Promote.reasonNoMatch` "No matching record" (`promote.reason.noMatch`). Comment changed: `promote.reason.weak` (D-4). Old-scope copy: `promote.reason.weak` "Too little agrees to suggest a match" now speaks only of a menu record, which its wording already fits; no other string names a proposed New or Skip.

## 7. Steps

Every Go test command is run with `CGO_ENABLED=1` and `-tags fts5`, from the repo root. The Swift tests run only on macOS (`xcodebuild` through `scripts/check-pr.sh`).

**STEP-1 — Go: unplaced rows carry no target**
Files: `core/graphalign/types.go`, `core/graphalign/rows.go`, `core/graphalign/align.go`. Skill: `add-match-rule`.
Do: D-1, D-2, D-3, D-10, D-14's `rows.go` lines, and D-9's `Align` comment.
Check: `CGO_ENABLED=1 go build -tags fts5 ./core/... ./api/...` passes (no non-test file used `ReasonNoMatch`). `CGO_ENABLED=1 go vet -tags fts5 ./core/graphalign/ ./core/database/promotealign/ ./api/ffi/...` reports only test files that still name `graphalign.ReasonNoMatch`, which STEP-2 and STEP-3 fix.

**STEP-2 — Go: graphalign tests**
Files: `core/graphalign/align_test.go`. Rule: `go-tests.mdc`.
Do: per D-11, D-12, D-16. `TestGracieObituaryGolden` (543–548): the first message "should be New or Skip" → "should have no target"; the second check becomes `r.Target != ""` with message "place %s: %+v, want no target". `TestIsolatedSubjectScoresOnTheWalkScale` "a different first name is no candidate": `wantTarget: ""`, `wantReason: graphalign.ReasonUnplaced`. `TestCandidateLoserDoesNotNominate`: `row.Target != "" || row.Reason != graphalign.ReasonUnplaced || row.Via != nil`. `TestRowReasons`: T-1. `TestNewDuplicateCheckComparesDirectedNeighborsByEnd`: `r.Target != ""`, message "want unplaced and a possible duplicate". `TestBirthIsNeverProposedOnADeath` (three sites), `TestLifespanRulesOutAPerson`, `TestOnlyCandidatesOrNominatedHandlesAreCompared` (two sites): `Target != ""` and `Reason != graphalign.ReasonUnplaced`, messages "unplaced". `TestContestedRow` row `n`: `r.Target != ""`, message "an unplaced row is not contested". `TestPossibleDuplicates`: D-12 plus T-3's cases.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/...` passes.

**STEP-3 — Go: catalog and FFI tests, proto comments**
Files: `core/database/promotealign/promotealign_test.go`, `api/ffi/handlers/promote_align_test.go`, `api/ffi/swift_vocabulary_test.go`, `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: `promotealign_test.go`: lines 593, 1009, 1270 → `Target != "" || Reason != graphalign.ReasonUnplaced` (1270 keeps `len(row.Alternatives) != 0`); 710 → same, message "want no target (under the accept bar)"; 1161 → `first.Target != ""`, message "want no target". `promote_align_test.go`: `TestProposeSendsContested` case 1 wants target `""`, reason `""`; `TestProposalProtoSendsMove` rows drop `Target: graphalign.TargetSkip` (zero value). `swift_vocabulary_test.go`: "promote reasons" replaces `graphalign.ReasonNoMatch` with `graphalign.ReasonUnplaced`; add the "promote row targets" row (D-13). D-6 for the proto.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./api/ffi/handlers/...` passes. `TestSwiftVocabularyMatchesGo` fails on exactly "promote reasons" and "promote row targets" until STEP-4. `git diff api/proto/engine/engine.pb.go macos/App/Platform/Generated/engine.pb.swift` shows only the three comment changes of D-6.

**STEP-4 — Swift: wire codes, the reason and its words**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Features/Promote/PromoteFlow.swift` (`Reason`, lines 694 / 733 / 756), `macos/App/Features/Promote/PromoteModel.swift` (`reasonText`), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-4, D-5, D-13. In `Localizable.xcstrings`, delete the `promote.reason.noMatch` entry and add `promote.reason.unplaced` beside `promote.reason.taken` in the same shape (`extractionState` `manual`, `en` `translated`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `python3 scripts/check-localizable-xcstrings.py` passes; `rg -n "\.noMatch\b|reasonNoMatch|promote\.reason\.noMatch" macos/App` prints only `EvidencePropertyCause`'s `noMatch` lines in `ConclusionVocabulary.swift`.

**STEP-5 — Swift: draft, fake, tests**
Files: `macos/App/Features/Promote/PromoteFlow.swift` (`draftTarget`, D-14 comments), `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`, `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift`. Skill: `add-swift-test`.
Do: D-7, D-8, D-14. Test fixtures: `PromoteFlowTests` helper `row` takes `target: PromoteTargetCode = .handle`; `PromoteModelTests` helper `row` and the two `PromoteEvidenceBoardTests` rows pass `target: .handle`. In `PromoteFlowTests.swift`, every fixture row with `target: "new"` or `target: "skip"` (lines 173, 193, 257, 341, 356, 358, 361, 676, 830–832, 904, 907, 910, 947) passes `target: .unplaced`. `aWeakCandidateBelowTheBarStartsSelected` line 194 `below.reason = .weak` → `.unplaced`. `unmatched(_:)` sets `reason = .unplaced` and its doc reads "A row Promote found no record for: no target, no band, no alternatives, so it opens unset.". `reasonCodesParseThroughTheWireVocabulary`: T-7. `evidenceNeighborsAreTheRowsDoneFiles`: after `load`, call `flow.setTarget(subjectID: "n", token: "new")` and `flow.setTarget(subjectID: "s", token: "skip")` before the assertions, which stay. `duplicateNotesNameTheOtherRowAndItsRef`: after `load`, call `flow.setTarget(subjectID: "c", token: "new")` and `flow.setTarget(subjectID: "d", token: "new")` before the `note("c")` / `note("d")` assertions, which stay (AC-2.6); row `b`'s `sharesHandle` assertion stays. Add T-4, T-5, T-6.
Check: `rg -n 'target: "(new|skip|handle)"' macos/ProvenenciaTests macos/App/Platform/FakeStore.swift` prints only `CatalogPromoteGraphAlignmentFixed(...)` and batch-row lines (`PromoteModelTests.swift` 412–413 and FakeStore's batch `switch` at 1879–1907). On macOS, `xcodebuild` passes through STEP-7.

**STEP-6 — Docs and skill**
Files: `docs/promote-graph-alignment.md`, `docs/ideas/coverage-aware-similarity.md`, `docs/deployment-plan/spike-11/name-candidates.md`, `.agents/skills/add-match-rule/SKILL.md`.
Do: D-9 and D-15, word for word.
Check: `rg -n "comes back Skip|comes back \*\*Skip|New or Skip for every|Skip \`weak\`|opens New \(\`no_match\`\)" docs/promote-graph-alignment.md docs/ideas docs/deployment-plan/spike-11/name-candidates.md` prints nothing; `rg -n "A row or move reason is layer 6" .agents/skills/add-match-rule/SKILL.md` prints one line.

**STEP-7 — Whole suite**
Files: none.
Do: run `scripts/check-pr.sh`.
Check: it prints `check-pr: PASS` (on macOS it includes the Xcode tests; elsewhere report that the Swift half was not run). `rg -n "ReasonNoMatch|reasonNoMatch" core api macos docs/*.md .agents .cursor` prints nothing.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done, once the PR number is known.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-47-suggest-a-record-or-nothing.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. STEP-1 to STEP-5 are one commit ("Leave a row Promote cannot place unset instead of drafting New or Skip, and say no record is suggested."), so the vocabulary guard is never red at a commit; STEP-6 and STEP-8 are docs commits. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/graphalign/align_test.go` `TestRowReasons` (extended) | New column `wantTarget`, asserted with the reason (`row.Target != tt.wantTarget` fails with the target in the message). "a property-only match names the agreeing property" and "a neighbor-reached row is via": `TargetHandle`. "a neighbor-reached candidate below the bar is weak" → renamed "a neighbor-reached candidate below the bar suggests nothing", `""`, `ReasonUnplaced`. "a surname-only look-alike is offered as weak" → renamed "a surname-only look-alike is only an alternative", `""`, `ReasonUnplaced`. "no candidate is no match" → renamed "no candidate suggests nothing", `""`, `ReasonUnplaced`. "no values is empty": `""`, `ReasonEmpty`. "a held row is decided": `TargetSkip`, `ReasonDecided`. New "a row whose handle another row took is taken": layer `s-a` Ref `PER-A` and `s-b` Ref `PER-B`, both `nameVals("John", "Smith")`, handle `h-1` `nameVals("John", "Smith")`, subject `s-b` → `""`, `ReasonTaken` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, D-3 |
| T-2 | Go | `TestHeldNewAndSkipDecisions` (existing), `TestOnlyCandidatesOrNominatedHandlesAreCompared` and `TestBirthIsNeverProposedOnADeath` (edited, STEP-2) | Held rows keep `TargetNew` / `TargetSkip`; unplaced rows keep their alternatives | AC-1.2, AC-1.5 |
| T-3 | Go | `core/graphalign/align_test.go` `TestPossibleDuplicates` (extended) | Case 1 sets `taken: "s-b"`. The existing "two New rows…" case is renamed "two unplaced rows that match each other name each other". New cases: "two unplaced rows that share a below-bar candidate name each other": handles `[{ID: id("h-1"), Ref: "PER-1", Kind: "person", Values: surnameVals("Smith")}]` → `s-a`↔`s-b`. "a row decided New and its look-alike name each other": `fixed: [{SubjectID: id("s-a"), Target: graphalign.TargetNew}]` → `s-a`↔`s-b`. "a row decided Skip is not checked": `fixed: [{SubjectID: id("s-a"), Target: graphalign.TargetSkip}]` → no flags | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-10, D-12 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `anUnplacedRowNeverDraftsNewOrSkip` *(new)* | For each band `"medium"`, `"strong"`: `row(id: "a", target: .unplaced, handleID: "", assessment: band)` with `reason = .unplaced` and its default alternative `e2` at that band, loaded with `fact("a")` → target `.unset`, `targetsChosen == false`, message "band \(band)" | AC-1.6, D-7 |
| T-5 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aRowThatLosesItsRecordSaysNoRecordIsSuggested` *(new)* | Load `[row(id: "a"), row(id: "b", handleID: "e1")]` with `fact("a"), fact("b")`, then `merge` a proposal (revision 4) where `b` is `row(id: "b", target: .unplaced, handleID: "", assessment: "none")` with `reason = .unplaced`, `alternatives = []` → `b.target == .unset`, `b.updated == true`, `b.captionReason == .unplaced` | AC-3.1 |
| T-6 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `anUnplacedRowIsWordedNoRecordSuggested` *(new)* | `let model = await loadedAdaAndBea(row("a", handleID: "e1"), row("b", handleID: "e2"))`; `bea = model.flow.row("b")!` → `model.reasonText(.unplaced, on: bea) == L10n.string(L10n.Promote.reasonUnplaced)` and that string is "No record suggested" | AC-3.2, D-4 |
| T-7 | Swift | `PromoteFlowTests.swift` `reasonCodesParseThroughTheWireVocabulary` (edited) | Replace `("no_match", .noMatch)` with `("", .unplaced)`; add `#expect(PromoteReasonCode(wire: "no_match") == .unrecognized("no_match"))`; replace `reason(.noMatch) == .noMatch` with `reason(.unplaced) == .unplaced`; add `PromoteTargetCode(wire:)` for `""` → `.unplaced`, `"handle"` → `.handle`, `"new"` → `.new`, `"skip"` → `.skip`, `"later"` → `.unrecognized("later")` | AC-3.3, D-5, D-13 |
| T-8 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "promote reasons" (edited) and "promote row targets" *(new row)* | Go sets equal the Swift `PromoteReasonCode` and `PromoteTargetCode` wire sets | AC-3.3, D-2, D-13 |
| T-10 | Swift | `PromoteFlowTests.swift` `aWeakCandidateBelowTheBarStartsSelected` (edited, STEP-5: `target: .unplaced`, `reason = .unplaced`) | A weak-band unplaced row with alternative `e9` → `.handle(id: "e9", ref: "PER-9", title: "PER-9")`, as today | AC-1.7, D-7 |
| T-9 | Swift | `PromoteFlowTests.swift` `duplicateNotesNameTheOtherRowAndItsRef` (edited, STEP-5) | Rows `c`, `d` arrive `.unplaced`; after setting both to New → `.alsoNew` each | AC-2.6, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/graphalign` (`TestGracieObituaryGolden`), `core/database/promotealign` (goldens, `TestLargeSourceBudgets`), `api/ffi` (`TestSwiftVocabularyMatchesGo`), `PromoteModelTests` that load a FakeStore proposal and read an unmatched row.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: AC-1.1–1.5 → T-1, T-2; AC-1.6 → T-4; AC-1.7 → T-10; AC-2.1–2.4 → T-3; AC-2.6 → T-9; AC-3.1 → T-5; AC-3.2 → T-6; AC-3.3 → T-7, T-8; AC-1.8, AC-2.5 → manual check
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; promote.reason.noMatch is gone from both
- [ ] Sweep: `ReasonNoMatch`, `reasonNoMatch`, `PromoteReasonCode.noMatch` and `promote.reason.noMatch` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — n/a, none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no chrome
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status Landed #N; completed.md: one short paragraph of what shipped, newest first, in the voice of the S11-44 entry
- [ ] Brief moved to design/archive/ and design/README.md updated — n/a, no brief
- [ ] Domain docs updated: docs/promote-graph-alignment.md, docs/ideas/coverage-aware-similarity.md, name-candidates.md (D-9); add-match-rule (D-15)
- [ ] PR description: what shipped, AC evidence, deviations (including whether D-6's generated files were regenerated or hand-edited), manual check
```

## 10. Manual check

1. Open a project with a person record entered as surname `Ames` only. Add a Source whose Citation records four persons: given `Zed`, surname `Zimm` (no record like him); given `Ann`, surname `Ames` (the `Ames` record is a candidate under the weak bar); and two persons both given `John`, surname `Smith`.
2. Open Promote for that Source. See: the "Zed Zimm" row opens unset ("Select where this files") with no record in its menu; the "Ann Ames" row opens unset with "Ames" in its menu, not on Skip; both "John Smith" rows open unset (AC-1.8).
3. Set both "John Smith" rows to New. See: each shows the "Also new" note naming the other (AC-2.5).
4. Set "Zed Zimm" to New and "Ann Ames" to Skip, then Done. See: a new record for Zed Zimm and one for each John Smith; Ann Ames is not filed (AC-1.8).

This makes slice 10's Check line about "Amy" opening unset with "Amy Lee Anne Long" in its menu hold without a Skip fallback.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A code path reads an unplaced row's `TargetNew` to mean "will be new" | Searched (§4): only `flagAlikeNewRows` did; D-10 moves it. `promotealign` reads `TargetNew` / `TargetSkip` only from `Fixed` |
| More rows enter the alike-New check on a large Source | `wouldBeNew` adds only below-bar-candidate rows; `TestLargeSourceBudgets` runs in check-pr |
| `protoc` missing in the implementer's environment | D-6 gives the exact hand-edit text |
| The Swift half cannot run off macOS | STEP-7 says to report it; the reviewer runs check-pr on macOS |

**Stop and report if:**

- `core/graphalign/rows.go`'s unplaced branch no longer assigns `TargetSkip` / `TargetNew` / `ReasonWeak` / `ReasonNoMatch` as §4 says (S11-45 or another PR moved it).
- A non-test Go file fails to build after `ReasonNoMatch` is deleted — name the file.
- Any Go or Swift test not named in STEP-2, STEP-3 or STEP-5 fails — name it and its output; do not edit it.
- `TestLargeSourceBudgets` goes over budget.
- T-3's "two unplaced rows that share a below-bar candidate" is not flagged: `alikeNew` does not pair them, so D-10's claim is wrong.
- T-9's `.alsoNew` assertions fail after both rows are set to New: `noteDuplicates` does not read the proposal's duplicate once a row is decided.
- The generated-file diff after D-6 shows anything other than the three comment changes.
- In the manual check, a row opens on Skip or New without the researcher choosing it.

## 12. Open questions

None.

## 13. PR

**Title:** S11-47 — Promote suggests a record or nothing
**Body:** objective in one sentence; what shipped against US-1, US-2, US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-47-suggest-a-record-or-nothing.md`.
