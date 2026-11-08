# Evidence Explained citation map

**Status:** Spike 15 design input, not yet cut into PRs ([`README.md`](README.md)). It lists the kinds of record *Evidence Explained* (EE) gives citation models for, groups them into the citation patterns Provenencia has to build, and maps each pattern's elements to data the catalog has or will need. The last section lists every missing field with a recommended way to add it. It extends [`citation-formats.md`](citation-formats.md), which sets out the convention and the six-step path; this doc is the inventory behind steps 1, 2 and 4 of that path.

**Goal (user, 2026-11-08):** a researcher builds an Evidence graph, promotes its Subjects, and gets a correct EE citation with no further typing. Every field below either already exists, can be read from the graph, or is one the researcher fills once on the Source, Artifact or Citation.

## Read this first

- **Edition.** The list follows the chapter layout of *Evidence Explained: Citing History Sources from Artifacts to Cyberspace* (Elizabeth Shown Mills), 3rd edition. It was written from knowledge of the book, not from a copy beside the keyboard. Check the chapter list and the record kinds against the edition on hand before the formatter's templates are written, and add any kind this list misses to its pattern.
- **Not a copy of the book.** EE has hundreds of "QuickCheck" models, many of them variants of the same shape. This doc does not reproduce them. It names the record kinds and paraphrases the shape each one shares. Every example citation here is invented for this doc (made-up people, records and numbers) and is written in EE style, not taken from the book.
- **Keep it that way.** Every later doc, template, test and UI string follows the spike's working rule on *Evidence Explained* ([`README.md`](README.md#working-rules)): our own templates, invented examples, our own names, no implied endorsement.
- **Patterns, not models.** The formatter needs one template per *pattern* (about 17), with small per-kind variations, not one per EE model. The catalog in § 2 maps each record kind to its pattern.

## Contents

1. [How an EE citation is put together](#1-how-an-ee-citation-is-put-together)
2. [The record kinds EE covers](#2-the-record-kinds-ee-covers)
3. [Where the data comes from: the element codes](#3-where-the-data-comes-from-the-element-codes)
4. [The patterns](#4-the-patterns)
5. [Missing fields and how to add them](#5-missing-fields-and-how-to-add-them)
6. [Open questions](#6-open-questions)

---

## 1. How an EE citation is put together

Summarised from [`citation-formats.md`](citation-formats.md) "The convention", with the rules the formatter has to know.

**Three forms per record.**

| Form | What it identifies | Grain |
| --- | --- | --- |
| **Source list entry** (bibliography) | The source as a whole: the register, the census of a county, the book, the database | Source |
| **First reference note** (full note) | The exact item used: entry, page, line, the person named | Source + Artifact + Citation |
| **Subsequent note** (short note) | The same item again, abbreviated | Source + Citation |

**Layers.** A full note is up to three layers separated by semicolons: (1) the record and the item in it; (2) how it was seen (the copy: "digital image, Ancestry (URL : accessed date)", or a microfilm); (3) what that copy rests on ("citing …", the repository and its reference). A record seen in the original has layers 1 and 3 only ("…; County Clerk's Office, Minneapolis").

**Order flips between forms.** A source list entry runs from large to small and starts with what it is filed under, often the jurisdiction ("Kansas. Ottawa County. *Deed Book C*."). A note runs small to large ("Ottawa County, Kansas, Deed Book C: 214"). So the formatter needs the jurisdiction as **ordered parts**, not one string.

**Title styling.** The title of a published work is italic; a formal title of an unpublished manuscript is in quotation marks; a generic description ("baptisms, 1813–1837", "letter") is plain. So the formatter must know whether a title is the record's own exact title or the researcher's description of it.

**Names of people.** Authors are written in natural order in notes ("Alice Smith") and inverted for the first author in the source list ("Smith, Alice"). People *in* the record are named as the record names them ("entry for Wm. Robins"), not as the researcher concluded.

**Dates.** EE writes dates day-month-year ("3 March 1830"), abbreviates months only in short notes, and uses several dates in one citation: the record's or publication's date, the item's date (the baptism), and the date accessed.

---

## 2. The record kinds EE covers

Chapters 1 and 2 of EE are the fundamentals of evidence analysis and of citation; they have no record kinds of their own. Chapters 3 to 14 each cover a family of records. The **Pattern** column names the template in § 4 that each kind uses.

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

## 3. Where the data comes from: the element codes

Every element in § 4 names its source with one of these codes. **Have** means stored and readable today. **Graph** means readable from the Evidence graph the researcher already built. **Planned** means designed in another Spike 15 doc. **New** means it has to be added; § 5 says how.

### Have

| Code | What | Where |
| --- | --- | --- |
| `S.title` | Source title (required headline) | `sources.title` |
| `S.description` | Source description | `sources.description` |
| `S.type` | Source type label and key | `sources.source_type_id` → `source_types` |
| `S.meta[key]` | Free-text metadata value under a researcher's or seed's key | `source_metadata.value_text`; seeded keys exist only for `birth_certificate` in new catalogs |
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

The graph supplies the "entry for …" or "baptism of …" part of a full note. It reads the Source's own Subjects, never the canonical entities, because EE names people as the record names them. A Citation that supports several Subjects needs a rule for the principal one (§ 6).

### Planned

| Code | What | Where |
| --- | --- | --- |
| `S.date` | The Source's date: publication, or the date the record was made | `sources` → `date_values`, [`source-provenance-date.md`](source-provenance-date.md) |

### New (detailed in § 5)

| Code | What | Level |
| --- | --- | --- |
| `S.title_kind` | Whether `S.title` is the record's exact title or a description of it | Source |
| `S.contributors` | Creators with roles: author, editor, compiler, translator, creating agency, photographer, interviewer | Source |
| `S.jurisdiction` | Ordered place parts, largest to smallest | Source |
| `S.repository` | The holding repository: name and location | Source → shared Repository record |
| `S.publication` | Place of publication, publisher, edition | Source |
| `S.container` | The containing work: journal, newspaper, website, book holding the chapter, database | Source |
| `S.span` | The years a register or collection covers ("1813–1837") | Source |
| `X.ref[kind]` | A reference part of a given kind (volume, book, folio, page, entry, line, ED, sheet, dwelling, family, piece, schedule, certificate number, case number, record group, series, collection, call number, box, folder, issue, column) | Source, Artifact or Citation, whichever the researcher filled (§ 5.2) |
| `A.format` | How the copy was seen: original, photocopy, certified copy, microfilm, digital image, database entry, transcript, abstract, recording | Artifact |
| `A.provider` | Website or provider (Ancestry, FamilySearch, a record office's site) | Artifact |
| `A.database` | Database or collection title at the provider | Artifact |
| `A.url` | URL of the image, record page or database | Artifact (Citation can override) |
| `A.accessed` | Date accessed | Artifact |
| `A.film` | Microfilm: publication number and roll, or film and item number | Artifact |
| `A.citing` | What the provider says its copy rests on, as text | Artifact |
| `A.holder` | Who holds a privately held original, and where | Artifact |
| `C.item` | Override for the item description when the graph does not give one | Citation (today's `C.label` can serve) |

---

## 4. The patterns

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
| Creator of the item (writer, agency, firm) | F S | `S.contributors` (creator), else `G.parties` (writer) | New |
| Recipient (letters) | F S | `G.parties` (recipient role) | Graph; needs a `recipient` role term |
| Item description ("letter", "pension file") | F S | `C.item`, else `S.type` label | Partial |
| Item date | F S | `G.date`, else `S.date` | Graph / Planned |
| Collection title | L F S | `X.ref[collection]` | New (`S.meta[collection]` horizon key) |
| Collection span | L F | `S.span` | New |
| Call or MS number | L F | `X.ref[call_number]` | New |
| Record group, series | L F | `X.ref[record_group]`, `X.ref[series]` | New |
| Box, folder, item | F S | `X.ref[box]`, `X.ref[folder]`, `X.ref[item]` | New |
| Repository and its location | L F | `S.repository` | New |

### P2 — Local record book

A register kept by a public office in book form: deed books, will books, court minutes, marriage registers, tax rolls, passenger manifests (national variant).

> **L:** Kansas. Ottawa County. Deed Records, 1866–1920. Register of Deeds, Minneapolis.
>
> **F:** Ottawa County, Kansas, Deed Book C: 214–15, John Smith to William Robins, 3 May 1879; Register of Deeds, Minneapolis.
>
> **S:** Ottawa Co., Kans., Deed Book C: 214.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Jurisdiction | L F S | `S.jurisdiction` | New |
| Office or court that kept it | L F | `S.contributors` (creating agency) | New |
| Record series and span ("Deed Records, 1866–1920") | L | `S.title` + `S.span` | Have / New |
| Book or volume | F S | `X.ref[book]` or `X.ref[volume]` | Partial: often typed into `A.label` today |
| Page (printed) | F S | `C.page_label` | Have, composer gap |
| Entry, line, folio | F S | `X.ref[entry]`, `X.ref[line]`, `X.ref[folio]` | New |
| Parties and nature of the record ("John Smith to William Robins") | F | `G.parties` with roles (grantor, grantee) | Graph; needs role terms |
| Item date | F | `G.date` | Graph |
| Repository (the office today) | L F | `S.repository` | New |

### P3 — Case file or loose-paper packet

Court case files, probate packets, original wills, naturalization petitions, land entry files. The file is the unit; the item is one document inside it.

> **L:** Kansas. Ottawa County. Probate Court. Estate Files. Probate Court, Minneapolis.
>
> **F:** Ottawa County, Kansas, Probate Court, estate file 112, John Smith (1885), inventory, 2 March 1885; Probate Court, Minneapolis.
>
> **S:** Ottawa Co., Kans., estate file 112, John Smith, inventory.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Jurisdiction, court or office | L F S | `S.jurisdiction`, `S.contributors` | New |
| File type and number | F S | `X.ref[case_number]` | New |
| Principal party and year | F S | `G.person` (or parties), `S.date` | Graph / Planned |
| Document in the file | F S | `C.item`, else `C.label` | Partial |
| Document date | F | `G.date` | Graph |
| Repository | L F | `S.repository` | New |

### P4 — Vital certificate

An issued civil certificate: birth, death, marriage, delayed birth. In Provenencia this is usually one Source per certificate.

> **L:** Kansas. State Board of Health. Birth Certificates. Office of Vital Statistics, Topeka.
>
> **F:** Kansas State Board of Health, birth certificate no. 1902-04417, Ada Robins, 12 February 1902; Office of Vital Statistics, Topeka.
>
> **S:** Kans. birth cert. no. 1902-04417, Ada Robins.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Issuing authority and jurisdiction | L F S | `S.contributors` (creating agency), `S.jurisdiction` | New |
| Certificate type | L F S | `S.type` label | Have |
| Certificate or file number | F S | `X.ref[certificate_number]` | New (`S.meta[certificate_number]` horizon key) |
| Registration or volume and page (registers) | F | `X.ref[volume]`, `C.page_label` | New / Have |
| Person named | F S | `G.person` | Graph |
| Event date | F | `G.date` | Graph |
| Filing or issue date (delayed certificates) | F | `S.date` | Planned |
| Repository (issuing office) | L F | `S.repository` | New |

### P5 — Census

Every census in Chapter 6. The richest set of reference parts.

> **L:** Kansas. Ottawa County. 1880 U.S. census, population schedule. Digital images. Ancestry. https://www.ancestry.com : 2025.
>
> **F:** 1880 U.S. census, Ottawa County, Kansas, population schedule, Minneapolis, enumeration district (ED) 236, p. 14 (stamped), dwelling 112, family 118, Ada Robins; digital image, Ancestry (https://www.ancestry.com : accessed 12 March 2025); citing NARA microfilm publication T9, roll 393.
>
> **S:** 1880 U.S. census, Ottawa Co., Kans., pop. sch., Minneapolis, ED 236, p. 14 (stamped), dwell. 112, fam. 118, Ada Robins.

British variant: "1851 census of England, Yorkshire, York, St Michael le Belfrey, folio 23, p. 4, schedule 17, household of John Smith; …; citing The National Archives, HO 107/2353."

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Census year | L F S | `S.date` (year) | Planned |
| Census authority ("U.S.", state, "of England") | L F S | `S.contributors` (creating agency) or the jurisdiction's top part | New |
| Jurisdiction: state, county, civil division | L F S | `S.jurisdiction` | New |
| Schedule (population, agriculture, mortality, slave) | L F S | `X.ref[schedule]` | New |
| Enumeration district; district and sub-district (Canada) | F S | `X.ref[ed]` | New |
| Sheet, page, folio; which numbering is used (stamped, penned) | F S | `X.ref[sheet]` / `C.page_label` / `X.ref[folio]` | Partial |
| Piece (British) | F | `X.ref[piece]` | New |
| Schedule number (British household) | F | `X.ref[schedule_number]` | New |
| Dwelling, family | F S | `X.ref[dwelling]`, `X.ref[family]` | New |
| Line | F | `X.ref[line]` | New |
| Person or household named | F S | `G.person` (head of household, or the person cited) | Graph |
| Layer 2 and 3 (image, film, citing) | L F | P14 / P15 | New |

### P6 — Church register

> **L:** St Michael le Belfrey (York, Yorkshire, England). Register of Baptisms, 1813–1837. Borthwick Institute for Archives, York.
>
> **F:** St Michael le Belfrey (York, Yorkshire, England), *Register of Baptisms, 1813–1837*, vol. 3, p. 23, no. 182, baptism of Ada Smith, 3 March 1830; Borthwick Institute for Archives, York.
>
> **S:** St Michael le Belfrey, Register of Baptisms, 3: 23, no. 182.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Church or congregation (creator) | L F S | `S.contributors` (creating body) | New; an `organization` Subject later ([`organizations.md`](../../ideas/organizations.md)) |
| Church's place | L F | `S.jurisdiction` | New |
| Register title or description, and its span | L F S | `S.title` + `S.title_kind` + `S.span` | Have / New |
| Volume, page, entry number | F S | `X.ref[volume]`, `C.page_label`, `X.ref[entry]` | Partial |
| Event and person ("baptism of Ada Smith") | F | `G.event` + `G.person` | Graph |
| Event date | F | `G.date` | Graph |
| Repository (parish chest, diocesan or county archive) | L F | `S.repository` | New |

### P7 — Cemetery marker

> **L:** Kansas. Ottawa County. Minneapolis. Highland Cemetery. Grave markers.
>
> **F:** Highland Cemetery (Minneapolis, Ottawa County, Kansas), Ada Robins marker, section 4, lot 12; photographed by the author, 14 May 2024.
>
> **S:** Highland Cemetery (Minneapolis, Kans.), Ada Robins marker.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Cemetery name | L F S | `S.title` or `S.contributors` | Have / New |
| Cemetery location | L F S | `S.jurisdiction` | New |
| Person on the marker | F S | `G.person` | Graph |
| Section, lot, row | F | `X.ref[section]`, `X.ref[lot]` | New |
| Who read or photographed it, and when | F | `A.holder` / `S.contributors` (photographer), `A.accessed` or `S.date` | New |

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
| Title or description of the record | L F S | `S.title` + `S.title_kind` | Have / New |
| Span | L F | `S.span` | New |
| Creator or compiler (writer, photographer, studio, testing company) | L F S | `S.contributors` | New (horizon keys `photographer`, `studio`, `testing_company`) |
| Imprint of the containing book (Bible) | L F | `S.container` + `S.publication` | New |
| Item: page, letter date, match entry | F S | `C.item` / `C.label`, `G.date` | Partial |
| Medium (cabinet card, tintype) | F | `A.format` | New (horizon key `medium`) |
| Holder and place, year seen | L F | `A.holder`, `A.accessed` | New |
| Kit or match identifiers | F | `X.ref[kit]`, `X.ref[match]` | New (horizon keys) |
| Software and export date (GEDCOM) | F | `X.ref[software]`, `S.date` | New / Planned |

### P9 — Interview, oral history and personal correspondence

> **L:** Doe, Jane. Interview by the author. Calgary, Alberta, 2 February 2024. Recording privately held by the author.
>
> **F:** Jane Doe (Calgary, Alberta), interview by the author, 2 February 2024; recording privately held by the author.
>
> **S:** Doe, interview, 2 February 2024.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Interviewee or writer | L F S | `S.contributors` (interviewee) | New (horizon key `interviewee`) |
| Interviewer or recipient | L F | `S.contributors` (interviewer) | New |
| Place | L F | `S.jurisdiction` (place of interview) | New |
| Date | L F S | `S.date` | Planned |
| Where the recording or message is held | L F | `A.holder` | New |
| Timestamp of the cited passage | F | `C.item`; a time selector later ([`audio-video-sources.md`](../../ideas/audio-video-sources.md)) | Partial |

### P10 — Book and other standalone publication

Books, compiled and edited works, chapters, multi-volume works, published abstracts and transcriptions, CDs, maps, films.

> **L:** Smith, Alice. *The Smiths of Ottawa County*. Topeka: Prairie Press, 1965.
>
> **F:** Alice Smith, *The Smiths of Ottawa County* (Topeka: Prairie Press, 1965), 42.
>
> **S:** Smith, *Smiths of Ottawa County*, 42.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Authors, editors, compilers, translators, with roles and order | L F S | `S.contributors` | New (horizon key `author` is one string) |
| Title (italic), subtitle | L F S | `S.title` + `S.title_kind` | Have / New |
| Short title (short note) | S | Derived from `S.title`, with an override | New (derived) |
| Edition, volume, series | L F | `S.publication` (edition), `X.ref[volume]`, `X.ref[series]` | New |
| Place of publication, publisher | L F | `S.publication` | New |
| Year | L F | `S.date` | Planned |
| Chapter author and title (essay in an edited book) | L F | `S.contributors` + `S.title`, with the book as `S.container` | New |
| Page | F S | `C.page_label` | Have, composer gap |
| Medium (CD-ROM, DVD, map scale) | L F | `A.format` or `S.publication` | New |

### P11 — Periodical article and broadcast

Journal and magazine articles, newsletters, radio, TV and podcasts.

> **L:** Jones, Robert. "The Robins Family of York." *Yorkshire Genealogist* 12 (June 1990): 45–60.
>
> **F:** Robert Jones, "The Robins Family of York," *Yorkshire Genealogist* 12 (June 1990): 45–60, specifically 47.
>
> **S:** Jones, "Robins Family of York," 47.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Author | L F S | `S.contributors` | New |
| Article title (quoted) | L F S | `S.title` | Have |
| Periodical title (italic) | L F | `S.container` | New |
| Volume, issue | L F | `X.ref[volume]`, `X.ref[issue]` | New |
| Issue date | L F | `S.date` | Planned |
| Article's page range | L F | `X.ref[pages]` | New |
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
| Headline, or a description of an untitled item | F S | `S.title` + `S.title_kind` | Have / New |
| Newspaper title, with its place in parentheses | L F S | `S.container` + `S.jurisdiction` | New |
| Issue date | L F S | `S.date` | Planned |
| Edition (morning, evening) | F | `S.publication` (edition) | New |
| Page, column | F | `C.page_label`, `X.ref[column]` | Partial |

### P13 — Website, web page and online database

Web pages, blog posts, online indexes and databases, online trees, memorial sites, message boards. The Source *is* the online publication.

> **L:** "Kansas, Births, 1880–1910." Database. FamilySearch. https://www.familysearch.org : 2025.
>
> **F:** "Kansas, Births, 1880–1910," database, FamilySearch (https://www.familysearch.org : accessed 12 March 2025), entry for Ada Robins, 12 February 1902; citing Ottawa County birth register, vol. 2, p. 41.
>
> **S:** "Kansas, Births, 1880–1910," entry for Ada Robins, 1902.

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Creator or owner of the page or tree | L F | `S.contributors` | New |
| Page or database title (quoted) | L F S | `S.title` | Have |
| Kind of item ("database", "database with images", "blog") | L F | `A.format` | New |
| Website title (italic) | L F | `S.container` or `A.provider` | New (horizon key `site_name`) |
| URL | L F | `A.url` | New (horizon key `url`) |
| Date accessed | L F | `A.accessed` | New (horizon key `accessed_date`) |
| Entry for whom | F S | `G.person`, `G.date` | Graph |
| Memorial, record or profile id | F | `X.ref[record_id]` | New |
| Citing (the original the database rests on) | F | `A.citing` | New |

### P14 — Online image layer (layer 2)

Not a standalone pattern: the second layer appended to P1–P12 when the researcher saw the record as an image on a website.

> "…; digital image, Ancestry (https://www.ancestry.com : accessed 12 March 2025), image 37 of 412; citing NARA microfilm publication T9, roll 393."

| Element | Source | Status |
| --- | --- | --- |
| "digital image" | `A.format` | New |
| Provider or website | `A.provider` | New |
| Database or collection title (when cited) | `A.database` | New |
| URL | `A.url` (a Citation can override with its image URL) | New |
| Accessed date | `A.accessed` | New |
| Image number in the online set | `C.page` (page in the file), when the Artifact is the whole image set | Have |
| Citing | `A.citing`, or the film reference in P15 | New |

### P15 — Microfilm layer (layer 2)

> "…; FHL microfilm 1,234,567, item 3." or "…; NARA microfilm publication T9, roll 393."

| Element | Source | Status |
| --- | --- | --- |
| "microfilm" | `A.format` | New |
| Who filmed or publishes the film (NARA, FHL, a state archive) | `A.provider` | New |
| Publication number, roll, film number, item | `A.film` (or `X.ref[roll]`, `X.ref[film]`, `X.ref[item]` on the Artifact) | New (horizon key `roll`) |
| Frame or image | `C.page` | Have |

### P16 — Legal publication

Statutes, session laws, codes, legislative documents, published court reports. These follow legal citation conventions, compact and number-led.

> **F (statute):** Kansas, *Session Laws*, 1885, chap. 12, sec. 3, p. 45.
>
> **F (case):** *Smith v. Jones*, 12 Kan. 345 (1874).

| Element | Forms | Source | Status |
| --- | --- | --- | --- |
| Jurisdiction | L F S | `S.jurisdiction` | New |
| Compilation or reporter title, abbreviated | L F S | `S.container` (with an abbreviation) | New |
| Session or year | L F | `S.date` | Planned |
| Chapter, section, volume, page | F S | `X.ref[chapter]`, `X.ref[section]`, `X.ref[volume]`, `C.page_label` | New / Have |
| Case name | F S | `S.title` | Have |

### P17 — Fallback

Any Source whose type maps to no pattern, or whose pattern is missing a required element, gets today's line: Source title, Artifact label, Citation label, page ([`citation-formats.md`](citation-formats.md) "What ships in S13-11"). The citation preview names the missing elements so the researcher can fill them.

---

## 5. Missing fields and how to add them

### 5.1 The full list

Each row is one thing the catalog does not hold today. **Patterns** says how widely it is needed.

| Field | Level | Patterns | Recommended implementation |
| --- | --- | --- | --- |
| Source date | Source | all | **Planned**: `sources.date_value_id` ([`source-provenance-date.md`](source-provenance-date.md)) |
| Covered span ("1813–1837") | Source | P1, P2, P6, P8 | A second, optional date on `sources` (span end), so the Source date is the span's start. A DateValue `range` is an uncertainty window, not a period ([`structured-date-model.md`](../../structured-date-model.md)), so it cannot hold a span. Alternatively, leave the span in the title text |
| Title kind (exact title or description) | Source | all | Column `sources.title_kind` (`exact` \| `description`), defaulted by pattern |
| Contributors (author, editor, compiler, translator, creating agency, photographer, interviewer, interviewee) | Source | all | Table `source_contributors (source_id, position, role, display_name, sort_name)`. Text only, never resolved to Subjects (Source layer stays descriptive). `role` from a compiled product list |
| Jurisdiction | Source | P2–P7, P9, P12, P16 | Table `source_jurisdiction_parts (source_id, position, name)`, largest first, so the formatter can write either order. Text, not place Subjects (§ 6) |
| Repository | Source | P1–P7, P8 (holder) | New record `repositories (id, ref REP-…, name, location, url)` and `sources.repository_id`. Shared by many Sources, picked from a list, has its own small page later |
| Container (journal, newspaper, website, book holding a chapter, statute compilation) | Source | P8, P10–P13, P16 | Column `sources.container_title` (with an optional `container_short`), not a parent Source, which stays parked ([`source-structure-later.md`](source-structure-later.md) "Nested Sources") |
| Publication (place, publisher, edition) | Source | P8, P10–P12 | Columns `publication_place`, `publisher`, `edition` on `sources`, or reference parts of those kinds (§ 5.2) |
| Reference parts (volume, book, folio, entry, line, ED, sheet, dwelling, family, schedule, piece, certificate number, case number, record group, series, collection, call number, box, folder, issue, column, section, chapter, roll, item, record id, kit, match) | Source, Artifact or Citation | all | Typed tables `source_ref_parts`, `artifact_ref_parts`, `citation_ref_parts`, each `(owner_id, kind, position, value_text)`. `kind` is a **compiled Go registry** (§ 5.2) |
| Printed page in the composer | Citation | all | Composer writes `page_label` on the `page` selector; the field already exists ([`citation-formats.md`](citation-formats.md) step 3) |
| Copy format | Artifact | P8, P13–P15 | Column `artifacts.format` from a compiled list: `original`, `photocopy`, `certified_copy`, `microfilm`, `digital_image`, `database_entry`, `transcript`, `abstract`, `recording`, `photograph` |
| Provider, database title | Artifact | P13, P14, P15 | Columns `artifacts.provider`, `artifacts.database_title` (text). A shared provider record is not needed yet: the list of providers is short and stable |
| URL | Artifact (Citation override) | P13, P14 | Column `artifacts.url`; a Citation-level override as a `citation_ref_parts` row of kind `url` |
| Accessed date | Artifact | P8, P13, P14 | Column `artifacts.accessed_on`, an ISO calendar date. It is always one exact day the researcher knows, so a DateValue's uncertainty buys nothing |
| Film reference | Artifact | P15 | Reference parts on the Artifact: `film_publication`, `roll`, `film`, `item` |
| Citing | Artifact | P13, P14 | Column `artifacts.citing_text`. A structured link to the original stays with derivation links ([`source-structure-later.md`](source-structure-later.md)) |
| Holder of a private original | Artifact | P7, P8, P9 | Column `artifacts.holder_text` ("privately held by Jane Doe, Calgary, Alberta") |
| Item description | Citation | all | Derived from the graph (§ 5.3); `citations.label` remains the override |
| Pattern for a Source type | Source type | all | Column `source_types.citation_pattern`. Seeded types come mapped; a researcher's type picks one, default fallback |
| Role terms for parties | Graph vocabulary | P1, P2, P3 | Seed `role` terms the formatter reads: `grantor`, `grantee`, `writer`, `recipient`, `testator`, `head_of_household` ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 3.5) |
| Formatter | Engine | all | `core/citationformat`: one Go template per pattern, three forms each, replacing `CitationLineDisplay` ([`citation-formats.md`](citation-formats.md) step 4) |

### 5.2 Recommendations behind the list

**Fill at any level; resolve from the most specific.** The formatter reads each element from the Citation first, then the Artifact, then the Source. That answers the open question "do template fields live on the Source, the Artifact or the Citation" without forcing one grain on every researcher: a researcher who files one Source per census page puts the sheet number on the Source; one who files the whole county as a Source puts it on each Citation. The citation preview shows which level each value came from.

**A compiled registry, not seeded vocabulary.** Reference-part kinds, contributor roles, copy formats and patterns are product code, like name part types ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 4.1), not catalog rows. Two reasons: the formatter's templates are code and must know every kind they read, and seeded vocabulary is installed only in new catalogs ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 1), so existing projects would never get the fields. A compiled registry reaches every project on upgrade with no healing.

**Few new columns.** A field gets its own column (or record) only when something other than the formatter uses it: the date (sort), the repository (shared and listed), the jurisdiction (two orders), contributors (roles and sort names), and the Artifact's access fields (one per copy). Everything else is a reference part.

**This settles most of the metadata rethink.** Of the 48 horizon metadata keys in [`seeded-vocabulary.md`](../../seeded-vocabulary.md) § 2.2, all but a handful (`isbn`, `language`, `form_number`, `gedcom_version`) become reference parts, contributors, Artifact access fields or the Source date above. What would be left for `source_metadata` is free-form filing text the formatter never reads. That supports shrinking metadata to plain researcher notes-as-fields, or removing it and moving existing values into Source notes or reference parts during the migration ([`2026-10-25-rethink-source-metadata-and-type-suggestions.md`](2026-10-25-rethink-source-metadata-and-type-suggestions.md)). The type → field suggestions would be replaced by each pattern's list of elements, shown in the citation preview.

### 5.3 Reading the item from the graph

This is where "build the graph, get the citation" pays off. For a Citation, the formatter collects the Subjects whose Observations cite it and composes the item part:

1. If the Citation has a `label`, use it (the researcher's override).
2. Else, find the event Subject with Observations on this Citation: its `event_type` gives "baptism of", its `date` gives the item date, and its principal participant's recorded `name` gives the person.
3. Else, find a person Subject: "entry for *name*" (P13), or "household of *name*" when the role is head of household (P5).
4. Else, write the pattern's generic item ("entry", "page").

Names and dates are the ones the Source records (Observations), never the canonical person's kept name: a citation describes the record, not the conclusion.

### 5.4 Suggested order

Following [`citation-formats.md`](citation-formats.md) "How it could arrive":

1. **One Source schema pass:** Source date and span, title kind, contributors, jurisdiction parts, repository record, container, `source_types.citation_pattern`, reference-part tables. One migration, one Source page design.
2. **Artifact access fields:** format, provider, database title, URL, accessed date, citing, holder. Can share the migration; the Artifact edit form needs its own design.
3. **Printed page labels** in the composer, and Citation reference parts in the composer.
4. **Formatter** with P17 and the patterns the dogfood Sources need first. Proposed first set: P5 census, P4 vital certificate, P6 church register, P10 book, P12 newspaper, P13 web database, with the P14 layer.
5. **Citation preview** on the Source page, naming missing elements.
6. **Copy citation**, full or short.
7. Remaining patterns, one PR or batch each.

---

## 6. Open questions

- **Jurisdiction from place Subjects?** A Source's jurisdiction is text here, which keeps the Source layer descriptive. The graph often already has the place chain (town → county → state through `part_of`). Offering "fill jurisdiction from this graph's place" as a one-time copy into the text parts would save typing without making the Source depend on interpretation.
- **Principal Subject.** When a Citation supports several people (a census household, a marriage), which one names the item: head of household, both spouses, the person the researcher opened the composer from?
- **Span: column or text?** A second date on `sources`, or only the title text.
- **Which patterns first.** The proposed first set in § 5.4 is a guess; the product owner's dogfood Sources decide.
- **Editable templates.** Whether a researcher can adjust a pattern's wording, or only fill fields. Recommendation: fields only in Spike 15; a per-Source "edited citation" override text field if the formatter is wrong for an odd record.
- **Provider as a record.** Kept as text on the Artifact here. Revisit if researchers want a list of everything they took from one provider.
- **Metadata.** Shrink to filing text, or remove and migrate (§ 5.2).
- **Edition check.** Verify § 2 against the edition of *Evidence Explained* on hand, including how it treats DNA results and online trees.
