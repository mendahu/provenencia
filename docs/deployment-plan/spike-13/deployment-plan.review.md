# Review log — Spike 13 deployment plan

Map review per [`review-plan`](../../../.agents/skills/review-plan/SKILL.md#map-review-plan-spike). Round 1, 2026-10-26: one fresh maintainability reviewer (document, repo, `maintainability.md`; no planning conversation). Verdict: **Fits with changes**. Evidence spot-checked by the writer before answering.

| ID | Finding | Answer |
| --- | --- | --- |
| K-1 | S13-02's `PersonSummary` duplicates `PersonHeader`, whose doc already names "later tree nodes"; gender treatment would get two inputs | **Fixed.** S13-02 is "Gender in the person header"; rule "One person header for every tile"; `PersonGenderTreatment` reads only the header. Re-key by id noted in S13-04 / S13-05 |
| K-2 | S13-04 would be the third copy of the direct + implied merge (`promotealign.edgeSeenKey` in `canon.go`, `layer.go`) | **Fixed; user agreed 2026-10-26 and moved it to Spike 11 as S11-51** (right after S11-39). Was S13-19 "One person's links" in `canonlinks`, after S11-39; S13-04 depends on it; rule "One person's links, direct and implied, from one function"; updates `add-match-rule` |
| K-3 | Deleting the person on last unpromote would skip `deleteimpact`; a boolean flag skips the Impact preflight pattern | **Fixed.** S13-08 deletes through `deleteimpact` with an unpromote Impact preflight; `add-catalog-delete`, `add-audit-scope` named. Still blocked on open question 1 |
| K-4 | `GetIdentityClaim` would be a second claim evidence builder beside `promotealign.Evidence` | **Fixed.** S13-07 adds a stored-claim mode to `Evidence`; rule renamed "reuses the claim evidence query" |
| K-5 | S13-05 rebuilds `EventHeader` / `sortEvents` and walks links without a named hop | **Fixed.** Any-role hop in `core/hops`; `EventHeader` + role + {role, `PersonHeader`}; `sortEvents`; `EventSubject` fold logged for the plan |
| K-6 | S13-01 joins what `graphcache` holds; a third accepted-members filter | **Fixed.** S13-01 reads `graphcache`, adds one accepted-members helper and moves both private copies onto it; Subject card uses existing `CatalogSubject` / Observation messages |
| K-7 | Three per-person keys each need hand-written invalidation; relatives' changes touch other ids | **Fixed.** Rule "One invalidation rule for person-section keys"; S13-01 builds the per-handle descriptor entry (its own `CatalogQueryFollows`), S13-04 / S13-05 add entries; family and timeline also refresh on promote effects |
| K-8 | A Person-only page type forks `ConclusionDetailPage`, which Event and Place will want back | **Fixed.** S13-09 is "Sections slot on the Conclusion page"; rule "One sectioned Conclusion page, not a Person fork" |
| K-9 | Keeping the Why accordion on Event and Place leaves two provenance mechanisms | **To user.** Open question 3 now carries the recommendation: replace everywhere and delete the accordion |
| K-10 | S13-06 shaped for birth/death only; makes "its page owns the Why" comments false | **Fixed.** S13-06 "Derived rows carry their records", keyed by owning entity, off `LifeFacts`; comments updated in the PR |
| K-11 | S13-03 changes a pattern `add-match-rule` teaches ("siblings changes `Derive`") | **Fixed.** S13-03 updates `add-match-rule`; added to the docs-and-skills rule |
| K-12 | The link-reading move to its own pure package (kinship-view, S11-38 M-3) is not placed | **Declined.** The page reads kinship only in Go through `canonlinks`, which already imports `segmentcompare`; nothing new imports `core/kinship` from a lighter layer, so the weight is already paid. Stays in [`ideas/kinship-view.md`](../../ideas/kinship-view.md) |
| K-13 | Open question 2 ignores `ideas/source-provenance-date.md` | **Fixed.** S13-01 ships without a date; OQ2 links the idea and names the two options |
| K-14 | Owner package for the three new reads not named | **Fixed.** `core/database/persondetail` named in the rule and scopes |

No blocker remains open and no fix reshaped the spike's slices, so no round 2. K-2 and K-9 were answered by the user on 2026-10-26: S13-19 became S11-51; the Why accordion stays on Event and Place until they adopt sections.

## Slice 0 — Removing Observation pins (2026-10-29)

Map review of slice 0 (S13-23 – S13-25) and its re-plans in Spikes 11, 12 and 14. Round 1: one fresh maintainability reviewer (the maps, the repo, `maintainability.md`; no planning conversation). Verdict: **Fits with changes**. It confirmed the seam: with no pin ids and `default_pins` false, `promotealign.ResolvePins` skips the evidence read, so S13-23 can go first.

| ID | Finding | Answer |
| --- | --- | --- |
| K-15 | S13-23 left Swift that names the fields and Method S13-24 removes (batch-row fields, GoStore mappers, `CoreMethod`, the `OnboardingModelTests` conformer, pbxproj entries) | **Fixed.** S13-23 deletes them all; Check `rg` finds hits only in `Generated/` |
| K-16 | The sheet's wire enums, share wording, `recordedDirect` and their Go parity rows go dead; S13-23 can't remove them and stay Swift only | **Fixed.** S13-23 deletes them and edits `swift_vocabulary_test.go` (its only Go edit); `add-seeded-vocabulary` line 80 joins the docs rule |
| K-17 | S12-32 and S12-31 touch the same Done path and project file; the intro said Spike 12 was closed | **Fixed.** S12-32 depends on S13-24 in both maps; S12-31 / S13-23 rebase rule; intro rewritten |
| K-18 | The claim confidence grades read loses its only consumer | **Fixed (writer's call, below).** S13-23 drops the Swift key, warm, query and store method; S13-24 reserves `ListClaimConfidenceGrades`. The table and seed stay |
| K-19 | S13-24 missed `BatchRow` claim fields, `Reader.Relationships`, `Recording.Shown`, `Vocabulary.BindingOrder` and tests; `acceptedClaims` has no other caller | **Fixed.** Listed in S13-24; `acceptedClaims` deleted; S13-01 no longer moves it |
| K-20 | Skills, rules and docs made false but unassigned | **Fixed.** The docs rule assigns each to S13-23, S13-24 or S13-25 and each DoD `rg`s the deleted names |
| K-21 | S13-25 left the Swift pin facet (FakeStore, `PVDeleteImpact` case and key, tests) | **Fixed.** Listed in S13-25 with an `rg` Check |
| K-22 | "Audit history" rule named readers that don't exist | **Fixed.** Rule says nothing reads audit rows by type; Check is that the rows survive |
| K-23 | `PromoteReadout` comments and the row's button label assume the sheet | **Fixed.** In S13-23 |
| K-24 | S14-02's re-plan was unordered against S13-23 and its rule named `isLikeness` | **Fixed.** S14-02 depends on S13-23 in both maps; rule reworded |
| K-25 | Descoped Spike 11 entries kept "S12-… waits on this PR"; duplicate slice headings in the scope section | **Fixed** (past tense) / **Logged** (the headings predate slice 0) |

No blocker was raised and no fix reshaped the slice, so no round 2.

### Writer's own calls

- **The claim confidence read goes with Promote's grade picker** (K-18): `ListClaimConfidenceGrades` is reserved, not kept for a future editor; it is a few lines to bring back when open question 9 or claim editing needs it.
- **No brief for S13-23.** The assessment column keeps today's look and loses only its button and chevron.
- **S13-08 waits on S13-25** because both change `deleteimpact.ReleaseClaim`.
