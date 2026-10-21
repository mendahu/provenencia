# S11-19 — Claim evidence query

**Status:** Landed in [#513](https://github.com/mendahu/provenencia/pull/513) (2026-10-21)

**Landed 2026-10-21.** Differences from the plan: D-10's collapse key also holds the line's outcome and property cause, which are the same whenever the incoming Ref, relation, value cause and member value key are (so no line collapses differently). `Build` keeps a corresponding link whose bridges all name a Subject that is not a member as a relationship with no bundles and no sections (D-13 skips the bridge, not the link). The loader leaves out a filed bridge Subject whose endpoint Observations do not name both ends from positive Observations, or whose end names two different Subjects; no stop condition fired on the test catalogs. STEP-5's `Evidence` calls `claimevidence.Build`, so STEP-5 and STEP-6 landed as one commit, with `subjects.SourcesOf` in its own. No expected value of an existing `subjectcompare`, `graphalign` or `promotealign` test changed.
**Spike / slice:** Spike 11 · Slice 5 — Claim evidence
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a Go query with no caller on the wire; the RPC is S11-25 and the sheet S11-20)
**Board:** —
**Depends on:** S11-16 (landed [#501](https://github.com/mendahu/provenencia/pull/501)); S11-30 (landed [#509](https://github.com/mendahu/provenencia/pull/509): `subjectcompare.Pool.Terms`, which D-6 reads through `subjectcompare.Side`). Branch from `main`
**Blocks:** S11-25
**Branch:** `s11-19-claim-evidence-query`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (a registry section, its inventory rows, a pure package on the guard list; this PR extends the skill), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (the loader reads `graphcache`)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/catalog-session.mdc`
**Context:** [`claim-evidence.md`](../claim-evidence.md) (whole: Separate from matching, Lines, Writing pins, Evidence through relationships); [Match registry](../comparison-layers.md#match-registry) (`evidence.go` row); [Detail](../comparison-layers.md#detail) (value refs); [`docs/matching-model.md`](../../../matching-model.md) knob inventory rows `evidence.*`; [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §9; working rules "Build, then migrate", "High tolerance for churn", "A separate, lazy query per row", "Lines", "Relationships", "The pin policy has one owner, in Go", "One match registry", "One guard, not one per package", "One matching skill, grown with the layers"; S11-13 ship log MNT-3 and MNT-4
**Planned:** 2026-10-21 against `main` @ `e169b71`; rechecked at `c2312f1` (no Current state file changed)
**Review:** [`S11-19-claim-evidence-query.review.md`](S11-19-claim-evidence-query.review.md) — 2 rounds: Fits with changes / Not ready, every finding fixed; M-4 answered by the user

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

One Promote row's claim evidence is computed by its own Go query, from the row's target and its neighbors' targets as the page holds them. Every accepted member of the target entity gets its own line per property, members with an identical value collapse into one line, and a relationship that corresponds through segment compare is one line that bundles every Observation behind it. Which lines are pinnable and which start checked comes from a new registry section, `evidence`. Today `pairExhibits` shows one member per incoming value, and `attachExhibits` runs inside every proposal ([`claim-evidence.md`](../claim-evidence.md#why)).

**Why now.** Slice 5's first PR. S11-16 landed layer 6, so segment compare and entity compare are the layers the query composes. S11-25 puts the query behind a lazy RPC and moves Apply and Swift onto it; it needs the query first ("Build, then migrate").

**What stays.** `attachExhibits`, `pairExhibits`, `withRowLines`, `refLine`, `PromoteGraphAlignmentRow.comparisons` (field 9), the Promote handler, Swift and the sheet: all unchanged in behavior until S11-25. `Propose` returns the same proposal; its setup moves into a helper it shares with the query (D-3). The pin tables: `identity_claim_evidence` (`core/database/migrations/000032.sql`, rebuilt in `000033.sql`) has primary key `(identity_claim_id, observation_id)` and stores no incoming → member pairing, so one incoming Observation pins against many member Observations, and `identityclaims.PinTx` writes a shared incoming pin once (`INSERT … ON CONFLICT DO NOTHING`). No migration PR is cut ahead of this one.

## 2. User stories

Conventions. `name(f)` is `propertycomparetest.Name(f)`, `term(k)` is `propertycomparetest.Term(k)`, `date(y)` is `propertycomparetest.Date(datevaluetest.YMD(y, 0, 0))`; `ref(v, b)` is `propertycomparetest.Ref(v, b)`, and every incoming and member value in a US-1 row carries a distinct ref. Metas are `subjectcomparetest.Metas()`. Rules are `matchregistry.Default()`: `reg.Evidence`, `reg.Subject`, `reg.Value` (tests import `matchregistry` from the external `claimevidence_test` package). "Checked" and "pinnable" are `Line.Checked` and `Line.Pinnable`. Points are within 1e-9.

**US-1.** As the engine, I compare one incoming value with each member's values, so every member that agrees is evidence.

`claimevidence.Sections(kind, incoming, members, metas, reg.Evidence, reg.Subject, reg.Value)` (D-5).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person; incoming `name` "John Smith"; five members, each with `name` "John Smith" | One section `name`, one line: relation `same`, outcome agree, checked, pinnable, five `Members` in member order, points `subject.Weights.person.name.Agree` = 2.5 |
| AC-1.2 | Incoming "John Smith"; members "John Smith", "John Smith", "Jon Smith" | Two lines: `same` with two members (checked); `near` cause `spelling` with one member (pinnable, unchecked) |
| AC-1.3 | Incoming "José Smith"; member "Jose Smith" | One line: `equivalent` cause `fold`, checked, pinnable |
| AC-1.4 | Incoming "J. Smith"; member "James Smith" | One line: `contains` cause `initial`, pinnable, unchecked, points 2.5 × the speaking similarity |
| AC-1.5 | Event; incoming `date` 1817; member `date` 1830 | One line: outcome conflict, relation `different`, not pinnable, not checked, points −`subject.Weights.event.date.Conflict` |
| AC-1.6 | Incoming `name` "John Smith"; one member with `name`, one member with no `name` | One line, one member: a member with nothing to compare has no line (outcome unknown) |
| AC-1.7 | Incoming negative `name` "John Smith" (the record denies it); member "John Smith" | One line: outcome conflict, property cause `negative`, not pinnable |
| AC-1.8 | Incoming `gender` `man`; member `sex_at_birth` `male` | One line in section `sex_at_birth` (the lead): outcome agree, checked; `Line.Property` is `gender`; points `subject.Weights.person.sex_at_birth.Agree` |
| AC-1.9 | Event; incoming `start_date` 1817; member `start_date` 1817 and `date` 1817 | One line in section `date`: the incoming `start_date` against the member's `start_date`, relation `same`, checked, points `subject.Weights.event.date.Agree` = 2; the member's `date` is not compared with it (D-6) |
| AC-1.10 | Event; incoming `event_name` (a property with no lead; text, single in `subjectcomparetest.Metas()`) "Baptism of John"; member "Baptism of John" | Section `event_name`, one checked line, points 0 |
| AC-1.12 | Event; incoming `date` 1817; member with only `start_date` 1817 and `end_date` 1817 (no `date`) | No line: the speaking member value is the period `Side` builds, which names no Observation (D-5) |
| AC-1.11 | Person; incoming values include a derived lifespan (no ref) | No line for it (D-5) |

**US-2.** As the engine, I bundle a corresponding relationship into one line, so a researcher can say "they have a matching birth event" and check what is behind it.

AC-2.1–AC-2.6 hold for `claimevidence.Build` (D-12) in table tests with hand-built inputs, and end to end for `promotealign.Evidence(c, req)` (D-3) on a catalog built with the `promotealign_test.go` fixture.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A filed person P (one member, birth event E with `date` 1817, P subject of E). A Source with person "John Smith", a birth dated 1817, subject participation. Request: the person on P, neighbor the birth on E | One relationship: key `participation·subject·person·birth`, outcome agree, checked, points the segment evidence; `Incoming` holds the Source participation Subject's Observations and the Source birth's `event_type` Observation; one `Members` bundle with the filed participation's Observations and the filed birth's `event_type` Observation; one section `date` under it with one checked line |
| AC-2.2 | AC-2.1 with the Source birth dated 1830 | The relationship is partial: pinnable, unchecked. Its `date` line is a conflict, not pinnable |
| AC-2.3 | Filed P has a spouse S1 (relationship `spouse`); the Source person has a sibling, neighbor mapped to S1 | No relationship: `sibling` does not correspond to `spouse` |
| AC-2.4 | AC-2.1 with the neighbor's target New, Skip, or the neighbor not in the request; and a request whose neighbor is not a layer neighbor of the row's Subject | No relationship; no error |
| AC-2.5 | Filed P has two accepted members, each the subject of a birth that is a member of E | One relationship with two `Members` bundles |
| AC-2.6 | AC-2.1 with no `date` on either birth | One relationship, agree, checked, with no section under it (decided 2026-10-17: a relationship line works with nothing under it) |
| AC-2.7 | A request with: a Source that does not exist; a row Subject not a primary Subject on it; an empty row handle; a row or neighbor handle that is missing, merged, or of another kind | `promote.ErrInvalid` |
| AC-2.8 | AC-2.1; a member of P on a second Source titled "Census 1851" | Each line's and bundle's Source id is a key of `Evidence.Sources`, with that Source's title |

**US-3.** As the engine, a property line's points come from one helper, so no caller reads the speaking similarity itself (S11-13 MNT-3).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `subjectcompare.Points(res, Weights{Agree: 2.5, Conflict: 3})` with `res` agree; partial with speaking similarity 0.75; conflict; unknown | 2.5; 1.875; −3; 0 |
| AC-3.2 | `propertycompare.Result{}.Similarity()`; a result whose speaking pair has similarity 0.75 | 0; 0.75 |
| AC-3.3 | Every existing `subjectcompare`, `graphalign` and `promotealign` test | Passes with unchanged expectations: `Compare`'s similarity and evidence are the same numbers, and `Propose` the same proposal |

## 3. Scope

| In | Covered by |
| --- | --- |
| `propertycompare.Result.Similarity` *(new)*; `subjectcompare.Points` takes a `propertycompare.Result`; `Compare` builds score from `Points`; `promotealign`'s private `similarity` deleted | STEP-1, AC-3.1–AC-3.3 |
| `subjectprofile.ForSubject` sets each value's `Ref` to its Observation id | STEP-1, T-13 |
| `core/claimevidence` *(new)*: pure types, `Rules`, `Sections`, `Build` (lines, relationships, bundles), line state, collapse; on the guard list and the match rule globs | STEP-2, STEP-3, STEP-6, AC-1.1–AC-1.12, AC-2.1–AC-2.6 |
| Registry `evidence` section: `Registry.Evidence`, `core/matchregistry/evidence.go` *(new)*; `notYetBuilt` loses `"evidence."`; inventory row `evidence.Defining.event` renamed `evidence.Defining` | STEP-2, T-4 |
| `graphalign.Bridge.Associations` *(new)*; `layerValues` takes a set of Subject ids; `loadInputs` and `checkFixed` *(new)* shared by `propose` and the query | STEP-4 |
| `subjects.SourcesOf` *(new)* | STEP-5 |
| `promotealign.Evidence` *(new)*: the reads behind `Build`, and validation | STEP-5, STEP-7, AC-2.1–AC-2.8 |
| Docs: `add-match-rule`, `claim-evidence.md` status, `matching-model.md`, the map, an idea line for review M-6 | STEP-8 |

| Out | Home |
| --- | --- |
| The RPC, `engine.proto` messages, the handler | S11-25 |
| Default pins at Apply; Swift `draftedPins`, `draftedArgument`, `kept.pins` | S11-25 |
| Deleting `attachExhibits`, `pairExhibits`, `withRowLines`, `refLine`, field 9, `bridgeNeighbors`, `loadExhibitObservations`, `exhibitObs`, the pin-test helper `obs` (S11-13 MNT-5, PRF-2) | S11-25 |
| Swift parsing of relation and cause codes (`ConclusionVocabulary.swift`) | S11-25 |
| Copy for lines and relationships; the "why suggested" summary; the sheet's order of sections and relationships beyond D-11's | S11-20, S11-D2 |
| The lifespan line (`Comparison.Derived`) and likeness lines (depiction values): their engine part and their sheet | S11-20 (user, 2026-10-21; S10-54 landed in #433) |
| Provisional members on the sheet | not planned (D-9: accepted members only, as `acceptedMembers` today) |
| A bridge's own property lines (dates on a place relationship): bundled, never compared | not planned ([`claim-evidence.md`](../claim-evidence.md#evidence-through-relationships)) |
| `neighborTypeProperty` into the registry's segment section (review M-6) | [`docs/ideas/matching-test-suite.md`](../../../ideas/matching-test-suite.md) known gaps line |
| Timing the query on a large Source | S11-21 |

## 4. Current state

Verified at `e169b71`.

| Area | Where | Today |
| --- | --- | --- |
| Sheet lines | `core/database/promotealign/exhibits.go` `attachExhibits` | Runs after `graphalign.Align` in every `propose` (`promotealign.go:92`). Pairs each mapped row's own records with its handle's accepted members, and each mapped neighbor's records with members one bridge from the row's members (`bridgeNeighbors`, SQL over `observations`, any bridge type, no key check). Unchanged by this PR |
| Pairing | `exhibits.go` `pairExhibits` | One line per incoming Observation: the first `autoreconcile.Compatible` member value, else the most similar. Reads `similarity(res)` and calls `subjectcompare.Points(chosenRes.Outcome, sim, …)`. This PR changes only those reads (D-4) |
| Similarity helper | `exhibits.go` `similarity` | The speaking pair's similarity, 0 when none. Deleted by D-4 |
| Second Observation read | `exhibits.go` `loadExhibitObservations` | SQL read of Observations with sources and terms, building `autoreconcile.Value` beside `decode` (S11-13 MNT-4). The query does not use it (D-8); S11-25 deletes it |
| Setup | `core/database/promotealign/promotealign.go` `propose`, `mergeFixed` | `propose`: 16-byte check, `SELECT 1 FROM sources`, `valuerules.Load`, `AlignConfig()` with `cfg.Value = rules`, `loadTerms`, `loadLayer`, then `mergeFixed`, `newProfiler`, `loadCanon`, `Align`, `attachExhibits`. `mergeFixed`'s loop checks each caller `Fixed` (Subject in `primary`; New / Skip carry no handle; a 16-byte handle; the handle unmerged and of the Subject's kind), then adds already-promoted Subjects as anchors |
| Row decoder | `core/database/promotealign/profiles.go` `decode`, `layerValues(sg, primary, props, terms)` | Each listed Subject's product Observations, both polarities, as `autoreconcile.Candidate` (`ObservationID`, `Value`, `Negative`), by property key, in Observation id order. Reads `primary` only for membership. No Subject-valued Observations |
| Plain-Subject side | `core/subjectprofile/subjectprofile.go` `ForSubject`, `Value` | `ForSubject` turns candidates into `subjectcompare.Values` with `Value(c.Value, c.Negative, namePattern)`, which leaves `Ref` empty, then adds the lifespan values (no ref). One caller: `layer.go` `loadLayer` |
| Layer load | `core/database/promotealign/layer.go` `loadLayer`, `loadLayerBridges` | Builds every primary Subject's values once (`graphalign.Subject.Values`) and every bridge's keys from both ends (`segmentcompare.Key(linkOf(…))`) through `canonicalgraph.WalkSource`, whose `Edge.Association` is the bridge Subject id. The id is dropped; a second edge with the same `edgeSeenKey` is skipped. `export_test.go` calls `loadLayer` |
| Layer types | `core/graphalign/types.go` `Bridge{A, B; KeyA, KeyB}` | No association ids |
| Canon keys | `core/database/promotealign/canon.go` `canonKeys` | Reads a `graphcache.Link` as an edge with its key from each end. The one keying rule for the canon side |
| Handle profiles | `profiles.go` `newProfiler`, `profiler.profile`, `profiler.values` | Builds a handle's strict profile (and a person's life) from the cache |
| Members | `core/database/graphcache/graph.go` `Node.Members`, `Member{SubjectID, Accepted}`; `exhibits.go` `acceptedMembers` | Accepted and provisional claims on the entity, in claim id order. No Source id |
| Association members | `core/database/identityclaims/bridgefile.go` `fileBridge` | Files each bridge Subject as an accepted claim on its association entity, so `g.Node(link.Association).Members` are the filed bridge Subjects |
| Source graph | `core/database/graphcache/source.go` `Graph.Source`, `SourceGraph`, `SourceObservation` | One Source's Subjects and Observations (with `SubjectID`, `PropertyID`, `ValueSubjectID`, `Polarity`), cached. No title |
| Subject reads | `core/database/subjects/subjects.go` | Owns reads of `subjects` (`getTx` selects `source_id`). No batched Source lookup |
| Value refs | `core/valuecompare/valuecompare.go` `Value.Ref`, `Detail.RefA` / `RefB` | `Compare` echoes refs; `propertycompare.SpeakingPair.Detail` carries them. Nothing reads them yet |
| Points | `core/subjectcompare/subjectcompare.go` `Points(outcome, similarity, w)`, `Compare` | `Compare` reads `res.Speaking.Result.Similarity` for partial and repeats the agree / partial arithmetic inline for `score` |
| Pools and periods | `core/subjectcompare/pairs.go` `Side`; `subjectcompare.go` `Rules.Lead` | `Side`: a lead's own values, pooled values (terms mapped by `Pool.Terms`), or a period from start and end when no positive lead value. `Lead`: the lead a property counts under, for a weighted, pooled or period-end property; `false` otherwise |
| Segment compare | `core/segmentcompare/segmentcompare.go` `Compare`, `Segment{Key, NeighborKind, Neighbor}`, `Result{Outcome, Cause, Evidence, Neighbor}`; `key.go` `Corresponds` | `Compare(a, b, placed, own, metas, rules, subject, value)`: agree, or partial when the neighbors conflict; unknown when not corresponding or gated (`own` below `Gate.MinOwnEvidence`) |
| Registry | `core/matchregistry/registry.go` `Registry`, `Default` | Sections `Value`, `Subject`, `Segment`, `Alignment`. No `Evidence` |
| Inventory guard | `core/matchregistry/inventory_test.go` `notYetBuilt`, `TestKnobInventoryMatchesRegistry` | `notYetBuilt = {"evidence."}` keeps `docs/matching-model.md` rows `evidence.Checked`, `evidence.Pinnable`, `evidence.Defining.event` legal with no field. `registryKeys` expands a map only when its element is a struct or a map, so a `map[SubjectType][]PropertyKey` is one leaf (precedent: row `alignment.Shortlist.Words`) |
| Layer guard | `core/matchregistry/guard_test.go` `layerPackages`, `TestLayerPackages` | Purity and numeric-literal checks over non-test files of listed packages |
| Pin tables | `core/database/migrations/000032.sql`, `000033.sql` `identity_claim_evidence`; `core/database/identityclaims/pins.go` `PinTx`; `core/database/promote/promote.go` `pinPairs` | See **What stays** |
| Tests | `core/subjectcompare/subjectcompare_test.go` `TestPoints`; `core/database/promotealign/promotealign_test.go` `newFixture`, `newSource`, `person`, `birth`, `datedBirth`, `participation`, `relationship`, `join`, `promote`, `propose` | `TestPoints` calls the old signature |

**Precedent.** S11-14 ([#493](https://github.com/mendahu/provenencia/pull/493), [plan](S11-14-segment-compare.md)) added a pure package with its `Rules` type, a registry section file with a source comment per value, inventory rows and the guard line, and extended `add-match-rule`. Copy that shape for `core/claimevidence` *(new)* and `core/matchregistry/evidence.go` *(new)*. The loader shape is `loadLayer` / `loadCanon`: catalog reads in `promotealign`, pure decisions in the layer package. Do not copy `pairExhibits`' "first compatible, else closest" choice.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A new pure package, `core/claimevidence`, holds the evidence types, `Rules` and every line and relationship rule: `Sections` and `Build` (D-12, D-13). The loader, `core/database/promotealign/evidence.go` *(new)*, only reads: inputs, profiles, members, Source graphs, titles, and canon links keyed by `canonKeys`. `claimevidence` is on `layerPackages` with `numbers: true` and in `match-rules.mdc` globs. Its tests are an external `claimevidence_test` package and may import `matchregistry` | The policy is the view's own ([`claim-evidence.md`](../claim-evidence.md#separate-from-matching)) and the registry must name its `Rules` type without a cycle, as for every section. Pure means table tests without build tags | In `promotealign` — needs `-tags fts5` to test pure policy, and the registry cannot import it. In `graphalign` — layer 6 decides placement; evidence is a separate view |
| D-2 | `claimevidence.Rules{Checked, Pinnable []valuecompare.Relation; Defining map[vocabkey.SubjectType][]vocabkey.PropertyKey}`. `matchregistry.Default().Evidence`: Checked `Same`, `Equivalent`; Pinnable `Contains`, `Overlaps`, `Near`; Defining event: `event_type`. Delete `"evidence."` from `notYetBuilt`, and rename the row `evidence.Defining.event` to `evidence.Defining` in `docs/matching-model.md` (start value "event `event_type`") | The inventory walk reads a slice and a map of slices as one leaf each (`registryKeys`; row `alignment.Shortlist.Words`), so the keys are `evidence.Checked`, `evidence.Pinnable`, `evidence.Defining` | A per-relation struct — renames every row |
| D-3 | Entry point `func Evidence(c *database.Catalog, req EvidenceRequest) (claimevidence.Evidence, error)` *(new)*, `EvidenceRequest{SourceID, SubjectID, HandleID []byte; Neighbors []graphalign.Fixed}` *(new)*. Setup is `loadInputs(c, sourceID) (inputs, error)` *(new)*, moved out of `propose` (the 16-byte and `sources` checks, `valuerules.Load`, `cfg`, `namePattern`, `loadTerms`, `loadLayer`; `inputs{q, cfg, namePattern, terms, layer, primary}`), and both call it; the `MAX(revision)` read stays in `propose`; a nil `c` is `promote.ErrInvalid`. `checkFixed(q, primary, f) (graphalign.Fixed, error)` *(new)* is `mergeFixed`'s loop body; `mergeFixed` calls it. `Evidence` runs `checkFixed` on `{SubjectID, HandleID}` (an empty handle is `promote.ErrInvalid`) and on every `Neighbors` entry, and adds no anchors. A neighbor with Target New or Skip, or whose Subject is not a layer neighbor of the row's Subject, contributes nothing | Inputs are the page's, not matching internals ([`claim-evidence.md`](../claim-evidence.md#separate-from-matching)). One copy of setup and of the input rules (review M-3). `graphalign.Fixed` already says "Subject on handle, or New, or Skip" | A new `Target` struct — a copy of `Fixed`. Calling `mergeFixed` — it adds promoted Subjects the request did not name (AC-2.4) |
| D-4 | `propertycompare.Result.Similarity() float64` *(new)*: the speaking pair's similarity, 0 when `Speaking` is nil. `subjectcompare.Points(res propertycompare.Result, w Weights) float64`: agree `w.Agree`, partial `w.Agree × res.Similarity()`, conflict `−w.Conflict`, unknown 0. `Compare` sets `pr.Points = Points(res, w)`, adds `max(pr.Points, 0)` to `score`, and keeps its `counted` switch. `pairExhibits` calls `res.Similarity()` and the new `Points`; `promotealign.similarity` is deleted | The scope entry and S11-13 MNT-3: one place reads the speaking similarity. The method lives on the type that owns the speaking pair | A second helper beside `Points` — leaves `pairExhibits` reading the field |
| D-5 | `claimevidence.Sections(kind, incoming subjectcompare.Values, members []Member, metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) []Section`. For each incoming value with a non-empty `Ref` (positive or negative; a derived value has no ref and gives no line), for each member, one `propertycompare.Compare` call with the incoming value as side A (D-6) gives the line; an unknown outcome gives no line, and so does a speaking pair whose `Detail.RefB` is empty (a period `Side` built, AC-1.12). The member Observation is the speaking pair's `Detail.RefB` | One line per member per property ([Lines](../claim-evidence.md#lines)); comparisons use layer 2 only, never `Compatible` or Promote's scores | Comparing with the entity's profile — loses which member agrees |
| D-6 | What an incoming value of property p is compared with, and under which section, meta and weight. Let `lead, ok := subject.Lead(kind, p)`. (a) `ok` and p is the lead or in `subject.Pools[kind][lead].With`: both sides are `subjectcompare.Side(kind, lead, …)` (incoming: `Values{p: {v}}`; member: its `Values`), meta `metas[lead]`. (b) `ok` and p is a period end (`subject.Periods[kind][lead].Start` or `.End`): the member's `Values[p]` only, meta `metas[p]`. (c) not `ok`: the member's `Values[p]`, meta `metas[p]`. The section key is `lead` in (a) and (b), p in (c); weight `subject.Weights[kind][lead]` in (a) and (b), none in (c). A p with no `metas` entry gives no line | `Side` is the one place pooled terms are mapped (S11-30). A period built by `Side` is synthetic and has no `Ref`, so it cannot name a member Observation | Building periods per member — the line would pin no Observation |
| D-7 | Line state, `rules` applied to the speaking relation: checked when the outcome is agree and the relation is in `Checked`; pinnable when checked, or when the outcome is partial and the relation is in `Pinnable`. A conflict (relation `Different`, or property cause `negative`) is shown, not pinnable. Points: `subjectcompare.Points(res, weight)` with D-6's weight, else 0 | The table in [Lines](../claim-evidence.md#lines); points from the one helper (D-4) | Reading similarity in the view |
| D-8 | One value shape. `ForSubject` sets each built value's `Ref` to `c.ObservationID`; lifespan values keep an empty `Ref`. `layerValues(sg, subjects map[string]bool, props, terms)` takes the Subject ids to read; `loadLayer` passes its primary ids. The incoming side is the row Subject's `graphalign.Subject.Values` from `loadLayer`. A member side is `ForSubject(subjectprofile.Subject{Values: layerValues(memberSource, {member}, prof.props, prof.terms)[member], NamePattern: namePattern}, nil)`, with the profiler's `props` and `terms`. Member Source ids and titles come from `subjects.SourcesOf` (D-16) | One conversion path, one decode of the row's Source, one type fewer (review M-1); value refs are how the layers name an Observation ([Detail](../comparison-layers.md#detail)); no second Observation read (S11-13 MNT-4) | A `Record` type beside `Values`. `loadExhibitObservations`. `graphcache.Member.SourceID` — a cache shape change for one reader |
| D-9 | Members are the handle's accepted members (`acceptedMembers`), in claim id order; provisional members are left out | Today's rule, and pins backfill onto accepted claims only (`promote.pairClaim`) | Provisional members — a pin could not backfill |
| D-10 | Collapse: within a section, lines with the same incoming `Ref`, relation and value cause, whose member values have the same `valuecompare.Key(memberValue.Value, valuecompare.Same, value)`, are one line whose `Members` lists each member in member order. A value with no key (`ok` false) does not collapse | "Identical value" is the Same key by the working rule "A normalized `Key` per value type"; `Key` and `Compare` say `Same` together | Comparing display text — the six text rules S11-07 retired |
| D-11 | Order. Sections by ascending key; lines by incoming `Ref` bytes, then by the first member's claim order. Relationships: keys in `segment.One` (End cleared) first, then by `BridgeKey.String()`, then by the layer neighbor's ref | Deterministic, testable output. The sheet's own order is S11-D2's | Hand-listing "core" relationships here — a list S11-D2 has not drawn |
| D-12 | Relationships are decided in `claimevidence.Build(kind, row subjectcompare.Values, handle subjectcompare.Values, members []Member, neighbors []Neighbor, metas, rules Rules, segment segmentcompare.Rules, subject subjectcompare.Rules, value valuecompare.Rules) Evidence` *(new)*, which also returns `Sections(kind, row, members, …)`. A `Neighbor` *(new)* is one layer neighbor S′ the request places on handle H′: `Subject`, `Kind`, `Values` (S′'s layer values), `Handle` (H′'s profile), `Members` (H′'s accepted members as `Member`), `Layer []LayerLink` (each `layer.Bridges` entry between the row Subject and S′: the key from the row Subject's end, `KeyA` when it is `A`, and its `Associations`), `Canon []CanonLink` (each link in `g.Node(H).Links` to H′: the key from H's end by `canonKeys`, `keyFrom` when H is `from`, and its member bridges as `Bridge{Subject, Source, Member, Neighbor []byte; Observations [][]byte}`, Member and Neighbor read from the bridge's endpoint Observations). For each distinct layer key with a `CanonLink` that `segmentcompare.Corresponds`: one `Relationship`, state from `segmentcompare.Compare(Segment{key, S′ kind, S′ Values}, Segment{key, H′ kind, Handle}, Together, own, …)`, `own` = `subjectcompare.Compare(kind, row, handle, …).Evidence`; `Outcome`, `Cause` copied, `Points` = `Result.Evidence`; agree → checked and pinnable; partial → pinnable; unknown → none | One package owns every line and relationship rule (review M-7); the policy is table-testable without a catalog. [Evidence through relationships](../claim-evidence.md#evidence-through-relationships): a relationship exists only when the segments correspond; the page placed the neighbor, so `Together` | Correspondence and state in the loader — policy split across two packages. Every bridge between members, as `bridgeNeighbors` — the spouse-meets-sibling defect |
| D-13 | Bundles, in `Build`. Incoming: the row Source's Observations on the layer link's `Associations` (the loader passes them on `LayerLink.Observations`) plus the refs of S′'s values of a property in `rules.Defining[S′ kind]`. Members: a `Bridge` of a corresponding `CanonLink` whose `Member` is in `members` and whose `Neighbor` is in `Neighbor.Members`, else skipped; its bundle is the bridge's `Observations` plus the refs of that neighbor member's defining values, `Bundle{Member, Neighbor, Source}`. The relationship's sections are `Sections(S′ kind, S′ Values without defining properties, the neighbor members its bundles name, …)` | The registry rule "a segment's bundle is all of its bridge's Observations plus the neighbor kind's defining properties". Lines under it compare this birth with that birth only | Bundling the neighbor's other properties — they are their own checkable lines |
| D-14 | `graphalign.Bridge.Associations [][]byte` *(new)*: the bridge Subjects behind the edge, an opaque field the layers never read, documented as `Value.Ref` is. `loadLayerBridges` appends the association to the existing `Bridge` when `edgeSeenKey` was seen, instead of skipping | One owner for a bridge's facts; no side table, no discarded return on `Propose` (review M-2) | A fourth `loadLayer` return — a map `propose` throws away |
| D-15 | Wording: a line carries its relation, value cause (`valuecompare.Cause`) and property cause (`propertycompare.Cause`); a relationship carries its segment cause (`segmentcompare.Cause`). No separate wording-key vocabulary | [Wording](../claim-evidence.md#wording): lines are built from relation and `Detail.Cause`; S11-25 parses these codes in `ConclusionVocabulary.swift`. Every row of the wording table has its code (`fold`, `alias`, `precision`, `missing_part`, `spelling`, `gap`) | A Go `Wording` code set — a second vocabulary to mirror in Swift |
| D-16 | `subjects.SourcesOf(q sqlbatch.Querier, ids [][]byte) (map[string]SourceOf, error)` *(new)* in `core/database/subjects/sources.go` *(new)*, `SourceOf{SourceID []byte; Title string}`: one `sqlbatch.Query` over `subjects JOIN sources`. `Evidence.Sources map[string]string` holds the title of every Source a line, pin or bundle names, keyed by Source id as a string | Reads of `subjects` live in `core/database/subjects` (review M-5); titles are display data the pure package does not read | A title on each `Pin` — repeats it per line. SQL in `promotealign` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/claimevidence` *(new)* | package | `core/database/promotealign/exhibits.go` | Needs build tags and a database import; the registry cannot name its types |
| `claimevidence.Rules` *(new)* | type | `segmentcompare.Rules` | Each section's type lives with the code that reads it |
| `claimevidence.Member`, `Pin`, `Line`, `Section`, `Bundle`, `Relationship`, `Evidence` *(new)* | types | `graphalign.Exhibit` | `Exhibit` is one pair per incoming value with display strings; S11-25 deletes it with field 9. Values stay `subjectcompare.Values` |
| `claimevidence.Sections`, `Build` *(new)* | funcs | `pairExhibits`; `graphalign` `state.segments` | Replace the pairing in S11-25; built on `propertycompare.Compare`, `subjectcompare.Side`, `segmentcompare.Compare` |
| `claimevidence.Neighbor`, `LayerLink`, `CanonLink`, `Bridge` *(new)* | types | `graphalign.Bridge`, `CanonEdge` | Carry Observation ids and members per side, which the walk's types do not |
| `propertycompare.Result.Similarity` *(new)* | method | `promotealign.similarity` | Moves it to the type's owner; the copy is deleted |
| `graphalign.Bridge.Associations` *(new)* | field | `valuecompare.Value.Ref` | Same opaque-ref pattern for a segment |
| `matchregistry` `defaultEvidence`, `Registry.Evidence` *(new)* | func, field | `defaultSegment` | The section file the registry design names (`evidence.go`) |
| `promotealign.Evidence`, `EvidenceRequest` *(new)* | func, type | `Propose` | Lazy, per row, from page targets; no `Align` |
| `promotealign.loadInputs`, `inputs`, `checkFixed` *(new)* | funcs, type | `propose`, `mergeFixed` | Moved, not copied |
| `subjects.SourcesOf`, `SourceOf` *(new)* | func, type | `subjects.getTx` | Batched, ids and titles only |
| Tests in `core/claimevidence`, `core/database/promotealign/evidence_test.go`, `core/database/subjects` *(new)* | tests | `promotealign_test.go` fixture | Reused: `newFixture`, `newSource`, `person`, `birth`, `datedBirth`, `participation`, `relationship`, `join`, `promote` |

**Builds on.** `propertycompare.Compare` and value refs; `subjectcompare.Side`, `Lead`, `Points`; `segmentcompare.Compare` and `Corresponds`; `loadLayer`, `layerValues`, `decode`, `ForSubject`, `canonKeys`, `newProfiler`, `acceptedMembers`; the `add-match-rule` steps for a section and a package.
**Deletes.** `promotealign.similarity`; the inline partial arithmetic in `subjectcompare.Compare`; `propose`'s inline setup and `mergeFixed`'s inline checks (moved); `"evidence."` in `notYetBuilt`. The exhibit code goes in S11-25 ("Build, then migrate").
**Next adds.** S11-25 maps `claimevidence.Evidence` onto the RPC, and Apply's default pins are the checked lines' and relationships' pins. A new defining property is one `Defining` entry. S11-20 words lines from D-15's codes.

### Schema

None — `identity_claim_evidence` already holds one incoming Observation against many member Observations (see **What stays**).

### Go core

| Package | Change |
| --- | --- |
| `core/propertycompare` | `func (r Result) Similarity() float64` *(new)* (D-4) |
| `core/subjectcompare` | `func Points(res propertycompare.Result, w Weights) float64` (new signature); `Compare` uses it (D-4) |
| `core/subjectprofile` | `ForSubject` sets `Ref` (D-8) |
| `core/graphalign` | `Bridge.Associations` (D-14) |
| `core/claimevidence` *(new)* | Types and `Sections` below |
| `core/matchregistry` | `Registry.Evidence claimevidence.Rules`; `evidence.go` *(new)* `defaultEvidence()`, each value with the source comment `decided` citing [`claim-evidence.md`](../claim-evidence.md#lines) |
| `core/database/subjects` | `sources.go` *(new)*: `SourcesOf`, `SourceOf` (D-16) |
| `core/database/promotealign` | `promotealign.go`: `loadInputs`, `checkFixed` (D-3); `profiles.go`: `layerValues` signature (D-8); `layer.go`: associations (D-14); `exhibits.go`: D-4 call sites; `evidence.go` *(new)*: `EvidenceRequest`, `Evidence` |

```go
// package claimevidence — sketch

type Rules struct {
	Checked  []valuecompare.Relation                         // lines that start checked
	Pinnable []valuecompare.Relation                         // pinnable, unchecked
	Defining map[vocabkey.SubjectType][]vocabkey.PropertyKey // bundled into a relationship line
}

type Member struct {
	Subject, Source []byte
	Values          subjectcompare.Values // each value's Ref is its Observation id
}

type Pin struct{ Subject, Observation, Source []byte }

type Line struct {
	Property      vocabkey.PropertyKey   // the incoming value's property
	Incoming      propertycompare.Value  // Ref is the incoming Observation
	Outcome       propertycompare.Outcome
	PropertyCause propertycompare.Cause
	Relation      valuecompare.Relation
	Cause         valuecompare.Cause
	Points        float64
	Pinnable      bool
	Checked       bool
	Members       []Pin // one per member, member order
}

type Section struct {
	Key   vocabkey.PropertyKey // D-6
	Lines []Line
}

type Bundle struct {
	Member, Neighbor, Source []byte
	Observations             [][]byte
}

type Relationship struct {
	Key      vocabkey.BridgeKey // from the row Subject's end
	Neighbor []byte             // the layer neighbor
	Outcome  propertycompare.Outcome
	Cause    segmentcompare.Cause
	Points   float64 // segment evidence
	Pinnable bool
	Checked  bool
	Incoming [][]byte // bundled incoming Observations
	Members  []Bundle
	Sections []Section // the neighbor's other properties
}

type Evidence struct {
	Sections      []Section
	Relationships []Relationship
	Sources       map[string]string // Source id → title; the loader fills it
}

type LayerLink struct {
	Key          vocabkey.BridgeKey // from the row Subject's end
	Observations [][]byte           // on the layer bridge Subjects
}

type Bridge struct {
	Subject, Source, Member, Neighbor []byte
	Observations                      [][]byte
}

type CanonLink struct {
	Key     vocabkey.BridgeKey // from the handle's end
	Bridges []Bridge           // the association's accepted bridge Subjects
}

type Neighbor struct {
	Subject []byte
	Kind    vocabkey.SubjectType
	Values  subjectcompare.Values // the layer neighbor's
	Handle  subjectcompare.Values // its handle's profile
	Members []Member              // its handle's accepted members
	Layer   []LayerLink
	Canon   []CanonLink
}

func Build(kind vocabkey.SubjectType, row, handle subjectcompare.Values, members []Member, neighbors []Neighbor,
	metas subjectcompare.Metas, rules Rules, segment segmentcompare.Rules, subject subjectcompare.Rules,
	value valuecompare.Rules) Evidence

func Sections(kind vocabkey.SubjectType, incoming subjectcompare.Values, members []Member, metas subjectcompare.Metas,
	rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) []Section
```

`claimevidence` imports `propertycompare`, `subjectcompare`, `segmentcompare`, `valuecompare`, `vocabkey`, the standard library. No numeric literal other than 0 and 1 in non-test files.

`promotealign.Evidence` flow, reads only: `loadInputs` → `checkFixed` on the row and each neighbor (D-3) → `newProfiler`, `profile` for H and every H′ → `acceptedMembers` for H, every H′ and every link's association → `subjects.SourcesOf` for every member and member bridge → `g.Source` per Source → member sides (D-8), `LayerLink`s, `CanonLink`s with `Bridge`s → `claimevidence.Build` → `Evidence.Sources`. Catalog and cache errors are returned as is.

### FFI

None — S11-25.

### macOS

None — S11-25 and S11-20.

### Strings

None — no copy in Go.

## 7. Steps

**STEP-1 — One points helper, refs on plain-Subject values**
Files: `core/propertycompare/propertycompare.go`, `core/propertycompare/propertycompare_test.go`, `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go`, `core/subjectprofile/subjectprofile.go`, `core/subjectprofile/subjectprofile_test.go`, `core/database/promotealign/exhibits.go`. Skill: `add-match-rule`.
Do: D-4; D-8's `ForSubject` change. Update `TestPoints` (T-1); add T-2 and T-13. A `subjectprofile` test that asserts `ForSubject` output exactly gets the refs added, nothing else.
Check: `go test ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/graphalign/...` passes; `rg -n "Speaking\.Result\.Similarity" core --glob '!*_test.go'` prints only `core/propertycompare/propertycompare.go`; `rg -n "func similarity" core/database/promotealign` prints nothing.

**STEP-2 — The package and its registry section**
Files: `core/claimevidence/claimevidence.go` *(new)* (package doc: the claim evidence view's policy, pure; `Rules`, the types), `core/matchregistry/registry.go`, `core/matchregistry/evidence.go` *(new)*, `core/matchregistry/evidence_test.go` *(new)*, `core/matchregistry/inventory_test.go`, `core/matchregistry/guard_test.go`, `.cursor/rules/match-rules.mdc`, `docs/matching-model.md` (row `evidence.Defining`). Skill: `add-match-rule`.
Do: D-1, D-2. Add `{"core/claimevidence", true}` to `layerPackages`; add `core/claimevidence/**` to the rule's globs; delete `"evidence."` from `notYetBuilt`.
Check: `go test ./core/claimevidence/... ./core/matchregistry/...` passes, `TestKnobInventoryMatchesRegistry` and `TestLayerPackages` included.

**STEP-3 — Property lines**
Files: `core/claimevidence/sections.go` *(new)*, `core/claimevidence/sections_test.go` *(new)*, `core/claimevidence/invariants_test.go` *(new)*. Skill: `add-match-rule`.
Do: `Sections` per D-5, D-6, D-7, D-10, D-11.
Check: `go test ./core/claimevidence/...` passes with no build tags; T-3 and T-12 pass.

**STEP-4 — Shared inputs and association ids**
Files: `core/graphalign/types.go`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/layer.go`, `core/database/promotealign/profiles.go`. Skill: `use-catalog-session`.
Do: D-3's `loadInputs` and `checkFixed` (moved, `propose` and `mergeFixed` call them); D-8's `layerValues` signature; D-14. `loadLayer` keeps its three returns, so `export_test.go` is untouched.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./core/graphalign/...` passes with unchanged expectations.

**STEP-5 — Source lookup and the row's own lines**
Files: `core/database/subjects/sources.go` *(new)*, `core/database/subjects/subjects_test.go`, `core/database/promotealign/evidence.go` *(new)*, `core/database/promotealign/evidence_test.go` *(new)*. Skill: `use-catalog-session`.
Do: D-16; `Evidence` with D-3's validation, D-8, D-9, `Build` with no neighbors, and `Evidence.Sources`. Tests are in package `promotealign_test`, beside the fixture.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjects/... ./core/database/promotealign/...` passes; T-5, T-10, T-14 pass.

**STEP-6 — Relationships and bundles**
Files: `core/claimevidence/build.go` *(new)*, `core/claimevidence/build_test.go` *(new)*, `core/database/promotealign/evidence.go` *(new)* (STEP-5), `core/database/promotealign/evidence_test.go` *(new)* (STEP-5). Skill: `add-match-rule`, `use-catalog-session`.
Do: `Build` per D-12, D-13 and D-11's relationship order, with T-16 first; then the loader fills `Neighbor`s with `canonKeys` and calls `Build`. Never call `bridgeNeighbors`, `loadExhibitObservations` or `autoreconcile.Compatible`.
Check: `go test ./core/claimevidence/...` passes with no build tags; `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes; T-6, T-7, T-9, T-16 pass.

**STEP-7 — Neighbors the page did not map**
Files: `core/database/promotealign/evidence_test.go` *(new)* (STEP-5).
Do: AC-2.4 as T-8; fix `evidence.go` only if T-8 fails.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestEvidence ./core/database/promotealign/` passes.

**STEP-8 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md` (the `evidence` section; `core/claimevidence` in its package list and Check command), `docs/matching-model.md` (the sentence on the `evidence` section names `core/claimevidence`), `docs/deployment-plan/spike-11/claim-evidence.md` (Status: query built in S11-19; Open questions answered by **What stays**), `docs/ideas/matching-test-suite.md` (known gaps: move `neighborTypeProperty` into the segment section, review M-6), `docs/deployment-plan/spike-11/deployment-plan.md` (map row status), `docs/deployment-plan/spike-11/completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-19-claim-evidence-query.md` reports no errors.

Commits: one per step, STEP-6 and STEP-7 together. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Read a property line's points from one helper."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectcompare/subjectcompare_test.go` `TestPoints` | Results agree / partial 0.75 / conflict / unknown → 2.5 / 1.875 / −3 / 0 | AC-3.1, D-4 |
| T-2 | Go | `core/propertycompare/propertycompare_test.go` `TestResultSimilarity` *(new)* | Zero result → 0; speaking similarity 0.75 → 0.75 | AC-3.2 |
| T-3 | Go | `core/claimevidence/sections_test.go` `TestSections` *(new)*, table-driven | Each AC-1.x row → its sections, lines, state, members, points | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, D-5, D-6, D-7, D-10 |
| T-4 | Go | `core/matchregistry/evidence_test.go` `TestEvidenceSection` *(new)*; `inventory_test.go` `TestKnobInventoryMatchesRegistry` | `Default().Evidence` → Checked `same`, `equivalent`; Pinnable `contains`, `overlaps`, `near`; Defining event `event_type`; Checked and Pinnable disjoint; the inventory test passes with row `evidence.Defining` | D-2 |
| T-5 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceCollapsesFiveNamedMembers` *(new)* | Handle with five accepted "John Smith" members; Source person "John Smith" → one `name` line with five pins | AC-1.1, D-8, slice Check |
| T-6 | Go | same file `TestEvidenceBundlesAMatchingBirth` *(new)* | AC-2.1, AC-2.2 as subtests | AC-2.1, AC-2.2, D-12, D-13, slice Check |
| T-7 | Go | same file `TestEvidenceSpouseNeverMeetsSibling` *(new)* | AC-2.3 | AC-2.3 |
| T-8 | Go | same file `TestEvidenceOnlyMappedNeighbors` *(new)* | AC-2.4: New, Skip, absent, not a layer neighbor | AC-2.4, D-3 |
| T-9 | Go | same file `TestEvidenceRelationshipBundles` *(new)* | AC-2.5 (two bundles), AC-2.6 (no dates, no section) | AC-2.5, AC-2.6 |
| T-10 | Go | same file `TestEvidenceInvalidRequest` *(new)* | AC-2.7 rows → `promote.ErrInvalid` | AC-2.7, D-3 |
| T-11 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` | Lists `core/claimevidence`; passes | D-1 |
| T-12 | Go | `core/claimevidence/invariants_test.go` `TestSectionsInvariants` *(new)*, fixed seed | Random names and dates over 3–6 members: reordering members reorders only `Members`, and lines tied on the incoming `Ref`; adding a member with no values changes nothing; every pin names a ref the member holds; no line is both a conflict and pinnable | D-5, D-7, D-10 |
| T-13 | Go | `core/subjectprofile/subjectprofile_test.go` `TestForSubjectEchoesObservationRefs` *(new)* | Candidates with ids → each value's `Ref` is its id; a lifespan value's `Ref` is empty | D-8 |
| T-14 | Go | `core/database/subjects/subjects_test.go` `TestSourcesOf` *(new)*; `evidence_test.go` `TestEvidenceNamesEverySource` *(new)* | Two Subjects on two Sources → each id maps to its Source and title; AC-2.8 | AC-2.8, D-16 |
| T-16 | Go | `core/claimevidence/build_test.go` `TestBuild` *(new)*, table-driven | Hand-built `Neighbor`s for AC-2.1–AC-2.6: birth agree with bundle ids; birth 1830 partial; sibling key against spouse key → none; no neighbor → none; two bridges → two bundles; a bridge whose `Member` is not in `members` → skipped; no dates → relationship with no section | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, D-12, D-13 |
| T-15 | Go | `core/subjectcompare/...`, `core/graphalign/...`, `core/database/promotealign/...` existing tests | Unchanged expectations → pass | AC-3.3, D-3, D-4, D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/subjectcompare` and `core/graphalign` (D-4), `promotealign_test.go` (D-3, D-14), `subjectprofile` tests (D-8), `TestKnobInventoryMatchesRegistry` (D-2).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without Xcode, report the Go and plan-doc steps and say Xcode was not run
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings — none: no copy
- [ ] UI matches the brief inventory — none: no UI
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — none: not gated
- [ ] Domain docs: claim-evidence.md status, matching-model.md evidence sentence, add-match-rule skill, ideas/matching-test-suite.md line
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None — no UI and no RPC; nothing in the app changes until S11-25. The reviewer runs `CGO_ENABLED=1 go test -tags fts5 -run 'TestEvidence' -v ./core/database/promotealign/` and reads T-5 and T-6's output.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Loading every member's whole Source graph is slow for a member on a large Source | Lazy and per row; the cache keeps each Source. S11-21 times it |
| `loadInputs` builds every layer Subject's values for one row | The same cost as a proposal's layer load, without `Align`; S11-21 times it |
| D-4 changes a float sum's order in `Compare` | AC-3.3: every existing test passes unchanged |

**Stop and report if:**

- A `subjectcompare`, `graphalign` or `promotealign` test changes value after STEP-1 or STEP-4.
- `identity_claim_evidence`'s primary key is not `(identity_claim_id, observation_id)` at your base.
- A filed association's node `Members` are not its bridge Subjects (D-13 cannot find member bundles).
- A member's `SourceGraph.Observations` lacks a member bridge's endpoint Observations (`ValueSubjectID`).
- `TestKnobInventoryMatchesRegistry` needs a row name other than `evidence.Checked`, `evidence.Pinnable`, `evidence.Defining`.
- A non-test file in `core/claimevidence` needs an import under `core/database` or of `core/matchregistry`.

## 12. Open questions

None. Review M-4 was answered by the user on 2026-10-21: S11-20 owns the engine part of the lifespan and likeness lines.

## 13. PR

**Title:** S11-19 — Claim evidence query
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-19-claim-evidence-query.md`.
