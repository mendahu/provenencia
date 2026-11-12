# Ship log — S15-26 The metadata board takes an owner

**Plan:** [`S15-26-metadata-board-takes-an-owner.md`](S15-26-metadata-board-takes-an-owner.md) · **Branch:** `s15-26-metadata-board-takes-an-owner` · **PR:** #798
**State:** ready

## Round 0 — implement
Base `main` @ `8d349308e`. Head `6b3ddb5bf`. Gate PASS (orchestrator re-run). Report deviations: version bump 0.10.2 → 0.10.3 (main moved); `api/ffi/handlers/sources_test.go` `unmarshalSetMetadataEntry` read the removed `resp.Entry`, now reads the valued row from `resp.Metadata` (plan's `rg GetEntry api` missed it). Manual check not run (needs the app).

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`), SEC+PRF (`api/ffi/**`; `macos/App/Features/**` models). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | A board write patches the page's copy of `sourceWorkspace`, then reloads the whole key the write already described | To user → Later PR (S15-09, required) | User (2026-11-12): required in S15-09, not "only if slow"; written into the map's S15-09 scope entry ("Also (from S15-26 review, MNT-1…)"). Also verified the reload resets a row edit begun before it lands (`SourcePageModel.apply` compares against `appliedWorkspace`). Contradicts the plan's Out line (review M-2: patching deferred to S15-09, only if its measurement shows the reload is slow). Verified: `MetadataBoardHost.swift` `entries` setter + `didWrite` → `.mutatedSourceWorkspace` |
| MNT-2 | MNT | major | `FakeStore` keeps the engine's Clear rule and Set placement removed from the client | To user → Fix now → F-1 | User (2026-11-12): script results now; supersedes D-4's "keeps the scripted result". Contradicts D-4 (keeps both as scripted results); matches `swift-tests.mdc` line 23 ("does not re-implement… engine rules"). Verified `FakeStore.swift:775-796` |
| MNT-3 | MNT | minor | `metadataBySource` view: fake Source delete/copy go through the `.source`-only view and will miss Artifact owners | To user → Fix now → F-2 | User (2026-11-12): move sites now; supersedes D-4's computed view. Contradicts D-4 (computed view so seeds compile unchanged). Verified `FakeStore.swift:148-158`, `:550`, `:725` |
| MNT-4 | MNT | minor | `save` and `add` hand-route a set failure the same way | Fix now → F-3 | verified `MetadataBoardModel.swift:113-119`, `:150-155` |

SEC+PRF: clean. UX: clean. CON: clean.

Nits (logged): `UX-1 · MetadataBoardDialogs.swift:42 · cancelAdd() clears addTarget, combo options empty during sheet close; check in manual check` · `MNT-5 · MetadataBoardModel.swift:18-19 · host is read by no one outside; make private, drop comment` · `MNT-6 · MetadataBoardHost.swift:44-45 · rows/setRows pair could be one WritableKeyPath` · `CON-1 · FakeStore.swift:779 · clearMetadata keeps openHref on an emptied suggested row` · `CON-2 · MetadataBoardModel.swift · rename detected only at -M30%`


### Fix plan 1

**F-1 — FakeStore scripts metadata write results instead of applying engine rules** (from MNT-2; supersedes D-4's "keeps the scripted result"; CON-1 goes with it)
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift` (and any other test that relied on the fake's Set/Clear/Dismiss result). Skill: `add-swift-test`; rule `.cursor/rules/swift-tests.mdc`.
Failing test first: none new (the rule is pinned in Go by `TestMetadataOwners`); the Source-page tests that pin AC-1.2 and AC-1.6 must fail once the rule is removed from the fake, then pass after they seed the post-write rows.
Do: add one scripted map, `metadataWriteResults: [MetadataOwner: [CatalogMetadataEntry]]`, shaped like `nextWriteEffects`. `setMetadata`, `clearMetadata` and `dismissMetadataSuggestion` (all three apply an engine rule today: placement, empty-or-remove, remove-unfilled) each: throw their error hook as now; if `metadataWriteResults[owner]` is set, store it as `metadataByOwner[owner]` and return it; otherwise return the stored list unchanged. Keep `bumpActivity` for a `.source` owner. Leave `reorderMetadata` as it is unless it also derives rows by a rule. Tests that saw a row change after a write seed `metadataWriteResults` with the rows the engine would return.
Check: `rg -n "suggested \{|list.remove\(at:|list.append\(CatalogMetadataEntry|removeAll \{ \$0.field.id" macos/App/Platform/FakeStore.swift` prints no line inside the three metadata write methods; scripts/check-pr.sh PASS.

**F-2 — Delete the `metadataBySource` view; every site uses `metadataByOwner`** (from MNT-3; supersedes D-4's computed view)
Files: `macos/App/Platform/FakeStore.swift` (the view at ~148-158 and its internal uses: workspace read ~492, Copy Source ~550, Source delete ~725, `sourcesHoldingField` ~2938), every test under `macos/ProvenenciaTests` that names `metadataBySource`.
Do: replace each use with `metadataByOwner[MetadataOwner(level: .source, id: <id>)]`, then delete the view. Source delete and Copy Source keep today's behavior (the Source's own owner only); do not add Artifact-owner rules to the fake (S15-09 seeds those and its scope entry carries the note).
Check: `rg -n metadataBySource macos` prints nothing; scripts/check-pr.sh PASS.

**F-3 — One helper classifies a set failure for `save` and `add`** (from MNT-4)
Files: `macos/App/Features/Sources/MetadataBoardModel.swift`.
Do: one private helper that takes the thrown error and returns nil (owner gone: `host.failureMessage(for:)` is nil), `.invalid(message)` for `sourceMetadataInvalid`, or shows the failure through `host.showFailure` and returns a shown result. `save` maps `.invalid` to `fieldError`/`fieldErrorID`; `add` maps it to the dialog's value error. Behavior unchanged; existing T-8 (`invalidValueStaysOnTheFieldOrDialog`) and T-4 pin it.
Check: `rg -n "engineCode == .sourceMetadataInvalid" macos/App/Features/Sources/MetadataBoardModel.swift` prints one line; scripts/check-pr.sh PASS.

**Docs (rides along).** The plan's Landed note records: D-4 superseded by F-1 and F-2 (user, 2026-11-12); the Out row for patching the cached `sourceWorkspace` (review M-2) is now required in S15-09 per MNT-1 (user, 2026-11-12) — update that Out row's "Lands in" cell to say so.

## Round 1 — fix
Head `888215c81`. Gate PASS (orchestrator re-run). F-1, F-2, F-3, Docs done. Deviations: F-3's code landed in `f72739d8b` with F-1/F-2, and `888215c81` (subject names F-3) carries the plan docs — subjects and contents are crossed; left as is (pushed history). Two Source-page clear tests now assert the seeded result, so they pin little; the rule is pinned in Go by `TestMetadataOwners`.
F-3 (MNT-4, minor): **Resolved** — orchestrator read `MetadataBoardModel.swift`: one `classifySetFailure`; `save` and `add` map `.invalid` to their own place.

## Round 2 — review
Reviewers: MNT (had Fix-now major MNT-2). Not re-sent: CON, UX, SEC+PRF (no Fix-now blocker or major). Scope `6b3ddb5bf...888215c81`.

MNT-2: **Resolved** (round-2 MNT reviewer: `scriptedMetadataWrite` in `FakeStore.swift`; no Set/Clear/Dismiss rule left; tests seed post-write rows). MNT-3: Resolved (`rg -n metadataBySource macos` empty). No new blockers. Note: `reorderMetadata` in the fake still renumbers and appends, as F-1 left it.

## Result
Gate PASS at `888215c81` · `origin/main` unchanged since the branch was cut (`8d349308e`) · CI see PR #798 · open: none · follow-ups: none · later PR: MNT-1 → S15-09 (required) · to user: MNT-1, MNT-2, MNT-3 answered 2026-11-12.
