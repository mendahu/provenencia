# Review — S16-20 One match index that Apply keeps current

Plan: [`S16-20-match-index-apply.md`](S16-20-match-index-apply.md). Planned against `64a91ae`.

## Before the review

- User, 2026-11-13: bar 10's "first proposal after a handle commit no slower than 1.5 times a warm one" moves from S16-20's Check to S16-32's. Measured on the planning base (scratch benchmark, Xeon 2.1 GHz): with the index kept, re-ranking Source A's candidates costs 723 ms against a 226 ms warm proposal at 4,000 persons, 1,187 ms against 226 ms at 16,000.

## Round 1

Maintainability: **Fits with changes**. Executability: **Not ready**.

| ID | Sev | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | Build and Apply each restate the profiler's build order; Apply writes the inverse of `build`'s read by hand, keyed on `kind == Event` | Fixed — D-3: `profiles.go` owns `subjects` (the order) and `readers` (the inverse read, driven by the `Lifespans` rule); D-2 and D-5 call them |
| M-2 | major | `hopPersonToEvent` copies `hops.EventsOfParticipant` | Fixed — D-3 uses `hops.EventsOfParticipant` and `hops.ParticipantsOfEvent`; **Deletes** lists the private hop |
| M-3 | major | `add-match-rule` step 2 and `TestShortlistSection` justify the lead rule with "`keyValues` builds no life", false after D-8 | Fixed — D-8 keeps the rule with its real reason (a derived lead's keys would move with every event edit); STEP-7 rewords the skill and the test; T-10 |
| M-4 | major | `promote-graph-alignment.md` line 242 and `catalog-caches.md` line 5 become false | Fixed — STEP-7 edits both; sweep gains "drops the shortlist index" |
| M-5 | minor | `propose` becomes a forwarding wrapper | Fixed — D-9: `Propose` absorbs it |
| M-6 | minor | `profiler.eventTypeProp` loses its reader | Fixed — D-3 drops the field, D-9 has `loadCanon` read it. The `canonlinks`-owned lookup is an Out row → S16-21 |
| M-7 | minor | Three direct `ReadWhole` calls where siblings have one named read | Fixed — D-2 keeps `matchIndexOf(g) (*matchIndex, error)` over `ReadWhole`, used by `Warm`, `Propose`, `checkCandidates` and Apply |
| M-8 | minor | `filed` repeats each handle's keys | Declined — the strings share their bytes with `byKey`'s map keys, so it costs slice headers, which STEP-6 measures in `matchIndex-B`; taking the caller's old values would let one wrong caller corrupt the postings. Rejected column of D-4 says so |
| X-1 | blocker | `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` forbids D-9's always-read | Fixed — D-9 reads the entry always (it is built at open); D-14 deletes the test with the other lazy-fill test |
| X-2 | blocker | `TestDoKeepsWarmEntries` "handle commit" expects the index dropped | Fixed — D-14 flips the case and its comment; D-5 states a nonexistent id is a no-op; STEP-4 and T-9 |
| X-3 | major | D-5 classifies roots before loading them | Fixed — D-5 gives the order: load roots, `readers`, load the people |
| X-4 | major | T-3 fixtures missing (birth Subject discarded, no place, name edit inline) | Fixed — D-14: `warmFixture.memberBirth`, `renameMember`; T-2 / T-3 name the place and the date Cite |
| X-5 | major | `loadCanon`'s node map and Handle sharing undecided | Fixed — D-9: the node map stays; Handles are shared, not copied; T-2 runs Verify after two proposals |
| X-6 | major | Sweep terms hit unrelated docs; line 242 missing | Fixed — sweep scoped to named paths and qualified terms; line 242 in STEP-7 |
| X-7 | minor | STEP-7 quotes the map row wrongly; `structural-debt.md` row no longer exists | Fixed — STEP-7 quotes the row and the entry separately; Current state notes the row moved |
| X-8 | minor | `Put` with no keys or another kind; deleted event | Fixed — D-4 states both; T-4 rows; T-3 "an event deleted"; Edges |
| X-9 | minor | Figures under Verify; rg pattern incomplete; `session.go` line 112 | Fixed — STEP-6 and §11 say Verify unset; STEP-3's `rg` adds `ProfilesKeptForTest` and `hopPersonToEvent`; STEP-7 adds the `openResearcher` comment |

## Round 2

Maintainability: **Fits with changes** — every round-1 answer holds.

| ID | Sev | Finding | Answer |
| --- | --- | --- | --- |
| M-10 | blocker | D-2 called `subjects` once per kind, so Build would profile persons before events and give every person an empty life; Check reuses Build, so Verify would not catch it | Fixed — D-2: one `subjects` call over all three kinds, never per kind; T-2 asserts the person's lifespan lead is non-empty after `Warm` |

Executability: **Not ready** (one new blocker); every round-1 answer it checked holds.

| ID | Sev | Finding | Answer |
| --- | --- | --- | --- |
| X-10 | blocker | With a Build, `MatchIndex.Get` rebuilds on a miss, so every "emptied" assertion (`stored`, `warm_test.go`) loses its meaning; `TestNameFormatDropsItsReaders` was not in the plan | Fixed — D-15: `graphcache.Whole.Peek` (never builds), used by every held probe in `memo_test.go` and `warm_test.go`; Current state lists the `warm_test.go` tests; STEP-3 adds it with a `cache_test.go` case |
| X-11 | major | STEP-3 applies D-14 while Apply still drops | Fixed — STEP-3 excludes the two handle-commit flips, which STEP-4 makes |
| X-12 | major | Probe Values for `IndexedForTest` unspecified | Fixed — `HandleForTest` hook; T-3 and AC-1.1 probe with the held profile before and after |
| X-13 | major | T-3 / T-6 per-row choices | Fixed — each row names its helper, event type, year, role, term key and committed ids |
| X-14 | minor | STEP-3 is large | Fixed — STEP-3 states an order within the step with a build check after each part; not split, since the parts compile only together (the old profiler's callers go with it) |
| X-15 | minor | Map Ships sentence quoted badly; row status value unstated | Fixed — STEP-7 quotes the whole sentence and the status |
| X-16 | nit | `registry_test.go` line 46; `CandidatesHeldForTest` not listed | Fixed |

Round 2 is the last round. No blocker is open.
