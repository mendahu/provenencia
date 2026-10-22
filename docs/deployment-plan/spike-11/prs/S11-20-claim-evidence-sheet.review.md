# Review — S11-20 Claim evidence sheet (and brief S11-D2)

**Round 1, 2026-10-22,** on the draft plan (status Blocked) and the brief. Lint: `check-plan-docs.py` clean on both before and after. Reviewers ran in fresh context: maintainability over the brief and the plan, and an executability dry run.

**Verdicts.** Maintainability: brief Fits with changes; plan Fits with changes. Executability: Not ready (four blockers, all in the draft's assumptions about S11-32 and wording).

**User calls before review (2026-10-22):** split the engine half into S11-32; likeness lines with two different images are shown, pinnable and unchecked; no Weight column.

**Round 2** runs at the STEP-0 refresh, after S11-25 and S11-32 land and the board is agreed.

## Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | `verbatimTitle:` init on `PVSectionHeader` is a second copy pattern; the kit takes `PVCopy` | **Fixed.** D-11: `PVSectionHeader<Title: PVCopy, …>` as `PVDisclosureButton`; brief inventory row updated |
| M-2 | blocker | Section order owned twice: Go sorts by key, Swift moved likeness | **Fixed.** S11-32 scope entry: a lead order per kind in `claimevidence.Rules` / `defaultEvidence()`. D-3: Swift renders wire order; `ownSections`, `likenessSection` and the T-1 case are gone. Brief §2 Order, ES-5, ES-15 |
| M-3 | blocker | Relationship wording a second "A · term · B" mechanism beside `BridgeFact` | **Fixed.** D-14: the relationship names its Source bridge from `flow.bridges` (`endAName`, `phrase`, `endBName`); `BridgeFact.kind` added; `connectionLink` shared with S11-27. The S11-32 label bullet is dropped. Departs from `claim-evidence.md`'s draft "They have a matching birth event": noted to the user |
| M-4 | blocker | Lifespan span text would be a third formatter | **Fixed in part.** D-13: `PersonLifeDisplay.dateSpan` made internal and reused. **Declined** moving `PlacePeriodDisplay.span` onto it: its open and until copy (`placePeriodOpen`, `placePeriodUntil`) differ from a life span's; Out line added |
| M-5 | major | Summary should reuse the row's readout badges, not move `badgeTone` | **Fixed.** D-8: `PromoteReadoutBadges` used by the row and the summary |
| M-6 | major | Speak "partly checked" in `PVCheckbox` | **Fixed.** D-15; brief ES-22 and inventory (Extend); AC-5.1, T-6 |
| M-7 | major | The likeness tile now has a second call site | **Fixed.** D-16: `LikenessTile` shared with the gallery. Kept as a feature view, not a `DesignSystem/Recipes` type, because it wraps `Platform/CachedThumbnail` |
| M-8 | minor | Don't delete and re-add keys for the same jobs | **Fixed.** D-12 keeps `conflictCalloutTitle`, `conflictCalloutBody`, `sheetNeighbor`, `outcome` |
| M-9 | minor | `*Display` naming | **Fixed.** `PromoteEvidenceDisplay`, `PromoteEvidenceDisplayTests` |
| M-10 | minor | One neighbor-naming helper | **Fixed.** D-9 `PromoteModel.neighborName`, used by `reasonText` |

## Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | Lifespan sentences need an event and year S11-32 does not promise | **Fixed.** ES-16 and AC-4.1 – AC-4.3 word only each side's window and the no-lifespan reason, through D-13; S11-32's entry promises windows as date values and the reason |
| X-2 | blocker | Relationship end order and labels cannot be read from the key | **Fixed.** D-14 (see M-3); brief frame 4 rewritten |
| X-3 | blocker | Wording covers only a person row with a matching relationship | **Fixed.** D-14's three outcome sentences cover every row kind and conflict; `outcome` takes the outcome alone for relationships (D-4); AC-3.2, AC-3.3 |
| X-4 | blocker | T-1 contradicts D-2 on likeness; D-3 misstates Go's order | **Fixed.** Order moved to S11-32 (M-2); D-3 corrected |
| X-5 | major | D-11 contradicts its precedent | **Fixed.** By M-1 |
| X-6 | major | Subtitle, summary and row speak different reasons | **Fixed.** D-8: subtitle `readout.reason`; summary Updated line on `row.updated` with `captionReason`; AC-1.2, T-4 |
| X-7 | major | New file breaks private access; pbxproj not named | **Fixed.** D-1 (internal sheet), D-8 (badges in their own view), pbxproj in STEP-2, 3, 5 |
| X-8 | major | S11-25's plan still sends lifespan, likeness and the depiction case to S11-20; S11-32 entry contradicted itself on lifespan offers | **Fixed in the map** (S11-32: the lifespan has no offer). S11-25's plan is being implemented in another session and is not edited here; its STEP-10 docs or the S11-32 plan correct its Out and Next lines. **To user** as a note |
| X-9 | major | No DoD fallback without Xcode; STEP-5–7 checks only manual | **Fixed.** DoD clause as S11-25's; STEP-6 – STEP-8 checks name T-1 – T-3 |
| X-10 | minor | `model.projectDir` does not exist; this Source's caption unnamed | **Fixed.** D-10, D-16: `session.projectKey.projectDir`, caption "This source", first member's crop on a collapsed line |
| X-11 | minor | Manual data cannot show AC-4.3 or AC-4.5 | **Fixed.** §10 data: images A and B, deaths, a death in 1810 |
| X-12 | minor | Icons, loading meta, lifespan in the callout, claim fields for Skip, unrecognized relation, double-spoken state, repeating identifiers | **Fixed.** D-17, D-2 `hasConflict`, D-4, D-15 |

## To user

- **Relationship wording** departs from `claim-evidence.md`'s draft ("They have a matching birth event") to "They share a matching connection — John Smith · subject of · Birth of John Smith", so one owner words every link on the page (M-3, X-2, X-3). Reverting needs labels and a directed phrase from Go in S11-32.
- **S11-25's plan** still names S11-20 for the depiction case and lifespan and likeness lines (X-8); correct it when S11-25 lands.

## User request, 2026-10-22: example data for the board

The user asked for a complete, realistic example set so designers make no data calls. S11-D2 §5 now holds one research project (an 1890 obituary promoted against censuses, parish registers, an album and a funeral card), 13 frames, and specimen lines, each checked against the engine. The check corrected the draft:

| Was | Now | Evidence |
| --- | --- | --- |
| A relationship could read "A connection disagrees" | No relationship disagrees; it matches, or is similar when the neighbor's facts disagree | `core/claimevidence/build.go` `relationships` keeps agree and partial only; `segmentcompare.Compare` gives partial when `Neighbor.Conflicts()` |
| Lifespan sentences showed windows ("about 1815–1890") | They show spans of dated events; one date prints once; no Card when neither side has dates | A window without a birth starts `MaxYears` (110) before the first event (`subjectprofile` `lifespan`) |
| Different toponyms could show a disagreement | No line: toponym, interests and likeness hold many values | `propertycompare` outcome table: Different on a multiple-value property is unknown |
| A conflicting-date event row proposed at Medium | Shown as the researcher's choice, No match | Own evidence 0.5 − 3 with two 0.25 + 0.25 × evidence links stays below the accept bar 2 (`matchregistry` segment and alignment) |
| "The same, written differently", "Matches, with less detail" drawn | Marked unreachable fallbacks | No comparer returns equivalent or contains with another cause (`valuecompare` `text.go`, `name.go`, `date.go`, `words.go`) |

Plan updated to match: AC-3.3, AC-4.1 – AC-4.3, D-2, D-13, D-14, D-17, T-1, T-3, and §10 now walks the brief's frames. S11-32's entry sends spans.

## Board agreed, 2026-10-22

The user drew the board from S11-D2 and supplied its export (`Promote flow.dc.html`, `EvidenceCard.dc.html`). It was read frame by frame (E01 – E14) against the brief and the plan, and S11-25's Landed note (#517) was folded in at the same time.

| Board or landing | Plan change |
| --- | --- |
| Band and Contested Badges at the title's right end, the reason as their tooltip; no subtitle, no summary Card; Updated and contested lines under the title only when they apply | D-8 rewritten; D-18 `PVPanel` title accessory *(kit)*; `sheetSubtitle` removed (D-12); AC-1.1, AC-1.2, T-4 |
| Claim fields above the evidence: Status over Confidence beside the argument, labelled "Additional comments or arguments about this claim (optional)", no hints | D-19 with `PVField.optional` *(kit)*; `argument` text changed and `sheetStatusHint` removed (D-12); AC-1.5, T-7 |
| A sentence under the heading; one " · "-joined detail line; a 116 pt Badge column; micro-caps property labels under a relationship; an info icon on the lifespan line | D-20; `evidenceIntro`, `lineDetail` keys |
| File-search icon on both EmptyStates | D-17 `.searchEmpty` |
| Board's E04 count 15 | S11-D2 §5.3 Frame 4 corrected from 14 |
| Board left the failed message and E13e's Badges out | S11-D2 §5.3 and §9 supply them ("Catalog closed."; No match, "Your choice") |
| Board's italic crop captions | Not copied: the shared `LikenessTile` keeps the gallery's caption (D-16) |
| S11-25: argument hint already removed; `pinRelationship` added; relationship identifier `promote.sheet.relationship.<neighborSubjectID>.<key>` | §4 rows, D-12, D-17 |

Status is `Blocked — on S11-32`. Open questions: none. Round 2 of `review-plan` runs when S11-32 lands, as STEP-0 says.
