# Ship log — S14-01 Pictures engine

**Plan:** [`S14-01-pictures-engine.md`](S14-01-pictures-engine.md) · **Branch:** `s14-01-pictures-engine` · **PR:** #615
**State:** ready

## Round 0 — implement

Head `d4c82b3e`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report: deviation — `placePictureSubject` instead of D-12 `placeSubject` (name taken by `placechains_test.go`). S12-24 preferred-delete asserts preserved under `gallery_test.go`.

## Round 1 — review

Reviewers: CON, MNT, UX, SEC+PRF (Cursor Auto / inherit). Skipped: none (macos/App and api/ffi + core/database match UX and SEC+PRF).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Identical kind-case tables in both header lead tests | Fix now → F-1 | verified: both `cases` bodies byte-identical; plan Next adds wants one kind row |
| MNT-2 | MNT | minor | Three near-identical lead map helpers | Fix now → F-1 | same file; same-pass with F-1 |
| MNT-3 | MNT | minor | `eventSubject` / `placePictureSubject` near-copy | Fix now → F-1 | same file; same-pass with F-1 |
| MNT-4 | MNT | minor | Install smoke place-key list omits `place_picture` | Fix now → F-2 | verified: place loop is toponym/start_date/end_date only; event len bumped to 6 |
| UX-1 | UX | major | Properties VO rows with shared Picture label identical | Later PR → S14-02 | Seeds intentional (D-2); `PropertiesView` not in this diff; Out is app chrome → S14-02. Wrote into S14-02 scope |
| UX-2 | UX | minor | Promote still says likeness for Event/Place pictures | Later PR → S14-02 | Plan Out already; no change |
| SEC/PRF | SEC+PRF | — | — | — | Clean; no findings |

Nits (logged): `CON-1 · spike-14/README.md still says no PR has landed` · `CON-2 · TestForEntityLikenesses names` · `CON-3 · TestGetConclusionDetailListsLikenesses name` · `MNT-5 · placePictureSubject naming` · `MNT-6 · three RelPaths loops (D-11)` · `UX-3 · GenealogyStore setPreferredLikeness comment`

### Fix plan 1

**F-1 — One shared header-picture kind matrix** (from MNT-1, MNT-2, MNT-3)
Files: `core/database/conclusionheaders/conclusionheaders_test.go`. Skill: —.
Failing test first: none (refactor). After extract, `TestHeaderLeadsWithEarliestPicture` and `TestHeaderLeadsWithChosenPicture` must still pass for person/event/place.
Do: Extract one shared `headerPictureCases()` (or package-level cases) both tests range over. Fold `personLeads` / `eventLeads` / `placeLeads` into one helper. Fold `eventSubject` / `placePictureSubject` into one `subjectAndHandle(kind)` (keep a one-line wrapper only if needed). Grep check: no second copy of the person/event/place case table body in that file.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/ -run 'TestHeaderLeadsWith'` passes; `scripts/check-pr.sh` PASS.

**F-2 — Install smoke asserts `place_picture`** (from MNT-4)
Files: `core/database/subjectvocab/subjectvocab_test.go`. Skill: —.
Failing test first: none required; extend the place key loop.
Do: Add `"place_picture"` to the place Install smoke key list (alongside toponym/start_date/end_date). Optionally assert `event_picture` among event bindings, not only `len == 6`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/ -run TestInstall` (or the Install smoke test name) passes; `scripts/check-pr.sh` PASS.

**F-3 — Landing README + renamed gallery test names** (from CON-1, CON-2, CON-3)
Files: `docs/deployment-plan/spike-14/README.md`, `core/database/conclusiondetails/conclusiondetails_test.go`, `api/ffi/handlers/conclusion_detail_test.go`. Skill: —.
Do: Update README Status so it no longer says no PR has landed (S14-01 and S14-08 landed). Rename `TestForEntityLikenesses` / `TestForEntityLikenessesSkipNegatives` / `TestGetConclusionDetailListsLikenesses` to Pictures wording to match the package rename.
Check: `rg 'TestForEntityLikenesses|TestGetConclusionDetailListsLikenesses|no PR has landed' docs/deployment-plan/spike-14 core/database/conclusiondetails api/ffi/handlers` empty for those strings; `scripts/check-pr.sh` PASS.

**Later PR notes (written into map):** S14-02 scope adds: Properties list VoiceOver must distinguish two Properties that share label Picture (include key and/or bound kind labels in `rowAccessibilityLabel`); Promote depiction wording (already in scope).

## Round 1 — fix

Head `510de8e7`. Gate PASS (orchestrator). F-1 done; F-2 done; F-3 done. MNT-1–4 Resolved (round-2 MNT). CON nits Resolved via F-3.

## Round 2 — review

Reviewers: MNT only (had Fix-now major). Skipped: CON / UX / SEC+PRF (no Fix-now blocker or major in round 1).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| — | MNT | — | — | — | Clean; F-1–F-2 hold; no new blockers in `d4c82b3e...510de8e7` |

## Result

Gate PASS · CI pending · open: none · follow-ups: UX-1 (S14-02 Properties VO), UX-2 (S14-02 Promote wording) · to user: §10 manual check · CON miss: none
