# Review — Spike 16 map (Batches 1 and 2)

**Verdict:** Round 1 (Batch 1): Fits with changes. Round 2 (the whole map with Batch 2): Reshape for slices 2a – 2d, answered by reshaping them; Batch 1 and slice 2e fit with changes. No blocker open.
**Rounds:** 2. **Reviewers:** maintainability (map review, [`review-plan`](../../../.agents/skills/review-plan/SKILL.md) "Map review").

## To user

- **The re-cut narrows two consolidations.** You asked for PRs that overlap in scope, domain or files to be merged. The review split two back out, and I applied both, recommending you keep them: P-2 (the places cap) left S16-01 for its own PR, S16-08, because it is not a Person read, lives in another package, and would have made S16-01 and then S16-05 wait on Spike 14's S14-40 (K-3); P-6 (stale Source words) left S16-04 for its own PR, S16-09, because it shares no file with the failed-reload notice and would have waited on brief S16-D1 for no reason (K-7). The review also merged two (K-1: the Timeline's pages and the fold focus are now one PR, S16-07), so Batch 1 is nine PRs. Say if you would rather re-merge either split.
- ~~**Where the death stub sits between a baptism and a burial** (P-1's neighbour).~~ **Answered (user, 2026-11-08): just before the burial.** A placeholder death sits just before the earliest dated burial or funeral whenever one is recorded; the marker rule, bar 1, S16-02, the audit's P-1, R-24a and Spike 13's lifeline rule say so.

- **Kinship is composed, not indexed (K-20, applied).** You approved "indexes for the walks", which named a kinship index. Round 2 showed the existing per-event kinship memos are already exact, and only the two slots that store composed answers (`for`, `siblings`) empty on every write. S16-16 now composes a person's implied links and siblings from those memos at read time: no catalog-wide index, no reverse map, no warm at open, 10–20 MB less, and a term edit drops only the memos it touches. The cost is a few map reads per `For` call, which S16-16 measures against Promote's budgets. Recommendation: keep it composed; if that measures slow, the index comes back as a follow-up.
- **Batch 2 is 24 PRs, not 20 (K-22, applied).** Four PRs were past the ~10-step plan size and split: the records tier (S16-13) from its readers (S16-30); the place entries (S16-18) from the composers that read them (S16-31); the match index (S16-20) from the narrowed re-rank (S16-32); and a preparatory canon view (S16-33) ahead of the walk's removal (S16-22). The kinship PR shrank instead (K-20).

## Writer's own calls

Decisions no source dictated, so you can veto them:

- The cap sizes: 20 places per event; 100 relatives per role in the Family read; implied siblings only through at most `MaxPivotEnd` (10) parents; 10 records of one person per Evidence graph (reusing `MaxEntityMembers`); 75,000 record links lines (the number the diagram already draws).
- Past a crafted cap, the reads drop the rest silently; only the record links diagram says so.
- Every fold pages at 50, in every section, not only the Timeline: one fold rule (K-1). Real data only meets it on a long Timeline, where 150 hidden events take three clicks instead of one. You suggested "maybe the first 10 participants"; the participant fold already shows the first 5 at Regular and 12 at Wide (your call of 2026-11-03), so those stay.
- One brief, S16-D1, draws both notices (the failed reload and the partial diagram), since both are Person page notices.
- K-3's optional rider (dropping the unread `RecordLinksGraph` fields in S16-01) was declined at first, to keep the low items parked; later the same day you folded it in with seven other parked items that touch Batch 1's files (map, "Where the notes went").

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

## Round 2 — the whole map with Batch 2 (2026-11-08)

A fresh maintainability reviewer read the whole map, the design doc and the S14-40 and S14-42 branches. Every code claim the map and design make held (the write path's `graphcache.Over`; `canonicalgraph`'s test-only caller after S14-42; `DropOnHandleCommit`'s two users; `dropMatchIndex` on any handle commit; the `effects.Set` fields; `graphsOfSource`; `graphprogress`'s scans; `PersonLifeOwners`; `claims.go` reading the graph's Source fields; `conclusiondetails.go:8`; `promotealign.Warm` loading every node), as did the cross-spike dependencies in both maps and the Batch 1–2 orderings. Writer's calls added in round 2: the records tier's own cap, Source records and the graph catalog filled on first read rather than at open, and three wire lists for inputs.

### K-17 · blocker · The registry's contract does not fit most of the caches it is meant to hold
Apply fed only `effects.Set` would re-derive the roots `OnCommit` already computes; no field says "drop when another cache changes" (`dropDerived` on `SetAliasIndex` and `SetNameFormat`; profiles keyed on the name pattern); only a whole Build exists, though kinship pivots, profiles and places are results per key; `Activity` goes unmentioned.
**Answer:** Fixed — rule "One cache registry, two ways to fill": per-key (today's `NodeMemo` folds in) and whole fills; Apply(g, roots, set) after the reload with `dropMemos`' roots; declared Reads replace `DropOnVocabulary` and `dropDerived`; a lifetime; `Activity`, displays and labels stay typed fields, explained in `docs/catalog-caches.md`. S16-10 and the design § One cache registry.

### K-18 · blocker · S16-13 and S16-20 each build their own per-graph cache that outlives the 64-graph LRU
`EvidenceGraph` already holds Subjects and Observations, and the candidates memo is an Evidence-graph memo.
**Answer:** Fixed — rule "Per-graph records live in the Evidence-graph store": a records tier per graph (Subject records and memos that outlive the loaded graph), one drop path, filled from a loaded graph or a split-out `graphfill` step under graphfill's check. S16-13 builds it; S16-32 keeps Promote's candidates there.

### K-19 · blocker · Profiles as a node memo and the shortlist index as a registry entry work out "what changed" three times
`dropMemos` deletes a profile before an Apply can read its old keys; profiles are catalog-wide anyway.
**Answer:** Fixed — S16-20 is one match index entry (profiles, keys and postings by handle); Apply re-profiles the roots and their events' people once, diffs keys, patches postings and records the changed keys for S16-32.

### K-20 · blocker · The kinship index swaps exact memos for a whole-catalog build
**Answer:** Fixed, and listed under To user — S16-16 composes `For` and siblings from the existing exact pivot memos; no index, reverse map, warm or Apply; bar 9 reads "kinship pivots".

### K-21 · blocker · The place index's whole Build runs on every reprojecting write
**Answer:** Fixed — rule "A write's one-shot graph fills only what it reads": S16-18's place entries are per key, filled a frontier at a time by a narrow load that skips location links, and on a write's graph only the chains it composes fill; `graphcache` exports `nodeKey` and `followKeys` so entries record what `InputScope` would.

### K-22 · major · Five Batch 2 PRs are bigger than one plan of about 10 steps
**Answer:** Fixed — S16-13 / S16-30, S16-18 / S16-31, S16-20 / S16-32 and S16-33 / S16-22 split; S16-16 shrank with K-20 (listed under To user).

### K-23 · major · `claims.go` and `recordlinks.go` are rewritten by five PRs, and the graph misses real dependencies
**Answer:** Fixed — S16-15's claims and record links ordering moved into S16-14 (S16-15 → S16-14); S16-01 → S16-12; S16-24 now depends on S16-30, S16-14 and S16-19 beside S16-17; S16-12 → S16-25, whose Ships names S16-12's `.source` follows; the shared-files note lists each file's order.

### K-24 · major · S16-11's "`gallery.ForEntities` goes" misses a caller on the write path
**Answer:** Fixed — S16-11 names `SetPreferredLikeness`'s check (`canonicalentities.go:275`), which composes through the node on `graphcache.Over(tx)`, and measures open time and a write at the 6,000-event city beside memory.

### K-25 · major · Docs and skills made false that no PR names
**Answer:** Fixed — each PR's Context names them: S16-10 (`.cursor/rules/catalog-queries.mdc`, `use-catalog-session`), S16-12 (`effects/README.md`), S16-16 (`add-match-rule` lines 88 and 91, `matching-model.md`), S16-20 (`use-catalog-session`), S16-33 and S16-22 (`add-match-rule` line 33, `matching-model.md` line 31, `promote-graph-alignment.md` §7 and §12), S16-26 (`macos-client-patterns.md` rule 2 and "Conclusion keys", `add-workspace-place` Refresh), S16-27 (`add-workspace-place`, `macos-client-patterns.md` rule 5), S16-29 (`add-workspace-place`).

### K-26 · major · Apply adds work to every commit, and no PR states its fan-out
**Answer:** Fixed — rule "Apply states its fan-out" (per commit at Scale and at bar 2's crafted sizes, capped where crafted fan-out reaches it); S16-11, S16-15, S16-16, S16-18 and S16-20 name their case in Check. The open-versus-first-read contradiction is settled: only the match index fills at open (rule "Warm only what every proposal needs"); Source records and the catalog fill on first read.

### K-27 · major · S16-29's prefetch keys make a second list beside S16-03's section reads
**Answer:** Fixed — S16-29 adds a `prefetch` tag to S16-03's list of section reads, declared by Family and Timeline, and depends on S16-03.

### K-28 · minor · Put Batch 1's caps where Batch 2 keeps them, and keep `InputKey` to one job
**Answer:** Fixed — S16-01 caps relatives per role in `kinship.KinOf` and records per graph in one function (Batch 1 rule and S16-01 Ships); rule "Batch 1's caps stay where S16-01 and S16-08 put them" says the place entries hold place-to-place links and the composers keep S16-08's helper; S16-24 sends graphs and Sources as their own wire lists and leaves `graphcache.InputKey` with its one kind.

