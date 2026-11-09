# Deployment plan — Spike 15

Spike 15 makes Sources carry what a standard genealogical citation needs and turns it into citations. It pulls in every design doc in this folder: the Source date ([`source-provenance-date.md`](source-provenance-date.md)), citation formats and the seed ([`evidence-explained-citation-map.md`](evidence-explained-citation-map.md), which supersedes the arrival steps of [`citation-formats.md`](citation-formats.md)), moving an Artifact ([`move-artifact-between-sources.md`](move-artifact-between-sources.md)), the privacy toggle ([`2026-10-27-source-privacy-toggle.md`](2026-10-27-source-privacy-toggle.md)), the credibility field ([`2026-10-18-credibility-field-truncates-and-ignores-enter.md`](2026-10-18-credibility-field-truncates-and-ignores-enter.md)) and the metadata rethink ([`2026-10-25-rethink-source-metadata-and-type-suggestions.md`](2026-10-25-rethink-source-metadata-and-type-suggestions.md), settled: metadata stays). One behavior, one PR.

**A brief covers one view.** A PR that adds, changes or removes chrome waits on a brief for each view it touches; a removal is gated too. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). Only a PR whose views look the same before and after is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a researcher gets a correct, editable *Evidence Explained*-style citation for any Citation from data entered once on the Source, the Artifact and the Evidence graph, and can re-file Artifacts between Sources.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project:

1. **A Source has a date.** Set, change and clear it on the Source page with the app's date entry. The Sources list shows it on each row and sorts by it, undated last. The Person page's Sources section shows it on each Source.
2. **A Source can be marked not public.** A switch on the Source page; it is saved and survives reopening. Nothing else changes yet.
3. **Copy and item details live where they belong.** On a census Source, the Artifact holds provider, URL and accessed date (also offered when the Artifact is added), and each household's Evidence graph holds dwelling, family and line. The citation composer gains only the printed page.
4. **Citations read like citations.** A census Citation's provenance sheet reads "1880 U.S. census, Ottawa County, Kansas, population schedule, ED 236, p. 14, dwelling 112, family 118, Ada Robins; digital image, *Ancestry* (… : accessed 12 March 2025)". Copy citation gives the full or short form. The Source page previews its citation and names the fields its format expects and is missing.
5. **Formats are the researcher's.** Editing the census full format on its Source type page updates a live preview against a real Source. A new type can **Copy from** an existing one and starts with its formats and suggestions.
6. **A new project knows the citation styles.** It starts with the 26 seeded types, 62 fields, suggestions at each level and formats (map § 6.8). The product owner's dogfood project gets them by running Install once.
7. **One household, one citation.** The five Citations drawn on one household row show as one citation, with the highlighted pieces beneath it, on the provenance sheet and in the Person page's Sources section.
8. **An Artifact moves with its graphs.** **Move to another Source…** from an Artifact or a graph shows everything that moves together, keeps every ref, and the moved Citations then read the target Source.
9. **The credibility argument behaves like a text area.** A long argument wraps, and it saves with the Save button, like the Source description.

**Scale** (user, 2026-11-09). A census Source with **60 household graphs of 1 to 10 Citations each** (600 Citations); a Person page citing **30 Sources**, so a provenance sheet or Sources section renders up to **300 citations**; a move bundle of **5 Artifacts and 5 graphs**; the seed's **26 types and 62 fields**. Also: a project of **3,000 Sources** (Spike 16's Scale) for the Sources list; a Source with up to **60 Artifacts**; a move bundle's graphs holding up to **100 Subjects** in all; a format or field edit on a type used by **500 Sources**. Plans write their fan-out against these.

## Where the notes went

| Dogfood note or idea | PR | Brief |
| --- | --- | --- |
| [Credibility field truncates and ignores Enter](2026-10-18-credibility-field-truncates-and-ignores-enter.md) | S15-01 | — (working rule) |
| [Source privacy toggle](2026-10-27-source-privacy-toggle.md) | S15-02, S15-03 | S15-D1 |
| [Source provenance date](source-provenance-date.md) | S15-02, S15-03, S15-04, S15-05 | S15-D1, S15-D2, S15-D3 |
| [Rethink Source metadata](2026-10-25-rethink-source-metadata-and-type-suggestions.md) | Settled: metadata stays (map § 6.1); S15-06 – S15-12 build on it | — |
| [Evidence Explained citation map](evidence-explained-citation-map.md) | S15-06 – S15-23 | S15-D1, S15-D3 – S15-D9 |
| [Move an Artifact to another Source](move-artifact-between-sources.md) | S15-24, S15-25 | S15-D1, S15-D10 |

**Back in the ideas folder** — [`ideas/source-type-packages.md`](../../ideas/source-type-packages.md): citation styles as installable packages. [`source-structure-later.md`](source-structure-later.md) stays in this folder as reference and moves to `ideas/` at close: nested Sources, derivation links, named templates, per-graph credibility.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

- **Follow the *Evidence Explained* style; never copy the book.** Seeded formats, examples, tests, fixtures and UI copy are written for Provenencia; every example uses invented people, places and records; no product or feature name uses the book's name; nothing is called official. The full rule: [`README.md`](README.md#working-rules). Binds S15-14, S15-19, S15-20 and every PR with copy.
- **The Source date is one optional DateValue on `sources`.** `sources.date_value_id` into `date_values` ([`add-date-value`](../../../.agents/skills/add-date-value/SKILL.md)); its meaning is the original publication, or when the record was made; no backfill. The Sources list sorts by it, undated last. Metadata dates stay text. Binds S15-02 – S15-05. Design: [`source-provenance-date.md`](source-provenance-date.md).
- **Not public is a stored switch and nothing else.** `sources.is_public INTEGER NOT NULL DEFAULT 1`, toggled on the Source page. No filtering, hiding, list glyph, cascade or Copy Source carry. Binds S15-02, S15-03. Design: [`2026-10-27-source-privacy-toggle.md`](2026-10-27-source-privacy-toggle.md).
- **Metadata lives at three levels with one vocabulary.** The Source (the record), the Artifact (the copy used) and the Evidence graph (the item cited) each hold values and layout, with the same fields from `source_metadata_fields`, which keeps its name. Tables: `artifact_metadata`, `artifact_metadata_layout`, `evidence_graph_metadata`, `evidence_graph_metadata_layout`. Layout rules (order, dismiss) are the Source's at every level. Binds S15-07 – S15-12. Design: map § 6.6.
- **One level type.** Source, Artifact and graph are one Go `Level` type, used by suggestion rows, metadata owners, format prefixes and the RPCs, with one Swift enum and a row in `api/ffi/swift_vocabulary_test.go`. Binds S15-06 on. (K-6.)
- **Metadata saves with its owner's write.** Values entered while adding an Artifact or saving the Evidence graph form go in the same transaction as the Artifact or graph (`CreateArtifactRequest`, the graph create and update requests, through the owner form of `sourcemetadata.CreateAll`); the boards on existing owners save per field as the Source page does. Binds S15-08, S15-10, S15-11. (K-1.)
- **One metadata board.** The Swift board model and view take an owner (kind, id, its entries and its write operations); the Source page, the Artifact rows, the add flow and the graph form all use it. Binds S15-26, S15-09 – S15-11. (K-2.)
- **One date order.** The rule for where a date sits in an order (today `conclusionsummaries`' `datePlace` / `dateKeyOf`) is exported and gives each Source a date sort key on the wire; Swift's Sources list sorts by that key, undated last, and never orders dates itself. Binds S15-02, S15-04. (K-12.)
- **Suggestions have a level.** `source_type_metadata_fields.level` is `source`, `artifact` or `graph`, in the table's key. An Artifact's and a graph's suggestions are their Source's type's rows at that level. Binds S15-06, S15-09 – S15-12, S15-19.
- **Citations carry no metadata.** The composer's only new input is the printed page (`page_label` on the `page` selector, already in the locator). Binds S15-13. Design: map § 6.9.
- **Citation formats are data on the Source type.** Three nullable columns, `citation_full`, `citation_short`, `citation_list`, in the syntax of map § 6.3: `{level.meta.key}` and `{meta.key}` (graph, else Artifact, else Source), built-ins directly under their level (`{source.title}`, `{source.date:year}`, `{citation.page}`, `{citation.item}` …), `[optional groups]`, `{a|b}`, `*italic*`, backslash escapes. The list form reads the Source only. A type with no full format shows today's citation line. Binds S15-14 – S15-23.
- **No transforms on metadata.** Values print exactly as typed; only the Source date (a DateValue) has formatting options (`:year`, `:full`). Each field's description says to enter the value as it should appear. Refused: inverting, reversing, abbreviating. Binds S15-14, S15-19.
- **Formats name fields by id.** `core/citationformat` defines the stored token for a field id and one rewrite of the parse tree between keys and ids; one `sourcetypes` function maps keys to ids (preferring a given origin) and both the type write and `sourcevocab.Install` call it. A field named in any format is in use: its delete is refused by a delete-impact probe over the three format columns, listing the types ([`add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md), which gains the text-reference case). Field keys stay unrenamable, as today. Binds S15-14, S15-15, S15-17, S15-20. (K-7.)
- **One engine, in Go.** `core/citationformat` parses and renders; values go in after parsing, so a `*` or `[` in a value never formats. The app renders only through FFI. `CitationLineDisplay` goes when the provenance sheet moves to the engine. Binds S15-14, S15-16, S15-21 – S15-23.
- **Citations render inside the reads that list them.** One Go function renders a batch of Citations and returns, per Citation, the styled runs, the plain text, the missing fields and a group key; the provenance records read (S16-30) and the Sources section read (S16-13) call it, so a screen makes no separate render call and Swift keeps no cache of rendered text. The resolver reads each Source, Artifact and graph's values once per batch (300 Citations at Scale) through one batch read by owners. A format or field edit is signalled as a Source vocabulary change (`ListSourceVocabulary`) and Artifact and graph values as Source-scoped changes, so the reads refresh through Spike 16's declared signals. Binds S15-16, S15-21 – S15-23. (Map review K-3.)
- **One render request for previews.** The editor and the Source page preview through one request: a form, an optional unsaved format, and a target (Citation ids, or a Source for the list form); it returns runs, missing fields and parse diagnostics. Swift never parses a format. Binds S15-16, S15-17, S15-23. (K-10.)
- **Built-in placeholders are one table.** `core/citationformat` declares each built-in once (name, level, allowed in the list form, options such as `:year`, `:full`); parsing, the list-form rule, missing fields, the resolver and the editor's Insert field list (over FFI) all read it. Binds S15-14, S15-16, S15-17. (K-5.)
- **Today's line stays the fallback, in Swift.** For a type with no full format the engine returns no text, and the app shows today's localized line from `CitationLineDisplay`, which stays for that and for its other callers (Gallery captions, the likeness viewer, a denier's Source name). Go does not write a second copy. Binds S15-16, S15-21. (K-4.)
- **Build on Spike 16's caches.** Source words come from the Source records (S16-12), per-graph Subject records from the records tier (S16-13), graph shells from the graph catalog (S16-15) and provenance records from S16-30. Spike 15 PRs that read or move those land after them; the rest run alongside Spike 16 (user, 2026-11-09). Spike 16's map names this spike in **Shared with Spike 15**. Spike 15's new Swift query keys are added to today's `CatalogQueryKey` enum and converted with the rest by S16-25, rather than waiting for it, so slices 1 to 4 can run alongside Spike 16 (user, 2026-11-09). Binds S15-05, S15-16, S15-21, S15-22, S15-24.
- **The seed is the product default.** The 26 types, 62 fields, suggestions and formats of map § 6.8 go into `core/database/sourcevocab` and install at create through `sourcevocab.Install`. No plugin, no opt-out, no install on open, no seed-version row. The dogfood projects are updated by running Install once against them; the PR says how. Binds S15-19, S15-20.
- **Grouping is by the rendered full citation, computed when read.** Citations whose full citation renders the same are one citation with their Citations beneath; nothing is stored. Binds S15-21, S15-22.
- **A move is the whole bundle.** The bundle is the closure of the picked Artifact or graph over `evidence_graph_artifacts`. One flow from an Artifact or a graph; one function computes the bundle for the preflight and the write; one transaction updates `source_id` on `artifacts`, `evidence_graphs` and `subjects` with deferred foreign keys, recording per-row changes so audit scopes (both Sources), `Reads`, search and Spike 16's caches all follow from the effects registry, with no hand-written cache update or scope tag; a Source left with no graph gets `evidencegraphs.CreateDefault`. Source-level values are not looked at. Binds S15-24, S15-25. Design: [`move-artifact-between-sources.md`](move-artifact-between-sources.md).
- **The credibility fix is idiomatic and needs no brief.** The argument becomes the same `PVTextArea` the Source description uses, with the platform's text-area behavior (Return is a new line) and the existing Save button. It adds no new chrome beyond the control the page already uses (user, 2026-11-09). Binds S15-01.
- **Refused in this spike.** A shared Repository record; per-Source format overrides; a fourth citation form; source type packages ([idea](../../ideas/source-type-packages.md)); install on open; undo of a move; Merge Sources; moving to a new Source; filling the jurisdiction from the graph; a per-Artifact page map. Each is recorded in the map (§ 7) or the move doc.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S15-01 | The credibility argument is a text area | 1 | Swift: `PVTextArea` with the description's behavior; Save as today | — | — | — | Scoped |
| S15-02 | Sources carry a date and a not-public flag | 1 | Go: one migration (`date_value_id`, `is_public`), Source read and write, a date sort key from the shared date order | — | — | — | Scoped |
| S15-03 | The Source page shows the date and the not-public switch | 1 | Swift: set, change, clear the date; the switch | S15-D1 | S15-02, S15-D1 | — | Scoped |
| S15-04 | The Sources list shows and sorts by the date | 1 | Swift: the date on each row; a date sort by Go's key, undated last | S15-D2 | S15-02, S15-D2 | — | Scoped |
| S15-05 | The Person page's Sources section shows each Source's date | 5 | Go and Swift: the date on the Source record and in the section | S15-D3 | S15-02, S16-12, S16-13, S16-03, S16-06, S16-07, S15-D3 | — | Scoped |
| S15-06 | Type suggestions have a level | 2 | Go: one `Level` type; migration rebuilding `source_type_metadata_fields` with `level`; readers and RPCs take it | — | — | — | Scoped |
| S15-07 | Metadata operations take an owner | 2 | Preparatory, Go: every `sourcemetadata` function on an owner, plus a batch read by owners; Sources unchanged | — | — | — | Scoped |
| S15-08 | Artifacts and graphs hold metadata | 2 | Go: the four tables, model, audit, delete, Copy Source, FFI; values in the Artifact and graph writes; Used by from delete impact | — | S15-06, S15-07 | — | Scoped |
| S15-09 | Artifacts show their metadata on the Source page | 2 | Swift: a metadata board per Artifact | S15-D1 | S15-08, S15-26, S15-D1 | — | Scoped |
| S15-10 | Adding an Artifact asks for its copy details | 2 | Swift: the Artifact's suggested fields in the add flow | S15-D4 | S15-08, S15-26, S15-D4 | — | Scoped |
| S15-11 | Evidence graphs show their metadata | 2 | Swift: a metadata board in the graph form | S15-D5 | S15-08, S15-26, S15-D5 | — | Scoped |
| S15-12 | Source types suggest fields at each level | 2 | Swift: the type page edits three suggestion lists | S15-D6 | S15-06, S15-D6 | — | Scoped |
| S15-13 | The composer records the printed page | 3 | Swift: `page_label` on the page selector | S15-D8 | S15-D8 | — | Scoped |
| S15-14 | The citation format engine | 4 | Go: `core/citationformat` parse, render, diagnostics, the built-ins table and the field-id token; table tests; syntax doc | — | — | — | Scoped |
| S15-15 | Source types store citation formats | 4 | Go: migration (three columns), id-keyed storage, delete impact, type read and write | — | S15-14 | — | Scoped |
| S15-16 | Citations render through their type's formats | 5 | Go: the resolver; one batch render for the reads; one render request for previews; built-ins over FFI | — | S15-08, S15-15, S15-05, S16-12, S16-13 | — | Scoped |
| S15-17 | The format editor | 5 | Swift: three formats, Insert field by level, live preview, warnings | S15-D6 | S15-16, S15-12, S15-D6 | — | Scoped |
| S15-18 | A new type can copy another's formats and suggestions | 4 | Go and Swift: one write; **Copy from** in the new type dialog | S15-D7 | S15-15, S15-06, S15-D7 | — | Scoped |
| S15-19 | Seed the citation types, fields and suggestions | 4 | Go: `sourcevocab` registry and Install for 26 types, 62 fields, levelled suggestions | — | S15-06 | — | Scoped |
| S15-20 | Seed the citation formats | 4 | Go: each seeded type's three formats; a render test per type | — | S15-19, S15-15 | — | Scoped |
| S15-21 | The provenance sheet shows formatted citations | 6 | Go and Swift: the provenance read renders and groups; the sheet shows them; Copy citation copies rich text | S15-D9 | S15-16, S16-30, S15-D9 | — | Scoped |
| S15-22 | The Sources section groups a Source's Citations by citation | 6 | Go and Swift: the Sources section read renders and groups | S15-D3 | S15-16, S15-05, S16-03, S16-06, S16-07, S15-D3 | — | Scoped |
| S15-23 | The Source page previews its citation | 6 | Swift: preview, missing fields, Copy citation | S15-D1 | S15-16, S15-D1 | — | Scoped |
| S15-24 | Move a bundle to another Source in the engine | 7 | Go: closure, preflight and move RPCs, one transaction through the effects registry | — | S16-12, S16-13, S16-15 | — | Scoped |
| S15-25 | Move from an Artifact or a graph | 7 | Swift: the menu items and the bundle confirm | S15-D10, S15-D1 | S15-24, S15-D10, S15-D1 | — | Scoped |
| S15-26 | The metadata board takes an owner | 2 | Preparatory, Swift: the board model and view take an owner; the Source page unchanged | — | — | — | Scoped |
| S15-99 | Dogfood close / docs | — | Bar pass, the ship logs' follow-ups triaged, the dogfood Install run, the spike's `citation-formats.md` retired in favour of the domain doc, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S15-D1 | Source page | **S15-03**, S15-09, S15-23, S15-25 | Where the date sits (it reads as the Source's own date, not a metadata row) and whether the page moves onto Page columns ([`wide-layout.md`](../../ideas/wide-layout.md)); the not-public switch; an Artifact's metadata board; the citation preview with missing fields and Copy citation; Move on Artifact rows and graphs. One board for the page, drawn before S15-03 |
| S15-D2 | Sources list | **S15-04** | The date on rows and as a sort |
| S15-D3 | Person page Sources section | **S15-05**, S15-22 | The date on each Source; Citations grouped under one citation |
| S15-D4 | Adding an Artifact | **S15-10** | The copy fields in the add flow |
| S15-D5 | Evidence graph form | **S15-11** | The item's metadata board beside name and description |
| S15-D6 | Source type page | **S15-12**, S15-17 | Suggestions at three levels; the three formats, Insert field, live preview, warnings |
| S15-D7 | New Source type dialog | **S15-18** | Copy from |
| S15-D8 | Citation composer | **S15-13** | The printed page beside the page in the file |
| S15-D9 | Provenance sheet | **S15-21** | The formatted citation, grouped Citations, Copy citation |
| S15-D10 | Move confirm | **S15-25** | The bundle, the target picker, the credibility note |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Source date, privacy, credibility           (alongside Spike 16)
  S15-01  The credibility argument is a text area
  S15-02  Sources carry a date and a not-public flag
  ✎ S15-D1 ──▶ S15-03  The Source page shows the date and the not-public switch
  ✎ S15-D2 ──▶ S15-04  The Sources list shows and sorts by the date
  Check: bar 1 (Source page and list), bar 2, bar 9.

SLICE 2 — Metadata at three levels                    (alongside Spike 16)
  S15-06  Type suggestions have a level
  S15-07  Metadata operations take an owner
  S15-08  Artifacts and graphs hold metadata
  S15-26  The metadata board takes an owner
  S15-09  Artifacts show their metadata on the Source page   (S15-D1)
  ✎ S15-D4 ──▶ S15-10  Adding an Artifact asks for its copy details
  ✎ S15-D5 ──▶ S15-11  Evidence graphs show their metadata
  ✎ S15-D6 ──▶ S15-12  Source types suggest fields at each level
  Check: bar 3 except the composer.

SLICE 3 — Printed page                                (alongside Spike 16)
  ✎ S15-D8 ──▶ S15-13  The composer records the printed page
  Check: bar 3's composer.

SLICE 4 — Formats as data                             (alongside Spike 16)
  S15-14  The citation format engine
  S15-15  Source types store citation formats
  ✎ S15-D7 ──▶ S15-18  A new type can copy another's formats and suggestions
  S15-19  Seed the citation types, fields and suggestions
  S15-20  Seed the citation formats
  Check: bar 6 on a new project; bar 5's Copy from.

SLICE 5 — Rendering                                   (after S16-12, S16-13)
  S15-05  The Person page's Sources section shows each Source's date   (S15-D3)
  S15-16  Citations render through their type's formats
  S15-17  The format editor                                            (S15-D6)
  Check: bar 1 on the Person page; bar 5.

SLICE 6 — Citations in the app                        (after S16-30)
  ✎ S15-D9 ──▶ S15-21  The provenance sheet shows formatted citations
  S15-22  The Sources section groups a Source's Citations by citation  (S15-D3)
  S15-23  The Source page previews its citation                        (S15-D1)
  Check: bar 4, bar 7.

SLICE 7 — Move                                        (after S16-12, S16-13, S16-15)
  S15-24  Move a bundle to another Source in the engine
  ✎ S15-D10 ──▶ S15-25  Move from an Artifact or a graph
  Check: bar 8.

S15-99 — Dogfood close / docs
```

**In parallel with Spike 16:** slices 1 to 4 need nothing from it. Within them, S15-01, S15-02, S15-06, S15-07, S15-14 and S15-26 have no dependencies and can start at once; slice 1's UI waits on S15-D1 and S15-D2. Slice 5 starts when S16-12 and S16-13 land (S15-05 also after S16-03, S16-06 and S16-07, which edit the same Sources section files), slice 6 when S16-30 lands, slice 7 when S16-15 lands as well.

### Dependencies at a glance

```text
S15-02 ──▶ S15-03, S15-04, S15-05, S15-16
S15-06 ──▶ S15-08, S15-12, S15-18, S15-19
S15-07 ──▶ S15-08 ──▶ S15-09, S15-10, S15-11, S15-16
S15-26 ──▶ S15-09, S15-10, S15-11
S15-05 ──▶ S15-16
S15-14 ──▶ S15-15 ──▶ S15-16, S15-18, S15-20
S15-19 ──▶ S15-20
S15-16 ──▶ S15-17, S15-21, S15-22, S15-23
S15-12 ──▶ S15-17
S15-05 ──▶ S15-22
S15-24 ──▶ S15-25

Spike 16:
S16-12, S16-13 ──▶ S15-05, S15-16
S16-03, S16-06, S16-07 ──▶ S15-05, S15-22
S16-30 ──▶ S15-21
S16-12, S16-13, S16-15 ──▶ S15-24

Briefs:
S15-D1 ──▶ S15-03, S15-09, S15-23, S15-25     S15-D2 ──▶ S15-04
S15-D3 ──▶ S15-05, S15-22                      S15-D4 ──▶ S15-10
S15-D5 ──▶ S15-11                              S15-D6 ──▶ S15-12, S15-17
S15-D7 ──▶ S15-18                              S15-D8 ──▶ S15-13
S15-D9 ──▶ S15-21                              S15-D10 ──▶ S15-25
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Two to six lines each. Link, do not restate.

### Slice 1 — Source date, privacy, credibility

#### S15-01 — The credibility argument is a text area

**Ships.** The credibility argument on the Source page is a `PVTextArea` like the description (wraps, grows, then scrolls), with the platform's text-area keys and the existing Save button.
**Not this PR.** A shared Save shortcut for every inline editor (the note's idea about `PVInlineEditActions`).
**Check.** A 500-character argument is fully visible and editable; Return adds a line; Save saves it.
**Context.** Rule "The credibility fix is idiomatic and needs no brief"; [note](2026-10-18-credibility-field-truncates-and-ignores-enter.md); `SourcePageCredibilityView.swift`, `SourcePageDescriptionView.swift`.

#### S15-02 — Sources carry a date and a not-public flag

**Ships.** One migration adds `sources.date_value_id` and `sources.is_public`; the catalog model, the Source read and write, the FFI `Source` message and the Sources list read carry both, plus a date sort key from the exported date-order rule. Audit picks the columns up through the `sources` effect. The list read joins `date_values` once for the whole list (3,000 Sources at Scale).
**Not this PR.** Any chrome (S15-03, S15-04); the Person page (S15-05); Copy Source carrying either (refused: privacy; the date is a Copy Source question for its plan).
**Check.** Handler tests: set, read and clear a date and the flag; sort keys order exact, `abt.`, `bef.` and range dates as the Timeline does; an upgraded catalog reads every Source as undated and public.
**Context.** Rules "The Source date is one optional DateValue", "Not public is a stored switch", "One date order" (`core/database/conclusionsummaries/compose.go` `datePlace`, `dateKeyOf`); [`source-provenance-date.md`](source-provenance-date.md); [privacy note](2026-10-27-source-privacy-toggle.md) "Likely code locations"; skills `add-catalog-migration`, `add-catalog-model`, `add-date-value`, `add-ffi-handler`. Updates [`source-layer-data-model.md`](../../source-layer-data-model.md) § 4.

#### S15-03 — The Source page shows the date and the not-public switch

**Ships.** On the Source page, the date is set, changed and cleared with the app's structured date entry where S15-D1 puts it, and the not-public switch is shown and saved.
**Not this PR.** The Sources list (S15-04); the rest of S15-D1's board (S15-09, S15-23, S15-25).
**Check.** Bar 1 on the Source page; bar 2.
**Context.** S15-D1; `SourcePageIdentityHeader.swift`, `SourceIdentitySection.swift`, `PVSwitch`; [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md).

#### S15-04 — The Sources list shows and sorts by the date

**Ships.** Each Sources list row shows the date as S15-D2 draws it, and `SourcesModel`'s `Sort` gains the date, ordered by Go's sort key, undated last.
**Not this PR.** Filtering by date or by not public.
**Check.** Bar 1's list half, on a project with dated and undated Sources.
**Context.** S15-D2; `SourcesListRow.swift`, `SourcesListSettings.swift`, `SourcesModel.swift`.

### Slice 2 — Metadata at three levels

#### S15-06 — Type suggestions have a level

**Ships.** One Go `Level` type (Swift enum and vocabulary parity row). A migration rebuilds `source_type_metadata_fields` with `level` in its key (existing rows become `source`); `sourcevocab`, `sourcetypes`, the effects registry and the catalog model read it; the suggestion RPCs take a level, and the Source page's board reads `source` only.
**Not this PR.** Artifact and graph values (S15-08); the type page's three lists (S15-12).
**Check.** The Source page's suggestions are unchanged; handler tests add and list suggestions per level; a field suggested at two levels keeps both rows.
**Context.** Rule "Suggestions have a level"; map § 6.6; skills `add-catalog-migration`, `add-catalog-model`, `add-seeded-vocabulary` (update its suggestion shape).

#### S15-07 — Metadata operations take an owner

**Ships.** Preparatory: every exported function of `core/database/sourcemetadata` (`Set`, `Clear`, `DismissSuggestion`, `Reorder`, `ListWorkspace`, `ListBySourceTx`, `ListLayoutTx`, `CreateAll` and the rest) takes an owner (`Level` and id) instead of a Source id, with only the Source wired, and one batch read of values by owners is added; `ListWorkspace` reads its fields in one query instead of one per extra value. S15-08 then adds tables, not a second copy of the logic.
**Not this PR.** New tables (S15-08).
**Check.** Every existing metadata test passes unchanged; the Source page behaves the same.
**Context.** Rule "Metadata lives at three levels"; [`source-layer-data-model.md`](../../source-layer-data-model.md) § 5.4 (the merge rules); `core/database/sourcecopy` (its callers). Updates the `sourcemetadata.ListLayoutTx` example in [`add-catalog-query`](../../../.agents/skills/add-catalog-query/SKILL.md).

#### S15-08 — Artifacts and graphs hold metadata

**Ships.** The four tables of map § 6.6, wired for the Artifact and graph owners: catalog model, audit under the Source's scope, delete impact (a field in use at any level, six tables probed), release with the Artifact or graph, Copy Source copies a chosen Artifact's values and layout; FFI for the board operations per owner, with suggestions from the Source's type at that level; metadata values on `CreateArtifactRequest` and the graph create and update requests, written in the same transaction; `metadatafields.UsedBy` and the Metadata page's count come from the delete-impact totals. The Source page's workspace read returns its Artifacts' boards in one query (60 Artifacts at Scale).
**Not this PR.** Search over Artifact or graph values (not this spike); any chrome (S15-09 – S15-11).
**Check.** Handler tests per owner: set, clear, dismiss, reorder, list with suggestions; an Artifact created with values is one audited write; a field in use only on a graph refuses delete and its Used by counts it; Copy Source carries Artifact values; a graph delete releases its rows.
**Context.** Map § 6.6 "Knock-on work"; skills `add-catalog-migration`, `add-catalog-model`, `add-audit-scope`, `add-catalog-delete`, `add-ffi-handler`. Updates [`source-layer-data-model.md`](../../source-layer-data-model.md) §§ 5, 7, 7.1 and [`catalog-deletes.md`](../../catalog-deletes.md).

#### S15-26 — The metadata board takes an owner

**Ships.** Preparatory, Swift: `SourceMetadataSection` and `SourcePageMetadataView` become one board model and view that take an owner (level, id, its entries and its write operations, or a draft for a form); the Source page moves onto it with no visible change.
**Not this PR.** Any new board (S15-09 – S15-11).
**Check.** The Source page's metadata behaves as before; Swift tests drive the board against a fake owner.
**Context.** Rule "One metadata board"; `macos/App/Features/Sources/SourceMetadataSection.swift`, `SourcePageMetadataView.swift`; [`add-swift-test`](../../../.agents/skills/add-swift-test/SKILL.md).

#### S15-09 — Artifacts show their metadata on the Source page

**Ships.** Each Artifact on the Source page has a metadata board, the S15-26 board with the Artifact as owner, as S15-D1 draws it.
**Not this PR.** The add flow (S15-10).
**Check.** Bar 3's Artifact half on a census Source.
**Context.** S15-D1; `SourcePageArtifactsView.swift`, `SourcePageMetadataView.swift`, `SourceMetadataSection.swift`.

#### S15-10 — Adding an Artifact asks for its copy details

**Ships.** When an Artifact is added, its type's `artifact`-level suggestions can be filled in the add flow (the S15-26 board on a draft), and they save with the Artifact in one write.
**Not this PR.** Filling them automatically ([`registry-source-import.md`](../../ideas/registry-source-import.md)).
**Check.** Adding a downloaded census image offers provider, URL and accessed date, and the values show on the Artifact afterwards.
**Context.** S15-D4; map § 1 "Copy details are captured when the Artifact is added"; `IngestFileDropRow.swift`, `SourceFormLoadArea.swift`.

#### S15-11 — Evidence graphs show their metadata

**Ships.** The Evidence graph form has a metadata board for the item (the S15-26 board on a draft), with the type's `graph`-level suggestions, saved with the form's one write.
**Not this PR.** Nudging researchers toward one graph per household (map § 7).
**Check.** Bar 3's graph half: dwelling, family and line on a household graph.
**Context.** S15-D5; map § 6.9; `SourceGraphFormDialog.swift`, `SourceGraphFormModel.swift`.

#### S15-12 — Source types suggest fields at each level

**Ships.** The Source type page edits three suggestion lists (Source, Artifact, Evidence graph), each ordered.
**Not this PR.** Formats (S15-17).
**Check.** Adding `line_number` to a type's graph list makes it a suggestion on that type's graphs.
**Context.** S15-D6; `SourceTypesDetailPane.swift`, `SourceTypesModel.swift`.

### Slice 3 — Printed page

#### S15-13 — The composer records the printed page

**Ships.** The citation composer offers the printed page beside the page in the file and writes it as `page_label` on the `page` selector.
**Not this PR.** A per-Artifact page map (map § 7).
**Check.** A Citation saved with printed page "14" shows "p. 14" on today's citation line.
**Context.** S15-D8; rule "Citations carry no metadata"; [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) § 3.2; `CitationLocatorDraft.swift`, `CitationFieldsDraft.swift`.

### Slice 4 — Formats as data

#### S15-14 — The citation format engine

**Ships.** `core/citationformat`: `Parse` to a tree (literals, placeholders with alternatives, groups, italic spans), `Render` through a resolver interface to styled runs and plain text, a missing-field list, and diagnostics with positions for unknown syntax and unbalanced brackets; the built-ins table; the stored token for a field id and the rewrite between keys and ids. No database.
**Not this PR.** The resolver over the catalog (S15-16); storage (S15-15).
**Check.** Table tests: every syntax rule of map § 6.3, nested groups, escapes, a value holding `*` and `[`, the census example of map § 6.8.
**Context.** Rules "Citation formats are data", "No transforms", "One engine, in Go", "Built-in placeholders are one table", "Formats name fields by id"; map § 6.3. Writes the syntax as a domain doc (new `docs/citation-formats.md`, linked from [`source-layer-data-model.md`](../../source-layer-data-model.md)); S15-99 retires the spike's `citation-formats.md`.

#### S15-15 — Source types store citation formats

**Ships.** A migration adds `citation_full`, `citation_short`, `citation_list` to `source_types`; writes resolve `{…meta.key}` to field ids through the one `sourcetypes` key-to-id function, and reads show keys; a delete-impact probe over the format columns refuses deleting a field a format names, listing the types; the type read and write RPCs carry the formats and their diagnostics; the list form rejects non-Source placeholders. A format edit reports `ListSourceVocabulary`.
**Not this PR.** Rendering a Citation (S15-16); the editor (S15-17).
**Check.** Handler tests: save and read a format; a key held by a product and a user field resolves to the one chosen; relabel a field and the format still reads; delete a field in use is refused with its types.
**Context.** Rule "Formats name fields by id"; map § 6.3 "Storage"; `core/database/deleteimpact/register.go` (`joinEdge`); skills `add-catalog-migration`, `add-catalog-delete` (gains the text-reference case), `add-ffi-handler`. A format edit refreshes the citations of every Source of the type through the vocabulary signal (500 Sources at Scale).

#### S15-18 — A new type can copy another's formats and suggestions

**Ships.** One engine write creates a type as a copy of another's three formats and suggestions at every level; the new type dialog offers **Copy from**.
**Not this PR.** Copying Sources or metadata values.
**Check.** Bar 5's Copy from: a new type copied from Census has its formats and its graph suggestions.
**Context.** S15-D7; map § 6.3 "The editor"; `SourceTypesListPane.swift`, `SourceTypesModel.swift`.

#### S15-19 — Seed the citation types, fields and suggestions

**Ships.** `sourcevocab`'s registry gains the 26 types (with icons from `sourcetypes/icons.go`), 62 fields (with descriptions) and suggestions at their levels in format order (map § 6.8); `birth_certificate` stops suggesting `record_date` and `issue_date`.
**Not this PR.** Formats (S15-20).
**Check.** A new project lists the 26 types; a census Source suggests its Source fields, its Artifacts the copy fields, its graphs the item fields; Install run twice changes nothing.
**Context.** Rule "The seed is the product default"; map § 6.8; skill `add-seeded-vocabulary`. Updates [`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 2.

#### S15-20 — Seed the citation formats

**Ships.** Each seeded type's three formats, as in map § 6.8, installed with the types; Install writes them after the fields so keys resolve to ids. Says in the PR how to run Install against an existing dogfood project.
**Not this PR.** Rendering them in the app (S15-16 on).
**Check.** A test renders every seeded format in all three forms against an invented Source of its type and compares the expected text.
**Context.** Rules "Follow the style; never copy the book", "The seed is the product default", "Formats name fields by id" (Install calls the same key-to-id function); map § 6.8 "Formats".

### Slice 5 — Rendering

#### S15-05 — The Person page's Sources section shows each Source's date

**Ships.** The Source record (S16-12) carries the date, and each Source in the Person page's Sources section shows it as S15-D3 draws it. Lands after S16-03, S16-06 and S16-07, which edit `PersonSourcesSection.swift` and `PersonSourcesContent.swift`.
**Not this PR.** Grouped citations (S15-22).
**Check.** Bar 1's Person page half.
**Context.** S15-D3; rule "Build on Spike 16's caches"; S16-12 and S16-13 scopes in [Spike 16's map](../spike-16/deployment-plan.md); `PersonSourcesContent.swift`, `PersonSourcesSection.swift`.

#### S15-16 — Citations render through their type's formats

**Ships.** The catalog resolver for `core/citationformat`: built-ins from the Source record (with the date, S15-05), the Citation's locator and the graph (`{citation.item}` per map § 6.5); metadata at each level with the `meta.` fallback, through S15-07's batch read; one batch render function for the reads (runs, plain text, missing fields, group key per Citation); one render request for previews; the built-ins table over FFI. A type with no full format renders no text.
**Not this PR.** Any screen (S15-17, S15-21 – S15-23).
**Check.** Handler tests: the census example renders from values spread over Source, Artifact and graph; 300 Citations render in one call with each Source, Artifact and graph read once.
**Context.** Rules "One engine, in Go", "Citations render inside the reads that list them", "One render request for previews", "Today's line stays the fallback", "Build on Spike 16's caches"; map §§ 6.3, 6.5, 6.6.

#### S15-17 — The format editor

**Ships.** The Source type page edits its three formats with Insert field (the built-ins table and the fields, grouped by level, inserting the prefixed form), a live preview against a Source of that type the researcher picks, and the engine's diagnostics shown at their position. Swift does not parse formats.
**Not this PR.** Copy from (S15-18).
**Check.** Bar 5: editing the census full format updates the preview.
**Context.** S15-D6; map § 6.3 "The editor".

### Slice 6 — Citations in the app

#### S15-21 — The provenance sheet shows formatted citations

**Ships.** The provenance records read (S16-30) renders and groups through S15-16's batch function; the sheet shows each record's full citation with its Citations beneath, or today's line when the type has no format. The existing Copy citation (`PVClipboardButton`) copies the full or short form as rich and plain text through one rich-payload write added to `PVClipboardCopyState`. `CitationLineDisplay` stays as the fallback and for its other callers.
**Not this PR.** The Sources section (S15-22).
**Check.** Bar 4's provenance sheet; bar 7 on the sheet; copying pastes italics into a rich-text editor.
**Context.** S15-D9; rules "Grouping is by the rendered full citation", "Citations render inside the reads that list them", "Today's line stays the fallback"; S16-30; `PVProvenanceSheetContent.swift`, `PVProvenanceRecordRow.swift`, `PVClipboardCopyState`, `CitationLineDisplay.swift`.

#### S15-22 — The Sources section groups a Source's Citations by citation

**Ships.** The Sources section read (S16-13) renders and groups through S15-16's batch function, and a Source's Citations that render the same full citation show as one citation with the pieces beneath. Lands after S16-03, S16-06 and S16-07.
**Not this PR.** Re-ordering Sources.
**Check.** Bar 7 in the Sources section, on a person cited 5 times on one household.
**Context.** S15-D3; rule "Grouping is by the rendered full citation"; `PersonSourcesContent.swift`.

#### S15-23 — The Source page previews its citation

**Ships.** Through the one render request, the Source page shows its list entry and a full citation for one of its Citations, names the fields the format expects and is missing, and offers Copy citation through the same rich-payload write as S15-21.
**Not this PR.** Editing the format here (it is on the type, S15-17).
**Check.** Bar 4's Source page half: a census Source with no ED names "Enumeration district" as missing.
**Context.** S15-D1; map § 6.3 "Empty placeholders outside a group".

### Slice 7 — Move

#### S15-24 — Move a bundle to another Source in the engine

**Ships.** The bundle closure, a preflight RPC (the bundle, its Subject counts, the credibility difference) and a move RPC that updates `source_id` on `artifacts`, `evidence_graphs` and `subjects` in one transaction with deferred foreign keys, recording per-row changes so audit (both Sources), `Reads`, search and Spike 16's caches follow from the effects registry; `evidencegraphs.CreateDefault` when the old Source loses its last graph. At Scale a move writes 5 Artifacts, 5 graphs and up to 100 Subjects.
**Not this PR.** Chrome (S15-25); undo; Merge Sources (refused).
**Check.** Handler tests: a single pair, a shared graph pulling in its Artifacts, a bundle of 5 Artifacts and 5 graphs, the last-graph case; refs unchanged; Artifact and graph metadata move with the bundle; both Sources' histories show the move.
**Context.** Rule "A move is the whole bundle"; [`move-artifact-between-sources.md`](move-artifact-between-sources.md); `core/database/effects/registry.go`, `effects/path.go` (`readChangeColumn`); skills `use-catalog-session`, `add-audit-scope`, `add-ffi-handler`. Updates [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) § 4.2 and [`source-layer-data-model.md`](../../source-layer-data-model.md) § 7 (a Subject's and an Artifact's Source can change).

#### S15-25 — Move from an Artifact or a graph

**Ships.** **Move to another Source…** on Artifact rows and graphs on the Source page opens the confirm S15-D10 draws: the target picker, the bundle, and the credibility note when the two Sources' grades differ. After the move, the app lands on the moved Artifact under its new Source.
**Not this PR.** Undo.
**Check.** Bar 8.
**Context.** S15-D10, S15-D1; [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md).
