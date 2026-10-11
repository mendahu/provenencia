# Name Matching Enhancements

**Status:** pulled into [Spike 10](deployment-plan.md). This note is the design for the PRs. What this spike does not build is in [`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md). These come from an adversarial review of the matching engine during S9-10 ([`matching.md`](../../matching.md)). The small, contained fixes from that review shipped in S9-10:
- hyphens as word breaks;
- adjacent-letter swaps;
- short-name insertions;
- date spans and before/after bounds;
- explicit zero settings;
- initials not hiding a surname conflict.

What this spike builds is below. Accents, the Spanish profile, and `patronymic` are in [`international-names.md`](international-names.md).

## Names

- **Abbreviations.** Jas., Wm., Chas., Thos., Geo., Eliz. and similar are everywhere in historical records, and they score as different names today ("Jas." vs "James" gives the given name 0). This needs a seeded abbreviation table (extensible per project, like the nickname tables) that maps an abbreviation to its full name, scored just below an exact match. Edit distance can't learn these.
- **Name frequency.** "John Smith" vs "John Smith" scores 1.0, the same as a rare name, and that's the main source of false suggestions in genealogy. Weight words by how common they are in the project (inverse frequency across all name words), so a shared rare surname counts for more than a shared "Smith". A seeded population table is later ([`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md)).
- **Spaced particles.** "O Brien" vs "O'Brien": the apostrophe joins, so "OBrien" matches but the spaced form is only a partial. A rule that rejoins a lone O, Mc, Mac, St, D, Fitz… with the next word would fix it. It needs care, because the same rule would wrongly join a real initial "O.".
- **Sound-alikes.** Daitch–Mokotoff below the near-spelling tier. It matters for English records too (Smyth/Smith already works by edit distance; Leigh/Lee does not). Beider–Morse is later.

## The scoring model

- **Weigh values by support.** Each feature takes the *best* pair across all values, so:
  - A handle with conflicting values escapes contradictions. One "female" observation among several "male" ones cancels the −8, so a mixed handle scored above a clean one.
  - A stray grouping counts fully: one "James Robins" observation on a handle that's mostly "Mary Smith" scores a full match.

  Weigh each candidate value by its share of support, and judge contradictions against the consensus value. A concluded value dominating the rest waits on Reconciliation Claims ([`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md)).
- **Negative evidence (S9-14).** Only positive observations are loaded. A negative observation ("not born in York") should become a contradiction.
- **Event types and places relate.** Birth vs baptism, death vs burial should partly match. That's a term-affinity table, the same pattern as name roles. Place comparison can use the `part_of` chain already shipped. Flat text stays the strong match.
- **Date precision below the year (S9-21).** Spans and gaps are counted in whole years:
  - December 1817 vs January 1818 is "a year apart" (0.53), lower than May vs September of the same year (0.6).
  - Ranges ignore their months and days: "BET MAR 1817 AND JUN 1817" reads as all of 1817.

  S9-21 stored the windows. This spike measures spans and gaps in days from them.
- **ABT on spans.** "About" doubles the tolerance only when both dates are points. "ABT 1817" against a range or bound uses the normal tolerance.
- **Cross-property dates.** An event's `date` can overlap another's `start_date` / `end_date` by window. Calendars, double dating, and researcher-defined property weights are later ([`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md)).

## The engine

- **Explain the matched values.** Reasons name the two values (or their observation ids) that scored, not only the property.
- **Age-to-birth-year, rejected-claim memory, and blocking** are later ([`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md)). Graph alignment already compares neighbors ([`promote-graph-alignment.md`](../../promote-graph-alignment.md)).

## Decided for this spike

- Name frequency is this project's counts, not a population table.
- Abbreviation and nickname tables are seeded vocabulary a project can extend.
- Support weighting uses share of support, not rank and not claim confidence.
