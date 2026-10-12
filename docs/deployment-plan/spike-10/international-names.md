# International names: accents, scripts, and non-Western name structures

**Status:** pulled into [Spike 10](deployment-plan.md). This note is the design for the PRs. Patterns and scripts this spike does not build are in [`ideas/international-names.md`](../../ideas/international-names.md). Other matcher leftovers are in [`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md). It touches matching ([`matching.md`](../../matching.md)), the name reconciler (S9-13), and the name-format work ([`structured-name-model.md`](../../structured-name-model.md) §4).

## Why

Names are compared, clustered, sorted and displayed with assumptions that hold mainly for Western, Latin-script names:

- **Accents.** S10-24 scores a fold-only pair at 0.95 in `wordSimilarity`. `autoreconcile.NormalizeForm` still keeps diacritics, so José and Jose stay separate groupings on a Person page.
- **Only Western roles exist.** `core/match` compares names word by word with part types as data (S9-10). Mismatched types discount, they don't block. Culture-specific rules live in a `NamePattern` fetched through a `NamePatterns` source, but the only pattern so far is `western`, built in. That's right for "James K. Robins" and incomplete for many others (below).
- **The catalog holds one profile.** S10-29 installs `western` and matching reads those rows. Further cultures are unscheduled.

## Part 1: accent and character folding

Fold for **comparison**, never for **display**, and don't use one fold for **sorting**.

- **What to fold:**
  - Decompose text (Unicode NFD) and drop combining marks: é→e, ü→u, ñ→n, å→a.
  - `LetterFolds` in `core/match/registry.go` for letters that don't decompose: ß→ss, æ→ae, œ→oe, ø→o, ł→l, đ→d, þ→th, ı→i. Adding a letter extends comparison.
  - Needs `golang.org/x/text` (`unicode/norm`) or a hand-written table.
- **Where (decided): matching only.** "José" and "Jose" match in suggestions but stay separate groupings on a Person's page. The auto-reconciler may grow an accent rule later, but merging them is left to manual reconciliation. Folding therefore lives in `core/match`'s word comparison, not in `autoreconcile.NormalizeForm`: the auto-reconciler and the cache `sort_key` don't change, and no cache bump is needed.
- **Exact still beats folded.** S10-24 scores a folded-only match at 0.95 in `wordSimilarity`. "José" vs "José" stays 1.
- **Sorting is a different problem.** Folding for sort order is wrong in some languages: Swedish files å, ä, ö after z, and Spanish once filed ch and ll as letters. List order wants locale collation (CLDR, via `x/text/collate`), not the comparison fold. **Decided:** the project picks the collation locale (a project setting).
- **Display never folds.** `form` and part values stay exactly as recorded.

## Part 2: non-Western name structures

Matching takes its roles from the project's **name format profile**. S10-29 seeds `western` only. Spanish dual surnames and a patronymic part are unscheduled, with the rest, in [`ideas/international-names.md`](../../ideas/international-names.md).

**Decided: types are data, not locks.** Comparison must work across name formats. A surname in one name that matches a given name in another still connects, just scored lower because the types differ. S9-10 already compares this way (see [`matching.md`](../../matching.md)):
- part types map to roles;
- any word can pair with any word;
- role affinity discounts mismatched types.

What's in place:
- **Profiles supply the pattern.** S10-29 reads `name_format_profiles` (roles and weights) and sets each name's `Value.NamePattern` from the project default. A catalog with no default still uses the built-in Western pattern. A per-person `name_format` claim is later.
- **Which profile applies** when the two names have different formats: each name's words take roles from its own profile. With only the project default, both names share that pattern until per-person format exists. No `maternal_surname` part type.

## Part 3: sound-alikes and nicknames

- **Sound-alikes:** Daitch–Mokotoff as a `wordSimilarity` tier below near-spelling. Beider–Morse and cross-script transliteration are later ([`ideas/international-names.md`](../../ideas/international-names.md)).
- **Nicknames:** S10-25 scores a table hit at 0.98, above the fold. The product seed is `productAliasTables` in `core/match/registry.go`, written at catalog create. Scoring reads the catalog only. Opening a project loads those rows once and keeps them until the project closes; `namealiases.Invalidate` reloads them. A project can add a `user` or `plugin:<id>` table. Every installed table is consulted. No editor UI.

## Order of work (rough)

1. Accent folding in matching's word comparison, with the 0.95 folded-match tier. S10-24 did this. It needs no cache change.
2. Name format profiles carry role maps and weights. S10-29 did this for `western`. Further profiles are unscheduled.
3. Locale collation for sorted lists. The locale is a project setting, default `en`, with no picker. A picker is later.
4. Nickname tables, then Daitch–Mokotoff. Transliteration is later.

## Decisions

- **Accent variants match in suggestions only.** On a Person's page they stay separate groupings. Manual reconciliation merges them; the auto-reconciler may add a rule later.
- **Types are data.** Names in different formats compare word by word, and mismatched types discount rather than block. This shipped in S9-10.
- **The project picks the sorting locale.** This spike stores it and does not offer a control.

Spanish dual surnames, a patronymic part, the father's-handle link, a collation picker, and a viewer's own locale are in [`ideas/international-names.md`](../../ideas/international-names.md).
