# S11-19 — Claim evidence query

**Status:** Draft
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
**Planned:** 2026-10-21 against `main` @ `e169b71`
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

One Promote row's claim evidence is computed by its own Go query, from the row's target and its neighbors' targets as the page holds them. Every accepted member of the target entity gets its own line per property, members with an identical value collapse into one line, and a relationship that corresponds through segment compare is one line that bundles every Observation behind it. Which lines are pinnable and which start checked comes from a new registry section, `evidence`. Today `pairExhibits` shows one member per incoming value, and `attachExhibits` runs inside every proposal ([`claim-evidence.md`](../claim-evidence.md#why)).

**Why now.** Slice 5's first PR. S11-16 landed layer 6, so segment compare and entity compare are the layers the query composes. S11-25 puts the query behind a lazy RPC and moves Apply and Swift onto it; it needs the query first ("Build, then migrate").

**What stays.** `Propose`, `attachExhibits`, `pairExhibits`, `withRowLines`, `refLine`, `PromoteGraphAlignmentRow.comparisons` (field 9), the Promote handler, Swift and the sheet: all unchanged in behavior until S11-25. The pin tables: `identity_claim_evidence` (`core/database/migrations/000032.sql`, rebuilt in `000033.sql`) has primary key `(identity_claim_id, observation_id)` and stores no incoming → member pairing, so one incoming Observation pins against many member Observations, and `identityclaims.PinTx` writes a shared incoming pin once (`INSERT … ON CONFLICT DO NOTHING`). No migration PR is cut ahead of this one.

## 2. User stories

Conventions. `name(f)` is `propertycomparetest.Name(f)`, `term(k)` is `propertycomparetest.Term(k)`, `date(y)` is `propertycomparetest.Date(datevaluetest.YMD(y, 0, 0))`; `ref(v, b)` is `propertycomparetest.Ref(v, b)`. Metas are `subjectcomparetest.Metas()`. Rules are `matchregistry.Default()`: `reg.Evidence`, `reg.Subject`, `reg.Value`. "Checked" and "pinnable" are `Line.Checked` and `Line.Pinnable`. Points are within 1e-9.

**US-1.** As the engine, I compare one incoming Observation with each member's values, so every member that agrees is evidence.

`claimevidence.Sections(kind, incoming, members, metas, reg.Evidence, reg.Subject, reg.Value)` (D-5).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person; incoming `name` "John Smith"; five members, each with `name` "John Smith" | One section `name`, one line: relation `same`, outcome agree, checked, pinnable, five `Members` in member order, points `subject.Weights.person.name.Agree` = 2.5 |
| AC-1.2 | Incoming "John Smith"; members "John Smith", "John Smith", "Jon Smith" | Two lines: `same` with two members (checked); `near` cause `spelling` with one member (pinnable, unchecked) |
| AC-1.3 | Incoming "José Smith"; member "Jose Smith" | One line: `equivalent` cause `fold`, checked, pinnable |
| AC-1.4 | Incoming "J. Smith"; member "James Smith" | One line: `contains` cause `initial`, pinnable, unchecked, points 2.5 × the speaking similarity |
| AC-1.5 | Event; incoming `date` 1817; member `date` 1830 | One line: outcome conflict, relation `different`, not pinnable, not checked, points −`subject.Weights.event.date.Conflict` |
| AC-1.6 | Incoming `name` "John Smith"; one member with `name`, one member with no `name` | One line, one member: a member with nothing to compare has no line (outcome unknown) |
| AC-1.7 | Incoming negative `name` "John Smith" (the record denies it); member "John Smith" | One line: outcome conflict, cause `negative`, not pinnable |
| AC-1.8 | Incoming `gender` `man`; member `sex_at_birth` `male` | One line in section `sex_at_birth` (the lead): outcome agree, checked; `Line.Property` is `gender` |
| AC-1.9 | Event; incoming `start_date` 1817; member `start_date` 1817 and `date` 1817 | One line in section `date`: the incoming `start_date` against the member's `start_date`, relation `same`; the member's `date` is not compared with it (D-6) |
| AC-1.10 | Person; incoming `occupation` (a property with no lead) "farmer"; member "farmer" | Section `occupation`, one checked line, points 0 |

**US-2.** As the engine, I bundle a corresponding relationship into one line, so a researcher can say "they have a matching birth event" and check what is behind it.

`promotealign.Evidence(c, req)` (D-3) on a catalog built with the `promotealign_test.go` fixture.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A filed person P (one member, birth event E with `date` 1817, P subject of E). A Source with person "John Smith", a birth dated 1817, subject participation. Request: the person on P, neighbor the birth on E | One relationship: key `participation·subject·person·birth`, outcome agree, checked; `Incoming` holds the Source participation Subject's Observations and the Source birth's `event_type` Observation; one `Members` bundle with the filed participation's Observations and the filed birth's `event_type` Observation; one section `date` under it with one checked line |
| AC-2.2 | AC-2.1 with the Source birth dated 1830 | The relationship is partial: pinnable, unchecked. Its `date` line is a conflict, not pinnable |
| AC-2.3 | Filed P has a spouse S1 (relationship `spouse`); the Source person has a sibling, neighbor mapped to S1 | No relationship: `sibling` does not correspond to `spouse` |
| AC-2.4 | AC-2.1 with the neighbor's target New, or the neighbor not in the request | No relationship |
| AC-2.5 | Filed P has two accepted members, each the subject of a birth that is a member of E | One relationship with two `Members` bundles |
| AC-2.6 | AC-2.1 with no `date` on either birth | One relationship, agree, checked, with no section under it (decided 2026-10-17: a relationship line works with nothing under it) |
| AC-2.7 | Request with a subject not on the Source, a handle of another kind, or a merged handle | `promote.ErrInvalid` |

**US-3.** As the engine, a property line's points come from one helper, so no caller reads the speaking similarity itself (S11-13 MNT-3).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `subjectcompare.Points(res, Weights{Agree: 2.5, Conflict: 3})` with `res` agree; partial with speaking similarity 0.75; conflict; unknown | 2.5; 1.875; −3; 0 |
| AC-3.2 | `propertycompare.Result{}.Similarity()`; a result whose speaking pair has similarity 0.75 | 0; 0.75 |
| AC-3.3 | Every existing `subjectcompare` and `graphalign` test | Passes unchanged: `Compare`'s similarity and evidence are the same numbers |

## 3. Scope

| In | Covered by |
| --- | --- |
| `propertycompare.Result.Similarity` *(new)*; `subjectcompare.Points` takes a `propertycompare.Result`; `Compare` builds score from `Points`; `promotealign`'s private `similarity` deleted | STEP-1, AC-3.1–AC-3.3 |
| `core/claimevidence` *(new)*: pure types, `Rules`, `Sections`, line state, collapse; on the guard list and the match rule globs | STEP-2, STEP-3, AC-1.1–AC-1.10 |
| Registry `evidence` section: `Registry.Evidence`, `core/matchregistry/evidence.go` *(new)*; `notYetBuilt` loses `"evidence."` | STEP-2, T-4 |
| `promotealign.Evidence` *(new)*: the loader, member records from the members' Source graphs through `layerValues` and `decode`, relationships through segment compare with bundles | STEP-4, STEP-5, AC-2.1–AC-2.7 |
| `loadLayer` / `loadLayerBridges` return each layer edge's association ids | STEP-4 |
| Docs: `add-match-rule` evidence section, `claim-evidence.md` status, `matching-model.md`, the map | STEP-6 |

| Out | Home |
| --- | --- |
| The RPC, `engine.proto` messages, the handler | S11-25 |
| Default pins at Apply; Swift `draftedPins`, `draftedArgument`, `kept.pins` | S11-25 |
| Deleting `attachExhibits`, `pairExhibits`, `withRowLines`, `refLine`, field 9, `bridgeNeighbors`, `loadExhibitObservations`, `exhibitObs`, the pin-test helper `obs` (S11-13 MNT-5, PRF-2) | S11-25 |
| Swift parsing of relation and cause codes (`ConclusionVocabulary.swift`) | S11-25 |
| Copy for lines and relationships; the "why suggested" summary; the sheet's order of sections and relationships beyond D-11's | S11-20, S11-D2 |
| The lifespan line (`Comparison.Derived`) | S11-20 |
| Likeness lines (depiction values) | S11-20 (after S10-54) |
| Provisional members on the sheet | not planned (D-9: accepted members only, as `acceptedMembers` today) |
| A bridge's own property lines (dates on a place relationship): bundled, never compared | not planned ([`claim-evidence.md`](../claim-evidence.md#evidence-through-relationships)) |
| Timing the query on a large Source | S11-21 |

## 4. Current state

Verified at `e169b71`.

| Area | Where | Today |
| --- | --- | --- |
| Sheet lines | `core/database/promotealign/exhibits.go` `attachExhibits` | Runs after `graphalign.Align` in every `propose` (`promotealign.go:92`). Pairs each mapped row's own records with its handle's accepted members, and each mapped neighbor's records with members one bridge from the row's members (`bridgeNeighbors`, SQL over `observations`, any bridge type, no key check). Unchanged by this PR |
| Pairing | `exhibits.go` `pairExhibits` | One line per incoming Observation: the first `autoreconcile.Compatible` member value, else the most similar. Reads `similarity(res)` and calls `subjectcompare.Points(chosenRes.Outcome, sim, …)`. This PR changes only those two reads (D-4) |
| Similarity helper | `exhibits.go` `similarity` | The speaking pair's similarity, 0 when none. Deleted by D-4 |
| Second Observation read | `exhibits.go` `loadExhibitObservations` | SQL read of Observations with sources and terms, building `autoreconcile.Value` beside `decode` (S11-13 MNT-4). The new query does not use it (D-8); S11-25 deletes it |
| Row decoder | `core/database/promotealign/profiles.go` `decode`, `layerValues` | `layerValues(sg, primary, props, terms)` returns each primary Subject's product Observations, both polarities, as `autoreconcile.Candidate` (`ObservationID`, `Value`, `Negative`), keyed by property, in Observation id order; values via `decode`. Only product properties (`OriginProvenencia`), no Subject-valued Observations |
| Profile value | `core/subjectprofile/subjectprofile.go` `Value` | Builds a `propertycompare.Value` from an `autoreconcile.Value`; leaves `Ref` empty |
| Layer load | `core/database/promotealign/layer.go` `loadLayer`, `loadLayerBridges` | Builds every primary Subject's values (`subjectprofile.ForSubject`) and every bridge's keys from both ends (`segmentcompare.Key(linkOf(…))`) through `canonicalgraph.WalkSource`, whose `Edge.Association` is the bridge Subject id. The association id is dropped; edges are deduped by `edgeSeenKey` |
| Canon keys | `core/database/promotealign/canon.go` `canonKeys` | Reads a `graphcache.Link` as an edge with its key from each end. The one keying rule for the canon side |
| Handle profiles | `profiles.go` `newProfiler`, `profiler.profile`, `profiler.values` | Builds a handle's strict profile (and a person's life) from the cache |
| Members | `core/database/graphcache/graph.go` `Node.Members`, `Member{SubjectID, Accepted}`; `exhibits.go` `acceptedMembers` | Accepted and provisional claims in claim id order. No Source id |
| Source graph | `core/database/graphcache/source.go` `Graph.Source`, `SourceGraph`, `SourceObservation` | One Source's Subjects and Observations (with `SubjectID`, `PropertyID`, `ValueSubjectID`, `Polarity`), cached. No title |
| Value refs | `core/valuecompare/valuecompare.go` `Value.Ref`, `Detail.RefA` / `RefB` | `Compare` echoes refs; `propertycompare.SpeakingPair.Detail` carries them. Nothing reads them yet |
| Points | `core/subjectcompare/subjectcompare.go` `Points(outcome, similarity, w)`, `Compare` | `Compare` reads `res.Speaking.Result.Similarity` for partial and repeats the agree / partial arithmetic inline for `score` |
| Pools | `core/subjectcompare/pairs.go` `Side` | A lead's values: its own, pooled (terms mapped by `Pool.Terms`), or a period built from start and end when the lead has no positive value |
| Leads | `subjectcompare.go` `Rules.Lead` | The lead a property counts under; `false` for a property no lead weighs or pools |
| Segment compare | `core/segmentcompare/segmentcompare.go` `Compare`, `Segment`, `Result`; `key.go` `Corresponds` | `Compare(a, b, placed, own, metas, rules, subject, value)`: agree, or partial when the neighbors conflict; unknown when not corresponding or gated. `Result.Neighbor` is layer 3's verdict |
| Registry | `core/matchregistry/registry.go` `Registry`, `Default` | Sections `Value`, `Subject`, `Segment`, `Alignment`. No `Evidence` |
| Inventory guard | `core/matchregistry/inventory_test.go` `notYetBuilt`, `TestKnobInventoryMatchesRegistry` | `notYetBuilt = {"evidence."}` keeps `docs/matching-model.md` rows `evidence.Checked`, `evidence.Pinnable`, `evidence.Defining.event` legal with no field |
| Layer guard | `core/matchregistry/guard_test.go` `layerPackages`, `TestLayerPackages` | Purity and numeric-literal checks over a list of packages |
| Pin tables | `core/database/migrations/000032.sql`, `000033.sql` `identity_claim_evidence`; `core/database/identityclaims/pins.go` `PinTx`; `core/database/promote/promote.go` `pinPairs` | See **What stays**. `pinPairs` pins both Observations of a pair on both claims |
| Tests | `core/subjectcompare/subjectcompare_test.go` `TestPoints`; `core/database/promotealign/promotealign_test.go` `newFixture`, `person`, `birth`, `participation`, `relationship`, `join`, `promote`, `datedBirth`; `exhibits_pin_test.go` | `TestPoints` calls the old signature |

**Precedent.** S11-14 ([#493](https://github.com/mendahu/provenencia/pull/493), [plan](S11-14-segment-compare.md)) added a pure package with its `Rules` type, a registry section file with a source comment per value, inventory rows and the guard line, and extended `add-match-rule`. Copy that shape for `core/claimevidence` *(new)* and `core/matchregistry/evidence.go` *(new)*. The loader shape is `loadLayer` / `loadCanon`: catalog reads in `promotealign`, pure decisions in the layer package. Do not copy `pairExhibits`' "first compatible, else closest" choice.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A new pure package, `core/claimevidence`, holds the evidence types, `Rules`, `Sections` and the line policy. The loader is `core/database/promotealign/evidence.go` *(new)*. `claimevidence` is on `layerPackages` with `numbers: true` and in `match-rules.mdc` globs | The policy is the view's own ([`claim-evidence.md`](../claim-evidence.md#separate-from-matching)) and the registry must name its `Rules` type without a cycle, as for every section. Pure means table tests without build tags | In `promotealign` — needs `-tags fts5` to test pure policy, and the registry cannot import it. In `graphalign` — layer 6 decides placement; evidence is a separate view |
| D-2 | `claimevidence.Rules{Checked, Pinnable []valuecompare.Relation; Defining map[vocabkey.SubjectType][]vocabkey.PropertyKey}`. `matchregistry.Default().Evidence`: Checked `Same`, `Equivalent`; Pinnable `Contains`, `Overlaps`, `Near`; Defining event: `event_type`. The inventory rows already exist; delete `"evidence."` from `notYetBuilt` | The keys must be the rows' (`evidence.Checked`, `evidence.Pinnable`, `evidence.Defining.event`); a slice of relations and a map of slices are leaves the inventory walk reads that way | A per-relation struct — renames every row |
| D-3 | Entry point: `func Evidence(c *database.Catalog, req EvidenceRequest) (claimevidence.Evidence, error)` *(new)*, with `EvidenceRequest{SourceID, SubjectID, HandleID []byte; Neighbors []graphalign.Fixed}` *(new)*. `Neighbors` are the page's targets for the row's layer neighbors, suggested or decided; only `TargetHandle` (or empty Target) entries with a 16-byte handle count. Invalid: a source that does not exist, a subject not a primary Subject on it, a handle that is missing, merged, or of another kind → `promote.ErrInvalid` (the checks `mergeFixed` runs) | Inputs are the page's, not matching internals ([`claim-evidence.md`](../claim-evidence.md#separate-from-matching)). `graphalign.Fixed` already says "Subject on handle, or New, or Skip"; reusing it saves a type S11-25 would map twice | A new `Target` struct — a copy of `Fixed`. Passing a `graphalign.Proposal` — matching internals |
| D-4 | `propertycompare.Result.Similarity() float64` *(new)*: the speaking pair's similarity, 0 when `Speaking` is nil. `subjectcompare.Points(res propertycompare.Result, w Weights) float64`: agree `w.Agree`, partial `w.Agree × res.Similarity()`, conflict `−w.Conflict`, unknown 0. `Compare` computes `pr.Points = Points(res, w)`, adds `max(pr.Points, 0)` to `score` and keeps its `counted` switch. `pairExhibits` calls `res.Similarity()` and the new `Points`; `promotealign.similarity` is deleted | The scope entry and S11-13 MNT-3: one place reads the speaking similarity. The method lives on the type that owns the speaking pair | A second helper beside `Points` in `subjectcompare` — still leaves `pairExhibits` reading the field |
| D-5 | `claimevidence.Sections(kind, incoming []Record, members []Member, metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) []Section`. For each incoming record (positive or negative), for each member, one `propertycompare.Compare` call (D-6) gives the line; an unknown outcome gives no line. The member Observation is the speaking pair's `Detail.RefB` | One line per member per property ([Lines](../claim-evidence.md#lines)); comparisons use layer 2 only, never `Compatible` or Promote's scores | Comparing the incoming value with the entity's profile — loses which member agrees |
| D-6 | What a record is compared with. A property with a lead that pools it, or a lead that has pools: both sides go through `subjectcompare.Side(kind, lead, …)` (the incoming side with only that record), with `metas[lead]`. Every other property, period ends included, compares with the member's values of the same property, with `metas[property]` | `Side` is the one place pooled terms are mapped (S11-30). A period built by `Side` is a synthetic value with no `Ref`, so it cannot name a member Observation | Building periods per member — the line would pin no Observation |
| D-7 | Line state, `rules` applied to the speaking relation: checked when the outcome is agree and the relation is in `Checked`; pinnable when checked, or when the outcome is partial and the relation is in `Pinnable`. A conflict (relation `Different`, or cause `negative` or `cardinality`) is shown, not pinnable. Points: `subjectcompare.Points(res, subject.Weights[kind][lead])` when the property has a lead (`subject.Lead`), else 0 | The table in [Lines](../claim-evidence.md#lines); points from the one helper (D-4) | Reading similarity in the view |
| D-8 | Records come from Source graphs only: the row's Source graph for the incoming side, each member's Source graph for the member side, both through `layerValues` (and so `decode`), with `Value.Ref` and `Record.Observation` set to `Candidate.ObservationID` and `subjectprofile.Value(c.Value, c.Negative, namePattern)` as the value. Each member Subject's Source id and every Source title come from one batched read, `memberSources(q, ids)` *(new)*: `SELECT s.id, s.source_id, src.title FROM subjects s JOIN sources src ON src.id = s.source_id WHERE s.id IN (…)` through `sqlbatch.Query` | The scope entry and S11-13 MNT-4: one row decoder, no second Observation read. A Subject's Source id is not in the cache; one small read is cheaper than a cache shape change | `loadExhibitObservations` — the second reader MNT-4 retires. `graphcache.Member.SourceID` — changes the cache, its load and `Verify` for one reader |
| D-9 | Members are the handle's accepted members (`acceptedMembers`), in claim id order; provisional members are left out | Today's rule, and pins backfill onto accepted claims only (`promote.pairClaim`) | Provisional members — a pin could not backfill |
| D-10 | Collapse: within a section, lines with the same incoming Observation, relation and value cause, whose member values have the same `valuecompare.Key(v, valuecompare.Same, value)`, are one line whose `Members` lists each member in member order. A value with no key (`ok` false) does not collapse | "Identical value" is the Same key by the working rule "A normalized `Key` per value type"; `Key` and `Compare` say `Same` together | Comparing display text — the six text rules S11-07 retired |
| D-11 | Order. Sections by ascending lead key; lines by incoming Observation id, then by the first member's claim order. Relationships: keys in `segment.One` first, then by `BridgeKey.String()`, then by the layer neighbor's ref | Deterministic, testable output. The sheet's own order is S11-D2's | Hand-listing "core" relationships here — a list S11-D2 has not drawn |
| D-12 | Relationships. For each layer edge of the row's Subject whose other end is a neighbor with a handle target H′: every canon link of the handle H to H′ (`g.Node(H).Links`) whose key from H's end (`canonKeys`) `Corresponds` to the layer key from the Subject's end. One `Relationship` per (layer neighbor, key), with a bundle per corresponding member segment from every such link. Its state: `segmentcompare.Compare(layer, canon, segmentcompare.Together, own, …)` where `own` is `subjectcompare.Compare` of the Subject's layer values and H's profile; agree → checked and pinnable; partial → pinnable; unknown → no relationship | [Evidence through relationships](../claim-evidence.md#evidence-through-relationships): a section exists only when the segments correspond; segment compare decides. The page placed the neighbor, so the placement is `Together` | Every bridge between members, as `bridgeNeighbors` — the spouse-meets-sibling defect |
| D-13 | Bundles. Incoming: every Observation on the layer edge's association Subjects (`SourceObservation.SubjectID` = the association), plus the neighbor's Observations of `rules.Defining[neighbor kind]`. Member: for each canon link's association node, each accepted member bridge Subject b (`g.Node(link.Association).Members`); b's Observations from its Source graph; the endpoints (`ValueSubjectID`) must be an accepted member m of H and an accepted member n of H′, else b is skipped; the bundle is b's Observations plus n's defining Observations, with `Bundle.Member` = m and `Bundle.Neighbor` = n. The relationship's sections are `Sections(neighbor kind, the neighbor's non-defining records, the n's as members, …)` | The registry rule "a segment's bundle is all of its bridge's Observations plus the neighbor kind's defining properties". Member lines under it compare this birth with that birth only | Bundling the neighbor's other properties — they are their own checkable lines |
| D-14 | `loadLayerBridges` returns, beside the bridges, `map[seenEdge][][]byte` of every association id behind each deduped edge; `loadLayer` passes it out as a fourth return value, and `propose` ignores it. The evidence loader calls `loadLayer` for the layer's values, bridges and associations | One reader of the layer's bridges and keys; `loadLayer` already builds each Subject's values once | A second in-memory bridge reader over the Source graph — a second keying path beside `linkOf` |
| D-15 | Wording: a line carries its relation and value cause (`valuecompare.Cause`), its property cause (`propertycompare.Cause`), and a relationship carries its segment cause (`segmentcompare.Cause`). No separate wording-key vocabulary | [Wording](../claim-evidence.md#wording): lines are built from relation and `Detail.Cause`; S11-25 parses these codes in `ConclusionVocabulary.swift`. Every row of the wording table has its code (`fold`, `alias`, `precision`, `missing_part`, `spelling`, `gap`) | A Go `Wording` code set — a second vocabulary to mirror in Swift beside the cause codes |
| D-16 | `Evidence.Sources map[string]string` holds the title of every Source a record or bundle names, keyed by Source id as a string | Each line names the member's Source ([Lines](../claim-evidence.md#lines)); titles are display data the pure package does not read | A title on each `Pin` — repeats the title per line |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/claimevidence` *(new)* | package | `core/database/promotealign/exhibits.go` | Needs build tags and a database import; the registry cannot name its types |
| `claimevidence.Rules` *(new)* | type | `segmentcompare.Rules` | Each section's type lives with the code that reads it |
| `claimevidence.Record`, `Member`, `Pin`, `Line`, `Section`, `Bundle`, `Relationship`, `Evidence` *(new)* | types | `graphalign.Exhibit` | `Exhibit` is one pair per incoming value with display strings; S11-25 deletes it with field 9 |
| `claimevidence.Sections` *(new)* | func | `pairExhibits` | Replaces its pairing in S11-25; built on `propertycompare.Compare` and `subjectcompare.Side` |
| `propertycompare.Result.Similarity` *(new)* | method | `promotealign.similarity` | Moves it to the type's owner; the private copy is deleted |
| `matchregistry` `defaultEvidence`, `Registry.Evidence` *(new)* | func, field | `defaultSegment` | The section file the registry design names (`evidence.go`) |
| `promotealign.Evidence`, `EvidenceRequest` *(new)* | func, type | `Propose` | Lazy, per row, from page targets; does not run `Align` |
| `promotealign.memberSources` *(new)* | func | `loadExhibitObservations` (joins `sources`) | Reads only Source ids and titles; Observations come through `layerValues` |
| `evidence_test.go` in `core/claimevidence` and `core/database/promotealign` *(new)* | tests | `promotealign_test.go` fixture | Reused: `newFixture`, `person`, `birth`, `participation`, `relationship`, `join`, `promote` |

**Builds on.** `propertycompare.Compare` and its value refs; `subjectcompare.Side`, `Lead`, `Points`; `segmentcompare.Compare` and `Corresponds`; `loadLayer`, `layerValues`, `decode`, `canonKeys`, `newProfiler`, `acceptedMembers`; the `add-match-rule` steps for a section and a package.
**Deletes.** `promotealign.similarity`; the inline partial arithmetic in `subjectcompare.Compare`; `"evidence."` in `notYetBuilt`. The exhibit code is deleted in S11-25, not here ("Build, then migrate").
**Next adds.** S11-25 maps `claimevidence.Evidence` to the RPC response, and Apply's default pins are the `Checked` lines' and relationships' pins. A new defining property is one `Defining` entry. S11-20 words lines from the codes D-15 carries.

### Schema

None — `identity_claim_evidence` already holds one incoming Observation against many member Observations (see **What stays**).

### Go core

| Package | Change |
| --- | --- |
| `core/propertycompare` | `func (r Result) Similarity() float64` *(new)* (D-4) |
| `core/subjectcompare` | `func Points(res propertycompare.Result, w Weights) float64` (new signature); `Compare` uses it (D-4) |
| `core/claimevidence` *(new)* | Types and `Sections` below |
| `core/matchregistry` | `Registry.Evidence claimevidence.Rules`; `evidence.go` *(new)* `defaultEvidence()` with a source comment per value (`decided`, citing [`claim-evidence.md`](../claim-evidence.md#lines)) |
| `core/database/promotealign` | `evidence.go` *(new)*: `EvidenceRequest`, `Evidence`, `memberSources`; `layer.go`: association ids (D-14); `exhibits.go`: D-4 call sites only |

```go
// package claimevidence — sketch

type Rules struct {
	Checked  []valuecompare.Relation // lines that start checked
	Pinnable []valuecompare.Relation // pinnable, unchecked
	Defining map[vocabkey.SubjectType][]vocabkey.PropertyKey // bundled into a relationship line
}

type Record struct {
	Observation, Subject, Source []byte
	Key                          vocabkey.PropertyKey
	Value                        propertycompare.Value // Ref is Observation
}

type Member struct {
	Subject []byte
	Records []Record
}

type Pin struct{ Subject, Observation, Source []byte }

type Line struct {
	Property      vocabkey.PropertyKey // the incoming record's property
	Incoming      Record
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
	Key   vocabkey.PropertyKey // the lead, or the property when it has none
	Lines []Line
}

type Bundle struct {
	Member, Neighbor []byte // member of the handle, member of the neighbor's handle
	Source           []byte
	Observations     [][]byte
}

type Relationship struct {
	Key      vocabkey.BridgeKey   // from the row Subject's end
	Neighbor []byte               // the layer neighbor
	Outcome  propertycompare.Outcome
	Cause    segmentcompare.Cause
	Points   float64              // segment evidence
	Pinnable bool
	Checked  bool
	Incoming [][]byte             // bundled incoming Observations
	Members  []Bundle
	Sections []Section            // the neighbor's other properties
}

type Evidence struct {
	Sections      []Section
	Relationships []Relationship
	Sources       map[string]string // Source id → title
}

func Sections(kind vocabkey.SubjectType, incoming []Record, members []Member, metas subjectcompare.Metas,
	rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) []Section
```

`claimevidence` imports `propertycompare`, `subjectcompare`, `segmentcompare` (for `Cause` only), `valuecompare`, `vocabkey`, the standard library. No numeric literal other than 0 and 1.

`promotealign.Evidence` flow: validate (D-3) → `valuerules.Load` → `matchregistry.Default()` with `Value` replaced as `propose` does → `loadTerms` → `loadLayer` (values, bridges, associations; D-14) → `newProfiler` and `profile` for H and every neighbor H′ → `acceptedMembers` for H and every H′ → `memberSources` for all members and member bridges → `g.Source` per member Source → `layerValues` per Source for the member Subjects → `Sections` for the row → relationships (D-12, D-13) → `Evidence.Sources` (D-16). Errors: catalog and cache errors are returned as is.

### FFI

None — S11-25.

### macOS

None — S11-25 and S11-20.

### Strings

None — no copy in Go.

## 7. Steps

**STEP-1 — One points helper**
Files: `core/propertycompare/propertycompare.go`, `core/propertycompare/propertycompare_test.go`, `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go`, `core/database/promotealign/exhibits.go`. Skill: `add-match-rule`.
Do: D-4. Update `TestPoints` to build results (T-1); add T-2.
Check: `go test ./core/propertycompare/... ./core/subjectcompare/... ./core/graphalign/...` passes; `rg -n "Speaking\.Result\.Similarity" core --glob '!*_test.go'` prints only `core/propertycompare/propertycompare.go`; `rg -n "func similarity" core/database/promotealign` prints nothing.

**STEP-2 — The package and its registry section**
Files: `core/claimevidence/claimevidence.go` *(new)* (package doc: the claim evidence view's policy, pure; `Rules`, the types), `core/matchregistry/registry.go`, `core/matchregistry/evidence.go` *(new)*, `core/matchregistry/evidence_test.go` *(new)*, `core/matchregistry/inventory_test.go`, `core/matchregistry/guard_test.go`, `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule`.
Do: D-1, D-2. Add `{"core/claimevidence", true}` to `layerPackages`; add `core/claimevidence/**` to the rule's globs; delete `"evidence."` from `notYetBuilt`.
Check: `go test ./core/claimevidence/... ./core/matchregistry/...` passes, `TestKnobInventoryMatchesRegistry` and `TestLayerPackages` included.

**STEP-3 — Property lines**
Files: `core/claimevidence/sections.go` *(new)*, `core/claimevidence/sections_test.go` *(new)*. Skill: `add-match-rule`.
Do: `Sections` per D-5, D-6, D-7, D-10, D-11.
Check: `go test ./core/claimevidence/...` passes with no build tags; T-3 covers AC-1.1–AC-1.10.

**STEP-4 — Association ids on the layer**
Files: `core/database/promotealign/layer.go`, `core/database/promotealign/promotealign.go`. Skill: `use-catalog-session`.
Do: D-14. `propose` discards the new return value.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes unchanged.

**STEP-5 — The query**
Files: `core/database/promotealign/evidence.go` *(new)*, `core/database/promotealign/evidence_test.go` *(new)*. Skill: `use-catalog-session`.
Do: D-3, D-8, D-9, D-12, D-13, D-16. Build relationships with `canonKeys` and `segmentcompare.Corresponds`; never call `bridgeNeighbors`, `loadExhibitObservations` or `autoreconcile.Compatible`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes; T-5–T-10 cover AC-2.1–AC-2.7.

**STEP-6 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md` (the `evidence` section and `core/claimevidence` in its section list, package list and Check command), `docs/matching-model.md` (the `evidence` rows' source column unchanged; the sentence at "the `evidence` section" names `core/claimevidence`), `docs/deployment-plan/spike-11/claim-evidence.md` (Status: query built in S11-19; Open questions answered by What stays), `deployment-plan.md` (map row; working rule "Build, then migrate" says `pairExhibits` is deleted in S11-25), `completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-19-claim-evidence-query.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Read a property line's points from one helper."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectcompare/subjectcompare_test.go` `TestPoints` | Results agree / partial 0.75 / conflict / unknown → 2.5 / 1.875 / −3 / 0 | AC-3.1, D-4 |
| T-2 | Go | `core/propertycompare/propertycompare_test.go` `TestResultSimilarity` *(new)* | Zero result → 0; speaking similarity 0.75 → 0.75 | AC-3.2 |
| T-3 | Go | `core/claimevidence/sections_test.go` `TestSections` *(new)*, table-driven | Each AC-1.x row → its sections, lines, state, members, points | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, D-5, D-6, D-7, D-10 |
| T-4 | Go | `core/matchregistry/evidence_test.go` `TestEvidenceSection` *(new)* | `Default().Evidence` → Checked `same`, `equivalent`; Pinnable `contains`, `overlaps`, `near`; Defining event `event_type`; Checked and Pinnable disjoint | D-2 |
| T-5 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceCollapsesFiveNamedMembers` *(new)* | Handle with five accepted "John Smith" members; Source person "John Smith" → one `name` line with five pins, each naming its member's Source in `Sources` | AC-1.1, D-8, D-16, slice Check |
| T-6 | Go | same file `TestEvidenceBundlesAMatchingBirth` *(new)* | AC-2.1, AC-2.2 as subtests | AC-2.1, AC-2.2, D-12, D-13, slice Check |
| T-7 | Go | same file `TestEvidenceSpouseNeverMeetsSibling` *(new)* | AC-2.3 | AC-2.3 |
| T-8 | Go | same file `TestEvidenceOnlyMappedNeighbors` *(new)* | AC-2.4: New, Skip, absent | AC-2.4, D-3 |
| T-9 | Go | same file `TestEvidenceRelationshipBundles` *(new)* | AC-2.5 (two bundles), AC-2.6 (no dates, no section) | AC-2.5, AC-2.6 |
| T-10 | Go | same file `TestEvidenceInvalidRequest` *(new)* | AC-2.7 rows → `promote.ErrInvalid` | AC-2.7 |
| T-11 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` | Lists `core/claimevidence`; passes | D-1 |
| T-12 | Go | `core/subjectcompare/...`, `core/graphalign/...` existing tests (`TestCompare`, `TestInvariants`, `align_test.go`) | Unchanged expectations → pass | AC-3.3, D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/subjectcompare` and `core/graphalign` (D-4), `promotealign_test.go` (D-14), `TestKnobInventoryMatchesRegistry` (D-2).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no copy
- [ ] UI matches the brief inventory — none: no UI
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; "Build, then migrate" names S11-25 for pairExhibits; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — none: not gated
- [ ] Domain docs: claim-evidence.md status, matching-model.md evidence sentence, add-match-rule skill
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None — no UI and no RPC; nothing in the app changes until S11-25. The reviewer runs `CGO_ENABLED=1 go test -tags fts5 -run 'TestEvidence' -v ./core/database/promotealign/` and reads T-5 and T-6's output.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Loading every member's whole Source graph is slow for a member on a large Source | Lazy and per row; the cache keeps each Source. S11-21 times it |
| `loadLayer` builds every Subject's values for one row | The same cost as a proposal's layer load, without `Align`; S11-21 times it |
| D-4 changes a float sum's order in `Compare` | AC-3.3: every existing test passes unchanged |

**Stop and report if:**

- A `subjectcompare` or `graphalign` test changes value after STEP-1.
- `identity_claim_evidence`'s primary key is not `(identity_claim_id, observation_id)` at your base.
- A canon association node's `Members` are not the bridge Subjects of that association, so D-13 cannot find member bundles.
- A member's `SourceGraph.Observations` lacks the member bridge's endpoint Observations (`ValueSubjectID`).
- `TestKnobInventoryMatchesRegistry` needs a row name other than `evidence.Checked`, `evidence.Pinnable`, `evidence.Defining.event`.
- `claimevidence` needs an import under `core/database` or of `core/matchregistry` to pass a test.

## 12. Open questions

None.

## 13. PR

**Title:** S11-19 — Claim evidence query
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-19-claim-evidence-query.md`.
