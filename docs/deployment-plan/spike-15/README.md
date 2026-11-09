# Spike 15 — Sources and citations

## Status

**Not yet planned.** Design inputs sit here until the spike is scoped with [`plan-spike`](../../../.agents/skills/plan-spike/SKILL.md). There is no PR map (`deployment-plan.md`) yet, and nothing here is a commitment to a PR cut.

The theme: give Sources the fields a standard genealogical citation needs, starting with a first-class Source date, and let researchers re-file Artifacts between Sources as those Sources take shape. One Source schema pass (one migration) and one Source page design should cover the date and the first citation fields together. Parked beside that (2026-11-04): Source privacy, credibility-field UX, a rethink of Source metadata / type suggestions, and Spike 12's longer Source-structure leftovers.

## Documents

| Doc | Role |
| --- | --- |
| [**Source provenance date**](source-provenance-date.md) | The first-class Source publication date, descoped from Spike 13 (S13-20 – S13-22). Its meaning, storage, backfill and where it shows are settled |
| [**Citation formats**](citation-formats.md) | What an *Evidence Explained* citation needs, layer by layer, against what the catalog has; a six-step path from Source fields to a Go formatter and Copy citation |
| [**Evidence Explained citation map**](evidence-explained-citation-map.md) | Every record kind *Evidence Explained* covers, grouped into 17 citation patterns; each pattern's elements mapped to catalog data, graph data or a seeded metadata field. Built on the existing Source types, metadata fields and type → field suggestions; new schema is the Source date plus the Source's metadata structure (values, layout, type templating) repeated on Artifacts and Citations; § 5.8 is the proposed first seed, for review (revised 2026-11-09) |
| [**Move an Artifact to another Source**](move-artifact-between-sources.md) | Move an Artifact, with the Evidence graphs built on it, to another Source; the schema change, which graphs move, side effects |
| [**Source structure later**](source-structure-later.md) | What Spike 12 parked: nested Sources, derivation links, named templates, per-graph credibility |
| [**Source privacy toggle**](2026-10-27-source-privacy-toggle.md) | Per-Source "not public" switch; persist and toggle only; no share behaviour yet (dogfood 2026-10-27) |
| [**Credibility field UX**](2026-10-18-credibility-field-truncates-and-ignores-enter.md) | Credibility argument truncates and ignores Enter (dogfood 2026-10-18) |
| [**Rethink Source metadata**](2026-10-25-rethink-source-metadata-and-type-suggestions.md) | Whether Source metadata and type→field suggestions stay, shrink, or go; footprint map (dogfood 2026-10-25) |

## Working rules

Carry these into `deployment-plan.md` as spike-wide working rules when the spike is planned, and into every PR plan that touches citations.

**Evidence Explained: follow the style, never copy the book (user, 2026-11-08).** A citation style is a method, and building citations that follow it is fine. The book's text is not ours to reuse. So:

- **Templates are our own.** Write every formatter template from our understanding of the style's rules: element order, punctuation, layers, the three forms. Never transcribe a QuickCheck model, or the book's wording for one, into code, tests, docs, help or UI copy.
- **Examples are invented.** Every example citation, in docs, tests, fixtures, previews, onboarding and marketing, uses made-up people, places, records and numbers written for Provenencia. Never use an example from the book, even with names changed.
- **Our own names.** Patterns, element kinds and UI labels get Provenencia's names ("Census", "Local record book"), not the book's chapter or model titles copied as a set. Organizing the set the same way the book does is a compilation risk.
- **No endorsement implied.** In the product, help and marketing, refer to the style only factually (for example, "citations in the *Evidence Explained* style") or generically ("standard genealogical citations"). Never put the name or the book's branding in a product or feature name, and never call anything "official" or "certified". The book's title may be a trademark of its publisher.
- **Check against the book; don't copy from it.** The team checks templates against a purchased copy of the current edition. Notes taken from it are paraphrased into rules, not quoted.
- **Before selling:** an IP lawyer reviews the template set and the marketing wording, and we consider asking the publisher (Evidence Explained, Inc. / Elizabeth Shown Mills) for permission or an endorsement for using the name. Guidance given 2026-11-08 was general, not legal advice.

## Open questions

Gathered from the three core docs; each doc keeps its own context.

**Source date** ([`source-provenance-date.md`](source-provenance-date.md))

- Where the date sits on the Source page so it reads as the Source's own date and sort key, not another metadata row.
- Whether the Source page redesign moves it onto Page columns ([`ideas/wide-layout.md`](../../ideas/wide-layout.md)).

**Citation formats** ([`citation-formats.md`](citation-formats.md))

- ~~Which elements are first-class columns and which are seeded metadata~~: seeded metadata, except the Source date ([`evidence-explained-citation-map.md`](evidence-explained-citation-map.md) § 5, user 2026-11-09).
- ~~Whether a repository is its own record~~: no, `repository` text, carried by Copy Source.
- Which Source types to template first: the product owner's dogfood Sources decide.
- ~~Whether a template's fields live on the Source, the Artifact or the Citation~~: any of them. The field vocabulary is shared; `artifact_metadata` and `citation_metadata` hold per-copy and per-item values, and the most specific value wins (user, 2026-11-09).
- How the seed reaches existing projects: once per seed version on open, or an action the researcher runs.
- Whether researchers can edit a template, or only the fields under it.

**Move an Artifact** ([`move-artifact-between-sources.md`](move-artifact-between-sources.md))

- Whether the entry point is only the Artifact, or also a graph ("move this graph and its Artifacts").
- Whether moving several Artifacts at once is in the first cut, or only the refuse-and-name path for shared graphs.
- Whether "move to a **new** Source" (split a Source in two, using Copy Source to prefill it) belongs here or follows.
- Undo: rely on moving it back, or offer an undo right after the move.
- How the Subjects' composite key follows a moved graph: `PRAGMA defer_foreign_keys` inside the transaction, or a migration adding `ON UPDATE CASCADE`.

**Metadata rethink** ([`2026-10-25-rethink-source-metadata-and-type-suggestions.md`](2026-10-25-rethink-source-metadata-and-type-suggestions.md))

- ~~Stay, shrink, or remove~~: stay. Metadata and type suggestions become what citations are built from (user, 2026-11-09).

## Not this spike

Still in [`ideas/`](../../ideas/): relationships between Sources ([`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)); pasting a registry URL to create a Source ([`registry-source-import.md`](../../ideas/registry-source-import.md)). Pull any of them in when the spike is scoped.
