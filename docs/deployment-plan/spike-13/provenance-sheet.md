# The provenance sheet as a kit recipe: where the boundary sits

Spike 13 feature note for brief [S13-D13](design/archive/S13-D13-provenance-sheet-in-the-kit.md) and PR S13-36. Written 2026-11-03 against `main` @ `0e20544`. It answers one question (user, 2026-11-03): should a caller hand the sheet raw values for the sheet to assemble into citation rows, or hand it rows already assembled?

**Answer.** The sheet does the assembly. The kit ships three things:
1. a **view** that draws display-ready content;
2. an **assembler** that builds that content from the wire types every caller already holds;
3. a **link callback** that hands each link back to the caller as wire data.

A caller chooses *which* records (a field, a span of fields, a derived row's owners, one record), gives the sheet a title, and decides where a link goes. It formats nothing. The content type stays public so that a caller whose records are not reconciler outcomes, such as a future relationship's provenance, can fill the view without the assembler.

## 1. Callers

| Caller | Shows | Data it holds |
| --- | --- | --- |
| Person · Vitals rows (today, S13-11) | One field's records; a life row's records grouped by owning event or place | `CatalogConclusionDetail.fields`, `.derivedRows` |
| Person · likeness viewer (S13-15) | One record row on its own, under the image | Same detail |
| Person · other sections (Family, Timeline, Sources) | A record or a field, when R-13b is revisited | Same detail for the person; a relative's or event's own detail for theirs |
| Event and Place pages (their redesign) | Their fields' records, as Vitals | Their `CatalogConclusionDetail` (the same message) |
| Family tree, other views (later) | A person's fact; possibly a relationship | A relative's detail; for a relationship, the family read's "what it rests on", which is **not** a reconciler outcome |

## 2. Where the data comes from

Every caller except a future relationship reads the same wire shape: `CatalogConclusionField`, holding `outcomes: [CatalogReconcilerOutcome]`, `displayedValues`, `state` and `propertyKey`, plus `CatalogDerivedRow.owners: [CatalogOwnerField]` for life rows. It all comes from one RPC, `GetConclusionDetail`, cached per entity in `WorkspaceSession` under `CatalogQueryKey.conclusionDetail` (`macos/App/Features/Workspace/Session/CatalogQueryKey.swift`) and refreshed by the touched-handles rule (`macos/App/Platform/GenealogyStore.swift` `CatalogConclusionDetail`).

- **Person, Event and Place pages:** the page has already loaded its detail. Opening a sheet costs **no fetch**.
- **Gallery and viewer:** the Person page's detail. No fetch.
- **A family tree, for a relative's fact:** that relative's detail, read through the same key when the sheet opens. **One cached fetch**, invalidated by the existing rule. The kit never fetches: the caller warms the key and hands the sheet what it read (working rule "Each section reads its own query").
- **A relationship's provenance:** not reconciler outcomes. A new read and a new record kind, when it is designed. The content seam (§ 4) is where it plugs in.

So the inputs are uniform, already on the client, and of a type the kit may read: recipes may read `Catalog*` types and `Platform/` formatters ([`docs/design-system-layers.md`](../../design-system-layers.md)).

## 3. What assembling a record means today

Before S13-36 all of it lived in `macos/App/Features/Conclusions/`, across about 800 lines in four files (not all of it provenance). The Code column names those homes. S13-36 moved them: `ReconciliationRecord` is deleted, and `Record(outcome:in:)`, `valueGroups`, `pooledValueGroups`, `derived`, `spanning`, `ownerLine` and the summary join are in `PVProvenanceSheetContent`; the header is `PVProvenanceSheetContent.Shown(field:)`; `ReconciliationOutcome` is `PVProvenanceOutcome`; `ProvenanceSheet` is `PVProvenanceSheet`, with `ProvenanceSheetItem` left in Features to follow a refresh; the record wording stayed in `ReconciledValueDisplay` and `CitationLineDisplay` in `Platform/`.

| Step | Code | Depends on |
| --- | --- | --- |
| A record from an outcome: Read as, the "not …" phrase, counted, outcome phrase and mark | `ReconciliationRecord.init(outcome:in:)`; `ReconciledValueDisplay.readAs`, `readAsIsPhrase`, `counted`, `outcomePhrase`, `outcomeMark`, `weakCauses`, `deniedSomething` | the field's state and values; `Platform/` value formatters (`DateRowDisplay`, `NameValueDisplay`, `PropertyTermDisplay`) |
| The citation line | `CitationLineDisplay.line` | the outcome only |
| Value groups: runs by `groupRank`, when headings show, the "other records" bucket | `ReconciledValueRowModel.valueGroups`, `pooledValueGroups` | the field |
| Owner groups: the shown owner first, its line titled as its list titles it, records of other owners in secondary ink | `ReconciledValueRowModel.derived`, `ownerLine` | `CatalogOwnerField`; `EventTitleDisplay`, `PlaceTitleDisplay` (both in Features today) |
| Header: value lines and faces, the state Badge, the summary | `ReconciledValueRowModel.init(field:)`, `spanning`, `derived`, `joinSummary`; `ReconciledValueDisplay.stateBadge`, `spokenState`, `againstLine`, `sourceCount` | the field |
| Link destinations | `ReconciliationRecord.composerLocation(for:)`, `sourceLocation`, `citationLocation` (S13-36 moved them to `WorkspaceLocation.composer(for:focusingObservation:)` and `.provenance(_:_:)`) | `WorkspaceLocation` (a Features type) |

Only the last row depends on where the sheet sits. Everything above it is the same for every caller: it is how the app speaks provenance. Four rules carry decisions that were designed and tested:
- counted ink;
- headings only over two or more buckets (S13-D2 board note);
- the shown owner first and in primary ink (S13-D12);
- "not …" phrases for negative records.

## 4. The options

| | A. Caller pre-assembles rows | B. Sheet takes raw wire only | **C. Assembler and view, with an open content seam (chosen)** |
| --- | --- | --- | --- |
| Caller's work | Build every record, group and header line | Pass fields, owners, shown owner, title | Pass fields, owners, shown owner, title and the `Shown` value line (one call); or build content directly for a new record kind |
| Where the rules live | In every caller, or in a Features helper that a family tree in another feature must reach into | The kit | The kit (assembler), next to the view that draws them |
| Drift risk | High: each page words provenance its own way | None | None |
| A relationship's provenance (not an outcome) | Fits | Does not fit: the view only accepts outcomes | Fits: it builds `PVProvenanceSheetContent` itself |
| Testing | Per caller | Needs wire fixtures for every view test | Assembler: pure, table-driven Swift tests from wire fixtures. View: previews from content |
| Precedent | — | — | `PVPersonCardContent` with `init(summary: CatalogPersonSummary)` (S13-35): content in the recipe, built from a `Catalog*` type |

A spreads one rulebook across pages; it is the drift the kit is meant to stop. B is right about where the rules live, but it shuts out the one known caller whose records are not outcomes, and it makes the view untestable without wire data. **C keeps B's single rulebook and A's freedom.** The assembler is the normal way in, and content built by hand is the exception.

## 5. The boundary

| Piece | Owner | Home | Notes |
| --- | --- | --- | --- |
| `PVProvenanceRecordRow`, `PVProvenanceRecordTable`, `PVProvenanceSheet` (views, § 5.1) | Kit recipe | `DesignSystem/Recipes/ProvenanceSheet/` | Three levels a caller mixes: one record, a table of records, the whole sheet. Each draws its own piece of content; know nothing of wire types; ids from a prefix the caller passes |
| `PVProvenanceSheetContent` and its `Shown` (the value line, its face and Badge), `Group`, `OwnerLine`, `ValueGroup`, `Record` | Kit recipe | same folder | Display-ready: strings, faces, Badge, ink. Each `Record` keeps its `CatalogReconcilerOutcome`, a wire type the kit may read, so a link can hand it back |
| The assembler: `PVProvenanceSheetContent.Record(outcome:in field:)` for one record; `PVProvenanceSheetContent.init(field:title:shown:…)`, `.spanning(fields:…)`, `.derived(owners:shownOwnerID:kind:…)` for a sheet (no `.narrowed`: nothing calls it, S13-36 D-10) | Kit recipe | same folder | Today's `ReconciledValueRowModel` provenance logic and `ReconciliationRecord.init`, moved. Pure; tested |
| Record and state wording: `ReconciledValueDisplay`'s record and state functions, `CitationLineDisplay` | Platform | `Platform/` | Value formatting, as S13-34 moved names, dates and terms. Vitals rows keep calling them |
| Owner titles: `EventTitleDisplay`, `PlaceTitleDisplay` | Platform | `Platform/` | The owner line titles an event or place as its list does; S13-34 moved `PersonTitleDisplay` the same way |
| Link callback: `open(_ record: PVProvenanceSheetContent.Record, _ link: PVProvenanceLink)` with `.source`, `.citation`, `.observation` | Kit → caller | — | The record carries its outcome. The caller maps it to a `WorkspaceLocation` (`WorkspaceLocation.provenance(_:_:)` and `.composer(for:focusingObservation:)`, in Features), and the sheet closes itself first (`\.pvSheetClose`). The kit names no navigation type. A later non-outcome record changes `Record.outcome` and that one mapping |
| Choosing the records and the title; presenting; following a refresh (`ProvenanceSheetItem`, `row(in:)`) | Caller | `Features/…` | A Vitals row, the viewer, a tree node |
| The provenance control (`PVIconButton(.info)`) | Caller | — | As today |

### 5.1 Three levels, mixed freely (user, 2026-11-03)

A caller uses whichever level it needs, and every level is built from the level below. Nothing is drawn twice.

| Level | View | Takes | Typical caller |
| --- | --- | --- | --- |
| One record | `PVProvenanceRecordRow` | a `Record`, a face for Read as, `PVProvenanceColumns`, an `idPrefix`, the link callback | The viewer's row under the image (S13-15); any view that cites one record inline (a tree node's detail, a Timeline event) |
| A table of records | `PVProvenanceRecordTable` | `[Group]` (owner lines, value headings, records), the face, `PVProvenanceColumns`, an `idPrefix`, the link callback; a caller places `PVProvenanceColumnHeadings` above it | A section that lists several records without a modal |
| The sheet | `PVProvenanceSheet` | `PVProvenanceSheetContent` (title, `Shown`, summary, groups), `PVProvenanceColumns` (which carry the width class), an `idPrefix`, the link callback, Done | Vitals, an Event or Place field, a tree node |

- **The row stands alone.** It draws one record's cells (Read as, citation line with its copy button, outcome, refs, links) with no Panel, heading row or group line, and it is focusable and readable by VoiceOver on its own. The table is rows plus group lines; the sheet is a Panel, a header and a table.
- **`PVProvenanceColumns`** *(new)* is one small value shared by the row and the table: whether Read as shows (off, the record's spoken line moves to the row's leading edge, ahead of the citation and the links; the reserved column always shows) and the width class that sets the caps (PSK-W1). A lone row and a table built from the same value line up, so a caller can stack rows of its own beside a table.
- **One link callback for all three:** `(PVProvenanceSheetContent.Record, PVProvenanceLink) -> Void`. A caller wires navigation once, whatever level it places.
- **The assembler mirrors the levels.** `Record(outcome:in:)` builds one row's content, and the group and sheet builders are made of it. A caller that wants one row does not build a sheet and pick it apart.

**Build on open.** The caller builds the content when the sheet opens, and on each redraw while it is open, from the detail it already holds. It is not built for every row on every page load. The sheet then follows a refresh for free, and nothing carries a precomputed sheet. Vitals still needs to know whether a row has records, to show its control: a count, not a sheet.

**The Event and Place Why accordion** keeps its own record table (`ReconciliationRecordRows`, a `Grid`) until those pages move onto the sheet (user, 2026-11-03). It reads the same `PVProvenanceSheetContent` from the kit assembler and draws the recipe's parts (`PVProvenanceOwnerLine`, `PVProvenanceOutcome`, `PVProvenanceReadAs`), so only the table layout is the Why's, never the rules or the cells.

## 6. Consequences for the plans

- **S13-36** moves more than the view: the assembler and the `Platform/` formatters too. By `plan-pr`'s size rule it may be two PRs:
  1. the wording and titles move to `Platform/`, no visible change;
  2. the assembler, the view and Vitals move onto the recipe.

  The S13-36 planner decides the cut and updates the map.
- **S13-14** uses no provenance piece: tiles have no provenance control (user, 2026-11-03).
- **S13-15** places one `PVProvenanceRecordRow` under the image, built with `Record(outcome:in:)` from the likeness's record (its D-9). It is refreshed after S13-36.
- **Event and Place, Family, Timeline, a family tree:** each adds one call to the assembler and one link mapping.
