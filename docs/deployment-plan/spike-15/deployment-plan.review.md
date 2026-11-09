# Review — Spike 15 PR map

**Verdict:** Fits with changes (after round 1). **Rounds:** 1. **Reviewers:** maintainability (Opus, fresh context, map review per [`review-plan`](../../../.agents/skills/review-plan/SKILL.md)).

## To user

- **K-2 · Preparatory Swift PR S15-26** (the metadata board takes an owner). Added to the map ahead of S15-09, S15-10 and S15-11; a preparatory PR is the user's call. Recommendation: keep it. Without it, three PRs each copy the Source page's board.
- **K-16 · Which Subject names the item** (`{citation.item}`) when a Citation supports several people. Needed before S15-16 is planned. Recommendation: the Citation label when set; else the event Subject with Observations on the Citation and its principal participant; else the first person Subject by ref; a census head of household only through the label or the graph metadata.

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
**Answer:** To user, with a recommended default (above).

### K-17 · minor · Docs and skills each PR makes false are not named
**Answer:** Fixed — S15-07 (`add-catalog-query` example), S15-15 (`add-catalog-delete` text references), S15-24 (two domain docs), S15-99 (retires the spike's `citation-formats.md` for the new domain doc). Spike 15 adds no cache, so `docs/catalog-caches.md` is untouched.
