# Citation formats

**Status:** idea only — not roadmapped. The first step ships in Spike 13: the provenance sheet (S13-11, brief S13-D2) shows each record's **citation line** built from the data the catalog has (user, 2026-10-31). This file lists what a standard genealogical citation needs that the catalog does not have yet. It replaces the "Layered citation output" note in [`source-structure-later.md`](source-structure-later.md).

## The convention

Genealogy cites with *Evidence Explained* (Elizabeth Shown Mills), which builds on the Chicago Manual of Style's notes style. Three things set it apart:

- **Templates per kind of record.** A census, a parish register, a civil certificate, a newspaper item and an online database image are each cited in their own pattern.
- **Layers.** A citation says what was used and what that rests on, separated by semicolons: *"1851 England census, Yorkshire, York, St Michael le Belfrey, folio 23, p. 4; image, Ancestry, accessed 12 Mar 2025; citing The National Archives, HO 107/2353."*
- **Three forms.** A full first reference note, a short form for later mentions ("1851 census, York, St Michael le Belfrey, fol. 23"), and a source list entry for a bibliography.

APA and MLA are not used in genealogy.

## What ships in S13-11

The citation line is the Source title, the Artifact label, the Citation label and the page, each part present only when it has a value, joined by commas: *"Register of Baptisms, St Michael le Belfrey, Vol. 3, Baptism of Ada, p. 23"*.

- The page is the locator's printed page label ("p. 23") when the locator has one, else its page in the file ("image 37"). The citation composer writes only the page in the file today, so most lines read "image N".
- Go reads the parts (`GetConclusionDetail`'s `ReconcilerOutcomeDetail`: Artifact label, Citation label, page in the file, printed page label); Swift joins and localizes them (`CitationLineDisplay`).
- It is one line for every Source type. It does not know the record's kind, its jurisdiction, its repository or how it was accessed.

## What a standard citation still needs

| Missing | Today | Needed for |
| --- | --- | --- |
| **Fields per Source type** | One seeded type (`birth_certificate`: document number, record date, issue date). Every other descriptive value is free-text metadata under a key the researcher chose ([`seeded-vocabulary.md`](../seeded-vocabulary.md)) | Any template: a formatter cannot read "author" or "repository" if every project names them differently |
| **Creator / author, publisher, place of publication** | Free text, if any | Published works, newspapers |
| **Jurisdiction** (country, county, parish or town) | Free text, if any | Census, parish and civil records |
| **Repository and its reference** (archive, call number, film or piece number) | Free text, if any | Every unpublished record |
| **Access layer** (provider, collection, URL, accessed date) | Free text on the Source ([`source-structure-later.md`](source-structure-later.md) "Derivation links" parks Artifact-level access metadata) | The "image, Ancestry, accessed …" layer |
| **What it is a copy of** | Not modeled ([`source-structure-later.md`](source-structure-later.md) "Derivation links") | The "citing …" layer |
| **Source date** | Being built: S13-20 adds an optional publication date to `sources` | Dates in the note, and the source list order |
| **Printed page labels** | The locator can hold `page_label`, but the composer writes only `artifact_page` | "p. 23" instead of "image 37"; folio, entry and line numbers |
| **Entry or line within a page** | Not modeled; a region selector marks where, not which entry | Census line numbers, register entry numbers |
| **Formatter** | None | A Go package (for example `core/citationformat`) with one template per Source type, driven by a registry like the seeded vocabulary, so the Mac app, exports and any later report share it. It replaces `CitationLineDisplay` |
| **Short form and source list entry** | None | Repeated mentions and a bibliography |

## How it could arrive

1. **Seed the fields** for the common types (census, parish register, civil certificate, newspaper item, online image, published book) with [`add-seeded-vocabulary`](../../.agents/skills/add-seeded-vocabulary/SKILL.md): new projects only, and existing ones keep their keys.
2. **Printed page labels in the composer**, so the locator says "p. 23".
3. **The formatter in Go** with a template per seeded type and a fallback to today's line for any other type or a missing field. The citation line on the provenance sheet reads from it.
4. **A citation preview on the Source page** that shows the formatted note and names the fields it is missing, so researchers can fill them.
5. **Copy citation** (full or short) from the provenance sheet and the Source page.

## Open questions

- Which Source types to template first: the product owner's dogfood Sources decide.
- Whether a template's fields live on the Source, the Artifact or the Citation (a census piece number is per Source; a folio is per Citation).
- Whether researchers can edit a template, or only the fields under it.
