# International names — later

**Status:** idea, not scheduled. Spike 10 takes accent folding in matching (not display, not the auto-reconciler), spaced particles, and a short nickname table. Collation was descoped (S10-37, 2026-10-16) and is below. Daitch–Mokotoff was descoped (S10-26); the design is [`daitch-mokotoff.md`](daitch-mokotoff.md). S10-29 installed the Western name-format profile and reads it from the catalog. Spanish dual surnames and a `patronymic` part are unscheduled again, below. Design: [`spike-10/international-names.md`](../deployment-plan/spike-10/international-names.md). Abbreviations, project name frequency, and the rest of the matcher are in [`name-matching-enhancements.md`](name-matching-enhancements.md) for the leftovers, and in that same spike for what it builds.

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

`maternal_surname` as its own part type is not seeded. GEDCOM has no patronymic type, so export needs a mapping when export exists.

## Format and sort, beyond the project default

- **Per-person name format.** A Reconciliation Claim on `name_format`. Until that exists, every name uses the project default profile. Each name already may carry its own pattern once a loader sets `Value.NamePattern`.
- **Accent merging on the Person page.** José and Jose match in suggestions and stay separate groupings. An auto-reconciler rule that folds them is later, and it is manual reconciliation until then.
- **Collation locale.** A project setting (default `en`) whose locale orders name lists through `golang.org/x/text/collate`, so accented and unaccented names sort in that locale's order. Comparison fold is not the sort, and the match score does not change. Descoped from S10-37 (2026-10-16); nothing stores a locale today.
- **Collation picker.** A control for that setting. A viewer who wants a locale other than the project's (a shared tree read abroad) is the same gap.
