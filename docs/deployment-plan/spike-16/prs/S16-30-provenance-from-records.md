# S16-30 — Provenance records compose from cached records

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2b — Entity records
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the provenance sheet, the Gallery captions and the life rows show the same records; only where the engine reads them from moves)
**Board:** —
**Depends on:** S16-11 (landed [#793](https://github.com/mendahu/provenencia/pull/793), outcome rows on the node), S16-12 (landed [#800](https://github.com/mendahu/provenencia/pull/800), Source records), S16-13 (landed [#809](https://github.com/mendahu/provenencia/pull/809), per-graph records)
**Blocks:** S16-24 (the detail read reports its inputs), S15-21 (the provenance sheet's citations), S15-28 (the records read by (owner, Property) pairs)
**Branch:** `s16-30-provenance-from-records`
**Version bump:** none — no change a researcher can see; the wire is unchanged ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc` (a reader composes from registered entries; a derived cache's row in `docs/catalog-caches.md` and its fill in `BenchmarkCacheSizes`), `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`
**Context:** [design § Entity records](../cache-architecture.md#entity-records); [`docs/catalog-caches.md`](../../../catalog-caches.md); [`provenance-loading.md`](../../spike-15/provenance-loading.md); working rules: "(Batch 2) The point of this batch is simpler code" (as clarified 2026-11-13), "(Batch 2) Per-graph records and memos are registry entries keyed by graph", "(Batch 2) A Source's words live once", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) Orders are defined and results do not move", "(Batch 2) Memory is budgeted", "(Batch 2) Measure growth", "(Batch 2) No Go cache for a view that is one query"; goal bars 4 and 11
**Planned:** 2026-11-13 against `main` @ `1689359`
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

A Conclusion page's records (the Vitals provenance sheet, each Gallery picture's record, and the Person's life rows' records) stop being read from SQL. Today `conclusiondetails` takes each record's outcome from the canonical node (S16-11), then reads the record shell (the Observation, its Subject, its Citation) in one SQL query by Subject, then looks up the recorded dates, names and a Place's toponym in three more. After this PR a provenance record is three cached pieces composed: the node's outcome row, the Subject record of the graph's records (S16-13, which now carry each Observation's Citation label, page, Artifact and uncertainty), and the Source record (S16-12). `sqlRecords`, `sqlToponyms`, `toponymsOf` and the date and name lookups go; a Person's life rows compose only the one Property each owner is read for; and the handler reads the subject type from the graph vocabulary. With the catalog warm, opening a Person on Vitals with its provenance sheet reads nothing from SQL except the picture paths (`files.RelPaths`).

**Why now.** S16-11, S16-12 and S16-13 put each piece in the cache; this PR is their reader. S16-24 instruments the detail read after it, and Spike 15's S15-21 and S15-28 edit the sheet and its records read after it.

**What stays.** The wire (`GetConclusionDetail`); every field, value, outcome, picture, life row and their order; `ReconciledValue` and `Outcome` as Go types; the summary rows and `PersonLifeOwners` (S16-19); the picture paths read; the claims and record links reads (S16-13); the records tier's key, cap, Apply and Check.

## 2. User stories

**US-1.** As a researcher, when I open a Person, Event or Place page on a warm catalog, its records come from memory, so that Vitals, the provenance sheet, the Gallery captions and the life rows open with no catalog read.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The handler test's Person with a name record, a birth event (event type, date) and a participation bridge; `GetConclusionDetail` once, then again under `databasetest.CountSelects` and the graph query counter | The second read reads none of `observations`, `citations`, `subjects`, `date_values`, `name_values`, `property_terms`, `subject_types`, `identity_claims`, `evidence_graphs`, `sources`, `artifacts`, `auto_reconciler_outcomes`; zero graph queries; the response equals the first |
| AC-1.2 | The likeness fixture (two pictures) under the same counters on a second read | None of the AC-1.1 tables and no `depiction_values`; zero graph queries; two pictures (the `files` read for their paths stays) |
| AC-1.3 | `conclusiondetails.ForEntity` and `PersonDerivedRows` on a warm graph (vocabulary, Source records, the person's graph records held), under `databasetest.CountSelects` | `N()` is 0 for each |
| AC-1.4 | Every existing case of `TestForEntity*`, `TestFieldsOf`, `TestDetail*`, `TestLinkRecordReadsAsItsPlacesToponym`, `TestPersonDerivedRows`, `TestGetConclusionDetail*`, `TestConclusionDetailProto*` | Each passes with its assertions unchanged (the tests D-11 edits change setup or counting only) |
| AC-1.5 | `households(f, 80)` (a person in 80 Evidence graphs): `ForEntity(son)` then `PersonDerivedRows(son)` cold (`catalogsessiontest.Cold`), then warm | Cold within 1 s, warm within 50 ms; the planning-base figures on the same machine are in the ship log |

**US-2.** As a researcher, when I fix a Citation's label, page or uncertainty, the provenance sheet shows the new words (bar 4), as it did when they came from SQL.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Person whose name record's Citation is labelled "Baptism of Ada" on page 37, not uncertain; `ForEntity`; then `citations.Update` to label "Baptism of Ada Lovelace", page 38, uncertain | The second `ForEntity` outcome has `CitationLabel` "Baptism of Ada Lovelace", `ArtifactPage` 38, `Provenance.Uncertain` true |
| AC-2.2 | The same edit under `PROVENENCIA_GRAPH_VERIFY=1` | The write passes Verify (the records Check sees the new words) |
| AC-2.3 | `TestForEntityAfterSourceEdit` (Source rename, Artifact relabel, credibility) | Unchanged result |

**US-3.** As the engine, each graph's records carry its Observations' Citation words, so that a reader composes a provenance record with no Citation read.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A graph whose Observations cite a Citation on Artifact A, labelled "Baptism of Ada", page 37 (`{"version":1,"selectors":[{"type":"page","artifact_page":37}]}`), uncertain; `graphfill.RecordsOf` cold and from the loaded graph | Every Observation of the Citation has `ArtifactID` A, `CitationLabel` "Baptism of Ada", `ArtifactPage` 37, `PageLabel` "", `TranscriptionUncertain` true, in both fills; the two `*GraphRecords` are `reflect.DeepEqual` |
| AC-3.2 | A Citation with no label and a locator with no page | `CitationLabel` "", `ArtifactPage` 0, `PageLabel` "" |
| AC-3.3 | `SubjectRecord.FirstText(toponym)` for a Place Subject with toponyms "York" (ref b), "Eboracum" (ref a), a negative "Leeds" and a blank one | "Eboracum" (the lowest ref, positive, not blank; `observations.FirstText`) |
| AC-3.4 | A Place Subject with no toponym | `FirstText` returns "" |

**US-4.** As the engine, a Person's life rows compose only the Property each owner is read for, so that the derived rows read no other field.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `FieldsOf` with `only` naming one Property of a handle that records three | One field, that Property |
| AC-4.2 | `FieldsOf` with `only` naming a subject-valued Property (a place end) | That field, with its records read as their Place's toponym |
| AC-4.3 | `FieldsOf` with `only` nil over the same handle | Every non-depiction, non-subject-valued field, as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache.EvidenceObservation` carries `ArtifactID`, `CitationLabel`, `ArtifactPage`, `PageLabel`, `TranscriptionUncertain`, read by `graphfill`'s Observation loader | STEP-2, AC-3.1, AC-3.2 |
| `graphfill.SubjectRecord.FirstText` *(new)* | STEP-2, AC-3.3, AC-3.4 |
| `conclusiondetails` composes each outcome from the node's row, the Subject record and the Source record; `sqlRecords`, `sqlToponyms`, `loadRecords`, `toponymsOf`, the date and name lookups and the `Querier` parameter go | STEP-4, AC-1.3, AC-1.4, AC-2.1, AC-2.2, AC-2.3 |
| `PersonDerivedRows` composes only each owner's Property (`fieldsOf`'s `only` replaces `subjectValued`) | STEP-3, AC-4.1, AC-4.2, AC-4.3 |
| `GetConclusionDetail` reads the subject type key from the graph vocabulary | STEP-5, AC-1.1, AC-1.2 |
| Budgets at 80 graphs and the records' size, measured before and after | STEP-1, STEP-6, AC-1.5 |
| Docs: `catalog-caches.md`, the package comment, the map, `completed.md` | STEP-7 |

| Out | Home |
| --- | --- |
| The summary rows' life owners (`PersonLifeOwners`) | S16-19 |
| The picture paths read (`files.RelPaths`): one query, kept by rule "No Go cache for a view that is one query" (user, 2026-11-13) | not planned |
| The detail read reporting the handles, graphs and Sources it used | S16-24 |
| The sheet's records loading on open, by (owner, Property) pairs, and `FieldRecordCounts` | S15-28 |
| Rendered citations in the sheet | S15-21 |
| The claims and record links reads | S16-13 (landed) |
| The Source page's Observation list (`source_read.go` `listedFromSource`) reading the new Citation fields | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Detail read | `core/database/conclusiondetails/conclusiondetails.go` `ForEntity` | Takes the node (`Handle`), `c.DB()`, the vocabulary, the gallery tiles (`gallery.Of`), the lead rows; calls `fieldsOf(c.Graph(), db, vocab, []*graphcache.Node{n}, leadRows, false)` |
| Composition | same file `fieldsOf(g, q Querier, v, nodes, leadRows, subjectValued bool)` | Dedupes nodes; `loadProperties`, `attachValues`; `sources.RecordsOf(g)`; `loadRecords(q, recs, toponym, use, byProp, leadRows)`; `finish` per field |
| Field list | same file `loadProperties(g, nodes, subjectValued)` | A node's fields are `n.Outcomes`' Property ids, minus depiction, minus subject-valued unless `subjectValued` |
| Record shells | same file `loadRecords`, `sqlRecords` | Collects every outcome row with a Subject; one `sqlbatch.Query` of `sqlRecords` by the rows' Subject ids (chunks of 500) joining `observations`, `subjects`, `citations`, `property_terms`; joins each row to its node outcome by Observation id; Source words from `recs.Artifact(o.ArtifactID)` (error when missing); page from `locator.Page(locatorJSON)`; then `datevalues.LookupManyTx`, `namevalues.LookupManyTx`, `toponymsOf` |
| Toponyms | same file `toponymsOf`, `sqlToponyms` | One batched query of the value Subjects' positive toponym Observations (`INDEXED BY observations_subject_id_idx`), `observations.FirstText` per Subject |
| Life rows | same file `PersonDerivedRows` | `PersonLifeOwners`; loads owner nodes; `c.DB()`; `fieldsOf(g, db, v, nodes, nil, true)` composes every field of every owner; `EventsByIDs`; a second loop keeps, per owner, the field whose `PropertyID` matches `o.PropertyID` |
| Package comment | same file, lines 1–20 | "The record shells (Observation, Subject, Citation, Artifact, Source words) are one SQL read keyed by the rows' Subjects until the records tier holds them." |
| Test exports | `core/database/conclusiondetails/export_test.go` `FieldsOf`, `countingQuerier`, `CountFieldsOfQueries`, `ToponymPlan`, `SQLRecordsForTest` | Reach `fieldsOf` with a counting `Querier`; the toponym query plan; the shell SQL text |
| Node rows | `core/database/graphcache/graph.go` `Node.Outcomes`, `Outcome` (`ObservationID`, `SubjectID` nil when the Observation is gone, `Reason`, `ValueRank`, `DeniedBy`, `VoteSupport`, `VoteOf`), `Node.Members` (`SubjectID`, `SourceID`, `GraphID`, `Accepted`) | Exact on `Set.Handles` (S16-11) |
| Records tier | `core/database/graphfill/records.go` `SubjectRecord` (`Subject`, `Accepted`, `Observations` in ref order), `GraphRecords.Subject`, `Labels`, `Entity`, `Records` (Cap 4,096, `DropNamedGraphs`, `checkRecords`), `RecordsOf(g, graphIDs)` | Filled from the loaded graph when held, else `loadGraphRows`; one fill per call |
| Observation rows | `core/database/graphcache/evidence_graph.go` `EvidenceObservation`; `core/database/graphfill/load.go` `loadGraphObservations` | Scans id, ref, citation, Subject, Property, polarity, scalar, value ids, depiction (with `dc.locator_json` from `LEFT JOIN citations dc`), then decodes dates and names into `Date`, `Name` blobs. No Citation label, Artifact, page or uncertainty |
| Source records | `core/database/sources/records.go` `RecordsOf(g)`, `RecordIndex.Artifact(id) (Record, label, bool)`, `Record.CredibilityKey`, `Credibility` | Whole entry, exact on `Set.SourceScopes` |
| Effects | `core/database/effects/registry.go` `"citations"` (`EvidenceGraph: citationObservationGraphs`), `"observations"`, `ownedValue` for date and name values | Every Citation, Observation, date or name edit names the graphs of the Observations it touches, so `graphfill.Records` drops them |
| Edge values | `core/database/observations` `ErrOtherGraph` (tests `AC-1.5`, `AC-1.6`, `AC-1.11` in `observations_test.go`) | An association end's value Subject is on the same graph as the Observation |
| Handler | `api/ffi/handlers/conclusion_detail.go` `GetConclusionDetail`, `attachSummary` (`subjecttypes.GetByID`, then by key `PersonsByIDs` + `PersonDerivedRows`, `EventsByIDs`, `PlacesByIDs`), `attachPictures` (`files.RelPaths`) | `GetByID` is one SQL read per open |
| Vocabulary | `core/database/graphcache/vocab.go` `Vocabulary.TypeKeysByID` (provenencia types only), `TermsByID` (every term) | Cached; drops on `ListPropertyVocabulary` |
| Tests | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntity`, `TestForEntityCitationParts`, `TestForEntityEvidence`, `TestForEntityQueryCountIsConstant` (asserts equal graph-query counts and that `sqlRecords` reads no Source tables), `TestForEntityAfterSourceEdit`, `TestFieldsOf`; `placerecords_test.go` `TestLinkRecordReadsAsItsPlacesToponym`, `TestLinkRecordReadsKeepQueriesConstant` (`CountFieldsOfQueries`), `TestForEntityDropsSubjectValuedFields`, `TestToponymLoadUsesSubjectIndex` (`ToponymPlan`); `core/database/conclusionsummaries/placechains_test.go` `TestPersonDerivedRows`; `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailWarmReadsNoReconcilerOrGallerySQL`, `TestGetConclusionDetailCarriesDerivedRows`; `core/database/graphfill/records_test.go` `TestGraphRecordsFill`; `core/database/graphfill/records_bench_test.go` `BenchmarkGraphRecords` (budgets 2 s, 35 MB); `core/database/persondetail/claims_budget_test.go` `TestClaimsBudgets`, `recordlinks_perf_test.go` `households` | Green at the planning base |

**Precedent.** S16-13 ([#809](https://github.com/mendahu/provenencia/pull/809), `ab521b8`) composed the claims read from the same three pieces: `core/database/persondetail/claims.go` `claims` takes graph ids from the node's members, one `graphfill.RecordsOf` call, `records[graph].Subject(m.SubjectID)`, Source words from `sources.RecordsOf`. Copy that shape (members → graph ids → one fill → lookups). S16-12 ([#800](https://github.com/mendahu/provenencia/pull/800), `048751f`) moved the shells' Source words to `RecordIndex.Artifact`; keep its missing-record error. `TestClaimsBudgets` is the budget test to copy.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `graphcache.EvidenceObservation` gains `ArtifactID []byte`, `CitationLabel string`, `ArtifactPage int`, `PageLabel string`, `TranscriptionUncertain bool` (the Observation's Citation's). `loadGraphObservations` adds `dc.artifact_id, COALESCE(dc.label, ''), dc.transcription_uncertain` to its SELECT (the `LEFT JOIN citations dc` is already there) and sets `ArtifactPage, PageLabel = locator.Page(citationLocator.String)` when `citationLocator.Valid`; `ArtifactID` is cloned like the other ids | One loader feeds the loaded graph and the records, so both fills and both Checks carry the words with no new query; the names match `conclusiondetails.Outcome`'s, so composition copies field to field (S16-13 D-4: "one Subject-record shape that S16-30 extends") | A `Citation` map on `GraphRecords` filled by a fourth query — a second row shape and a query the loaded-graph fill would also need; a nested `CitationWords` type — a new type for five fields one reader copies |
| D-2 | `graphfill.SubjectRecord` gains `FirstText(propertyID []byte) string` *(new)*: `observations.FirstText` over its Observations of that Property (`TextRecord{Ref, Polarity, Text}`) | The "a Place reads as its first toponym" rule stays `observations.FirstText`; the cache layer builds the input, the reader asks | `toponymsOf` kept over the records — reader-side construction; a `toponyms` map stored in `GraphRecords` — memory for one reader's lookup |
| D-3 | `conclusiondetails` composes outcomes through `recordReader` *(new, unexported)*: `newRecordReader(g, v, nodes, rows)` takes each row's graph from its node's `Members` (accepted and provisional, by `SubjectID`), calls `graphfill.RecordsOf` once with the distinct graph ids and `sources.RecordsOf` once; `(r *recordReader) outcome(row graphcache.Outcome) (Outcome, bool, error)` returns false for a row with no `SubjectID`, a Subject on no member, a Subject not in its graph's records, or an Observation not among the Subject's | The S16-13 shape; one fill per read; a row the cache cannot place is dropped, as `loadRecords` drops a row with no shell | Per-row `RecordsOf` calls — N fills; reading every member's graph — fills graphs no row needs |
| D-4 | `recordReader` indexes Observations by id lazily, one Subject at a time: the first row of a Subject builds `obs[string(id)]` from that Subject's `Observations`; later rows look up | A Subject with 20,000 Observations (bar 2) costs one pass, not one per row | A scan of the Subject's slice per row — quadratic on crafted data; an index stored in `GraphRecords` — memory every graph pays for one reader |
| D-5 | An outcome copies: `ObservationID`, `Reason`, `ValueRank`, `DeniedBy`, `Vote` from the row (as today); `ObservationRef`, `CitationID`, `ArtifactID`, `CitationLabel`, `ArtifactPage`, `PageLabel`, `Provenance.Uncertain` (from `TranscriptionUncertain`) from the Observation; `SubjectID`, `SubjectRef` from the Subject record; `GraphID` from the records' `GraphID`; `SourceID`, `SourceTitle`, `ArtifactLabel`, `CredibilityKey`, `Provenance.Credibility` from `recs.Artifact(ArtifactID)`, which when missing is the error `conclusiondetails: Artifact %x has no Source record` (as today); `Recorded.Scalar` from `Scalar`; `Recorded.TermID`, `TermKey`, `TermLabel`, `TermOrigin` from `v.TermsByID()`; `Recorded.Date`, `Recorded.Name` from `valuecodec.UnmarshalDate(Date)` / `UnmarshalName(Name)` when the blob is set; and for a `ValueSubjectID`, `Recorded.Text` = the value Subject's `FirstText(toponym)` in the same graph's records and `HasText` = text != "" | Every field `loadRecords` fills, from the piece that owns it; dates and names decode the way `attachValues` and `listedFromSource` already do | Copying `Recorded.Date.ID` / `Name.ID` — the blobs do not carry the row ids, and the wire never sent them (D-12, Edges) |
| D-6 | Gallery lead rows compose through the same `outcome`; a depiction Observation has no date or name, so the result equals today's ("their Recorded value has no Date or Name"); the comment says so | One composition for fields and leads | A lead path of its own |
| D-7 | `fieldsOf(g *graphcache.Graph, v graphcache.Vocabulary, nodes []*graphcache.Node, leadRows []graphcache.Outcome, only map[string]map[string]bool)` — the `Querier` and `subjectValued` parameters go. `only` nil: each node's Properties except depiction and subject-valued (today's page fields). `only` non-nil: for each node, the Properties named under `string(n.ID)` (any value type except depiction); a node not named has no field. `loadProperties(g, nodes, only)` applies it | The life rows read one Property per owner; composing every field and keeping one is the "second pass" the map names | Keeping `subjectValued` beside `only` — two switches for one choice; a separate composer for life rows — a second composition |
| D-8 | `PersonDerivedRows` builds `only` from the four owner lists (`o.Entity.ID` → `o.PropertyID`), calls `fieldsOf(g, v, nodes, nil, only)`, and picks each owner's field by Property id from the at most two fields its owner returns. `c.DB()` goes | Names the Property before composing; the pick stays because one owner can carry two rows (an event that owns both a birth and a death date in crafted data) | Composing every field (today) |
| D-9 | `attachSummary` reads the type key as `vocab.TypeKeysByID()[string(d.Entity.SubjectTypeID)]` from `c.Graph().Vocabulary()`; an empty key (no provenencia type) returns `nil, nil` as a non-provenencia origin does today. `subjecttypes.GetByID`'s call and the file's `subjecttypes` import go | The user's answer (2026-11-13): drop the type read, keep the paths read; the vocabulary already holds every subject type | Keeping `GetByID` — one SQL read per open that bar 11 counts |
| D-10 | `sqlRecords`, `sqlToponyms`, `loadRecords`, `toponymsOf`, the `Querier` type and the `database/sql`, `datevalues`, `namevalues`, `scalarcols`, `sqlbatch`, `locator` imports of `conclusiondetails` go (keep any still used after the change; `go vet` decides). `ForEntity` stops calling `c.DB()` | Each loses its last caller | — |
| D-11 | Test edits: `export_test.go` `FieldsOf(c, ids, only)` takes `only`; `countingQuerier`, `CountFieldsOfQueries`, `ToponymPlan`, `SQLRecordsForTest` go. `TestToponymLoadUsesSubjectIndex` goes (no toponym query left). `TestForEntityQueryCountIsConstant` keeps its equal-count assertion and replaces the `SQLRecordsForTest` half with: a second `ForEntity` of each handle prepares no SELECT (`databasetest.CountSelects`). `TestLinkRecordReadsKeepQueriesConstant` counts `PersonDerivedRows(f.C, h)`'s graph queries after `g.Drop()` and a warm `Vocabulary` (the `TestPersonDerivedRows` pattern) for 2 and 8 unrelated toponyms, and asserts a second read issues none | Their subjects (the shell SQL, the toponym plan, the counting `Querier`) are gone; the query-count tests keep their meaning on the graph counter | Deleting the query-count tests — they hold "cost does not grow with unrelated data" |
| D-12 | `Recorded.Date.ID` and `Recorded.Name.ID` are nil after this PR | `valuecodec`'s blobs carry the value, not the row id; `conclusionValueProto` sends `DateToProto` / `NameToProto`, which carry no id; no reader in `core/` or `api/` reads `Recorded.Date.ID` or `Recorded.Name.ID` (`rg -n "Recorded\.(Date|Name)\.ID"` finds nothing at the planning base) | Looking the ids up — the SQL this PR removes |
| D-13 | Budgets (`core/database/persondetail/detail_budget_test.go` *(new)* `TestDetailBudgets` *(new)*, on `households(f, 80)`): `budgetDetailCold = time.Second`, `budgetDetailWarm = 50 * time.Millisecond`, timing `ForEntity(son)` then `PersonDerivedRows(son)`. The test lands first (STEP-1) and is measured on the planning base | The claims read's budgets on the same fixture (`TestClaimsBudgets`); rule "Measure growth" | A budget in `api/ffi/handlers` — the handler adds summary and path reads this PR does not change |
| D-14 | No Apply changes. `graphfill.Records` and `graphfill.Graphs` keep `DropNamedGraphs`; a Citation edit already names its Observations' graphs (`"citations"` `EvidenceGraph: citationObservationGraphs`), so the new fields stay exact | Rule "Every cache names the signal that updates it": the signal exists | A Citation drop of its own — the effect already names the graphs |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `EvidenceObservation.ArtifactID`, `CitationLabel`, `ArtifactPage`, `PageLabel`, `TranscriptionUncertain` *(new fields)* | fields | `conclusiondetails.sqlRecords`'s Citation columns; `loadGraphObservations`' `LEFT JOIN citations dc` | Extends the one Observation row both fills share (D-1); the columns move from the reader's SQL to the loader |
| `graphfill.SubjectRecord.FirstText` *(new)* | method | `observations.FirstText`; `grapheventtitles` builds `TextRecord`s from a loaded graph | Calls `observations.FirstText`; a Subject record is the input, built in the cache layer (D-2). `grapheventtitles` reads a loaded `EvidenceGraph`, not records, and keeps its loop |
| `conclusiondetails.recordReader`, `newRecordReader`, `outcome` *(new, unexported)* | type, funcs | `conclusiondetails.loadRecords` (deleted); `core/database/persondetail/claims.go` `claims` (members → graph ids → `RecordsOf`) | Replaces `loadRecords`. The members-to-graphs loop is five lines in `claims`, which also filters by the graph index and `boundedMembers`; a shared helper would carry both callers' filters |
| `fieldsOf`'s `only` parameter *(replaces `subjectValued`)* | param | `subjectValued bool` | Replaces it (D-7) |
| `TestDetailBudgets`, `budgetDetailCold`, `budgetDetailWarm` *(new)* | test | `TestClaimsBudgets` | The same fixture and helpers; a separate read |
| `TestGraphRecordsCarryCitationWords`, `TestForEntityAfterCitationEdit`, `TestFirstText` cases, `TestGetConclusionDetailWarmReadsNoRecordSQL` *(new)* | tests | §8 | — |

**Builds on.** The records tier (`graphfill.Records`, S16-13), the Source records (`sources.Records`, S16-12), the node's outcome rows (S16-11), the graph vocabulary; the claims read's composition shape; `observations.FirstText`; `databasetest.CountSelects`.
**Deletes.** `conclusiondetails.sqlRecords`, `sqlToponyms`, `loadRecords`, `toponymsOf`, `Querier`; the `subjectValued` parameter; `fieldsOf`'s and `ForEntity`'s and `PersonDerivedRows`' `c.DB()` calls; `export_test.go` `countingQuerier`, `CountFieldsOfQueries`, `ToponymPlan`, `SQLRecordsForTest`; `TestToponymLoadUsesSubjectIndex`; `attachSummary`'s `subjecttypes.GetByID` call; the package comment's sentence "The record shells (Observation, Subject, Citation, Artifact, Source words) are one SQL read keyed by the rows' Subjects until the records tier holds them".
**Next adds.** S16-24 reports the graphs the detail read used: `recordReader` already holds the distinct graph ids and the Sources it asked `recs.Artifact` for. S15-28's records read by (owner, Property) pairs is `fieldsOf` with `only` over the pairs. S15-21 reads `CitationID` as today.

### Reach

None — no new path into existing code, no new state reachable. The detail read's callers (`GetConclusionDetail`, `autoreconciler_test`, `conclusionsummaries_test`) are unchanged.

### Scale

At the map's Scale (Batch 2: 16,000 persons; a person in 50 and in 80 graphs; 3,000 Sources, 10,000 graphs, 50,000 Observations):

- **Warm read.** Lookups only: per outcome row one map hit for the graph, one for the Subject, one for the Observation (D-4), one for the Artifact. A person in 80 graphs with two records per graph and ~10 fields: hundreds of lookups.
- **Cold read.** The person's graphs' records fill once, through `RecordsOf` (three queries plus date and name lookups per graph not held), the cost the Sources section's cold claims read already pays (`TestClaimsBudgets` cold, 13.2 ms at 80 graphs on the S16-13 ship log's machine). Today the shells were one query by Subject. AC-1.5 measures both at 80 graphs.
- **Graphs past the cap.** A person in more than 4,096 graphs refills the least recently used graphs on each read (`GetMany` returns every value, so the read is correct). Real data: 50 graphs (Scale); the cap is S16-13's.
- **Memory.** Five fields more per Observation in the records and the 64 loaded graphs: two words for `ArtifactID`'s slice header plus 16 bytes, a string header plus the label, two ints and a bool (about 70 bytes plus the label). At 50,000 Observations, about 3.5 MB more at most when all 10,000 graphs' records were held; at the cap (4,096 graphs) less. `BenchmarkGraphRecords` (budget 35 MB, today 17.3 MB) and `BenchmarkCacheSizes`' `graphRecords` line measure it before and after (STEP-1, STEP-6).
- **Writes.** No Apply change (D-14). A Citation edit drops its graphs' records and loaded graphs as today; the next read refills them.

### Edges

- **Assumptions.** (1) Each outcome row's Subject is one of its node's members: `Node.Members` and `Node.Outcomes` load together and reload on `Set.Handles` (S16-11). A row whose Subject is on no member is dropped (D-3); T-3 covers the warm path with every existing fixture. (2) An association end's value Subject is on the Observation's graph (`ErrOtherGraph`, `observations_test.go` AC-1.6, AC-1.11), so its toponym comes from the same records; `TestLinkRecordReadsAsItsPlacesToponym` holds it. (3) A Subject's Observations in the records are every Observation of the Subject (the loader reads the whole graph); `TestGraphRecordsFill` holds it. (4) Outcome order is set by `finish` (group rank, value rank, Observation id), not by read order, so composing in row order changes nothing; AC-1.4.
- **Neighbors.** A subject-valued Property on a Person page (`TestForEntityDropsSubjectValuedFields`): still not composed, `only` nil (AC-4.3). A subject-valued Property named in `only` whose value Subject is not on the graph (a `mentions` of a reification on another Source, which `ErrOtherGraph` allows for non-edge rows): `Recorded.Text` is "" where today it read that Subject's toponym. The life rows name only an owner's date or place-end Property, so no read reaches this today; S15-28's pairs read names Properties from a row's `RecordsSource` and inherits the rule (a Stop condition if a test reaches it). A negative toponym and a blank one are skipped (AC-3.3). A Citation with no page (AC-3.2).
- **Removed inputs.** An Observation deleted after its node loaded: the node reloads on `Set.Handles` (its members' handles), its row goes. A Citation, date or name edit drops the graph's records (D-14), so no stale words. A Source rename reaches the record through `sources.Records` (AC-2.3).
- **Real data.** None — no table, seed or configuration drives the composition.
- **Untrusted size.** Outcome rows per node: bounded by the node's records; crafted 20,000 records of one person in one graph (bar 2) is one graph fill and one Observation index per Subject (D-4). Graphs per read: the node's distinct member graphs; past 4,096 they refill (Scale). Observations per Subject: one index pass per Subject per read. Toponyms per value Subject: one `FirstText` pass per distinct value Subject per read, memoised in `recordReader` by Subject id.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `EvidenceObservation` gains the five fields (D-1); doc comment names them as the Citation's |
| `core/database/graphfill` | `loadGraphObservations` reads them (D-1); `SubjectRecord.FirstText(propertyID []byte) string` *(new)* (D-2) |
| `core/database/conclusiondetails` | `recordReader` *(new)*; `fieldsOf`, `loadProperties` take `only` (D-7); `ForEntity`, `PersonDerivedRows` (D-8) stop opening the DB; deletions (D-10); package comment: a record is the node's outcome row, the graph's Subject record and the Source record, composed with no SQL |

Sketch:

```go
// recordReader composes provenance records from the node's outcome rows, the
// graphs' Subject records and the Source records.
type recordReader struct {
	recs      *sources.RecordIndex
	graphs    map[string]*graphfill.GraphRecords // by graph id
	graphOf   map[string][]byte                  // Subject id -> graph id
	terms     map[string]graphcache.VocabTerm
	toponym   []byte
	obs       map[string]graphcache.EvidenceObservation // Observation id, filled per Subject
	indexed   map[string]bool                           // Subjects whose Observations are in obs
	toponyms  map[string]string                         // value Subject id -> first toponym
}

func newRecordReader(g *graphcache.Graph, v graphcache.Vocabulary, nodes []*graphcache.Node, rows []graphcache.Outcome) (*recordReader, error)
func (r *recordReader) outcome(row graphcache.Outcome) (Outcome, bool, error)
```

Errors: none new; the missing Source record error stays.

### FFI

None — the wire is unchanged. `api/ffi/handlers/conclusion_detail.go` `attachSummary` reads the type key from the vocabulary (D-9).

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Measure the planning base**
Files: `core/database/persondetail/detail_budget_test.go` *(new)*.
Do: Add `TestDetailBudgets` per D-13 (fixture subtest: `ForEntity(son)` has a name field with 80 outcomes; cold and warm subtests). Run it and `BenchmarkGraphRecords` and `BenchmarkCacheSizes` on the planning base; write the figures (cold, warm, records build ms and heap MB, `graphRecords` size) in the ship log.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestDetailBudgets -count=1 ./core/database/persondetail/` passes; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkGraphRecords|BenchmarkCacheSizes' -benchtime=1x ./core/database/graphfill/ ./core/database/promotealign/` prints the figures.

**STEP-2 — The records carry Citation words**
Files: `core/database/graphcache/evidence_graph.go`, `core/database/graphfill/load.go`, `core/database/graphfill/records.go`, `core/database/graphfill/records_test.go`. Skill: `add-catalog-query`.
Do: D-1, D-2. Add T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/graphfill/... ./core/database/evidencelayer/... ./core/database/persondetail/... ./api/ffi/handlers/` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/graphfill/` passes.

**STEP-3 — Life rows compose only their Property**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `export_test.go`, `conclusiondetails_test.go` (`TestFieldsOf`), `placerecords_test.go`.
Do: D-7, D-8: `fieldsOf` and `loadProperties` take `only` in place of `subjectValued` (the `Querier` parameter stays until STEP-4); `ForEntity` passes nil; `PersonDerivedRows` builds `only` from its owner lists and picks each owner's field by Property id. `export_test.go` `FieldsOf(c, ids, only)`; `CountFieldsOfQueries` takes `only` until STEP-4 deletes it. Add T-5's `TestFieldsOf` cases.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/ ./core/database/conclusionsummaries/ ./api/ffi/handlers/` passes; `rg -n subjectValued core/database/conclusiondetails` finds nothing.

**STEP-4 — The detail read composes its records**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `export_test.go`, `conclusiondetails_test.go`, `placerecords_test.go`. Skill: `add-catalog-query`.
Do: D-3, D-4, D-5, D-6, D-10, D-11. Add T-3, T-4 and T-5's `TestLinkRecordReadsKeepQueriesConstant` edit. Update the package comment.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/ ./core/database/conclusionsummaries/ ./core/database/autoreconciler/ ./api/ffi/handlers/` passes; `rg -n "sqlRecords|sqlToponyms|toponymsOf|loadRecords|LookupManyTx|sqlbatch|c\.DB\(\)" core/database/conclusiondetails` finds nothing.

**STEP-5 — The handler reads the subject type from the vocabulary**
Files: `api/ffi/handlers/conclusion_detail.go`, `api/ffi/handlers/conclusion_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-9. Add T-6; widen `TestGetConclusionDetailWarmReadsNoReconcilerOrGallerySQL`'s table list to AC-1.2's.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/` passes; `rg -n subjecttypes api/ffi/handlers/conclusion_detail.go` finds nothing.

**STEP-6 — Measure after**
Files: none new.
Do: Run STEP-1's commands on the branch; write the after figures in the ship log beside the base. Run the full Go suite under Verify.
Check: `TestDetailBudgets` passes; `BenchmarkGraphRecords` within its 2 s and 35 MB budgets; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-7 — Docs**
Files: `docs/catalog-caches.md` (the `graphfill.records` and `graphfill.graphs` rows: the Observations carry their Citation's label, page, Artifact and uncertainty; the new sizes), `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-30-provenance-from-records.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphfill/records_test.go` `TestGraphRecordsCarryCitationWords` *(new)* | A Citation labelled "Baptism of Ada", page 37, uncertain → every Observation of it in the records (cold fill and from the loaded graph) carries Artifact A, the label, 37, "", true, and the two fills are `DeepEqual`; a Citation with no label and no page → "", 0, "" | AC-3.1, AC-3.2, D-1 |
| T-2 | Go | `core/database/graphfill/records_test.go` `TestSubjectRecordFirstText` *(new, table)* | toponyms York (ref b), Eboracum (ref a), negative Leeds, blank → "Eboracum"; none → "" | AC-3.3, AC-3.4, D-2 |
| T-3 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityQueryCountIsConstant` (edited per D-11) | Small and big Person cold → equal graph-query counts; a second `ForEntity` of each → `CountSelects` `N()` 0 | AC-1.3, D-11 |
| T-4 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityAfterCitationEdit` *(new)*; run once under `t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")` | label "Baptism of Ada" page 37 → `citations.Update` label "Baptism of Ada Lovelace", page 38, uncertain → the outcome shows the new label, 38, `Provenance.Uncertain` true; Verify passes | AC-2.1, AC-2.2, D-14 |
| T-5 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestFieldsOf` (new cases) and `placerecords_test.go` `TestLinkRecordReadsKeepQueriesConstant` (edited per D-11) | `only` naming one of three Properties → one field; `only` naming a place end → its records read "York"; `only` nil → today's fields; 2 vs 8 unrelated toponyms → equal cold graph-query counts, 0 warm | AC-4.1, AC-4.2, AC-4.3, AC-1.3, D-7 |
| T-6 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailWarmReadsNoRecordSQL` *(new)*, on `TestGetConclusionDetailCarriesDerivedRows`' fixture | second read → none of AC-1.1's tables, 0 graph queries, the same response bytes | AC-1.1, D-9 |
| T-7 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailWarmReadsNoReconcilerOrGallerySQL` (table list widened) | second read → none of AC-1.2's tables; two pictures | AC-1.2 |
| T-8 | Go | `core/database/persondetail/detail_budget_test.go` `TestDetailBudgets` *(new)* | `households(f, 80)` cold ≤ 1 s, warm ≤ 50 ms | AC-1.5, D-13 |
| T-9 | Go | the existing suites named in AC-1.4 and AC-2.3 | unchanged | AC-1.4, AC-2.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `./core/database/conclusiondetails/`, `./core/database/conclusionsummaries/` (`TestPersonDerivedRows`), `./api/ffi/handlers/` (`TestGetConclusionDetail*`), and anything that compares `EvidenceObservation` values with `reflect.DeepEqual` (`./core/database/graphfill/`, `./core/database/evidencelayer/`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`sqlRecords`, `sqlToponyms`, `toponymsOf`, `loadRecords`, `CountFieldsOfQueries`, `ToponymPlan`, `SQLRecordsForTest`, `subjectValued`, "record shells"); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] Ship log: STEP-1 and STEP-6 figures (TestDetailBudgets cold and warm, BenchmarkGraphRecords build and heap, BenchmarkCacheSizes graphRecords), base and after, same machine
- [ ] docs/catalog-caches.md: graphfill.records and graphfill.graphs rows name the Citation words; sizes updated
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none (docs/conclusion-reconciliation.md §12's row stays true)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Person cited in several Sources, one record of which has a Citation label and a page. Open the Person on Vitals and open a name's provenance sheet: each record shows its Source, Artifact, Citation label and page as before.
2. In the Source page, edit that Citation's label and page. Return to the Person (Back): the sheet shows the new label and page.
3. Open the Person's Gallery: each picture's caption names its Source as before. Open a life row's (birth date, birth place) records: they show as before, the place records as the Place's name.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A cold Person open costs more than today's one shell query, since it fills whole graphs' records | AC-1.5 budgets it at 80 graphs; the Sources section pays the same fill (`TestClaimsBudgets`), so a visit fills it once for both |
| The five fields grow the records past their 35 MB budget | Measured in STEP-1 and STEP-6 against `BenchmarkGraphRecords`' budget |
| A decoded date or name differs from the row SQL read | AC-1.4's tests compare the recorded values; D-12 names the one known difference (the row id) |

**Stop and report if:**

- An existing `TestForEntity*`, `TestPersonDerivedRows` or `TestGetConclusionDetail*` assertion fails on a recorded value (other than a `Date.ID` / `Name.ID` the test compares, D-12).
- An outcome row in any existing test has a Subject that is not on one of its node's `Members` (D-3 would drop a record today's SQL showed).
- A test reaches a subject-valued record whose value Subject is on another graph (Edges, Neighbors).
- `TestDetailBudgets` cold or warm misses its budget after STEP-6, or `BenchmarkGraphRecords` passes 35 MB.
- `rg -n "Recorded\.(Date|Name)\.ID" core api` finds a reader.
- `Vocabulary.TypeKeysByID` lacks a subject type that `subjecttypes.GetByID` returns with origin provenencia.

## 12. Open questions

None. Answered 2026-11-13: the handler reads the subject type from the vocabulary; the picture paths stay one SQL read, and bar 11 says so.

## 13. PR

**Title:** S16-30 — Provenance records compose from cached records
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; TestDetailBudgets, BenchmarkGraphRecords and BenchmarkCacheSizes figures, base and after; `Plan: docs/deployment-plan/spike-16/prs/S16-30-provenance-from-records.md`.
