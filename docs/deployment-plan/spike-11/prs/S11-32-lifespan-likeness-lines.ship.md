# Ship log — S11-32 Lifespan and likeness evidence lines

**Plan:** [`S11-32-lifespan-likeness-lines.md`](S11-32-lifespan-likeness-lines.md) · **Branch:** `s11-32-lifespan-likeness-lines` · **PR:** —
**State:** implementing (resumed)

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
