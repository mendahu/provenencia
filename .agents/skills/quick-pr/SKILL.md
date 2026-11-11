---
name: quick-pr
description: >-
  Ships one small Provenencia fix (a bug, a wrong value, a missing guard, a
  copy correction) as a light PR without a PR plan, a plan review, an
  implementer agent, or a ship log: check the change fits the quick limits,
  write the failing test, make the smallest fix, pass scripts/check-pr.sh, then
  one fresh reviewer through the quick lens, one round. Escalates to plan-pr
  the moment the change outgrows the limits. Use when the user asks for a
  quick fix, quick PR, light PR, small correction, or hotfix, or when a PR map
  row's Plan column says quick-pr. Not for features, schema, FFI, or chrome
  changes; those go through plan-pr and ship-pr.
---

# Ship a quick PR

The full loop ([`plan-pr`](../plan-pr/SKILL.md) →
[`review-plan`](../review-plan/SKILL.md) → [`ship-pr`](../ship-pr/SKILL.md))
exists so that a smaller model can build a large change without making a
decision. A ten-line fix has no decisions to hand over. Writing a 300-line plan,
reviewing it twice, and sending four reviewers over the result costs more than
the fix is worth.

This skill keeps the parts that catch bugs and drops the parts that hand work
between agents:

| Kept | Dropped |
| --- | --- |
| A failing test before the fix | The PR plan and its review log |
| `scripts/check-pr.sh`, unchanged | The implementer agent: this session writes the fix |
| One fresh reviewer that never saw this session's reasoning | Lens routing, verifiers, and the second review round |
| CI on the ready PR | The ship log: the PR body is the record |

This session writes the code **and** decides what happens to each finding. The
reviewer is the only independent check, so it gets the diff and the PR body and
nothing else.

## Fits a quick PR

Every line must be true. Check them before you write code, and again before
you push.

```
- [ ] One behavior, stated in one sentence a researcher would recognize
      ("Copy Source keeps the date's qualifier"), with a reproducible wrong result
- [ ] At most about 40 changed lines of non-test code, in at most 3 non-test files
- [ ] No migration, no catalog table or column, no catalogmodel change
- [ ] No new or changed FFI method, proto message, or engine.proto field
- [ ] No new package, registry, seeded vocabulary, searchable kind, or workspace place
- [ ] No chrome change: every view looks the same before and after, except
      copy (a corrected string through add-localized-string is fine).
      A view that looks different needs a brief; the user can waive it, and the
      PR body says so
- [ ] No product call open, and no plan Decision or map working rule reversed
- [ ] The fix reuses existing code. It adds no private copy of a helper that already exists
```

When one line is false, it is not a quick PR. Tell the user which line and
recommend [`plan-pr`](../plan-pr/SKILL.md): a one-off PR goes to the open
even-numbered spike as a new batch ([odd and even
spikes](../../../docs/deployment-plan/README.md#odd-and-even-spikes)). Keep the
failing test you wrote; it goes into that plan's Current state.

Security-sensitive or hot paths do not disqualify a fix. They make the
reviewer read more (step 5).

## Checklist

```
- [ ] 1. Fit: the list above; read the code the fix lands in and its callers
- [ ] 2. Branch from an up-to-date main
- [ ] 3. Red: a test that fails for the reported reason
- [ ] 4. Fix: the smallest change at the layer that owns the behavior; green; gate
- [ ] 5. Review: one pr-reviewer, quick lens, fresh
- [ ] 6. Triage and fix in place; gate again
- [ ] 7. Finish: PR body, ready, CI, report
```

### 1. Fit

Find the code: the function that computes the wrong result, its callers
(`rg -n "<FuncName>\("`), and its tests. Read them. Find the **root cause**, not
the line where the symptom shows: a fix in one caller that leaves the others
wrong is a patch, and the reviewer will rate it a major.

Load the skill for the layer you touch when the table in
[`plan-pr`](../plan-pr/SKILL.md) §4 names one. A quick PR follows the same
rules a planned one does (`add-localized-string` for copy, `go-tests.mdc` for
table-driven Go tests, `add-swift-test` for Swift tests).

### 2. Branch

`git fetch origin main`, then `git switch -c quick/<slug> origin/main`. When the
session names a branch, use that one instead. If a map row named this PR, the
slug is the row's (`s16-41-<slug>`).

### 3. Red

Write the test that shows the reported wrong result, and run it. It **fails**,
for the reason the report gives. Keep the output: the PR body quotes it.

- Go: `go test -tags fts5 ./core/<pkg>/ -run '<TestName>'`.
- Swift: the test under `macos/ProvenenciaTests`, run through `xcodebuild test`
  with `-only-testing:`.

No test is possible only for a copy-only change, or a fix in code that has no
test seam today. Say which in the PR body, and give the manual check instead.
"Hard to test" is not "no seam": if the behavior has a unit-test layer, use it.

If the test passes before any fix, the bug is not where you thought. Go back to
step 1. Do not fix code you cannot show is wrong.

### 4. Fix

Make the smallest change, at the layer that owns the behavior, that turns the
test green. Then run the package's tests, then the gate:

```bash
scripts/check-pr.sh
```

`check-pr: PASS` is the bar, as in [`implement-pr`](../implement-pr/SKILL.md).
Never skip, disable, or weaken a test or a check.

**Off a Mac** the gate's `xcode-test` check fails by design. Run the checks that
can run (`gofmt`, `go vet -tags fts5 ./api/... ./core/...`,
`go test -tags fts5 ./api/... ./core/...`, `python3 scripts/check-guidance-names.py`,
`python3 scripts/check-localizable-xcstrings.py`) and let CI on the ready PR
run the Xcode part (step 7). Say so in the PR body. A fix under `macos/` that
cannot be built here needs the user to build it before merge; say that too.

**Version.** A fix that ships product code bumps PATCH in `VERSION`,
`core/version.go` and `MARKETING_VERSION` together ([`docs/versioning.md`](../../../docs/versioning.md)).
A docs-only or test-only fix does not bump.

**Sweep.** When you renamed or deleted a symbol, `rg` its old name across
`.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift`, as `implement-pr`
step 6 does.

Commit per behavior: usually one commit for the test and the fix together, one
for the version bump. The subject is one full imperative sentence ending in a
period that says what changed for the researcher, like the rest of `git log`.

Re-check the fit list against `git diff --stat origin/main...HEAD`. Over the
limits now → stop and escalate (above).

Push and open a **draft** PR (`gh pr create --draft --base main`, or the GitHub
tools where `gh` is not available) with the body below, so the reviewer reads
the same text the user will.

### 5. Review

Dispatch **one** `pr-reviewer`:

```text
Review a Provenencia quick PR through the Quick lens. Follow
.agents/skills/review-pr/SKILL.md and .agents/skills/review-pr/lenses/quick.md.
PR body: <path to the body file, or the PR number>. Base: <merge-base sha>.
Head: <head sha>. Read-only. Return findings in the review-pr format, then the verdict.
```

Do not send it this session's reasoning, the user's conversation, or your view
of what might be wrong. The PR body is its only context beyond the code.

The quick lens covers correctness, regression, test proof, scope, and reuse,
and pulls in the security, performance and UX checks that the diff's paths call
for (the routing table in [`ship-pr`](../ship-pr/SKILL.md) §3). There is one
reviewer even when several of those apply.

The user can ask for the full reviewer set on any quick PR. Then route and
dispatch as `ship-pr` §3 does, with the PR body standing in for the plan.

### 6. Triage

Each finding gets one decision. Write the table in the PR body.

| Decision | When |
| --- | --- |
| **Fix now** | A blocker or major in code this PR adds or changes, confirmed by reading the code. A minor in a line this PR already touches. A finding that the fix adds a private copy of existing code: switch to the existing code |
| **Follow-up** | A problem in code this PR does not add, change, or make worse. A cleanup of copies this PR does not touch. Listed in the PR body; ask the user where each goes before you finish |
| **Rejected** | Refuted by the code or a test. One line of evidence |
| **To user** | A product call, or a finding that contradicts a working rule but looks right |
| **Escalate** | The right fix breaks the fit list: a shared helper across packages, a migration, an FFI change, a second view. Stop; take the finding and the failing test to the user and recommend `plan-pr` |

Nits are listed in one line each and not fixed.

Fix every **Fix now** item here, test first when it is a behavior, then run the
gate again and push. There is **no second review round**. When a fixed item was a
**blocker**, dispatch one `pr-reviewer` to verify that item alone:

```text
Try to disprove that this finding is fixed, against the code at <head sha>. Read-only.
<finding>
Answer: Fixed / Not fixed / Unclear, with the evidence.
```

Not fixed after one more try → stop and take it to the user.

### 7. Finish

1. `git fetch origin`; if `origin/main` moved, merge it and run the gate again.
2. Update the PR body (template below). Push.
3. If a map row named this PR, set its Status to `Landed #N` and add the
   `completed.md` paragraph in one commit, "Record S#-NN as landed in pull
   request N.", as `implement-pr` step 8 does. An ad-hoc quick PR (no map row)
   changes no planning docs.
4. Mark the PR ready (`gh pr ready <n>`), so CI runs, and watch the checks. A red
   check is fixed once here; red again → back to the user.
5. Report (below). The user reviews and merges. Never merge.

## PR body

```markdown
**Quick PR** ([`quick-pr`](.agents/skills/quick-pr/SKILL.md))

**Problem.** What the researcher saw, with the input and the wrong result.
**Cause.** `path/to/file.go:42` `FuncName`: why it computed that.
**Fix.** What changed, at which layer, and the existing code it reuses.
**Test.** `TestName/case`: failed before the fix (quote the line), passes after.
Or: no test, and why; the manual check.
**Gate.** check-pr: PASS (paste the PASS/FAIL lines). Off a Mac: which checks ran; CI runs Xcode.
**Version.** 0.9.0 → 0.9.1, or none and why.
**Not in this PR.** What the fix left alone on purpose.

### Review (quick lens)
| ID | Severity | Title | Decision | Evidence |
| --- | --- | --- | --- | --- |
Nits: …
**Follow-ups:** one line each, with a suggested home.
```

## Report to the user

```markdown
**Quick PR ready for review:** <PR link>
- Fixed: one sentence.
- Test: the test name, red then green.
- Review: findings fixed / follow-up / rejected; verdict.
- Needs you: To user items and follow-ups with a suggested home. "Nothing."
- Build: on your Mac, when the fix touches `macos/` or the gate ran off a Mac.
```

## Never

- Start a quick PR for a change that fails the fit list, or keep going once it does.
- Fix without a red test, except the two cases step 3 names.
- Send the reviewer this session's reasoning.
- Run a second review round, mark the PR ready with a blocker open, or merge.
- Choose a follow-up's home without asking the user, or put it in `docs/dogfood/`.
