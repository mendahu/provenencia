# International names — later

**Status:** idea, not scheduled. Spike 10 shipped accent folding in matching, spaced particles, nickname tables, and the Western profile ([decisions](../deployment-plan/archive/spike-10/)). Daitch–Mokotoff: [`daitch-mokotoff.md`](daitch-mokotoff.md).

## Scripts and sound

- **Across scripts.** Cyrillic vs Latin (*Иванов* / *Ivanov*), Hebrew or Yiddish vs Latin, and romanization of CJK names (Wade-Giles vs Pinyin). Needs transliteration tables or a library. Hold it until a researcher has mixed-script evidence.
- **Beider–Morse.** Language-aware phonetics. More precise than Daitch–Mokotoff, and a larger rule set. Still later. The Daitch–Mokotoff design, descoped from S10-26, is [`daitch-mokotoff.md`](daitch-mokotoff.md).
- **Calendars and double dating.** Julian vs Gregorian (the 1752 switch in Britain and its colonies, at different times elsewhere) and "1717/18" double dating. Matching ignores both today. They vary by place and church.

## Names this product does not pattern yet

S10-29 seeds `western` only. Both surname words keep the single family weight. Still wanted:

| Pattern | Example | What matching needs |
| --- | --- | --- |
| Spanish dual surnames | *García Márquez* (paternal then maternal) | Both parts stay `surname`. The first weighs more than the second. A new weight slot, then a profile row. |
| Portuguese dual surnames | Maternal name first | The opposite weight order from the Spanish profile. |
| Patronymic | Icelandic *Jónsdóttir*; Russian *Ivanovich* | A `patronymic` part type and role, compared to another patronymic. Linking it to the father's given name needs the father's handle. |
| Farm and locational names | Norwegian *Haugen*, which changed when the family moved | A surname that is evidence of place, not lineage. Weak as a family role. |
| Surname-first order | Chinese, Japanese, Korean, Hungarian, Vietnamese | Order is display if parts are typed. Profiles should drive entry, not only display, or data entry will type by position. |
| Name chains | Arabic *ibn* / *bint*; Hebrew *ben* / *bat* | Several generations in one name. Compare the chain in order. Particles are joiners, not words. |
| Particle sort | Dutch *van der*; German *von*; Portuguese *da* | Matching already ignores them. Sorting needs a per-profile rule: *van Gogh* under V or G. |
| Single names | Mononyms; enslaved people recorded by given name only | A missing role is already left out. Context (owner, place, date) has to carry more of the weight, on the profile. |

A `generation` weight slot, and one declared slot list read by both `nameformats.LoadID` and `seedProfiles`, would replace the registry's `value.Names.Generation` knob (S11-03 kept it a knob).

`maternal_surname` as its own part type is not seeded. GEDCOM has no patronymic type, so export needs a mapping when export exists.

## Name changes

- **Name changes as a signal.** Married, religious and anglicized names connect today only when one of a handle's names matches. A "known as" relationship between names, or the marriage event, could carry it. Spike 11 treats Mary Smith / Mary Robins as `Different` at the value level, and leaves marriage evidence to links. Maiden names in the conclusion are the intended path ([`comparison-layers.md`](../deployment-plan/archive/spike-11/comparison-layers.md#entities-compare-against-the-conclusion)).

## Format and sort, beyond the project default

- **Per-person name format.** A Reconciliation Claim on `name_format`. Until that exists, every name uses the project default profile. Each name already may carry its own pattern once a loader sets `Value.NamePattern`.
- **Accent merging on the Person page** — decided 2026-10-17 (S11-06): an accent fold is value compare's `Equivalent`, which never groups or folds, so *José* and *Jose* show as two values.
- **Collation locale.** A project setting (default `en`) whose locale orders name lists through `golang.org/x/text/collate`, so accented and unaccented names sort in that locale's order. Comparison fold is not the sort, and the match score does not change. Descoped from S10-37 (2026-10-16); nothing stores a locale today.
- **Collation picker.** A control for that setting. A viewer who wants a locale other than the project's (a shared tree read abroad) is the same gap.

## Whole-phrase copy, app-wide

S11-33 made every sentence on the Promote evidence sheet one whole translatable phrase (see `add-localized-string`, "Whole phrases"): an argument is data or a complete noun phrase from one table entry, never a bare translated word the sentence must agree with. Other copy still splices a translated word into a sentence, so a language that inflects or agrees an article cannot translate it: the `EventTitle` formats (`L10n.EventTitle`), `evidenceGraph.bridge.summary.*` (a term label inside "… participated as … at …"), the `CitationComposer` term formats, `sources.sortedBy`, and `conclusions.outcomeOutvotedBy` (a count and a Sources word). Each becomes a key per sentence, or takes its word whole from `L10n.PropertyTerm.wording`, when a second language is added and the copy is next touched.
