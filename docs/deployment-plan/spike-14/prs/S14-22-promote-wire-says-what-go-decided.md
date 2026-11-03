# S14-22 — Promote's wire says what Go decided

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes (batch 5; was Spike 12's S12-48)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a connection that would close a loop is drawn with the self-link subrow's existing warning callout and wording pattern; no view gains or loses a control, and every other subrow looks as it does today)
**Board:** —
**Depends on:** S13-23 (landed [#633](https://github.com/mendahu/provenencia/pull/633))
**Blocks:** —
**Branch:** `s14-22-promote-wire-says-what-go-decided`
**Version bump:** patch — a fix (the footer counted a connection Done refuses) with an additive wire field, no feature ([`docs/versioning.md`](../../../versioning.md)); the next patch after `VERSION` at merge (0.2.11 → 0.2.12 at planning)
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (for `core/graphalign` only), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md) (reading the term seed, no new seed), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc` (every wire enum is a typed Swift enum in `Platform/` with a guard row), `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [audit 2](../app-health-audit-2.md) B-23 (map review K-38, K-40); [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) §9.1 Bridge filing; working rules "(Batch 5) Audit items keep their IDs", "(Batch 5) FakeStore is a seeded fake, not a second engine", "(Batch 5) Proposals are liberal; the researcher decides"; Spike 13 working rules "The score ranks, unseen" and "Spike 14's Promote PRs build on the removal"; S13-23 plan **Next adds** ([`S13-23-promote-stops-asking-for-pins.md`](../../spike-13/prs/S13-23-promote-stops-asking-for-pins.md)); the map scope entry's Spike 11 leftover (guard rows read their list from Go)
**Planned:** 2026-11-03 against `main` @ `0e205443`
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

On the Promote page, every connection subrow and the footer's connection count come from Go, which decides each connection by the same rule Done files it with. Today Swift decides which connections file (both ends on records, not the same record), so a place `part_of` or `succeeded_by` link that would close a loop is counted and shown as "Promotes on Done", Done leaves it unfiled, and nothing says so (audit 2 B-23). After this PR the proposal reports, per connection, whether it files and why not; such a connection reads "Not promoted · these places would form a loop", and the footer counts only what Done files.

**Why now.** Bar outcome 19 ("the footer's connection count is what Done files"). S13-23 landed, so the Promote rows are one grid and Swift's filing rule is the only copy left to remove. Every later refusal the filer adds would otherwise be copied into Swift.

**What stays.** Which rows Promote proposes and how it ranks them (Spike 13 "The score ranks, unseen"). What Done writes: the filer's behavior is unchanged, it calls the shared rule instead of its own copy. The subrows' layout, ordering and the self-link warning. A connection's sentence, mark and ends come from the graph snapshot as today.

## 2. User stories

**US-1.** As a researcher, when a place link on the page would close a loop with places already in my tree, I see that it will not be promoted and why, and the footer does not count it, so that what Done promotes is what the page said.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Tree: York `part_of` Ontario (filed). Source B records Ontario `part_of` York; both rows are on their records (Ontario on Ontario's, York on York's) | The subrow under each row shows the sentence muted and the warning "Not promoted · these places would form a loop"; the footer reads "… · 0 connections" |
| AC-1.2 | AC-1.1, then Done | No `place_relationship` association is minted for the Source B link; the toast's connection count is 0 |
| AC-1.3 | Tree: York `part_of` Ontario (filed). Source B records Ontario `succeeded_by` York | The connection files ("Promotes on Done", counted 1): `part_of` and `succeeded_by` loops are checked apart |
| AC-1.4 | Source B records A `part_of` B and B `part_of` A, both rows New | The connection whose bridge Subject id is lower promotes; the other reads the loop warning; footer "1 connections"; Done files one association |
| AC-1.5 | AC-1.4, then the researcher sets A to Skip | Both connections wait for A ("Waits for A"); footer 0 |
| AC-1.6 | VoiceOver on AC-1.1's subrow | Reads "<sentence>. Not promoted, these places would form a loop." |

**US-2.** As a researcher, the page's other connection states read as they do today, now decided by Go.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Ann (row on a record) and her birth event (row set to Skip) are joined by a participation | "Waits for <birth name>"; VoiceOver "…, who is skipped."; not counted |
| AC-2.2 | A row with no target (Unset) at one end | "Waits for <name>"; VoiceOver "…, who has no target yet." |
| AC-2.3 | Page opened on one Subject (not "Promote all"); the other end is an unpromoted Subject | "Waits for <name>"; VoiceOver "…, who is not on this page."; after "Map the rest" the page proposes again and the connection reads by that end's target |
| AC-2.4 | Two rows set to the same record, joined by a relationship | "Not promoted · both ends are PER-…" as today; not counted |
| AC-2.5 | A bridge whose end names no Subject, names Subjects that land on two records, or a typed link with no type term | "Waits for its other end"; not counted; Done does not file it |
| AC-2.6 | Both ends already promoted, link not filed | Counted, shown under no row (as today) |
| AC-2.7 | Any page | The footer's connection count equals the number of `place_relationship`, `relationship`, `participation` and `location` associations Done files, as long as no proposal failed after the last decision (see Out) |

**US-3.** As a contributor, the filing rule lives in one Go function that the filer and the proposal both call, and a new refusal reaches the page without a Swift rule.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `rg "placeRelationshipWouldCycle\|EndResolution\|func resolve" core macos/App` | No hit |
| AC-3.2 | A Go connection outcome Swift's `PromoteConnectionOutcome` lacks, or the reverse | `TestSwiftVocabularyMatchesGo` fails; its expected list is `graphalign.ConnectionOutcomes()`, not a hand-written list |
| AC-3.3 | The filer refuses York→Upper Canada→Canada→York (the existing `bridges_test.go` case) | It still files two associations, through `connectrules.Refuse` |

## 3. Scope

| In | Covered by |
| --- | --- |
| One filing rule in Go: `connectrules.Refuse` (self-link, loop) and `connectrules.ClosesCycle`, the loop-checked terms on the bridge registry; the filer calls it (B-23) | STEP-1, STEP-2, AC-3.1, AC-3.3 |
| The proposal decides each unfiled connection on the graph: files, or why not (skipped, unset, not on the page, self-link, loop), in the filer's order, from the page's targets, the canon edges and the page's own connections (B-23) | STEP-3, STEP-4, AC-1.1 – AC-1.5, AC-2.1 – AC-2.6 |
| The wire: `ProposePromoteGraphAlignmentRequest.only_subject_id`, `ProposePromoteGraphAlignmentResponse.connections` | STEP-5, AC-2.3 |
| A typed Swift enum `PromoteConnectionOutcome` in `Platform/ConclusionVocabulary.swift` with a guard row that reads Go's list | STEP-5, STEP-6, AC-3.2 |
| Swift's subrows and footer read Go's outcomes; Swift's own rule (`resolve`) goes; "Map the rest" proposes again | STEP-7, AC-1.1, AC-2.3, AC-2.7 |
| Loop warning copy and VoiceOver | STEP-8, AC-1.1, AC-1.6 |
| `promote-graph-alignment.md` §9.1 says where a refusal is reported | STEP-9 |

| Out | Home |
| --- | --- |
| A proposal that fails after a decision leaves the subrows and the footer on the last answer, and Done stays enabled (today's behavior for the rows) | S14-20 (the serial latest-wins proposal runner and Promote error recovery) |
| Swift's weak-row draft (B-26) | Spike 13's S13-24 (moved 2026-10-29) |
| Unread wire fields (B-37) and `graphalign.Subject.Provenance` | S13-24 (wire half, landed); S14-20 (`Provenance`) |
| The typed band (B-24's other half) | Not planned: the band is deprecated (Spike 13 "The score ranks, unseen") |
| Converting the other hand-written guard rows (`promote reasons`, `promote row targets` and the rest) to read their lists from Go | Not planned |
| Wording a loop by its term or naming the places | Not planned: the user chose one string (2026-11-03) |
| A separate "incomplete" state for a bridge with a missing end, two records on one end, or no type | Not planned: the user chose "Waits for its other end" (2026-11-03) |
| Reclaiming the empty freed slot on a working row (`provisional-identity-claims.md` names S14-22 as a candidate) | Not planned here: no chrome change; the idea keeps it |
| A filing rule for a non-place bridge (person relationship loops) | Not planned |

## 4. Current state

Verified at `0e205443`.

| Area | Where | Today |
| --- | --- | --- |
| Filer | `core/database/identityclaims/bridgefile.go` `fileBridge` | Per bridge: skips a filed one; resolves each endpoint to one accepted handle (`resolveEndpoint`, else waits); refuses a self-link (`len(seenEntity) < len(ends)`); reads the term for a type-keyed bridge (`keysOnType`; no term → not filed); for `place_relationship` with `propertyterms.KeyPartOf` or `KeySucceededBy`, reorders ends to from/to by property key and calls `placeRelationshipWouldCycle`; then joins or mints the association |
| Filer loop check | same file `placeRelationshipWouldCycle(tx, termKey, fromID, toID)` | One SQL read of every filed `place_relationship` of that term (entity ids of both ends, hard-coded `'place_relationship'`, `'place_relationship_type'`, `'from'`, `'to'`), then a BFS from `toID` looking for `fromID`. The only loop rule in Go |
| Filer order | `FileEvidenceGraphBridgesTx` → `bridgeHeads` (`ORDER BY s.id`) → `fileHeads` | Done (`core/database/promote/batch.go` `SaveBatch`) files claims first, then every unfiled bridge on the graph in bridge Subject id order; a link filed earlier in the loop is visible to a later loop check |
| Acyclic terms | `core/database/propertyterms/propertyterms.go` `KeyPartOf`, `KeySucceededBy`; seeds in `core/database/subjectvocab/registry.go` `seedTerms` (`part_of`, `succeeded_by`: `Directed: true`, no `Inverse`) | The filer names the two terms inline |
| Connect matrix | `core/connectrules/connectrules.go` `Bridge`, `productBridges`, `LookupBridge` | `place_relationship` endpoints `[from, to]`, disambiguation `place_relationship_type`. Pure, imports only `vocabkey` |
| Proposal | `core/database/promotealign/promotealign.go` `Propose(c, graphID, fixed)` → `propose` → `loadInputs`, `mergeFixed` (caller fixed + already-promoted primaries as anchors), `loadCanon`, `graphalign.Align` | Returns rows only. No connection is decided in Go |
| Canon | `core/database/promotealign/canon.go` `loadCanon` | Walks every link of every reached handle (the whole connected component of each fixed, anchor and candidate handle), primary kinds only; each link is a `graphalign.CanonEdge` keyed by `canonlinks.EdgeOf` / `segmentcompare.Keys` (`KeyFrom.End` is the property of `From`'s end, `Term` is the recorded term; `part_of` and `succeeded_by` have no inverse, so they are not re-keyed) |
| Graph read | `graphcache.EvidenceGraph` `Subjects` (`TypeKey`, `TypeOrigin`), `Observations` (`SubjectID`, `PropertyID`, `ValueSubjectID`, `TermID`), `Members` | All in memory once the graph is loaded; `Vocabulary().PropertiesByID()`, `TermsByID()` name properties and terms |
| Proposal types | `core/graphalign/types.go` `Proposal{Rows}`, `Row.Start()`, `Fixed`, `Target` | `Start` is where the page starts an undecided row; a held handle row with band none starts empty, so the page's target for a decided row is its `Fixed`, not `Start` |
| Handler | `api/ffi/handlers/promote_align.go` `ProposePromoteGraphAlignment`, `proposalProto` | Maps rows; no connections |
| Proto | `api/proto/engine.proto` `ProposePromoteGraphAlignmentRequest` (fields 1, 3, 4; 2 reserved), `ProposePromoteGraphAlignmentResponse` (1 rows, 2 revision) | — |
| Swift rule | `macos/App/Features/Promote/PromoteFlow.swift` `connections`, `resolve(_:in:)` ("The one place that decides"), `EndResolution`, `lined`, `connectionCount`, `ConnectionState`, `WaitReason`, `stateRank`, `connectionsByRow` | Swift decides each state from the rows' targets and `mappedRest`; misses the loop refusal and the missing-term refusal (B-23) |
| Swift bridges | `macos/App/Features/Promote/PromoteModel.swift` `bridgeFacts`, `load`, `repropose`, `mapRest`, `promote` | Bridges are built once at load from the graph snapshot. `mapRest()` only flips `flow.mappedRest`, no proposal |
| Swift wire | `macos/App/Platform/GenealogyStore.swift` `CatalogPromoteGraphAlignmentProposal`, protocol `proposePromoteGraphAlignment(projectDir:graphID:fixed:)`; `GoStore.swift` `proposePromoteGraphAlignment`; `FakeStore.swift` `proposePromoteGraphAlignment`, `buildPromoteProposal`, `promoteProposals`, `lastPromoteFixed`; `macos/ProvenenciaTests/OnboardingModelTests.swift` stub | No connections on the wire |
| Swift vocabulary | `macos/App/Platform/ConclusionVocabulary.swift` `PromoteMoveReason`, `PromoteTargetCode` | The wire-enum pattern this PR copies |
| View | `macos/App/Features/Promote/PromoteView.swift` `PromoteConnectionSubrows.subrow` | Mutes a non-promoting sentence; shows `connectionState` text; a self-link gets `PVCallout(tone: .warning, …, compact: true)` |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote.connectionState`, `connectionSpoken`, `connectionSelfLink` | Self-link: "Not promoted · both ends are %@" / "Not promoted, both ends are %@." |
| Guard rows | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`, helpers `listStrings`, `partTypeKeys` | The newer rows read their list from Go (`effects.Lists`, `namevalue.PartTypes()`) |
| Tests | `core/database/promote/bridges_test.go` "place relationship files; cycles refuse; split and amalgamation ok"; `core/database/promotealign/promotealign_test.go` `TestProposePlacePartOfChain`, `TestProposeQueryCountDoesNotGrowWithTheLayer`, `TestProposeQueryCountStaysFlatWhenWarm`, fixture `f.placeRel`, `f.promote`, `f.propose`; `api/ffi/handlers/promote_align_test.go` `TestProposePromoteGraphAlignment`; `macos/ProvenenciaTests/PromoteFlowTests.swift` `connectionsFollowTheirEnds`, `aConnectionWithOneRowEndShowsOnlyThere`, `waitingNamesTheEndAndWhy`, `bothEndsOnOneRecordNameTheRecord`, `rowConnectionsAreOrderedByStateThenSentence`, `promoteAllStartsExpandedAndHidesPromotedBridges`; `macos/ProvenenciaTests/PromoteModelTests.swift` `opensOnTheClickedSubjectThenMapsTheRest`, `reproposeKeepsBridges` | Swift tests pin Swift's rule by setting targets and reading `connections` |

**Precedent.** `move_reason` (S13-24, [#639](https://github.com/mendahu/provenencia/pull/639)): a `graphalign` code string on the proposal wire, a typed `PromoteMoveReason` in `ConclusionVocabulary.swift` with `init(wire:)` and `.unrecognized`, decoded once in `GoStore`, a guard row. Copy that shape. For a guard row whose list comes from Go, copy `partTypeKeys()` (S14-09, [#649](https://github.com/mendahu/provenencia/pull/649)). Do not copy the hand-written `want:` lists of the older rows.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The filing rule lives in `core/connectrules` *(new file `filing.go`)*: `type Edge struct{ From, To string }`, `type Refusal string` with `RefusalSelfLink = "self_link"` and `RefusalCycle = "cycle"`, `func Refusals() []Refusal`, `func ClosesCycle(edges []Edge, from, to string) bool`, `func Refuse(bridgeTypeKey, termKey string, ends []string, filed func() ([]Edge, error)) (Refusal, error)` (`""` files). `ends` are one record key per endpoint, in the bridge's endpoint order. `Refuse` returns `RefusalSelfLink` when two ends are equal; else, when the term is acyclic for the bridge, it calls `filed` and returns `RefusalCycle` when `ClosesCycle(edges, ends[0], ends[1])`; `filed` is not called otherwise | connectrules is the pure owner of bridge types, endpoints and disambiguation, and both callers already import it; a callback keeps the filer's SQL read lazy (only for acyclic terms) | A new `core/bridgefile` package — a second home for bridge rules; passing edges eagerly — the filer would read every place link for every bridge |
| D-2 | `connectrules.Bridge` gains `AcyclicTerms []string` *(new field)*; `place_relationship` sets `{vocabkey.TermPartOf, "succeeded_by"}` (add `vocabkey.TermSucceededBy` *(new const)* beside `TermPartOf`; `propertyterms.KeySucceededBy` becomes `string(vocabkey.TermSucceededBy)`, as `KeyPartOf` is). An acyclic bridge's first endpoint is the `from` end. Each term is checked apart (a `part_of` edge never makes a `succeeded_by` loop) | The registry names which links may not loop, so the next one is one entry; today the filer names them inline | A map keyed by term in `identityclaims` — stays in the database layer, unreadable by `graphalign` |
| D-3 | The filer keeps its loop: `fileBridge` builds `ends` from its resolved entity ids (`string(entityID)`), calls `connectrules.Refuse` with a `filed` closure that runs today's SQL, generalized to the rule's keys (`st.key = rule.BridgeTypeKey`, disambiguation `rule.Disambiguation`, ends `rule.Endpoints[0/1].PropertyKey`) and returns `[]connectrules.Edge`. `placeRelationshipWouldCycle`, the inline self-link check and the from/to reorder go. A refusal still leaves the bridge unfiled and is not an error | One rule, two loaders. Behavior at Done is unchanged | Moving filing out of SQL — out of scope |
| D-4 | `core/graphalign` gets `connections.go` *(new)*: `type ConnectionOutcome string` with `ConnectionFiles = "files"`, `ConnectionSkipped = "skipped"`, `ConnectionUnset = "unset"`, `ConnectionNotOnPage = "not_on_page"`, `ConnectionSelfLink = ConnectionOutcome(connectrules.RefusalSelfLink)`, `ConnectionCycle = ConnectionOutcome(connectrules.RefusalCycle)`; `func ConnectionOutcomes() []ConnectionOutcome` (all six); `type PageBridge struct { ID []byte; TypeKey, TermKey string; Ends [][][]byte }` (per endpoint in rule order, the Subjects its Observations name); `type EndTarget struct { Target Target; HandleID []byte; Waits ConnectionOutcome }` (a Subject's page target: a handle, `TargetNew`, or `Waits` one of skipped / unset / not_on_page); `type Connection struct { BridgeID []byte; Outcome ConnectionOutcome; Waiting []byte }`; `func Connections(bridges []PageBridge, ends map[string]EndTarget, canon []CanonEdge) []Connection`. `Proposal` gains `Connections []Connection` *(new field)*, set by `promotealign` after `Align` | graphalign is Promote's pure engine and owns `Target`; the outcome joins the move and target codes it already sends | Deciding in `promotealign` — not pure, harder to table-test; in Swift — the bug |
| D-5 | `Connections` runs the filer's order: bridges sorted by `ID` bytes. Per bridge: (1) an endpoint with no Subject, or a type-keyed bridge (`connectrules.HasDisambiguation` and not `DisambiguationRole`) with no `TermKey` → `not_on_page`, `Waiting` nil; (2) the first Subject, in endpoint then value order, whose `EndTarget.Waits` is set → that outcome, `Waiting` = that Subject; (3) an endpoint whose Subjects land on two different record keys → `not_on_page`, `Waiting` nil; (4) `connectrules.Refuse` over the record keys, with `filed` = the canon edges of that type and term plus the page connections already decided `files` of that type and term → `self_link`, `cycle` or `files`. A record key is `"h:" + string(HandleID)` for a handle and `"n:" + string(subjectID)` for New | Mirrors `fileBridge` step by step, and the batch's order, so an earlier page link that files is seen by a later one (AC-1.4). Missing end, two records and no type all read "Waits for its other end" (user, 2026-11-03) | An "incomplete" outcome (user declined); ignoring page links — AC-1.4 would count 2 |
| D-6 | A canon edge becomes a directed `connectrules.Edge` when `KeyFrom.Type` is the bridge type and `KeyFrom.Term` the term: `From → To` when `KeyFrom.End` is the bridge's first endpoint property, else `To → From` | `loadCanon` keys links through `segmentcompare.Keys`; an acyclic term has no inverse (T-3), so `Term` is the recorded term and `End` says the direction | Re-reading the catalog in the proposal — breaks the flat query count |
| D-7 | The canon already holds every link the loop check needs: each end that is a handle is a fixed, anchor or candidate handle, which `loadCanon` seeds, and the walk follows every link of every reached handle, so the whole component of each end is in `Canon.Edges`. A New end has no filed link. No new catalog read | Map scope: "a proposal runs no catalog scan" | A bounded walk for places — would miss a long chain |
| D-8 | `promotealign.Propose(c, graphID, fixed, only []byte)` *(new parameter)*; nil or empty means every row. `propose` builds `ends` after `Align`: anchors (already-promoted primaries from `mergeFixed`) → their handle, regardless of scope; with `only` set, every other unanchored Subject except `only` → `not_on_page`; a caller fixed row → its handle, `TargetNew`, or `Waits: skipped`; otherwise the row's `Start()`: handle, or `Waits: unset` when empty. A Subject that is not a primary Subject of the graph → `not_on_page`. Then `pageBridges(eg, vocab)` *(new, unexported, in `core/database/promotealign/connections.go` (new))* reads every bridge Subject on the graph whose type is a `connectrules.BridgeTypeKeys()` key of origin `provenencia` and that has no accepted member, with its endpoint and disambiguation Observations of origin `provenencia`, any polarity (as the filer reads them) | Mirrors exactly what Done sends (`PromoteFlow.batchRows`: in-scope, non-anchor rows with a target) and what the filer reads; one place that knows the page's scope | A Swift overlay for "not on the page" — keeps part of the rule in Swift; returning both scopes — twice the work every proposal |
| D-9 | Wire: `ProposePromoteGraphAlignmentRequest.only_subject_id = 5` *(new)*; `message PromoteGraphAlignmentConnection { string bridge_subject_id = 1; string outcome = 2; string waiting_subject_id = 3; }` *(new)*; `ProposePromoteGraphAlignmentResponse.connections = 3` *(new)*, in `Connections` order. `only_subject_id` that does not parse is `invalid_argument` as `graph_id` is (`parseID` via `optionalID`) | Additive fields; the outcome is a code, the page words it | Folding outcomes into the row message — a connection is not a row |
| D-10 | Swift: `PromoteConnectionOutcome` *(new)* in `Platform/ConclusionVocabulary.swift` (`files`, `skipped`, `unset`, `notOnPage`, `selfLink`, `cycle`, `unrecognized(String)`, `init(wire:)`); `CatalogPromoteGraphAlignmentConnection` *(new)* `{ bridgeSubjectID: String; outcome: PromoteConnectionOutcome; waitingSubjectID: String }` in `GenealogyStore.swift`; `CatalogPromoteGraphAlignmentProposal.connections: [CatalogPromoteGraphAlignmentConnection] = []` *(new)*. The protocol method gains `onlySubjectID: String?` (nil sends empty). `GoStore` decodes once | `macos-client.mdc`: a wire enum is typed in `Platform/` with a guard row | A `String` outcome on the flow |
| D-11 | `PromoteFlow` stores `outcomes: [String: CatalogPromoteGraphAlignmentConnection]` *(new)*, keyed by bridge id, set by `load` and `merge` from the proposal. `connections` maps each non-`alreadyPromoted` `BridgeFact` that has an outcome; a bridge with no outcome is not shown and not counted. Mapping: `files` → `.promotes`; `skipped` / `unset` / `notOnPage` → `.waits(endID: waitingSubjectID` or nil when empty`, why:)`; `selfLink` → `.selfLink(ref:)` with `ref` = `row(fact.endA)?.shownRef ?? row(fact.endB)?.shownRef ?? ""`; `cycle` → `.closesCycle` *(new case)*; `unrecognized` → `.waits(endID: nil, why: .notOnPage)`. `resolve`, `EndResolution` and `lined` go. `stateRank`: `.closesCycle` ranks with `.selfLink` (2) | Swift words what Go decided and decides nothing; the ref is display, read from the rows as the page shows them | Sending the ref from Go — New ends have only a Source ref, which the row already holds |
| D-12 | The subrows and the footer show the latest proposal's outcomes. After a decision they keep the previous outcomes until the next proposal lands (Done already waits for it: `canFinish` checks `isProposing`). No Swift-side prediction | A prediction is a second rule | Optimistic local states — the bug this PR removes |
| D-13 | `PromoteFlow.onlySubjectID` *(new computed)*: nil when `mappedRest`, else `entryID`. `PromoteModel.load` and `repropose` pass it. `PromoteModel.mapRest()` calls `flow.mapRest()` then `proposeAfterDecision()` | The scope is part of what Done files, so changing it asks Go again | Leaving `mapRest` local — AC-2.3's states would stay "not on this page" |
| D-14 | `FakeStore.buildPromoteProposal` returns `connections: []`; scripted `promoteProposals` carry the connections a test needs. `FakeStore` records `lastPromoteOnlySubjectID: String?` *(new)* beside `lastPromoteFixed` | Working rule "FakeStore is a seeded fake": no filing rule in the fake | Computing outcomes in the fake — a second engine |
| D-15 | Copy (user, 2026-11-03): warning "Not promoted · these places would form a loop" (`promote.connection.cycle`), VoiceOver "Not promoted, these places would form a loop." (`promote.connection.spoken.cycle`). `connectionState` returns nil for `.closesCycle`, as for `.selfLink`; the view shows `PVCallout(tone: .warning, message: L10n.Promote.connectionCycle, compact: true)` where the self-link callout is | One string for both terms, drawn as the self-link warning (map: no new chrome) | Per-term or naming the places (user declined) |
| D-16 | Guard row: `{name: "promote connection outcomes", path: swiftVocabulary, enum: "PromoteConnectionOutcome", re: wireCase, want: sorted(connectionOutcomes()...)}` with `connectionOutcomes()` *(new helper beside `partTypeKeys`)* returning each `graphalign.ConnectionOutcomes()` string | Spike 11 leftover in the map scope: the new enum reads its list from Go | A hand-written list |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/connectrules/filing.go` `Edge`, `Refusal`, `Refusals`, `ClosesCycle`, `Refuse` *(new)* | file / types / funcs | `identityclaims.placeRelationshipWouldCycle` (searched `cycle`, `Cycle` in `core/`: also `autoreconcile` fold cycles, `placechains` display walk, `objectpath`, none a filing rule) | It is the code being moved; it reads SQL and cannot be called from the proposal. Extending `connectrules` rather than a new package |
| `connectrules.Bridge.AcyclicTerms` *(new field)* | registry field | `propertyterms.KeyPartOf` / `KeySucceededBy` inline in the filer | The registry already owns each bridge's endpoints and disambiguation |
| `vocabkey.TermSucceededBy` *(new const)* | const | `vocabkey.TermPartOf` | Same shape; `propertyterms.KeySucceededBy` derives from it |
| `core/graphalign/connections.go` `ConnectionOutcome`, `ConnectionOutcomes`, `PageBridge`, `EndTarget`, `Connection`, `Connections` *(new)* | file / types / func | `graphalign.Reason` / move codes (searched `Outcome`, `Connection` in `core/`) | No connection type exists in Go; codes follow `Reason`'s shape |
| `graphalign.Proposal.Connections` *(new field)* | field | `Proposal.Rows` | The proposal is what the handler maps |
| `core/database/promotealign/connections.go` `pageBridges`, `pageEnds` *(new)*, unexported | funcs | `loadLayerBridges` (walks `WalkEvidenceGraph`, positive polarity only, merged links, both ends required) | The layer merges links and drops a bridge with a missing end; the filer reads per bridge Subject with any polarity |
| `promotealign.Propose` `only` *(new param)* | param | — | — |
| `PromoteGraphAlignmentConnection`, `only_subject_id`, `connections` *(new proto)* | FFI | `PromoteGraphAlignmentRow.move_reason` | Additive |
| `PromoteConnectionOutcome` *(new Swift enum)* | type | `PromoteMoveReason` | One enum per wire code set (`macos-client.mdc`) |
| `CatalogPromoteGraphAlignmentConnection` *(new Swift struct)*, `CatalogPromoteGraphAlignmentProposal.connections` | type / field | `CatalogPromoteGraphAlignmentRow` | — |
| `PromoteFlow.outcomes`, `onlySubjectID`, `ConnectionState.closesCycle` *(new)* | field / computed / case | `PromoteFlow.ConnectionState.selfLink` | Extends the existing state |
| `FakeStore.lastPromoteOnlySubjectID` *(new)* | test hook | `lastPromoteFixed` | Same shape |
| `L10n.Promote.connectionCycle`, `promote.connection.spoken.cycle` *(new keys)* | key | `connectionSelfLink`, `promote.connection.spoken.selfLink` | Same namespace |
| `connectionOutcomes()` *(new test helper)* | helper | `partTypeKeys()` | Same shape |

**Builds on.** `connectrules` (bridge registry), `graphalign` codes on the wire (`Reason`, `Target`), `loadCanon`'s canon edges, the `macos-client.mdc` wire-enum pattern, S14-09's Go-read guard row.
**Deletes.** `identityclaims.placeRelationshipWouldCycle`; the inline self-link check and from/to reorder in `fileBridge`; Swift `PromoteFlow.resolve(_:in:)`, `EndResolution`, `lined(_:_:)`; the Swift tests that pin Swift's rule by targets alone (`connectionsFollowTheirEnds`, `waitingNamesTheEndAndWhy`, `bothEndsOnOneRecordNameTheRecord` are rewritten to script Go's outcomes, T-9). The `propertyterms.go` comment "Walk / cycle behaviour is keyed here in Go" is rewritten to name `connectrules.Bridge.AcyclicTerms`. `promote-graph-alignment.md` §9.1's false claim.
**Next adds.** A new filing refusal is one `Refusal` constant, one `ConnectionOutcome`, one Swift case (the guard row fails until it is added) and one string. A new loop-checked term is one `AcyclicTerms` entry. S14-20's runner replaces `proposeAfterDecision`'s body and keeps `onlySubjectID`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| `.closesCycle` connection state | `PromoteConnectionSubrows.subrow`, `L10n.Promote.connectionState`, `connectionSpoken`, `PromoteFlow.stateRank` | Every `switch` over `ConnectionState` gets the case (the compiler lists them); the callout and spoken copy are D-15 |
| A snapshot bridge with no outcome (snapshot read after the proposal at load, or a bridge added in another window) | `PromoteFlow.connections` | Not shown, not counted (D-11); a later write makes Done fail `promote.stale`, which proposes again |
| "Map the rest" now proposes | `PromoteModel.mapRest` → `proposeAfterDecision` → `repropose` | `isProposing` holds Done; latest-wins by `proposeGeneration` as for a decision |
| An outcome naming a waiting Subject that is not a row | `bridge.name(of:)` returns "" → "Waits for its other end" | As today |

### Scale

At the map's Scale (a census graph with 300 Citations; 37,000 Subjects; 6,000 events on one place): `Propose` adds no catalog read (`TestProposeQueryCountDoesNotGrowWithTheLayer` and `TestProposeQueryCountStaysFlatWhenWarm` hold it). `pageBridges` is one pass over the loaded graph's Subjects and Observations. `Connections` is linear in the bridges and their values, plus, for each `place_relationship` connection of an acyclic term whose ends are records, one BFS over that term's canon edges and earlier page edges (each edge visited once). A place with 6,000 events adds `location` edges, which no acyclic term reads. The filer's work at Done is unchanged (one SQL read per acyclic bridge, as today).

### Edges

- **Assumptions.** `ClosesCycle` must terminate on an input that already loops (a crafted catalog): BFS with a visited set (T-1 "edges already loop"). The filer's order is bridge Subject id order (`bridgeHeads ORDER BY s.id`); `Connections` sorts by `bytes.Compare` on the same 16-byte ids (T-4 "page links in id order"). An acyclic term has no inverse and is directed, so a canon edge's `Term` and `End` are the recorded ones (T-3); if a seed ever gives one an inverse, T-3 fails.
- **Neighbors.** `part_of` loop refused, `succeeded_by` over the same places files (AC-1.3, T-1, T-2, T-5); a split (A `part_of` B, A `part_of` C) files (T-1); a person `relationship` between two records never calls `filed` (T-2); a self-link wins over a loop (T-2); an anchor end resolves to its record whatever the scope (T-4).
- **Removed inputs.** A Skip or Unset end makes its connections wait, and a later page link that would have looped files (AC-1.5, T-4 "skipped end frees the loop"). A decided row outside the scope cannot happen: the page decides only visible rows. An already-filed bridge is not reported (T-5).
- **Real data.** The acyclic table is two terms on one bridge type; the whole shipped `seedTerms` is checked by T-3.
- **Untrusted size.** Bridges, values per endpoint and canon edges come from the catalog. Each is read once; BFS marks each record key once. No bound is added: the work is linear in what the graph cache already holds.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `TermSucceededBy BridgeTerm = "succeeded_by"` *(new)* |
| `core/connectrules` | `Bridge.AcyclicTerms []string` *(new)*; `place_relationship` sets `{string(vocabkey.TermPartOf), string(vocabkey.TermSucceededBy)}`. `filing.go` *(new)*: `Edge`, `Refusal`, `RefusalSelfLink`, `RefusalCycle`, `Refusals()`, `ClosesCycle`, `Refuse` (D-1) |
| `core/database/propertyterms` | `KeySucceededBy = string(vocabkey.TermSucceededBy)`; comment names `connectrules.Bridge.AcyclicTerms` |
| `core/database/identityclaims` | `fileBridge` calls `connectrules.Refuse`; `filedLinks(tx, rule, termKey)` *(new, unexported)* replaces `placeRelationshipWouldCycle` (D-3) |
| `core/graphalign` | `connections.go` *(new)* (D-4, D-5, D-6); `Proposal.Connections` |
| `core/database/promotealign` | `Propose(c, graphID, fixed, only)`; `propose` fills `prop.Connections` from `pageBridges`, `pageEnds` and `canon.Edges` (D-7, D-8) |

```go
// connectrules/filing.go (sketch)
type Edge struct{ From, To string }
type Refusal string
const (
	RefusalSelfLink Refusal = "self_link"
	RefusalCycle    Refusal = "cycle"
)
func Refusals() []Refusal
func ClosesCycle(edges []Edge, from, to string) bool
func Refuse(bridgeTypeKey, termKey string, ends []string, filed func() ([]Edge, error)) (Refusal, error)
```

Errors: none new. `Refuse` returns `filed`'s error.

### FFI

`api/proto/engine.proto`: `ProposePromoteGraphAlignmentRequest.only_subject_id = 5` *(new)*; `PromoteGraphAlignmentConnection` *(new)*; `ProposePromoteGraphAlignmentResponse.connections = 3` *(new)* (D-9). Handler `api/ffi/handlers/promote_align.go`: parse `only_subject_id` with `optionalID`, pass to `Propose`; `proposalProto` maps `prop.Connections`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Loop warning | `PVCallout(tone: .warning, compact: true)` | `PromoteView.swift` `PromoteConnectionSubrows.subrow` | Beside the self-link callout; same placement |
| Muted sentence | existing `promotes ? … : PVColor.textMuted` | same | Unchanged; `.closesCycle` is not `.promotes` |

States: a connection is shown only once Go has answered for it. Accessibility: the subrow's `accessibilityLabel` via `connectionSpoken` (D-15).

### Strings

L10n keys *(new)*: `promote.connection.cycle` = "Not promoted · these places would form a loop"; `promote.connection.spoken.cycle` = "Not promoted, these places would form a loop.". Removed: none. Old-scope copy: "Promotes on Done" and the footer's "%lld connections" now count Go's `files` outcomes; their wording stays true.

## 7. Steps

**STEP-1 — One filing rule in connectrules**
Files: `core/vocabkey/vocabkey.go`, `core/connectrules/connectrules.go`, `core/connectrules/filing.go` *(new)*, `core/connectrules/connectrules_test.go`, `core/database/propertyterms/propertyterms.go`, `core/database/subjectvocab/subjectvocab_test.go`. Skill: `add-seeded-vocabulary` (read only).
Do: D-1, D-2. Add T-1, T-2, T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/connectrules/... ./core/database/subjectvocab/... ./core/database/propertyterms/...` passes.

**STEP-2 — The filer calls the rule**
Files: `core/database/identityclaims/bridgefile.go`.
Do: D-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/identityclaims/... ./core/database/promote/...` passes, including "place relationship files; cycles refuse; split and amalgamation ok" (T-6); `rg placeRelationshipWouldCycle core` has no hit.

**STEP-3 — Connections in graphalign**
Files: `core/graphalign/connections.go` *(new)*, `core/graphalign/types.go`, `core/graphalign/connections_test.go` *(new)*. Skill: `add-match-rule` (layer package conventions; no knob is added).
Do: D-4, D-5, D-6. Add T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/...` passes.

**STEP-4 — The proposal decides connections**
Files: `core/database/promotealign/promotealign.go`, `core/database/promotealign/connections.go` *(new)*, `core/database/promotealign/promotealign_test.go`, `core/database/promotealign/bench_test.go`. Skill: `use-catalog-session`.
Do: D-7, D-8; every existing `Propose` caller passes `nil`. Add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes, including `TestProposeQueryCountDoesNotGrowWithTheLayer` and `TestProposeQueryCountStaysFlatWhenWarm`.

**STEP-5 — Wire and handler**
Files: `api/proto/engine.proto`, generated code (`./scripts/generate-proto.sh`), `api/ffi/handlers/promote_align.go`, `api/ffi/handlers/promote_align_test.go`. Skill: `add-ffi-handler`.
Do: D-9. Add T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-6 — Swift wire type and guard row**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift` (stub signature), `api/ffi/swift_vocabulary_test.go`.
Do: D-10, D-14, D-16. Every caller of `proposePromoteGraphAlignment` passes `onlySubjectID` (STEP-7 sets the real value; this step passes `nil`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes (T-8); the app builds.

**STEP-7 — The page reads Go's outcomes**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: D-11, D-12, D-13. Rewrite the tests in T-9; add T-10, T-11.
Check: `xcodebuild test` on `ProvenenciaTests` passes; `rg "EndResolution|func resolve\(" macos/App` has no hit.

**STEP-8 — Loop copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Promote/PromoteView.swift`. Skill: `add-localized-string`.
Do: D-15.
Check: the app builds with no Swift warning; T-11 passes.

**STEP-9 — Docs**
Files: `docs/promote-graph-alignment.md` §9.1, `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (map row), `docs/deployment-plan/spike-14/app-health-audit-2.md` (B-23 row: fixed in this PR), this plan (status), `VERSION`, `core/version.go`, `MARKETING_VERSION`.
Do: §9.1's last two refusal bullets say: the filer and the proposal decide a self-link and a loop through `connectrules.Refuse`; the proposal reports each connection's outcome (`graphalign.Connections`), and the page words it under the connection and leaves it out of the footer's count. Then the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-22-promote-wire-says-what-go-decided.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/connectrules/connectrules_test.go` `TestClosesCycle` *(new)* | chain a→b, b→c; add c→a → true; add a→c → false; split a→b, a→c; add b→c → false; empty edges, a→a → true; edges already loop (a→b, b→a), add c→d → false and returns | D-1, AC-1.1 |
| T-2 | Go | same file `TestRefuse` *(new)* | ends equal → `self_link`, `filed` not called; `place_relationship`/`part_of` with filed b→a, ends [a, b] → `cycle`; same with `succeeded_by` and filed `part_of` edges passed only for `part_of` → `""`; `relationship`/`spouse` ends [a, b] → `""`, `filed` not called; `filed` error → returned | D-1, D-2, AC-1.3 |
| T-3 | Go | `core/database/subjectvocab/subjectvocab_test.go` `TestAcyclicTermsAreSeededDirectedWithNoInverse` *(new)* | every `connectrules.Bridges()` `AcyclicTerms` key is a `seedTerms` term of that bridge's disambiguation Property, `Directed`, empty `Inverse`; each acyclic bridge's first endpoint is `from` | D-2, D-6 |
| T-4 | Go | `core/graphalign/connections_test.go` `TestConnections` *(new)*, table | both ends on records → `files`; one end `Waits: skipped` → `skipped`, `Waiting` that Subject; one end `Waits: unset` → `unset`; not_on_page (scope); first waiting end in endpoint order; endpoint with no Subject → `not_on_page`, `Waiting` nil; endpoint Subjects on two records → `not_on_page`, nil; `relationship` with empty `TermKey` → `not_on_page`, nil; both ends one record → `self_link`; canon York `part_of` Ontario + page Ontario `part_of` York → `cycle`; canon edge stored To→From (KeyFrom.End = `to`) read the right way; "page links in id order": two New places A↔B `part_of` → lower id `files`, higher `cycle`; "skipped end frees the loop": same with A Skip → both `skipped`; New + New → `files` | D-4 – D-6, AC-1.1, AC-1.4, AC-1.5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| T-5 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeReportsEachConnection` *(new)* | Tree York `part_of` Ontario (`f.placeRel`, `f.promote`); Source B Ontario2 `part_of` York2 fixed on the two handles → the connection's outcome `cycle`; then `promote.SaveBatch` with the rows → no new `place_relationship` handle; `succeeded_by` case → `files` and Done mints one; `only` = York2 with Ontario2 unfixed and unpromoted → `not_on_page` waiting Ontario2; an already-filed bridge → not in `Connections`; the number of `files` outcomes equals the associations Done mints | D-7, D-8, AC-1.1 – AC-1.3, AC-2.3, AC-2.7 |
| T-6 | Go | `core/database/promote/bridges_test.go` "place relationship files; cycles refuse; split and amalgamation ok" (existing) | unchanged expectations | D-3, AC-3.3 |
| T-7 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposeSendsEachConnection` *(new)* | a graph with one participation, event row Skip → `connections[0]` `{bridge_subject_id, outcome: "skipped", waiting_subject_id: event}`; `only_subject_id: "nope"` → invalid argument | D-9 |
| T-8 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row "promote connection outcomes" *(new)* | Swift cases = `graphalign.ConnectionOutcomes()` | D-16, AC-3.2 |
| T-9 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `connectionsFollowGosOutcome` *(rewrites `connectionsFollowTheirEnds`)*, `waitingNamesTheEndAndWhy`, `bothEndsOnOneRecordNameTheRecord`, `aConnectionWithOneRowEndShowsOnlyThere`, `rowConnectionsAreOrderedByStateThenSentence` (rewritten to script outcomes), `aConnectionGoDidNotDecideIsNotShown` *(new)*, `aLoopIsNotCounted` *(new)* | each outcome → its `ConnectionState`; `unrecognized("x")` → `.waits(endID: nil, why: .notOnPage)`; a bridge without an outcome → absent from `connections`, count 0; `.cycle` → `.closesCycle`, not counted, ranked with self-link; a self-link's ref is the end row's `shownRef` | D-11, AC-1.1, AC-2.1, AC-2.2, AC-2.4, AC-2.5, AC-2.6 |
| T-10 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `theProposalCarriesThePagesScope` *(new)*, `mappingTheRestProposesAgain` *(new)* | load on Subject "a" → `lastPromoteOnlySubjectID == "a"`; Promote all → nil; `mapRest()` → a second proposal with nil, and the scripted second answer's outcomes replace the first's; footer count = scripted `files` count | D-12, D-13, AC-2.3, AC-2.7 |
| T-11 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `reasonCodesParseThroughTheWireVocabulary` (extended) and `aLoopIsWordedAsALoop` *(new)* | the six wire codes parse; `"later"` → `.unrecognized`; `L10n.Promote.connectionState(.closesCycle, …) == nil`; `connectionSpoken` ends "Not promoted, these places would form a loop." | D-10, D-15, AC-1.6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promotealign` (query-count tests, `bench_test.go` signature), `core/database/promote/bridges_test.go`, `PromoteFlowTests`, `PromoteModelTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`placeRelationshipWouldCycle`, `PromoteFlow.resolve`, `EndResolution`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: only the existing PVCallout)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: none)
- [ ] Domain doc updated: docs/promote-graph-alignment.md §9.1; app-health-audit-2.md B-23 row marked fixed in this PR
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. New project. Source A: places York and Ontario, York `part_of` Ontario. Promote all, both New, Done. The toast counts 1 connection.
2. Source B: places Ontario and York, Ontario `part_of` York. Promote all. Set Ontario to Ontario's record and York to York's. The subrow under each reads the sentence muted with "Not promoted · these places would form a loop"; the footer reads "Promote 2 · 0 connections". With VoiceOver on, the subrow reads "…. Not promoted, these places would form a loop." (AC-1.1, AC-1.6)
3. Done. The toast counts 0 connections; York's page shows no new place relationship (AC-1.2).
4. Source C: Ontario `succeeded_by` York, both on their records: "Promotes on Done", footer 1 connection; Done files it (AC-1.3).
5. `rg "placeRelationshipWouldCycle|EndResolution|func resolve" core macos/App` prints nothing (AC-3.1).
6. Source D: a person and a birth event joined by a participation. Open Promote on the person only: the subrow reads "Waits for <birth>"; choose "Map the rest", set the birth New: "Promotes on Done" once the page settles (AC-2.3). Set the birth Skip: "Waits for <birth>" (AC-2.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S14-20 (scoped, not planned) rewrites the proposal runner in `PromoteModel` | Whichever lands second rebases; D-13 adds one call to `proposeAfterDecision`, which S14-20 keeps as the entry point |
| The subrows lag a decision by one proposal | D-12; Done waits for the proposal; T-10 |
| `Propose`'s signature change touches tests and the bench | STEP-4 updates every caller (`rg "promotealign.Propose\("`) |

**Stop and report if:**

- `fileBridge` refuses or waits on a case §4 does not list (another `return nil, nil, false, nil` path); `Connections` would miss it.
- `TestProposeQueryCountDoesNotGrowWithTheLayer` or `TestProposeQueryCountStaysFlatWhenWarm` fails after STEP-4: the canon does not hold an edge D-7 says it does.
- T-5's count of `files` outcomes differs from the associations `SaveBatch` mints on any case.
- A `place_relationship` end that is a record is not in `Canon.Handles` (D-7's claim is false).
- `seedTerms` gives `part_of` or `succeeded_by` an inverse (T-3 fails).
- Done's filing order is not bridge Subject id order.

## 12. Open questions

None.

## 13. PR

**Title:** S14-22 — Promote's wire says what Go decided
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-22-promote-wire-says-what-go-decided.md`.
