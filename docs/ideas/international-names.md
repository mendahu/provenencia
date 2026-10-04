# International names: accents, scripts, and non-Western name structures

**Status:** idea, not scheduled. This is the start of an internationalization plan for names. It touches matching ([`docs/matching.md`](../matching.md)), the name reconciler (S9-13), and the name-format work ([`structured-name-model.md`](../structured-name-model.md) §4).

## Why

Names are compared, clustered, sorted and displayed with assumptions that hold mainly for Western, Latin-script names:

- **Accents break matches.** "José" vs "Jose" and "Müller" vs "Muller" don't match. `resolve.NormalizeForm` keeps diacritics, and short words fall below the near-spelling floor. Records disagree on accents all the time: clerks, transcribers and OCR drop or add them.
- **Matching roles are hard-coded.** `core/match` treats every name the same way: `surname` is the family name, the first `given` leads, and `suffix` separates generations. That is right for "James K. Robins" and wrong or incomplete for many others (below).
- **Only one profile exists.** Name format profiles are designed to carry culture, but only `western` is seeded, and nothing reads a profile yet.

## Part 1: accent and character folding

Fold for **comparison**, never for **display**, and don't use one fold for **sorting**.

- **What to fold:**
  - Decompose text (Unicode NFD) and drop combining marks: é→e, ü→u, ñ→n, å→a.
  - A small table for letters that don't decompose: ß→ss, æ→ae, œ→oe, ø→o, ł→l, đ→d, þ→th, ı→i.
  - Needs `golang.org/x/text` (`unicode/norm`) or a hand-written table.
- **Where:** in one shared normalizer, so the cluster key, the cache `sort_key`, search, and matching can't disagree. Today that normalizer is `resolve.NormalizeForm`; folding there changes cached keys, so it needs a cache version bump.
- **Exact still beats folded.** "José" vs "José" should outrank "José" vs "Jose". One option: a folded-only match scores 0.95 in `wordSimilarity`, so an accent difference is a near-match, not identity.
- **Sorting is a different problem.** Folding for sort order is wrong in some languages: Swedish files å, ä, ö after z, and Spanish once filed ch and ll as letters. List order wants locale collation (CLDR, via `x/text/collate`), chosen by the project or the Person's name format, not the comparison fold.
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

Sketch:
- **Profiles also define roles.** Each name format profile maps part types to roles: family (with an order or weight), given, lineage (patronymic), generation (suffix), ignored.
- **The comparer takes roles.** `NameComparer` receives the role map from the candidate's profile instead of the Western grouping hard-coded in `rolesOf`. A Western default keeps today's behaviour.
- **New part types** (`patronymic`, perhaps `maternal_surname`) go through the compiled `namevalues` registry and [`seeded-vocabulary.md`](../seeded-vocabulary.md) §4.1. GEDCOM has no patronymic type, so export needs a mapping.

## Part 3: scripts, transliteration, sound-alikes

- **Across scripts:** Cyrillic vs Latin (*Иванов* / *Ivanov*), Hebrew or Yiddish vs Latin, and romanization of CJK names (Wade-Giles vs Pinyin). This needs transliteration tables or a library. Hold it until a researcher actually has mixed-script evidence.
- **Sound-alikes:** Daitch–Mokotoff Soundex suits Eastern European and Jewish names better than American Soundex; Beider–Morse is language-aware. Either could be a `wordSimilarity` tier below the near-spelling one.
- **Nicknames:** equivalence tables per language (Jim ↔ James, Hans ↔ Johann, Pepe ↔ José), ideally seeded vocabulary so projects can extend them.

## Order of work (rough)

1. Accent folding in the shared normalizer, with the 0.95 folded-match tier and a cache version bump. This is small and has the highest payoff.
2. Profiles carry roles; `NameComparer` reads them; a `patronymic` part type; a second seeded profile (Spanish dual surname, or Icelandic).
3. Locale collation for sorted lists, alongside the name-format display styles.
4. Nickname tables, then phonetics, then transliteration.

## Open questions

- Is the matching profile the **candidate's** name format (its Person), the **probe's**, or the project default when they differ?
- Does accent folding belong in `resolve.NormalizeForm` (affecting clustering), or only in matching, so that "José" and "Jose" stay separate clusters on a Person's page and merge only when suggested?
- Who chooses the collation locale: the project, the viewer's system locale, or the Person's name format?
- Should a patronymic link actually use the father's handle, once family edges exist (S9-28)?
