# S11-41 — One link reading for bridges

**Status:** Landed in [#561](https://github.com/mendahu/provenencia/pull/561) (2026-10-25)
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an engine refactor with no behavior change)
**Board:** —
**Depends on:** —
**Blocks:** S11-38
**Branch:** `s11-41-link-reading`
**Version bump:** none — internal refactor, no behavior change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (layer 5 rules; STEP-4 edits it)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [Kinship shapes § Shape](../kinship-shapes.md#shape) ("keyed through the same link reading the loaders use"); [segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6); working rules: "High tolerance for churn", "Pure means the package builds and tests with no build tags and no `core/database` import", "One guard, not one per package", "One pure package, `core/kinship`"
**Planned:** 2026-10-25 against `main` @ `74bb416`
**Landed:** None.
**Review:** [`S11-41-link-reading.review.md`](S11-41-link-reading.review.md) — Fits with changes / Ready with changes, 1 round (findings fixed or logged)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Reading a bridge as a `segmentcompare.Link` moves out of `promotealign` into `core/segmentcompare`, beside `Key`, so every reader of a bridge builds its link the same way. Today the one builder, `linkOf`, and its helper `eventEndpoint` are private to `core/database/promotealign`, so the pure `core/kinship` package (S11-38) cannot reach them and would have to build its implied links' keys by hand. The layer loader, `loadCanon`'s `canonKeys` and, through `canonKeys`, the evidence query's `canonLinks` call the moved functions. Nothing changes for researchers.

**Why now.** S11-38 keys every implied link through this reading and `segmentcompare.Key` ([kinship shapes](../kinship-shapes.md#shape)); the slice 9 map review added this PR so that reading has a pure home first (K-4, [`deployment-plan.review.md`](../deployment-plan.review.md)).

**What stays.** Every key the loaders produce, every proposal, band and evidence line. `canonKeys`, `loadLayerBridges`, `edgeSeenKey` and `observedTerms` stay in `promotealign`: they read the cache.

## 2. User stories

**US-1.** As a contributor writing `core/kinship`, when I need the key a direct link between two Subjects would have, I get both of its keys from `segmentcompare.Keys` (built on `LinkOf` and `Key`), so an implied link keys exactly like a recorded one from each end and the two readings cannot drift.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `LinkOf(participation bridge, "subject", "birth", firstEnd true)` | `Link{Type: participation, Term: subject, End: person, Other: event, EventType: birth, SameType: false}` |
| AC-1.2 | Same bridge, `firstEnd false` | `End: event, Other: person`; everything else as AC-1.1 |
| AC-1.3 | `LinkOf(relationship bridge, "parent", "birth", firstEnd true)` | `EventType` is empty (only a participation keeps one); `End: person, Other: related_to, SameType: true` |
| AC-1.4 | `EventEndpoint` on the participation, relationship, place relationship and location bridges | `(false, true)`, `(false, false)`, `(false, false)`, `(true, true)` |
| AC-1.6 | `LinkOf(location bridge, "", "birth", firstEnd true)` | `End: event, Other: place, SameType: false`, `EventType` empty: the location bridge has an event end, but only a participation keeps the event type |
| AC-1.7 | `Keys(relationship bridge, "parent", "", segmentcomparetest.Terms())` | `first` = `{relationship, child, End: related_to}`, `second` = `{relationship, child, End: person}`: "A parent of B" reads as `child` with the ends swapped, from each end |
| AC-1.8 | `Keys(participation bridge, "subject", "birth", terms)` | `first` = `{participation, subject, End: person, EventType: birth}`, `second` = `{participation, subject, End: event, EventType: birth}` |
| AC-1.5 | `segmentcompare` imports `core/connectrules` | `TestLayerPackages` still passes: no build tag, no path to `core/database` |

**US-2.** As the engine, after the move, Promote and the evidence query read every bridge exactly as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `CGO_ENABLED=1 go test -tags fts5 ./core/...` | Passes with no test changed outside `core/segmentcompare` |
| AC-2.2 | `rg -n 'linkOf|eventEndpoint' core/database` | No match |

## 3. Scope

| In | Covered by |
| --- | --- |
| `LinkOf` and `EventEndpoint` *(new, exported)* in `core/segmentcompare/key.go`, with the bodies of `linkOf` and `eventEndpoint` from `core/database/promotealign/terms.go` (D-1, D-2, D-3) | STEP-1, AC-1.1 – AC-1.6 |
| `Keys` *(new)* in `core/segmentcompare/key.go`: a bridge's keys from its first and its second endpoint's end (D-8) | STEP-1, AC-1.7, AC-1.8 |
| Table tests for all three in `core/segmentcompare/segmentcompare_test.go` | STEP-1, T-1, T-2 |
| `loadLayerBridges` and `canonKeys` call `EventEndpoint` and `Keys`; `linkOf` and `eventEndpoint` deleted from `core/database/promotealign/terms.go` | STEP-2, AC-2.1, AC-2.2 |
| The two "linkOf alone decides" comments name `segmentcompare.Keys` | STEP-2 |
| `add-match-rule` layer 5 bullet names the link reading | STEP-3 |
| Docs owed (map row, `completed.md`, this plan) | STEP-4 |

| Out | Home |
| --- | --- |
| `core/kinship`, its patterns and its `vocabkey` term constants | S11-38 |
| Any caller deriving implied links (layer loader, `loadCanon`, `canonLinks`) | S11-39, S11-40 |
| Moving `canonKeys`, `edgeSeenKey`, `observedTerms` or `loadTerms` | not planned — they read `graphcache` and belong in the adapter |
| A `core/connectrules` row in the guard's `layerPackages` | not planned — D-4 |
| Changing `ReadFrom` or `Key`. `ReadFrom` takes a bridge's endpoints unpacked while `LinkOf` takes the `Bridge` (review M-3) | not planned here — logged in the review log for S11-40's planner, whose PR reworks `evidence.go`'s relationship reading |
| `core/database/identityclaims/bridgefile.go` `sameEndpointType`, a second copy of the same-type rule (review M-2) | not planned — logged in the review log |
| Rewriting hand-built `segmentcompare.Link` literals in tests (`core/graphalign/align_test.go`, `TestKey`, `TestReadFrom`) | not planned — tests build links by hand on purpose |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Link builder | `core/database/promotealign/terms.go` `linkOf(b connectrules.Bridge, term, eventType string, firstEnd bool) segmentcompare.Link` | Reads a bridge from its first endpoint's end or its second: `End`/`Other` from the endpoints' `PropertyKey`, swapped when `!firstEnd`; `SameType` when both endpoints share `TypeKey`; `EventType` only for `vocabkey.BridgeParticipation`. Private, so a pure package cannot call it |
| Event end | `core/database/promotealign/terms.go` `eventEndpoint(b connectrules.Bridge) (first, ok bool)` | `(true, true)` when `Endpoints[0].TypeKey == vocabkey.Event`, `(false, true)` when `Endpoints[1]` is, else `(false, false)`. Private |
| Layer caller | `core/database/promotealign/layer.go` `loadLayerBridges` | Calls `eventEndpoint(b)` once per bridge type and `linkOf(b, term, eventType, true/false)` per edge; `term` and `eventType` are `string`. Comment above the `eventEndpoint` call: "linkOf alone decides whether the key keeps it." |
| Canon caller | `core/database/promotealign/canon.go` `canonKeys` | Same two calls, with `l.TermKey` (`string`) and `eventType` from `keptTerm`. Same comment |
| Evidence caller | `core/database/promotealign/evidence.go` `(*evidenceReader).canonLinks` | Reads each filed link through `canonKeys`; no direct call to `linkOf` |
| Destination | `core/segmentcompare/key.go` | Holds `Link`, `Term`, `Terms`, `Key`, `Corresponds`, `ReadFrom`. Imports only `core/vocabkey` |
| Package imports | `core/segmentcompare/segmentcompare.go` | `propertycompare`, `subjectcompare`, `valuecompare`, `vocabkey`. Not `connectrules` |
| Bridge spec | `core/connectrules/connectrules.go` `Bridge`, `Endpoint`, `LookupBridge`, `Bridges` | Pure: imports `strings`, `sync`, `core/vocabkey` only, no build tags. Every product bridge has two endpoints; in `productBridges`: participation is person → event, relationship person → related_to (both person), location event → place, place relationship from → to (both place) |
| Guard | `core/matchregistry/guard_test.go` `TestLayerPackages` | `core/segmentcompare` row with `numbers: true`; the walker follows imports transitively, so a `connectrules` path to `core/database` would fail it |
| Tests | `core/segmentcompare/segmentcompare_test.go` `TestKey`, `TestReadFrom` | External package `segmentcompare_test`, table-driven, build links by hand with `segmentcomparetest.Terms()`. No test of `linkOf` exists anywhere |
| Tests | `core/database/promotealign/promotealign_test.go` `TestLayerBridgesCoverEveryConnectRule`, `TestProposeInverseKinshipTermsCorrespond`, `TestProposeRelationshipFromEitherEnd`, `TestProposeChildOfReadsAsParentOf`, `TestProposeBirthOnAnotherPersonSubtracts`; `evidence_test.go` | Pin the keys and correspondences the two loaders produce today |

Verified at the planning base: adding `import "github.com/mendahu/provenencia/core/connectrules"` to `core/segmentcompare` builds `./core/...` with `-tags fts5` and `TestLayerPackages` passes (probe, reverted).

**Precedent.** `10dec362` ("Read each relationship from the neighbor's side in the engine…", S11-33) added `segmentcompare.ReadFrom` beside `Key` for a reading `promotealign` needed, with its table test `TestReadFrom`. Copy its placement (in `key.go`, after `Key`), its doc-comment style and its test shape; this PR differs only in moving an existing function instead of writing a new one.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Names: `LinkOf` and `EventEndpoint`, in `core/segmentcompare/key.go`, after `Key` and before `Corresponds` | The map's scope entry puts them beside `Key`; the names are today's, exported | `ReadLink` / a method on `connectrules.Bridge` — `connectrules` is a lower package many callers import, and must not import `segmentcompare` |
| D-2 | Signature `func LinkOf(b connectrules.Bridge, term vocabkey.BridgeTerm, eventType vocabkey.EventType, firstEnd bool) Link`. Callers convert at the call: `vocabkey.BridgeTerm(term)`, `vocabkey.EventType(eventType)` | `segmentcompare`'s API is typed on `vocabkey` (`Link`, `Key`, `ReadFrom`); the string comes from the catalog at the adapter, which is where the conversion belongs | Keep `string` parameters — a typed package taking untyped strings, and `core/kinship` would hold typed terms |
| D-3 | `func EventEndpoint(b connectrules.Bridge) (first, ok bool)`, body unchanged, beside `LinkOf` | It says where a caller reads the event type that `LinkOf` decides whether to keep; the two halves of one rule stay in one file, as the map's scope entry has it | A method on `connectrules.Bridge` — workable (it reads only `TypeKey`), but it splits the event-type rule across two packages; review M-2 |
| D-4 | Bodies move unchanged; `LinkOf` keeps today's precondition that `b` has two endpoints, written in its doc comment. No length check is added | No behavior change; every caller already skips a bridge with `len(b.Endpoints) != 2`, and every product bridge has two | A guard returning a zero `Link` — a new behavior with no caller that needs it |
| D-5 | `core/segmentcompare` imports `core/connectrules`. No `layerPackages` row is added for `connectrules` | `connectrules` is pure (vocabkey only) and the guard's walker already fails if it ever reaches `core/database` through `segmentcompare`; the guard's rows are layer and pure value packages | A row for `connectrules` — it is neither, and the transitive walk already covers it |
| D-6 | Tests are new tables in `core/segmentcompare/segmentcompare_test.go`, against the four product bridges from `connectrules.LookupBridge` only | The product bridges are what every caller passes, and between them they take every branch: location's first endpoint is an event, participation's second, relationship and place relationship have none | Hand-built bridges — they would test shapes no caller passes; a new `key_test.go` — the package keeps one test file |
| D-8 | `func Keys(b connectrules.Bridge, term vocabkey.BridgeTerm, eventType vocabkey.EventType, terms Terms) (first, second vocabkey.BridgeKey)` *(new)*: `Key(LinkOf(b, term, eventType, true), terms)` and the same with `false`. Both loaders call it in place of their `Key(linkOf(…, true))` / `Key(linkOf(…, false))` pair. `LinkOf` stays exported: `Link` and `Key` are exported and a one-end reader can use it | `graphalign.Bridge` (`core/graphalign/types.go`) needs `KeyA` and `KeyB`, so S11-39's implied links need both keys too; one function keeps a third caller from repeating the pair or hand-swapping one end (review M-1) | Keep the pair at each call site — three copies once S11-39 lands |
| D-7 | No `promotealign` test changes. The existing suite is the regression check | No key changes; a changed expected value means behavior changed (Stop condition) | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `segmentcompare.LinkOf` *(new)* | func | `promotealign.linkOf` | It is that function, moved and exported; the old one is deleted |
| `segmentcompare.EventEndpoint` *(new)* | func | `promotealign.eventEndpoint` | Same; moved and deleted |
| `segmentcompare.Keys` *(new)* | func | the `keyA`/`keyB` pair in `loadLayerBridges` and `keyFrom`/`keyTo` in `canonKeys` | Replaces both copies of the pair |
| `TestLinkOf`, `TestEventEndpoint`, `TestKeys` *(new)* | test | `TestKey`, `TestReadFrom` | Same file and shape; no existing test covers the builder |

**Builds on.** `segmentcompare.Link` and `Key` (S11-14); the `ReadFrom` precedent (`10dec362`); `connectrules.LookupBridge`.
**Deletes.** `linkOf` and `eventEndpoint` in `core/database/promotealign/terms.go`; the two-call key pair in `layer.go` and `canon.go`.
**Next adds.** S11-38's `core/kinship` keys an implied link from both ends as `segmentcompare.Keys(relationshipBridge, term, "", terms)` (its term constants are S11-38's), the same reading the loaders use, ready for `graphalign.Bridge`'s `KeyA` / `KeyB` in S11-39. S11-40 needs nothing further from this PR.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/segmentcompare` | `key.go` imports `core/connectrules`. `func LinkOf(b connectrules.Bridge, term vocabkey.BridgeTerm, eventType vocabkey.EventType, firstEnd bool) Link` *(new)* and `func EventEndpoint(b connectrules.Bridge) (first, ok bool)` *(new)*, bodies from `terms.go`; `func Keys(b connectrules.Bridge, term vocabkey.BridgeTerm, eventType vocabkey.EventType, terms Terms) (first, second vocabkey.BridgeKey)` *(new)* per D-8 |
| `core/database/promotealign` | `terms.go` loses `linkOf` and `eventEndpoint`. `layer.go` `loadLayerBridges` and `canon.go` `canonKeys` call `segmentcompare.EventEndpoint(b)` and `keyA, keyB := segmentcompare.Keys(b, vocabkey.BridgeTerm(term), vocabkey.EventType(eventType), terms)` (`keyFrom, keyTo = …(l.TermKey)…` in `canonKeys`) |

Doc comment for `LinkOf` (adapted from today's):

```go
// LinkOf is a bridge as recorded, read from its first endpoint's end or its
// second. It is the one builder of a link: every reader of a bridge goes
// through it, or through Keys, so their keys cannot drift. b has two
// endpoints, and only a participation carries an event type.

// Keys are a bridge's keys read from its first endpoint's end and from its
// second.
```

Errors: none.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Add `LinkOf`, `EventEndpoint` and `Keys` to `segmentcompare`, with tests**
Files: `core/segmentcompare/key.go`, `core/segmentcompare/segmentcompare_test.go`. Skill: `add-match-rule` (layer 5).
Do: add the three functions per D-1 – D-4 and D-8 and the doc comments in §6, with `EventEndpoint`'s comment from `terms.go` ("the one rule for where a link reads its event type"). Add T-1, T-2 and T-5: the test file imports `core/connectrules`; a `bridge(t, key)` helper inside each test calls `connectrules.LookupBridge` and `t.Fatalf`s when it returns false. Terms in the rows: `vocabkey.BridgeTerm(vocabkey.RoleSubject)` (as `partKey` in the same file does), `vocabkey.BridgeTerm("parent")` (no constant exists until S11-38), `vocabkey.TermPartOf`; event type `vocabkey.EventBirth`.
Check: `go test ./core/segmentcompare/... ./core/matchregistry/...` passes with no build tags.

**STEP-2 — Move the loaders onto them and delete the private copies**
Files: `core/database/promotealign/terms.go`, `core/database/promotealign/layer.go`, `core/database/promotealign/canon.go`.
Do: replace each `eventEndpoint(` with `segmentcompare.EventEndpoint(`. Replace each two-line `segmentcompare.Key(linkOf(…, true), terms)` / `segmentcompare.Key(linkOf(…, false), terms)` pair with one `segmentcompare.Keys(…)` call per §6, converting `term` / `l.TermKey` and `eventType` per D-2. Change both "linkOf alone decides whether the key keeps it" comments to "`segmentcompare.Keys` alone decides whether the key keeps it". Delete `linkOf` and `eventEndpoint` from `terms.go`; drop any import that becomes unused.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes; `rg -n 'linkOf|eventEndpoint' core/database` prints nothing (AC-2.2).

**STEP-3 — Name the reading in the matching skill**
Files: `.agents/skills/add-match-rule/SKILL.md`.
Do: in the "Layer 5 (`core/segmentcompare`) scores one link." bullet, after "Keys come from `Key` with the catalog's `Terms`;", add: "outside tests, a bridge reads as its keys through `Keys` (or as a `Link` through `LinkOf`), with its event end from `EventEndpoint`;".
Check: `rg -n 'EventEndpoint' .agents/skills/add-match-rule/SKILL.md` prints the line.

**STEP-4 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: map row S11-41 in `deployment-plan.md`: Status `Landed #N` (the Plan cell already links this file). `completed.md`: a new section at the top (newest first), `## S11-41 — One link reading for bridges`, then `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).` and two or three sentences on what moved, ending "Nothing changes for researchers." This plan: Status `Landed in [#N](https://github.com/mendahu/provenencia/pull/N) (YYYY-MM-DD)`, and a `**Landed:**` header line after **Planned** listing every difference from the plan ("None." when there is none).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-41-link-reading.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits. For this PR: one code commit ("Read every bridge's keys through one function in segment compare.") and one docs commit.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/segmentcompare/segmentcompare_test.go` `TestLinkOf` *(new)* | Table, bridges from `connectrules.LookupBridge`: participation, `subject`, `birth`, first end → AC-1.1's link; second end → AC-1.2's; relationship, `parent`, `birth`, first end → AC-1.3's (no event type, `SameType`); relationship second end → `End: related_to, Other: person`; place relationship, `part_of`, `""`, first end → `End: from, Other: to, SameType: true`; location, `""`, `birth`, first end → AC-1.6's link. Each row compares the whole `Link` | AC-1.1, AC-1.2, AC-1.3, AC-1.6, D-2, D-4 |
| T-2 | Go | `core/segmentcompare/segmentcompare_test.go` `TestEventEndpoint` *(new)* | participation → `(false, true)`; relationship → `(false, false)`; place relationship → `(false, false)`; location → `(true, true)` | AC-1.4, D-6 |
| T-5 | Go | `core/segmentcompare/segmentcompare_test.go` `TestKeys` *(new)* | relationship, `parent`, `""`, `segmentcomparetest.Terms()` → AC-1.7's pair; participation, `subject`, `birth` → AC-1.8's pair; every row also equals `Key(LinkOf(…, true))`, `Key(LinkOf(…, false))` | AC-1.7, AC-1.8, D-8 |
| T-3 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` (existing) | `core/segmentcompare/pure` and `/numbers` pass with the new import | AC-1.5, D-5 |
| T-4 | Go | `core/database/promotealign` suite (existing, unchanged) | `TestLayerBridgesCoverEveryConnectRule`, `TestProposeInverseKinshipTermsCorrespond`, `TestProposeRelationshipFromEitherEnd`, `TestProposeChildOfReadsAsParentOf`, `TestProposeBirthOnAnotherPersonSubtracts` and `evidence_test.go` pass unchanged | AC-2.1, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promotealign` (a wrong type conversion at a call site) and `TestLayerPackages` (the new import).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none changes; add-match-rule layer 5 bullet updated (STEP-3)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Run `rg -n 'linkOf|eventEndpoint' core/` — only `segmentcompare.LinkOf` / `segmentcompare.EventEndpoint` and their call sites match (AC-2.2; the slice Check: `promotealign` has no private link reading left).
2. Open any project with a promoted Source that has a marriage and a birth with parents, open its Promote page: the rows, bands and evidence sheet read as on `main`. (No behavior change; the Go suite is the real check.)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| An import cycle through `connectrules` | Probed at the planning base: none. Stop condition below if one appears |
| A call-site conversion swaps term and event type (both string-backed) | D-2's types differ (`BridgeTerm`, `EventType`), so a swap fails to compile; T-4 catches anything else |

**Stop and report if:**

- `core/segmentcompare` importing `core/connectrules` gives an import cycle, or `TestLayerPackages` fails on it.
- Any `promotealign` or `graphalign` test needs a changed expected value to pass (that is a behavior change, which D-7 rules out).
- A caller of `linkOf` or `eventEndpoint` exists outside `layer.go` and `canon.go` at your base.
- `connectrules.Bridge` no longer has `Endpoints []Endpoint` with `PropertyKey` and `TypeKey`.

## 12. Open questions

None.

## 13. PR

**Title:** S11-41 — One link reading for bridges
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-41-link-reading.md`.
