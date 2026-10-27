# Review — S12-20 Docs and skills match the code

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **M-7: enforce README rows with a check?** The STEP-4 loop is the whole check: every `Components/*` and `Recipes/*` folder under `macos/App/DesignSystem/` needs a README row. The map's S12-20 "Not this PR" excludes a `check-pr.sh` check, so this plan keeps it Out and adds only a checklist line. **Recommendation:** add the loop to `scripts/check-guidance-names.py` in this PR. It is about five lines; the hand-kept rule is what let six rows go missing; and the PR stops being strictly docs-only by one script edit. If you'd rather keep S12-20 docs-only, the check goes into the map as a line in S12-31, which also changes how kit folders are registered.

## Writer's own calls

- `docs/matching-model.md` line 3's "until then" sentence is corrected here, though the audit names only `matching.md` (D-4). It makes the same false claim.
- `docs/ideas/README.md` line 30's description of `shared-list-helpers.md` is rewritten (D-7). It names no section the file has.
- `api/proto/README.md`: the stale sentence is deleted rather than reworded (D-5 d).
- The audit's `swift-tests.mdc` and `design-system-layers.mdc` bullets are Out, because `22708a81` already fixed them.

## Round 1

### M-1 · major · D-4 adds a second copy of the six-layer table
`docs/matching-model.md` "## The stack" owns all six layers; the `matching.md` table was a copy that drifted.
**Answer:** Fixed. D-4 replaces the `matching.md` table with a link to `matching-model.md#the-stack`. AC-2.2 and Codebase fit follow.

### M-2 · major · The `GPH` fix leaves two stale copies of the reserved-prefix list
`docs/interpretation-layer-data-model.md:542` and `docs/seeded-vocabulary.md:312` list the prefixes without `GPH`.
**Answer:** Fixed. D-8 (new): the rule and both docs name `core/ref` `reservedPrefixes` instead of listing prefixes. AC-1.3 sweeps for a list that ends at `OBS`.

### M-3 · major · `.cursor/rules/catalog-deletes.mdc:12` gate list lacks `last_graph`
**Answer:** Fixed. D-5 (g) points the rule at `extraGateList`. AC-2.4 checks it. The wire enum sketch in `docs/catalog-deletes.md:73` still lists values, because it describes the wire enum.

### M-4 · minor · README does not name `Search` / `Structure` as fields
**Answer:** Fixed. D-1 names `type Effect` in `registry.go` for the field set and the README for what each field moves.

### M-5 · minor · D-6 rows list callers
**Answer:** Fixed. The rows state role and contract only. The FloatingMenuSelection row says not to write another highlight mover.

### M-6 · minor · D-5(e) cites `artifacts.Create` call shape that S12-22 replaces
**Answer:** Fixed. D-5 (e) states the behavior instead.

### M-7 · minor · README-row rule is only a checklist line
**Answer:** To user (above).

### M-8 · minor · `Effect.Source` sweep hits intended examples
`implement-pr` and `conformance.md` use `Effect.Source` as an example.
**Answer:** Fixed. The DoD sweep now searches for the stale text itself (`CitationSource`, the prefix list, the idea path), with this plan and its review log excluded.

### X-1 · blocker · STEP-4 loop cannot pass with D-6's wording
**Answer:** Fixed. **D-6 rows** gives each row verbatim with the `` `Components/<Name>/…` `` path the loop matches.

### X-2 · major · STEP-5 / DoD sweep match the plan itself
**Answer:** Fixed. Each such `rg` excludes `**/S12-20-*`. The archive links read `ideas/archive/…` and do not match.

### X-3 · major · D-7 leaves the archive text and README lines open
**Answer:** Fixed. **D-7 text** gives the archived file, the Archived line and its position, and the whole replacement for line 30. The Leftovers section is a table, per `archive-docs`.

### X-4 · minor · weak or wrap-sensitive checks
**Answer:** Fixed. AC-1.2 keeps the name on one line. AC-2.6 searches for the new sentence.

### X-5 · minor · STEP-6 under-specified
**Answer:** Fixed. STEP-6 says where the `completed.md` paragraph goes and what the map status becomes. This planning commit links the map row.

### X-6 · minor · DoD off macOS
**Answer:** Fixed. The stop condition names the two script checks as the evidence to report.

Also added from the dry run: an Out line saying the S11-16 rows in `matching-model.md`'s "Replaced" table stay.

## Round 2

### M-9 · blocker · D-8 misses `docs/data-model-source-interpretation-conclusion.md:219`
It has a hand-kept inline list (correct today, ending at `GPH`). AC-1.3's sweep could not see it.
**Answer:** Fixed. D-8 rewrites line 219 to point at `reservedPrefixes` and the doc's own table. AC-1.3 sweeps for any inline list (`OBS.[,.]`); the only expected hit is `interpretation-layer-data-model.md:74`, which names the layer's own refs. The three prefix tables are correct and stay. Next adds now says a new prefix is one constant plus a row in each table, with no inline list to find. M-2 holds with this.

### M-10 · blocker · Manual check 2 contradicts D-4
**Answer:** Fixed. §10 step 2 now says there is no layer table, one link to the stack, and only merge hints under "Not built yet". This is the same finding as X-1 below.

### X-1 · major · Manual check 2 contradicts D-4
**Answer:** Fixed (as M-10).

### X-2 · minor · STEP-1 effects check passes before the edit
**Answer:** Fixed. STEP-1 now checks `rg -n 'type Effect'` → one hit in step 6. T-1 no longer claims AC-1.3; T-4 covers it.

### X-3 · minor · wording choices
**Answer:** Fixed. D-0 (new) fixes the link text and relative paths and the one-line rule for code names. D-5 (f) reads "on that graph's memo". D-7 changes both the link text and the target of the four map links. STEP-6 copies the S12-15 `completed.md` shape.

### X-4 · nit · branch and status
**Answer:** Logged. The header's branch is the implementation branch; this planning branch is separate. Status stays Draft until M-7 is answered.
