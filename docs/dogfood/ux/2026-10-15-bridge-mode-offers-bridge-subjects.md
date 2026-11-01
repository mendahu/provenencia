# Bridge mode subject selector still offers bridge subjects

- **Date:** 2026-10-15
- **Type:** Bug (selection rules)
- **Where:** Citation composer / new observation modal / subject selector (bridge mode)
- **Annoyance:** While creating a bridge observation, the two subject selectors list the existing bridge subjects (relationships, participations, locations, place links) alongside the primary subjects. A bridge is by definition between two primary subjects, so picking a bridge there can only end in a refusal.
- **Wanted:** In bridge mode the subject selectors offer primary subjects only; bridge subjects are excluded (or shown disabled with a reason).

## Code notes (observed, not fixed)

- `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift` — `make(...)` builds one `graphSubjects` list from `snapshot.subjects` **and** `snapshot.bridges` (bridges get their edge sentence as the label), then derives the single `subjectComboOptions` menu from it plus the "New <type>…" rows. The doc comment says "One subject menu for every observation row", which is the design that leaks bridges into bridge mode.
- `macos/App/Features/CitationComposer/CitationComposerModel.swift` — `subjectOptions` just returns `vocabulary.subjectComboOptions`; there is no mode-specific variant.
- `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` — both the bridge end pickers (the `PVComboBox` built in the end-selector helper around `draft.endA` / `draft.endB`) and the property-mode subject picker pass `options: model.subjectOptions`. A fix is a filtered list for the bridge pickers: `graphSubjects` whose `typeKey` is not an `EvidenceBridgeKind` raw value (`relationship`, `participation`, `location`, `place_relationship`) — `EvidenceBridgeKind` lives in `macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift`. Whether the "New <type>…" rows belong in bridge mode is a separate product call.
- `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` — `resolvePair(endA:endB:termID:vocabulary:)` looks up a connect rule by `(fromTypeKey, toTypeKey)`; with a bridge on either end no rule matches and it returns the generic `refusedPairTitle` / `refusedPairBody`. That is why the selection is allowed but never works. Keeping that guard is still useful as a backstop after filtering the menu.
- `macos/ProvenenciaTests/` — the existing citation composer tests (e.g. `CitationComposerVocabulary` / authoring tests) are where a "bridge mode options exclude bridges" case would go.
