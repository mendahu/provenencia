# Source structure: what Spike 12 left for later

**Status:** idea only — not roadmapped. Spike 12 ([design](../deployment-plan/spike-12/source-structure.md)) solved large graphs and the "obituary or newspaper?" confusion with **Copy Source** and **multiple Evidence graphs per Source**. These options were weighed on 2026-10-18 and parked.

## Nested Sources

A `sources.parent_source_id`: publication → issue → obituary, census → household. The parent carries shared metadata and children inherit it.

- **Why parked:** you never cite the Medicine Hat News on its own; you always cite an item in an issue. Copy Source puts the publication details on each item without a new level. Inheritance (which fields, overrides, depth) was real design work with little payoff.
- **Revisit if:** fixing one publication typo across dozens of copied Sources becomes a real chore, or researchers want to browse "everything from this publication". A "replace this metadata value everywhere" tool or a Sources list filter by metadata value may be enough first.
- **If it is built:** `parent_source_id` means *physically part of*, never *copied from*.

## Derivation links

A FamilySearch index entry is a transcription of an image, which is a film frame of a parish register. That is derivation (copy of), not containment (part of).

- **Today:** one Source for the thing you grabbed; provider, collection, film, URL and accessed date go in its metadata. The original vs derivative judgment lives in the credibility argument.
- **Later:** a `source_derivations (source_id, derived_from_source_id, kind)` link (`image`, `transcription`, `abstract`, `index`, `translation`) for derivatives that are authored works in their own right (a book of abstracts, a published transcription), pointing at an original that may be a placeholder Source. This is a structural fact about where a record came from, not a cited Observation, which sets it apart from [`source-to-source-relationships.md`](source-to-source-relationships.md).
- **With it:** the GPS original / derivative / authored form as a first-class Source attribute, and Artifact-level access metadata (provider, film or item number, URL, accessed date).

## Named Source templates

"Stamp this with Medicine Hat News": a saved set of type and metadata values. Copy Source from a recent sibling covers it. Watch that a template does not quietly become the publication entity that nested Sources would have been.

## Per-graph credibility

Credibility stays per Source. A census household with a poor informant can only say so in a Citation note. Revisit if dogfooding shows it matters.

## Layered citation output

Formatting Evidence Explained style citations ("…; citing …; image, FamilySearch, accessed …") from Source metadata. Depends on the metadata being consistent, which seeded fields for obituary, census and record-collection types would help.
