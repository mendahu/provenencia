# Provenance Genealogy — Data Model Philosophy

## Status

Draft architecture notes. This document describes the relationship between Provenance's three principal research layers.

Authoritative layer schemas:

- Source: [`source-layer-data-model.md`](source-layer-data-model.md)
- Interpretation: [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md)
- Conclusion: [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md)

Shared value models:

- Dates: [`structured-date-model.md`](structured-date-model.md)
- Names: [`structured-name-model.md`](structured-name-model.md)

Seed catalog for new projects: [`seeded-vocabulary.md`](seeded-vocabulary.md).

Cross-cutting audit/revision: [`audit-revision-history.md`](audit-revision-history.md).

---

# 1. The three layers

Provenance uses an evidence-first model with three principal layers:

```text
Source
  ↓
Interpretation
  ↓
Conclusion
```

The boundaries are intentional. Evidence should be preserved without forcing interpretation, and interpretation should remain revisable without rewriting the evidence from which it was derived.

## 1.0 Rigorous science and a casual tree

The data model is built so **good science is convenient**: Citations, Observation-sized assertions, sameness with exhibit, Reconciliation with a stored value and optional pins, provenance badges on working subjects. Use-case write-ups in these docs often describe that path — how the author wants to do genealogy.

The same model must **allow a family tree with little or no evidence**. Empty canonical Persons, asserted Places, names typed only as Reconciliation, no Sources: that is a valid project. It will not withstand genealogical proof review. That is acceptable. The application should encourage and surface rigor (warnings, badges, “unlinked” filters, suggested Citations), not refuse to save a tree because the user skipped science.

Layer boundaries still matter when evidence *is* entered: do not rewrite a Source to match a later conclusion. They do not require every handle to have a Source.

## 1.1 Source

The Source layer answers:

> What evidence do we possess?

It contains the evidence as acquired, descriptive catalog metadata, concrete Artifacts, and locally managed Files.

The Source layer deliberately does not identify historical people, normalize historical assertions, resolve places, or decide what facts are true.

Its complete schema and storage rules are maintained in [`source-layer-data-model.md`](source-layer-data-model.md).

## 1.2 Interpretation

The Interpretation layer answers:

> What does this particular source appear to say?

The conceptual progression is:

```text
Artifact
  ↓ researcher selects a meaningful portion
Citation
  ↓ supports
Observation
  ↓ asserts a Property about
Node
```

A Citation identifies an addressable portion of an Artifact and may preserve the researcher's transcription and description of that evidence.

A Node is a source-local, globally addressable thing encountered while interpreting evidence. Nodes deliberately carry very little domain structure themselves. Their semantics come from their Node Type and the Observations that describe and connect them.

Each Node records a home `source_id` for the Source in which it was primarily encountered. That is a UI and organization aid, not a confinement rule: Observations from Citations under other Sources may still target the Node.

An Observation is the smallest independently addressable unit of interpreted evidence. It is an atomic, cited assertion with the general shape:

```text
subject Node -- Property --> typed value
```

The assertion may be positive or negative. A negative Observation records that the source appears to deny the proposition (for example, that a person's name is not Jake), rather than merely omitting a positive assertion.

The value may be a scalar/structured value or another Node. A Node-valued Observation therefore forms an edge in the Interpretation graph.

The Interpretation layer is an extensible, schema-described property graph. It stores explicit normalized assertions derived from evidence but does not persist relationships or conclusions that can merely be inferred from those assertions. Genealogical inference and higher-order semantics belong to application and Conclusion logic.

A Node's assertion provenance is derived through its Observations:

```text
Node
  ← Observation
  ← Citation
  ← Artifact
  ← Source
```

There is no persisted `source_stack` or specialized Record hierarchy. Disconnected semantic matrices arise naturally from the resulting Node/Observation graph.

Authoritative schema: [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md).

## 1.3 Conclusion

The Conclusion layer answers:

> What does the researcher currently conclude about the historical world after considering one or more sources?

This layer contains canonical research entities such as:

```text
Person
Event
Place
Relationship
Participation
Location
```

These are rows in one `canonical_entities` table distinguished by `kind`, not parallel per-kind tables.

Canonical does not mean complete, final, or universally authoritative. A Person may be unnamed. A Place may be an asserted handle with no Interpretation Node. Two canonical entities may later be merged.

Interpretation keeps source-local Nodes separate for traceability: two Sources that mention the same historical person normally produce two `person` Nodes. The Conclusion layer gives the researcher a durable **working subject** (the canonical row). Creating that row is origination of the handle, not a third claim type.

The two Conclusion **claims** are **sameness** (Node co-reference) and **reconciliation** (committed Property values on a handle):

```text
Node A ── sameness claim (same_as / distinct_from) ── Node B

canonical_entities E (kind = person)
  identity_anchor_id → optional Node (live same_as cluster when set)
  members(E) = ∅, or Nodes reachable from that anchor via accepted same_as
  argument → optional existence rationale on the handle
```

Sameness Claims are higher-order than Observations. An Observation says what a particular Source appears to assert about a Node. A Sameness Claim says the researcher concludes that two Nodes do or do not co-refer, with its own evidence chain. Exhibit pins do not retarget Observations.

The same handle pattern applies across Node Types (`person`, `event`, `place`, `relationship`, `participation`, `location`). Domain payload is projected from member-Node Observations and overridden by Reconciliation Claims. For `value_type = 'node'`, Reconciliation points at another **canonical** entity (`value_entity_id`), so a nodeless Place can still be an association end.

**Convention:** Places work best as thin features at one grain; Locations as M:N event–place membership; gazetteer packs as ordinary Sources; extra grains as additional Location handles rather than `same_as` between town and colony. The schema does not enforce grain. See [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §1 and §8.

Authoritative schema: [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

---

# 2. Shared persistence conventions

Persistent rows use globally unique machine identifiers, currently UUIDv7 stored as 16-byte SQLite `BLOB` values.

Ordinary schema tables use SQLite `STRICT` typing.

Structured genealogical dates use the shared model in [`structured-date-model.md`](structured-date-model.md).

Structured personal names use the shared model in [`structured-name-model.md`](structured-name-model.md).

Selected user-facing rows receive a required short human-readable `ref`, unique within the project (application: unique across all ref-bearing tables).

Catalog and Interpretation rows that are not typed Nodes use `{PREFIX}-{token}`:

```text
SRC   sources          e.g. SRC-F4N2P
ART   artifacts        e.g. ART-3K9M2
CIT   citations        e.g. CIT-3K9M2
OBS   observations     e.g. OBS-2F8Q1
```

Nodes and canonical entities share a **type prefix** from `node_types.ref_prefix`. The concluded working subject uses the short form; Interpretation Nodes are marked as **candidates** so they do not read as “final”:

```text
PER-7KD45      Person (canonical entity)
PER-C-7KD45    person candidate Node (Interpretation)
```

```text
canonical    {type_prefix}-{token}
node         {type_prefix}-C-{token}
```

`C` means candidate (source-local, not the concluded Person/Event/…). It is not “canonical.” **Hypothetical (`H`) is not used:** a census person Node is cited evidence, not a guess.

`C` is a fixed application layer code, not a column and not user-extensible. Do not use `C` as a `ref_prefix`. Catalog prefixes `SRC`, `ART`, `CIT`, `OBS` stay reserved.

Creating a Node Type includes a `ref_prefix` that is unique and not in that reserved set. The Node Type / canonical `kind` is immutable; a mistaken type is a new row, not an in-place change.

Machine identity remains the UUID. `ref` is assigned by the application on insert, stable, and not recycled. After canonical merge, old refs keep resolving to the surviving entity.


Generic persistence bookkeeping such as `created_at`, `updated_at`, and user attribution does not belong on core domain tables. Creation, modification, deletion, user attribution, and revision ordering are recorded by the append-only audit/revision model in [`audit-revision-history.md`](audit-revision-history.md).

---

# 3. Cross-layer examples

See also the worked examples in [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md#12-cross-layer-examples).

## 3.1 Photograph and testimony

```text
Photograph Source
  Artifact: scan.jpg
  Citation: crop around one person
  Observation: cited crop depicts person Node N1
  Node N1 (person, home Source = photograph)

Testimony Source
  Artifact: audio or research note
  Citation: "That's my grandfather"
  Observation: testimony refers to the same depicted person as Node N1
  Node N2 (person, home Source = testimony)

Conclusion
  canonical_entities E1 (kind=person, identity_anchor_id = N1)
  sameness_claim: N1 same_as N2 (accepted), with evidence Observations
  members(E1) = {N1, N2}
```

## 3.2 Conflicting certificate and letter

```text
Birth certificate Source C
  person Node NC, birth_date Observation → 1 JAN 1800
  source Node SC reifying C

Letter Source L
  Observations:
    SC -- remark --> "date of birth on certificate mistyped"
    person Node NL -- birth_date --> 2 JAN 1800
    person Node NL -- birth_date --> 1 JAN 1800 (polarity negative)

Conclusion
  sameness_claim: NC same_as NL (accepted), evidence cites both Sources' Observations
  canonical_entities E (kind=person, identity_anchor → NC or NL); members = {NC, NL}
  reconciliation_claim on E, property birth_date:
    value → DateValue(2 JAN 1800)   # researcher choice, or synthesized
    evidence → the birth_date Observations (and related letter Observations as needed)
    argument → why the letter's correction is preferred
```

## 3.3 DNA evidence

```text
Source: DNA match report
Artifact: locally retained export/screenshot/JSON/CSV
Citation: match result
Observations on person Node ND:
  shared DNA
  predicted relationship
  match display name

Conclusion
  sameness_claims may later correlate ND with other person Nodes
  canonical person entity created/extended via identity_anchor + accepted same_as closure
```

---

# 4. Documentation ownership

To avoid competing schema definitions:

- [`source-layer-data-model.md`](source-layer-data-model.md) is authoritative for Source-layer tables and Artifact/File storage.
- [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md) is authoritative for Interpretation-layer tables and vocabulary.
- [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) is authoritative for Conclusion-layer tables and Claims.
- [`structured-date-model.md`](structured-date-model.md) is authoritative for shared DateValue persistence.
- [`structured-name-model.md`](structured-name-model.md) is authoritative for shared NameValue persistence.
- [`seeded-vocabulary.md`](seeded-vocabulary.md) is the horizon catalog for intended keys and starter open-vocabulary lists (not a v1 ship list).
- [`audit-revision-history.md`](audit-revision-history.md) is authoritative for audit and revision history.
- This document summarizes three-layer philosophy and cross-layer examples; it does not own layer schemas.
