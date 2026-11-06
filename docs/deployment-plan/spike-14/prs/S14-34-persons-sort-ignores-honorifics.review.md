# Review — S14-34 Persons sort ignores honorifics

**Verdict:** Fits with changes — Ready with changes (all changes applied; no open blocker)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. (Answered before review: the working-label title stripping is descoped, user 2026-11-06.)

## Writer's own calls
- A name made only of prefix parts ("Rev.") keeps them in its key, so it sorts under R. The alternative was to fall through to the label, which would push it to the ref-only tail.
- `SortByTitle` computes each key once for all five callers, not only for Persons (M-1).
- `sortKeyOmits` is a list in code, not a profile field. A profile field needs a migration, so it stays with the "Particle sort" idea.
- No precompute of collation keys: computing keys once already measured 0.37 s at 37,000 rows.

## Round 1

### M-1 · blocker · Compute each sort key once, then drop D-3
`SortByTitle` calls `title()` twice per comparison. Measured at Scale (37,000 Persons): 1.25 s per comparison, 0.37 s with keys computed once (`sortByDateKey` pattern). Make `SortByTitle` keyed and drop the allocation special case.
**Answer:** Fixed — D-4, STEP-2, T-4; draft D-3 removed; §6 Scale cites the measurement; Out row now covers only collation-key precompute.

### M-2 · blocker · The name sort key belongs in `nameformats`, not as a private filter in `conclusionsummaries`
`Format` claims to be the sort key; the "Particle sort" idea in `docs/ideas/international-names.md` is the next add to the same key. Add `nameformats.SortKey` with a declared omit list; never read `form`.
**Answer:** Fixed — D-1, D-2, STEP-1, T-1; `withoutPrefixes` and the `Form`-clearing workaround and its stop condition removed; Next adds names the idea.

### M-3 · major · `SortPersons` restates the production sort line; extract it instead
**Answer:** Fixed — D-5: `sortPersons` / `sortPlaces` in `order.go`; `export_test.go` exposes them as vars; the restated `SortPlaces` goes.

### M-4 · major · Docs the change makes false are not all in STEP-3
`Format` doc; `structured-name-model.md` closing paragraph and requirement 5.
**Answer:** Fixed — D-2 (Format doc), STEP-4 and DoD list requirement 2, requirement 5 and the closing paragraph.

### M-5 · minor · Codebase fit misses the in-file precedents
**Answer:** Fixed — Codebase fit rows for `Format` and `sortByDateKey`.

### X-1 · blocker · Most AC cases and tests give the same order with or without the fix
AC-1.1, AC-1.2, T-1 (f) and the manual check ordered the same before and after.
**Answer:** Fixed — AC-1.1 uses Nolan, AC-1.2 uses Dr. Zed / Mr. John, (f) gives the Mr. row the lower ref, the manual check now uses Mrs. Smith / Dr. Zed Smith / Jane Smith / Pat Nolan with the before order stated; STEP-3's check names which cases fail on revert.

### X-2 · blocker · T-2 has no recipe for building a name that has a prefix part
**Answer:** Fixed — T-3 names the construction (inline `namevalue.Value` with `Form` and `Idx`, a local closure, no `namevaluetest` helper) and asserts that the listed name keeps its prefix.

### X-3 · major · Definition-of-done checks I cannot run here
`check-pr.sh` needs macOS for its Xcode steps.
**Answer:** Fixed — DoD says to report the Go results and "Xcode steps not run" when `xcodebuild` is absent.

### X-4 · minor · Smaller choices left open
**Answer:** Fixed — STEP-4 names the version files and the next patch, the `completed.md` position and heading, and the map status `Landed #N`.

## Round 2

### M-6 · blocker · The round-1 fix copies `sortByDateKey`'s keyed sort instead of sharing one
A generic `sortByKey[T, K]` serves both `sortByDateKey` and `SortByTitle`; the "key is a string" reason for copying was wrong.
**Answer:** Fixed — D-4 rewritten (`sortByKey`, `titleKey`, `titleKeyLess`); Codebase fit and Deletes updated; STEP-2 order.

Reviewer notes below blocker, taken: Reach now lists `PersonsByIDs` → `persondetail` participant order (wanted, covered by the `persondetail` tests in the checks); Next adds no longer claims `surname_prefix` becomes a `sortKeyOmits` entry (the idea needs a per-profile rule); the `sortKeyOmits` row cites `seedProfiles` Roles and says why `RoleIgnored` is not reused (it covers `surname_prefix`).

### X-5 · blocker · T-2 does not say where `SortPersons`'s `order` argument comes from
**Answer:** Fixed — T-1 and T-2 load the Western order through `Install` + `Load` / `LoadFor`; no order literal; case (c) passes `nil`.

### X-6 · major · The sketch calls an undefined `dropTypes`
**Answer:** Fixed — the sketch inlines the filter into a new slice with `slices.Contains`; `name.Parts` is not changed.

### X-7 · major · The PR number `#N` is needed in STEP-4 before the PR exists
**Answer:** Fixed — STEP-4 says `#N` is the draft PR's number (the docs commit comes after `implement-pr` opens it), the date is the commit date, and the scope entry's Context link to the deleted note is rewritten the way S14-33's is.

### X-8 · minor · The plan disagrees with itself about how the splitter reads "Mrs. Smith"
**Answer:** Fixed — the manual check states the parts the splitter stores (from the dogfood note's code notes). §1, the Out row and §11 agree: today it is prefix + given; prefix + surname is the Out change.

### X-9 · minor · The manual check and `check-pr.sh` cannot run here
**Answer:** Fixed — the DoD's PR-description box allows "manual check not run: no macOS".

### X-10 · nit · The STEP-3 revert claim is hard to act on
Logged.
