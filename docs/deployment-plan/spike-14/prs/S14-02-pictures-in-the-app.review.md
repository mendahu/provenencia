# Review — S14-02 Pictures in the app

**Verdict:** Refresh round 3 (2026-10-30): Fits with changes — Not ready (no board yet). Every finding answered below. Rounds 1–2 (2026-10-29) reviewed the earlier plan.
**Rounds:** 3 (one refresh round). **Reviewers:** round 3: maintainability (Opus), executability (Sonnet, the implementer's model); rounds 1–2: maintainability (cursor-grok-4.6-xhigh), executability (composer-2.5)

## To user

- **S14-02 now depends on S13-10** (M-5). The S14-D2 board puts the Event and Place pages on Page columns, which S13-10 builds, and the plan is written against S13-10's `PersonDetailSections`. The map row, its dependency graph, and Spike 13's S13-14 row (which now also waits on S14-02) are updated. Recommendation: accept; it removes a two-branch design.
- **A confirmation round still runs.** Round 3 had blockers (M-1, M-2, X-2, X-3), all Fixed. The confirming round runs with the post-board refresh, which `plan-pr` already requires before `Planned`, so the plan is reviewed once against the real frames instead of twice.

## Writer's own calls

- Copy with arguments is formatted once in the content through L10n functions; tiles carry ready Strings (M-1).
- One lead model per page, always built from `userID`; copy read at call time from `gallery.choose` (M-2).
- Ids from `pageIdentifier` with fixed leaves; the Person's `persons.detail.portrait*` ids become `persons.detail.lead*` (no test reads them) (M-3).
- Rename `ConclusionPortraitModel` → `ConclusionLeadModel`, with a hand pbxproj edit (M-9).
- `offersEarliest` moves to the gallery; one `ConclusionDetailThumbnail.leading(…)` builder (M-4).
- Delete the `thumbnailRelPath` default and the dead `L10n.Promote.ValueFamily`.
- Properties label on `PropertiesModel`, origin spoken as the cell shows it, no trailing comma for an unbound Property.
- New `pictureOpenTip` key rather than borrowing the Person-namespaced `likenessOpenTip`.

## Refresh round 3 (2026-10-30)

Scope change before review: S13-23 (#633) removed the Promote evidence sheet, so US-3 (Promote), D-9/D-10/D-14 (Promote wording, wire `kind`) and their tests left the plan; the user chose to gate the PR on a new board, S14-D2.

### M-1 · blocker · Format templates stored as data and filled outside L10n
**Answer:** Fixed — D-2, D-12, § Strings: `ConclusionDetailPicture` carries `accessibilityLabel`, `useAsLabel`, `chosenToastBody` built by `tiles(from:tileLabel:useAsLabel:chosenBody:)` from L10n functions; the slot and `LeadCopy` hold only static resources; Person functions unchanged.

### M-2 · blocker · `leadChoose` and `gallery.choose` carry the same pack
**Answer:** Fixed — D-6, D-7: `leadChoose` removed; the page always builds `ConclusionLeadModel(session:userID:entityId:)`; `choose(_:copy:)` / `useEarliest(copy:)` take `content.gallery.choose`; `choose` is non-optional.

### M-3 · major · Id and mark fields repeat what the page knows
**Answer:** Fixed — D-3: ids from `pageIdentifier` with fixed leaves; context and header carry `pageIdentifier`; tile fallback is `content.thumbnail.mark`; `identifierRoot`, suffixes, `headerIdentifier`, `fallbackMark` dropped.

### M-4 · major · Thumbnail logic repeated; `offersEarliest` a second copy
**Answer:** Fixed — D-5: `ConclusionDetailThumbnail.leading(…)`; `offersEarliest` computed on the gallery, deleted from the thumbnail.

### M-5 · major · Both landing orders planned; S13-10 overlap half described
**Answer:** Fixed — Depends on S13-10 (plan, brief, both maps); D-7 and D-13 written once against S13-10's full init, context and `PersonDetailSections`; §4 rows marked *S13-10*. To user (dependency).

### M-6 · major · Docs and comments left false
**Answer:** Fixed — STEP-6 and DoD: `macos-client-patterns.md` line 102, `event-place-pictures.md` lines 32/33/64, `CachedThumbnail.swift` and `ConclusionListRow.swift` comments, the model doc comment.

### M-7 · major · `thumbnailRelPath` default left with no user
**Answer:** Fixed — D-8 deletes it; Deletes lists it.

### M-8 · minor · Person-only empty fields on the shared slot
**Answer:** Fixed — D-13: `PersonDetailSections` reads `galleryEmptyTitle` / `galleryEmptyMessage` directly.

### M-9 · minor · Kind-neutral layer still says portrait
**Answer:** Fixed — D-6: `ConclusionLeadModel`, `context.lead`, `LeadCopy`; pbxproj lines named in §4.

### M-10 · minor · Raw origin in a new key; label away from the feature's spoken text
**Answer:** Fixed — D-11: `PropertiesModel.rowAccessibilityLabel(for:)`, `PropertiesOriginCell.spokenOrigin(_:)`.

### M-11 · nit · Shared page uses another feature's toast modifier
**Answer:** Logged — Out row; not this PR.

### X-1 · blocker · The board does not exist
**Answer:** Declined as a plan change — the gate already exists: Status is `Blocked — on board S14-D2 and on S13-10`, and `ship-pr` starts only from a `Planned` plan (`.claude/skills/ship-pr/SKILL.md` line 47). The Implementer block is copied verbatim from the template and is not edited.

### X-2 · blocker · "More tiles per row at Wide" cannot be true
**Answer:** Fixed — brief EPP-W1, § 4.1, frame 9; plan §6 Wide row and §10 step 4: the primary column goes from 1038pt to at most 1180pt, so Wide gains about one tile per row and is otherwise the Regular page, centered.

### X-3 · blocker · Page init change not specified
**Answer:** Fixed — D-7 names both hand-written inits, the non-optional `@State` model, the overlay binding `$lead.toast`, and the context and header parameters.

### X-4 · major · `LocalizedStringResource` cannot take `%@` after the fact
**Answer:** Fixed — with M-1; § Strings lists the Person members reused.

### X-5 · major · STEP-1 check unreachable; split steps do not build
**Answer:** Fixed — STEP-1 is one buildable unit (slot, model, contents, views, host); checks scoped to `macos/App`; tests compile in STEP-5.

### X-6 · major · Person behavior changes undecided
**Answer:** Fixed — D-5 keeps the Person label keys and the `personLikenessThumb` fallback (`?? L10n.string(withImage)`); § Strings lists the Person tip and tile functions; AC-2.4, T-1, T-3.

### X-7 · major · Lead held in two places
**Answer:** Fixed — with M-2.

### X-8 · major · Id gaps across packs
**Answer:** Fixed — with M-3: one rule for every kind, Person ids old → new in D-3.

### X-9 · major · Undecided edges
**Answer:** Fixed — Reach row (saving guard, failure then success) and Edges (empty Source title, header path with no lead tile → AC-2.4).

### X-10 · minor · Properties label slots and trailing comma
**Answer:** Fixed — D-11: five slots, `bound` is `rowBoundAnnouncement`, kinds appended only when present; AC-3.3, T-7.

### X-11 · minor · AC-4.1 grep scope
**Answer:** Fixed — scoped to `macos/App`.

### X-12 · nit · Host row wording
**Answer:** Fixed — §4 Host row.

### X-13 · major · Brief: relationship sections, tile spec source, list thumb size
**Answer:** Fixed — brief frame 1 adds a "Part of" section and cites S13-D1 frames 4i/4j for tiles; EPP-8 names the List's 44pt slot.

### X-14 · minor · Toast naming; Event frames
**Answer:** Fixed — brief inventory Toast row names `PVToast` via `vocabularyToastOverlay`; frame 4 states the Event header lines and rows.

### X-15 · minor · D-13's second order has no test
**Answer:** Fixed — with M-5: one order only.

## Round 1 (2026-10-29, earlier plan)

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

### X-6 · minor · Per-kind `gallery.identifierRoot` not named
**Answer:** Fixed — STEP-2 sets `identifierRoot` to each kind's `pageIdentifier` (`persons.detail` / `events.detail` / `places.detail`).

### X-7 · minor · Depiction crop mark on expanded Promote members
**Answer:** Fixed — D-10 / STEP-5: the group's crop mark applies to line and depiction `memberRow` crops.

### M-7 · blocker · D-10 rebuilds neighbor kind from page rows and falls back to the sheet kind
**Plan:** D-10 used `neighbor()?.kind ?? kind.markKey`; rejected adding relationship kind to proto.
**Codebase:** `claimevidence.Relationship.Kind` already exists; handler uses `rel.Kind` for sections; proto omits it; `PromoteFlow.row` is nil for off-page neighbors.
**Better shape:** Wire `kind` on `PromoteClaimEvidenceRelationship`; board reads `relationship.kind`.
**Answer:** Fixed — D-10 / AC-3.5 / T-7b / T-7d / STEP-5 / FFI section: field 20, handler map, no page-row lookup, no sheet-kind fallback (Stop on empty/unknown). Skills + Out updated.

## Close

No open blockers. Executability round 2: **Ready** (X-6/X-7 Fixed). Maintainability round 2: **Fits with changes** → M-7 Fixed. Plan may go to **Planned**.
