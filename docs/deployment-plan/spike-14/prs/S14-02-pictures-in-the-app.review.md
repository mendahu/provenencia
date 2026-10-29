# Review — S14-02 Pictures in the app

**Verdict:** Fits with changes — Ready with changes (round 1 Fixed; round 2 pending)
**Rounds:** 1. **Reviewers:** maintainability (cursor-grok-4.6-xhigh), executability (composer-2.5)

## To user

None.

## Writer's own calls

- Gallery slot as `ConclusionDetailGallery` carrying copy (not a kind enum in the section).
- `ChooseCopy` static packs carry copy + a11y id leaves; section/header stay kind-blind.
- Page owns `ConclusionPortraitModel` from `userID` + `leadChoose`.
- Keep type name `ConclusionPortraitModel`.
- Promote picture sentences for every kind (value type); drop `isPicture` bool; neighbor crop marks from `PromoteFlow.Row.kind`.
- Promote `worded` maps depiction → `pictureValue`; leave `ReconciledValueDisplay.likenessValue` for Person rows.
- Properties VO includes machine key and bound type labels.

## Round 1

### X-1 · major · STEP-5 grep cannot mean “empty” as written
**Answer:** Fixed — greps `ValueFamily\.likeness` only.

### X-2 · major · D-3 header menu ids vs today’s stable `portrait.*` ids
**Answer:** Fixed — then superseded by M-1: ids live on `ChooseCopy` packs; Person pack keeps `persons.detail.portrait`.

### X-3 · minor · Tile choose-control accessibility suffixes not named
**Answer:** Fixed — then superseded by M-1: `badgeSuffix` / `useAsSuffix` on `ChooseCopy`.

### X-4 · minor · STEP-1 check passes while the tree does not compile
**Answer:** Fixed — STEP-1+2 one commit / one green gate.

### X-5 · minor · §8 Tests omits spoken-label updates implied by D-9
**Answer:** Fixed — T-6b.

### M-1 · blocker · D-3 puts per-kind identifier leaves beside the gallery slot
**Plan:** Kind forks for badge/header ids in D-3 while rejecting a kind enum in the section.
**Codebase:** Section/header hard-code Person ids today; working rule one slot; S13-14 reuses the slot.
**Better shape:** Id leaves on `ChooseCopy`; static packs; section kind-blind.
**Answer:** Fixed — D-3 / D-2 / sketch / STEP-1: `ChooseCopy` carries `badgeSuffix`, `useAsSuffix`, `headerIdentifier`; `.personPortrait` / `.eventThumbnail` / `.placeThumbnail`.

### M-2 · blocker · D-10 stamps the sheet’s kind on every depiction line, including a neighbor’s
**Plan:** All picture lines used row `kind.markKey`.
**Codebase:** Relationship sections are built for `n.Kind`; Swift `neighbor` → `PromoteFlow.Row.kind`.
**Answer:** Fixed — D-10 / AC-3.5 / T-7b: own cards use row kind; relationship groups use neighbor kind with row-kind fallback.

### M-3 · major · `isPicture` stays a second flag beside `Value.crops`
**Answer:** Fixed — D-9 deletes `isLikeness` and does not add `isPicture`; branch on `.crops`; AC-5.1 / T-7 updated.

### M-4 · major · Out of `likenessValue` leaves Promote member wording on “Likeness”
**Answer:** Fixed — D-14 Promote `worded` → `pictureValue`; Out narrowed to not changing `ReconciledValueDisplay`; T-7c.

### M-5 · major · Event/Place copy PersonDetailView’s portrait machine instead of the page owning it
**Answer:** Fixed — D-7 / STEP-3: page takes `userID` + `leadChoose`, owns model and overlay; kind views stay thin.

### M-6 · minor · `ConclusionDetailGallery.empty` is a fake Person/Pictures hybrid
**Answer:** Fixed — D-1: required `gallery` like `thumbnail`; no `.empty` default with mixed chrome.

## Round 2

Pending — fresh maintainability + executability on the Fixed reshape (M-1, M-2, M-5 especially).
