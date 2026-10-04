# International names: accents, scripts, and non-Western name structures

**Status:** idea, not scheduled. Other matching improvements (abbreviations, name frequency, support weighting, derived features) are in [`name-matching-enhancements.md`](name-matching-enhancements.md). This is the start of an internationalization plan for names. It touches matching ([`docs/matching.md`](../matching.md)), the name reconciler (S9-13), and the name-format work ([`structured-name-model.md`](../structured-name-model.md) §4).

## Why

Names are compared, clustered, sorted and displayed with assumptions that hold mainly for Western, Latin-script names:

- **Accents break matches.** "José" vs "Jose" and "Müller" vs "Muller" don't match. `resolve.NormalizeForm` keeps diacritics, and short words fall below the near-spelling floor. Records disagree on accents all the time: clerks, transcribers and OCR drop or add them.
- **Only Western roles exist.** `core/match` compares names word by word with part types as data (S9-10). Mismatched types discount, they don't block. Culture-specific rules live in a `NamePattern` fetched through a `NamePatterns` source, but the only pattern so far is `western`, built in. That's right for "James K. Robins" and incomplete for many others (below).
- **Only one profile exists.** Name format profiles are designed to carry culture, but only `western` is seeded, and nothing reads a profile yet.

## Part 1: accent and character folding

Fold for **comparison**, never for **display**, and don't use one fold for **sorting**.

- **What to fold:**
  - Decompose text (Unicode NFD) and drop combining marks: é→e, ü→u, ñ→n, å→a.
  - A small table for letters that don't decompose: ß→ss, æ→ae, œ→oe, ø→o, ł→l, đ→d, þ→th, ı→i.
  - Needs `golang.org/x/text` (`unicode/norm`) or a hand-written table.
- **Where (decided): matching only.** "José" and "Jose" match in suggestions but stay separate groupings on a Person's page. The auto-reconciler may grow an accent rule later, but merging them is left to manual reconciliation. Folding therefore lives in `core/match`'s word comparison, not in `resolve.NormalizeForm`: cluster keys and the cache `sort_key` don't change, and no cache bump is needed.
- **Exact still beats folded.** "José" vs "José" should outrank "José" vs "Jose". One option: a folded-only match scores 0.95 in `wordSimilarity`, so an accent difference is a near-match, not identity.
- **Sorting is a different problem.** Folding for sort order is wrong in some languages: Swedish files å, ä, ö after z, and Spanish once filed ch and ll as letters. List order wants locale collation (CLDR, via `x/text/collate`), not the comparison fold. **Decided:** the project picks the collation locale (a project setting).
- **Display never folds.** `form` and part values stay exactly as recorded.

## Part 2: non-Western name structures

Matching should take its roles from the Person's **name format profile**, not from fixed part types. Patterns to cover:

| Pattern | Example | What matching needs |
| --- | --- | --- |
| Spanish / Portuguese dual surnames | *García Márquez* (paternal then maternal); Portuguese puts the maternal name first | The paternal surname weighed above the maternal one, with the order coming from the profile. Records often give only one of them. |
| Patronymics | Icelandic *Jónsdóttir*; historical Scandinavian *Andersen*; Russian *Ivanovich* | A `patronymic` role (no part type exists yet), compared as a link to the father's given name rather than as a surname. The family name changes every generation. |
| Farm and locational names | Norwegian *Haugen*, which changed when the family moved | A surname that is evidence of place, not lineage. Weak as a family role. |
| Surname-first order | Chinese, Japanese, Korean, Hungarian, Vietnamese | Order is only display, if parts are typed. The risk is data entry that types by position; profiles should drive entry, not just display. |
| Name chains | Arabic *ibn* / *bint*; Hebrew *ben* / *bat* | Several generations in one name. Compare the chain in order, with particles as joiners and not words. |
| Surname particles | Dutch *van der*; German *von*; Portuguese *da* | Already ignored in matching. Sorting needs a per-profile rule: *van Gogh* under V or G. |
| Name changes | Married names, religious names, anglicized immigrant names | Not a comparer rule. They're multiple name Observations, and matching already takes the best pair. A "known as" relationship could help later. |
| Single names | Mononyms; enslaved people recorded by given name only | Work today: a role neither name has is left out. Context (owner, place, date) must carry more of the weight; that belongs in the profile, not the name. |

**Decided: types are data, not locks.** Comparison must work across name formats. A surname in one name that matches a given name in another still connects, just scored lower because the types differ. S9-10 already compares this way (see [`docs/matching.md`](../matching.md)):
- part types map to roles;
- any word can pair with any word;
- role affinity discounts mismatched types.

What's left here:
- **Profiles supply the pattern.** Each name format profile becomes a `NamePattern`: part type → role (family, given, lineage, generation, ignored), plus role weights, for example a maternal surname weighing less. What's ready:
  - the seam: a `NamePatterns` source on `NameComparer`, today `BuiltinNamePatterns` in `core/match/registry.go`;
  - `Value.NamePattern`, which picks each name's pattern.

  What's left: a source that reads `name_format_profiles`, and loaders that set each name's pattern from its Person's name format.
- **Which profile applies** when the two names have different formats: each name's words take roles from its own profile, so no single profile has to win.
- **New part types** (`patronymic`, perhaps `maternal_surname`) go through the compiled `namevalues` registry and [`seeded-vocabulary.md`](../seeded-vocabulary.md) §4.1. GEDCOM has no patronymic type, so export needs a mapping. A `lineage` role might pair a patronymic with the father's given name, at a discount.

## Part 3: scripts, transliteration, sound-alikes

- **Across scripts:** Cyrillic vs Latin (*Иванов* / *Ivanov*), Hebrew or Yiddish vs Latin, and romanization of CJK names (Wade-Giles vs Pinyin). This needs transliteration tables or a library. Hold it until a researcher actually has mixed-script evidence.
- **Sound-alikes:** Daitch–Mokotoff Soundex suits Eastern European and Jewish names better than American Soundex; Beider–Morse is language-aware. Either could be a `wordSimilarity` tier below the near-spelling one.
- **Nicknames:** equivalence tables per language (Jim ↔ James, Hans ↔ Johann, Pepe ↔ José), ideally seeded vocabulary so projects can extend them.

## Order of work (rough)

1. Accent folding in matching's word comparison, with the 0.95 folded-match tier. This is small, has the highest payoff, and needs no cache change.
2. Name format profiles carry role maps (and weights); `NameComparer` reads each name's own profile; a `patronymic` part type; a second seeded profile (Spanish dual surname, or Icelandic).
3. Locale collation for sorted lists, with the locale as a project setting, alongside the name-format display styles.
4. Nickname tables, then phonetics, then transliteration.

## Decisions

- **Accent variants match in suggestions only.** On a Person's page they stay separate groupings. Manual reconciliation merges them; the auto-reconciler may add a rule later.
- **Types are data.** Names in different formats compare word by word, and mismatched types discount rather than block. This shipped in S9-10.
- **The project picks the sorting locale.**

## Open questions

- Should a patronymic link actually use the father's handle, once family edges exist (S9-28)?
- Does a viewer ever need a locale other than the project's, for example a shared tree read abroad?
