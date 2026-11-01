# S13-02 — Gender in the person header

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 3 — Family (the header terms feed the Family and Timeline tiles, S13-12 and S13-13)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Person page's badge and thumb read the same terms from a different input and look the same)
**Board:** —
**Depends on:** S11-30 (landed [#509](https://github.com/mendahu/provenencia/pull/509))
**Blocks:** S13-04, S13-05, S13-12, S13-13
**Branch:** `s13-02-gender-in-the-person-header`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-01 and S13-06 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md](../person-detail.md) R-3, R-21; working rules: "One person header for every tile", "Colors reuse `PersonGenderTreatment`"; [`docs/deployment-plan/archive/spike-11/README.md`](../../archive/spike-11/README.md) ("Gender meets sex at birth through a registry crosswalk")
**Planned:** 2026-11-01 against `main` @ `fdd37f19`
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

Every place that shows a person as a row or a tile can color it by gender, from the one header type every tile carries. `PersonHeader` gains the person's displayed gender and sex at birth terms, and `PersonGenderTreatment` reads them from the header instead of from the Person page's detail fields. The Person page header moves onto the new input; its badge, its spoken label and its thumb colors stay exactly as they are.

**Why now.** The Family (S13-12) and Timeline (S13-13) tiles color every relative and participant per R-3, and they carry only `PersonHeader`s (working rule "One person header for every tile"). S13-04 and S13-05 return those headers. The terms have to be on the header before those reads are planned. S11-30's crosswalk landed in #509, so the gender and sex at birth terms are final.

**What stays.** The treatment's rules (gender wins; sex at birth stands in only when no gender is displayed; a disagreement or a non-product term is neutral or "other"), its colors, the badge text, its accessibility label, and the Vitals rows for Gender and Sex at birth. The Persons list, Promote and search hits get the terms on the wire and do not draw them.

## 2. User stories

**US-1.** As the engine, when I compose a Person's header, I carry its displayed gender and sex at birth terms and how many values each displays, so every surface that shows the person can color it without reading the detail.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Person whose one member records gender `woman` (product term) | `PersonHeader.Gender` = {Term {Key `woman`, Origin `provenencia`, Label `Woman`}, Count 1, PropertyLabel `Gender`}; `SexAtBirth.Count` = 0, `SexAtBirth.Term` nil, `SexAtBirth.PropertyLabel` `Sex at birth` |
| AC-1.2 | Two members of one Person, from two Sources, record gender `woman` and `man` | `Gender.Count` = 2; `Gender.Term` is the rank-1 kept term |
| AC-1.3 | A Person with only a provisional member recording gender `woman` | `Gender.Count` = 0, `Gender.Term` nil |
| AC-1.4 | A researcher's own gender term "Agender" (origin `user`, on the product `gender` Property) is recorded, the header is composed and stored, then the term is relabeled "No gender" | The next `PersonsByIDs` returns `Gender.Term.Label` "No gender" and `Gender.Term.Origin` `user`. Product terms cannot be relabeled (`propertyterms.Update` returns `ErrLocked`) |
| AC-1.5 | The product Property `gender` is relabeled "Gender identity" | The next read returns `Gender.PropertyLabel` "Gender identity" |
| AC-1.6 | 51 Persons, each with a gender Observation, read cold | `ListPersonsQueryCount` equals the count for 1 Person |
| AC-1.7 | `ListPersonHeaders`, `GetConclusionDetail`, Promote alignment and `SearchCatalog` | Each `engine.PersonHeader` carries `gender` and `sex_at_birth` as `TermFacts` (one mapper, `personHeaderProto`) |

**US-2.** As a researcher, when I open a Person page, I see the same gender badge, spoken label and thumb colors as before, now read from the person header.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Header gender {`woman`, product, "Woman", count 1}, sex at birth {`female`, count 1} | Family `.woman`, basis `.gender`, badge "Woman", spoken "Gender, Woman" |
| AC-2.2 | Header gender count 0, sex at birth {`male`, product, "Male", count 1} | Family `.man`, basis `.sexAtBirth`, badge "Male · at birth", spoken "Sex at birth, Male" |
| AC-2.3 | Header gender count 2 (mixed), sex at birth {`female`, count 1} | Treatment nil (no fallback to sex at birth) |
| AC-2.4 | Header gender {`woman`, origin `user`, count 1} | Family `.other` |
| AC-2.5 | Header gender {`two_spirit`, product, count 1} | Family `.other`; sex at birth `intersex` alone → `.other`; gender `non_binary` → `.nonBinary` |
| AC-2.6 | Header gender count 1 with no term; header gender {key "", label ""} count 1; both counts 0; no header | Treatment nil |
| AC-2.7 | `PersonDetailContent` on a detail whose header carries gender `woman` | `treatment?.family == .woman`, `badge?.text == "Woman"`, thumb colors set; the Vitals rows still include "Gender" from the detail fields |
| AC-2.8 | `PersonDetailContent` on a detail whose fields carry gender `woman` but whose header carries none | Treatment nil: the fields no longer feed it |

## 3. Scope

| In | Covered by |
| --- | --- |
| `conclusionheaders.PersonHeader` gains `Gender` and `SexAtBirth` as `TermFacts` *(new)*; `EventType` struct renamed `Term` and shared | STEP-1, STEP-2, AC-1.1 – AC-1.6 |
| Term label and origin, and the Property label, resolved at read time (`finishPerson`), as `finishEvent` does | STEP-2, AC-1.4, AC-1.5 |
| Proto `TermFacts` *(new)*; `PersonHeader.gender = 9`, `sex_at_birth = 10`; `personHeaderProto` fills them | STEP-3, AC-1.7 |
| Swift `CatalogTermFacts` *(new)* on `CatalogPersonHeader`; `GoStore.mapPersonHeader` maps it | STEP-4, AC-1.7 |
| `PersonGenderTreatment.read(_ header:)` replaces `read(_ fields:)`; the Person page header reads it; the field read is deleted | STEP-5, AC-2.1 – AC-2.8 |
| Docs: package and proto comments, the treatment's doc, `person-detail.md`'s S11-30 status, `completed.md`, the map row | STEP-6 |

| Out | Home |
| --- | --- |
| Any tile chrome: the Family and Timeline person tile | S13-12, S13-13 |
| Coloring the Persons list rows, Promote rows or omnibar hits by gender | not planned: each is a chrome change that needs its own brief |
| `GetPersonFamily` and `ListPersonEvents` reading the headers | S13-04, S13-05 |
| Gender on `EventSubject` (the Event header's subject list) | not planned: the Timeline reads participants as `PersonHeader`s (working rule "Timeline is direct participation, any role") |
| Localizing seeded Property labels ("Gender", "Sex at birth") | not planned: the Vitals row label and the spoken label both stay the catalog label |
| FakeStore composing gender into its person headers | not planned: no FakeStore detail or UI test carries a gender field today (`rg gender macos/App/Platform/FakeStore.swift` is empty) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Header type | `core/database/conclusionheaders/conclusionheaders.go` `PersonHeader` | Entity, Name, NameValueCount, Birth, Death (`LifeFacts`), PictureImageFileID, Updated. No gender or sex at birth |
| Term type | `core/database/conclusionheaders/events.go` `EventType` | `{ID, Key, Origin, Label}`; used only as `EventHeader.EventType` (`events.go:30`) and built at `compose.go:828`. No other package names `conclusionheaders.EventType` (`rg 'conclusionheaders\.EventType\b'` is empty) |
| Compose | `core/database/conclusionheaders/compose.go` `onePerson` | Builds Name, Birth, Death off the cached `graphcache.Node`. `keptTerm(n, propID)` returns the rank-1 kept term's id and key and the kept count; used for event type only |
| Finish step | `compose.go` `finishEvent`, called from `composeEvents`' output loop for fresh and cached rows | Fills `EventType.Label` and `.Origin` from `vocab.termLabel` / `termOrigin` at read time, so a term rename does not drop the stored row. `composePersons` has no finish step |
| Vocabulary | `compose.go` `vocab`, `loadVocab`, `scanVocab` | `propID` (key → id) is read with `SELECT id, key FROM properties WHERE origin = ?`; no Property labels. The map is held by `g.Labels()` and dropped by `graphcache.OnCommit` on `effects.ListVocabulary`, which a `properties` or `property_terms` write raises (`core/database/effects/registry.go` `"properties"`) |
| Kept semantics | `core/database/graphcache/graph.go` `Node.Kept`, `Rank1`; `core/autoreconcile/autoreconcile.go` `Result.Values` | Kept = reason `kept`. Displayed values come first, so one kept value is rank 1. Swift `CatalogConclusionField.displayedValues` is the same filter (`isDisplayed` = `.kept`) |
| Vocab keys | `core/vocabkey/vocabkey.go` `PropertyGender`, `PropertySexAtBirth` | `"gender"`, `"sex_at_birth"` |
| Wire | `api/proto/engine.proto` `message PersonHeader` (fields 1 – 8), `message ConclusionTerm {id, key, label, origin}` | No gender |
| Mapper | `api/ffi/handlers/conclusion_headers.go` `personHeaderProto` | The one Go → proto mapper; called by `ListPersonHeaders`, `conclusion_detail.go` `attachHeader`, `promote_align.go`, `search_catalog.go` |
| Swift model | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonHeader`, `CatalogLifeFacts` | No gender |
| Swift mapper | `macos/App/Platform/GoStore.swift` `mapPersonHeader` | Maps fields 1 – 8 |
| Treatment | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` `PersonGenderTreatment.read(_ fields: [CatalogConclusionField])` | Picks `fields.first(where: propertyKey == gender)`; if it has displayed values, treats it (exactly one displayed `.term`, non-empty text), else falls back to sex at birth. `fieldLabel` = the field's catalog label. Matches on `propertyKey` only, so a researcher's own Property keyed `gender` can be the one picked |
| Page | `macos/App/Features/Conclusions/PersonDetailContent.swift` `init(detail:…)` | `treatment = PersonGenderTreatment.read(detail.fields)`; `header = detail.header?.person` already feeds the life lines |
| Detail header | `api/ffi/handlers/conclusion_detail.go` `attachHeader` | A product Person's detail always carries its `PersonHeader` (one `PersonsByIDs` call) |
| Tests | `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonsByIDs`, `TestHeadersShowKeptValuesOnly`, `TestListPersonsQueryCountIsConstant`; fixtures `person`, `subjectAndHandle`, `prop`, `term`, and `cite` in `walks_test.go` | No gender case |
| Tests | `api/ffi/handlers/conclusion_headers_test.go` `TestListPersonHeaders` (package `handlers`) | No gender case |
| Tests | `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift` (all six tests build `CatalogConclusionField`s); `macos/ProvenenciaTests/PersonDetailTests.swift` `treatmentRidesOnContentAndRowsStay`, `headerThumbIsTheLeadLikeness` (put gender in `extra:` fields) | All read fields |

**Precedent.** `17223d66` ("Lead a Person header with the earliest likeness.") added one field to `PersonHeader`, filled in `onePerson`, with a `conclusionheaders_test.go` case: copy that shape for the Go header. `finishEvent` and `EventHeader.EventType` (`ConclusionTerm` on the wire, mapped in `eventHeaderProto`) are the precedent for a term read at read time: copy its split (id and key stored, label and origin filled on read). Do not copy `EventHeader`'s flat `event_type` + `event_type_count` wire pair: the treatment also needs the Property label (D-3).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The header carries both terms raw: per Property, the rank-1 kept term (nil when none, or when the rank-1 kept value has no term key) and the count of kept values. Go does not pick gender over sex at birth and does not choose a color family | Go returns structures; the app formats (package doc). The treatment's rules and colors are Swift-owned since S10-23 and stay in one place, `PersonGenderTreatment` | A Go-chosen `basis`/`family` enum — moves a display rule into the engine and adds a wire enum; one kept term only — loses "mixed" (count > 1), which must stay neutral (AC-2.3) |
| D-2 | One shape for both: `TermFacts {Term *Term; Count int; PropertyLabel string}` *(new)* in `conclusionheaders`, mirrored by proto `TermFacts` and Swift `CatalogTermFacts`. `EventType` is renamed `Term` and `EventHeader.EventType` becomes `*Term` | Two Properties, one shape, next to `LifeFacts`; a term struct named for events would be wrong on a person. The rename touches one package only (§4) | Two more flat fields per Property on `PersonHeader` — six fields that differ only by prefix; a second `GenderTerm` struct beside `EventType` — a near-copy |
| D-3 | `TermFacts.PropertyLabel` is the product Property's catalog label, loaded into `vocab` (new `propLabel` map keyed by property id) by the existing properties query, now `SELECT id, key, label …` read by a dedicated `scanProperties` (as `scanTerms` reads two columns) | The badge's spoken label is "<field label>, <value>" today, and a researcher may relabel a product Property (`properties.Update` allows origin `provenencia`). Reading it from the catalog keeps the spoken label equal to the Vitals row label. Same query, no new round trip | Swift L10n constants "Gender" / "Sex at birth" — diverge from the Vitals row after a relabel; the property label on every `ConclusionTerm` — widens a shared message for one caller |
| D-4 | `finishPerson(v, h)` *(new)* fills `Term.Label`, `Term.Origin` and `PropertyLabel` at read time, called in `composePersons`' output loop on fresh and cached rows. The stored display row keeps `Term.ID` and `Term.Key` only. `finishEvent` and `finishPerson` share one helper, `vocab.finishTerm(*Term) *Term` *(new)* | A term or Property relabel raises `ListVocabulary`, which drops `vocab` but must not be the reason a stored header is wrong; this is `finishEvent`'s existing rule (`vocab`'s doc: term labels are not stored on the row, so a rename does not throw the row away) | Storing labels on the row — a relabel would show the old label until the person's node is replaced |
| D-5 | Only the **product** Properties feed the header (`v.prop(vocabkey.PropertyGender)`, `v.prop(vocabkey.PropertySexAtBirth)`, origin `provenencia`) | `vocab.propID` is product-only already; the treatment's own rule is that a key names a product term only with the product's origin. Today's `fields.first(where:)` could pick a researcher's own Property keyed `gender`; that is the latent bug this read removes (Edges) | Matching by key across origins — repeats the bug |
| D-6 | `PersonGenderTreatment.read(_ header: CatalogPersonHeader?) -> PersonGenderTreatment?` replaces `read(_ fields:)`, which is deleted. Rule, unchanged: `gender.count > 0` → treat gender (needs count 1, a `.term` value, non-empty text), else treat sex at birth the same way; nil header → nil | The map's Ships line says the treatment reads only the header's terms and the read from detail fields is deleted. Same rules as today, so AC-2.x restate the existing tests | Keeping both reads — two inputs for one treatment, which the map review (K-1) removed |
| D-7 | `CatalogTermFacts.term` is a `CatalogConclusionValue?` that `GoStore` builds only as `.term(id:key:label:origin:)` | The treatment already destructures `.term` and formats it with `ReconciledValueDisplay.string(for:)`; reusing the value keeps that one formatting path | A new Swift term struct — a third term shape beside `CatalogPropertyTerm` and `CatalogConclusionValue.term` |
| D-8 | Wire: `message TermFacts { ConclusionTerm term = 1; int32 count = 2; string property_label = 3; }`; `PersonHeader.gender = 9`, `PersonHeader.sex_at_birth = 10`. `personHeaderProto` always sets both (an empty `TermFacts` when the Property is absent from the catalog); `term` is unset when `Term` is nil | Reuses `ConclusionTerm`; additive fields, no renumbering | `optional` presence on the message — Swift reads `count`, so an always-present message is simpler |
| D-9 | Test layers: Go table-driven tests in `conclusionheaders` for the composed values, one handler test for the wire, Swift Testing for the treatment and the page content. No UI test | The behavior is a mapping and a pure rule; the page chrome is unchanged | An XCUITest of the badge — FakeStore carries no gender (Out) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `conclusionheaders.TermFacts` *(new)* | type | `conclusionheaders.LifeFacts`; `EventHeader.EventType` + `EventTypeCount` | `LifeFacts` is an event read; the event pair has no Property label and would be duplicated per Property (D-2) |
| `conclusionheaders.Term` | type (rename of `EventType`) | `conclusionheaders.EventType` | Reused, renamed so a person's term is not an "event type" (D-2) |
| `finishPerson` *(new)*, `vocab.finishTerm` *(new)* | func | `finishEvent` | `finishEvent` now calls `finishTerm`; one term-finishing path (D-4) |
| `vocab.propLabel` *(new)*, `scanProperties` *(new)* | field, func | `vocab.propID`, `scanTerms` | Same query as `propID`; `scanVocab` scans two columns only (D-3) |
| `keptTermFacts(n, propID) TermFacts` *(new)* | func | `keptTerm` | Thin wrapper over `keptTerm` that builds the struct; no second kept-term read |
| proto `TermFacts` *(new)* | message | `ConclusionTerm`, `LifeFacts` | Carries `ConclusionTerm` plus count and label (D-8) |
| `termFactsProto` *(new)* | func | `lifeFactsProto`, the `ConclusionTerm` literal in `eventHeaderProto` | Shared by both Properties; `eventHeaderProto` keeps its literal (not this PR's code) |
| `CatalogTermFacts` *(new)* | type | `CatalogLifeFacts`, `CatalogPropertyTerm` | `CatalogPropertyTerm` is a vocabulary row (description, property id), not a displayed value (D-7) |
| `GoStore.mapTermFacts` *(new)* | func | `mapLifeFacts` | Same shape as its neighbor |
| `PersonGenderTreatment.read(_ header:)` | func (replaces `read(_ fields:)`) | itself | Replaced, not beside (D-6) |

**Builds on.** The header package's split of stored row vs read-time labels (`finishEvent`), the display cache (`g.Display` / `SetDisplay`), `effects.ListVocabulary` dropping `vocab`, and the one mapper `personHeaderProto` that every Person header on the wire goes through.
**Deletes.** `PersonGenderTreatment.read(_ fields: [CatalogConclusionField])` and its private `treatment(_ field: CatalogConclusionField, basis:)`; the `EventType` type name (renamed). The field-shaped helpers `gender(_:state:)` and `sex(_:state:)` in `PersonGenderTreatmentTests.swift` are replaced by header-shaped helpers. No other caller of `read(_ fields:)` exists (`rg 'PersonGenderTreatment.read' macos/App` → `PersonDetailContent.swift:48` only).
**Next adds.** S13-04 and S13-05 return `PersonHeader`s through `conclusionheaders.ByIDs` and get the terms for free; S13-12 and S13-13 call `PersonGenderTreatment.read(header)` per tile. A later term-valued header fact (an Event's or a Place's type on another kind) is one more `TermFacts` field.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| The Person page treatment now depends on `detail.header` | `PersonDetailContent.init` | Content is rebuilt per detail, so it resets on a new id. `attachHeader` sets the header for every product Person (§4); a nil header gives no badge and default thumb colors (AC-2.6), the same as a Person with no gender today |
| Promote, search and list headers now carry the terms | `PromoteFlow`, `CatalogSearch`, Persons list Swift models | They ignore the new fields; nothing draws them (Out). `CatalogPersonHeader`'s new properties have defaults, so every existing initializer compiles unchanged |

### Scale

At the map's Scale (a Source of 100 new people on one Promote page; the 41-Subject obituary): the header read adds no query. The terms are read off the cached `graphcache.Node` (`n.Kept`), and the Property labels come in the properties query `loadVocab` already runs once per vocabulary load. AC-1.6 holds the count constant at 1 vs 51 Persons. The display cache and its invalidation do not change: gender and sex at birth are the person's own node values, so a commit that changes them replaces the node and drops its stored header, as it does for the name. The wire grows by two small messages per header.

### Edges

- **Assumptions.** The treatment assumes the one displayed value is the one it reads. The header reads `Rank1` of the kept rows; `autoreconcile.Result.Values` puts displayed values first, so a single kept value is rank 1 and the two reads agree (T-1 "one member"). A kept count of 2 with a rank-1 term still gives nil (T-1 "two Sources disagree" for the count, T-6 for the nil, AC-2.3).
- **Neighbors.** Gender displayed-empty but recorded (only `no_evidence` outcomes): count 0, so sex at birth stands in, as today (T-6 `sexAtBirthBorrowsAFamily`). Gender count 1 but the rank-1 kept row carries no term key: `Term` nil, treatment nil, as today's `.text` case (T-7). A provisional member's gender: not kept, count 0 (T-1 "provisional", AC-1.3). A researcher's own Property keyed `gender`: no longer feeds the header (D-5); a Person whose only gender is on such a Property loses its badge — today it gets one only when `fields.first(where:)` happens to pick it (T-1 "user Property").
- **Removed inputs.** A member unpromoted or its Observation deleted: the reconciler reruns, the node is replaced, the stored header drops, and the next read recomputes count and term (existing display-cache behavior, held by `display_cache_test.go`; no new mechanism). A term or Property relabel: labels are read-time (AC-1.4, AC-1.5).
- **Real data.** The seeded gender terms (`man`, `woman`, `non_binary`, others → `.other`) and sex at birth terms (`female`, `male`, `intersex`, `indeterminate`, `unknown`) are the ones `PersonGenderTreatment.family` already maps; no table-driven rule changes.
- **Untrusted size.** None — two Properties per header, each one term and a count.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/conclusionheaders/events.go` | `type EventType struct` → `type Term struct` (same fields, doc: "a term value of a header; Key names a product term only when Origin is the product's"). `EventHeader.EventType *Term` |
| `core/database/conclusionheaders/conclusionheaders.go` | `PersonHeader` gains `Gender, SexAtBirth TermFacts` (doc: "the displayed gender and sex at birth terms; the app chooses which colors a person"). `type TermFacts struct { Term *Term; Count int; PropertyLabel string }` *(new)* — doc: "Term is the rank-1 kept term, nil when none; Count is the number of kept values, more than one is a disagreement; Term label and origin and PropertyLabel are filled at read time." Package doc: unchanged except "later tree nodes" stays |
| `core/database/conclusionheaders/compose.go` | `vocab.propLabel map[string]string` *(new)*; `loadVocab` calls `scanProperties` *(new)* (`SELECT id, key, label FROM properties WHERE origin = ?`) filling `propID` and `propLabel`; `func (v vocab) finishTerm(t *Term) *Term` *(new)* (copy with Label and Origin from `termLabel` / `termOrigin`; nil in → nil out); `finishEvent` uses it; `func keptTermFacts(n *graphcache.Node, propID []byte) TermFacts` *(new)* wraps `keptTerm`; `onePerson` sets `h.Gender = keptTermFacts(n, v.prop(vocabkey.PropertyGender))` and `h.SexAtBirth` likewise; `func finishPerson(v vocab, h PersonHeader) PersonHeader` *(new)* fills both `Term`s through `finishTerm` and `PropertyLabel` from `v.propLabel[string(v.prop(key))]`; `composePersons`' output loop returns `finishPerson(v, h)` for fresh and cached rows. The stored row (`g.SetDisplay`) is the unfinished one, as for events |

Errors: none new.

### FFI

`api/proto/engine.proto`:

```proto
// TermFacts is one term-valued Property of a header: the rank-1 displayed
// term (absent when none), how many values are displayed (> 1 is a
// disagreement), and the Property's label.
message TermFacts {
  ConclusionTerm term = 1;
  int32 count = 2;
  string property_label = 3;
}

message PersonHeader {
  // … fields 1 – 8 unchanged …
  // gender and sex_at_birth are the displayed terms of the product
  // Properties; the app chooses the colors (gender first). Always set.
  TermFacts gender = 9;
  TermFacts sex_at_birth = 10;
}
```

`api/ffi/handlers/conclusion_headers.go`: `termFactsProto(f conclusionheaders.TermFacts) *engine.TermFacts` *(new)*; `personHeaderProto` sets `Gender` and `SexAtBirth`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Header terms | `CatalogTermFacts` *(new)*: `var term: CatalogConclusionValue?`, `var count = 0`, `var propertyLabel = ""` | `macos/App/Platform/GenealogyStore.swift`, beside `CatalogLifeFacts` | `CatalogPersonHeader` gains `var gender = CatalogTermFacts()`, `var sexAtBirth = CatalogTermFacts()` after `death` |
| Mapping | `GoStore.mapTermFacts` *(new)* | `macos/App/Platform/GoStore.swift` | `term: f.hasTerm ? .term(id: f.term.id, key: f.term.key, label: f.term.label, origin: f.term.origin) : nil`, `count: Int(f.count)`, `propertyLabel: f.propertyLabel`; `mapPersonHeader` passes `gender: mapTermFacts(h.gender), sexAtBirth: mapTermFacts(h.sexAtBirth)` |
| Treatment | `PersonGenderTreatment.read(_ header: CatalogPersonHeader?)` | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` | D-6. Private `treatment(_ facts: CatalogTermFacts, basis:)`: `facts.count == 1`, `case .term(_, let key, _, let origin)? = facts.term`, text = `ReconciledValueDisplay.string(for:)` non-empty, `fieldLabel: facts.propertyLabel`. `family`, colors, `badge`, `thumbnailColors` unchanged. Doc comment: "reads the Person header's terms (S13-02)" |
| Page | `PersonDetailContent` | `macos/App/Features/Conclusions/PersonDetailContent.swift` | `treatment = PersonGenderTreatment.read(header)` (the `header` local already exists) |

States: no header → no badge, default thumb. No view changes. Accessibility: the badge's spoken label is unchanged ("Gender, Woman").

### Strings

None — no key added or removed; `L10n.Conclusions.sexAtBirthBadge` and `a11yList` are reused as they are.

## 7. Steps

**STEP-1 — The header composes the terms**
Files: `core/database/conclusionheaders/events.go`, `conclusionheaders.go`, `compose.go`, `conclusionheaders_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: rename `EventType` → `Term`; add `TermFacts`, `keptTermFacts`, the two `PersonHeader` fields, and set them in `onePerson` (§6 Go core). Add the test helper `citeOn` and write T-1 (keys, counts and term presence; labels arrive in STEP-2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/...` passes, T-1 included; `go vet ./core/...` clean.

**STEP-2 — Labels are read at read time**
Files: `compose.go`, `conclusionheaders_test.go`.
Do: `propLabel` and `scanProperties` in `loadVocab`; `finishTerm`; `finishEvent` calls it; `finishPerson` applied in `composePersons`' output loop. `finishTerm` returns a new `*Term` and never writes through the pointer it is given: a cached `PersonHeader` from `g.Display` is a value copy whose `Gender.Term` still points at the stored row's `Term`. Write T-2; T-3 stays green; extend `TestListPersonsQueryCountIsConstant` (T-4).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/... ./core/database/searchindex/...` passes; `TestListEvents` still sees `EventType.Label == "Birth"`.

**STEP-3 — The wire carries the terms**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` (generated), `api/ffi/handlers/conclusion_headers.go`, `api/ffi/handlers/conclusion_headers_test.go`. Skill: `add-ffi-handler`.
Do: add `TermFacts` and fields 9, 10 (§6 FFI); run `./scripts/generate-proto.sh`; add `termFactsProto` and set both in `personHeaderProto`. Write T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; `git diff --stat` shows both generated files changed. If `./scripts/generate-proto.sh` exits because `protoc`, `protoc-gen-go` or `protoc-gen-swift` is missing, stop and report; do not hand-edit generated files.

**STEP-4 — Swift maps the header terms**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`.
Do: `CatalogTermFacts`, the two `CatalogPersonHeader` properties with defaults, `mapTermFacts`, `mapPersonHeader` (§6 macOS).
Check: `scripts/check-pr.sh` gets through the Xcode build with no new Swift warning (its Swift test stage may still fail on the old treatment tests until STEP-5).

**STEP-5 — The treatment reads only the header**
Files: `macos/App/Features/Conclusions/PersonGenderTreatment.swift`, `PersonDetailContent.swift`, `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-swift-test`.
Do: replace `read(_ fields:)` with `read(_ header:)` (D-6) and delete the field-shaped private `treatment`; `PersonDetailContent` reads `header`. Rewrite the treatment tests on headers (T-6, T-7, T-8); move the gender in `treatmentRidesOnContentAndRowsStay` and `headerThumbIsTheLeadLikeness` onto the header (T-9) and add the fields-only case (T-10).
Check: `rg -n 'read\(_ fields|CatalogConclusionField' macos/App/Features/Conclusions/PersonGenderTreatment.swift` is empty; `scripts/check-pr.sh` passes, Swift tests included.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-13/person-detail.md` (the S11-30 row says "In review"; it landed in #509).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-02-gender-in-the-person-header.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonHeaderTerms` *(new)*, table-driven; read through `PersonsByIDs`. Asserts keys, counts and term presence only (labels arrive in STEP-2). Helper `func (f *fixture) citeOn(art artifacts.Artifact, s subjects.Subject, in ...observations.Input)` *(new)*: the body of `f.cite` (`walks_test.go`) with `GraphID: evidencegraphstest.SoleGraph(f.t, f.c, art.SourceID).ID` and `ArtifactID: art.ID` | "one member": `s, _, handle := f.person("Ada")`, `f.cite(s, observations.Input{PropertyID: f.prop("gender").ID, ValueTermID: f.term("gender", "woman").ID})` → `Gender.Term.Key` `woman`, `Gender.Count` 1, `SexAtBirth.Term` nil, `SexAtBirth.Count` 0; "sex only": the same with `sex_at_birth` / `male` → `SexAtBirth.Term.Key` `male`, Count 1, `Gender.Count` 0; "two Sources disagree": after "one member", `_, art := f.book("Census")`, `s2 := subjectstest.OfType(f.t, f.c, userID, evidencegraphstest.SoleGraph(f.t, f.c, art.SourceID).ID, f.typeID("person"), "")`, `f.citeOn(art, s2, …gender man…)`, `promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s2.ID, EntityID: handle})` → `Gender.Count` 2, `Gender.Term` non-nil; "provisional": the `TestHeadersShowKeptValuesOnly` shape (create handle, Subject, `identityclaims.StatusProvisional` claim, `autoreconciler.RecomputeTx`) citing gender `woman` → Count 0, Term nil; "user Property": `properties.Create(tx, userID, "Gender", propertymeta.ValueTypeTerm, "", "")` (action `create_property`), `propertyterms.Create(tx, userID, prop.ID, "Woman", "")` (action `create_property_term`), `subjectvocab.EnsureBinding(f.c, f.typeID("person"), prop.ID, 99)` (observations need a binding: `observations.go` `sqlBindingExists`), cite that term on a promoted person → `Gender.Count` 0 | AC-1.2, AC-1.3, D-1, D-5 |
| T-2 | Go | same file, `TestPersonHeaderTermLabels` *(new)* | "labels": the T-1 "one member" Person → `Gender.Term` {Label `Woman`, Origin `provenencia`}, `Gender.PropertyLabel` `Gender`, `SexAtBirth.PropertyLabel` `Sex at birth`; "term relabel": `propertyterms.Create(tx, userID, f.prop("gender").ID, "Agender", "")` (action `create_property_term`), cite it, read once (stored), `propertyterms.Update(tx, userID, term.ID, "No gender", "")` (action `update_property_term`) → next read `Gender.Term.Label` "No gender", Origin `user`; "Property relabel": `p := f.prop("gender")`, `properties.Update(tx, userID, p.ID, "Gender identity", p.ValueType, p.Description, p.Cardinality)` (action `update_property`, called as in `core/database/autoreconciler/autoreconciler_test.go` `TestCardinalityChangeRecomputes`) → `Gender.PropertyLabel` "Gender identity" | AC-1.1, AC-1.4, AC-1.5, D-3, D-4 |
| T-3 | Go | `TestListEvents` (existing) | still passes: `EventType` {Key `birth`, Label `Birth`, Origin `provenencia`} after the rename and `finishTerm` | D-2, D-4 |
| T-4 | Go | `TestListPersonsQueryCountIsConstant` (extended) | each of the 50 added Persons also cites gender `woman` → `many == one` | AC-1.6 |
| T-5 | Go | `api/ffi/handlers/conclusion_headers_test.go` `TestPersonHeaderProtoCarriesTerms` *(new)* | `personHeaderProto` of a `PersonHeader` with `Gender` {Term {ID 16 bytes, Key `woman`, Origin `provenencia`, Label `Woman`}, Count 1, PropertyLabel `Gender`} and empty `SexAtBirth` → `gender.term` {id = `database.UUIDString(ID)`, key, label, origin}, `gender.count` 1, `gender.property_label` `Gender`; `sex_at_birth` set, `term` unset, count 0 | AC-1.7, D-8 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift` (rewritten on `CatalogPersonHeader`; helpers `private func facts(_ key: String, _ label: String, origin: String = CatalogOrigin.provenencia, count: Int = 1, propertyLabel: String) -> CatalogTermFacts` and `private func header(gender: CatalogTermFacts = CatalogTermFacts(), sex: CatalogTermFacts = CatalogTermFacts()) -> CatalogPersonHeader` replace `gender(_:state:)` and `sex(_:state:)`; a disagreement is `facts(…, count: 2, …)`; a count-1 facts with no term is `CatalogTermFacts(term: nil, count: 1, propertyLabel: "Gender")`) `genderWinsOverSexAtBirth`, `sexAtBirthBorrowsAFamily`, `mixedGenderIsNeutralAndDoesNotFallBack` | AC-2.1, AC-2.2, AC-2.3, AC-2.5 inputs → expected families and bases | AC-2.1, AC-2.2, AC-2.3, AC-2.5, D-6 |
| T-7 | Swift | same file `nothingUsableIsNeutral`, `aResearchersTermKeyedLikeAProductOneKeepsTheOtherFamily` | AC-2.4, AC-2.6 inputs → `.other` / nil | AC-2.4, AC-2.6 |
| T-8 | Swift | same file `badgeAndSpokenText` | header gender `woman` → badge "Woman", spoken "Gender, Woman"; sex `male` with `propertyLabel` "Sex at birth" → "Male · at birth", "Sex at birth, Male"; `propertyLabel` "Gender identity" → spoken "Gender identity, Woman" | AC-2.1, AC-2.2, D-3 |
| T-9 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `treatmentRidesOnContentAndRowsStay`, `headerThumbIsTheLeadLikeness` (gender moved onto `header:`; the gender field stays in `extra:` for the Vitals row) | → `.woman`, badge "Woman", thumb colors set, rows include "Gender" | AC-2.7 |
| T-10 | Swift | same file `treatmentReadsOnlyTheHeader` *(new)* | gender `woman` in `extra:` fields, header with no gender → `treatment == nil`, `badge == nil` | AC-2.8, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `conclusionheaders` (rename, finish step), `searchindex` (composes `PersonHeader`s via `RowsForSearch`), `api/ffi/handlers` promote and search tests that compare whole `PersonHeader` protos.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg 'conclusionheaders\.EventType\b|type EventType struct|PersonGenderTreatment\.read\(detail\.fields' .agents/ .cursor/rules/ docs/*.md core api macos/App` finds no hit; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief; no chrome change)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] person-detail.md: the S11-30 row reads Landed (#509)
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Person whose member records Gender = Woman. The Person page header shows the "Woman" badge above the name and the woman tint on the thumb, as on `main`.
2. On a Person with only Sex at birth = Male, the badge reads "Male · at birth" and the thumb takes the man tint.
3. On a Person with two Sources recording Gender Woman and Man, the header shows no badge and the default thumb; the Vitals Gender row shows the mixed state.
4. With VoiceOver, the badge reads "Gender, Woman" on the first Person.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test compares a whole `engine.PersonHeader` proto and now sees two new fields | Update the expected proto in that test; it is a fixture, not a behavior change |
| A relabel in T-2 does not raise `effects.ListVocabulary`, so `vocab` keeps the old label | Stop condition below; D-4 rests on it |
| The second Subject in T-1 "two Sources disagree" is outvoted instead of kept | Give each Subject its own Source (one Artifact each, as `joinPlace` does with its own Artifact); if both are still not kept, stop |

**Stop and report if:**

- `composePersons` no longer returns cached rows from `g.Display` (the finish step would then have nowhere to run).
- A relabel of a researcher's term or of a product Property does not change the next header read after STEP-2 (D-4's invalidation assumption is false).
- A single kept gender value is not rank 1 in `n.Kept` (D-1 / Edges "Assumptions" is false).
- `rg 'PersonGenderTreatment\.read' macos/App` finds a caller other than `PersonDetailContent.swift` at the base.
- `TestListPersonsQueryCountIsConstant` needs a higher count after STEP-2 (the label query is not the existing one).

## 12. Open questions

None.

## 13. PR

**Title:** S13-02 — Gender in the person header
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-02-gender-in-the-person-header.md`.
