# First-class Source provenance date

**Status:** idea only — not roadmapped. Parked after de-scoping structured DateValue from Source metadata (Spike 8 **S8-D4** / **S8-07**). Authoritative catalog rule: [`source-layer-data-model.md`](../source-layer-data-model.md) §1.2 and §5.

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

1. **Whether the feature is worth it.** Dogfood has not proven a research workflow that needs “Sources from 1880” as a catalog sort rather than as interpreted event dates.
2. **The attribute’s name and grain.** Publication vs issuance vs registration vs enumeration are different legal moments. The first-class field has to pick one meaning (or an explicit “as-of” that the researcher chooses per Source).
3. **Storage.** A column on `sources` (or a single dedicated FK), not `data_type = 'date'` on `source_metadata_fields`. Schema and DateValue-or-text is decided when a spike pulls this.
4. **UI.** Where it lives on the Source page so it is obvious it is the sort key, not another metadata row.

Until a spike pulls this, metadata date fields remain ordinary text.
