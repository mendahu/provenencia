# Provenencia Genealogy — Seeded Vocabulary

## Status

Draft **horizon catalog** of keys and open-vocabulary values Provenencia may seed as workflows need them.

This document is the place to record intended types, metadata fields, Properties, and picker values so they do not fork across layer docs. It is **not** a required insert list for the first implementation.

Table schemas and architectural rules remain in the layer and shared-value documents. When a key is seeded, it is data rather than a SQL enum: researchers may still add keys without migrations. First-class application behavior may recognize well-known keys.

Related schema docs:

- Source: [`source-layer-data-model.md`](source-layer-data-model.md)
- Interpretation: [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md)
- Conclusion: [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md)
- Names: [`structured-name-model.md`](structured-name-model.md)
- Dates: [`structured-date-model.md`](structured-date-model.md)

Items marked **TBD** are expected seeds whose exact set is still being refined.

---

# 1. Principles

1. **Implement small; grow from use.** Early project databases should seed only what a concrete workflow requires. Prefer adding a field when cataloging a real Source over shipping this entire list on day one.
2. This file may stay ahead of the product. Keeping a larger catalog here is documentation, not a commitment to pre-make unused pickers, metadata fields, or event types.
3. Seeds are data inserted into a project database, not SQL enums.
4. Built-in / shipped keys are not structurally privileged subclasses; they are convenient defaults with optional first-class UX.
5. Open text vocabularies (`event_type`, `role`, name part `type`, and similar) stay extensible even when a starter list exists for pickers.
6. Do not seed a fine-grained source-quality ontology (`is_authentic`, defect codes, and similar) on Source catalog rows or as Observation defect codes unless a concrete workflow requires it. First-class **Source credibility** uses the three-point assessment vocabulary in §3.0 and [`research-judgment-model.md`](research-judgment-model.md), not ad hoc Source metadata.
7. Expanding this catalog does not require a schema migration when the underlying tables already use open keys.
8. Source credibility grades and Claim confidence grades share a three-point *shape* but **must not share keys or labels** — they answer different questions.

---

# 2. Source layer

Schema: [`source-layer-data-model.md`](source-layer-data-model.md).

## 2.1 `source_types`

```text
birth_certificate
census
photograph
oral_testimony
interview
dna_match_report
book
website
gedcom_file
```

`website` is the evidentiary website or page. A local capture (screenshot, WARC, PDF print) is an Artifact of that Source, not a separate Source type.

`interview` is a recorded or published conversation. `oral_testimony` is informal family testimony that may not have a formal interviewer.

## 2.2 `source_metadata_fields`

Catalog fields only. Values stay descriptive text (or structured dates); they do not resolve to Interpretation Nodes or canonical entities.

```text
key                     data_type   typical use
creator                 text        generic credited creator when a more specific role field is not used
author                  text        book, website, GEDCOM submitter
photographer            text        photograph
studio                  text        photograph studio or imprint
interviewer             text        interview, oral testimony
interviewee             text        interview, oral testimony
publisher               text        book, website host/organization
publication_place       text        book imprint place
edition                 text        book
isbn                    text        book
series                  text        book series; census microfilm series
volume                  text        book volume; vital-record volume
page                    text        catalog page/sheet identity for a record-as-source
url                     text        originating website, vendor report, or download location
site_name               text        website
software                text        GEDCOM exporting program
gedcom_version          text        GEDCOM
testing_company         text        DNA match report
kit_number              text        DNA kit belonging to the researcher
match_identifier        text        vendor match name or kit id as printed on the report
medium                  text        photograph process (cabinet card, daguerreotype, …)
jurisdiction            text        civil records, census geography as printed
civil_division          text        census township, ward, or equivalent
enumeration_district    text        census
sheet                   text        census sheet
line_number             text        census
dwelling_number         text        census
family_number           text        census
roll                    text        microfilm roll
certificate_number      text        vital record
registration_number     text        vital record local/state file number
form_number             text        vital record or census form
repository              text        archive, library, or civil office holding the evidence
collection              text        archival or vendor collection name
call_number             text        repository shelf/call number
box                     text        archival box
folder                  text        archival folder
record_group            text        archival record group
language                text        interview, oral testimony, book, website
publication_date        date        book, website page
issue_date              date        certificate issuance
record_date             date        date the record was made or filed
census_date             date        census day / enumeration year
taken_date              date        photograph
interview_date          date        interview, oral testimony
accessed_date           date        website
export_date             date        GEDCOM
report_date             date        DNA match report
```

`url` and `accessed_date` describe Source provenencia. The bytes of a capture still live on an Artifact/File.

## 2.3 `source_type_metadata_fields`

Suggested fields per type for cataloging UI. Not required; a Source may use fields not listed for its type.

```text
birth_certificate
  jurisdiction
  certificate_number
  registration_number
  form_number
  volume
  page
  record_date
  issue_date
  repository
  collection
  call_number
  box
  folder

census
  jurisdiction
  civil_division
  census_date
  enumeration_district
  sheet
  page
  line_number
  dwelling_number
  family_number
  form_number
  series
  roll
  repository
  collection
  call_number

photograph
  photographer
  studio
  taken_date
  medium
  creator
  repository
  collection
  call_number
  box
  folder

oral_testimony
  interviewee
  interviewer
  interview_date
  language
  creator

interview
  interviewee
  interviewer
  interview_date
  language
  publisher
  publication_date
  url
  accessed_date

dna_match_report
  testing_company
  kit_number
  match_identifier
  report_date
  url
  accessed_date
  collection

book
  author
  publisher
  publication_place
  publication_date
  edition
  isbn
  series
  volume
  repository
  collection
  call_number
  language

website
  url
  site_name
  author
  publisher
  publication_date
  accessed_date
  language

gedcom_file
  author
  software
  gedcom_version
  export_date
  url
  accessed_date
```

---

# 3. Interpretation layer

Schema: [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md). Judgment semantics: [`research-judgment-model.md`](research-judgment-model.md).

## 3.0 Source credibility grades

Open vocabulary for `source_credibility_assessments.credibility_key`. Three-point scale with a baseline middle. **Do not** reuse Claim confidence keys.

```text
key             sort_order    label (draft product copy)
low_trust       1             Low trust
standard        2             Standard
high_trust      3             High trust
```

Missing assessment may display as Standard without inserting a row.

## 3.1 `node_types`

```text
key             ref_prefix    summary
person          PER           A person represented by interpreted evidence.
event           EVT           An occurrence represented by interpreted evidence.
place           PLC           A geographic feature at one grain (town, township, colony, farm, …).
relationship    REL           A general association when evidence is not a more specific event/context structure.
participation   PTN           Association between a person and an event, including role.
location        LOC           Association between an event and a place.
source          SRN           A Source reified so other evidence can refer to or comment on it.
```

`ref_prefix` is required for every Node Type, including researcher-defined types. Do not reuse reserved prefixes `SRC`, `ART`, `CIT`, `OBS`, or the candidate layer code `C`. Canonical refs are `{prefix}-{token}`; Node refs are `{prefix}-C-{token}`. `SRN` is used for source-Nodes so they do not collide in speech with Source catalog refs (`SRC-…`).

## 3.2 `properties`

```text
key                 value_type    notes
name                name
name_format         text          primarily Conclusion reconciliation; profile key
birth_date          date
age_at_event        integer
occupation          text
event_type          text          open values; see §3.4
date                date          event date
role                text          open values; see §3.5
relationship_type   text          open values; TBD starter list
person              node          app target hint: person
event               node          app target hint: event
place               node          app target hint: place
participant         node          app target hint: person
mentions            node          app target hint: source
remark              text          free-text commentary about a source Node
toponym             text          place name as interpreted from a Source (not a personal NameValue)
```

Target Node Type hints are application-only (not SQL allow-lists). See the Interpretation doc.

Additional Properties may be seeded as workflows need them (shared DNA, predicted relationship, and similar). Treat those as **TBD** until a concrete UI requires them.

## 3.3 `node_type_properties`

```text
node_type       property
person          name
person          birth_date
person          occupation

event           event_type
event           date

place           toponym

participation   person
participation   event
participation   role

location        event
location        place

relationship    participant
relationship    relationship_type

source          mentions
source          remark
```

`name_format` is not required on Interpretation `person` Nodes unless a Source itself asserts a naming convention.

## 3.4 Open values: `event_type`

Starter picker values (**extensible**):

```text
birth
death
marriage
baptism
burial
census
residence
immigration
military
probate
other
```

Exact GEDCOM alignment and additional vital/event kinds are **TBD**.

## 3.5 Open values: `role` (participation)

Starter picker values (**extensible**):

```text
subject
father
mother
spouse
child
witness
informant
other
```

## 3.6 Open values: `relationship_type`

Starter picker values are **TBD** (for example `cousin`, `guardian`, household roles). Keep open text; seed only what pickers need.

---

# 4. Structured names

Schema: [`structured-name-model.md`](structured-name-model.md).

## 4.1 Name part `type` starter set

GEDCOM-aligned open vocabulary:

```text
prefix              -- NPFX
given               -- GIVN
initial
nick                -- NICK
surname_prefix      -- SPFX
surname             -- SURN
suffix              -- NSFX
undetermined
```

## 4.2 `name_format_profiles`

```text
key = western
label = Western
description = GEDCOM-aligned spoken/display order
```

## 4.3 `name_format_profile_parts` (`western`)

```text
idx   part_type
0     prefix
1     given
2     initial
3     nick
4     surname_prefix
5     surname
6     suffix
```

## 4.4 `project_settings`

```text
default_name_format_key = western
```

Additional cultural profiles (for example dual-surname ordering) are **TBD**.

---

# 5. Conclusion layer

Schema: [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

Conclusion seeds are mostly workflow vocabularies, not large type catalogs.

## 5.1 Sameness claim `status`

Closed workflow vocabulary (not researcher-extensible like `event_type`):

```text
provisional       -- persisted; UI should distinguish from accepted; not in membership
accepted          -- only accepted same_as participates in membership closure
rejected          -- retained for audit; not in membership
```

`superseded` is omitted until a concrete workflow needs it. Changing a conclusion updates the existing claim row (and is audited) rather than inserting a successor claim.

## 5.2 Reconciliation claim `status`

Same closed vocabulary as Sameness Claims:

```text
provisional       -- persisted working value; distinct UI; not the committed conclusion
accepted          -- the committed concluded value for that entity + Property
rejected          -- retained for audit; not the working value
```

## 5.3 Reconciliation `origin`

Not used. Soft blends are display-only. A Reconciliation Claim is a researcher-persisted row or it is absent.

## 5.4 Canonical handles

There is no `origination_claims` table and no `representative_node_id`. A `canonical_entities` row may have a nullable `identity_anchor_id` and optional `argument`. For node-typed Reconciliation values, Conclusion uses `value_entity_id` (canonical), not Observation `value_node_id`.

Handle provenencia badges (from records / inferred / asserted / unlinked) are computed, not seeded workflow enums.

## 5.5 Claim confidence grades

Open vocabulary for `sameness_claims.confidence_key` and `reconciliation_claims.confidence_key`. Three-point scale with a baseline middle. **Do not** reuse Source credibility keys — same shape, different semantics ([`research-judgment-model.md`](research-judgment-model.md)).

```text
key                 sort_order    label (draft product copy)
low_confidence      1             Low confidence
moderate            2             Moderate
high_confidence     3             High confidence
```

`confidence_key` is nullable. Claim `status` remains the separate workflow vocabulary in §5.1–5.2.

---

# 6. Documentation ownership

- This document is the horizon catalog for intended keys, starter open-vocabulary lists, and project-default settings. It does not require every listed row to ship in the first app version.
- Layer and shared-value documents remain authoritative for table schemas and invariants.
- [`research-judgment-model.md`](research-judgment-model.md) is authoritative for judgment semantics (Source credibility, Citation transcription certainty, Claim confidence).
- When a key is actually seeded in product, update this catalog if needed; do not fork competing lists into layer docs.
