# Ship log — S11-32 Lifespan and likeness evidence lines

**Plan:** [`S11-32-lifespan-likeness-lines.md`](S11-32-lifespan-likeness-lines.md) · **Branch:** `s11-32-lifespan-likeness-lines` · **PR:** #523
**State:** ready

## Round 0 — implement
Base `3c2ff0d`.
Head `948864c` (local, not pushed). Gate not run. Report: **Stopped on** "Changing the `subject_type_properties` effects entry fails an `effects` test for a reason other than the entry itself".

- STEP-1 – STEP-3 done (`bd81258`, `391dd1f`, `eeff78e`); STEP-3 also carries D-13's loader plumbing, needed to compile after D-5.
- STEP-4 partial (`948864c`): D-6 seed order and `graphcache.Vocabulary.Bindings` landed; the `effects` registry change was reverted.
- Evidence (verified by orchestrator): setting `"subject_type_properties"` to `{Lists: []List{ListVocabulary}}` fails `TestCascadeIntoNonEmptyEffectIsAudited` (`subject_type_id` and `property_id` cascade into a table with effects and are not Audited). Fixing that needs `Audited: true` on both FKs in `catalogmodel/model.go:178-179`, which pulls in the deleteimpact audit/release path: a catalog-delete change the plan does not scope.
- Premise gap (verified): `subjectvocab.EnsureBinding` / `AppendBinding` / `DeleteBinding` (`subjectvocab.go:102,119,196`) are raw `db.Exec` outside `writes.Call`, so no effect would fire anyway. The only non-Install caller is `api/ffi/handlers/subject_defs.go:173,203`. No binding reorder RPC exists (`engine.proto` has only `ReorderSourceMetadata`), so AC-3.3 and Manual check step 4 have no code path.
- STEP-5 – STEP-9 not started.

Back to `plan-pr` for D-7 and AC-3.3.

Plan refreshed with `plan-pr` (user, 2026-10-22): D-7 rewritten, D-19 – D-22 and STEP-5 added; review-plan refresh rounds 1–2, verdict Fits / Ready with changes, no open blocker. Commit `0506850`.

## Round 0 (resumed) — implement
Base `3c2ff0d`. Resumed from STEP-4 at `0506850`.
Head `4133efa`. PR #523 (draft). Gate PASS (run by orchestrator). Report: 7 commits after the resume; deviations — AC-2.5's 1838 burial overlaps the row's span, so the tests use 9 Jan 1810 to get `conflict`; D-20's "no facet for `subject_type_id`" failed `TestFacetReleaseHonesty` (`subject_types` counts as deletable), so a `KindSubjectType` facet was added; `datedEvent` lost its unused `polarity` parameter; `BindingOrder` has a value receiver, as `TermsByID`; T-10's lifespan dates are a unit test on `claimEvidenceProto`; STEP-6 passed an empty `Inputs{}` until STEP-7. Manual check not walked.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (`macos/App/**` and `Localizable.xcstrings` route UX; `api/ffi/**`, the matching layers, `core/database/**` and `catalogmodel` route SEC+PRF).

All four verdicts Clean; no blocker or major.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor | `BindingOrder` maps types by key with no origin rule; conclusion page walks `v.Types` to get back from id to key | Fix now → F-1 | `vocab.go:29-66`, `conclusiondetails.go:264-270`; `subject_types.key` not unique (`000021.sql`); code this PR adds |
| MNT-2 | MNT | minor | "First lifespan lead in key order" chosen twice | Fix now → F-2 | `subjectprofile.go:146-150` and `claimevidence/build.go:51-65`; code this PR adds |
| MNT-3 | MNT | minor | `writes.Call` + `AppendBinding` wrapper copied five times; action string repeated | Fix now → F-3 | same-pass cleanup: four test packages and `subject_defs.go` |
| MNT-4 | MNT | minor | `conclusionValueProto` gains a parameter used by one caller | Rejected | plan D-16, chosen in plan review M-3 so one encoder owns `ConclusionValue`; the reviewer marks it deliberate |
| UX-1 | UX | minor | Today's sheet words every likeness line and pin alike, also to VoiceOver | Later PR → S11-20 | AC-5.3 added to the S11-20 plan (labels name the member's Source and outcome); crops and copy are S11-20's (D-18, Out) |
| UX-2 | UX | nit → minor | `promote.sheet.unknown` translator comment does not cover `unjudged` | Fix now → F-4 | stale comment in `L10n.swift`, a file this PR changes |
| SEC note | SEC | — | Binding audit entity id is the property id alone | Rejected | plan D-19 decision (as `source_metadata_layout` uses `field_id`); the change's fields carry `subject_type_id` |

Nits (logged): `CON-1 · promotealign_test.go:869 · TestProposeIgnoresLikenesses covers only the unjudged pair` · `CON-2 · values.go:100 · missing RelPaths entry leaves an empty path` · `MNT-5 · layer.go:109-116 · forSubject closure writes primary while ranging` · `MNT-6 · subjectvocab.go:233-237 · ReorderBindings diff built beside bindingFields` · `MNT-7 · subjectvocab_test.go:566,576 · appendBinding/deleteBinding signatures differ` · `MNT-8 · evidence.go:194-201 · bindingOrder(g) forwarder`

### Fix plan 1

**F-1 — Binding order by type id, provenencia types by kind** (from MNT-1)
Files: `core/database/graphcache/vocab.go`, `core/database/graphcache/vocab_test.go`, `core/database/conclusiondetails/conclusiondetails.go`.
Failing test first: `TestVocabularyBindings/a plugin type with a seeded key` — insert a `subject_types` row with key `person` and origin `plugin:x`, bind one property to it → `BindingOrder()[person]` lists only the provenencia person's properties.
Do: add `(v Vocabulary) TypeOrder(typeID []byte) []vocabkey.PropertyKey` *(new)*: that type's bound property keys by `SortOrder`, then key. `BindingOrder()` keeps only `OriginProvenencia` types (as `promotealign/canon.go` `typeByKey`) and builds each list with `TypeOrder`. `loadProperties` calls `v.TypeOrder(subjectTypeID)` and drops its `v.Types` walk.
Check: conclusion tests unchanged; scripts/check-pr.sh PASS.

**F-2 — One rule for the lifespan lead** (from MNT-2)
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectprofile/subjectprofile.go`, `core/claimevidence/build.go`, a `subjectcompare` test.
Failing test first: `TestFirstLifespanLead` — `{"b": …, "a": …}` → `"a", true`; empty → `"", false`.
Do: `subjectcompare.FirstLifespanLead(leads map[vocabkey.PropertyKey]Lifespan) (vocabkey.PropertyKey, bool)` *(new)* — the smallest key. `addLifespans` and `lifespanOf` both call it; neither sorts or mins keys itself.
Check: `rg -n 'maps.Keys\((lifespans|leads)\)' core` prints nothing; T-2, T-5 pass; scripts/check-pr.sh PASS.

**F-3 — One binding test helper and the action names** (from MNT-3)
Files: `core/database/subjectvocab/subjectvocab.go`, `core/database/subjectvocab/subjectvocabtest/subjectvocabtest.go` *(new)*, `api/ffi/handlers/subject_defs.go`, `core/database/deleteimpact/impact_test.go`, `core/database/subjects/delete_test.go`, `core/database/properties/properties_delete_test.go`, `core/database/subjectvocab/subjectvocab_test.go`.
Do: export `subjectvocab.ActionAssign = "assign_subject_type_property"`, `ActionRemove = "remove_subject_type_property"`, `ActionReorder = "reorder_subject_type_properties"`; the three handlers use them. `subjectvocabtest.Assign(t testing.TB, c *database.Catalog, userID, typeID, propertyID []byte)` *(new)* wraps `writes.Call` with `ActionAssign` and fails the test on error. Move every test site onto it; `subjectvocab_test.go`'s in-package `appendBinding` / `deleteBinding` helpers take the same `(t, c, userID, typeID, propertyID)` shape (MNT-7 rides along). If an import cycle blocks the in-package test, keep its private helpers but use the exported action constants.
Check: `rg -n '"(assign|remove)_subject_type_propert' core api` prints only `subjectvocab.go`; `rg -n 'writes\.Call\(.*\n?.*AppendBinding' core` finds only `subjectvocabtest`; scripts/check-pr.sh PASS.

**F-4 — Translator comment** (from UX-2)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: the `promote.sheet.unknown` comment covers both meanings: nothing to compare, or two likenesses the researcher judges by eye.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `d931438`. Gate PASS (run by orchestrator). F-1 – F-4 done (`b3776f7`, `696cd6b`, `c79667f`, `21820fe`). Orchestrator read the fix commits: MNT-1 **Resolved** (`TypeOrder`; `BindingOrder` provenencia only; `loadProperties` has no `v.Types` walk), MNT-2 **Resolved** (`FirstLifespanLead` in both sites), MNT-3 **Resolved** (`subjectvocabtest.Assign`, `Action*` constants; the remaining literals are `unmarshalErr` RPC names and the test's independent `assertLatestAuditAction` checks), UX-2 **Resolved**. Implementer also gave S11-20's AC-5.3 (added by orchestrator in `d4fc2ce`) a T-5 row and manual step 8, which `check-plan-docs.py` required (`d931438`).

No round 2: round 1 had no Fix-now blocker or major.

## Result
Gate PASS · CI pending · open: none · follow-ups: none · later PR: UX-1 → S11-20 AC-5.3 · to user: — · CON miss: none
