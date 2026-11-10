# Wide layout: pages waiting for a Wide frame

**Status:** idea, 2026-10-29. A queue, not a spike. Each line becomes a brief ([`add-design-brief`](../../.agents/skills/add-design-brief/SKILL.md) "Wide layout (binding)") and a PR when its page is next touched, or as a batch in an even spike.

Spike 14's Batch 6 gives every page a Regular / Wide width class (1600pt of pane, S14-24, landed) and a Page columns kit component (brief S14-D1). Page columns has no migration PR: each page moves onto it in its own redesign (user, 2026-10-29), starting with Spike 13's Person page frame (S13-D1, S13-10), which builds it. The rules and patterns are in [`macos-client-patterns.md` § 9](../macos-client-patterns.md#9-window-widths-regular-and-wide). These pages still show their Regular layout at Wide:

| Page | Wide pattern to start from | Notes |
| --- | --- | --- |
| Source page | Page columns with a secondary column; the pinned header slot | Landed (S15-03): the page is on Page columns with the pinned header, `primaryFill` and `fillSplitSideMax`; `SourcePageOverviewLayout` keeps Description and Credibility beside Metadata at Regular and puts Metadata in a side column at Wide, over the same four subviews so drafts survive crossing 1600. Spike 15 adds the Artifact metadata board and the citation preview to it ([`source-provenance-date.md`](../deployment-plan/spike-15/source-provenance-date.md)) |
| Person, Event and Place pages | Page columns with a secondary column | On Page columns since S13-10: `ConclusionDetailPage` is built on it, so Person, Event and Place are centered at Wide. Person's sections are primary-only, except Sources, which fills the width with a pinned side column (S13-28); Event and Place keep their own Wide arrangement for their redesigns |
| Sources, Persons, Events and Places lists | List beside the selected detail | `ConclusionListPage` moves onto Page columns with the lists' redesign; the Sources list is drawn at Wide with the Source date (Spike 15) and keeps `SourcesListColumns` |
| Properties, Metadata, Source types | List with a side pane | Each has a fixed side pane today; its brief decides whether the list or the pane takes the width (S14-24 D-6). `MetadataView` and `SourceTypesView` are near-copies of each other |
| Citation composer | Canvas / editor | Has its own 1400pt two-up / three-up switch in `CitationComposerView.composeSplit`; moves onto the width class with its brief (S14-24 review M-2) |
| Evidence graph | Canvas with the inspector docked open | |

Promote widened in S13-23: its list goes to 1600 pt and its target menu to 460 pt at Wide (S13-D11 frame 27W).

## Narrow panes

The window can be as narrow as 700pt (a pane of about 436pt). Pages are drawn at Regular and Wide; below the Regular frame some do not fit:

- **Person page, Sources rows.** The Source rows' fixed Type, Date and Ref columns (442pt with their gaps, `PersonSourcesColumns`) do not fit the 372pt column of a 436pt pane, so the citation column has no room. Found on the S13-D10 board (frame 15r, 2026-11-07); the rows are S13-D6's layout, and their narrow arrangement waits for the next change to the Sources section (user, 2026-11-07).
