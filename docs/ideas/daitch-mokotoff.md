# Daitch–Mokotoff matching

**Status:** idea, 2026-10-12. Not scheduled. Descoped from Spike 10 (S10-26).

**Spike 11 moves the code this note names.** The spelling helpers that lived in `core/autoreconcile` moved to value compare (`core/valuecompare`) in S11-02; `wordSimilarity` lives in `core/valuecompare/words.go` and its numbers in `core/matchregistry/value.go`; S11-05 deleted `match`'s copy ([`comparison-layers.md`](../deployment-plan/spike-11/comparison-layers.md)). Read the code paths below as "today's". **One conflict to settle when this is pulled in:** a shared code scores 0.7 below, but a name pair now needs at least 0.75 to be `Near`, so a sound-alike would read `Different`. The likely shape is `Near` with cause `phonetic` and its own minimum in the registry.

A sound-alike tier under near-spelling, for two surnames that are not a one-edit misspelling and still share a code. Beider–Morse is the more precise phonetic algorithm and stays a separate later idea ([`international-names.md`](international-names.md)). This note is the Daitch–Mokotoff design if that is the one pulled back.

Call it Daitch–Mokotoff. Do not call it a Jewish chart, Jewish Soundex, Eastern European Soundex, or a JewishGen chart. The authors discourage those nicknames because the rules are not tied to one community. The letter-to-digit rules are the procedure. Implementing them does not need a citation, a publisher link, or a borrowed copyright header. Do not copy anyone's explanatory prose. A one-line comment that names Daitch–Mokotoff Soundex is enough.

## Where it sits

[`wordSimilarity`](../../core/valuecompare/words.go) already returns, in order: exact (1), initial, alias (0.98), fold (0.95), then the edit-distance ratio. Phonetics runs only when that ratio is 0, on the already-folded word.

A shared code scores **0.7**. That is above 0 and below the lowest one-edit score the spelling tier returns (Ann/Anne is 0.75, Smyth/Smith is 0.8). Exact, alias, fold, and near-spelling never reach it. An initial is decided earlier and does not fall through.

The tier applies to every role. Alias tables stay given, nick, and untyped. Surnames are why this tier exists.

Do not change `SpellingSimilarity` in [`core/valuecompare/words.go`](../../core/valuecompare/words.go). The reconciler keeps merging only edit-distance variants. Display, `NormalizeForm`, and `sort_key` do not code, so the Person page still keeps both spellings.

## Pairs

Lee codes as `800000` and Leigh as `850000`. They do not share a code, so this tier does not make them meet. The Spike 10 line that said they would was wrong.

Lewinsky and Levinski both code as `876450`. Two substitutions on eight letters is ratio 0.75, and the spelling tier returns a below-floor ratio only for a single added or dropped letter, so today the pair scores 0. A surname-only comparison would score 0.7.

Smyth and Smith stay on edit distance (0.8) and never reach this tier. Moskowitz and Moskovitz share `645740` and are also one substitution on nine letters (ratio about 0.89), so they meet by spelling and do not need the tier.

## How a word becomes digits

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

## When it is built

`WordRules` gains a phonetic chart. Nil uses the default chart. An empty chart scores no phonetic hits, the same shape as an empty fold map. `valuecompare.WordRules` (numbers in [`core/matchregistry/value.go`](../../core/matchregistry/value.go)) points at the chart so the registry test stays green.

Coder tests lock the codes above: Auerbach is `097400` and `097500`; Uhrbach shares both; Moskowitz and Moskovitz are `645740`; Lewinsky and Levinski are `876450`; Lee is `800000` and Leigh is `850000`. A surname-only Lewinsky/Levinski comparison equals 0.7. The existing Smyth/Smith case stays the edit-distance score. `José`/`Jose` stays 0.95. An empty chart makes Lewinsky/Levinski score 0.

[`docs/matching.md`](../matching.md) then lists the tier after near-spelling. No UI, no migration, no product-version bump.
