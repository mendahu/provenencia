# Conformance and regression lens

The question: **does the change do what the plan says — every acceptance
criterion, nothing outside scope — and does everything that worked before
still work?** This lens always runs. It catches more than any other.

## Read first

- The plan (`prs/S#-NN-<slug>.md`): **User stories** and their ACs, **Scope**
  (In and Out), **Decisions**, **Tests**, **Definition of done**.
- The fix plan for this round in the ship log (`S#-NN-<slug>.ship.md`), if any.
- The diff against its base: `git diff <base>...HEAD`, and `git log <base>..HEAD`
  for what each commit claims.
- The domain doc the plan's **Context** links, for the rules the behavior must keep.

## How to work

1. Build a table for yourself: each AC → the code that makes it true → the test
   that proves it. Any AC without code or without a test is a finding.
2. Walk the diff file by file. Each hunk must trace to an In line, a Step, or a
   Decision. A hunk that traces to nothing is scope creep; one that implements
   an Out line is a blocker.
3. For each changed function, type, query, or view: find its **other callers**
   (`rg` the symbol). Does the change hold for them? This is where regressions
   live.
4. Read the new and changed tests: do they prove the AC, or only that the code
   runs?
5. You may run targeted Go tests (`CGO_ENABLED=1 go test -tags fts5 ./core/<pkg>/...`).
   Do not run `xcodebuild`; the gate already did, and parallel builds collide.
   Do not edit files.

## Look for

### Plan conformance

- An AC that is not met, or met only for the example in the plan and not the rule.
- A Decision the code ignores or swaps for a different design.
- An Out line implemented "while I was there".
- A Step's files untouched, or files changed that no Step names, without a reason
  in the commit.
- A Definition of done line left undone: `completed.md`, map status, plan
  **Landed** note, brief archived, domain doc, version bump as the header says.

### Functional regression

- A shared function, query, or component whose contract changed (arguments,
  ordering, nil/empty handling, error codes) with callers that assume the old one.
- A changed default that every existing call site now inherits.
- Ordering and identity: a list now sorted differently, an id that changes on
  save, a key that is no longer stable for `ForEach` or a cache.
- Edge cases the old code handled: empty, one, many, very long text, missing
  file, deleted referent, a catalog created before this PR (no migration, no
  seed), a project opened by another user.
- Persistence: does a value written by the old code still read back? Does a
  value written by the new code break an older reader in the same release?
- Audit and delete: a new write without its audit effect, a new reference a
  delete probe does not see ([`add-audit-scope`](../../add-audit-scope/SKILL.md),
  [`add-catalog-delete`](../../add-catalog-delete/SKILL.md)).
- FFI: a proto field renumbered or reused, a method whose Swift and Go sides
  disagree, an error code with no Swift mapping.

### Rules and algorithms

For every rule, comparison, ranking, fold, key or algorithm the diff adds or
changes. Read the code the new case flows through, not only the diff.

- **Assumptions a new case breaks.** List what the surrounding code relies
  on: transitive, acyclic, symmetric, order-free, one key per value, two lists
  of equal length, a result with a bound. Code written when every case had a
  property keeps relying on it after a new kind of case arrives without it.
  Test each property with the new case.
- **Excluded is not neutral.** When a decision, a filter, a delete, a merge,
  or another item's claim removes an input from a computation, ask what that
  input contributed (a conflict, a constraint, a count) and whether removing
  it flips the result. Test the result with the input decided, excluded and
  held elsewhere, not only present.
- **Results that depend on changing state.** For each cached, memoized or
  queued result, list the state it was computed from. Every change to that
  state must recompute or invalidate it; a score kept from before a
  dependency changed reaches the output stale.
- **Claims over all inputs.** "The order does not matter", "accepting the
  output changes nothing", "equal keys mean equal values", "compare is
  symmetric": prove or refute these with generated inputs (a seeded random
  generator or a fuzz target over many small cases), not a few hand fixtures.
- **A rule a table drives.** When behavior comes from a table, seed or
  configuration, run the rule over the whole shipped table and read what it
  produces: the counts and the surprising rows. A review of the plan's three
  examples does not review the other thousand.
- **Production configuration in tests.** Invariant, fuzz and property tests
  run at least once with the shipped tables and settings. A minimal fixture
  that omits the new data never reaches the new path.
- **One equivalence, two implementations.** When two layers each implement
  "the same" (a normalizer in the engine and in the client, a fold in code and
  in the database's text search), check them against one table of hard cases
  chosen to differ: decomposed and composed text, compatibility characters,
  invisible characters, letters a library folds and the other does not.
- **Neighbors of a widened rule.** A rule that now accepts more reaches
  inputs next to its examples. List the nearest ones it now also reaches and
  check the plan or the user accepted them; one nobody looked at is a finding.

### Tests

- Domain changes in `core/` without Go tests; new store APIs without
  `FakeStore` and model tests; handlers without `runRPC` coverage of the happy
  path and the important errors.
- Go tests that are not table-driven (`.cursor/rules/go-tests.mdc`).
- Tests that assert raw English instead of `L10n` lookups; Mac unit tests that
  reach the dylib or real SQLite.
- A test that would still pass if the feature were deleted.
- A migration tested directly (the repo forbids it — [`add-catalog-migration`](../../add-catalog-migration/SKILL.md));
  the behavior it enables must be tested instead.
- Do not demand line coverage. Demand coverage on boundaries and rules that
  edits tend to break.

## Evidence: write the failing test

For every functional or regression finding, include the **test that fails
today**: file, test name, and the code, small enough to paste. The fix round
adds it first and confirms it fails before the fix. If it passes before the
fix, the finding was wrong and is dropped. A regression finding without a
failing test is at most **minor**, unless the path is a UI interaction no unit
test can reach; then give exact manual steps.

## Severity in this lens

| Severity | Use when |
| --- | --- |
| **blocker** | An AC not met; an Out line implemented; a regression in an existing flow; data written wrong or lost; a Decision ignored |
| **major** | A missing test for an AC or a boundary; an edge case the old code handled; a DoD docs line skipped |
| **minor** | A weak test; scope noise with no behavior change |
| **nit** | Test naming, commit wording |

## Known traps

Classes that reached `main` past this lens and were found later, by an
app-health audit or a landing. Each line names the class, the shape it took,
and what to check. Add a line when an audit traces an item back to this
lens ([`review-pr`](../SKILL.md) "Close the loop").

- **A new way into old code.** A PR that adds a path into an existing
  destination (an omnibar hit, a history entry, a new button) or makes a state
  reachable (no links, a deleted referent, many where there was one) is
  checked at the destination: does the view reset on the new id, handle the
  missing referent, show the state? Once S12-15 let the omnibar open canvas
  after canvas, the canvas's `@State` model kept the first graph and wrote to
  it. S12-13 let a card uncheck every graph, and a Citation with no
  Observation became unreachable.
- **An old name left behind.** For each symbol the diff renames or deletes,
  `rg` the old name, qualified (`Effect.Source`), across `.agents/`,
  `.cursor/rules/` and `docs/*.md`. `scripts/check-guidance-names.py` catches
  the distinct names, not the short ones. A hit in a skill or rule is a
  blocker: it trains the next agent wrong.
- **A re-scoped action, unchanged copy.** When a PR changes what an action
  covers (Promote per Source → per graph), `rg` `L10n.swift` in the touched
  features for the old scope's word. "Strings: none" in a plan is a claim to
  check.
- **The last caller removed.** For each call the diff deletes, `rg` the
  callee. One with no caller left goes in this PR: a kit component kept after
  a revision removed its second caller, a session accessor, two RPCs whose
  only caller a rewrite replaced.
- **A fix beside the thing it fixes.** A narrower mechanism (a by-Source
  read) retires the wider one it replaces (the catalog-wide list key), or both
  run and the fix does nothing.
- **A lock taken twice.** A new function that acquires a lock, a decode slot
  or the session that its callers may already hold deadlocks at the smallest
  pool. Name who holds it.
- **A new kind of case in an old loop.** Two new fold kinds reached a loop
  that assumed folds chain; three ordinary values with no chain left a nil
  map and the process died. The invariant test ran without the shipped
  tables that create the new kind, so it never generated one. See "Rules and
  algorithms".
- **A decision that drops evidence.** Marking an item decided removed it from
  scoring, and with it the one conflict that kept a neighbor from matching
  the wrong record. The row then read as agreement. Check what a skipped input
  contributed.
- **A table nobody read.** A rule took its direction from a third-party list
  and was reviewed from its examples; run over the whole list it reversed
  hundreds of entries.

## Good finding shape

The shape to copy. Names in the example are illustrative.

"AC-2.1 says a filed Source drops out of the In-progress filter. The query in
`core/database/graphprogress` counts a primary as filed when *any* claim
is accepted, not when every primary has one. Failing test:
`TestEvidenceStatus/one_of_two_primaries_filed` → want In progress, got Filed."
