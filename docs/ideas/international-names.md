# International names — later

**Status:** idea, not scheduled. Spike 10 shipped accent folding in matching, spaced particles, nickname tables, and the Western profile ([decisions](../deployment-plan/archive/spike-10/)). Includes the Daitch–Mokotoff design descoped from S10-26.

## Scripts, sound, and dates

- **Across scripts.** Cyrillic vs Latin (*Иванов* / *Ivanov*), Hebrew or Yiddish vs Latin, and romanization of CJK names (Wade-Giles vs Pinyin). Needs transliteration tables or a library. Hold it until a researcher has mixed-script evidence.
- **Beider–Morse.** Language-aware phonetics. More precise than Daitch–Mokotoff, and a larger rule set. Still later than the Daitch–Mokotoff section below.
- **Calendars and double dating.** Julian vs Gregorian (the 1752 switch in Britain and its colonies, at different times elsewhere) and "1717/18" double dating. Matching ignores both today. They vary by place and church.
- **Impossible calendar days.** Date validation accepts day 31 in any month (`datevalue` `dayOK`), so 30 February saves, but value compare finds no usable bound for it and matching silently ignores the date (Spike 11 S11-04 ship log; dogfood 2026-10-29). Refuse a day past the month's length when the date is entered.

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

- **Name changes as a signal.** Married, religious and anglicized names connect today only when one of a handle's names matches. A "known as" relationship between names, or the marriage event, could carry it. Spike 11 treats Mary Smith / Mary Robins as `Different` at the value level, and leaves marriage evidence to links. Maiden names in the conclusion are the intended path ([`comparison-layers.md`](../deployment-plan/archive/spike-11/comparison-layers.md#entities-compare-against-the-conclusion)). The `former_surname` part type (S14-09) records a maiden or earlier surname, and a record carrying the former surname pairs it as a family name; reading a former and a current surname as one person across records is still open.
- **Typed forms with decoration.** `NameFormSplit` does not recognise a quoted nickname or a parenthesized former surname when it splits a typed form; if it learns to, `NamePartType.decoration` is the table to read.
- **The unused `Sort` field.** `namevalue.PartTypeInfo.Sort` orders the registry and nothing else reads it.
- **Seed roles from `matchregistry`.** The Western seed's roles in `nameformats` repeat `matchregistry`'s Western `PartRoles`; deriving one from the other would retire `TestSeedMatchesRegistryWestern`.
- **A fixture pin of the undecorated join.** A Go and Swift fixture that pins `nameformats.Profile.Format` against `NameValueDisplay` would catch drift in order and separators.

## Format and sort, beyond the project default

- **Per-person name format.** A Reconciliation Claim on `name_format`. Until that exists, every name uses the project default profile. Each name already may carry its own pattern once a loader sets `Value.NamePattern`.
- **Accent merging on the Person page** — decided 2026-10-17 (S11-06): an accent fold is value compare's `Equivalent`, which never groups or folds, so *José* and *Jose* show as two values.
- **Collation locale.** A project setting (default `en`) whose locale orders name lists through `golang.org/x/text/collate`, so accented and unaccented names sort in that locale's order. Comparison fold is not the sort, and the match score does not change. Descoped from S10-37 (2026-10-16); nothing stores a locale today.
- **Collation picker.** A control for that setting. A viewer who wants a locale other than the project's (a shared tree read abroad) is the same gap.

## Whole-phrase copy, app-wide

S11-33 made every sentence on the Promote evidence sheet (removed in S13-23) one whole translatable phrase (see `add-localized-string`, "Whole phrases"): an argument is data or a complete noun phrase from one table entry, never a bare translated word the sentence must agree with. Other copy still splices a translated word into a sentence, so a language that inflects or agrees an article cannot translate it: the `EventTitle` formats (`L10n.EventTitle`), `evidenceGraph.bridge.summary.*` (a term label inside "… participated as … at …"), the `CitationComposer` term formats, `sources.sortedBy`, `conclusions.outcomeOutvotedBy` (a count and a Sources word), and `properties.inspector.holdsSingleOnly` (a value type label inside "… Properties hold one value"). Each becomes a key per sentence, or takes its word whole from `L10n.PropertyTerm.resource`, when a second language is added and the copy is next touched.

## Daitch–Mokotoff matching

Descoped from Spike 10 (S10-26). Spike 11 moved the spelling helpers this note names into `core/valuecompare` (S11-02); `wordSimilarity` lives in `core/valuecompare/words.go` and its numbers in `core/matchregistry/value.go` ([`comparison-layers.md`](../deployment-plan/archive/spike-11/comparison-layers.md)). **One conflict to settle when this is pulled in:** a shared code scores 0.7 below, but a name pair now needs at least 0.75 to be `Near`, so a sound-alike would read `Different`. The likely shape is `Near` with cause `phonetic` and its own minimum in the registry.

A sound-alike tier under near-spelling, for two surnames that are not a one-edit misspelling and still share a code. Beider–Morse stays a separate later idea (above). This section is the Daitch–Mokotoff design if that is the one pulled back.

Call it Daitch–Mokotoff. Do not call it a Jewish chart, Jewish Soundex, Eastern European Soundex, or a JewishGen chart. The authors discourage those nicknames because the rules are not tied to one community. The letter-to-digit rules are the procedure. Implementing them does not need a citation, a publisher link, or a borrowed copyright header. Do not copy anyone's explanatory prose. A one-line comment that names Daitch–Mokotoff Soundex is enough.

### Where it sits

[`wordSimilarity`](../../core/valuecompare/words.go) already returns, in order: exact (1), initial, alias (0.98), fold (0.95), then the edit-distance ratio, then a prefix (S11-45). Phonetics runs only when that ratio is 0, and is ordered against the prefix rung, on the already-folded word.

A shared code scores **0.7**. That is above 0 and below the lowest one-edit score the spelling tier returns (Ann/Anne is 0.75, Smyth/Smith is 0.8). Exact, alias, fold, and near-spelling never reach it. An initial is decided earlier and does not fall through.

The tier applies to every role. Alias tables stay given, nick, and untyped. Surnames are why this tier exists.

Do not change `SpellingSimilarity` in [`core/valuecompare/words.go`](../../core/valuecompare/words.go). The reconciler keeps merging only edit-distance variants. Display, `NormalizeForm`, and `sort_key` do not code, so the Person page still keeps both spellings.

### Pairs

Lee codes as `800000` and Leigh as `850000`. They do not share a code, so this tier does not make them meet. The Spike 10 line that said they would was wrong.

Lewinsky and Levinski both code as `876450`. Two substitutions on eight letters is ratio 0.75, and the spelling tier returns a below-floor ratio only for a single added or dropped letter, so today the pair scores 0. A surname-only comparison would score 0.7.

Smyth and Smith stay on edit distance (0.8) and never reach this tier. Moskowitz and Moskovitz share `645740` and are also one substitution on nine letters (ratio about 0.89), so they meet by spelling and do not need the tier.

### How a word becomes digits

The chart below is the data. The encoder does not name letters. Follow this chart, not Apache Commons' extra branches.

- Walk the word from the left. At each position take the longest letter sequence that has a row.
- Read the cell for where that sequence sits: start of the word, immediately before a vowel, or anywhere else. `NC` means skip it. A cell with two codes, or a "try both" row, branches, so one word can yield several six-digit codes.
- Pad with zeros to six digits. The first letter is coded.
- When the next sound would repeat the code just emitted, emit it once. `MN` and `NM` are the exception: their letters are coded separately as `66`, as in Kleinman `586660`.
- Two words hit when any code is shared. A shared `000000` does not count. A character with no row is skipped. A word that produces no code does not match.

Matching folds the word before this step, so `Ą`, `Ę`, and `Ţ` are already `A`, `E`, and `T`. Those three rows stay in the table because they are part of the chart, and they are not reached from `wordSimilarity`.

A blank start cell means that sequence is not used at the start. The shorter row applies (`MN` at the start is `M` then `N`).

| Letters | Alternates | Start | Before a vowel | Other |
| --- | --- | --- | --- | --- |
| AI | AJ, AY | 0 | 1 | NC |
| AU | | 0 | 7 | NC |
| Ą | Polish a-ogonek | NC | NC | 6 or NC |
| A | | 0 | NC | NC |
| B | | 7 | 7 | 7 |
| CHS | | 5 | 54 | 54 |
| CH | try KH (5) and TCH (4) | | | |
| CK | try K (5) and TSK (45) | | | |
| CZ | CS, CSZ, CZS | 4 | 4 | 4 |
| C | try K (5) and TZ (4) | | | |
| DRZ | DRS | 4 | 4 | 4 |
| DS | DSH, DSZ | 4 | 4 | 4 |
| DZ | DZH, DZS | 4 | 4 | 4 |
| D | DT | 3 | 3 | 3 |
| EI | EJ, EY | 0 | 1 | NC |
| EU | | 1 | 1 | NC |
| Ę | Polish e-ogonek | NC | NC | 6 or NC |
| E | | 0 | NC | NC |
| FB | | 7 | 7 | 7 |
| F | | 7 | 7 | 7 |
| G | | 5 | 5 | 5 |
| H | | 5 | 5 | NC |
| IA | IE, IO, IU | 1 | NC | NC |
| I | | 0 | NC | NC |
| J | try Y (1) and DZH (4) | | | |
| KS | | 5 | 54 | 54 |
| KH | | 5 | 5 | 5 |
| K | | 5 | 5 | 5 |
| L | | 8 | 8 | 8 |
| MN | | | 66 | 66 |
| M | | 6 | 6 | 6 |
| NM | | | 66 | 66 |
| N | | 6 | 6 | 6 |
| OI | OJ, OY | 0 | 1 | NC |
| O | | 0 | NC | NC |
| P | PF, PH | 7 | 7 | 7 |
| Q | | 5 | 5 | 5 |
| RZ, RS | try RTZ (94) and ZH (4) | | | |
| R | | 9 | 9 | 9 |
| SCHTSCH | SCHTSH, SCHTCH | 2 | 4 | 4 |
| SCH | | 4 | 4 | 4 |
| SHTCH | SHCH, SHTSH | 2 | 4 | 4 |
| SHT | SCHT, SCHD | 2 | 43 | 43 |
| SH | | 4 | 4 | 4 |
| STCH | STSCH, SC | 2 | 4 | 4 |
| STRZ | STRS, STSH | 2 | 4 | 4 |
| ST | | 2 | 43 | 43 |
| SZCZ | SZCS | 2 | 4 | 4 |
| SZT | SHD, SZD, SD | 2 | 43 | 43 |
| SZ | | 4 | 4 | 4 |
| S | | 4 | 4 | 4 |
| TCH | TTCH, TTSCH | 4 | 4 | 4 |
| TH | | 3 | 3 | 3 |
| TRZ | TRS | 4 | 4 | 4 |
| TSCH | TSH | 4 | 4 | 4 |
| TS | TTS, TTSZ, TC | 4 | 4 | 4 |
| TZ | TTZ, TZS, TSZ | 4 | 4 | 4 |
| Ţ | Romanian t-cedilla | 3 or 4 | 3 or 4 | 3 or 4 |
| T | | 3 | 3 | 3 |
| UI | UJ, UY | 0 | 1 | NC |
| U | UE | 0 | NC | NC |
| V | | 7 | 7 | 7 |
| W | | 7 | 7 | 7 |
| X | | 5 | 54 | 54 |
| Y | | 1 | NC | NC |
| ZDZ | ZDZH, ZHDZH | 2 | 4 | 4 |
| ZD | ZHD | 2 | 43 | 43 |
| ZH | ZS, ZSCH, ZSH | 4 | 4 | 4 |
| Z | | 4 | 4 | 4 |

Worked paths:

- Lewinsky: L=8, E skipped, W=7, I skipped, N=6, S=4, K=5, Y skipped, pad → `876450`. Levinski takes the same path with V=7 instead of W.
- Leigh: L=8, E skipped, I skipped, G=5, H skipped, pad → `850000`. Lee: L=8, then both E's skipped, pad → `800000`.
- Auerbach branches on CH. As KH the tail is 5 (`097500`); as TCH the tail is 4 (`097400`). Uhrbach shares both.

### When it is built

`WordRules` gains a phonetic chart. Nil uses the default chart. An empty chart scores no phonetic hits, the same shape as an empty fold map. `valuecompare.WordRules` (numbers in [`core/matchregistry/value.go`](../../core/matchregistry/value.go)) points at the chart so the registry test stays green.

Coder tests lock the codes above: Auerbach is `097400` and `097500`; Uhrbach shares both; Moskowitz and Moskovitz are `645740`; Lewinsky and Levinski are `876450`; Lee is `800000` and Leigh is `850000`. A surname-only Lewinsky/Levinski comparison equals 0.7. The existing Smyth/Smith case stays the edit-distance score. `José`/`Jose` stays 0.95. An empty chart makes Lewinsky/Levinski score 0.

[`docs/matching-model.md`](../matching-model.md) then lists the tier's knob in the inventory. No UI, no migration, no product-version bump.
