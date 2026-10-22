# Review — S11-32 Lifespan and likeness evidence lines

**Round 1, 2026-10-22.** Lint clean before the reviewers ran. Maintainability (fresh context): **Fits with changes**. Executability dry run (fresh context): **Ready with changes**, one blocker. No round 2: the one blocker of each lens was a missing decision, now made, and no answer reshaped the plan; the user answered the one product call.

**User call (2026-10-22):** sections follow the catalog's binding order (the order the conclusion page uses), and the seeded person binding puts Likeness second, after Name. New catalogs only.

## Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | `Rules.Order` is a second per-kind property order beside `subject_type_properties.sort_order` | **To user → answered:** binding order; seed likeness second. D-6 (seed), D-7 (bindings on the cached vocabulary, dropped on a binding edit), D-8 (`Inputs.Order` passed to `Sections`). `Rules.Order` and its inventory row removed. Map scope entry updated |
| M-2 | major | `Value.Image []byte` breaks the pointer-per-type shape | **Fixed.** D-1: `Value.Depiction *depictionvalue.Value`; D-3, D-4 share the pointer |
| M-3 | major | Paths resolved in the handler through a `conclusiondetails` field its owner never sets | **Fixed.** D-15: `EvidenceValue.ImageRelPath` filled by `EvidenceValues.ResolvePaths`; D-16: `conclusionValueProto` takes a depiction input |
| M-4 | major | `LifeSpan` beside `Lifespan`; `Gap` beside the `gap` value cause | **Fixed.** `LifeSide` and `LifespanMissing` in Go, proto (`missing`) and Swift |
| M-5 | major | `LifeOf` re-runs the lifespan pass | **Fixed.** D-5: `Build` and `ForSubject` return the side with the values; `LifeOf` removed |
| M-6 | minor | `Lifespan` repeats `Line`'s verdict fields | **Fixed in part.** D-11: one `verdictOf` helper fills both. A shared struct is declined (Out line): it churns every `Line` literal for four fields |
| M-7 | minor | "Likeness" value copy outlives its screen | **Fixed.** Out line: S11-20 retires `likenessValue` |

## Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | An empty relation has no Swift value | **Fixed.** D-18: `ValueRelation.none` for `""`, and `""` in the vocabulary row (T-11) |
| X-2 | major | The profile map is shared with `matchIndex.profiles` | **Fixed.** D-13 names `index.go`, `export_test.go` and the readers; STEP-6 files |
| X-3 | major | Test metas lack `likeness` and `interests`; no depiction builder | **Fixed.** STEP-5: `subjectcomparetest.Metas()` gains both as seeded; STEP-1: `propertycomparetest.Depiction` |
| X-4 | major | Year-only date fixtures | **Fixed.** D-14: one `datedOn` helper; `datedEvent` and `datedBirth` call it |
| X-5 | major | No catalog in `claimEvidenceProto` | **Fixed.** D-15/D-16: paths resolved in `promotealign` inside `withProjectCatalog` before encoding |
| X-6 | major | No handler fixture for likeness and lifespan | **Fixed.** T-10 is likeness paths only on `newEvidenceFixture` with `SaveLikeness`; the lifespan's dates are T-7's |
| X-7 | minor | `ok` dropped; "one lead" undefined | **Fixed.** D-5: has a lifespan exactly when `Missing` is empty; the first lead in key order |
| X-8 | minor | Test mechanics for the lifespan table and `Build` | **Fixed.** STEP-3 and T-5 say how each side is built |
| X-9 | minor | Refused-image precedent not cited | **Fixed.** §4 row and T-6 |
| X-10 | minor | Files with no Do; encoder test missing; case order | **Fixed.** STEP-5 says what `TestEvidenceSection` asserts; `invariants_test.go` dropped; T-9; D-16 puts the depiction first |
| X-11 | minor | Swift small choices | **Fixed.** D-18: `mapDepictionValue`, conformances, the L10n helper |
| X-12 | nit | Proto and Xcode tools missing on Linux | Logged: §6 FFI says a macOS host runs `generate-proto.sh`; a stop condition covers it |
