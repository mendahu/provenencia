# First-class Source provenance date

**Status:** idea, planned for Spike 15 with standard citation formats ([`citation-formats.md`](citation-formats.md)), because the Source date is one of the fields an *Evidence Explained* citation needs (user, 2026-10-31). It was pulled into Spike 13 on 2026-10-26 as S13-20 – S13-22, so the Person page's Sources section could show each Source's date, and descoped from it on 2026-10-31 before any of them was planned; the Sources section shows title and type only. Was parked before that after de-scoping structured DateValue from Source metadata (Spike 8 **S8-D4** / **S8-07**). Authoritative catalog rule: [`source-layer-data-model.md`](../source-layer-data-model.md) §1.2 and §5.

## Problem

Researchers may want to group or sort Sources by **when the evidence itself was created or published** — a book’s publication year, a certificate’s issue day, a census enumeration date, a deed’s registration date. That is catalog time (when the record exists in the world), not the interpreted event dates on the Evidence graph.

Source metadata already has type-specific date *labels* for filing and keyword search (`publication_date`, `issue_date`, `record_date`, `census_date`, …). Those keys are not one field. A global “sort by publication date” over EAV rows would miss every Source that filed the same fact under a different key.

## Settled (do not reopen)

| Call | Why |
| --- | --- |
| **Catalog dates stay text on `source_metadata`** | Filing + omnibar search. Structured DateValue on metadata made Source-page entry tedious and did not buy a working filter. |
| **Do not sort the Sources list by metadata date keys** | The printed name of the date depends on the Source type. Unnormalized keys cannot drive a project-wide order. |
| **DateValue stays on Interpretation / Conclusion** | Birth, census day, marriage, death — those dates already have a structured home. Do not rip `date_values` out of Observations. |
| **If provenance-time sort becomes first-class, it is not an EAV key** | One Source-level attribute (name TBD) that every type can fill, independent of whether the researcher also typed “Issue date” or “Enumeration date” in metadata. |

## What this idea still owns

Answered 2026-10-26 (they were Spike 13 working rules until the descope; they carry to Spike 15):

- **Meaning and name:** the original publication date, when the Source was created. Shown as the Source's publication date.
- **Storage:** a structured date, `date_value_id` into the shared `date_values`.
- **Backfill:** none; the field is optional and filled by hand.
- **Sources list:** shows it and sorts by it, undated last.
- **Where it shows:** the Source page (set, change, clear, with the app's structured date entry), the Sources list (each row, and the sort), and the Person page's Sources section (each Source group, R-35).

The original questions, kept for the record:

1. ~~Whether the feature is worth it.~~ Yes: the product owner wants it, first for the Person page's Sources section (2026-10-26).
2. **The attribute’s name and grain.** Publication vs issuance vs registration vs enumeration are different legal moments. The first-class field has to pick one meaning (or an explicit “as-of” that the researcher chooses per Source).
3. **Storage.** A column on `sources` (or a single dedicated FK), not `data_type = 'date'` on `source_metadata_fields`. Schema and DateValue-or-text: answered above (a structured date).
4. **UI.** Where it lives on the Source page so it is obvious it is the sort key, not another metadata row.

Metadata date fields remain ordinary text.

## What Spike 13 had cut

Kept so Spike 15 can re-cut them with the citation fields:

- **Engine:** an optional publication date on `sources`, a `date_value_id` into `date_values` ([`add-date-value`](../../.agents/skills/add-date-value/SKILL.md)): migration, catalog model, read on the Source, the Sources list rows and the Person page's claim groups (`ListPersonClaims`, S13-01), and the write RPC. Existing Sources start empty.
- **Source page:** set, change and clear the date where its brief puts it, so it reads as the Source's own date and not another metadata row; the brief also decides whether the Source page moves onto Page columns ([`wide-layout.md`](wide-layout.md)).
- **Sources list:** each row shows the date; the list sorts by it, undated last; drawn at Wide.
- **Person page:** each Source group in the Sources section shows the date (R-35).

With the citation fields, one migration pass and one Source page design can cover the date and the fields together ([`citation-formats.md`](citation-formats.md) "What a standard citation still needs").
