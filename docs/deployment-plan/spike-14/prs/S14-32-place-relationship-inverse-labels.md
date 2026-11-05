# S14-32 — Place relationship rows read from the far end

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 7c — Evidence graph polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the saved-line row keeps its layout, mark and type; only the label word on the far end changes)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-32-place-relationship-inverse-labels`
**Version bump:** patch — a researcher-visible labelling bug fix in the Mac app ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) Place relationship rows use the inverse label on the far end" (reworded by this plan, see §5); bar outcome 27; dogfood note [`ux/2026-10-15-place-relationship-inverse-label.md`](../ux/2026-10-15-place-relationship-inverse-label.md) (deleted on land)
**Planned:** 2026-11-05 against `main` @ `b6cac73d`
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

The Citation composer lists a saved place-relationship bridge under both of its places. Today both rows use the forward term: with *Calgary is part of Alberta*, Alberta's row reads "Part of Calgary", which is wrong. After this PR the far end reads the relationship from its own side: Alberta's row reads "Contains Calgary", and for *Old Town was succeeded by New Town* New Town's row reads "Succeeded Old Town". This is the same wording the Place conclusion page uses for those two ends.

**Why now.** Batch 7 dogfood fix (note 2026-10-15, bar outcome 27).

**What stays.** The near end (the bridge's `from` place) reads as today: "Part of Alberta", "Succeeded by New Town". The reading sentence, the end titles in the authoring box, the Evidence graph cards, the Place page, and person relationship rows are unchanged. Place relationship terms stay product-only with no vocabulary `Inverse`.

## 2. User stories

**US-1.** As a researcher, when I look at the saved lines under a place in the Citation composer, each place-relationship row reads true from that place's side, so I can check the reading without working out the direction.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Saved bridge `part_of`, from Calgary to Alberta; Calgary's group | Row label "Part of", value "Alberta" → body "Part of Alberta" (unchanged) |
| AC-1.2 | Same bridge; Alberta's group | Row label "Contains", value "Calgary" → body "Contains Calgary" |
| AC-1.3 | Saved bridge `succeeded_by`, from Old Town to New Town; Old Town's group | Body "Succeeded by New Town" (unchanged) |
| AC-1.4 | Same bridge; New Town's group | Body "Succeeded Old Town" |
| AC-1.5 | A researcher-origin term keyed `part_of` labelled "Inside of" (origin `user`), from a to b | Both rows read "Inside of" + the other place: a non-product term keeps its own label on both ends |
| AC-1.6 | A place-relationship bridge whose term cannot be found (empty `roleTermID`) | Both rows read "Place relationship" + the other place (unchanged) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `listedLine`'s `placeRelationship` case picks the far-end label when `viewing == to.id` for the two product terms | STEP-2, AC-1.1 – AC-1.6 |
| Two L10n keys for the far-end labels, in `L10n.CitationComposer` beside `atWord` / `siteOfWord`, with String Catalog entries | STEP-1, AC-1.2, AC-1.4 |
| Swift test over `CitationComposerAuthoring.groups` for both ends of both terms, a researcher's term, and a missing term | STEP-3, T-1 – T-3 |
| Map: working rule, bar outcome 27 and S14-32 scope entry say "Succeeded" for the far end of `succeeded_by` (user, 2026-11-05) | STEP-4 |

| Out | Home |
| --- | --- |
| Inverse terms in the seeded `place_relationship_type` vocabulary | not planned (working rule: copy from L10n) |
| Person relationship inverses | not planned — already handled by `inverseTerm` |
| Localizing the near-end label (`term.label` from the catalog) through `PropertyTermDisplay` | not planned — the relationship and participation rows read `term.label` the same way; one change for every kind if it is ever wanted |
| End titles in the authoring box (`endTitles`) and the reading sentence (`readingSentence`) | not planned — they already word each end ("Part" / "Whole", "Earlier" / "Later") |
| Evidence graph bridge cards, `ConnectionSentenceRegistry` | not planned — they word the whole sentence, not one end |
| Changing the Place page's "Succeeded" / "Succeeded by" wording | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Saved-line grouping | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` `CitationComposerAuthoring.groups(observations:connections:vocabulary:nameOrders:)` | For each saved connection, adds one `CitationSavedLine` per end (`fromSubjectID`, `toSubjectID`), calling `listedLine(…, viewing: subjectID)` for `label` and `valueText`; `body` joins them |
| Per-end label | same file, `private static func listedLine(from:to:rule:termID:termProperty:vocabulary:viewing:)` | `relationship` flips through `inverseTerm` when `viewing == to.id`; `location` picks `atWord` / `siteOfWord` by end; **`placeRelationship` returns `term?.label ?? bridgeKindPlace` on both ends** — the bug |
| Person inverse | same file, `private static func inverseTerm(of:property:vocabulary:)` | Looks up `SeededRelationshipTypeTerm.inverse` (person terms only); has no place keys. Not touched |
| Place term keys | `macos/App/Platform/ConclusionVocabulary.swift` `SeededPlaceRelationshipTypeTerm.partOf` (`"part_of"`), `.succeededBy` (`"succeeded_by"`) | Constants used by `endTitles`, `readingSentence`, `CitationBridgeSentence.displayLabel` |
| Product key | `macos/App/Platform/GenealogyStore.swift` `CatalogPropertyTerm.productKey` | `key` when origin is product, else nil — a researcher's term keyed `part_of` gets nil |
| Direction | `CitationComposerAuthoring.readingSentence`; `core/database/subjectvocab/registry.go` seed rows for `place_relationship_type` | `part_of`: `isPartOf(from, to)`, seed description "From is part of to" → from is the part, to the whole. `succeeded_by`: `wasSucceededBy(from, to)`, seed description "From was succeeded by to" → from is earlier, to later. Neither seed row has an `Inverse` |
| Line-label copy | `macos/App/Platform/L10n.swift` `L10n.CitationComposer.atWord` ("At"), `.siteOfWord` ("Site of"), `.bridgeKindPlace` ("Place relationship") | Per-end location labels; precedent for the new keys |
| Place page wording | `macos/App/Features/Conclusions/PlaceDetailContent.swift` `relationshipSections(summary:)`; `L10n.Conclusions.placeContains` ("Contains", section heading), `.placeSucceeded` ("Succeeded", role on the later place's row) | The far-end wording this PR matches |
| Same-place link | `CitationComposerAuthoring.resolvePair` | Refuses `endA == endB` (`sameSubjectTitle`), so `from.id == to.id` does not reach `groups` |
| Tests | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `savedBridgeIsListedUnderBothEnds`, `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel`; helper `termVocabulary(propertyKey:terms:endTypeKey:)` | Participation bridge listed under both ends; no test covers a place-relationship row's label |

**Precedent.** `4745bcb7` added `atWord` / `siteOfWord` and the `location` case of `listedLine`: a per-end label word chosen by `viewing`, copy in `L10n.CitationComposer`. Copy that shape. `inverseTerm` is the other per-end precedent, but it needs a vocabulary row to look up and place terms have none, so do not copy it.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The far end of `succeeded_by` reads **"Succeeded"** ("Succeeded Old Town") | User, 2026-11-05: match the Place page's `placeSucceeded` | "Preceded by" (the map's first wording and the dogfood note) — the user chose consistency with the Place page |
| D-2 | The far end of `part_of` reads **"Contains"** | Matches the Place page's `placeContains` section and the working rule | — |
| D-3 | Two new keys in `L10n.CitationComposer`: `containsWord` (`citationComposer.containsWord`, "Contains") and `succeededWord` (`citationComposer.succeededWord`, "Succeeded"), placed right after `siteOfWord` | The composer's line labels live in its namespace (`atWord`, `siteOfWord`); a translator sees a line-label comment, not a section heading | Reusing `L10n.Conclusions.placeContains` / `placeSucceeded` — their comments describe a Place page heading and role, and a later change to the Place page would silently change the composer |
| D-4 | A private static `placeFarEndLabel(_ term: CatalogPropertyTerm?) -> String?` *(new)* in `CitationComposerAuthoring`, switching on `term?.productKey`: `partOf` → `containsWord`, `succeededBy` → `succeededWord`, anything else → nil | One switch beside `listedLine`; a third product place term adds one case | An entry in `SeededRelationshipTypeTerm.inverse` — that map returns a term key looked up in vocabulary, and no inverse place term row exists; a place map in `ConclusionVocabulary.swift` — it holds keys, not copy |
| D-5 | `listedLine`'s `placeRelationship` case uses `placeFarEndLabel(term)` only when `viewing == to.id`; otherwise, and when it returns nil, it keeps `term?.label ?? bridgeKindPlace` | `from` is the part / earlier place (see §4 Direction), so its row already reads right; a researcher's term (`productKey` nil) has no known inverse and keeps its label on both ends | Lowercasing or re-wording the near end — Out |
| D-6 | Test at the `CitationComposerAuthoring.groups` level with a hand-built `CitationComposerVocabulary` (extend the `termVocabulary` helper's shape), not through `CitationComposerModel` + FakeStore | `groups` is a pure static function; no store call is involved | A model test — needs place types, properties and terms seeded in `makeStore()` for no extra coverage |

**Rules this plan bends.** Working rule "(Batch 7) Place relationship rows use the inverse label on the far end" names "preceded by" for `succeeded_by`. The rule also says "matching the Place conclusion page", which reads "Succeeded". User answer 2026-11-05: "Succeeded". STEP-4 rewords the rule, bar outcome 27 and the S14-32 scope entry.

## 6. Design

### Codebase fit

Searched: `inverse`, `SeededPlaceRelationshipTypeTerm`, `Contains`, `Succeeded`, `atWord`, `siteOfWord`, `viewing` in `macos/App`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CitationComposerAuthoring.placeFarEndLabel(_:)` *(new)* | func | `CitationComposerAuthoring.inverseTerm(of:property:vocabulary:)` | `inverseTerm` resolves a vocabulary row; place terms have no inverse row (product-only, no `Inverse` in `core/database/subjectvocab/registry.go`) |
| `L10n.CitationComposer.containsWord` *(new)* | key | `L10n.Conclusions.placeContains` | D-3 |
| `L10n.CitationComposer.succeededWord` *(new)* | key | `L10n.Conclusions.placeSucceeded` | D-3 |
| `placeRelationshipRowsReadFromEachEnd` *(new)* | test | `savedBridgeIsListedUnderBothEnds` | That test drives the model for a participation bridge; this one checks labels for a place bridge (D-6) |

**Builds on.** `listedLine`'s per-end switch (`location` case, `4745bcb7`); `SeededPlaceRelationshipTypeTerm`; `CatalogPropertyTerm.productKey`; the `termVocabulary` test helper.
**Deletes.** None — nothing becomes dead.
**Next adds.** A new product place term (for example a "merged into" term) adds one constant in `SeededPlaceRelationshipTypeTerm`, one case in `placeFarEndLabel`, and one key.

### Reach

None — no new path into existing code, no new state reachable.

### Scale

None — no write, list or loop changes; the label is chosen per line already built.

### Edges

- **Assumptions.** `from` is the part / earlier place. Held by `readingSentence` (`isPartOf(from, to)`, `wasSucceededBy(from, to)`) and the Place page; T-1 and T-2 pin both ends.
- **Neighbors.** The near end of each product term (unchanged, T-1/T-2); a researcher's term keyed like a product term (no flip, T-3); a missing term (fallback, T-3); person relationship rows (unchanged, covered by existing tests).
- **Removed inputs.** None — a label choice; nothing is decided, excluded or deleted.
- **Real data.** `place_relationship_type` seeds exactly `part_of` and `succeeded_by` (`SeededPlaceRelationshipTypeTerm`); both are cased.
- **Untrusted size.** None — no list is read beyond what `groups` already walks.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Saved-line label | existing `CitationSavedLine.label` | `CitationComposerAuthoring.listedLine` | Only the string changes |

States: unchanged. Accessibility: `body` (read by VoiceOver) is built from the same `label`, so it reads "Contains Calgary" too.

### Strings

L10n keys *(new)*, in `L10n.CitationComposer` after `siteOfWord`, with matching `Localizable.xcstrings` entries (`extractionState: manual`, en `translated`):

- `containsWord` → `citationComposer.containsWord` = "Contains"; comment "Place-relationship line label on the whole place, read from its side of part of".
- `succeededWord` → `citationComposer.succeededWord` = "Succeeded"; comment "Place-relationship line label on the later place, read from its side of succeeded by".

Removed: none.

## 7. Steps

**STEP-1 — Add the two line-label keys**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: add `containsWord` then `succeededWord` as `LocalizedStringResource` immediately after `siteOfWord`, as §6 Strings says. In the catalog, insert `citationComposer.containsWord` then `citationComposer.succeededWord` immediately after the `citationComposer.siteOfWord` entry, in its exact shape (`comment`, `extractionState: manual`, `en` `stringUnit` `state: translated`).
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-2 — Word the far end**
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`.
Do: add `private static func placeFarEndLabel(_ term: CatalogPropertyTerm?) -> String?` (D-4) beside `inverseTerm`. In `listedLine`, `case EvidenceBridgeKind.placeRelationship.rawValue:` becomes: when `viewing == to.id` and `placeFarEndLabel(term)` is non-nil, use it; otherwise `term?.label ?? L10n.string(L10n.CitationComposer.bridgeKindPlace)` (D-5). Value stays `other.label`.
Check: the app target builds (`xcodebuild build` through `scripts/check-pr.sh`, or Xcode) with no new warning.

**STEP-3 — Test both ends**
Files: `macos/ProvenenciaTests/CitationComposerModelTests.swift`. Skill: `add-swift-test`.
Do: add T-1 – T-3 (§8) as methods of `CitationComposerModelTests` in the "term keys honour the term's origin" MARK section (the helper is a private instance method). Build one vocabulary per case with `termVocabulary(propertyKey: "place_relationship_type", terms: …, endTypeKey: "place")`: T-1 `[(key: "part_of", label: "Part of", origin: CatalogOrigin.provenencia)]`, T-2 `[(key: "succeeded_by", label: "Succeeded by", origin: CatalogOrigin.provenencia)]`, T-3 `[(key: "part_of", label: "Inside of", origin: CatalogOrigin.user)]`. The helper names term ids `term-<origin>-<key>` (`term-provenencia-part_of`, `term-user-part_of`); then overwrite `vocabulary.subjectsByID["a"]` and `["b"]` with `CitationComposerModel.GraphSubjectOption(id:label:ref:typeID:typeKey:)` values whose labels are "Calgary" / "Alberta" (T-1) and "Old Town" / "New Town" (T-2), `typeKey: "place"` (the helper uses the id as the label). Pass one saved `ConnectionRow` to `CitationComposerAuthoring.groups(observations: [], connections: […], vocabulary:)`, with arguments in declaration order: `id: UUID()`, `isPending: false`, `fromSubjectID: "a"`, `toSubjectID: "b"`, `fromLabel`/`toLabel` the place labels, `bridgeTypeKey: "place_relationship"`, `bridgeID: "bridge-place"`, `sentence: ""`, `termProperty: vocabulary.propertiesByID["p-place_relationship_type"]`, `roleTermID:` the term id (`""` for AC-1.6), `termTouched: false`, `isSaving: false`. Find each group by `subjectID` (`"a"`, `"b"`), never by index: `groups` sorts by type, then line `sortKey`, then name. Each group holds one line. Compare the far-end labels with `L10n.string(L10n.CitationComposer.containsWord)` / `succeededWord` and the fallback with `L10n.string(L10n.CitationComposer.bridgeKindPlace)`, not English literals; the near-end labels are the term labels given above.
Check: `xcodebuild test` for `CitationComposerModelTests` passes (the command `scripts/check-pr.sh` runs). On a host without Xcode (`scripts/check-pr.sh` skips `xcode-test` off macOS), run the Python checks only, report the Swift build and T-1 – T-3 as **not run**, and name them in the PR description so they are run on a Mac before merge.

**STEP-4 — Version bump**
Files: `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`. Rule: `.cursor/rules/versioning.mdc`.
Do: raise PATCH by one from what `main` holds when you branch (0.3.3 → 0.3.4 at the planning base). If `main` moves past your number before merge (S14-31 is also a patch), rebase and take the next PATCH again; all three places always agree.
Check: `rg -n "0\.3\.3" VERSION core/version.go macos/Provenencia.xcodeproj/project.pbxproj` prints nothing after the bump (with the old number substituted if `main` moved).

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-14/deployment-plan.md` (working rule "(Batch 7) Place relationship rows use the inverse label on the far end", bar outcome 27, S14-32 scope entry and map row), `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/ux/2026-10-15-place-relationship-inverse-label.md` (delete), "Where the notes went" row, this plan.
Do: per the Definition of done. The working rule, bar 27 and scope entry are already reworded to "Succeeded" by the planning commit; check they still say so.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-32-place-relationship-inverse-labels.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Read a place relationship from the far place's side in the composer's saved lines.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `placeRelationshipRowsReadFromEachEnd` *(new)* | product `part_of`, Calgary → Alberta: Calgary's line label "Part of" (the term's label), value "Alberta"; Alberta's line label `containsWord`, value "Calgary", body "Contains Calgary" | AC-1.1, AC-1.2, D-2, D-5 |
| T-2 | Swift | same test | product `succeeded_by`, Old Town → New Town: Old Town's label "Succeeded by", value "New Town"; New Town's label `succeededWord`, value "Old Town" | AC-1.3, AC-1.4, D-1 |
| T-3 | Swift | `aResearchersPlaceTermKeepsItsLabelOnBothEnds` *(new)* | user-origin `part_of` "Inside of": both lines labelled "Inside of"; and a bridge with `roleTermID: ""`: both lines labelled `bridgeKindPlace` | AC-1.5, AC-1.6, D-5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerModelTests` (`savedBridgeIsListedUnderBothEnds`, `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel`), `scripts/check-localizable-xcstrings.py`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: no new component)
- [ ] Version bump: patch (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; "Where the notes went" row says the note was deleted on land; completed.md: one short paragraph of what shipped
- [ ] Dogfood note ux/2026-10-15-place-relationship-inverse-label.md deleted
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source and an Evidence graph holding two places, Calgary and Alberta.
2. In the Citation composer, connect Calgary → Alberta with Part of and save. Calgary's group reads "Part of Alberta"; Alberta's group reads "Contains Calgary".
3. Add Old Town and New Town; connect Old Town → New Town with Succeeded by and save. Old Town reads "Succeeded by New Town"; New Town reads "Succeeded Old Town".
4. Open Alberta's and New Town's Place pages after Promote: their "Contains" and "Succeeded" rows use the same words.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The near end's `term.label` is a catalog string ("Part of") while the far end comes from L10n | Accepted for English; Out line names the shared fix |
| A translator renders "Succeeded" without context | The catalog comment names the later place and the term |

**Stop and report if:**

- `listedLine`'s `placeRelationship` case no longer returns `term?.label` on both ends (someone fixed it first).
- `readingSentence` no longer treats `from` as the part / earlier place, or Go writes `part_of` with the whole as `from` (then D-5 flips the wrong end).
- `place_relationship_type` gains a product term beyond `part_of` and `succeeded_by` at the base you build on.
- `CitationComposerAuthoring.groups` is no longer callable from tests (`private` or renamed).

## 12. Open questions

None.

## 13. PR

**Title:** S14-32 — Place relationship rows read from the far end
**Body:** objective in one sentence; what shipped against US-1; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-32-place-relationship-inverse-labels.md`.
