# Review — S14-09 Quote nicknames and add former surnames (was S11-29)

Reviewed as S11-29 on 2026-10-19; moved to Spike 14 on 2026-10-29. The refresh by `plan-pr` gets its own review.

**Verdict:** Fits with changes — Not ready → fixed in round 1; round 2 Ready
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. Product calls were asked before planning (2026-10-19): curly quotes, parentheses + italic, new catalogs only.

## Writer's own calls
- Former surname sits before the surname (*Mary (Smith) Jones*), role `family` for matching.
- Italics only on the Persons list row and Person page title; every other surface shows the parentheses as plain text.
- No design brief: existing titles gain punctuation and one italic run.
- A Swift golden that shows a `nick` part is updated to the quoted string.

## Round 1

### X-1 · blocker · Run splitting and the L10n shape unspecified
**Answer:** Fixed — D-7 (`NamePartDecoration.runs`, private-use slot, two-piece rule with bare-word fallback), D-8 (`NameDisplayRun`, separators are plain runs), Strings (static resource constants), T-7.

### X-2 · blocker · How the title knows a run is emphasized
**Answer:** Fixed — D-9 / D-10: one `attributed` call; `.name(_ styled:)` keeps `styled` only when a run has `.emphasized`; form fallback is plain.

### X-3 · major · AC-1.4 callers
**Answer:** Fixed — §4 lists every `NameValueDisplay.string` caller; D-13 / Out: the header-less omnibar hit keeps its indexed title; AC-1.4 by T-4 (shared function) and manual step 2.

### X-4 · major · Enum payload change sites
**Answer:** Fixed — D-9 splits the `text` arm; macOS table names the three matches.

### X-5 · major · Checks need macOS
**Answer:** Fixed — §7 preamble and a stop condition: Swift checks and the manual check go to a macOS runner.

### X-6 · major · No fallback for an upright italic
**Answer:** Fixed — D-11 binding fallback (italic `.font` on the run).

### X-7 · major · Test mechanics
**Answer:** Fixed — `western` fixture; T-2 `TestWesternSeedOrder`; T-6 `TestFormerSurnamePairsAsFamily` in `core/matchregistry` (value compare tests can't import the registry); `partCodes` unchanged (§4 Precedent); T-5 counts emphasized runs only.

### X-8 · minor · Doc anchors
**Answer:** Fixed — §4.2a, §4.5, fenced §4.1 line, international-names wording in STEP-5.

### X-9 · minor · Row vs sort key
**Answer:** Fixed — §4.5 rewrite (M-1) and a risk row.

### X-10 · nit — logged.

### M-1 · major · structured-name-model §4.5 becomes false
**Answer:** Fixed — STEP-5 rewrites requirement 2 and adds the Decoration row; comments on `Format` and `runs`.

### M-2 · major · Name stored and joined twice
**Answer:** Fixed — D-8, D-9, D-10; T-5 asserts `text == String(styled.characters)`.

### M-3 · minor · Two switches per decorated type
**Answer:** Fixed — `NamePartDecoration` and one `decoration` property (D-6); split lives on it (D-7).

### M-4 · minor · Magic count; Sort and parity
**Answer:** Fixed — T-1 asserts the key list; `Sort` and Go↔Swift parity logged in international-names (Out).

### M-5 · minor · `NameFormSplit` asymmetry
**Answer:** Fixed — Out line, logged in international-names.

### M-6 · nit · Name `TestSeedMatchesRegistryWestern`
**Answer:** Fixed — §4, D-5, T-2.

### M-7 · nit · Key segment
**Answer:** Fixed — `nameValue.display.former_surname`.

## Round 2

Reviewer (opus, both lenses): no new blockers; every round-1 answer accepted. Verdict: Ready. Notes, all fixed:
- R2-1 · minor · `conclusionheaders/order.go` `personSortKey` comment says the key is the row string — added to STEP-1.
- R2-2 · minor · D-7 split call — `components(separatedBy:)`, count then drop empties; T-7 case for `(%@`.
- R2-3 · nit · `attributed` fallback narrower than `string` — "when the runs are empty".
- R2-4 · nit · list-row italic font named (`PVFont.body(size: PVTypeScale.body)`); map row status updated.

---

# Refresh review — 2026-10-31 (against `main` @ `b6a8f3e2`)

**Verdict:** Fits with changes — Ready with changes → all fixed; no blocker open
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet), fresh each round

## To user (answered)
- **Board.** The 2026-10-29 board rule covers any change to how a view looks. User (2026-10-31): no board for S14-09; they will ask for a prompt to update the board's examples. Recorded in the plan's **Rules this plan bends**.
- **M-2 guard row.** User (2026-10-31): move B-24's part-type half (the `swift_vocabulary_test.go` row and `add-name-value` step 6) from S14-22 into S14-09. Map re-cut in the same commit as this refresh.

## Writer's own calls
- No `CacheVersion` bump (D-14): no stored name can hold `former_surname` before this PR.
- The seeded order is pinned by the existing `TestGetNameFormat` (D-16), not a new Go test.
- `NameDisplay.styled` is built only when a run is emphasized (D-8), so names without a former surname build no `AttributedString`.
- Every `NamePartType` case spells its raw value so one regex (`listCase`) reads the enum (D-18).
- The enum stays in `Features/Names/` (macos-client.mdc allows an enum beside its owner).

## Round 1

### M-1 · blocker · D-7's slot splitter is a second styled-format mechanism beside `StyledFormat`
**Answer:** Fixed — D-7 renders a decorated word with `L10n.format` (plain) or `StyledFormat.attributed` (styled); the sentinel, splitter and bad-translation fallback are gone. D-8 runs carry a `decoration`, `NameDisplay` holds text and an optional styled copy. T-7 shrinks; a Stop condition covers `StyledFormat`'s placement. Rejected the reviewer's "always build `AttributedString` and strip it": every `string` caller would pay it; instead the styled copy is built only when needed.

### M-2 · blocker · New case in an unguarded wire enum; guard row left to S14-22
**Answer:** To user → moved into this PR (2026-10-31). US-3, D-18, T-8, STEP-2, STEP-5 skill text; map S14-22 row, scope and dependency lines updated.

### M-3 · major · T-6 premise false; belongs in `nameCases`
**Answer:** Fixed — D-17: `partCodes` gains `f`; AC-2.8 is a `nameCases` row (T-6b); AC-2.7 is `TestFormerSurnamePairsAsFamily` in `name_test.go` (a `nameCases` row would pin a similarity number). Codebase fit row corrected.

### M-4 · major · Skill guidance made incomplete
**Answer:** Fixed — STEP-5 rewrites `add-name-value` step 6, including the `decoration` arm and "display only".

### M-5 · minor · D-11 fallback written twice
**Answer:** Fixed — D-11 names one helper, `ConclusionListRow.italicizingEmphasis(_:font:)`, written only if the fallback fires.

### M-6 · minor · `.name` factory overloaded on type
**Answer:** Fixed with X-4 — the static constructor is gone (D-9); callers pass `styled:` explicitly.

### M-7 · minor · `sortedOrderUsesTheCatalogSeparator` keeps an inline order
**Answer:** Fixed — D-15, T-4.

### M-8 · minor · Western roles in two hand-written lists
**Answer:** Logged — Out row; parked in international-names (STEP-5). Pre-existing, and derivation touches create-time seeding.

### M-9 · nit · Go/Swift join held by comments only
Logged — Out row; parked in international-names (STEP-5).

### X-1 · blocker · "one pair smith/smith" is wrong; `Mary`/`Mary` pairs too
**Answer:** Fixed — AC-2.7, T-6 and the Stop condition say `Parts` holds the `Smith` pair beside the `Mary` pair and `OnlyB` is `Jones`.

### X-2 · major · T-6 has no recipe; `valuecompare` can import `matchregistry` in tests
**Answer:** Fixed with M-3 — T-6 gives the exact `nv` specs and assertions.

### X-3 · major · `CacheVersion` is 21, not 18
**Answer:** Fixed — §4 row. D-14 unchanged.

### X-4 · major · `.name("…")` ambiguous with an `AttributedString` overload
**Answer:** Fixed — D-9 drops the static constructor; only the case with `styled: AttributedString? = nil` remains.

### X-5 · major · T-7's `(%@` case cannot be built
**Answer:** Fixed by M-1 — no splitter, so the case is gone; T-7 writes its expectations against `decoration` and `L10n.format`.

### X-6 · minor · Edge behaviour of runs
**Answer:** Fixed — D-8: one run per word, separator runs, never merged; styled copy built per run then concatenated. The sentinel edge is gone with M-1.

### X-7 · minor · D-11 fallback under-specified
**Answer:** Fixed with M-5.

### X-8 · minor · Seed role row position
**Answer:** Fixed — D-4: right after the `surname` role.

### X-9 · minor · Docs and sweep loose
**Answer:** Fixed — DoD names the landing status `Landed in #N`; the map re-cut is done at planning. The sweep line stays (the template requires it; a zero-hit sweep is the evidence). The plan is committed before `ship-pr` runs.

## Round 2

Both reviewers accepted the round-1 answers except as below. One correction to the round-1 log: M-5's answer said the helper is written only if a fallback fires; X-12 made it unconditional (below).

### M-10 · blocker · D-7 calls `L10n.format` from a feature file, against `add-localized-string` step 4
**Answer:** Fixed with the reviewer's option (b), which bends no rule — D-7 renders the plain text as `String(StyledFormat.attributed(format, [.init(text: word)]).characters)`, so `StyledFormat` is the one parser of a decoration format and `L10n.format` stays in `L10n.swift`. T-7 updated. Not To user: option (a) would have changed the skill; (b) keeps it.

### M-2 answer check · major · Map S14-09 "Not this PR" still names the parity check
**Answer:** Fixed — clause removed from the map's S14-09 scope entry.

### M-note · Next adds promised a `StyledFormat` slot for a styled name
**Answer:** Fixed — Next adds now says Event-title italics would first need `StyledFormat.Argument` to carry an `AttributedString`; not planned.

### X-10 · blocker · T-6 asserts capitalised part text; parts are folded lowercase
**Answer:** Fixed — AC-2.7, T-6 and the Stop condition use `smith`, `jones`, `mary`; T-6 compares `OnlyB` with `slices.Equal`.

### X-11 · major · Steps do not map to the named commits
**Answer:** Fixed — §7: one commit per STEP, each passing its Check, with subjects.

### X-12 · major · Italic check cannot run unattended; D-11 trigger is manual
**Answer:** Fixed — D-11 sets the italic font explicitly through `ConclusionListRow.italicizingEmphasis(_:font:)`, always; T-5 asserts the helper sets `.font` on the `Smith` run only; the Stop condition says an agent that cannot see the app reports "italics unverified" and leaves manual step 3 to the user.

### X-13 · minor · `Equatable` on a type holding `LocalizedStringResource`
**Answer:** Fixed — D-6 drops `Equatable`; `== nil` on the optional needs none.

### X-14 · minor · T-1 must replace `len(types) != 7`
**Answer:** Fixed — T-1 says so.

### X-15 · minor · Empty separator; typed quotes
**Answer:** Fixed — D-8: an empty separator is a one-space run; a value already holding quotes shows as typed, inside the decoration (accepted).

### X-16 · minor · Out row misquotes `macos-client.mdc`
**Answer:** Fixed — reworded: new wire enums go to `Platform/`; an existing one keeps its home.

### X-17 · nit · `.tracking` on `Text(AttributedString)`
Logged.
