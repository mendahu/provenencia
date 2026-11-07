# S13-26 — Scoring two Evidence graphs

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 7 — How a person's records connect
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a pure Go layer package; nothing reads it until S13-27, and nothing draws it until S13-28)
**Board:** —
**Depends on:** S13-24 (landed [#639](https://github.com/mendahu/provenencia/pull/639))
**Blocks:** S13-27
**Branch:** `s13-26-scoring-two-evidence-graphs`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, and this PR adds no catalog format step ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (Rules that hold everywhere; Add a knob; Add a layer package; Check)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [`../evidence-graph-comparison.md`](../evidence-graph-comparison.md) (Requirement 1, First version); working rules "Scoring and the view are two concerns", "Slice 7 starts with a simple count", "Each PR updates what it makes false" (slice 7); [conclusion §5.1](../../../conclusion-layer-data-model.md#51-two-lenses); [`matching-model.md`](../../../matching-model.md) (The stack, Knob inventory)
**Planned:** 2026-11-07 against `main` @ `aaf3b15`
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

The Person page will draw a person's Evidence graphs around a circle with one arc for each other person two graphs share, thin or thick by how much the two records agree about that person (bar 10). This PR builds the count behind each arc: a pure layer package, `core/recordlinks`, that takes two Evidence graphs and returns one line per shared person with the number of facts the two records share about them and a weak or strong weight. It is the first version the user decided on 2026-11-07, built simple so it can be dogfooded before anything richer is designed.

**Why now.** It is slice 7's first PR: S13-27 reads it for the Person page, and S13-28 draws it. It needs only S13-24 (landed), which moved `Rules.Role` and `Rules.Meta` into `subjectcompare` as this comparison's starting point.

**What stays.** Promote and graph alignment are untouched: `CountedOnce` keeps its behavior (STEP-1 builds it on a new helper). Subject compare's weights, points and lifespan are not read. Nothing reaches the catalog, the wire or the Mac app.

## 2. User stories

**US-1.** As the engine, given two Evidence graphs of one person, I return one line for each other person both graphs hold, so the Person page can draw an arc per shared person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Graph A holds person members on entities P (the page), F and M; graph B holds members on P, F and S | One line, for F. P is the page's own person and is never a line; M and S are each in one graph only |
| AC-1.2 | B holds two members filed on F (F1, F2); A's F shares 1 fact with F1 and 2 with F2 | One line for F, with 2 facts |
| AC-1.3 | A's and B's members on F share no fact (names "John Smith" / "Mary Smith", nothing else) | One line for F, 0 facts, weak (user, 2026-11-07: sharing the person is itself the line) |
| AC-1.4 | A member whose entity is nil (not filed, or not accepted) | Never a line, and never the far end of a counted link |

**US-2.** As the engine, I count each fact the two records share about a shared person once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | F's names "John Smith" / "John Smith" | 1 fact (name agrees) |
| AC-2.2 | F's names "Amy" / "Amy Lee Anne Long" | 1 fact (name partial) |
| AC-2.3 | F's names "John Smith" / "Mary Smith" | 0 facts (the name compare says conflict: two first names left over) |
| AC-2.4 | A: gender `man`; B: sex at birth `male` | 1 fact (gender pools with sex at birth) |
| AC-2.5 | A: gender `man`, sex at birth `male`; B: the same two | 1 fact, not 2 |
| AC-2.6 | A: sex at birth `male`; B: sex at birth `female` | 0 facts |
| AC-2.7 | A: religion `catholic`; B: religion `catholic`, `anglican` (multiple) | 1 fact |
| AC-2.8 | Both members have a lifespan and lifespan span that agree | 0 facts (derived; not counted) |
| AC-2.9 | Both members are subjects of a birth event; the two events are filed on one entity E | 1 fact |
| AC-2.10 | As AC-2.9, but A's role is `subject` and B's is `witness` | 0 facts |
| AC-2.11 | As AC-2.9, but the events are filed on two entities, or B's event is filed on none | 0 facts |
| AC-2.12 | A: F is recorded as the parent of the page's person P; B: the same | 1 fact (a link to the page's own person counts) |
| AC-2.13 | A: F and P are spouses (recorded relationship); B: F and P are both subjects of a marriage filed on entity M, so the spouse link is only implied | 1 fact (the spouse links correspond; B's marriage has no partner in A) |
| AC-2.14 | A and B: F and P are both subjects of a marriage, both filed on entity M; the spouse link is implied on both sides | 1 fact (the marriage participation); the implied spouse link is not counted again |
| AC-2.15 | As AC-2.14, and both A and B also record the spouse relationship directly | 1 fact, not 2 |
| AC-2.16 | A and B: F is the father in a birth of P, both births filed on entity E (an implied parent link on each side) | 1 fact for F's line |
| AC-2.17 | Same name, and both subjects of births filed on one entity | 2 facts, strong |

**US-3.** As the engine, I give each line a weight from the registry.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Shipped registry (`recordlinks.StrongAt` = 2); a line with 0, 1, 2, 3 facts | weak, weak, strong, strong |
| AC-3.2 | A test copy with `StrongAt` = 3; a line with 2 facts | weak |

## 3. Scope

| In | Covered by |
| --- | --- |
| `kinship.SameShape` *(new)*, with `CountedOnce` built on it | STEP-1, AC-2.14, AC-2.15 |
| `core/recordlinks` *(new)*: `Rules`, `Config`, `Member`, `Graph`, `NewGraph`, `Line`, `Strength`, `Lines` | STEP-2 – STEP-5, US-1 – US-3 |
| The registry's `RecordLinks` section, `defaultRecordLinks`, `Registry.RecordLinksConfig`, and its inventory row | STEP-3, AC-3.1, AC-3.2 |
| The layer package guard row, the `match-rules.mdc` glob, the `add-match-rule` skill | STEP-2, STEP-7 |
| An exact table test and a fixed-seed invariants test | STEP-4 – STEP-6 |
| Docs: `matching-model.md` (The stack), conclusion §5.1 and §13 item 4 | STEP-7 |

| Out | Home |
| --- | --- |
| Loading graphs, members, values and links from the catalog; the RPC | S13-27 (with the shared loader, S13-29) |
| The circular link diagram | S13-28 (brief S13-D10) |
| Listing what a line counted, collecting conflicts, conflicts or Source quality changing a weight | [`../evidence-graph-comparison.md`](../evidence-graph-comparison.md) Open (later version) |
| Lines for shared events or places | [`../evidence-graph-comparison.md`](../evidence-graph-comparison.md) Open |
| Catalog-defined Properties (non-product origin): the loader reads product Observations only | not planned (the loader's scope, `promotealign.layerValues`) |
| Any change to Promote's ranking with this count | [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Property compare | `core/propertycompare/propertycompare.go` `Compare` | Gives one outcome (agree, partial, conflict, unknown) for all values two Subjects hold for one property, under its meta. A multiple-cardinality difference is unknown, not conflict. Reads relations, never numbers. This PR calls it as is |
| Name relations | `core/valuecompare/name.go` `name.compare` | Same words → `Same`; paired words with leftovers that are not both first names or both family names → a partial relation; no pair, or both sides leaving a first name or a family name → `Different` (`CauseNoMatch`). So "Amy" / "Amy Lee Anne Long" is partial and "John Smith" / "Mary Smith" is conflict |
| Subject rules | `core/subjectcompare/subjectcompare.go` `Rules.Role`, `Rules.Meta`, `Rules.Derived` | `Role` gives a property's lead and how it counts (lead, pooled, period end); `Meta` the key its values compare under; `Derived` names the lifespan lead and its span (no catalog row). Moved here by S13-24 as slice 7's starting point |
| Subject side | `core/subjectcompare/pairs.go` `Side` | The values `Compare` reads for one lead: its own, its pool's (gender mapped into sex at birth through `Pool.Terms`), or a period built from start and end when the lead has no positive value |
| Subject compare | `core/subjectcompare/subjectcompare.go` `Compare` | Compares only the weighted leads (`Weights[kind]`: name, sex at birth, lifespan for a person) and turns outcomes into points. Not used here: this count is over every property and has no points |
| Person registry | `core/matchregistry/subject.go` `defaultSubject` | Person weights name, sex at birth (pools gender), lifespan; `Lifespans[person][lifespan]` with `Span: lifespan_span`; `Periods` only for events |
| Correspondence | `core/segmentcompare/key.go` `Corresponds` | "The one correspondence rule": two `BridgeKey`s mean the same link read from the same end (today, equality) |
| Links | `core/kinship/links.go` `Link`, `LinkSet`, `LinkSet.WithImplied` | One side's links, recorded and implied, merged once per ends and keys; each carries a `Recording` (its shapes: pattern and pivot; `OnlyImplied`) |
| Shapes | `core/kinship/kinship.go` `Recording`, `Shape`, `CountedOnce` | `CountedOnce(a, b, paired)` is true when at least one side is only implied and a shape of `a` and a shape of `b` share a pattern with paired pivots. Two links both recorded are never counted once |
| CountedOnce caller | `core/graphalign/align.go` `state.corresponding` | Graph alignment's one caller; must not change behavior |
| Edge reading | `core/kinship/kinship.go` `Edge.From` | Reads an edge from one end: the far end and the key from this end |
| Registry | `core/matchregistry/registry.go` `Registry`, `Default`, `AlignConfig` | Sections `Value`, `Subject`, `Segment`, `Alignment`, `Kinship`; `AlignConfig` bundles the sections graph alignment reads |
| Layer guard | `core/matchregistry/guard_test.go` `layerPackages`, `TestLayerPackages` | Each layer package is pure (no `core/database`, no build tags, no non-test `core/matchregistry` import) and, when `numbers`, holds no numeric literal but 0 and 1 |
| Inventory guard | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry`, `registryKeys` | Every registry field needs a row in `docs/matching-model.md`'s knob inventory; a top-level field's name is lowercased whole (`Kinship` → `kinship.MaxPivotEnd`) |
| Test helpers | `core/propertycompare/propertycomparetest` (`Name`, `Term`, `Neg`), `core/subjectcompare/subjectcomparetest` `Metas`, `core/segmentcompare/segmentcomparetest` (`Terms`, `Bridge`) | Values, metas and bridge keys for layer tests. `Metas` has no `religion` (tests add it to their copy) |
| Link test builders | `core/kinship/kinship_test.go` `part`, `rel`, `link`, `TestDeriveInvariants` | How a layer test builds participation and relationship edges with `segmentcompare.Keys`, and a fixed-seed invariants test |
| Tests | `core/kinship/kinship_test.go` `TestCountedOnce` | Seven cases of `CountedOnce`; stay green unchanged |
| Docs | `docs/conclusion-layer-data-model.md` §5.1, §13 item 4 | Say the entity comparison's rules are open and that it "must count conflicts … and show its reasons as sentences" |

**Precedent.** `a5c0e0ff` ("Derive the spouse, parent and grandparent links …", S11-39) added the `core/kinship` layer package: the package with a package doc, an external `_test` package with an exact table and a fixed-seed invariants test, its `layerPackages` row, its `match-rules.mdc` glob and its `add-match-rule` lines. Copy that shape. Do not copy its `vocabkey` additions: this PR adds no vocabulary key.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A new layer package `core/recordlinks`, `numbers` true in `layerPackages`. Its package doc names it the record-links comparison, beside the matching stack: it judges two members the researcher already joined to one entity, and it is not a layer under graph alignment | Working rule "Scoring and the view are two concerns": pure and registry-driven like matching. A separate package keeps Subject compare's points and Promote's walk out of it | Inside `core/subjectcompare` — that layer turns outcomes into points for Promote; this count has no points and reads links. Inside `core/graphalign` — the working rule forbids reading its scoring |
| D-2 | Inputs are pure types. `Member{ID []byte; Entity []byte; Kind vocabkey.SubjectType; Values subjectcompare.Values}` and `NewGraph(members []Member, links []kinship.Link) Graph`. `Entity` is the canonical entity of the member's accepted Identity Claim, nil when it has none; the caller sets it. `links` is the graph's `kinship.LinkSet.Links()` after `WithImplied` | The shared loader (S13-29) yields values from `subjectprofile.ForSubject` and links as `kinship` edges with recordings; `graphalign.Bridge` is `kinship.Edge` + `kinship.Recording`, which is `kinship.Link`, so S13-27 converts without loss | Taking `graphalign.Layer` — imports graph alignment's types into a package the working rule keeps apart from it. Taking `graphcache` types — a layer may not reach `core/database` |
| D-3 | `NewGraph` indexes once: members by entity, the entity of each member id, and each member's links read from its end (`Edge.From`: far end, key, `Recording`). `Lines` reads only the index | 50 graphs give 1,225 pairs; each graph is indexed once by the caller, not once per pair | Indexing inside `Lines` — 49 rebuilds per graph at Scale |
| D-4 | `Lines(page []byte, a, b Graph, metas subjectcompare.Metas, cfg Config) []Line`. A line is made for each entity that has a person member (`Kind == vocabkey.Person`) in both graphs, other than `page` and other than nil. `Line{Entity []byte; Facts int; Strength Strength}`. Lines are sorted by `Entity` bytes ascending | AC-1.1, AC-1.4; working rule "A line per shared person". A fixed order makes the output comparable and order-free (T-8) | Lines for events or places — Out (user, 2026-11-07) |
| D-5 | When a graph holds more than one member on the line's entity, every pair (one member from each graph) is counted and the line takes the largest count | AC-1.2. Two members of one graph on one entity are the researcher's call that both are this person; the better-corroborated pair speaks for the line, as the best pair speaks in property compare | Summing pairs — counts one fact twice; first pair by id — the result would depend on ids |
| D-6 | Properties. For each key in the union of both members' `Values`, in ascending key order: skip it when `cfg.Subject.Derived(kind, key)`; skip it when `Role` gives `RolePeriodEnd` (its values count through the lead's period, as `Side` reads them); otherwise its unit is `cfg.Subject.Meta(kind, key)`, counted once per unit. A unit with no entry in `metas` is skipped. The unit's values are `subjectcompare.Side(kind, unit, values, cfg.Subject)` when `Role` found a lead, else the key's own values. `propertycompare.Compare(sideA, sideB, metas[unit], cfg.Value)`; agree or partial adds 1 | Working rule "A matching Property counts": every product property both record, pooled ones once with their lead, lifespan out. `Meta` and `Side` are the existing rules for how a property counts; nothing new decides it | `subjectcompare.Compare` — reads only the weighted leads and includes the lifespan. A new loop over raw keys — gender and sex at birth would count twice (AC-2.5) |
| D-7 | Links. For each link of the member in A (far end `x`, key `k`, recording `ra`) whose far end has an entity, and each link of the member in B (far end `y`, key `k'`, recording `rb`) whose far end has the same entity: it is a match when `segmentcompare.Corresponds(k, k')` and not `kinship.SameShape(ra, rb, same)`, where `same(pa, pb)` is true when both pivots have an entity and the two are equal. Each distinct pair (key `k`, far entity) adds 1 | Working rule "A matching link counts": same `BridgeKey`, far ends on one canonical entity; a relationship to anyone (the page's person included, AC-2.12) or a participation in the same event in the same role (AC-2.9, AC-2.10). `Corresponds` is the one correspondence rule | Matching links on the far entity alone — counts a witness and a groom at one wedding as a shared fact (AC-2.10). Keying a map on `BridgeKey` — bypasses `Corresponds` |
| D-8 | `kinship.SameShape(a, b Recording, same func(pivotA, pivotB []byte) bool) bool` *(new)*: true when a shape of `a` and a shape of `b` have the same `Pattern` and `same` holds for their pivots. `CountedOnce` becomes: false when neither side is only implied; else `SameShape(a, b, …)` with `same` true when `paired(pivotA)` is `pivotB`. A pair of links that `SameShape` matches stands for one fact already counted through its pivot, so it is skipped even when both sides also record it directly (AC-2.15) | User, 2026-11-07: a marriage and the spouse link it implies are one fact. The pivot's own fact always counts: the event's participations, or the parent link to the middle parent of a chain, are links of the same members (AC-2.14, AC-2.16). One shape rule serves both callers | Calling `CountedOnce` — returns false when both sides record the link (AC-2.15 would count 2). A copy of the shape loop in `recordlinks` — a second mechanism beside `kinship`'s |
| D-9 | `Facts` is the property count plus the link count. `Strength` is `StrengthStrong` *(new)* `"strong"` when `Facts >= cfg.Links.StrongAt`, else `StrengthWeak` *(new)* `"weak"`, 0 facts included | Working rule "Two weights"; user, 2026-11-07: a shared person with no matching fact is still a thin line (AC-1.3) | Dropping 0-fact lines — refused by the user. Reusing `graphalign.Assessment` — Promote's band, which the working rule "The score ranks, unseen" keeps out of new work |
| D-10 | Registry. `recordlinks.Rules{StrongAt int}` with a doc comment; `Registry.RecordLinks recordlinks.Rules`; `defaultRecordLinks()` in `core/matchregistry/recordlinks.go` *(new)* sets `StrongAt: 2` with the comment `// decided: user, 2026-11-07 (Spike 13 slice 7): two or more shared facts make a strong line`. `recordlinks.Config{Value valuecompare.Rules; Subject subjectcompare.Rules; Links Rules}` and `func (r Registry) RecordLinksConfig() recordlinks.Config`, the way `AlignConfig` bundles graph alignment's sections. Inventory row key: `recordlinks.StrongAt` | Numbers live only in the registry; the layer takes its sections as arguments (`add-match-rule`). The caller (S13-27) fills `Value` from `valuerules.Load` as Promote does | A constant in the layer — the numbers guard forbids it |
| D-11 | Tests live in an external package `recordlinks_test` and build links the way `core/kinship/kinship_test.go` does: edges from `segmentcompare.Keys` with `segmentcomparetest.Bridge` and `segmentcomparetest.Terms`, merged through `kinship.NewLinkSet`, `Add`, `WithImplied(terms, matchregistry.Default().Kinship)`, then `Links()`. Values from `propertycomparetest`; metas from `subjectcomparetest.Metas()` plus `"religion"` (term, multiple) on the test's copy | Implied links in the tests are the ones the loader produces, not hand-written recordings | Hand-built `Recording`s — they could disagree with `Derive` |

**Rules this plan bends.** None. The conclusion doc's §5.1 sentence that the comparison must count conflicts and show its reasons as sentences is a design aim, not a working rule; STEP-7 rewrites it to say the first version does neither, by the user's decision of 2026-11-07, and keeps the aim for a later version.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/recordlinks` *(new)* | package | `core/subjectcompare`, `core/graphalign` | D-1 |
| `recordlinks.Member`, `Graph`, `NewGraph` *(new)* | type / func | `graphalign.Subject`, `graphalign.Layer` | D-2: graph alignment's types belong to the walk the working rule keeps apart; `Member` adds `Entity` |
| `recordlinks.Lines`, `Line` *(new)* | func / type | `subjectcompare.Compare`, `Result` | D-6: Compare weighs leads and the lifespan; Lines counts every property and links |
| `recordlinks.Strength` *(new)* | type | `graphalign.Assessment` | D-9 |
| `recordlinks.Rules`, `Config` *(new)* | type | `graphalign.Rules`, `graphalign.Config` | Same shape, own section (D-10) |
| `kinship.SameShape` *(new)* | func | `kinship.CountedOnce` | `CountedOnce` is rebuilt on it (D-8); one shape loop |
| `Registry.RecordLinks`, `RecordLinksConfig`, `defaultRecordLinks` *(new)* | field / func | `Registry.Kinship`, `AlignConfig`, `defaultKinship` | The registry's pattern for a section |

**Builds on.** `propertycompare.Compare`, `subjectcompare.Rules.Role` / `Meta` / `Derived`, `subjectcompare.Side`, `segmentcompare.Corresponds`, `kinship.Link` / `Recording` / `Edge.From`, and the registry's section pattern (`add-match-rule` "Add a knob", "Add a layer package").
**Deletes.** None — nothing becomes dead. `CountedOnce`'s body moves onto `SameShape`; its signature and behavior stay.
**Next adds.** S13-27 calls `NewGraph` once per graph and `Lines` per pair. A later version adds an explanation by returning the counted keys beside `Facts`, and a conflict rule as a `Rules` field.

### Reach

None — no new path into existing code, no new state reachable. The package has no caller until S13-27.

### Scale

At the map's Scale (slice 7: 50 Evidence graphs, 1,225 pairs; a household's few people per pair, an obituary's few dozen at most): `NewGraph` runs once per graph (D-3). Per pair, `Lines` walks A's person entities with a map lookup into B (linear in A's members), and per shared person compares each property once (values × values, as Subject compare does) and each of A's member's links against B's member's links on the same far entity (an index by far entity, so linear in the two members' links). There is no write and no cache. The read's timing at Scale is S13-27's budget subtest, which owns the call.

### Edges

- **Assumptions.** (1) Symmetry: `Lines(page, a, b)` and `Lines(page, b, a)` give the same entities, facts and strengths. Property compare and the name compare are side-free (the name pairing uses the side with the smaller key); `Corresponds` is equality; `SameShape` checks both orders through `same`. T-8 holds it. (2) Order-free: shuffling members and links does not change the result (T-8). (3) `Edge.From` reads a self-loop from A's end only; the loader does not produce self-loops, and a self-loop would count at most once per key and entity.
- **Neighbors.** A name that shares a word but leaves two first names (AC-2.3) counts nothing; one that leaves a middle name (AC-2.2) counts. A multiple-cardinality property with one shared and one extra value is partial-or-agree (AC-2.7), one with no shared value is unknown and counts nothing. A neutral term (sex `unknown`) is dropped by property compare and counts nothing. A negative value that matches the other side's positive is a conflict and counts nothing. A participation in the same event under another role counts nothing (AC-2.10). A spouse recorded directly on one side and only implied on the other counts once (AC-2.13).
- **Removed inputs.** A member with no entity is neither a line nor a far end (AC-1.4, AC-2.11): a record the researcher has not filed adds nothing. An unaccepted claim is the caller's: `Entity` is nil for it (D-2). A skipped implied link loses nothing: its pivot's own link is counted (D-8; T-8 checks that skipping by shape never removes a line's only link fact when its pivot link is present).
- **Real data.** The rule reads the shipped subject section: for a person, `Role` finds name and sex at birth as leads, gender pooled under sex at birth, lifespan and lifespan span derived, and no period ends. T-1 runs on `matchregistry.Default().Subject`, not a fixture.
- **Untrusted size.** Values per property, links per member and members per entity come from the Source's Evidence graph. Property compare is values × values, already bounded the way Subject compare is (and names by `tooManyWords`). Links are indexed by far entity (D-7), so a member with many links is linear. Members per entity (D-5) multiply pairs: two graphs each holding k members on one entity give k² member pairs; the loader keeps one member per Subject, and a graph holding several Subjects on one entity is a researcher's explicit filing, bounded by the graph's Subjects (an obituary's 41).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/kinship` | `func SameShape(a, b Recording, same func(pivotA, pivotB []byte) bool) bool` *(new)*; `CountedOnce` calls it (D-8) |
| `core/recordlinks` *(new)* | `recordlinks.go`: package doc (D-1); `type Rules struct{ StrongAt int }`, `type Config struct{ Value valuecompare.Rules; Subject subjectcompare.Rules; Links Rules }`, `type Member struct{…}`, `type Graph struct{ /* unexported index */ }`, `func NewGraph(members []Member, links []kinship.Link) Graph`, `type Strength string` with `StrengthWeak`, `StrengthStrong`, `type Line struct{ Entity []byte; Facts int; Strength Strength }`, `func Lines(page []byte, a, b Graph, metas subjectcompare.Metas, cfg Config) []Line`; unexported `properties` (D-6) and `links` (D-7) |
| `core/matchregistry` | `Registry.RecordLinks recordlinks.Rules`; `Default` sets it from `defaultRecordLinks()`; `recordlinks.go` *(new)*; `func (r Registry) RecordLinksConfig() recordlinks.Config` |

```go
// Lines is one line for each person both graphs hold on one canonical entity,
// other than the page's own: the facts the two records share about that
// person and the line's weight. Lines are sorted by entity.
func Lines(page []byte, a, b Graph, metas subjectcompare.Metas, cfg Config) []Line
```

Errors: none — the function is total and returns nil for no shared person.

### FFI

None — no wire change; S13-27 adds the RPC.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — One shape rule in kinship**
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`. Skill: `add-match-rule`.
Do: add `SameShape` (D-8) beside `CountedOnce` with a doc comment; rewrite `CountedOnce`'s loop as a call to it. Add `TestSameShape` (T-2).
Check: `go test ./core/kinship/... ./core/graphalign/...` passes; `TestCountedOnce` is unchanged and green.

**STEP-2 — The package and its guard**
Files: `core/recordlinks/recordlinks.go` *(new)*, `core/matchregistry/guard_test.go`, `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule` "Add a layer package".
Do: the package doc and every exported type of §6 Go core; `NewGraph` builds the index (D-3); `Lines` returns nil for now. Add `{"core/recordlinks", true}` to `layerPackages` after `core/kinship`; add `core/recordlinks/**` to the glob.
Check: `go test ./core/matchregistry/...` passes (`TestLayerPackages` includes `core/recordlinks/pure` and `/numbers`).

**STEP-3 — The registry section**
Files: `core/matchregistry/registry.go`, `core/matchregistry/recordlinks.go` *(new)*, `docs/matching-model.md`. Skill: `add-match-rule` "Add a knob".
Do: D-10. Add the inventory row after the `kinship.MaxPairSteps` row: `| \`recordlinks.StrongAt\` | 2 | user, 2026-11-07 (Spike 13 slice 7) | Fewest facts two records share about a person for their line to be strong; fewer is weak |`.
Check: `go test ./core/matchregistry/...` passes (`TestKnobInventoryMatchesRegistry`, `TestDefaultIsFresh`).

**STEP-4 — Count properties**
Files: `core/recordlinks/recordlinks.go` *(new)* (from STEP-2), `core/recordlinks/recordlinks_test.go` *(new)*. Skill: `add-match-rule`.
Do: `properties` per D-6; `Lines` assembles lines (D-4, D-5, D-9) with property facts only. Add `TestLines` (T-1) with the rows for AC-1.1 – AC-1.4, AC-2.1 – AC-2.8, AC-3.1 and AC-3.2, built per D-11.
Check: `go test ./core/recordlinks/...` passes.

**STEP-5 — Count links**
Files: `core/recordlinks/recordlinks.go` *(new)* (from STEP-2), `core/recordlinks/recordlinks_test.go` *(new)* (from STEP-4).
Do: `links` per D-7, added into `Facts`. Add the `TestLines` rows for AC-2.9 – AC-2.17.
Check: `go test ./core/recordlinks/...` passes.

**STEP-6 — Invariants**
Files: `core/recordlinks/invariants_test.go` *(new)*.
Do: `TestLinesInvariants` (T-8), fixed seed `rand.NewSource(1)`, 200 rounds, on the model of `TestDeriveInvariants`.
Check: `go test ./core/recordlinks/...` passes; the add-match-rule Check command passes with `./core/recordlinks/...` added.

**STEP-7 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md`, `docs/matching-model.md`, `docs/conclusion-layer-data-model.md`, `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), this plan (status).
Do: (a) `add-match-rule`: `core/recordlinks` in the description's package list; a "Rules that hold everywhere" bullet: "**Record links (`core/recordlinks`) count what two records share about a person.** `NewGraph` indexes one Evidence graph's members (each with the entity of its accepted claim) and links; `Lines` gives one line per person both graphs hold, other than the page's own, counting each property that agrees or partly agrees (`Meta` and `Side` decide the unit; derived properties and period ends do not count on their own) and each corresponding link to the same entity, skipping a pair `kinship.SameShape` matches. Its one knob is `recordlinks.StrongAt`. It reads no weights, points or band."; `./core/recordlinks/...` in the Check command. (b) `matching-model.md` The stack: a paragraph after the kinship paragraph: "Beside the stack, `core/recordlinks` counts what two Evidence graphs of one person share about each other person they hold (Spike 13 slice 7): agreeing properties and corresponding links to the same canonical entity, weak or strong by `recordlinks.StrongAt`. It reads layers 1 to 3's rules and kinship's shapes, never graph alignment's scoring." (c) conclusion §5.1: replace the sentences from "How the score is built is open" through "show its reasons as sentences." with: "The first version (2026-11-07, Spike 13's design, linked as the paragraph's existing design link is) counts, for each other person two graphs share, the properties and links the two records agree on, and weighs the line weak or strong; it ignores conflicts and shows no reasons. Counting conflicts, weighing a nickname alone as weak, and showing reasons as sentences stay the aim for a later version." Keep the following sentences. §13 item 4: "Requirements are captured; how it is built and expressed is open there." becomes "A first version that counts shared facts is built (`core/recordlinks`, S13-26); conflicts, reasons and Source quality are open there." (d) `completed.md`: an S13-26 paragraph at the top. (e) Map row Plan and Status; this plan's Status and Landed note.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-26-scoring-two-evidence-graphs.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/recordlinks/recordlinks_test.go` `TestLines` *(new)* | Table, one row per AC-1.1 – AC-1.4, AC-2.1 – AC-2.17 and AC-3.1 – AC-3.2, each the two graphs and the page entity → the exact `[]Line` (entities, facts, strengths) | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, AC-2.11, AC-2.12, AC-2.13, AC-2.14, AC-2.15, AC-2.16, AC-2.17, AC-3.1, AC-3.2, D-4 – D-10 |
| T-2 | Go | `core/kinship/kinship_test.go` `TestSameShape` *(new)* | Same pattern, `same` true → true; same pattern, `same` false → false; different patterns → false; empty shapes → false; both recorded (neither only implied) with paired pivots → true | D-8 |
| T-3 | Go | `core/kinship/kinship_test.go` `TestCountedOnce` (unchanged) | Its seven cases → as today | D-8 |
| T-4 | Go | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` (unchanged) | `recordlinks.StrongAt` field and row → pass | D-10 |
| T-5 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` (new row) | `core/recordlinks` → pure, no numbers | D-1 |
| T-6 | Go | `core/matchregistry/registry_test.go` `TestDefaultIsFresh` (unchanged) | Still passes with the new section | D-10 |
| T-7 | Go | `core/graphalign/...` (unchanged) | Promote's alignment tests stay green after STEP-1 | D-8 |
| T-8 | Go | `core/recordlinks/invariants_test.go` `TestLinesInvariants` *(new)* | 200 random pairs of graphs (2 – 6 people, 1 – 3 events, random names from a small pool, random roles and relationship terms, random entities with some nil): (a) `Lines(p, a, b)` equals `Lines(p, b, a)`; (b) shuffled members and links give the same lines; (c) no line for `p` or nil; a line for exactly each person entity in both graphs; (d) `Strength` is strong exactly when `Facts >= StrongAt`; (e) a graph compared with itself gives each shared person at least one fact when that person has a name | D-4 – D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/kinship` and `core/graphalign` (STEP-1), `core/matchregistry` guards (STEP-2, STEP-3).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: the TestLines row for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: no name was renamed or deleted; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory — no UI
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated — not gated
- [ ] Domain doc updated: conclusion §5.1 and §13 item 4, matching-model.md The stack and knob inventory, add-match-rule (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app: the package has no caller until S13-27, whose manual check and bar 10 exercise it. The reviewer reads `TestLines`'s rows against §2.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Rebuilding `CountedOnce` on `SameShape` changes graph alignment | T-3 and T-7 run unchanged; `CountedOnce`'s signature stays |
| The name compare gives a partial for a pair the user would not call a match | AC-2.2 and AC-2.3 pin the boundary; the relation stays the name compare's, so a later change to it reaches both Promote and this count |
| The symmetry invariant fails on a tie in property compare | T-8 (a) surfaces it; property compare picks its speaking pair by outcome, similarity, relation and Same key, never by side |

**Stop and report if:**

- `subjectcompare.Rules.Meta`, `Role`, `Derived` or `Side` no longer exist with the signatures in §4.
- `kinship.Link` no longer carries `Recording`, or `LinkSet.WithImplied` no longer takes `(segmentcompare.Terms, kinship.Rules)`.
- `TestCountedOnce` or any `core/graphalign` test fails after STEP-1 and the fix would change `CountedOnce`'s behavior.
- `TestLayerPackages` fails because `core/recordlinks` needs a numeric literal other than 0 and 1, or an import of `core/matchregistry` or `core/database` outside tests.
- T-8 (a) fails for a case where `propertycompare.Compare` itself is not symmetric: report the values; do not add a tiebreak in `recordlinks`.
- An AC's expected count cannot be met without counting a lifespan, a conflict, or a link to a far end with no entity.

## 12. Open questions

None.

## 13. PR

**Title:** S13-26 — Scoring two Evidence graphs
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (`TestLines` rows); deviations from the plan; manual check (none: no caller until S13-27); `Plan: docs/deployment-plan/spike-13/prs/S13-26-scoring-two-evidence-graphs.md`.
