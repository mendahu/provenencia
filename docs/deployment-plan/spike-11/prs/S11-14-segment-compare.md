# S11-14 — Segment compare

**Status:** Blocked — on S11-13
**Spike / slice:** Spike 11 · Slice 4 — Graph alignment
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the wire fields, the page and its copy are unchanged)
**Board:** —
**Depends on:** S11-13 (Planned, [plan](S11-13-promote-subject-compare.md); not landed at drafting). Branch from `main` after S11-13 merges
**Blocks:** S11-16
**Branch:** `s11-14-segment-compare`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (a new layer package, a registry section, knobs and inventory rows; this PR extends the skill)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6), [Match registry](../comparison-layers.md#match-registry) (typed keys, `segment.go`), [Rules for every layer](../comparison-layers.md#rules-for-every-layer); [`docs/matching-model.md`](../../../matching-model.md) knob inventory (`segment.*` rows) and worked example step 5; working rules "Segment compare is one hop and pure", "Build, then migrate", "High tolerance for churn", "Sparse data first", "Registry numbers: carry over, then estimate", "One matching skill, grown with the layers", "One wire vocabulary"; user calls of 2026-10-19 written as D-5 and D-10
**Planned:** 2026-10-19 against `main` @ `6ecfd88` (S11-13's plan merged, S11-13 not implemented) — a draft; redo §4 and the self-review gate when S11-13 lands
**Review:** — (not reviewed yet: `review-plan` runs on the refresh after S11-13 lands, since §4 is written against S11-13's plan)

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

Promote judges a link between two Subjects with layer 5, segment compare, instead of matching edge signatures and measuring fan-out from the catalog. A segment is primary – bridge – primary. Its bridge reads as a typed `vocabkey.BridgeKey`, so a spouse never corresponds to a sibling, and "Mary parent of John" corresponds to "John child of Mary" with the ends swapped. Whether a link is one-cardinality is declared in the registry: subject of a birth and subject of a death, from both ends. A differing neighbor there subtracts; on every other link it is unknown. `core/segmentcompare` is pure. The walk in `core/graphalign` calls it, and the fan-out scan goes.

**Why now.** Slice 4's Check ("a spouse never corresponds to a sibling; parent ↔ child corresponds with ends swapped; differing births subtract; differing residences are unknown") is S11-14's scope entry. S11-16's layer 6 walk compares segments around accepted pairs, so it needs this layer first. S11-13 left edge support, credit, fan-out and signatures in place for this PR.

**What changes for the researcher** (each is the design, not a regression; tests are updated to it, D-11):
- A link's support no longer depends on how many such links the catalog holds. Subject of a birth or a death earns 1.5 plus 0.5 of the neighbor's own evidence. Every other link (parents, spouses, residences, witnesses, locations, place parts) earns 0.25 plus 0.25. On a real catalog, a parent or spouse link with fan-out near 1 used to earn 1.5 + 0.5.
- A person whose birth (or death) is already placed on a different event than the handle's birth loses 3 points. A birth whose subject is already placed on a different person than the handle's subject loses 3 points too (user, 2026-10-19). Before, both were silent.
- A differing neighbor on any other link stays silent, as today. `part_of` keeps its gates: support needs 2 points of own evidence, and a different parent cancels the part's own evidence.

**What stays.** The walk, the fallback pass, `refineOneHop`, `ConflictWithFixed`, the bands and the accept bar (S11-16). Which neighbors are compared stays the walk's choice: only neighbors already placed are compared, as today. The wire (`engine.proto`), Swift and copy. A bridge's own properties (dates on a place relationship) are not compared (D-10).

## 2. User stories

Conventions. `name(f)` is `propertycomparetest.Name(f)`, `term(k)` is `propertycomparetest.Term(k)`, `date(y)` is `propertycomparetest.Date(datevaluetest.YMD(y, 0, 0))`. Metas are `subjectcomparetest.Metas()` unless a row says otherwise. Rules are `matchregistry.Default()`. Scores are within 1e-9. A key is written as its `String()` (D-2): `participation·subject·person·birth`. "Placed together" and "placed apart" are `segmentcompare.Together` and `segmentcompare.Apart` (D-4).

**US-1.** As the engine, I read a bridge as one typed key, so links correspond only when they mean the same thing.

`segmentcompare.Key(link, terms)` with `terms` holding the seeded relationship terms (`parent` and `child` directed, inverse of each other; `spouse` and `sibling` symmetric).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Relationship `spouse` from either end; relationship `sibling` | Both spouse keys are `relationship·spouse`; the sibling key is `relationship·sibling`; they differ |
| AC-1.2 | "Mary parent of John" (`person` Mary, `related_to` John, term `parent`) from Mary's end; "John child of Mary" (`person` John, `related_to` Mary, term `child`) from Mary's end | Both are `relationship·parent·person`; from John's end both are `relationship·parent·related_to` |
| AC-1.3 | Participation, role `subject`, event type `birth`, from the person's end and from the event's end | `participation·subject·person·birth` and `participation·subject·event·birth` |
| AC-1.4 | Place relationship `part_of` from the part (`from`) and from the whole (`to`) | `place_relationship·part_of·from` and `place_relationship·part_of·to` |
| AC-1.5 | `segmentcompare.Compare` on two segments whose keys differ (spouse against sibling), placed together | `Corresponds` false, outcome unknown, cause `not_corresponding`, evidence 0 |

**US-2.** As the engine, I compare one segment with one research segment, and the cardinality table decides what a differing neighbor means.

`Compare(a, b, placement, own, metas, rules)` where `a` and `b` share a key, `own` is the pair's own evidence, and `rules` is `matchregistry.Default()` split into its sections.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Key `participation·subject·person·birth`; neighbors event {`term(birth)`, `date(1817)`} on both sides; together; own 2.5 | Outcome agree; evidence `OneSupport + OneCredit × (event_type.Agree + date.Agree)` = 1.5 + 0.5 × 2.5 = 2.75; `Neighbor.Evidence` 2.5 |
| AC-2.2 | AC-2.1, apart | Outcome conflict; cause `placed_apart`; evidence −`OneConflict` = −3 |
| AC-2.3 | Key `participation·subject·event·birth`; apart | Outcome conflict; evidence −3 |
| AC-2.4 | Key `participation·subject·person·residence`; apart | Outcome unknown; evidence 0 |
| AC-2.5 | AC-2.4, together, neighbors {`term(residence)`} both sides | Outcome agree; evidence `OtherSupport + OtherCredit × 0.5` = 0.25 + 0.125 = 0.375 |
| AC-2.6 | Key `participation·subject·person·birth`; together; neighbor events {`term(birth)`, `date(1817)`} against {`term(birth)`, `date(1830)`} | Outcome partial; evidence 1.5 + 0.5 × (0.5 − 3) = 0.25 |
| AC-2.7 | Key `place_relationship·part_of·from`; together; neighbors {toponym "York"} both sides; own 1.9 | Outcome unknown; cause `gated`; evidence 0 |
| AC-2.8 | AC-2.7 with own 3 | Outcome agree; evidence 0.25 + 0.25 × 3 = 1 |
| AC-2.9 | Key `place_relationship·part_of·from`; apart; own 3 | Outcome conflict; evidence −`Gates[…from].ConflictFactor` × 3 = −3 |
| AC-2.10 | Key `place_relationship·part_of·to`; apart; own 3 | Outcome unknown; evidence 0 |

**US-3.** As the engine, the walk scores links through segment compare, with no fan-out.

`graphalign` tests. A person's values are `{name: name("Ann Ames")}` with person metas; an event's are `{event_type: term(t), date: date(y)}`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Layer person P "Ann Ames", subject of birth E {birth, 1817}. Canon PER-1 "Ann Ames" subject of EVT-1 {birth, 1817}; PER-2 "Ann Ames" subject of EVT-2 {birth, 1817}. Fixed P → PER-1, E → EVT-2 | P's row score: 2.5 (name) − 3 = −0.5. E's row score: 2.5 (event) − 3 = −0.5 |
| AC-3.2 | AC-3.1 with both events `residence` instead of `birth` | P's row score 2.5; E's row score 2.5 |
| AC-3.3 | Layer Ann fixed to H-Ann; layer "Ann spouse Bob"; canon "H-Ann sibling PER-B", PER-B "Bob Ames" | Bob's row reason is not `via`; its score is the name weight 2.5 |
| AC-3.4 | `TestSparseNameAndBirthDate` | Unchanged: person and birth each strong at 5.25 |
| AC-3.5 | Every other `core/graphalign` test, ported (D-11) | Passes; a changed expected number is one of §1's listed changes |

**US-4.** As the engine (the Promote loader), I list a Source's links from the connect rules and key both sides the same way.

Catalog tests in `core/database/promotealign` with the existing `fixture`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A Source with one participation (person subject of a birth), one location (birth at a place), one relationship (person parent of person) and one place relationship (place part of place) | `loadLayerBridges` returns four bridges; their `KeyA.Kind` values, sorted, equal `connectrules.BridgeTypeKeys()` sorted |
| AC-4.2 | `TestProposeRelationshipFromEitherEnd` (Source A records "Mary parent of John", Source B records "John child of Mary") | Passes: John's row on Source B proposes John's handle through Mary |
| AC-4.3 | Source A: person "Ann Ames" subject of a dated birth 1817, both promoted. Source B: person "Ann Ames" subject of a dated birth 1817. Propose B with B's birth fixed to a third handle, a promoted birth 1817 whose subject is another person | B's person row: score 2.6 − 3 = −0.4, not placed on A's person |
| AC-4.4 | `go vet ./...`; `rg -n 'FanOut|loadStats|computeStats|EdgeSignature|SignatureScale|EdgePoints|EdgeFact' core api` | Builds; the grep prints nothing |

**US-5.** As a contributor, the segment section is one registry entry per rule and the skill says how to add one.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `TestKnobInventoryMatchesRegistry` | Passes with `segment.` removed from `notYetBuilt` |
| AC-5.2 | `TestSegmentSectionMatchesSeed` *(new)*, on `matchregistry.Default().Segment` | Every key in `One` and `Gates` names a registered bridge type, one of its endpoint properties as `End`, a seeded term of its disambiguation property as `Term`, and a seeded `event_type` term as `EventType` when set |
| AC-5.3 | `TestLayerPackages` | `core/segmentcompare` is listed with `numbers` true and passes |

## 3. Scope

| In | Covered by |
| --- | --- |
| `vocabkey.BridgeKey`, `BridgeTerm`, bridge type and endpoint constants (D-2) | STEP-1, US-1 |
| `core/segmentcompare`: `Link`, `Terms`, `Key`, `Segment`, `Placement`, `Compare`, `Result`, `Rules`, `Gate`, causes (D-1, D-3, D-4) | STEP-2, US-1, US-2 |
| Registry segment section, inventory rows, seed check (D-5) | STEP-3, US-5 |
| `graphalign` on segment compare; edge signatures, edge facts, signature scales, fan-out config and `Stats` deleted (D-6) | STEP-4, US-3 |
| Promote loader: link kinds from `connectrules`, keys on both sides, terms from the vocabulary; `stats.go` deleted (D-7) | STEP-5, US-4 |
| FFI `Via` fields from the key (D-8) | STEP-5 |
| Catalog tests for the slice Check (D-11) | STEP-6 |
| Skill and docs the change makes false (D-12) | STEP-7 |

| Out | Home |
| --- | --- |
| Comparing a bridge's own properties (place relationship dates) through subject compare (user, 2026-10-19) | S11-16 (scope entry updated with this plan) |
| Comparing neighbors that are not yet placed, by entity compare alone (a third `Placement`) | S11-16 |
| The layer 6 walk, top K, dirty-pair settling, `ConflictWithFixed` | S11-16 |
| `graphcache`'s stats slot (`Stats`, `SetStats`, `statsOK`, `ResetStatsForTest`) | S11-15 renames and reuses it |
| Moving the `"subject"` / `"birth"` / `"death"` literals in `core/hops/hops.go` and `core/database/conclusionheaders/compose.go` onto `vocabkey` constants | Not planned — not matching layers; each moves when next edited (S11-12 Out) |
| `connectrules`' own bridge type literals (`"participation"`, `"relationship"`, `"location"`, `"place_relationship"`) onto the new `vocabkey` constants | Not planned — the bridge registry stays the source of the keys; the constants name them for the match registry |
| Evidence lines for a corresponding segment (relationship bundles) | S11-19 |
| Plugin bridges in the match registry (a plugin's link is an "other" link with no entry) | Not planned in Spike 11 |

## 4. Current state

Written against S11-13's plan at `6ecfd88`. Rows marked *S11-13* describe code that plan adds or changes; check them against the landed code first (Implementer step 2).

| Area | Where | Today |
| --- | --- | --- |
| Edge signature | `core/graphalign/types.go` `EdgeSignature{BridgeType, RoleOrType, NeighborKind, NeighborTypeTerm, Directed}`, `Key()`, `ScaleKey(fromEnd)`; `Bridge{A, B, Signature}`; `CanonEdge{From, To, Signature}`; `Via{NeighborSubjectID, Signature}`; `Stats{FanOut}` | Untyped pipe-joined strings; a directed term carries `Directed` and ends through `fromEnd` |
| Walk links | `core/graphalign/align.go` `layerLink`, `canonLink` (`neighbor`, `sig`, `fromEnd`), `corresponds`, `newState` (builds `layerAdj` / `canonAdj`), `pushFromAnchor`, `supports`, `scoreFixed` (`fromNeighbor` view), `scoredCand.viaSig`, `enqueue(…, via EdgeSignature, …)`, `recordBest` | Correspondence is `sig.Key()` equality plus same end for a directed term |
| Edge scoring | `core/graphalign/align.go` `edgeFacts`, `nodeScore`, `scoreCandidate`; `core/graphalign/score.go` `EdgeFact`, `EdgePoints` | Support and credit by fan-out (`FanOutLowMax`); a mismatch subtracts only through `SignatureScale.ConflictFactor` |
| New-row duplicates | `core/graphalign/align.go` `alikeNew`, `newPairFacts`, `pairGroup`, `neighborsMatch`, `pairNode` | Groups neighbors by signature and end; fan-out per group |
| Walk config | `core/graphalign/registry.go` `Config{EdgeSupportLow, EdgeSupportHigh, NeighborCreditLow, NeighborCreditHigh, FanOutLowMax, FanOutUnknown, SignatureScale}`, `SignatureScale{MinNode, ConflictFactor}`; *S11-13*: `Config.Subject` | The `part_of` scale is keyed `place_relationship|part_of|place||from` / `…|to` |
| Entry | `core/graphalign/align.go` `Align(layer, canon, stats, fixed, cfg)`, `state.stats` | Takes fan-out stats |
| Scoring | *S11-13*: `core/graphalign/score.go` `ScoreCandidate(r subjectcompare.Result, edge, provenance float64) Scored`; `align.go` `evaluate(kind, a, b) subjectcompare.Result` | `edge` is `EdgePoints` |
| Walk tests | `core/graphalign/align_test.go` helpers `partSig`, `relSig`, `partOfSig`; tests that pass `graphalign.Stats{FanOut: …}` (`rg -n 'Stats\{' core/graphalign/align_test.go`); `TestHighFanOutNameCreditStaysUnderTheBar`, `TestConflictingNeighborLowersTheScore`, `TestEdgePointsHonorsTheSignatureScale`, `TestDefaultPartOfScale`, `TestPartOfDoesNotFileADifferentToponym`, `TestPartOfBreaksASharedToponym`, `TestDirectedRelationshipsMatchFromTheSameEnd`, `TestNewDuplicateCheckComparesDirectedNeighborsByEnd`, `TestNewEventsDisagreeWhenTheirParticipantsDo` | |
| Layer links | `core/database/promotealign/layer.go` `hopPersonToEvent`, `hopEventToPlace`, `hopPersonRelated`, `hopPlaceParent`, `hopPlaceSuccessor`, `loadLayerBridges` (four hand-written walks), `bridgeKey`, `neighborTypeProperty`; *S11-13*: `loadLayerBridges` also returns `[]participation`, and the profiler follows `hopPersonToEvent` | Link kinds listed in code |
| Canon links | `core/database/promotealign/canon.go` `canonSignature`, `keptTerm`, the edge loop in `loadCanon` (`ek` dedupe by `sig.Key()`) | Signs with `edgeStamp` |
| Terms | `core/database/promotealign/terms.go` `kinship`, `loadKinship`, `edgeStamp`, `kinship.canonical`, `observedTerms`, `termsBy` | Directed and inverse marks, keyed `propertyKey|termKey` |
| Fan-out | `core/database/promotealign/stats.go` `loadStats`, `cloneStats`, `computeStats`, `sqlFanOut`, `fanOutSum`, `addFanOut`; `export_test.go` `LoadStats`; `promotealign.go` `Propose` calls `loadStats` and passes `stats` to `Align`; tests `TestStatsCacheInvalidatesOnWrite`, `TestStatsFanOut` | One SQL scan per bridge type, cached on `graphcache` |
| Stats slot | `core/database/graphcache/graph.go` `Stats`, `SetStats`, `ResetStatsForTest`, fields `stats`, `statsOK` | Only `promotealign` uses it |
| Connect rules | `core/connectrules/connectrules.go` `Bridges()`, `Bridge{Origin, BridgeTypeKey, Endpoints, Disambiguation}`, `Endpoint{PropertyKey, TypeKey}`, `HasDisambiguation`, `LookupBridge`, `BridgeTypeKeys` | The one product table of bridges |
| Hops | `core/hops/hops.go` `NewHop(bridgeTypeKey, fromProperty, toProperty, filter)`; `core/database/canonicalgraph/canonicalgraph.go` `WalkSource` | `NewHop` checks a hop against `connectrules` |
| FFI | `api/ffi/handlers/promote_align.go` `ViaBridgeType: r.Via.Signature.BridgeType`, `ViaRole: r.Via.Signature.RoleOrType` | |
| Keys | `core/vocabkey/vocabkey.go` `SubjectType`, `PropertyKey`, `EventType`, `Role`, `EventBirth`, `EventDeath`, `RoleSubject`; `vocabkey_test.go`; `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` | No bridge key |
| Registry | `core/matchregistry/registry.go` `Registry{Value, Subject}`; `inventory_test.go` `notYetBuilt` (`segment.`, `alignment.`, `evidence.`); `guard_test.go` `layerPackages` | No segment section |
| Seed | `core/database/subjectvocab/registry.go` `seedTerms` (`role` terms, `relationship_type` with `Directed` and `Inverse`, `place_relationship_type`), `seedProperties`; `subject_section_test.go` `TestSubjectSectionMatchesSeed` (the shape to copy) | |
| Docs | `docs/matching-model.md` knob inventory rows `segment.*`, "Removed knobs"; `docs/promote-graph-alignment.md` (signature, fan-out and support sections: `rg -n 'fan-out|FanOut|signature|EdgeSupport|NeighborCredit' docs/promote-graph-alignment.md`); `docs/deployment-plan/spike-11/comparison-layers.md` §Segments; `.agents/skills/add-match-rule/SKILL.md` | |

**Precedent.** S11-10 ([#471](https://github.com/mendahu/provenencia/pull/471), [plan](S11-10-subject-compare.md)) built a layer package with its registry section, inventory rows, guard line, seed check and skill section: copy its shape for STEP-2 and STEP-3 (`TestSubjectSection` and `TestSubjectSectionMatchesSeed` are the models for D-5's tests). S11-13 ([plan](S11-13-promote-subject-compare.md)) moved the walk onto a layer and ported `align_test.go` in place, reading expected numbers from the config: copy that for STEP-4. S11-12's **Landed** note: a signature change that crosses packages is one commit (D-13).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Package.** `core/segmentcompare` *(new)*, package doc "Package segmentcompare is layer 5 of matching: whether a segment, two primary Subjects joined by a bridge, lines up with a research segment. Pure: no I/O, no catalog, no numbers; the segment section, layer 3's rules and the bridge terms come in as arguments." Files: `segmentcompare.go` (`Rules`, `Gate`, `Segment`, `Placement`, `Result`, `Cause`, `Compare`), `key.go` (`Link`, `Terms`, `Term`, `Key`). External tests in `segmentcompare_test` | Layer 5's package per the working rule "Packages" | Folding it into `graphalign` — layer 6 must compose layer 5, not contain it |
| D-2 | **Typed keys in `vocabkey`.** `type BridgeTerm string` ("a bridge's disambiguation term key: a participation role, a relationship type or a place relationship type, typed for registry rules"). `type BridgeKey struct { Kind SubjectType; Term BridgeTerm; End PropertyKey; EventType EventType }` with doc "BridgeKey is one bridge read from one end: its type, its term in canonical direction, the endpoint Property of the end it is read from (empty for a symmetric link between two of one kind), and the type of the event on it (participation only)". `func (k BridgeKey) String() string`: the non-empty fields joined with `·` in field order. Constants, untyped like `Person`: `Participation = "participation"`, `Relationship = "relationship"`, `Location = "location"`, `PlaceRelationship = "place_relationship"`; seeded endpoint Property keys `PropertyPerson = "person"`, `PropertyEvent = "event"`, `PropertyFrom = "from"`, `PropertyTo = "to"`; term `TermPartOf BridgeTerm = "part_of"`. The role term is written `BridgeTerm(RoleSubject)`. `TestVocabKeysMatchSeed` checks the four bridge type constants against `seedTypes`, the four endpoint constants against `seedProperties`, and `TermPartOf` under `place_relationship_type` | The registry rule "Typed keys, not pipe-joined strings"; S11-12 named `BridgeKey` as the next user of its typed term keys and asked for "a role-or-term key built beside `vocabkey.Role`"; the skill: "no other package declares a key type". `String()` makes the inventory path readable (`TestKnobInventoryMatchesRegistry` prints map keys with `fmt.Sprint`) | `End` as a bool (`fromEnd`) — it cannot say which end of a participation; `NeighborKind` in the key — the endpoint Property already implies it; keeping `EdgeSignature` — untyped, and its `Key()` drops the end |
| D-3 | **Key rule.** `segmentcompare.Link{Kind vocabkey.SubjectType; Term vocabkey.BridgeTerm; End, Other vocabkey.PropertyKey; EventType vocabkey.EventType; OneKind bool}`: a bridge as recorded, read from the end whose endpoint Property is `End`; `OneKind` is true when both endpoints hold the same Subject type. `Terms map[vocabkey.SubjectType]map[vocabkey.BridgeTerm]Term`, `Term{Directed bool; Inverse vocabkey.BridgeTerm}`: the catalog's term facts per bridge type. `Key(l Link, terms Terms) vocabkey.BridgeKey`: start from `l.Term`, `l.End`, `l.Other`; when `terms[l.Kind][term].Inverse` is non-empty and sorts before `term`, take the inverse and swap `End` and `Other`; then when `l.OneKind` and the resulting term is not `Directed`, `End` is empty. `EventType` is copied | "A term and its inverse correspond with the ends swapped" is layer 5's rule ([segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6)); directed and inverse marks are data-model facts passed in ([match registry](../comparison-layers.md#match-registry)). Same canonical choice as today's `kinship.canonical` (the key that sorts first), so wire values do not move | Normalizing in the loader (today's `edgeStamp`) — layer 5's rule outside layer 5, untested at its level |
| D-4 | **Compare.** `Segment{Key vocabkey.BridgeKey; NeighborKind vocabkey.SubjectType; Neighbor subjectcompare.Values}`. `type Placement string`: `Together = "together"` (the layer neighbor is placed on this research neighbor), `Apart = "apart"` (placed on another handle). `Rules{One []vocabkey.BridgeKey; OneSupport, OneCredit, OneConflict, OtherSupport, OtherCredit float64; Gates map[vocabkey.BridgeKey]Gate}`, `Gate{MinOwnEvidence, ConflictFactor float64}`. `Result{Corresponds bool; Outcome propertycompare.Outcome; Cause Cause; Evidence float64; Neighbor subjectcompare.Result}`. Causes: `CauseNotCorresponding = "not_corresponding"`, `CauseGated = "gated"`, `CausePlacedApart = "placed_apart"`. `Compare(a, b Segment, placed Placement, own float64, metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) Result`, in order: (1) `a.Key != b.Key` → unknown, `not_corresponding`, `Corresponds` false. (2) `g := rules.Gates[a.Key]`; `one := slices.Contains(rules.One, a.Key)`. (3) Apart: evidence `−(OneConflict if one) − g.ConflictFactor × own`; outcome conflict when `one` or `g.ConflictFactor > 0`, else unknown; cause `placed_apart`. (4) Together and `g.MinOwnEvidence > 0` and `own < g.MinOwnEvidence` → unknown, `gated`, 0. (5) Together: `Neighbor = subjectcompare.Compare(a.NeighborKind, a.Neighbor, b.Neighbor, metas, subject, value)`; evidence `support + credit × Neighbor.Evidence` with the One pair when `one`, else the Other pair; outcome partial when any `Neighbor.Properties` entry's outcome is conflict, else agree | The cardinality table in [segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6), with today's credit ("segment evidence includes how well the neighbor itself matched") and today's `part_of` gates carried over. Placement comes from the walk because only layer 6 knows where a neighbor is placed; `Result.Neighbor` carries layer 3's speaking values for S11-19's bundle ("output carries value refs") | Entity compare alone deciding "differs" in this PR — the walk today compares placed neighbors only, and changing that is S11-16's walk; a `Similarity` field — `Neighbor.Similarity` holds it, and nothing reads a second copy |
| D-5 | **Registry section (user, 2026-10-19: both ends).** `matchregistry.Registry` gains `Segment segmentcompare.Rules` ("Segment is layer 5: what a matching or differing link counts."), set by `defaultSegment()` in `core/matchregistry/segment.go` *(new)*: `One` = the four keys `participation·subject·person·birth`, `participation·subject·person·death`, `participation·subject·event·birth`, `participation·subject·event·death` (comment: decided: a person has one birth and one death, and each has one subject); `OneSupport` 1.5 and `OneCredit` 0.5 (`graphalign` `EdgeSupportLow`, `NeighborCreditLow`); `OtherSupport` 0.25 and `OtherCredit` 0.25 (`EdgeSupportHigh`, `NeighborCreditHigh`); `OneConflict` 3 (estimate: a differing birth or death is as strong as a differing name or date, `subject.Weights.person.name.Conflict`); `Gates[place_relationship·part_of·from]` = {2, 1} and `Gates[place_relationship·part_of·to]` = {2, 0} (`SignatureScale` `part_of`). Inventory rows replace `segment.Cardinality.*` and `segment.PartOf.*` (STEP-3) | Numbers are the inventory's starting values; `OneConflict` is the one new knob, estimated by the working rule "Registry numbers: carry over, then estimate". The user chose both ends | A separate cardinality enum per key — `One` is the only declared value; `ConflictFactor × own` for births — a sparse person's own evidence is small, so a differing birth would barely count |
| D-6 | **Walk on segments.** `graphalign.Bridge{A, B []byte; KeyA, KeyB vocabkey.BridgeKey}` (KeyA is the bridge read from A's end); `CanonEdge{From, To []byte; KeyFrom, KeyTo vocabkey.BridgeKey}`; `Via{NeighborSubjectID []byte; Key vocabkey.BridgeKey}`. `Config` loses `EdgeSupportLow`, `EdgeSupportHigh`, `NeighborCreditLow`, `NeighborCreditHigh`, `FanOutLowMax`, `FanOutUnknown`, `SignatureScale` and gains `Segment segmentcompare.Rules` (after `Subject`, doc "Segment is the registry's segment section: what a link counts. DefaultConfig holds the registry's."). `Align(layer, canon, fixed, cfg)` loses `stats`; `Stats`, `state.stats`, `EdgeSignature`, `SignatureScale`, `EdgeFact`, `EdgePoints`, `corresponds`, `pairGroup`, `nodeScore` and `pairNode` are deleted. `layerLink` and `canonLink` become `{neighbor []byte; key, back vocabkey.BridgeKey}` (`back` is the key from the neighbor's end); correspondence is `link.key == ce.key`; `scoreFixed`'s neighbor view compares `link.back`. `edgeFacts` becomes `segments(subjectID []byte, h *Handle, own float64) float64`: for each placed layer neighbor, the first canon link with the same key landing on its handle is compared together; failing that, the first landing elsewhere is compared apart; no such link adds nothing; it returns the sum of `Evidence`. `scoreCandidate` computes the Subject's `subjectcompare.Result` first and passes its `Evidence` as `own`. `newPairFacts` becomes `newPairSegments(a, b *Subject, own float64) (float64, bool)`: neighbors grouped by `link.key`; a key both have disagrees when no neighbor pair is the same Subject or passes `neighborsMatch`; otherwise it adds the best `Compare` evidence over matching pairs, together. `scoredCand.viaSig` becomes `viaKey`. `DefaultConfig` sets `Segment: reg.Segment` | Layer 6 composes layer 5 and decides placement (rule 7); one key comparison replaces signature equality plus `Directed` plus `fromEnd` | Passing `Terms` into `graphalign` to key links there — the loader has the vocabulary, and keys travel ready-made |
| D-7 | **Loader.** `loadLayerBridges` walks each `connectrules.Bridges()` entry with two endpoints whose first endpoint's type has Subjects on the layer: `hops.NewHop(b.BridgeTypeKey, ep[0].PropertyKey, ep[1].PropertyKey, nil)` (an error is returned), `canonicalgraph.WalkSource`, the term through `observedTerms(q, assocs, b.Disambiguation)` when `connectrules.HasDisambiguation`, and the event type through `observedTerms(q, events, neighborTypeProperty[vocabkey.Event])` for the endpoint whose type is `event`. Each edge becomes a `Bridge` whose `KeyA` is `segmentcompare.Key(linkOf(b, term, true, eventType), terms)` and `KeyB` the same with `false`. `linkOf(b connectrules.Bridge, term, eventType string, firstEnd bool) segmentcompare.Link` *(new, `terms.go`)* sets `End` / `Other` from `b.Endpoints` and `OneKind` from their `TypeKey`s. `loadTerms(g) (segmentcompare.Terms, error)` *(new, `terms.go`)* reads `g.Vocabulary()` terms whose `PropertyKey` is a bridge's disambiguation, under that bridge's type. `hopEventToPlace`, `hopPersonRelated`, `hopPlaceParent`, `hopPlaceSuccessor`, `kinship`, `loadKinship`, `edgeStamp`, `canonical` are deleted; `hopPersonToEvent` stays (the profiler reads it). `canonSignature` becomes `canonKeys(selfKind string, n *graphcache.Node, l graphcache.Link, neighbor *graphcache.Node, terms segmentcompare.Terms) (from, to []byte, keyFrom, keyTo vocabkey.BridgeKey)` with the event type from `keptTerm` as today. `bridgeKey` dedupes by sorted ids plus both keys' `String()`; `loadCanon`'s `ek` uses `KeyFrom.String()`. `stats.go`, `export_test.go`'s `LoadStats`, `TestStatsCacheInvalidatesOnWrite` and `TestStatsFanOut` are deleted; `Propose` drops `loadStats` | The map: "The layer loader reads link kinds from `connectrules` instead of a list in code"; a plugin's registered bridge reaches the walk with no edit here. Both sides key through one function | Keeping the four named walks — the list in code the map retires; deleting `graphcache`'s slot — S11-15 renames and reuses it |
| D-8 | **Wire.** The FFI handler sends `ViaBridgeType: string(r.Via.Key.Kind)`, `ViaRole: string(r.Via.Key.Term)`. No proto change | Same values as `BridgeType` and `RoleOrType` today (the canonical term), so Swift is untouched ("One wire vocabulary") | A new `via_end` field — nothing on the page reads it |
| D-9 | **No reordering of the walk.** Which pairs are scored, when `refineOneHop` runs, the fallback, `ConflictWithFixed` and the bands stay as S11-13 leaves them. Only the points a link adds change | S11-16 rewrites the walk; this PR replaces the link scorer under it | Comparing unplaced neighbors now — S11-16's step 4 |
| D-10 | **Bridge properties are not compared (user, 2026-10-19).** `Segment` holds no bridge values, and neither loader reads them. The key carries the role, relationship type and event type. The map's scope entry for S11-14 loses "bridge by subject compare on its properties"; S11-16's gains "a bridge's own properties (place relationship dates), through subject compare, if a golden needs them" | No inventory weight exists for any bridge property, so the comparison would score nothing, and both loaders would read rows for it | Plumbing bridge values with no weights — loader work and a code path no test can show working |
| D-11 | **Ported tests.** In `align_test.go`, `partSig(role, neighborKind, typeTerm)` becomes `partKeys(role vocabkey.BridgeTerm, eventType vocabkey.EventType) (person, event vocabkey.BridgeKey)`; `relSig(typ)` becomes `relKeys(term vocabkey.BridgeTerm, directed bool) (a, b vocabkey.BridgeKey)` (symmetric: both `relationship·term`; directed: `·person` and `·related_to`); `partOfSig()` becomes `partOfKeys() (part, whole vocabkey.BridgeKey)`. Every `graphalign.Stats{…}` argument goes. A test reading `cfg.EdgeSupportLow` / `NeighborCreditLow` reads `cfg.Segment.OneSupport` / `OneCredit` when its key is in `One`, else `OtherSupport` / `OtherCredit`; `SignatureScale` reads `cfg.Segment.Gates`. `TestHighFanOutNameCreditStaysUnderTheBar` is renamed `TestOtherLinkCreditStaysUnderTheBar` (residence: the Other pair). `TestEdgePointsHonorsTheSignatureScale` and `TestDefaultPartOfScale` are deleted; their cases are T-2's rows AC-2.7 to AC-2.10. **Expected changes, and only these:** (a) a link's support and credit come from `One` / Other, not fan-out (a catalog test with fan-out near 1 on a non-birth link drops to the Other pair); (b) a birth or death link placed apart subtracts 3 from both ends; (c) a test whose name states the behavior (`TestSparseNameAndBirthDate`, `TestGracieObituaryGolden`, `TestPartOfChainIsStrongWithoutADecision`, `TestPartOfChainIsAStrongMatch`, `TestPartOfDoesNotFileADifferentToponym`, `TestPartOfBreaksASharedToponym`, `TestDirectedRelationshipsMatchFromTheSameEnd`, `TestProposeRelationshipFromEitherEnd`, `TestProposePartOfMatchesTheSameEnd`) keeps every Target and handle it asserts. Any other change is a stop condition | The S11-13 precedent; the goldens are the acceptance tests ("Sparse data first") | Rewriting the suite on new fixtures — numbers would move for no behavior |
| D-12 | **Skill and docs.** `add-match-rule`: the description and the intro list `core/segmentcompare`; a rule bullet "Layer 5 (`core/segmentcompare`) scores one link" (keys come from `Key` with the catalog's `Terms`; correspondence is key equality; `One` and `Gates` keyed by `vocabkey.BridgeKey`; placement comes from layer 6; neighbor evidence through layer 3); a section "Add a link rule" (a `One` key or a `Gates` entry in `core/matchregistry/segment.go` (new) with a source comment, its inventory row named by the key's `String()`, `TestSegmentSectionMatchesSeed` and a `TestCompare` row); the Check command gains `./core/segmentcompare/...`. `.cursor/rules/match-rules.mdc` globs gain `core/segmentcompare/**`. `matching-model.md`: the segment rows (STEP-3), "Removed knobs" gains `graphalign` `SignatureScale` (→ `segment.Gates`) and the fan-out scan (`Stats.FanOut`, `sqlFanOut`), worked example step 5 names `segment.One`. `comparison-layers.md` §Segments: a "Built in S11-14" paragraph. `promote-graph-alignment.md`: its signature, fan-out and support text describes keys and the segment section | Working rule "One matching skill, grown with the layers"; each migration PR fixes the docs it makes false | — |
| D-13 | **Commits.** STEP-4 and STEP-5 are one commit: `graphalign`'s types change breaks `promotealign` and the FFI handler until both move. Every other step is its own commit | S11-12's Landed note | Two commits that do not build |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/segmentcompare` *(new)*: `Compare`, `Result`, `Rules`, `Gate`, `Segment`, `Placement`, `Cause` | package | `core/graphalign/score.go` `EdgeFact`, `EdgePoints`; `registry.go` `SignatureScale` | Layer 5 is its own pure layer; these move into it and are deleted (D-1, D-4) |
| `segmentcompare.Key`, `Link`, `Terms`, `Term` *(new)* | func, type | `promotealign` `edgeStamp`, `kinship.canonical`; `graphalign.corresponds` | The same rule, moved to the layer that owns it (D-3) |
| `vocabkey.BridgeKey`, `BridgeTerm`, `BridgeKey.String` *(new)*; constants `Participation`, `Relationship`, `Location`, `PlaceRelationship`, `PropertyPerson`, `PropertyEvent`, `PropertyFrom`, `PropertyTo`, `TermPartOf` *(new)* | type, const | `graphalign.EdgeSignature`; `vocabkey.Role`, `EventType` | The key home (D-2); `EdgeSignature` is untyped and deleted |
| `matchregistry.Registry.Segment`, `core/matchregistry/segment.go` *(new)* `defaultSegment` | field, file | `subject.go` `defaultSubject` | Same shape, one file per section ([match registry](../comparison-layers.md#match-registry)) |
| `graphalign.Config.Segment` *(new field)*; `Bridge.KeyA/KeyB`, `CanonEdge.KeyFrom/KeyTo`, `Via.Key` *(new fields)* | field | `Config.SignatureScale`, `EdgeSupport*`; `Signature` fields | Replace them (D-6) |
| `state.segments`, `state.newPairSegments` *(new)* | method | `edgeFacts`, `newPairFacts` | Renamed rewrites on the layer (D-6) |
| `promotealign` `loadTerms`, `linkOf`, `canonKeys` *(new)* | func | `loadKinship`, `edgeStamp`, `canonSignature` | Same jobs on the typed key (D-7) |
| `TestKey`, `TestCompare`, `TestInvariants` in `core/segmentcompare/segmentcompare_test.go` *(new)*; `TestSegmentSection` in `core/matchregistry/segment_test.go` *(new)*; `TestSegmentSectionMatchesSeed` in `core/database/subjectvocab/segment_section_test.go` *(new)*; `TestBirthPlacedApartSubtracts`, `TestResidencePlacedApartIsUnknown`, `TestSpouseNeverCorrespondsToSibling`, `TestOtherLinkCreditStaysUnderTheBar` (renamed) in `align_test.go`; `TestLayerBridgesCoverEveryConnectRule`, `TestProposeBirthOnAnotherPersonSubtracts` in `promotealign_test.go`; `LayerBridges` in `export_test.go` | test, helper | `TestSubjectSection`, `TestSubjectSectionMatchesSeed`, `subjectcompare` `TestCompare` / invariants | Same table shapes |

**Builds on.** `subjectcompare.Compare` and the subject section (S11-10), the handle profiles S11-13 builds, `connectrules` and `hops.NewHop`, the `add-match-rule` skill, the guard and inventory tests.
**Deletes.** `graphalign` `EdgeSignature`, `Stats`, `SignatureScale`, `EdgeFact`, `EdgePoints`, `corresponds`, `pairGroup`, `nodeScore`, `pairNode`, the seven edge `Config` fields; `promotealign` `stats.go`, `LoadStats`, `kinship`, `loadKinship`, `edgeStamp`, `canonical`, four of five hop vars, `canonSignature`; tests `TestEdgePointsHonorsTheSignatureScale`, `TestDefaultPartOfScale`, `TestStatsCacheInvalidatesOnWrite`, `TestStatsFanOut`; inventory rows `segment.Cardinality.*`, `segment.PartOf.*`; `segment.` from `notYetBuilt`.
**Next adds.** S11-16 adds a third `Placement` for unplaced neighbors (entity compare decides) and calls `Compare` around accepted pairs; bridge-property comparison is a `Segment.Bridge` field and subject section weights for a bridge type, if a golden needs it (D-10). A new axiomatic one-cardinality link is one `One` entry; a new gate one `Gates` entry. S11-19 bundles a corresponding segment's evidence from `Result.Neighbor`. S11-15's shortlist index takes `graphcache`'s stats slot.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `BridgeKey`, `BridgeTerm`, `String`, constants (D-2) |
| `core/segmentcompare` *(new)* | `segmentcompare.go` (new), `key.go` (new), tests (D-1, D-3, D-4) |
| `core/matchregistry` | `registry.go` `Segment` field; `segment.go` *(new)*; `segment_test.go` *(new)*; `inventory_test.go` `notYetBuilt`; `guard_test.go` `layerPackages` (D-5) |
| `core/graphalign` | `types.go`, `registry.go`, `score.go`, `align.go`; tests (D-6, D-11) |
| `core/database/promotealign` | `layer.go`, `canon.go`, `terms.go`, `promotealign.go`; `stats.go` deleted; tests (D-7) |
| `core/database/subjectvocab` | `segment_section_test.go` *(new)*; `TestVocabKeysMatchSeed` rows (D-2, D-5) |
| `api/ffi/handlers` | `promote_align.go` via fields (D-8) |

```go
// vocabkey

type BridgeTerm string

type BridgeKey struct {
	Kind      SubjectType // the bridge type
	Term      BridgeTerm  // canonical direction; empty when the bridge has no term
	End       PropertyKey // the endpoint read from; empty for a symmetric link between two of one kind
	EventType EventType   // the event on a participation; empty otherwise
}

func (k BridgeKey) String() string

// segmentcompare

type Link struct {
	Kind       vocabkey.SubjectType
	Term       vocabkey.BridgeTerm
	End, Other vocabkey.PropertyKey
	EventType  vocabkey.EventType
	OneKind    bool
}
type Term struct {
	Directed bool
	Inverse  vocabkey.BridgeTerm
}
type Terms map[vocabkey.SubjectType]map[vocabkey.BridgeTerm]Term

func Key(l Link, terms Terms) vocabkey.BridgeKey

type Segment struct {
	Key          vocabkey.BridgeKey
	NeighborKind vocabkey.SubjectType
	Neighbor     subjectcompare.Values
}
type Placement string
type Cause string
type Gate struct{ MinOwnEvidence, ConflictFactor float64 }
type Rules struct {
	One                                  []vocabkey.BridgeKey
	OneSupport, OneCredit, OneConflict   float64
	OtherSupport, OtherCredit            float64
	Gates                                map[vocabkey.BridgeKey]Gate
}
type Result struct {
	Corresponds bool
	Outcome     propertycompare.Outcome
	Cause       Cause
	Evidence    float64
	Neighbor    subjectcompare.Result
}

func Compare(a, b Segment, placed Placement, own float64, metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) Result

// graphalign

func Align(layer Layer, canon Canon, fixed []Fixed, cfg *Config) Proposal
```

`Rules` fields carry the doc comments the inventory's "Changes" column states. `Gates` doc: "per key, a minimum own evidence before support counts and a factor of own evidence lost when placed apart; a missing key is the zero gate".

### FFI

None — `api/proto/engine.proto` is unchanged; the handler fills the same fields from the key (D-8).

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Typed bridge keys**
Files: `core/vocabkey/vocabkey.go`, `core/vocabkey/vocabkey_test.go`, `core/database/subjectvocab/vocabkey_test.go`. Skill: `add-match-rule` ("Add a Subject property or pair", step 1).
Do: D-2. `vocabkey_test.go` gains `TestBridgeKeyString` *(new)* (T-1).
Check: `go test ./core/vocabkey/...` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` pass.

**STEP-2 — Segment compare**
Files: `core/segmentcompare/segmentcompare.go`, `key.go`, `segmentcompare_test.go` (new); `core/matchregistry/guard_test.go`. Skill: `add-match-rule` ("Add a layer package").
Do: D-1, D-3, D-4. Add `{"core/segmentcompare", true}` to `layerPackages` after `core/subjectprofile`. Tests T-2 to T-4 build their rules in the test (a literal `Rules` in the test file, the registry's numbers come in STEP-3).
Check: `go test ./core/segmentcompare/... ./core/matchregistry/...` passes.

**STEP-3 — Registry segment section**
Files: `core/matchregistry/registry.go`, `segment.go` (new), `segment_test.go` (new), `inventory_test.go`; `core/database/subjectvocab/segment_section_test.go` (new); `docs/matching-model.md`. Skill: `add-match-rule` ("Add a knob").
Do: D-5. Delete `"segment."` from `notYetBuilt`. In the knob inventory, replace `segment.Cardinality.birth`, `segment.Cardinality.death` and the three `segment.PartOf.*` rows with: `segment.One` (the four keys; "decided 2026-10-17, both ends 2026-10-19"; "Links where a differing neighbor conflicts"), `segment.OneConflict` (3; the D-5 estimate; "Points lost when a one-cardinality link's neighbor is placed elsewhere"), `segment.Gates.place_relationship·part_of·from.MinOwnEvidence` (2), `.ConflictFactor` (1), `segment.Gates.place_relationship·part_of·to.MinOwnEvidence` (2), `.ConflictFactor` (0; "a different whole is silent"), each with today's source column. Switch STEP-2's test rules to `matchregistry.Default().Segment` where a row reads a registry number (external test packages may import the registry).
Check: `go test ./core/matchregistry/... ./core/segmentcompare/...` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` pass.

**STEP-4 — Walk on segments**
Files: `core/graphalign/types.go`, `registry.go`, `score.go`, `align.go`, `align_test.go`, `pick_internal_test.go` if it names a removed field. Skill: `add-match-rule`.
Do: D-6, D-9, D-11. Add T-5 to T-7.
Check: `go test ./core/graphalign/...` passes (the tree does not build outside `core/graphalign` until STEP-5; D-13).

**STEP-5 — Loader and wire**
Files: `core/database/promotealign/layer.go`, `canon.go`, `terms.go`, `promotealign.go`, `export_test.go`, `promotealign_test.go`, `bench_test.go` if it calls `loadStats`; delete `stats.go`; `api/ffi/handlers/promote_align.go`.
Do: D-7, D-8. Before deleting `stats.go`, run `rg -n 'loadStats|LoadStats|computeStats|graphalign.Stats' core api`; a hit outside the files above is a stop condition.
Check: `go build ./... && CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./api/ffi/...` passes; AC-4.4's grep prints nothing.

**STEP-6 — Slice Check in the catalog**
Files: `core/database/promotealign/promotealign_test.go`, `export_test.go`.
Do: T-8 and T-9. `export_test.go` gains `LayerBridges(c *database.Catalog, sourceID []byte) ([]graphalign.Bridge, error)`, which loads the layer the way `Propose` does and returns its bridges.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes.

**STEP-7 — Skill and docs**
Files: `.agents/skills/add-match-rule/SKILL.md`, `.cursor/rules/match-rules.mdc`, `docs/matching-model.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `completed.md`, `deployment-plan.md`, this plan.
Do: D-12; the Definition of done's docs lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-14-segment-compare.md` reports no errors; `go test ./core/matchregistry/...` passes (the inventory test reads `matching-model.md`); `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior (STEP-4 and STEP-5 together, D-13). The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/vocabkey/vocabkey_test.go` `TestBridgeKeyString` *(new)* | `{participation, subject, person, birth}` → `participation·subject·person·birth`; `{relationship, spouse, "", ""}` → `relationship·spouse` | D-2 |
| T-2 | Go | `core/segmentcompare/segmentcompare_test.go` `TestKey` *(new)* | AC-1.1 to AC-1.4, one row each | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-3 |
| T-3 | Go | same file, `TestCompare` *(new)*, rows named by AC | AC-1.5, AC-2.1 to AC-2.10, one row each | AC-1.5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, D-4 |
| T-4 | Go | same file, `TestInvariants` *(new)*, fixed seed, 500 random segment pairs from the fixtures | Non-corresponding keys → evidence 0; together and ungated, evidence equals the key's support + credit × `Neighbor.Evidence`; apart on a key in neither `One` nor `Gates` → 0 | D-4 |
| T-5 | Go | `core/graphalign/align_test.go` `TestBirthPlacedApartSubtracts` *(new)* | AC-3.1 | AC-3.1, D-5 |
| T-6 | Go | same file, `TestResidencePlacedApartIsUnknown` *(new)* | AC-3.2 | AC-3.2 |
| T-7 | Go | same file, `TestSpouseNeverCorrespondsToSibling` *(new)* | AC-3.3 | AC-3.3 |
| T-8 | Go | `core/database/promotealign/promotealign_test.go` `TestLayerBridgesCoverEveryConnectRule` *(new)* | AC-4.1 | AC-4.1, D-7 |
| T-9 | Go | same file, `TestProposeBirthOnAnotherPersonSubtracts` *(new)* | AC-4.3 | AC-4.3 |
| T-10 | Go | `core/matchregistry/segment_test.go` `TestSegmentSection` *(new)* | `Default().Segment`: every `One` key is a participation key with `Term` `subject`; every `Gates` key's `Kind` is a `connectrules` bridge type; all support and credit numbers are positive | AC-5.1, D-5 |
| T-11 | Go | `core/database/subjectvocab/segment_section_test.go` `TestSegmentSectionMatchesSeed` *(new)* | AC-5.2 | AC-5.2 |
| T-12 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages`; `inventory_test.go` `TestKnobInventoryMatchesRegistry` | Pass with the new row and section | AC-5.1, AC-5.3 |
| T-13 | Go | Ported: `TestSparseNameAndBirthDate`, `TestProposeRelationshipFromEitherEnd`, the rest of `core/graphalign` and `core/database/promotealign` | Pass under D-11's expected changes | AC-3.4, AC-3.5, AC-4.2, D-11 |
| T-14 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposePromoteGraphAlignment` | Passes; a via row's `via_bridge_type` and `via_role` are unchanged | D-8 |

AC-4.4 is the STEP-5 build and grep. Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestGracieObituaryGolden`, the `part_of` tests in both packages, `BenchmarkPropose*` in `bench_test.go` (it may call `loadStats`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: matching-model.md, promote-graph-alignment.md, comparison-layers.md §Segments (D-12)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

`completed.md` gets, above the newest entry: "## S11-14 — Segment compare" and "**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). Promote judges each link with segment compare. A link is a typed key, so a spouse never stands for a sibling and a parent link reads the same from either record. Subject of a birth or a death is one per person and one per event: a birth placed on a different event, or on a different person, subtracts 3. Every other link earns the same small support whatever the catalog holds; the fan-out scan is gone."

## 10. Manual check

1. Open a project with two Sources about one family: Source A records "Mary Smith parent of John Smith", both promoted. Source B records "John Smith child of Mary Smith". Open Promote on B and decide Mary onto Mary's handle. John's row proposes John's handle, read "through Mary".
2. On Source B, add a spouse link from Mary to a new person "Robert Smith" while A records Robert as Mary's sibling. Robert's row is not proposed through Mary.
3. Source C records Ann Ames, subject of a birth on 1 May 1817. The research graph holds two Ann Ames, each the subject of a birth on that date. Decide C's birth onto the second Ann's birth. Ann's row on C does not propose the first Ann.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Spouse and parent links drop from 1.5 + 0.5 to 0.25 + 0.25 on real catalogs, so a family placed by its links files weaker | The inventory's starting values (S11-01); D-11 (c) holds the goldens' Targets; dogfood notes go to the knob inventory, not into this PR |
| Both-end birth conflicts push a decided row negative where the old walk was silent | AC-3.1 pins the number; `ConflictWithFixed` is unchanged (S11-16 replaces it with the contested mark) |
| A plugin bridge walked generically trips `hops.NewHop` | `NewHop` returns its error and `Propose` fails loudly; T-8 covers every product bridge |

**Stop and report if:**

- S11-13 landed with a different `graphalign` scoring shape than §4's *S11-13* rows (`ScoreCandidate`, `evaluate`, `Config.Subject`, the profiler's `hopPersonToEvent`).
- A test named in D-11 (c) changes a Target or handle it asserts.
- A ported number changes for a reason not in D-11 (a) or (b).
- `connectrules.Bridges()` returns a bridge with other than two endpoints that the layer has Subjects for.
- The FFI test's `via_role` changes for any row (D-3's canonical choice differs from `kinship.canonical`).
- STEP-5's `rg` finds a stats caller outside the files it names.

## 12. Open questions

None.

## 13. PR

**Title:** S11-14 — Segment compare
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-14-segment-compare.md`.
