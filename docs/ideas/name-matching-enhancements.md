# Name Matching Enhancements

**Status:** shelved. These come from an adversarial review of the matching engine during S9-10 ([`docs/matching.md`](../matching.md)). The small, contained fixes from that review shipped in S9-10:
- hyphens as word breaks;
- adjacent-letter swaps;
- short-name insertions;
- date spans and before/after bounds;
- explicit zero settings;
- initials not hiding a surname conflict.

What's left is below: improvements to how names are compared, and to the matching engine around them.

Accents, scripts, non-Western name structures and nickname tables are in [`international-names.md`](international-names.md), so they aren't repeated here.

## Names

- **Abbreviations.** Jas., Wm., Chas., Thos., Geo., Eliz. and similar are everywhere in historical records, and they score as different names today ("Jas." vs "James" gives the given name 0). This needs a seeded abbreviation table (extensible per project, like the nickname tables) that maps an abbreviation to its full name, scored just below an exact match. Edit distance can't learn these.
- **Name frequency.** "John Smith" vs "John Smith" scores 1.0, the same as a rare name, and that's the main source of false suggestions in genealogy. Weight words by how common they are in the project (inverse frequency across all name words), so a shared rare surname counts for more than a shared "Smith". Open: whether frequency comes from the project, a seeded population table, or both.
- **Spaced particles.** "O Brien" vs "O'Brien": the apostrophe joins, so "OBrien" matches but the spaced form is only a partial. A rule that rejoins a lone O, Mc, Mac, St, D, Fitz… with the next word would fix it. It needs care, because the same rule would wrongly join a real initial "O.".
- **Sound-alikes.** A phonetic tier (Daitch–Mokotoff, Beider–Morse) below the near-spelling tier. Also listed under international names; it's here because it matters for English records too (Smyth/Smith already works by edit distance, but Leigh/Lee doesn't).
- **Name changes as a signal.** Today, married, religious and anglicized names only connect if one of a handle's names happens to match. A "known as" relationship between names, or using the marriage event once edges exist, could carry it.

## The scoring model

- **Weigh values by support.** Each feature takes the *best* pair across all values, so:
  - A handle with conflicting values escapes contradictions. One "female" observation among several "male" ones cancels the −8, so a mixed handle scored above a clean one.
  - A stray grouping counts fully: one "James Robins" observation on a handle that's mostly "Mary Smith" scores a full match.

  Weigh each candidate value by its share of support (or by rank), and judge contradictions against the consensus value.
- **Concluded values first.** Once Reconciliation Claims ship, a concluded value should dominate, and the other groupings should count less.
- **Negative evidence (S9-14).** Only positive observations are loaded. A negative observation ("not born in York") should become a contradiction.
- **Calibration.** Weights and minimum scores are hand-set. Build a labelled fixture of known-same and known-different pairs (real-world shaped: common names, generations, spelling drift) and tune against it.
  - Later, accepted and rejected claims are labels.
  - Fellegi–Sunter record linkage learns per-feature agree/disagree weights from exactly that, and our additive points already have its shape.
- **Event types and places relate.** Birth vs baptism, death vs burial should partly match. That's a term-affinity table, the same pattern as name roles. Places need their hierarchy and renamings (York → Toronto) from the Place work (S9-25); flat text can't do it.
- **Date precision below the year (S9-21).** Spans and gaps are counted in whole years:
  - December 1817 vs January 1818 is "a year apart" (0.53), lower than May vs September of the same year (0.6).
  - Ranges ignore their months and days: "BET MAR 1817 AND JUN 1817" reads as all of 1817.

  When S9-21 adds date windows, measure spans and gaps in days or months from the windows.
- **ABT on spans.** "About" doubles the tolerance only when both dates are points. "ABT 1817" against a range or bound uses the normal tolerance.
- **Calendars and double dating.** Julian vs Gregorian (the 1752 switch in Britain and its colonies, at different times elsewhere) and "1717/18" double dating are ignored. This belongs with the international work ([`international-names.md`](international-names.md)), since calendars vary by place and church.
- **Cross-property dates.** An event's `date` is never compared with another's `start_date` / `end_date`. When S9-21 adds date windows, compare by window overlap across these properties.
- **Researcher-defined properties.** `ComparerFor` exists, but default profiles only list product properties. A profile could include user properties (occupation, residence text) at a modest default weight, or let a project opt them in.

## The engine

- **Derived and edge features (S9-28 / S9-29).** Features read only a property directly on the probe or the handle. Persons are therefore matched on name and sex alone, and can't tell two men of the same name a generation apart. Needed:
  - **Derived values:** birth year through a participation edge to a birth event; an age converted to a birth year using the census date; a spouse's or parent's name.
  - **A feature key that names a path,** not just a property.
  - **Per-candidate context for `Rank`,** so the walk's related-first ordering is a feature or boost rather than a separate code path.
- **Remember researcher decisions (Spike 10).**
  - A handle whose claim the researcher *rejected* for this Subject is suggested again today.
  - A handle with a *provisional* claim isn't flagged.
  - Merge hints have no memory of "not the same person".

  The claim statuses already exist; matching should exclude or flag those handles. Merge hints need a stored "distinct" decision.
- **Explain the matched values.** Reasons say which property contributed, but not which two values matched. The compare step (S9-19) and the UI's explanation need the pair (probe value and candidate value, or their observation ids).
- **Blocking (S9-34).** Every candidate of the kind is scored. The `loadCandidates` hook should narrow candidates through the search index once catalogs grow.

## Open questions

- Is name frequency per project only, or seeded from population data so small projects get sensible weights?
- Where do abbreviation and nickname tables live: seeded vocabulary that projects extend, or compiled data?
- Should support-weighting use cluster support, rank, or the researcher's claim confidence?
- What does a labelled evaluation set look like before real data exists: synthetic records, or a curated public-domain sample?
