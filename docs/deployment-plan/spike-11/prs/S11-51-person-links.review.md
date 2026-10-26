# Review — S11-51 One person's links for every reader

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- M-3: where does the one link merge live? Answered 2026-10-26: `core/kinship`, not `canonlinks` as the map said. The map's working rule, the S11-51 row and scope entry, and Spike 13's rule follow (D-8).

## Writer's own calls
- The read is `Reader.Relationships` and returns only `relationship`-bridge links plus `For`'s implied links; participation links stay in the evidence query's own `canonLinks` (D-4).
- `Relationships` does not filter merged or unaccepted neighbors and is not memoized (D-5).
- `PlaceImplied` returns indexes instead of merging, so `loadCanon` keeps F-9's cost shape (D-3).

## Round 1

### M-1 · blocker · D-3 keeps a second merge in `loadCanon` and widens the exports to keep it alive
`loadCanon`'s `impliedAt` loop is a second same-ends, same-keys merge; the first draft exported `LinkKey`, `LinkKeyOf` and `IndexOf` to keep it outside the set. Better: a `LinkSet` method that places implied links without growing the set.
**Answer:** Fixed — D-3 adds `LinkSet.PlaceImplied`; the key stays unexported (D-1); AC-2.6 and T-2 cover it; `impliedAt` joins the STEP-2 grep.

### M-2 · major · `Reader.Links` would be the third copy of the relationship-links loop
`chainLinks` and `forLinks` already filter `n.Links` to relationship links through `d.edge`.
**Answer:** Fixed — D-6 `deriver.relationshipEdges`, used by all three.

### M-3 · major · A pure merge over `kinship` types goes into a database package the Source loader also uses
**Answer:** To user → answered: `core/kinship` (D-1, D-8).

### M-4 · minor · `Link` / `Links` collide with `graphcache.Link` and `Node.Links`
**Answer:** Fixed — the type is `kinship.Link` (another package), and the read is `Reader.Relationships` (D-4).

### M-5 · minor · The S11-40 handoff and the map row
**Answer:** Fixed — Next adds says S11-40 replaces `canonLinks`' relationship links with `Relationships(from)` kept to `to`, participation links stay; D-8 rewrites the S11-51 row.

### M-6 · nit · `spouse` helper copies `parent`
**Answer:** Fixed — `world.relationship(name, a, b, term)`; `parent` calls it.

### X-1 · blocker · The names grep cannot print nothing
`core/database/effects/evidence_graphs_test.go` has a local `linkSet`.
**Answer:** Fixed — the STEP-2 grep covers `core/kinship`, `core/database/canonlinks`, `core/database/promotealign`; the STEP-4 grep names the four docs.

### X-2 · major · Unused imports after the move
**Answer:** Fixed — D-7: each touched file drops the imports it leaves unused, and only those.

### X-3 · major · Orientation of the returned link is not fixed
**Answer:** Fixed — T-1: AC-1.1 and AC-1.4 compare ends as a set; AC-1.2 and AC-1.3 in order.

### X-4 · minor · Keys for parent and grandparent
**Answer:** Fixed — AC-1.2, AC-1.3 name `w.keys(relationship, parent)` and `w.keys(relationship, grandparent)`, A the parent end.

### X-5 · minor · Scope rows point at STEP-2
**Answer:** Fixed — they point at STEP-1.

### X-6 · minor · Map row names the evidence query
**Answer:** Fixed — D-8 (same as M-5).

### X-7 · minor · Other stale mentions
**Answer:** Fixed — Out line: landed plans, logs, `completed.md`'s S11-39 paragraph and the S11-39 row and scope entry stay; line 104 is rewritten (D-8).

Nits (logged): X-8 · `Links()` on an empty set is nil (now in D-2); the `spouse` term exists in `subjectvocab/registry.go`, and a stop line covers it.

## Round 2

Both reviewers confirmed every round-1 answer in the plan.

### M-7 · blocker · The implied link's recording is built in two packages
`addImplied` and `loadCanon` (`canon.go:212`) each write `Recording{Shapes: []Shape{imp.Shape()}, OnlyImplied: true}`.
**Answer:** Fixed — D-3a `Implied.Recording()` beside `Implied.Shape()`; `AddImplied` and `loadCanon` use it; Codebase fit row; T-2 subtest.

### X-9 · major · `TestLinkSet`'s `rules` source
**Answer:** Fixed — D-7: the package-level `terms` and `rules` of `kinship_test.go`; the local closures stay and shadow.

### X-10 · major · D-8 gives meaning, not text
**Answer:** Fixed — D-8 gives the replacement text for line 104 and Spike 13's line 64; the domain docs change only the clause naming the old home.

### X-11 · minor · AC-1.5 / AC-1.6 inputs
**Answer:** Fixed — 16 zero bytes; deep copies first, link 0 and every association overwritten.

### X-12 · minor · S11-40's scope entry
**Answer:** Fixed — Out line: S11-40's plan words the split.

Nits (logged): X-13 · `PlaceImplied`'s index rule now says `j` is 0-based and `at[i] == len(edges)` means append.
