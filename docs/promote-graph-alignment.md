# Provenencia — Promote Graph Alignment

## Status

**Agreed 2026-10-06; scheduled as Spike 9 slice 9** (§10). Renamed from “Promote alignment” to **Promote graph alignment** (2026-10-07) so the name names the graph-to-graph proposal, not property-only matching. It replaced the per-Property compare checklist and the per-Subject walk; [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §5.3–§5.4 now state its model rules. How we got here, and the directions we turned down: [`ideas/archive/promote-matching.md`](ideas/archive/promote-matching.md). Matching today: [`matching.md`](matching.md).

**The data model doesn't change:** one Identity Claim per Subject and handle, with Observation pins on the claim and backfill (§5, §5.1).

---

# 1. The problem

Promote files Interpretation Subjects onto canonical handles. A Subject's identity is carried mostly by its **neighbors** (its birth event, its places, its family), not by its own Properties. An Evidence graph is also much bigger than one Subject: the reference case, an obituary, has 41 Subjects. Ticking agreements Property by Property, Subject by Subject, is the wrong amount of work. Most of those agreements the machine can decide, and none of them can see the structure that actually identifies people.

**The picture** ([`ideas/archive/promote-matching.md`](ideas/archive/promote-matching.md#the-stack-how-to-picture-the-problem)):
- every Evidence graph is a **layer**;
- the layers stack so that Subjects about the same thing line up in **columns**;
- a handle is a column, and an Identity Claim is a vertical link;
- Promote **lays a new layer onto the stack**.

A vertical link's strongest evidence is the horizontal structure around it lining up too.

**The goal:** Promote proposes the whole graph alignment automatically, and the researcher fine-tunes the few links that need judgment.

---

# 2. Principles

1. **The machine proposes; the researcher decides.** Nothing is filed without the researcher's Done. Promote stays create-only.
2. **Decisions are fixed.** A choice the researcher made is never overridden by a re-run. Everything not yet decided follows the decisions.
3. **One claim per Subject.** A batch is many independent claims, and any one can be rejected later on its own.
4. **No hard-coded lists of Properties or relationships.** What counts as evidence, and how much, comes from metadata (value types, cardinality, bridge types and roles) and from the catalog's own data. A Property or bridge kind added later takes part with no code change.
5. **Precision over recall.** A false merge is the costliest mistake. The core never proposes a merge below the accept bar: such a row comes back with no target, its candidates as alternatives. The page starts it **unset**, so Done waits for the researcher's Skip or New, and drafts a weak one on its best match for review.
6. **Explainable, deterministic, local.**
   - Every suggestion carries its reasons.
   - The same inputs give the same proposal.
   - Everything runs on device.

---

# 3. The flow: one page

Promote is one workspace page; the choose-target screen and the separate claim step go away.

- **Entry:** the Promote button on any unpromoted Subject card, as today. The page opens with that Subject's row, preselected to its best match, plus **"Map the rest of this graph (N Subjects)"**. Promoting one Subject is the same page with one row.
- **Rows:** one per primary Subject on the Evidence graph (person, event, place).
  - **Already promoted:** shown read-only, with no controls. They are anchors for the graph alignment, and editing them is a separate workflow ([`ideas/identity-claim-review.md`](ideas/identity-claim-review.md)), not this page.
  - **Unpromoted:** a target dropdown with the best match preselected, the next few alternatives, **New**, and **Skip**.
- **Assessment:** each row shows **strong**, **medium**, **weak** or **no match**. Clicking it opens a **sheet** that opens with why the row was suggested (its band and Contested marks at the title, the reason as their tooltip, the Updated caption under it) and then the evidence that the row is the record. The sheet loads the row's claim evidence when it opens and draws it as sentences, with no weights: a Card per property with a line per distinct value (identical members collapse into one line that expands to each record), likenesses as two crops side by side, and a relationship line for each mapped neighbor that corresponds, named by the connection that matched. The lines and relationships that agree are preselected as pins and can be toggled.
- **Claim fields per row:** status, confidence and argument. The argument is the researcher's own: Promote drafts none.
- **Bridges aren't rows** (participation, location, relationship, place relationship). Each shows as a subrow under every end that is a row on the page (a connection between two rows shows under both). A subrow says whether it files on Done, waits for a named end, or has both ends on one record; when the other end is not a row, a note says it is already filed or not on the page. The footer counts the connections that file. Filing rules: §9.1.
- **Possible duplicates:** two rows landing on the same existing handle, or two rows that would both be new records (unplaced or decided New) whose properties clear the accept bar, get a warning ("these may be the same record; combine them on the Evidence graph"). A link key both rows have neighbors for suppresses the warning when none of those neighbors are the same subject or a match; a directed key compares only neighbors on the same end (a son with a son, not with a father). A row whose best record another row took starts with no target and the warning, whatever its band, so Done can't file both on one handle without a choice. There is no row-to-row option.
- **Done** writes the whole batch in one transaction (§9).
- **Leaving:** the leave guard covers accidental navigation. Drafts aren't persisted: re-opening re-proposes everything, and only manual changes are lost.

## 3.1 Suggested and decided rows

Every row is **suggested** (the machine filled it) or **decided**: the researcher changed its dropdown or toggled anything in its sheet.

- **On open,** the only fixed points are the already-promoted rows; the clicked Subject's row is proposed like the rest.
- **On any change,** that row becomes decided, and graph alignment re-runs over the whole page with every decided row held fixed. Suggested rows and their sheets are recomputed; decided rows are never touched. A row decided New or Skip is held too: it takes no handle and the walk doesn't pass through it.
- **Rows whose suggestion changed** get a brief "updated" mark with the reason ("via Gracie → PER-Z"), or, when settling moved the row because of a neighbor, what that neighbor did ("after Gracie moved to PER-Z").
- **A contested row** (a comparison behind its pair conflicts) shows a Contested mark beside its band, at any band; the band is unchanged.
- **A decided row that the new context contradicts** keeps its choice and gets a warning. The researcher resolves it.

---

# 4. The graph alignment function

One pure function, called on open and after every decision:

```go
// Align proposes a handle or nothing for every primary Subject on one
// Evidence layer (graph alignment), holding the fixed pairs as given.
func Align(layer Layer, canon Canon, fixed []Fixed, cfg Config) Proposal
```

| Input | Holds |
| --- | --- |
| `layer` | One Evidence graph: Subjects, bridges, Observations with values and provenance |
| `canon` | The part of the canonical graph reachable from the seed handles: handles, their links, and their auto-reconciled values (every rank) |
| `fixed` | Decided rows, plus every already-promoted Subject on the layer, whose existing claim is a free anchor |

`Proposal` is one row per Subject:
- the target, plus the next 2–3 alternatives;
- the assessment;
- the comparisons (the score's inputs); pins come from the claim evidence (§9);
- the reasons;
- flags (conflicts with decided rows, possible duplicates).

**Pure and deterministic,** like the `autoreconcile` / `autoreconciler` split. A pure package (`graphalign` or similar) runs `Align`; a database package loads the inputs; ties break by id. The obituary is a golden test.

## 4.1 Abstraction, composition, and tunables

`core/graphalign` composes the matching layers: entity compare judges a pair, segment compare judges a link, and `Align` walks, places and settles. It talks to no database and reads its numbers from the registry's alignment section (`graphalign.Rules`, filled in `core/matchregistry/alignment.go`); no layer imports the registry. The catalog adapter (`core/database/promotealign`) builds `Layer` and `Canon` and calls `Align`; the Promote page presents rows. Every tunable and its start is in the [knob inventory](matching-model.md#knob-inventory).

---

# 5. Walking: best-first propagation

This is collective entity resolution: graph alignment by propagation (cf. PARIS, similarity flooding). The walk is **best-first, not recursive**. A depth-first walk commits to whatever it reaches first; a priority queue lets the strongest matches settle first, so the weaker ones are decided with the most context.

1. **Handles.** A Subject's handles come from its shortlist or from a link to an accepted pair. The loader shortlists by cheap keys and `graphalign.Candidates` keeps the handles with similarity at least `alignment.MinSimilarity`, the best K by similarity and every handle tied with the K-th, up to the tie cap. `Align` runs no shortlist of its own; it reads the loader's lists (`Canon.Candidates`). A handle reached through a corresponding link of an accepted pair is nominated at any score; one placed pair nominates at most the most similar `TopK` plus ties, up to `TieCap`, for each neighbor (S11-21).
2. **One queue.** Every fixed pair is placed first. Each candidate is scored (entity compare, then the segments around it) and enters one best-first queue. A pair is popped best first, and placed when it clears the accept bar and its handle is free (one-to-one: within one layer a handle takes at most one Subject). A handle another Subject holds records who took it, so a loser can read `taken`.
3. **Unplaced neighbors.** A link whose neighbor is not placed yet is judged by entity compare on the neighbor's own properties: alike counts as lined up, a differing one counts against the pair from the start, and an unsure one adds nothing. Once the neighbor is placed on that canon neighbor the same number follows, so a score does not depend on which pair the walk placed first.
4. **Placing a pair** nominates, for each neighbor, the most similar of the handles its links point at (as step 1 bounds them), re-scores the neighbors that are not placed, and marks the placed ones dirty. A candidate a free neighbor already had keeps its own reason; only a fixed row's existing candidate gains the nominating `via`, so a decided row has the handles its neighbors point at.
5. **Settling.** When the queue empties, only dirty pairs are re-scored. A pair is released when it falls under the accept bar or a free rival leads it by the settle margin; the released Subjects walk again. At most `SettleRounds` rounds. A released Subject that ends on another handle carries a move reason (`neighbor_placed`, `neighbor_moved`, `neighbor_released`) and the neighbor behind it. A fixed pair is never released. Every score is replaced, never kept higher, so support from a neighbor that moved is dropped and a "via" reason stays only while that neighbor still supports the pair.
6. **Rows.** A row's band is its evidence against `Alignment.Bands`. A row is **contested** when any comparison behind its reported pair conflicts, whatever its band. A row with no accepted candidate comes back with no target and its candidates as alternatives; a decided row whose neighbors point at a handle that leads by the settle margin is flagged `ConflictWithFixed`.

The work is about (Subjects + bridges) × candidates × log: well under a millisecond at obituary scale.

## 5.1 Layer shape: every bridge pair, and orphans

A layer is **one Evidence graph’s Subjects**, not one connected component. *Map the rest of this graph* lists every primary Subject on that graph, including stragglers.

**Every Evidence bridge is a pair** of primary ends (call them subject and object). Promotion status is independent on each end. Graph alignment and bridge filing together cover all four:

| Subject | Object | Who acts |
| --- | --- | --- |
| Unpromoted | Unpromoted | Alignment proposes both (seed/propagate once either end is fixed). Done may mint two handles, then file the bridge. |
| Promoted | Unpromoted | The promoted end is a **fixed anchor**; seed walks to the unpromoted end from that handle’s canonical neighbors. |
| Unpromoted | Promoted | Same mechanism, other direction (e.g. open Promote on a new person whose event neighbor was filed earlier). |
| Promoted | Promoted | Not an identity proposal for those rows. It is a **missing / unfiled relationship**: the Evidence bridge exists and both ends are handles, so §9.1 files the canonical association (shown on the page as a connection that will be filed). |

**Orphans and disconnected islands** are expected. A primary Subject with no bridges, or a cluster that shares no path with any fixed/promoted anchor, is **unreachable** (§5 step 1): property-only matching only; a match below the bar comes back as an alternative with no target, and the row starts unset on the page. Structure-backed scoring never assumes a path from the entry card to every row.

---

# 6. Scoring

Nothing names a Property or a relationship. Every weight comes from metadata or from data.

**Node comparisons.** The score does not name a property. Entity compare (`subjectcompare.Compare` on the two sides' profiles) runs layer 2 on each lead property in the registry's subject section and returns evidence in points: agree earns the agree weight, partial the agree weight × similarity, conflict loses the conflict penalty, unknown counts nothing. Weights are in the [knob inventory](matching-model.md#knob-inventory): name 2.5 / 3, sex at birth (with gender) 0.6 / 1, toponym 3 / 3, event type 0.5 / 20, date 2 / 3, a person's lifespan 0.1 / 20. A property the subject section does not list scores nothing. A neutral term is no evidence.

A line starts pinned only when value compare says the values are the same; the researcher can pin any other line. A partial name can make the row weak without starting pinned.

**Provenance scales each comparison:** Source credibility, transcription uncertainty, and the claim confidence of the member the evidence came from. These are the same inputs the auto-reconciler's provenance uses ([`conclusion-reconciliation.md`](conclusion-reconciliation.md) §4).

**Edge comparisons.**
- A bridge is read as a typed **key** (`vocabkey.BridgeKey`): bridge type, term in canonical direction, the end it is read from (empty for a symmetric link between two Subjects of one type), and the event type on a participation. An example: `participation·subject·person·birth`. A term and its inverse read as the one that sorts first with the ends swapped, so "Mary parent of John" and "John child of Mary" give the same key from the same end, and a spouse never corresponds to a sibling. Two links correspond when their keys are equal (`segmentcompare.Corresponds`).
- A corresponding neighbor is compared by segment compare (`core/segmentcompare`), with no fan-out. The segment section of the registry declares which links are one-cardinality: subject of a birth and subject of a death, from both ends. A one-cardinality link whose neighbor lines up earns `OneSupport + OneCredit × neighborEvidence` (1.5 + 0.5 ×); one whose neighbor is already placed on another handle loses `OneConflict` (3), and one whose neighbor is not placed yet is judged by entity compare on the neighbor's own properties (`Unplaced`: alike earns as above, a differing neighbor loses `OneConflict`, an unsure one adds nothing). Weights: [knob inventory](matching-model.md#knob-inventory). Every other link earns `OtherSupport + OtherCredit ×` the neighbor's evidence (0.25 + 0.25 ×) and is silent when the neighbor is placed elsewhere. `neighborEvidence` is the pairwise total for that pair; its own links stay out, so credit stops at one hop. A negative total subtracts.
- A **gate**, looked up by key, can require the candidate's own evidence before support is added, and can cancel a factor of that evidence when the same-end neighbor is a different handle. A key with no gate does neither. The shipped `part_of` gates both require evidence of 2 (the accept bar). The part end also cancels with factor 1, so a shared parent does not file a different toponym, and the same toponym under a different parent falls below the accept bar. The whole end does not cancel: a whole has many parts, and cancelling it while its part is still settling never lets the chain agree. `succeeded_by` and location have no gate.

---

# 7. Loading: prefetch, then walk in memory

There are no database calls during the walk. Everything it reads is loaded first:

1. **The layer:** one Evidence graph's Subjects, bridges, Observations and provenance, held on the catalog's in-memory graph (`graphcache`, `g.EvidenceGraph`).
2. **Seed handles:**
   - the fixed pairs;
   - the handles of already-promoted Subjects;
   - the candidates `graphalign.Candidates` keeps for each unpromoted Subject (`alignment.MinSimilarity`, the best `TopK`, ties up to `TieCap`; see the knob inventory in [`matching-model.md`](matching-model.md)), passed to `Align` in `Canon.Candidates`, ranked over the handles the shortlist index offers (`core/shortlist`, built by `loadShortlist`), not every handle of the kind.
3. **Walk the in-memory catalog graph** from the seeds through every link to a person, event or place, with no hop limit. A node not yet in memory loads in one batched read per wave of the walk; a warm one costs nothing. Handles carry their cached auto-reconciled values.
Every read is batched (`sqlbatch`, 500 ids per IN list), so the query count follows the waves of the walk, not the size of the layer or the catalog.

**Canonical adjacency without S9-28.** It can be derived through members: H's member m → bridge → neighbor n → n's accepted claim → G. That's a join, and it loads in batch. S9-28 caches it as canonical edges and adds bridge filing, so a prototype of graph alignment (`Align`) doesn't need to wait for it.

---

# 8. Caching and decisions

- **Decisions live in the page's draft:** the Swift row list (suggested or decided, target, sheet toggles, claim fields). They're sent as `fixed` on every call and never stored.
- **Loads read the catalog's in-memory graph** (`graphcache`), which keeps handles, Evidence graphs (keyed by graph id) and the match index between calls, so re-runs after decisions skip loading. Each commit drops what it touched (its effects), not a whole revision. It keeps at most `maxEvidenceGraphs` Evidence graphs (`graphcache`); loading one more drops the least recently used graph with its memos, and a dropped graph loads again on its next read. One Evidence request reads each graph once, whatever the cap.
- **The match index** (the shortlist index) is cached on the graph and dropped by any commit that names a handle, by a vocabulary commit (its keys read term keys and metas), and by a new alias index (`SetAliasIndex`) or name format (`SetNameFormat`).
- **The candidates memo** holds each Subject's ranked candidate ids on that graph's memo. It is valid for one match index generation: it is dropped with the Evidence graph memo and ignored when the index generation changes.
- **Open-time warms** (`catalogsession.WarmOnOpen`, behind the open spinner) are held on the graph until the project closes: the built alias index, the default name format, then `promotealign.Warm`, which loads the shortlist kinds (every person, event and place) and builds the shortlist index, so the first proposal pays only for its own Source. `SetAliasIndex` and `SetNameFormat` drop the caches derived from them. The shortlist index is built at open only, and a failure to build it never fails the open: the next proposal builds it and reports the error. After a call that drops the warm state (a vocabulary or structure commit, `Drop`, a new alias index or name format), `catalogsession.Do` warms the alias index and name format again before it returns and leaves the shortlist index to the next proposal. A handle commit drops only the shortlist index, and the next proposal rebuilds it ([`name-index-on-commit.md`](ideas/name-index-on-commit.md)).

---

# 9. Writing

- **Done is one transaction for the whole batch:** all or nothing. Bridges need handles minted in the same batch, so a partial write could leave things half-linked. It's re-validated against the current state first: a write elsewhere since the proposal fails Done cleanly, and the page re-proposes. One audit revision covers the batch.
- **Per row:** the Identity Claim with status, confidence and argument; a minted handle for New; nothing for Skip.
- **Pins:** Go pins what the claim evidence offers. For a row whose sheet was never opened, Done pins the lines and relationships that start checked; for an opened row, the ones the researcher left checked (a pin the evidence no longer offers is dropped). A pin's Observations land on the new claim and on the claim of each member it names, so joining an entity with five named members pins all five. The evidence reaches **one hop through a bridge** ("her birth date matches" is pinned on the person's claim), and compares a neighbor only with the handle that neighbor is filed on, in this Done or earlier: this birth with that birth, never with another of the handle's events. A neighbor filed New or skipped pins nothing.
- **Duplicate pins are accepted on purpose.** The same agreement may be pinned on several claims (the person's and the birth event's). The machine drafts the pins, so this costs the researcher nothing.
- **Bridges** are filed by the rules in §9.1.

## 9.1 Bridge filing

A bridge (participation, relationship, location, place relationship) has **no identity of its own**: once both its ends are handles, which canonical association it belongs to follows from the ends. So filing is automatic everywhere, not only on the Promote page.

- **When:** after **any** claim create, every bridge Subject whose ends are now both handles is filed in the same transaction. On the Promote page and everywhere else, filing just happens. A bridge with an end that is unpromoted, skipped or simply not recorded stays unfiled on the graph, and files when that end is promoted later.
- **Which association it joins:** an existing association of the same kind between the same two handles, matched by its key; otherwise a new one is minted.

  | Bridge kind | Key | Notes |
  | --- | --- | --- |
  | Participation | person + event | The role is a reconciled value on the association. Sources that disagree on the role show it as mixed; a person who was really witness *and* informant shows both if `role` is multi-valued (S9-36). |
  | Location | event + place | |
  | Relationship | the two people + relationship type | The type is the identity: spouse *and* cousin are two relationships. Sources that disagree on the type (son vs stepson) make two relationships, not a mixed value. |
  | Place relationship (S9-38) | the two places + type (`part_of` vs `succeeded_by`) | |

- **Direction:** directed types ("A parent of B", "part of") match only in the same direction. Symmetric types (spouse) ignore order. Whether a type is directed is a property of the type (term or bridge kind), not a list in code.
- **Inverse terms:** a directed term can name its inverse (`property_terms.inverse_key`: parent ↔ child, grandparent ↔ grandchild, pibling ↔ nibling, guardian ↔ ward). Matching reads a term and its inverse under one key, the one that sorts first, with the ends swapped, so "Mary parent of John" and "John child of Mary" correspond. Filing still keys an association on the recorded term: joining the two readings onto one association needs the auto-reconciler to swap a member's ends too, which it does not yet.
- **A link from a handle to itself** (both ends on one handle: a duplicate on the graph, or a wrong match) is refused. On the page it's flagged before Done.
- **Refused filings don't fail the claim.** A place relationship that would close a hierarchy cycle (S9-38), or a self-link, leaves the bridge unfiled, with its reason visible. On the page, graph alignment flags it before Done, so the batch never fails on it.
- **An end that leaves later** (its member deleted) leaves the bridge's claim in place. The association loses that end's evidence, and the reconciler shows it. That's §5.2 review territory, not a special case.

---

# 10. Effect on the plan

Spike 9 was replanned around this on 2026-10-06. The close note is [`deployment-plan/archive/spike-9/README.md`](deployment-plan/archive/spike-9/README.md).

- **Events and Places come first** (slices 5–6): graph alignment needs the date module (S9-21), per-Property cardinality (S9-36), and the Event and Place headers for its dropdowns.
- **The canonical graph gets its own slice** (slice 7): the subject module and automatic bridge filing (S9-28), then the derived values that walk it (S9-31 / S9-32).
- **Place hierarchy comes before Promote** (slice 8): part-of and succession links (splits and amalgamations included) are filed as bridges from the start, so trying Promote on real research captures them, and graph alignment is tested with them as edges.
- **Promote graph alignment is one slice** (slice 9):
  - **S9-17**, reshaped from #265 / #266: `Compatible`, pins and backfill, the pinned-delete tests. The per-Subject comparison read and its UI plumbing are dropped.
  - **S9-41:** `Align`, a pure package composed over pairwise matching (§4.1), with a single config registry for walk/band tunables.
  - **S9-42:** the loader and the proposal read.
  - **S9-43:** the batch write.
  - **S9-44:** the page, gated by brief **S9-D16**.
- **Retired:**
  - S9-19 and S9-D11 (compare);
  - S9-29 (neighborhood read);
  - S9-30 and S9-D12 (walk);
  - S9-18, folded into S9-17 and S9-44.

  S9-D16 is a **rethink**: it replaces the S9-D9 / S9-D10 frames on the Promote board.
- **Later, learned weights:** start logging the suggestions shown and the decisions made, so m and u can be fitted from real accepts and rejects ([`ideas/archive/promote-matching.md`](ideas/archive/promote-matching.md#directions-we-circled-none-chosen)). This is out of scope for Spike 9.

---

# 11. Open questions

- **Thresholds and bands:** all live in the registry's alignment section (`core/matchregistry/alignment.go`, §4.1). The walk has no depth limit; the size knobs are `TopK` and `TieCap`, which bound both the shortlisted candidates and each placed pair's nominations. Initial values from the obituary golden; refine by dogfooding whole Sources, not by guessing in isolation.
- ~~**Places at several grains:**~~ **Decided 2026-10-07:** fold Locations that share a `part_of` graph into one chain; unrealted Places stay competing values and reconcile ([`conclusion-reconciliation.md`](conclusion-reconciliation.md) §11.1). No place-nature Property this spike.
- **Typed name parts:** a name entered as one form ("Gracie Gray Gates (Frickleton)") weakens name comparison. Is that a composer nudge, or a parsing step?
- **Naming:** what is the "map the rest of this graph" action called?
- **The decision log:** what to record, and where, so it stays local and private.

---

# 12. Implementation status

§5 and §6 describe the built walk (Spike 11, S11-16). What differs:

- **Query count (§7):** every read is batched (`sqlbatch`), and the canon comes from the in-memory graph, a warm proposal reads `MAX(revision)` only, and a warm Evidence request reads nothing from the catalog (`TestLargeSourceBudgets` `warm queries`). After a handle or vocabulary commit drops the shortlist index, the next proposal reads it again (one read of every person, event and place node). Cold, the graph loads one batch per wave of the walk.
- **Properties compared (§6):** only the lead properties of the registry's subject section, with their pools and periods, not every Property both sides carry.
- **Provenance (§6):** not applied. Every Subject's provenance is 1; credibility, transcription certainty, and member claim confidence don't scale comparisons yet.
- **Kinship shapes:** both loaders add the links `core/kinship` implies (spouse, parent, grandparent) through their one merge (`kinship.LinkSet`, whose implied links enter through `AddImplied`, or `PlaceImplied` in `loadCanon`, each recorded by `Implied.Recording`): the layer loader derives them, `loadCanon` reads them from the graph cache through `canonlinks.Reader.For`. Implied links are never filed and never shown as connections; a row reached through one says "via <neighbor>" with no relationship phrase. A link both sides imply through a paired event or middle parent counts once.
- **Bridge kinds in the loaders:** the layer loader lists its hops (participation, location, relationship, part of, succeeded by) in code. The canon follows every link the graph holds for a `connectrules` bridge, as pins and filing do.
