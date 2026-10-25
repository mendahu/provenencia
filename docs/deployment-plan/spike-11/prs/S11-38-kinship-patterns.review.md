# Review — S11-38 Kinship patterns

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None.

## Writer's own calls

- **`core/kinship.Edge` mirrors `graphalign.Bridge`'s fields exactly** (D-1, after M-1), but is its own type, so the package that trees and person pages read later does not depend on layer 6.
- **The chain is one rule in code, not a table** (D-10, M-2 option (b)). The reviewer recommended a two-step row with middle ends (option (a)) so siblings and great-grandparents could be rows; the writer chose (b) because no roadmapped pattern needs that shape and the working rule stops at grandparent. Only event patterns are table rows, and the skill says so.
- **One implied link per event and pair of participation edges** (D-5, X-3): the same couple in two marriages gives two links, which the loaders' merge (S11-39) joins.
- **One grandparent link per triple** (D-6), gathering every parent step behind it, so a parent recorded in two shapes does not double the chain.
- **Self-links are dropped** (D-7) in every pattern and as chain steps.
- **A `child` birth role implies nothing** (D-5): [kinship shapes](../kinship-shapes.md#open-questions) open question 3, as proposed there.

## Round 1

### M-1 · major · Give `kinship.Edge` the exact field set of `graphalign.Bridge` (`Bridges` → `Associations`)
**Plan:** D-1 named the field `Bridges`. **Codebase:** `core/graphalign/types.go` `Bridge.Associations`; `Layer.Bridges` and `connectrules.Bridge` already use "bridges" for other things.
**Better shape:** name it `Associations`, field order `A, B, KeyA, KeyB, Associations`, so S11-39 converts with `kinship.Edge(b)` / `graphalign.Bridge(imp.Edge)`.
**Answer:** Fixed — D-1, D-5, D-6, the §6 sketch and Codebase fit row, AC-1.1 – AC-1.16, T-2 (b), STEP-5 bullet and Next adds.

### M-2 · major · `chainPatterns` rows fit only parent-of-parent, yet US-2 and the skill promise "a new pattern is one row"
**Codebase:** `docs/ideas/kinship-view.md` names siblings and great-grandparents, neither a `{step, term}` row.
**Better shape:** (a) a two-step row with middle ends, or (b) drop `chainPatterns` and keep one chain rule.
**Answer:** Fixed with (b) — D-10 drops the chain table; US-2, §6, STEP-2 and the skill section say only event patterns are rows and another chain shape changes `Derive`. (a) declined for now: no map row or roadmapped idea needs it, and the working rule "nothing deeper than grandparent" leaves great-grandparent out of this spike.

### M-3 · minor · `core/kinship` reaches the link reading through the whole compare stack
**Answer:** Fixed by logging — a "Link reading below the layers" line in [`ideas/kinship-view.md`](../../../ideas/kinship-view.md). Not this PR: S11-41 just placed the reading in `segmentcompare`, and the working rule names it.

### X-1 · major · STEP-5's "Layer 5 bullet" is not in `match-rules.mdc`
**Answer:** Fixed — STEP-5 puts the bullet in the skill's "Rules that hold everywhere" list; the `.mdc` gets only the glob. AC-2.3 lists each skill change.

### X-2 · major · Spouse pair order versus `From` / `Associations` order
**Answer:** Fixed — D-5: for spouses, `From` and `Associations` follow the link's `A`, `B` after D-4; AC-1.15 (input with `P2` first).

### X-3 · major · Duplicate or repeated participation edges
**Answer:** Fixed — D-5: one link per event, `A` edge and `B` edge; AC-1.16 (one couple, two marriages).

### X-4 · minor · Chain self-edges, triple order, nil versus empty
**Answer:** Fixed — D-7 covers self steps; D-11 defines the nested loop order, `nil` for no output, `nil` `Event` / `Through`, non-empty chain `Associations`.

### X-5 · minor · No top-level `rel` helper in `segmentcompare_test.go`
**Answer:** Fixed — Codebase fit row names the closure in `TestKey`, and `TestInvariants` as the model for the seed and loop only.

### X-6 · minor · Test ids have no byte values
**Answer:** Fixed — US-1 fixtures: people `0x11…`, events `0x21…`, bridges `0x31…`.

### X-7 · minor · `vocabkey` const block layout
**Answer:** Fixed — D-9: `TermPartOf` becomes a `const (…)` block with a comment per group; `gofmt` re-aligns.

### X-8 · minor · Numbers guard and the pairing loops
**Answer:** Fixed — the Risks row names the literals to avoid and says test files are not checked.

### X-9 · minor · `check-pr.sh` needs Xcode
**Answer:** Fixed — STEP-6 Check says what to report on a host with no Xcode.

### X-10 · minor · PR number before the PR exists
**Answer:** Fixed — STEP-6 opens the draft PR after the code commit, then makes the docs commit. Status is now Planned.

### X-11 · nit · AC-2.3's grep is loose
**Answer:** Logged; AC-2.3 now lists each change the grep must show.
