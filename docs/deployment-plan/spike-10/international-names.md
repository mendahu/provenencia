# International names: accents, scripts, and non-Western name structures

**Status:** pulled into [Spike 10](deployment-plan.md). This note is the design for the PRs. Patterns and scripts this spike does not build are in [`ideas/international-names.md`](../../ideas/international-names.md). Other matcher leftovers are in [`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md). It touches matching ([`matching.md`](../../matching.md)), the name reconciler (S9-13), and the name-format work ([`structured-name-model.md`](../../structured-name-model.md) §4).

## Why

Names are compared, clustered, sorted and displayed with assumptions that hold mainly for Western, Latin-script names:

- **Accents break matches.** "José" vs "Jose" and "Müller" vs "Muller" don't match. `autoreconcile.NormalizeForm` keeps diacritics, and short words fall below the near-spelling floor. Records disagree on accents all the time: clerks, transcribers and OCR drop or add them.
- **Only Western roles exist.** `core/match` compares names word by word with part types as data (S9-10). Mismatched types discount, they don't block. Culture-specific rules live in a `NamePattern` fetched through a `NamePatterns` source, but the only pattern so far is `western`, built in. That's right for "James K. Robins" and incomplete for many others (below).
- **Only one profile exists.** Name format profiles are designed to carry culture, but only `western` is seeded, and nothing reads a profile yet.

## Part 1: accent and character folding

Fold for **comparison**, never for **display**, and don't use one fold for **sorting**.

- **What to fold:**
  - Decompose text (Unicode NFD) and drop combining marks: é→e, ü→u, ñ→n, å→a.
  - A small table for letters that don't decompose: ß→ss, æ→ae, œ→oe, ø→o, ł→l, đ→d, þ→th, ı→i.
  - Needs `golang.org/x/text` (`unicode/norm`) or a hand-written table.
- **Where (decided): matching only.** "José" and "Jose" match in suggestions but stay separate groupings on a Person's page. The auto-reconciler may grow an accent rule later, but merging them is left to manual reconciliation. Folding therefore lives in `core/match`'s word comparison, not in `autoreconcile.NormalizeForm`: the auto-reconciler and the cache `sort_key` don't change, and no cache bump is needed.
- **Exact still beats folded.** "José" vs "José" should outrank "José" vs "Jose". One option: a folded-only match scores 0.95 in `wordSimilarity`, so an accent difference is a near-match, not identity.
- **Sorting is a different problem.** Folding for sort order is wrong in some languages: Swedish files å, ä, ö after z, and Spanish once filed ch and ll as letters. List order wants locale collation (CLDR, via `x/text/collate`), not the comparison fold. **Decided:** the project picks the collation locale (a project setting).
- **Display never folds.** `form` and part values stay exactly as recorded.

## Part 2: non-Western name structures

Matching should take its roles from the Person's **name format profile**, not from fixed part types. This spike covers two of those patterns. The rest are in [`ideas/international-names.md`](../../ideas/international-names.md).

| Pattern | Example | What this spike does |
| --- | --- | --- |
| Spanish dual surnames | *García Márquez* (paternal then maternal) | Both parts stay `surname`. The first weighs more than the second. Records that give only one still match on that one. Portuguese maternal-first is later. |
| Patronymics | Icelandic *Jónsdóttir*; Russian *Ivanovich* | A `patronymic` part type and role, compared to another patronymic. Linking it to the father's given name is later. | |

**Decided: types are data, not locks.** Comparison must work across name formats. A surname in one name that matches a given name in another still connects, just scored lower because the types differ. S9-10 already compares this way (see [`matching.md`](../../matching.md)):
- part types map to roles;
- any word can pair with any word;
- role affinity discounts mismatched types.

What's left here:
- **Profiles supply the pattern.** Each name format profile becomes a `NamePattern`: part type → role (family, given, lineage, generation, ignored), plus role weights, for example a maternal surname weighing less. What's ready:
  - the seam: a `NamePatterns` source on `NameComparer`, today `BuiltinNamePatterns` in `core/match/registry.go`;
  - `Value.NamePattern`, which picks each name's pattern.

  This spike: a source that reads `name_format_profiles`, and loaders that set each name's pattern from the **project default**. A per-person `name_format` claim is later.
- **Which profile applies** when the two names have different formats: each name's words take roles from its own profile, so no single profile has to win. With only the project default, both names share that pattern until per-person format exists.
- **`patronymic`** goes through the compiled `namevalues` registry and [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §4.1. No `maternal_surname` part type. GEDCOM export mapping is later.

## Part 3: sound-alikes and nicknames

- **Sound-alikes:** Daitch–Mokotoff as a `wordSimilarity` tier below near-spelling. Beider–Morse and cross-script transliteration are later ([`ideas/international-names.md`](../../ideas/international-names.md)).
- **Nicknames:** a short seeded table projects can extend (Jim ↔ James, and the set S10-25 names). One table, shared with abbreviations.

## Order of work (rough)

1. Accent folding in matching's word comparison, with the 0.95 folded-match tier. This is small, has the highest payoff, and needs no cache change.
2. Name format profiles carry role maps (and weights); `NameComparer` reads each name's own profile; a `patronymic` part type; a second seeded profile (Spanish dual surname, or Icelandic).
3. Locale collation for sorted lists. The locale is a project setting, default `en`, with no picker. A picker is later.
4. Nickname tables, then Daitch–Mokotoff. Transliteration is later.

## Decisions

- **Accent variants match in suggestions only.** On a Person's page they stay separate groupings. Manual reconciliation merges them; the auto-reconciler may add a rule later.
- **Types are data.** Names in different formats compare word by word, and mismatched types discount rather than block. This shipped in S9-10.
- **The project picks the sorting locale.** This spike stores it and does not offer a control.

The father's-handle patronymic, a collation picker, and a viewer's own locale are in [`ideas/international-names.md`](../../ideas/international-names.md).
