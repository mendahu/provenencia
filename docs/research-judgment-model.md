# Provenencia Genealogy — Research Judgment Model

## Status

Draft architecture notes. This document is authoritative for how Provenencia persists researcher judgment about **Source credibility**, **Citation transcription certainty**, and **Claim confidence**, and for what remains derived UI rather than stored data.

Layer schemas remain in:

- Source: [`source-layer-data-model.md`](source-layer-data-model.md)
- Interpretation: [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md)
- Conclusion: [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md)

Seed keys: [`seeded-vocabulary.md`](seeded-vocabulary.md). Audit: [`audit-revision-history.md`](audit-revision-history.md).

---

# 1. Design philosophy

## 1.1 Store judgment; derive display pills

Persist only researcher judgments that cannot be recomputed from the evidence trail.

Do **not** store a single subject- or fact-level “likelihood” pill. Product UI may synthesize badges from stored judgments, exhibit shape (pins, polarity mix), claim `status`, and handle grounding. Those pills are recomputable app policy and must not be written back as catalog truth.

Do **not** auto-multiply Source credibility × Citation certainty × Claim confidence into a stored rollup. The claim editor should **surface** Source assessments and Citation flags as context when the researcher sets Claim confidence.

## 1.2 Three distinct questions

| Question | Axis | Attachment |
|---|---|---|
| How much do I trust this evidence? | Source **credibility** | Dedicated Interpretation assessment row (not a column on `sources`) |
| How faithfully does this Citation represent the cited media? | Citation **transcription certainty** | Light columns on `citations` |
| How sure am I of this conclusion? | Claim **confidence** | Columns on Sameness and Reconciliation Claims |

These are related in research practice but must not share vocabulary keys. Credibility is about trusting evidence; confidence is about standing behind a conclusion. Both use a **three-point scale with a baseline middle**, with separate labels and grade tables.

## 1.3 What this model deliberately excludes

- **No confidence score on Observations.** Competing or alternate readings are separate Observation rows (and polarity). See [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md).
- **No credibility column on `sources`.** The Source layer answers what evidence we possess. Judging evidentiary value is Interpretation. Putting a grade on the catalog row is a layer leak.
- **No claim polarity.** Polarity (`positive` / `negative`) lives on Observations. Sameness uses `relation` (`same_as` / `distinct_from`). Reconciliation stores a concluded value.
- **Disputed / undocumented are not stored confidence grades.** Prefer derivation from exhibit conflict, empty pins, handle grounding badges, and claim `status`. An optional explicit conflict flag may be added later if derivation is too noisy; it remains separate from confidence.

## 1.4 Authorship and changing your mind

Sameness Claims, Reconciliation Claims, Source credibility assessments, and Citation fields follow the existing **single working-state** pattern: one row per slot; updates change that row; [`audit-revision-history.md`](audit-revision-history.md) records who/when/old/new.

That covers one researcher revising a grade and sequential multi-contributor edits (last write wins for working state). It does **not** model concurrent live peer opinions (Alice high and Bob low both standing). Standing multi-opinion assessments are out of scope until a concrete collaboration workflow requires them; audit history is not a substitute for that domain concept.

---

# 2. Source credibility assessments

## 2.1 Role

A Source credibility assessment records how much the researcher trusts a Source **as evidence in general**. It is reusable across many Claims and may be set during intake before any Claim exists.

Credibility is often claim-relative in practice (a census may be strong for household composition and weak for an exact birth year). Claim confidence absorbs that locality. Source credibility remains the reusable baseline impression of the Source itself.

## 2.2 Layer and table

Assessments are **Interpretation** data: judgment about evidentiary value, not Source catalog possession, and not a third Conclusion claim verb about historical Persons/Events/Places.

```sql
CREATE TABLE source_credibility_grades (
    key         TEXT PRIMARY KEY,
    label       TEXT NOT NULL,
    sort_order  INTEGER NOT NULL,
    builtin     INTEGER NOT NULL DEFAULT 1,
    CHECK (builtin IN (0, 1))
) STRICT;

CREATE TABLE source_credibility_assessments (
    id                  BLOB PRIMARY KEY,
    source_id           BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    credibility_key     TEXT NOT NULL REFERENCES source_credibility_grades(key),
    argument            TEXT,

    UNIQUE (source_id)
) STRICT;
```

There is **at most one working assessment per Source**. Changing the grade updates that row (and is audited).

Optional later: a nullable link to the reifying `source` Node for graph navigation. Cited commentary about a Source (another Source challenges authenticity, remarks, `mentions`) continues to use Observations on the Source Node as today; that path does not replace this assessment table.

## 2.3 Three-point scale (separate vocabulary)

Seeded keys (exact product copy TBD; keys must stay credibility-specific):

```text
low_trust       -- below baseline; soft, hearsay, or otherwise weakly trusted
standard        -- baseline / ordinary trust (default mental model)
high_trust      -- above baseline; evidence the researcher strongly trusts
```

Missing assessment: UI may treat the Source as `standard` for display without inserting a row. Casual projects need not rate every Source.

Do **not** reuse Claim confidence keys here.

## 2.4 Not on the Source catalog row

[`source-layer-data-model.md`](source-layer-data-model.md) stays free of genealogical and epistemic judgment columns. `source_notes` and descriptive metadata are not substitutes for structured credibility.

---

# 3. Citation transcription certainty

## 3.1 Role

Records whether the Citation’s transcription (or equivalent representation of spoken/visual content) may **not faithfully represent** the cited Artifact portion. This is media-agnostic: faint ink, damaged scan, garbled audio, and muddy video are the same class of problem.

This is **not** Claim confidence and **not** Source credibility. A clear transcription of a weakly trusted Source is possible; an uncertain transcription of a highly trusted Source is possible. Partial uncertainty (one surname damaged; rest clear) often matters for some Claims and not others — Claim confidence absorbs that locality, with detail in `transcription_note` and in-line uncertainty marks in `transcription`.

## 3.2 Columns on `citations`

Not a dedicated table. Extend Citations:

```sql
-- additions to citations (see interpretation-layer-data-model.md)
transcription_uncertain INTEGER NOT NULL DEFAULT 0,  -- 0 | 1
transcription_note      TEXT,

CHECK (transcription_uncertain IN (0, 1))
```

`transcription_uncertain = 1` means the researcher marks the transcription as low-fidelity / uncertain relative to the cited media.

`transcription_note` is optional free text (“garbled audio ~1:02”; “surname damaged”).

Existing `transcription` and `description` remain. In-line marks such as `[?]` in transcription remain valid and complementary.

Do **not** put a rich ordinal confidence ladder on Citations. A boolean (plus note) is enough; epistemic weight belongs on the Claim.

---

# 4. Claim confidence

## 4.1 Role

Records how sure the researcher is of **this** conclusion (this Sameness or Reconciliation Claim), given its exhibit and reasoning.

Confidence is claim-relative by design. It is the primary structured epistemic field for Conclusions.

## 4.2 Orthogonal to claim `status`

| Field | Values | Answers |
|---|---|---|
| `status` | `provisional` \| `accepted` \| `rejected` | Is this in the working graph / committed conclusion? |
| `confidence_key` | three-point confidence vocabulary | How sure am I of this conclusion? |

They are **not** the same thing. Valid combinations include accepted + low confidence (committed weak best guess) and provisional + high confidence (strong candidate not yet committed). Do not overload `provisional` to mean “I am unsure.”

## 4.3 Columns on Claims

```sql
-- additions to sameness_claims and reconciliation_claims
confidence_key  TEXT REFERENCES claim_confidence_grades(key),  -- nullable
```

```sql
CREATE TABLE claim_confidence_grades (
    key         TEXT PRIMARY KEY,
    label       TEXT NOT NULL,
    sort_order  INTEGER NOT NULL,
    builtin     INTEGER NOT NULL DEFAULT 1,
    CHECK (builtin IN (0, 1))
) STRICT;
```

`confidence_key` is **nullable** so casual trees and quick Claims need not set a grade. `argument` remains the free-text reasoning chain; confidence is the structured, filterable stance.

There remains at most one Sameness Claim per Node pair and one Reconciliation Claim per `(entity, property)`. Changing confidence updates that row (and is audited).

## 4.4 Three-point scale (separate vocabulary)

Seeded keys (exact product copy TBD; keys must stay confidence-specific):

```text
low_confidence     -- below baseline; weak or shaky conclusion
moderate           -- baseline / ordinary confidence
high_confidence    -- above baseline; strongly trusted conclusion
```

Same **shape** as Source credibility (below / baseline / above); **different keys and labels**. Do not share grade rows between the two scales.

## 4.5 Polarity and conflict in the exhibit

Observation polarity factors into Claims only through the exhibit and `argument`:

- Mixed positive and negative pinned Observations may justify lower confidence and a contested UI badge.
- Sameness `relation` (`same_as` / `distinct_from`) is the claim’s content stance for identity — not Observation polarity copied onto the claim.
- Prefer **deriving** “disputed” / “contested exhibit” and “undocumented” (no pins) for display rather than storing them as confidence grades.

---

# 5. Product composition (informative)

When authoring a Claim, the UI should make it easy to see:

1. Pinned Observations (exhibit)
2. Each pin’s Citation (`transcription_uncertain`, note)
3. Each Citation’s root Source and any `source_credibility_assessments` row

The researcher then sets `confidence_key` deliberately. Derived subject/fact pills may later combine claim confidence, `status`, exhibit shape, and optional display hints from Source/Citation judgments — without persisting that synthesis.

Handle provenencia badges (from records / inferred / asserted / unlinked) remain a separate computed grounding story; see [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

---

# 6. Architectural rules

1. Persist researcher judgment at Source assessment, Citation, and Claim attachment points only as defined here.
2. Do not store subject/fact likelihood pills or automatic Source × Citation × Claim rollups.
3. Do not put credibility on `sources` or confidence scores on Observations.
4. Source credibility and Claim confidence use parallel three-point shapes with **separate** vocabularies.
5. Citation transcription certainty is a boolean (+ optional note), media-agnostic, columns on `citations`.
6. Claim `status` and Claim `confidence_key` are orthogonal.
7. Observation polarity is not Claim polarity.
8. Disputed / undocumented display states are derived unless a later explicit flag is justified.
9. Working state is single-row per slot; history and attribution are audit.
10. Grade keys are open vocabulary tables (seeded builtins), not SQL enums on every domain CHECK — except Citation’s `0|1` flag.

---

# 7. Documentation ownership

- This document is authoritative for research judgment semantics and the assessment/grade schemas above.
- [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md) owns Citation columns and hosts the Source credibility assessment tables in the Interpretation schema surface.
- [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) owns Claim tables including `confidence_key`.
- [`seeded-vocabulary.md`](seeded-vocabulary.md) lists intended grade keys and labels.
- [`source-layer-data-model.md`](source-layer-data-model.md) remains free of credibility columns.
- [`audit-revision-history.md`](audit-revision-history.md) remains authoritative for revision attribution.
