# S11-40 — Implied links in claim evidence

**Status:** Landed in [#579](https://github.com/mendahu/provenencia/pull/579) (2026-10-27)
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
**Environment:** STEP-1 – STEP-7 need Go with cgo and the `fts5` tag. STEP-8 (`./scripts/generate-proto.sh`), STEP-9 (`xcodebuild test`), `scripts/check-pr.sh` and the manual check need a Mac with the repo's toolchain. A Linux session does STEP-1 – STEP-7 and hands the rest to a Mac session.
**Context:** [Kinship shapes § On the evidence sheet](../kinship-shapes.md#on-the-evidence-sheet), [§ Counting once](../kinship-shapes.md#counting-once), [§ Where implied links live](../kinship-shapes.md#where-implied-links-live); [`claim-evidence.md` § Evidence through relationships](../claim-evidence.md#evidence-through-relationships); S11-51 plan D-4, D-5 and its **Next adds**; S11-39's Landed note; working rules: "The canon's implied links live in the graph cache", "One fact, one link per side", "Counted once across sides", "One wire vocabulary", "The pin policy has one owner, in Go"; Spike 13 working rule "One person header for every tile"
**Planned:** 2026-10-26 against `main` @ `f22b19a`
**Landed:** Differences from the plan: STEP-4 and STEP-5 landed in one commit, because reading `Relationships` makes the old subtest "AC-3.3" (no line for a tree marriage) fail before STEP-5 flips it. `kinship.Shape` holds a `[]byte`, so tests compare shapes by `Pattern` or `reflect.DeepEqual`. Swift reads the canon pivot header through the generated oneof (`case .person(let h)?`), not `hasPerson`. `CatalogClaimEvidenceShape` holds the pivot header as one `pivot: CatalogConclusionHeader?`, not D-13's `pivotPerson` / `pivotEvent`: the same Person/Event header the Go side reads through `conclusionheaders.ByIDs`. The canon `Pivot.Kind` comes from the bridge index's type map. The manual check was not run.
**Review:** [`S11-40-implied-links-evidence.review.md`](S11-40-implied-links-evidence.review.md) — Fits with changes / Ready with changes, 2 rounds (every finding answered)

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

The evidence sheet shows a relationship line when the tree records a kinship fact in a different shape from the Source. Today only the Source side reads the kinship patterns: a Source marriage meets a filed spouse relationship, but a tree marriage never meets a Source spouse relationship, and nothing on the line says how either side recorded the fact. After this PR the evidence query's canon side reads a person's recorded and implied links through `canonlinks.Reader.Relationships`; each relationship line says how each side recorded the fact (directly, or through which pattern and pivot); the tree's pivot travels with its existing `EventHeader` or `PersonHeader`; an implied link's bundle pins every bridge behind it; and the line follows the count-once rule. The RPC carries it to Swift. The sheet's words do not change until S11-42.

**Why now.** S11-39 made matching read both shapes and S11-51 gave one read of a person's merged links. The evidence sheet is the last reader that still sees only filed links on the canon side, so a row Promote proposes because of a tree marriage opens a sheet with no relationship line for it (S11-39's `TestEvidenceReadsASourceMarriageAsSpouses` "AC-3.3" pins that gap).

**What stays.** Every property line, likeness line and section order. Every relationship line between filed links and recorded layer links keeps its key, outcome, checked state, incoming Observations, bundles, sections and pin ids. Matching, bands and proposals. The sheet's sentences and chrome (S11-42). `canonlinks` is read, not changed; `core/kinship` gains three small read helpers (D-9, D-12).

## 2. User stories

Fixture names: "tree" is the first Source, every Subject on it promoted (`f.promote`) unless the AC says otherwise; "record" is a second Source, unpromoted. "A parent of B" is `f.relationship(src, art, A, B, "parent")` (the fixture's "person is typ of related"). The row is the record's Bob on the tree Bob's handle, and the request places the record's Ann on the tree Ann's handle, unless the AC names other rows.

**US-1.** As a researcher, when the tree records a kinship fact through a marriage, a birth or a parent chain and the record records it directly (or the other way round), the sheet shows one relationship line for it whose pins hold every bridge behind it, so that I can pin the evidence matching already counted.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Tree: marriage M with Ann and Bob each `subject`. Record: Ann and Bob in a `spouse` relationship | One relationship: key `relationship·spouse`, neighbor the record's Ann, `Checked` true. `Layer` is the zero `kinship.Shape` (direct). `Canon` is `{marriage_spouses, M's handle}`. One bundle: `Member` the tree Bob, `Neighbor` the tree Ann, `Bridges` = [tree Bob's participation, tree Ann's participation], `Observations` = every Observation on Bob's participation, then on Ann's, then M's `event_type` Observation |
| AC-1.2 | Tree: Ann and Bob in a `spouse` relationship. Record: marriage M′ with Ann and Bob each `subject` | One relationship, `Checked` true. `Layer` is `{marriage_spouses, M′}`; `Canon` is zero. `Incoming` is the 6 Observations of the two record participations, then M′'s `event_type` Observation (7 in all). One bundle with `Bridges` = [the tree relationship Subject] |
| AC-1.3 | Tree: Ann and Bob in a `spouse` relationship **and** marriage M with both as `subject`. Record: `spouse` relationship | One relationship. `Canon` is zero (a side that records the link directly reads as direct). Two bundles, both `Member` tree Bob, `Neighbor` tree Ann: first `Bridges` = [the relationship Subject], then [Bob's participation, Ann's participation] |
| AC-1.4 | Tree: birth B with Ann `father` and Cal `subject`. Record: Ann parent of Cal. Row: the record's Cal on Cal's handle, the record's Ann placed on Ann's handle | One relationship whose key is the record's parent link read from Cal's end. `Canon` is `{birth_parents, B's handle}`. One bundle: `Member` tree Cal, `Neighbor` tree Ann, `Bridges` = [Cal's participation, Ann's participation], Observations end with B's `event_type` Observation |
| AC-1.5 | Tree (one Source): Ann parent of Bea, Bea parent of Cal. Record: Ann `grandparent` of Cal. Row: record Cal on Cal's handle, record Ann on Ann's handle | One relationship. `Canon` is `{parent_of_parent, Bea's handle}`. One bundle: `Member` tree Cal, `Neighbor` tree Ann, `Bridges` = [the Bea–Cal relationship Subject, the Ann–Bea relationship Subject] |
| AC-1.6 | As AC-1.5, but Ann parent of Bea is on a third Source, whose Bea is joined to the tree Bea's handle (`f.join`) and whose Ann is joined to the tree Ann's handle | One relationship, `Canon` as AC-1.5. One bundle: `Member` tree Cal, `Neighbor` nil, `Bridges` = [the Bea–Cal relationship Subject]. The third Source's part has no member of Cal's entity and gives no bundle (D-6) |
| AC-1.7 | AC-1.1, with the record's Ann and the tree Ann both named "Ann Ames" | The relationship's name section compares the record Ann's name with the tree Ann's (the group's `Neighbor` is named, as a recorded bundle's is). In AC-1.6, the relationship has no section member from the tree Ann (nothing named) |
| AC-1.8 | Tree: marriage M (Ann, Bob `subject`). Record: marriage M′ (Ann, Bob `subject`). The request places the record's Ann on Ann's handle **and M′ on M's handle** | No `relationship·spouse` line (count-once: both sides imply it through `marriage_spouses` and the pivots are paired). A participation line Bob — M′ (key type `participation`, event type `marriage`) is there |
| AC-1.9 | AC-1.8's catalog, but the request does not name M′ | One `relationship·spouse` line: `Layer` `{marriage_spouses, M′}`, `Canon` `{marriage_spouses, M's handle}` |
| AC-1.10 | Tree and record both record Ann and Bob as `spouse` directly (today's case) | Unchanged from `main`: one line, both shapes zero, one bundle per tree relationship Subject with `Bridges` of length 1, the same pin id and sections as `main` |

**US-2.** As a researcher, when I press Done on a row whose sheet I never opened, the implied line's bundle is pinned like any checked relationship, so that the claim keeps the evidence the sheet would have checked.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1's catalog; `ResolvePins` with Bob's row `DefaultPins` on Bob's handle and Ann's row on Ann's handle (no pins asked) | Bob's row has a `PinSet` with `Member` the tree Bob and `Observations` = the line's `Incoming` (the record's spouse relationship Observations) followed by the bundle's Observations. Ann's row is not asserted |
| AC-2.2 | AC-1.1's evidence | `PinOffers` offers the bundle once, checked, with id `BundlePinID(rel, bundle)` = `b:<record Ann>:relationship·spouse:<tree Bob's participation>`; `ResolvePins` with that id in Bob's `PinIDs` gives AC-2.1's set |

**US-3.** As the app, I receive how each side recorded a relationship line, and the tree pivot's header, so that S11-42 can word it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `claimEvidenceProto` on a relationship whose `Layer` is direct and whose `Canon` is `{marriage_spouses, E}`, E a promoted, dated event's handle | `layer.pattern` "", `layer.pivot_id` "", no `layer` header; `canon.pattern` "marriage_spouses", `canon.pivot_id` = E as a UUID string, `canon.event` = E's `EventHeader` with its date, `canon.person` unset |
| AC-3.2 | `claimEvidenceProto` on a relationship whose `Canon` is `{parent_of_parent, P}`, P a promoted person's handle | `canon.pattern` "parent_of_parent", `canon.person` = P's `PersonHeader` |
| AC-3.3 | `claimEvidenceProto` on a relationship whose `Layer` is `{marriage_spouses, S}`, S a layer Subject id, and whose `Canon` is direct | `layer.pattern` "marriage_spouses", `layer.pivot_id` = S as a UUID string, no `layer` header; `canon.pattern` "" |
| AC-3.4 | `TestSwiftVocabularyMatchesGo` | Swift `KinshipPattern` parses "" and every `kinship.Patterns()` code, and has an `unrecognized` case |
| AC-3.5 | The app on AC-1.1's data (manual check) | The Promote sheet for Bob shows a relationship Card for Ann, worded as today; Done pins without error |

## 3. Scope

| In | Covered by |
| --- | --- |
| `canonLinks` reads relationship links from `canonlinks.Reader.Relationships`; participation and other links stay on `Node.Links` | STEP-4, STEP-5, AC-1.1, AC-1.4, AC-1.5 |
| One bridge reader for every canon link, reading each bridge Subject's ends through its own bridge spec and classifying them as the row's member, the neighbor's member, or a pivot | STEP-3, AC-1.10 |
| `claimevidence`: `Relationship.Layer` / `.Canon` say how each side records the fact (`kinship.Shape`) | STEP-2, STEP-4, STEP-5, AC-1.1 – AC-1.3 |
| An implied link's bundle holds every bridge behind it in one Source, and the pivot's defining Observations; `Bundle.Bridges` replaces `Bundle.Bridge` | STEP-1, STEP-2, STEP-5, AC-1.1, AC-1.4 – AC-1.7 |
| The layer side's `Incoming` adds its pivot's defining Observations | STEP-2, STEP-5, AC-1.2 |
| Count-once on relationship lines | STEP-2, STEP-6, AC-1.8, AC-1.9 |
| `core/kinship`: `Recording.Shown`, `Pattern.PivotKind`, `Patterns` | STEP-1, AC-1.3, AC-3.2, AC-3.4 |
| Default pins at Apply (no code: `ResolvePins` calls `evidence`) | STEP-5, AC-2.1, AC-2.2 |
| `conclusionheaders.ByIDs`: one read of every header kind by id, used by the pivots, `proposalProto` and `fillHitHeaders` | STEP-7, AC-3.1, AC-3.2 |
| Proto `PromoteClaimEvidenceShape` with the canon pivot's header, handler, `GoStore` mapping, Swift `KinshipPattern` under `swift_vocabulary_test.go` | STEP-8, STEP-9, AC-3.1 – AC-3.5 |
| Docs: `claim-evidence.md`, `kinship-shapes.md`, `add-match-rule` skill, `completed.md`, map | STEP-10 |

| Out | Home |
| --- | --- |
| The sheet's sentences for an implied line ("both are spouses in the marriage on <date>") and their L10n | S11-42 |
| A Source that records only the far part of an implied link (no member of the row's entity on its bridges) pinning that part (D-6) | not planned |
| A chain step that is itself implied through a birth adding its participations to the chain's bundle (D-7) | not planned |
| Changing `canonlinks`, `kinship.LinkSet`, `Reader.Relationships` or `Reader.For` | not planned (S11-51 owns them; this PR only reads) |
| One `canonlinks.Reader` shared by `loadCanon` and the evidence query through `inputs` (review M-7) | not planned here; logged in the review log |
| Changing `segmentcompare.ReadFrom` / `Key` (S11-41 review M-3) | not planned |
| A stored-claim mode on `promotealign.Evidence` | S13-07 (rebases on whichever lands second) |
| Trees and person pages reading implied links | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md), S13-04 |
| `FakeStore` changes | not planned: it returns canned `CatalogClaimEvidence`, and the new fields have defaults |

## 4. Current state

Verified at `f22b19a`, and again by the executability review's dry run.

| Area | Where | Today |
| --- | --- | --- |
| Evidence query | `core/database/promotealign/evidence.go` `evidence` | Builds `placed` (layer neighbor → handle), `layerLinks` from `in.layer.Bridges` (`b.From(row.SubjectID)`; each link's Observations are its `Associations`' Observations), profiles the handle and neighbor nodes once (`q.prof.profile(nodes)`, `q.prof.keep()`), reads the row's `members` (`[]claimevidence.Member`), and per wanted neighbor calls `q.canonLinks(handleNode, hn)` **before** `q.accepted(hid)`. Calls `claimevidence.Build`, then `readFromNeighbor` per relationship, then `q.titles` |
| Canon links | same file `(*evidenceReader).canonLinks` | Iterates `from.Links` with `Neighbor == to.ID`, keys each through `canonlinks.EdgeOf`, reads bridges through `bridges(l)`. Never reads implied links: a tree marriage gives no `relationship·spouse` canon link |
| Bridge reading | same file `(*evidenceReader).bridges` | For one `graphcache.Link`: `connectrules.LookupBridge(l.BridgeTypeKey)`, `r.accepted(l.Association)`, reads each bridge Subject's Observations from its member graph, sets `Member` / `Neighbor` from the spec's endpoint properties (swapped when `!l.FromEnd`), from positive Observations only; an end with two different Subjects, or a missing end, drops the bridge |
| Neighbor members | same file `bridgedOnly`, `memberValues`, `accepted` | `bridgedOnly` keeps neighbor members some `Bridge.Neighbor` names; `accepted` records each Subject's `graphcache.Member` in `memberOf` |
| Titles | same file `(*evidenceReader).titles` | `add(b.Source, b.Bridge)` per bundle |
| Evidence graph Subjects | `core/database/graphcache/evidence_graph.go` `EvidenceSubject.TypeKey` | Each Subject on a graph carries its type key; a bridge Subject's is its bridge type (`participation`, `relationship`) |
| Policy types | `core/claimevidence/claimevidence.go` | `Bundle{Member, Neighbor, Source, Bridge, Observations}`; `Relationship` has no shapes; `LayerLink{Key, Observations}`; `Bridge{Subject, Source, Member, Neighbor, Observations}`; `CanonLink{Key, Bridges}`; `Inputs{Order}` |
| Build | `core/claimevidence/build.go` `relationships`, `bundles` | Groups layer links by key; a canon link corresponds by `segmentcompare.Corresponds` alone; `Incoming` = the layer links' Observations + the neighbor's defining refs; `bundles` keeps a bridge only when both `Member` and `Neighbor` are members, sets `named[ni]`, one bundle per bridge, stable-sorted by member then neighbor index; Sections compare the neighbor's other properties with the `named` members |
| Layer links | `core/database/promotealign/layer.go` `loadLayerBridges` | Merges recorded and implied links through one `kinship.LinkSet`; each `graphalign.Bridge` carries `kinship.Recording` (its shapes' pivots are layer Subjects) and the `Associations` behind it |
| Person links | `core/database/canonlinks/reader.go` `(*Reader).Relationships` | A canonical person's relationship links, recorded (`Associations` = [the filed association]) then implied (`Associations` = the bridges behind it, `Recording.Shapes` with handle pivots), merged through `kinship.LinkSet`; the caller owns the result; not filtered |
| Kinship | `core/kinship/kinship.go` `Recording`, `CountedOnce`; `patterns.go` `eventPatterns` (unexported) | `Recording` has `Clone`, `Merge`; nothing outside `core/kinship` reads `OnlyImplied`. No exported list of pattern codes |
| Pins | `core/database/promotealign/pins.go` `BundlePinID`, `PinOffers`, `ResolvePins` | `BundlePinID` = `b:` + neighbor + `:` + key + `:` + `b.Bridge`. `ResolvePins` builds each default-pin row's offers through `evidence` |
| Handler | `api/ffi/handlers/promote_evidence.go` `GetPromoteClaimEvidence`, `claimEvidenceProto(e, values)` | Maps each relationship; bundles by `BundlePinID`. `claimEvidenceProto` takes no catalog |
| Headers | `api/ffi/handlers/promote_align.go` `proposalProto`; `api/ffi/handlers/conclusion_headers.go` `personHeaderProto`, `eventHeaderProto`; `core/database/conclusionheaders` `PersonsByIDs`, `EventsByIDs` | Promote's alternatives carry a handle's header through `oneof header { PersonHeader person; EventHeader event; … }`. Two hand-written copies read headers by kind and re-key them by id: `proposalProto` (raw-byte keys) and `search_catalog.go` `fillHitHeaders` (UUID-string keys). `conclusionheaders.RowsForSearch` reads the three kinds in one call but returns slices |
| Proto | `api/proto/engine.proto` `PromoteClaimEvidenceRelationship` | Fields 1–3, 7, 12–17 used; 4–6, 8–11 reserved |
| Swift | `macos/App/Platform/GenealogyStore.swift` `CatalogClaimEvidenceRelationship`; `GoStore.swift` `getPromoteClaimEvidence`, `mapPersonHeader`, `mapEventHeader`; `ConclusionVocabulary.swift` | No shapes; no kinship pattern enum |
| Vocabulary test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | Rows read a Swift enum with `wireCase` (`case "…": self = `); some rows derive `want` from a Go list (`sortedRelations()`, `vocabkey.PrimaryKinds()`); a subtest lists the enums that need `unrecognized` |
| Handler fixtures | `api/ffi/handlers/conclusion_headers_test.go` `citedEvent`; `promote_evidence_test.go` `newEvidenceFixture` | `citedEvent` promotes one event with a birth type and a date and returns its project dir and handle; `newEvidenceFixture` promotes a person handle (`handle`). The handlers package has no marriage or relationship helpers |
| Fixtures | `core/database/promotealign/promotealign_test.go` `fixture.participation`, `relationship` (return nothing; `writes.Call` returns a `connect.Result` whose `Subject` is the bridge), `marriage`, `eventOn`, `promote`, `join`, `newSource` | Enough to build every AC once `participation` and `relationship` return their bridge Subject |
| Tests | `evidence_test.go` `TestEvidenceReadsASourceMarriageAsSpouses`, `marriageEvidence`; `pins_test.go` `TestPinOffersOfAnImpliedLayerLink` (calls `marriageEvidence`); `core/claimevidence/build_test.go` `TestBuild` (AC-2.1 reads `Members[0].Bridge`); `api/ffi/handlers/promote_evidence_test.go` `TestClaimEvidenceNeighborReading` (calls `claimEvidenceProto(e, values)` in `withProjectCatalog`) | "AC-3.1" there asserts 6 incoming Observations; "AC-3.3" asserts **no** relationship for a tree marriage (this PR flips it) |
| Skill | `.agents/skills/add-match-rule/SKILL.md` kinship bullet; "Add a kinship pattern" | Says nothing of the evidence query reading `Relationships`, nor of Swift's pattern enum |
| Budget | `core/database/promotealign/perf_test.go` `budgetEvidence` (100 ms) | Must stay green |

**Precedent.** S11-25 ([plan](S11-25-claim-evidence-rpc.md)) and S11-20's neighbor reading added relationship fields the loader fills after `Build` (`readFromNeighbor`), with a handler test that also checks reserved names. `proposalProto` and `fillHitHeaders` read handle headers in one batch per kind; D-17 extracts that read before the pivots use it. S11-39's `marriageEvidence` is the fixture shape for every new evidence test here.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `evidence` reads `rels, err := q.kin.Relationships(row.HandleID)` once, after the handle node check. Per wanted neighbor it calls `all, err := q.accepted(hid)` first, then `q.canonLinks(handleNode, hn, rels, rowSet, nbrSet)`; `rowSet` and `nbrSet` are `map[string]bool` of Subject ids (`rowSet` built once from `members`, `nbrSet` from `all`). `canonLinks` returns `from.Links` with `Neighbor == to.ID` and `BridgeTypeKey != vocabkey.BridgeRelationship`, each read as today (one `CanonLink` per link, zero `Recording`), then one `CanonLink` per `kinship.Link` of `rels` whose other end is `to.ID`, keyed from `from`'s end (`KeyA` when `A` is `from.ID`, else `KeyB`), with that link's `Recording` and its bridges from one `bridgesOf` call over all its `Associations` | S11-51 D-4 and **Next adds**: participation links stay in `canonLinks`, relationship links come from `Relationships`, so recorded and implied shapes of one fact are one `CanonLink` | `Relationships` per neighbor — the same answer for every neighbor of a row; reading `Reader.For` and merging here — a second merge beside `kinship.LinkSet` |
| D-2 | One bridge reader, `(*evidenceReader).bridgesOf(assocs [][]byte, row, nbr, pivot map[string]bool) ([]claimevidence.Bridge, error)` *(new)*, replaces `bridges`. For each association, in order: `r.accepted(assoc)`; each bridge Subject's Observations from its member graph (each graph read once, as today); its spec is `connectrules.LookupBridge` of the Subject's `EvidenceSubject.TypeKey` on that graph (none, or not two endpoints: dropped); its ends are its positive Observations on the spec's two endpoint properties that carry a `ValueSubjectID`. An endpoint with two different Subjects drops the bridge. Each end Subject in `row` sets `Member`, else in `nbr` sets `Neighbor`, else in `pivot` sets `Pivot`; a bridge with fewer than two of the three set is dropped. Filed links call it with `[][]byte{l.Association}` and a nil `pivot` | One reader for recorded and implied bridges, reading ends through the bridge's own spec like every other reader (`canonlinks.EdgeOf`, `loadLayerBridges`); ends by membership, so a participation, a relationship and a chain's parent link read alike, with no `FromEnd` swap | Keeping `bridges(l)` beside a second reader for implied associations — two readers of bridge ends; reading ends from any registered endpoint property — misreads a future bridge type (review M-6) |
| D-3 | `claimevidence.Bridge` gains `Pivot []byte`. `canonLinks` builds `pivot` for one `kinship.Link` as the accepted members (`r.accepted`) of each `Shape.Pivot` handle in its `Recording.Shapes`; a recorded-only link has an empty set | A recorded bridge never has a pivot end, so `Pivot != nil` is exactly "behind an implied link" and `Build` needs no other flag | Marking implied associations on `kinship.Link` — S11-51 owns that type |
| D-4 | `Bundle.Bridge []byte` becomes `Bundle.Bridges [][]byte`. A bridge with `Pivot == nil` gives one bundle as today (`Bridges` of length 1; same `named`, same Observations). Bridges with a `Pivot` are grouped by `Pivot` across the corresponding canon links. A group gives one bundle when one of its bridges has a row-member `Member`: `Member` is the first such bridge's member, `Source` its Source, `Neighbor` the first group bridge's `Neighbor` that is a neighbor member (nil when none; a set `Neighbor` is `named` like a recorded bundle's, a nil one names nothing), `Bridges` the member's bridge first, then the group's other bridges in input order; `Observations` = each of those bridges' Observations in that order, then the neighbor member's defining refs (when `Neighbor` is set), then the pivot's defining refs (D-8), each id once (`appendNew`) | A pivot Subject lives in one Evidence graph, so one group is one Source's recording of the fact, and its backfill lands on the member whose Source recorded it ([kinship shapes: "every bridge behind the implied link"](../kinship-shapes.md#on-the-evidence-sheet)) | One bundle per underlying bridge — half of a participation pair has no row member to backfill onto; one bundle across every Source — backfills one Source's Observations onto another Source's member's claim |
| D-5 | Bundle order: bundles are appended in the order of their first bridge as `bundles` walks the corresponding canon links and their bridges in input order (a recorded bridge at its own position, a group at its first bridge's position), then stable-sorted by member index, then neighbor index, a nil `Neighbor` after every neighbor member | AC-1.3's recorded bundle precedes its marriage bundle because `Relationships` lists the recorded association first | Sorting groups after recorded bundles explicitly — the walk order already gives it |
| D-6 | A group with no row-member bridge is dropped (AC-1.6) | `promote.PinSet` needs one accepted `Member`; backfilling onto a neighbor's member would change §9's backfill rule | Pinning the far part on the neighbor member's claim — a product change |
| D-7 | The pivot set (D-3) holds only `Shape.Pivot` handles' members. A chain step that is itself a birth's implied parent link brings participations whose event end is no pivot, so those bridges have one classified end and are dropped (D-2); the chain's bundle holds its recorded parent links only | Keeps the pivot rule one line; the birth carries its participations on its own line | Adding every event reached by a step's associations to the pivot set — a second walk for a rare case |
| D-8 | `claimevidence.Pivot` *(new)* `{Subject []byte; Kind vocabkey.SubjectType; Values subjectcompare.Values}`. `LayerLink` gains `Recording kinship.Recording` and `Pivots []Pivot`: each `Recording.Shapes` pivot (a layer Subject) with `valuesOf[pivot]` and its kind from `in.primary`. `CanonLink` gains `Recording kinship.Recording` and `Pivots []Pivot`: each distinct `Bridge.Pivot` of its bridges, values through `r.memberValues`, kind the pivot Subject's `EvidenceSubject.TypeKey`. `Build` appends `refsOf(p.Values, rules.Defining[p.Kind])`: the layer pivots' after the neighbor's defining refs in `Incoming`, a group's pivot's into its bundle (D-4) | "What a bundle holds" stays the evidence section's `Defining` rule in the pure package, as the neighbor's does | Reading the pivot's `event_type` Observation ids in the loader — moves the rule out of `claimevidence` |
| D-9 | `Relationship` gains `Layer, Canon kinship.Shape` (the zero `Shape` is direct). `core/kinship` gains `func (r Recording) Shown() Shape` *(new)*: the zero `Shape` when `!r.OnlyImplied` or `r.Shapes` is empty, else `r.Shapes[0]`. `Build` sets `Layer` from the key's layer links' recordings merged (`Recording.Merge` into a `Clone` of the first) and `Canon` from the corresponding canon links' recordings merged, each through `Shown` | One shape per side is what S11-42 words; a side that records the fact directly reads as direct; `Shapes` keep first-added order, so the choice is stable. The direct-vs-implied rule has one home S13-04 can call (review M-2) | A list of every shape per side — no sentence words two; a private helper in `claimevidence` — S13-04 would copy it |
| D-10 | Count-once: `Inputs` gains `Placed map[string][]byte` (layer Subject → the handle the request places it on; the loader passes `evidence`'s `placed`). In `relationships`, a canon link corresponds when `segmentcompare.Corresponds(key, c.Key) && !kinship.CountedOnce(layerRec, c.Recording, paired)`, with `paired := func(p []byte) ([]byte, bool) { h, ok := in.Placed[string(p)]; return h, ok }` | Working rule "Counted once across sides": the evidence lines follow matching's `corresponding`, through the same function | A second count-once check in the loader — the rule would have two homes |
| D-11 | Wire: `message PromoteClaimEvidenceShape { string pattern = 1; string pivot_id = 2; oneof pivot_header { PersonHeader person = 3; EventHeader event = 4; } }` *(new)*; `PromoteClaimEvidenceRelationship` gains `PromoteClaimEvidenceShape layer = 18; PromoteClaimEvidenceShape canon = 19;`, always set. `pattern` is `kinship.Pattern` ("" when direct); `pivot_id` is the pivot as a UUID string ("" when direct). Only `canon` carries a header: the handler collects every canon pivot, splits them by `Shape.Pattern.PivotKind()`, reads them with one `conclusionheaders.ByIDs` call (D-17), and sets `person` / `event` through `personHeaderProto` / `eventHeaderProto`. A layer pivot is a Subject on the page; Swift finds it by id as it finds `neighbor_subject_id`. `claimEvidenceProto` becomes `claimEvidenceProto(c *database.Catalog, e claimevidence.Evidence, values promotealign.EvidenceValues) (*engine.GetPromoteClaimEvidenceResponse, error)` | The tree pivot is shown through the one handle-header mechanism Promote already uses (review M-1); the header already picks the shown name or date; Spike 13's "One person header for every tile" | Pivot values typed through `EvidenceValues` with a kind → property table — a second display rule beside the headers (M-1) |
| D-12 | `core/kinship` gains `func (p Pattern) PivotKind() vocabkey.SubjectType` *(new)* (`vocabkey.Event` for each `eventPatterns` code, `vocabkey.Person` for `PatternParentOfParent`, "" otherwise) and `func Patterns() []Pattern` *(new)* (each `eventPatterns` code in order, then `PatternParentOfParent`) | A pattern's pivot kind is a fact of the pattern table; the vocabulary test derives its row from Go (review M-4), as `sortedRelations()` does | A pattern → kind switch in the handler — a second table beside `eventPatterns` |
| D-13 | Swift: `enum KinshipPattern` *(new)* in `ConclusionVocabulary.swift` after `EvidencePropertyCause`, cases `direct`, `marriageSpouses`, `birthParents`, `parentOfParent`, `unrecognized(String)`, `init(wire:)` written exactly as `EvidenceOutcome`'s (`case "": self = .direct`, then one `case "<code>": self = .<case>` per pattern). `struct CatalogClaimEvidenceShape` *(new)* in `GenealogyStore.swift` `{pattern: KinshipPattern = .direct; pivotID = ""; pivotPerson: CatalogPersonHeader? = nil; pivotEvent: CatalogEventHeader? = nil}`; `CatalogClaimEvidenceRelationship` gains `layer` and `canon`, each defaulting to `CatalogClaimEvidenceShape()`. `GoStore` maps through `mapEvidenceShape` *(new, private static)* with `mapPersonHeader` / `mapEventHeader`. No Swift unit test: the enum is pinned by `TestSwiftVocabularyMatchesGo`, as every evidence enum is | Working rule "One wire vocabulary"; defaults keep `FakeStore` and existing Swift tests compiling | Parsing the pattern only in S11-42 — the map puts the vocabulary here |
| D-14 | `BundlePinID` keys on `b.Bridges[0]`; `titles` reads `b.Bridges[0]` | Bundles of one relationship never share a first bridge (a group's member bridge is in no other group; a recorded bundle's one bridge has no pivot); recorded bundles keep `main`'s ids | Joining every bridge id — longer, no extra uniqueness |
| D-15 | The evidence reader holds one `canonlinks.Reader`: `evidenceReader.kin` *(new)*, `canonlinks.NewReader(g, matchregistry.Default().Kinship)` in `newEvidenceReader` | `ResolvePins` reuses one reader across a page, as it reuses the profiler | A reader per `evidence` call — reloads the terms per row |
| D-16 | The test fixture's `participation` and `relationship` return the bridge `subjects.Subject` (`res.Subject` from `writes.Call`'s `connect.Result`); every current caller ignores it. `marriageEvidence` returns a `marriageScene` *(new)* with the evidence (`e`), the record's `ann` and `bob`, the tree's `bob`, the tree and record marriages (zero when absent), the tree's two participations (zero when absent), and the two handles; `TestPinOffersOfAnImpliedLayerLink` reads `.e`. `marriageEvidence` and every new test build a marriage with `eventOn(…, "marriage", d)` and two `participation` calls instead of `f.marriage`, so the participations are kept | ACs name bridge Subjects; the fixture writes them already | Looking bridge Subjects up by query in each test |
| D-17 | `core/database/conclusionheaders` gains `func ByIDs(c *database.Catalog, persons, events, places [][]byte) (Headers, error)` *(new)* and `type Headers` *(new)* with `Person(id []byte) (PersonHeader, bool)`, `Event(id []byte) (EventHeader, bool)`, `Place(id []byte) (PlaceHeader, bool)`, keyed by raw id bytes, built on `PersonsByIDs`, `EventsByIDs`, `PlacesByIDs` (a kind with no ids is not read). `proposalProto` and `fillHitHeaders` move onto it and their `personByID` / `eventByID` / `placeByID` maps are deleted; each caller keeps its own kind switch and `oneof` | A third copy of "read headers by kind, re-key by id" is extracted in the PR that would add it (review M-9); S13-04 and S13-05 become one call | Copying `proposalProto`'s block for pivots — a third copy |
| D-18 | T-7 hand-builds `claimevidence.Evidence` as `TestClaimEvidenceNeighborReading` does and runs `claimEvidenceProto(c, e, values)` inside `withProjectCatalog`: AC-3.1 and AC-3.3 in `citedEvent`'s project, `Canon` = `{marriage_spouses, the citedEvent handle}` (its header is dated; the event's type is not checked) and `Layer` = `{marriage_spouses, a fresh UUID}`; AC-3.2 in `newEvidenceFixture`'s project, `Canon` = `{parent_of_parent, f.handle}` | The handlers package has no marriage or chain helpers, and the header read does not depend on the pivot's role in the catalog; the catalog-built cases are T-4 | Adding marriage and chain fixtures to the handlers package — a copy of `promotealign_test.go`'s |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `claimevidence.Pivot` | type | `claimevidence.Member` (`Subject`, `Source`, `Values`) | A pivot has a kind and no Source, and is a member of neither the row's nor the neighbor's entity |
| `Bridge.Pivot`, `Bundle.Bridges`, `LayerLink.Recording/Pivots`, `CanonLink.Recording/Pivots`, `Relationship.Layer/Canon`, `Inputs.Placed` | fields | the same types | Extends the types the query already passes; `Layer` / `Canon` reuse `kinship.Shape` |
| `kinship.Recording.Shown`, `Pattern.PivotKind`, `kinship.Patterns` | methods, func | `Recording.Clone`/`Merge`; `eventPatterns` | Reads of what `core/kinship` owns, for S11-42, S13-04 and the vocabulary test |
| `bridgesOf` | func | `(*evidenceReader).bridges` | Replaces it (Deletes) |
| `evidenceReader.kin` | field | `canonlinks.NewReader` in `canon.go` `loadCanon` | The evidence query's own reader; sharing one through `inputs` is logged (M-7) |
| `PromoteClaimEvidenceShape` | proto message | `PromoteGraphAlignmentAlternative`'s `oneof header` | Same header shape, with a pattern and pivot id beside it |
| `evidenceShapeProto` | func | `evidenceTermProto` | Different message; headers come from `conclusionheaders.ByIDs` |
| `conclusionheaders.ByIDs`, `Headers` | func, type | `proposalProto` and `fillHitHeaders` blocks; `RowsForSearch` | Extracts the two copies (Deletes); `RowsForSearch` returns slices and reads through the graph for search |
| `KinshipPattern` | Swift enum | `EvidenceOutcome` | New wire vocabulary, same idiom |
| `CatalogClaimEvidenceShape`, `mapEvidenceShape` | Swift struct, func | `CatalogEvidenceTerm`, `mapEvidenceTerm`; `mapPersonHeader`, `mapEventHeader` | Different message; reuses the header mappers |
| `marriageScene` | test type | `marriageEvidence`'s three return values | Grows the one fixture instead of adding a second |
| `TestByIDs` | test | `conclusionheaders_test.go` `TestPersonsByIDs`, `TestEventsByIDs` | Pins the lookup |
| `TestBuildImpliedLinks`, `TestRecordingShown`, `TestPatterns`, `TestEvidenceImpliedCanonLinks`, `TestEvidenceCountsASharedMarriageOnce`, `TestResolvePinsImpliedCanonLink`, `TestClaimEvidenceShapes` | tests | `TestBuild`, `TestRecordingMerge`, `TestEvidenceReadsASourceMarriageAsSpouses`, `TestResolvePinsBundlesARelationship`, `TestClaimEvidenceNeighborReading` | New behavior; each reuses the existing fixtures |

Searched: `rg "Bridge\b" core/claimevidence core/database/promotealign api/ffi/handlers`, `rg "Relationships\(" core api`, `rg "OnlyImplied" core api`, `rg "marriage_spouses|parent_of_parent" macos api`, `rg "conclusionheaders\.\w+ByIDs" api`, `rg "func \(r \*evidenceReader\)" core/database/promotealign`.

**Builds on.** `canonlinks.Reader.Relationships` (S11-51); `kinship.Recording` / `CountedOnce` (S11-39); the evidence section's `Defining` rule; `conclusionheaders` headers as `proposalProto` sends them; the proto reserved-names test; the `wireCase` vocabulary test.
**Deletes.** `(*evidenceReader).bridges`; the `personByID` / `eventByID` / `placeByID` maps in `proposalProto` and `fillHitHeaders`; `Bundle.Bridge`; the canon side's own reading of relationship links in `canonLinks`; `TestEvidenceReadsASourceMarriageAsSpouses` subtest "AC-3.3" (no relationship for a tree marriage), replaced by T-4.
**Next adds.** S11-42 words `Layer` / `Canon` from `KinshipPattern`, the canon header, and the layer pivot's row on the page. A new kinship pattern adds its `eventPatterns` row (which `Patterns` and `PivotKind` then list) and one `KinshipPattern` case, which `TestSwiftVocabularyMatchesGo` demands. S13-04 reads `Recording.Shown` for "recorded directly or implied", and S13-04 / S13-05 read headers through `conclusionheaders.ByIDs`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/kinship` | `kinship.go`: `func (r Recording) Shown() Shape` *(new)*. `patterns.go`: `func (p Pattern) PivotKind() vocabkey.SubjectType` *(new)*, `func Patterns() []Pattern` *(new)* |
| `core/claimevidence` (`claimevidence.go`) | `type Pivot struct { Subject []byte; Kind vocabkey.SubjectType; Values subjectcompare.Values }` *(new)*. `Bundle.Bridges [][]byte` replaces `Bridge` (doc: the member's bridge first). `Relationship` gains `Layer, Canon kinship.Shape` (doc: how the row's Source and the entity record the fact; the zero Shape is direct; a layer pivot is a layer Subject, a canon pivot a handle). `LayerLink` gains `Recording kinship.Recording; Pivots []Pivot`. `Bridge` gains `Pivot []byte`. `CanonLink` gains `Recording kinship.Recording; Pivots []Pivot`. `Inputs` gains `Placed map[string][]byte` |
| `core/claimevidence` (`build.go`) | `relationships`: per key, merge the layer links' recordings and pivots (each pivot Subject once); canon links correspond per D-10; `Incoming` = observations, the neighbor's defining refs, then each layer pivot's defining refs; `rel.Layer` / `rel.Canon` per D-9. `bundles`: D-4, D-5, D-6, the pivots looked up from the corresponding canon links' `Pivots` |
| `core/database/promotealign` (`evidence.go`) | `evidenceReader.kin`; `evidence` per D-1 and D-10 (`Inputs.Placed: placed`); each layer link gets `Recording` (from the `graphalign.Bridge`) and `Pivots` (D-8); `canonLinks(from, to *graphcache.Node, rels []kinship.Link, row, nbr map[string]bool) ([]claimevidence.CanonLink, error)` per D-1, D-3, D-8; `bridgesOf` *(new)* replaces `bridges`; `titles` reads `b.Bridges[0]` |
| `core/database/promotealign` (`pins.go`) | `BundlePinID` reads `b.Bridges[0]` (D-14) |
| `core/database/conclusionheaders` (`compose.go`) | `ByIDs`, `Headers` *(new)* (D-17) |

`export_test.go` needs no change: `BridgedOnly` keeps its signature.

Errors: none new; every new read returns the graph cache's or `conclusionheaders`' error.

### FFI

`api/proto/engine.proto`: `message PromoteClaimEvidenceShape` *(new)* before `PromoteClaimEvidenceRelationship`, commented "How one side records a relationship's fact (S11-40): the kinship pattern ("" when recorded directly) and the event or parent between it passes through (a layer Subject on the row's side, a handle on the entity's). Only the entity's side carries the pivot's header."; `PromoteClaimEvidenceRelationship` gains `layer = 18`, `canon = 19`. Regenerate with `./scripts/generate-proto.sh`.

Handler `api/ffi/handlers/promote_evidence.go`: `claimEvidenceProto` takes the catalog and returns an error (D-11); `GetPromoteClaimEvidence` passes `c`. `evidenceShapeProto(s kinship.Shape) *engine.PromoteClaimEvidenceShape` *(new)* sets `Pattern` and `PivotId` (`database.UUIDString(s.Pivot)` when `len(s.Pivot) == 16`, else ""); `claimEvidenceProto` sets `Layer` and `Canon` on every relationship, then reads the canon pivots' headers through one `conclusionheaders.ByIDs` call and sets `Canon.PivotHeader`. `proposalProto` (`promote_align.go`) and `fillHitHeaders` (`search_catalog.go`) read through `ByIDs` (D-17). `TestClaimEvidenceNeighborReading` passes `c` and checks the error.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Wire enum | `KinshipPattern` *(new)* | `macos/App/Platform/ConclusionVocabulary.swift` | D-13 |
| Model | `CatalogClaimEvidenceShape` *(new)*; `CatalogClaimEvidenceRelationship.layer`, `.canon` | `macos/App/Platform/GenealogyStore.swift` | D-13 |
| Mapping | `mapEvidenceShape` *(new)* | `macos/App/Platform/GoStore.swift` | `pivotPerson: s.hasPerson ? mapPersonHeader(s.person) : nil`, `pivotEvent` likewise |

No view or state changes. `FakeStore` unchanged.

### Strings

None — no user-facing copy (S11-42).

## 7. Steps

**STEP-1 — Kinship reads and the bundle rename**
Files: `core/kinship/kinship.go`, `patterns.go`, `kinship_test.go`; `core/claimevidence/claimevidence.go`, `build.go`, `build_test.go`; `core/database/promotealign/pins.go`, `evidence.go` (`titles` only). Skill: `add-match-rule`.
Do: D-9's `Shown`, D-12; T-1. Rename `Bundle.Bridge` → `Bridges` and fix every reader in the same step (`build.go` sets `Bridges: [][]byte{b.Subject}`; `BundlePinID` and `titles` read `Bridges[0]`, D-14); T-3.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./core/... ./api/...` and `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/... ./core/claimevidence/... ./core/matchregistry/... ./core/database/promotealign/...` pass.

**STEP-2 — Shapes, implied bundles, pivot defining refs, count-once in `Build`**
Files: `core/claimevidence/claimevidence.go`, `build.go`, `build_test.go`.
Do: the new types and fields (§6), D-4 – D-6, D-8 (`Build` side), D-9 (`Layer` / `Canon`), D-10. T-2.
Check: `go test ./core/claimevidence/... ./core/matchregistry/...` passes (no build tags; the guard passes).

**STEP-3 — One bridge reader, recorded links only**
Files: `core/database/promotealign/evidence.go`, `promotealign_test.go` (D-16 fixture returns).
Do: D-2 (`bridgesOf` with a nil `pivot`), D-15, and D-1's `rowSet` / `nbrSet` and `accepted`-first order. `canonLinks` takes D-1's signature and is called with a nil `rels`; in this step it still reads every `from.Links` entry, relationship links included, through `bridgesOf`. Delete `bridges`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes with no test changed beyond D-16's fixture returns; `rg -n "func \(r \*evidenceReader\) bridges\(" core/` prints nothing.

**STEP-4 — Relationship links from `Relationships`, with their recordings**
Files: `core/database/promotealign/evidence.go`.
Do: D-1 in full: `evidence` reads `rels`; `canonLinks` skips `from.Links` relationship links and reads `rels` with each link's `Recording`; each layer link gets its `Recording` from the `graphalign.Bridge`. No pivot set yet (`bridgesOf` gets a nil `pivot`, so implied bridges drop).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes with no test changed; AC-1.10 holds (`TestEvidenceReadsEachLinkFromTheNeighbor`, `TestResolvePinsBundlesARelationship` unchanged).

**STEP-5 — Pivots: implied bundles on both sides, and their pins**
Files: `core/database/promotealign/evidence.go`, `evidence_test.go`, `pins_test.go`.
Do: D-3 (the pivot set in `canonLinks`), D-8 loader side (`LayerLink.Pivots`, `CanonLink.Pivots`). `marriageScene` (D-16); rename `TestEvidenceReadsASourceMarriageAsSpouses`'s subtests "AC-3.1" → "AC-1.2" and "AC-3.3" → "AC-1.1", and flip "AC-1.1" from no relationship to AC-1.1's line; `TestPinOffersOfAnImpliedLayerLink` reads `.e`. T-4, T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes.

**STEP-6 — Count-once**
Files: `core/database/promotealign/evidence.go`, `evidence_test.go`.
Do: D-10 loader side (`Inputs.Placed: placed`). T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes, `TestLargeSourceBudgets` included.

**STEP-7 — One header read by id**
Files: `core/database/conclusionheaders/compose.go`, `conclusionheaders_test.go`, `api/ffi/handlers/promote_align.go`, `search_catalog.go`.
Do: D-17. T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/... ./api/ffi/handlers/...` passes with no handler test changed.

**STEP-8 — Wire** (Mac)
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/promote_evidence.go`, `promote_evidence_test.go`. Skill: `add-ffi-handler`.
Do: §6 FFI, D-11. T-7 (D-18).
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-9 — Swift model and vocabulary** (Mac)
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `GenealogyStore.swift`, `GoStore.swift`, `api/ffi/swift_vocabulary_test.go`.
Do: D-13. T-8: a "kinship patterns" row (`enum: "KinshipPattern"`, `re: wireCase`, `want` = `""` plus each `kinship.Patterns()` code, sorted; import `core/kinship`), and `"KinshipPattern"` in the "tolerate a value this build doesn't know" list.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `scripts/check-pr.sh` passes (Xcode tests, no Swift warning).

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-11/claim-evidence.md` (§ Evidence through relationships: a paragraph "Implied links (S11-40)" stating D-4 – D-6, D-9 and the count-once rule), `kinship-shapes.md` (§ On the evidence sheet: one bundle per Source recording, the member's bridge first and keyed by it; open question 2 answered: yes, S11-40 and S11-42), `.agents/skills/add-match-rule/SKILL.md` (layer-5 bullet: `kinship.CountedOnce` also skips a pair in claim evidence; kinship bullet: claim evidence's canon side reads `Reader.Relationships`, and `Recording.Shown` is the one direct-vs-implied reading; "Add a kinship pattern": add the Swift `KinshipPattern` case, which `TestSwiftVocabularyMatchesGo` demands, and S11-42's wording), `docs/deployment-plan/spike-13/deployment-plan.md` (working rule "One person header for every tile": S13-04 and S13-05 read through `conclusionheaders.ByIDs` instead of re-keying), `completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-40-implied-links-evidence.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a tree marriage as a spouse line on the evidence sheet."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/kinship_test.go` `TestRecordingShown` *(new)*, `TestPatterns` *(new)* | Zero `Recording` → zero `Shape`; `{Shapes: [s1, s2], OnlyImplied: false}` → zero; `{Shapes: [s1, s2], OnlyImplied: true}` → `s1`; `{OnlyImplied: true}` → zero. `Patterns()` = [`marriage_spouses`, `birth_parents`, `parent_of_parent`]; `PivotKind` → `event`, `event`, `person`; an unknown code → "" | D-9, D-12 |
| T-2 | Go | `core/claimevidence/build_test.go` `TestBuildImpliedLinks` *(new)* | (a) layer link `Recording{Shapes: [{marriage_spouses, P}], OnlyImplied: true}`, canon recorded → `Layer == {marriage_spouses, P}`, `Canon` zero. (b) canon link with bridges `{Subject 0x70, Member m, Pivot e}`, `{Subject 0x71, Neighbor n, Pivot e}`, `Pivots [{e, Event, event_type ref 0x90}]` → one bundle `Member m`, `Neighbor n`, `Bridges [0x70 0x71]`, Observations = 0x70's, 0x71's, n's defining refs, 0x90; the neighbor's section lines pin n. (c) only the 0x71 bridge → no bundle; only 0x70 → one bundle, `Neighbor` nil, no section member from n, sorted after a recorded bundle of the same member. (d) both sides only implied through `marriage_spouses`, `Placed[layer pivot] = canon pivot` → no relationship; `Placed` empty → one. (e) both sides recorded, pivots placed → one relationship. (f) layer `Pivots [{p, Event, event_type ref 0x91}]` → `Incoming` ends with 0x91 | AC-1.1 – AC-1.3, AC-1.6 – AC-1.9, D-4 – D-10 |
| T-3 | Go | `core/claimevidence/build_test.go` `TestBuild` AC-2.1, AC-2.5 | Unchanged inputs → `Members[i].Bridges == [[bridge]]` | AC-1.10, D-4, D-14 |
| T-4 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceReadsASourceMarriageAsSpouses` (subtests "AC-1.2", "AC-1.1"); `TestEvidenceImpliedCanonLinks` *(new)* | `marriageEvidence(t, true)` → AC-1.2. `marriageEvidence(t, false)` → AC-1.1 (`Bridges` = the tree Bob's then the tree Ann's participation). New subtests "AC-1.3 both shapes", "AC-1.4 birth parents", "AC-1.5 chain", "AC-1.6 chain across Sources", "AC-1.7 the neighbor is named", "AC-1.10 recorded both sides" (`Bridges` length 1, `BundlePinID` of the `b:<neighbor>:<key>:<relationship Subject>` form, the sections as on `main`) | AC-1.1 – AC-1.7, AC-1.10 |
| T-5 | Go | `evidence_test.go` `TestEvidenceCountsASharedMarriageOnce` *(new)* | AC-1.8 (no `relationship·spouse`; a line whose key type is `participation` with event type `marriage`); AC-1.9 (one spouse line with both shapes) | AC-1.8, AC-1.9, D-10 |
| T-6 | Go | `core/database/promotealign/pins_test.go` `TestResolvePinsImpliedCanonLink` *(new)* | AC-1.1's catalog: `ResolvePins` with Bob's row `DefaultPins` → AC-2.1's set; `PinOffers` id → AC-2.2; that id in Bob's `PinIDs` → the same set | AC-2.1, AC-2.2 |
| T-7 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestClaimEvidenceShapes` *(new)* | Per D-18: AC-3.1 (`canon.event` is the `citedEvent` header, its date set; `canon.person` unset), AC-3.2 (`canon.person` is `f.handle`'s header), AC-3.3 (`layer.pivot_id` the UUID, no header); every relationship has `layer` and `canon` set | AC-3.1 – AC-3.3, D-11, D-18 |
| T-9 | Go | `core/database/conclusionheaders/conclusionheaders_test.go` `TestByIDs` *(new)* | One promoted person and one promoted event: `ByIDs(c, [person], [event], nil)` → `Person(person)` and `Event(event)` found, `Person(event)` and `Place(person)` not; an id of no entity → not found | D-17 |
| T-8 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "kinship patterns" *(new row)* and the `unrecognized` subtest | Swift `KinshipPattern` cases equal `kinship.Patterns()` plus "" | AC-3.4, D-12, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestEvidenceBridgeEndpoints`, `TestEvidenceLoadsOnlyBridgedNeighborMembers`, `TestEvidenceReadsEachLinkFromTheNeighbor`, `TestEvidenceSpouseNeverMeetsSibling`, `TestResolvePinsBundlesARelationship`, `TestPinOffersOfAnImpliedLayerLink` (they pin `bridgesOf` against `bridges`' behavior), and `TestLargeSourceBudgets` `budgetEvidence` (100 ms).

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
- [ ] Domain doc and skill updated: claim-evidence.md; kinship-shapes.md (§ On the evidence sheet, open question 2); add-match-rule SKILL.md
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project. On Source "Tree", add Ann Ames and Bob Ames and a marriage dated 12 May 1842 with both as `subject`; promote all three. On Source "Register", add Ann Ames and Bob Ames in a `spouse` relationship.
2. Open Promote for "Register". Bob's row targets the tree Bob, Ann's the tree Ann.
3. Open Bob's evidence sheet: a relationship Card for Ann is shown, worded as today ("…they share a spouse …"), checked (AC-3.5). On `main` this Card is absent.
4. Press Done without unchecking. `identity_claim_evidence` for Bob's new claim holds both tree marriage participations' Observations.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `bridgesOf`'s membership reading differs from `bridges`' endpoint reading for a filed link | STEP-3 swaps the reader alone and must pass every existing test unchanged |
| `Relationships` per request pushes the evidence budget | One call per request (D-1, D-15), and `Reader.For` is memoized; `budgetEvidence` must pass |
| A pivot Subject is also a member of the row's or the neighbor's entity | Classification order in D-2 is row, then neighbor, then pivot; a pivot handle is an event or the parent between, never the row's or neighbor's handle |

**Stop and report if:**

- `canonlinks.Reader.Relationships` no longer returns recorded links with `Associations` = [the filed association] and implied links with handle pivots in `Recording.Shapes`.
- STEP-3 needs any existing test's expected values changed.
- `TestLargeSourceBudgets` fails `budgetEvidence` with this change and passes without it.
- A filed link's bridge Subject has no `EvidenceSubject` on its member graph, or its `TypeKey` is not its bridge type.
- `conclusionheaders.PersonsByIDs` or `EventsByIDs` refuses an id of the other kind instead of leaving it out.
- `claimevidence` cannot import `core/kinship` without `TestLayerPackages` failing.
- S13-07 has landed and changed `evidence` or `canonLinks`: rebase, then report the rows of §4 that moved.

## 12. Open questions

None.

## 13. PR

**Title:** S11-40 — Implied links in claim evidence
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-40-implied-links-evidence.md`.
