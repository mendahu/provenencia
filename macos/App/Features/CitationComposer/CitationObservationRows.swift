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
    var valueText: String = ""
    var valueIntegerText: String = ""
    var valueTermID: String = ""
    var valueSubjectID: String = ""
    var dateDraft: DateValueDraft = .empty()
    var nameDraft: NameValueDraft = .empty()
    var depictionDraft: DepictionValueDraft = .empty()

    /// Resets every value field; subject, property, and polarity stay.
    mutating func clearValue() {
        valueText = ""
        valueIntegerText = ""
        valueTermID = ""
        valueSubjectID = ""
        dateDraft = .empty()
        nameDraft = .empty()
        depictionDraft = .empty()
    }
}

struct ObservationRow: Identifiable, Equatable {
    var id: UUID
    var persistedID: String?
    var persistedRef: String?
    var values: ObservationRowValues
    var baseline: ObservationRowValues?
    var state: ObservationRowState
    var propertyError: String?

    var subjectID: String {
        get { values.subjectID }
        set { values.subjectID = newValue }
    }

    var propertyID: String {
        get { values.propertyID }
        set { values.propertyID = newValue }
    }

    var polarity: String {
        get { values.polarity }
        set { values.polarity = newValue }
    }

    var valueText: String {
        get { values.valueText }
        set { values.valueText = newValue }
    }

    var valueIntegerText: String {
        get { values.valueIntegerText }
        set { values.valueIntegerText = newValue }
    }

    var valueTermID: String {
        get { values.valueTermID }
        set { values.valueTermID = newValue }
    }

    var valueSubjectID: String {
        get { values.valueSubjectID }
        set { values.valueSubjectID = newValue }
    }

    var dateDraft: DateValueDraft {
        get { values.dateDraft }
        set { values.dateDraft = newValue }
    }

    var nameDraft: NameValueDraft {
        get { values.nameDraft }
        set { values.nameDraft = newValue }
    }

    var depictionDraft: DepictionValueDraft {
        get { values.depictionDraft }
        set { values.depictionDraft = newValue }
    }

    var isEmptyDraft: Bool {
        baseline == nil
            && values.propertyID.isEmpty
            && values.valueText.isEmpty
            && values.valueIntegerText.isEmpty
            && values.valueTermID.isEmpty
            && values.valueSubjectID.isEmpty
            && values.dateDraft == .empty()
            && values.nameDraft == .empty()
            && values.depictionDraft == .empty()
    }

    var canSave: Bool {
        switch state {
        case .saving:
            return false
        case .draft, .edited, .error:
            return !values.subjectID.isEmpty
                && !values.propertyID.isEmpty
                && propertyError == nil
        case .saved:
            return false
        }
    }

    var isNegated: Bool {
        values.polarity == ObservationPolarity.negative.rawValue
    }

    var polarityMenuTitle: String {
        if isNegated {
            return L10n.string(L10n.CitationComposer.affirmObservation)
        }
        return L10n.string(L10n.CitationComposer.negateObservation)
    }

    mutating func deriveState() {
        if case .saving = state { return }
        if baseline == nil {
            state = .draft
            return
        }
        state = values == baseline ? .saved : .edited
    }

    mutating func markSaved(id: String, ref: String) {
        persistedID = id
        persistedRef = ref
        baseline = values
        propertyError = nil
        state = .saved
    }

    static func draft(subjectID: String = "") -> ObservationRow {
        ObservationRow(
            id: UUID(),
            persistedID: nil,
            persistedRef: nil,
            values: ObservationRowValues(subjectID: subjectID),
            baseline: nil,
            state: .draft,
            propertyError: nil
        )
    }

    static func loaded(from observation: CatalogObservation) -> ObservationRow {
        var values = ObservationRowValues(
            subjectID: observation.subjectID,
            propertyID: observation.propertyID,
            polarity: observation.polarity.isEmpty
                ? ObservationPolarity.positive.rawValue
                : observation.polarity,
            valueText: observation.valueText,
            valueIntegerText: observation.valueInteger.map(String.init) ?? "",
            valueTermID: observation.valueTermID,
            valueSubjectID: observation.valueSubjectID,
            dateDraft: observation.date.map(DateValueDraft.init(from:)) ?? .empty(),
            nameDraft: NameValueDraft(from: observation),
            depictionDraft: DepictionValueDraft(from: observation)
        )
        if values.valueText.isEmpty, !observation.nameForm.isEmpty {
            values.valueText = observation.nameForm
        }
        return ObservationRow(
            id: UUID(),
            persistedID: observation.id,
            persistedRef: observation.ref,
            values: values,
            baseline: values,
            state: .saved,
            propertyError: nil
        )
    }

    func valueFields() -> CitationObservationValue.Fields {
        CitationObservationValue.Fields(
            valueText: values.valueText,
            valueIntegerText: values.valueIntegerText,
            valueTermID: values.valueTermID,
            valueSubjectID: values.valueSubjectID,
            dateDraft: values.dateDraft,
            nameDraft: values.nameDraft,
            depictionDraft: values.depictionDraft
        )
    }

    func asDraft(property: CatalogProperty) -> (CatalogObservationDraft?, String?) {
        if values.subjectID.isEmpty {
            return (nil, L10n.string(L10n.CitationComposer.subjectRequiredError))
        }
        var draft = CatalogObservationDraft(
            subjectID: values.subjectID,
            propertyID: property.id,
            polarity: values.polarity == ObservationPolarity.negative.rawValue
                ? ObservationPolarity.negative.rawValue
                : ObservationPolarity.positive.rawValue
        )
        if let failure = CitationObservationValue.apply(
            valueType: property.valueType,
            fields: valueFields(),
            to: &draft
        ) {
            return (nil, CitationObservationValue.message(for: failure))
        }
        return (draft, nil)
    }

    func asCatalogObservation(property: CatalogProperty) -> (CatalogObservation?, String?) {
        guard let persistedID, let persistedRef else {
            return (nil, L10n.string(L10n.CitationComposer.missingPropertyError))
        }
        let (draft, message) = asDraft(property: property)
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
                valueDepictionID: values.depictionDraft.valueDepictionID
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
    var focusedID: UUID?

    private weak var context: CitationComposerContext?

    func attach(_ context: CitationComposerContext) {
        self.context = context
    }

    func replace(_ next: [ObservationRow]) {
        rows = next
        focusedID = next.first?.id
    }

    func addDraft(subjectID: String) {
        let row = ObservationRow.draft(subjectID: subjectID)
        rows.append(row)
        focusedID = row.id
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
        if rows[index].baseline == nil {
            rows.remove(at: index)
            return
        }
        rows[index].values = rows[index].baseline!
        rows[index].propertyError = nil
        rows[index].state = .saved
    }

    func deletePersisted(rowID: UUID) async {
        guard let context else { return }
        await performDelete(rowID: rowID, context: context)
    }

    func applyProperty(rowID: UUID, propertyID: String) {
        mutate(rowID) { row in
            row.propertyID = propertyID
            row.values.clearValue()
            row.propertyError = propertyID.isEmpty
                ? L10n.string(L10n.CitationComposer.dialogPropertyRequired)
                : nil
            row.deriveState()
        }
    }

    func applyText(rowID: UUID, text: String) {
        mutate(rowID) { row in
            row.valueText = text
            row.deriveState()
        }
    }

    func applyInteger(rowID: UUID, text: String) {
        mutate(rowID) { row in
            row.valueIntegerText = text
            row.deriveState()
        }
    }

    func applyTerm(rowID: UUID, termID: String) {
        mutate(rowID) { row in
            row.valueTermID = termID
            row.deriveState()
        }
    }

    func replaceValues(rowID: UUID, values: ObservationRowValues) {
        mutate(rowID) { row in
            row.values = values
            row.deriveState()
        }
    }

    func togglePolarity(rowID: UUID) {
        mutate(rowID) { row in
            row.polarity = row.polarity == ObservationPolarity.negative.rawValue
                ? ObservationPolarity.positive.rawValue
                : ObservationPolarity.negative.rawValue
            row.deriveState()
        }
    }

    func markSaved(rowID: UUID, id: String, ref: String) {
        mutate(rowID) { row in
            row.markSaved(id: id, ref: ref)
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

    var hasEditedOrSaving: Bool {
        rows.contains {
            switch $0.state {
            case .edited, .saving, .error: return true
            case .draft: return !$0.isEmptyDraft
            case .saved: return false
            }
        }
    }

    var unsavedObservationCount: Int {
        rows.filter {
            if case .saved = $0.state { return false }
            return !$0.isEmptyDraft
        }.count
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
        let (draftOrNil, draftError) = row.asDraft(property: property)
        if let draftError {
            markError(rowID: rowID, message: draftError)
            return
        }
        guard var draft = draftOrNil else { return }
        draft.depiction = depictionInput
        do {
            let citationID = try await context.ensureCitationID(citation)
            if row.persistedID == nil {
                let written = try await context.store.addObservationsToCitation(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    citationID: citationID,
                    observations: [draft]
                )
                if let saved = written.0.first {
                    markSaved(rowID: rowID, id: saved.id, ref: saved.ref)
                } else {
                    markError(rowID: rowID, message: L10n.string(L10n.CitationComposer.observationSaveFailed))
                }
                context.applyWrite(written.1)
            } else {
                let (observationOrNil, updateError) = row.asCatalogObservation(property: property)
                if let updateError {
                    markError(rowID: rowID, message: updateError)
                    return
                }
                if var observation = observationOrNil {
                    observation.citationID = citationID
                    let written = try await context.store.updateObservation(
                        projectDir: context.projectDir,
                        userID: context.userID,
                        observation: observation,
                        depictionInput: depictionInput
                    )
                    markSaved(rowID: rowID, id: written.0.id, ref: written.0.ref)
                    context.applyWrite(written.1)
                } else {
                    markError(rowID: rowID, message: L10n.string(L10n.CitationComposer.observationSaveFailed))
                }
            }
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
        focusedID = rowID
    }
}
