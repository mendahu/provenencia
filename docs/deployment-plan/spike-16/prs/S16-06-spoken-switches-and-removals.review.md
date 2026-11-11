# Review — S16-06 Section switches and removals are spoken

**Verdict:** Fits — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None.

## Writer's own calls
- Only a switch chosen on the switcher speaks; Back, Forward and links do not (D-5, AC-1.6).
- Focus is claimed only when it was in the sections column or on the switcher, read with a Bool focused value (D-3), the app's first use of `@FocusedValue`.
- The Remove announcement names the record, not the person (D-11).
- `conclusions.memberCount` keeps its key; only its values and comment change (D-12).
- No brief (user, 2026-11-11) and "{Section} shown." (user, 2026-11-11) are the user's answers, not the writer's.

## Round 1

### X-1 · blocker · The "finds nothing" `rg` check cannot pass
The sweep pattern hit docs, the audit, archives, a code comment and the plan.
**Answer:** Fixed — §6 Strings scopes the check to `macos/App` and `macos/ProvenenciaTests` with `'and the handle stays|%lld members?"|[0-9] members?"'`; docs, audit, archives and comments are an Out row; the Stop condition asking to sweep docs is gone.

### X-2 · major · Two preview literals, not one
**Answer:** Fixed — §4 Preview copy, §6 Deletes and STEP-6 name both previews.

### X-3 · major · EventDetailTests assertion has no home
**Answer:** Fixed — T-8 adds `EventDetailTests.headerCountsRecords` with 4 and 1.

### X-4 · major · T-4/T-5 navigation and session setup
**Answer:** Fixed — T-4 writes the setup lines and says the nil-history path is intended.

### X-5 · major · `confirm(onPageOf:)` edge behaviour
**Answer:** Fixed — D-8 states the nil-pending guard, that a failure calls neither closure, and that the closure runs before `pending = nil`; T-6 *(new)* and AC-2.5 cover the failure.

### X-6 · major · Focused-value key under-specified
**Answer:** Fixed — D-3 writes the key struct, the `FocusedValues` property, the call site and the read (Bool, after M-4).

### X-7 · major · D-2 and D-1 disagree for a disabled selection
**Answer:** Fixed — D-2 says it holds because the Conclusion switcher has no disabled item; a Stop condition covers one appearing.

### X-8 · minor · STEP-3's check reaches ahead of STEP-6
**Answer:** Fixed — STEP-3 checks manual steps 1–3; the header copy moved to step 4.

### X-9 · minor · AC-1.1's Tab clause has no producing step
**Answer:** Fixed — kept in the AC as SwiftUI's key loop; a Stop condition reports it if it fails.

### X-10 · minor · `announce:` default closure annotation
**Answer:** Fixed — §11 Risks names the `@MainActor in` form.

### X-11 · nit · T-3 fixtures
**Answer:** Fixed — T-3 names the helpers and the nil-listed case.

### M-1 · major · A second public `confirm` beside the `apply:` hook it leaves unused
**Answer:** Fixed — D-9: the `apply:` variant becomes the private `write`; `CatalogMutation.swift`'s comment is updated (STEP-4); sweep names added.

### M-2 · major · The shared flow reaches into the Sources section's namespace
**Answer:** Fixed — D-10: both spoken strings live in `L10n.Unpromote`; `conclusions.sources.personDeletedAnnouncement` goes (STEP-5).

### M-3 · major · `pendingSwitch` guards a hold that cannot happen
Verified: `leaveGuard` is set only by `CitationComposerView.swift` and `PromoteView.swift`.
**Answer:** Fixed — D-5 posts right after `select` when the open section equals the target; the state and `onChange` are gone; a Stop condition covers a later async path.

### M-4 · minor · The focused value can only equal `from`
**Answer:** Fixed — D-3/D-4 use a Bool `sectionsFocused`.

### M-5 · minor · `onPageOf entityId` restates `handleID`
**Answer:** Fixed — D-7 decides on `item.handleID`; `ownHandleDeletedRef` is shared with `message`.

### M-6 · minor · Codebase fit omits `pvAnnounce`
**Answer:** Fixed — Codebase fit row explains why a switch posts directly.

### M-7 · minor · Kit README row
**Answer:** Fixed — STEP-1 updates the README row.

### M-8 · nit · Return the outcome instead of injecting `announce`
**Answer:** Logged — the view would hold the post again (D-8 Rejected).

### Outside lens (CON) · Space on a non-selected segment reports `switcherFocused` false
**Answer:** Fixed — §6 Edges and manual check step 2 cover it; AC-1.2 now uses ⌘4 from the focused selected segment, which is what T-2's case describes.

## Round 2

Maintainability: **Fits**, no new findings; every round-1 answer checked against the code. Executability: **Ready with changes**.

### Q10 · note · `macos-client-patterns.md` Cache contract item 2 still names `UnpromoteFlow.confirm`'s `apply`
**Answer:** Fixed — STEP-7 rewrites that sentence to name `confirm(session:userID:leavingPage:)`.

### M-3 follow-up · note · The band could build `select` itself now that it reads navigation
**Answer:** Logged — the band keeps its `select` parameter; Event and Place share the frame, and the change is out of this PR's findings.

### X-1 · blocker · STEP-4 makes `apply:` private while `PersonSourcesSection` still calls it
**Answer:** Fixed — the `PersonSourcesSection.confirm()` rewrite moves into STEP-4; STEP-5 only removes the old string.

### X-2 · blocker · STEP-5's `rg` would match `UnpromotePageOutcome.after`
**Answer:** Fixed — `deletedHandles.contains` is searched only in `macos/App/Features/Conclusions`.

### X-3 · major · `@discardableResult` unstated
**Answer:** Fixed — D-8 and D-9 put it on both public `confirm` methods; D-9 says the no-`apply` overload is new.

### X-4 · major · Enum labels and the `switch` local
**Answer:** Fixed — D-7's cases are unlabelled; D-5 names the local `sectionSwitch`.

### X-5 · minor · Test fixtures
**Answer:** Fixed — STEP-4 says to copy the effects shape before deleting the old test; T-6 names the error and drops the unassertable "no write".

### X-6 · minor · Mouse click from the column
**Answer:** Fixed — manual check step 2 adds the click; a Stop condition covers it failing.

### X-7 · minor · Catalog fields, imports, temptation
**Answer:** Logged — `add-localized-string` covers the catalog fields; the rest needs no change.
