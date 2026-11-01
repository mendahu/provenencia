# Review — S13-04 Person family read

**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet).
**Verdict:** round 1 — Fits with changes / Not ready; every finding Fixed. Round 2 below.

## Decisions the user made while planning (2026-11-01)

- **Parents pair into sets by evidence.** Two parents are one set when they are partners (spouse link, recorded or marriage-implied) or both stand as the person's parent in one birth event; a parent paired with nobody stands alone (D-4, D-5). "One set unless 3+" was offered and declined.
- **A co-parent gets a unit.** A child's other parent who is not a spouse is a partner with no link to the person (D-6). Putting such children in the one-parent group was offered and declined.

Both are in the map's working rule "Family groups in Go".

## Round 1 — maintainability

### M-1 · blocker · A second batched parent-step walk beside `siblingsOf`
**Answer:** Fixed — D-8: one `deriver.loadSteps` (four batches, per-id node, parent edges, step people) under both `siblingsOf` and `LoadRelationships`; `forLinks` uses `stepPeople`. STEP-1 moves the existing walk first with no behavior change.

### M-2 · blocker · The handler enforces a grouping rule again
**Answer:** Fixed — D-10: `PersonFamily` drops relatives with no summary and units left empty; D-11: the handler only maps; T-3 gains the row.

### M-3 · blocker · `Kin` field and `KinOf` branch per role, with the next role already named
**Answer:** Fixed — D-2: a declared `nuclearRoles` table like `eventPatterns`; `Kin` is `map[Role][]Link`; a grandparent row is one entry.

### M-4 · major · A third copy of family-fixture builders
**Answer:** Fixed — D-14: builders on `graphcachetest.Canon`; `canonlinks`' `world` delegates. The raw merge updates in other packages' tests are not this PR's files and stay.

### M-5 · major · Name the shown-person predicate once
**Answer:** Fixed — D-7: `shownPerson`, documented as the package's one canonical-relative rule for S13-05; why `nodesOf` checks only half is stated.

### M-6 · major · Pattern mapping falls back with no guard
**Answer:** Fixed — D-11 and T-5 `TestKinshipPatternProto`.

### M-7 · minor · `Load…` means two things
**Answer:** Fixed — `LoadRelationships` derives and stores `For`'s answer (D-8).

### M-8 · minor · Relationship-key lookup copied again
**Answer:** Fixed — `relationshipKeys`, used by `ParentStep`, `chains`, `Siblings` and `KinOf` (D-2, STEP-2).

### M-9 · minor · Page-only grouping in a matching-layer package
**Answer:** Fixed — kept in `core/kinship` (pure, needs `relationshipKeys`, which is unexported); the Out row states why K-12 still holds, and STEP-8 adds that sentence to K-12.

## Round 1 — executability

### X-1 · blocker · The canonical filter needs relative nodes nothing loads
**Answer:** Fixed — D-9: a second `LoadRelationships` over every relative of all four roles before the filter, which reads only `g.Cached`.

### X-2 · blocker · `ThrowingStore` in `OnboardingModelTests.swift` not named
**Answer:** Fixed — STEP-6 Files.

### X-3 · major · Checks not runnable here (protoc, xcodebuild)
**Answer:** Fixed — a Stop condition: the proto and Swift steps run on macOS; never hand-edit generated files.

### X-4 · major · Choices left open
**Answer:** Fixed — `stepPeople` signature (D-8); missing ids skipped; duplicate ids give one `Family` (D-9); `kinOf` keyed by `string(id)`; `association_ids` holds whatever `Link.Associations` holds; an empty unit is one with no partner and no child (D-10); `.personFamily` placed after `.personClaims` in every switch (D-12); three parents in one birth pair all three ways (D-4, T-1).

### X-5 · minor · The numbers guard on `core/kinship`
**Answer:** Fixed — Rules that bind.

### X-6 · minor · Terms with no `vocabkey` constant
**Answer:** Fixed — T-1 says `vocabkey.BridgeTerm("cousin")`.

### X-7 · minor · Test package and commit ids
**Answer:** Fixed — T-3 names `persondetail_test`; AC-2.5 commits the association id too.
