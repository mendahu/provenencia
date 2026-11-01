# Spike 15 — Sources and citations

## Status

**Not yet planned.** Three ideas were pulled out of [`ideas/`](../../ideas/) on 2026-11-01 so they sit together until the spike is scoped with [`plan-spike`](../../../.agents/skills/plan-spike/SKILL.md). There is no PR map (`deployment-plan.md`) yet, and nothing here is a commitment to a PR cut.

The theme: give Sources the fields a standard genealogical citation needs, starting with a first-class Source date, and let researchers re-file Artifacts between Sources as those Sources take shape. One Source schema pass (one migration) and one Source page design should cover the date and the first citation fields together.

## Documents

| Doc | Role |
| --- | --- |
| [**Source provenance date**](source-provenance-date.md) | The first-class Source publication date, descoped from Spike 13 (S13-20 – S13-22). Its meaning, storage, backfill and where it shows are settled |
| [**Citation formats**](citation-formats.md) | What an *Evidence Explained* citation needs, layer by layer, against what the catalog has; a six-step path from Source fields to a Go formatter and Copy citation |
| [**Move an Artifact to another Source**](move-artifact-between-sources.md) | Move an Artifact, with the Evidence graphs built on it, to another Source; the schema change, which graphs move, side effects |

## Open questions

Gathered from the three docs; each doc keeps its own context.

**Source date** ([`source-provenance-date.md`](source-provenance-date.md))

- Where the date sits on the Source page so it reads as the Source's own date and sort key, not another metadata row.
- Whether the Source page redesign moves it onto Page columns ([`ideas/wide-layout.md`](../../ideas/wide-layout.md)).

**Citation formats** ([`citation-formats.md`](citation-formats.md))

- Which elements are first-class columns (like the Source date) and which are seeded metadata per Source type.
- Whether a repository is its own record that many Sources share.
- Which Source types to template first: the product owner's dogfood Sources decide.
- Whether a template's fields live on the Source, the Artifact or the Citation (a census piece number is per Source; a folio is per Citation).
- Whether researchers can edit a template, or only the fields under it.

**Move an Artifact** ([`move-artifact-between-sources.md`](move-artifact-between-sources.md))

- Whether the entry point is only the Artifact, or also a graph ("move this graph and its Artifacts").
- Whether moving several Artifacts at once is in the first cut, or only the refuse-and-name path for shared graphs.
- Whether "move to a **new** Source" (split a Source in two, using Copy Source to prefill it) belongs here or follows.
- Undo: rely on moving it back, or offer an undo right after the move.
- How the Subjects' composite key follows a moved graph: `PRAGMA defer_foreign_keys` inside the transaction, or a migration adding `ON UPDATE CASCADE`.

## Not this spike

Still in [`ideas/`](../../ideas/): nested Sources, derivation links ("citing …" as a link between Sources), named Source templates and per-graph credibility ([`source-structure-later.md`](../../ideas/source-structure-later.md)); relationships between Sources ([`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)); pasting a registry URL to create a Source ([`registry-source-import.md`](../../ideas/registry-source-import.md)). Pull any of them in when the spike is scoped.
