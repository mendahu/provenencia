import Foundation

/// Text and marks for a Conclusion detail's fields (S9-15, board S9-D5): a
/// field's state badge, its Source count and disagreement, its other values,
/// each value as text, and what the auto-reconciler did with each record. Go
/// decides every state and reason; this only words them. Shared by Person,
/// Event and Place pages.
enum ReconciledValueDisplay {
    /// The badge beside a field's value. Single and empty fields have none:
    /// their Source count (or empty text) says it.
    enum StateBadge: Equatable {
        case merged, mixed, concluded
    }

    static func stateBadge(_ field: CatalogConclusionField) -> StateBadge? {
        switch field.state {
        case "merged": .merged
        case "mixed": .mixed
        case "concluded": .concluded
        default: nil
        }
    }

    /// Sources behind the shown values: the lead value's support, or for a
    /// mixed field the distinct Sources across every shown value (one Source
    /// can support two values of a multi-valued Property).
    static func sourceCount(_ field: CatalogConclusionField) -> Int {
        let shown = field.displayedValues
        guard shown.count > 1 else { return shown.first?.support ?? 0 }
        let ranks = Set(shown.map(\.rank))
        let sources = Set(field.outcomes.filter { outcome in
            (outcome.reason == "kept" || outcome.reason == "folded") && outcome.valueRank.map(ranks.contains) == true
        }.map(\.sourceID))
        return sources.isEmpty ? shown.reduce(0) { $0 + $1.support } : sources.count
    }

    /// "1 Source", "3 Sources", "by you"; nil for an empty field.
    static func countLine(_ field: CatalogConclusionField) -> String? {
        switch field.state {
        case "": nil
        case "concluded": L10n.string(L10n.Conclusions.countConcluded)
        default: L10n.Conclusions.sourceCount(sourceCount(field))
        }
    }

    /// "1 record disagrees": negative records that count against the lead
    /// value without eliminating it. Nil when none do.
    static func againstLine(_ field: CatalogConclusionField) -> String? {
        let against = field.displayedValues.first?.against ?? 0
        return against > 0 ? L10n.Conclusions.againstCount(against) : nil
    }

    /// The shown values after the lead, which a mixed field discloses.
    static func otherValues(_ field: CatalogConclusionField) -> [CatalogReconciledValue] {
        guard field.state == "mixed" else { return [] }
        return Array(field.displayedValues.dropFirst())
    }

    /// "2 other values", or nil when there are none to disclose.
    static func otherValuesLabel(_ field: CatalogConclusionField) -> String? {
        let others = otherValues(field).count
        return others > 0 ? L10n.Conclusions.otherValues(others) : nil
    }

    /// Honest empty text: "No name recorded", or "Nothing recorded" for a
    /// Property with no wording of its own.
    static func emptyText(propertyKey: String) -> String {
        switch propertyKey {
        case "name": L10n.string(L10n.Conclusions.emptyName)
        case "sex_at_birth": L10n.string(L10n.Conclusions.emptySexAtBirth)
        default: L10n.string(L10n.Conclusions.stateEmpty)
        }
    }

    /// One value as text: names and dates through their own displays, terms
    /// by label (else key). Empty for no value.
    static func string(for value: CatalogConclusionValue, locale: Locale = .autoupdatingCurrent) -> String {
        switch value {
        case .none:
            return ""
        case .text(let text):
            return text.trimmingCharacters(in: .whitespacesAndNewlines)
        case .integer(let integer):
            return integer.formatted(.number.grouping(.never).locale(locale))
        case .term(_, let key, let label):
            return label.isEmpty ? key : label
        case .date(let date):
            return DateValueDisplay.string(for: date, locale: locale)
        case .name(let name):
            return NameValueDisplay.string(for: name)
        }
    }

    /// A negative record: it says "not this". Its outcome is always `against`.
    static func isNegative(_ outcome: CatalogReconcilerOutcome) -> Bool {
        outcome.reason == "against"
    }

    /// What the record said: its value, "not …" for a negative, or "—".
    static func readAs(_ outcome: CatalogReconcilerOutcome, locale: Locale = .autoupdatingCurrent) -> String {
        let value = string(for: outcome.recorded, locale: locale)
        if value.isEmpty { return L10n.string(L10n.Conclusions.readAsNone) }
        return isNegative(outcome) ? L10n.Conclusions.readAsNot(value) : value
    }

    /// Whether "Read as" is a plain phrase ("not …", "—") rather than a value
    /// in its value type's face.
    static func readAsIsPhrase(_ outcome: CatalogReconcilerOutcome) -> Bool {
        isNegative(outcome) || string(for: outcome.recorded).isEmpty
    }

    /// The records that went into the shown value, read in primary ink; the
    /// rest are secondary.
    static func counted(_ outcome: CatalogReconcilerOutcome, in field: CatalogConclusionField) -> Bool {
        switch outcome.reason {
        case "kept", "folded": true
        case "against": deniedSomething(outcome, in: field)
        default: false
        }
    }

    /// A negative record that eliminated a value in this field.
    static func deniedSomething(_ outcome: CatalogReconcilerOutcome, in field: CatalogConclusionField) -> Bool {
        field.outcomes.contains { $0.deniedByObservationID == outcome.observationID }
    }

    /// What happened to one record, in the field it belongs to (S9-D5):
    /// "kept", "folded into James Robins", "outvoted (2 of 3 Sources)",
    /// "weak · low-trust Source", "denied by Court deposition, 1862", and a
    /// negative is "kept" when it denied something, else "disagrees · did
    /// not eliminate".
    static func outcomePhrase(
        _ outcome: CatalogReconcilerOutcome,
        in field: CatalogConclusionField,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        switch outcome.reason {
        case "kept":
            return L10n.string(L10n.Conclusions.outcomeKept)
        case "folded":
            let into = field.values.first { $0.rank == outcome.valueRank }.map { string(for: $0.value, locale: locale) } ?? ""
            return into.isEmpty
                ? L10n.string(L10n.Conclusions.outcomeFolded)
                : L10n.Conclusions.outcomeFoldedInto(into)
        case "outvoted":
            return outcome.voteTotal > 0
                ? L10n.Conclusions.outcomeOutvotedBy(outcome.voteSupport, sources: L10n.Conclusions.sourceCount(outcome.voteTotal))
                : L10n.string(L10n.Conclusions.outcomeOutvoted)
        case "weak":
            let causes = weakCauses(outcome)
            return causes.isEmpty
                ? L10n.string(L10n.Conclusions.outcomeWeak)
                : L10n.Conclusions.outcomeWeakBecause(ListFormatter.localizedString(byJoining: causes))
        case "denied":
            let denier = field.outcomes.first { $0.observationID == outcome.deniedByObservationID }
            let name = denier.map { $0.sourceTitle.isEmpty ? $0.observationRef : $0.sourceTitle } ?? ""
            return name.isEmpty
                ? L10n.string(L10n.Conclusions.outcomeDenied)
                : L10n.Conclusions.outcomeDeniedBy(name)
        case "provisional":
            return L10n.string(L10n.Conclusions.outcomeProvisional)
        case "no_evidence":
            return L10n.string(L10n.Conclusions.outcomeNoEvidence)
        case "against":
            return deniedSomething(outcome, in: field)
                ? L10n.string(L10n.Conclusions.outcomeKept)
                : L10n.string(L10n.Conclusions.outcomeAgainst)
        default:
            return ""
        }
    }

    /// The outcome's mark: one glyph per phrase, so the Why reads without
    /// colour. Decorative beside the phrase.
    static func outcomeMark(_ outcome: CatalogReconcilerOutcome, in field: CatalogConclusionField) -> PVSymbol {
        switch outcome.reason {
        case "kept": .check
        case "folded": .gitMerge
        case "outvoted": .scale
        case "weak": .signalLow
        case "denied": .ban
        case "provisional": .circleDashed
        case "against": deniedSomething(outcome, in: field) ? .check : .circleMinus
        default: .minus
        }
    }

    /// Why a record's evidence is weak, in the auto-reconciler's order: the
    /// Source, the transcription, then the claim.
    static func weakCauses(_ outcome: CatalogReconcilerOutcome) -> [String] {
        var causes: [String] = []
        if outcome.isLowTrustSource { causes.append(L10n.string(L10n.Conclusions.weakLowTrustSource)) }
        if outcome.transcriptionUncertain { causes.append(L10n.string(L10n.Conclusions.weakUncertainTranscription)) }
        if outcome.isLowConfidenceClaim { causes.append(L10n.string(L10n.Conclusions.weakLowConfidenceClaim)) }
        return causes
    }

    /// The spoken state: "merged from 2 Sources", "single, 1 Source", …;
    /// nil for an empty field.
    static func spokenState(_ field: CatalogConclusionField) -> String? {
        let sources = L10n.Conclusions.sourceCount(sourceCount(field))
        switch field.state {
        case "single": return L10n.Conclusions.a11ySingle(sources)
        case "merged": return L10n.Conclusions.a11yMerged(sources)
        case "mixed": return L10n.Conclusions.a11yMixed(sources)
        case "concluded": return L10n.string(L10n.Conclusions.a11yConcluded)
        default: return nil
        }
    }

    /// One field as VoiceOver reads it: "Name, James Robins, merged from 2
    /// Sources, 1 record disagrees", or "Birth date, empty".
    static func accessibilityLabel(label: String, lead: String?, field: CatalogConclusionField?) -> String {
        guard let field, let lead, !lead.isEmpty, let state = spokenState(field) else {
            return L10n.Conclusions.a11yEmpty(label)
        }
        let parts = [label, lead, state] + [againstLine(field)].compactMap { $0 }
        return parts.dropFirst().reduce(parts[0]) { L10n.Conclusions.a11yList($0, rest: $1) }
    }
}
