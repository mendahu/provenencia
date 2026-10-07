# Promote matching: what proves a join

**Status:** direction chosen 2026-10-06; see [`promote-graph-alignment.md`](../promote-graph-alignment.md) for the flow and the graph alignment function. This note keeps the problem and the directions we explored to get there. Written 2026-10-05 while planning the Promote compare step (S9-D11 / S9-19). The data model doesn't change, so S9-17 (pins + backfill, PR #265) and S9-18 (pinned deletes, PR #266) stand.

Model rules today: [`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §5 (exhibit pins, backfill, review), §5.3 (compare), §5.4 (walk). Matching today: [`matching.md`](../matching.md).

## The current design, and why it isn't right

§5.3 describes the compare step as a **checklist**. For each Property, line the incoming Subject's Observations up against every member's. The researcher ticks the pairs that agree, and each ticked pair is pinned on both claims (backfill, §5.1). S9-17 built the engine for exactly that.

Working through real cases shows three problems.

1. **Identity is mostly carried by neighbors, not by the Subject's own Properties.** A person card usually holds a name and little else. Birth and death dates sit on *event* Subjects, two hops away (person → participation → event). Places are toponyms on place Subjects, linked through location bridges. "This is the same Grace because she was born 1 May 1901 in Polemont, Scotland" is evidence about the Subject's neighborhood. A per-Property checklist on the Subject can't see it, and S9-17's pair check (same Subject, same Property) can't store it.
2. **Pins duplicate.** If neighbors count, the same agreement is evidence for several claims at once. The matching birth date supports both the person's claim and the birth event's claim. With backfill, each confirmation becomes up to 2 × (claims that can see it) rows. In a large graph almost every Observation is evidence for several Subjects, so pins grow with Observations × claims, while the researcher made one confirmation. Backfill only exists because the evidence hangs off claims.
3. **The checklist is the wrong amount of work.** Most of what the researcher would tick is something the machine can already decide: same name parts, the same date window, the same toponym. Asking for a tick on each adds friction, not judgment. The judgment that matters is "is this the same person?", not "do these two strings agree?".

## Scale: the obituary

The dogfood case is Grace Gray Gates (Frickleton)'s obituary, *Medicine Hat News*: **41 Subjects**, about 70 Observations, on one Evidence graph.

| Kind | Count | What identifies them on this graph |
| --- | --- | --- |
| Persons | 5: Gracie, two children (Bill Davies, Marion Robins), their spouses (Aida, Doug) | A name only. Gracie also reaches birth, emigration, death and burial events. |
| Events | 6: Birth (1 May 1901), Emigration (1925), Death (27 Dec 1990), Burial (after 27 Dec 1990), two Residences (before 27 Dec 1990) | Type + date |
| Places | 8: Polemont, Scotland, Canada, Edmonton, Alberta, Mountain Park, Mountain Park Cemetery, Medicine Hat | A toponym only |
| Bridges | ~22: participations, relationships (parent, spouse), locations | A role or relationship type |

Promoting Gracie onto a tree that already has her means placing about 19 primaries and 22 bridges, nearly all of them by **where they hang off Gracie**, not by their own fields. A per-Subject walk (§5.4) with a per-Property checklist at each step is dozens of screens of ticking.

What the graph also shows:
- **Names come as one form.** Gracie's is "Gracie Gray Gates (Frickleton)", with the maiden name in parentheses. Matching her to "Grace Gates" needs typed parts.
- **Places come at several grains, not as a hierarchy.** Birth links to both Polemont *and* Scotland through separate location bridges, and Death to both Edmonton and Alberta. Nothing records that Polemont is in Scotland.
- **Dates carry qualifiers.** "Before" and "after 27 Dec 1990" are windows, not points.

## The stack: how to picture the problem

Picture every Evidence graph as a **layer**, and the layers stacked so that Subjects about the same historical thing line up **vertically**.

- **Horizontal** is within one Source: bridges inside one Evidence graph (Gracie → participation → Birth).
- **Vertical** is across Sources: a **column** of Subjects from different layers that are the same person, event or place. A canonical entity is a column, and an Identity Claim is one vertical link.
- **The canonical graph is the stack seen from above:** columns collapsed into handles, and horizontal edges reconciled into canonical edges (the subject module, S9-28).
- **Traversal goes both ways:** horizontally to stay inside a Source, vertically to hop to the same entity in another Source.
- **Promote lays a new layer onto the stack:** it decides which of the layer's Subjects drop into existing columns and which start new ones.

How a new layer meets the stack varies:
- **Large overlap:** most of the layer lines up with an existing region, with a few branches that are new information.
- **A single connecting node:** one Subject lines up, and everything else is new.
- **No overlap:** the layer starts its own columns.

What the picture makes clear:
- **A vertical link's strongest evidence is the horizontal structure around it.** Gracie belongs in column PER-X partly because the names agree. Mostly, though, it's because her neighbors line up too: her Birth in PER-X's birth column, her daughter in PER-X's child's column, their dates and places agreeing. Vertical links in one layer **support each other** through the horizontal edges between them.
- **That's why pins duplicated.** One matching pair of edges (her Birth here, PER-X's Birth there) supports two vertical links, Gracie's and the Birth's. The evidence is a **correspondence between layers**, not a list of Observations on a claim.
- **Overlap measures confidence.** A large, self-consistent overlap is a strong alignment. A single connecting node, with no structure around it, is the weak case, and that's where the researcher's attention belongs.
- **New branches are placed by position.** A Subject with no column of its own is identified by its horizontal link to one that has a column ("the spouse of Marion, who is in PER-Y"). That link is both the genealogical linkage and the reason it's filed where it is.
- **It's how genealogists already argue.** The Genealogical Proof Standard's *correlation of evidence* is this stacking: identity established by consistent agreement across independent sources, not by one matching field.

**A candidate unit of evidence (not decided): the alignment.**
- **Decisions stay per Subject:** one vertical link, one claim, each rejectable alone.
- **The evidence is the alignment of the layer against the stack,** recorded once per Promote: which edge pairs line up, which property pairs agree, which conflict.
- **A claim's support is the part of that alignment touching it,** worked out rather than copied. Nothing is duplicated, there's no backfill, and "why is Gracie here?" has a structural answer: "this layer lines up with PER-X through her birth, her death and two children".

**The goal for the interaction:** a Promote that is **mostly automated, with a few fine-tuning actions**. The app proposes the alignment. The researcher accepts the well-matched regions together, then fixes the few links that need judgment (a single connecting node, a conflict, an ambiguous child) by moving a Subject onto a different column or into a new one. The UI for this is still open.

## What any answer has to keep

- **Membership stays per Subject.** One Identity Claim per Subject and handle, so a wrong match can be rejected alone. An obituary mixes people you file, people you skip, and people who belong to different handles.
- **Create-only Promote** with a Done off-ramp. Editing claims is Spike 10.
- **A chain of evidence a later reader can follow.** "Why was this Subject filed here?" has to have a machine-readable answer, not just prose in `argument`.
- **Stable history.** Evidence recorded at decision time must not quietly change when the reconciler is tuned or a value is concluded. Changes since then should be visible: that's the §5.2 review alert.
- **No circularity.** A member's own records can't vouch for its membership. Agreement among members is what a wrong merge looks like too.
- **Precision over recall.** A false merge is the costliest mistake. A missed suggestion is cheap.
- **Explainable and local.** Reasons per feature, computed on device.

## Directions we circled (none chosen)

- **Matches, not pins.** Store one row per confirmed pair (`observation_matches`) and *derive* which claims it supports: those whose Subject, or a one-hop neighbor, holds one side. That removes the duplication and backfill. Open: should support be derived or chosen?
- **No stored evidence; corroboration as a read.** A claim is corroborated when its Subject's (or a neighbor's) record is `autoreconcile.Compatible` with another member's from a different Source. This needs no storage, but it loses the researcher's intent, and it shouldn't use the reconciler's outcomes, which shift with tuning.
- **Automatic scoring.** Score each candidate on agree / conflict / unknown per feature. Freeze the result on the claim at promote time, keep a live re-score, and treat drift between them as the review alert. Rejected for now: bare percentages, because commonness ("John Smith") matters more than coverage.
- **Key properties per kind,** instead of everything. Person: name, birth, death (+ age → birth window, parents/spouse). Place: toponym, containment. Event: type, date, then participants. An "identifying" flag on a Property would add researcher-specific facts. This fits the `core/match` profiles, which already work this way.
- **Anchor and propagate.** Score only the starting matches (Gracie). Place the rest by position:
  - a role that allows one value (her birth, her death) → the anchor's existing neighbor unless it conflicts;
  - a role that allows several (children, residences) → narrowed to the anchor's neighbors, then picked by name;
  - otherwise → new.

  One review screen for the whole graph instead of a step per Subject. Bridges are filed once both ends are handles.
- **Learned weights, later.** Probabilistic linkage (Fellegi–Sunter): u from the catalog's own value frequencies, m from priors, refined from the researcher's accepts and rejects. That needs a log of the suggestions shown and the decisions made, which nothing records today.

## Open questions

- What is the unit of evidence: a pin per claim, a match, a frozen score, or nothing stored?
- Does the researcher confirm individual agreements, or only the overall match?
- How far do neighbors count: one hop through bridges, or anything an anchor reaches?
- Is Promote one Subject at a time (§5.4 walk), or one review for the whole graph from an anchor?
- Events and persons each help identify the other. What gets matched first?
- What does laying a layer onto the stack look like on screen? Where does the proposed alignment show, and what are the fine-tuning actions (move to another column, start a new one, detach)?
- Is the alignment stored per Promote, and how does a later Promote of an overlapping layer relate to it?
- What happens to #265 / #266 (pins, backfill, pin release) and to briefs S9-D11 / S9-D12?
