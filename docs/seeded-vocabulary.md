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
4. Shipped keys are not structurally privileged subclasses; they are convenient defaults with optional first-class UX. Their privilege is **origin**, not a separate table or enum.
5. **Property terms** (`event_type`, `role`, `relationship_type`, and similar kind/edge vocabularies) are vocabulary-definition rows with `origin` (§1.1), not free-text Observation strings. **Term-typed Properties are Install/registry only** (product or plugin) — researchers do not create Properties with `value_type = term`. Product/plugin Install seeds large term sets; researchers may add `origin=user` **term rows** under those Properties without a dedicated Event types / Roles admin destination. True prose Properties (`remark`, `toponym`) stay `value_type = text`. Name part `type` on NameValue remains a separate open part vocabulary.
6. Do not seed a fine-grained source-quality ontology (`is_authentic`, defect codes, and similar) on Source catalog rows or as Observation defect codes unless a concrete workflow requires it. First-class **Source credibility** uses the three-point assessment vocabulary in §3.0 and [`research-judgment-model.md`](research-judgment-model.md), not ad hoc Source metadata.
7. Expanding this catalog does not require a schema migration when the underlying tables already use open keys.
8. Source credibility grades and Claim confidence grades share a three-point *shape* but **must not share keys or labels** — they answer different questions.

## 1.1 Vocabulary origin (namespace)

Vocabulary-definition tables — rows the UI lists as managed taxonomy, not instance data — carry an **`origin`** column. This is the reserved namespace for who contributed the term. It is **not** an evidentiary Source (`sources`).

Applies to (authoritative schemas in the linked docs):

| Table | Doc |
| --- | --- |
| `source_types`, `source_metadata_fields` | [`source-layer-data-model.md`](source-layer-data-model.md) |
| `subject_types`, `properties`, `property_terms`, `source_credibility_grades` | [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md), [`research-judgment-model.md`](research-judgment-model.md) |
| `claim_confidence_grades` | [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md), [`research-judgment-model.md`](research-judgment-model.md) |
| `name_format_profiles` | [`structured-name-model.md`](structured-name-model.md) |

Does **not** apply to join/suggestion tables (`source_type_metadata_fields`, `subject_type_fields`, `name_format_profile_parts`), domain instance rows, or NameValue part `type` strings (open part vocabulary, not `property_terms`).

### Reserved `origin` values

```text
provenencia          -- product-seeded vocabulary shipped with the app
user                 -- created by the researcher in this project
plugin:<plugin_id>   -- reserved prefix for future plugin-contributed vocabulary
```

`<plugin_id>` is a stable plugin slug (lowercase ASCII letters, digits, hyphens). No other `origin` prefixes are reserved yet; do not invent ad hoc origins outside these three shapes without updating this catalog.

### Uniqueness and foreign keys

- Uniqueness is **`UNIQUE (key, origin)`**, not bare `key`. The same machine `key` (and user-visible label) may exist under different origins so a later product seed never collides with a user- or plugin-created term.
- Vocabulary rows use a UUID **`id`** primary key. Domain tables reference vocabulary **by `id`**, never by bare `key`, so colliding keys across origins stay unambiguous.
- Application recognition of well-known shipped terms looks up `(key, origin = 'provenencia')` (optional first-class UX). Using a term on a Source, subject, or Claim always stores the vocabulary row’s `id`.
- UI should surface origin (badge/caption: app / custom / plugin) without treating non-`provenencia` rows as second-class when *using* them.
- Vocabulary rows of any origin may be **deleted when unused** (no Sources referencing a type; no `source_metadata` referencing a field). Delete is refused with a conflict error while in use. Suggestion joins cascade on delete.
- Subject type `ref_prefix` remains **globally** unique across origins (refs must not collide in speech).

There is no separate `builtin` boolean; `origin = 'provenencia'` replaces that flag.

### Implementation (Source seed today)

Product-seeded Source types/fields/suggestions live in a **code registry** (`core/database/sourcevocab/registry.go`) and are installed once by `sourcevocab.Install` at catalog **create** via `onboarding.createCatalog` (see `core/onboarding/ready.go`). Opens (`OpenCatalog`) do **not** re-install or heal deleted seed rows. Do not put seed rows in SQL migrations. The current create-time starter is a single `birth_certificate` type plus a few suggested fields — the §2 lists below remain a horizon catalog of suggested keys, not a ship commitment.

When extending or adding another vocabulary domain’s seed, follow the project skill [`.cursor/skills/add-seeded-vocabulary/SKILL.md`](../.cursor/skills/add-seeded-vocabulary/SKILL.md).

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

Catalog fields only. Values stay descriptive text (or structured dates); they do not resolve to Interpretation subjects or canonical entities.

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

Open vocabulary for `source_credibility_assessments.credibility_grade_id` (via `source_credibility_grades`). Three-point scale with a baseline middle. **Do not** reuse Claim confidence keys.

```text
key             sort_order    label (draft product copy)
low_trust       1             Low trust
standard        2             Standard
high_trust      3             High trust
```

Missing assessment may display as Standard without inserting a row.

## 3.1 `subject_types`

```text
key             ref_prefix    candidate_ref_prefix    summary
person          PER           CPR                     A person represented by interpreted evidence.
event           EVT           CEV                     An occurrence represented by interpreted evidence.
place           PLC           CPL                     A geographic feature at one grain (town, township, colony, farm, …).
relationship    REL           CRL                     A general association when evidence is not a more specific event/context structure.
participation   PTN           CPA                     Association between a person and an event, including role.
location        LOC           CLO                     Association between an event and a place.
source          SRN           CSR                     A Source reified so other evidence can refer to or comment on it.
```

Both prefixes are required for every Subject type, including researcher-defined types. `ref_prefix` mints canonical entity refs in the Conclusion layer (`PER-7KD45`); `candidate_ref_prefix` mints Interpretation subject refs (`CPR-7KD45`). Same format, different prefix — see [`catalog-refs.md`](catalog-refs.md) §2.

Do not reuse the reserved catalog prefixes `USR`, `SRC`, `ART`, `CIT`, `OBS`. Both columns share **one** three-letter namespace, so a new prefix must not collide with any existing value in *either* column. The leading `C` on candidate prefixes is convention only and is not validated.

`SRN` is used for source-subjects so they do not collide in speech with Source catalog refs (`SRC-…`); `CSR` is its candidate counterpart.

## 3.2 `properties`

```text
key                 value_type    notes
name                name
event_type          term          kind identity; terms §3.4
date                date          point-in-time (or best single date); locked on event
start_date          date          span start; locked on event — leave empty if only Date applies
end_date            date          span end; locked on event — leave empty if only Date applies
role                term          participation edge label; terms §3.5
relationship_type   term          relationship edge label; terms §3.6
person              subject       app target hint: person
event               subject       app target hint: event
place               subject       app target hint: place
participant         subject       app target hint: person
mentions            subject       app target hint: source
remark              text          free-text commentary about a source subject
toponym             text          place name as interpreted from a Source (not a personal NameValue)
```

Target Subject type hints are application-only (not SQL allow-lists). See the Interpretation doc.

`name_format` (Conclusion naming profiles) is not an Interpretation Property in the create-time seed. `integer` remains a valid value_type with no seed row yet.

Create-time Install seeds the full §3.2 matrix, including kind/edge Properties as `value_type = term` with `property_terms` (§3.4–3.6).

Event date Properties (`date`, `start_date`, `end_date`) are locked on `event`: Conclusion ordering and timelines may key into them; Subject fields must not unbind. **Coexistence:** use `date` for a single point (birth, death, marriage day); use `start_date` / `end_date` when the event spans time (residence, service, voyage). Instantaneous events leave start/end empty; spanned events may leave `date` empty when only the range is known.

Additional Properties may be seeded as workflows need them (shared DNA, predicted relationship, and similar). Treat those as **TBD** until a concrete UI requires them.

## 3.3 `subject_type_fields`

```text
subject_type    property
person          name

event           event_type          # term
event           date                # locked
event           start_date          # locked
event           end_date            # locked

place           toponym

participation   person
participation   event
participation   role                # term

location        event
location        place

relationship    participant
relationship    relationship_type   # term

source          mentions
source          remark
```

Create-time Install binds the full matrix above (including kind/edge term Properties).

## 3.4 Property terms: `event_type`

Product-seeded **term keys** (horizon list; Install grows with use). Researchers may add further `origin=user` terms; do not treat these as free-text Observation strings.

```text
birth
death
marriage
baptism
burial
census
residence
migration
```

Exact GEDCOM alignment and additional vital/event kinds are **TBD**. Researchers add long-tail kinds as `origin=user` terms — there is no product `other` escape hatch. First-class facets (e.g. birthday) will attach to recognized keys such as `birth` in the subject registry when that behavior lands — not part of the create-time term seed itself.

## 3.5 Property terms: `role` (participation)

Product-seeded **term keys** (horizon list):

```text
subject
father
mother
spouse
child
witness
informant
```

Long-tail participation roles are `origin=user` terms — there is no product `other` escape hatch. Tree / connect behavior will attach to recognized keys (e.g. `father`, `mother`, `subject`) in the registry when that behavior lands — not part of the create-time term seed itself.

## 3.6 Property terms: `relationship_type`

Product-seeded **term keys** (starter set for connect macros; grow with use):

```text
spouse
sibling
parent_child
cousin
guardian
```

Prefer expanding the product set as connect macros need them. Long-tail labels are `origin=user` terms — there is no product `other` escape hatch. First-class connect behavior attaches to recognized keys in the subject registry when that behavior lands.

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
default_name_format_id = <uuid of western profile>
```

Additional cultural profiles (for example dual-surname ordering) are **TBD**.

---

# 5. Conclusion layer

Schema: [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

Conclusion seeds are mostly workflow vocabularies, not large type catalogs.

## 5.1 Sameness claim `status`

Closed workflow vocabulary (not researcher-extensible like user-minted Property terms on `event_type`):

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

There is no `origination_claims` table and no `representative_subject_id`. A `canonical_entities` row may have a nullable `identity_anchor_id` and optional `argument`. For subject-typed Reconciliation values, Conclusion uses `value_entity_id` (canonical), not Observation `value_subject_id`.

Handle provenencia badges (from records / inferred / asserted / unlinked) are computed, not seeded workflow enums.

## 5.5 Claim confidence grades

Open vocabulary for `sameness_claims.confidence_grade_id` and `reconciliation_claims.confidence_grade_id` (via `claim_confidence_grades`). Three-point scale with a baseline middle. **Do not** reuse Source credibility keys — same shape, different semantics ([`research-judgment-model.md`](research-judgment-model.md)).

```text
key                 sort_order    label (draft product copy)
low_confidence      1             Low confidence
moderate            2             Moderate
high_confidence     3             High confidence
```

`confidence_grade_id` is nullable. Claim `status` remains the separate workflow vocabulary in §5.1–5.2.

---

# 6. Documentation ownership

- This document is the horizon catalog for intended keys, starter open-vocabulary lists, project-default settings, and vocabulary **origin** namespaces (§1.1). It does not require every listed row to ship in the first app version.
- Layer and shared-value documents remain authoritative for table schemas and invariants.
- [`research-judgment-model.md`](research-judgment-model.md) is authoritative for judgment semantics (Source credibility, Citation transcription certainty, Claim confidence).
- When a key is actually seeded in product, update this catalog if needed; do not fork competing lists into layer docs.
