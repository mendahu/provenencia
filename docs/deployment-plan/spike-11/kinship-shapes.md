# Kinship shapes: one fact, several ways to record it

**Status:** pulled into [Spike 11](deployment-plan.md) on 2026-10-24 (S11-41, S11-38, S11-39, S11-40, S11-42); re-cut 2026-10-26 with S11-49 (implied links in the graph cache) and S11-50 (matching warms while the project opens). Idea first written 2026-10-19 while shipping S11-14, as `docs/ideas/relationships-from-shared-events.md`; grandparent added by the user on 2026-10-24. What stays later is in [`ideas/kinship-view.md`](../../ideas/kinship-view.md).

Matching today: [`comparison-layers.md`](comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6) (segments, layer 6). Seeded terms: [`seeded-vocabulary.md`](../../seeded-vocabulary.md).

## The problem

A kinship fact can be recorded in more than one shape:

| Fact | Direct | Through an event | Through a chain |
| --- | --- | --- | --- |
| A and B are spouses | `relationship · spouse` between A and B | A and B are both `subject` of one `marriage` | — |
| A is B's parent | `relationship · parent`, A parent of B | A is `father` or `mother`, B is `subject`, in one `birth` | — |
| A is C's grandparent | `relationship · grandparent`, A grandparent of C | — | A is a parent of B, and B is a parent of C (each in any parent shape) |

Matching does not know these are one fact. `segmentcompare.Corresponds` lines up keys by type, term and direction, so a participation never corresponds to a relationship. If a Source records the marriage and the research graph records the spouse relationship, the link adds nothing to the match. A sparse Source (a birth record naming the parents) against a tree built from relationships (A parent of B) gets no help from the parents.

## Shape

**Implied links.** Before alignment, each side's links are read through a small table of **kinship patterns**. A pattern that holds adds an **implied link**: a relationship link between the two people, with the same key a direct link would have. Graph alignment and segment compare see it as an ordinary link and need no change. They stay one hop: an implied grandparent link is one hop from A to C, made from two parent links.

| Pattern | Holds when | Implies |
| --- | --- | --- |
| Marriage spouses | Two people are each `subject` of one `marriage` event | `relationship · spouse` between them |
| Birth parents | A is `father` or `mother` and B is `subject` in one `birth` event | `relationship · parent`, A parent of B |
| Parent of a parent | A is a parent of B and B is a parent of C, each by a direct or implied link | `relationship · grandparent`, A grandparent of C |

- **Directed where the roles say so.** Birth roles name the parent end (`father`, `mother`) and the child end (`subject`), so the implied parent link is directed like a recorded one, and reads through `segmentcompare.Key` with its inverse (`child`).
- **One way only.** A chain implies a grandparent. A grandparent link does not imply a parent in between; there is no one to name.
- **Grandparent stops at two generations.** Great-grandparent and deeper are not seeded terms and are not derived.
- **Seeded terms only.** The patterns name product terms through `core/vocabkey` constants. Every term is already seeded (`subjectvocab/registry.go`); `birth` and `subject` have `vocabkey` constants, and S11-38 adds constants for `marriage`, the roles `father` and `mother`, and the relationship terms `spouse`, `parent` and `grandparent`, checked by `TestVocabKeysMatchSeed`. A catalog's own terms imply nothing.

- **The subject is an event's principal** (user, 2026-10-26). A marriage's couple are its two `subject`s, as a birth's child and a death's decedent are its subject. There is no `spouse` participation role; the seed, `vocabkey` and Swift drop it (S11-39). An existing catalog keeps its own `spouse` role term, which implies nothing.
- **Bounded per pivot** (user, 2026-10-26). A pivot with more than `kinship.MaxPivotEnd` (10) participants at either end, or a person with more than 10 parents, reads as malformed data and implies nothing, so one event cannot turn a linear list of links into a quadratic one.

## Where implied links live

Re-cut 2026-10-26 after S11-39's first ship loop measured the first design on a 12-generation tree (8,190 parent links, 4,095 marriages, about 18,000 handles reached through one shared place).

**One result per pivot, in the graph cache.** Every pattern reads one node's links. A marriage's spouses and a birth's parent come from that event node's participations. A grandparent through B comes from B's filed parent links and the birth parents B's event neighbors imply: one hop out. The graph cache keeps each pivot's result, derived the first time a reader needs it, and drops it when a commit touches the pivot or a node one hop out (`canonlinks`' deps function, which the commit calls for every handle it touches), or the vocabulary changes; a first load of a node keeps it. `canonlinks.Reader.For` answers one person's implied links from the results of its events and of its parents and children. `loadCanon` (S11-39), the evidence query's `canonLinks` (S11-40), and later trees and person pages read it. There is one copy per open catalog; nothing is written to the catalog, and graph verify mode checks each result against a fresh derivation (S11-49).

**The Source side derives when its layer loads.** A Source's layer is small and already memoized with the layer; the layer loader calls `core/kinship` over it (S11-39).

**Why not per proposal.** Deriving over the walked component on every proposal cost about 20–30 ms on the tree, and memoizing the whole walk per Source missed on every pin (a pin changes the walk's seeds), was thrown away by every commit, and kept one catalog-sized copy per Source. The graph cache keeps one copy, current through commits.

**Warm at open.** Loading every person, event and place and building the name index was most of the first proposal's cost (about 0.65 s of 1.05 s locally, three times that on CI). It moves behind the open spinner (S11-50, user 2026-10-26). Kinship results are not warmed; they are derived on first read, about 20–30 ms across that catalog.

**One pure package.** `core/kinship` holds the pattern table and the derivation: it takes a side's links (each with its two ends, its key and its bridge Subjects) and the catalog's `segmentcompare.Terms`, and returns only the implied ones, each carrying the links and the event it was made from. Their keys come from `segmentcompare.Key` over the same link reading the loaders use (moved to `segmentcompare` by S11-41), never built by hand. It imports no `core/database`, joins the one purity guard, and is the place a family tree or a person page will read later ([`ideas/kinship-view.md`](../../ideas/kinship-view.md)). Reading the canon's links into it lives in `core/database/canonlinks`: `EdgeOf` reads one filed link as an edge with its keys, and `LoadTerms` reads the catalog's bridge terms; the walk, the evidence query and `Reader.For` share them.

## Counting once

- **One fact, one link per side.** When a side records the same fact in two shapes (a spouse relationship and a shared marriage between the same two people), the side carries one spouse link between them, whose bridge Subjects are both shapes'. It is credited once. The loaders already merge same-ends, same-key edges (`layer.go`'s merge on `edgeSeenKey`, `loadCanon`'s dedupe); implied links go through that merge, so there is no second one.
- **Counted once across sides** (user, 2026-10-26). When both sides record the marriage and the walk pairs the two marriage events, the people already meet through the event path, so the implied spouse link adds nothing. When the events are not paired (a sparse marriage whose event row is not placed), the implied link credits. The same holds for a birth's parents and for a chain through a paired middle parent. Each link carries its shapes (pattern and pivot), and graph alignment's `corresponding` skips the pair (S11-39). A link both sides also record directly keeps today's credit. Planning measured the effect: an undated marriage on both sides stays at today's 2.875 (weak) for a named couple, while a marriage against a spouse relationship scores 3.375 (medium), the event path passing on the marriage's own evidence.

## Weight

An implied link is weighed as the direct link it stands for: spouse, parent and grandparent are all the segment registry's Other pair. No new knob. Whether parent links deserve more is the S11-14 dogfood question, separate from this.

## On the evidence sheet

A relationship line today comes from a corresponding segment ([`claim-evidence.md`](claim-evidence.md#evidence-through-relationships)). An implied link that corresponds gives a relationship line too, and it says how it was recorded:

- "They are spouses: both are spouses in the marriage on 12 May 1842" (an implied link on the Source side), or "…the record has them as spouses" (direct).
- Checking the line pins its bundle: every bridge behind the implied link (both participations, or both parent links of a chain), and the event's defining properties, as a relationship bundle does today.

The line is drawn by the sheet's existing relationship Card. No new chrome, but it is a wire change: `claimevidence.Relationship` gains how each side recorded the fact and the event behind it, which crosses the FFI to Swift (S11-40), and the sheet words it (S11-42). Today a bundle is one bridge between a member and the neighbor's member (`claimevidence/build.go` `bundles`, `BundlePinID`); an implied link has several bridges and passes through an event, so S11-40 decides the bundle's shape and its pin id.

## What stays later

In [`ideas/kinship-view.md`](../../ideas/kinship-view.md):

- **Family trees and person pages** reading both shapes as one fact (a canonical kinship view on `core/kinship`).
- **Other patterns:** siblings from shared parents, aunts and uncles (`pibling`), godparents at a baptism, divorce, census households, great-grandparents.
- **Conflicts between shapes** (a direct `sibling` and a shared marriage as spouses).
- **A project's own patterns** (see [`ideas/term-management.md`](../../ideas/term-management.md)).

## Open questions

1. **Counting once across paths.** Answered (user, 2026-10-26): drop the implied link's credit when both sides share the shape and its pivot is paired; see [Counting once](#counting-once).
2. **The sheet line.** Is a line that says how a fact was recorded wanted in this spike (S11-40), or is matching enough for now?
3. **Birth roles beyond `father` / `mother`.** A birth participant recorded with role `child` (rather than `subject`), or a baptism's parents. Proposed: birth with `subject` only; the rest waits.
