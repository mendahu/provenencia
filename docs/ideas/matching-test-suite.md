# Matching test suite and labelled calibration

**Status:** idea, not scheduled; likely the core of **Spike 11**, a large-scale matching refinement. Captured 2026-10-16 from a design conversation. It expands the **Calibration** bullet and the open question in [`name-matching-enhancements.md`](name-matching-enhancements.md). Model: [`promote-graph-alignment.md`](../promote-graph-alignment.md), [`matching.md`](../matching.md).

Spike 11 scoping has started. Layers 1–2 of the comparison stack (part of "make the algorithm graspable") are in [`deployment-plan/spike-11/comparison-layers.md`](../deployment-plan/spike-11/comparison-layers.md).

## The problem

Promote graph alignment works reasonably well, but the space of layouts is far too large for manual QA. One change to the walk or to a registry number can move rows anywhere: names against dates, edge fan-outs, directed signatures, orphans, contested handles, decided rows. We need to change the algorithm with confidence, and to tune its numbers from evidence rather than by hand.

What makes this tractable is that `graphalign.Align(layer, canon, stats, fixed, cfg)` is pure and deterministic over plain structs. Almost all of the hard testing can happen without SQLite or the app.

Today's tests are good but expensive. Each scenario is 50–120 lines of hand-built Go (`TestGracieObituaryGolden`), and most assert only a few targets. That cost per scenario is why coverage stays thin.

## First: make the algorithm graspable

Before building any of the tooling below, the algorithm itself has to be structurally sound and understood. Today it is too complex to hold on one page, and there is no shared abstraction to talk about it with. The test suite and the labeler only pay off if there are clear levers to pull, and if everyone means the same thing by "score", "support" or "band".

This is a refinement pass. It may be only documentation and renames, or it may need code changes. To do:

- **One-page model.** A single diagram and page covering the whole pipeline: pairwise comparison → property-only `Rank` → seed → best-first walk → one-hop refinement → bands and flags. Each stage gets one sentence on its input, its output, and the knobs that affect it.
- **A shared vocabulary.** One glossary (similarity, points, weight, scale, support, credit, signature, fan-out, band, margin, anchor, …), with code, docs and UI copy renamed to match. Where two names mean one thing, or one name means two, fix it.
- **A knob inventory.** Every tunable, its default, what it changes, and which tests pin it, across `core/match/registry.go`, `core/match/profiles.go`, `core/graphalign/registry.go` and `core/autoreconcile`. Find magic numbers living outside the registries and move them in.
- **Are the levers right?** Check that the knobs are independent and cover the behaviors we actually want to tune. The learning step needs each row's score to break down cleanly into parts, each one a weight times an observed feature. If a stage can't be expressed that way, decide whether to restructure it or tune it by search.
- **Known gaps.** List where the algorithm is "not quite there" today, from dogfood and from [`name-matching-enhancements.md`](name-matching-enhancements.md), so the refinement and the test suite aim at the same problems.

Pieces 1 and 2 below can start alongside this, since goldens and invariants protect the refactor. Pieces 3–7 wait until the model and the knobs settle.

## Five pieces

### 1. Scenario files with golden output

Move scenarios out of Go into `core/graphalign/testdata/scenarios/*.txt`, in a compact text format:

```
# Gracie obituary
layer:
  PER-G  person  name="Gracie Gray Gates"      fixed=PER-Z
  EVT-B  event   event_type=birth
  PER-C1 person  name="Alice Gates"
  PER-G -[subject@birth]-> EVT-B
  PER-G -[parent]-> PER-C1
canon:
  PER-Z  person  name="Gracie Gray Gates"
  EVT-B1 event   event_type=birth
  PER-Z -[subject@birth]-> EVT-B1
stats:
  fanout subject@birth=1 parent=3
expect:
  EVT-B  -> EVT-B1 strong
  PER-C1 -> new|skip
```

- One table-driven test runs every file.
- A `-update` flag writes a `.golden` file next to each scenario with the full proposal: target, band, score, alternatives and reasons.
- The hand-written `expect:` block holds what matters. The golden file catches everything else.
- A registry change that moves 14 scenarios shows up as a reviewable diff, row by row, with reasons.

Adding a case should take a couple of minutes. Pieces 3–5 all produce scenarios in this format.

### 2. Invariant tests over random graphs

Generate small random layers and canons, and check rules that must always hold. [`pgregory.net/rapid`](https://github.com/flyingmutant/rapid) shrinks a failure to a minimal graph. Most rules come straight from the design principles:

- **Order doesn't matter.** Shuffling subjects, bridges, handles or edges gives the same proposal.
- **One-to-one.** Within a layer, no handle is assigned twice.
- **Decided rows don't move.** Every `fixed` row comes back as given, including New and Skip.
- **Stable on re-run.** Re-running with the accepted rows added as `fixed` changes no other row.
- **Monotone evidence.** An added agreeing property or matching neighbor never lowers a candidate's score. An added conflict never raises it.
- **Unrelated data doesn't interfere.** Handles that share no values or edges with the layer don't change a strong match.
- **Ambiguity is never strong.** Two identical handles in canon: neither is proposed at the strong band.

These catch whole classes of bugs in propagation, refinement and tie-breaking that single examples miss.

### 3. A simulated world with known truth

The standard way to evaluate entity resolution:

1. **Generate the truth.** A few families over three generations, with births, marriages, deaths, residences, and places in a `part_of` hierarchy.
2. **Generate Sources from it.** Obituaries, census rows, marriage records. Each sees part of the world through controlled noise:
   - typos and nicknames (the nickname seed helps);
   - dropped given names, surname changes at marriage;
   - imprecise or off-by-one dates;
   - missing places, or the same place named at another level of the hierarchy;
   - deliberate look-alikes: two John Smiths born a year apart in the same town.
3. **Promote Source by Source,** accepting every suggestion, like an over-trusting researcher. Compare the resulting handles to the truth.

This gives real metrics: false-merge rate (the one that matters most, principle 5), false-split rate, and calibration per band (what fraction of strong proposals are right). It also allows:

- **Order independence.** Promoting A then B should cluster the same as B then A, or at least never falsely merge.
- **Scale.** A few thousand people catches frequency and fan-out statistics behaving differently than at fixture size.

The noise settings are the knobs for covering permutations. Coverage grows by turning them, not by writing scenarios by hand.

### 4. A scoreboard that gates changes

An evaluation run (`go test -run TestEval` or a small command) over pieces 1, 3 and 5:

- prints a confusion matrix per band and per kind;
- compares against a checked-in baseline, and `scripts/check-pr.sh` fails when false merges go up or recall drops past a tolerance;
- reports each case's **margin** to the nearest band edge, so fragile cases are visible;
- optionally nudges each registry knob ±10% and records which cases flip, to show which tunables matter.

### 5. Dogfood sessions become cases

Every Promote session already produces labelled data: the researcher's Done is the truth, and every overridden suggestion is a disagreement. A debug export (FFI call or hidden menu item) writes that session's `Layer`, `Canon`, `Stats`, `fixed` and final decisions as a scenario file. Manual QA stops being repeated work and becomes a source of regression cases.

Pair it with a few **adapter equivalence** tests in `promotealign`: build a small catalog, and check that the loaded `Layer` and `Canon` equal a hand-written struct. That keeps scenarios tested at the pure level honest about the real loader.

## The labeler

A small standalone tool that shows generated cases and asks the researcher for a band. A thousand answers become training data for the registry and regression cases at once. This is the "learned weights" of [`promote-graph-alignment.md`](../promote-graph-alignment.md) §4.1, reached long before real accept/reject history exists.

### Shape

- **`cmd/labeler`,** a Go program in the repo. A small local web server with one HTML page. It reuses `graphalign`, `match` and the simulator directly, and appends answers to a JSONL file under `testdata/`. No Swift, FFI or catalog.
- **Cases come from the simulated world (piece 3), not uniform random graphs.** Random graphs are mostly obvious (unrelated or identical) and teach nothing. Pick a person in the world, generate a noisy Source mentioning them, and pair that Subject with either the true handle or a deliberate look-alike. Cases look like genealogy, and the generator knows the truth.
- **Active sampling.** After the first ~100 labels, favor cases near a band edge, cases where a first fit disagrees with the current registry, and noise patterns with few labels so far (a nickname plus an imprecise date). Well-chosen labels teach far more than random ones.

### Screen

- Side by side: the Evidence Subject and its neighbors on the left, the candidate handle and its canonical neighbors on the right, values lined up row by row.
- **The algorithm's band stays hidden** until after answering, so it can't anchor the label. Revealing it afterward is a cheap bug finder.
- One keystroke per answer: `1` no match, `2` weak, `3` medium, `4` strong, `S` skip. An optional note for "why".
- About 5% of cases repeat unannounced, to measure the labeller's own consistency. Disagreement with yourself is the ceiling no fit can beat, and it shows which band edges are fuzzy even to a human.

At roughly 10 seconds a case, 1,000 labels is about 3 hours, in sessions of any length.

### Each label is also a test

A labelled case saves as a scenario file (piece 1) with the label as its `expect:`. Agreement with the labels becomes a scoreboard metric (piece 4).

## How the learning works

### No single disagreement explains itself

The labeler does not work out why one answer differs from the algorithm. One disagreement is ambiguous: names may matter more, or birth events, or the label was a slip. Learning happens across all the labels at once.

### What gets fitted

A row's score is already a sum of parts: name points, date points, sex points, edge support per signature, minus conflict penalties. Each part is a registry **weight** times **what this case observed**. The band is a cutoff on the total. So each labelled case stores:

- the observed parts, as a feature vector;
- the label, and the generator's truth (same entity or not).

| name sim | date agree | birth edge | parent edges | … | label | algorithm |
|---|---|---|---|---|---|---|
| 1.0 | 0 | 1 | 0 | | strong | medium |
| 1.0 | 0 | 0 | 2 | | medium | medium |
| 0.8 | 1 | 1 | 0 | | strong | strong |

An **ordered logistic regression** finds the weights and band cutoffs that make the formula's band agree with the labels as often as possible. It tunes every registry dial at once.

It knows what to change from consistent patterns. If 60 cases with "same name plus a matching birth event" are labelled strong and the algorithm says medium, the only change that fixes all 60 is a higher birth-edge weight or a lower strong cutoff. The other labels decide which, by marking where each change starts to overshoot.

The output is a proposed diff to the registry files:

```
birth-edge support   1.5 → 2.1
name weight          2.5 → 2.3
strong cutoff        6.0 → 5.4
```

A human reviews it, runs the scoreboard and goldens, and commits it or doesn't. Nothing changes the shipped algorithm by itself.

Safeguards in the fit:

- **Uneven costs.** Labelled weak but proposed strong costs more than the reverse (principle 5), so the fit leans cautious.
- **Stay near today's numbers.** A pull toward the current values (regularization), so a few odd labels can't swing a weight.
- **Keep the structure.** Learn the registry's numbers, not a black box. The algorithm stays explainable, deterministic and on device (principle 6). With 20–40 tunables, 1,000 labels is plenty. A neural net would need far more data and would lose "every suggestion carries its reasons."

**Watch out:** a band mixes "how sure is it the same entity" with "how much evidence is there". Same name and nothing else may feel weak even when it is the same person. Recording both the label and the generator's truth lets the fit calibrate against each.

### The leftovers are the useful part

Disagreements that survive the fit fall into four groups:

1. **Wrong weights.** The fit fixed these; they are gone.
2. **A missing feature.** The labeller used evidence the algorithm can't see: birth years 30 years apart, the same spouse's name. No weight fixes this because the formula has no slot for it. Cluster the remaining misses by which parts were present and which noise the generator applied ("every miss involves a maiden-name change"). Notes help. A cluster is a code change: a new comparison or feature.
3. **A bug.** A case far off, such as strong with nothing truly agreeing. The row's reasons show which part produced the points; the cause is usually in the walk.
4. **Label noise.** The repeated cases measure it. A 50/50 split on "same name only" means that edge is fuzzy, and no formula will do better.

The labeler can show a per-case breakdown for triage: each part's points, and what would flip the band ("without the birth-edge support this would be medium"). `Row` reasons carry most of that already.

### What it can't do

The fit tunes numbers inside the current structure: a sum of parts and cutoffs. It doesn't invent comparisons, change how the walk propagates, or find bugs. Knobs that aren't simple weights (the 0.75 name floor, propagation, one-to-one refusal margins) need a slower search: try many settings, keep the one that scores best against the labels.

The loop: label → fit proposes numbers → leftovers point at missing features or bugs → a human changes code → label again.

## Loose plan

Each phase is usable on its own. Rough sizes, not commitments.

0. **Make the algorithm graspable** (docs first, then whatever code changes it turns up). The one-page model, glossary and renames, knob inventory, and lever review above. Phases 1–2 can run alongside it; the rest waits for it.
1. **Scenario format and golden runner** (one PR). Parser for the text format into `Layer` / `Canon` / `Stats` / `fixed`, the table-driven test, `-update`. Port the obituary golden and a handful of `align_test.go` cases as proof. Leave the rest of the Go tests in place.
2. **Invariant tests** (one PR). A random small-graph generator under `rapid`, plus the invariants in piece 2. Expect it to find a few real bugs; fix them in their own PRs.
3. **World simulator** (one or two PRs). Truth generator (families, events, place hierarchy), Source renderer with noise knobs, and a writer to scenario files. The hard part; everything after it depends on it.
4. **Scoreboard** (one PR). Evaluation run over the corpus and the simulator: confusion matrix, margins, baseline file, a gate in `check-pr.sh`. Start with false merges only, add recall once the baseline is stable.
5. **Labeler** (one PR). `cmd/labeler` web page, case selection from the simulator, JSONL answers, repeats for self-consistency, scenario export. Label the first few hundred cases.
6. **Fitting** (one PR). A script (Go, or Python if the stats libraries earn it) that reads the labels, fits ordered logistic regression with uneven costs and regularization, prints a registry diff, and reports the leftover misses clustered by feature and noise tag. Add active sampling to the labeler once a first fit exists.
7. **Dogfood export** (one PR, any time after phase 1). Debug export of a Promote session as a scenario, plus adapter equivalence tests in `promotealign`.

Later: once claim editing ships ([`identity-claim-review.md`](identity-claim-review.md)), real accepts and rejects become labels in the same shape.

## Open

- Where the labels live: committed under `testdata/` (reviewable, grows the repo) or outside the repo.
- Whether the fit runs in Go or Python. Go keeps one toolchain; Python has ordinal regression ready-made.
- How much of the simulator should model real record types (census columns, obituary phrasing) versus abstract noise.
- Whether pairwise `Rank` tuning ([`matching.md`](../matching.md) registry) uses the same labels, or only the promote scale.
