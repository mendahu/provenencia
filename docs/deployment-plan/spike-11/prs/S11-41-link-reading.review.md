# Review — S11-41 One link reading for bridges

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (Opus, fresh agent), executability (Sonnet, fresh agent). No blocker and no reshape, so no round 2.

## To user
- None.

## Writer's own calls
- D-2: `LinkOf` and `Keys` take `vocabkey.BridgeTerm` / `vocabkey.EventType`, not `string`; the adapter converts at the call.
- D-5: no `layerPackages` row for `core/connectrules`; the guard's transitive walk covers it.
- D-6: tests use only the four product bridges (location covers `EventEndpoint`'s first-endpoint branch).
- D-8 (from M-1): a new `segmentcompare.Keys` returning both ends' keys; `LinkOf` stays exported.
- D-3 (M-2): `EventEndpoint` stays in `segmentcompare` beside `LinkOf`, as the map's scope entry has it, rather than becoming a `connectrules.Bridge` method.

## Round 1

### M-1 · major · Three callers will repeat the "key from both ends" pair
**Plan:** STEP-2 kept `Key(LinkOf(…, true))` / `Key(LinkOf(…, false))` at each call site; Next adds showed S11-38 keying one end only.
**Codebase:** `promotealign/layer.go` `loadLayerBridges` and `canon.go` `canonKeys` each hold the pair; `core/graphalign/types.go` `Bridge` needs `KeyA, KeyB`, so S11-39's implied links need both.
**Better shape:** one `segmentcompare.Keys(b, term, eventType, terms) (first, second)`.
**Answer:** Fixed — D-8, §3 In row, §6 Codebase fit and Go core, STEP-1/STEP-2, AC-1.7, AC-1.8, T-5; Next adds now shows `Keys`.

### M-2 · minor · `EventEndpoint` is a bridge-spec fact, not link reading
**Plan:** D-1 / D-3 export it into `segmentcompare`.
**Codebase:** reads only `TypeKey`; `core/database/identityclaims/bridgefile.go` `sameEndpointType` repeats the same-type rule.
**Answer:** Declined with reason — D-3 now says why: it is the half of the event-type rule that `LinkOf` completes, and the map's S11-41 scope entry moves both to `segmentcompare`. `sameEndpointType` logged in §3 Out (not planned).

### M-3 · minor · `ReadFrom` takes unpacked endpoints beside a Bridge-taking reader
**Answer:** Logged — §3 Out. For S11-40's planner: S11-40 reworks `evidence.go`'s relationship reading and is the natural place to have `ReadFrom` take a `connectrules.Bridge`, if it still pays then.

### M-4 · minor · "never built by hand" is false for tests
**Answer:** Fixed — STEP-3 wording is now "outside tests, …"; §3 Out notes test literals stay.

### M-5 · nit · `LinkOf` doc comment names `core/kinship`, which does not exist yet
**Answer:** Fixed — §6 comment says "every reader of a bridge".

### X-1 · major · AC-2.2's `rg` used `\|`, which never matches
**Answer:** Fixed — AC-2.2 and STEP-2 use `rg -n 'linkOf|eventEndpoint' core/database`.

### X-2 · major · STEP-4 left map-row and `completed.md` text undefined
**Answer:** Fixed — STEP-4 gives the map Status, the `completed.md` heading and form, and the plan's Landed status and note. The map's Plan cell links the plan now.

### X-3 · major · Plan was Draft and unreviewed
**Answer:** Fixed — this log; status Planned.

### X-4 · minor · T-1 left term constants open
**Answer:** Fixed — STEP-1 names `vocabkey.BridgeTerm(vocabkey.RoleSubject)`, `vocabkey.BridgeTerm("parent")`, `vocabkey.TermPartOf`, `vocabkey.EventBirth`.

### X-5 · minor · Test bookkeeping
**Answer:** Fixed — STEP-1 names the `connectrules` import and a `bridge(t, key)` helper that fails on a missing bridge; AC-1.4 lists all four bridges.

### X-6 · minor · `check-pr.sh` needs `xcodebuild`
**Answer:** Declined — every Spike 11 plan carries this DoD line; `ship-pr` runs `scripts/check-pr.sh` itself on a machine with Xcode. An implementer on Linux reports the Go checks and `check-plan-docs.py`.

### X-7 · nit · Manual check 2 needs the macOS app
Logged.
