# S11-42 — Implied links on the evidence sheet

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: one muted caption line inside the existing relationship Card header, in the style its record count already uses; user, 2026-10-27)
**Board:** —
**Depends on:** S11-40 (landed [#579](https://github.com/mendahu/provenencia/pull/579))
**Blocks:** —
**Branch:** `s11-42-implied-links-sheet`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Environment:** every step needs a Mac with the repo's toolchain (`xcodebuild test`, `scripts/check-pr.sh`) and the app for §10. No Go, proto or FFI change. A session without a Mac does not start: it reports "needs a Mac" to the orchestrator, which hands the PR to a Mac session.
**Context:** [Kinship shapes § On the evidence sheet](../kinship-shapes.md#on-the-evidence-sheet); [`claim-evidence.md` § Wording](../claim-evidence.md#wording) and [§ Evidence through relationships](../claim-evidence.md#evidence-through-relationships); S11-40 plan D-9, D-11, D-13 and its Landed note; working rules "One wire vocabulary", "Counted once across sides"
**Planned:** 2026-10-27 against `main` @ `c545a482`
**Review:** [`S11-42-implied-links-sheet.review.md`](S11-42-implied-links-sheet.review.md) — Fits with changes / Ready with changes, 1 round (every finding answered)

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

The evidence sheet's relationship Card says how each side recorded a fact when either side recorded it through a kinship pattern. Under "…they share a spouse (refs)" a muted line reads "This source: married in the marriage on 12 May 1842 · The record: recorded as spouses". S11-40 sends Swift how each side recorded the fact (`CatalogClaimEvidenceRelationship.layer` / `.canon`) and the tree pivot's header, but the sheet still words the Card as if both sides recorded the link directly. This PR adds the caption line, its VoiceOver reading, and its L10n.

**Why now.** S11-40 landed the engine and the wire; its ship log left UX-1 ("an implied tree link's Card does not say how the tree recorded it") to this PR. This is the last PR of slice 9, and the slice Check needs the sheet to say it.

**What stays.** The Card header ("…they share a spouse (refs)"), its pin box, badge, record count, groups and pins. Every property and likeness line. A Card whose two sides both record the link directly looks exactly as on `main`. Go, proto, FFI and `FakeStore` are not touched.

## 2. User stories

Fixtures: "spouse line" is a `CatalogClaimEvidenceRelationship` with `key` "relationship·spouse", `bridgeType` "relationship", `neighborTerm` the product `spouse` term, `neighborTermNamesNeighbor` true, `outcome` `.agree`. "M" is a page row of kind `.event`, `name` "Marriage of Ann Ames and Bob Ames", `ref` "EVT-M", `date` "12 May 1842". "Mʹ" is a `CatalogEventHeader` whose ref is "EVT-T" and whose `date` is 3 June 1840 (shown by `DateRowDisplay` as "3 Jun 1840"; the AC compares with `DateRowDisplay.line`'s output, not a literal). "Bea" is a `CatalogPersonHeader` named Bea Ames, ref "PER-B".

**US-1.** As a researcher, when the Source or the tree records a relationship through a marriage, a birth or a parent in between, the relationship Card says how each side recorded it, so that I can see why the two records count as one fact before I pin it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Spouse line, `layer` `{marriageSpouses, pivotID "M"}`, `canon` direct; row M on the page | Card `recorded` = "This source: married in the marriage on 12 May 1842 · The record: recorded as spouses" |
| AC-1.2 | Spouse line, `layer` direct, `canon` `{marriageSpouses, pivot .event(Mʹ)}` | "This source: recorded as spouses · The record: married in the marriage on <Mʹ's date line>" |
| AC-1.3 | Both sides `marriageSpouses` (layer M, canon Mʹ) | "This source: married in the marriage on 12 May 1842 · The record: married in the marriage on <Mʹ's date line>" |
| AC-1.4 | Both sides direct (today's case) | `recorded` is nil; the Card draws no caption; `spoken` equals `main`'s |
| AC-1.5 | `layer` `{marriageSpouses, pivotID "M"}`, M has `date` "" | Source side reads "married in the marriage EVT-M" |
| AC-1.6 | `layer` `{marriageSpouses, pivotID "X"}`, no page row X | Source side reads "married in a marriage" |
| AC-1.7 | Parent line (`key` "relationship·parent"), `canon` `{birthParents, .event(birth header dated 1 Mar 1850)}`, `layer` direct | "This source: recorded as parent and child · The record: parent and child in the birth on <date line>" |
| AC-1.8 | Grandparent line, `canon` `{parentOfParent, .person(Bea)}`, `layer` direct | "This source: recorded as grandparent and grandchild · The record: grandparent and grandchild through Bea Ames" |
| AC-1.9 | `canon` `{parentOfParent, pivot nil}` | Record side reads "grandparent and grandchild through a parent" |
| AC-1.10 | Sibling line (`neighborTerm` the product `sibling` term), `layer` `{unrecognized("sibling_parents"), pivotID "M"}`, `canon` direct | "This source: through Marriage of Ann Ames and Bob Ames · The record: recorded directly" |
| AC-1.11 | AC-1.1, the Card's own neighbor not on the page (header "they share a spouse") | Card `spoken` = `spokenRelationship` of "they share a spouse. This source: married in the marriage on 12 May 1842. The record: recorded as spouses" and the outcome; the caption is hidden from VoiceOver |
| AC-1.12 | The app on S11-40's manual-check data (tree marriage; record spouse relationship) | Bob's sheet shows the Card "…they share a spouse (…)" with the caption "This source: recorded as spouses · The record: married in the marriage on 12 May 1842"; checking it and pressing Done pins both tree participations |

**US-2.** As the page, a Source event row knows its date as the row lists it, so that the sheet can name a Source marriage or birth by its date.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A placed `.event` Subject with a positive `date` Observation whose `date` is 12 May 1842 | `EvidenceGraphPlacedSubject.dateLine` = `DateRowDisplay.line(date: that date)` |
| AC-2.2 | An event with only `start_date` and `end_date` Observations | `dateLine` = `DateRowDisplay.line(date: nil, start:, end:)` of them |
| AC-2.3 | An event whose only `date` Observation is negative; a person with a `date` Observation | `dateLine` = "" for both |
| AC-2.4 | `PromoteFlow.load` with a `SubjectFact` whose `date` is "12 May 1842" | The row's `date` is "12 May 1842", for an anchor row and for a proposed row |

## 3. Scope

| In | Covered by |
| --- | --- |
| A Source event's date on its page row: `EvidenceGraphPlacedSubject.dateLine`, `SubjectFact.date`, `Row.date` | STEP-1, AC-2.1 – AC-2.4 |
| One display of any `CatalogConclusionHeader` (title, secondary line); Promote's menu and search hits move onto it | STEP-2, T-5 |
| The two side phrases and their L10n, as one pure function | STEP-3, AC-1.1 – AC-1.10 |
| The board fills `RelationshipCard.recorded` from the shapes, the page row of a Source pivot and the tree pivot's header; the spoken label reads it | STEP-4, AC-1.1, AC-1.2, AC-1.4, AC-1.11 |
| The Card draws the caption | STEP-5, AC-1.4, AC-1.12 |
| Docs: `claim-evidence.md`, `kinship-shapes.md`, `add-match-rule` skill, `completed.md`, map | STEP-6 |

| Out | Home |
| --- | --- |
| Engine, proto, FFI, `GoStore` mapping, `FakeStore` | not planned (S11-40 landed them) |
| A related person's lines comparing with every member | S11-52 |
| Changing the header sentence ("they share a spouse") or its refs | not planned |
| A Source pivot that is not a page row named by anything but the no-pivot phrase (a second read of the Source graph) | not planned |
| Linking the pivot name to its row or record | not planned |
| Trees and person pages wording implied links | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md), S13-04 |

## 4. Current state

Verified at `c545a482`.

| Area | Where | Today |
| --- | --- | --- |
| Shapes on the wire | `macos/App/Platform/GenealogyStore.swift` `CatalogClaimEvidenceShape {pattern, pivotID, pivot: CatalogConclusionHeader?}`; `CatalogClaimEvidenceRelationship.layer`, `.canon` | Filled by `GoStore.mapEvidenceShape`; read by nothing in `Features/` |
| Pattern enum | `macos/App/Platform/ConclusionVocabulary.swift` `KinshipPattern` | `direct`, `marriageSpouses`, `birthParents`, `parentOfParent`, `unrecognized(String)` |
| Board | `macos/App/Features/Promote/PromoteEvidenceBoard.swift` `EvidenceBoard.RelationshipCard`, `EvidenceBoard.build(evidence:neighbor:labels:worded:)`, local `relationshipCard` | Card has `header`, `pinIDs`, `outcome`, `spoken`, `groups`; `spoken` from `PromoteEvidenceDisplay.spokenRelationship(header:outcome:count:)`. `build` has no name order |
| Card view | `macos/App/Features/Promote/PromoteEvidenceCards.swift` `EvidenceRelationshipCard.header` | `EvidencePinBox`, then an `HStack` of the header text and the `recordCount` caption (`PVFont.body(size: PVTypeScale.caption)`, `PVColor.textMuted`), `accessibilityHidden(true)`, then `EvidenceOutcomeBadge` |
| Display | `macos/App/Features/Promote/PromoteEvidenceDisplay.swift` `PromoteEvidenceDisplay` | `sharePhrase`, `shareHeader`, `spokenRelationship` — no shape wording |
| Board caller | `macos/App/Features/Promote/PromoteModel.swift` `board(of:)` | Calls `EvidenceBoard.build` with `neighbor: { flow.row(id) }`; `session.nameFormat` is the name order (`syncNameOrders`) |
| Page rows | `macos/App/Features/Promote/PromoteFlow.swift` `SubjectFact {id, kind, name, ref, anchor}`, `Row`, `makeRow` | Two `Row(…)` constructions in `makeRow` (anchor, proposed); a decided row keeps `previous`. No date |
| Row facts | `PromoteModel.subjectFacts(_:)` | `name: placed.displayName(orders:)`: an event's title, which never holds its date |
| Placed Subject | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` `EvidenceGraphPlacedSubject` | `observations: [CatalogObservation]`; each Observation carries `propertyKey`, `polarity`, `date: CatalogDateValueInput?` (`GoStore.mapCatalogObservation`'s reverse) |
| Dates | `macos/App/Features/Dates/DateRowDisplay.swift` `DateRowDisplay.line(date:start:end:locale:)` | The Events list's date column |
| Headers | `macos/App/Features/Conclusions/PersonHeaderDisplay.swift` `PersonHeaderDisplay.title`; `EventTitleDisplay.title`; `PlaceTitleDisplay.title`; `EventsListView.swift` `EventSecondaryDisplay.content(_:).date`, `.line(_:)`; `PersonLifeDisplay.line`; `PlaceChainDisplay.line` | The per-kind title and secondary line of a header |
| Header switches | `macos/App/Features/Promote/PromoteFlow.swift` private `conclusionTitle(kind:person:event:place:fallback:orders:style:)`, `conclusionSubtitle(kind:person:event:place:)` (called from `item`); `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` `title(for:orders:)`, `secondary(for:)` | Two copies of one switch over person / event / place calling the per-kind helpers; Promote's takes a kind and three optional headers, search's a `CatalogConclusionHeader?` with the hit's own text when nil |
| Relationship terms | `macos/App/Platform/ConclusionVocabulary.swift` `SeededRelationshipTypeTerm` | `spouse`, `parent`, `child`, `grandparent`, `grandchild`, `sibling`, … ; the Card's `neighborTerm.productKey` is one of them for a product term |
| Keys | `macos/App/Platform/ConclusionVocabulary.swift` `SeededPropertyKey.date`, `.startDate`, `.endDate` | Spelled once |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Promote` (`shareHeader`, `recordCount`, `thisSource`, `spokenRelationship`); `macos/App/Resources/Localizable.xcstrings` | No shape copy |
| Tests | `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift` `build` helper, `twoHundredLines` (loads a `PromoteFlow` with one `SubjectFact`), `aRelationshipCardIsWordedFromTheNeighborsSide`; `PromoteEvidenceDisplayTests.swift` `relationshipShareFromTheEngine`, a point-date helper at line 43; `EvidenceGraphSnapshotTests.swift` private `observation(_:text:nameForm:polarity:termKey:)` (no date); `PromoteFlowTests.swift`; `CatalogSearchTests.swift` | Pin today's header, spoken label, menu titles and hit titles |
| Skill | `.agents/skills/add-match-rule/SKILL.md` "Add a kinship pattern" step 6 | "The claim evidence sheet's wording for it (S11-42) follows." |

**Precedent.** `8aab38bf` ("…a share phrase on each relationship") added the Card's header words as a `PromoteEvidenceDisplay` function tested in `PromoteEvidenceDisplayTests`, with the board filling a `RelationshipCard` field once per load: copy that split (pure words in `PromoteEvidenceDisplay`, assembly in `EvidenceBoard.build`, the view only draws a `String`). The `recordCount` caption in `EvidenceRelationshipCard.header` is the style the new caption copies.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The Card keeps its header. A caption line under the header text shows when either side's `pattern` is not `.direct`; it is nil when both are direct | User, 2026-10-27 (placement); a direct–direct Card stays as `main` | Folding into the header — one side only, long header; a brief — no new component |
| D-2 | The caption names both sides: `L10n.Promote.recorded(source:record:)` *(new)*, key `promote.sheet.recorded`, "This source: %1$@ · The record: %2$@". "This source" is the layer, "The record" the canon | User, 2026-10-27 (both sides); one whole format, as S11-33 requires; "record" is the sheet's word for the matched entity (`shareHeader`'s "connected record") | Only the implied side |
| D-3 | An implied side's phrase comes from its pattern's **wording entry** `PatternWording` *(new, private struct in `PromoteEvidenceDisplay`)* `{pivot: PivotShape; named: (String) -> String; none: LocalizedStringResource}`, with `enum PivotShape { case event, person }` *(new, private)*. `private static func wording(_ pattern: KinshipPattern) -> PatternWording` is one exhaustive `switch`: `.marriageSpouses` → `(.event, recordedMarriage, recordedMarriageNone)`; `.birthParents` → `(.event, recordedBirth, recordedBirthNone)`; `.parentOfParent` → `(.person, recordedChain, recordedChainNone)`; `.unrecognized` → `(.person, recordedOther, recordedOtherNone)`; `.direct` is never passed (D-4). One ladder, `private static func implied(_ w: PatternWording, _ pivot: EvidenceBoard.Pivot?) -> String`: `.event` — date ≠ "" → `w.named(L10n.Promote.recordedOn(date))`, else ref ≠ "" → `w.named(ref)`, else `L10n.string(w.none)`; `.person` — name ≠ "" → `w.named(name)`, else ref ≠ "" → `w.named(ref)`, else `L10n.string(w.none)`. Copy is §6 Strings | User, 2026-10-27 (event by date, person by name); one ladder, so a new pattern is one `switch` arm and two keys; the `switch` stays exhaustive, so a new `KinshipPattern` case does not compile without wording (review M-2) | A dated / ref / none key per pattern — the ladder copied per pattern; splicing an event-type word into a shared template — S11-33 forbids it |
| D-4 | A **direct** side's phrase comes from the link's term, not from the other side's pattern: `private static func direct(term: String?) -> String`, keyed by `relationship.neighborTerm?.productKey`: `SeededRelationshipTypeTerm.spouse` → `recordedDirectSpouses`; `.parent` or `.child` → `recordedDirectParentChild`; `.grandparent` or `.grandchild` → `recordedDirectGrandparent`; anything else or nil → `recordedDirectOther` | The term is the fact both sides record; a later pattern that implies a parent (a baptism's parents, `docs/ideas/kinship-view.md`) reuses "recorded as parent and child" with no new key (review M-2). `neighborTerm` is the Card's existing reading of the link (S11-20) | Keying by the other side's pattern — two patterns implying one link would need two keys |
| D-5 | `PromoteEvidenceDisplay.recorded(layer: KinshipPattern, layerPivot: EvidenceBoard.Pivot?, canon: KinshipPattern, canonPivot: EvidenceBoard.Pivot?, term: String?) -> (source: String, record: String)?` *(new)*: nil when both patterns are `.direct`; otherwise each side is `direct(term:)` when its pattern is `.direct`, else `implied(wording(pattern), pivot)` | Pure and testable without a view, as `sharePhrase` and `shareHeader` are | Wording in the view — untestable |
| D-6 | `EvidenceBoard.Pivot` *(new)* `struct Pivot: Equatable { var name: String; var date: String; var ref: String }`. From a Source pivot: `neighbor(shape.pivotID)` (the page row) → `Pivot(name: row.name, date: row.date, ref: row.ref)`; nil when `pivotID` is "" or no row has it. From the tree pivot: `init?(header: CatalogConclusionHeader?, orders: NameFormatOrders)` *(new)* — nil when `header` is nil; else `name: ConclusionHeaderDisplay.title(header, orders: orders)`, `date: header.event.map { EventSecondaryDisplay.content($0).date } ?? ""`, `ref` the header's entity ref (`ConclusionHeaderDisplay.ref`, D-12). A place header is named like any other. The layer shape's `pivot` header is never read (S11-40 D-11 sends none) | One value both sides fill; the header is titled where every header is (D-12) | Reading the Source pivot's header from Go — S11-40 D-11 decided a layer pivot is found on the page |
| D-7 | `EvidenceBoard.build` gains `orders: NameFormatOrders = NameFormatOrders()` as its last parameter; `PromoteModel.board(of:)` passes `session.nameFormat`. The `neighbor` closure's doc comment becomes "the page row a Subject is (a connected row, or a relationship's Source pivot)" | The tree pivot's person name follows the researcher's name order, as every other name on the sheet does | A second closure for pivots — `neighbor` already answers any page row |
| D-8 | `RelationshipCard` gains `recorded: String?` *(new field, after `header`)*: `L10n.Promote.recorded(source:record:)` of D-5's pair, formatted in `relationshipCard`. When the pair is set, `spoken` is `spokenRelationship(header: L10n.Promote.spokenRecorded(header: header, source:, record:), outcome:, count:)`, where `header` is the Card's `header` as built (refs included) and `spokenRecorded` *(new)* is key `promote.sheet.spokenRecorded`, "%1$@. This source: %2$@. The record: %3$@"; otherwise `spoken` is unchanged | The board words once and the view only draws a `String` (`EvidenceBoard`'s doc comment; review M-3). VoiceOver reads the pin box's `spoken`; a spoken "·" reads badly, so the spoken form is its own whole format | Holding the two parts on the Card and formatting in the view — words on each draw; joining with `L10n.Conclusions.a11yList` — a Conclusions format with another meaning |
| D-9 | `EvidenceRelationshipCard.header`: the `HStack` holding the header text and record count becomes the first child of a `VStack(alignment: .leading, spacing: PVSpacing.space1)`; when `relationship.recorded` is set, the second child is `Text(verbatim: recorded)` in `PVFont.body(size: PVTypeScale.caption)`, `PVColor.textMuted`, `.fixedSize(horizontal: false, vertical: true)`. The `VStack` takes the `.frame(maxWidth: .infinity, alignment: .leading)` and `.accessibilityHidden(true)` the `HStack` has today. The outer `HStack` (pin box, text, badge) keeps `.center` | The record count's caption style; no new kit component (`add-ui-component`: a line inside an existing feature Card) | A `PVSectionHeader` subtitle — the Card header is not a section header |
| D-10 | `EvidenceGraphPlacedSubject.dateLine: String` *(new, computed)*: "" unless `kind == .event`; else `DateRowDisplay.line(date:start:end:)` of the `date` of the first Observation each with `polarity != "negative"`, a non-nil `date`, and `propertyKey` `SeededPropertyKey.date` / `.startDate` / `.endDate` | The Events list's date rule on the Source's own Observations, so a Source marriage and a tree marriage read alike | Reading it in the board from the snapshot — the board has rows, not the snapshot |
| D-11 | `PromoteFlow.SubjectFact` gains `var date: String = ""` and `Row` gains `var date: String = ""`, each the last stored property. `PromoteModel.subjectFacts` sets `date: placed.dateLine`; both `Row(…)` calls in `makeRow` pass `date: fact.date`; a decided row keeps `previous` as today | The row already carries the Subject's name and ref the same way; defaults keep every test construction compiling | A dictionary of dates on `PromoteFlow` — a second lookup beside `row(_:)` |
| D-12 | `ConclusionHeaderDisplay` *(new)* in `macos/App/Features/Conclusions/ConclusionHeaderDisplay.swift`: `static func title(_ header: CatalogConclusionHeader, style: NameDisplayStyle = .natural, orders: NameFormatOrders = NameFormatOrders()) -> String` (person → `PersonHeaderDisplay.title(h, style:, orders:)`, event → `EventTitleDisplay.title(h, orders:)`, place → `PlaceTitleDisplay.title(h)`), `static func secondary(_ header: CatalogConclusionHeader) -> String` (person → `PersonLifeDisplay.line(h).text`, event → `EventSecondaryDisplay.line(h)`, place → `PlaceChainDisplay.line(parents: h.parents, candidates: h.parentsAreCandidates)`), `static func ref(_ header: CatalogConclusionHeader) -> String` (the header's `entity.ref`). `CatalogSearchHitPresentation.title(for:orders:)` becomes `hit.header.map { ConclusionHeaderDisplay.title($0, style: .sorted, orders: orders) } ?? hit.title`, and `secondary(for:)` likewise with `hit.subtitle`. In `PromoteFlow`, `conclusionTitle` and `conclusionSubtitle` are deleted; `item` builds `let header: CatalogConclusionHeader?` from `kind` and the three optionals (`.person` with `person`, `.event` with `event`, `.place` with `place`, else nil) and calls `ConclusionHeaderDisplay.title(header, style: .sorted / .natural, orders:)` / `.secondary(header)`, with `ref` / "" when `header` is nil | `Pivot` would be the third switch over the header kinds; the third copy is extracted in the PR that adds it (review M-1). Same helpers, so no title changes | A private switch in `Pivot` — a third copy |
| D-13 | Test fixtures: `EvidenceGraphSnapshotTests.observation` gains `date: CatalogDateValueInput? = nil` (passed to `CatalogObservation`'s `date`). `PromoteEvidenceBoardTests` gains `private func page(_ facts: [PromoteFlow.SubjectFact]) -> PromoteFlow` *(new)*, which loads a `PromoteFlow` with one proposal row per fact, as `twoHundredLines` does (kind from the fact, `target: "new"`, `assessment: "none"`), and T-2's `neighbor` is `{ page.row($0) }`. The Card's own neighbor is no fact, so its header stays "they share a spouse". Dates are `CatalogDateValueInput(kind: "point", startYear:startMonth:startDay:)`; expected text is computed through `DateRowDisplay.line`, never a literal | The fixtures the suites already use, grown by one parameter; rows come from the real `load`, so `Row`'s many fields need no factory (review X-3, X-4) | A `Row` memberwise factory in the tests — fifteen arguments to keep in step |
| D-14 | Tests are Swift Testing in the existing suites (`add-swift-test`), comparing English as the neighboring tests do (`card.header == "they share a similar spouse"`) | The suites already run under the English catalog | A new suite file |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionHeaderDisplay` *(new)* `title`, `secondary`, `ref` | enum, funcs | `PromoteFlow.conclusionTitle` / `conclusionSubtitle`; `CatalogSearchHitPresentation.title(for:)` / `secondary(for:)` | Extracts those two copies (Deletes) before `Pivot` would add a third |
| `EvidenceBoard.Pivot`, `Pivot(header:orders:)` | type, init | `PromoteFlow.Alternative`'s title fields | The phrase needs the name, the date alone and the ref apart; built through `ConclusionHeaderDisplay` |
| `RelationshipCard.recorded` | field | `Line.thisSource` (formatted in `build`) | A Card-level line, formatted where `Line.thisSource` is |
| `PromoteEvidenceDisplay.recorded`, private `wording`, `implied`, `direct`, `PatternWording`, `PivotShape` | funcs, types | `PromoteEvidenceDisplay.sharePhrase`; `ConnectionSentenceRegistry.share` | The share registry words a bridge kind and term from the neighbor's side; this words a kinship pattern per side. Its direct half keys on the same term `neighborTerm` the share phrase reads |
| `EvidenceGraphPlacedSubject.dateLine` | computed property | `displayName(orders:)`; `EventSecondaryDisplay.content` | Same date rule, read from Observations instead of a header |
| `SubjectFact.date`, `Row.date` | fields | `SubjectFact.name`, `Row.name` | Extends the same carrier |
| `L10n.Promote.recorded…` (§6 Strings) | L10n | `L10n.Promote.shareHeader`, `thisSource`, `spokenRelationship` | New copy; same nest and key prefix `promote.sheet.` |
| Test helper `page(_:)`; `observation(…, date:)` | test helpers | `twoHundredLines`; `observation` | Grows the existing fixtures |
| Tests `recordedNamesHowEachSideRecordsIt`, `aRelationshipCardSaysHowEachSideRecordedIt`, `anEventsDateLineReadsItsDateObservations`, `aRowCarriesItsEventsDate`, `aHeaderIsTitledByItsKind` | tests | `relationshipShareFromTheEngine`, `aRelationshipCardIsWordedFromTheNeighborsSide` | New behavior in the existing suites |

Searched: `rg "KinshipPattern|\.layer\b|\.canon\b" macos/App/Features`, `rg "DateRowDisplay.line" macos/App`, `rg "PersonHeaderDisplay.title|EventTitleDisplay.title|PlaceTitleDisplay.title" macos/App`, `rg "case .person\(let header\)" macos/App`, `rg "struct SubjectFact|struct Row\b" macos/App/Features/Promote`, `rg "promote.sheet.(thisSource|recordCount|shareHeader)" macos/App`, `rg "SeededRelationshipTypeTerm" macos/App`.

**Builds on.** S11-40's `CatalogClaimEvidenceShape` and `KinshipPattern`; the board/display split from `8aab38bf`; the record-count caption style; `DateRowDisplay` and the per-kind header displays; the Card's `neighborTerm` reading.
**Deletes.** `PromoteFlow.conclusionTitle` and `conclusionSubtitle`; the two `switch hit.header` blocks in `CatalogSearchHitPresentation`. The add-match-rule step 6 sentence "(S11-42) follows" is replaced (STEP-6).
**Next adds.** A new kinship pattern adds its `KinshipPattern` case (S11-40), one arm in `PromoteEvidenceDisplay.wording` and two keys (named, none); a pattern implying a link the sheet already words directly needs no direct key. The exhaustive `switch` in `wording` fails the build until the arm is there. A new relationship term an implied link can carry adds one arm in `direct(term:)`. A second reader of these phrases (a tree citation, [`kinship-view.md`](../../../ideas/kinship-view.md) "Citing an implied fact on a tree") moves `recorded`, `Pivot` and the wording out of Promote into a neutral home in that PR.

### Schema

None — no catalog change.

### Go core

None — no Go change.

### FFI

None — S11-40 carries the shapes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Header display | `ConclusionHeaderDisplay` *(new)* | `macos/App/Features/Conclusions/ConclusionHeaderDisplay.swift` *(new)* | D-12 |
| Date on a placed Subject | `EvidenceGraphPlacedSubject.dateLine` *(new)* | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` | D-10 |
| Row date | `SubjectFact.date`, `Row.date` *(new)* | `macos/App/Features/Promote/PromoteFlow.swift`; set in `PromoteModel.swift` `subjectFacts` | D-11 |
| Pivot, caption | `EvidenceBoard.Pivot`, `RelationshipCard.recorded` *(new)* | `macos/App/Features/Promote/PromoteEvidenceBoard.swift` | D-6, D-8 |
| Phrase choice | `PromoteEvidenceDisplay.recorded` and its private helpers *(new)* | `macos/App/Features/Promote/PromoteEvidenceDisplay.swift` | D-3 – D-5 |
| Caption view | `EvidenceRelationshipCard.header` | `macos/App/Features/Promote/PromoteEvidenceCards.swift` | D-9; no new `PV*` |

States: both sides direct → no caption (AC-1.4); a Source pivot not on the page → the none phrase (AC-1.6); a tree pivot with no header → the none phrase (AC-1.9); a long caption wraps (`fixedSize(horizontal: false, vertical: true)`). Accessibility: the caption is hidden; the pin box's `spoken` carries it (D-8).

### Strings

L10n under `L10n.Promote` *(new)*. Each catalog entry is written as the `add-localized-string` skill's "Steps" show (key, `comment`, `extractionState: manual`, `localizations.en.stringUnit`), with the comment below. Argument-free keys are `static let … = LocalizedStringResource(…)`; others are `static func … -> String` through `L10n.format`, as `shareHeaderBare` is.

| Function / key | English | Comment |
| --- | --- | --- |
| `recorded(source:record:)` · `promote.sheet.recorded` | "This source: %1$@ · The record: %2$@" | Muted line under a relationship Card's header: how each side records the link; 1 = this Source's way, 2 = the matched record's way |
| `spokenRecorded(header:source:record:)` · `promote.sheet.spokenRecorded` | "%1$@. This source: %2$@. The record: %3$@" | VoiceOver reading of a relationship Card's header and how each side records the link; 1 = the header, 2 = this Source's way, 3 = the record's way |
| `recordedOn(_:)` · `promote.sheet.recorded.on` | "on %@" | An event named by its date inside a recorded phrase; the argument is the date |
| `recordedMarriage(_:)` · `promote.sheet.recorded.marriage` | "married in the marriage %@" | How a side records a couple: through a marriage; the argument is "on <date>" or the marriage's ref |
| `recordedMarriageNone` · `promote.sheet.recorded.marriage.none` | "married in a marriage" | How a side records a couple through a marriage that cannot be named |
| `recordedBirth(_:)` · `promote.sheet.recorded.birth` | "parent and child in the birth %@" | How a side records a parent through a birth; the argument is "on <date>" or the birth's ref |
| `recordedBirthNone` · `promote.sheet.recorded.birth.none` | "parent and child in a birth" | How a side records a parent through a birth that cannot be named |
| `recordedChain(_:)` · `promote.sheet.recorded.chain` | "grandparent and grandchild through %@" | How a side records a grandparent through a parent between; the argument is that parent's name or ref |
| `recordedChainNone` · `promote.sheet.recorded.chain.none` | "grandparent and grandchild through a parent" | The same, when the parent between cannot be named |
| `recordedOther(_:)` · `promote.sheet.recorded.other` | "through %@" | How a side records a link through a pattern this build has no words for; the argument is the record between |
| `recordedOtherNone` · `promote.sheet.recorded.other.none` | "through another record" | The same, when the record between cannot be named |
| `recordedDirectSpouses` · `promote.sheet.recorded.direct.spouses` | "recorded as spouses" | A side that records the couple directly |
| `recordedDirectParentChild` · `promote.sheet.recorded.direct.parentChild` | "recorded as parent and child" | A side that records the parent directly |
| `recordedDirectGrandparent` · `promote.sheet.recorded.direct.grandparent` | "recorded as grandparent and grandchild" | A side that records the grandparent directly |
| `recordedDirectOther` · `promote.sheet.recorded.direct.other` | "recorded directly" | A side that records the link directly, of a kind with no words of its own |

The dated marriage reads "married in the marriage on 12 May 1842"; by ref, "married in the marriage EVT-M". Removed: none.

## 7. Steps

**STEP-1 — A Source event's date on its row**
Files: `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift`, `macos/App/Features/Promote/PromoteFlow.swift`, `PromoteModel.swift`, `macos/ProvenenciaTests/EvidenceGraphSnapshotTests.swift`, `PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: D-10, D-11, D-13's `observation(…, date:)`. T-3, T-4.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes T-3 and T-4 and every other test unchanged.

**STEP-2 — One display of a conclusion header**
Files: `macos/App/Features/Conclusions/ConclusionHeaderDisplay.swift` *(new)*, `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift`, `macos/App/Features/Promote/PromoteFlow.swift`, `macos/ProvenenciaTests/CatalogSearchTests.swift`.
Do: D-12. T-5.
Check: `xcodebuild test` passes T-5, `CatalogSearchTests`, `PromoteFlowTests` and `PromoteModelTests` with no expected value changed; `rg -n "func conclusionTitle|func conclusionSubtitle" macos/App` prints nothing.

**STEP-3 — The side phrases**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Promote/PromoteEvidenceBoard.swift` (the `Pivot` type and its init only), `PromoteEvidenceDisplay.swift`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`. Skill: `add-localized-string`.
Do: §6 Strings; D-3 – D-6 (`Pivot`). T-1. This step's commit carries the `Pivot` type.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test` passes T-1.

**STEP-4 — The board fills the Card**
Files: `macos/App/Features/Promote/PromoteEvidenceBoard.swift`, `PromoteModel.swift`, `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift`.
Do: D-6 (the two pivots in `relationshipCard`), D-7, D-8, D-13's `page(_:)`. T-2.
Check: `xcodebuild test` passes T-2; `aRelationshipCardIsWordedFromTheNeighborsSide` passes unchanged.

**STEP-5 — The Card draws the caption**
Files: `macos/App/Features/Promote/PromoteEvidenceCards.swift`. Skill: `add-ui-component`.
Do: D-9.
Check: `scripts/check-pr.sh` prints `check-pr: PASS` (no Swift warning); §10 steps 3 and 5 show the caption on one Card and none on the other. AC-1.4's "no caption drawn" is checked by hand only (a view); T-2 pins `recorded == nil`.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-11/claim-evidence.md` (§ Evidence through relationships, "Implied links (S11-40)": a bullet "The Card says how each side recorded it (S11-42)": a muted line under the header, both sides named, an implied side by its pattern and pivot, a direct side by the link's term, words in `PromoteEvidenceDisplay.recorded`; no English copy), `docs/deployment-plan/spike-11/kinship-shapes.md` (§ On the evidence sheet: the first bullet's example becomes "This source: married in the marriage on 12 May 1842 · The record: recorded as spouses", under the unchanged "they share a spouse" header), `.agents/skills/add-match-rule/SKILL.md` ("Add a kinship pattern" step 6: replace "The claim evidence sheet's wording for it (S11-42) follows." with "Then add its arm to `PromoteEvidenceDisplay.wording` and its two `promote.sheet.recorded.*` keys; the `switch` does not compile without it."), `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status, Landed note). The plan's "Landed in #N" and the map status are written after the draft PR is open, so its number is known.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-42-implied-links-sheet.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Say on the evidence sheet how the Source and the record each recorded a relationship."). STEP-2's subject says the titles are one display ("Title every conclusion header through one display, for search and Promote alike."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift` `recordedNamesHowEachSideRecordsIt` *(new)* | M = `Pivot(name: "Marriage of Ann Ames and Bob Ames", date: "12 May 1842", ref: "EVT-M")`, spouse term "spouse". (a) `(marriageSpouses, M, direct, nil, "spouse")` → `("married in the marriage on 12 May 1842", "recorded as spouses")`. (b) mirrored → swapped. (c) both marriage → both dated. (d) both direct → nil. (e) M with date "" → "married in the marriage EVT-M". (f) layer pivot nil → "married in a marriage". (g) `(direct, nil, birthParents, Pivot(name: "Birth of Cal Ames", date: "1 Mar 1850", ref: "EVT-B"), "parent")` → `("recorded as parent and child", "parent and child in the birth on 1 Mar 1850")`; term "child" → the same direct phrase. (h) `(direct, nil, parentOfParent, Pivot(name: "Bea Ames", date: "", ref: "PER-B"), "grandparent")` → `("recorded as grandparent and grandchild", "grandparent and grandchild through Bea Ames")`; canon pivot nil → "grandparent and grandchild through a parent". (i) `(unrecognized("x"), M, direct, nil, "sibling")` → `("through Marriage of Ann Ames and Bob Ames", "recorded directly")`. `Pivot(header: .event(Mʹ), orders: NameFormatOrders())` → date `EventSecondaryDisplay.content(Mʹ).date`, ref "EVT-T"; `Pivot(header: .person(Bea), orders:)` → name "Bea Ames", date ""; `Pivot(header: nil, orders:)` → nil | AC-1.1, AC-1.2, AC-1.3, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, D-3, D-4, D-5, D-6 |
| T-2 | Swift | `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift` `aRelationshipCardSaysHowEachSideRecordedIt` *(new)* | `page([SubjectFact(id: "M", kind: .event, name: "Marriage of Ann Ames and Bob Ames", ref: "EVT-M", anchor: nil, date: "12 May 1842")])`; spouse line on neighbor "n1" (no page row) with `layer` `{marriageSpouses, "M"}` → `recorded` = `L10n.Promote.recorded(source: "married in the marriage on 12 May 1842", record: "recorded as spouses")`; `spoken` = `spokenRelationship(header: spokenRecorded(header: "they share a spouse", source: …, record: …), outcome: .agree, count: 1)`. Spouse line with `canon.pivot = .event(Mʹ)` → AC-1.2's caption. Direct–direct → `recorded` nil and `spoken` as on `main` | AC-1.1, AC-1.2, AC-1.4, AC-1.11, D-7, D-8, D-13 |
| T-3 | Swift | `macos/ProvenenciaTests/EvidenceGraphSnapshotTests.swift` `anEventsDateLineReadsItsDateObservations` *(new)* | Event with a positive `date` 1842-05-12 → `DateRowDisplay.line(date:)` of it; event with `start_date` 1840 and `end_date` 1845 → `DateRowDisplay.line(date: nil, start:, end:)`; event with only a negative `date` → ""; person with a `date` Observation → "" | AC-2.1, AC-2.2, AC-2.3, D-10 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aRowCarriesItsEventsDate` *(new)* | `flow.load` with an `.event` `SubjectFact` of `date` "12 May 1842", once with an `anchor` and once without → `flow.row(id)?.date` "12 May 1842" both times | AC-2.4, D-11 |
| T-5 | Swift | `macos/ProvenenciaTests/CatalogSearchTests.swift` `aHeaderIsTitledByItsKind` *(new)* | `ConclusionHeaderDisplay.title(.person(h), style: .sorted)` == `PersonHeaderDisplay.title(h, style: .sorted)`; `.event(e)` == `EventTitleDisplay.title(e)`; `.place(p)` == `PlaceTitleDisplay.title(p)`; `secondary` likewise per kind; `ref` == each header's `entity.ref` | D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CatalogSearchTests` and `PromoteFlowTests` (titles through the new display), `PromoteEvidenceBoardTests` (the `build` helper gains a defaulted parameter), `EvidenceGraphSnapshotTests` (the `observation` helper gains one), and `scripts/check-localizable-xcstrings.py`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — not gated; no new PV* component
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; docs/deployment-plan/spike-11/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc and skill updated: claim-evidence.md; kinship-shapes.md (§ On the evidence sheet); add-match-rule SKILL.md step 6
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

This check also runs S11-40's manual check, which S11-40 did not run. If step 3 shows no relationship Card for Ann at all, stop: that is an S11-40 defect, reported with the sheet's screenshot, not worked around here.

1. Open a project. On Source "Tree", add Ann Ames and Bob Ames and a marriage dated 12 May 1842 with both as `subject`; promote all three. On Source "Register", add Ann Ames and Bob Ames in a `spouse` relationship.
2. Open Promote for "Register". Bob's row targets the tree Bob, Ann's the tree Ann.
3. Open Bob's evidence sheet: the Card "…they share a spouse (…)" shows under its header "This source: recorded as spouses · The record: married in the marriage on 12 May 1842" (AC-1.12). With VoiceOver on, the Card's box reads the header, both sides and the outcome.
4. Press Done without unchecking. `identity_claim_evidence` for Bob's new claim holds both tree marriage participations' Observations.
5. On Source "Census", add Ann Ames and Bob Ames in a `spouse` relationship and promote it against a tree that records them only as `spouse`: the Card shows no caption (AC-1.4).
6. On Source "Wedding", add Ann Ames and Bob Ames as `subject`s of a marriage dated 12 May 1842, against a tree that records them only as `spouse`: Bob's sheet reads "This source: married in the marriage on 12 May 1842 · The record: recorded as spouses".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Source pivot is not a page row, so its date is unknown | The none phrase (AC-1.6); the Source graph is not read again (Out) |
| The caption crowds the header at narrow widths | It wraps on its own line (D-9); the pin box and badge keep their places |
| A new pattern ships without wording | The exhaustive `switch` in `wording` stops the build (Next adds) |
| STEP-2 changes a title somewhere | It calls the same per-kind helpers with the same style; `CatalogSearchTests`, `PromoteFlowTests` and `PromoteModelTests` must pass unchanged |

**Stop and report if:**

- `CatalogClaimEvidenceRelationship.layer` / `.canon` or `CatalogClaimEvidenceShape` no longer have the fields §4 lists, or `KinshipPattern` has a case §4 does not list.
- §10 step 3 shows no relationship Card for Ann (S11-40 defect).
- A Source graph's `CatalogObservation.date` is nil for a `date` Observation that the Events list shows with a date.
- A layer shape's `pivotID` is not a page row's `subjectID` in §10 step 6.
- STEP-2 needs an expected title or subtitle changed in any existing test.
- `EvidenceBoard.build` has a caller other than `PromoteModel.board(of:)`, `PromoteEvidenceBoardTests` and `PromoteFlowTests`.
- Adding `Row.date`, `SubjectFact.date` or the `observation` helper's `date` needs any existing test's expected value changed.

## 12. Open questions

None.

## 13. PR

**Title:** S11-42 — Implied links on the evidence sheet
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-42-implied-links-sheet.md`.
