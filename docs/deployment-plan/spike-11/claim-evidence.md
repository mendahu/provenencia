# Claim Evidence

**Status:** requirements; the query is built in S11-19 (`core/claimevidence`, `promotealign.Evidence`), the RPC and Go-owned pins are built in S11-25 (2026-10-21), the likeness lines in S11-32 (its lifespan line was withdrawn in S11-20 ship round 11), and the sheet in S11-20 (2026-10-22), audited in S11-33 (2026-10-23). Part of [Spike 11](README.md), after the matching algorithm is re-implemented ([`comparison-layers.md`](comparison-layers.md)). Today's model: [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 (assessment sheet) and §9 (pins).

## Why

After matching, each Promote row proposes an Identity Claim: this Subject belongs with this entity. The row's sheet is where the researcher pins the evidence for that claim. It falls short in two ways.

- **It only shows one member.** `pairExhibits` (`core/database/promotealign/exhibits.go`) builds one line per incoming value. It pairs that value with the first member Observation that is the same value, else the single closest one. Members load in claim order, so the line pairs with whichever member joined first. Joining an entity with five members who all have names shows one name line, not five. Most of the evidence that could back the claim is never shown or pinned, and the gap grows as entities grow.
- **The wording doesn't say what's happening.** The researcher ticks rows without being told what a tick means.

## The question the view answers

**"What's the evidence that this record is <entity>?"** For each property, it compares the incoming Subject's values with every member's values and lets the researcher pin the agreements.

This is a different question from **"Why was this suggested?"**, which is matching's output (band, links that lined up, contested marker). The sheet keeps the two apart: a short summary of why it was suggested, then the evidence to pin.

## Requirements

### Separate from matching

- **Its own query, run lazily** when the sheet opens, for that one row (decided 2026-10-17). It reads the in-memory catalog graph (`graphcache`): the target entity's members, their Observations, and for one-hop lines the neighbors' targets' members. Nothing in the walk gathers evidence along the way, and re-proposing the page no longer pays for every row's sheet.
- **Its inputs come from the page, not from matching internals:** the row's target handle and the targets of its layer neighbors, as the page's draft holds them, suggested or decided. The view stays right after the researcher changes a row.
- **Comparisons use the comparison layers.** Each line is value compare + property compare on one incoming Observation and one member's Observation. No second comparer, and no use of `autoreconcile.Compatible` or Promote's scores.
- **Pinning policy belongs to this view** (comparison layers rule 7). The layers say how two values relate; this view decides what can be pinned and what starts checked.

### Lines

- **One line per member per property.** Five members with names give five name lines.
- **Grouped by property, not by Subject** (decided 2026-10-17): a Name section, a Sex section, and so on. One-hop groups are their own sections (§ Evidence through relationships).
- **Members with an identical value collapse into one line** (decided 2026-10-17): the line shows "Show 12 records" in its Source column, which expands to each member and its Source. Checking the line pins all of them; expanding lets the researcher uncheck one.
- **No "select all".** Lines that match start checked; the researcher changes from there.
- **Pairs compared together show as one section** (decided 2026-10-17): sex at birth with gender, and an event's `date` with `start_date` / `end_date`.
- **What starts checked** (decided 2026-10-17):

  | Relation | Shown | Starts checked |
  |---|---|---|
  | `Same`, `Equivalent` | yes | yes |
  | `Contains`, `Overlaps`, `Near` | yes | **no** |
  | `Different`, or a negative that contradicts | yes, as a disagreement | **no** |
  | `Unknown`, cause `unjudged` (two different likeness images) | yes, to judge by eye | **no** |
  | `Unknown`, any other cause | no | — |

  Every line, member and relationship can be pinned (S11-20 round 16, from the board's notes: a conflict starts unchecked and can still be pinned). Only the default check depends on the relation.

- **Likenesses are judged by eye** (S11-32). Two likeness values are `Same` (`exact`) when they are the same image File and `Unknown` (`unjudged`) otherwise; the app never calls two different images the same face. `evidence.ByEye` lists the causes of an unknown line that is kept, never checked. The line carries the stored crop's path, not its region. Likeness has no weight in matching, so these lines never move a band.
- **Sections follow the catalog's binding order** for the kind (`subject_type_properties.sort_order`, cached with the vocabulary): the person's `name`, `likeness`, `sex_at_birth`, `gender`, `interests` as seeded, then any property with no binding, by key. Reordering a type's bindings (`ReorderSubjectTypeProperties`) shows on the next open.
- **The lifespan is a matching signal only.** It ranks suggestions (`subjectprofile`, `subjectcompare`) and is not a sheet line: the evidence carries no lifespan, on the wire or in the app (removed in S11-20 ship round 11, at the user's request; S11-32 had carried it).
- **Each line names the member's Source,** so the researcher sees which record agrees.

### Wording

The copy is overhauled so the screen says what it is for. Lines are sentences built from value compare's relation and `Detail.Cause`, never from a number. Settled in [S11-D2](design/archive/S11-D2-evidence-sheet.md) (the copy table, ES-8) and built in S11-20 as `L10n.Promote.lineSentence`, which is the one place the sentences are written; this document keeps no English copy of them. A relationship is named by the connection that matched, read from the neighbor's side as the engine reports it: its Card header is "…they share <phrase> (<refs>)", where the phrase is the neighbor's part in the link ("a spouse", "a birth event", "a place within it") and the refs are the connected row's and its record's. It is not named by its neighbor ("Through <name>"). Every sentence is one whole translatable phrase, with no translated word spliced into another (S11-33).

**Requirement on value compare:** its cause codes must be rich enough to word every line without reading the similarity number.

### Writing pins

- A pinned line pins the incoming Observation on the new claim, and backfills that member's Observation onto the member's claim (today's rule, §9).
- One incoming Observation can now pair with many member Observations. Verify the pin tables allow that, and that one incoming pin is written once however many lines share it.

### UI process

New chrome and copy: a Claude Design brief ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md)) gates the PR, and all copy goes through L10n ([`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md)).

### Evidence through relationships

**Today** (`attachExhibits`, `bridgeNeighbors`):
- **Every bridge counts, on both sides, with no key relationships.** On the layer, every bridge the loader reads (participation, location, relationship, `part_of`, `succeeded_by`) gives a neighbor group. On the entity side, any bridge of any type between the handle's members and the neighbor's handle's members qualifies.
- **The bridge type is never checked.** A spouse on the Source can pair with a member who is that person's *sibling* in another Source. Matching checks that (`corresponds`); the evidence view doesn't.
- **The lines compare only the neighbor's own values** (her husband's name), never the relationship itself.
- **The heading was "Through <name>"** (`promote.sheet.through`, gone in S11-20): no relationship, and no reason it matters.

**Decided (2026-10-17): the relationship is abstracted; the neighbor's properties are shown.**

The researcher wants to say "these are the same person because they have a matching birth event", then check the evidence that the births really match. The plumbing that connects them is handled for them.

- **Use segments.** A relationship section exists only when the Source segment (this Subject – bridge – neighbor) corresponds to a member segment (member – bridge – neighbor's member). Correspondence comes from segment comparison (layer 5): same bridge kind, role or term, and direction, with inverse terms (parent ↔ child). A spouse only meets a spouse. Every corresponding relationship counts, ordered core first (birth, death, parents and children, spouses), then the rest.
- **The relationship is its own checkable line,** and it says what matched and why it counts: "They have a matching birth event", "They have a matching spouse: Robert Gates". Checking it pins the bundle below. It starts checked when the segment matches, and is unchecked when the segment is partial. It works even when there is nothing under it to check: an event's place link, or two births with no dates (decided 2026-10-17).
- **Abstracted (bundled, not shown):** everything that only connects the two, on both sides:
  - the bridge's two end Observations (who, and which event or person);
  - everything on the bridge itself: its role, its relationship type, any other bridge values;
  - the neighbor's **defining properties**, declared per kind in the registry. For an event that is `event_type`: "birth" is what the heading already says.
- **Shown as checkable lines under it,** each with its own checkbox: the neighbor's other properties, exactly like the Subject's own property lines. Under "They have a matching birth event": the birth date lines. Under "They have a matching spouse": Robert's name, sex, and so on. Same rules as property lines: grouped by property, identical member values collapse into one line, matches start checked, partial matches and disagreements start unchecked, and every line can be pinned.
  - **Which members (decided 2026-10-27, ships in S11-52).** An event neighbor's lines compare with the members of the Source that records the link: this birth with that birth only (S11-19 D-13). A person neighbor's lines compare with all of that entity's accepted members, as the row's own lines do: a spouse is one person across many Sources, so a Source's "Jake Robins" meets every Source that calls him Jake, not only the marriage certificate that calls him James Kenneth Robins. Which kinds widen is a registry entry.
- **Registry rule, not code:** "a segment's bundle is all of its bridge's Observations plus the neighbor kind's defining properties". Event's defining property is `event_type`; adding one for another kind is one entry.

**Implied links (S11-40).** The entity side reads a person's links through `canonlinks.Reader.Relationships`, so a tree marriage or birth gives a relationship line to a Source that records the fact directly, and the other way round; each line says how each side recorded the fact (`Relationship.Layer`, `.Canon`: the zero `kinship.Shape` is direct, and `Recording.Shown` is the one reading of it).

- **One bundle per Source's recording.** A bridge behind an implied link reaches the pivot (the event, or the parent between) and the row's member or the neighbor's member. The bridges that reach one pivot make one bundle: the row member's bridge first, then the others, and the pivot's defining Observations. A recorded link's bundle is one bridge, as before. The bundle's pin id is keyed by its first bridge.
- **A pivot with no bridge of a row member makes no bundle.** A Source that records only the far part of the link (a parent of the neighbor, not of the row's member) pins nothing: a pin needs a member to land on.
- **Counted once.** When both sides record a fact through the same pattern and the request pairs their pivots, the line is dropped, as matching drops the credit (`kinship.CountedOnce`); a participation line to the paired event stays.
- **The Card says how each side recorded it (S11-42).** A muted line under the Card's header names both sides, "This source" and "The record": an implied side by its pattern and pivot (the marriage by its date, a parent by name), a direct side by the link's term. A Card whose sides are both direct has no line. The words are in `PromoteEvidenceDisplay.recorded`; a new pattern adds one arm to its `wording` switch and two keys.

## Retires

Done in S11-25:

- `pairExhibits`, `comparerDiscount` and `exhibitCompatible` (`core/database/promotealign/exhibits.go`).
- `attachExhibits` running inside every proposal read. `withRowLines` / `refLine` are replaced by the "why suggested" summary (S11-20).

## Open questions

- **Where a bundle's pins land.** On the person's claim (with backfill onto each member's claim), as today. The neighbor event gets its own claim and pins from its own row; duplicates across claims are accepted on purpose (§9). Answered for the schema by S11-19: `identity_claim_evidence` is keyed `(identity_claim_id, observation_id)` and stores no incoming-to-member pairing, so one incoming Observation pins against many members' Observations and no migration is needed.

