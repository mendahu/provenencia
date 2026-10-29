# Review — S14-02 Pictures in the app

**Verdict:** (pending maintainability) — Ready with changes → fixes applied
**Rounds:** 1 (executability returned; maintainability still open). **Reviewers:** maintainability (cursor-grok-4.6-xhigh, pending), executability (composer-2.5)

## To user

None yet.

## Writer's own calls

- Gallery slot as `ConclusionDetailGallery` carrying copy (not a kind enum in the section).
- Keep type name `ConclusionPortraitModel`; inject `ChooseCopy`.
- Promote picture sentences for every kind (value type), including Person.
- Person header a11y ids stay `persons.detail.portrait*`; Event/Place use `.lead*`.
- Properties VO includes both machine key and bound type labels.
- Out: `ReconciledValueDisplay.likenessValue` and composer catalog-comment “person” wording.

## Round 1

### X-1 · major · STEP-5 grep cannot mean “empty” as written
**Plan:** STEP-5 Check — `rg '…|ValueFamily' macos` empty.
**Codebase:** `L10n.swift` private `enum ValueFamily` stays after D-9 (`.picture` case).
**Better shape:** Drop bare `ValueFamily` from the pattern.
**Answer:** Fixed — STEP-5 greps `ValueFamily\.likeness` (and other retired tokens only). Deletes clarify the enum stays.

### X-2 · major · D-3 header menu ids vs today’s stable `portrait.*` ids
**Plan:** D-3 used `\(identifierRoot).lead.*` for every kind when choose is set.
**Codebase:** `persons.detail.portrait.menu` / `.useEarliest` / `.portrait` (S12-17).
**Better shape:** Person keeps portrait ids or rename is in Deletes/sweep.
**Answer:** Fixed — D-3: Person keeps `persons.detail.portrait*`; Event/Place use `\(identifierRoot).lead*`. Deletes/Strings updated.

### X-3 · minor · Tile choose-control accessibility suffixes not named
**Plan:** “choose controls under the same root” without leaf names.
**Codebase:** `…likeness.{id}.portraitBadge`, `…usePortrait`.
**Answer:** Fixed — D-3 names Person `.portraitBadge` / `.usePortrait` and Event/Place `.thumbnailBadge` / `.useThumbnail`.

### X-4 · minor · STEP-1 check passes while the tree does not compile
**Plan:** STEP-1 replaces protocol member; Check was `rg` only; STEP-2 updates conformers.
**Answer:** Fixed — STEP-1 requires a green App build with STEP-2 (same commit or stub); STEP-1+2 are one green gate before STEP-3.

### X-5 · minor · §8 Tests omits spoken-label updates implied by D-9
**Plan:** T-6 covered sentence keys only.
**Codebase:** `PromoteEvidenceDisplayTests` spokenLikeness cases.
**Answer:** Fixed — T-6b covers `spokenPicture*` and renamed spoken tests.

## Round 2

Pending maintainability round-1 result.
