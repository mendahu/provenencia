# Spike 15 — Sources and citations

## Status

**Open.** Planned 2026-11-09. PR map: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

Sources gain a date, a not-public switch and metadata at three levels (the Source, the Artifact, the Evidence graph); citation formats become editable data on each Source type, seeded with *Evidence Explained*-style formats; Artifacts move between Sources with their graphs. Slices 1 to 4 run alongside Spike 16; the rest build on its caches.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order |
| [**Source provenance date**](source-provenance-date.md) | The first-class Source publication date, descoped from Spike 13 (S13-20 – S13-22). Its meaning, storage, backfill and where it shows are settled |
| [**Citation formats**](citation-formats.md) | What an *Evidence Explained* citation needs, layer by layer, against what the catalog has; a six-step path from Source fields to a Go formatter and Copy citation. Its arrival steps are superseded by the citation map and the deployment plan |
| [**Evidence Explained citation map**](evidence-explained-citation-map.md) | Every record kind *Evidence Explained* covers, grouped into 17 patterns, with each element traced to catalog, graph or metadata data. The design: metadata on the Source (the record), the Artifact (the copy) and the Evidence graph (the item), and **citation formats as editable data on each Source type** (full, short, list) in a small template syntax, seeded with EE formats for the first types. § 1 separates *your copy* (the locator) from *finding the record again*; § 6.8 is the seed: 26 types, 63 fields and their formats, covering every pattern (revised 2026-11-09) |
| [**Move an Artifact to another Source**](move-artifact-between-sources.md) | Move an Artifact, with the Evidence graphs built on it, to another Source; the schema change, which graphs move, side effects |
| [**Source structure later**](source-structure-later.md) | What Spike 12 parked: nested Sources, derivation links, named templates, per-graph credibility. Reference only; not in this spike, moves to `ideas/` at close |
| [**Source privacy toggle**](2026-10-27-source-privacy-toggle.md) | Per-Source "not public" switch; persist and toggle only; no share behaviour yet (dogfood 2026-10-27) |
| [**Credibility field UX**](2026-10-18-credibility-field-truncates-and-ignores-enter.md) | Credibility argument truncates and ignores Enter (dogfood 2026-10-18) |
| [**Rethink Source metadata**](2026-10-25-rethink-source-metadata-and-type-suggestions.md) | Footprint map of Source metadata (dogfood 2026-10-25). Settled: metadata stays and citations are built from it |
| [**Completed**](completed.md) | Finished steps |
| [Design briefs](design/) | Claude Design, written just before the PR they gate |
| [PR plans](prs/) | One per PR, written just before the PR |

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

The design questions gathered here before planning are settled; their answers are working rules in [`deployment-plan.md`](deployment-plan.md#working-rules), and the remaining small ones are in the [citation map § 7](evidence-explained-citation-map.md#7-open-questions). Two stay with the briefs: where the date sits on the Source page, and whether the page moves onto Page columns ([`ideas/wide-layout.md`](../../ideas/wide-layout.md)) (S15-D1).

## Not this spike

Still in [`ideas/`](../../ideas/): relationships between Sources ([`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)); pasting a registry URL to create a Source ([`registry-source-import.md`](../../ideas/registry-source-import.md)); citation styles as installable packages ([`source-type-packages.md`](../../ideas/source-type-packages.md), raised 2026-11-09: Spike 15 seeds the *Evidence Explained*-style set as the product default, with no plugin system and no opt-out). Pull any of them in when the spike is scoped.
