# S11-40 — Implied links in claim evidence

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the sheet's relationship Card keeps today's sentence; S11-42 words the new fields in the same Card)
**Board:** —
**Depends on:** S11-39 (landed [#572](https://github.com/mendahu/provenencia/pull/572)), S11-51 (landed [#575](https://github.com/mendahu/provenencia/pull/575))
**Blocks:** S11-42
**Branch:** `s11-40-implied-links-evidence`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [Kinship shapes § On the evidence sheet](../kinship-shapes.md#on-the-evidence-sheet), [§ Counting once](../kinship-shapes.md#counting-once), [§ Where implied links live](../kinship-shapes.md#where-implied-links-live); [`claim-evidence.md` § Evidence through relationships](../claim-evidence.md#evidence-through-relationships); S11-51 plan D-4, D-5 and its **Next adds**; S11-39's Landed note; working rules: "The canon's implied links live in the graph cache", "One fact, one link per side", "Counted once across sides", "One wire vocabulary", "The pin policy has one owner, in Go"
**Planned:** 2026-10-26 against `main` @ `f22b19a`
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

The evidence sheet shows a relationship line when the tree records a kinship fact in a different shape from the Source. Today only the Source side reads the kinship patterns: a Source marriage meets a filed spouse relationship, but a tree marriage never meets a Source spouse relationship, and nothing on the line says how either side recorded the fact. After this PR the evidence query's canon side reads a person's recorded and implied links through `canonlinks.Reader.Relationships`, each relationship line says how each side recorded the fact (directly, or through which pattern and pivot, with the pivot's date or name), its bundle pins every bridge behind an implied link, and the line follows the count-once rule. The RPC carries it to Swift. The sheet's words do not change until S11-42.

**Why now.** S11-39 made matching read both shapes and S11-51 gave one read of a person's merged links. The evidence sheet is the last reader that still sees only filed links on the canon side, so a row Promote proposes because of a tree marriage opens a sheet with no relationship line for it (S11-39's `TestEvidenceReadsASourceMarriageAsSpouses` AC-3.3 pins that gap).

**What stays.** Every property line, likeness line and section order. Every relationship line between filed links and recorded layer links keeps its key, outcome, checked state, incoming Observations, bundles and pin ids. Matching, bands and proposals. The sheet's sentences and chrome (S11-42). `canonlinks` and `core/kinship` are read, not changed.

## 2. User stories

Fixture names: "tree" is the first Source, every Subject on it promoted (`f.promote`) unless the AC says otherwise; "record" is a second Source, unpromoted. The row is the record's Bob on the tree Bob's handle, and the page places the record's Ann on the tree Ann's handle, unless the AC names other rows.

**US-1.** As a researcher, when the tree records a kinship fact through a marriage, a birth or a parent chain and the record records it directly (or the other way round), the sheet shows one relationship line for it whose pins hold every bridge behind it, so that I can pin the evidence matching already counted.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Tree: marriage M with Ann and Bob each `subject`. Record: Ann and Bob in a `spouse` relationship | One relationship: key `relationship·spouse`, neighbor the record's Ann, `Checked` true. `Layer.Shape` is the zero `kinship.Shape` (direct). `Canon.Shape` is `{marriage_spouses, M's handle}`. One bundle: `Member` the tree Bob, `Neighbor` the tree Ann, `Bridges` = [tree Bob's participation, tree Ann's participation], `Observations` hold every Observation on both participations and M's `event_type` Observation |
| AC-1.2 | Tree: Ann and Bob in a `spouse` relationship. Record: marriage M′ with Ann and Bob each `subject` | One relationship, `Checked` true. `Layer.Shape` is `{marriage_spouses, M′}`; `Canon.Shape` is zero. `Incoming` is the 6 Observations of the two record participations, then M′'s `event_type` Observation (7 in all). One bundle with `Bridges` = [the tree relationship Subject] |
| AC-1.3 | Tree: Ann and Bob in a `spouse` relationship **and** marriage M with both as `subject`. Record: `spouse` relationship | One relationship. `Canon.Shape` is zero (a side that records the link directly reads as direct). Two bundles, both `Member` tree Bob, `Neighbor` tree Ann: first `Bridges` = [the relationship Subject], then [Bob's participation, Ann's participation] |
| AC-1.4 | Tree: birth B with Ann `father` and Cal `subject`. Record: Ann `parent` of Cal. Row: the record's Cal on Cal's handle, the record's Ann placed on Ann's handle | One relationship whose key is the record's parent link read from Cal's end. `Canon.Shape` is `{birth_parents, B's handle}`. One bundle: `Member` tree Cal, `Neighbor` tree Ann, `Bridges` = [Cal's participation, Ann's participation], Observations include B's `event_type` Observation |
| AC-1.5 | Tree (one Source): Ann `parent` of Bea, Bea `parent` of Cal. Record: Ann `grandparent` of Cal. Row: record Cal on Cal's handle, record Ann on Ann's handle | One relationship. `Canon.Shape` is `{parent_of_parent, Bea's handle}`. One bundle: `Member` tree Cal, `Neighbor` tree Ann, `Bridges` = [the Bea–Cal relationship Subject, the Ann–Bea relationship Subject] |
| AC-1.6 | As AC-1.5, but Ann `parent` of Bea is on a third Source (Bea there joined to Bea's handle) | One relationship, `Canon.Shape` as AC-1.5. One bundle: `Member` tree Cal, `Neighbor` nil, `Bridges` = [the Bea–Cal relationship Subject]. The Ann–Bea Source has no member of Cal's entity and gives no bundle (D-6) |
| AC-1.7 | AC-1.1 with M dated 12 May 1842; AC-1.2 with M′ dated 12 May 1842; AC-1.5 | AC-1.1: `Canon.PivotProperty` `date`, `Canon.PivotValues` = [12 May 1842] (from M's handle profile). AC-1.2: `Layer.PivotProperty` `date`, `Layer.PivotValues` = [12 May 1842] (M′'s values). AC-1.5: `Canon.PivotProperty` `name`, `Canon.PivotValues` hold Bea's name. A direct side has no pivot property and no values |
| AC-1.8 | Tree: marriage M (Ann, Bob `subject`). Record: marriage M′ (Ann, Bob `subject`). The page places the record's Ann on Ann's handle **and M′ on M's handle** | No `relationship·spouse` line (count-once: both sides imply it through `marriage_spouses` and the pivots are paired). The participation line Bob — M′ (`participation·subject·marriage`) is there |
| AC-1.9 | AC-1.8's catalog, but the request does not place M′ | One `relationship·spouse` line: `Layer.Shape` `{marriage_spouses, M′}`, `Canon.Shape` `{marriage_spouses, M's handle}` |
| AC-1.10 | Tree and record both record Ann and Bob as `spouse` directly (today's case) | Unchanged from `main`: one line, both shapes zero, one bundle per tree relationship Subject with `Bridges` of length 1, the same pin id as `main` |

**US-2.** As a researcher, when I press Done on a row whose sheet I never opened, the implied line's bundle is pinned like any checked relationship, so that the claim keeps the evidence the sheet would have checked.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1's catalog; `ResolvePins` with Bob's row `DefaultPins` on Bob's handle and Ann's row on Ann's handle | Bob's row has a `PinSet` with `Member` the tree Bob and `Observations` = the line's `Incoming` (the record's spouse relationship Observations) followed by the bundle's Observations |
| AC-2.2 | AC-1.1's evidence | `PinOffers` offers the bundle once, checked, with id `BundlePinID(rel, bundle)` = `b:<record Ann>:relationship·spouse:<tree Bob's participation>`; `ResolvePins` with that id in `PinIDs` gives AC-2.1's set |

**US-3.** As the app, I receive how each side recorded a relationship line, so that S11-42 can word it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `claimEvidenceProto` of a relationship with `Layer` direct and `Canon` `{marriage_spouses, M}`, `PivotProperty` `date`, one date value | `layer.pattern` "", `layer.pivot_id` "", no `layer.pivot_values`; `canon.pattern` "marriage_spouses", `canon.pivot_id` = M as a UUID string, `canon.pivot_values` one `ConclusionValue` whose `date` is that date |
| AC-3.2 | `TestSwiftVocabularyMatchesGo` | Swift `KinshipPattern` parses "", `marriage_spouses`, `birth_parents`, `parent_of_parent`, and has an `unrecognized` case |
| AC-3.3 | The app on AC-1.1's data (manual check) | The Promote sheet for Bob shows a relationship Card for Ann, worded as today; checking it and pressing Done pins without error |

## 3. Scope

| In | Covered by |
| --- | --- |
| `canonLinks` reads relationship links from `canonlinks.Reader.Relationships`; participation and other links stay on `Node.Links` | STEP-3, AC-1.1, AC-1.4, AC-1.5 |
| One bridge reader for every canon link, classifying each bridge Subject's ends as the row's member, the neighbor's member, or a pivot | STEP-3, AC-1.1, AC-1.10 |
| `claimevidence`: `Relationship.Layer` / `.Canon` say how each side records the fact, with its pivot and the pivot's date or name | STEP-1, STEP-2, STEP-3, AC-1.1 – AC-1.3, AC-1.7 |
| An implied link's bundle holds every bridge behind it in one Source, and the pivot's defining Observations; `Bundle.Bridges` replaces `Bundle.Bridge` | STEP-1, STEP-2, AC-1.1, AC-1.4 – AC-1.6 |
| The layer side's `Incoming` adds its pivot's defining Observations | STEP-2, AC-1.2 |
| Count-once on relationship lines | STEP-2, AC-1.8, AC-1.9 |
| `BundlePinID` keys on the bundle's first bridge | STEP-4, AC-2.2, AC-1.10 |
| Default pins at Apply (no code: `ResolvePins` calls `evidence`) | STEP-4, AC-2.1 |
| Proto `PromoteClaimEvidenceRecording`, handler, `GoStore` mapping, Swift `KinshipPattern` under `swift_vocabulary_test.go` | STEP-5, STEP-6, AC-3.1 – AC-3.3 |
| Docs: `claim-evidence.md`, `kinship-shapes.md`, `completed.md`, map | STEP-7 |

| Out | Home |
| --- | --- |
| The sheet's sentences for an implied line ("both are spouses in the marriage on <date>") and their L10n | S11-42 |
| A Source that records only the far part of an implied link (no member of the row's entity on its bridges) pinning that part (D-6) | not planned |
| A chain step that is itself implied through a birth adding its participations to the chain's bundle (D-5) | not planned |
| Changing `canonlinks`, `kinship.LinkSet`, `Reader.Relationships` or `Reader.For` | not planned (S11-51 owns them; this PR only reads) |
| Changing `segmentcompare.ReadFrom` / `Key` (S11-41 review M-3) | not planned |
| A stored-claim mode on `promotealign.Evidence` | S13-07 (rebases on whichever lands second) |
| Trees and person pages reading implied links | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md), S13-04 |
| `FakeStore` changes | not planned: it returns canned `CatalogClaimEvidence`, and the new fields have defaults |

## 4. Current state

Verified at `f22b19a`.

| Area | Where | Today |
| --- | --- | --- |
| Evidence query | `core/database/promotealign/evidence.go` `evidence` | Builds `placed` (layer neighbor → handle), `layerLinks` from `in.layer.Bridges` (`b.From(row.SubjectID)`; each link's Observations are its `Associations`' Observations), profiles the handle and neighbor nodes once (`q.prof.profile(nodes)`, `q.prof.keep()`), reads the row's `members`, and per wanted neighbor calls `q.canonLinks(handleNode, hn)` **before** `q.accepted(hid)`. Calls `claimevidence.Build`, then `readFromNeighbor` per relationship, then `q.titles` |
| Canon links | same file `(*evidenceReader).canonLinks` | Iterates `from.Links` with `Neighbor == to.ID`, keys each through `canonlinks.EdgeOf`, reads bridges through `bridges(l)`. Never reads implied links: a tree marriage gives no `relationship·spouse` canon link |
| Bridge reading | same file `(*evidenceReader).bridges` | For one `graphcache.Link`: `r.accepted(l.Association)`, reads each bridge Subject's Observations from its member graph, sets `Member` / `Neighbor` from the bridge spec's endpoint properties (swapped when `!l.FromEnd`), from positive Observations only; an end with two different Subjects, or a missing end, drops the bridge |
| Neighbor members | same file `bridgedOnly`, `memberValues`, `accepted` | `bridgedOnly` keeps neighbor members some `Bridge.Neighbor` names; `accepted` records each Subject's `graphcache.Member` in `memberOf` |
| Titles | same file `(*evidenceReader).titles` | `add(b.Source, b.Bridge)` per bundle |
| Policy types | `core/claimevidence/claimevidence.go` | `Bundle{Member, Neighbor, Source, Bridge, Observations}`; `Relationship` has no recording; `LayerLink{Key, Observations}`; `Bridge{Subject, Source, Member, Neighbor, Observations}`; `CanonLink{Key, Bridges}`; `Inputs{Order}` |
| Build | `core/claimevidence/build.go` `relationships`, `bundles` | Groups layer links by key; a canon link corresponds by `segmentcompare.Corresponds` alone; `Incoming` = the layer links' Observations + the neighbor's defining refs; `bundles` keeps a bridge only when both `Member` and `Neighbor` are members, one bundle per bridge, sorted by member then neighbor index |
| Layer links | `core/database/promotealign/layer.go` `loadLayerBridges` | Merges recorded and implied links through one `kinship.LinkSet`; each `graphalign.Bridge` carries `kinship.Recording` (its shapes' pivots are layer Subjects) and the `Associations` behind it |
| Person links | `core/database/canonlinks/reader.go` `(*Reader).Relationships` | A canonical person's relationship links, recorded (`Associations` = [the filed link's association]) then implied (`Associations` = the bridges behind it, `Recording.Shapes` with handle pivots), merged through `kinship.LinkSet`; the caller owns the result; not filtered |
| Count once | `core/kinship/kinship.go` `CountedOnce`; `core/graphalign/align.go` `corresponding` | Matching skips a corresponding pair both sides imply through one pattern once the walk paired the pivots |
| Pins | `core/database/promotealign/pins.go` `BundlePinID`, `PinOffers`, `ResolvePins` | `BundlePinID` = `b:` + neighbor + `:` + key + `:` + `b.Bridge`. `ResolvePins` builds each default-pin row's offers through `evidence` |
| Values | `core/database/promotealign/values.go` `EvidenceValues` | Types a line's value through unexported `value(property, val, search)` |
| Handler | `api/ffi/handlers/promote_evidence.go` `claimEvidenceProto` | Maps each relationship; bundles by `BundlePinID` |
| Proto | `api/proto/engine.proto` `PromoteClaimEvidenceRelationship` | Fields 1–3, 7, 12–17 used; 4–6, 8–11 reserved. No recording |
| Swift | `macos/App/Platform/GenealogyStore.swift` `CatalogClaimEvidenceRelationship`; `macos/App/Platform/GoStore.swift` `getPromoteClaimEvidence`; `macos/App/Platform/ConclusionVocabulary.swift` | No recording; no kinship pattern enum |
| Vocabulary test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | Pins `EvidenceOutcome`, `EvidencePropertyCause`, … with `wireCase`; a subtest lists the enums that need `unrecognized` |
| Tests | `core/database/promotealign/evidence_test.go` `TestEvidenceReadsASourceMarriageAsSpouses`, `marriageEvidence`; `pins_test.go` `TestPinOffersOfAnImpliedLayerLink`; `core/claimevidence/build_test.go` `TestBuild`; `api/ffi/handlers/promote_evidence_test.go` `TestClaimEvidenceNeighborReading` | AC-3.1 there asserts 6 incoming Observations; AC-3.3 asserts **no** relationship for a tree marriage (this PR flips it); `TestBuild` AC-2.1 reads `Members[0].Bridge` |
| Budget | `core/database/promotealign/perf_test.go` `budgetEvidence` (100 ms) | Must stay green |
| Fixtures | `promotealign_test.go` `fixture.marriage`, `eventOn`, `participation`, `relationship`, `promote`, `join`, `newSource` | Enough to build every AC; `eventOn(src, art, "marriage", &d)` gives a dated marriage |

**Precedent.** S11-25 ([plan](S11-25-claim-evidence-rpc.md); `claimEvidenceProto`) and S11-20's neighbor reading added the relationship's neighbor reading as fields the loader fills after `Build` (`readFromNeighbor`) and a proto block with a handler test that checks reserved names; copy that split: `Build` decides, the loader annotates, the handler maps. S11-39's `TestEvidenceReadsASourceMarriageAsSpouses` / `marriageEvidence` is the fixture shape for every new evidence test here.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `canonLinks(from, to, rels, row, nbr, pivots)` reads `from.Links` with `Neighbor == to.ID` and `BridgeTypeKey != vocabkey.BridgeRelationship` as today, then each `kinship.Link` of `rels` (the row handle's `Reader.Relationships`, read once per request) whose other end is `to.ID`, keyed from `from`'s end (`KeyA` when `A` is `from.ID`, else `KeyB`), with that link's `Recording` | S11-51 D-4 and **Next adds**: participation links stay in `canonLinks`, relationship links come from `Relationships`, so the recorded and implied shapes are already one link | Calling `Relationships` per neighbor — the same answer for every neighbor of one row; reading `Reader.For` and merging here — a second merge beside `kinship.LinkSet` |
| D-2 | One bridge reader, `(*evidenceReader).bridgesOf(assocs [][]byte, row, nbr, pivot map[string]bool) ([]claimevidence.Bridge, error)` *(new)*, replaces `bridges`. For each association: `r.accepted(assoc)`; each bridge Subject's Observations from its member graph (read once per graph, as today); its ends are its positive Observations whose property is an endpoint property of any `connectrules.Bridges()` entry and that carry a `ValueSubjectID`. An endpoint property with two different Subjects drops the bridge. Each end Subject in `row` sets `Member`, in `nbr` sets `Neighbor`, in `pivot` sets `Pivot`; a bridge with fewer than two of the three set is dropped. Filed links pass `[][]byte{l.Association}` | One reader for recorded and implied bridges; ends by membership, so a participation, a relationship and a chain's parent link read alike, with no `FromEnd` swap | Keeping `bridges(l)` beside a second reader for implied associations — two readers of bridge ends (near-copy) |
| D-3 | `claimevidence.Bridge` gains `Pivot []byte`: the end that is a member of a pivot handle (the event, or the parent between). `pivot` (D-2) is the accepted members of every `Shape.Pivot` handle of the links D-1 keeps for this neighbor, read through `r.accepted` | A recorded bridge never has a pivot end, so `Pivot != nil` is exactly "behind an implied link" and `Build` needs no other flag | Marking implied associations on `kinship.Link` — S11-51 owns that type and it has no such field |
| D-4 | `Bundle.Bridge []byte` becomes `Bundle.Bridges [][]byte`. A bridge with `Pivot == nil` gives one bundle as today (`Bridges` of length 1). Bridges with a `Pivot` are grouped by `Pivot` (first-seen order across the corresponding canon links): a group gives one bundle when one of its bridges has a `Member` that is a row member. `Member` is that bridge's member (the first such bridge in input order), `Source` its Source, `Neighbor` the first group bridge's `Neighbor` that is a neighbor member (nil when none), `Bridges` the member's bridge first, then the others in input order; `Observations` = every group bridge's Observations in that order, then the neighbor member's defining refs (when `Neighbor` is set), then the pivot's defining refs, each id once (`appendNew`) | A pivot Subject lives in one Evidence graph, so one group is one Source's recording of the fact, and its backfill lands on the member whose Source recorded it ([kinship shapes: "every bridge behind the implied link"](../kinship-shapes.md#on-the-evidence-sheet)) | One bundle per underlying bridge — a participation's ends are a person and an event, so half the bridges have no row member to backfill onto; one bundle across every Source — backfills one Source's Observations onto another Source's member's claim |
| D-5 | The pivot set (D-3) holds only `Shape.Pivot` handles' members. A chain step that is itself a birth's implied parent link brings participations whose event end is not a pivot, so those bridges have one classified end and are dropped (D-2); the chain's bundle holds its recorded parent links only | Keeps the pivot rule one line; the birth-implied chain step on the canon side is rare and its own line (the birth) carries its participations | Adding every event reached by a step's associations to the pivot set — a second walk for a rare case |
| D-6 | A group with no row member is dropped (AC-1.6). Bundle sort: member index, then neighbor index with a nil `Neighbor` sorting after every member | `promote.PinSet` needs one accepted `Member`; backfilling onto a neighbor's member is a new pin rule | Pinning the far part on the neighbor member's claim — a product change to §9's backfill rule |
| D-7 | `claimevidence.Pivot` *(new)* `{Subject []byte; Kind vocabkey.SubjectType; Values subjectcompare.Values}`. `LayerLink` gains `Recording kinship.Recording` and `Pivots []Pivot` (each `Recording.Shapes` pivot, a layer Subject, with its layer values and kind from `in.primary`). `CanonLink` gains `Recording kinship.Recording` and `Pivots []Pivot` (each pivot Subject its bridges reach, values through `r.memberValues`, kind of the pivot handle through `Vocabulary.TypeKeysByID`). `Build` adds `refsOf(p.Values, rules.Defining[p.Kind])` for each pivot: the layer's into `Incoming` after the neighbor's defining refs, a group's into its bundle (D-4) | "What a bundle holds" stays the evidence section's `Defining` rule in the pure package, as the neighbor's does | Reading the pivot's `event_type` Observation ids in the loader — moves the defining rule out of `claimevidence` |
| D-8 | `Relationship` gains `Layer, Canon Recorded`. `Recorded` *(new)* `{Shape kinship.Shape; PivotProperty vocabkey.PropertyKey; PivotValues []propertycompare.Value}`. `Build` sets `Shape` from the side's merged `kinship.Recording` (the layer links of the key merged with `Recording.Merge`; the corresponding canon links likewise) through `recordedShape` *(new, unexported)*: the zero `Shape` when the recording is not `OnlyImplied` or has no shapes, else `Shapes[0]`. The loader fills `PivotProperty` / `PivotValues` after `Build` (D-10) | One code per side is what S11-42 words; a side that records the fact directly reads as "the record has them as spouses"; `Shapes` keep first-added order, so the choice is stable | A list of every shape per side — no sentence words two shapes; a new `claimevidence` shape enum — `kinship.Pattern` already names the shape and goes on the wire as is (D-11) |
| D-9 | Count-once: `Inputs` gains `Placed map[string][]byte` (layer Subject → the handle the page places it on; the loader passes `evidence`'s `placed`). In `relationships`, a canon link corresponds when `segmentcompare.Corresponds(key, c.Key) && !kinship.CountedOnce(layerRec, c.Recording, paired)`, `paired` reading `Placed` | Working rule "Counted once across sides": the evidence lines follow matching's `corresponding`, through the same function | A second count-once check in the loader — the rule would have two homes |
| D-10 | `pivotShown` *(new, unexported, `evidence.go`)* `= map[vocabkey.SubjectType]vocabkey.PropertyKey{vocabkey.Event: vocabkey.PropertyDate, vocabkey.Person: vocabkey.PropertyName}`. After `Build`, `readPivots` *(new)* fills each relationship's `Layer` from the layer pivot's values (`valuesOf[pivot][prop]`, kind from `in.primary`) and `Canon` from the pivot handle's profile (`q.prof.values(pivot)[prop]`). The canon pivot handles are loaded and added to `nodes` before the one `q.prof.profile(nodes)` call | The sentence names the event's date or the parent between's name (S11-42); a neighbor-side reading is filled after `Build` the same way (`readFromNeighbor`). The handle's profile is the tree's reconciled value | Wording-only data in the pure policy package; carrying only the pivot id and reading headers in Swift — the canon pivot is a handle the Promote page has no header for |
| D-11 | Wire: `message PromoteClaimEvidenceRecording { string pattern = 1; string pivot_id = 2; repeated ConclusionValue pivot_values = 3; }` *(new)*; `PromoteClaimEvidenceRelationship` gains `PromoteClaimEvidenceRecording layer = 18; PromoteClaimEvidenceRecording canon = 19;`, always set. `pattern` is `kinship.Pattern` ("" when direct). `EvidenceValues.Pivot(r claimevidence.Recorded) []EvidenceValue` *(new)* types each value as `value(r.PivotProperty, v, nil)` does | `kinship.Pattern` is already a stable code; one wire vocabulary (working rule) | A separate `recorded_how` code beside `pattern` — `pattern` says direct, event (marriage, birth) or chain already |
| D-12 | Swift: `enum KinshipPattern` *(new)* in `ConclusionVocabulary.swift`, cases `direct` (""), `marriageSpouses`, `birthParents`, `parentOfParent`, `unrecognized(String)`, `init(wire:)` like `EvidenceOutcome`. `struct CatalogClaimEvidenceRecording` *(new)* in `GenealogyStore.swift` `{pattern: KinshipPattern = .direct; pivotID = ""; pivotValues: [CatalogConclusionValue] = []}`; `CatalogClaimEvidenceRelationship` gains `layer` and `canon` defaulting to it. `GoStore` maps through `mapEvidenceRecording` *(new, private static)*. No Swift unit test: the wire enum is pinned by `TestSwiftVocabularyMatchesGo`, as every other evidence enum is | Working rule "One wire vocabulary"; defaults keep `FakeStore` and every existing test call compiling | Parsing the pattern only in S11-42 — the map puts the vocabulary in this PR |
| D-13 | `BundlePinID` keys on `b.Bridges[0]`; `titles` reads `b.Bridges[0]`. Bundles of one relationship never share a first bridge (D-4: a group's member bridge is in no other group, and a recorded bundle's one bridge has no pivot) | Recorded bundles keep `main`'s ids (AC-1.10); an implied bundle is named by its member's own bridge | Joining every bridge id — a longer id with no extra uniqueness |
| D-14 | The evidence request reads `Relationships` through a `canonlinks.Reader` built once in `newEvidenceReader` (`kin` field *(new)*, `canonlinks.NewReader(g, matchregistry.Default().Kinship)`) | `ResolvePins` reuses one reader across a page, as it reuses the profiler | A reader per `evidence` call — reloads the terms per row |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `claimevidence.Recorded` | type | `kinship.Shape`, `kinship.Recording` | It embeds the chosen `kinship.Shape` and adds the pivot's shown values; `Recording` keeps every shape, which the line does not word |
| `claimevidence.Pivot` | type | `claimevidence.Member` (`Subject`, `Source`, `Values`) | A pivot has a kind and no Source; `Member` is a member of the row's or neighbor's entity, and reusing it would blur which list a value belongs to |
| `Bridge.Pivot`, `Bundle.Bridges`, `LayerLink.Recording/Pivots`, `CanonLink.Recording/Pivots`, `Relationship.Layer/Canon`, `Inputs.Placed` | fields | the same types | Extends the types the query already passes |
| `recordedShape` | func | `kinship.Recording` | A two-line reading of a `Recording` for one line; the kinship package has no "one shape to word" rule |
| `bridgesOf` | func | `(*evidenceReader).bridges` | Replaces it (Deletes) |
| `readPivots`, `pivotShown` | func, var | `readFromNeighbor`; `layer.go` `neighborTypeProperty` | Same post-`Build` shape and same kind → property map idiom; different data |
| `evidenceReader.kin` | field | `canonlinks.NewReader` in `canon.go` `loadCanon` | The evidence query's own reader, as `loadCanon` has its own |
| `EvidenceValues.Pivot` | method | `EvidenceValues.Incoming`, `.Member` | Same typing through `value`; a pivot has no section |
| `PromoteClaimEvidenceRecording` | proto message | `ConclusionTerm`, `PromoteClaimEvidenceBundle` | Neither carries a pattern and pivot |
| `evidenceRecordingProto` | func | `evidenceTermProto` | Different message |
| `KinshipPattern` | Swift enum | `EvidenceOutcome` | New wire vocabulary, same idiom |
| `CatalogClaimEvidenceRecording`, `mapEvidenceRecording` | Swift struct, func | `CatalogEvidenceTerm`, `mapEvidenceTerm` | Different message |
| `TestBuildImpliedLinks`, `TestEvidenceImpliedCanonLinks`, `TestEvidenceCountsASharedMarriageOnce`, `TestResolvePinsImpliedCanonLink`, `TestClaimEvidenceRecording` | tests | `TestBuild`, `TestEvidenceReadsASourceMarriageAsSpouses`, `TestResolvePinsBundlesARelationship`, `TestClaimEvidenceNeighborReading` | New behavior; each reuses the existing fixtures (`newScene`, `marriageEvidence`'s shape, `fixture`) |

Searched: `rg "Bridge\b" core/claimevidence core/database/promotealign api/ffi/handlers`, `rg "Relationships\(" core api`, `rg "marriage_spouses|parent_of_parent" macos api`, `rg "func \(r \*evidenceReader\)" core/database/promotealign`.

**Builds on.** `canonlinks.Reader.Relationships` (S11-51) and `kinship.Recording` / `CountedOnce` (S11-39); the evidence section's `Defining` rule; the post-`Build` annotation pattern (`readFromNeighbor`); the proto reserved-names test; the `wireCase` vocabulary test.
**Deletes.** `(*evidenceReader).bridges`; `Bundle.Bridge`; the canon side's own reading of relationship links in `canonLinks`; `TestEvidenceReadsASourceMarriageAsSpouses` subtest "AC-3.3" (no relationship for a tree marriage), replaced by T-3.
**Next adds.** S11-42 words `Relationship.Layer` / `.Canon` from `KinshipPattern` and the pivot values in the existing Card. A new kinship pattern adds a `kinship.Pattern` and one `KinshipPattern` case, and its pivot kind one `pivotShown` entry when it is a new kind. S13-07's stored-claim mode keeps `canonLinks` and `bridgesOf` as they are.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/claimevidence` (`claimevidence.go`) | `type Pivot struct { Subject []byte; Kind vocabkey.SubjectType; Values subjectcompare.Values }` *(new)*. `type Recorded struct { Shape kinship.Shape; PivotProperty vocabkey.PropertyKey; PivotValues []propertycompare.Value }` *(new)*. `Bundle.Bridges [][]byte` replaces `Bridge` (doc: the member's bridge first). `Relationship` gains `Layer, Canon Recorded` (doc: the loader fills `PivotProperty` and `PivotValues` after Build). `LayerLink` gains `Recording kinship.Recording; Pivots []Pivot`. `Bridge` gains `Pivot []byte`. `CanonLink` gains `Recording kinship.Recording; Pivots []Pivot`. `Inputs` gains `Placed map[string][]byte` |
| `core/claimevidence` (`build.go`) | `relationships`: per key, merge the layer links' `Recording`s (`Merge` into a clone of the first) and `Pivots` (each Subject once); canon links correspond per D-9; `Incoming` = observations, neighbor defining refs, then each layer pivot's defining refs; `rel.Layer.Shape` / `rel.Canon.Shape` per D-8 (canon: the corresponding links' recordings merged). `bundles`: D-4 / D-6, pivots looked up from the corresponding canon links' `Pivots`. `recordedShape(kinship.Recording) kinship.Shape` *(new)* |
| `core/database/promotealign` (`evidence.go`) | `evidenceReader.kin *canonlinks.Reader` *(new)*. `evidence`: after the handle node check, `rels, err := q.kin.Relationships(row.HandleID)`; canon pivot handles of links to wanted neighbors' handles are loaded (`g.Node`) and added to `nodes` before `q.prof.profile(nodes)`; per neighbor, `q.accepted(hid)` before `q.canonLinks(...)`; `Inputs.Placed: placed`; each layer link gets `Recording` and `Pivots`; after `Build`, `readFromNeighbor` then `readPivots`. `canonLinks` per D-1, building `pivot` per D-3 and calling `bridgesOf` per D-2, filling `CanonLink.Pivots` per D-7. `bridgesOf` *(new)* replaces `bridges`. `titles` reads `b.Bridges[0]`. `pivotShown`, `readPivots` *(new)* per D-10 |
| `core/database/promotealign` (`pins.go`) | `BundlePinID` reads `b.Bridges[0]` (D-13) |
| `core/database/promotealign` (`values.go`) | `func (v EvidenceValues) Pivot(r claimevidence.Recorded) []EvidenceValue` *(new)* |

```go
// Recorded is how one side records a relationship line's fact. Shape is the
// zero kinship.Shape when the side records the link directly; else the first
// shape that implies it, whose Pivot is a layer Subject on the row's side and a
// handle on the entity's. PivotValues are the pivot's PivotProperty values
// (an event's date, the parent between's name); the loader fills both.
type Recorded struct {
	Shape         kinship.Shape
	PivotProperty vocabkey.PropertyKey
	PivotValues   []propertycompare.Value
}
```

Errors: none new; every new read returns the graph cache's error as today.

### FFI

`api/proto/engine.proto`: `message PromoteClaimEvidenceRecording` *(new)* before `PromoteClaimEvidenceRelationship`, comment "How one side records a relationship's fact (S11-40): the kinship pattern ("" when recorded directly), the event or parent between it passes through (a layer Subject on the row's side, a handle on the entity's), and that pivot's date or name."; `PromoteClaimEvidenceRelationship` gains `layer = 18`, `canon = 19`. Regenerate with `./scripts/generate-proto.sh` (Go and Swift). Handler `api/ffi/handlers/promote_evidence.go`: `evidenceRecordingProto(r claimevidence.Recorded, values promotealign.EvidenceValues) *engine.PromoteClaimEvidenceRecording` *(new)*: `Pattern: string(r.Shape.Pattern)`, `PivotId: database.UUIDString(r.Shape.Pivot)` when `len(r.Shape.Pivot) == 16` (else ""), `PivotValues` from `values.Pivot(r)` through `evidenceValueProto`; `claimEvidenceProto` sets `Layer` and `Canon` on every relationship.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Wire enum | `KinshipPattern` *(new)* | `macos/App/Platform/ConclusionVocabulary.swift`, after `EvidencePropertyCause` | D-12 |
| Model | `CatalogClaimEvidenceRecording` *(new)*; `CatalogClaimEvidenceRelationship.layer`, `.canon` | `macos/App/Platform/GenealogyStore.swift` | D-12 |
| Mapping | `mapEvidenceRecording` *(new)* | `macos/App/Platform/GoStore.swift` | `pivotValues: r.pivotValues.map(mapConclusionValue)` |

No view changes; no state changes. `FakeStore` unchanged.

### Strings

None — no user-facing copy (S11-42).

## 7. Steps

**STEP-1 — Policy types and recorded shapes**
Files: `core/claimevidence/claimevidence.go`, `build.go`, `build_test.go`. Skill: `add-match-rule` (the package stays pure).
Do: add `Pivot`, `Recorded`, the new fields (§6 Go core), rename `Bundle.Bridge` → `Bridges` (bundles of a recorded bridge set `Bridges: [][]byte{b.Subject}`), `recordedShape`, merging of the layer links' recordings and pivots per key, `Relationship.Layer.Shape` and `Relationship.Canon.Shape` (D-8). Update `TestBuild` AC-2.1 to read `Members[0].Bridges`. Add T-1 subtests (a) and (d).
Check: `go test ./core/claimevidence/... ./core/matchregistry/...` passes (no build tags; the guard passes).

**STEP-2 — Implied bundles, pivot defining refs, count-once**
Files: `core/claimevidence/build.go`, `build_test.go`.
Do: D-4, D-6, D-7 (`refsOf` for pivots), D-9 (`Inputs.Placed`, `kinship.CountedOnce`). Add T-1 subtests (b), (c), (e), (f), (g).
Check: `go test ./core/claimevidence/...` passes.

**STEP-3 — The canon side reads person links**
Files: `core/database/promotealign/evidence.go`, `evidence_test.go`, `export_test.go` (if `BridgedOnly` or a new export needs it). Skill: `add-match-rule`.
Do: D-1, D-2, D-3, D-7 (loader side), D-10, D-14. Delete `bridges`. Update `marriageEvidence` callers: T-3. Add T-4, T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes; `rg -n "func \(r \*evidenceReader\) bridges\(" core/` prints nothing.

**STEP-4 — Pins**
Files: `core/database/promotealign/pins.go`, `pins_test.go`.
Do: D-13. Add T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes, `TestResolvePins*` and `TestPinOffersOfAnImpliedLayerLink` included.

**STEP-5 — Wire**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/promote_evidence.go`, `promote_evidence_test.go`, `core/database/promotealign/values.go`. Skill: `add-ffi-handler`.
Do: §6 FFI, D-11. Add T-7.
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-6 — Swift model and vocabulary**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `GenealogyStore.swift`, `GoStore.swift`, `api/ffi/swift_vocabulary_test.go`.
Do: D-12. Add the "kinship patterns" row to `TestSwiftVocabularyMatchesGo` (`want: sorted("", string(kinship.PatternMarriageSpouses), string(kinship.PatternBirthParents), string(kinship.PatternParentOfParent))`) and `"KinshipPattern"` to the "tolerate a value this build doesn't know" list: T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `xcodebuild test` (the `scripts/check-pr.sh` command) passes with no Swift warning.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-11/claim-evidence.md` (§ Evidence through relationships: a paragraph "Implied links (S11-40)" stating D-4, D-6, D-8 and the count-once rule), `kinship-shapes.md` (§ On the evidence sheet: the bundle is one per Source recording with the member's bridge first, keyed by it; open question 2 answered: yes, S11-40 and S11-42), `completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-40-implied-links-evidence.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a tree marriage as a spouse line on the evidence sheet."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/claimevidence/build_test.go` `TestBuildImpliedLinks` *(new)* | (a) layer link `Recording{Shapes: [{marriage_spouses, P}], OnlyImplied: true}`, canon recorded → `Layer.Shape == {marriage_spouses, P}`, `Canon.Shape` zero. (b) canon link with bridges `{Subject 0x70, Member m, Pivot e}`, `{Subject 0x71, Neighbor n, Pivot e}`, `Pivots [{e, Event, event_type ref 0x90}]` → one bundle `Member m`, `Neighbor n`, `Bridges [0x70 0x71]`, Observations = 0x70's, 0x71's, n's defining refs, 0x90. (c) same with only the 0x71 bridge → no bundle; with only 0x70 → one bundle, `Neighbor` nil, sorted after a recorded bundle of the same member. (d) a recording with a shape and `OnlyImplied` false → zero `Shape`. (e) both sides `{marriage_spouses, …}` only implied, `Placed[layer pivot] = canon pivot` → no relationship; `Placed` empty → one. (f) both sides recorded, pivots placed → one relationship (two recorded links are never skipped). (g) layer `Pivots [{p, Event, event_type ref 0x91}]` → `Incoming` ends with 0x91 | AC-1.1 – AC-1.3, AC-1.6, AC-1.8, AC-1.9, D-4, D-6 – D-9 |
| T-2 | Go | `core/claimevidence/build_test.go` `TestBuild` AC-2.1, AC-2.5 | Unchanged inputs → `Members[i].Bridges == [[bridge]]` | AC-1.10, D-4 |
| T-3 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceReadsASourceMarriageAsSpouses` (subtests renamed "AC-1.2", "AC-1.1") | `marriageEvidence(t, true)` → AC-1.2 (7 incoming, `Layer.Shape.Pattern == marriage_spouses`, `Layer.Shape.Pivot` = the record's marriage). `marriageEvidence(t, false)` → AC-1.1 (one line, `Canon.Shape`, one bundle of two bridges whose first is the tree Bob's participation; Observations contain the tree marriage's `event_type` Observation). `marriageEvidence` returns the tree marriage and the record marriage too | AC-1.1, AC-1.2 |
| T-4 | Go | `evidence_test.go` `TestEvidenceImpliedCanonLinks` *(new)* | Subtests "AC-1.3 both shapes", "AC-1.4 birth parents", "AC-1.5 chain", "AC-1.6 chain across Sources", "AC-1.7 pivot values" (dated marriage through `eventOn(…, "marriage", &d)`; chain → Bea's name), "AC-1.10 recorded both sides" (bundle `Bridges` length 1 and `BundlePinID` equal to the `b:<neighbor>:<key>:<relationship Subject>` form) | AC-1.3 – AC-1.7, AC-1.10 |
| T-5 | Go | `evidence_test.go` `TestEvidenceCountsASharedMarriageOnce` *(new)* | AC-1.8 (no `relationship·spouse`; a `participation·subject·marriage` line); AC-1.9 (one spouse line with both shapes) | AC-1.8, AC-1.9, D-9 |
| T-6 | Go | `core/database/promotealign/pins_test.go` `TestResolvePinsImpliedCanonLink` *(new)* | AC-1.1's catalog: `ResolvePins` with Bob's row `DefaultPins` → AC-2.1's set; `PinOffers` id → AC-2.2; that id in `PinIDs` → the same set | AC-2.1, AC-2.2 |
| T-7 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestClaimEvidenceRecording` *(new)* | AC-3.1 through `claimEvidenceProto` on a hand-built `claimevidence.Evidence` (as `TestClaimEvidenceNeighborReading` does); every relationship has `layer` and `canon` set | AC-3.1, D-11 |
| T-8 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "kinship patterns" *(new row)* and the `unrecognized` subtest | Swift `KinshipPattern` cases equal Go's patterns plus "" | AC-3.2, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestEvidenceBridgeEndpoints`, `TestEvidenceLoadsOnlyBridgedNeighborMembers`, `TestEvidenceReadsEachLinkFromTheNeighbor`, `TestResolvePinsBundlesARelationship`, `TestPinOffersOfAnImpliedLayerLink` (they pin `bridgesOf` against `bridges`' behavior), and `TestLargeSourceBudgets` `budgetEvidence` (100 ms).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI change
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: claim-evidence.md § Evidence through relationships; kinship-shapes.md § On the evidence sheet and open question 2
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project. On Source "Tree", add Ann Ames and Bob Ames and a marriage with both as `subject`; promote all three. On Source "Register", add Ann Ames and Bob Ames in a `spouse` relationship.
2. Open Promote for "Register". Bob's row targets the tree Bob, Ann's the tree Ann.
3. Open Bob's evidence sheet: a relationship Card for Ann is shown, worded as today ("…they share a spouse …"), checked (AC-3.3). On `main` this Card is absent.
4. Press Done without unchecking. `identity_claim_evidence` for Bob's new claim holds both tree marriage participations' Observations.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `bridgesOf`'s membership reading differs from `bridges`' endpoint reading for a filed link | The existing bridge tests (§8) pin today's bundles; any change there is a stop |
| `Relationships` per request pushes the evidence budget | One call per request (D-1, D-14), and `Reader.For` is memoized; `budgetEvidence` must pass |
| A pivot Subject is also a member of the row's or the neighbor's entity | Classification order in D-2 is row, then neighbor, then pivot; a pivot handle is an event or the parent between, never the row's or neighbor's handle |

**Stop and report if:**

- `canonlinks.Reader.Relationships` no longer returns recorded links with `Associations` = [the filed association] and implied links with handle pivots in `Recording.Shapes`.
- Any test under "Most likely to break" in §8 needs its expected values changed (other than `Bridge` → `Bridges[0]` reads).
- `TestLargeSourceBudgets` fails `budgetEvidence` with this change and passes without it.
- A filed participation link's bridge Subject has a positive endpoint Observation on a property that no `connectrules.Bridges()` entry names as an endpoint, so D-2 drops a bridge `bridges` kept.
- `claimevidence` cannot import `core/kinship` without `TestLayerPackages` failing.
- S13-07 has landed and changed `evidence` or `canonLinks`: rebase, then report the rows of §4 that moved.

## 12. Open questions

None.

## 13. PR

**Title:** S11-40 — Implied links in claim evidence
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-40-implied-links-evidence.md`.
