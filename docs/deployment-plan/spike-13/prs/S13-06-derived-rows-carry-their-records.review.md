# Review — S13-06 Derived rows carry their records

**Verdict:** round 1 Fits with changes — Not ready; every finding answered below. Round 2 pending (round 1 had blockers and reshaping fixes).
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet, the implementing model). Brief S13-D12 reviewed in the same pass.

## To user
- None. The open item, the brief's LW-4 grouping option, is the board's call (plan §12).

## Writer's own calls
- **The Why appears now, with a brief.** You chose "Why shows now", so S13-06 is gated on a new brief, S13-D12, and stays `Blocked — on board S13-D12`.
- **Group line copy** (D-13): the owner's title as every list titles it, then its ref ("Birth of Ada Lovelace · EVT-4KX2M"). Two birth events of one person often share a title, and only the ref tells them apart. The board may reword it.
- **Ink rule** (D-12): records of owners the row does not show are secondary. When no owner is the shown one, all records are secondary.
- **One shown Place** (D-10): the place row's open arrow now opens the shown Place (the first one with a name) instead of the first Place in the list. This fixes a mismatch the review found.
- **Owner order**: the shown owner first; then Go's order, which is `lifeBefore` for events and fold order per event for Places.

## Round 1 — maintainability

### M-1 · blocker · Group-line titles bypass the one Event and Place title mechanisms
**Answer:** Fixed — D-13 titles with `EventTitleDisplay.title` (the wire's `ConclusionOwnerField.event` carries the `EventHeader`, from one `EventsByIDs` call, D-6, D-7) and `PlaceTitleDisplay.titleSource`. Brief LW-4 and LW-7 examples updated.

### M-2 · blocker · `ownedField` is a second copy of `ForEntity`'s sequence
**Answer:** Fixed — D-5 splits `ForEntity` into `fieldsOf` (plus `finish`); `ForEntity` and `PersonDerivedRows` both call it. `ownedField` is gone.

### M-3 · blocker · One wire field per row; the map names the next row
**Answer:** Fixed — D-6, D-7, D-9: a tagged `repeated DerivedRow { DerivedRowKind kind; owners }` in Go, proto and Swift. The Event place row becomes one more kind.

### M-4 · blocker · The shown owner is decided in three places
**Answer:** Fixed — D-10 adds `DerivedPlace.shown`, which `name`, `extra`, the open arrow and the owner ordering use. The date row matches `header.event.id`. AC-2.5 and T-5 cover a nameless first Place.

### M-5 · major · `lifeFrom`'s date rule has a copy in `eventWhen`
**Answer:** Fixed — D-2 adds `eventDate`, called by both.

### M-6 · major · `spanning` already maps several fields to records
**Answer:** Fixed — D-11's shared `records(_:locale:orders:)`, used by `init(field:)`, `spanning` and `derived`.

### M-7 · major · `FieldOwner` carries a key that is re-resolved by origin
**Answer:** Fixed — `FieldOwner.PropertyID` is the id the value was read from (D-4). The origin check is gone. This also answers X-4.

### M-8 · major · One SQL round per owner
**Answer:** Fixed — D-5 batches `sqlPropertyIDs` and `sqlOutcomes` with `IN`. AC-3.3 and T-8 guard the graph query count.

### M-9 · major · The group-line branch loses its caller after S13-11
**Answer:** Fixed — STEP-9 adds the removal (or move into the sheet) to S13-11's scope entry. The brief's handoff tells S13-D2 to keep the grouping.

### M-10 · major · The brief invents an Event/Place title format
**Answer:** Fixed — the brief's LW-4, LW-7 and domain facts use list titles.

### M-11 · minor · Docs left unchanged
**Answer:** Fixed — STEP-9 updates `macos-client-patterns.md` and `conclusion-reconciliation.md` §9. DoD line updated.

### M-12 · minor · Three names for one concept
**Answer:** Fixed — `DerivedRow`, `DerivedRowKind` and `OwnerField` in every layer (`ConclusionOwnerField`, `CatalogOwnerField`).

### M-13 · minor · Grouping re-derived in the view
**Answer:** Fixed — D-11: `ReconciliationRecordGroup` built by the row model; the view iterates groups (D-14).

### M-14 · minor · "Rules this plan bends" out of date
**Answer:** Fixed — the rule was updated in the planning commit, and §5 now says so (same as X-1).

### M-15 · minor · The Event page picks its date property in Swift
**Answer:** Logged — §3 Out. The Swift choice predates this PR, and this PR does not widen it.

## Round 1 — executability

### X-1 · blocker · STEP-8's working-rule edit already done
**Answer:** Fixed — §5 and the Docs step no longer edit the rule.

### X-2 · major · STEP-7's check is wrong and misses `CatalogLifeFacts.event`
**Answer:** Fixed — STEP-8 lists `CatalogLifeFacts.event`; the check is `owns[^.]*Why|Why stays on the Event or Place` → no hits.

### X-3 · minor · Planned commit is not HEAD
**Answer:** Declined — HEAD is the plan commit on top of `705cd98`; no code file differs.

### X-4 · blocker · Origin filter has no data
**Answer:** Fixed by M-7 (property ids, no origin check).

### X-5 · major · "Narrowed to one field" unspecified
**Answer:** Fixed by M-2: D-5 specifies `fieldsOf`; D-6 says how each owner selects its field. `TestForEntityQueryCountIsConstant` stays.

### X-6 · major · Place order, fold and dedup underdetermined
**Answer:** Fixed — D-1 and D-4 give the algorithm: for each event in `lifeBefore` order, call `eventPlaces` and append leaves not yet seen by id. The place graph is built as `composePersons` builds it.

### X-7 · major · Event group-line text does not match the examples
**Answer:** Fixed by M-1 (list titles). T-7 builds the header with the subject rule.

### X-8 · major · Spoken label cannot be built
**Answer:** Fixed — `OwnerLine { id, text, spoken }` (D-11, D-13).

### X-9 · major · Fixtures not specified
**Answer:** Fixed — T-1 names the `placechains_test.go` helpers and the parent link (`placeRel(…, KeyPartOf, …)`), the `start_date` input, and the non-folding pair. T-3 lists its FFI calls. T-8 copies the T-1 helpers.

### X-10 · minor · Edges left to guess
**Answer:** Fixed — merged and nil nodes are handled as `onePerson` handles them (D-1, D-6); `OwnerField.Entity` comes from `FieldOwner` (no second `entityOf`); `derivedRowsProto` is named (D-8); an absent header leaves `derived_rows` empty (D-8).

### X-11 · minor · Swift edges
**Answer:** Fixed — `derived` takes `locale` and `orders` (D-12). STEP-6 says the existing `records.isEmpty` expectations stay and why.

### X-12 · minor · Scope notes
**Answer:** Fixed — the unreachable "no outcomes" skip is now "owner dropped when `fieldsOf` returns no field with its id" (D-6). `swift_vocabulary_test.go` checks string vocabularies; a proto enum needs no parity row (D-7).

### X-13 · major · Brief examples conflict with the data
**Answer:** Fixed with M-1 and M-10.

### X-14 · major · Wide frames for 2, 4 and 5
**Answer:** Fixed — the brief's §5 says the Wide frames of 1 and 3 stand for 2, 4 and 5, since nothing moves at Wide.

### X-15 · minor · Inventory vs requirements
**Answer:** Fixed with M-1 (LW-7 copy). The rest needs no change: the group row's style is named, and the S13-D1 references resolve.
