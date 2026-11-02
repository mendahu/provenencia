# Review — S13-03 Siblings from a shared parent

**Verdict:** Fits with changes — Ready with changes (both rounds; every finding answered, no blocker open)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None.

## Writer's own calls

- **Siblings come through `Relationships(id, WithSiblings)`, not a read of their own** (M-1). The reviewer sent this To user because it reverses the draft's Out row and D-3. That Out row was the draft's own reading, not the map's: the map's scope entry says "reachable only from an explicit option, not from the pattern set `canonlinks.Reader.For` gives matching", and working rule "One person's links, direct and implied, from one function" makes `Relationships` S13-04's one call. The option matches both, bends no rule, and keeps siblings out of `For`. The scope entry now names the option.
- **No size bound on siblings** (D-5). `MaxPivotEnd` (10) would hide every sibling in a family of 11 children.
- **A relative in two roles is returned in both, unflagged** (D-6, X-2). Cal shows as both Bea's parent and her sibling when the data says so. Whether to flag that stays in `ideas/kinship-view.md` "Conflicts between shapes".
- **A half-sibling is a sibling through the one shared parent** (AC-1.6). Grouping by parent set is S13-04's.
- **`Relationships` takes a plain `Options` struct** (round 2 note): the repo's shape (`subjectprofile.Options`), and its one call site is a test.
- **`PivotKind(shared_parent)` is `Person`** (M-8, X-8). Matching does not read `PivotKind`; only `TestPatterns` calls it.

## Round 1

### M-1 · blocker · Recorded sibling links read twice; S13-04 merges anyway
`kinship.Siblings` re-scanned recorded sibling links that `Reader.Relationships` already returns through `relationshipEdges`, so S13-04 would merge two lists. The reviewer proposed: implied siblings only from `kinship.Siblings`, stored in the slot, and added to `Relationships` behind an explicit page-only option.
**Answer:** Fixed. `Siblings` returns `[]Implied` (D-1, D-3). `Relationships(id, opts ...Option)` with `WithSiblings` (D-11). The Out row is now "Siblings in `Reader.For`". US-1's ACs read sibling links from `Relationships`. The map's scope entry names the option. See Writer's own calls.

### M-2 · blocker · `Siblings` rebuilds `chains`' parent-step index in the same package
`Derive`'s step loop and `chains`' `byPair`, `order`, `childrenOf` and `from`/`assoc` dedup are the same join one generation up.
**Answer:** Fixed. D-4 adds `stepsOf`, `parentSteps`, `indexSteps` and `through`, shared by `chains` and `Siblings`. STEP-2 lands it with no behavior change and no test changed.

### M-3 · major · A person's parents and birth links are found in three or four places
`forLinks` filters its own `births` and walks `relationshipEdges` through `ParentStep`; `siblingLinks` and `Siblings` would each find parents again.
**Answer:** Fixed. D-9's `kinEdges(n, atEnd)` serves `chainLinks` and `forLinks` (its `births` filter goes). D-1 adds `kinship.Parents`, used by both `siblingsOf` (to choose what to load) and `Siblings`.

### M-4 · major · One id per call; the next add batches many ids
Working rule "The family read is the family tree's first step" (3) asks for many ids in a fixed number of queries.
**Answer:** Fixed. D-8's `siblingsOf(ids, …)` loads N ids in four batches; `Reader.LoadSiblings(ids)` stores the answers. AC-2.2 compares 1 person with 3. `For`'s per-person `loadAround` is logged as S13-04's to batch (Out row).

### M-5 · major · Third copy of the check closure and the stored-answer lookup
**Answer:** Fixed. D-7 adds `cachedNodes`, `personAnswer` (replacing `forResult`, shared by both slots) and `Reader.answer` for `For` and siblings. `Reader.pivot` keeps its own lookup, because its result also carries a `kind`.

### M-6 · minor · Two ways to inject loading into the deriver
**Answer:** Fixed for this PR. D-7 adds the `deriver.load` field, set in `newDeriver`. `forLinks`' `around` parameter is logged as S13-04's (Out row), since moving it is not needed here.

### M-7 · minor · Grouping a half-sibling under a parent set needs more than shared parents
**Answer:** Fixed as a note. **Next adds** says S13-04 groups by parent set, reading a half-sibling's other parents from that sibling's own links, and may need one more batch for them, which is S13-04's to count. This is not a product call: R-19 already asks for each parent set.

### M-8 · minor · `PivotKind` says `shared_parent` has no pivot
**Answer:** Fixed. D-2: `PivotKind` returns `Person`. T-2 adds the row.

### M-9 · nit · Second hard-coded memo test helper
**Answer:** Fixed. D-12: `hasMemo(slot, id)`.

### X-1 · major · Whether a missing person stores an answer is not decided
**Answer:** Fixed. D-8 and D-11: an id that is no node gets no entry and stores nothing. Nodes and parents are read through `g.Cached` after the loads on both the read and the check path, so a missing parent is skipped the same way in both. AC-1.9 and AC-2.7 (under verify) cover it.

### X-2 · major · Cycles and relatives in more than one role are not decided
**Answer:** Fixed. D-6 and AC-1.10: such a relative is returned in each role, unflagged; flagging is in the idea file (Out row).

### X-3 · major · T-7's fixture cannot be built from its row alone
**Answer:** Fixed. T-8 (renumbered) spells out `kinTree`'s tree half: helpers, the `"parent"` and `"sibling"` type strings, which `Entity.ID`s to pass, and `f.c.Graph()`. §4 says a relationship filed with `f.relationship` reaches the canon when its people are promoted.

### X-4 · minor · AC-1.8 has no names and no source for the order
**Answer:** Fixed. AC-1.8 names the people `C1` … `C14`; T-3 asserts the order and checks it against `w.g.Cached(ann).Links`.

### X-5 · minor · AC-2.2 can pass trivially when the nodes are already warm
**Answer:** Fixed. AC-2.2 uses fresh catalogs, with no read or commit before the counter is set, and asserts each count is greater than 0.

### X-6 · minor · STEP-6 does not link the plan in the map row
**Answer:** Fixed. The map row's Plan column is linked with this plan; STEP-7 changes only the Status.

### X-7 · minor · `check-pr.sh` cannot pass on Linux
**Answer:** Fixed. STEP-7 and the DoD: on a host without Xcode, report every other check's PASS and `xcode-test` as not run (the PR has no Swift change).

### X-8 · minor · `PivotKind` blank for a person pivot
**Answer:** Fixed with M-8.

### X-9 · minor · AC-2.7 "same slice" has no assertion shape
**Answer:** Fixed by M-1, which removed the case. `Relationships` returns a slice the caller owns, so the old AC-2.7 is gone. AC-1.11 now tests that the caller owns the result, and AC-2.7 is now the missing-parent case.

## Round 2

Round-2 reviewers checked M-1 – M-9 and X-1 – X-9 against the code. All held except M-6 (M-10 below), X-1's AC-1.9 claim (X-1 below) and X-9 / AC-2.7 (X-2 below).

### M-10 · blocker · The M-6 fix added a third way to say whether the deriver may load
D-7's `load` field sat beside `deriver.node` and `forLinks`' `around` parameter. A wrong pairing (`cachedNodes` with `g.Nodes`) would compile and silently load on a check path.
**Answer:** Fixed. D-7: one `nodeSource{node, load}`, built by `loadingNodes(g)` or `cachedNodes(g)`; `newDeriver(g, rules, src)`; `Reader.loadAround` becomes `deriver.around`, and `forLinks` loses its `around` parameter. Both paths behave as today. The Out row deferring it to S13-04 is gone.

### M-note · `Option` value vs the repo's `Options` struct (not raised as a finding)
**Answer:** Fixed. D-11 uses `type Options struct{ Siblings bool }` and `Relationships(id, opts Options)`; the one call passes `Options{}`.

### M-note · `PivotKind`'s doc should name the shared parent
**Answer:** Fixed. D-2.

### X-1 · blocker · AC-1.9's self-sibling claim is false
`LinkSet` does not reject a link whose two ends are the same person, so `Relationships` already returns a filed Bea–Bea sibling link.
**Answer:** Fixed. AC-1.9 now expects that recorded link, unchanged, with nothing implied added. D-6 is corrected: only `ParentStep` rejects one-id ends. `Relationships`' recorded links do not change.

### X-2 · major · AC-2.7 never reaches the "parent not cached" branch
**Answer:** Fixed. AC-2.7 keeps the removal case (no siblings, no error, Verify passes). It now says the not-cached-parent skip is defensive and no test reaches it, because the catalog cannot hold a link to a parent that is not there.

### X-3 · major · D-4 left interface choices open
**Answer:** Fixed. D-4 gives `stepPair` at package level, the `parentSteps` fields, how `through` orders and dedups, the signature of `chains`, and `Derive`'s two `stepsOf` calls concatenated before one `indexSteps`.

### X-4 · minor · AC-3.5's command was not runnable from a table cell
**Answer:** Fixed. The command is a fenced block in STEP-6 with a plain `|`, and it now also matches `kinship.Parents`.

### X-5 · minor · T-8 did not compile
**Answer:** Fixed. T-8 now uses three `f.person` calls, defines `sbea` and `scal`, and names the imports.

### X-6 · minor · D-8 left small choices open
**Answer:** Fixed. D-8 now covers: duplicate ids are read once; an id that is not 16 bytes gets no entry; a node with no siblings stores an empty answer; an uncached event on the check path is an error. D-11 says what happens when there is no stored answer.

### X-7 · minor · AC-1.5 said "sorted" but nothing sorts
**Answer:** Fixed. AC-1.5 expects the shapes in `Parents` order.
