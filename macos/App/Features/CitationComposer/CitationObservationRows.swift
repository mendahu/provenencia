import Foundation
import Observation

enum ObservationRowState: Equatable {
    case draft
    case saved
    case edited
    case saving
    case error(String)
}

struct ObservationRowValues: Equatable {
    var subjectID: String = ""
    var propertyID: String = ""
    var polarity: String = ObservationPolarity.positive.rawValue
    var valueFields = CitationObservationValue.Fields()

    /// Resets every value field; subject, property, and polarity stay.
    mutating func clearValue() {
        valueFields = .init()
    }
}

struct ObservationRow: Identifiable, Equatable {
    var id: UUID
    var values: ObservationRowValues
    var baseline: ObservationRowValues?
    var state: ObservationRowState
    /// The catalog row last stored; nil on a draft. The saved line reads
    /// every part of itself from it, never from `values`.
    var saved: CatalogObservation?

    var persistedID: String? { saved?.id }
    var persistedRef: String? { saved?.ref }

    var propertyID: String { values.propertyID }

    var canSave: Bool {
        switch state {
        case .saving:
            return false
        case .draft, .edited, .error:
            return !values.subjectID.isEmpty
                && !values.propertyID.isEmpty
        case .saved:
            return false
        }
    }

    mutating func deriveState() {
        if case .saving = state { return }
        if baseline == nil {
            state = .draft
            return
        }
        state = values == baseline ? .saved : .edited
    }

    mutating func markSaved(_ written: CatalogObservation) {
        saved = written
        baseline = values
        state = .saved
    }

    static func loaded(from observation: CatalogObservation) -> ObservationRow {
        var fields = CitationObservationValue.Fields(
            valueText: observation.valueText,
            valueIntegerText: observation.valueInteger.map(String.init) ?? "",
            valueBoolean: observation.valueBoolean,
            valueTermID: observation.valueTermID,
            valueSubjectID: observation.valueSubjectID,
            dateDraft: observation.date.map(DateValueDraft.init(from:)) ?? .empty(),
            nameDraft: NameValueDraft(from: observation),
            depictionDraft: DepictionValueDraft(from: observation)
        )
        if fields.valueText.isEmpty, !observation.nameForm.isEmpty {
            fields.valueText = observation.nameForm
        }
        let values = ObservationRowValues(
            subjectID: observation.subjectID,
            propertyID: observation.propertyID,
            polarity: observation.polarity.isEmpty
                ? ObservationPolarity.positive.rawValue
                : observation.polarity,
            valueFields: fields
        )
        return ObservationRow(
            id: UUID(),
            values: values,
            baseline: values,
            state: .saved,
            saved: observation
        )
    }

    func asDraft(
        property: CatalogProperty,
        vocabulary: CitationComposerVocabulary
    ) -> (CatalogObservationDraft?, String?) {
        if values.subjectID.isEmpty {
            return (nil, L10n.string(L10n.CitationComposer.subjectRequiredError))
        }
        var draft = CatalogObservationDraft(
            subjectID: values.subjectID,
            propertyID: property.id,
            polarity: values.polarity == ObservationPolarity.negative.rawValue
                && vocabulary.offersNegative(propertyID: property.id)
                ? ObservationPolarity.negative.rawValue
                : ObservationPolarity.positive.rawValue
        )
        if let failure = CitationObservationValue.apply(
            valueType: property.valueType,
            fields: values.valueFields,
            to: &draft
        ) {
            return (nil, CitationObservationValue.message(for: failure))
        }
        return (draft, nil)
    }

    func asCatalogObservation(
        property: CatalogProperty,
        vocabulary: CitationComposerVocabulary
    ) -> (CatalogObservation?, String?) {
        guard let persistedID, let persistedRef else {
            return (nil, L10n.string(L10n.CitationComposer.missingPropertyError))
        }
        let (draft, message) = asDraft(property: property, vocabulary: vocabulary)
        if let message { return (nil, message) }
        guard let draft else { return (nil, message) }
        return (
            CatalogObservation(
                id: persistedID,
                ref: persistedRef,
                citationID: "",
                subjectID: draft.subjectID,
                propertyID: draft.propertyID,
                polarity: draft.polarity,
                valueText: draft.valueText,
                valueInteger: draft.valueInteger,
                valueBoolean: draft.valueBoolean,
                valueDateID: draft.valueDateID,
                date: draft.date,
                valueNameID: draft.valueNameID,
                nameForm: draft.nameForm,
                nameParts: draft.nameParts,
                valueSubjectID: draft.valueSubjectID,
                valueTermID: draft.valueTermID,
                propertyKey: property.key,
                propertyLabel: property.label,
                propertyValueType: property.valueType,
                valueDepictionID: values.valueFields.depictionDraft.valueDepictionID
            ),
            nil
        )
    }
}

/// Ordinary Observation rows (never edges) and their row-level commands.
@MainActor
@Observable
final class CitationObservationRows {
    var rows: [ObservationRow] = []

    private weak var context: CitationComposerContext?

    func attach(_ context: CitationComposerContext) {
        self.context = context
    }

    func replace(_ next: [ObservationRow]) {
        rows = next
    }

    /// `depictionInput` is a new region and image for a depiction Observation;
    /// nil keeps the saved one.
    func commit(
        rowID: UUID,
        citation: CitationFieldsDraft.Values,
        depictionInput: CatalogDepictionValueInput? = nil
    ) async {
        guard let context, let index = rows.firstIndex(where: { $0.id == rowID }) else { return }
        if case .saving = rows[index].state { return }
        guard rows[index].canSave else { return }
        rows[index].state = .saving
        await performCommit(
            rowID: rowID,
            context: context,
            citation: citation,
            depictionInput: depictionInput
        )
    }

    func revert(rowID: UUID) {
        guard let index = rows.firstIndex(where: { $0.id == rowID }) else { return }
        guard let baseline = rows[index].baseline else { return }
        rows[index].values = baseline
        rows[index].state = .saved
    }

    func deletePersisted(rowID: UUID) async {
        guard let context else { return }
        await performDelete(rowID: rowID, context: context)
    }

    func replaceValues(rowID: UUID, values: ObservationRowValues) {
        mutate(rowID) { row in
            row.values = values
            row.deriveState()
        }
    }

    func markSaved(rowID: UUID, written: CatalogObservation) {
        mutate(rowID) { row in
            row.markSaved(written)
        }
    }

    func markError(rowID: UUID, message: String) {
        mutate(rowID) { row in
            row.state = .error(message)
        }
    }

    func remove(rowID: UUID) {
        rows.removeAll { $0.id == rowID }
    }

    private func performCommit(
        rowID: UUID,
        context: CitationComposerContext,
        citation: CitationFieldsDraft.Values,
        depictionInput: CatalogDepictionValueInput?
    ) async {
        guard let index = rows.firstIndex(where: { $0.id == rowID }) else { return }
        let row = rows[index]
        guard let property = context.vocabulary.property(id: row.propertyID) else {
            markError(rowID: rowID, message: L10n.string(L10n.CitationComposer.missingPropertyError))
            return
        }
        let (draftOrNil, draftError) = row.asDraft(property: property, vocabulary: context.vocabulary)
        if let draftError {
            markError(rowID: rowID, message: draftError)
            return
        }
        guard draftOrNil != nil else { return }
        do {
            let citationID = try await context.ensureCitationID(citation)
            let (observationOrNil, updateError) = row.asCatalogObservation(property: property, vocabulary: context.vocabulary)
            if let updateError {
                markError(rowID: rowID, message: updateError)
                return
            }
            guard var observation = observationOrNil else {
                markError(rowID: rowID, message: L10n.string(L10n.CitationComposer.observationSaveFailed))
                return
            }
            observation.citationID = citationID
            let written = try await context.store.updateObservation(
                projectDir: context.projectDir,
                userID: context.userID,
                observation: observation,
                depictionInput: depictionInput
            )
            markSaved(rowID: rowID, written: written.0)
            context.applyWrite(written.1)
        } catch is CitationComposerNeedArtifact {
            markError(rowID: rowID, message: L10n.string(L10n.CitationComposer.needArtifact))
        } catch {
            markError(rowID: rowID, message: L10n.Errors.message(for: error))
        }
    }

    private func performDelete(rowID: UUID, context: CitationComposerContext) async {
        guard let row = rows.first(where: { $0.id == rowID }),
              let persistedID = row.persistedID
        else { return }
        do {
            let effects = try await context.store.deleteObservation(
                projectDir: context.projectDir,
                userID: context.userID,
                observationID: persistedID
            )
            remove(rowID: rowID)
            context.applyWrite(effects)
        } catch {
            markError(rowID: rowID, message: L10n.Errors.message(for: error))
        }
    }

    private func mutate(_ rowID: UUID, _ body: (inout ObservationRow) -> Void) {
        guard let index = rows.firstIndex(where: { $0.id == rowID }) else { return }
        body(&rows[index])
    }
}
