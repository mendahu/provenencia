# Ship log — S11-08 One normalization helper in Swift

**Plan:** [`S11-08-swift-normalization.md`](S11-08-swift-normalization.md) · **Branch:** `s11-08-swift-normalization` · **PR:** #466
**State:** ready

## Round 0 — implement
Head `24c5df7e`. Gate PASS (orchestrator run: diff-check, gofmt, go-vet, go-test, localizable, plan-docs, xcode-test). Report deviations:
- `TextMatch.range(of:in:)` compares Unicode scalars rather than `Character`s (D-5), so a key ending in a combining scalar cannot merge with its neighbor; same results on all fixtures.
- `EvidenceGraphModelTests.promotedPersonOpensItsPageAndShowsItsName` seed given `nameParts` (assertions unchanged; allowed by stop conditions).
- `scoreSearchFields` builds its value-key map with `uniquingKeysWith`.
- `EvidenceGraphModelTests.deleteUncitedSubjectSucceeds` flaked once in a full gate run; passed on re-runs; unchanged.
- Implementer deleted eight stale `~/Library/Developer/Xcode/DerivedData/Provenencia-*` caches (disk full).
- Manual check (§10) left for the user.

## Round 1 — review
Base `d49fde71`, head `24c5df7e`. Verdicts: CON changes needed · MNT changes needed · UX changes needed · SEC clean · PRF changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Sources filter ~200 ms main-thread per keystroke at 2,000 Sources (per-character Foundation keying, re-keyed every `body`) | Fix now → F-1 | Reviewer benchmark of the shipped helper (212 ms -O vs 12.7 ms old). Reviewer's fix part 1 (lowercase the whole string) is rejected by D-2 (Greek final sigma vs Go), so F-1 keeps D-2's per-character rule and adds an ASCII fast path plus a per-row key cache in `SourcesModel` |
| CON-1 | CON | major | `Query.matches`/`matchesKey`, combo bands and type-select compare keys by grapheme (`String.contains`/`hasPrefix`), not by scalar like Go's byte `strings.Contains` (`가` misses `간장`, `क` misses `किताब`); filter and highlight disagree | Fix now → F-2 | Carries a failing test; verified by the fix round. Code this PR adds |
| MNT-1 | MNT | major | Go guard's backslash refusal can never fire (`swiftFixturePair` excludes `\`), so escaped fixtures are silently skipped | Fix now → F-3 | Verified: `api/ffi/swift_vocabulary_test.go:59,70` — the refusal requires `MatchString` of a pattern with `[^"\\]*` |
| MNT-2 | MNT | major→minor | `macos-client.mdc` rule says mirrors keep a `static let fixtures` table; `ProjectSlug` has `folderNameFixtures` | Fix now → F-4 (rides along) | Verified: `ProjectSlug.swift:9`. Doc accuracy; re-rated minor, few lines |
| UX-1 | UX | major | Title lookup on `TextMatch.key` drops every `.`, so Split reads dotted initials as titles (`S.R.`→Sister, `M.S.`→Ms, `I.I.`→II) | To user → answered → F-5 | Contradicts D-10 / AC-5.1 (`D.r.` → true). User (2026-10-18): "Refuse X.Y. initials" — keep `TextMatch.key`, a token with a period between two letters is never a title; drop `D.r.` from AC-5.1; add S.R. / M.S. / I.I. as false |
| PRF-2 | PRF | minor | Properties board reads `visibleProperties` twice per body | Follow-up | Mostly covered by F-1's fast path; `PropertiesView.swift` not otherwise touched |
| PRF-3 | PRF | minor | Combo filter keys each label twice (key + `words`) | Logged | Covered by F-1's fast path; bounded vocab lists |
| SEC note / CON note | PRF | minor | `TextMatch.range` allocates an `Array` slice at every offset | Fix now → F-2 (rides along) | F-2 rewrites the scalar window search in the same file |
| MNT-3 | MNT | nit | `nameField(memberIDs:sources:)` takes `sources` only for titles | Logged | Harmless |
| CON-2 | CON | nit | T-16 already passed before the change | Logged | Guards S11-06 policy, as planned |

### Fix plan 1

**F-1 — Make the Sources filter cheap per keystroke without changing D-2's key** (from PRF-1)
Files: `macos/App/Platform/TextMatch.swift`, `macos/App/Features/Sources/SourcesModel.swift`, `macos/ProvenenciaTests/TextMatchTests.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`. Skill: `add-swift-test`.
Failing test first: `TextMatchTests.asciiFastPathKeysLikeFoundation` — for every ASCII scalar 0x00–0x7F, alone and inside `"aXb"`, `TextMatch.key(_, at:, locale:)` at both levels and in `en_US`, `tr_TR` and `TextMatch.posix` equals the D-2 formula computed inline in the test (`String(ch).lowercased(with:)`, decompose at `.equivalent`, the same scalar filter and folds). This one is a guard, so it passes before and after; the speed-up is the change.
Do:
1. In `characterKey`, add a fast path before the Foundation calls: when `ch.unicodeScalars` is exactly one scalar, that scalar `isASCII`, and it is not `"I"` (Turkish/Azeri lower it to `ı`), lowercase `A`–`Z` by adding 32 and run the existing whitespace / punctuation filter on it; do not call `lowercased(with:)` or `decomposedStringWithCanonicalMapping`. Everything else takes the D-2 path unchanged. Do not lowercase the whole string at once (D-2 rejects it: Greek final sigma). Avoid building an `[Unicode.Scalar]` per character on the fast path (e.g. let `characterKey` append into an `inout` buffer that `key` and `range` both pass).
2. In `SourcesModel`, key each Source's title and ref once and reuse it: a cache from `source.id` to `(revision: source.updated.revision, titleKey: String, refKey: String)` (two keys, never one joined key, so a match cannot cross fields); filter with `query.matchesKey(titleKey) || query.matchesKey(refKey)`. Recompute an entry only when its revision differs or the id is new; drop ids no longer in `sources`. Mark the cache `@ObservationIgnored` (or the model's equivalent) so filling it inside `visibleSources` does not invalidate the view. The cache uses the query's locale (`.autoupdatingCurrent`); if the locale can change while the model lives, include it in the cache's validity.
3. Add `SourcesModelTests.searchSeesAnEditedTitle`: filter, change one Source's title and revision in the fake, filter again → the new title matches and the old one does not.
Check: tests pass; `scripts/check-pr.sh` PASS. In the report, give a before/after timing for 2,000 Sources × 56-character titles (a scratch benchmark, not committed), as PRF-1 did.

**F-2 — Compare keys by scalar, like Go's `strings.Contains`** (from CON-1; SEC/CON note on `range` allocation rides along)
Files: `macos/App/Platform/TextMatch.swift`, `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift`, `macos/App/DesignSystem/Components/FloatingMenuSelection/PVFloatingMenuSelection.swift`, `macos/ProvenenciaTests/TextMatchTests.swift`.
Failing test first: `TextMatchTests.queryMatchesInsideAGraphemeLikeGo` — `TextMatch.Query("가", locale: enUS).matches("간장")`, `TextMatch.Query("क", locale: enUS).matches("किताब")`, `PVTypeSelectMatcher.index(in: ["간장"], prefix: "가") == 0`, and a combo box case: `PVComboBoxMatch.filter` with query `가` keeps an option labelled `간장` in band 0.
Do: add private scalar-level helpers in `TextMatch` (contains and has-prefix over `unicodeScalars`) sharing one window search with `range(of:in:)`; that search compares in place (no `Array(keyed[i..<i + n])` per offset). Use them in `Query.matches` / `matchesKey`. Add `Query.isPrefix(ofKey:)` and use it for the combo box band checks (`labelKey.hasPrefix`, word `hasPrefix`) and in `PVTypeSelectMatcher.index` instead of `String.hasPrefix` on keys.
Check: test passes; existing `TextMatchTests`, `PVComboBoxTests`, `PVTableTypeSelectMatcherTests` pass; `scripts/check-pr.sh` PASS.

**F-3 — Make the Go guard actually refuse escaped fixtures** (from MNT-1)
Files: `api/ffi/swift_vocabulary_test.go`. Rules: `.cursor/rules/go-tests.mdc`.
Failing test first: `TestSwiftFixturePairsRefusesEscapes` — write a temp Swift file whose `enum E { static let fixtures = [("God\u{2010}father", "godfather"), ("a", "b")] }` and assert the reader reports the escape (refactor the reader so the refusal is testable: return an error, or take a `testing.TB` you can fake); a second case with only plain pairs returns them.
Do: find candidate pairs with an escape-tolerant pattern (`\("((?:[^"\\]|\\.)*)", "((?:[^"\\]|\\.)*)"\)`); fail when either capture contains `\`; return only passing pairs. Read only the `static let fixtures` array literal, not the whole enum body. Apply the same narrowing to `swiftFoldEntry` (read only the `letterFolds` literal).
Check: new test passes; `TestSwiftTextMatchMirrorsTextKey` still passes; `scripts/check-pr.sh` PASS.

**F-4 — Make the mirror-fixture rule true for `ProjectSlug`** (from MNT-2)
Files: `.cursor/rules/macos-client.mdc`, `docs/ideas/shared-list-helpers.md`.
Do: reword the bullet to "keeps a fixture table in its enum (`TextMatch.fixtures`, `FieldSlug.fixtures`, `ProjectSlug.folderNameFixtures`)"; do not rename code. In the ideas item "Go reads every Swift mirror", add that `ProjectSlug`'s `"Robins\\Family"` / `"Robins\"Family"` rows are escapes the reader refuses, like `FieldSlug`'s `\u{2019}`.
Check: `scripts/check-plan-docs.py` 0 errors.

**F-5 — Dotted initials are never titles** (from UX-1; user decision 2026-10-18)
Files: `macos/App/Features/Names/NameTitleRegistry.swift`, `macos/ProvenenciaTests/NameValueDraftTests.swift`, plan `S11-08-swift-normalization.md` (AC-5.1 row and the D-10 doc-comment text, recorded as a user-approved deviation in the Landed note).
Failing test first: in `registryIgnoresCaseAndPunctuation`, `isPrefix("S.R.")`, `isPrefix("M.S.")`, `isPrefix("D.R.")`, `isSuffix("I.I.")`, `isSuffix("V.I.")` → false; `isPrefix("D.r.")` → false (was true); `Dr.`, `dr`, `JR.`, `Jr.,` stay true. Add `NameFormSplit.parts(from: "S.R. Ranganathan")` → given `S.R.`, surname `Ranganathan`, no prefix.
Do: in `isPrefix` / `isSuffix`, return false when the token has a `.` with a letter immediately before and after it; otherwise keep the `TextMatch.key(token, at: .same, locale: TextMatch.posix)` lookup. Doc comment: "Case and edge punctuation do not count (`Dr`, `Dr.`, `Jr.,`); dotted initials (`S.R.`) are never titles." Update the plan's AC-5.1 row (drop `D.r.` → true, add the dotted-initials cases → false) and note it in the Landed note as the user's call on UX-1.
Check: tests pass; `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `44dfd0b7`. Gate PASS (orchestrator run). F-1 to F-5 done. F-1 timing (scratch, 2,000 Sources, -O): per-keystroke filter 218 ms → 1.4 ms on cached keys (16 ms keying every row with the ASCII fast path). Deviations: fast path also excludes NUL (Foundation drops it); Sources cache resets on locale change; `characterKey` became `appendKey(of:level:locale:to:)` + `keeps(_:level:)` with a shared in-place `firstWindow` search; Go reader split into `readSwiftFixturePairs` (returns an error) and reads only the named literal; F-5 also updated the `completed.md` paragraph (`D.r.` → `Dr.`). Implementer ran a stray `git stash pop` that conflicted with the user's pre-existing `main` stash; it restored the file and the stash is intact (verified: `stash@{0}: On main: wip: Localizable.xcstrings`).

## Round 2 — review
Scope `24c5df7e...44dfd0b7`. Lenses: CON, MNT, UX, PRF (SEC had no Fix-now finding). All four Clean. Every round-1 Fix-now finding (PRF-1, CON-1, MNT-1, MNT-2, UX-1, `range` allocation note) reported Resolved.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-3 | CON | minor | `FakeStore.scoreSearchFields` still uses grapheme `String.contains` on keys (`FakeStore.swift:3358`) | Follow-up | Round 2 adds only blockers; fake-store only, no shipped behavior |
| MNT-4 | MNT | note | `SourcesModel` is the only per-row key cache; Properties and the rotor key each read | Follow-up (with PRF-2) | A shared helper earns its place when a second list needs it |

No Fix plan 2.

## Result
Gate PASS · CI see PR #466 · open: none · follow-ups: PRF-2 (+MNT-4), CON-3 · to user: UX-1 (answered 2026-10-18: refuse X.Y. initials)
