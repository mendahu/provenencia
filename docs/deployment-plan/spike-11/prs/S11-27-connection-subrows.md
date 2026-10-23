# S11-27 — Connection subrows on Promote

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 6 — Connection subrows
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S11-D1](../design/S11-D1-promote-rows.md) §4.2 (CS-1 – CS-9) and §8 board notes, board agreed 2026-10-20
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), file `Promote flow.dc.html`, frames **16** – **19** (S11-27's; 14 – 15 were S11-18's). If you cannot open it, build from §6 macOS, which describes the frames in words. Where the board and this plan differ, this plan wins: D-4 (wording) and D-5 (the other end's note) are deliberate departures the user chose on 2026-10-23
**Depends on:** S11-26 (landed [#527](https://github.com/mendahu/provenencia/pull/527)), S11-D1 (agreed)
**Blocks:** —
**Branch:** `s11-27-connection-subrows`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [Connection subrows](../promote-connection-subrows.md) "The idea" and "Open questions for design"; [S11-D1](../design/S11-D1-promote-rows.md) §2 (domain facts: states, one host per connection, busy rows), §4.2, §6 inventory, §8; [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3 ("Bridges aren't rows"); working rules "High tolerance for churn"; S11-26 [plan](S11-26-connections-by-ends.md) D-3 (`connections`)
**Planned:** 2026-10-23 against `main` @ `e33db0f`
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

On the Promote page, each connection (participation, location, relationship, place relationship) shows as a subrow under a row it joins. Its state comes only from its ends: it files, it waits for a named end, or both ends are one record. A researcher who sets John to Skip sees "Mary Smith is the parent of John Smith" dim under John's row at once, and sees it turn back to "Files on Done" when John is set to New. The bottom connections panel and the footer button that opened it go away. The footer keeps a count of what files.

**Why now.** S11-26 (#527) removed the switch-off, so a connection's state is a pure function of its ends. This PR moves that state next to the rows that decide it ([connection subrows](../promote-connection-subrows.md) "The problem" §1). It closes slice 6 and the dogfood bar item 9.

**What stays.** Filing (Go is not touched). `PromoteFlow.connections` and `connectionCount` keep their meaning: every connection not yet filed, and how many of them file. The row's columns, the target menu, the band, Contested, Updated and Decided badges, the Updated caption, the conflict and duplicate callouts, both groupings, the anchors block, Done and Cancel. The toast after Done (`L10n.Promote.filedTitle`). The connection sentence wording (`EvidenceBridgeEdgeSummary.sentence`).

## 2. User stories

**US-1.** As a researcher, I see each connection under a row it joins, so I can see what Done will file without opening anything.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Promote all on a Source with person John (row), event Birth (row), and a participation John → Birth | The subrow is under **Birth** (participation is hosted by its event, D-2), not under John |
| AC-1.2 | Location Birth → Toronto, both rows | Under **Birth** |
| AC-1.3 | Place relationship "Toronto is part of Ontario", both rows | Under **Toronto** (the from end) |
| AC-1.4 | Relationship "Mary Smith is the parent of John Smith" (`parent`), both rows | Under **John** (directed: the end named last) |
| AC-1.5 | Relationship "Ann is the spouse of Bill" (`spouse`), both rows | Under **Ann** (symmetric: the end named first) |
| AC-1.6 | Relationship with a researcher's own type, or no type, Mary → John | Under **John** (anything not a product symmetric type reads as directed) |
| AC-1.7 | The participation John → Birth when Birth is an anchor (already filed) and John is a row | Under **John**, with the note "Birth is already filed" (CS-4) |
| AC-1.8 | Promote opened on John alone (not mapped) and a participation John → Birth | Under **John**, state "Waits for Birth", note "Birth is not on this page" |
| AC-1.9 | A connection whose ends are both anchors, or neither of whose ends is a row on the page | No subrow anywhere; it still counts in the footer if it files (D-7) |
| AC-1.10 | A connection that is already filed (`alreadyFiled`) | No subrow, not counted (as today) |

**US-2.** As a researcher, a subrow's state follows its ends as I change targets, so the page shows what Done files.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Mary and John both New, relationship between them | State "Files on Done", sentence in primary text |
| AC-2.2 | Set John to Skip | The same subrow reads "Waits for John Smith" and its sentence is muted; nothing re-lays out |
| AC-2.3 | Then set John to New | Back to "Files on Done" |
| AC-2.4 | John with no target yet (unset) | "Waits for John Smith"; VoiceOver says "…, who has no target yet" |
| AC-2.5 | Both ends waiting (both Skip) | The state names the end named first in the connection (end A): "Waits for Mary Smith" |
| AC-2.6 | Mary and John both set to record PER-7QH3D | The state column is empty and a compact warning callout under the sentence reads "Not filed · both ends are PER-7QH3D" |

**US-3.** As a researcher, a busy row stays short, and a row whose connections live elsewhere points to them.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Birth hosts four participations, one waiting | Three subrows show, ordered files, waits, self-link, then by sentence; a "+1 more" ghost button expands the fourth in place and becomes "Show fewer" |
| AC-3.2 | John is the non-host end of two connections hosted by Birth | Under John's row, on its own line at the note indent: a ghost button "+2 connections" |
| AC-3.3 | Press "+2 connections" | The page scrolls to Birth's row; the two subrows that join John fill with `PVColor.surfaceSelected` for about 1.6 s; if one of them was past the third, Birth's list expands first |
| AC-3.4 | Group by assessment | Every subrow stays under its host row, wherever that row's section is |

**US-4.** As a VoiceOver user, I hear each connection and why it waits.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Focus a filing subrow under John | One element: "Mary Smith is the parent of John Smith. Files on Done." |
| AC-4.2 | John is Skip | "Mary Smith is the parent of John Smith. Waits for John Smith, who is skipped." |
| AC-4.3 | Self-link | "… Not filed, both ends are PER-7QH3D." |
| AC-4.4 | The subrow list | Is a container named "Connections hosted by John Smith" |
| AC-4.5 | The "+2 connections" button | Reads "2 connections hosted on other rows" |

**US-5.** As a researcher, the footer says what Done does, without a panel.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | 12 rows file, 3 skipped, 0 unset, 22 connections file | Footer reads "File 12 · Skip 3 · 22 connections"; no connections button, no panel |
| AC-5.2 | 12 file, 0 skipped, 2 unset, 1 connection | "File 12 · 2 unselected · 1 connection" |
| AC-5.3 | 5 file, 0 skipped, 0 unset, 0 connections | "File 5 · 0 connections" |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConnectionState` carries why a connection waits and the self-link's ref (D-1) | STEP-1, AC-2.1 – AC-2.6 |
| `BridgeFact.kind` and `.symmetric`, filled in `PromoteModel.bridgeFacts` (D-3) | STEP-1, STEP-2, AC-1.4 – AC-1.6 |
| `PromoteFlow.connectionHosts`: host by rule, fallback to the other end, pointers, order (CS-1, CS-3, CS-4, CS-5, CS-7; D-2, D-6) | STEP-1, AC-1.1 – AC-1.10, AC-3.1, AC-3.2 |
| `PromoteFlow.unsetCount` for the footer (D-8) | STEP-1, AC-5.2 |
| Strings: subrow state, spoken state, other-end note, self-link callout, container label, "+N connections", "+N more", "Show fewer", footer parts; delete the panel's strings (D-9) | STEP-3, AC-4.*, AC-5.* |
| Subrows under `PromoteAlignmentRow`, the "+N connections" pointer with scroll and highlight, "+N more" / "Show fewer" (CS-1 – CS-8) | STEP-4, US-1 – US-4 |
| Remove `connectionsPanel`, `connectionRow`, `connectionMeta`, the footer toggle, `bridgesOpen` (flow, model, view) (CS-9) | STEP-5, AC-5.1 |
| Footer summary with the unselected part, zero Skip and unselected parts left out (CS-9, §8 frame 19) | STEP-5, AC-5.1 – AC-5.3 |
| Docs: `promote-graph-alignment.md` §3 "Bridges aren't rows"; completed.md, map row, brief archive | STEP-6 |

| Out | Home |
| --- | --- |
| Opening the evidence sheet from a subrow, or the sheet listing a row's connections | not planned (S11-20 decided, map scope entry) |
| An "open on Evidence graph" action, switch, menu or hover action on a subrow | not planned (S11-D1 §6 non-goals) |
| Showing already-filed connections dimmed | not planned (S11-D1 §2: hidden) |
| The board's "A · term · B" three-part wording | not planned (D-4, user 2026-10-23) |
| Any change to Go filing, the proposal, or the wire | not planned |
| Plural rule for `promote.toast.filedBody` (S11-26 ship log UX-1 rest) | dogfood log; this PR deletes `promote.connections.count`, the other key in that follow-up |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Connection state | `macos/App/Features/Promote/PromoteFlow.swift` `ConnectionState` (`.files`, `.endSkipped`, `.selfLink`), `Bridge`, `BridgeFact`, `lined(_:_:)`, `connections`, `connectionCount`, `resolve(_:)` | `connections` maps every bridge with `alreadyFiled == false` to `.endSkipped` when either end does not resolve, `.selfLink` when both resolve to one id, else `.files`. No case says which end waits or why. `resolve` returns nil for a missing row, a non-entry row when `mappedRest` is false, and a Skip or unset row; an anchor resolves to its handle id; New resolves to `"new:" + subjectID` |
| Bridge facts | `PromoteFlow.BridgeFact` (`id`, `sentence`, `phrase`, `term`, `endAName`, `endBName`, `mark`, `endA`, `endB`, `alreadyFiled`) | No bridge kind and no relationship direction; `mark` is derived from the kind and is the only trace of it |
| Building facts | `macos/App/Features/Promote/PromoteModel.swift` `static func bridgeFacts(_:subjects:)` | Maps `snapshot.bridges` (`SourceGraphPlacedBridge`, with `kind: EvidenceBridgeKind`) to `BridgeFact`; `sentence` from `EvidenceBridgeEdgeSummary.sentence(for:in:subjects:bridges:)` |
| End order | `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift` `Slots` doc | Participation A = person, B = event; location A = event, B = place; relationship and place relationship A = from, B = to. Relationship sentence: "%1$@ is the %2$@ of %3$@" (A is the term of B) |
| Term key | `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift` `private static func termKey(kind:in:)` | Returns the bridge's term key only when it is a product term; private |
| Inverse map | `macos/App/Platform/ConclusionVocabulary.swift` `SeededRelationshipTypeTerm.inverse` | "A symmetric term is its own inverse. The one inverse map." `spouse`, `sibling`, `cousin` map to themselves |
| Page | `macos/App/Features/Promote/PromoteView.swift` `PromoteView.body` | `ScrollView` of sections, then `if bridgesOpen.wrappedValue { connectionsPanel }`, then `footer`. No `ScrollViewReader` |
| Panel | same file, `connectionsPanel`, `connectionRow(_:)`, `connectionMeta`, `bridgesOpen` binding | The bottom panel: title, meta "N filed · M not filed", hint, a list of `connectionRow`s (mark, `endAName` + `phrase` + `endBName` or `sentence`, caption state). `phrase` is "Is the parent of" (capitalized), so the row reads "Mary Is the parent of John" |
| Footer | same file, `footer` | Ghost `PVButton(L10n.Promote.connections(count:))` with chevron that toggles `bridgesOpen` (`promote.connections.toggle`); `Text(verbatim: L10n.Promote.doneSummary(file:skip:connections:))`; Cancel; Done |
| Row | same file, `PromoteAlignmentRow` (`row`, `model`), `PromoteLayout` (`columnGap`, `kindTile`, `targetColumn`, `badgeColumn`, `noteIndent`) | A `VStack`: the column `HStack`, then the Updated caption, conflict callout and duplicate callout, each `.padding(.leading, PromoteLayout.noteIndent)` |
| Sections | same file, `sections`, `sectionBlock(_:)` | `ForEach(section.rows) { PromoteAlignmentRow(row:model:) }`; anchors render as `PromoteAnchorRow` in the "already filed" block |
| Model | `PromoteModel.setBridgesOpen(_:)`; `PromoteFlow.bridgesOpen` | Panel open state |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote.connections(count:)`, `connectionState(_:)`, `doneSummary(file:skip:connections:)`, `connectionsTitle`, `connectionsMeta(filed:unfiled:)`, `connectionsHint`; `macos/App/Resources/Localizable.xcstrings` keys `promote.connections.count`, `.filed`, `.self`, `.skipped`, `.title`, `.meta`, `.hint`, `promote.done.summary` | Panel and footer copy. `promote.connections.count` has no plural rule (S11-26 ship log UX-1 rest) |
| Kit | `macos/App/DesignSystem/Components/Button/PVButton.swift` `init(_:variant:size:icon:iconRight:loading:iconOnly:flushLeading:action:)`; `macos/App/DesignSystem/Components/Callout/PVCallout.swift` (`tone:title:message:compact:`); `macos/App/DesignSystem/Recipes/Marks/PVMark.swift`; `macos/App/DesignSystem/Tokens/PVColor.swift` `surfaceSelected`, `borderSubtle`, `textMuted`, `textPrimary`, `textSecondary` | All exist |
| Tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` `doneSendsDefaultPinsAndConnectionsFollowEnds` (expects `.endSkipped`), `promoteAllStartsExpandedAndHidesFiledBridges`, helpers `proposal(rows:)`, `row(id:target:handleID:assessment:)`, `fact(_:anchor:handleID:)` (kind `.person`); `macos/ProvenenciaTests/PromoteModelTests.swift` `viaReason(kind:edgeKeys:term:)` (builds a bridge Subject of a kind with a term through `FakeStore`), `reproposeKeepsBridges` | Cover `connections` and states; nothing covers hosting |
| Docs | `docs/promote-graph-alignment.md` §3 bullet "**Bridges aren't rows**" | Describes the summary line and the expanding list |

**Precedent.** S11-18 ([#507](https://github.com/mendahu/provenencia/pull/507), [plan](S11-18-contested-marker.md)) extended `PromoteAlignmentRow` under the same brief: flow logic and its tests first (`PromoteFlowTests`), then strings, then the view. Copy that order and its Board-line convention. S11-26 ([#527](https://github.com/mendahu/provenencia/pull/527), [plan](S11-26-connections-by-ends.md) D-3) made `connections` the one list; build on it, do not add a second walk over `bridges`. The footer's joined parts copy `CitationComposerEvidenceColumn.swift` (`parts.joined(separator: " · ")`). Plural catalog JSON copies `promote.sheet.recordCount`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ConnectionState` becomes `case files`, `case waits(endID: String?, why: WaitReason)`, `case selfLink(ref: String)`, with *(new)* `enum WaitReason: Equatable, Sendable { case skipped, unset, notOnPage }`. `.endSkipped` is deleted. In `connections`: an end waits when `resolve` returns nil. The named end is end A when A waits, else end B. `why` is `.notOnPage` when the end id is nil, has no row, or is not in `visibleRows`; `.skipped` for target `.skip`; `.unset` for `.unset`. `selfLink(ref:)` takes `row(endA)?.shownRef ?? ""`. `Bridge.files` stays `state == .files` | The subrow names the end and VoiceOver says why (CS-2, CS-8); the self-link callout names the record. One enum keeps the state a function of the ends | Separate `waitingEnd` fields on `Bridge` — two sources of truth for one state |
| D-2 | Host by rule, from `BridgeFact.kind` and A/B order: participation → end B (event); location → end A (event); place relationship → end A (from; the part for `part_of`); relationship → end A when `symmetric`, else end B. A **working row** is a row in `visibleRows` with `anchor == false`. When the rule's end is a working row, it hosts. Else, when the other end is a working row, that end hosts (CS-4). Else the connection has no host and is not shown. The non-host end, when it is a working row and differs from the host, gets a pointer to the connection | S11-D1 §2 "One host per connection", "An end that is not a row", §8 frame 18 ("A connection with neither end a row on the page is not shown"). Anchors render in the "already filed" block with no subrows | Mirroring under both ends — every connection twice (idea §Open questions 1) |
| D-3 | `BridgeFact` gains `var kind: EvidenceBridgeKind = .relationship` and `var symmetric = false`. `PromoteModel.bridgeFacts` sets `kind: bridge.kind` and `symmetric: EvidenceBridgeEdgeSummary.isSymmetric(bridge)`. *(new)* `static func isSymmetric(_ placed: SourceGraphPlacedBridge) -> Bool` in `EvidenceBridgeEdgeSummary`: true only when `placed.kind == .relationship`, the term is a product term (`termKey(kind:in:)`), and `SeededRelationshipTypeTerm.inverse[key] == key` | The facts already carry everything else a subrow needs; the inverse map is the one place symmetry is known | Inferring the kind from `mark` — a display value; a researcher's term with an inverse — the catalog has no inverse for those (AC-1.6) |
| D-4 | A subrow words its connection with `Bridge.sentence` (today's `EvidenceBridgeEdgeSummary.sentence`), one line, truncating at the tail. `phrase`, `endAName` and `endBName` stay on `Bridge` (the evidence sheet and `reasonText` do not use them; they stay because S11-26 left them and removing them is Out) | User, 2026-10-23: reuse the localized sentence for every kind and event ("John Smith was born") instead of the board's "A · term · B", which would need a new lowercase phrase per kind | The board's three parts — new copy per kind; the panel's `endAName phrase endBName` — reads "Mary Is the parent of John" |
| D-5 | When the non-host end is not a working row, the subrow shows a second line, caption, `PVColor.textMuted`, *(new)* `L10n.Promote.connectionOtherEnd(name:filed:)`: "Birth is already filed" for an anchor end, "Birth is not on this page" otherwise. The name is the end's row `name`, else the bridge's `endAName` / `endBName`, else no note. `HostedConnection.otherEnd` *(new)* holds `OtherEnd(name: String, filed: Bool)?` | With a whole sentence (D-4) the board's ", already filed" suffix after a name cannot be spliced in; a second line keeps the sentence whole and translatable | Appending ", already filed" to the sentence — breaks word order in other languages |
| D-6 | *(new)* `struct ConnectionHosts: Equatable, Sendable { var hosted: [String: [HostedConnection]]; var pointers: [String: [Pointer]] }`, *(new)* `struct HostedConnection: Equatable, Sendable, Identifiable { var bridge: Bridge; var otherEnd: OtherEnd?; var id: String { bridge.id } }`, *(new)* `struct Pointer: Equatable, Sendable { var bridgeID: String; var hostID: String }`, and *(new)* `var connectionHosts: ConnectionHosts` on `PromoteFlow`, built from `connections` in one pass. `hosted` is keyed by host subject id; each list is sorted by state rank (files 0, waits 1, selfLink 2), then `sentence` with `localizedStandardCompare`, then `id`. `pointers` is keyed by the pointing row's subject id, in `connections` order. The view reads `model.flow.connectionHosts` once per `body` and passes slices down. This replaces the map's working name `connections(hostedBy:)` | One pass, bound once per render (S11-26 ship log PRF-1: the panel header built `connections` three times); a per-row function would walk every bridge once per row | `connections(hostedBy:)` per row — O(rows × bridges) per render, and the pointers need the same walk again |
| D-7 | `connectionCount` is unchanged: every connection that files, shown or not. The footer and the toast use it | User, 2026-10-23: the footer and the toast say what Done files (AC-1.9) | Counting only shown subrows — would undercount what Done files |
| D-8 | *(new)* `var unsetCount: Int { scope.filter { !$0.anchor && $0.target == .unset }.count }` beside `skipCount` | Board §8 frame 19 adds "N unselected" (user, 2026-10-23: adopt the board) | — |
| D-9 | Footer text is `L10n.Promote.doneSummary(file:skip:unselected:connections:)` *(rewritten)*: the parts `promote.done.file` "File %lld", `promote.done.skip` "Skip %lld" (only when skip > 0), `promote.done.unselected` "%lld unselected" (only when unselected > 0), `promote.done.connections` plural one "%lld connection" / other "%lld connections" (always), joined with `" · "` in Swift. Key `promote.done.summary` is deleted | Each part is one translatable phrase; the separator is a glyph, as in `CitationComposerEvidenceColumn`; leaving out zero parts would need 4 whole keys otherwise | Four whole-sentence keys — four plurals to keep in step |
| D-10 | Expansion and highlight are `PromoteView` `@State`: `expandedHosts: Set<String>` (host subject ids) and `highlighted: Set<String>` (bridge ids), with `highlightGeneration: Int`. Not in `PromoteFlow` | View-only state like scroll position; `PromoteFlow` holds what Done and the proposal read. `bridgesOpen` was the one exception and goes | Flow state — would make `PromoteFlow` Equatable churn on every highlight tick |
| D-11 | The cap is `PromoteLayout.subrowCap = 3` *(new)*. A host with more than 3 hosted connections shows the first 3 and a ghost `.sm` `PVButton(L10n.Promote.moreConnections(count:))` "+N more" (N = hidden count); expanded, it shows all and `PVButton(L10n.Promote.fewerConnections)` "Show fewer" | CS-5 | A disclosure triangle — S11-D1 §6 non-goals |
| D-12 | Pressing a pointer: let `ids = pointers[row].map(\.bridgeID)`; the target host is the first host id, in page order (the order `sections` lists rows), among `pointers[row].map(\.hostID)`. If any of `ids` sits past index 2 in its host's list, insert that host in `expandedHosts`. Then `proxy.scrollTo(hostID, anchor: .top)` inside `withAnimation(.default)` unless `accessibilityReduceMotion`, set `highlighted = Set(ids)`, bump `highlightGeneration`, and after `Task.sleep(for: .milliseconds(1600))` clear `highlighted` only if the generation is unchanged | CS-3, §8 frame 17 (1.6 s); the generation stops an earlier timer from clearing a later highlight | A kit highlight modifier — one call site (S11-D1 §6) |
| D-13 | Rows get `.id(row.subjectID)` and the page `ScrollView` is wrapped in a `ScrollViewReader` | `scrollTo` needs ids; `PVList` uses the same pattern | — |
| D-14 | Subrow identifiers: `promote.connection.<bridgeID>` (kept from the panel row), pointer `promote.row.<ref>.connections`, expander `promote.row.<ref>.connectionsMore` | Stable ids for later UI tests, same shape as `promote.row.<ref>.evidence` | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteFlow.WaitReason` *(new)*; `ConnectionState.waits`, `.selfLink(ref:)` *(changed)* | type | `PromoteFlow.ConnectionState` | Extends it in place (D-1) |
| `PromoteFlow.ConnectionHosts`, `HostedConnection`, `OtherEnd`, `Pointer` *(new)*; `connectionHosts` *(new)* | type / var | `PromoteFlow.connections` | Built on it; `connections` stays the one list of states (D-6) |
| `PromoteFlow.unsetCount` *(new)* | var | `skipCount` | A sibling, same shape |
| `BridgeFact.kind`, `.symmetric` *(new)* | field | `BridgeFact.mark` | `mark` is display (D-3) |
| `EvidenceBridgeEdgeSummary.isSymmetric(_:)` *(new)* | func | `SeededRelationshipTypeTerm.inverse`, private `termKey(kind:in:)` | Reads both; searched `isSymmetric`, `symmetric` in `macos/App`: only the inverse map's doc comment |
| `PromoteConnectionSubrows` *(new, private view)* | component (snowflake) | `PromoteView.connectionRow(_:)` (deleted) | S11-D1 §6 "Rethink": the panel row becomes the subrow |
| `PromoteLayout.subrowCap` *(new)* | constant | `PromoteLayout` | Same enum |
| L10n keys in §6 Strings | key | `promote.connections.*` (deleted) | Rewritten under `promote.connection.*` for the subrow |
| `PromoteFlowTests` T-1 – T-6, `PromoteModelTests` T-7, `PromoteEvidenceDisplayTests` T-8 *(new)* | test | `doneSendsDefaultPinsAndConnectionsFollowEnds`, `viaReason`, `countsArePlural` | Reuse the helpers named in §8 |

**Builds on.** `PromoteFlow.connections` and `resolve` (S11-26 D-3); `EvidenceBridgeEdgeSummary.sentence`; `SeededRelationshipTypeTerm.inverse`; kit `PVButton`, `PVCallout`, `PVMark`; the S11-18 row-extension pattern.
**Deletes.** `ConnectionState.endSkipped`; `PromoteFlow.bridgesOpen`; `PromoteModel.setBridgesOpen(_:)`; `PromoteView.connectionsPanel`, `connectionRow(_:)`, `connectionMeta`, the `bridgesOpen` binding and the footer toggle button; `L10n.Promote.connections(count:)`, `connectionState(_:)` (replaced by `connectionStateText` and `connectionStateSpoken`), `connectionsTitle`, `connectionsMeta(filed:unfiled:)`, `connectionsHint`; catalog keys `promote.connections.count`, `.filed`, `.self`, `.skipped`, `.title`, `.meta`, `.hint`, `promote.done.summary`; the §3 doc sentence about the expanding list.
**Next adds.** None named in the map. A later "open on Evidence graph" action would be one button in `PromoteConnectionSubrows`.

### Schema

None — no catalog change.

### Go core

None — Swift only; filing is unchanged since S11-26.

### FFI

None — no wire change.

### macOS

`PromoteFlow` (`macos/App/Features/Promote/PromoteFlow.swift`):

```swift
enum WaitReason: Equatable, Sendable { case skipped, unset, notOnPage }   // new
enum ConnectionState: Equatable, Sendable {
    case files
    case waits(endID: String?, why: WaitReason)
    case selfLink(ref: String)
}
struct OtherEnd: Equatable, Sendable { var name: String; var filed: Bool }          // new
struct HostedConnection: Equatable, Sendable, Identifiable { var bridge: Bridge; var otherEnd: OtherEnd?; var id: String { bridge.id } }  // new
struct Pointer: Equatable, Sendable { var bridgeID: String; var hostID: String }    // new
struct ConnectionHosts: Equatable, Sendable {                                        // new
    var hosted: [String: [HostedConnection]] = [:]
    var pointers: [String: [Pointer]] = [:]
}
var connectionHosts: ConnectionHosts   // new; D-2, D-6
var unsetCount: Int                     // new; D-8
```

`Bridge` gains `kind` and `symmetric` copied by `lined(_:_:)` from `BridgeFact` (D-3), so `connectionHosts` reads only `connections`. The rule (D-2) is a `private static func ruleHost(_ bridge: Bridge) -> String?`. The waiting end's display name for the state text comes from `row(endID)?.name`, else `endAName` / `endBName` by which end it is; the view asks the model for it (below).

`PromoteModel` (`macos/App/Features/Promote/PromoteModel.swift`): `bridgeFacts` fills `kind` and `symmetric` (D-3). *(new)* `func connectionStateText(_ bridge: PromoteFlow.Bridge) -> String?` and `func connectionStateSpoken(_ bridge: PromoteFlow.Bridge) -> String`, which resolve the waiting end's name and call the L10n functions in §6 Strings (nil text for `.selfLink`, whose copy is the callout). `setBridgesOpen` is deleted.

View (`macos/App/Features/Promote/PromoteView.swift`), from the S11-D1 §6 inventory:

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Promote row | Snowflake, **Extend** | `PromoteAlignmentRow` | Gains `hosted: [PromoteFlow.HostedConnection]`, `pointerCount: Int`, `expanded: Bool`, `highlighted: Set<String>`, and closures `onPointer`, `onToggleExpanded`. After the duplicate callout: the pointer button, then `PromoteConnectionSubrows` |
| Connection subrow list | Snowflake, **Rethink** of `connectionRow` | `PromoteConnectionSubrows` *(new, private struct in PromoteView.swift)* | Leading 1 pt `PVColor.borderSubtle` rule at the note indent, content padded `PVSpacing.space4` after it. Each subrow: `HStack(spacing: PVSpacing.space3)`: `PVMark(bridge.mark, size: 14, decorative: true)` in `textMuted`; `Text(verbatim: bridge.sentence)` `PVFont.body(size: PVTypeScale.bodySmall)`, `.lineLimit(1)`, `textPrimary` when files, `textMuted` when waits or self-link; `Spacer(minLength: PVSpacing.space4)`; state text caption italic (`textSecondary` files, `textMuted` waits; none for self-link). Under it, the D-5 note, then for a self-link `PVCallout(tone: .warning, message: L10n.Promote.connectionSelfLink(ref:), compact: true)`. Vertical padding `PVSpacing.space1` per subrow. Highlighted subrows get `.background(PVColor.surfaceSelected)` with `PVRadius.sm` |
| Self-link warning | Component, Ship | `PVCallout` (`compact: true`, `.warning`) | As the row's conflict and duplicate callouts |
| "+N connections", "+N more", "Show fewer" | Component, Ship | `PVButton(_, variant: .ghost, size: .sm)` | The pointer sits on its own line at `PromoteLayout.noteIndent` (§8 frame 17) |
| Subrow highlight | Snowflake, Extend | `PromoteView` state (D-10, D-12) | `PVColor.surfaceSelected`; no new token or animation |
| Connections panel, footer toggle | **Remove** | `PromoteView` | STEP-5 |
| Footer summary | Snowflake, Ship | `PromoteView.footer` | Text from D-9 |

Frames in words (for an implementer without the board):

- **Frame 16, event row with subrows.** A Birth row; under its callouts, indented to the name column, a thin vertical rule and three subrows: two participation sentences ("John Smith was born", "Mary Smith was the mother at the birth of John Smith") reading "Files on Done" at the right in secondary italic, one muted reading "Waits for Ann Ames"; under them a small ghost button "+1 more".
- **Frame 17, the other end.** John's row; on its own line under it, at the name indent, a small ghost button "+2 connections". After pressing, the page has scrolled so Birth's row is at the top and two of its subrows have a pale selected fill that fades out after about 1.6 s.
- **Frame 18, edges.** A subrow whose second line reads "Harold Gates is already filed" in muted caption (this plan's D-5 wording); a subrow with no state text and a compact warning callout under it, "Not filed · both ends are PER-7QH3D".
- **Frame 19, footer.** Left: nothing (the connections button is gone). Right: "File 12 · Skip 3 · 22 connections" in mono caption, then Cancel and Done.

States: a page with no unfiled connections shows no subrows and no pointers; the footer ends in "0 connections" (AC-5.3). A row whose hosted list is empty draws no rule. Large: 100 rows and a few hundred bridges compute `connectionHosts` once per render (D-6). Accessibility: each subrow is `.accessibilityElement(children: .ignore)` with `.accessibilityLabel(Text(verbatim: model.connectionStateSpoken(bridge)))` (CS-8, which leads with the sentence); the list is `.accessibilityElement(children: .contain)` with `.accessibilityLabel(Text(verbatim: L10n.Promote.connectionsHostedBy(name: row.name)))`; the pointer's label is `L10n.Promote.connectionsElsewhereSpoken(count:)`. The highlight is visual only.

### Strings

L10n keys *(new)* in `L10n.Promote`, beside today's connection keys, each with a catalog entry (`extractionState: manual`, English only):

| Function | Key | English |
| --- | --- | --- |
| `connectionStateText(_ state: …, endName: String) -> String?` | `promote.connection.files` | "Files on Done" |
| (same) | `promote.connection.waits` | "Waits for %@" |
| (same, `endName` empty) | `promote.connection.waitsEnd` | "Waits for its other end" |
| `connectionStateSpoken(sentence:state:endName:) -> String` | `promote.connection.spoken.files` | "%@. Files on Done." |
| (same) | `promote.connection.spoken.skipped` | "%1$@. Waits for %2$@, who is skipped." |
| (same) | `promote.connection.spoken.unset` | "%1$@. Waits for %2$@, who has no target yet." |
| (same) | `promote.connection.spoken.notOnPage` | "%1$@. Waits for %2$@, who is not on this page." |
| (same, `endName` empty) | `promote.connection.spoken.waitsEnd` | "%@. Waits for its other end." |
| (same) | `promote.connection.spoken.selfLink` | "%1$@. Not filed, both ends are %2$@." |
| `connectionSelfLink(ref:)` | `promote.connection.selfLink` | "Not filed · both ends are %@" |
| `connectionOtherEnd(name:filed:)` | `promote.connection.otherEnd.filed` / `.notOnPage` | "%@ is already filed" / "%@ is not on this page" |
| `connectionsHostedBy(name:)` | `promote.connection.hostedBy` | "Connections hosted by %@" |
| `connectionsElsewhere(count:)` | `promote.connection.elsewhere` (plural) | one "+%lld connection" / other "+%lld connections" |
| `connectionsElsewhereSpoken(count:)` | `promote.connection.elsewhereSpoken` (plural) | one "%lld connection hosted on another row" / other "%lld connections hosted on other rows" |
| `moreConnections(count:)` | `promote.connection.more` | "+%lld more" (no noun: no plural) |
| `fewerConnections` (static resource) | `promote.connection.fewer` | "Show fewer" |
| `doneSummary(file:skip:unselected:connections:)` | `promote.done.file`, `promote.done.skip`, `promote.done.unselected`, `promote.done.connections` (plural) | D-9 |

Removed: the keys and functions listed under **Deletes**. Comments say where each string shows and what each argument is.

## 7. Steps

The implementer evidences every step with XT and `scripts/check-pr.sh`. §10's manual checks need a person at the app; the implementer lists them as "for the user" and does not tick them.

Xcode check, used below as **XT**: `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/<Suite> CODE_SIGNING_ALLOWED=NO` (from `scripts/check-pr.sh`).

**STEP-1 — States, hosts and the unset count in the flow**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-6, D-8, and the `BridgeFact` / `Bridge` fields of D-3 (defaults keep every existing `BridgeFact(…)` literal compiling). In `doneSendsDefaultPinsAndConnectionsFollowEnds`, `.endSkipped` becomes `.waits(endID: "b", why: .skipped)`. `PromoteView` still compiles: its `connectionRow` reads `L10n.Promote.connectionState(bridge.state)`; update that function's switch to the new cases with today's three strings (`.waits` → `promote.connections.skipped`, `.selfLink` → `promote.connections.self`) so the tree builds; STEP-3 replaces it. Add T-1 – T-6.
Check: XT with `PromoteFlowTests` passes.

**STEP-2 — Kind and symmetry from the graph**
Files: `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`.
Do: D-3's `isSymmetric(_:)` and `bridgeFacts` fields. Add T-7, built on a helper shaped like `viaReason(kind:edgeKeys:term:)` that returns `model.flow.bridges.first`.
Check: XT with `PromoteModelTests` passes.

**STEP-3 — Strings**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`. Skill: `add-localized-string`.
Do: every key in §6 Strings; `PromoteModel.connectionStateText(_:)` and `connectionStateSpoken(_:)`. Delete `connectionState(_:)` and point `connectionRow` at `model.connectionStateText(bridge) ?? ""` for this step only (STEP-5 deletes `connectionRow`). Keep the panel and footer keys until STEP-5. Add T-8.
Check: XT with `PromoteEvidenceDisplayTests` and `PromoteModelTests` passes; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-4 — Subrows, pointer, expand and highlight**
Files: `macos/App/Features/Promote/PromoteView.swift`. Skill: `add-ui-component`.
Do: D-10 – D-14 and §6 macOS: `ScrollViewReader`, `.id(row.subjectID)`, bind `let hosts = model.flow.connectionHosts` once in `body` and pass `hosts` into `sectionBlock`; `PromoteAlignmentRow` gains the inputs in §6; add `PromoteConnectionSubrows`. The panel still exists in this step.
Check: XT with `PromoteFlowTests` and `PromoteModelTests` passes (the app builds); §10 steps 1 – 4 by hand if a Mac is at hand, else listed for the user.

**STEP-5 — Remove the panel; footer**
Files: `macos/App/Features/Promote/PromoteView.swift`, `PromoteModel.swift`, `PromoteFlow.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: delete everything under **Deletes** that is still present; footer per D-9 with `unsetCount`.
Check: XT with `PromoteFlowTests`, `PromoteModelTests`, `PromoteEvidenceDisplayTests` passes; `rg -n "bridgesOpen|setBridgesOpen|connectionsPanel|connectionRow|connectionMeta|connectionsHint|connectionsTitle|connectionsMeta|endSkipped|promote\.connections\.|promote\.done\.summary" macos` prints nothing; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-6 — Docs**
Files: `docs/promote-graph-alignment.md` §3, `docs/deployment-plan/spike-11/completed.md` (a `## S11-27 — Connection subrows on Promote` section above S11-26's, in its shape), `deployment-plan.md` (map row), this plan (status), `design/S11-D1-promote-rows.md` → `design/archive/`, `design/README.md`.
Do: in §3 replace the bullet "**Bridges aren't rows** (participation, relationship, location). A summary line reads "22 connections will be filed". It expands to the list of connections, each saying whether it files. Filing rules: §9.1." with "**Bridges aren't rows** (participation, location, relationship, place relationship). Each shows as a subrow under one row it joins: a participation or location under its event, a place relationship under its from place, a directed relationship under the end named last and a symmetric one under the end named first; when that end is not a row on the page, under the other end. A subrow says whether it files on Done, waits for a named end, or has both ends on one record. The other end shows how many of its connections are hosted elsewhere and scrolls to them. The footer counts the connections that file. Filing rules: §9.1." The S11-D1 brief moves to `design/archive/`; `design/README.md` moves its row to **Archived** ("S11-D1 — Promote page rows; board agreed 2026-10-20, built by S11-18 (#507) and S11-27 (#N)"). The rest per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-27-connection-subrows.md` reports no errors; links resolve (`rg -n "S11-D1-promote-rows.md" docs` points only at `design/archive/`).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Swift helpers in `PromoteFlowTests`: `proposal(rows:)`, `row(id:target:handleID:assessment:)`, `fact(_:anchor:handleID:)`. Facts default to kind `.person`; a test that needs another kind sets `kind` on the fact after the call. `BridgeFact(id:sentence:endA:endB:)` with `kind:` and `symmetric:` set explicitly.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `connectionsAreHostedByRule` *(new)* | `mappedRest = true`; rows j, b, t, o, m, n, a, c, all handles on distinct ids. Bridges: participation `endA: "j", endB: "b"`; location `endA: "b", endB: "t"`; place relationship `endA: "t", endB: "o"`; relationship `m`→`n` `symmetric: false`; relationship `a`→`c` `symmetric: true`. → `connectionHosts.hosted` keys: participation under `b`, location under `b`, place under `t`, directed under `n`, symmetric under `a`. `pointers["j"] == [Pointer(bridgeID: <participation>, hostID: "b")]`, `pointers["m"]` names the directed bridge, `pointers["c"]` the symmetric one | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, D-2 |
| T-2 | Swift | same file · `aConnectionFallsBackToTheEndThatIsARow` *(new)* | (a) `mappedRest = true`; row j (handle), b loaded with `fact("b", anchor: true)`; participation j→b → hosted under `j`, `otherEnd == OtherEnd(name: "b", filed: true)`, state `.files`, no pointer. (b) Not mapped, entry `j`; rows j, b; participation j→b → hosted under `j`, state `.waits(endID: "b", why: .notOnPage)`, `otherEnd == OtherEnd(name: "b", filed: false)`. (c) Both ends anchors (handles e-a, e-b): no host, `connectionCount == 1`. (d) A bridge with `alreadyFiled: true` between two rows: in no `hosted` list, not in `connections`, not counted | AC-1.7, AC-1.8, AC-1.9, AC-1.10, D-5, D-7 |
| T-3 | Swift | same file · `waitingNamesTheEndAndWhy` *(new)* | `mappedRest = true`; rows a, b (handles e1, e2); relationship a→b. `setTarget("b", "skip")` → `.waits(endID: "b", why: .skipped)`; `setTarget("a", "skip")` → `.waits(endID: "a", why: .skipped)`; reload with row b drafted unset (`row(id: "b", target: "skip", handleID: "", assessment: "none")`) → `.waits(endID: "b", why: .unset)`; bridge with `endB: nil` → `.waits(endID: nil, why: .notOnPage)` | AC-2.2, AC-2.4, AC-2.5, D-1 |
| T-4 | Swift | same file · `bothEndsOnOneRecordNameTheRecord` *(new)* | rows a, b both `handleID: "e1"` (helper ref `PER-1`); relationship a→b → `.selfLink(ref: "PER-1")`, `connectionCount == 0` | AC-2.6, D-1 |
| T-5 | Swift | same file · `hostedConnectionsAreOrderedByStateThenSentence` *(new)* | Event row `b`; person rows p1 – p4; participations p1→b "d", p2→b "c", p3→b "b", p4→b "a"; then `setTarget("p2", "skip")` and `setTarget("p3", "skip")`; a fifth row p5 on `handleID: "e-b"` (b's handle) with participation p5→b "e". → `hosted["b"].map(\.bridge.sentence) == ["a", "d", "b", "c", "e"]` (files by sentence, then waits by sentence, then the self-link) | AC-3.1, D-6 |
| T-6 | Swift | same file · `theUnsetCountCountsRowsWithNoTarget` *(new)* | rows a (handle), b drafted unset, c anchor → `unsetCount == 1`; `setTarget("b", "new")` → 0 | AC-5.2, D-8 |
| T-7 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `bridgeFactsCarryKindAndSymmetry` *(new)* | Through `FakeStore`, as `viaReason`: relationship with `spouse` (provenencia) → `kind == .relationship`, `symmetric == true`; `parent` → `symmetric == false`; a researcher's term key `spouse` with origin `researcher` → `false`; participation with role `subject` → `kind == .participation`, `symmetric == false` | AC-1.4, AC-1.5, AC-1.6, D-3 |
| T-8 | Swift | `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift` `connectionCopyIsWorded` *(new)* | `doneSummary(file: 12, skip: 3, unselected: 0, connections: 22) == "File 12 · Skip 3 · 22 connections"`; `(12, 0, 2, 1) == "File 12 · 2 unselected · 1 connection"`; `(5, 0, 0, 0) == "File 5 · 0 connections"`; `connectionsElsewhere(count: 1) == "+1 connection"`, `(2) == "+2 connections"`; `connectionsElsewhereSpoken(count: 2) == "2 connections hosted on other rows"`; `connectionStateSpoken(sentence: "Mary Smith is the parent of John Smith", state: .waits(endID: "j", why: .skipped), endName: "John Smith") == "Mary Smith is the parent of John Smith. Waits for John Smith, who is skipped."` | AC-4.1, AC-4.2, AC-4.5, AC-5.1, AC-5.2, AC-5.3, D-9 |

AC-2.1 and AC-2.3 are covered by the rewritten `doneSendsDefaultPinsAndConnectionsFollowEnds`. AC-3.2 – AC-3.4 and the subrow look are checked in §10.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests` (the state cases), `L10nPluralTests` (the new plural keys), `scripts/check-localizable-xcstrings.py` (removed keys).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (STEP-5 rg prints nothing)
- [ ] UI matches the brief inventory; no new kit primitive
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] S11-D1 moved to design/archive/ and design/README.md updated
- [ ] docs/promote-graph-alignment.md §3 "Bridges aren't rows" rewritten (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a Source with: persons Mary Smith, John Smith and Ann Ames; a Birth event of John with participations John (subject), Mary (mother), Ann (witness) and one more person; the Birth located at Toronto; Toronto part of Ontario, with Ontario already filed; and a relationship "Mary Smith is the parent of John Smith".

1. Promote all. Birth's row shows three participations and "+1 more"; the location is under Birth; "Toronto is part of Ontario" is under Toronto with "Ontario is already filed"; the parent relationship is under John. Press "+1 more": four show; "Show fewer" collapses.
2. Mary's row shows "+N connections". Press it: the page scrolls to the host row and its subrows that join Mary are highlighted briefly (AC-3.2, AC-3.3). Collapse Birth first and press it again: Birth expands before the highlight.
3. Set John to Skip: his subrows and the participation under Birth read "Waits for John Smith" and dim; the footer's connections count drops. Set John to New: they read "Files on Done".
4. Set Mary and John to the same record: the relationship shows "Not filed · both ends are" and that record's ref, as a warning callout (AC-2.6).
5. Group by assessment: every subrow is still under its host (AC-3.4). The footer reads as AC-5.1 with no button; there is no bottom panel.
6. With VoiceOver, focus a filing subrow (AC-4.1), a waiting one with John skipped (AC-4.2), the self-link from step 4 (AC-4.3), the subrow list (AC-4.4) and the pointer (AC-4.5): each reads as its AC says.
7. Open Promote on John alone: the participation is under John and reads "Waits for Birth" with "Birth is not on this page".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A connection's ends are not in the A/B order `ConnectionSentenceRegistry.Slots` documents for some kind | T-1 pins the order; T-7 builds facts from real edge keys |
| `scrollTo` lands on a row inside a lazily built section | Rows sit in a plain `VStack` inside the `ScrollView` (no `LazyVStack`), so every id exists |
| The footer joins parts in an order another language would change | Accepted; the separator is a glyph, and each part is whole (D-9) |

**Stop and report if:**

- `PromoteFlow.connections` or `resolve(_:)` no longer does what §4 says.
- A seeded `place_relationship_type` other than `part_of` and `succeeded_by` exists, or `part_of` stores the part at end B (D-2's place rule would host under the whole).
- `ConnectionSentenceRegistry.Slots` documents a different A/B order than §4.
- Any AC needs a Go or wire change.
- A frame of the board, if you can open it, contradicts an AC other than D-4 and D-5's wording.

## 12. Open questions

None.

## 13. PR

**Title:** S11-27 — Connection subrows on Promote
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-27-connection-subrows.md`.
