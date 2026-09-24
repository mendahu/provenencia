# Provenencia Genealogy — Structured Date Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for Provenencia's shared genealogical date value model.

Structured dates are cross-layer infrastructure. They may be referenced by Source metadata, Interpretations, Claims, Conclusions, and other future domain objects that need to represent genealogical dates faithfully.

---

# 1. Design philosophy

A **DateValue always represents one genealogical date** — one researched moment (a day, a year, a season, a clock time on a day, and so on). It is never “how long something lasted.” Duration / period events (a residence FROM–TO, a war spanning years) are modeled as **two** DateValues (or a future domain concept), not as one DateValue with overloaded start/end meaning.

Genealogical sources rarely hand us a clean SQL `DATE`. They give:

```text
14 MAY 1985
MAY 1985
1985
ABT 1985
BEF 1900
AFT 1872
BET 1880 AND 1885
14 MAY 1985 15:30
14 MAY 1985 15:30 Eastern Standard Time
Christmas 1887
harvest 1842
```

What varies is **precision** and **how uncertainty is stated**, not whether we are talking about one date or two.

Provenencia therefore stores:

1. **Shape** — a single point, or a **bounded uncertainty window** (between earliest and latest).
2. **Approximation** — as stated, about, before, or after (on a point).
3. **Components** — whichever civil fields the evidence supports (year, month, day, time…), with missing finer fields meaning **unknown / not asserted**, never midnight or zero.
4. **Optional phrase** — unstructured wording that is part of the DateValue itself when structure alone cannot carry the meaning.
5. **Optional free-text timezones** — as written (`start_tz` / `end_tz`), not forced IANA conversion to UTC.

A DateValue is a value object. Its database UUID is persistence identity only; it is not an independent genealogical entity.

---

# 2. Shape and approximation

## 2.1 `kind` — point vs bounded window

| `kind` | Meaning |
| --- | --- |
| `point` | One date at whatever precision the components assert. |
| `range` | One date that falls **between** two bounds (GEDCOM-style `BET … AND …`). Earliest and latest are limits on **where that single date lies**, not the start and end of an event. |

There is **no** separate `exact` vs `year` kind. Year-only, month+year, full day, and day+time are all `point` values distinguished by which components are set.

## 2.2 Approximation on a point

| `qualifier` | Meaning (point only) |
| --- | --- |
| *(empty)* | As stated — no extra approximation beyond the components. |
| `ABT` | About / approximately the point. |
| `BEF` | Before / no later than the point. |
| `AFT` | After / no earlier than the point. |

**Between** is not a qualifier. It is `kind = range` (two bounds, qualifier empty).

`ABT` / `BEF` / `AFT` are invalid on `range`: the window *is* the approximation.

## 2.3 Mapping common source forms

| Source wording (examples) | Structured shape |
| --- | --- |
| `14 MAY 1985` | `point`, Y+M+D |
| `MAY 1985` | `point`, Y+M |
| `1985` | `point`, Y |
| `ABT 1985` | `point`, Y, `ABT` |
| `BEF 1900` | `point`, Y, `BEF` |
| `AFT 1872` | `point`, Y, `AFT` |
| `BET 1880 AND 1885` | `range`, start Y=1880, end Y=1885 |
| `14 MAY 1985 15:30` | `point`, Y+M+D+H+M |
| `Christmas 1887` | `point`, Y=1887, `phrase` = Christmas (or similar) |

`FROM 1880 TO 1885` as a **period lasting those years** is **not** this `range` kind. Prefer two DateValues on the domain object (period start / period end) or a later period model.

---

# 3. `date_values`

```sql
CREATE TABLE date_values (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    kind            TEXT NOT NULL,             -- point | range
    qualifier       TEXT,                      -- '' | ABT | BEF | AFT (point only)
    calendar        TEXT,

    start_year      INTEGER,
    start_month     INTEGER,
    start_day       INTEGER,
    start_hour      INTEGER,
    start_minute    INTEGER,
    start_second    INTEGER,
    start_millisecond INTEGER,
    start_tz        TEXT,

    end_year        INTEGER,
    end_month       INTEGER,
    end_day         INTEGER,
    end_hour        INTEGER,
    end_minute      INTEGER,
    end_second      INTEGER,
    end_millisecond INTEGER,
    end_tz          TEXT,

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

DDL does not CHECK `kind` / `qualifier` enums so vocabulary can evolve in application validation without a migration for every refinement.

## 3.1 Component precision

On each side (`start_*` / `end_*`), any subset of civil components may be set. Sources are often fragmentary — “May”, “the 14th”, a clock time with no year — and the model must store what the evidence actually asserts.

```text
year, month, day, hour, minute, second, millisecond
```

Rules:

1. A DateValue is valid when **`phrase` is set** or **at least one** of year / month / day / hour / minute / second is set. Empty structure with no phrase is invalid. Timezone alone is not enough.
2. **Gaps are allowed.** Day without month, month without year, hour without a civil date are all legal. Do not invent the missing fields.
3. Absent fields mean **unknown / not asserted**, not zero or midnight.
4. Present fields must be in range (month 1–12, day 1–31, hour 0–23, and so on).
5. `point` uses only the start side (`end_*` including `end_tz` must be empty).
6. `range` requires both sides, each with at least one civil component; start ≤ end when shared fields can be compared.

Timezone labels (`start_tz` / `end_tz`) are independent of the components: optional free text. Empty means unspecified. They are not parsed into UTC offsets at the storage layer. `point` uses `start_tz` only; `range` may set either or both.

## 3.2 `phrase`

`phrase` is optional text that belongs to the **DateValue** when the date is partly or wholly verbal (`Christmas 1887`, dual OS/NS wording, seasonal labels) and structure alone is incomplete.

It is **not** the place for Source-layer transcription fidelity. Callers that preserve “as written on this record” keep that on the referencing row (e.g. Source metadata `value_text`) and attach `date_value_id` for structure. See §4.

A DateValue may be phrase-forward (structure sparse or empty) when the evidence is only verbal; prefer adding whatever components can be asserted without inventing precision.

## 3.3 Calendar

`calendar` records the calendar system when known (default in product UI: Gregorian). It does not replace dual-date wording in `phrase` when both old and new style appear in the source.

---

# 4. Cross-layer use

Tables in any domain layer may reference a structured date:

```sql
date_value_id BLOB REFERENCES date_values(id)
```

Source metadata (and similar) may preserve both wording and structure:

```text
publication_date
  value_text = "about the year 1890"    -- fidelity for this attachment
  date_value_id = DateValue(point, ABT, 1890)
```

The structured DateValue adds machine-readable semantics for sorting, filtering, comparison, and timelines. It does not replace original wording where that wording is meaningful evidence.

Later layers reuse the same DateValue model for interpreted or concluded dates without a second incompatible representation. Refined conclusions should usually be **new** DateValue rows (or later-layer assertions), not silent mutation of Source evidence.

---

# 5. Architectural rules

1. `date_values` is shared cross-layer infrastructure, not part of the Source layer.
2. Every DateValue is **one** genealogical date; `range` is a bounded uncertainty window for that date, not an event duration.
3. Genealogical dates are not reduced to SQL `DATE` / `DATETIME` values.
4. Precision is carried by whichever optional components the evidence asserts (gaps allowed), not by proliferating kinds (`point` vs `range` only).
5. Approximation on a point is `qualifier` (`ABT` / `BEF` / `AFT` / empty); **between** is `kind = range`.
6. Original textual wording for a domain attachment may be retained on the referencing object (`value_text`, etc.); `phrase` is optional wording on the DateValue itself.
7. The DateValue persistence UUID does not imply genealogical entity identity.
8. All tables use SQLite `STRICT` typing.
9. Missing time components are unknown, not midnight.
10. Timezone labels are optional free text (not a forced IANA/offset enum); empty means unspecified; do not treat them as UTC converters at the storage layer.

This shared model gives Source, Interpretation, and Conclusion data one consistent representation for genealogical dates while allowing each layer to preserve its own evidentiary or interpretive context.
