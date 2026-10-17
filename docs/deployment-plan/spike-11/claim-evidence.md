# Claim Evidence

**Status:** requirements, being scoped (2026-10-17). Part of [Spike 11](README.md), after the matching algorithm is re-implemented ([`comparison-layers.md`](comparison-layers.md)). Today's model: [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 (assessment sheet) and §9 (pins).

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
- **Grouped by property, not by Subject** (decided 2026-10-17): a Name section, a Sex section, and so on, each listing every member's line. One-hop groups ("her birth's date") are their own sections, as today.
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

## Retires

- `pairExhibits`, `comparerDiscount` and `exhibitCompatible` (`core/database/promotealign/exhibits.go`).
- `attachExhibits` running inside every proposal read. `withRowLines` / `refLine` are replaced by the "why suggested" summary.

## Open questions

- **Large entities.** With fifty members, a property section is fifty lines. Is a per-section "select all" enough, or do members that share one value collapse into an expandable line?
- **Lines across properties that span properties.** Sex with gender, and an event's `date` against `start_date` / `end_date`, are compared as pairs in Subject comparison. Does the evidence view show them as one section?
