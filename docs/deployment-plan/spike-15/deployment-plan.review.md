# Review — Spike 15 PR map

**Verdict:** Fits with changes (after round 2; slice 8 reviewed in round 3; two To user items open). **Rounds:** 3 (round 3 covers slice 8 only, added after round 2). **Reviewers:** maintainability (Opus, fresh context each round, map review per [`review-plan`](../../../.agents/skills/review-plan/SKILL.md)).

## To user

- ~~K-2 · S15-26~~: kept (user, 2026-11-09).
- ~~K-18 · Format edits and the Person page reads~~: closed by the user's call (2026-11-09) to render citations on demand from SQL with nothing cached; see rule "Citations render on demand, from SQL", which replaces the round-1 answer to K-3. To be measured at Scale in S15-16.
- ~~K-16 · `{citation.item}`~~: the item is a graph-level metadata field, `item`, typed once per graph; no per-Citation derivation and no primary-subject toggle in this spike (user, 2026-11-09). Rule "The item is a graph field"; map § 6.5.

- **K-25 · Preparatory PR S15-30** (the paged fold moves to the kit). Added after S16-07 so Spike 16's S16-07 keeps its scope; the alternative is to widen S16-07 to build the fold in `Platform/` and `PVShowMoreRow` directly. A preparatory PR is the user's call. Recommendation: keep S15-30.
- **K-30 · The deprecated Why accordion.** After S15-28 the detail read carries no records, and the Why on the Event and Place pages reads them. Options: (a) a PR before S15-29 removes the Why and moves Event and Place rows to the provenance sheet, as the Person page does (a visible change to two pages, so briefs); (b) the map states what Event and Place rows show until the Why is removed elsewhere. Recommendation: (a).

## Writer's own calls

- **K-4:** `CitationLineDisplay` stays in Swift as the fallback for a type with no format and for its other callers (Gallery captions, the likeness viewer, a denier's Source name), because its page words are localized there. The engine returns no text when a type has no format.
- **K-14:** Spike 15's new Swift query keys go into today's `CatalogQueryKey` enum, and S16-25 converts them with the rest, instead of waiting for S16-25, so slices 1 to 4 run alongside Spike 16 as the user asked.
- **K-15:** Scale additions (3,000 Sources from Spike 16; 60 Artifacts per Source; 100 Subjects per move bundle; 500 Sources per type) are the writer's estimates beside the user's numbers.

## Round 1

### K-1 · blocker · Two client-only PRs would save one user action as several writes
**Answer:** Fixed — rule "Metadata saves with its owner's write"; S15-08 adds values to `CreateArtifactRequest` and the graph create and update requests through the owner form of `CreateAll`; S15-10 and S15-11 send drafts.

### K-2 · blocker · The Swift metadata board only works for a Source, and three PRs would each copy it
**Answer:** To user (preparatory PR); added as S15-26 with rule "One metadata board", pending the user's call.

### K-3 · blocker · Rendered citations would be a second client cache with no update signal, and grouping is built twice
**Answer:** Fixed — rule "Citations render inside the reads that list them" replaces "Render in batches": one Go batch function returns runs, plain text, missing fields and a group key; the S16-30 and S16-13 reads call it; signals named (`ListSourceVocabulary` for format and field edits, Source scopes for values). S15-16, S15-21, S15-22 rescoped.

### K-4 · blocker · Retiring `CitationLineDisplay` is under-scoped, and the fallback would exist in Go and Swift
**Answer:** Fixed — rule "Today's line stays the fallback, in Swift"; S15-21 no longer retires it; S15-16 renders no text for a type without a format. Listed under Writer's own calls.

### K-5 · blocker · The built-in placeholders need one declared table
**Answer:** Fixed — rule "Built-in placeholders are one table"; S15-14 declares it, S15-16 serves it over FFI, S15-17's Insert field reads it.

### K-6 · major · The three levels are one concept with three names
**Answer:** Fixed — rule "One level type"; S15-06 adds the Go `Level` type, Swift enum and parity row.

### K-7 · major · Turning field keys into ids happens in two places, and the stored form is not decided
**Answer:** Fixed — rule "Formats name fields by id" rewritten: S15-14 defines the token and rewrite; one `sourcetypes` key-to-id function used by the type write and Install; delete refusal is a probe over the format columns; the rename Check is replaced (keys stay unrenamable); `add-catalog-delete` gains the text-reference case.

### K-8 · major · The "used by" count is a private copy of the in-use rule
**Answer:** Fixed — S15-08 computes `metadatafields.UsedBy` and the Metadata page count from the delete-impact totals.

### K-9 · major · S15-07 leaves out the functions S15-08 and S15-16 will need
**Answer:** Fixed — S15-07 moves every exported function to the owner form and adds a batch read by owners; `ListWorkspace` reads its fields in one query.

### K-10 · major · Three render RPCs where one would do, and the editor's warnings have no source
**Answer:** Fixed — rule "One render request for previews"; S15-14 returns diagnostics with positions; citation map § 7's bracket question closed.

### K-11 · major · Copy citation already exists; extend it
**Answer:** Fixed — S15-21 and S15-23 use the existing `PVClipboardButton` with one rich-payload write added to `PVClipboardCopyState`; Go returns runs and plain text, not rich text.

### K-12 · major · Sorting by date belongs to one rule
**Answer:** Fixed — rule "One date order"; S15-02 exports the rule and sends a sort key; S15-04 sorts by it in `SourcesModel`.

### K-13 · major · The move should rely on the effects registry
**Answer:** Fixed — rule "A move is the whole bundle" and S15-24 now record per-row changes and let scopes, `Reads`, search and caches follow from the registry; `evidencegraphs.CreateDefault`; S15-24 names the domain docs it updates.

### K-14 · major · Spike 16's map never names Spike 15
**Answer:** Fixed — Spike 16's map gains "Shared with Spike 15"; S15-05 and S15-22 depend on S16-03, S16-06, S16-07; S15-16 depends on S15-05. Ordering the new Swift keys after S16-25: Declined, by the user's instruction (2026-11-09) to run slices 1 to 4 alongside Spike 16; S16-25 converts them instead (Writer's own calls).

### K-15 · major · Scale misses sizes some entries depend on
**Answer:** Fixed — Scale line extended; S15-02, S15-08, S15-15 and S15-24 state their fan-out.

### K-16 · major · S15-16's `{citation.item}` needs a decision first
**Answer:** To user; settled 2026-11-09 (above).

### K-17 · minor · Docs and skills each PR makes false are not named
**Answer:** Fixed — S15-07 (`add-catalog-query` example), S15-15 (`add-catalog-delete` text references), S15-24 (two domain docs), S15-99 (retires the spike's `citation-formats.md` for the new domain doc). Spike 15 adds no cache, so `docs/catalog-caches.md` is untouched.

## Round 2

Fresh reviewer, document plus this log. Round-1 answers: K-2, K-3 (Go shape), K-4, K-5, K-8, K-9, K-10, K-11, K-13, K-14, K-15 hold; K-1, K-6, K-7, K-12, K-17 held only in part and are completed below; K-16 stays with the user.

### K-18 · blocker · A format or field edit has no signal that reaches the Person page's citations
**Answer:** To user (a change to Spike 16's S16-24 and S16-25 scope), with the recommendation above. The rule "Citations render inside the reads that list them" now says the refresh is open instead of claiming it works.

### K-19 · blocker · The graph form's one write cannot go through `CreateAll` on an existing graph
**Answer:** Fixed — rule "Metadata saves with its owner's write": drafts save through `sourcemetadata.Apply(tx, owner, draft)`, added in S15-07, with `CreateAll` as its empty-owner case; S15-08 wires it.

### K-20 · blocker · `Level` has no named home, and the format engine starts before it exists
**Answer:** Fixed — rule "One level type": `Level` lives in `core/vocabkey`; S15-06 adds it; S15-07 and S15-14 depend on S15-06 (map row, slice note, dependency graph).

### K-21 · blocker · Two binding rules contradict each other on `CitationLineDisplay`
**Answer:** Fixed — the sentence is gone from "One engine, in Go"; the citation map's § 6.2 row and the older `citation-formats.md` "Formatter" row now say it stays as the fallback.

### K-22 · blocker · The key-to-id lookup is put in `sourcetypes`
**Answer:** Fixed — `metadatafields.ResolveKeysTx(tx, keys, preferOrigin)` beside `lookupTx`, called by the `sourcetypes` format write and `sourcevocab.Install` (rule and S15-15, S15-20).

### K-23 · blocker · The shared date order would be exported from the Conclusion summary composer
**Answer:** Fixed — rule "One date order": the rule moves to `core/valuecompare` beside `DateEnds`, called by `conclusionsummaries` and `sources` (S15-02).

### Smaller round-2 notes
- **K-14 follow-up:** Spike 16's "Shared with Spike 15" S16-25 row no longer lists S15-09 (its boards come in the workspace read); Spike 15 keys added after S16-25 use its value keys (S15-17, S15-23 Context). Fixed.
- **K-17 follow-up:** S15-17 and S15-23 add their keys to `docs/catalog-caches.md` once S16-10 creates it. Fixed.
- **K-8 note:** what "Used by" counts across Sources, Artifacts, graphs and types is left to S15-08's plan.
- **S15-20:** the plan names how the dogfood Install run happens (`add-seeded-vocabulary` keeps `Install` out of FFI, Swift and open) and warns that a re-run overwrites edited product rows and restores deleted ones. Fixed in S15-20's Context and the citation map § 6.4.
- **S15-08:** "six tables probed" corrected: layout rows cascade and do not keep a field. Fixed.

## Round 3 — slice 8 (S15-27, S15-28, S15-29)

Fresh reviewer, slice 8 only. Writer's call to note: the rule "Citations render on demand, from SQL" now lets the provenance sheet hold its on-open call as one `CatalogQueryKey` while it is shown, evicted when hidden, so it gets S16-04's failure reading and the registry's refresh; nothing outlives the sheet (K-26).

### K-24 · blocker · A new grouped SQL query would duplicate facts the node and the wire already carry
**Answer:** Fixed — rule "Records load when the sheet opens": the detail read keeps `Support`, `Against`, values and state, and adds only `FieldRecordCounts` (record count, a mixed field's distinct Sources) composed from the node and the records tier; no new SQL. S15-29 removes Swift's count rules.

### K-25 · blocker · A kit recipe cannot use the Features fold
**Answer:** To user (preparatory PR); added as S15-30 after S16-07, with the rule rewritten for the kit fold (`Platform/` state, `PVShowMoreRow`).

### K-26 · blocker · The on-open call had no home for S16-04's reading or refetch while open
**Answer:** Fixed — the sheet's on-open call is one `CatalogQueryKey` evicted when hidden (S15-21, which now depends on S16-04); S15-28's read reports its inputs and S15-29 refreshes through the registry. The rule "Citations render on demand, from SQL" says so.

### K-27 · major · One (owner, Property) key does not cover spanning or derived rows
**Answer:** Fixed — the records read takes a row's (owner, Property) pairs, mirroring `RecordsSource`, and returns records grouped by pair.

### K-28 · major · S15-27 re-derives an engine fact in Swift
**Answer:** Fixed — each outcome carries whether it denied a value and the denier's Source title (S15-28); `deniedSomething` and the denier search go in S15-29; S15-27 is the fold only.

### K-29 · major · Spike 16 PRs slice 8 touches are missing from both maps
**Answer:** Fixed — S15-28 depends on S16-24, S15-29 on S16-09, S15-21 on S16-04; Spike 16's shared table gains S16-09, S16-24 and S15-21's key under S16-25; S15-29 restates `.conclusionDetail`'s Source follow.

### K-30 · major · The deprecated Why reads the records S15-28 removes
**Answer:** To user (above).

### K-31 · major · Scale and fan-out at 1,200 records not stated
**Answer:** Fixed — S15-16 times 300 and 1,200 Citations; S15-28 times the records read at 1,200; S15-29 times a 1,200-record sheet and states one refetch per write touching the read's inputs.

### K-32 · major · Docs and skills S15-28 and S15-29 make false
**Answer:** Fixed — S15-28: the `conclusiondetails` package comment and `conclusion-reconciliation.md`; S15-29: `macos-client-patterns.md`, `design-system-layers.md`, `add-ui-component`, and retiring `provenance-loading.md`.

### K-33 · minor · The fold's unit
**Answer:** Fixed — the fold counts citation groups and never splits one (rule and S15-27).

### K-34 · minor · "Summary" already names something else
**Answer:** Fixed — `FieldRecordCounts`.

### K-35 · nit · `leadOutcomes` will be gone
**Answer:** Fixed — S15-28 says each `PictureTile.Record` keeps its outcome.
