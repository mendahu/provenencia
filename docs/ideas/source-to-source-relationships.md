# Source-to-source relationships

**Status:** idea only — not roadmapped. Parked here from [`interpretation-graph-ui.md`](archive/interpretation-graph-ui.md) leftovers **12** (Source-page `mentions` / `remark`) and **13** (placeholder Source + merge), plus the descoped **source cards on the Evidence graph**.

The Interpretation model already knows how one Source talks about another. The app does not. This note is the home for that domain until a spike pulls it.

## Problem

A book cites a marriage certificate. A letter says the certificate mistyped a date. Those are **not** person / event / place assertions, and they are not Source-layer filing columns. They are cited Interpretation about another Source.

The catalog already has the rows:

- a `source` subject that reifies exactly one `sources` row (at most one per Source);
- Properties `mentions` (subject-valued, target is typically a source subject) and `remark` (text about that subject);
- Citations that may sit under a *different* Source than the reified subject’s home — cross-source by construction.

What never shipped: any UI to create or read those Observations, any way to mint a Source you do not hold, and any merge when two placeholders turn out to be the same document.

An earlier graph-brainstorm draft also considered **drawing `source` subjects as cards on the Evidence graph**. That is refused. The domain still has to live somewhere.

## Settled (do not reopen)

Authoritative reasoning is [`interpretation-graph-ui.md`](archive/interpretation-graph-ui.md) §4.6 and [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) §4 / §5. Do not invent a parallel assertion table.

| Call | Why |
| --- | --- |
| **`source` subjects never render on the Evidence graph** | A filing-cabinet document is not a historical person, event, or place. Query filter, not a schema change. |
| **Keep `mentions` / `remark` as ordinary Observations** | Conclusion exhibit pins Observations. A bespoke `source_remarks` table would freeze vocabulary and break “the letter says this certificate is a forgery” as Reconciliation evidence. |
| **At most one `source` subject per Source** | Source identity is already `sources.id` / `SRC-…`. Duplicates have no Conclusion reconciliation path (reification subjects do not get canonical rows). |
| **Do not enforce same-Source Observations in SQL** | A trigger would break this feature and fight the model’s generic-graph posture. |
| **Do not give Sources a `canonical_entities` row** | That would be a second canonical handle for a row that already has one. |
| **Home the UI on the Source page** | A “what other Sources say about this one” section — inbound commentary *and* outbound mentions from this Source. |

Cross-source Observations stay legal. Person reuse across Sources stays forbidden on the canvas (Sameness later). Those two facts are independent.

## What this idea still owns

1. **Source-page commentary surface.** What the section looks like: inbound remarks, outbound `mentions`, jump to the citing or mentioned Source, polarity on a denial, empty state.
2. **Creating the reifying `source` subject.** Lazy on first mention vs create-with-Source. Application invariant, not a unique index (type id is a runtime UUID).
3. **Placeholder Sources.** Mentioning an unheld document requires a `sources` row (`subjects.source_id` is `NOT NULL`). Fileless Sources already exist; this is the research-to-do case.
4. **Source merge.** Two Sources each mention “the same” unseen certificate → two placeholders. Resolving that is a Source-layer merge that does not exist. Blocker for shipping `mentions`, not for the Evidence graph.
5. **How the researcher writes the Citation.** Same composer place vs a Source-page-local flow. The Citation still needs an Artifact on the *citing* Source.
6. **Later vocabulary.** `supersedes`, `is_transcription_of`, `derived_from` stay ordinary Properties once a workflow needs them — no migration if `mentions` / `remark` stay Observations.

## Why source cards left the graph

Decision 6 in the graph brainstorm: render filter **(a)**, not remodel **(b)**. Taking (b) (Citations point at `sources` directly) would:

- break Conclusion exhibit (pins are Observations only);
- close an open Property vocabulary;
- duplicate Citation machinery (second ref, audit, notes, FFI, search).

The graph slices therefore query `source_id = ?` and omit `source`-type subjects. That leftover is **this product surface**, not more canvas chrome.

## Open questions

- What does the Source-page section actually look like?
- Does mentioning an unheld document create a placeholder Source immediately, and what title / type / empty-state does that row get?
- What resolves two placeholders that are the same document? Merge? Link? Researcher-only note until merge exists?
- Who mints the `source` subject — Source create, first `mentions` / `remark`, or a repair job?
- Is the writer the citing Source’s page, the mentioned Source’s page, or both?
- How does this sit next to first-class [credibility assessments](../research-judgment-model.md) (those are not Observations and do not replace remarks)?

## Explicitly out of this note

- Spike 8 **S8-D4** / **S8-07** (Open Evidence graph on the Source page — navigation only)
- Putting `source` bubbles back on the Evidence graph
- Composer rethink / pinning ([`docs/dogfood/ux.md`](../dogfood/ux.md))
- Schema work unless a future spike deliberately reopens §4.6

## Related docs

- [`interpretation-graph-ui.md`](archive/interpretation-graph-ui.md) §4.6, §14
- [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) (source subjects, `mentions` / `remark`)
- [`seeded-vocabulary.md`](../seeded-vocabulary.md) §3.2–3.3
- [`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) (exhibit pins; no canonical rows for source subjects)
- [`research-judgment-model.md`](../research-judgment-model.md) (credibility vs cited commentary)
- [`source-layer-data-model.md`](../source-layer-data-model.md) (fileless Sources; merge does not exist)
