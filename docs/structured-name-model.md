# Provenencia Genealogy — Structured Name Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for Provenencia's shared personal name value model.

Structured names are cross-layer infrastructure. They may be referenced by Interpretations, Conclusions, and other future domain objects that need to represent personal names for display, search, and reconciliation without forcing a single cultural naming schema.

---

# 1. Design philosophy

Personal names cannot be represented adequately by a single undifferentiated string when researchers need to search, compare, and reconcile evidence across Sources.

At the same time, names must not be forced into a closed Western `given` / `surname` pair. Cultures differ in how many name parts exist, which parts are family names, whether particles or patronymics appear, whether initials are used, and how nicknames relate to formal names.

The structured name model therefore:

```text
always stores a full-form normalized reading
optionally stores ordered parts with an open type vocabulary
```

Citation transcription remains the fidelity layer for what the Source actually said (`Wm Robins`, uncertain marks, and so on). A NameValue is the Interpretation (or later Conclusion) normalization of a name assertion, parallel to DateValue for dates.

A NameValue is conceptually a value object. Its database UUID provides persistence identity and referential convenience; it does not mean the name itself has independent genealogical identity as a research subject.

---

# 2. `name_values`

```sql
CREATE TABLE name_values (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    form            TEXT NOT NULL
) STRICT;
```

`form` is the required full-form normalized reading of the name as a unit, for example:

```text
James K. Robins
Maria da Silva Costa
蒋浩
```

Researchers may leave parts empty and rely on `form` alone when further segmentation is unnecessary or culturally unclear.

---

# 3. `name_value_parts`

```sql
CREATE TABLE name_value_parts (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    name_value_id   BLOB NOT NULL REFERENCES name_values(id) ON DELETE CASCADE,
    idx             INTEGER NOT NULL,
    value           TEXT NOT NULL,
    type            TEXT,

    UNIQUE (name_value_id, idx),
    CHECK (idx >= 0)
) STRICT;
```

Parts are ordered by `idx` within a NameValue. `type` is stored as TEXT (nullable). **Product-known keys** live in the compiled `namevalues` registry (`PartTypes()`); Insert rejects unknown non-empty types. Empty type means an untyped segment. The GEDCOM-aligned key set is catalogued in [`seeded-vocabulary.md`](seeded-vocabulary.md) §4.1. UI shows **localized labels** for those keys (S7-02b) — not free-text type entry. User-minted part types (catalog rows, like `property_terms`) are deferred.

Applications use the same keys for search, reconcile, and format profiles. Absence of parts is valid. Culture-specific *display order* is handled by name format profiles rather than by closing the column in DDL.

Example:

```text
NameValue
  form = "James K. Robins"
  parts:
    idx=0 value="James"  type=given
    idx=1 value="K."     type=initial
    idx=2 value="Robins" type=surname
```

---

# 4. Name format profiles

Part types alone do not define how a name should be entered or displayed. A tree that spans cultures needs per-person preferences and a project default.

## 4.1 `name_format_profiles`

```sql
CREATE TABLE name_format_profiles (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    key             TEXT NOT NULL,
    origin          TEXT NOT NULL,             -- provenencia | user | plugin:<id>
    label           TEXT NOT NULL,
    description     TEXT,

    UNIQUE (key, origin)
) STRICT;
```

A profile is a named cultural/display convention such as Western / GEDCOM-style ordering. Profiles are seeded data (`origin = 'provenencia'`), not schema enums. Projects may add more profiles later (for example Latin American dual-surname ordering) without migrations. Vocabulary origin: [`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1.

## 4.2 `name_format_profile_parts`

```sql
CREATE TABLE name_format_profile_parts (
    profile_id      BLOB NOT NULL REFERENCES name_format_profiles(id) ON DELETE CASCADE,
    idx             INTEGER NOT NULL,
    part_type       TEXT NOT NULL,

    PRIMARY KEY (profile_id, idx),
    CHECK (idx >= 0)
) STRICT;
```

`idx` is the preferred assembly / entry order for parts of the given `part_type`. Types not listed in a profile may still appear on a NameValue; the profile guides UI and display assembly when parts are present.

### Seeded Western profile

New projects should include at least the `western` profile (`origin = 'provenencia'`). The authoritative profile key, labels, and part order are in [`seeded-vocabulary.md`](seeded-vocabulary.md).

This is a display/entry convention, not a claim that every Person uses every part type.

## 4.3 Project default

The portable project database should store a default profile used when a person entity has no concluded format:

```sql
CREATE TABLE project_settings (
    key             TEXT PRIMARY KEY,
    value_text      TEXT
) STRICT;
```

Known setting for name formats (also listed in [`seeded-vocabulary.md`](seeded-vocabulary.md)):

```text
default_name_format_id = <uuid of western profile>
```

`value_text` holds the `name_format_profiles.id` (hex UUID text). Additional project settings may reuse this table later. Do not store bare `key` alone — keys are unique only within an `origin`.

## 4.4 Person preference via Reconciliation Claim

Per-person format is not a column on `canonical_entities`. It is concluded like other Person-scoped facts:

```text
Property name_format -> text
Reconciliation Claim on person entity E:
  property = name_format
  value_text = <name_format_profiles.id>   # not bare key
```

Resolution order for UI display/entry:

1. Accepted Reconciliation Claim for `name_format` on that person entity, if any.
2. Otherwise `project_settings.default_name_format_id`.

Evidence pins on a `name_format` claim are optional (the choice is often a researcher preference rather than source-derived). `argument` may still record why that profile was chosen. Concluded genealogical names remain separate claims on Property `name` (NameValue).

The profile does not replace `NameValue.form` or cited Observations. It tells the application how to present and edit structured parts for that Person (and how to suggest ordering when capturing new name evidence). Reconciliation Claims themselves are defined in [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

---

# 5. Cross-layer use

Tables in any domain layer may reference a structured name:

```sql
name_value_id BLOB REFERENCES name_values(id)
```

For Interpretation Observations with `properties.value_type = 'name'`:

```text
Citation.transcription
  "Wm Robins"

Observation
  Person N1 -- name --> NameValue
    form = "William Robins"
    parts: William (given), Robins (surname)
```

Search and reconciliation should use parts when present and fall back to `form` (including fuzzy or full-text strategies) when parts are absent or untyped.

Canonical entity `label` is a researcher working identifier, not a genealogical name. Structured NameValue Observations on member Nodes remain authoritative for name evidence. Cultural display/entry ordering for a Person is the concluded Property `name_format` (falling back to the project default).

---

# 6. Architectural rules

1. `name_values` and `name_value_parts` are shared cross-layer infrastructure, not part of the Source layer.
2. Personal names are not reduced to a single undifferentiated string when structure is known and useful.
3. A full-form `form` is always required; parts are optional.
4. Part `type` is a product registry key (or empty for untyped), not free text and not a culture-specific closed schema; user-minted types are deferred.
5. Name format profiles define cultural display/entry ordering; they are seeded rows with `origin` namespaces, not enums.
6. Projects have a default name format; each Person may override it via a `name_format` Reconciliation Claim.
7. One cited name assertion is one NameValue, not one Observation per token.
8. Multiple name Observations on the same person Node remain valid (alternate forms, nicknames asserted separately, name changes over time, and so on).
9. Citation text preserves source wording; NameValue holds normalized interpretation.
10. The NameValue persistence UUID does not imply genealogical entity identity.
11. All tables use SQLite `STRICT` typing.

This shared model gives Interpretation and Conclusion one consistent representation for personal names while allowing each layer to preserve its own evidentiary or interpretive context.
