# Claim Evidence

**Status:** requirements; the query is built in S11-19 (`core/claimevidence`, `promotealign.Evidence`), the RPC is S11-25 (2026-10-21). Part of [Spike 11](README.md), after the matching algorithm is re-implemented ([`comparison-layers.md`](comparison-layers.md)). Today's model: [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 (assessment sheet) and §9 (pins).

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
- **Members with an identical value collapse into one line** (decided 2026-10-17): "The names match · 12 records", expandable to each member and its Source. Checking the line pins all of them; expanding lets the researcher uncheck one.
- **No "select all".** Lines that match start checked; the researcher changes from there.
- **Pairs compared together show as one section** (decided 2026-10-17): sex at birth with gender, and an event's `date` with `start_date` / `end_date`.
- **What starts checked** (decided 2026-10-17):

  | Relation | Shown | Pinnable | Starts checked |
  |---|---|---|---|
  | `Same`, `Equivalent` | yes | yes | yes |
  | `Contains`, `Overlaps`, `Near` | yes | yes | **no** |
  | `Different`, or a negative that contradicts | yes, as a disagreement | no | — |
  | `Unknown` | no | — | — |

- **Each line names the member's Source,** so the researcher sees which record agrees.

### Wording

The copy is overhauled so the screen says what it is for. Lines are sentences built from value compare's relation and `Detail.Cause`, never from a number. Draft wording, to be settled in the design brief:

| Relation and cause | Line reads |
|---|---|
| `Same` | "The names match" / "The dates match" |
| `Equivalent`, `fold` | "Same name, different spelling" |
| `Equivalent`, `alias` | "Same name (nickname)" |
| `Contains`, `precision` | "Same date, recorded less precisely" |
| `Contains`, `missing_part` | "The name matches; a given name is missing" |
| `Near` | "Similar name" / "Dates are close" |
| `Different` | "The names differ" |

**Requirement on value compare:** its cause codes must be rich enough to word every line in this table without reading the similarity number.

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
- **The heading is "Through <name>"** (`promote.sheet.through`): no relationship, and no reason it matters.

**Decided (2026-10-17): the relationship is abstracted; the neighbor's properties are shown.**

The researcher wants to say "these are the same person because they have a matching birth event", then check the evidence that the births really match. The plumbing that connects them is handled for them.

- **Use segments.** A relationship section exists only when the Source segment (this Subject – bridge – neighbor) corresponds to a member segment (member – bridge – neighbor's member). Correspondence comes from segment comparison (layer 5): same bridge kind, role or term, and direction, with inverse terms (parent ↔ child). A spouse only meets a spouse. Every corresponding relationship counts, ordered core first (birth, death, parents and children, spouses), then the rest.
- **The relationship is its own checkable line,** and it says what matched and why it counts: "They have a matching birth event", "They have a matching spouse: Robert Gates". Checking it pins the bundle below. It starts checked when the segment matches, and is pinnable but unchecked when the segment is partial. It works even when there is nothing under it to check: an event's place link, or two births with no dates (decided 2026-10-17).
- **Abstracted (bundled, not shown):** everything that only connects the two, on both sides:
  - the bridge's two end Observations (who, and which event or person);
  - everything on the bridge itself: its role, its relationship type, any other bridge values;
  - the neighbor's **defining properties**, declared per kind in the registry. For an event that is `event_type`: "birth" is what the heading already says.
- **Shown as checkable lines under it,** each with its own checkbox: the neighbor's other properties, exactly like the Subject's own property lines. Under "They have a matching birth event": the birth date lines. Under "They have a matching spouse": Robert's name, sex, and so on. Same rules as property lines: grouped by property, identical member values collapse into one line, matches start checked, partial matches pinnable and unchecked, disagreements shown and not pinnable.
- **Registry rule, not code:** "a segment's bundle is all of its bridge's Observations plus the neighbor kind's defining properties". Event's defining property is `event_type`; adding one for another kind is one entry.

## Retires

- `pairExhibits`, `comparerDiscount` and `exhibitCompatible` (`core/database/promotealign/exhibits.go`).
- `attachExhibits` running inside every proposal read. `withRowLines` / `refLine` are replaced by the "why suggested" summary.

## Open questions

- **Where a bundle's pins land.** On the person's claim (with backfill onto each member's claim), as today. The neighbor event gets its own claim and pins from its own row; duplicates across claims are accepted on purpose (§9). Answered for the schema by S11-19: `identity_claim_evidence` is keyed `(identity_claim_id, observation_id)` and stores no incoming-to-member pairing, so one incoming Observation pins against many members' Observations and no migration is needed.

