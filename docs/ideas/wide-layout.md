# Wide layout: pages waiting for a Wide frame

**Status:** idea, 2026-10-29. A queue, not a spike. Each line becomes a brief ([`add-design-brief`](../../.agents/skills/add-design-brief/SKILL.md) "Wide layout (binding)") and a PR when its page is next touched, or as a batch in an even spike.

Spike 14's Batch 6 gives every page a Regular / Wide width class (1600pt of pane, S14-24) and a Page columns kit component (S14-D1, S14-25). The rules and patterns are in [`macos-client-patterns.md` § 9](../macos-client-patterns.md#9-window-widths-regular-and-wide). These pages still show their Regular layout at Wide:

| Page | Wide pattern to start from | Notes |
| --- | --- | --- |
| Source page | Page columns with a secondary column; the pinned header slot | First caller of Page columns' header slot and pinned mode (S14-D1 WL-3, WL-8, WL-W4), which S14-25 does not build. Today it left-aligns its column and pins `SourcePageIdentityHeader` with its own copy of the column frame |
| Person, Event and Place pages | Page columns with a secondary column | On Page columns, primary only, after S14-25 |
| Sources, Persons, Events and Places lists | List beside the selected detail | `ConclusionListPage` is on Page columns, primary only, after S14-25; the Sources list keeps `SourcesListColumns` |
| Properties, Metadata, Source types | List with a side pane | Each has a fixed side pane today; its brief decides whether the list or the pane takes the width (S14-24 D-6). `MetadataView` and `SourceTypesView` are near-copies of each other |
| Citation composer | Canvas / editor | Has its own 1400pt two-up / three-up switch in `CitationComposerView.composeSplit`; moves onto the width class with its brief (S14-24 review M-2) |
| Evidence graph | Canvas with the inspector docked open | |
| Promote | To be decided in its brief | |
