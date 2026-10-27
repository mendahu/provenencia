# S11-42 — Implied links on the evidence sheet

**Status:** Draft
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
**Environment:** every step needs a Mac with the repo's toolchain (`xcodebuild test`, `scripts/check-pr.sh`). No Go, proto or FFI change.
**Context:** [Kinship shapes § On the evidence sheet](../kinship-shapes.md#on-the-evidence-sheet); [`claim-evidence.md` § Wording](../claim-evidence.md#wording) and [§ Evidence through relationships](../claim-evidence.md#evidence-through-relationships); S11-40 plan D-9, D-11, D-13 and its Landed note; working rules "One wire vocabulary", "Counted once across sides"
**Planned:** 2026-10-27 against `main` @ `c545a482`
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
| AC-1.10 | `layer` `{unrecognized("sibling_parents"), pivotID "M"}`, `canon` direct | "This source: through Marriage of Ann Ames and Bob Ames · The record: recorded directly" |
| AC-1.11 | AC-1.1 | Card `spoken` = `spokenRelationship` of "they share a spouse. This source: married in the marriage on 12 May 1842. The record: recorded as spouses" and the outcome; the caption is hidden from VoiceOver |
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
| The two side phrases and their L10n, as one pure function | STEP-2, AC-1.1 – AC-1.10 |
| The board fills `RelationshipCard.recorded` from the shapes, the page row of a Source pivot and the tree pivot's header; the spoken label reads it | STEP-3, AC-1.1, AC-1.2, AC-1.4, AC-1.11 |
| The Card draws the caption | STEP-4, AC-1.4, AC-1.12 |
| Docs: `claim-evidence.md`, `kinship-shapes.md`, `add-match-rule` skill, `completed.md`, map | STEP-5 |

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
| Headers | `macos/App/Features/Conclusions/PersonHeaderDisplay.swift` `PersonHeaderDisplay.title`; `EventTitleDisplay.title`; `EventsListView.swift` `EventSecondaryDisplay.content(_:).date` | Name and date of a header, as Promote's menu and the Events list show them |
| Keys | `macos/App/Platform/ConclusionVocabulary.swift` `SeededPropertyKey.date`, `.startDate`, `.endDate` | Spelled once |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Promote` (`shareHeader`, `recordCount`, `thisSource`, `spokenRelationship`); `macos/App/Resources/Localizable.xcstrings` | No shape copy |
| Tests | `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift` `build` helper, `aRelationshipCardIsWordedFromTheNeighborsSide`; `PromoteEvidenceDisplayTests.swift` `relationshipShareFromTheEngine`; `EvidenceGraphSnapshotTests.swift`; `PromoteFlowTests.swift` | Pin today's header and spoken label |
| Skill | `.agents/skills/add-match-rule/SKILL.md` "Add a kinship pattern" step 6 | "The claim evidence sheet's wording for it (S11-42) follows." |

**Precedent.** `8aab38bf` ("…a share phrase on each relationship") added the Card's header words as a `PromoteEvidenceDisplay` function tested in `PromoteEvidenceDisplayTests`, with the board filling a `RelationshipCard` field once per load: copy that split (pure words in `PromoteEvidenceDisplay`, assembly in `EvidenceBoard.build`, the view only draws a `String`). The `recordCount` caption in `EvidenceRelationshipCard.header` is the style the new caption copies.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The Card keeps its header. A caption line under the header text shows when either side's `pattern` is not `.direct`; it is nil when both are direct | User, 2026-10-27 (placement); a direct–direct Card stays as `main` | Folding into the header — one side only, long header; a brief — no new component |
| D-2 | The caption names both sides: `L10n.Promote.recorded(source:record:)` *(new)*, key `promote.sheet.recorded`, "This source: %1$@ · The record: %2$@". "This source" is the layer, "The record" the canon | User, 2026-10-27 (both sides); one whole format, as S11-33 requires; "record" is the sheet's word for the matched entity (`shareHeader`'s "connected record") | Only the implied side |
| D-3 | A side's phrase by its pattern and pivot (`EvidenceBoard.Pivot`, D-5), English defaults, each one L10n key: **marriageSpouses** — date ≠ "" "married in the marriage on %@" (`promote.sheet.recorded.marriage.dated`), else ref ≠ "" "married in the marriage %@" (`….marriage.ref`), else "married in a marriage" (`….marriage.none`). **birthParents** — "parent and child in the birth on %@" (`….birth.dated`), "parent and child in the birth %@" (`….birth.ref`), "parent and child in a birth" (`….birth.none`). **parentOfParent** — name ≠ "" "grandparent and grandchild through %@" (`….chain.named`, the name), else ref ≠ "" the same key with the ref, else "grandparent and grandchild through a parent" (`….chain.none`). **unrecognized** — "through %@" (`….other.named`) with the name, else the ref, else "through another record" (`….other.none`). **direct** — by the other side's pattern: marriageSpouses "recorded as spouses" (`….direct.spouses`), birthParents "recorded as parent and child" (`….direct.parentChild`), parentOfParent "recorded as grandparent and grandchild" (`….direct.grandparent`), unrecognized or direct "recorded directly" (`….direct.other`) | User, 2026-10-27 (event by date, person by name); whole phrases with one argument each; a direct side reads in the other side's terms, which is the one key the line stands for | Splicing an event-type word into a shared template — S11-33 forbids it; naming the event by title — repeats the couple's names |
| D-4 | The words live in `PromoteEvidenceDisplay.recorded(layer: KinshipPattern, layerPivot: EvidenceBoard.Pivot?, canon: KinshipPattern, canonPivot: EvidenceBoard.Pivot?) -> EvidenceBoard.Recorded?` *(new)*, returning `Recorded {source: String; record: String}` *(new)*, nil when both patterns are `.direct`; a private `phrase(_:pivot:other:)` *(new)* holds D-3's switch | Pure and testable without a view, as `sharePhrase` and `shareHeader` are | Wording in the view — untestable; wording in `L10n.Promote` — L10n holds formats, not choices |
| D-5 | `EvidenceBoard.Pivot` *(new)* `{name: String; date: String; ref: String}`. From a Source pivot: `neighbor(shape.pivotID)` (the page row) → `Pivot(name: row.name, date: row.date, ref: row.ref)`; nil when `pivotID` is "" or no row has it. From the tree pivot: `Pivot(header:orders:)` *(new, failable)* — `.person(h)` → `(PersonHeaderDisplay.title(h, orders: orders), "", h.entity.ref)`, `.event(h)` → `(EventTitleDisplay.title(h, orders: orders), EventSecondaryDisplay.content(h).date, h.entity.ref)`, `.place` or nil → nil. The layer shape's `pivot` header is never read (S11-40 D-11 sends none) | One value both sides fill, so D-3 has one input shape; reuses the display helpers Promote's menu titles with | Reading the Source pivot's header from Go — S11-40 D-11 decided a layer pivot is found on the page |
| D-6 | `EvidenceBoard.build` gains `orders: NameFormatOrders = NameFormatOrders()` as its last parameter; `PromoteModel.board(of:)` passes `session.nameFormat`. The `neighbor` closure's doc comment changes to "the page row a Subject is (a connected row, or a relationship's Source pivot)" | The tree pivot's person name follows the researcher's name order, as every other name on the sheet does | A second closure for pivots — `neighbor` already answers any page row |
| D-7 | `RelationshipCard` gains `recorded: Recorded?` *(new field)*. When set, `spoken` is `spokenRelationship(header: L10n.Promote.spokenRecorded(header:source:record:), …)` with `spokenRecorded` *(new)*, key `promote.sheet.spokenRecorded`, "%1$@. This source: %2$@. The record: %3$@"; otherwise `spoken` is unchanged | VoiceOver reads the pin box's `spoken`, and the header and caption are hidden from it; a spoken "·" reads badly, so the spoken form is its own whole format | Joining with `L10n.Conclusions.a11yList` — a Conclusions format with another meaning |
| D-8 | `EvidenceRelationshipCard.header`: the `HStack` holding the header text and record count becomes the first child of a `VStack(alignment: .leading, spacing: PVSpacing.space1)`; when `relationship.recorded` is set, the second child is `Text(verbatim: L10n.Promote.recorded(source:record:))` in `PVFont.body(size: PVTypeScale.caption)`, `PVColor.textMuted`, `.fixedSize(horizontal: false, vertical: true)`. The `VStack` takes the `.frame(maxWidth: .infinity, alignment: .leading)` and `.accessibilityHidden(true)` the `HStack` has today. The pin box's and badge's `HStack` alignment stays `.center` | The record count's caption style; no new kit component (`add-ui-component`: a snowflake line inside an existing feature Card) | A `PVSectionHeader` subtitle — the Card header is not a section header |
| D-9 | `EvidenceGraphPlacedSubject.dateLine: String` *(new, computed)*: "" unless `kind == .event`; else `DateRowDisplay.line(date:start:end:)` of the first Observation each with `polarity != "negative"`, a non-nil `date`, and `propertyKey` `SeededPropertyKey.date` / `.startDate` / `.endDate` | The Events list's date rule (`EventSecondaryDisplay.content`) on the Source's own Observations, so a Source marriage and a tree marriage read alike | Reading it in the board from the snapshot — the board has rows, not the snapshot |
| D-10 | `PromoteFlow.SubjectFact` gains `var date: String = ""` and `Row` gains `var date: String = ""`, each the last stored property. `PromoteModel.subjectFacts` sets `date: placed.dateLine`; both `Row(…)` calls in `makeRow` pass `date: fact.date`; a decided row keeps `previous` as today | The row already carries the Subject's name and ref the same way; defaults keep every test construction compiling | A dictionary of dates on `PromoteFlow` — a second lookup beside `row(_:)` |
| D-11 | Tests are Swift Testing in the existing suites (`add-swift-test`), comparing English through the L10n functions or literal English as the neighboring tests do (`card.header == "they share a similar spouse"`) | The suites already run under the English catalog | A new suite file |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `EvidenceBoard.Pivot`, `Pivot(header:orders:)` | type, init | `PromoteFlow.conclusionTitle` / `conclusionSubtitle` (private) | Those return a menu title and subtitle line; the phrase needs the name, the date alone and the ref apart. The init calls the same `PersonHeaderDisplay` / `EventTitleDisplay` / `EventSecondaryDisplay` helpers |
| `EvidenceBoard.Recorded`, `RelationshipCard.recorded` | type, field | `Group.thisSource` | A Card-level second line with two parts, read by both the view and the spoken label |
| `PromoteEvidenceDisplay.recorded`, `phrase` | funcs | `PromoteEvidenceDisplay.sharePhrase`, `ConnectionSentenceRegistry.share` | The share registry words a bridge kind and term from the neighbor's side; this words a kinship pattern per side. Adding it to the registry would put a pattern switch in a kind table |
| `EvidenceGraphPlacedSubject.dateLine` | computed property | `displayName(orders:)`; `EventSecondaryDisplay.content` | Same date rule, read from Observations instead of a header |
| `SubjectFact.date`, `Row.date` | fields | `SubjectFact.name`, `Row.name` | Extends the same carrier |
| `L10n.Promote.recorded`, `spokenRecorded`, `recordedPhrase…` keys | L10n | `L10n.Promote.shareHeader`, `thisSource`, `spokenRelationship` | New copy; same nest and key prefix `promote.sheet.` |
| Tests `recordedNamesHowEachSideRecordsIt`, `aRelationshipCardSaysHowEachSideRecordedIt`, `anEventsDateLineReadsItsDateObservations`, `aRowCarriesItsEventsDate` | tests | `relationshipShareFromTheEngine`, `aRelationshipCardIsWordedFromTheNeighborsSide` | New behavior in the existing suites |

Searched: `rg "KinshipPattern|\.layer\b|\.canon\b" macos/App/Features`, `rg "DateRowDisplay.line" macos/App`, `rg "func conclusionTitle|PersonHeaderDisplay.title|EventSecondaryDisplay" macos/App/Features/Promote`, `rg "struct SubjectFact|struct Row\b" macos/App/Features/Promote`, `rg "promote.sheet.(thisSource|recordCount|shareHeader)" macos/App`.

**Builds on.** S11-40's `CatalogClaimEvidenceShape` and `KinshipPattern`; the board/display split from `8aab38bf`; the record-count caption style; `DateRowDisplay`, `PersonHeaderDisplay`, `EventTitleDisplay`.
**Deletes.** None — nothing becomes dead. The add-match-rule step 6 sentence "(S11-42) follows" is replaced (STEP-5).
**Next adds.** A new kinship pattern adds its `KinshipPattern` case (S11-40), one `case` in `PromoteEvidenceDisplay.phrase`, its dated/ref/none (or named/none) keys and its direct phrase key. Swift's exhaustive `switch` over `KinshipPattern` in `phrase` fails the build until it is added. S13-04's Person page can call `PromoteEvidenceDisplay.recorded`'s phrase switch if it words the same shapes; it is not moved now.

### Schema

None — no catalog change.

### Go core

None — no Go change.

### FFI

None — S11-40 carries the shapes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Date on a placed Subject | `EvidenceGraphPlacedSubject.dateLine` *(new)* | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` | D-9 |
| Row date | `SubjectFact.date`, `Row.date` *(new)* | `macos/App/Features/Promote/PromoteFlow.swift`; set in `PromoteModel.swift` `subjectFacts` | D-10 |
| Pivot, recorded words | `EvidenceBoard.Pivot`, `EvidenceBoard.Recorded`, `RelationshipCard.recorded` *(new)* | `macos/App/Features/Promote/PromoteEvidenceBoard.swift` | D-5, D-7 |
| Phrase choice | `PromoteEvidenceDisplay.recorded`, `phrase` *(new)* | `macos/App/Features/Promote/PromoteEvidenceDisplay.swift` | D-3, D-4 |
| Caption | `EvidenceRelationshipCard.header` | `macos/App/Features/Promote/PromoteEvidenceCards.swift` | D-8; no new `PV*` |

States: both sides direct → no caption (AC-1.4); a Source pivot not on the page → the no-pivot phrase (AC-1.6); a tree pivot with no header → the no-pivot phrase (AC-1.9); a long caption wraps (`fixedSize(horizontal: false, vertical: true)`). Accessibility: the caption is hidden; the pin box's `spoken` carries it (D-7).

### Strings

L10n under `L10n.Promote` *(new)*, each with a translator comment naming its argument:

| Function / key | English |
| --- | --- |
| `recorded(source:record:)` · `promote.sheet.recorded` | "This source: %1$@ · The record: %2$@" |
| `spokenRecorded(header:source:record:)` · `promote.sheet.spokenRecorded` | "%1$@. This source: %2$@. The record: %3$@" |
| `recordedMarriageDated(_:)` · `promote.sheet.recorded.marriage.dated` | "married in the marriage on %@" |
| `recordedMarriageRef(_:)` · `promote.sheet.recorded.marriage.ref` | "married in the marriage %@" |
| `recordedMarriageNone` · `promote.sheet.recorded.marriage.none` | "married in a marriage" |
| `recordedBirthDated(_:)` · `promote.sheet.recorded.birth.dated` | "parent and child in the birth on %@" |
| `recordedBirthRef(_:)` · `promote.sheet.recorded.birth.ref` | "parent and child in the birth %@" |
| `recordedBirthNone` · `promote.sheet.recorded.birth.none` | "parent and child in a birth" |
| `recordedChainNamed(_:)` · `promote.sheet.recorded.chain.named` | "grandparent and grandchild through %@" |
| `recordedChainNone` · `promote.sheet.recorded.chain.none` | "grandparent and grandchild through a parent" |
| `recordedOtherNamed(_:)` · `promote.sheet.recorded.other.named` | "through %@" |
| `recordedOtherNone` · `promote.sheet.recorded.other.none` | "through another record" |
| `recordedDirectSpouses` · `promote.sheet.recorded.direct.spouses` | "recorded as spouses" |
| `recordedDirectParentChild` · `promote.sheet.recorded.direct.parentChild` | "recorded as parent and child" |
| `recordedDirectGrandparent` · `promote.sheet.recorded.direct.grandparent` | "recorded as grandparent and grandchild" |
| `recordedDirectOther` · `promote.sheet.recorded.direct.other` | "recorded directly" |

Argument-free keys are `static let … = LocalizedStringResource(…)`; one-argument keys are `static func … -> String` through `L10n.format`, as `shareHeaderBare` is. Removed: none.

## 7. Steps

**STEP-1 — A Source event's date on its row**
Files: `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift`, `macos/App/Features/Promote/PromoteFlow.swift`, `PromoteModel.swift`, `macos/ProvenenciaTests/EvidenceGraphSnapshotTests.swift`, `PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: D-9, D-10. T-3, T-4.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes `EvidenceGraphSnapshotTests` and `PromoteFlowTests`, with no other test changed.

**STEP-2 — The side phrases**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Promote/PromoteEvidenceBoard.swift` (`Pivot`, `Recorded` types only), `PromoteEvidenceDisplay.swift`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`. Skill: `add-localized-string`.
Do: §6 Strings; D-3, D-4, D-5's `Pivot` type and `Pivot(header:orders:)`. T-1.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test` passes `PromoteEvidenceDisplayTests`.

**STEP-3 — The board fills the Card**
Files: `macos/App/Features/Promote/PromoteEvidenceBoard.swift`, `PromoteModel.swift`, `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift`.
Do: D-5 (the Source and tree pivots in `relationshipCard`), D-6, D-7. T-2.
Check: `xcodebuild test` passes `PromoteEvidenceBoardTests`; `aRelationshipCardIsWordedFromTheNeighborsSide` passes unchanged.

**STEP-4 — The Card draws the caption**
Files: `macos/App/Features/Promote/PromoteEvidenceCards.swift`. Skill: `add-ui-component`.
Do: D-8.
Check: `scripts/check-pr.sh` prints `check-pr: PASS` (no Swift warning); the manual check (§10) shows the caption under the header and none on a direct–direct Card.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-11/claim-evidence.md` (§ Evidence through relationships, "Implied links (S11-40)": a bullet "The Card says how each side recorded it (S11-42)": a muted line under the header, both sides named, words in `PromoteEvidenceDisplay.recorded`; no English copy), `kinship-shapes.md` (§ On the evidence sheet: the first bullet's example becomes "This source: married in the marriage on 12 May 1842 · The record: recorded as spouses", under the unchanged "they share a spouse" header), `.agents/skills/add-match-rule/SKILL.md` ("Add a kinship pattern" step 6: replace "The claim evidence sheet's wording for it (S11-42) follows." with "Then add its case to `PromoteEvidenceDisplay.phrase` and its `promote.sheet.recorded.*` keys; the `switch` does not compile without it."), `completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-42-implied-links-sheet.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Say on the evidence sheet how the Source and the record each recorded a relationship."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift` `recordedNamesHowEachSideRecordsIt` *(new)* | Pivots built directly: M = `Pivot(name: "Marriage of Ann Ames and Bob Ames", date: "12 May 1842", ref: "EVT-M")`. (a) `(marriageSpouses, M, direct, nil)` → `source` "married in the marriage on 12 May 1842", `record` "recorded as spouses". (b) mirrored → swapped. (c) both marriage → both dated phrases. (d) both direct → nil. (e) M with date "" → "married in the marriage EVT-M". (f) layer pivot nil → "married in a marriage". (g) `(direct, nil, birthParents, Pivot(…, date: "1 Mar 1850", …))` → "recorded as parent and child" / "parent and child in the birth on 1 Mar 1850". (h) `(direct, nil, parentOfParent, Pivot(name: "Bea Ames", date: "", ref: "PER-B"))` → "recorded as grandparent and grandchild" / "grandparent and grandchild through Bea Ames"; canon pivot nil → "grandparent and grandchild through a parent". (i) `(unrecognized("x"), M, direct, nil)` → "through Marriage of Ann Ames and Bob Ames" / "recorded directly". Also `Pivot(header: .event(Mʹ), orders:)` → date = `EventSecondaryDisplay.content(Mʹ).date`, ref "EVT-T"; `Pivot(header: .person(Bea), orders:)` → name "Bea Ames", date ""; `.place` → nil | AC-1.1, AC-1.2, AC-1.3, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, D-3, D-4, D-5 |
| T-2 | Swift | `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift` `aRelationshipCardSaysHowEachSideRecordedIt` *(new)* | Spouse line with `layer` `{marriageSpouses, "M"}`, `neighbor` returning row M (kind `.event`, date "12 May 1842") for "M" → `recorded` (a); `spoken` == `spokenRelationship(header: spokenRecorded(header: "they share a spouse", …), outcome: .agree, count: 1)`. Spouse line with `canon.pivot = .event(Mʹ)` → (b). Direct–direct → `recorded` nil and `spoken` as `aRelationshipCardIsWordedFromTheNeighborsSide` computes it | AC-1.1, AC-1.2, AC-1.4, AC-1.11, D-6, D-7 |
| T-3 | Swift | `macos/ProvenenciaTests/EvidenceGraphSnapshotTests.swift` `anEventsDateLineReadsItsDateObservations` *(new)* | Event with positive `date` 12 May 1842 → `DateRowDisplay.line(date:)` of it; event with `start_date` and `end_date` → the span; event with only a negative `date` → ""; person with a `date` Observation → "" | AC-2.1, AC-2.2, AC-2.3, D-9 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aRowCarriesItsEventsDate` *(new)* | `flow.load` with a `SubjectFact` of kind `.event`, `date` "12 May 1842", once anchored and once proposed → `flow.row(id)?.date` "12 May 1842" both times | AC-2.4, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteEvidenceBoardTests` (the `build` helper gains a defaulted parameter), `PromoteFlowTests` and `PromoteModelTests` (`Row` and `SubjectFact` gain defaulted fields), and `scripts/check-localizable-xcstrings.py`.

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
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc and skill updated: claim-evidence.md; kinship-shapes.md (§ On the evidence sheet); add-match-rule SKILL.md step 6
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project. On Source "Tree", add Ann Ames and Bob Ames and a marriage dated 12 May 1842 with both as `subject`; promote all three. On Source "Register", add Ann Ames and Bob Ames in a `spouse` relationship.
2. Open Promote for "Register". Bob's row targets the tree Bob, Ann's the tree Ann.
3. Open Bob's evidence sheet: the Card "…they share a spouse (…)" shows under its header "This source: recorded as spouses · The record: married in the marriage on 12 May 1842" (AC-1.12). With VoiceOver on, the Card's box reads the header, both sides and the outcome.
4. Press Done without unchecking. `identity_claim_evidence` for Bob's new claim holds both tree marriage participations' Observations.
5. Open Promote for a Source whose Bob and Ann are recorded as spouses on a tree that also records them directly as spouses: the Card shows no caption (AC-1.4).
6. On Source "Wedding", add Ann Ames and Bob Ames as `subject`s of a marriage dated 12 May 1842; with the tree recording them directly as spouses, Bob's sheet reads "This source: married in the marriage on 12 May 1842 · The record: recorded as spouses".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Source pivot is not a page row, so its date is unknown | The no-pivot phrase (AC-1.6); the Source graph is not read again (Out) |
| The caption crowds the header at narrow widths | It wraps on its own line (D-8); the pin box and badge keep their places |
| A new pattern ships without wording | The exhaustive `switch` in `phrase` stops the build (Next adds) |

**Stop and report if:**

- `CatalogClaimEvidenceRelationship.layer` / `.canon` or `CatalogClaimEvidenceShape` no longer have the fields §4 lists, or `KinshipPattern` has a case §4 does not list.
- A Source graph's `CatalogObservation.date` is nil for a `date` Observation that the Events list shows with a date (the snapshot does not carry structured dates).
- A layer shape's `pivotID` is not a page row's `subjectID` in the manual check's step 6.
- `EvidenceBoard.build` has a caller other than `PromoteModel.board(of:)` and the tests.
- Adding `Row.date` or `SubjectFact.date` needs any existing test's expected value changed.

## 12. Open questions

None.

## 13. PR

**Title:** S11-42 — Implied links on the evidence sheet
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-42-implied-links-sheet.md`.
