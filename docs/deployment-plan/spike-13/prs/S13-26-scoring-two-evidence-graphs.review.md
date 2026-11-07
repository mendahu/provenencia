# Review — S13-26 Scoring two Evidence graphs

**Verdict:** Fits with changes / Ready with changes after two rounds; every finding Fixed; M-2 and M-5 went To user and were answered 2026-11-07 (below). Plan status Planned.

Plan: [`S13-26-scoring-two-evidence-graphs.md`](S13-26-scoring-two-evidence-graphs.md)

## Round 1 (2026-11-07)

Fresh reviewers, plan only. Maintainability: **Fits with changes** (2 blockers, 2 majors, 3 minors). Executability: **Ready with changes** (2 blockers, 6 majors, 4 minors). Lint clean before the round.

### Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | D-6 composed `Side` and `propertycompare.Compare` in `recordlinks`, a second copy of `subjectcompare.Compare`'s property step, against `add-match-rule`'s Do-not line | **Fixed.** D-6: `subjectcompare.Rules.Unit` and `CompareProperty` *(new)*; `Compare` calls `CompareProperty`; `recordlinks` counts through them. The "else own values" branch and the missing-meta rule are gone: `Side` returns a key's own values, and a zero meta compares as unsupported, as in `Compare` (§4 Value compare row). STEP-2, T-9, §6 Codebase fit, STEP-8 (a) rewords the skill bullet |
| M-2 | blocker | D-8 drops `CountedOnce`'s only-implied gate while the map's working rule "A matching link counts" names `CountedOnce`; §5 said no rule is bent | **To user** (relabelled in round 2: the refusal does not name the both-recorded case). Plan text: §5 Rules this plan bends names it, with the user's 2026-11-07 refusal ("counting a marriage and the spouse link it implies as two facts"), which does not depend on whether the spouse link is also written down: the planner named `CountedOnce` in the rule before checking that it never skips two recorded links. STEP-8 (e) rewrites the rule and the S13-26 scope entry. Raised to the user in the report |
| M-3 | major | D-11 added a third private copy of the participation and relationship key builders | **Fixed.** D-12: `segmentcomparetest.PartKeys` / `RelKeys`; `kinship_test` `part` / `rel` and `align_test` `partKeys` / `relKeys` move onto them (STEP-1). The other one-off `Keys` calls in tests are an Out line with a reason |
| M-4 | major | `religion` added to a test-local metas copy instead of the seed-pinned `subjectcomparetest.Metas` | **Fixed.** STEP-2 adds it to `Metas`; the subjectvocab seed test covers it |
| M-5 | minor | `graphalign.Bridge` duplicates `kinship.Link`; S13-27 would convert | **To user.** Recommendation: S13-29's scope entry says the shared loader yields `[]kinship.Link` and `graphalign.Bridge` becomes `kinship.Link`. A map change to a preparatory PR, so the user's call. Out line added |
| M-6 | minor | `Config.Links` reused "links" for the section | **Fixed.** D-10: `Config.RecordLinks`, matching `Registry.RecordLinks` and the inventory key |
| M-7 | minor | Helpers returned bare counts though an explanation is the written next add | **Fixed.** D-9: the helpers return the counted units and pairs; `Facts` is their total; Next adds updated |
| Sweep | — | The skill's "starting point for slice 7" bullet goes stale | **Fixed.** STEP-8 (a) |

### Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | D-7 did not say how several matching link pairs for one (key, entity) combine with the shape skip | **Fixed.** D-7: the (key, entity) pair counts once when any corresponding pair of links is not skipped. AC-2.19 and T-2's two-shape case pin a link with two shapes |
| X-2 | blocker | D-4 / D-5 did not say which members form a pair when an entity holds members of several kinds | **Fixed.** D-4: only person members are paired, compares run with kind person, other kinds on the entity are ignored; AC-1.5 |
| X-3 | major | The pivot lookup for `same` was implied | **Fixed.** D-3 and D-7: a pivot's entity is looked up in its own graph's index; a pivot that is no member or has no entity is never the same, so the link counts. AC-2.18 |
| X-4 | major | `NewGraph`'s odd inputs unspecified | **Fixed.** D-3: duplicate ids (first wins), links with no member end, far ends with no entity, empty and zero `Graph`, no copying and no mutation |
| X-5 | major | T-8's generator left to the implementer | **Fixed.** §8 T-8 specifies the pools, ids, page, properties, links, config and shuffle; the builders live in `recordlinks_test.go` (D-11) |
| X-6 | major | T-8 (e) could fail on names | **Fixed.** The name pool is plain names that agree with themselves; a failure is a Stop condition |
| X-7 | major | Test config not named | **Fixed.** D-11: `matchregistry.Default().RecordLinksConfig()`, `StrongAt` edited on a copy; Stop condition for AC-2.2 / AC-2.3 under the shipped value rules |
| X-8 | major | `review-pr/lenses/security.md` and `ship-pr` list layer packages | **Fixed.** STEP-8 (b) |
| X-9 | minor | The §5.1 replacement could not keep the design link it replaced; the earlier sentence still lists conflicts | **Fixed.** STEP-8 (d) replaces the whole decided paragraph with the link written in |
| X-10 | minor | `completed.md`, Landed note and status words need the PR number | **Fixed.** STEP-8 (f): written at ship, with the status words |
| X-11 | minor | STEP-6's check named a command STEP-7 edits; Xcode steps skip on Linux | **Fixed.** STEP-7's check spells out the command; the DoD says Linux skips Xcode |
| X-12 | minor | Sort comparator, `Entity` copy, outcome constants, imports, test file layout, which rows per step | **Fixed.** D-1 imports, D-4 (`bytes.Compare`, `bytes.Clone`), D-6 constants, D-11 file layout, STEP-5 / STEP-6 rows |

### Writer's own calls

- **No missing-meta rule** (M-1): `CompareProperty` reads `metas[unit]` as `Compare` did; a zero meta is unsupported in value compare, so it is unknown.
- **Only `part` / `rel` and `partKeys` / `relKeys` move** onto the shared key builders (M-3); one-off `Keys` calls inside single tests stay.

## Round 2 (2026-11-07)

Fresh reviewers with the plan and this log. Maintainability: **Fits with changes** (1 new blocker). Executability: **Ready with changes** (1 new blocker, the same one; 1 major, 3 minors). Both checked every round-1 answer against the code; all held except M-3's fix.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-8 / X-13 | blocker | D-12's `PartKeys(t, …)` cannot be called from `align_test.go`'s `partKeys` / `relKeys`, which take no `t`, build through `segmentcompare.Key(Link{…})` (not `Keys`), and feed `partLink` / `relLink` at about 25 call sites; §4's row was half wrong | **Fixed.** D-12: the helpers take no `t` and build as `align_test` does; `partKeys` / `relKeys` become one-line calls; no caller's signature changes. T-10 holds them equal to `Keys(Bridge(t, …))` for the seeded bridges. §4 row corrected |
| M-1 note | — | `Rules.Meta`'s doc comment restates the `Side` rule `CompareProperty` now owns | **Fixed.** STEP-2 rewords it |
| M-2 / M-5 | process | Both belong under To user, with a Verdict header | **Fixed.** Relabelled; To user section and Verdict added |
| X-14 | major | T-8's link generation and shuffle point were open | **Fixed.** §8 T-8: participation and relationship counts, same-graph endpoints, association ids, half the rounds sharing events with B, and the shuffle of edges before a new `LinkSet` |
| X-15 | minor | A far entity equal to the line's own entity | **Fixed.** D-7: any far entity counts, the page's and the line's own included |
| X-16 | minor | Helper return types unnamed | **Fixed.** D-9: `[]vocabkey.PropertyKey` and `[]linkFact{key, entity}` |
| X-17 | minor | D-1's import list missed `maps` | **Fixed.** D-1 says "at least" and lists `maps` |

Loop closed at round 2 (the cap). No blocker open.

## To user

1. **M-2 — both-recorded spouse links.** When both records hold a marriage on one entity and also write the spouse link down, the plan counts one fact (`kinship.SameShape` without `CountedOnce`'s gate) and rewrites the working rule "A matching link counts" to say so. Recommendation: one fact. **Answered 2026-11-07: one fact.** §5 Rules this plan bends records it; STEP-8 (e) updates the rule.
2. **M-5 — one link type from the loader.** Recommendation: S13-29's scope entry says the shared loader yields `[]kinship.Link` and `graphalign.Bridge` becomes `kinship.Link`, so S13-27 hands the loader's output to `NewGraph` with no conversion. A map change to a preparatory PR. **Answered 2026-11-07: yes.** S13-29's map row and scope entry say so.
