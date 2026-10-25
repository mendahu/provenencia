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

## Refresh — 2026-10-22

The ship loop stopped at STEP-4: binding writes were unaudited, so D-7's effects entry could not fire and failed `TestCascadeIntoNonEmptyEffectIsAudited` ([ship log](S11-32-lifespan-likeness-lines.ship.md)). **User call (2026-10-22):** route binding writes through `writes.Run` with an audited `subject_type_properties` change and the deleteimpact work, and add a binding reorder RPC for AC-3.3. The plan gained D-19 – D-22, STEP-5, AC-3.5 – AC-3.7, T-13, T-14, and the reorder gesture went Out to `docs/ideas/binding-order-editing.md`.

**Writer's own calls.** Entity id of a binding change is the property id (as `source_metadata_layout` uses the field id). The reorder takes the full permutation only. No `GenealogyStore` method until a gesture calls the RPC. `EnsureBinding` stays un-audited for `Install`. The reorder response carries `WriteEffects`.

### Refresh round 1

Lint clean before the reviewers ran. Maintainability (fresh, Opus): **Fits with changes**, one blocker. Executability (fresh, Sonnet): **Ready with changes**, one blocker.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-20 | blocker | `conclusiondetails` keeps its own binding-order SQL and sort beside `Vocabulary.Bindings` | **Fixed.** D-7: `Vocabulary.BindingOrder()` and `vocabkey.CompareByOrder`; `loadProperties` moves onto them and `sqlPropertyIDs` drops its join; D-8 sorts with `CompareByOrder`; STEP-4 does it |
| M-21 | major | `catalog-deletes.md`, `core/writes/README.md`, `properties.Delete` comment become false | **Fixed.** STEP-10 and DoD docs; D-19 updates the comment (merged with X-21) |
| M-22 | major | Third inline `engine.SubjectTypeProperty` build | **Fixed.** D-22: `subjectTypePropertyProto`; L141 and L295 move onto it |
| M-23 | major | Twin `source_type_metadata_fields` stays silent and un-audited | **To user → answered (2026-10-22):** source type ↔ metadata field links are UI hints and need no audit. Out line, not planned |
| M-24 | minor | Binding fields spelled three times | **Fixed.** D-19: `bindingFields` with `FullRow` / `DeletedRow`, as `evidencegraphs.linkFields` |
| X-20 | blocker | Test callers of `AppendBinding` / `DeleteBinding` not listed or directed | **Fixed.** D-19 lists every caller and the `writes.Call` shape; STEP-5 Files and an `rg` check |
| X-21 | major | `docs/catalog-deletes.md` becomes false | **Fixed** with M-21; the `add-catalog-delete` skill row too |
| X-22 | major | Create's `sort_order` and delete's old values undecided | **Fixed.** D-19: `SELECT` after a non-empty insert; read before delete; unbound pair → no change |
| X-23 | major | Empty list, missing type, response effects undecided | **Fixed.** D-22: both `ErrInvalid`; response gains `WriteEffects effects = 2` |
| X-24 | minor | Seed row stale against the branch | **Fixed.** Row marked landed in `948864c` |
| X-25 | minor | `TestVocabularyBindings` file and assertions unnamed | **Fixed.** `graphcache/vocab_test.go` *(new)*; T-3 asserts `BindingOrder()[person]` |
| X-26 | minor | Proto placement, revision helper, assign-of-bound case | **Fixed.** §6 proto note and dispatch names; T-14 names `assertLatestAuditAction`; AC-3.6 covers assign of a bound pair |

### Refresh round 2

Maintainability (fresh, Opus): **Fits**, no new blockers. Executability (fresh, Sonnet): **Ready with changes**, no new blockers.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-30 | minor | `GetByIDTx` takes `*sql.Tx` | **Fixed.** D-19 passes `tx.Tx` |
| (MNT note) | — | STEP-5's `rg` guard misses unqualified in-package calls | **Fixed.** Guard drops the package prefix |
| (MNT note) | — | `ReorderBindings`' read of bound rows unnamed | **Fixed.** D-22: private `sqlBindingOrders` on the tx |
| (MNT note) | — | `ListVocabulary` comment and `interpretation-layer-data-model.md` L693 go stale | **Fixed.** STEP-5 and STEP-10 |
| (MNT note) | — | Ranking by key, not id | **Accepted.** Risks row: order only; evidence already keys by `PropertyKey` |
| (MNT note) | — | Reorder's update change spells fields a third time | **Logged** (minor, round 1 half) |

**Verdict:** Fits — Ready with changes, all applied. No open blocker; To user answered.
