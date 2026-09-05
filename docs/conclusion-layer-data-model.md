# Provenencia Genealogy — Conclusion Layer Data Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for the Provenencia Conclusion layer.

The Conclusion layer answers:

> What does the researcher currently conclude about the historical world after considering one or more sources?

Cross-layer philosophy is summarized in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md). Interpretation subjects (Nodes, Observations, Properties) are defined in [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md). Shared date and name value models are [`structured-date-model.md`](structured-date-model.md) and [`structured-name-model.md`](structured-name-model.md). Audit history is [`audit-revision-history.md`](audit-revision-history.md). Researcher judgment (Claim confidence, and how it relates to Source credibility and Citation certainty) is [`research-judgment-model.md`](research-judgment-model.md).

This draft treats canonical rows as the **working set** (handles the researcher is investigating). The Conclusion **claims** are still only two kinds: Sameness and Reconciliation. Creating a handle is not a third claim type; the entity row is the origination.

---

# 1. Invariants versus conventions

The graph is built to **accept** research the schema cannot judge. A toponym `York, Upper Canada, North America` on one Place, a `same_as` between that Place and a colony-level Node, a Location with only one end filled, or an Observation that stretches what a Citation supports: these must **persist**. The evidence trail (Citations, Observation subjects, claim pins, `argument`) is how a later reader evaluates them. Product UI may warn, badge, or suggest splits; it must not refuse to save because a convention was skipped.

This document therefore uses three strengths of “must”:

| Strength | What it is | If the researcher ignores it |
|---|---|---|
| **Schema invariant** | SQLite `NOT NULL`, FKs, `UNIQUE`, `CHECK` | Insert/update fails |
| **Application invariant** | Writer/resolver rules that keep the generic graph typed (value column matches `value_type`, sameness endpoints share `node_type_key`, membership is the `same_as` component of the identity anchor) | Treat as malformed data (repair, ignore extras, or show an error on *that row*) — not as “you used Places wrong” |
| **Convention** | How to get the most from search, maps, merge, and honest reading of Sources | Data remains valid; features may be weaker or the proof harder to trust |

Seeded vocabulary (`location` + `event`/`place`, one-grain Places, gazetteer-as-Source) is convention plus picker defaults, not a closed ontology. Researchers may add Properties and Node Types; first-class UI may only light up for keys it knows.

Use-case text in this document often describes **rigorous, evidence-backed** genealogy: that is the path the product should make easy. A user who only wants a family tree — unlinked Persons, typed names, no Citations — must still be able to save. The result need not be scientifically defensible. Encourage and badge; do not block. Broader product stance: [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) §1.0.

---

# 2. Conclusion-layer principles

## 2.1 Working subjects may be sparse or ungrounded

A researcher may add a working Person, Event, Place, or association before any Interpretation Node exists for it.

```text
Person PER-7KD45
  label = "Mother of James"
  identity_anchor_id = NULL
  argument = "James existed; he had a mother"
```

This is valid. It is a **workspace handle**, not an Observation. Genealogical names, when known, come from cited NameValue Observations on member Nodes and/or name Reconciliation Claims. The canonical `label` is only a researcher working identifier.

**Schema:** Observations still require a Citation and still target a Node — you cannot make an Observation whose subject is a canonical entity. **Convention:** hang inferred geography on a knowledge or gazetteer Source (or a Reconciliation exhibit), not on a Citation that did not say it. The database will store the latter; the chain will look like the Source asserted it.

## 2.2 Identity cluster versus committed properties

**Membership** (which Interpretation Nodes are this subject) is the accepted `same_as` component of an optional **identity anchor** Node.

**Committed values** (name, date, location ends, …) are accepted Reconciliation Claims. If there is no claim, the UI may **project** values from Observations on member Nodes. That projection is not stored.

```text
with identity_anchor     members = same_as cluster of that Node
                         display values = accepted Reconciliation, else member Observations
without identity_anchor  members = ∅
                         display values = accepted Reconciliation only
```

## 2.3 Sameness and merge are related but distinct

**Node sameness** correlates Interpretation subjects:

```text
Node A same_as Node B
Node A distinct_from Node B
```

**Canonical merge** joins two Conclusion handles that were previously treated as distinct working subjects:

```text
PER-A merged_into PER-B
PLC-A merged_into PLC-B
```

Accepting Node sameness may connect two entities' membership components; the application should then merge or otherwise reconcile those canonical rows. Historical relationships or succession are not sameness or merge.

**Convention:** `same_as` is most useful between Places of the same grain (two readings of York the town). The schema will accept `same_as` between a composite “York, Upper Canada” Node and a colony Node; maps and merge will then treat them as one feature. Extra grains of “where this event happened” are usually additional Location handles rather than place identity — again a convention, not a CHECK.

## 2.4 Negation and conflict at Conclusion

Node co-reference uses Sameness Claims (`same_as` / `distinct_from`). Absence of a Sameness Claim is not `distinct_from`.

Attribute-level conflicts across member Observations are handled by soft display merges and, when durable, by Reconciliation Claims. That is separate from Observation polarity and from Node sameness.

## 2.5 Resolver logic is application-level

The database preserves multiple source-backed values. Resolvers may synthesize display without creating Claims. **Provenencia badges** on a handle (from records / inferred / asserted / unlinked) are computed from whether an identity anchor exists, whether Reconciliations pin Observations, and whether `argument` is set. They are not a stored enum.

## 2.6 Persistence conventions

Persistent rows use globally unique machine identifiers, currently UUIDv7 stored as 16-byte SQLite `BLOB` values. Ordinary schema tables use SQLite `STRICT` typing.

Structured genealogical dates use [`structured-date-model.md`](structured-date-model.md). Structured personal names use [`structured-name-model.md`](structured-name-model.md).

Generic change metadata belongs to [`audit-revision-history.md`](audit-revision-history.md).

Selected user-facing entities receive a required short human-readable `ref`. Canonical entities use `{ref_prefix}-{token}` (no candidate mark). Shared rules are in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md).

---

# 3. Design summary

Interpretation is a cited property graph (Nodes, Observations, Citations). Conclusion adds working handles and two claim verbs. The family tree and maps are **views** over that graph.

```text
Observations ──subject──► Node ◄──same_as──► other Nodes
                              │
                              │ identity_anchor_id (optional)
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
  - identity_anchor_id (nullable Node; live cluster when set)
  - argument (optional existence rationale)
  - label (researcher working identifier)
  - members = ∅, or accepted same_as closure from the identity anchor
  - payload = member-Node Observations projected, overridden by reconciliation_claims
```

Important separations:

1. **Canonical entity** — working subject; creating it *is* origination. No `origination_claims` table.
2. **Sameness Claim** — two Nodes co-refer (or not), with Observation exhibit.
3. **Membership** — derived from the identity anchor + `same_as` graph; not a join table of Nodes on the entity.
4. **Reconciliation Claim** — concluded value of one Property on one entity. Exhibit Observations may be about **other** Nodes; they are not retargeted.
5. **Canonical merge** — `merged_into_id` within the same `kind`.
6. **Independent handles** — creating a Person does not auto-create related Events, Places, or Locations.

SQLite enforces foreign keys and basic checks. Cluster membership, kind/anchor consistency, split repointing, and “a Location is usable only when both ends are known” are **application invariants**.

---

# 4. `sameness_claims`

A Sameness Claim asserts that two Nodes do or do not refer to the same historical thing.

Claim confidence grades (shared with Reconciliation Claims):

```sql
CREATE TABLE claim_confidence_grades (
    key         TEXT PRIMARY KEY,
    label       TEXT NOT NULL,
    sort_order  INTEGER NOT NULL,
    builtin     INTEGER NOT NULL DEFAULT 1,
    CHECK (builtin IN (0, 1))
) STRICT;
```

Seed keys: [`seeded-vocabulary.md`](seeded-vocabulary.md) §5.5. Semantics: [`research-judgment-model.md`](research-judgment-model.md).

```sql
CREATE TABLE sameness_claims (
    id              BLOB PRIMARY KEY,
    node_a_id       BLOB NOT NULL,
    node_b_id       BLOB NOT NULL,
    node_type_key   TEXT NOT NULL,
    relation        TEXT NOT NULL,
    status          TEXT NOT NULL,
    confidence_key  TEXT REFERENCES claim_confidence_grades(key),
    argument        TEXT,

    CHECK (relation IN ('same_as', 'distinct_from')),
    CHECK (status IN ('provisional', 'accepted', 'rejected')),
    CHECK (node_a_id < node_b_id),
    FOREIGN KEY (node_a_id, node_type_key)
        REFERENCES nodes (id, node_type_key),
    FOREIGN KEY (node_b_id, node_type_key)
        REFERENCES nodes (id, node_type_key),
    UNIQUE (node_a_id, node_b_id)
) STRICT;
```

`relation = same_as` means the researcher concludes the Nodes co-refer.  
`relation = distinct_from` means the researcher concludes they do not.

There is at most one Sameness Claim per unordered Node pair. Endpoint order is canonicalized with `node_a_id < node_b_id` so `(A, B)` and `(B, A)` cannot both exist. A pair therefore cannot simultaneously carry `same_as` and `distinct_from`; changing the conclusion updates the existing row (and is audited), rather than inserting a second Claim.

`status` is workflow state: `provisional`, `accepted`, or `rejected`. Only **accepted** `same_as` Claims participate in membership closure. `distinct_from` never enlarges membership regardless of status. See [`seeded-vocabulary.md`](seeded-vocabulary.md).

`provisional` is a real persisted conclusion-in-progress, not a missing row. The UI should surface provisional Claims differently from accepted ones (for example a distinct stroke, badge, or filter) so researchers can see candidates without treating them as membership. Rejected Claims remain in history for audit and should not drive the working graph.

`confidence_key` is optional epistemic stance (three-point Claim confidence vocabulary). It is **orthogonal to `status`**: accepted + low confidence and provisional + high confidence are both valid. Do not overload `provisional` to mean “I am unsure.” Authoritative rules: [`research-judgment-model.md`](research-judgment-model.md). Grade keys: [`seeded-vocabulary.md`](seeded-vocabulary.md).

`argument` holds the researcher's reasoning chain for the claim — correlation narrative, circumstantial synthesis, or a short note when a single Observation already makes the case. Reasoning stays here rather than scattered across evidence rows.

Sameness Claims point at Nodes, not at canonical entities. Each pairwise correlation has its own row and evidence so `A same_as B` and `B same_as C` can be justified independently. Transitive membership follows from the accepted graph; evidence does not have to be repeated onto a single “cluster claim.”

Both endpoints must share one Node Type. `node_type_key` on the claim is copied from those Nodes so SQLite can enforce that with composite foreign keys (`nodes` has `UNIQUE (id, node_type_key)`). That copy cannot drift: Node type is immutable, and a Claim cannot point at a Node whose type differs from `node_type_key`.

---

# 5. `sameness_claim_evidence`

Evidence rows are pointers to Observations that participate in the claim's exhibit list. An individual Observation does not carry a stance toward the Sameness Claim; the claim's `relation` and `argument` express the researcher's conclusion over the whole set.

```sql
CREATE TABLE sameness_claim_evidence (
    sameness_claim_id   BLOB NOT NULL REFERENCES sameness_claims(id) ON DELETE CASCADE,
    observation_id      BLOB NOT NULL REFERENCES observations(id),

    PRIMARY KEY (sameness_claim_id, observation_id)
) STRICT;
```

Pin every relevant Observation here; write the conclusion and inference in `sameness_claims.argument`. Exhibit pins are **Observations only** for now — not Citations, Sources, or other Claims. If a prior correlation matters, say so in `argument` (and pin the Observations that support this claim). Broader exhibit types can wait until a concrete workflow needs them.

The Observation's `subject_node_id` is unchanged. Pins mean “this assertion is part of my proof,” not “this Observation is now about the claim.”

---

# 6. `canonical_entities`

Because canonical rows are thin handles, they share one table instead of parallel `persons` / `events` / … tables.

```sql
CREATE TABLE canonical_entities (
    id                      BLOB PRIMARY KEY,
    kind                    TEXT NOT NULL REFERENCES node_types(key),
    ref                     TEXT UNIQUE NOT NULL,
    identity_anchor_id      BLOB REFERENCES nodes(id),
    argument                TEXT,
    label                   TEXT,
    merged_into_id          BLOB REFERENCES canonical_entities(id)
) STRICT;
```

`kind` is the same vocabulary as Interpretation Node Types. The table is an implementation detail; in the UI a `person` row is a Person (`PER-7KD45`), not a “canonical entity.” `kind` is immutable after insert.

`ref` is required: `{ref_prefix}-{token}`. Candidate person Nodes use `PER-C-…` and stay distinct.

`identity_anchor_id` is the optional Node whose identity cluster *is* this subject. It replaces the older required `representative_node_id`. Creating a handle from a record sets the anchor to that Node. **Adopt** (first grounding of an inferred handle) sets it later. When the current anchor leaves the accepted `same_as` component, the application repoints to any remaining member Node; the canonical id does not change for that reason alone.

```text
members(entity) =
  ∅  if identity_anchor_id IS NULL
  else all Nodes reachable from that Node
       via accepted sameness_claims where relation = 'same_as'
```

`argument` is optional existence rationale on the handle (“working subject for Canada; not taken from a citation in this project”). Accumulating commentary stays in `canonical_entity_notes`. Do not treat `argument` as a genealogical Property.

Rules:

- When `identity_anchor_id` is set, `kind` must match that Node's `node_type_key` (application invariant).
- `merged_into_id`, when set, should reference another entity of the same `kind`.
- `label` is an optional researcher working identifier (for example `Mother of James`). It is not a genealogical name and must not substitute for NameValue Observations or name Reconciliation Claims.
- `ref` is the stable short public reference, assigned on insert as `{ref_prefix}-{token}`. After merge, old refs keep resolving to the surviving entity.
- Later accepted `same_as` Claims expand membership automatically from the anchor. Do not list member Nodes on the entity row.
- Creating one handle does not auto-create related kinds. Reification `source` Nodes are not typically given canonical rows.
- Primary kinds (`person`, `event`, `place`) may be created with no anchor, no `argument`, and no Reconciliations (stubs). Association kinds (`location`, `participation`, `relationship`) are edges: **conventionally** they are not shown as complete until endpoint Properties are known (see §7.3). Incomplete Location handles still persist.

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
| **From records** | `identity_anchor_id` set (cluster may still be sparse) |
| **Inferred** | no anchor; at least one accepted Reconciliation with Observation pins |
| **Asserted** | no anchor; `argument` set; no pinned Reconciliations |
| **Unlinked** | no anchor, empty `argument`, no accepted Reconciliations |

Unlinked stubs should stay off graphs and maps; list them in an unlinked / stubs view.

A Node-backed handle is a better **trace to Interpretation**, not automatically better historical geography than a well-pinned inferred Location.

---

# 7. `reconciliation_claims`

A Reconciliation Claim asserts:

```text
for canonical entity E, Property P has concluded value V
```

The value is stored **on the claim**. Observation pins are optional exhibit. Pins may point at Observations whose subjects are other Nodes (sibling births, a gazetteer line). Those Observations are not retargeted and are not copied onto a fictional Location Node.

```sql
CREATE TABLE reconciliation_claims (
    id              BLOB PRIMARY KEY,
    entity_id       BLOB NOT NULL REFERENCES canonical_entities(id) ON DELETE CASCADE,
    property_key    TEXT NOT NULL REFERENCES properties(key),
    status          TEXT NOT NULL,
    confidence_key  TEXT REFERENCES claim_confidence_grades(key),
    argument        TEXT,

    value_text      TEXT,
    value_integer   INTEGER,
    value_real      REAL,
    value_boolean   INTEGER,
    value_date_id   BLOB REFERENCES date_values(id),
    value_name_id   BLOB REFERENCES name_values(id),
    value_entity_id BLOB REFERENCES canonical_entities(id),

    CHECK (status IN ('provisional', 'accepted', 'rejected')),
    CHECK (value_boolean IS NULL OR value_boolean IN (0, 1)),
    UNIQUE (entity_id, property_key)
) STRICT;
```

`entity_id` is a real foreign key to `canonical_entities`. The entity's `kind` is available via that join; it is not duplicated on the claim row.

`property_key` is the extensibility hook. Any Property in the vocabulary may be concluded on a suitable entity (subject to application rules about which Properties make sense for that `kind`).

Exactly one value representation must be populated, and it must match `properties.value_type`, parallel to Observations. For `value_type = 'node'`, Conclusion stores **`value_entity_id`** (another canonical handle). Interpretation Observations still use `value_node_id` (Node to Node). That split is intentional: a nodeless Place (asserted Canada) has no Node to point at. Matching `kind` to the Property's target hint is an application invariant.

As with Observations, typed-column matching is an **application write invariant** for now. Readers prefer the column matching `value_type` if extras are present. The concluded value may match one exhibit Observation, be copied from a related Node's value, or be synthesized (for example a DateValue spanning Apr–May 1985).

There is at most one Reconciliation Claim per `(entity, property)`. Changing the concluded value, `status`, or `confidence_key` updates that row (and is audited).

`status` matches Sameness Claims: `provisional`, `accepted`, or `rejected`. **Only `accepted` is the committed concluded value.** `provisional` is a persisted working choice the UI should show differently. `rejected` is kept for audit and is not the working value.

`confidence_key` matches Sameness Claims: optional three-point Claim confidence, orthogonal to `status`. See [`research-judgment-model.md`](research-judgment-model.md). Grades live in `claim_confidence_grades` (§4).

From a modeling standpoint a Property on an entity either has a Reconciliation Claim or it does not. Soft blends of member Observations are **stateless display**. They are never persisted as automatic claims. There is no `origin` column and no `association_endpoints` table: association ends are ordinary Properties (`event`, `place`, `person`, `participant`) resolved like any other.

`argument` holds researcher reasoning when needed.

## Name format as reconciliation

Cultural name display/entry ordering is not a column on `canonical_entities`. It is concluded like other Person-scoped facts:

```text
Property name_format -> text   # value is a name_format_profiles.key, e.g. "western"
```

- `project_settings.default_name_format_key` supplies the UI default when a person entity has no accepted `name_format` Reconciliation Claim.
- When the researcher commits a format for a Person (including as part of reconciling names across cultures), they persist a Reconciliation Claim for `name_format`.
- Evidence pins are optional for `name_format` (it is often a preference rather than source-derived); `argument` may still record why that profile was chosen.
- Concluded `name` values remain separate Reconciliation Claims (`property_key = name`, NameValue). Format and name content are related in the UI but distinct Properties.

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
  identity_anchor → N1
  members: person Nodes N1, N2

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

On **Location Nodes**, `event` and `place` are node-valued Observations (Interpretation). A canonical Location **with** an identity anchor projects those Observations and resolves each target Node to a canonical Event/Place when those handles exist.

A canonical Location **without** an anchor has no such Observations. Its ends are accepted Reconciliations:

```text
LOC-…  (kind = location, identity_anchor_id = NULL)
  argument = optional rationale for the whole edge
  Reconciliation event → value_entity_id = EVT-…   (this birth)
  Reconciliation place → value_entity_id = PLC-…   (Upper Canada)
```

Pins typically sit on the claim they justify (geography Observations on `place`; identifying the birth on `event` if needed). **Convention:** do not invent a Location Node for a conclusion-only edge, and do not add Observations to a Citation that did not support that end. Both are storeable; they mislabel the Source.

The same pattern applies to Participation (`person`, `event`, `role`) and Relationship (`participant`, `relationship_type`): cluster projection when anchored; Reconciliation when not. `role` / `relationship_type` stay scalar Reconciliations (or member Observations) as today.

A Location is **M:N membership** of Events in Places: one Event may have several Location handles (York *and* Upper Canada); one Place many events. Two Locations on one Event do not imply gazetteer containment.

---

# 8. Places and Locations

This section is **convention** unless noted. The schema does not know a township from a comma-separated blob.

**Convention that pays off:** treat a Place as one geographic feature at **one grain** (this town, this township, this colony, this farm), and treat Location as M:N membership of Events in those Places. Genealogy then groups events and can later draw a map without a locality-history encyclopedia.

**What the graph still allows:** a single Place whose `toponym` is `York, Upper Canada, North America`; `same_as` between that Place and any other place Node of the same type; a Location with only `event` filled; zero Citations. Save succeeds. Search, gazetteer bind, and “events in Upper Canada” roll-up will be weaker or misleading. The exhibit (or lack of one) is how another researcher judges it.

## 8.1 Interpretation

A source phrase `"York, Upper Canada"` **can** be one place Node and one toponym string. **Convention:** split into two place Nodes (or two Location memberships) at the grains the Citation actually supports, so York does not `same_as` the colony when a second source only says Upper Canada. Toponyms on place Nodes use Property `toponym` in [`seeded-vocabulary.md`](seeded-vocabulary.md) (open: other Properties are fine). Event date is usually the time context rather than a date on the Place Node — convention, not a CHECK.

A Location Node is the usual source-local association: this **event Node** at this **place Node**, each end an Observation with a Citation. Other shapes remain valid graph data.

## 8.2 Canonical Places

Create `PLC-…` like any other handle:

- **From a record:** identity anchor = a place Node (census line, gazetteer feature, …).
- **Asserted:** no anchor; `label` = Canada; optional `argument`. Citing a baptism that never named Canada is allowed and will look like that baptism supported Canada.
- **Gazetteer as Source (convention for citable geography):** ingest an editioned dump as a Source/Artifact; interpret a feature into a place Node; ground the canonical Place there. Google Maps is a renderer. Live geocoders make poor frozen Citations; dated packs (Who's On First / GeoNames SQLite, Newberry shapefiles, dated OHM extracts) match Source+Artifact better. Wikidata Q-ids are concordances; a snapshot is what you cite if Wikidata is the Source.

Farms and unnamed lots may stay project-only with no gazetteer feature.

**Containment:** two Locations on one Event do not imply a parent polygon. Search may offer an explicit “include gazetteer parents” lens. A concluded `contained_in` Property is optional later. **Convention:** do not assume York ⇒ Upper Canada unless some exhibit says so.

## 8.3 Extra grain (inferred Location)

Sibling births say Upper Canada; this birth says only York. **Convention:** keep three honest Interpretation Locations, then add a nodeless Location handle with Reconciliations (example below). **Also valid:** one composite Place, or `same_as` across grains — the app should not throw.

```text
EVT-this-birth     identity-anchored on this birth Node
PLC-york           from this register's place Node (settlement grain)
PLC-upper-canada   from sibling place Nodes and/or asserted/gazetteer Place
LOC-york           anchored on this birth's Location Node (event + York)
LOC-uc             no anchor
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

Typical trigger: an accepted `same_as` Claim connects Nodes that currently sit under two different canonical entities of that kind. The application merges the entities and keeps a single identity anchor in the combined component.

Old ids and human refs should continue resolving to the surviving entity. Merge remains explicit and auditable. Reconciliation Claims on the absorbed entity must be merged or re-pointed by application rules.

---

# 10. Schema and application invariants

These are **schema and application invariants** (see §1). Grain, gazetteer use, and interpolation are conventions in §8, not items here.

1. Sameness Claims reference two distinct Nodes and never substitute for Observations.
2. There is at most one Sameness Claim per unordered Node pair (`UNIQUE (node_a_id, node_b_id)` with canonical endpoint order).
3. Both endpoints of a Sameness Claim have the same Node Type (`node_type_key` plus composite FKs to `nodes`). Same-type is what the schema can enforce; “same grain of place” is convention.
4. Each accepted `same_as` component should correspond to at most one non-merged `canonical_entities` row of the matching `kind` (application).
5. Members are ∅ if `identity_anchor_id` is NULL; otherwise the accepted `same_as` component of that Node (application resolver).
6. Membership is not stored as an independently editable join table.
7. `identity_anchor_id` may change when the previous anchor leaves the component; the canonical entity id does not change for that reason alone.
8. `kind` is immutable. When an anchor is set, it matches that Node's `node_type_key` (application).
9. Every canonical entity has a required `ref` of the form `{ref_prefix}-{token}`.
10. `distinct_from` and rejected Claims do not enlarge membership.
11. Observations always target Nodes and always have a Citation (Interpretation schema). Canonical handles are not Observation subjects.
12. Creating one canonical entity does not require or imply creating related entities (no cascade in schema).
13. There is at most one Reconciliation Claim per `(entity_id, property_key)`.
14. A Reconciliation Claim's typed value must match `properties.value_type` (application write). For `value_type = 'node'`, Conclusion uses `value_entity_id`.
15. Soft display merges are stateless UI projections; only an accepted Reconciliation Claim is a committed concluded value.
16. Person name format is a Property (`name_format`) or the project default, not a column on `canonical_entities`.
17. Association ends are Properties on the association entity (projected from member Observations and/or Reconciliation), not a separate endpoints table.
18. There is no `origination_claims` table; the canonical row is the working-subject insert.
19. Optional `confidence_key` on Sameness and Reconciliation Claims is epistemic stance only; it does not replace `status` and must use Claim confidence vocabulary, not Source credibility keys ([`research-judgment-model.md`](research-judgment-model.md)).

---

# 11. Claims scope note

This draft's Conclusion Claims are:

- **Sameness Claims** (`sameness_claims`) — Node co-reference;
- **Reconciliation Claims** (`reconciliation_claims`) — concluded Property values on `canonical_entities`.

The entity insert is origination of a **handle**, not a third claim kind. The older generic `claims` / Record-resolution tables, per-kind canonical tables (`persons`, `events`, …), required `representative_node_id`, derived-only association FKs, and a parked Place domain are superseded by this model.

---

# 12. Cross-layer examples

## 12.1 Photograph and testimony

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

## 12.2 Conflicting certificate and letter

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

## 12.3 DNA evidence

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

## 12.4 Asserted Place (Canada)

```text
PLC-canada
  kind = place
  identity_anchor_id = NULL
  label = "Canada"
  argument = "working subject for the country; not cited from a family record"
```

Optional later: gazetteer Source → place Node → set `identity_anchor_id`. Events appear in Canada only via Location handles, not by this insert alone.

## 12.5 Inferred extra Location (siblings)

```text
Interpretation (unchanged, honest)
  Birth A: Location Node → place York
  Birth B (older sibling): Location Node → place Upper Canada
  Birth C (younger sibling): Location Node → place Upper Canada

Conclusion
  EVT-A, PLC-york, PLC-uc, LOC-york as usual (anchors on the corresponding Nodes)
  LOC-uc-extra
    kind = location
    identity_anchor_id = NULL
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
