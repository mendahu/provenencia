import Foundation

/// Text for a Conclusion detail's fields (S9-15): a field's state line, its
/// *+N*, each value, and what the auto-reconciler did with each record. Go
/// decides every state and reason; this only words them. S9-D5's
/// `ReconciledValueRow` draws with it.
enum ReconciledValueDisplay {
    /// "1 Source", "merged · 3 Sources", "mixed", or "Nothing recorded".
    static func stateLine(_ field: CatalogConclusionField) -> String {
        let support = field.displayedValues.first?.support ?? 0
        switch field.state {
        case "":
            return L10n.string(L10n.Conclusions.stateEmpty)
        case "single":
            return L10n.Conclusions.sourceCount(support)
        case "merged":
            return L10n.Conclusions.stateMerged(sourceCount: support)
        case "mixed":
            return L10n.string(L10n.Conclusions.stateMixed)
        case "concluded":
            return L10n.string(L10n.Conclusions.stateConcluded)
        default:
            return ""
        }
    }

    /// Displayed values beyond the first; the field shows *+N* when above 0.
    static func additionalCount(_ field: CatalogConclusionField) -> Int {
        max(0, field.displayedValues.count - 1)
    }

    /// "+2", or nil when every displayed value is shown.
    static func additionalLabel(_ field: CatalogConclusionField) -> String? {
        let count = additionalCount(field)
        return count > 0 ? L10n.Conclusions.additionalValues(count) : nil
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

    /// What happened to one record, in the field it belongs to: "kept",
    /// "folded into James Robins", "weak · low-trust Source", "denied by
    /// OBS-7KD45", and so on.
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
            return L10n.string(L10n.Conclusions.outcomeOutvoted)
        case "weak":
            let causes = weakCauses(outcome)
            return causes.isEmpty
                ? L10n.string(L10n.Conclusions.outcomeWeak)
                : L10n.Conclusions.outcomeWeakBecause(ListFormatter.localizedString(byJoining: causes))
        case "denied":
            let ref = field.outcomes.first { $0.observationID == outcome.deniedByObservationID }?.observationRef ?? ""
            return ref.isEmpty
                ? L10n.string(L10n.Conclusions.outcomeDenied)
                : L10n.Conclusions.outcomeDeniedBy(ref)
        case "provisional":
            return L10n.string(L10n.Conclusions.outcomeProvisional)
        case "no_evidence":
            return L10n.string(L10n.Conclusions.outcomeNoEvidence)
        case "against":
            return L10n.string(L10n.Conclusions.outcomeAgainst)
        default:
            return ""
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
}
