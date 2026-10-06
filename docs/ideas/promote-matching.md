# Promote matching: what proves a join

**Status:** open problem, unsolved. Written 2026-10-05 while planning the Promote compare step (S9-D11 / S9-19). It sits on top of S9-17 (pins + backfill, PR #265) and S9-18 (pinned deletes, PR #266), which are on hold until this is settled. This note sets out the scope only. Nothing here is decided.

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
- What happens to #265 / #266 (pins, backfill, pin release) and to briefs S9-D11 / S9-D12?
