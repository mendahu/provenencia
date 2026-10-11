# International names — later

**Status:** idea, not scheduled. Spike 10 takes accent folding in matching (not display, not the auto-reconciler), spaced particles, a short nickname table, Daitch–Mokotoff, `patronymic` as a part type, a Spanish dual-surname profile, catalog-backed name patterns from the project default, and collation with a default locale and no picker. Design: [`spike-10/international-names.md`](../deployment-plan/spike-10/international-names.md). Abbreviations, project name frequency, and the rest of the matcher are in [`name-matching-enhancements.md`](name-matching-enhancements.md) for the leftovers, and in that same spike for what it builds.

## Scripts and sound

- **Across scripts.** Cyrillic vs Latin (*Иванов* / *Ivanov*), Hebrew or Yiddish vs Latin, and romanization of CJK names (Wade-Giles vs Pinyin). Needs transliteration tables or a library. Hold it until a researcher has mixed-script evidence.
- **Beider–Morse.** Language-aware phonetics. Spike 10 stops at Daitch–Mokotoff, under the near-spelling tier.
- **Calendars and double dating.** Julian vs Gregorian (the 1752 switch in Britain and its colonies, at different times elsewhere) and "1717/18" double dating. Matching ignores both today. They vary by place and church.

## Names Spike 10 does not pattern

Matching still has only the Western pattern plus one Spanish profile (first surname weighs more than the second; both stay `surname`). Still wanted:

| Pattern | Example | What matching needs |
| --- | --- | --- |
| Portuguese dual surnames | Maternal name first | The opposite weight order from the Spanish profile. |
| Patronymic as a family link | Icelandic *Jónsdóttir*; Russian *Ivanovich* | Spike 10 compares the part to another patronymic. Linking it to the father's given name needs the father's handle. |
| Farm and locational names | Norwegian *Haugen*, which changed when the family moved | A surname that is evidence of place, not lineage. Weak as a family role. |
| Surname-first order | Chinese, Japanese, Korean, Hungarian, Vietnamese | Order is display if parts are typed. Profiles should drive entry, not only display, or data entry will type by position. |
| Name chains | Arabic *ibn* / *bint*; Hebrew *ben* / *bat* | Several generations in one name. Compare the chain in order. Particles are joiners, not words. |
| Particle sort | Dutch *van der*; German *von*; Portuguese *da* | Matching already ignores them. Sorting needs a per-profile rule: *van Gogh* under V or G. |
| Single names | Mononyms; enslaved people recorded by given name only | A missing role is already left out. Context (owner, place, date) has to carry more of the weight, on the profile. |

`maternal_surname` as its own part type is not seeded. GEDCOM has no patronymic type, so export needs a mapping when export exists.

## Format and sort, beyond the project default

- **Per-person name format.** A Reconciliation Claim on `name_format`. Until that exists, every name uses the project default profile. Each name already may carry its own pattern once a loader sets `Value.NamePattern`.
- **Accent merging on the Person page.** José and Jose match in suggestions and stay separate groupings. An auto-reconciler rule that folds them is later, and it is manual reconciliation until then.
- **Collation picker.** The project has a locale setting (default `en`) and no control. A viewer who wants a locale other than the project's (a shared tree read abroad) is the same gap.
