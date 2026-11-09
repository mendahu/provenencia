# Evidence Explained citation map

**Status:** Spike 15 design input, not yet cut into PRs ([`README.md`](README.md)). It lists the kinds of record *Evidence Explained* (EE) gives citation models for, groups them into the citation patterns Provenencia has to build, and maps each pattern's elements to data the catalog has or will need. The last section lists every missing field and how to add it. **Revised 2026-11-09 (user):** the first draft added tables and columns for contributors, jurisdiction, repositories, reference parts and Artifact access fields; this version uses the existing Source types, metadata fields and type → field suggestions instead. Beyond the planned Source date, the only new schema repeats the Source's metadata structure (values, layout, type templating) on the Artifact (the copy used) and the Evidence graph (the item cited) (§ 6.6, § 6.9). The Citation composer gets no new fields. It extends [`citation-formats.md`](citation-formats.md), which sets out the convention and the six-step path; this doc is the inventory behind steps 1, 2 and 4 of that path.

**Goal (user, 2026-11-08):** a researcher builds an Evidence graph, promotes its Subjects, and gets a correct EE citation with no further typing. Every field below either already exists, can be read from the graph, or is a seeded metadata field the researcher fills once: on the Source, or on the Artifact or Evidence graph when it differs per copy or per item.

## Read this first

- **Edition.** The list follows the chapter layout of *Evidence Explained: Citing History Sources from Artifacts to Cyberspace* (Elizabeth Shown Mills), 3rd edition. It was written from knowledge of the book, not from a copy beside the keyboard. Check the chapter list and the record kinds against the edition on hand before the seeded citation formats are written, and add any kind this list misses to its pattern.
- **Not a copy of the book.** EE has hundreds of "QuickCheck" models, many of them variants of the same shape. This doc does not reproduce them. It names the record kinds and paraphrases the shape each one shares. Every example citation here is invented for this doc (made-up people, records and numbers) and is written in EE style, not taken from the book.
- **Keep it that way.** Every later doc, template, test and UI string follows the spike's working rule on *Evidence Explained* ([`README.md`](README.md#working-rules)): our own templates, invented examples, our own names, no implied endorsement.
- **Patterns, not models.** Each *pattern* (about 17) becomes a seeded Source type with its citation formats as data (§ 6.3), with small per-kind variations, not one per EE model. The catalog in § 3 maps each record kind to its pattern.

## Contents

1. [Two kinds of information: your copy and the record](#1-two-kinds-of-information-your-copy-and-the-record)
2. [How an EE citation is put together](#2-how-an-ee-citation-is-put-together)
3. [The record kinds EE covers](#3-the-record-kinds-ee-covers)
4. [Where the data comes from: the element codes](#4-where-the-data-comes-from-the-element-codes)
5. [The patterns](#5-the-patterns)
6. [Missing fields and how to add them](#6-missing-fields-and-how-to-add-them)
7. [Open questions](#7-open-questions)

---

## 1. Two kinds of information: your copy and the record

A Provenencia Citation and an EE citation answer different questions (user, 2026-11-09):

- **A Provenencia Citation says where the evidence is in *your* copy:** Artifact `ART-…`, image 37, this region. It is for the researcher and the app, to get back to the exact pixels an Observation came from. It assumes you hold the file.
- **An EE citation says how *anyone* finds the record again:** the record in the world, in its own numbering (Deed Book C, page 214; ED 236, sheet 14, line 12), then the path to a copy (the provider, URL and accessed date, or a film number) and where the original lives. It assumes the reader has nothing.

The two overlap only by accident: "image 37" means nothing without the file, and "folio 23, entry 182" means nothing to the viewer. The locator already keeps the two apart for pages: `artifact_page` is where the page is in the file, and `page_label` is how the page numbers itself in the original ([`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) § 3.2). EE needs that second kind everywhere, so the metadata in this plan sits **beside** the locator rather than replacing it:

| Level | Your copy (the locator side) | Finding the record again (the metadata side) |
| --- | --- | --- |
| Source | — | What the record is and where the original lives: creator, jurisdiction, title, volume, repository |
| Artifact | The File | How to get this copy again: format, provider, URL, accessed date, film number, citing |
| Evidence graph | — | The item in the record's own numbering: entry, line, dwelling, family, folio, column (§ 6.9) |
| Citation | Page in the file, region | The printed page (`page_label`), the one item detail that can differ inside a graph |

What follows from it:

- **Citation formats do not read the locator,** except `{citation.image}`, the image number in the online-copy layer ("image 37 of 412") when the Artifact is a provider's image set.
- **Missing access details are flagged, never invented.** A scan or download with no provider, URL or film number gets a citation whose copy layer is incomplete, and the preview says so.
- **Copy details are captured when the Artifact is added,** while the researcher still knows where it came from: the import flow shows the Artifact's metadata fields, not only the Artifact page afterwards. This also raises the value of filling them automatically ([`registry-source-import.md`](../../ideas/registry-source-import.md)).
- **A physical original has no access path;** `holder` ("privately held by …"), on the Source, takes the place of the copy layer.
- **One EE citation usually covers several Provenencia Citations.** A researcher draws one Citation per cell or name they want to highlight; EE cites the row or the entry once. § 6.9 covers what that means for data entry and display.

---

## 2. How an EE citation is put together

Summarised from [`citation-formats.md`](citation-formats.md) "The convention", with the rules the seeded formats follow.

**Three forms per record.**

| Form | What it identifies | Grain |
| --- | --- | --- |
| **Source list entry** (bibliography) | The source as a whole: the register, the census of a county, the book, the database | Source |
| **First reference note** (full note) | The exact item used: entry, page, line, the person named | Source + Artifact + Citation |
| **Subsequent note** (short note) | The same item again, abbreviated | Source + Citation |

**Layers.** A full note is up to three layers separated by semicolons: (1) the record and the item in it; (2) how it was seen (the copy: "digital image, Ancestry (URL : accessed date)", or a microfilm); (3) what that copy rests on ("citing …", the repository and its reference). A record seen in the original has layers 1 and 3 only ("…; County Clerk's Office, Minneapolis").

**Order flips between forms.** A source list entry runs from large to small and starts with what it is filed under, often the jurisdiction ("Kansas. Ottawa County. *Deed Book C*."). A note runs small to large ("Ottawa County, Kansas, Deed Book C: 214"). Provenencia does not reverse text values (§ 6.3), so a seeded list format prints the jurisdiction as typed; a researcher who wants the large-to-small order there can add a second field for it.

**Title styling.** The title of a published work is italic; a formal title of an unpublished manuscript is in quotation marks; a generic description ("baptisms, 1813–1837", "letter") is plain. The pattern decides which: published patterns style the title, record patterns print it as a description.

**Names of people.** Authors are written in natural order in notes ("Alice Smith") and inverted for the first author in the source list ("Smith, Alice"). People *in* the record are named as the record names them ("entry for Wm. Robins"), not as the researcher concluded.

**Dates.** EE writes dates day-month-year ("3 March 1830"), abbreviates months only in short notes, and uses several dates in one citation: the record's or publication's date, the item's date (the baptism), and the date accessed.

---

## 3. The record kinds EE covers

Chapters 1 and 2 of EE are the fundamentals of evidence analysis and of citation; they have no record kinds of their own. Chapters 3 to 14 each cover a family of records. The **Pattern** column names the template in § 5 that each kind uses; in Provenencia, a pattern is a seeded Source type with its suggested fields (§ 6).

### Chapter 3 — Archives and artifacts

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Archived manuscript item (letter, diary, journal, ledger, memoir) | P1 Archived manuscript | Item → collection → repository |
| Manuscript collection cited as a whole | P1 | Source list grain only |
| Archived maps, plats, drawings | P1 | Map title and date as the item |
| Archived photographs, portraits | P1 | Photographer, studio, date |
| Archived audio or video recordings | P1 | Speaker, recorder, date |
| Family Bible records | P8 Private holdings | Bible's own imprint (publisher, year) plus the record pages; who holds it |
| Family letters, diaries and papers in private hands | P8 | "Privately held by …" |
| Heirlooms and artifacts (samplers, quilts, engraved objects) | P8 | Description of the object |
| Private photographs, cabinet cards, tintypes | P8 | Medium and studio |
| Personal interviews and oral histories | P9 Interview | Interviewee, interviewer, date, place, where the recording is |
| Personal correspondence and email to the researcher | P9 | Writer, recipient, date |
| Research reports, family group sheets, compiled family files | P8 | Compiler as creator |
| DNA test results | P8 | **Verify:** the 3rd edition's treatment of DNA results was thin; Provenencia needs a pattern for its `dna_match_report` type whatever the book says |

### Chapter 4 — Business and institutional records

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Bank, company and corporate records | P1 | Corporate body as creator |
| Funeral home and mortuary records | P1 | Firm, register title, entry |
| Hospital, asylum and medical records | P1 | Often restricted; repository matters |
| Insurance records | P1 | Policy number as the item |
| Lineage society applications (DAR, SAR and similar) | P1 | National number as the item |
| School, college and university records | P1 | Register or transcript |
| Orphanage, poorhouse and charity records | P1 | |
| Fraternal and benevolent society records (lodges, unions) | P1 | |
| Business records seen on an online image or database | P1 + layer 2 (P14 / P13) | |

### Chapter 5 — Cemetery records

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Grave markers, headstones, monuments | P7 Cemetery marker | Cemetery and its place, the marker, who photographed or read it and when |
| Cemetery office records: interment register, plot map, deed | P1 | Cemetery or association as creator |
| Sexton's and burial registers | P1 | |
| Published cemetery transcriptions and abstracts | P10 Book | Compiler as author |
| Online memorial sites (memorial pages with photographs) | P13 Web database | Memorial id; the memorial's creator; "citing" the cemetery |

### Chapter 6 — Census records

| Record kind | Pattern | Notes |
| --- | --- | --- |
| National population schedules (U.S. federal 1790–1950 and similar) | P5 Census | Year, jurisdiction, schedule, ED, sheet or page, dwelling, family, line, person |
| Nonpopulation schedules: agricultural, industry or manufactures, mortality, social statistics | P5 | Schedule name changes |
| U.S. slave schedules (1850, 1860) | P5 | Owner as the named person |
| Special schedules (veterans 1890, Indian schedules) | P5 | |
| State and territorial censuses | P5 | State as the creating jurisdiction |
| Local and municipal censuses | P5 | |
| Canadian censuses | P5 | District and sub-district numbers |
| British censuses (England, Wales, Scotland) | P5 | Piece, folio, page, schedule number; repository reference (HO 107, RG 9 and similar) |
| Other national censuses (France, Germany and others) | P5 | Commune or parish as the smallest part |
| Census seen on microfilm | P5 + P15 Microfilm layer | Microfilm publication and roll |
| Census seen as an online image | P5 + P14 Online image layer | |
| Census index or database only (no image) | P13 Web database | "citing" the schedule |

### Chapter 7 — Church records

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Parish registers: baptisms, marriages, burials | P6 Church register | Church and its place, register title and span, volume, page, entry |
| Church books: membership rolls, minutes, communion and confirmation lists | P6 | |
| Bishop's transcripts and diocesan copies | P6 | The copy is the record cited; the diocese is the repository |
| Church records held by an archive | P6 | Repository and its reference |
| Church records on microfilm | P6 + P15 | |
| Church records as online images | P6 + P14 | |
| Published or transcribed church registers | P10 Book | Compiler or society as author |

### Chapter 8 — Local and state records: courts and governance

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Court minute books, order books, dockets | P2 Local record book | Jurisdiction, court, book, page |
| Court case files (loose papers) | P3 Case file | Case or file number, parties, the document in the file |
| Council and commission minutes; ordinances | P2 | |
| Coroner's inquests | P3 | |
| Jury lists, bonds, oaths | P2 / P3 | Book form or loose |
| State court records (supreme court case files) | P3 | |

### Chapter 9 — Local and state records: licenses, registrations, rolls and vital records

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Birth, death and marriage certificates (issued documents) | P4 Vital certificate | Issuing authority, certificate or file number, person, date, office |
| Birth, death and marriage registers (book form) | P2 | Register volume, page, entry |
| Delayed and amended birth certificates | P4 | Filing date beside the event date |
| Marriage licenses, bonds and returns | P2 / P3 | Book form or loose |
| Civil registration indexes (national indexes such as GRO) | P13 / P10 | Index quarter, district, volume, page |
| Tax rolls and assessment lists | P2 | Year, district, page, line |
| Voter registrations and poll books | P2 | |
| Business and professional licenses | P2 | |
| Militia and muster rolls kept locally | P2 | |
| School census and enumeration rolls | P2 | |
| Divorce records | P3 | Usually court case files |

### Chapter 10 — Local and state records: property and probate

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Deed books (deeds, mortgages, leases, powers of attorney) | P2 | Book and page; grantor to grantee; date |
| Land grants, patents and surveys (state level) | P2 / P3 | Grant or patent number |
| Plat books and maps | P2 | |
| Will books (recorded copies) | P2 | |
| Original wills | P3 | Packet or file |
| Probate packets and estate files: inventories, accounts, bonds | P3 | Estate file number; the document in it |
| Guardianship records | P2 / P3 | |
| Orphans' court records | P2 / P3 | |

### Chapter 11 — National government records

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Military service records (compiled service records) | P1 Archived manuscript (national variant) | Soldier, unit, war; record group; national archive |
| Pension and bounty-land application files | P1 | File or certificate number |
| Draft and registration cards | P1 | Board, card, registrant |
| Passenger lists and ship manifests | P2 (national variant) | Ship, port, arrival date, page or sheet, line |
| Naturalization records (federal courts) | P3 | Petition or certificate number |
| Passport applications | P1 | |
| Federal land entry files and homestead files | P3 | Land office, final certificate number |
| Freedmen's Bureau and similar agency records | P1 | |
| Indian rolls and agency records | P2 | Roll and enrollment number |
| Social Security applications and claims | P1 | |
| Prison and penitentiary records | P1 | |
| Equivalent records of other national archives (TNA, LAC, others) | P1 | The archive's own reference scheme, for example TNA "WO 97/1234" |
| National records on microfilm or online | P1 + P15 / P14 | |

### Chapter 12 — Publications: books, CDs, DVDs, maps, leaflets and videos

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Authored book | P10 Book | Author, title, place, publisher, year, page |
| Edited, compiled or translated book | P10 | Editor, compiler, translator roles |
| Multi-volume work; volume in a series | P10 | Volume and series |
| Chapter or essay in an edited book | P10 | Chapter author and title inside the book |
| Reprint and later edition | P10 | Original year and reprint year |
| Book seen as an online image (digitized book) | P10 + P14 | Website, URL, accessed date |
| CD-ROM and DVD publications | P10 | Medium named |
| Published maps and atlases | P10 | Cartographer, scale, sheet |
| Leaflets, brochures, pamphlets | P10 | |
| Films, videos and documentaries | P10 | Producer or director, distributor |

### Chapter 13 — Publications: legal works and government documents

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Statutes and session laws | P16 Legal publication | Legal citation style: jurisdiction, compilation, chapter, section, page |
| Codes and compiled laws | P16 | |
| Legislative journals, reports and documents | P16 | |
| Court reports (published cases) | P16 | Case name, volume, reporter, page, year |
| Published government documents (serial sets, state papers, gazettes) | P16 | |

### Chapter 14 — Publications: periodicals, broadcasts and web miscellanea

| Record kind | Pattern | Notes |
| --- | --- | --- |
| Journal articles (genealogical and historical journals) | P11 Periodical article | Author, article title, journal, volume, issue, date, pages |
| Magazine articles | P11 | |
| Newspaper articles, notices, obituaries, advertisements | P12 Newspaper item | Item title or description, paper, place, date, page, column |
| Newsletters | P11 | |
| Newspapers seen as online images | P12 + P14 | |
| Radio and television broadcasts; podcasts | P11 | Program, network or host, broadcast date |
| Websites and web pages (articles, blog posts) | P13 Web page | Creator, page title, website, URL, accessed date |
| Online databases and indexes | P13 Web database | Database title, website, URL, accessed date, entry, "citing" |
| Online family trees | P13 | Tree owner, tree title, profile |
| Message boards, mailing lists, social media posts | P13 | Poster, thread, date |
| Email (as a published or archived message) | P9 | |

### Provenencia's own types against these

The product seed today is one type (`birth_certificate`); the horizon list is in [`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 2.1. Each horizon type falls under a pattern:

| Seeded type | Pattern |
| --- | --- |
| `birth_certificate` | P4 |
| `census` | P5 |
| `photograph` | P8 (privately held) or P1 (archived) |
| `oral_testimony`, `interview` | P9 |
| `dna_match_report` | P8 (DNA variant) |
| `book` | P10 |
| `website` | P13 |
| `gedcom_file` | P8 (compiled file) or P13 (downloaded online tree) |

Kinds EE covers that have no seeded type yet, and that the dogfood Sources will want first: church register (P6), newspaper item (P12), deed and probate records (P2, P3), cemetery marker (P7), death and marriage certificates (P4), passenger list (P2 national), military and pension files (P1).

---

## 4. Where the data comes from: the element codes

Every element in § 5 names its source with one of these codes. **Have** means stored and readable today. **Graph** means readable from the Evidence graph the researcher already built. **Planned** means designed in another Spike 15 doc. **Seed** means a product-seeded Source metadata field, most of them already on the horizon list; § 6 says how.

### Have

| Code | What | Where |
| --- | --- | --- |
| `S.title` | Source title (required headline) | `sources.title` |
| `S.description` | Source description | `sources.description` |
| `S.type` | Source type label and key | `sources.source_type_id` → `source_types` |
| `S.meta[key]` | Free-text metadata value under a researcher's or seed's key | `source_metadata.value_text`; seeded keys exist only for `birth_certificate` in new catalogs. The `M.` codes below are these values under product keys |
| `A.label` | Artifact label (headline) | `artifacts.label` |
| `A.description` | Artifact description | `artifacts.description` |
| `F.media` | The File's media type (image, PDF, audio) | `files.media_type` |
| `C.label` | Citation label | `citations.label` |
| `C.page` | Page in the file ("image 37") | locator `page` selector, `artifact_page` |
| `C.page_label` | Printed page label ("23", "iv", "23a") | locator `page` selector, `page_label`; stored, but the composer does not write it yet |
| `C.transcription` | The researcher's transcription | `citations.transcription` |

### Graph (read from the Observations the Citation supports)

| Code | What | Read from |
| --- | --- | --- |
| `G.person` | Principal person named in the item, **as recorded** | `name` Observation on the person Subject cited by this Citation |
| `G.event` | Event type of the item (baptism, death, sale) | `event_type` Observation on the event Subject |
| `G.date` | Date of the event in the item | `date` Observation (DateValue) on that event |
| `G.place` | Place of the event, as recorded | `toponym` of the place Subject linked to that event by a `location` |
| `G.parties` | Other named participants (grantor and grantee, bride and groom, head of household) | `participation` Subjects with their `role` terms |

The graph supplies the "entry for …" or "baptism of …" part of a full note. It reads the Source's own Subjects, never the canonical entities, because EE names people as the record names them. A Citation that supports several Subjects needs a rule for the principal one (§ 7).

### Planned

| Code | What | Where |
| --- | --- | --- |
| `S.date` | The Source's date: publication, or the date the record was made | `sources` → `date_values`, [`source-provenance-date.md`](source-provenance-date.md) |

### Seed (detailed in § 6)

`M.key` is a value under a **product-seeded metadata field** (`origin = 'provenencia'`) with that key: on the Source (`source_metadata`), or on the Artifact or Evidence graph when it differs per copy or per item (§ 6.6). The format engine takes the most specific one. The type → field suggestions (`source_type_metadata_fields`) give each seeded Source type the fields its pattern reads, so the Source page's existing Suggestions board is what asks the researcher for them.

Fields already on the horizon list in [`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 2.2:

| Group | Keys |
| --- | --- |
| Who made it | `creator`, `author`, `photographer`, `studio`, `interviewer`, `interviewee`, `testing_company`, `software` |
| Where | `jurisdiction` (written smallest first: "York, Yorkshire, England") |
| Publication | `publisher`, `publication_place`, `edition`, `series`, `volume`, `site_name` |
| Archive | `repository` (name and place: "Kansas Historical Society, Topeka"), `collection`, `record_group`, `call_number`, `box`, `folder` |
| Reference numbers | `certificate_number`, `enumeration_district`, `sheet`, `line_number`, `dwelling_number`, `family_number`, `roll`, `kit_number`, `match_identifier` |
| Copy | `medium`, `url`, `accessed_date` |

Fields to add to the seed:

| Group | Keys |
| --- | --- |
| Who made it | `author_sort` ("Smith, Alice", for the source list), `editor`, `compiler`, `translator`, `recipient` |
| Title | `short_title` (override for short notes), `published_in` (journal, newspaper, the book holding a chapter, a statute compilation), `coverage` (span such as "1813–1837") |
| Reference numbers | `book`, `folio`, `entry_number`, `schedule`, `schedule_number`, `piece`, `case_number`, `issue`, `column`, `page_range`, `section`, `lot`, `chapter`, `record_id`, `item_number` |
| Copy | `format` (digital image, microfilm, database, photocopy, …), `provider`, `database_title`, `film_publication`, `film_number`, `citing` |
| Holding | `holder` (who holds a private original, and where) |

The copy fields (`format`, `provider`, `database_title`, `url`, `accessed_date`, `film_publication`, `roll`, `film_number`, `citing`) are usually filled on the Artifact, and the item numbers (`line_number`, `dwelling_number`, `family_number`, `schedule_number`, `entry_number`, `folio`, `column`, `record_id`) on the Evidence graph.

---

## 5. The patterns

Each pattern has its three forms shown with an invented example, then an element table. In the element tables, **Forms** says which forms use the element: **L** = source list, **F** = full note, **S** = short note.

### P1 — Archived manuscript

Unpublished records held by an archive or library: letters, diaries, institutional and business records, national government files.

> **L:** Smith Family Papers, 1840–1910. MS 1234. Kansas Historical Society, Topeka.
>
> **F:** John Smith (Fort Scott, Kansas) to Mary Smith, letter, 4 June 1862; Smith Family Papers, 1840–1910, MS 1234, box 2, folder 7; Kansas Historical Society, Topeka.
>
> **S:** John Smith to Mary Smith, 4 June 1862, Smith Family Papers, box 2, folder 7.

National variant (Chapter 11): the creator is the agency, and the archival reference is the record group and series, for example "…; Record Group 15, Records of the Department of Veterans Affairs; National Archives, Washington, D.C."

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Creator of the item (writer, agency, firm) | F S | `M.creator`, else `G.parties` (writer) | Seed |
| Recipient (letters) | F S | `G.parties` (recipient role) | Graph; needs a `recipient` role term |
| Item description ("letter", "pension file") | F S | `C.label`, else `S.type` label | Partial |
| Item date | F S | `G.date`, else `S.date` | Graph / Planned |
| Collection title | L F S | `M.collection` | Seed (horizon key) |
| Collection span | L F | `M.coverage` | Seed |
| Call or MS number | L F | `M.call_number` | Seed |
| Record group, series | L F | `M.record_group`, `M.series` | Seed |
| Box, folder, item | F S | `M.box`, `M.folder`, `M.item_number` | Seed |
| Repository and its location | L F | `M.repository` | Seed |

### P2 — Local record book

A register kept by a public office in book form: deed books, will books, court minutes, marriage registers, tax rolls, passenger manifests (national variant).

> **L:** Kansas. Ottawa County. Deed Records, 1866–1920. Register of Deeds, Minneapolis.
>
> **F:** Ottawa County, Kansas, Deed Book C: 214–15, John Smith to William Robins, 3 May 1879; Register of Deeds, Minneapolis.
>
> **S:** Ottawa Co., Kans., Deed Book C: 214.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Jurisdiction | L F S | `M.jurisdiction` | Seed |
| Office or court that kept it | L F | `M.creator` | Seed |
| Record series and span ("Deed Records, 1866–1920") | L | `S.title` + `M.coverage` | Have / Seed |
| Book or volume | F S | `M.book` or `M.volume` | Partial: often typed into `A.label` today |
| Page (printed) | F S | `C.page_label` | Have, composer gap |
| Entry, line, folio | F S | `M.entry_number`, `M.line_number`, `M.folio` | Seed |
| Parties and nature of the record ("John Smith to William Robins") | F | `G.parties` with roles (grantor, grantee) | Graph; needs role terms |
| Item date | F | `G.date` | Graph |
| Repository (the office today) | L F | `M.repository` | Seed |

### P3 — Case file or loose-paper packet

Court case files, probate packets, original wills, naturalization petitions, land entry files. The file is the unit; the item is one document inside it.

> **L:** Kansas. Ottawa County. Probate Court. Estate Files. Probate Court, Minneapolis.
>
> **F:** Ottawa County, Kansas, Probate Court, estate file 112, John Smith (1885), inventory, 2 March 1885; Probate Court, Minneapolis.
>
> **S:** Ottawa Co., Kans., estate file 112, John Smith, inventory.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Jurisdiction, court or office | L F S | `M.jurisdiction`, `M.creator` | Seed |
| File type and number | F S | `M.case_number` | Seed |
| Principal party and year | F S | `G.person` (or parties), `S.date` | Graph / Planned |
| Document in the file | F S | `C.label` | Partial |
| Document date | F | `G.date` | Graph |
| Repository | L F | `M.repository` | Seed |

### P4 — Vital certificate

An issued civil certificate: birth, death, marriage, delayed birth. In Provenencia this is usually one Source per certificate.

> **L:** Kansas. State Board of Health. Birth Certificates. Office of Vital Statistics, Topeka.
>
> **F:** Kansas State Board of Health, birth certificate no. 1902-04417, Ada Robins, 12 February 1902; Office of Vital Statistics, Topeka.
>
> **S:** Kans. birth cert. no. 1902-04417, Ada Robins.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Issuing authority and jurisdiction | L F S | `M.creator`, `M.jurisdiction` | Seed |
| Certificate type | L F S | `S.type` label | Have |
| Certificate or file number | F S | `M.certificate_number` | Seed (horizon key) |
| Registration or volume and page (registers) | F | `M.volume`, `C.page_label` | Seed / Have |
| Person named | F S | `G.person` | Graph |
| Event date | F | `G.date` | Graph |
| Filing or issue date (delayed certificates) | F | `S.date` | Planned |
| Repository (issuing office) | L F | `M.repository` | Seed |

### P5 — Census

Every census in Chapter 6. The richest set of reference numbers.

> **L:** Kansas. Ottawa County. 1880 U.S. census, population schedule. Digital images. Ancestry. https://www.ancestry.com : 2025.
>
> **F:** 1880 U.S. census, Ottawa County, Kansas, population schedule, Minneapolis, enumeration district (ED) 236, p. 14 (stamped), dwelling 112, family 118, Ada Robins; digital image, Ancestry (https://www.ancestry.com : accessed 12 March 2025); citing NARA microfilm publication T9, roll 393.
>
> **S:** 1880 U.S. census, Ottawa Co., Kans., pop. sch., Minneapolis, ED 236, p. 14 (stamped), dwell. 112, fam. 118, Ada Robins.

British variant: "1851 census of England, Yorkshire, York, St Michael le Belfrey, folio 23, p. 4, schedule 17, household of John Smith; …; citing The National Archives, HO 107/2353."

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Census year | L F S | `S.date` (year) | Planned |
| Census authority ("U.S.", state, "of England") | L F S | `M.creator`, else the jurisdiction's top part | Seed |
| Jurisdiction: state, county, civil division | L F S | `M.jurisdiction` | Seed |
| Schedule (population, agriculture, mortality, slave) | L F S | `M.schedule` | Seed |
| Enumeration district; district and sub-district (Canada) | F S | `M.enumeration_district` | Seed |
| Sheet, page, folio; which numbering is used (stamped, penned) | F S | `M.sheet` / `C.page_label` / `M.folio` | Partial |
| Piece (British) | F | `M.piece` | Seed |
| Schedule number (British household) | F | `M.schedule_number` | Seed |
| Dwelling, family | F S | `M.dwelling_number`, `M.family_number` | Seed |
| Line | F | `M.line_number` | Seed |
| Person or household named | F S | `G.person` (head of household, or the person cited) | Graph |
| Layer 2 and 3 (image, film, citing) | L F | P14 / P15 | Seed |

### P6 — Church register

> **L:** St Michael le Belfrey (York, Yorkshire, England). Register of Baptisms, 1813–1837. Borthwick Institute for Archives, York.
>
> **F:** St Michael le Belfrey (York, Yorkshire, England), *Register of Baptisms, 1813–1837*, vol. 3, p. 23, no. 182, baptism of Ada Smith, 3 March 1830; Borthwick Institute for Archives, York.
>
> **S:** St Michael le Belfrey, Register of Baptisms, 3: 23, no. 182.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Church or congregation (creator) | L F S | `M.creator` | Seed; an `organization` Subject later ([`organizations.md`](../../ideas/organizations.md)) |
| Church's place | L F | `M.jurisdiction` | Seed |
| Register title or description, and its span | L F S | `S.title` + `M.coverage` | Have / Seed |
| Volume, page, entry number | F S | `M.volume`, `C.page_label`, `M.entry_number` | Partial |
| Event and person ("baptism of Ada Smith") | F | `G.event` + `G.person` | Graph |
| Event date | F | `G.date` | Graph |
| Repository (parish chest, diocesan or county archive) | L F | `M.repository` | Seed |

### P7 — Cemetery marker

> **L:** Kansas. Ottawa County. Minneapolis. Highland Cemetery. Grave markers.
>
> **F:** Highland Cemetery (Minneapolis, Ottawa County, Kansas), Ada Robins marker, section 4, lot 12; photographed by the author, 14 May 2024.
>
> **S:** Highland Cemetery (Minneapolis, Kans.), Ada Robins marker.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Cemetery name | L F S | `S.title` | Have / Seed |
| Cemetery location | L F S | `M.jurisdiction` | Seed |
| Person on the marker | F S | `G.person` | Graph |
| Section, lot, row | F | `M.section`, `M.lot` | Seed |
| Who read or photographed it, and when | F | `M.holder` / `M.photographer`, `M.accessed_date` or `S.date` | Seed |

### P8 — Private holdings, artifacts and compiled files

Family Bibles, letters and photographs in private hands, heirlooms, research reports, GEDCOM files, DNA results.

> **L:** Smith Family Bible Records, 1850–1931. *The Holy Bible*. New York: American Bible Society, 1850. Privately held by Jane Doe, Calgary, Alberta.
>
> **F:** Smith Family Bible Records, 1850–1931, *The Holy Bible* (New York: American Bible Society, 1850), "Births" page; privately held by Jane Doe, Calgary, Alberta, 2024.
>
> **S:** Smith Family Bible Records, "Births" page.

DNA variant: "AncestryDNA, autosomal DNA test results for Jane Doe, match list entry for J. R. (shared 212 cM), report viewed 12 March 2025; privately held by Jane Doe, Calgary, Alberta."

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Title or description of the record | L F S | `S.title` | Have / Seed |
| Span | L F | `M.coverage` | Seed |
| Creator or compiler (writer, photographer, studio, testing company) | L F S | `M.creator`, `M.photographer`, `M.studio`, `M.testing_company` | Seed (horizon keys) |
| Imprint of the containing book (Bible) | L F | `M.published_in` + `M.publisher`, `M.publication_place` | Seed |
| Item: page, letter date, match entry | F S | `C.label`, `G.date` | Partial |
| Medium (cabinet card, tintype) | F | `M.medium` | Seed (horizon key) |
| Holder and place, year seen | L F | `M.holder`, `M.accessed_date` | Seed |
| Kit or match identifiers | F | `M.kit_number`, `M.match_identifier` | Seed (horizon keys) |
| Software and export date (GEDCOM) | F | `M.software`, `S.date` | Seed / Planned |

### P9 — Interview, oral history and personal correspondence

> **L:** Doe, Jane. Interview by the author. Calgary, Alberta, 2 February 2024. Recording privately held by the author.
>
> **F:** Jane Doe (Calgary, Alberta), interview by the author, 2 February 2024; recording privately held by the author.
>
> **S:** Doe, interview, 2 February 2024.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Interviewee or writer | L F S | `M.interviewee` | Seed (horizon key `interviewee`) |
| Interviewer or recipient | L F | `M.interviewer` | Seed |
| Place | L F | `M.jurisdiction` (place of interview) | Seed |
| Date | L F S | `S.date` | Planned |
| Where the recording or message is held | L F | `M.holder` | Seed |
| Timestamp of the cited passage | F | `C.label`; a time selector later ([`audio-video-sources.md`](../../ideas/audio-video-sources.md)) | Partial |

### P10 — Book and other standalone publication

Books, compiled and edited works, chapters, multi-volume works, published abstracts and transcriptions, CDs, maps, films.

> **L:** Smith, Alice. *The Smiths of Ottawa County*. Topeka: Prairie Press, 1965.
>
> **F:** Alice Smith, *The Smiths of Ottawa County* (Topeka: Prairie Press, 1965), 42.
>
> **S:** Smith, *Smiths of Ottawa County*, 42.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Authors, editors, compilers, translators | L F S | `M.author` as written, `M.author_sort` for the source list; `M.editor`, `M.compiler`, `M.translator` | Seed (`author` is a horizon key) |
| Title (italic), subtitle | L F S | `S.title` | Have / Seed |
| Short title (short note) | S | Derived from `S.title`; `M.short_title` overrides | Derived / Seed |
| Edition, volume, series | L F | `M.edition`, `M.volume`, `M.series` | Seed |
| Place of publication, publisher | L F | `M.publisher`, `M.publication_place` | Seed |
| Year | L F | `S.date` | Planned |
| Chapter author and title (essay in an edited book) | L F | `M.author` + `S.title`, with the book as `M.published_in` | Seed |
| Page | F S | `C.page_label` | Have, composer gap |
| Medium (CD-ROM, DVD, map scale) | L F | `M.format` | Seed |

### P11 — Periodical article and broadcast

Journal and magazine articles, newsletters, radio, TV and podcasts.

> **L:** Jones, Robert. "The Robins Family of York." *Yorkshire Genealogist* 12 (June 1990): 45–60.
>
> **F:** Robert Jones, "The Robins Family of York," *Yorkshire Genealogist* 12 (June 1990): 45–60, specifically 47.
>
> **S:** Jones, "Robins Family of York," 47.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Author | L F S | `M.author` (+ `M.author_sort`) | Seed |
| Article title (quoted) | L F S | `S.title` | Have |
| Periodical title (italic) | L F | `M.published_in` | Seed |
| Volume, issue | L F | `M.volume`, `M.issue` | Seed |
| Issue date | L F | `S.date` | Planned |
| Article's page range | L F | `M.page_range` | Seed |
| Page cited | F S | `C.page_label` | Have, composer gap |

### P12 — Newspaper item

> **L:** *Minneapolis (Kansas) Messenger*. 1931.
>
> **F:** "Death of Ada Robins," *Minneapolis (Kansas) Messenger*, 3 March 1931, p. 2, col. 4.
>
> **S:** "Death of Ada Robins," *Minneapolis Messenger*, 3 March 1931.

Under Spike 12's rule the Source is the item, and Copy Source carries the paper's details from a sibling ([`source-structure-later.md`](source-structure-later.md) "Nested Sources").

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Headline, or a description of an untitled item | F S | `S.title` | Have / Seed |
| Newspaper title, with its place in parentheses | L F S | `M.published_in` + `M.jurisdiction` | Seed |
| Issue date | L F S | `S.date` | Planned |
| Edition (morning, evening) | F | `M.edition` | Seed |
| Page, column | F | `C.page_label`, `M.column` | Partial |

### P13 — Website, web page and online database

Web pages, blog posts, online indexes and databases, online trees, memorial sites, message boards. The Source *is* the online publication.

> **L:** "Kansas, Births, 1880–1910." Database. FamilySearch. https://www.familysearch.org : 2025.
>
> **F:** "Kansas, Births, 1880–1910," database, FamilySearch (https://www.familysearch.org : accessed 12 March 2025), entry for Ada Robins, 12 February 1902; citing Ottawa County birth register, vol. 2, p. 41.
>
> **S:** "Kansas, Births, 1880–1910," entry for Ada Robins, 1902.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Creator or owner of the page or tree | L F | `M.author` (+ `M.author_sort`) | Seed |
| Page or database title (quoted) | L F S | `S.title` | Have |
| Kind of item ("database", "database with images", "blog") | L F | `M.format` | Seed |
| Website title (italic) | L F | `M.site_name` | Seed (horizon key) |
| URL | L F | `M.url` | Seed (horizon key) |
| Date accessed | L F | `M.accessed_date` | Seed (horizon key) |
| Entry for whom | F S | `G.person`, `G.date` | Graph |
| Memorial, record or profile id | F | `M.record_id` | Seed |
| Citing (the original the database rests on) | F | `M.citing` | Seed |

### P14 — Online image layer (layer 2)

Not a standalone pattern: the second layer appended to P1–P12 when the researcher saw the record as an image on a website.

> "…; digital image, Ancestry (https://www.ancestry.com : accessed 12 March 2025), image 37 of 412; citing NARA microfilm publication T9, roll 393."

| Element | Source | Status |
| --- | --- | --- |
| "digital image" | `M.format` | Seed |
| Provider or website | `M.provider` | Seed |
| Database or collection title (when cited) | `M.database_title` | Seed |
| URL | `M.url` | Seed |
| Accessed date | `M.accessed_date` | Seed |
| Image number in the online set | `C.page` (page in the file), when the Artifact is the whole image set | Have |
| Citing | `M.citing`, or the film reference in P15 | Seed |

### P15 — Microfilm layer (layer 2)

> "…; FHL microfilm 1,234,567, item 3." or "…; NARA microfilm publication T9, roll 393."

| Element | Source | Status |
| --- | --- | --- |
| "microfilm" | `M.format` | Seed |
| Who filmed or publishes the film (NARA, FHL, a state archive) | `M.provider` | Seed |
| Publication number, roll, film number, item | `M.film_publication`, `M.roll`, `M.film_number`, `M.item_number` | Seed (`roll` is a horizon key) |
| Frame or image | `C.page` | Have |

### P16 — Legal publication

Statutes, session laws, codes, legislative documents, published court reports. These follow legal citation conventions, compact and number-led.

> **F (statute):** Kansas, *Session Laws*, 1885, chap. 12, sec. 3, p. 45.
>
> **F (case):** *Smith v. Jones*, 12 Kan. 345 (1874).

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Jurisdiction | L F S | `M.jurisdiction` | Seed |
| Compilation or reporter title, abbreviated | L F S | `M.published_in` | Seed |
| Session or year | L F | `S.date` | Planned |
| Chapter, section, volume, page | F S | `M.chapter`, `M.section`, `M.volume`, `C.page_label` | Seed / Have |
| Case name | F S | `S.title` | Have |

### P17 — Fallback

Any Source whose type maps to no pattern, or whose pattern is missing a required element, gets today's line: Source title, Artifact label, Citation label, page ([`citation-formats.md`](citation-formats.md) "What ships in S13-11"). The citation preview names the missing elements so the researcher can fill them.

---

## 6. Missing fields and how to add them

### 6.1 The approach: metadata plus citation formats as data

**Decision (user, 2026-11-09).** EE's models are conventions for one goal: give enough to identify the record, find it again and judge it. So Provenencia does not code EE's patterns. It stores them as **citation formats**, editable text on each Source type, and seeds the EE formats for the product's types. A researcher's own types get formats the same way.

The pieces:

- **Metadata fields** hold the values, on the Source (the record), the Artifact (the copy used) and the Evidence graph (the item cited) (§ 6.6). Values are free text and print exactly as typed.
- **Source types are templates twice over:** their suggestions say which fields to fill at each level, and their citation formats say how those fields read in a citation (§ 6.3).
- **The seed** supplies the EE types, fields, suggestions and formats for a new project (§ 6.8). Nothing about EE is in code.
- **The engine** is one small interpreter: it parses a format, fills each placeholder from the Citation's graph, Artifact and Source, drops empty optional groups, and renders plain and styled text.

This also settles the metadata rethink ([`2026-10-25-rethink-source-metadata-and-type-suggestions.md`](2026-10-25-rethink-source-metadata-and-type-suggestions.md)): metadata stays, and becomes what citations are built from.

### 6.2 The full list

| What | Change | Kind |
| --- | --- | --- |
| Source date | `sources.date_value_id` ([`source-provenance-date.md`](source-provenance-date.md)) | **Schema** (planned) |
| Copy and item values | `artifact_metadata`, `evidence_graph_metadata` and their layout tables, reusing the field vocabulary (§ 6.6) | **Schema** |
| Suggestions per level | `source_type_metadata_fields.level` (`source`, `artifact`, `graph`) | **Schema** |
| Citation formats | `source_types.citation_full`, `citation_short`, `citation_list` (text, optional) (§ 6.3) | **Schema** |
| Format engine | `core/citationformat`: parse, resolve, render; `CitationLineDisplay` stays as the fallback for a type with no format (deployment plan, "Today's line stays the fallback") ([`citation-formats.md`](citation-formats.md) step 4) | Engine |
| Format editor | On the Source type: the three formats, an insert-field picker grouped by level, a live preview against a real Source, warnings for unknown fields; **Copy from** another type in the new type dialog | UI |
| Item description | The `{citation.item}` placeholder: the Citation label, else read from the graph (§ 6.5) | Engine |
| Printed page | The composer writes `page_label` on the locator's `page` selector; the field already exists | Engine and UI |
| Seed | Types, fields, suggestions and EE formats for every pattern (§ 6.8), installed at create (§ 6.4) | Seed |
| Party roles | `role` terms the item reads: grantor, grantee, writer, recipient, testator, head of household ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 3.5) | Seed |

So the schema changes are the Source date (already planned), four tables for Artifact and graph values and layout, a level on the type → field join, and three format columns on `source_types`. If the privacy toggle and moving an Artifact are pulled in, they add `sources.is_public` and possibly `ON UPDATE CASCADE` on the Subjects' composite key ([`move-artifact-between-sources.md`](move-artifact-between-sources.md)).

### 6.3 Citation formats

#### The forms

Each Source type has up to three formats, one per EE form (§ 2): **full** (first reference note), **short** (subsequent note) and **list** (source list entry). A type with no full format falls back to today's citation line (§ 5 P17).

#### The syntax

A small syntax of our own, modelled on the convention genealogy programs use for editable source templates (fields in brackets, optional text that drops out when its field is empty), with Markdown for italics:

| Syntax | Meaning |
| --- | --- |
| `{source.meta.key}`, `{artifact.meta.key}`, `{graph.meta.key}` | The value of the metadata field with that key **on that level**, as typed: the Source, the Citation's Artifact, or the Citation's Evidence graph |
| `{meta.key}` | Shorthand: the most specific value, the graph's, else the Artifact's, else the Source's. Seeded formats use it so they work however finely the researcher files (§ 6.9) |
| `{source.title}` and the other built-ins | Values that are not metadata (below) |
| `{a\|b\|…}` | The first that has a value (`{meta.short_title\|source.title}`, `{meta.author_sort\|meta.author}`) |
| `[ … ]` | Optional group: removed entirely, its text and punctuation included, if any placeholder inside it is empty. Groups nest |
| `*…*` | Italic |
| `\{` `\[` `\*` `\\` | A literal character |

Everything else is literal text, including quotation marks and parentheses. The level prefix is what lets a format pick a value from a particular level (user, 2026-11-09): `{source.meta.url}` for the site a Source is from, `{artifact.meta.url}` for the image the Citation was drawn on. Metadata always sits under `meta.`, so a built-in and a metadata key can never collide, whatever the researcher names their fields (user, 2026-11-09).

**Empty placeholders outside a group** print nothing, and the preview names them as missing. So a format says which fields it expects (outside groups) and which are optional (inside groups), and the "missing fields" warning in the citation preview comes from the format itself.

**No transforms on metadata (user, 2026-11-09).** Metadata is free text, so the engine never reorders, inverts or abbreviates it. A researcher who wants "Smith, Alice" in the source list fills `author_sort`; one who wants "Co." in short notes types it that way, or leaves the full form. Each field's description says to enter the value as it should appear.

#### Built-in placeholders

Values that are not metadata, named directly under their level. Metadata keys are under `meta.`, so no name is reserved.

| Placeholder | Value |
| --- | --- |
| `{source.title}` | The Source title |
| `{source.type}` | The Source type label |
| `{source.date}`, `{source.date:year}`, `{source.date:full}` | The Source date. It is a structured DateValue, so formatting it is safe: the year only, or day-month-year ("3 March 1930") at whatever precision the date has. `{source.date}` is `{source.date:full}` |
| `{artifact.label}` | The Artifact label |
| `{graph.name}` | The Evidence graph name |
| `{citation.page}` | The Citation's printed page (`page_label`) |
| `{citation.image}` | The Citation's page in the file (`artifact_page`), for the online-image layer ("image 37") |
| `{citation.item}` | The item: the Citation label, else read from the graph (§ 6.5) |

The Citation level has only built-ins: Citations carry no metadata (§ 6.9).

#### Resolving a placeholder

A level-prefixed placeholder reads that level only. `{meta.key}` reads the Citation's Evidence graph, else its Artifact, else the Source (§ 6.6). The **list** form describes the Source as a whole, so it reads Source values only: `{meta.key}` there means `{source.meta.key}`, and an `artifact.`, `graph.` or `citation.` placeholder is an error the editor reports.

The editor's **Insert field** picker groups fields by level (Source, Artifact, Evidence graph, Citation) and inserts the prefixed form (`{graph.meta.line_number}`); the researcher can shorten it to `{meta.line_number}` to read the most specific level.

#### Storage

- Three nullable text columns on `source_types`: `citation_full`, `citation_short`, `citation_list`. Editing them is a Source type edit, audited as one today.
- **Field references by id.** The editor shows `{graph.meta.enumeration_district}`, but the stored text names the field by id, so renaming a key cannot break a format, and a key held by both a product field and a researcher's field is never ambiguous. The editor resolves a typed key when the format is saved and asks if two fields share it.
- **Delete impact.** A field named in any type's format is in use: deleting it is refused, with the types that use it listed ([`add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md)).
- Copy Source copies nothing new here: formats belong to the type, and the copy chooses its type.

#### The engine

`core/citationformat` in Go: `Parse` builds a tree of literal runs, placeholders, groups and italic spans; `Render` fills placeholders through a resolver and returns styled runs, from which plain text and rich text (italic kept) are produced. Values are inserted as literal text after parsing, so a value containing `*` or `[` never changes the formatting. The Mac app renders through FFI (one RPC for a Citation's citation in a form, one for previewing an unsaved format against a chosen Source), so there is one implementation. The parser and renderer are small and tested exhaustively with table tests.

#### The editor

On the Source type's page: the three formats as text fields, an **Insert field** picker (metadata fields and built-ins), a **live preview** against a real Source of that type the researcher picks, and warnings for unknown fields and unbalanced brackets. Researchers should not have to remember the syntax to use it.

The **new type** dialog has **Copy from**: pick an existing type, and the new one starts with its three formats and its suggestions at every level, ready to edit. Without it, a new type starts blank (user, 2026-11-09).

### 6.4 Getting the seed into existing projects

**Descoped (user, 2026-11-09).** The seed installs at create only, as it does today ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 1). Provenencia is not in production, and the only existing projects are a few dogfood catalogs. When the seed PR lands, the product owner gets a one-off command to add the new types, fields and suggestions to their own project. It is not product code: no install on open, no seed-version row. `Install` already writes with upserts (`ON CONFLICT (key, origin)` for types and fields, and for suggestions), so the one-off command can run it against an existing catalog: it adds what is missing and resets the product rows' labels, suggestion order and formats to the seed, and restores product rows that were deleted. [`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md) keeps `Install` out of FFI, Swift and project open, so the run is a developer command or test harness, named in S15-20's plan.

Revisit before release, when existing projects belong to other researchers.

### 6.5 Reading the item from the graph

This is where "build the graph, get the citation" pays off. The `{citation.item}` placeholder (§ 6.3) is built for each Citation from the Subjects whose Observations cite it:

1. If the Citation has a `label`, use it (the researcher's override).
2. Else, find the event Subject with Observations on this Citation: its `event_type` gives "baptism of", its `date` gives the item date, and its principal participant's recorded `name` gives the person.
3. Else, find a person Subject: "entry for *name*" (P13), or "household of *name*" when the role is head of household (P5).
4. Else, it is empty, and a group around it drops.

Names and dates are the ones the Source records (Observations), never the canonical person's kept name: a citation describes the record, not the conclusion.

### 6.6 Artifact and graph values: where new schema earns its keep

Keeping everything on the Source fails in two places, and both come from the app's own model, not from EE:

- **The copy layer belongs to the Artifact.** A Source can hold several representations of one record as separate Artifacts ([`source-layer-data-model.md`](../../source-layer-data-model.md) § 7): a microfilm printout and an Ancestry download of the same register page. EE cites the copy that was used, and a Citation already points at one Artifact. With the provider, URL and accessed date on the Source, one of those two Artifacts is always cited with the wrong copy layer.
- **Item numbers belong below the Source.** One census page is one Artifact, and a researcher often cites several households on it, each with its own dwelling, family and line numbers. The same goes for entries on a register page and items on a newspaper page. These cannot be one value on the Source. They belong to the Evidence graph, one per household or entry (§ 6.9).

**Decision (user, 2026-11-09): the Source's metadata structure, repeated on the Artifact and the Evidence graph.** Values, layout (order and dismiss) and type templating all work the same at the three levels.

```sql
-- Values: one per owner and field, as on the Source.
CREATE TABLE artifact_metadata (
    id          BLOB PRIMARY KEY,
    artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id),
    value_text  TEXT,
    UNIQUE (artifact_id, field_id)
) STRICT;

CREATE TABLE evidence_graph_metadata (
    id          BLOB PRIMARY KEY,
    graph_id    BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
    field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id),
    value_text  TEXT,
    UNIQUE (graph_id, field_id)
) STRICT;

-- Layout: order and dismissed suggestions, as source_metadata_layout.
CREATE TABLE artifact_metadata_layout (
    artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
    sort_order  INTEGER NOT NULL,
    dismissed   INTEGER NOT NULL DEFAULT 0 CHECK (dismissed IN (0, 1)),
    PRIMARY KEY (artifact_id, field_id)
) STRICT;

CREATE TABLE evidence_graph_metadata_layout (
    graph_id    BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
    field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
    sort_order  INTEGER NOT NULL,
    dismissed   INTEGER NOT NULL DEFAULT 0 CHECK (dismissed IN (0, 1)),
    PRIMARY KEY (graph_id, field_id)
) STRICT;

-- Templating: a type suggests fields per level.
-- source_type_metadata_fields gains a level; the table is rebuilt so the
-- key includes it (a field may be suggested at more than one level).
--   level TEXT NOT NULL DEFAULT 'source' CHECK (level IN ('source', 'artifact', 'graph'))
--   PRIMARY KEY (source_type_id, field_id, level)
-- Existing rows become level = 'source'.
```

- **Same fields, three levels.** All three levels take any field from the one vocabulary, seeded and researcher-made alike. No new columns on `artifacts` or `citations`.
- **The most specific value wins.** For each key, the format engine reads the value on the Citation's graph, else on its Artifact, else on the Source. A researcher who files one Source per household never needs the graph level; one who files the whole page as a Source fills line numbers per graph, one graph per household (§ 6.9).
- **Templating follows the Source's type.** An Artifact's and a graph's suggestions are the `artifact` and `graph` rows of their Source's type, so a census Source suggests provider and URL on each Artifact and line, dwelling and family on each graph. The Source types admin edits the three lists the way it edits one today.
- **Layout works as on the Source.** `ListWorkspace`'s merge rules (dismiss hides an empty suggestion, saved order first, then suggestions, then extras) apply unchanged per Artifact and per graph ([`source-layer-data-model.md`](../../source-layer-data-model.md) § 5.4). The Artifact form and the graph's edit view each get a metadata board built from the Source page's. The Citation composer gets none.
- **Readers of the join** filter by level: `sourcetypes`, `sourcevocab` (suggestion ops and Install), the `effects` registry and the catalog model.
- **Name.** `source_metadata_fields` now serves all three levels. Renaming it to `metadata_fields` would be accurate but touches every reader; keeping the name and noting the wider use is the cheaper choice.

**Knock-on work** (each with its skill): register the four tables in the catalog model (`add-catalog-model`); audit them under their Source's scope so the Source's history shows them, with dismiss and reorder audited as on the Source (`add-audit-scope`); add them to the field delete-impact probe, so a field in use on an Artifact or graph cannot be deleted (`add-catalog-delete`); release them with their Artifact or graph (a graph delete is already audited per facet, [`source-layer-data-model.md`](../../source-layer-data-model.md) § 7.1). Copy Source copies a chosen Artifact's values and layout with it, and copies no graph values, since it copies no graphs. Moving an Artifact to another Source carries its values and its graphs' values for free, because they hang off `artifact_id` and `graph_id`.

**Still given up, with the workaround:**

| Given up | Workaround |
| --- | --- |
| A shared Repository record | `repository` text, carried between siblings by Copy Source |
| Several authors with roles in order | `author` as written plus `author_sort`; `editor`, `compiler`, `translator` as separate fields |
| Reordering, inverting or abbreviating values | Enter values as they should appear; `author_sort` and `short_title` give a second form where a format wants one (§ 6.3) |

### 6.7 Suggested order

Following [`citation-formats.md`](citation-formats.md) "How it could arrive":

1. **Source date** and the Source page design that shows it.
2. **Artifact and graph metadata:** the value and layout tables, the level on suggestions, a metadata board on the Artifact form and on the graph, and the Artifact's fields in the import flow.
3. **Printed page labels** in the composer.
4. **Format engine:** `core/citationformat` (parse, resolve, render), the three format columns, the FFI calls, and the provenance sheet's citation line reading the full format, with today's line as the fallback.
5. **Format editor** on the Source type, with the live preview.
6. **Seed:** every type, field, suggestion and format in § 6.8, installed at create; a one-off command updates the dogfood projects (§ 6.4). It is data, so it can land in batches if one PR is too large.
7. **Citation preview** on the Source page, naming missing fields.
8. **Grouped display:** the provenance sheet and the Person page's Sources section show one full note per rendered citation (§ 6.9).
9. **Copy citation**, full or short.

### 6.8 The seed

**Decision (user, 2026-11-09): seed everything the patterns in § 5 use.** Researchers can delete, rename or edit any type, field, suggestion or format, so the seed does not need trimming. It installs at create (§ 6.4).

**It is the product default, not a plugin (user, 2026-11-09).** Every new project gets it, with no option to leave it out. Packaging styles as plugins is parked in [`source-type-packages.md`](../../ideas/source-type-packages.md).

**Existing infrastructure.** The product seed already exists: `core/database/sourcevocab/registry.go` lists the seeded types, fields and suggestions; `sourcevocab.Install` writes them with `origin = 'provenencia'`; `core/onboarding/ready.go` calls it when a catalog is created ([`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md)). Spike 15 extends it:

- `seedField` entries for the 62 fields, each with a description that says to enter the value as it should appear.
- `seedType` entries for the 26 types, each with an icon from the existing set in `core/database/sourcetypes/icons.go` (`type_census`, `type_certificate`, `type_book`, `type_newspaper`, `type_website`, `type_dna`, `type_gedcom`, `type_grave`, `type_photograph`, `type_oral_history`, `type_email`, `type_passport`, `type_folder_archive`, `type_document`, `type_scroll`), and its three formats written with keys.
- `seedSuggestion` gains a level (`source`, `artifact`, `graph`).
- `Install` writes the formats after the fields, turning each `{…meta.key}` into the field's id (§ 6.3, Storage). Written for Provenencia from the conventions in § 2, not copied from *Evidence Explained* (working rule, [`README.md`](README.md#working-rules)); check it against the edition on hand before it is merged.

#### Source types

26 types. Every horizon type in [`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 2.1 is among them. P14 (online image) and P15 (microfilm) are layers inside other types' formats, not types; P17 is the fallback for a type with no format.

| Key | Label | Pattern (§ 5) | Covers |
| --- | --- | --- | --- |
| `census` | Census | P5 | Every census and schedule in § 3 |
| `birth_certificate` | Birth certificate | P4 | Already seeded |
| `marriage_certificate` | Marriage certificate | P4 | |
| `death_certificate` | Death certificate | P4 | |
| `church_register` | Church register | P6 | Parish registers, church books, bishop's transcripts |
| `record_book` | Record book | P2 | Deed, will, court, marriage and tax books kept by a public office |
| `passenger_list` | Passenger list | P2 (national) | Manifests and passenger lists |
| `case_file` | Case file | P3 | Court and probate files, original wills, naturalization, land entry files |
| `manuscript` | Archived manuscript | P1 | Letters, diaries, business and institutional records held by an archive |
| `government_file` | Government file | P1 (national) | Military service, pension, draft, passport and agency files |
| `cemetery_marker` | Grave marker | P7 | Headstones and monuments |
| `family_bible` | Family Bible | P8 | Bible records |
| `photograph` | Photograph | P8 | Photographs and portraits in private hands |
| `private_papers` | Private papers | P8 | Letters, diaries, research reports and compiled family files in private hands |
| `dna_match_report` | DNA match report | P8 | Test results and match lists |
| `gedcom_file` | GEDCOM file | P8 | Exported family files |
| `interview` | Interview | P9 | Recorded or published interviews |
| `oral_testimony` | Oral testimony | P9 | Informal family testimony |
| `correspondence` | Correspondence | P9 | Letters and email to the researcher |
| `book` | Book | P10 | Books, chapters, maps, CDs, films |
| `journal_article` | Journal article | P11 | Journal, magazine and newsletter articles |
| `broadcast` | Broadcast | P11 | Radio, television, podcasts |
| `newspaper_item` | Newspaper item | P12 | Articles, notices, obituaries |
| `website` | Website or online database | P13 | Web pages, databases, online trees, memorial sites |
| `statute` | Statute | P16 | Statutes, session laws, codes |
| `court_report` | Court report | P16 | Published cases |

#### Fields

62 fields; one (`document_number`) is already seeded. **Usual level** is where a type suggests the field: Source unless listed otherwise. Any field can be filled at any level.

| Group | Keys | Usual level |
| --- | --- | --- |
| Who made it | `creator`, `author`, `author_sort`, `editor`, `compiler`, `translator`, `photographer`, `studio`, `interviewer`, `interviewee`, `recipient`, `testing_company`, `software` | Source |
| Where | `jurisdiction` | Source |
| Publication | `published_in`, `publisher`, `publication_place`, `edition`, `series`, `volume`, `issue`, `page_range`, `site_name`, `short_title`, `coverage` | Source |
| Archive | `repository`, `collection`, `record_group`, `call_number`, `box`, `folder` | Source |
| Record numbers | `document_number` (certificate or document number), `book`, `case_number`, `enumeration_district`, `schedule`, `piece`, `chapter`, `section`, `lot`, `kit_number`, `match_identifier` | Source |
| Holding | `holder` (who holds a private original, and where), `medium` | Source |
| The copy | `format`, `provider`, `database_title`, `url`, `accessed_date`, `citing`, `film_publication`, `roll`, `film_number`, `item_number` | Artifact |
| The item | `entry_number`, `line_number`, `dwelling_number`, `family_number`, `schedule_number`, `folio`, `column`, `record_id` | Evidence graph |

Not seeded, from the horizon list: the date keys (`publication_date`, `issue_date`, `record_date`, `census_date`, `taken_date`, `interview_date`, `export_date`, `report_date`), which the Source date replaces; `certificate_number` and `registration_number`, which `document_number` covers; and `isbn`, `language`, `form_number`, `gedcom_version`, `civil_division`, `sheet` and `page`, which no format uses. `birth_certificate`'s existing `record_date` and `issue_date` stay as fields so existing values are kept, but are no longer suggested.

#### Suggestions

One rule, so the suggestions and the formats cannot drift: **each type suggests every metadata field its full format uses, at the field's usual level, in the order the format uses them.** Two exceptions:

- `website` suggests `url` and `accessed_date` on the Source, since the website is the Source.
- Record types (census, certificates, church register, record book, passenger list, case file, manuscript, government file) suggest the copy fields on the Artifact even where the format reads only some of them.

#### Formats

Lines are wrapped here for reading; a format is one line. Types that share a pattern differ only where shown.

**Census**

```text
full:  {source.date:year} [{meta.creator} ]census, {meta.jurisdiction}[, {meta.schedule}]
       [, ED {meta.enumeration_district}][, piece {meta.piece}][, fol. {meta.folio}][, p. {citation.page}]
       [, schedule {meta.schedule_number}][, dwelling {meta.dwelling_number}][, family {meta.family_number}]
       [, line {meta.line_number}][, {citation.item}][; {meta.format}, *{meta.provider}*]
       [ ({meta.url}[ : accessed {meta.accessed_date}])][; citing {meta.citing}].
short: {source.date:year} [{meta.creator} ]census, {meta.jurisdiction}[, ED {meta.enumeration_district}]
       [, fol. {meta.folio}][, p. {citation.page}][, dwell. {meta.dwelling_number}][, fam. {meta.family_number}]
       [, {citation.item}].
list:  {meta.jurisdiction}. {source.date:year} [{meta.creator} ]census[, {meta.schedule}].
```

> 1880 U.S. census, Ottawa County, Kansas, population schedule, ED 236, p. 14, dwelling 112, family 118, Ada Robins; digital image, *Ancestry* (https://www.ancestry.com : accessed 12 March 2025); citing NARA microfilm publication T9, roll 393.

**Birth, marriage and death certificates** (each with its own word)

```text
full:  [{meta.creator}, ]birth certificate[ no. {meta.document_number}][ ({source.date:full})][, {citation.item}]
       [, vol. {meta.volume}][, p. {citation.page}][, entry {meta.entry_number}][; {meta.format}, *{meta.provider}*]
       [ ({meta.url}[ : accessed {meta.accessed_date}])][; {meta.repository}].
short: [{meta.creator}, ]birth certificate[ no. {meta.document_number}][, {citation.item}].
list:  {meta.jurisdiction}.[ {meta.creator}.] Birth certificates.[ {meta.repository}.]
```

**Church register**

```text
full:  {meta.creator}[ ({meta.jurisdiction})], *{source.title}*[, vol. {meta.volume}][, p. {citation.page}]
       [, fol. {meta.folio}][, no. {meta.entry_number}][, {citation.item}][; {meta.format}, *{meta.provider}*]
       [ ({meta.url}[ : accessed {meta.accessed_date}])][; film {meta.film_number}[, item {meta.item_number}]]
       [; {meta.repository}].
short: {meta.creator}, *{meta.short_title|source.title}*[, vol. {meta.volume}][, p. {citation.page}]
       [, no. {meta.entry_number}].
list:  {meta.creator}[ ({meta.jurisdiction})]. *{source.title}*[, {meta.coverage}].[ {meta.repository}.]
```

**Record book**

```text
full:  {meta.jurisdiction}, [{meta.creator}, ]{meta.book|source.title}[: {citation.page}][, entry {meta.entry_number}]
       [, {citation.item}][; {meta.format}, *{meta.provider}*][ ({meta.url}[ : accessed {meta.accessed_date}])]
       [; film {meta.film_number}][; {meta.repository}].
short: {meta.jurisdiction}, {meta.book|source.title}[: {citation.page}].
list:  {meta.jurisdiction}.[ {meta.creator}.] {source.title}[, {meta.coverage}].[ {meta.repository}.]
```

**Passenger list**

```text
full:  {source.title}[, p. {citation.page}][, line {meta.line_number}][, {citation.item}]
       [; {meta.format}, *{meta.provider}*][ ({meta.url}[ : accessed {meta.accessed_date}])]
       [; citing {meta.citing}][; {meta.film_publication}[, roll {meta.roll}]].
short: {meta.short_title|source.title}[, p. {citation.page}][, line {meta.line_number}].
list:  [{meta.creator}. ]{meta.collection|source.title}.[ {meta.repository}.]
```

**Case file**

```text
full:  {meta.jurisdiction}, [{meta.creator}, ]{source.title}[ {meta.case_number}][, {citation.item}]
       [; {meta.format}, *{meta.provider}*][ ({meta.url}[ : accessed {meta.accessed_date}])][; {meta.repository}].
short: {meta.jurisdiction}, {meta.short_title|source.title}[ {meta.case_number}][, {citation.item}].
list:  {meta.jurisdiction}.[ {meta.creator}.] {meta.collection|source.title}.[ {meta.repository}.]
```

**Archived manuscript**

```text
full:  [{meta.creator}, ]{source.title}[ ({source.date:full})][, {citation.item}][; {meta.collection}]
       [, {meta.call_number}][, box {meta.box}][, folder {meta.folder}][; {meta.format}, *{meta.provider}*]
       [ ({meta.url}[ : accessed {meta.accessed_date}])]; {meta.repository}.
short: [{meta.creator}, ]{meta.short_title|source.title}[, {meta.collection}][, box {meta.box}][, folder {meta.folder}].
list:  {meta.collection|source.title}[, {meta.coverage}].[ {meta.call_number}.] {meta.repository}.
```

**Government file**

```text
full:  {source.title}[ ({source.date:full})][, {citation.item}][; {meta.collection}][; Record Group {meta.record_group}]
       [; {meta.format}, *{meta.provider}*][ ({meta.url}[ : accessed {meta.accessed_date}])]
       [; {meta.film_publication}[, roll {meta.roll}]]; {meta.repository}.
short: {meta.short_title|source.title}[, {citation.item}].
list:  {meta.repository}.[ Record Group {meta.record_group}.] {meta.collection|source.title}.
```

**Grave marker** (`creator` is the cemetery)

```text
full:  {meta.creator}[ ({meta.jurisdiction})], {citation.item|source.title}[, section {meta.section}][, lot {meta.lot}]
       [; photographed by {meta.photographer}[, {meta.accessed_date}]].
short: {meta.creator}[ ({meta.jurisdiction})], {citation.item|source.title}.
list:  {meta.jurisdiction}. {meta.creator}. Grave markers.
```

**Family Bible**

```text
full:  {source.title}[, *{meta.published_in}*][ ([{meta.publication_place}: ]{meta.publisher}[, {source.date:year}])]
       [, {citation.item}][; {meta.holder}].
short: {meta.short_title|source.title}[, {citation.item}].
list:  {source.title}[, {meta.coverage}].[ *{meta.published_in}*.][ {meta.holder}.]
```

**Photograph**

```text
full:  {source.title}[, {meta.medium}][, {source.date:full}][, {meta.photographer|meta.studio}][; {meta.holder}].
short: {meta.short_title|source.title}[, {source.date:year}].
list:  {source.title}.[ {meta.medium}.][ {meta.holder}.]
```

**Private papers**

```text
full:  [{meta.creator}, ]{source.title}[, {source.date:full}][, {citation.item}][; {meta.holder}].
short: [{meta.creator}, ]{meta.short_title|source.title}[, {citation.item}].
list:  [{meta.creator}. ]{source.title}[, {meta.coverage}].[ {meta.holder}.]
```

**DNA match report**

```text
full:  {meta.testing_company}, {source.title}[, kit {meta.kit_number}][, match {meta.match_identifier}]
       [, {citation.item}][, report dated {source.date:full}][; {meta.holder}].
short: {meta.testing_company}, {meta.short_title|source.title}[, match {meta.match_identifier}].
list:  {meta.testing_company}. {source.title}.[ {meta.holder}.]
```

**GEDCOM file**

```text
full:  [{meta.creator}, ]{source.title}[, GEDCOM file exported from {meta.software}][, {source.date:full}]
       [, {citation.item}][; {meta.holder}].
short: [{meta.creator}, ]{meta.short_title|source.title}[, {citation.item}].
list:  [{meta.creator}. ]{source.title}.[ {meta.holder}.]
```

**Interview** (oral testimony: the same, with `interview by` inside an optional group)

```text
full:  {meta.interviewee}[ ({meta.jurisdiction})], interview by {meta.interviewer}, {source.date:full}
       [, {citation.item}][; {meta.holder}].
short: {meta.interviewee}, interview, {source.date:full}.
list:  {meta.interviewee}. Interview by {meta.interviewer}.[ {meta.jurisdiction},] {source.date:full}.[ {meta.holder}.]
```

**Correspondence**

```text
full:  {meta.creator}[ ({meta.jurisdiction})] to {meta.recipient}, {source.title}, {source.date:full}[; {meta.holder}].
short: {meta.creator} to {meta.recipient}, {source.date:full}.
list:  {meta.creator}. Correspondence to {meta.recipient}[, {meta.coverage}].[ {meta.holder}.]
```

**Book**

```text
full:  {meta.author}[, ed. {meta.editor}][, comp. {meta.compiler}][, trans. {meta.translator}], *{source.title}*
       [, {meta.edition}][, vol. {meta.volume}][, {meta.series}][ ([{meta.publication_place}: ]{meta.publisher}
       [, {source.date:year}])][, {citation.page}][; {meta.format}, *{meta.provider}*]
       [ ({meta.url}[ : accessed {meta.accessed_date}])].
short: {meta.author}, *{meta.short_title|source.title}*[, {citation.page}].
list:  {meta.author_sort|meta.author}. *{source.title}*.[ {meta.edition}.][ [{meta.publication_place}: ]{meta.publisher}
       [, {source.date:year}].]
```

**Journal article** (broadcast: `published_in` is the programme, `publisher` the network)

```text
full:  {meta.author}, "{source.title}," *{meta.published_in}*[ {meta.volume}][, no. {meta.issue}][ ({source.date:full})]
       [: {meta.page_range}][, specifically {citation.page}].
short: {meta.author}, "{meta.short_title|source.title}"[, {citation.page}].
list:  {meta.author_sort|meta.author}. "{source.title}." *{meta.published_in}*[ {meta.volume}][, no. {meta.issue}]
       [ ({source.date:full})][: {meta.page_range}].
```

**Newspaper item**

```text
full:  "{source.title}," *{meta.published_in}*[ ({meta.jurisdiction})], {source.date:full}[, p. {citation.page}]
       [, col. {meta.column}][; {meta.format}, *{meta.provider}*][ ({meta.url}[ : accessed {meta.accessed_date}])].
short: "{meta.short_title|source.title}," *{meta.published_in}*, {source.date:full}.
list:  *{meta.published_in}*[ ({meta.jurisdiction})].
```

**Website or online database**

```text
full:  [{meta.creator}, ]"{source.title}," [{meta.format}, ]*{meta.site_name}*
       [ ({meta.url}[ : accessed {meta.accessed_date}])][, {citation.item}][; citing {meta.citing}].
short: "{meta.short_title|source.title}"[, {citation.item}].
list:  [{meta.creator}. ]"{source.title}." *{meta.site_name}*.[ {meta.url}.]
```

**Statute**

```text
full:  {meta.jurisdiction}, *{meta.published_in}*[ {source.date:year}][, vol. {meta.volume}][, chap. {meta.chapter}]
       [, sec. {meta.section}][, p. {citation.page}].
short: {meta.jurisdiction}, *{meta.published_in}*[ {source.date:year}][, chap. {meta.chapter}].
list:  {meta.jurisdiction}. *{meta.published_in}*[, {source.date:year}].
```

**Court report**

```text
full:  *{source.title}*, [{meta.volume} ]{meta.published_in} {citation.page|meta.page_range}[ ({source.date:year})].
short: *{meta.short_title|source.title}*[, {citation.page}].
list:  *{source.title}*. [{meta.volume} ]{meta.published_in}[ ({source.date:year})].
```

### 6.9 One EE citation, many Provenencia Citations

A census row is one EE citation (the household, its dwelling, family and line), but a researcher may draw five Citations on it: the name, the age, the birthplace, the occupation, the relationship column. All five share the item and differ only in their locators.

**Decision (user, 2026-11-09): the item lives on the Evidence graph.** The Citation composer gets no new fields.

- **Why the graph.** An Evidence graph already groups the Subjects drawn from one part of a Source, and Spike 12 made several graphs per Source for exactly this: one per household, entry or article in a large record. Every Citation belongs to one graph and never moves ([`source-layer-data-model.md`](../../source-layer-data-model.md) § 7.1). So the item's numbers are entered once, on the graph, and every Citation drawn in it inherits them.
- **Why not the Artifact.** An Artifact is one file, and one file often holds several items (a census page with three households, a register page with eight entries), while one item can run onto a second page. Item fields on the Artifact would need one Artifact per household, each repeating the copy fields, and would change what an Artifact means. Graphs and Artifacts are already many-to-many, so a household over two pages is one graph linked to two Artifacts.
- **Why not the Citation.** It repeats the same values on every Citation of the row, adds input to the composer, which is already the heaviest step in building a graph, and needs tools to keep the copies in step.

**What follows:**

1. **Entering the item:** once, on the graph's metadata board, where its name and description are edited. A census Source's type suggests dwelling, family and line there.
2. **Showing the item once:** wherever Citations are listed for a reader (the provenance sheet, the Person page's Sources section, an export), Citations whose rendered full citation is the same are one EE citation; in practice, the same graph, Artifact and printed page. The full note shows once, with its Citations (the highlighted cells) beneath it. The same key gives the short note on second mention; a source list groups one level up, by Source. The key is computed when read; nothing new is stored.
3. **Editing the item:** once, on the graph.

**When one graph covers several items.** A researcher who keeps the Default graph for a page with two households of interest gets one set of item fields for both. The answer is the one Spike 12 intended: one graph per household. The Citation label stays the escape hatch ("line 14"), and the citation preview shows the result.

**The printed page stays on the Citation,** because it is the one part of the item that can differ inside a graph. It already exists on the locator (`page_label`). If that is still too much input, a later option is a page map per Artifact ("image 37 is page 23"), entered once, so every Citation inherits the printed page from the image it is drawn on (§ 7).

**Researchers who file finely skip the graph level.** If the Source is the household (one Source per household, each with an Artifact of the same page image, which the content-addressed File store already shares), the item values sit on the Source.

---

## 7. Open questions

- ~~Which patterns first~~: all of them, as seed data (§ 6.8).
- ~~Install for existing projects~~: descoped; create-time only, plus a one-off command for the dogfood projects (user, 2026-11-09, § 6.4).
- **Same key, two origins.** Whether to offer moving a researcher's own `author` (or similar) values into the product field, and when.
- ~~Generic template or a pattern choice for user types~~: any type has its own editable formats (user, 2026-11-09, § 6.3).
- ~~Starting formats for a new type~~: the researcher chooses. The new type dialog has **Copy from**, which copies another type's three formats and its suggestions (user, 2026-11-09).
- **Jurisdiction from the graph.** The graph often has the place chain (town → county → state through `part_of`). Offering "fill jurisdiction from this graph's place" as a one-time copy into the text would save typing without making the Source depend on interpretation.
- **Principal Subject.** When a Citation supports several people (a census household, a marriage), which one names the item: head of household, both spouses, the person the researcher opened the composer from?
- ~~Editable templates~~: yes, citation formats are data on the type (user, 2026-11-09, § 6.3).
- ~~Per-Source override~~: no; `{a|b}` fallbacks and optional groups are enough for now (user, 2026-11-09).
- ~~Layout at the Artifact and item levels~~: same as the Source, with templating per level (user, 2026-11-09).
- ~~Where the item lives~~: on the Evidence graph; the composer gets no new fields (user, 2026-11-09, § 6.9).
- ~~How much seed~~: everything the patterns use; researchers remove what they don't want (user, 2026-11-09, § 6.8).
- **Field vocabulary name.** Keep `source_metadata_fields` for all three levels, or rename it to `metadata_fields`.
- **Printed page map.** Whether a per-Artifact page map ("image 37 is page 23") should replace typing the printed page in the composer (§ 6.9).
- **One graph, several items.** Whether the app should nudge a researcher toward one graph per household when a Source's type suggests item fields at the graph level.
- ~~Syntax details~~: `{a|b}` is enough for now (user, 2026-11-09). The engine returns diagnostics with positions, and the editor shows them (map review K-10, [`deployment-plan.md`](deployment-plan.md)).
- ~~More forms~~: three (full, short, list) are enough for now (user, 2026-11-09).
- **Edition check.** Verify § 3 against the edition of *Evidence Explained* on hand, including how it treats DNA results and online trees.
