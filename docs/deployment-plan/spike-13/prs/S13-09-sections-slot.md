# S13-09 — Sections slot on the Conclusion page

**Status:** Landed in #624 (2026-10-29)
**Spike / slice:** Spike 13 · Slice 1 — Page frame and Vitals
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: preparatory layout slot only; Person, Event and Place look the same)
**Board:** —
**Depends on:** —
**Blocks:** S13-10
**Branch:** `s13-09-sections-slot`
**Version bump:** none — preparatory Swift layout reshape with no change a researcher sees ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-1; working rules: "One sectioned Conclusion page, not a Person fork", "Each PR updates the docs and skills it makes false" (S13-09 → [`macos-client-patterns.md`](../../../macos-client-patterns.md)); [`../deployment-plan.md`](../deployment-plan.md) scope S13-09
**Planned:** 2026-10-28 against `main` @ `504f4f9b`
**Review:** [`S13-09-sections-slot.review.md`](S13-09-sections-slot.review.md) — Fits / Ready, 2 rounds; every blocker Fixed

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

`ConclusionDetailPage` places the shared header, then fills everything below it from a `pageSections` slot. Person, Event and Place keep today's call sites and receive today's Details rows, likeness gallery and Place relationship rows through a convenience initializer that supplies the default fill, so the three pages look the same and S13-10 can pass a custom Person fill without forking a Person-only page.

**Why now.** Working rule "One sectioned Conclusion page, not a Person fork"; Slice 1 preparatory PR; map dependency of S13-10.

**What stays.** The header (thumb, badge, title, summary, ref, members); `ConclusionDetailBody` fields (`rows`, `likenesses`, Place `sections`); the Details section header copy and refreshing spinner; likeness tiles and Place Part of / Contains / Succession; one `conclusionDetail` query key; Event and Place layout this spike (R-1); the three `*DetailView` labeled arguments (no new `pageSections:` argument this PR; only the explicit `<Content, Summary>` specialization is dropped for inference).

## 2. User stories

**US-1.** As a contributor, when I configure a Conclusion detail page, below-header content comes from a `pageSections` slot defaulted like `portrait`, so S13-10 can replace Person's fill with five sections without a second page type and without touching Event or Place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `ConclusionDetailPage` type parameters | Three generics: `Content: ConclusionDetailBody`, `Summary: View`, `PageSections: View`; full initializer takes `pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections` *(new)* |
| AC-1.2 | `ConclusionDetailLoaded` content branch | Draws `ConclusionDetailHeader`, then `pageSections(content, context)` — no inline Details / likeness / relationship stack left in `ConclusionDetailLoaded` |
| AC-1.3 | Convenience initializer | `extension ConclusionDetailPage where PageSections == ConclusionDetailDefaultPageSections<Content>` supplies `pageSections: { ConclusionDetailDefaultPageSections(content: $0, context: $1) }`; the three detail views call `ConclusionDetailPage(` without a type specialization and without a `pageSections:` argument |

**US-2.** As a researcher, when I open a Person, Event or Place page after this PR, the page looks and behaves as it did before the slot existed.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Open a Person with reconciled rows and likenesses | Header, Details (rows + refreshing aside), likeness gallery and portrait controls match pre-PR layout |
| AC-2.2 | Open an Event | Header, Details with Date and Place rows; no likeness section when empty |
| AC-2.3 | Open a Place with Part of / Contains / Succession | Header, Details (Names, Period), then the three relationship sections in the same order |
| AC-2.4 | `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` | All existing tests pass |

**US-3.** As a contributor, when I read `macos-client-patterns.md` and the detail page source, the sectioned detail page and the Place hierarchy `sections` field are named separately, so the next Person section add does not collide with Place relationship data.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `docs/macos-client-patterns.md` under Conclusion keys | A bullet documents the `pageSections` slot, the default fill, and that `ConclusionDetailBody.sections` / `ConclusionDetailSection` stay Place hierarchy rows inside the default page section |
| AC-3.2 | `ConclusionDetailBody` doc comment in `ConclusionDetailPage.swift` | States that kinds fill the body fields, the page places the header and the kind-supplied `pageSections` slot, and a kind does not get its own page type |
| AC-3.3 | Map scope S13-09 Ships line | Says `ConclusionDetailPage` gains a sections slot (not `ConclusionDetailBody`) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionDetailPageSectionsContext` *(new)* with `projectDir`, `refreshing`, `portrait` (plain struct, not `Equatable`) | STEP-1, AC-1.1, D-1 |
| `ConclusionDetailPage` / `ConclusionDetailLoaded` third generic `PageSections` and `pageSections` parameter; page `body` passes `pageSections` into `ConclusionDetailLoaded`; content branch calls the slot after the header | STEP-1, AC-1.1, AC-1.2, D-2 |
| Internal `ConclusionDetailDefaultPageSections` *(new)* in `ConclusionDetailPage.swift`: today's Details header + rows, likeness section, Place relationship `ForEach`; takes `content` + `ConclusionDetailPageSectionsContext` | STEP-1, AC-1.3, AC-2.*, D-3, D-4 |
| Convenience `init` where `PageSections == ConclusionDetailDefaultPageSections<Content>`; three detail views drop `<Content, Summary>` only | STEP-1, AC-1.3, D-5 |
| Rewrite `ConclusionDetailBody` doc comment (no Person-page fork; page places header + slot) | STEP-1, AC-3.2, D-7 |
| `docs/macos-client-patterns.md` sectioned-detail-page bullet; map Ships line drops `(and ConclusionDetailBody)` | STEP-2, AC-3.1, AC-3.3, D-6 |
| Map row status, `completed.md`, this plan Landed note (at ship) | STEP-2 |

| Out | Home |
| --- | --- |
| Visible layout change; section switcher; five named Person sections; section in `WorkspaceLocation` | S13-10 (brief S13-D1) |
| Custom `pageSections:` at any `*DetailView` call site | S13-10 (Person); Event/Place keep the convenience default until they adopt sections |
| Moving likenesses into a Gallery page section | S13-14 |
| Kind-neutral gallery slot / Swift gallery rename on the body | S14-02 |
| Renaming `ConclusionDetailBody.sections` or `ConclusionDetailSection` (Place hierarchy) | not planned this PR; named in the patterns doc (AC-3.1) so later work can choose |
| Provenance sheet; fixed header redesign beyond today's header | S13-11; S13-10 |
| New `DesignSystem/` component or recipe | not planned — snowflake stays in `Features/Conclusions/` (`add-ui-component`) |
| New `.swift` file or `project.pbxproj` entry | not planned — new types stay in `ConclusionDetailPage.swift` (D-4) |
| Engine, FFI, proto, L10n keys | not planned — no copy or wire change |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Page | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage` | Generics `Content`, `Summary`; fields `session`, `entityId`, `pageIdentifier`, `errorIdentifier`, `make`, `summary`, optional `portrait` (`= nil`); warms `.conclusionDetail`; `body` builds `ConclusionDetailLoaded` with `make`, `summary`, `portrait` only |
| Loaded body | same file `ConclusionDetailLoaded` lines 191–218 | On `.content`: `ConclusionDetailHeader`, then a `VStack` with `PVSectionHeader` (`L10n.Conclusions.detailsSection` + refreshing `ProgressView`), `ForEach(content.rows)`, optional `ConclusionDetailLikenessSection`, `ForEach(content.sections)` → `ConclusionDetailRelationshipSection` |
| Body protocol | same file `ConclusionDetailBody` lines 88–91 | Doc comment: "The page only places them. Giving a kind its own page later does not need a new query or history entry." Fields: `title`, `ref`, `showsRef`, `members`, `rows`, `sections` (`[ConclusionDetailSection]`, default `[]`), `thumbnail`, `badge`, `likenesses` |
| Place hierarchy type | same file `ConclusionDetailSection` | Place Part of / Contains / Succession data (id, title, aside, emptyText, relationships, omittedCount) — **not** a Person page section |
| Portrait default | same file `ConclusionDetailPage.portrait` | `var portrait: ConclusionPortraitModel? = nil`; only `PersonDetailView` passes it |
| Person call site | `macos/App/Features/Conclusions/PersonDetailView.swift` | `ConclusionDetailPage<PersonDetailContent, PersonDetailVitals>(…, portrait:)` — no page-sections parameter |
| Event call site | `macos/App/Features/Conclusions/EventDetailView.swift` | `ConclusionDetailPage<EventDetailContent, EventDetailSummary>(…)` — omits `portrait` |
| Place call site | `macos/App/Features/Conclusions/PlaceDetailView.swift` | `ConclusionDetailPage<PlaceDetailContent, PlaceDetailChain>(…)`; content fills `sections` with three relationship groups |
| Portrait model | `macos/App/Features/Conclusions/ConclusionPortraitModel.swift` `ConclusionPortraitModel` | `@MainActor @Observable final class` — not `Equatable` |
| Patterns doc | `docs/macos-client-patterns.md` Conclusion keys | Documents one detail key and shared rows; no sectioned-page / `pageSections` bullet |
| Map Ships | `docs/deployment-plan/spike-13/deployment-plan.md` S13-09 scope | Already names `pageSections` on `ConclusionDetailPage` and the convenience default fill (updated while planning); STEP-2 verifies it still matches AC-3.3 and updates status/`completed.md` at ship |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift`, `EventDetailTests.swift`, `PlaceDetailTests.swift` | Content, presentation, portrait, Place relationships — no view-tree assert of the loaded stack |
| Precedent | `ConclusionDetailPage.portrait` default; `summary: (Content) -> Summary`; `ConclusionDetailThumbnail` "one slot" comment (S10-23) | Copy the optional-`portrait` defaulting pattern for the page-sections fill; keep extracted chrome private in the same file |

**Precedent.** Optional `portrait` is the call-site pattern when every kind shares a default and only a later PR customizes one kind. `summary` stays a required factory because each kind's summary view differs. The thumbnail comment ("One slot, so a later PR adds to this struct") is the naming precedent for a preparatory slot. Do not copy `ConclusionListKind`'s static protocol reshape — detail content is instance data on `ConclusionDetailBody`, not static kind metadata.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Add `struct ConclusionDetailPageSectionsContext` *(new)* with `projectDir: String`, `refreshing: Bool`, `portrait: ConclusionPortraitModel?`. No `Equatable`. `ConclusionDetailLoaded` builds it when calling `pageSections` | Call sites and the default fill need projectDir / refreshing / portrait in one value; `ConclusionPortraitModel` is a class and cannot synthesize `Equatable` | `: Equatable` (will not compile); four loose args; capturing only from the outer view |
| D-2 | `ConclusionDetailPage` gains generic `PageSections: View` and stored `pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections`. `ConclusionDetailLoaded` gains the same stored closure (passed from the page `body` initializer call) and, on `.content`, draws header then `pageSections(content, context)` only | Working rule: one page, sections slot; S13-10 replaces Person's fill | Hardcoding Person sections inside the page; a Person-only `PersonDetailPage` fork; putting page sections on `ConclusionDetailBody` as another `[…]` array (collides with Place `sections`) |
| D-3 | Extract today's Details + likeness + relationship stack into `struct ConclusionDetailDefaultPageSections<Content: ConclusionDetailBody>: View` *(new, internal)* with `content: Content` and `context: ConclusionDetailPageSectionsContext`. Move the likeness and relationship call sites into it unchanged (`ConclusionDetailLikenessSection` and `ConclusionDetailRelationshipSection` stay `private` in the same file). Internal access matches kit defaults such as `PVSelectPlainRow` so the convenience `where` constraint is visible to `*DetailView.swift` | One shared default fill; no visible change; a `private` default type cannot appear in a `where PageSections == …` constraint used from other files (review M-4); S13-10 reuses it by name for Event/Place or Person Vitals | Copy-pasting the stack into three call sites; inventing five empty Person section shells this PR; `private` default fill (blocks the convenience init at the three call sites) |
| D-4 | New types stay in `ConclusionDetailPage.swift`. No new file, no `project.pbxproj` change | File already owns the page and private section views; pbxproj still lists this file by hand | A new `ConclusionDetailDefaultPageSections.swift` (pbxproj churn for a private extract) |
| D-5 | Add `extension ConclusionDetailPage where PageSections == ConclusionDetailDefaultPageSections<Content>` with a convenience `init(session:entityId:pageIdentifier:errorIdentifier:make:summary:portrait:)` that matches today's call-site arguments and supplies `pageSections: { ConclusionDetailDefaultPageSections(content: $0, context: $1) }`. At the three call sites, drop the explicit `<Content, Summary>` specialization so inference can pick the constrained init (kit pattern: `PVCallout(` / `PVSelect(` with no specialization of the default). Do not add a `pageSections:` argument. If inference fails, write the three type arguments including `ConclusionDetailDefaultPageSections<…>` by name; do not introduce a factory, typealias, or `AnyView` beside `summary` | Internal default type + unconstrained call site matches `PVSelect`/`PVSelectPlainRow`; Event/Place omit customization; working rule "pass a single section" is the default fill | Three identical `pageSections:` closures; keeping `ConclusionDetailPage<A, B>(` two-arg specializations that cannot name the third generic; a parallel factory beside the convenience init |
| D-6 | Update `docs/macos-client-patterns.md` with one Conclusion-keys bullet, and strike `(and ConclusionDetailBody)` from the map S13-09 Ships line so it matches D-2 | Working rule names this doc for S13-09; map must not teach a Body slot | Leaving the collision undocumented; renaming Place `sections` in this PR |
| D-7 | Rewrite the `ConclusionDetailBody` doc comment: kinds fill the body fields; the page places the header and the kind-supplied `pageSections` slot; a kind does not get its own page type | Removes the "own page later" fork the working rule forbids | Leaving the old comment for S13-10 to misread |
| D-8 | No new L10n keys; no VERSION bump | Docs + layout reshape only; researcher-visible copy unchanged | Patch bump for an internal slot |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionDetailPageSectionsContext` *(new)* | type | Fields already threaded into `ConclusionDetailLoaded` / likeness section | Bundle for the slot factory; nothing else is a context value today |
| `pageSections` on `ConclusionDetailPage` *(new parameter)* | API | `portrait` optional default; `summary` required factory | Required on the full init; defaulted via convenience init like `portrait` for call sites |
| `ConclusionDetailDefaultPageSections` *(new, internal)* | view | Inline stack in `ConclusionDetailLoaded` lines 198–217; kit precedent `PVSelectPlainRow` | Extract, not a second layout beside the inline one; must be internal for the convenience `where` constraint |
| Convenience `init` *(new)* | API | Today's memberwise / synthesized call shape with `portrait = nil` | Constrains `PageSections` to the default fill so call sites stay unchanged |

**Builds on.** `ConclusionDetailPage` + `ConclusionDetailBody` shared page; optional `portrait` defaulting; working rule "One sectioned Conclusion page, not a Person fork".
**Deletes.** The inline Details / likeness / relationship stack inside `ConclusionDetailLoaded`'s `.content` branch (moved into `ConclusionDetailDefaultPageSections`). No callee loses its last caller: `ConclusionDetailLikenessSection` and `ConclusionDetailRelationshipSection` keep one caller (the default page sections view). The "own page later" clause in the `ConclusionDetailBody` doc comment (replaced by D-7).
**Next adds.** S13-10: Person uses the full init with a custom `pageSections` (switcher + five sections); Event and Place keep the convenience init and may name `ConclusionDetailDefaultPageSections` for Vitals-equivalent content. S13-14: Gallery page section replaces likenesses inside the Person fill. S14-02: gallery wording moves onto the body; edits the default fill / likeness section in the same file.

### Reach

None — no new navigation path, no new empty/missing-referent state; same three detail places.

### Scale

None — no write, list or loop changes; same one `conclusionDetail` read per page.

### Schema

None — no catalog change.

### Go core

None — Swift layout only.

### FFI

None — no proto or handler change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Page sections slot | `ConclusionDetailPage.pageSections` *(new)* | `ConclusionDetailPage.swift` | Full-init factory; convenience init supplies the default |
| Default one section | `ConclusionDetailDefaultPageSections` *(new, internal)* | same file | Today's Details + likenesses + Place relationships; takes `content` + context |
| Header | `ConclusionDetailHeader` | same file | Unchanged |
| Call sites | `PersonDetailView`, `EventDetailView`, `PlaceDetailView` | `Features/Conclusions/` | Same labeled arguments; drop `<Content, Summary>` for inference (D-5) |

States: loading / failed / content unchanged. Accessibility: existing identifiers on the page and likeness tiles stay; no new ids.

Sketch (signatures only):

```swift
struct ConclusionDetailPageSectionsContext {
    var projectDir: String
    var refreshing: Bool
    var portrait: ConclusionPortraitModel?
}

struct ConclusionDetailPage<Content: ConclusionDetailBody, Summary: View, PageSections: View>: View {
    // …existing fields…
    let pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections
}

extension ConclusionDetailPage where PageSections == ConclusionDetailDefaultPageSections<Content> {
    init(
        session: WorkspaceSession,
        entityId: String,
        pageIdentifier: String,
        errorIdentifier: String,
        make: @escaping (CatalogConclusionDetail) -> Content,
        summary: @escaping (Content) -> Summary,
        portrait: ConclusionPortraitModel? = nil
    ) { /* supplies pageSections → ConclusionDetailDefaultPageSections */ }
}

struct ConclusionDetailDefaultPageSections<Content: ConclusionDetailBody>: View {
    let content: Content
    let context: ConclusionDetailPageSectionsContext
    // Details header + rows; likenesses; Place relationship sections
}
```

### Strings

None — no L10n add or remove; Details / likenesses / Place section copy unchanged.

## 7. Steps

**STEP-1 — Sections slot, default fill, convenience init (one compile unit)**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`. Skill: `add-ui-component` (compose/extend the existing snowflake page; no new kit primitive).
Do: Add `ConclusionDetailPageSectionsContext` (D-1). Add generic `PageSections` and `pageSections` to `ConclusionDetailPage` and `ConclusionDetailLoaded`; in the page `body`, pass `pageSections` into `ConclusionDetailLoaded`'s initializer (D-2). Move the Details `VStack` (section header + rows), the likeness `if`, and the relationship `ForEach` into internal `struct ConclusionDetailDefaultPageSections` taking `content` + `context` (D-3, D-4). On `.content`, keep `ConclusionDetailHeader`, then call `pageSections(content, .init(projectDir:projectDir, refreshing:refreshing, portrait:portrait))` — no leftover inline stack. Add the convenience `init` (D-5). At each of `PersonDetailView`, `EventDetailView` and `PlaceDetailView`, drop the explicit `<Content, Summary>` type arguments so inference selects the convenience init; keep the same labeled arguments (`session`, `entityId`, identifiers, `make`, `summary`, and Person's `portrait`). Rewrite the `ConclusionDetailBody` doc comment (D-7).
Check: the three detail views compile with no `pageSections:` label; `rg "ConclusionDetailPage<" macos/App/Features/Conclusions/*DetailView.swift` has no hits (inference, not a two-arg specialization); `rg "ForEach\\(content\\.rows\\)" macos/App/Features/Conclusions/ConclusionDetailPage.swift` hits only inside `ConclusionDetailDefaultPageSections`; `ConclusionDetailLoaded` `.content` has no rows/likenesses/relationship loop.

**STEP-2 — Docs**
Files: `docs/macos-client-patterns.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map Ships line + map row status at ship), `docs/deployment-plan/spike-13/completed.md`, this plan (status + Landed note at ship).
Do: Add the AC-3.1 bullet under Conclusion keys (D-6). Verify the S13-09 Ships line still matches AC-3.3 (already updated while planning). At ship: map Status `Landed #N`; `completed.md` one paragraph; plan Status `Landed in #N (date)` with Landed note of any deviation.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-09-sections-slot.md` clean after status update; Ships line still has no `ConclusionDetailBody` slot claim.

Commits: one per behavior. Suggested subjects:
1. "Give Conclusion detail pages a below-header sections slot."
2. "Document the sectioned Conclusion detail page in client patterns."

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` existing suite | Unchanged cases still pass (presentation, rows, likenesses, portrait) | AC-2.4, AC-2.1 |
| T-2 | Swift | `macos/ProvenenciaTests/EventDetailTests.swift` existing suite | Unchanged cases still pass | AC-2.4, AC-2.2 |
| T-3 | Swift | `macos/ProvenenciaTests/PlaceDetailTests.swift` existing suite | Relationship sections still built on content (`periodAndRelationshipsFillFromTheHeader` and empty counterpart) | AC-2.4, AC-2.3 |
| T-4 | Manual | Manual check §10 steps 1–4 | Person / Event / Place visual parity | AC-2.1, AC-2.2, AC-2.3 |
| T-5 | Structural | STEP-1 `rg` / typecheck | `ConclusionDetailPage` has three generics and `pageSections`; `ConclusionDetailLoaded` `.content` calls the slot after the header with no inline rows/likeness/relationship loop; convenience init exists; `ConclusionDetailDefaultPageSections` is internal (not `private`); `*DetailView.swift` has no `pageSections:` argument and no `ConclusionDetailPage<` two-arg specialization | AC-1.1, AC-1.2, AC-1.3, D-2, D-3, D-5 |
| T-6 | Docs | `docs/macos-client-patterns.md`, map Ships line, `ConclusionDetailBody` comment + Manual check §10 steps 5–6 | Patterns bullet names `pageSections` and Place hierarchy `sections`; Ships line names only `ConclusionDetailPage`; protocol comment matches AC-3.2 | AC-3.1, AC-3.2, AC-3.3, D-6, D-7 |

No new Swift test file. View-tree inspection of the slot is not in the repo's Swift Testing practice (`add-swift-test`: no view snapshots unless asked). Slot wiring is T-5; researcher parity is T-1–T-4.

Must stay green: everything `scripts/check-pr.sh` runs. Suites most likely to break: `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` if the convenience init signature drifts from today's call sites. Xcode legs of `check-pr.sh` run on macOS or CI (`add-swift-test`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes (T-4 recorded in the PR description)
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md sectioned detail page bullet; map S13-09 Ships line
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with at least one Person that has a name, life rows and a likeness. Open the Person page. See the same header, Details block (rows + member/refresh chrome as before), and likeness gallery with portrait controls. (AC-2.1)
2. Open an Event. See the event header and Details (Date, Place) only — no empty likeness block. (AC-2.2)
3. Open a Place that has Part of / Contains / Succession data. See Names and Period under Details, then the three relationship sections in the same order and wording as before. (AC-2.3)
4. Flip back to the Persons / Events / Places lists and reopen the same entities: same pages, no new section switcher.
5. Confirm `docs/macos-client-patterns.md` Conclusion keys include the sectioned-detail-page bullet from AC-3.1.
6. Confirm the map S13-09 Ships line names only `ConclusionDetailPage` for the slot (AC-3.3) and the `ConclusionDetailBody` comment matches AC-3.2.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Name collision with Place `ConclusionDetailBody.sections` | D-2 / D-6: slot is `pageSections`; Place data type keeps its name; patterns doc spells the difference |
| S14-02 edits the same likeness stack | Next adds names S14-02; if S14-02 lands first, re-read Current state and move its gallery slot into `ConclusionDetailDefaultPageSections` |
| Convenience-init generic constraint fails to type-check at call sites | D-5: drop two-arg specializations first; if still failing, name all three type arguments including `ConclusionDetailDefaultPageSections`; stop rather than add a second defaulting mechanism |
| SwiftUI cannot infer `PageSections` from the convenience init | Same as above; do not add a factory, typealias, or `AnyView` |

**Stop and report if:**

- `ConclusionDetailLoaded` no longer owns the content branch described in §4 (another PR split the page).
- Adding `PageSections` forces a Person-only page type or a change to `WorkspaceLocation` (Out → S13-10).
- S14-02 (or another PR) has already replaced the likeness stack so STEP-1 cannot move the lines §4 names — refresh Current state first.
- Existing detail tests fail for a reason other than a convenience-init signature mismatch.
- The convenience init cannot be expressed so that today's three call sites compile unchanged.

## 12. Open questions

None.

## 13. PR

**Title:** S13-09 — Sections slot on the Conclusion page
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-09-sections-slot.md`.
