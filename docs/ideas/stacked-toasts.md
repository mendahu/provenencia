# Stacked toasts

**Status:** parked 2026-11-12 from dogfood while testing portrait swaps (user). The bug a quick replacement exposed, where a new toast vanished on the old toast's timer, was fixed in `PVToast` (quick PR #796). This file is the feature the bug pointed at.

## Problem

A screen shows one toast at a time. A second message replaces the first, so a researcher who acts twice in a row (two portrait choices, two saves) sees only the latest message, and the earlier one is gone before it could be read. Each message should stay its own time.

## Rough shape

- **A queue, not a slot.** Every overlay holds one optional toast today (`Binding<VocabularyToast?>` through `vocabularyToastOverlay`, plus `bridgeToast` in the Citation Composer and `error` in Onboarding). That becomes a small toast list a model appends to, with each entry carrying its own id and expiry.
- **A stack view in the kit.** A `PVToastStack` (design-system layer, [`add-ui-component`](../../.agents/skills/add-ui-component/SKILL.md)) draws the list top-trailing. New toasts enter at the top, older ones slide down, and each leaves on its own timer or its dismiss button. `PVToast` stays the single card; its content-keyed `.task(id:)` wait already gives each card an independent expiry.
- **Callers move from "set" to "add".** About 13 files show a toast today: `VocabularyChrome`, `ConclusionDetailPage`, `ConclusionPictureViewer`, `WorkspaceView`, `PropertiesView`, `SourcePageView`, `SourcesListView`, `EvidenceGraphView`, `SourceTypesView`, `MetadataView`, `CitationComposerSubjectGroups`, `OnboardingView`, and `PVToast` itself. The likeness viewer's rule that only one surface draws the toast (`ConclusionLeadModel.toastInViewer`, S13-15 D-10) carries over to the list.

## Open questions

- A cap on visible toasts (three?), and whether the oldest drops or collapses.
- Duplicates: a repeat of the same message refreshes its timer or adds a second card.
- VoiceOver: each toast already announces itself on appear; check a burst does not queue a long run of announcements.
- Reduce Motion: the stack's slide falls back to a fade, as the single toast does.

## Check

Two toasts added within five seconds both show, stacked, and each leaves five seconds after it appeared. Dismissing the lower one leaves the upper one's timer untouched.

Needs a Claude Design brief ([`add-design-brief`](../../.agents/skills/add-design-brief/SKILL.md)) before a plan: it changes how every toast looks when two arrive together.
