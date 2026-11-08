# Review — Spike 16 map (Batch 1)

**Verdict:** Fits with changes — all findings answered; no blocker open.
**Rounds:** 1. **Reviewers:** maintainability (map review, [`review-plan`](../../../.agents/skills/review-plan/SKILL.md) "Map review").

## To user

- **The re-cut narrows two consolidations.** You asked for PRs that overlap in scope, domain or files to be merged. The review split two back out, and I applied both, recommending you keep them: P-2 (the places cap) left S16-01 for its own PR, S16-08, because it is not a Person read, lives in another package, and would have made S16-01 and then S16-05 wait on Spike 14's S14-40 (K-3); P-6 (stale Source words) left S16-04 for its own PR, S16-09, because it shares no file with the failed-reload notice and would have waited on brief S16-D1 for no reason (K-7). The review also merged two (K-1: the Timeline's pages and the fold focus are now one PR, S16-07), so Batch 1 is nine PRs. Say if you would rather re-merge either split.
- ~~**Where the death stub sits between a baptism and a burial** (P-1's neighbour).~~ **Answered (user, 2026-11-08): just before the burial.** A placeholder death sits just before the earliest dated burial or funeral whenever one is recorded; the marker rule, bar 1, S16-02, the audit's P-1, R-24a and Spike 13's lifeline rule say so.

## Writer's own calls

Decisions no source dictated, so you can veto them:

- The cap sizes: 20 places per event; 100 relatives per role in the Family read; implied siblings only through at most `MaxPivotEnd` (10) parents; 10 records of one person per Evidence graph (reusing `MaxEntityMembers`); 75,000 record links lines (the number the diagram already draws).
- Past a crafted cap, the reads drop the rest silently; only the record links diagram says so.
- Every fold pages at 50, in every section, not only the Timeline: one fold rule (K-1). Real data only meets it on a long Timeline, where 150 hidden events take three clicks instead of one. You suggested "maybe the first 10 participants"; the participant fold already shows the first 5 at Regular and 12 at Wide (your call of 2026-11-03), so those stay.
- One brief, S16-D1, draws both notices (the failed reload and the partial diagram), since both are Person page notices.
- K-3's optional rider (dropping the unread `RecordLinksGraph` fields in S16-01) is declined: P-22 stays parked, as you asked for the low items.

## Round 1

### K-1 · blocker · Two PRs reshape the fold mechanism separately, and nothing orders them
S16-02 paged the Timeline's folds and S16-06 moved fold focus into `ConclusionShowMoreRow`, both on `ConclusionFold` ("the one fold rule of a Conclusion page's lists", reveal-all), with each section keeping its own fold state and no order between the PRs. **Better shape:** one preparatory PR for a paged fold that moves focus, ahead of both.
**Answer:** Fixed — new S16-07 "One paged fold that moves focus" (P-14 and P-7's pages); rule "One paged fold that moves focus" replaces "A Timeline fold bounds its work" and "Folds move focus through the Show-more row"; S16-02 depends on S16-07 and keeps P-1, building only shown rows and click targets; S16-06 keeps P-11 and P-12.

### K-2 · blocker · "Failure beside a value" is built for the Conclusion page alone, though three readers already need it
`EvidenceGraphModel.graphReadError`, `SourcesModel.statusCountsError` and `ConclusionListPresentation` each face the same silent refresh failure. **Better shape:** one reading on `QueryHandle` (loading, failed, or loaded with an optional refresh failure) and a failure revision for re-announcing; `read` unchanged; the skill and the doc teach it.
**Answer:** Fixed — rule "A refresh failure is read beside its value" and S16-04's Ships; `add-workspace-place` "Stale-while-revalidate" and the `QueryHandle` row of `macos-client-patterns.md` are in S16-04's Context; the Conclusion lists are parked in [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md) (they need a brief).

### K-3 · major · S16-01 ties the record-links change to Spike 14's S14-40, which blocks S16-05
Only P-2 touches S14-40's files, and it is not a Person read. **Better shape:** split P-2 into its own PR after S14-40.
**Answer:** Fixed — new S16-08 "Cap the places an event folds", depending on S14-40; S16-01 has no cross-spike dependency; S14-40's scope entry now names S16-08. The optional rider (P-22's unread fields) is declined (Writer's own calls). The split is listed under To user.

### K-4 · major · The places cap goes on one of four reads of an event's places; the Family cap needs two spots
`compose.go` reads an event's places in `eventPlaceGraph`, `lifeOwners`, `composeEvents` and `eventPlaces`; `Families` reads the person's kin in two loops.
**Answer:** Fixed — the rule and S16-08 name one capped helper for all four reads; the rule and S16-01 name one capped kin per person for both loops; the audit's P-2 and P-3 fixes say the same.

### K-5 · major · `RecordLinks.MaxLines` breaks the map's own cap rule
A display bound is not a matching tunable, and `Pairs` must count every line before keeping the strong ones.
**Answer:** Fixed — `maxRecordLinkLines` is a `persondetail` package constant carrying `maxArcs`'s `// decided:` provenance; `Pairs` reports whether it counted; no knob, inventory row or `add-match-rule` knob edit.

### K-6 · major · No PR deletes Swift's `maxArcs` and strong-first pass
**Answer:** Fixed — S16-01's Ships deletes them, in the PR that makes them dead.

### K-7 · major · S16-04 merges P-6 into a brief-gated PR it shares no file with
**Answer:** Fixed — new S16-09 "Conclusion pages follow the Sources their records cite", ungated, no dependencies. Listed under To user.

### K-8 · major · S16-06 is about 14 steps across three items, and P-16 belongs with P-12
**Answer:** Fixed — P-14 moved to S16-07; S16-06 is P-11, P-12 and P-16, with the leave-or-stay decision in a function a unit test reaches; P-16 left `structural-debt.md` and has a section in the audit.

### K-9 · major · S16-01 makes skill and doc lines false without naming them
**Answer:** Fixed — S16-01's Context lists `add-match-rule` line 33, the `kinship.Siblings` comment, the `matching-model.md` rows and record links paragraph, and the `PersonFamilyContent.swift` comment; the cap rule names the Spike 13 fold wording it supersedes for crafted data. `add-match-rule` line 34 ("six knobs") stays true now that K-5 adds no knob.

### K-10 · minor · The cap precedent named in the rule is wrong
**Answer:** Fixed — the rule cites `graphcache` `maxWalkEdges` and `PersonRecordLinksContent.maxArcs`.

### K-11 · minor · A second optional marker field allows "beside" and "before" to be set together
**Answer:** Fixed — rule "A marker's position is one value": a `oneof` on the wire and one Swift enum; S16-02's Ships.

### K-12 · minor · S16-04's new follow does not state its fan-out, and the skill omits the slot it uses
**Answer:** Fixed — the rule "A cached detail follows the Sources its records cite" states the fan-out (any write in a cited Source revalidates a visible detail, one read; hidden details evict), and S16-09 adds `.source` to `add-workspace-place`'s slot list.

### K-13 · minor · P-11 needs a kit change that the scope does not name
**Answer:** Fixed — S16-06's Ships changes `PVSegmentedControl`'s handle to land on the selected segment, in the component and its doc comment.

### K-14 · minor · The new focus keeper belongs beside `FocusClaim`
**Answer:** Fixed — the rule and S16-07 put it in `macos/App/Platform/` and cite Wide rule 5.

### K-15 · minor · The diagram note covers only Go's dropped lines, not Swift's dropped nodes
**Answer:** Fixed — one partial state covers "did not count", "lines dropped" and "graphs dropped"; S16-05's Ships and Check.

### K-16 · nit · S16-05's Context and the parallel-run note
**Answer:** Fixed — `CircularLinkDiagram/` left S16-05's Context (it does not edit it, so A-41 stays parked); the parallel-run note says S16-04 lands before S16-05.
