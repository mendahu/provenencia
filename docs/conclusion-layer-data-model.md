# Provenencia Genealogy — Conclusion Layer Data Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for the Provenencia Conclusion layer.

The Conclusion layer answers:

> What does the researcher currently conclude about the historical world after considering one or more sources?

Cross-layer philosophy is summarized in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md). Interpretation subjects (Subjects, Observations, Properties) are defined in [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md). Shared date and name value models are [`structured-date-model.md`](structured-date-model.md) and [`structured-name-model.md`](structured-name-model.md). Audit history is [`audit-revision-history.md`](audit-revision-history.md). Researcher judgment (Claim confidence, and how it relates to Source credibility and Citation certainty) is [`research-judgment-model.md`](research-judgment-model.md).

This draft treats canonical rows as the **working set** (handles the researcher is investigating). The Conclusion **claims** are still only two kinds: Identity and Reconciliation. Creating a handle is not a third claim type; the entity row is the origination. An Identity Claim says one Interpretation subject is a reading of one handle. Accepted claims are that handle's members.

---

# 1. Invariants versus conventions

The graph is built to **accept** research the schema cannot judge. A toponym `York, Upper Canada, North America` on one Place, an Identity Claim that puts a town subject and a colony subject on the same Place, a Location with only one end filled, or an Observation that stretches what a Citation supports: these must **persist**. The evidence trail (Citations, Observation subjects, claim pins, `argument`) is how a later reader evaluates them. Product UI may warn, badge, or suggest splits; it must not refuse to save because a convention was skipped.

This document therefore uses three strengths of “must”:

| Strength | What it is | If the researcher ignores it |
|---|---|---|
| **Schema invariant** | SQLite `NOT NULL`, FKs, `UNIQUE`, `CHECK` | Insert/update fails |
| **Application invariant** | Writer/resolver rules that keep the generic graph typed (value column matches `value_type`, an accepted Identity Claim is the only membership for that subject, merge re-points claims onto the survivor) | Treat as malformed data (repair, ignore extras, or show an error on *that row*) — not as “you used Places wrong” |
| **Convention** | How to get the most from search, maps, merge, and honest reading of Sources | Data remains valid; features may be weaker or the proof harder to trust |

Seeded vocabulary (`location` + `event`/`place`, one-grain Places, gazetteer-as-Source) is convention plus picker defaults, not a closed ontology. Researchers may add Properties and Subject types; first-class UI may only light up for keys it knows.

Use-case text in this document often describes **rigorous, evidence-backed** genealogy: that is the path the product should make easy. A user who only wants a family tree — unlinked Persons, typed names, no Citations — must still be able to save. The result need not be scientifically defensible. Encourage and badge; do not block. Broader product stance: [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) §1.0.

---

# 2. Conclusion-layer principles

## 2.1 Working subjects may be sparse or ungrounded

A researcher may add a working Person, Event, Place, or association before any Interpretation subject exists for it.

```text
Person PER-7KD45
  label = "Mother of James"
  members = ∅
  argument = "James existed; he had a mother"
```

This is valid. It is a **workspace handle**, not an Observation. Genealogical names, when known, come from cited NameValue Observations on member Subjects and/or name Reconciliation Claims. The canonical `label` is only a researcher working identifier.

**Schema:** Observations still require a Citation and still target a subject — you cannot make an Observation whose subject is a canonical entity. **Convention:** hang inferred geography on a knowledge or gazetteer Source (or a Reconciliation exhibit), not on a Citation that did not say it. The database will store the latter; the chain will look like the Source asserted it.

## 2.2 Membership versus committed properties

**Membership** (which Interpretation subjects are this subject) is the set of **accepted Identity Claims** that point at the handle. Each claim is one subject joining that handle, with its own exhibit. A new source subject joins by a new claim; it does not inherit membership through some other subject.

**Committed values** (name, date, location ends, …) are accepted Reconciliation Claims. If there is no claim, the UI may **project** values from Observations on member Subjects. That projection is not stored.

```text
with accepted identity claims    members = those subjects
                                 display values = accepted Reconciliation, else member Observations
with no accepted identity claims members = ∅
                                 display values = accepted Reconciliation only
```

## 2.3 Identity and merge are related but distinct

An **Identity Claim** places one Interpretation subject on one canonical entity:

```text
Subject A ── identity claim ──► canonical entity E
```

**Canonical merge** joins two Conclusion handles that were previously treated as distinct working subjects:

```text
PER-A merged_into PER-B
PLC-A merged_into PLC-B
```

Identifying a subject with an entity does not merge handles. Merge is an explicit decision that two handles are the same historical thing. The application then sets `merged_into_id` and re-points Identity Claims and Reconciliation Claims onto the survivor. If that subject already has a claim row for the survivor, the application updates that row instead of inserting a second one; an accepted claim wins over a provisional or rejected row for the same pair. Historical relationships or succession are not identity or merge.

**Convention:** an Identity Claim is most useful between a Place and place subjects of the same grain (two readings of York the town). The schema will accept claims that put a composite “York, Upper Canada” subject and a colony subject on the same Place; maps will then treat them as one feature. Extra grains of “where this event happened” are usually additional Location handles rather than place identity — again a convention, not a CHECK.

## 2.4 Negation and conflict at Conclusion

Subject identity uses Identity Claims. Absence of a claim is not a rejection. `status = rejected` on the claim for this subject and this entity means the researcher considered the subject for that handle and concluded it does not belong. A subject may be rejected for one entity and accepted for another.

Attribute-level conflicts across member Observations are handled by soft display merges and, when durable, by Reconciliation Claims. That is separate from Observation polarity and from subject identity.

## 2.5 Resolver logic is application-level

The database preserves multiple source-backed values. Resolvers may synthesize display without creating Claims. **Provenencia badges** on a handle (from records / inferred / asserted / unlinked) are computed from whether any accepted Identity Claim exists, whether Reconciliations pin Observations, and whether `argument` is set. They are not a stored enum.

## 2.6 Persistence conventions

Persistent rows use globally unique machine identifiers, currently UUIDv7 stored as 16-byte SQLite `BLOB` values. Ordinary schema tables use SQLite `STRICT` typing.

Structured genealogical dates use [`structured-date-model.md`](structured-date-model.md). Structured personal names use [`structured-name-model.md`](structured-name-model.md).

Generic change metadata belongs to [`audit-revision-history.md`](audit-revision-history.md).

Selected user-facing entities receive a required short human-readable `ref`. Canonical entities use `{ref_prefix}-{token}` (no candidate mark). Shared rules are in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md).

---

# 3. Design summary

Interpretation is a cited property graph (Subjects, Observations, Citations). Conclusion adds working handles and two claim verbs. The family tree, timeline, and map pins are **views** over that graph — **projections** in the proposed Narrative layer ([`narrative-layer-data-model.md`](narrative-layer-data-model.md) §4–§6). Conclusion owns the relationships and committed values those views read; saved layouts, prose, and curated map compositions would live downstream in Narrative when that layer ships.

```text
Observations ──subject──► subject
                              │
                              │ identity claim (accepted = member)
                              ▼
                       canonical_entities
                              │
                              ├── reconciliation_claims (committed Property values)
                              └── exhibit pins live on claims, not on the handle
```

```text
canonical_entities
  - kind (person | event | place | relationship | participation | location | …)
  - stable UUID and human ref
  - argument (optional existence rationale)
  - label (researcher working identifier)
  - members = subjects with an accepted identity claim for this entity
  - payload = member-subject Observations projected, overridden by reconciliation_claims
```

Important separations:

1. **Canonical entity** — working subject; creating it *is* origination. No `origination_claims` table.
2. **Identity Claim** — one Subject is a reading of one entity, with Observation exhibit. Accepting the claim makes the subject a member.
3. **Membership** — the accepted Identity Claims for that entity. Not a separate join table, and not a graph closure over other subjects.
4. **Reconciliation Claim** — concluded value of one Property on one entity. Exhibit Observations may be about **other** Subjects; they are not retargeted.
5. **Canonical merge** — `merged_into_id` within the same Subject type. Explicit; identifying a subject does not merge two handles.
6. **Independent handles** — creating a Person does not auto-create related Events, Places, or Locations. The promote UI may offer those neighbors afterward (§5.4). The schema still creates nothing until the researcher accepts each one.

SQLite enforces foreign keys, one claim per (subject, entity), one accepted claim per subject, and same Subject type on both ends. Merge re-pointing and “a Location is usable only when both ends are known” are **application invariants**.

---

# 4. `identity_claims`

An Identity Claim asserts that one Interpretation subject is a reading of one canonical entity — this census line, this gazetteer feature, this recorded event is that working person, place, or event. The name is the judgment. Accepting the claim makes the subject a member of that handle. The subject stays in Interpretation; Observations are not retargeted onto the entity.

This replaces pairwise sameness (`same_as` / `distinct_from` between subjects) and the identity anchor. Two subjects are the same historical thing when each has an accepted Identity Claim for the same entity. There is no edge between subjects and no transitive closure. A subject that has not been claimed yet is simply not a member of any handle; correlating two subjects before a handle exists means creating the handle (or choosing one) and claiming each subject.

Claim confidence grades (shared with Reconciliation Claims):

```sql
CREATE TABLE claim_confidence_grades (
    id          BLOB PRIMARY KEY,              -- UUIDv7, 16 bytes
    key         TEXT NOT NULL,
    origin      TEXT NOT NULL,                 -- provenencia | user | plugin:<id>
    label       TEXT NOT NULL,
    sort_order  INTEGER NOT NULL,

    UNIQUE (key, origin)
) STRICT;
```

Seed keys: [`seeded-vocabulary.md`](seeded-vocabulary.md) §5.5. Semantics: [`research-judgment-model.md`](research-judgment-model.md). Vocabulary origin: [`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1.

```sql
CREATE TABLE identity_claims (
    id                   BLOB PRIMARY KEY,
    subject_id           BLOB NOT NULL,
    entity_id            BLOB NOT NULL,
    subject_type_id      BLOB NOT NULL,
    status               TEXT NOT NULL,
    confidence_grade_id  BLOB REFERENCES claim_confidence_grades(id),
    argument             TEXT,

    CHECK (status IN ('provisional', 'accepted', 'rejected')),
    FOREIGN KEY (subject_id, subject_type_id)
        REFERENCES subjects (id, subject_type_id)
        ON DELETE CASCADE,
    FOREIGN KEY (entity_id, subject_type_id)
        REFERENCES canonical_entities (id, subject_type_id)
        ON DELETE CASCADE,
    UNIQUE (subject_id, entity_id)
) STRICT;

CREATE UNIQUE INDEX identity_claims_one_accepted_per_subject
    ON identity_claims (subject_id)
    WHERE status = 'accepted';
```

There is at most one Identity Claim per `(subject, entity)`. Changing `status`, confidence, or `argument` updates that row (and is audited), rather than inserting a second Claim for the same pair.

A subject has at most one **accepted** Identity Claim, so it belongs to at most one handle. Provisional claims against other entities may stand beside that accepted row: the subject is a member of one handle and still a candidate for others. Two accepted claims for one subject cannot exist.

`status` is workflow state: `provisional`, `accepted`, or `rejected`. Only **accepted** claims are members. See [`seeded-vocabulary.md`](seeded-vocabulary.md).

`provisional` is a real persisted conclusion-in-progress, not a missing row. The UI should surface provisional Claims differently from accepted ones (for example a distinct stroke, badge, or filter) so researchers can see candidates without treating them as membership. Rejected Claims remain for audit. A rejection means “not this entity,” which is scoped to that handle: the same subject can later be accepted for a different one. Absence of a row is not a rejection.

`confidence_grade_id` is optional epistemic stance (three-point Claim confidence vocabulary). It is **orthogonal to `status`**: accepted + low confidence and provisional + high confidence are both valid. Do not overload `provisional` to mean “I am unsure.” Authoritative rules: [`research-judgment-model.md`](research-judgment-model.md). Grade keys: [`seeded-vocabulary.md`](seeded-vocabulary.md).

`argument` holds the researcher's reasoning chain for why this subject is (or is not) this entity — name and age against the members already on the handle, a testimony that identifies a photograph, or a short note when a single Observation already makes the case. Reasoning stays here rather than scattered across evidence rows. Comparisons to another subject are written in `argument` and supported by pinning that subject's Observations; they are not a separate edge. Confirmed matches and exhibit backfill are §5.1.

The subject and the entity must share one Subject type. `subject_type_id` is copied onto the claim so SQLite can enforce that with composite foreign keys. `subjects` already has `UNIQUE (id, subject_type_id)`; `canonical_entities` carries the same unique pair (§6). That copy cannot drift: subject type is immutable on both rows.

---

# 5. `identity_claim_evidence`

Evidence rows are pointers to Observations that participate in the claim's exhibit list. An individual Observation does not carry a stance toward the Identity Claim; the claim's `status` and `argument` express the researcher's conclusion over the whole set.

```sql
CREATE TABLE identity_claim_evidence (
    identity_claim_id   BLOB NOT NULL REFERENCES identity_claims(id) ON DELETE CASCADE,
    observation_id      BLOB NOT NULL REFERENCES observations(id),

    PRIMARY KEY (identity_claim_id, observation_id)
) STRICT;
```

Pin every relevant Observation here; write the conclusion and inference in `identity_claims.argument`. Exhibit pins are **Observations only** — not Citations, Sources, or other Claims. A confirmed match pins Observations that already exist on the two subjects (§5.1). Broader exhibit types can wait until a concrete workflow needs them.

The Observation's `subject_id` is unchanged. Pins mean “this assertion is part of my proof,” not “this Observation is now about the claim.” Creating a match does not insert an Observation, a Citation, or a subject. The source did not state the match; the Identity Claim does.

## 5.1 Confirmed matches and backfill

One confirmed comparison is enough to accept an Identity Claim. The birth certificate can ground the person with no comparison at all. The marriage claim can pin only the birth certificate. The census claim can pin only the birth certificate and leave the marriage record uncited. Each claim justifies one arrival. Earlier proofs stay on earlier claims.

A confirmation records that two existing Observations agree. Both are pinned. A refusal leaves that pair unpinned. It does not create a negative Observation. A conflict the researcher wants to remember goes in `argument`, or stays visible because both Observations sit on members of the handle. Choosing a concluded value is a Reconciliation Claim.

When the incoming subject is confirmed against a subject that already has an **accepted** Identity Claim on the same entity, the application writes those Observation pins onto **both** claims. The existing member's exhibit grows. Its `argument` stays as written. The added pins are audited, so a later reader can see that a subsequent join contributed them.

That backfill is what keeps a proof standing when one member leaves. If the census claim pins the birth record and the marriage record, and the marriage claim also receives the census pins, removing the birth certificate still leaves each of those claims with a live comparison. Backfill onto the new claim alone does not repair an older claim that cited only the departing subject.

Pins are the machine-readable comparison. An `argument` that only names the other record in prose does not.

## 5.2 When a comparison subject leaves

Rejecting, deleting, or moving an accepted Identity Claim does not change any other subject's membership. Other Identity Claims on that entity whose exhibit pins an Observation of the departing subject are surfaced for review. The application does not auto-reject them and does not strip their remaining pins.

A claim that still pins another remaining member has a live comparison. A claim whose only comparison pins were the departing subject has an exhibit that no longer explains membership. Both are shown. Neither is evicted. The researcher re-pins against a member who is still there, accepts the claim again with no exhibit, or rejects it.

The same review applies when a pinned Observation is deleted or its value changes materially.

## 5.3 Promote comparison (future UI)

The screen below is the intended promote flow for a later spike. The schema does not require it. Accepting a claim with an empty exhibit stays valid, and that claim may show as undocumented. The product offers the comparison; it does not block the accept.

1. The researcher has a subject, its Observations, and an evidence graph.
2. They choose Promote and a canonical entity of the same Subject type.
3. If that entity has no accepted members yet, this claim grounds the handle. There is no comparison step.
4. Otherwise the UI lines the incoming subject's Observations up against the same Property on each accepted member. Compatible pairs start checked. Differing values start unchecked. A routine cross-property comparison (a census age against a birth date) may be suggested; confirming it still only pins the two existing Observations. Where a member has several Observations for one Property, each is its own row.
5. The researcher can accept the checked pairs for a Property in one gesture, clear a pair, or skip the comparison and accept with no pins. One confirmed member is enough. Further checked members are the resilient exhibit from §5.1.
6. Accept writes the incoming Identity Claim and the pins from §5.1, including backfill onto the other accepted members' claims. The UI may draft the new claim's `argument` from the confirmed rows; the researcher can edit that draft. Older claims' arguments are not rewritten.

A refusal is not stored as its own row. Opening the comparison again shows an unpinned pair as not yet cited.

The review in §5.2 is the other half of this UI: when a member leaves or a pinned Observation changes, list the affected Identity Claims on that entity and say whether each one still cites a remaining member.

## 5.4 Neighborhood walk (future UI)

Promoting one subject does not promote the subjects it is connected to. After the Identity Claim in §5.3 is accepted, the UI may continue with the unpromoted interpretation neighborhood of **the subject just filed**. That walk is a checklist the researcher can leave at any time. Whatever is left stays on the evidence graph, unconcluded.

The queue appends. It does not walk the canonical entity the subject was filed onto. That entity's existing relationships are a source of **suggested targets** only.

For each queued subject:

1. Offer canonical entities of the same Subject type that already relate to the handle just filed, plus others that resemble this subject (type, date, toponym, and similar text). The researcher picks one, or creates a new handle.
2. Run the property comparison from §5.3 against the chosen entity's accepted members, including backfill. A new handle with no members is grounding: there is no comparison step.
3. Append this subject's interpretation neighbors that are not already queued and not already handled in this pass. Do not append the other members of the canonical entity, or the neighbors of those members.

A subject handled in this pass is skipped, so a wife's relationships do not promote the husband again. Skipping a subject does not reach through it. Declining the birth event does not queue that event's places.

Bridge subjects (`participation`, `location`, `relationship`) are queue items, and they wait until both ends are handles. The step is then a short confirm: connect these two, or file this bridge subject onto the participation or location the canonical ends already share. The role or relationship type stays an Observation on the bridge. It is not its own canonical entity.

Worked shape, starting from a person on a birth record:

```text
accept person  → queue the birth event and its participation
accept event   → property match if the event already exists;
                 else ground a new event;
                 queue that event subject's places and location bridges
accept a place → suggest the place already linked to that event, if any;
                 do not queue that place's other events
both ends known → confirm the participation and each location
```

Three places are three checklist rows. Each can join a different existing place or start a new one.

---

# 6. `canonical_entities`

Because canonical rows are thin handles, they share one table instead of parallel `persons` / `events` / … tables.

```sql
CREATE TABLE canonical_entities (
    id                      BLOB PRIMARY KEY,
    subject_type_id         BLOB NOT NULL REFERENCES subject_types(id),
    ref                     TEXT UNIQUE NOT NULL,
    argument                TEXT,
    label                   TEXT,
    merged_into_id          BLOB REFERENCES canonical_entities(id),

    UNIQUE (id, subject_type_id)
) STRICT;
```

`subject_type_id` is the same vocabulary as Interpretation Subject types (referenced by id so origin-namespaced keys stay unambiguous). The table is an implementation detail; in the UI a `person` row is a Person (`PER-7KD45`), not a “canonical entity.” The Subject type is immutable after insert.

`ref` is required: `{ref_prefix}-{token}` from the Subject type — `PER-7KD45` for a person. Candidate person Subjects are minted off the same type's `candidate_ref_prefix` (`CPR-…`), so the two layers stay distinct by prefix while sharing one ref format.

There is no identity anchor and no `representative_subject_id`. Membership is not a column on this row. Creating a handle from a record inserts the entity and an accepted Identity Claim for that subject. **Adopt** (first grounding of an inferred handle) is the first accepted Identity Claim on a handle that had none. Dropping the last member does not delete the entity or change its id.

```text
members(entity) =
  subjects with an accepted identity_claims row whose entity_id is this entity
  (after merge, readers follow merged_into_id; the application re-points claims
   onto the survivor so membership is queried there)
```

`UNIQUE (id, subject_type_id)` exists so Identity Claims can foreign-key `(entity_id, subject_type_id)` and SQLite can reject a person subject claimed onto a place. `id` is already the primary key; the pair is the parent key for that composite reference, same as on `subjects`.

`argument` is optional existence rationale on the handle (“working subject for Canada; not taken from a citation in this project”). Accumulating commentary stays in `canonical_entity_notes`. Do not treat `argument` as a genealogical Property.

Rules:

- `merged_into_id`, when set, should reference another entity of the same Subject type.
- `label` is an optional researcher working identifier (for example `Mother of James`). It is not a genealogical name and must not substitute for NameValue Observations or name Reconciliation Claims.
- `ref` is the stable short public reference, assigned on insert as `{ref_prefix}-{token}`. After merge, old refs keep resolving to the surviving entity.
- Members are accepted Identity Claims. Do not list member Subjects on the entity row.
- Creating one handle does not auto-create related kinds. Reification `source` subjects are not typically given canonical rows.
- Primary kinds (`person`, `event`, `place`) may be created with no Identity Claims, no `argument`, and no Reconciliations (stubs). Association kinds (`location`, `participation`, `relationship`) are edges: **conventionally** they are not shown as complete until endpoint Properties are known (see §7.3). Incomplete Location handles still persist.

```sql
CREATE TABLE canonical_entity_notes (
    id                      BLOB PRIMARY KEY,
    canonical_entity_id     BLOB NOT NULL
        REFERENCES canonical_entities(id) ON DELETE CASCADE,
    body                    TEXT NOT NULL
) STRICT;
```

One notes table covers every kind because there is a single parent table and a real foreign key.

## 6.1 Handle provenencia (UI, not a column)

Compute a badge from the row and its claims. Suggested labels:

| Badge | Typical data |
|---|---|
| **From records** | at least one accepted Identity Claim (membership may still be one subject) |
| **Inferred** | no accepted Identity Claim; at least one accepted Reconciliation with Observation pins |
| **Asserted** | no accepted Identity Claim; `argument` set; no pinned Reconciliations |
| **Unlinked** | no accepted Identity Claim, empty `argument`, no accepted Reconciliations |

Unlinked stubs should stay off graphs and maps; list them in an unlinked / stubs view.

A handle with members is a better **trace to Interpretation**, not automatically better historical geography than a well-pinned inferred Location.

---

# 7. `reconciliation_claims`

A Reconciliation Claim asserts:

```text
for canonical entity E, Property P has concluded value V
```

The value is stored **on the claim**. Observation pins are optional exhibit. Pins may point at Observations whose subjects are other Subjects (sibling births, a gazetteer line). Those Observations are not retargeted and are not copied onto a fictional Location subject.

```sql
CREATE TABLE reconciliation_claims (
    id                   BLOB PRIMARY KEY,
    entity_id            BLOB NOT NULL REFERENCES canonical_entities(id) ON DELETE CASCADE,
    property_id          BLOB NOT NULL REFERENCES properties(id),
    status               TEXT NOT NULL,
    confidence_grade_id  BLOB REFERENCES claim_confidence_grades(id),
    argument             TEXT,

    value_text      TEXT,
    value_integer   INTEGER,
    value_real      REAL,
    value_boolean   INTEGER,
    value_date_id   BLOB REFERENCES date_values(id),
    value_name_id   BLOB REFERENCES name_values(id),
    value_entity_id BLOB REFERENCES canonical_entities(id),

    CHECK (status IN ('provisional', 'accepted', 'rejected')),
    CHECK (value_boolean IS NULL OR value_boolean IN (0, 1)),
    UNIQUE (entity_id, property_id)
) STRICT;
```

`entity_id` is a real foreign key to `canonical_entities`. The entity's Subject type is available via that join; it is not duplicated on the claim row.

`property_id` is the extensibility hook. Any Property in the vocabulary may be concluded on a suitable entity (subject to application rules about which Properties make sense for that Subject type).

Exactly one value representation must be populated, and it must match `properties.value_type`, parallel to Observations. For `value_type = 'subject'`, Conclusion stores **`value_entity_id`** (another canonical handle). Interpretation Observations still use `value_subject_id` (Subject to subject). That split is intentional: a nodeless Place (asserted Canada) has no subject to point at. Matching Subject type to the Property's target hint is an application invariant.

As with Observations, typed-column matching is an **application write invariant** for now. Readers prefer the column matching `value_type` if extras are present. The concluded value may match one exhibit Observation, be copied from a related subject's value, or be synthesized (for example a DateValue spanning Apr–May 1985).

There is at most one Reconciliation Claim per `(entity, property)`. Changing the concluded value, `status`, or confidence grade updates that row (and is audited).

`status` matches Identity Claims: `provisional`, `accepted`, or `rejected`. **Only `accepted` is the committed concluded value.** `provisional` is a persisted working choice the UI should show differently. `rejected` is kept for audit and is not the working value.

`confidence_grade_id` matches Identity Claims: optional three-point Claim confidence, orthogonal to `status`. See [`research-judgment-model.md`](research-judgment-model.md). Grades live in `claim_confidence_grades` (§4).

From a modeling standpoint a Property on an entity either has a Reconciliation Claim or it does not. Soft blends of member Observations are **stateless display**. They are never persisted as automatic claims. There is no `origin` column and no `association_endpoints` table: association ends are ordinary Properties (`event`, `place`, `person`, `participant`) resolved like any other.

`argument` holds researcher reasoning when needed.

## Name format as reconciliation

Cultural name display/entry ordering is not a column on `canonical_entities`. It is concluded like other Person-scoped facts:

```text
Property name_format -> text   # value is a name_format_profiles.key, e.g. "western"
```

- `project_settings.default_name_format_id` supplies the UI default when a person entity has no accepted `name_format` Reconciliation Claim.
- When the researcher commits a format for a Person (including as part of reconciling names across cultures), they persist a Reconciliation Claim for `name_format`.
- Evidence pins are optional for `name_format` (it is often a preference rather than source-derived); `argument` may still record why that profile was chosen.
- Concluded `name` values remain separate Reconciliation Claims (Property `name`, NameValue). Format and name content are related in the UI but distinct Properties.

See [`structured-name-model.md`](structured-name-model.md).

## Display versus persisted reconciliation

| Situation | Typical handling |
|---|---|
| Compatible or blendable member values | Stateless UI projection from Observations; no claim |
| Researcher commits a value (winner, synthesis, `name_format`, or nodeless ends) | Reconciliation Claim, usually `accepted` |
| Researcher is still weighing a value | Optional claim with `status = provisional` (distinct UI) |

Absence of an **accepted** Reconciliation Claim means “no committed concluded value yet.” The UI may still show member Observations and a stateless merge. For `name_format`, absence of an accepted claim means “use the project default.” For a nodeless Location, absence of accepted `event` or `place` means the edge is not usable yet.

## 7.1 `reconciliation_claim_evidence`

```sql
CREATE TABLE reconciliation_claim_evidence (
    reconciliation_claim_id BLOB NOT NULL
        REFERENCES reconciliation_claims(id) ON DELETE CASCADE,
    observation_id          BLOB NOT NULL REFERENCES observations(id),

    PRIMARY KEY (reconciliation_claim_id, observation_id)
) STRICT;
```

Evidence rows are pointers to Observations in the claim's exhibit list. Individual Observations do not carry a stance toward the claim; `argument` and the concluded value express the conclusion over the set.

## 7.2 Example (person dates)

```text
canonical_entities row E1
  kind = person
  identity claims: N1 and N2 accepted
  members: person Subjects N1, N2

N1 Observation: birth_date → DateValue(14 MAY 1985)
N2 Observation: birth_date → DateValue(MAY 1985)
  → soft display merge; no claim required

N1 Observation: birth_date → DateValue(MAY 1985)
N2 Observation: birth_date → DateValue(APR 1985)
  → stateless UI projection (e.g. a range); no claim
  → or researcher claim on E1 / birth_date if committing a value

N1 Observation: name → NameValue(form="James K. Robins", …)
N2 Observation: name → NameValue(form="James Robins", …)
  → often soft-mergeable; claim on E1 / name if committing a preferred NameValue
  → optional claim on E1 / name_format = "western" when committing display/entry convention
```

## 7.3 Association kinds (Location, Participation, Relationship)

On **Location subjects**, `event` and `place` are subject-valued Observations (Interpretation). A canonical Location **with members** projects those Observations and resolves each target subject to a canonical Event/Place when those handles exist.

A canonical Location **with no members** has no such Observations. Its ends are accepted Reconciliations:

```text
LOC-…  (kind = location, members = ∅)
  argument = optional rationale for the whole edge
  Reconciliation event → value_entity_id = EVT-…   (this birth)
  Reconciliation place → value_entity_id = PLC-…   (Upper Canada)
```

Pins typically sit on the claim they justify (geography Observations on `place`; identifying the birth on `event` if needed). **Convention:** do not invent a Location subject for a conclusion-only edge, and do not add Observations to a Citation that did not support that end. Both are storeable; they mislabel the Source.

The same pattern applies to Participation (`person`, `event`, `role`) and Relationship (`participant`, `relationship_type`): project member Observations when the handle has members; Reconciliation when it does not. `role` / `relationship_type` stay scalar Reconciliations (or member Observations) as today.

A Location is **M:N membership** of Events in Places: one Event may have several Location handles (York *and* Upper Canada); one Place many events. Two Locations on one Event do not imply gazetteer containment.

---

# 8. Places and Locations

This section is **convention** unless noted. The schema does not know a township from a comma-separated blob.

**Convention that pays off:** treat a Place as one geographic feature at **one grain** (this town, this township, this colony, this farm), and treat Location as M:N membership of Events in those Places. Genealogy then groups events and can later draw a map without a locality-history encyclopedia.

**What the graph still allows:** a single Place whose `toponym` is `York, Upper Canada, North America`; Identity Claims that put a town subject and a colony subject on the same Place; a Location with only `event` filled; zero Citations. Save succeeds. Search, gazetteer bind, and “events in Upper Canada” roll-up will be weaker or misleading. The exhibit (or lack of one) is how another researcher judges it.

## 8.1 Interpretation

A source phrase `"York, Upper Canada"` **can** be one place subject and one toponym string. **Convention:** split into two place Subjects (or two Location memberships) at the grains the Citation actually supports, so the York subject and the colony subject are not both claimed onto one Place when a second source only says Upper Canada. Toponyms on place Subjects use Property `toponym` in [`seeded-vocabulary.md`](seeded-vocabulary.md) (open: other Properties are fine). Event date is usually the time context rather than a date on the Place subject — convention, not a CHECK.

A Location subject is the usual source-local association: this **event subject** at this **place subject**, each end an Observation with a Citation. Other shapes remain valid graph data.

## 8.2 Canonical Places

Create `PLC-…` like any other handle:

- **From a record:** an accepted Identity Claim for a place subject (census line, gazetteer feature, …).
- **Asserted:** no Identity Claims; `label` = Canada; optional `argument`. Citing a baptism that never named Canada is allowed and will look like that baptism supported Canada.
- **Gazetteer as Source (convention for citable geography):** ingest an editioned dump as a Source/Artifact; interpret a feature into a place subject; accept an Identity Claim from that subject onto the canonical Place. Google Maps is a renderer. Live geocoders make poor frozen Citations; dated packs (Who's On First / GeoNames SQLite, Newberry shapefiles, dated OHM extracts) match Source+Artifact better. Wikidata Q-ids are concordances; a snapshot is what you cite if Wikidata is the Source.

Farms and unnamed lots may stay project-only with no gazetteer feature.

**Containment:** two Locations on one Event do not imply a parent polygon. Search may offer an explicit “include gazetteer parents” lens. A concluded `contained_in` Property is optional later. **Convention:** do not assume York ⇒ Upper Canada unless some exhibit says so.

## 8.3 Extra grain (inferred Location)

Sibling births say Upper Canada; this birth says only York. **Convention:** keep three honest Interpretation Locations, then add a memberless Location handle with Reconciliations (example below). **Also valid:** one composite Place, or Identity Claims that put different grains on one Place — the app should not throw.

```text
EVT-this-birth     identity claim accepted for this birth subject
PLC-york           identity claim for this register's place subject (settlement grain)
PLC-upper-canada   identity claims for sibling place Subjects, and/or an asserted Place
LOC-york           identity claim for this birth's Location subject (event + York)
LOC-uc             no identity claims
  Reconciliation event → EVT-this-birth
  Reconciliation place → PLC-upper-canada
  exhibit on place (and/or event) → the three location Observations
  argument → siblings usually born nearby, so this York is in that Upper Canada
```

## 8.4 Interpolation (convention)

An extra Observation on the original York Citation that asserts Upper Canada **will save**. It is a weak scientific move: readers cannot tell the register from the interpolation. A researcher-knowledge Source is a weaker but honest shortcut; a gazetteer Source is the usual geographic path. UI may warn; it should not block.

---

# 9. Canonical merge

When two canonical entities of the same `kind` should become one working subject:

```text
E-A label = "Unknown father"
E-B label = "William Smith"

Merge:
  E-A.merged_into_id = E-B
```

Typical trigger: the researcher decides two handles are one historical thing (two working fathers turn out to be one man), not a new Identity Claim. A claim only adds a subject to one handle. If that subject already belongs to the other handle, accepting it requires releasing the previous accepted claim first; the empty handle can then be merged or left as a stub.

Old ids and human refs should continue resolving to the surviving entity. Merge remains explicit and auditable. Identity Claims and Reconciliation Claims on the absorbed entity are re-pointed onto the survivor by application rules (§2.3). A subject that already has a row for the survivor updates that row; an accepted claim wins over a provisional or rejected one.

---

# 10. Schema and application invariants

These are **schema and application invariants** (see §1). Grain, gazetteer use, and interpolation are conventions in §8, not items here.

1. Identity Claims reference one Subject and one canonical entity and never substitute for Observations.
2. There is at most one Identity Claim per `(subject_id, entity_id)`.
3. A subject has at most one accepted Identity Claim (partial unique index on `status = 'accepted'`).
4. The subject and the entity share `subject_type_id` (composite foreign keys). “Same grain of place” is convention, not a check.
5. Members of an entity are the subjects with an accepted Identity Claim for that entity. After merge, claims live on the survivor.
6. Membership is not stored as an independently editable join table, and not as a column on `canonical_entities`.
7. Dropping the last member does not delete the canonical entity or change its id.
8. `subject_type_id` is immutable on the entity and on the subject.
9. Every canonical entity has a required `ref` of the form `{ref_prefix}-{token}`.
10. Provisional and rejected Identity Claims are not members. Absence of a claim is not a rejection.
11. Observations always target Subjects and always have a Citation (Interpretation schema). Canonical handles are not Observation subjects.
12. Creating one canonical entity does not require or imply creating related entities (no cascade in schema).
13. There is at most one Reconciliation Claim per `(entity_id, property_id)`.
14. A Reconciliation Claim's typed value must match `properties.value_type` (application write). For `value_type = 'subject'`, Conclusion uses `value_entity_id`.
15. Soft display merges are stateless UI projections; only an accepted Reconciliation Claim is a committed concluded value.
16. Person name format is a Property (`name_format`) or the project default, not a column on `canonical_entities`.
17. Association ends are Properties on the association entity (projected from member Observations and/or Reconciliation), not a separate endpoints table.
18. There is no `origination_claims` table and no `identity_anchor_id`; the canonical row is the working-subject insert, and membership is Identity Claims.
19. Optional `confidence_grade_id` on Identity and Reconciliation Claims is epistemic stance only; it does not replace `status` and must use Claim confidence vocabulary, not Source credibility grades ([`research-judgment-model.md`](research-judgment-model.md)).
20. Confirming that two Observations agree pins those Observations. It does not create an Observation, Citation, or subject (§5.1).
21. A confirmed match between an incoming subject and an existing accepted member pins both Observations on both Identity Claims. The existing claim's `argument` is not rewritten.
22. Removing one member does not remove the others. Identity Claims on that entity that pin the departing subject's Observations are surfaced for review (§5.2). Their remaining pins stay.

---

# 11. Claims scope note

This draft's Conclusion Claims are:

- **Identity Claims** (`identity_claims`) — one Subject is a reading of one canonical entity;
- **Reconciliation Claims** (`reconciliation_claims`) — concluded Property values on `canonical_entities`.

The entity insert is origination of a **handle**, not a third claim kind. The older generic `claims` / Record-resolution tables, per-kind canonical tables (`persons`, `events`, …), required `representative_subject_id`, `identity_anchor_id`, pairwise `sameness_claims` (`same_as` / `distinct_from`) with derived membership closure, derived-only association FKs, and a parked Place domain are superseded by this model.

---

# 12. Cross-layer examples

## 12.1 Photograph and testimony

```text
Photograph Source
  Artifact: scan.jpg
  Citation: crop around one person
  Observation: cited crop depicts person subject N1
  subject N1 (person, home Source = photograph)

Testimony Source
  Artifact: audio or research note
  Citation: "That's my grandfather"
  Observation: testimony refers to the same depicted person as subject N1
  subject N2 (person, home Source = testimony)

Conclusion
  canonical_entities E1 (kind=person)
  identity_claim: N1 → E1 (accepted)
  identity_claim: N2 → E1 (accepted), exhibit includes the testimony Observation
  members(E1) = {N1, N2}
```

## 12.2 Conflicting certificate and letter

```text
Birth certificate Source C
  person subject NC, birth_date Observation → 1 JAN 1800
  source subject SC reifying C

Letter Source L
  Observations:
    SC -- remark --> "date of birth on certificate mistyped"
    person subject NL -- birth_date --> 2 JAN 1800
    person subject NL -- birth_date --> 1 JAN 1800 (polarity negative)

Conclusion
  canonical_entities E (kind=person)
  identity_claim: NC → E (accepted)
  identity_claim: NL → E (accepted), exhibit cites both Sources' Observations
  members(E) = {NC, NL}
  reconciliation_claim on E, property birth_date:
    value → DateValue(2 JAN 1800)   # researcher choice, or synthesized
    evidence → the birth_date Observations (and related letter Observations as needed)
    argument → why the letter's correction is preferred
```

## 12.3 DNA evidence

```text
Source: DNA match report
Artifact: locally retained export/screenshot/JSON/CSV
Citation: match result
Observations on person subject ND:
  shared DNA
  predicted relationship
  match display name

Conclusion
  an identity claim may later make ND a member of a canonical person
  that claim is explicit for ND; other members are not pulled in by a path through ND
```

## 12.4 Asserted Place (Canada)

```text
PLC-canada
  kind = place
  members = ∅
  label = "Canada"
  argument = "working subject for the country; not cited from a family record"
```

Optional later: gazetteer Source → place subject → accepted Identity Claim onto `PLC-canada`. Events appear in Canada only via Location handles, not by this insert alone.

## 12.5 Inferred extra Location (siblings)

```text
Interpretation (unchanged, honest)
  Birth A: Location subject → place York
  Birth B (older sibling): Location subject → place Upper Canada
  Birth C (younger sibling): Location subject → place Upper Canada

Conclusion
  EVT-A, PLC-york, PLC-uc, LOC-york as usual (accepted identity claims for the corresponding Subjects)
  LOC-uc-extra
    kind = location
    members = ∅
    argument = "siblings usually born in a small area"
    Reconciliation event → EVT-A
    Reconciliation place → PLC-uc
    exhibit (typically on the place claim) → the three location Observations
```

---

# 13. Open schema questions

1. **Gazetteer runtime** — which editioned packs to ship or download (WOF vs GeoNames SQLite, Newberry, OHM extracts), license and ingest-as-Source UX. Product, not a third identity model.
2. **Place `contained_in`** — optional concluded Property vs search-only gazetteer parents. **Convention:** do not infer containment from two Locations on one Event; the schema will not stop a `contained_in` Observation or Reconciliation if someone adds that Property.
3. **Relationship endpoints** — seeded `participant` is a single Property; multi-party relationships may need repeated Properties, ordered parts, or a later shape. Out of scope for this Place pass.

---

# 14. Documentation ownership

To avoid competing schema definitions:

- [`source-layer-data-model.md`](source-layer-data-model.md) is authoritative for Source-layer tables and Artifact/File storage.
- [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md) is authoritative for Interpretation-layer tables and vocabulary.
- This document is authoritative for Conclusion-layer tables and Claims.
- [`structured-date-model.md`](structured-date-model.md) is authoritative for shared DateValue persistence.
- [`structured-name-model.md`](structured-name-model.md) is authoritative for shared NameValue persistence.
- [`seeded-vocabulary.md`](seeded-vocabulary.md) is the horizon catalog for intended keys and starter open-vocabulary lists (not a v1 ship list).
- [`audit-revision-history.md`](audit-revision-history.md) is authoritative for audit and revision history.
- [`research-judgment-model.md`](research-judgment-model.md) is authoritative for Claim confidence (and related Source/Citation judgment).
- [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) summarizes the three-layer philosophy and points here for schema detail.
