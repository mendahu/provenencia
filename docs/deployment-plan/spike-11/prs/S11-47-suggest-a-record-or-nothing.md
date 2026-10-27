# S11-47 — Promote suggests a record or nothing

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 10 — Name candidates
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the row, its menu, its Updated badge and caption already exist; this PR changes which target a row opens on and one caption sentence)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s11-47-suggest-a-record-or-nothing`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [name candidates § Promote suggests a record or nothing](../name-candidates.md#promote-suggests-a-record-or-nothing), [`promote-graph-alignment.md` § 3](../../../promote-graph-alignment.md#3-the-flow-one-page), map scope entry [S11-47](../deployment-plan.md#s11-47--promote-suggests-a-record-or-nothing); working rules: "Promote suggests a record or nothing", "High tolerance for churn"
**Planned:** 2026-10-27 against `main` @ `a70932e`
**Review:** — (not reviewed yet)

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

Promote proposes a record for a row, or nothing. Today Go gives every row it does not place a target the researcher never chose: Skip (reason `weak`) when a candidate fell short, New (reason `no_match`) when there was none. Swift mostly ignores them and opens the row unset, but a medium or strong row that was not placed and not taken still falls through to Skip, and the row reasons word a suggestion Promote did not make. After this PR a proposal row carries a handle or no target, with its alternatives as today. New and Skip are only ever the researcher's choice. When a suggested row loses its record on a re-proposal, its caption reads "Updated · No record suggested".

**Why now.** The user's call of 2026-10-24, answering S11-43's open question ([detail](../name-candidates.md#promote-suggests-a-record-or-nothing)); the caption wording is the user's call of 2026-10-27 (D-4).

**What stays.** Which candidates exist, every band and score, each alternative's own band and reason (`ReasonWeak` on an alternative stays), the weak-band draft onto the first alternative, a taken row opening unset with its duplicate note, the `empty` and `taken` row reasons, and decided New and Skip: the page still sends them as fixed pairs (`PromoteGraphAlignmentFixed`) and Go still holds them with `TargetNew` / `TargetSkip`.

## 2. User stories

**US-1.** As a researcher, when Promote finds no record for a row, or finds one that falls short, the row opens with nothing chosen so that New and Skip are only ever my decision.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A person Subject `Ann Ames` and no candidate (`graphalign.Align`, nothing fixed) | Row `Target` is empty, `Reason` is `ReasonUnplaced` (`""`), `Assessment` `AssessNone`, no alternatives |
| AC-1.2 | A person Subject `Ann Ames` and one handle `Ames` (surname only, a candidate under the weak bar) | Row `Target` is empty, `Reason` `ReasonUnplaced`, `Assessment` `AssessNone`, one alternative on that handle with `Assessment` `AssessNone` and `Reason` `ReasonWeak` |
| AC-1.3 | A person Subject with no values | Row `Target` is empty, `Reason` `ReasonEmpty` |
| AC-1.4 | Two `John Smith` Subjects and one handle `John Smith` | One row is placed; the other has `Target` empty, `Reason` `ReasonTaken`, `Flags.PossibleDuplicate`, `DuplicateOf` the placed row (as today, with an empty target instead of Skip) |
| AC-1.5 | A row fixed New, a row fixed Skip | Their rows keep `Target` `TargetNew` / `TargetSkip` and `Reason` `ReasonDecided` (unchanged) |
| AC-1.6 | Swift `PromoteFlow.load` with a proposal row whose `target` is `""` and `assessment` is `"strong"`, `"medium"` or `"none"`, reason not `taken` | The row's target is `.unset`, never `.skip` or `.newKind` |
| AC-1.7 | Swift, a proposal row with `target` `""`, `assessment` `"weak"` and one alternative `e9` | The row drafts `.handle(e9)`, as today |
| AC-1.8 | A real project: a Source with a person who has no record and a person whose only candidate falls short | Both rows open unset ("Select where this files"); choosing New or Skip and pressing Done files a New and leaves the Skip off |

**US-2.** As a researcher, two look-alike rows that would both become new records still warn me once I set both to New.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Two `John Smith` Subjects, no handles | Both rows are unplaced, `Flags.PossibleDuplicate`, each `DuplicateOf` the other (today this is found through `TargetNew`) |
| AC-2.2 | Two `John Smith` Subjects, one handle `Smith` (surname only, a candidate under the weak bar for both) | Both rows are unplaced with the handle as an alternative, and each is `DuplicateOf` the other (new: today these rows are Skip and are not checked) |
| AC-2.3 | Two `John Smith` Subjects, no handles, `s-a` fixed New | `s-a` keeps `TargetNew`; both rows name each other |
| AC-2.4 | Two `John Smith` Subjects, no handles, `s-a` fixed Skip | Neither row is flagged |
| AC-2.5 | A real project: a Source with two persons entered identically, no record | Both rows open unset; after setting both to New each row shows "Also new: …" |

**US-3.** As a researcher, when a suggested row loses its record on a re-proposal, its caption says no record is suggested.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Swift: a row loaded on handle `e1`, then merged with a proposal where that row has `target` `""`, `reason` `""`, `assessment` `"none"` | The row is `.unset`, `updated` is true, `captionReason` is `.unplaced` |
| AC-3.2 | `PromoteModel.reasonText(.unplaced, on:)` | `L10n.string(L10n.Promote.reasonUnplaced)` = "No record suggested", so the caption reads "Updated · No record suggested" |
| AC-3.3 | Wire vocabulary: `PromoteReasonCode(wire: "")` and `PromoteReasonCode(wire: "no_match")` | `.unplaced`; `.unrecognized("no_match")` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/graphalign/rows.go`: a row Align does not place and holds no decision for has an empty `Target`; its row reason is `ReasonUnplaced` (`""`), `ReasonEmpty` or `ReasonTaken`; alternatives, score and band unchanged | STEP-1, AC-1.1 – AC-1.5 |
| `ReasonNoMatch` deleted; `ReasonUnplaced` *(new)* = `""`; `ReasonWeak` documented as an alternative's reason only | STEP-1, D-2, D-3 |
| The possible-duplicate check (`flagAlikeNewRows`) checks unplaced rows and held New rows instead of `TargetNew` rows | STEP-1, AC-2.1 – AC-2.4 |
| Every Go test asserting `TargetNew` / `TargetSkip` / `ReasonNoMatch` / row `ReasonWeak` on an unplaced proposal row asserts an unplaced row | STEP-2, STEP-3 |
| `PromoteGraphAlignmentRow.target` and `.reason` comments in `engine.proto`, regenerated | STEP-3, D-6 |
| Swift: `PromoteReasonCode` and `PromoteFlow.Reason` lose `noMatch` and gain `unplaced`; `swift_vocabulary_test.go` follows | STEP-4, AC-3.3 |
| L10n: `reasonNoMatch` removed, `reasonUnplaced` *(new)* "No record suggested"; `reasonWeak`'s comment names alternatives | STEP-4, AC-3.2 |
| `PromoteFlow.draftTarget` drafts the proposed handle, the weak-band first alternative, or `.unset`; never `.skip` or `.newKind` | STEP-5, AC-1.6, AC-1.7 |
| `FakeStore.buildPromoteProposal`'s unmatched row sends target `""`, reason `.unplaced` | STEP-5, D-8 |
| Docs that say Promote proposes New or Skip (D-9) | STEP-6 |

| Out | Home |
| --- | --- |
| Decided New and Skip as fixed pairs, `TargetNew` / `TargetSkip` in `Fixed`, the batch and Done paths | Not planned: unchanged by design |
| Which record a weak band drafts (the first alternative) | Not planned: the map's scope entry keeps it |
| The `empty` and `taken` row reasons and their copy | Not planned: they still say something true about an unplaced row |
| `promote-find-another-record.md` line 9 ("A row with no candidate opens New (reason `no_match`)") | Not planned: it is that feature's problem statement as of S11-35; the spike folder is distilled at S11-99 |
| A Go constant for an empty row target | Not planned: D-1 |
| Changing the handler code in `api/ffi/handlers/promote_align.go` | Not planned: it copies `Target` and `Reason` as strings and needs no change (§4) |

## 4. Current state

Verified at `a70932e`.

| Area | Where | Today |
| --- | --- | --- |
| Unplaced row | `core/graphalign/rows.go` `proposal`, the final `else` branch (lines 58–84) | `row.Assessment = AssessNone`; with a best candidate: `Target = TargetSkip`, `Score`, `Alternatives = st.alts(sk, nil)`, `Reason = ReasonWeak`, band from the best score when at or above `Bands.Weak`; with none and `subjectHasValues`: `TargetNew`, `ReasonNoMatch`; with no values: `TargetSkip`, `ReasonEmpty`. Then `lostTo` overrides with `ReasonTaken` and the duplicate flags. Comment line 64: "Skip, never force a merge" |
| Held decision | `core/graphalign/rows.go` `proposal`, `st.held` branch | `Target = decided` (New or Skip), `AssessNone`, `ReasonDecided`. Unchanged |
| Alike-New check | `core/graphalign/rows.go` `flagAlikeNewRows` | Visits rows with `Target == TargetNew` and no `PossibleDuplicate`, pairs them with other `TargetNew` rows of the kind through `alikeNew`. Today that is unplaced no-candidate rows and held New rows; Skip-weak rows are never checked |
| Types | `core/graphalign/types.go` | `TargetHandle/New/Skip`; `Reason` consts with `ReasonWeak // the best candidate is below the accept bar`, `ReasonNoMatch // no candidate at all: New`, `ReasonEmpty // nothing to match on: Skip`; `Row.Target` has no doc comment; `Alternative.Reason // via, agrees, or weak`; `Fixed`'s comment: empty target on a `Fixed` means handle |
| Align doc | `core/graphalign/align.go` `Align` | "Align proposes a handle, New, or Skip for every primary Subject on layer" |
| Fixed validation | `core/database/promotealign/promotealign.go` `checkFixed` | Accepts `TargetNew` / `TargetSkip` with no handle. Unchanged |
| Handler | `api/ffi/handlers/promote_align.go` `proposalProto` | Copies `string(r.Target)` and `string(r.Reason)`; reads `TargetHandle` only to attach headers. No code change needed |
| Proto | `api/proto/engine.proto` `PromoteGraphAlignmentRow` | `string target = 3; // handle, new, skip`; field 20's comment lists "via, decided, agrees, weak, taken, no_match, empty". Generated: `api/proto/engine/engine.pb.go` line 13372, `macos/App/Platform/Generated/engine.pb.swift` lines 5569 and 5664. `PromoteGraphAlignmentFixed.target` (field 2, pb.go 13771, pb.swift 5759) keeps "handle, new, skip" |
| Swift wire vocabulary | `macos/App/Platform/ConclusionVocabulary.swift` `PromoteReasonCode` | Cases via, decided, agrees, weak, taken, noMatch, empty, three move codes, `unrecognized`; `""` decodes as `.unrecognized("")` |
| Vocabulary guard | `api/ffi/swift_vocabulary_test.go` "promote reasons" row (lines 323–328) | Requires Swift's `case "…": self =` wire set to equal the Go `graphalign` reason consts, `ReasonNoMatch` included. Regex `case "([a-z_]*)": self = ` matches `case "":` (precedent: `ReconciledState` `case "": self = .empty` mirrors `autoreconcile.StateEmpty = ""`) |
| Swift reasons | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.Reason` | Has `weak`, `noMatch`, `empty`; `init(code:…)` maps `.unrecognized` to `.empty`. So a row reason `""` would read "Nothing to match on" |
| Swift draft | `macos/App/Features/Promote/PromoteFlow.swift` `draftTarget` (line 727) | `.none` → `.unset`; proposed handle → handle; `taken` → `.unset`; `.weak` with a menu → first item; `.strong` and target `new` → `.newKind`; else `.skip` |
| Swift caption | `PromoteFlow.Row.captionReason` (`move ?? reason`); drawn by `PromoteView` and `PromoteEvidenceSheet` through `PromoteModel.reasonText` | `reasonText` words `.weak` → `L10n.Promote.reasonWeak`, `.noMatch` → `reasonNoMatch`, `.empty` → `reasonEmpty` |
| Alternative readout | `PromoteFlow.Row.selectionReadout` | Reads the selected menu item's own `reason` (an alternative's `weak` stays worded by `reasonWeak`) |
| L10n | `macos/App/Platform/L10n.swift` lines 1344–1347; `macos/App/Resources/Localizable.xcstrings` keys `promote.reason.weak`, `promote.reason.noMatch` (line 17676), `promote.reason.empty` | `reasonNoMatch` "No matching record" is read only by `reasonText` |
| FakeStore | `macos/App/Platform/FakeStore.swift` `buildPromoteProposal` (line 1828) | A Subject with no membership and no fixed handle gets `target: "skip"`, `assessment: "none"`, `reason: .empty` |
| Go tests | `core/graphalign/align_test.go`: `TestGracieObituaryGolden` (545), `TestIsolatedSubjectScoresOnTheWalkScale` (630), `TestCandidateLoserDoesNotNominate` (1053), `TestPossibleDuplicates` (1218–1255), `TestRowReasons` (1361, 1367, 1372), `TestNewDuplicateCheckComparesDirectedNeighborsByEnd` (1693), `TestBirthIsNeverProposedOnADeath` (1811, 1835, 1851), `TestLifespanRulesOutAPerson` (1882), `TestOnlyCandidatesOrNominatedHandlesAreCompared` (2088, 2127), `TestContestedRow` (2235); `core/database/promotealign/promotealign_test.go`: `TestSpellingOnlyToponymIsNotShortlisted` (593), `TestProposeRelationshipFromEitherEnd` (710), `TestProposeLifespanFromLifeEvents` (1009), `TestProposeSeesAHandlePromotedAfterTheIndexWasBuilt` (1161), `TestProposeDoesNotCompareAnUnshortlistedHandle` (1270); `api/ffi/handlers/promote_align_test.go`: `TestProposeSendsContested` (101), `TestProposalProtoSendsMove` (185–186) | Assert `TargetNew` / `TargetSkip` / `ReasonNoMatch` / row `ReasonWeak` on unplaced rows. The held-decision tests (`TestHeldNewAndSkipDecisions`, `TestRowReasons` "a held row is decided", `TestContestedRow` row `k`, `TestKinshipShapesCountOnce`, `TestProposeHoldsNewAndSkipDecisions`, `evidence_test.go`, `pins_test.go`, `core/database/promote/batch_test.go`) use fixed New/Skip and stay |
| Swift tests | `macos/ProvenenciaTests/PromoteFlowTests.swift`: `aRowWhoseRecordWasTakenStartsUnset` (173), `aWeakCandidateBelowTheBarStartsSelected` (193), `anUnmatchedRowStartsUnselected` (257), `aContestedRowShowsContestedInItsReadout` (341), `aContestedRowWithoutAHandleShowsNoReadout` (356–361), `reasonCodesParseThroughTheWireVocabulary` (436–462), `anUnresolvedTokenDecidesNothing` (676), `evidenceNeighborsAreTheRowsDoneFiles` (822–845), helper `unmatched(_:)` (946) | Build proposal rows with `target: "new"` / `"skip"`. `evidenceNeighborsAreTheRowsDoneFiles` relies on Swift drafting `.newKind` / `.skip` from a strong New / Skip row |

**Precedent.** `9cc42ba4` ("Leave a taken row open…") changed `draftTarget` so a taken row opens unset whatever its band, with a `PromoteFlowTests` test per band. Copy its shape for the draft change and its loop-over-bands test. `edb419f1` ("promote: typed reasons…") is how a reason code travels: `graphalign.Reason` const → proto comment → `PromoteReasonCode` wire case → `PromoteFlow.Reason` → `PromoteModel.reasonText` → L10n. Copy that path for `unplaced`; do not copy its size.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **An unplaced row's `Target` is the zero value `""`.** No new `Target` const. `Row.Target` gets the doc comment `// Target is TargetHandle when Align places the Subject; a held decision keeps its TargetNew or TargetSkip; otherwise empty: Align suggests nothing.` Tests compare with `""` | The map's wording ("handle or empty"); `Fixed` already reads an empty target as "handle", so a named empty `Target` const would mean two things | `TargetNone Target = ""` — a name whose meaning flips between `Row` and `Fixed` |
| D-2 | **`ReasonNoMatch` is deleted. `ReasonUnplaced Reason = ""` *(new)***, comment `// Align placed nothing and nothing else applies; the page says no record is suggested`, listed after `ReasonTaken`. The candidate branch and the no-candidate-with-values branch both set `row.Reason = ReasonUnplaced` explicitly. `ReasonEmpty`'s comment becomes `// nothing to match on`; `ReasonWeak`'s becomes `// an alternative below the weak bar: too little agrees to name` | The map: "no `weak` / `no_match` reason". The wire carries `""`, as the map says; a named const lets `swift_vocabulary_test.go` mirror it, as `autoreconcile.StateEmpty = ""` does. `ReasonWeak` stays because `alternative` still sets it | Leaving the reason unset with no const — the vocabulary guard cannot list `""`; keeping `ReasonWeak` on rows — the user's call says no |
| D-3 | **The unplaced branch** in `proposal` becomes, in this order: `row.Assessment = AssessNone`; with `best := st.bestNonAssigned(sk, nil)` non-nil: `row.Reason = ReasonUnplaced`, `row.Score = best.score`, `row.Alternatives = st.alts(sk, nil)`, `reported = best`, band from `best.score` when at or above `Bands.Weak` (unchanged); else if `st.subjectHasValues(s)`: `row.Reason = ReasonUnplaced`; else `row.Reason = ReasonEmpty`. The `lostTo` override is unchanged. No branch sets `row.Target`. The comment above the first branch reads `// A candidate existed but was not accepted (below the bar or lost one-to-one): suggest nothing, never force a merge. New and Skip are the researcher's.` | The smallest edit that removes both targets and both reasons and keeps score, band, alternatives and contested as they are | Folding the two `ReasonUnplaced` branches — `subjectHasValues` still picks `ReasonEmpty`, so the shape stays three-way |
| D-4 | **Caption copy** (user, 2026-10-27): an unplaced row's reason is worded "No record suggested". New `L10n.Promote.reasonUnplaced = LocalizedStringResource("promote.reason.unplaced", defaultValue: "No record suggested", comment: "Reason on a Promote row that Promote suggests no record for")`. `reasonNoMatch` and key `promote.reason.noMatch` are deleted. `reasonWeak`'s comment becomes "Reason on a record in a row's menu whose score is below the weak bar" (value unchanged) | The user chose one new line over keeping the two reasons or a bare badge | "Keep the two reasons" and "Badge only" — the user's answer |
| D-5 | **Swift names the case `unplaced`** in both `PromoteReasonCode` (`case "": self = .unplaced`, first in `init(wire:)`) and `PromoteFlow.Reason` (doc `/// Promote suggests no record for this row.`). `Reason.init(code:…)` maps `.unplaced` → `.unplaced`; `.noMatch` is deleted from both enums and from `reasonText` | `.none` would collide with `Optional.none` on `Reason?` (`move`) and `PromoteReasonCode?` (`moveReason`) | `.none` — ambiguous with optionals; `.noMatch` kept for `""` — the old meaning was "New" |
| D-6 | **Proto: comments only.** `PromoteGraphAlignmentRow.target`: `string target = 3; // handle, or empty when Promote suggests no record`. Field 20's comment: `// Why the row reads as it does: via, decided, agrees, taken, empty, or empty string when Promote suggests no record. The page words it. With agrees, the agreeing Property.` Regenerate with `./scripts/generate-proto.sh`. If `protoc` is not installed, edit the same comment lines by hand in `api/proto/engine/engine.pb.go` (13372 and the `Reason` field's comment) and `macos/App/Platform/Generated/engine.pb.swift` (5569, 5664) so they match the proto, and say so in the PR body. `PromoteGraphAlignmentFixed.target` is untouched | The wire shape is unchanged; generators copy field comments verbatim, so a hand edit gives the same bytes | A new field or enum — nothing new travels |
| D-7 | **`draftTarget`** after this PR, in order: `assessment == .none` → `.unset`; proposed handle first in the menu → `.handle`; `Reason(proposal) == .taken` → `.unset`; `assessment == .weak` with a menu item → first item; otherwise `.unset`. The two lines that return `.newKind` and `.skip` go. The first comment becomes `// No suggestion: leave the choice open so Skip and New are the researcher's.` | The map's scope entry; a medium or strong row that was not placed and not taken (a settle or a later registry) now opens unset instead of Skip | Drafting the first alternative for medium and strong — the map keeps "the weak-band first alternative as today" only |
| D-8 | **FakeStore**'s unmatched row: `target: ""`, `reason: .unplaced`, everything else as today | The fake mirrors Go; its rows have no values to say `empty` about | Keeping `.empty` — it would word "Nothing to match on" for a fake person with a name |
| D-9 | **Docs sweep**, word for word. `docs/promote-graph-alignment.md`: line 33 "such a row comes back **Skip** (New only when it has no candidate at all). The page starts a row with no match **unset**, so Done waits for the researcher's Skip or New, and drafts a weak one on its best match for review." → "such a row comes back with no target, its candidates as alternatives. The page starts it **unset**, so Done waits for the researcher's Skip or New, and drafts a weak one on its best match for review."; line 52 "or two New rows whose properties clear the accept bar" → "or two rows that would both be new records (unplaced or decided New) whose properties clear the accept bar"; line 73 "Align proposes a handle, New or Skip for every" → "Align proposes a handle or nothing for every"; line 108 "A candidate that cleared no bar comes back Skip with the candidate as an alternative (New only when there is no candidate at all);" → "A row with no accepted candidate comes back with no target and its candidates as alternatives;"; line 125 "a match below the bar comes back Skip, never New, and a row with no match starts unset on the page." → "a match below the bar comes back as an alternative with no target, and the row starts unset on the page.". `docs/ideas/coverage-aware-similarity.md` line 7 "so the row is Skip `weak` with the handle as an unbanded alternative" → "so the row has no target, with the handle as an unbanded alternative". `docs/deployment-plan/spike-11/name-candidates.md` line 11 "a Subject with values opens New (`no_match`) and an empty menu" → "a row opens unset with an empty menu"; line 69 first sentence "Go today gives every row it does not place a target:" → "Until S11-47, Go gave every row it did not place a target:" and "Swift drafts neither." → "Swift drafted neither."; line 73 "finds its rows by `TargetNew` today; S11-47 finds them as unplaced rows instead" → "found its rows by `TargetNew` before S11-47; it now checks unplaced rows and decided New rows". `core/graphalign/align.go` `Align`'s comment: "proposes a handle, New, or Skip for every primary Subject" → "proposes a handle or nothing for every primary Subject" | Each PR updates the docs it makes false | Rewording `promote-find-another-record.md` — Out |
| D-10 | **`flagAlikeNewRows`** visits a row when `rows[i].Target != TargetHandle && rows[i].Target != TargetSkip` (unplaced or held New) and it has no `PossibleDuplicate`, and pairs it with rows of the same kind that pass the same test. Its doc comment's first sentence becomes `// flagAlikeNewRows marks two rows of one kind that would both be new records (unplaced, or decided New) and would clear the accept bar against each other: filing both New mints two handles for what may be one entity.` The rest of the comment and `alikeNew` stay | The map: "found among unplaced rows". Held New rows must stay in, because the Swift "also new" note needs both rows `.newKind`, which Swift sends as fixed New. Held Skip rows file nothing. A no-values row scores 0 against anything, so it is never flagged and needs no special case | Only `Target == ""` — loses the note once the researcher sets both to New (AC-2.5) |
| D-11 | **Test edits are mechanical**: each assertion of `TargetNew` + `ReasonNoMatch` or `TargetSkip` + `ReasonWeak` on an unplaced row becomes `Target == ""` + `Reason == graphalign.ReasonUnplaced`; every other part of the assertion and its message keeps its meaning (messages say "unplaced" for "New"/"Skip"). The exact list is in STEP-2 and STEP-3. `TestPossibleDuplicates` gains a `fixed` field and a `taken` field (D-12) | AC-1.x are exactly the old assertions with the new target and reason | Deleting tests whose subject was the target — they also pin bands and alternatives |
| D-12 | **`TestPossibleDuplicates`** table gains `fixed []graphalign.Fixed` and `taken string` (the subject that must read `ReasonTaken`, empty for none). The loop's `r.Target == graphalign.TargetSkip && flagged && r.Reason != graphalign.ReasonTaken` check becomes `string(r.SubjectID) == tt.taken && r.Reason != graphalign.ReasonTaken`. Case 1 sets `taken: "s-b"` | The old check keyed "taken" on Skip, which no longer exists | Keying on `Target == ""` — the new weak-candidate case is unplaced and flagged but not taken |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

Searched: `rg -n "ReasonNoMatch|ReasonWeak|TargetNew|TargetSkip" core api`, `rg -n "noMatch|reasonNoMatch|case \"\": self" macos/App`, `rg -n "promote.reason" macos/App/Resources/Localizable.xcstrings`, `rg -n "draftTarget|captionReason" macos`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphalign.ReasonUnplaced` *(new)* | const, `core/graphalign/types.go` | `ReasonNoMatch` (deleted), `ReasonEmpty` | `no_match` meant "New"; `empty` means no values. The const names the empty wire code so the vocabulary guard can mirror it (D-2) |
| `PromoteReasonCode.unplaced` *(new)*, `PromoteFlow.Reason.unplaced` *(new)* | enum cases, `ConclusionVocabulary.swift`, `PromoteFlow.swift` | `.noMatch` (deleted), `.empty` | Replaces `.noMatch` with the new meaning; `.empty` is worded "Nothing to match on" (D-5) |
| `L10n.Promote.reasonUnplaced` *(new)*, key `promote.reason.unplaced` *(new)* | L10n key | `reasonNoMatch` (deleted) | New wording chosen by the user (D-4); the key is renamed with the meaning |
| `TestUnplacedRowsCarryNoTarget` *(new)* | Go test, `core/graphalign/align_test.go` | `TestRowReasons` | `TestRowReasons` asserts reasons only; this one pins target, reason, band and alternatives together for AC-1.1 – AC-1.4 |
| `anUnplacedRowNeverDraftsNewOrSkip` *(new)*, `aRowThatLosesItsRecordSaysNoRecordIsSuggested` *(new)* | Swift tests, `PromoteFlowTests.swift` | `aRowWhoseRecordWasTakenStartsUnset`, `decidedRowsSurviveAReproposal` | Same shapes (loop over bands; merge a re-proposal), new assertions |

**Builds on.** `graphalign`'s row building in `rows.go` (S11-16), the typed reason path of `edb419f1`, the taken-row draft of `9cc42ba4`, the wire-vocabulary guard `swift_vocabulary_test.go`.
**Deletes.** `graphalign.ReasonNoMatch`; the two `TargetNew` / `TargetSkip` assignments and the `ReasonWeak` / `ReasonNoMatch` assignments in `proposal`; `PromoteReasonCode.noMatch`; `PromoteFlow.Reason.noMatch`; `reasonText`'s `.noMatch` arm; `L10n.Promote.reasonNoMatch` and `promote.reason.noMatch` in `Localizable.xcstrings`; `draftTarget`'s `.newKind` and `.skip` returns. `ReasonWeak`, `.weak` and `reasonWeak` stay: alternatives read them.
**Next adds.** S11-45 (prefix names) adds candidates that open unset with no target. Any later reason for an unplaced row is a new `Reason` const plus its wire case and L10n line, through the same path.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A proposal row with `target` `""` and a band of medium or strong (not taken) | `PromoteFlow.draftTarget` → `.unset`; `PromoteView` target menu | `.unset` is already a row state (no band, placeholder "Select where this files", `targetsChosen` false); `selectionReadout` is nil for it, so no badge |
| A row reason `""` | `PromoteReasonCode(wire:)` → `.unplaced` → `PromoteFlow.Reason.unplaced` → `reasonText` | New case handled in every `switch` (the compiler enforces it); shown only in the Updated caption |
| An unplaced row with a weak candidate is now checked by `flagAlikeNewRows` | `PromoteFlow` duplicate notes (`proposedDuplicateOf`) | The note shows only when the other row has a handle (`sharesHandle`) or both rows are `.newKind` (`alsoNew`); an unset pair shows nothing until the researcher chooses |

### Scale

At the map's Scale (a 12-generation tree, the large-source budgets): `flagAlikeNewRows` now also visits rows with a below-bar candidate. It is pairwise over unplaced rows of one kind, as it was over New rows; on a large Source most unplaced rows had no candidate and were already visited. `TestLargeSourceBudgets` (check-pr) bounds it; going over budget is a stop condition.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/graphalign` | `types.go`: `ReasonUnplaced` *(new)*, `ReasonNoMatch` deleted, comments per D-1, D-2. `rows.go`: unplaced branch per D-3, `flagAlikeNewRows` per D-10. `align.go`: `Align` comment per D-9 |

### FFI

`api/proto/engine.proto`: comments on `PromoteGraphAlignmentRow` fields 3 and 20 (D-6). Regenerate with `./scripts/generate-proto.sh`. Handler `api/ffi/handlers/promote_align.go` unchanged.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Row draft | `PromoteFlow.draftTarget` | `macos/App/Features/Promote/PromoteFlow.swift` | D-7 |
| Row reason | `PromoteFlow.Reason.unplaced` | `macos/App/Features/Promote/PromoteFlow.swift` | D-5 |
| Wire code | `PromoteReasonCode.unplaced` | `macos/App/Platform/ConclusionVocabulary.swift` | D-5 |
| Caption wording | `PromoteModel.reasonText` | `macos/App/Features/Promote/PromoteModel.swift` | `.unplaced` → `L10n.Promote.reasonUnplaced` |
| Fake proposal | `FakeStore.buildPromoteProposal` | `macos/App/Platform/FakeStore.swift` | D-8 |

States: an unplaced row is the existing `.unset` state. Accessibility: the caption is a `Text`; no new control.

### Strings

L10n keys *(new)*: `L10n.Promote.reasonUnplaced` = "No record suggested" (`promote.reason.unplaced`). Removed: `L10n.Promote.reasonNoMatch` "No matching record" (`promote.reason.noMatch`). Comment changed: `promote.reason.weak` (D-4). Old-scope copy: `promote.reason.weak` "Too little agrees to suggest a match" now speaks only of a menu record, which its wording already fits; no other string names a proposed New or Skip.

## 7. Steps

Every Go test command is run with `CGO_ENABLED=1` and `-tags fts5`, from the repo root.

**STEP-1 — Go: unplaced rows carry no target**
Files: `core/graphalign/types.go`, `core/graphalign/rows.go`, `core/graphalign/align.go`. Skill: `add-match-rule`.
Do: D-1, D-2, D-3, D-10, and D-9's `Align` comment.
Check: `CGO_ENABLED=1 go build -tags fts5 ./core/... ./api/...` passes (no non-test file used `ReasonNoMatch`). `CGO_ENABLED=1 go vet -tags fts5 ./core/graphalign/ ./api/ffi/` reports only the test files that still name `graphalign.ReasonNoMatch`, which STEP-2 and STEP-3 fix.

**STEP-2 — Go: graphalign tests**
Files: `core/graphalign/align_test.go`. Rule: `go-tests.mdc`.
Do: per D-11 and D-12. `TestGracieObituaryGolden`: the second check becomes `r.Target != ""` with message "place %s: %+v, want no target". `TestIsolatedSubjectScoresOnTheWalkScale` "a different first name is no candidate": `wantTarget: ""`, `wantReason: graphalign.ReasonUnplaced`. `TestCandidateLoserDoesNotNominate`: `row.Target != "" || row.Reason != graphalign.ReasonUnplaced || row.Via != nil`. `TestRowReasons`: the two `ReasonWeak` cases and "no candidate is no match" want `graphalign.ReasonUnplaced`; rename "no candidate is no match" → "no candidate suggests nothing" "a neighbor-reached candidate below the bar is weak" → "a neighbor-reached candidate below the bar suggests nothing", and "a surname-only look-alike is offered as weak" → "a surname-only look-alike is only an alternative". `TestNewDuplicateCheckComparesDirectedNeighborsByEnd`: `r.Target != ""`, message "want unplaced and a possible duplicate". `TestBirthIsNeverProposedOnADeath` (three sites), `TestLifespanRulesOutAPerson` (AC-1.4), `TestOnlyCandidatesOrNominatedHandlesAreCompared` (two sites): `Target != ""` and `Reason != graphalign.ReasonUnplaced`, messages "unplaced". `TestContestedRow` row `n`: `r.Target != ""`, message "an unplaced row is not contested". `TestPossibleDuplicates`: D-12 plus T-3's cases. Add T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/...` passes.

**STEP-3 — Go: catalog and FFI tests, proto comments**
Files: `core/database/promotealign/promotealign_test.go`, `api/ffi/handlers/promote_align_test.go`, `api/ffi/swift_vocabulary_test.go`, `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: `promotealign_test.go`: lines 593, 1009, 1270 → `Target != "" || Reason != graphalign.ReasonUnplaced` (1270 keeps `len(row.Alternatives) != 0`); 710 → same, message "want no target (under the accept bar)"; 1161 → `first.Target != ""`, message "want no target". `promote_align_test.go`: `TestProposeSendsContested` case 1 wants target `""`, reason `""`; `TestProposalProtoSendsMove` rows drop `Target: graphalign.TargetSkip` (zero value). `swift_vocabulary_test.go` "promote reasons": replace `graphalign.ReasonNoMatch` with `graphalign.ReasonUnplaced`. D-6 for the proto.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./api/ffi/...` passes except `TestSwiftVocabularyMatchesGo` "promote reasons" (Swift not yet changed: it reports `""` missing and `no_match` extra). `git diff --stat api/proto/engine/engine.pb.go macos/App/Platform/Generated/engine.pb.swift` shows comment lines only.

**STEP-4 — Swift: the wire code, the reason and its words**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Features/Promote/PromoteFlow.swift` (`Reason`), `macos/App/Features/Promote/PromoteModel.swift` (`reasonText`), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-4, D-5. In `Localizable.xcstrings`, delete the `promote.reason.noMatch` entry and add `promote.reason.unplaced` beside `promote.reason.taken` in the same shape (`extractionState` `manual`, `en` `translated`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `python3 scripts/check-localizable-xcstrings.py` passes; `rg -n "noMatch|reasonNoMatch|promote\.reason\.noMatch" macos/App` prints only the unrelated `noMatches` omnibar/search and `EvidencePropertyCause.noMatch` lines.

**STEP-5 — Swift: draft and fake**
Files: `macos/App/Features/Promote/PromoteFlow.swift` (`draftTarget`), `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: D-7, D-8. Test fixtures: every `row(id:, target: "new", …)` and `row(id:, target: "skip", …)` in `PromoteFlowTests.swift` that stands for a proposal row (lines 173, 193, 257, 341, 356, 358, 361, 676, 946) uses `target: ""`; `unmatched(_:)` sets `reason = .unplaced` and its doc reads "A row Promote found no record for: no target, no band, no alternatives, so it opens unset."; `reasonCodesParseThroughTheWireVocabulary` replaces `("no_match", .noMatch)` with `("", .unplaced)`, adds `#expect(PromoteReasonCode(wire: "no_match") == .unrecognized("no_match"))`, and replaces `reason(.noMatch) == .noMatch` with `reason(.unplaced) == .unplaced`. `evidenceNeighborsAreTheRowsDoneFiles`: rows `n` and `s` become `row(id: "n", target: "", handleID: "", assessment: "strong")` and the same for `s`; after `load`, call `flow.setTarget(subjectID: "n", token: "new")` and `flow.setTarget(subjectID: "s", token: "skip")` before the assertions, which stay. Add T-4, T-5, T-6.
Check: on macOS, `xcodebuild test` through `scripts/check-pr.sh` (STEP-7) passes. Off macOS, `rg -n 'target: "new"|target: "skip"' macos/ProvenenciaTests/PromoteFlowTests.swift` prints only `CatalogPromoteGraphAlignmentFixed` / batch lines.

**STEP-6 — Docs that say Promote proposes New or Skip**
Files: `docs/promote-graph-alignment.md`, `docs/ideas/coverage-aware-similarity.md`, `docs/deployment-plan/spike-11/name-candidates.md`.
Do: D-9, word for word.
Check: `rg -n "comes back Skip|comes back \*\*Skip|New or Skip for every|Skip \`weak\`|opens New \(\`no_match\`\)" docs/promote-graph-alignment.md docs/ideas docs/deployment-plan/spike-11/name-candidates.md` prints nothing.

**STEP-7 — Whole suite**
Files: none.
Do: run `scripts/check-pr.sh`.
Check: it prints `check-pr: PASS`. `rg -n "ReasonNoMatch|reasonNoMatch" core api macos docs/*.md .agents .cursor` prints nothing.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done, once the PR number is known.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-47-suggest-a-record-or-nothing.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. STEP-1 to STEP-3 are one commit ("Leave a row Promote cannot place with no target instead of a suggested New or Skip."); STEP-4 and STEP-5 one commit ("Open every unplaced Promote row unset, and say no record is suggested when a row loses its record."); STEP-6 and STEP-8 docs commits. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/graphalign/align_test.go` `TestUnplacedRowsCarryNoTarget` *(new)*, table-driven | Metas `nameMetas()`, `alignWith(…, matchregistry.Default().AlignConfig())`. (a) "no candidate": Subject `s-a` `nameVals("Ann", "Ames")`, no handles → `Target ""`, `Reason ReasonUnplaced`, `AssessNone`, 0 alternatives. (b) "a candidate under the weak bar": same Subject, handle `h-a` `surnameVals("Ames")` → `Target ""`, `Reason ReasonUnplaced`, `AssessNone`, 1 alternative `h-a` with `AssessNone`, `ReasonWeak`. (c) "no values": Subject `{ID: id("s-a"), Ref: "A", Kind: "person"}` → `Target ""`, `ReasonEmpty`. (d) "taken": Subjects `s-a`, `s-b` both `nameVals("John", "Smith")`, handle `h-1` `nameVals("John", "Smith")` → the row of `s-b` has `Target ""`, `ReasonTaken`, `PossibleDuplicate`, `DuplicateOf` `s-a` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-3 |
| T-2 | Go | `core/graphalign/align_test.go` `TestHeldNewAndSkipDecisions` (existing) and `TestRowReasons` "a held row is decided" (existing) | Unchanged; they keep `TargetNew` / `TargetSkip` / `ReasonDecided` on fixed rows | AC-1.5 |
| T-3 | Go | `core/graphalign/align_test.go` `TestPossibleDuplicates` (extended) | New cases: "two unplaced rows that share a below-bar candidate name each other": handles `[{h-1, PER-1, surnameVals("Smith")}]` → `s-a`↔`s-b`. "a row decided New and its look-alike name each other": `fixed: [{SubjectID: id("s-a"), Target: graphalign.TargetNew}]` → `s-a`↔`s-b`. "a row decided Skip is not checked": `fixed: [{SubjectID: id("s-a"), Target: graphalign.TargetSkip}]` → no flags. The existing "two New rows…" case is renamed "two unplaced rows that match each other name each other" | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-10, D-12 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `anUnplacedRowNeverDraftsNewOrSkip` *(new)* | For each band `"none"`, `"medium"`, `"strong"`: `row(id: "a", target: "", handleID: "", assessment: band)` with `reason = .unplaced` and the default alternative `e2` at that band, loaded with `fact("a")` → target `.unset`, `targetsChosen == false`. Then `"weak"` → `.handle(id: "e2", ref: "PER-2", title: "PER-2")` | AC-1.6, AC-1.7, D-7 |
| T-5 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aRowThatLosesItsRecordSaysNoRecordIsSuggested` *(new)* | Load `[row(id: "a"), row(id: "b", handleID: "e1")]`, then `merge` a proposal (revision 4) where `b` is `row(id: "b", target: "", handleID: "", assessment: "none")` with `reason = .unplaced`, `alternatives = []` → `b.target == .unset`, `b.updated == true`, `b.captionReason == .unplaced` | AC-3.1 |
| T-6 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `anUnplacedRowIsWordedNoRecordSuggested` *(new)* | `loadedAdaAndBea(row("a", handleID: "e1"), row("b", handleID: "e2"))`; `model.reasonText(.unplaced, on: bea) == L10n.string(L10n.Promote.reasonUnplaced)` and that string is "No record suggested" | AC-3.2, D-4 |
| T-7 | Swift | `PromoteFlowTests.swift` `reasonCodesParseThroughTheWireVocabulary` (edited, STEP-5) | `""` → `.unplaced`; `"no_match"` → `.unrecognized("no_match")`; `reason(.unplaced) == .unplaced` | AC-3.3, D-5 |
| T-8 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "promote reasons" (edited, STEP-3) | Go set with `ReasonUnplaced` equals Swift `PromoteReasonCode` wire set | AC-3.3, D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/graphalign` (`TestGracieObituaryGolden`), `core/database/promotealign` (goldens, `TestLargeSourceBudgets`), `api/ffi` (`TestSwiftVocabularyMatchesGo`), `PromoteModelTests` that load a FakeStore proposal and read an unmatched row's reason.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: AC-1.1–1.4 → T-1; AC-1.5 → T-2; AC-1.6, 1.7 → T-4; AC-2.1–2.4 → T-3; AC-3.1 → T-5; AC-3.2 → T-6; AC-3.3 → T-7, T-8; AC-1.8, AC-2.5 → manual check
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
- [ ] Domain docs updated: docs/promote-graph-alignment.md, docs/ideas/coverage-aware-similarity.md, name-candidates.md (D-9)
- [ ] PR description: what shipped, AC evidence, deviations (including a hand-edited generated file, D-6), manual check
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
| More rows enter the alike-New check on a large Source | `TestLargeSourceBudgets` runs in check-pr |
| `protoc` missing in the implementer's environment | D-6's hand edit: comment lines only |

**Stop and report if:**

- `core/graphalign/rows.go`'s unplaced branch no longer assigns `TargetSkip` / `TargetNew` / `ReasonWeak` / `ReasonNoMatch` as §4 says (S11-45 or another PR moved it).
- A non-test Go file other than those in STEP-1 fails to build after `ReasonNoMatch` is deleted — name the file.
- Any Go or Swift test not named in STEP-2, STEP-3 or STEP-5 fails — name it and its output; do not edit it.
- `TestLargeSourceBudgets` goes over budget.
- `TestPossibleDuplicates`' new case "two unplaced rows that share a below-bar candidate" is not flagged: `alikeNew` does not pair them, so D-10's claim is wrong.
- The generated-file diff after D-6 shows anything other than comment lines.
- In the manual check, a row opens on Skip or New without the researcher choosing it.

## 12. Open questions

None.

## 13. PR

**Title:** S11-47 — Promote suggests a record or nothing
**Body:** objective in one sentence; what shipped against US-1, US-2, US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-47-suggest-a-record-or-nothing.md`.
