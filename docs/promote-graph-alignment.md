# Provenencia — Promote Graph Alignment

## Status

**Agreed 2026-10-06; scheduled as Spike 9 slice 9** (§10). Renamed from “Promote alignment” to **Promote graph alignment** (2026-10-07) so the name names the graph-to-graph proposal, not property-only matching. It replaced the per-Property compare checklist and the per-Subject walk; [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §5.3–§5.4 now state its model rules. How we got here, and the directions we turned down: [`ideas/promote-matching.md`](ideas/promote-matching.md). Matching today: [`matching.md`](matching.md).

**The data model doesn't change:** one Identity Claim per Subject and handle, with Observation pins on the claim and backfill (§5, §5.1).

---

# 1. The problem

Promote files Interpretation Subjects onto canonical handles. A Subject's identity is carried mostly by its **neighbors** (its birth event, its places, its family), not by its own Properties. An Evidence graph is also much bigger than one Subject: the reference case, an obituary, has 41 Subjects. Ticking agreements Property by Property, Subject by Subject, is the wrong amount of work. Most of those agreements the machine can decide, and none of them can see the structure that actually identifies people.

**The picture** ([`ideas/promote-matching.md`](ideas/promote-matching.md#the-stack-how-to-picture-the-problem)):
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
5. **Precision over recall.** A false merge is the costliest mistake. Weak and unreachable matches default to **Skip**, not to New.
6. **Explainable, deterministic, local.**
   - Every suggestion carries its reasons.
   - The same inputs give the same proposal.
   - Everything runs on device.

---

# 3. The flow: one page

Promote is one workspace page; the choose-target screen and the separate claim step go away.

- **Entry:** the Promote button on any unpromoted Subject card, as today. The page opens with that Subject's row, preselected to its best match, plus **"Map the rest of this graph (N Subjects)"**. Promoting one Subject is the same page with one row.
- **Rows:** one per primary Subject on the Evidence graph (person, event, place).
  - **Already promoted:** shown read-only, with no controls. They are anchors for the graph alignment, and editing them is a separate workflow (Spike 10).
  - **Unpromoted:** a target dropdown with the best match preselected, the next few alternatives, **New**, and **Skip**.
- **Assessment:** each row shows **strong**, **weak** or **no match**. Clicking it opens a **sheet** with every comparison that contributed: agree, conflict or unknown, with its weight. The agreeing comparisons are preselected as pins and can be toggled.
- **Claim fields per row:** status, confidence and argument. The argument can be drafted from the assessment.
- **Bridges aren't rows** (participation, relationship, location). A summary line reads "22 connections will be filed". It expands to a list where each bridge can be switched off, for a relationship the researcher doesn't accept from this Source. Filing rules: §9.1.
- **Possible duplicates:** two rows landing on the same existing handle, or two New rows whose properties clear the accept bar, get a warning ("these may be the same record; combine them on the Evidence graph"). A signature both rows have neighbors for suppresses the warning when none of those neighbors are the same subject or a match. There is no row-to-row option.
- **Done** writes the whole batch in one transaction (§9).
- **Leaving:** the leave guard covers accidental navigation. Drafts aren't persisted: re-opening re-proposes everything, and only manual changes are lost.

## 3.1 Suggested and decided rows

Every row is **suggested** (the machine filled it) or **decided**: the researcher changed its dropdown or toggled anything in its sheet.

- **On open,** the only fixed points are the clicked Subject's row and the already-promoted rows.
- **On any change,** that row becomes decided, and graph alignment re-runs over the whole page with every decided row held fixed. Suggested rows and their sheets are recomputed; decided rows are never touched. A row decided New or Skip is held too: it takes no handle and the walk doesn't pass through it.
- **Rows whose suggestion changed** get a brief "updated" mark with the reason ("because Gracie → PER-Z").
- **A decided row that the new context contradicts** keeps its choice and gets a warning. The researcher resolves it.

---

# 4. The graph alignment function

One pure function, called on open and after every decision:

```go
// Align proposes a handle, New or Skip for every primary Subject on one
// Evidence layer (graph alignment), holding the fixed pairs as given.
func Align(layer Layer, canon Canon, stats Stats, fixed []Fixed) Proposal
```

| Input | Holds |
| --- | --- |
| `layer` | One Evidence graph: Subjects, bridges, Observations with values and provenance |
| `canon` | A bounded piece of the canonical graph: handles, their adjacency, their members' Observations |
| `stats` | Catalog statistics for weighting: value frequencies, edge fan-outs (§6) |
| `fixed` | Decided rows, plus every already-promoted Subject on the layer, whose existing claim is a free anchor |

`Proposal` is one row per Subject:
- the target, plus the next 2–3 alternatives;
- the assessment;
- the comparisons, with drafted pins;
- the reasons;
- flags (conflicts with decided rows, possible duplicates).

**Pure and deterministic,** like the `autoreconcile` / `autoreconciler` split. A pure package (`graphalign` or similar) runs `Align`; a database package loads the inputs; ties break by id. The obituary is a golden test.

## 4.1 Abstraction, composition, and tunables

**Goal:** code that is well abstracted, strongly separated by concern, and highly composable. Graph alignment must not become a monolith that embeds sameness logic, catalog I/O, and UI drafting in one place.

**Compose smaller pure units upward:**

| Layer | Responsibility | Does not |
| --- | --- | --- |
| **Pairwise evaluation** | Given two sides’ values, return a comparer similarity per property. The promote scale turns that into agree / partial / conflict / unknown and points. `Compatible` is sameness for reconciliation and pins, not the score | Walk the graph, load the catalog, choose New/Skip for a whole layer |
| **Property-only matching** ([`matching.md`](matching.md)) | Rank one probe against many handles by looping the pairwise unit with a profile (`Rank`) | Know Evidence bridges, fixed anchors, or propagation |
| **Graph alignment** (`Align`) | Seed, priority walk, one-to-one, propagate structural support; call pairwise evaluation on each candidate; fall back to property-only matching when unreachable | Own a second “are these the same?” implementation; talk to SQLite |
| **Catalog adapter** (S9-42) | Build `Layer` / `Canon` / `Stats`, call `Align`, map the proposal to FFI | Contain scoring or walk policy |
| **Promote page** (S9-44) | Present rows, decisions, Done | Reimplement alignment |

Alignment **orchestrates**; matching / pairwise evaluation **judges a pair**. Alignment must not call only today’s bulk `Rank` API for every hop — that API is “score against the catalog.” It calls the **pairwise** unit (with edge terms when neighbors are already mapped). `Rank` and the promote score share `ComparerFor` and keep their own scales, so a menu point total is not a promote score.

**Package boundary:** graph alignment is **its own module**, not folded into Promote UI or FFI handlers. Config for the walk and scorer lives in **one registry file** in that pure package (same idea as matching’s `registry.go`): every tunable dogfood will tweak —

- accept / strong / weak / Skip score bands and the conflict penalty;
- cold-start priors for `m` (and fallbacks when `u` or fan-out is missing);
- propagation knobs (queue behaviour, how neighbor support accumulates, one-to-one refusal margins);
- expansion depth cap if the layer diameter is not enough on its own;
- any other weights that decide “is this the best candidate to promote?”

Pairwise / matching tunables stay in matching’s registry; alignment’s registry holds walk and band knobs. Do not scatter magic numbers through the walk.

Defaults ship for the obituary golden and table tests. Real tuning is **empirical**: file a batch of Sources, watch proposals, change the registry, re-run. Unit tests lock invariants and regressions; they cannot replace merging real graphs. Learned weights from accepts/rejects (§10) stay later; the registries are the hand-tuned front door until then.

`Align` takes an explicit config (or reads the package default) so tests can pin values without editing the registry file.

---

# 5. Walking: best-first propagation

This is collective entity resolution: graph alignment by propagation (cf. PARIS, similarity flooding). The walk is **best-first, not recursive**. A depth-first walk commits to whatever it reaches first; a priority queue lets the strongest matches settle first, so the weaker ones are decided with the most context.

1. **Seed.** For each fixed pair (s, H): look at each bridge between `s` and an unmapped neighbor `t` on the layer (**both directions** — the walk treats the Evidence graph as undirected; directedness matters only when filing, §9.1). For each corresponding edge from H to a handle G in `canon`, push the candidate (t, G).
2. **Pop the best-scoring candidate.** Accept it if it clears the threshold and keeps the graph alignment **one-to-one**: within one layer, a handle takes at most one Subject.
3. **Propagate.** The accepted pair becomes an anchor. Push its neighbors' candidates, and add **support** to queued candidates that are consistent with it. A candidate's score rises as more of its neighbors map consistently. That's the "the structure lines up" evidence.
4. **Repeat** until the queue is empty.
5. **Unreachable Subjects** (nothing reaches them from an anchor) fall back to property-only matching: `core/match` `Rank` picks their candidates, and each is scored by the same pairwise path the walk uses. They are assigned best-first across the layer, so a contest over one handle goes to the stronger match. Below the threshold, they default to Skip. Each pair this pass accepts then seeds the same walk as a fixed anchor, so a neighbor that never cleared `Rank` can still be nominated. A candidate `Rank` recorded but did not accept is not an anchor. The walk records a nominated handle even when its score stays under the accept bar; only an accepted pair propagates further.
6. **One-hop refinement.** With that mapping in hand, every recorded candidate is scored again. A layer neighbor adds edge support when its mapped handle is this candidate's canonical neighbor under the same signature: the fan-out bonus, plus a registry fraction of that neighbor's property score. Unfixed rows are reassigned from those scores. The pass repeats until the mapping stops changing, and at most once per subject. The fraction is of the neighbor's properties alone, so a city receives its parent's toponym points and does not receive the parent's credit from the country. The published band is this score: a toponym alone stays medium, and a corresponding low-fan-out `part_of` lifts it to strong. Decided rows are rescored afterward, so a neighbor that contradicts a decision can still raise a warning.

The work is about (Subjects + bridges) × candidates × log: well under a millisecond at obituary scale.

## 5.1 Layer shape: every bridge pair, and orphans

A layer is **one Source’s Subjects**, not one connected component. *Map the rest of this graph* lists every primary Subject on that Source, including stragglers.

**Every Evidence bridge is a pair** of primary ends (call them subject and object). Promotion status is independent on each end. Graph alignment and bridge filing together cover all four:

| Subject | Object | Who acts |
| --- | --- | --- |
| Unpromoted | Unpromoted | Alignment proposes both (seed/propagate once either end is fixed). Done may mint two handles, then file the bridge. |
| Promoted | Unpromoted | The promoted end is a **fixed anchor**; seed walks to the unpromoted end from that handle’s canonical neighbors. |
| Unpromoted | Promoted | Same mechanism, other direction (e.g. open Promote on a new person whose event neighbor was filed earlier). |
| Promoted | Promoted | Not an identity proposal for those rows. It is a **missing / unfiled relationship**: the Evidence bridge exists and both ends are handles, so §9.1 files the canonical association (shown on the page as a connection that will be filed; the researcher can switch it off). |

**Orphans and disconnected islands** are expected. A primary Subject with no bridges, or a cluster that shares no path with any fixed/promoted anchor, is **unreachable** (§5 step 5): property-only matching only; weak or no match defaults to Skip, not New. Structure-backed scoring never assumes a path from the entry card to every row.

---

# 6. Scoring

Nothing names a Property or a relationship. Every weight comes from metadata or from data.

**Node comparisons.** The score does not name a property. For each property both sides carry, `ComparerFor` returns a similarity (the same comparer `Rank` uses; `Rank`'s point total stays its own). A registry scale turns that similarity into points:

- Similarity at or above the scale's floor contributes `weight × similarity`. Similarity 1 agrees. A lower resemblance is partial.
- Similarity below the floor conflicts for a single-value property and is unknown for a multiple-value one (residences never conflict on a difference). A comparable zero is below every floor.
- Unknown (nothing to compare) is 0.

The scale is looked up by property, then by value type. A property added later is another entry.

- **Frequency scales** (text, terms, dates, integers, unless a property replaces them) derive the weight from catalog frequency: `u` is how often unrelated handles share the value, `m` is the prior that the same entity's records agree. A lone exact agreement is lifted to the weak bar, and any agreement bonus is kept. A toponym's bonus makes an exact toponym medium and leaves it short of strong until an edge adds support. A partial is that lifted weight times the similarity, so a one-letter toponym typo is weak.
- **Fixed scales** are not lifted. Name: floor 0.8, weight 2.5, contradiction 3. An exact surname plus a shared given name (James / James Kenneth, Lee / Lee-Ellen) is weak. An identical name is 2.5, still under medium. A shared surname with a different given name (about 0.55) conflicts. Sex at birth: floor 1, weight 0.6, contradiction 1. Agreement is a small nudge. A mismatch subtracts 1 and does not erase a name that cleared its floor.

`autoreconcile.Compatible` is sameness for reconciliation and for pins. A partial name can make the row weak without being pinned. A line is pinned only when the values are the same.

**Provenance scales each comparison:** Source credibility, transcription uncertainty, and the claim confidence of the member the evidence came from. These are the same inputs the auto-reconciler's provenance uses ([`conclusion-reconciliation.md`](conclusion-reconciliation.md) §4).

**Edge comparisons.**
- An edge's **signature** is (bridge type, role or relationship term, neighbor kind, the neighbor's own type term). An example: participation · subject · event · birth. Two edges correspond when their signatures match.
- Each signature's **fan-out** (how many neighbors a handle typically has through it) is measured from the catalog:
  - **≈1** ("subject at a birth"): a single correspondence is strong support;
  - **high** ("parent of", "subject at a residence"): it only narrows the candidates, and node comparison picks among them.
- A corresponding mapped neighbor contributes `Support(fan-out) + Credit(fan-out) × neighborPropertyScore`. Support is 1.5 when fan-out is low and 0.25 when it is high. Credit is 0.5 and 0.25 on that same split. `neighborPropertyScore` is the pairwise total for that pair. That pair's own link bonuses stay out, so credit stops at one hop. A negative total subtracts. A neighbor with nothing to compare adds only Support.

---

# 7. Loading: prefetch, then walk in memory

There are no database calls during the walk. Everything loads in a fixed number of batched queries:

1. **The layer:** one Source's Subjects, bridges, Observations and provenance, from the existing source-graph and comparison loaders.
2. **Seed handles:**
   - the fixed pairs;
   - the handles of already-promoted Subjects;
   - the top-k property-only candidates of each unpromoted Subject.
3. **Expand the canonical graph** breadth-first, with **one batched query per hop** (`WHERE entity_id IN (…)`), out to the layer's diameter (the obituary's is about 5). Each reached handle's member Observations load in batch, as `autoreconciler.load` does.
4. **`stats`:** value frequencies and signature fan-outs, computed lazily.

That's roughly 10–15 queries per proposal, whatever the graph size.

**Canonical adjacency without S9-28.** It can be derived through members: H's member m → bridge → neighbor n → n's accepted claim → G. That's a join, and it loads in batch. S9-28 caches it as canonical edges and adds bridge filing, so a prototype of graph alignment (`Align`) doesn't need to wait for it.

---

# 8. Caching and decisions

- **Decisions live in the page's draft:** the Swift row list (suggested or decided, target, sheet toggles, claim fields). They're sent as `fixed` on every call and never stored.
- **Stateless first:** every call loads and aligns from scratch. Tens of milliseconds is fine for an interactive page. The S9-33 deep-fixture timings will say if it isn't.
- **If it's slow,** keep the loaded snapshot in Go memory, keyed by (Source id, latest audit revision). The highest audit revision is a natural change stamp: any write invalidates the snapshot, and re-runs after decisions skip loading.
- **`stats`** is cached against the same stamp.

---

# 9. Writing

- **Done is one transaction for the whole batch:** all or nothing. Bridges need handles minted in the same batch, so a partial write could leave things half-linked. It's re-validated against the current state first: a write elsewhere since the proposal fails Done cleanly, and the page re-proposes. One audit revision covers the batch.
- **Per row:** the Identity Claim with status, confidence and argument; a minted handle for New; nothing for Skip.
- **Pins:** each toggled comparison pins its Observations on the row's claim and backfills them onto the member's claim (§5.1). The pair check widens from same Subject and same Property (S9-17) to **one-hop neighbors through a bridge**, so "her birth date matches" can be pinned on the person's claim. A one-hop pair compares a neighbor only with the handle that neighbor is filed on, in this Done or earlier: this birth with that birth, never with another of the handle's events. A neighbor filed New or skipped pins nothing.
- **Duplicate pins are accepted on purpose.** The same agreement may be pinned on several claims (the person's and the birth event's). The machine drafts the pins, so this costs the researcher nothing.
- **Bridges** are filed by the rules in §9.1; switched-off bridges are skipped. The switched-off set is the page's answer for the Source's unfiled bridges: it is recorded as declined (`subjects.filing_declined`), so a later Done or claim does not file them, and every other unfiled bridge is switched back on.

## 9.1 Bridge filing

A bridge (participation, relationship, location, place relationship) has **no identity of its own**: once both its ends are handles, which canonical association it belongs to follows from the ends. So filing is automatic everywhere, not only on the Promote page.

- **When:** after **any** claim create, every bridge Subject whose ends are now both handles is filed in the same transaction. On the Promote page the researcher can switch individual bridges off before Done. Anywhere else, filing just happens. A bridge with an end that is unpromoted, skipped or simply not recorded stays unfiled on the graph, and files when that end is promoted later.
- **Which association it joins:** an existing association of the same kind between the same two handles, matched by its key; otherwise a new one is minted.

  | Bridge kind | Key | Notes |
  | --- | --- | --- |
  | Participation | person + event | The role is a reconciled value on the association. Sources that disagree on the role show it as mixed; a person who was really witness *and* informant shows both if `role` is multi-valued (S9-36). |
  | Location | event + place | |
  | Relationship | the two people + relationship type | The type is the identity: spouse *and* cousin are two relationships. Sources that disagree on the type (son vs stepson) make two relationships, not a mixed value. |
  | Place relationship (S9-38) | the two places + type (`part_of` vs `succeeded_by`) | |

- **Direction:** directed types ("A parent of B", "part of") match only in the same direction. Symmetric types (spouse) ignore order. Whether a type is directed is a property of the type (term or bridge kind), not a list in code.
- **Inverse terms:** a directed term can name its inverse (`property_terms.inverse_key`: parent ↔ child, grandparent ↔ grandchild, pibling ↔ nibling, guardian ↔ ward). Matching reads a term and its inverse under one key, the one that sorts first, with the ends swapped, so "Mary parent of John" and "John child of Mary" correspond and share one fan-out. Filing still keys an association on the recorded term: joining the two readings onto one association needs the auto-reconciler to swap a member's ends too, which it does not yet.
- **A link from a handle to itself** (both ends on one handle: a duplicate on the graph, or a wrong match) is refused. On the page it's flagged before Done.
- **Refused filings don't fail the claim.** A place relationship that would close a hierarchy cycle (S9-38), or a self-link, leaves the bridge unfiled, with its reason visible. On the page, graph alignment flags it before Done, so the batch never fails on it.
- **An end that leaves later** (its member deleted) leaves the bridge's claim in place. The association loses that end's evidence, and the reconciler shows it. That's §5.2 review territory, not a special case.

---

# 10. Effect on the plan

Spike 9 was replanned around this on 2026-10-06: [`deployment-plan/spike-9/deployment-plan.md`](deployment-plan/spike-9/deployment-plan.md), slices 5–10.

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
- **Later, learned weights:** start logging the suggestions shown and the decisions made, so m and u can be fitted from real accepts and rejects ([`ideas/promote-matching.md`](ideas/promote-matching.md#directions-we-circled-none-chosen)). This is out of scope for Spike 9.

---

# 11. Open questions

- **Thresholds and bands / cold-start priors / expansion depth:** all live in the graph-alignment config registry (§4.1). Initial values from the obituary golden; refine by dogfooding whole Sources, not by guessing in isolation.
- ~~**Places at several grains:**~~ **Decided 2026-10-07:** fold Locations that share a `part_of` graph into one chain; unrealted Places stay competing values and reconcile ([`conclusion-reconciliation.md`](conclusion-reconciliation.md) §11.1). No place-nature Property this spike.
- **Typed name parts:** a name entered as one form ("Gracie Gray Gates (Frickleton)") weakens name comparison. Is that a composer nudge, or a parsing step?
- **Naming:** what is the "map the rest of this graph" action called?
- **The decision log:** what to record, and where, so it stays local and private.

---

# 12. Implementation status

As built at the end of Spike 9 slice 9 (S9-41 – S9-44 and the review fixes stacked on them). Read this before assuming a section above is in the code.

**Built as designed**

- One pure `core/graphalign.Align` over `core/match.Evaluate`, with every walk and band tunable in `core/graphalign/registry.go`. Deterministic: subjects, the queue, and candidate lists all break ties by ref, then id.
- Best-first walk, one-to-one, propagation: an accepted anchor re-scores the candidates it reaches, so support accumulates as neighbors map. Edge support counts each corresponding neighbor once.
- Directed terms (parent, part of, …) correspond only from the same end. Whether a term is directed is the catalog's `property_terms.directed`, not a list in code.
- Fan-out is measured per bridge from `connectrules`, keyed exactly like the edge signatures. Stats are cached per catalog file and audit revision.
- Decided rows (handle, New, Skip) are held. Warnings: a held handle its neighbors contradict (`ConflictWithFixed`); a row that lost a handle to another row; two New rows that match each other.
- Rows carry reason codes (`via`, `decided`, `agrees`, `weak`, `taken`, `no_match`, `empty`); the page words them.
- Done is one transaction with the stale check, one-hop pins (§9), and bridge filing with switch-offs (§9.1).

**Differs from the text above**

- **Query count (§7):** every read is batched, so a proposal's query count doesn't grow with the layer or catalog. It is about 60 queries, not 10–15: the expansion runs one walk per bridge direction per hop (10 × up to 5 hops: both directions of every bridge). Rows are bounded separately: a handle follows at most 50 associations per step and the frontier stops at 3,000 handles (`maxEdgesPerStep`, `maxCanonHandles`), so a hub place or event can't pull in the catalog. The kind-wide candidate scan that seeds the canon is cached per catalog revision, like the stats, and the page waits 250 ms after a decision so quick decisions share one proposal.
- **Properties compared (§6):** only the Properties in each kind's `core/match` default profile (name and sex; event type and dates; toponym), not every Property both sides carry. Value frequencies (`u`) cover name, sex, event type, and toponym, and structured names have no frequency key yet, so names score on the cold-start prior.
- **Provenance (§6):** not applied. Every Subject's provenance is 1; credibility, transcription certainty, and member claim confidence don't scale comparisons yet.
- **Bridge kinds in the loaders:** the layer and canon loaders list their hops (participation, location, relationship, part of, succeeded by) in code. Pins and filing read bridge kinds from `connectrules`.
- **Fallback islands:** an accepted property match seeds the same walk as a fixed anchor. A `Rank` loser does not. One-hop refinement still does not add handles; it only rescores candidates the walk or `Rank` already recorded.
- **Exhibit values:** a date travels as a structured value. The Mac client formats it with `DateValueDisplay` for the user's locale. The display string is the portable fallback (`Before 2001-03-31`).
