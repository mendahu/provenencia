# Provenencia Genealogy — Structured Date Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for Provenencia's shared genealogical date value model.

Structured dates are cross-layer infrastructure. They may be referenced by Source metadata, Interpretations, Claims, Conclusions, and other future domain objects that need to represent genealogical dates faithfully.

---

# 1. Design philosophy

Genealogical dates cannot be represented adequately by a conventional SQL `DATE` or `DATETIME` alone.

Research data commonly contains:

```text
14 MAY 1985
MAY 1985
1985
ABT 1985
BEF 1900
AFT 1872
BET 1880 AND 1885
FROM 1880 TO 1885
14 MAY 1985 15:30
```

A date may therefore be exact, partial, approximate, bounded, ranged, period-based, calendar-specific, clock-precise, or expressed as a phrase.

The structured date model preserves those semantics while providing enough structure for useful operations such as sorting, filtering, comparison, and timeline display.

Precision is cascading and optional: a value may assert only a year, a calendar day, or a full timestamp through milliseconds. Absent finer fields mean **unknown / not asserted**, not midnight or zero.

A DateValue is conceptually a value object. Its database UUID provides persistence identity and referential convenience; it does not mean the date itself has independent genealogical identity.

Timezone / offset is not modeled in the first schema; local or unspecified clock times are stored as civil components only.

---

# 2. `date_values`

```sql
CREATE TABLE date_values (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    kind            TEXT NOT NULL,
    qualifier       TEXT,
    calendar        TEXT,

    start_year      INTEGER,
    start_month     INTEGER,
    start_day       INTEGER,
    start_hour      INTEGER,
    start_minute    INTEGER,
    start_second    INTEGER,
    start_millisecond INTEGER,

    end_year        INTEGER,
    end_month       INTEGER,
    end_day         INTEGER,
    end_hour        INTEGER,
    end_minute      INTEGER,
    end_second      INTEGER,
    end_millisecond INTEGER,

    phrase          TEXT,

    CHECK (start_month IS NULL OR start_month BETWEEN 1 AND 12),
    CHECK (end_month   IS NULL OR end_month BETWEEN 1 AND 12),
    CHECK (start_day   IS NULL OR start_day BETWEEN 1 AND 31),
    CHECK (end_day     IS NULL OR end_day BETWEEN 1 AND 31),
    CHECK (start_hour IS NULL OR start_hour BETWEEN 0 AND 23),
    CHECK (end_hour   IS NULL OR end_hour BETWEEN 0 AND 23),
    CHECK (start_minute IS NULL OR start_minute BETWEEN 0 AND 59),
    CHECK (end_minute   IS NULL OR end_minute BETWEEN 0 AND 59),
    CHECK (start_second IS NULL OR start_second BETWEEN 0 AND 59),
    CHECK (end_second   IS NULL OR end_second BETWEEN 0 AND 59),
    CHECK (start_millisecond IS NULL OR start_millisecond BETWEEN 0 AND 999),
    CHECK (end_millisecond   IS NULL OR end_millisecond BETWEEN 0 AND 999)
) STRICT;
```

Components cascade on each side (`start_*` / `end_*`): year → month → day → hour → minute → second → millisecond. A finer field must not be set unless all coarser fields on that side are set.

The exact vocabulary and validation rules for `kind`, `qualifier`, and `calendar` should be refined as the date parser and domain behavior are implemented. The schema deliberately leaves room for the date semantics already identified without reducing genealogical dates to a single normalized timestamp.

---

# 3. Cross-layer use

Tables in any domain layer may reference a structured date:

```sql
date_value_id BLOB REFERENCES date_values(id)
```

For example, Source metadata can preserve both the source wording and a structured value:

```text
publication_date
  value_text = "about the year 1890"
  date_value_id = DateValue(ABT 1890)
```

The textual representation remains important for source fidelity. The structured DateValue adds machine-readable semantics; it does not replace the original wording where that wording is meaningful evidence.

Later layers may use the same DateValue model for interpreted or concluded dates without introducing a second incompatible date representation.

---

# 4. Architectural rules

1. `date_values` is shared cross-layer infrastructure, not part of the Source layer.
2. Genealogical dates are not reduced to SQL `DATE` / `DATETIME` values.
3. Partial and qualified dates are first-class values; precision is cascading.
4. Ranges and periods can preserve distinct start and end components (including optional clock time).
5. Original textual wording may be retained by the referencing domain object when fidelity requires it.
6. The DateValue persistence UUID does not imply genealogical entity identity.
7. All tables use SQLite `STRICT` typing.
8. Missing time components are unknown, not midnight; timezone is deferred.

This shared model gives Source, Interpretation, and Conclusion data one consistent representation for genealogical dates while allowing each layer to preserve its own evidentiary or interpretive context.
