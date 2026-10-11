# Name matching — later

**Status:** idea, not scheduled. Spike 10 takes abbreviations, project-only name frequency, spaced particles, Daitch–Mokotoff, support-weighted scores, negative observations as contradictions, reasons that name the two values, date windows in the comparer, and birth~baptism / death~burial plus place chains. Design: [`spike-10/name-matching-enhancements.md`](../deployment-plan/spike-10/name-matching-enhancements.md).

The small S9-10 fixes already shipped: hyphens as word breaks, adjacent-letter swaps, short-name insertions, date spans and before/after bounds, explicit zero settings, initials not hiding a surname conflict.

## Names

- **Population frequency.** Spike 10 weighs words by how common they are in this project. A seeded population table, so a small project does not treat every surname as rare, is still open. The project weight needs a floor either way; that floor is part of S10-27, not this note.
- **Name changes as a signal.** Married, religious, and anglicized names connect today only when one of a handle's names happens to match. A "known as" relationship between names, or the marriage event, could carry it. Multiple name Observations and taking the best pair already exist.

## Scoring

- **Concluded values first.** Once Reconciliation Claims ship, a concluded value should dominate and the other groupings should count less. Spike 10 weighs by share of support only.
- **Calibration.** Weights and minimum scores are hand-set. A labelled fixture of known-same and known-different pairs (common names, generations, spelling drift) would tune them. Later, accepted and rejected claims are labels. Fellegi–Sunter learns per-feature agree/disagree weights from that shape.
- **Researcher-defined properties.** `ComparerFor` exists. Default profiles only list product properties. A profile could include user properties (occupation, residence text) at a modest default weight, or let a project opt them in.

## Engine

- **An age converted to a birth year** from the census date, as a derived value on the Subject. Graph alignment already compares neighbors. This is the derived feature that alignment does not replace.
- **Remember researcher decisions.** A handle the researcher rejected for this Subject is suggested again. A provisional claim is not flagged. Merge hints have no memory of "not the same person." Claim statuses exist. Matching should exclude or flag those handles. Merge hints need a stored "distinct" decision. This waits with claim editing ([`identity-claim-review.md`](identity-claim-review.md)).
- **Blocking.** Every candidate of the kind is scored. `loadCandidates` should narrow through the search index once catalogs grow.

## Open

- What a labelled evaluation set looks like before real projects exist: synthetic records, or a curated public-domain sample.
