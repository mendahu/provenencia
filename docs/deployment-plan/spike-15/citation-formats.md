# Citation formats

**Status:** Spike 15 design input, not yet cut into PRs ([`README.md`](README.md)); planned together with the first-class Source date ([`source-provenance-date.md`](source-provenance-date.md)), which is one of the fields a standard citation needs (user, 2026-10-31). The first step ships in Spike 13: the provenance sheet (S13-11, brief S13-D2) shows each record's **citation line** built from the data the catalog has (user, 2026-10-31). This file lists what a standard genealogical citation needs that the catalog does not have yet. It replaces the "Layered citation output" note in [`source-structure-later.md`](../../ideas/source-structure-later.md).

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

## The elements, layer by layer

What an *Evidence Explained* citation draws on, and what Provenencia has for each (2026-10-31). "Lives on" is a first guess for Spike 15's design, not a decision.

**Layer 1: the record itself**

| Element | Example | Have today | Probably lives on |
| --- | --- | --- | --- |
| Creator (person, agency or jurisdiction that made the record) | "England, Yorkshire, York, St Michael le Belfrey Parish"; an author | No (free-text metadata at best) | Source |
| Title, or a description of an unpublished record | *Register of Baptisms*; "Baptisms, 1813 – 1837" | Yes: the Source title | Source |
| Record type | "baptism register", "birth certificate" | Partly: the Source type's label | Source (type) |
| Jurisdiction | country, county, parish or town | No | Source |
| Date of the record or publication | 1837; 15 May 1880 | No: [`source-provenance-date.md`](source-provenance-date.md) | Source |
| Publication details (published works, newspapers) | place, publisher, edition; newspaper name and issue | No (free text) | Source |
| Volume, piece or register number | "Vol. 3"; "HO 107/2353" | Partly: the Artifact label | Artifact or Source |
| Item within the record | certificate or entry number, folio, page, column, line | Partly: the Citation label and the page in the file | Citation |
| Person or event named in the item | "Baptism of Ada Lovelace" | Yes: the Citation label, when filled | Citation |

**Layer 2: how it was seen (the copy used)**

| Element | Example | Have today | Probably lives on |
| --- | --- | --- | --- |
| Format | "digital image", "database", "microfilm", "original" | No | Source or Artifact |
| Website or provider | Ancestry, FamilySearch, a record office | No | Source or Artifact |
| URL | the image or record page | No | Artifact or Citation |
| Date accessed | "accessed 12 Mar 2025" | No | Artifact |
| Film, fiche or image-set number | "FHL film 1,234,567, item 3" | No | Artifact |

**Layer 3: what that copy rests on ("citing …")**

| Element | Example | Have today | Probably lives on |
| --- | --- | --- | --- |
| Repository and its location | The National Archives, Kew | No | Source, or a shared Repository record |
| Collection, series and call number | "HO 107/2353"; "PR/Y/MB/3" | No | Source |
| What it is a copy of | an index entry citing an image citing the register | No ([`source-structure-later.md`](../../ideas/source-structure-later.md) "Derivation links") | A link between Sources |

An analysis after the citation ("informant was the deceased's son") fits the Citation's notes.

Of about 20 elements, 2 are there (title, item label) and 3 are partial (record type, volume through the Artifact label, the page in the file only).

## What a standard citation still needs

| Missing | Today | Needed for |
| --- | --- | --- |
| **Fields per Source type** | One seeded type (`birth_certificate`: document number, record date, issue date). Every other descriptive value is free-text metadata under a key the researcher chose ([`seeded-vocabulary.md`](../../seeded-vocabulary.md)) | Any template: a formatter cannot read "author" or "repository" if every project names them differently |
| **Creator / author, publisher, place of publication** | Free text, if any | Published works, newspapers |
| **Jurisdiction** (country, county, parish or town) | Free text, if any | Census, parish and civil records |
| **Repository and its reference** (archive, call number, film or piece number) | Free text, if any | Every unpublished record |
| **Access layer** (provider, collection, URL, accessed date) | Free text on the Source ([`source-structure-later.md`](../../ideas/source-structure-later.md) "Derivation links" parks Artifact-level access metadata) | The "image, Ancestry, accessed …" layer |
| **What it is a copy of** | Not modeled ([`source-structure-later.md`](../../ideas/source-structure-later.md) "Derivation links") | The "citing …" layer |
| **Source date** | Designed, not built: [`source-provenance-date.md`](source-provenance-date.md) (descoped from Spike 13 on 2026-10-31 to land with these fields) | Dates in the note, and the source list order |
| **Printed page labels** | The locator can hold `page_label`, but the composer writes only `artifact_page` | "p. 23" instead of "image 37"; folio, entry and line numbers |
| **Entry or line within a page** | Not modeled; a region selector marks where, not which entry | Census line numbers, register entry numbers |
| **Formatter** | None | A Go package (for example `core/citationformat`) with one template per Source type, driven by a registry like the seeded vocabulary, so the Mac app, exports and any later report share it. It replaces `CitationLineDisplay` |
| **Short form and source list entry** | None | Repeated mentions and a bibliography |

## How it could arrive

1. **The Source date and the first fields in one Source schema pass** ([`source-provenance-date.md`](source-provenance-date.md)): one migration and one Source page design.
2. **Seed the fields** for the common types (census, parish register, civil certificate, newspaper item, online image, published book) with [`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md): new projects only, and existing ones keep their keys.
3. **Printed page labels in the composer**, so the locator says "p. 23".
4. **The formatter in Go** with a template per seeded type and a fallback to today's line for any other type or a missing field. The citation line on the provenance sheet reads from it.
5. **A citation preview on the Source page** that shows the formatted note and names the fields it is missing, so researchers can fill them.
6. **Copy citation** (full or short) from the provenance sheet and the Source page.
7. **Move an Artifact to another Source**, with the Evidence graphs built on it, so researchers can re-file Artifacts as their Sources take on these fields ([`move-artifact-between-sources.md`](move-artifact-between-sources.md)).

## Open questions

- Which elements are first-class columns (like the Source date) and which are seeded metadata per Source type.
- Whether a repository is its own record that many Sources share.

- Which Source types to template first: the product owner's dogfood Sources decide.
- Whether a template's fields live on the Source, the Artifact or the Citation (a census piece number is per Source; a folio is per Citation).
- Whether researchers can edit a template, or only the fields under it.
