# S13-26 — Scoring two Evidence graphs

**Status:** Landed in #736 (2026-11-07)
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
**Review:** [`S13-26-scoring-two-evidence-graphs.review.md`](S13-26-scoring-two-evidence-graphs.review.md) — round 1: Fits with changes / Ready with changes; round 2: Fits with changes / Ready with changes, every finding Fixed; two To user items answered 2026-11-07

**Landed note.** One difference: `TestLines` AC-2.2 builds "Amy" with `pt.Given("Amy")`, not `pt.Name("Amy")`, because `Name` reads a one-word form as a surname and the name compare then conflicts (the graphalign test builds it the same way). Review round 1 changed these, by Decision (the Decisions and Edges above are kept as planned):

- **D-5 and Edges "Untrusted size":** one `Lines` call is now bounded. `recordlinks.Rules` gains `MaxEntityMembers` = 10 and `MaxUnitValues` = 20 (user, 2026-11-07), each with an inventory row. A person on an entity with more members of any kind in either graph keeps its line at 0 facts; links to such an entity count nothing; a property where either member holds more values (pooled ones added) counts nothing. Inside them, `recordlinks.Rules.MaxPairWork` = 25,000 (user, 2026-11-07, SEC-1 restated) caps the work of the whole pair, counted before any compare as, per member pair, the words of one side times the other's for each property that counts plus the shapes of one side's links times the other's for each entity both link to (a link with no shape counts one); a pair over it gives every shared person a line at 0 facts. A 41-Subject obituary does about 1,000, and `BenchmarkLinesAtCap` times the slowest input under the cap on each axis (names, long text, shared people, shapes, the longest values). `recordlinks.Rules.MaxValueBytes` = 256 (user, 2026-11-07, SEC-3) limits one value (the bytes of its text, term, and a name's form and parts): a property where either member holds a longer value counts nothing and charges no work, so a crafted value cannot hang a read; a value's words are the name's and the text's added.
- **D-6:** `Rules.Unit` is `Rules.CountsUnder`, `CompareProperty`'s `unit` parameter and `Side`'s `lead` are `key`, and `propertycompare.Outcome.Matches` is the one reading of agree or partial for `recordlinks`, `graphalign` and `subjectcompare`.
- **D-7 and D-8:** a pair of links is skipped through a pivot only when the pivot's own fact counted: the pivot is filed on one entity E in both graphs, E is within the member bound, and the two members' links to E hold a corresponding pair. This supersedes the skip rule in D-7 and corrects D-8's Why ("the pivot's own fact always counts"): father on one record and mother on the other no longer lose the parent link.
- **D-11 and D-12:** edge fixtures for participations and relationships are `kinshiptest.Part` and `kinshiptest.Rel` (a new package, a `layerPackages` row), used by `core/kinship`, `core/recordlinks` and `core/graphalign` tests; `part`, `rel` and `link` in `core/kinship` tests no longer take `t`, and `graphalign`'s `partKeys` and `relKeys` are gone.
- **Pinned:** a gender no pooled term maps counts nothing (decided, as for Promote), the first member of an id wins, and `FuzzLines` holds symmetry over random graphs.

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

**What stays.** Promote and graph alignment behave as today: `subjectcompare.Compare` and `kinship.CountedOnce` are rebuilt on shared helpers (STEP-1, STEP-2) with no change in output. Subject compare's weights, points and lifespan are not read by the new package. Nothing reaches the catalog, the wire or the Mac app.

## 2. User stories

**US-1.** As the engine, given two Evidence graphs of one person, I return one line for each other person both graphs hold, so the Person page can draw an arc per shared person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Graph A holds person members on entities P (the page), F and M; graph B holds members on P, F and S | One line, for F. P is the page's own person and is never a line; M and S are each in one graph only |
| AC-1.2 | B holds two person members filed on F (F1, F2); A's F shares 1 fact with F1 and 2 with F2 | One line for F, with 2 facts |
| AC-1.3 | A's and B's members on F share no fact (names "John Smith" / "Mary Smith", nothing else) | One line for F, 0 facts, weak (user, 2026-11-07: sharing the person is itself the line) |
| AC-1.4 | A member whose entity is nil (not filed, or not accepted) | Never a line, and never the far end of a counted link |
| AC-1.5 | Entity X holds a person member in A and an event member in B | No line for X (a line needs a person member in both graphs) |

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
| AC-2.18 | As AC-2.14, but A's marriage is filed on no entity | 1 fact: the spouse links correspond and their pivots are not on one entity, so the spouse link counts; the unfiled participation does not |
| AC-2.19 | A holds two marriages of F and P, M1 (filed on entity M) and M2 (filed on none), each implying the spouse link, so A's one spouse link carries two shapes; B holds one marriage filed on M | 1 fact: the spouse link is skipped because one shape pair is on one entity (D-7), and the M participation counts |

**US-3.** As the engine, I give each line a weight from the registry.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Shipped registry (`recordlinks.StrongAt` = 2); a line with 0, 1, 2, 3 facts | weak, weak, strong, strong |
| AC-3.2 | A test copy of the config with `RecordLinks.StrongAt` = 3; a line with 2 facts | weak |

## 3. Scope

| In | Covered by |
| --- | --- |
| `subjectcompare.Rules.Unit` and `subjectcompare.CompareProperty` *(new)*, with `Compare` built on `CompareProperty` | STEP-2, AC-2.4, AC-2.5, AC-2.8 |
| `kinship.SameShape` *(new)*, with `CountedOnce` built on it | STEP-1, AC-2.14, AC-2.15 |
| `segmentcomparetest.PartKeys` and `RelKeys` *(new)*, with no `t`; `core/kinship/kinship_test.go` `part` / `rel` and `core/graphalign/align_test.go` `partKeys` / `relKeys` move onto them | STEP-1, T-10 |
| `religion` in `subjectcomparetest.Metas` | STEP-2, AC-2.7 |
| `core/recordlinks` *(new)*: `Rules`, `Config`, `Member`, `Graph`, `NewGraph`, `Line`, `Strength`, `Lines` | STEP-3, STEP-5 – STEP-7, US-1 – US-3 |
| The registry's `RecordLinks` section, `defaultRecordLinks`, `Registry.RecordLinksConfig`, and its inventory row | STEP-4, AC-3.1, AC-3.2 |
| The layer package guard row, the `match-rules.mdc` glob, the package lists in `add-match-rule`, `.agents/skills/review-pr/lenses/security.md` and `.agents/skills/ship-pr/SKILL.md` | STEP-3, STEP-8 |
| An exact table test and a fixed-seed invariants test | STEP-5 – STEP-7 |
| Docs: `matching-model.md` (The stack), conclusion §5.1 and §13 item 4, the map's working rule "A matching link counts" and S13-26 entry | STEP-8 |

| Out | Home |
| --- | --- |
| Loading graphs, members, values and links from the catalog; the RPC | S13-27 (with the shared loader, S13-29) |
| One link type from the loader (`graphalign.Bridge` folded into `kinship.Link`) | S13-29 (user, 2026-11-07; review M-5) |
| The circular link diagram | S13-28 (brief S13-D10) |
| Listing what a line counted, collecting conflicts, conflicts or Source quality changing a weight | [`../evidence-graph-comparison.md`](../evidence-graph-comparison.md) Open (later version) |
| Lines for shared events or places | [`../evidence-graph-comparison.md`](../evidence-graph-comparison.md) Open |
| Catalog-defined Properties (non-product origin): the loader reads product Observations only | not planned (the loader's scope, `promotealign.layerValues`) |
| Moving the other hand-built key pairs in tests (`core/kinship/links_test.go`, `siblings_test.go`, `core/database/canonlinks/reader_test.go`) onto the new helpers | not planned: they build one-off keys inside a test; `part` / `rel` and `partKeys` / `relKeys` are the reusable builders |
| Any change to Promote's ranking with this count | [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Property compare | `core/propertycompare/propertycompare.go` `Compare` | Gives one outcome (`OutcomeAgree`, `OutcomePartial`, `OutcomeConflict`, `OutcomeUnknown`) for all values two Subjects hold for one property, under its meta. A multiple-cardinality difference is unknown, not conflict. A neutral value is dropped |
| Value compare | `core/valuecompare/valuecompare.go` `Compare` | A value type with no comparer, or two different types, relates as `Unknown` (`CauseUnsupported`), so a property compared under a missing (zero) meta is unknown |
| Name relations | `core/valuecompare/name.go` `name.compare` | Same words → `Same`; paired words with leftovers that are not both first names or both family names → a partial relation; no pair, or both sides leaving a first name or a family name → `Different` (`CauseNoMatch`). So "Amy" / "Amy Lee Anne Long" is partial and "John Smith" / "Mary Smith" is conflict |
| Subject rules | `core/subjectcompare/subjectcompare.go` `Rules.Role`, `Rules.Meta`, `Rules.Derived` | `Role` gives a property's lead and how it counts (lead, pooled, period end); `Meta` the key its values compare under; `Derived` names the lifespan lead and its span (no catalog row). Moved here by S13-24 as slice 7's starting point |
| Subject side | `core/subjectcompare/pairs.go` `Side` | The values `Compare` reads for one lead: its own, its pool's (gender mapped into sex at birth through `Pool.Terms`), or a period built from start and end when the lead has no positive value. For a key with no pool and no period it returns a copy of the key's own values |
| Subject compare | `core/subjectcompare/subjectcompare.go` `Compare` | For each weighted lead: the lifespan rule for a `Lifespans` lead, else `propertycompare.Compare(Side(…a…), Side(…b…), metas[key], value)` (line 117), then points. The only place that composes `Side` with property compare. Compares only the weighted leads (name, sex at birth, lifespan for a person) |
| Do-not rule | `.agents/skills/add-match-rule/SKILL.md` "Do not" | "Write a points switch or a lead lookup outside `core/subjectcompare`" |
| Person registry | `core/matchregistry/subject.go` `defaultSubject` | Person weights name, sex at birth (pools gender), lifespan; `Lifespans[person][lifespan]` with `Span: lifespan_span`; `Periods` only for events |
| Correspondence | `core/segmentcompare/key.go` `Corresponds` | "The one correspondence rule": two `BridgeKey`s mean the same link read from the same end (today, equality) |
| Links | `core/kinship/links.go` `Link`, `LinkSet`, `LinkSet.WithImplied` | One side's links, recorded and implied, merged once per ends and keys; each carries a `Recording` (its shapes: pattern and pivot; `OnlyImplied`). A link implied by two events gets two shapes |
| Shapes | `core/kinship/kinship.go` `Recording`, `Shape`, `CountedOnce` | `CountedOnce(a, b, paired)` is true when at least one side is only implied and a shape of `a` and a shape of `b` share a pattern with paired pivots. Two links both recorded are never counted once. A shape's pivot is the event (marriage, birth) or the middle parent (chain) |
| CountedOnce caller | `core/graphalign/align.go` `state.corresponding` | Graph alignment's one caller; must not change behavior |
| Edge reading | `core/kinship/kinship.go` `Edge.From` | Reads an edge from one end: the far end and the key from this end |
| Registry | `core/matchregistry/registry.go` `Registry`, `Default`, `AlignConfig` | Sections `Value`, `Subject`, `Segment`, `Alignment`, `Kinship`; `AlignConfig` bundles the sections graph alignment reads |
| Layer guard | `core/matchregistry/guard_test.go` `layerPackages`, `TestLayerPackages` | Each layer package is pure (no `core/database`, no build tags, no non-test `core/matchregistry` import) and, when `numbers`, holds no numeric literal but 0 and 1 |
| Inventory guard | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry`, `registryKeys` | Every registry field needs a row in `docs/matching-model.md`'s knob inventory; a top-level field's name is lowercased whole (`Kinship` → `kinship.MaxPivotEnd`) |
| Test metas | `core/subjectcompare/subjectcomparetest/subjectcomparetest.go` `Metas` | The metas the matching tests use, pinned to the seed by `core/database/subjectvocab/subject_section_test.go`; holds `"interests"` but no `religion` (seeded term, multiple, `core/database/subjectvocab/registry.go`) |
| Test keys | `core/segmentcompare/segmentcomparetest` `Terms`, `Bridge`; `core/kinship/kinship_test.go` `part`, `rel`, `link`; `core/graphalign/align_test.go` `partKeys`, `relKeys`, `partLink`, `relLink` | Two private copies of "a participation's or relationship's key pair": `kinship_test`'s take `t` and go through `segmentcompare.Keys(segmentcomparetest.Bridge(t, …))`; `align_test`'s take no `t` and hand-build `segmentcompare.Key(segmentcompare.Link{Type, Term, End, Other, EventType, SameType}, Terms())`, and are reached through `partLink` / `relLink` from about 25 call sites |
| Test values | `core/propertycompare/propertycomparetest` `Name`, `Term`, `Neg`, `Neutral` | Values for layer tests |
| Invariants model | `core/kinship/kinship_test.go` `TestDeriveInvariants` | A fixed-seed random test (`rand.NewSource(1)`, 200 rounds) of one edge set |
| Tests | `core/kinship/kinship_test.go` `TestCountedOnce`; `core/subjectcompare/subjectcompare_test.go` | Stay green unchanged |
| Package lists | `.agents/skills/review-pr/lenses/security.md` "Stored numbers and lists" row; `.agents/skills/ship-pr/SKILL.md` line 106 | Name every layer package, `core/kinship` included |
| Map | `docs/deployment-plan/spike-13/deployment-plan.md` working rule "A matching link counts"; S13-26 scope entry | Name `kinship.CountedOnce` as what keeps a marriage and its implied spouse link one fact |
| Docs | `docs/conclusion-layer-data-model.md` §5.1, §13 item 4 | Say the entity comparison's rules are open and that it must count conflicts and show its reasons as sentences |

**Precedent.** `a5c0e0ff` ("Derive the spouse, parent and grandparent links …", S11-39) added the `core/kinship` layer package: the package with a package doc, an external `_test` package with an exact table and a fixed-seed invariants test, its `layerPackages` row, its `match-rules.mdc` glob and its `add-match-rule` lines. Copy that shape. Do not copy its `vocabkey` additions: this PR adds no vocabulary key.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A new layer package `core/recordlinks`, `numbers` true in `layerPackages`. Its package doc names it the record-links comparison, beside the matching stack: it judges two members the researcher already joined to one entity, and it is not a layer under graph alignment. It imports at least `bytes`, `maps`, `slices`, `core/kinship`, `core/propertycompare`, `core/segmentcompare`, `core/subjectcompare`, `core/valuecompare`, `core/vocabkey` | Working rule "Scoring and the view are two concerns": pure and registry-driven like matching. A separate package keeps Subject compare's points and Promote's walk out of it | Inside `core/subjectcompare` — that layer turns outcomes into points for Promote; this count has no points and reads links. Inside `core/graphalign` — the working rule forbids reading its scoring |
| D-2 | Inputs are pure types. `Member{ID []byte; Entity []byte; Kind vocabkey.SubjectType; Values subjectcompare.Values}` and `NewGraph(members []Member, links []kinship.Link) Graph`. `Entity` is the canonical entity of the member's accepted Identity Claim, nil when it has none; the caller sets it. `links` is the graph's `kinship.LinkSet.Links()` after `WithImplied` | The shared loader (S13-29) yields values from `subjectprofile.ForSubject` and links as `kinship` edges with recordings | Taking `graphalign.Layer` — imports graph alignment's types into a package the working rule keeps apart from it. Taking `graphcache` types — a layer may not reach `core/database` |
| D-3 | `NewGraph` indexes once and keeps: the entity of each member id; the person members by entity; each member's links read from its end (`Edge.From`: far end, key, `Recording`), grouped by the far end's entity. Rules: a later member with an id already seen is ignored (first wins); a link with neither end a member is ignored; a link's far end that is no member, or a member with a nil entity, has no entity, so the link never counts; an empty input, and the zero `Graph`, are valid and give no lines. `NewGraph` keeps references to its inputs and does not copy them; neither `NewGraph` nor `Lines` changes them. A pivot's entity is looked up in its own graph's index (D-7) | 50 graphs give 1,225 pairs; each graph is indexed once by the caller, not once per pair. The loader's input is not under this package's control, so each odd input has one stated result | Indexing inside `Lines` — 49 rebuilds per graph at Scale. Copying the inputs — the caller owns them and does not change them |
| D-4 | `Lines(page []byte, a, b Graph, metas subjectcompare.Metas, cfg Config) []Line`. A line is made for each entity that has a member of kind `vocabkey.Person` in both graphs, other than `page` and other than nil; only person members on the entity are paired, and every compare runs with kind `vocabkey.Person`. A member of another kind on the same entity is ignored for lines. `Line{Entity []byte; Facts int; Strength Strength}`; `Entity` is a copy (`bytes.Clone`). Lines are sorted by `bytes.Compare` on `Entity` | AC-1.1, AC-1.4, AC-1.5; working rule "A line per shared person". A fixed order makes the output comparable and order-free (T-8) | Lines for events or places — Out (user, 2026-11-07) |
| D-5 | When a graph holds more than one person member on the line's entity, every pair (one member from each graph) is counted and the line takes the pair with the most facts | AC-1.2. Two members of one graph on one entity are the researcher's call that both are this person; the better-corroborated pair speaks for the line, as the best pair speaks in property compare | Summing pairs — counts one fact twice; first pair by id — the result would depend on ids |
| D-6 | Properties go through `core/subjectcompare`. New `func (r Rules) Unit(kind vocabkey.SubjectType, key vocabkey.PropertyKey) (vocabkey.PropertyKey, bool)`: false when `r.Derived(kind, key)` or `Role` gives `RolePeriodEnd`; otherwise `r.Meta(kind, key)`, true. New `func CompareProperty(kind vocabkey.SubjectType, unit vocabkey.PropertyKey, a, b Values, metas Metas, rules Rules, value valuecompare.Rules) propertycompare.Result` = `propertycompare.Compare(Side(kind, unit, a, rules), Side(kind, unit, b, rules), metas[unit], value)`. `Compare`'s non-lifespan branch calls `CompareProperty`. `recordlinks` takes the units of the union of both members' `Values` keys through `Unit` (each unit once, ascending), calls `CompareProperty` for each, and counts a unit whose outcome is `propertycompare.OutcomeAgree` or `OutcomePartial`. A unit with no meta is not special-cased: it compares under the zero meta, which value compare reads as unsupported, so it is unknown and counts nothing, exactly as in `Compare` | Working rule "A matching Property counts": every product property both record, pooled ones once with their lead, lifespan out. One place decides how a property's values compare, for Promote's points and for this count (review M-1; the `add-match-rule` Do-not line) | A loop in `recordlinks` composing `Side` and property compare itself — a second copy of `Compare`'s property step. `subjectcompare.Compare` — reads only the weighted leads and the lifespan |
| D-7 | Links. For a pair of members x (in A) and y (in B): for each far entity e that both have links to, and each pair of a link of x to a member on e (key `k`, recording `ra`) and a link of y to a member on e (key `k'`, recording `rb`) with `segmentcompare.Corresponds(k, k')`: the pair (k, e) counts once when **any** such pair of links is not skipped, where a pair is skipped when `kinship.SameShape(ra, rb, same)` and `same(pa, pb)` is true when `pa`'s entity in A and `pb`'s entity in B are both non-nil and equal (a pivot that is not a member, or has no entity, is never the same). So a fact recorded through a pivot filed on one entity in both graphs is counted through the pivot, and one whose pivots are not both filed on one entity is counted as the link (AC-2.18). A link that carries several shapes is skipped when any of its shapes pairs (AC-2.19). Any far entity counts, the page's own (AC-2.12) and the line's own included (two members of one graph on F linked to each other) | Working rule "A matching link counts": same `BridgeKey`, far ends on one canonical entity; a relationship to anyone (the page's person included, AC-2.12) or a participation in the same event in the same role (AC-2.9, AC-2.10). `Corresponds` is the one correspondence rule. "Any pair not skipped" makes the result independent of how many links a member holds to one entity (review X-1) | Matching links on the far entity alone — counts a witness and a groom at one wedding as a shared fact (AC-2.10). Keying a map on `BridgeKey` — bypasses `Corresponds`. Skipping when any pair is skipped — a second, unpaired link would vanish |
| D-8 | `kinship.SameShape(a, b Recording, same func(pivotA, pivotB []byte) bool) bool` *(new)*: true when a shape of `a` and a shape of `b` have the same `Pattern` and `same` holds for their pivots. `CountedOnce` becomes: false when neither side is only implied; else `SameShape(a, b, func(pa, pb) bool { p, ok := paired(pa); return ok && bytes.Equal(p, pb) })`. `recordlinks` calls `SameShape` with no only-implied gate, so a pair matched by shape is skipped even when both sides also record the link directly (AC-2.15) | User, 2026-11-07: a marriage and the spouse link it implies are one fact, whether or not the spouse link is also written down. The pivot's own fact always counts: the event's participations, or the parent link to the middle parent of a chain, are links of the same members (AC-2.14, AC-2.16). One shape loop serves both callers | Calling `CountedOnce` — returns false when both sides record the link (AC-2.15 would count 2). A copy of the shape loop in `recordlinks` — a second mechanism beside `kinship`'s |
| D-9 | `Facts` is the number of counted property units plus the number of counted (key, entity) link pairs. The unexported helpers return what they counted, not bare counts: `properties` a `[]vocabkey.PropertyKey` of units, `links` a `[]linkFact` *(new, unexported)* `{key vocabkey.BridgeKey; entity string}`; D-5 compares their total lengths. `Strength` is `StrengthStrong` *(new)* `"strong"` when `Facts >= cfg.RecordLinks.StrongAt`, else `StrengthWeak` *(new)* `"weak"`, 0 facts included | Working rule "Two weights"; user, 2026-11-07: a shared person with no matching fact is still a thin line (AC-1.3). Returning what was counted leaves the written next add (an explanation) one exported field away (review M-7) | Dropping 0-fact lines — refused by the user. Reusing `graphalign.Assessment` — Promote's band, which the working rule "The score ranks, unseen" keeps out of new work |
| D-10 | Registry. `recordlinks.Rules{StrongAt int}` with a doc comment; `Registry.RecordLinks recordlinks.Rules`; `defaultRecordLinks()` in `core/matchregistry/recordlinks.go` *(new)* sets `StrongAt: 2` with the comment `// decided: user, 2026-11-07 (Spike 13 slice 7): two or more shared facts make a strong line`. `recordlinks.Config{Value valuecompare.Rules; Subject subjectcompare.Rules; RecordLinks Rules}` and `func (r Registry) RecordLinksConfig() recordlinks.Config`, the way `AlignConfig` bundles graph alignment's sections. Inventory row key: `recordlinks.StrongAt` | Numbers live only in the registry; the layer takes its sections as arguments (`add-match-rule`). The field name matches its registry field and inventory key, so "links" keeps one meaning in the package (review M-6). The caller (S13-27) fills `Value` from `valuerules.Load` as Promote does | A constant in the layer — the numbers guard forbids it |
| D-11 | Tests live in the external package `recordlinks_test`. `core/recordlinks/recordlinks_test.go` holds the builders and `TestLines`; `invariants_test.go` uses the same builders. Links are built as the loader builds them: keys from `segmentcomparetest.PartKeys` / `RelKeys` *(new)*, edges as `kinship.Link`s added to `kinship.NewLinkSet()`, then `WithImplied(segmentcomparetest.Terms(), matchregistry.Default().Kinship)` and `Links()`. Values from `propertycomparetest`; metas from `subjectcomparetest.Metas()`, which gains `religion`. The config is `matchregistry.Default().RecordLinksConfig()`; AC-3.2 edits `StrongAt` on its own copy | Implied links in the tests are the ones the loader produces. One builder for test keys instead of a third copy (review M-3); the seed pin covers `religion` (review M-4) | Hand-built `Recording`s — they could disagree with `Derive`. A test-local metas copy — drifts from the seed without failing |
| D-12 | `segmentcomparetest.PartKeys(role vocabkey.BridgeTerm, eventType vocabkey.EventType) (person, event vocabkey.BridgeKey)` and `RelKeys(term vocabkey.BridgeTerm) (person, relatedTo vocabkey.BridgeKey)` *(new)* take no `t` and build as `align_test.go` does today: `segmentcompare.Key(segmentcompare.Link{…}, Terms())` from each end (participation: `End` person / `Other` event, then the reverse, with `EventType`; relationship: `End` person / `Other` related_to, then the reverse, `SameType: true`). `align_test.go`'s `partKeys` and `relKeys` become one-line calls to them; `kinship_test.go`'s `part` and `rel` call them for their keys (and keep `t` for their other uses). No caller's signature changes. `TestKeysMatchSeededBridges` *(new, T-10)* holds them equal to `segmentcompare.Keys(Bridge(t, …), …, Terms())` | The shared builder home for test keys is `segmentcomparetest`. Taking no `t` keeps `partLink` / `relLink` and their 25 call sites as they are (review M-8, X-13); T-10 ties the hand-built link to the seeded bridge's endpoint order | Threading `t` through `align_test`'s builders — about 25 call-site edits. Leaving the copies — a change to `Key` or the seeded terms is fixed in three places |

**Rules this plan bends.** The map's working rule "A matching link counts" names `kinship.CountedOnce` as what keeps a marriage and the spouse link it implies one fact; D-8 uses `kinship.SameShape` without `CountedOnce`'s only-implied gate. The behavior is the user's answer of 2026-11-07 ("counting a marriage and the spouse link it implies as two facts" refused), which does not depend on whether the spouse link is also written down; the planner named `CountedOnce` before checking that it never skips two recorded links. The user confirmed on 2026-11-07 that a shared marriage and a spouse link both records write down are one fact. STEP-8 rewrites the rule and the S13-26 scope entry to name `SameShape` and the both-recorded case. The conclusion doc's §5.1 aim (count conflicts, show reasons as sentences) is a design aim, not a rule; STEP-8 rewrites it to say the first version does neither and keeps the aim for later.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/recordlinks` *(new)* | package | `core/subjectcompare`, `core/graphalign` | D-1 |
| `recordlinks.Member`, `Graph`, `NewGraph` *(new)* | type / func | `graphalign.Subject`, `graphalign.Layer` | D-2: graph alignment's types belong to the walk the working rule keeps apart; `Member` adds `Entity` |
| `recordlinks.Lines`, `Line` *(new)* | func / type | `subjectcompare.Compare`, `Result` | D-6: Compare weighs leads and the lifespan; Lines counts every property and links |
| `recordlinks.Strength` *(new)* | type | `graphalign.Assessment` | D-9 |
| `recordlinks.Rules`, `Config` *(new)* | type | `graphalign.Rules`, `graphalign.Config` | Same shape, own section (D-10) |
| `subjectcompare.Rules.Unit`, `CompareProperty` *(new)* | func | `Rules.Meta`, `Compare`'s property step | Extracted from them; `Compare` calls `CompareProperty` (D-6) |
| `kinship.SameShape` *(new)* | func | `kinship.CountedOnce` | `CountedOnce` is rebuilt on it (D-8) |
| `segmentcomparetest.PartKeys`, `RelKeys` *(new)* | test helper | `kinship_test` `part` / `rel`, `align_test` `partKeys` / `relKeys` | Those two copies move onto them (D-12) |
| `Registry.RecordLinks`, `RecordLinksConfig`, `defaultRecordLinks` *(new)* | field / func | `Registry.Kinship`, `AlignConfig`, `defaultKinship` | The registry's pattern for a section |

**Builds on.** `propertycompare.Compare`, `subjectcompare.Rules.Role` / `Meta` / `Derived`, `subjectcompare.Side`, `segmentcompare.Corresponds`, `kinship.Link` / `Recording` / `Edge.From`, the registry's section pattern, and `segmentcomparetest` as the test-key home (`add-match-rule` "Add a knob", "Add a layer package").
**Deletes.** The key-building bodies of `kinship_test.go` `part` / `rel` and `align_test.go` `partKeys` / `relKeys` (they call the shared helpers). `CountedOnce`'s loop and `Compare`'s inline property step move onto the new helpers; their signatures and behavior stay.
**Next adds.** S13-27 calls `NewGraph` once per graph and `Lines` per pair. A later version adds an explanation by exporting the counted units and pairs the helpers already return (D-9), and a conflict rule as a `Rules` field. [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) can rank with `CompareProperty` and `Lines` instead of a third property compare.

### Reach

None — no new path into existing code, no new state reachable. The package has no caller until S13-27; `Compare` and `CountedOnce` keep their outputs (T-3, T-7).

### Scale

At the map's Scale (slice 7: 50 Evidence graphs, 1,225 pairs; a household's few people per pair, an obituary's few dozen at most): `NewGraph` runs once per graph (D-3). Per pair, `Lines` walks A's person entities with a map lookup into B (linear in A's members), and per shared person compares each property unit once (values × values, as Subject compare does) and, per far entity both members link to, the two members' links on it (indexed by far entity, D-3). There is no write and no cache. The read's timing at Scale is S13-27's budget subtest, which owns the call.

### Edges

- **Assumptions.** (1) Symmetry: `Lines(page, a, b)` and `Lines(page, b, a)` give the same entities, facts and strengths. Property compare and the name compare are side-free (the name pairing uses the side with the smaller key); `Corresponds` is equality; `same` compares entities, so it is symmetric. T-8 holds it. (2) Order-free: shuffling members and links does not change the result, given D-3's first-wins rule for a duplicate id (T-8 shuffles only inputs with distinct ids). (3) `Edge.From` reads a self-loop from A's end only; the loader does not produce self-loops, and a self-loop counts at most once per key and entity.
- **Neighbors.** A name that shares a word but leaves two first names (AC-2.3) counts nothing; one that leaves a middle name (AC-2.2) counts. A multiple-cardinality property with one shared and one extra value agrees or partly agrees (AC-2.7); one with no shared value is unknown and counts nothing. A neutral term (sex `unknown`) is dropped by property compare and counts nothing. A negative value that matches the other side's positive is a conflict and counts nothing. A participation in the same event under another role counts nothing (AC-2.10). A spouse recorded directly on one side and only implied on the other counts once (AC-2.13). A spouse link whose marriages are not both filed on one entity counts as the link (AC-2.18).
- **Removed inputs.** A member with no entity is neither a line nor a far end (AC-1.4, AC-2.11): a record the researcher has not filed adds nothing. An unaccepted claim is the caller's: `Entity` is nil for it (D-2). A link skipped by shape loses nothing: the shape's pivot is on one entity in both graphs, so the participation or parent link through it is a link of the same members and counts (D-8; AC-2.14, AC-2.16, AC-2.19).
- **Real data.** The rule reads the shipped subject section: for a person, `Unit` maps name → name, sex at birth → sex at birth, gender → sex at birth, religion, languages spoken, interests and likeness → themselves, lifespan and lifespan span → not counted; there are no person period ends. T-1 runs on `matchregistry.Default()`, not a fixture. Likeness compares depictions by image, so two records' likenesses are unknown unless they share an image, and count nothing in practice.
- **Untrusted size.** Values per property, links per member and members per entity come from the Source's Evidence graph. Property compare is values × values, bounded as Subject compare is (and names by `tooManyWords`). Links are indexed by far entity (D-3), so a member with many links is linear in them, plus the pairs on one far entity. Members per entity (D-5) multiply pairs: two graphs each holding k person members on one entity give k² member pairs; a graph holding several Subjects on one entity is the researcher's explicit filing, bounded by the graph's Subjects (an obituary's 41).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/subjectcompare` | `func (r Rules) Unit(kind vocabkey.SubjectType, key vocabkey.PropertyKey) (vocabkey.PropertyKey, bool)` *(new)*; `func CompareProperty(kind vocabkey.SubjectType, unit vocabkey.PropertyKey, a, b Values, metas Metas, rules Rules, value valuecompare.Rules) propertycompare.Result` *(new)*; `Compare` calls it (D-6) |
| `core/kinship` | `func SameShape(a, b Recording, same func(pivotA, pivotB []byte) bool) bool` *(new)*; `CountedOnce` calls it (D-8) |
| `core/segmentcompare/segmentcomparetest` | `PartKeys`, `RelKeys` *(new)* (D-12) |
| `core/subjectcompare/subjectcomparetest` | `Metas` gains `"religion": {ValueType: propertymeta.ValueTypeTerm, Cardinality: propertymeta.CardinalityMultiple}` |
| `core/recordlinks` *(new)* | `recordlinks.go`: package doc (D-1); `type Rules struct{ StrongAt int }`, `type Config struct{ Value valuecompare.Rules; Subject subjectcompare.Rules; RecordLinks Rules }`, `type Member struct{…}`, `type Graph struct{ /* unexported index, D-3 */ }`, `func NewGraph(members []Member, links []kinship.Link) Graph`, `type Strength string` with `StrengthWeak`, `StrengthStrong`, `type Line struct{ Entity []byte; Facts int; Strength Strength }`, `func Lines(page []byte, a, b Graph, metas subjectcompare.Metas, cfg Config) []Line`; unexported `properties` (D-6) returning the counted units and `links` (D-7) returning the counted (key, entity) pairs |
| `core/matchregistry` | `Registry.RecordLinks recordlinks.Rules`; `Default` sets it from `defaultRecordLinks()`; `recordlinks.go` *(new)*; `func (r Registry) RecordLinksConfig() recordlinks.Config` |

```go
// Lines is one line for each person both graphs hold on one canonical entity,
// other than the page's own: the facts the two records share about that
// person and the line's weight. Lines are sorted by entity.
func Lines(page []byte, a, b Graph, metas subjectcompare.Metas, cfg Config) []Line
```

Errors: none — every function is total; `Lines` returns nil for no shared person.

### FFI

None — no wire change; S13-27 adds the RPC.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — One shape rule and one test-key builder**
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/segmentcompare/segmentcomparetest/terms.go`, `core/graphalign/align_test.go`. Skill: `add-match-rule`.
Do: add `SameShape` (D-8) beside `CountedOnce` with a doc comment; rewrite `CountedOnce`'s loop as a call to it. Add `TestSameShape` (T-2). Add `PartKeys` and `RelKeys` (D-12) beside `Bridge`; make `part`, `rel`, `partKeys` and `relKeys` call them. Add `TestKeysMatchSeededBridges` (T-10) in `core/segmentcompare/segmentcomparetest/terms_test.go` *(new)*.
Check: `go test ./core/kinship/... ./core/graphalign/... ./core/segmentcompare/...` passes; `TestCountedOnce` is unchanged and green; `TestKeysMatchSeededBridges` passes.

**STEP-2 — One property step in subjectcompare**
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go`, `core/subjectcompare/subjectcomparetest/subjectcomparetest.go`. Skill: `add-match-rule` "Rules that hold everywhere" (layer 3).
Do: add `Unit` and `CompareProperty` (D-6) with doc comments; `Compare`'s non-lifespan branch calls `CompareProperty`. Reword `Rules.Meta`'s doc comment so it no longer restates how values reach `Side` (the clause on comparing through Side when there is a lead): it says `CompareProperty` compares a unit through `Side`, and `Unit` names the unit. Add `TestUnit` (T-9). Add `religion` to `Metas`.
Check: `go test ./core/subjectcompare/... ./core/graphalign/...` passes with every existing test unchanged; `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` passes (the seed pin covers `religion`).

**STEP-3 — The package and its guard**
Files: `core/recordlinks/recordlinks.go` *(new)*, `core/matchregistry/guard_test.go`, `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule` "Add a layer package".
Do: the package doc and every exported type of §6 Go core; `NewGraph` builds the index (D-3); `Lines` returns nil for now. Add `{"core/recordlinks", true}` to `layerPackages` after `core/kinship`; add `core/recordlinks/**` to the glob.
Check: `go test ./core/matchregistry/...` passes (`TestLayerPackages` includes `core/recordlinks/pure` and `/numbers`).

**STEP-4 — The registry section**
Files: `core/matchregistry/registry.go`, `core/matchregistry/recordlinks.go` *(new)*, `docs/matching-model.md`. Skill: `add-match-rule` "Add a knob".
Do: D-10. Add the inventory row after the `kinship.MaxPairSteps` row: `| \`recordlinks.StrongAt\` | 2 | user, 2026-11-07 (Spike 13 slice 7) | Fewest facts two records share about a person for their line to be strong; fewer is weak |`.
Check: `go test ./core/matchregistry/...` passes (`TestKnobInventoryMatchesRegistry`, `TestDefaultIsFresh`).

**STEP-5 — Count properties**
Files: `core/recordlinks/recordlinks.go` *(new)* (from STEP-3), `core/recordlinks/recordlinks_test.go` *(new)*. Skill: `add-match-rule`.
Do: `properties` per D-6; `Lines` assembles lines (D-4, D-5, D-9) with property facts only. Write the builders (D-11) in `recordlinks_test.go` and `TestLines` (T-1) with the rows for AC-1.1 – AC-1.5, AC-2.1 – AC-2.8, AC-3.1 and AC-3.2. The rows for AC-2.9 onward are added in STEP-6.
Check: `go test ./core/recordlinks/...` passes.

**STEP-6 — Count links**
Files: `core/recordlinks/recordlinks.go` *(new)* (from STEP-3), `core/recordlinks/recordlinks_test.go` *(new)* (from STEP-5).
Do: `links` per D-7, added into `Facts`. Add the `TestLines` rows for AC-2.9 – AC-2.19.
Check: `go test ./core/recordlinks/...` passes.

**STEP-7 — Invariants**
Files: `core/recordlinks/invariants_test.go` *(new)*.
Do: `TestLinesInvariants` (T-8), as specified in §8.
Check: `go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/segmentcompare/... ./core/kinship/... ./core/recordlinks/... ./core/shortlist/... ./core/graphalign/... ./core/matchregistry/...` passes.

**STEP-8 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md`, `.agents/skills/review-pr/lenses/security.md`, `.agents/skills/ship-pr/SKILL.md`, `docs/matching-model.md`, `docs/conclusion-layer-data-model.md`, `docs/deployment-plan/spike-13/deployment-plan.md`, this plan.
Do:
(a) `add-match-rule`: `core/recordlinks` in the description's package list after `core/kinship`; replace the bullet that begins "**`subjectcompare.Rules.Role` says what a Property counts under and how**" with: "**`subjectcompare.Rules.Role` says what a Property counts under and how, `Rules.Meta` the meta its values compare under, `Rules.Unit` the unit a key counts as (none for a derived key or a period end), and `CompareProperty` compares one unit;** `Compare` and `core/recordlinks` both compare properties through `CompareProperty`." Add a bullet after the kinship one: "**Record links (`core/recordlinks`) count what two records share about a person.** `NewGraph` indexes one Evidence graph's members (each with the entity of its accepted claim) and links; `Lines` gives one line per person both graphs hold, other than the page's own, counting each unit that agrees or partly agrees and each corresponding link to the same entity, skipping a pair `kinship.SameShape` matches on pivots filed on one entity. Its one knob is `recordlinks.StrongAt`. It reads no weights, points or band." Add a line under "Add a layer package": test keys come from `segmentcomparetest.PartKeys` and `RelKeys`. Add `./core/recordlinks/...` to the Check command after `./core/kinship/...`.
(b) `.agents/skills/review-pr/lenses/security.md` "Stored numbers and lists" row and `.agents/skills/ship-pr/SKILL.md` line 106: add `core/recordlinks` after `core/kinship`.
(c) `matching-model.md` The stack: a paragraph after the kinship paragraph: "Beside the stack, `core/recordlinks` counts what two Evidence graphs of one person share about each other person they hold (Spike 13 slice 7): agreeing properties (through `subjectcompare.CompareProperty`) and corresponding links to the same canonical entity, weak or strong by `recordlinks.StrongAt`. It reads layers 1 to 3's rules and kinship's shapes, never graph alignment's scoring."
(d) conclusion §5.1: replace the whole paragraph that begins "**The entity comparison compares a handle's records pairwise**" up to and including "show its reasons as sentences." with: "**The entity comparison compares a handle's records pairwise** (decided 2026-10-29): the connection between two of the handle's Evidence graphs is read from the Subjects they share on canonical entities, and the handle's page shows the graphs and their connections, computed on read and never stored. A record linking to nothing, or two groups of records that do not connect, is what the entity lens surfaces. The first version (2026-11-07, Spike 13's design) draws one line for each other person two graphs share, counts the properties and links the two records agree on about that person, and weighs the line weak or strong; it ignores conflicts and shows no reasons. Counting conflicts, weighing a link like a nickname alone (Bob for Robert) as weak against one carried by names, family and places together, and showing reasons as sentences stay the aim for a later version." In it, "Spike 13's design" is a Markdown link to `deployment-plan/spike-13/evidence-graph-comparison.md#first-version-2026-11-07` (relative to `docs/`). Keep the two sentences after it, which begin "A Reconciliation Claim that settles a conflict" and "Promote's alignment score (§5.3) is not this comparison". §13 item 4: "Requirements are captured; how it is built and expressed is open there." becomes "A first version that counts shared facts is built (`core/recordlinks`, S13-26); conflicts, reasons and Source quality are open there."
(e) `deployment-plan.md`: in the working rule "A matching link counts", replace "`kinship.CountedOnce` keeps a marriage and the spouse link it implies one fact." with "`kinship.SameShape` keeps a marriage and the spouse link it implies one fact, also when both records write the spouse link down: a pair of links whose shapes pass through pivots filed on one entity is counted through the pivot." In the S13-26 scope entry, replace "with `kinship.CountedOnce` for implied ones" with "skipping a pair `kinship.SameShape` matches"; replace "Properties compare through `propertycompare.Compare` under `subjectcompare.Rules.Meta`" with "Properties compare through `subjectcompare.CompareProperty` per `Rules.Unit`".
(f) At ship (`ship-pr`), once the PR number exists: the map row's Status becomes `Landed #N`; this plan's Status becomes `Landed in [#N](PR URL) (date)` with a Landed note; `docs/deployment-plan/spike-13/completed.md` gains a `## S13-26 — Scoring two Evidence graphs` paragraph at the top, starting "**Landed <date>** in [#N](PR URL)." like the entries below it.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-26-scoring-two-evidence-graphs.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/recordlinks/recordlinks_test.go` `TestLines` *(new)* | Table, one row per AC, each the two graphs and the page entity → the exact `[]Line` (entities, facts, strengths) | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, AC-2.11, AC-2.12, AC-2.13, AC-2.14, AC-2.15, AC-2.16, AC-2.17, AC-2.18, AC-2.19, AC-3.1, AC-3.2, D-3 – D-10 |
| T-2 | Go | `core/kinship/kinship_test.go` `TestSameShape` *(new)* | Same pattern, `same` true → true; same pattern, `same` false → false; different patterns → false; empty shapes → false; both recorded (neither only implied) with paired pivots → true; two shapes on `a`, one pairing → true | D-8 |
| T-3 | Go | `core/kinship/kinship_test.go` `TestCountedOnce` (unchanged) | Its seven cases → as today | D-8 |
| T-4 | Go | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` (unchanged) | `recordlinks.StrongAt` field and row → pass | D-10 |
| T-5 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` (new row) | `core/recordlinks` → pure, no numbers | D-1 |
| T-6 | Go | `core/matchregistry/registry_test.go` `TestDefaultIsFresh` (unchanged) | Still passes with the new section | D-10 |
| T-7 | Go | `core/subjectcompare/...`, `core/graphalign/...` (unchanged) | Subject compare's and Promote's alignment tests stay green after STEP-1 and STEP-2 | D-6, D-8, D-12 |
| T-8 | Go | `core/recordlinks/invariants_test.go` `TestLinesInvariants` *(new)* | `rand.NewSource(1)`, 200 rounds. Each round: an entity pool of 6 ids (the first is the page `p`); graph A and graph B each get 3 – 6 person members with distinct ids (prefix byte 0xA0 for A, 0xB0 for B), each filed on a random pool entity or nil (1 in 5), the page entity held by one person member in each graph; 1 – 3 event members per graph on a second pool of 3 event entities or nil; names from the pool "John Smith", "Mary Smith", "Amy Lee Long", "Amy", "Robert Brown", each a plain given and family name that agrees with itself; a random sex at birth (`male`, `female`) and a random gender (`man`, `woman`) or none; each event member gets a random event type (`birth` or `marriage`) and 1 – 3 participations from distinct random person members of the same graph, each with a random role (`subject`, `father`, `mother`, `witness`); each graph gets 0 – 2 relationships (`parent`, `spouse`) between two distinct random person members of the same graph; every link is a `kinship.Link` with keys from `PartKeys` / `RelKeys` and one association id unique in the graph (prefix byte 0xC0 for A, 0xD0 for B, then a counter); the edges are added to a `kinship.NewLinkSet()` and `WithImplied` runs. To raise the odds of shared shapes, half the rounds give B a copy of A's events and participations on the same entities (with B's own ids). Config `matchregistry.Default().RecordLinksConfig()`. Assert: (a) `Lines(p, a, b)` equals `Lines(p, b, a)`; (b) the same members and the same edges shuffled with `r.Shuffle` before they are added to a new `LinkSet` (and `WithImplied` run again), passed to `NewGraph`, give the same lines; (c) no line for `p` or nil, and a line for exactly each entity with a person member in both graphs; (d) `Strength` is strong exactly when `Facts >= StrongAt`; (e) `Lines(p, a, a)` gives each line at least 1 fact when the entity's member has a name | D-3 – D-9 |
| T-10 | Go | `core/segmentcompare/segmentcomparetest/terms_test.go` `TestKeysMatchSeededBridges` *(new)* | `PartKeys(subject, birth)`, `PartKeys(subject, "")`, `PartKeys(father, birth)`, `RelKeys(parent)`, `RelKeys(spouse)` → equal to `segmentcompare.Keys(Bridge(t, participation or relationship), term, eventType, Terms())` | D-12 |
| T-9 | Go | `core/subjectcompare/subjectcompare_test.go` `TestUnit` *(new)* | Shipped subject section, person: name → (name, true); gender → (sex_at_birth, true); sex_at_birth → (sex_at_birth, true); religion → (religion, true); lifespan, lifespan_span → (_, false). Event: start_date → (_, false); date → (date, true) | D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/kinship` and `core/graphalign` (STEP-1), `core/subjectcompare` (STEP-2), `core/database/subjectvocab` (the `religion` meta), `core/matchregistry` guards (STEP-3, STEP-4).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: the TestLines row for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks); on Linux it skips the Xcode steps, and the report says so
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: no exported name was renamed or deleted; `rg -n "CountedOnce" docs/deployment-plan/spike-13/deployment-plan.md` finds no line that says it keeps a marriage and its spouse link one fact; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (STEP-8 e)
- [ ] UI matches the brief inventory — no UI
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated — not gated
- [ ] Domain doc updated: conclusion §5.1 and §13 item 4, matching-model.md The stack and knob inventory, add-match-rule, security lens, ship-pr (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app: the package has no caller until S13-27, whose manual check and bar 10 exercise it. The reviewer reads `TestLines`'s rows against §2.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Rebuilding `CountedOnce` on `SameShape`, or `Compare` on `CompareProperty`, changes Promote | T-3 and T-7 run unchanged; both signatures stay |
| The name compare gives a partial for a pair the user would not call a match | AC-2.2 and AC-2.3 pin the boundary; the relation stays the name compare's, so a later change to it reaches both Promote and this count |
| The symmetry invariant fails on a tie in property compare | T-8 (a) surfaces it; property compare picks its speaking pair by outcome, similarity, relation and Same key, never by side |

**Stop and report if:**

- `subjectcompare.Rules.Meta`, `Role`, `Derived` or `Side` no longer exist with the signatures in §4.
- `kinship.Link` no longer carries `Recording`, or `LinkSet.WithImplied` no longer takes `(segmentcompare.Terms, kinship.Rules)`.
- Any existing test in `core/subjectcompare`, `core/kinship` or `core/graphalign` fails after STEP-1 or STEP-2, and the fix would change `Compare`'s or `CountedOnce`'s output.
- `TestLayerPackages` fails because `core/recordlinks` needs a numeric literal other than 0 and 1, or an import of `core/matchregistry` or `core/database` outside tests.
- AC-2.2 or AC-2.3 fails under `matchregistry.Default().Value` (for example an alias table relates "Amy" to another word): report the relation the name compare gave.
- T-8 (a) fails for a case where `propertycompare.Compare` itself is not symmetric, or T-8 (e) fails for a name in the pool: report the values; do not add a tiebreak or a special case in `recordlinks`.
- An AC's expected count cannot be met without counting a lifespan, a conflict, or a link to a far end with no entity.

## 12. Open questions

None.

## 13. PR

**Title:** S13-26 — Scoring two Evidence graphs
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (`TestLines` rows); deviations from the plan; manual check (none: no caller until S13-27); `Plan: docs/deployment-plan/spike-13/prs/S13-26-scoring-two-evidence-graphs.md`.
