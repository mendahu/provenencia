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

    var isEmptyDraft: Bool {
        baseline == nil
            && values.propertyID.isEmpty
            && values.valueText.isEmpty
            && values.valueIntegerText.isEmpty
            && values.valueTermID.isEmpty
            && values.valueSubjectID.isEmpty
            && values.dateDraft == .empty()
            && values.nameDraft == .empty()
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

    var polarityMenuTitle: String {
        if values.polarity == ObservationPolarity.negative.rawValue {
            return String(localized: L10n.CitationComposer.affirmObservation)
        }
        return String(localized: L10n.CitationComposer.negateObservation)
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
            nameDraft: NameValueDraft(from: observation)
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
            nameDraft: values.nameDraft
        )
    }

    func asDraft(property: CatalogProperty) -> (CatalogObservationDraft?, String?) {
        if values.subjectID.isEmpty {
            return (nil, String(localized: L10n.CitationComposer.subjectRequiredError))
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
            return (nil, String(localized: L10n.CitationComposer.missingPropertyError))
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
                propertyValueType: property.valueType
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
    struct PendingDelete: Identifiable, Equatable {
        var id: UUID
    }

    var pendingDelete: PendingDelete?

    private weak var context: CitationComposerContext?

    func attach(_ context: CitationComposerContext) {
        self.context = context
    }

    func replace(_ next: [ObservationRow]) {
        rows = next
        focusedID = next.first?.id
        pendingDelete = nil
    }

    func addDraft(subjectID: String) {
        let row = ObservationRow.draft(subjectID: subjectID)
        rows.append(row)
        focusedID = row.id
    }

    func commit(rowID: UUID) async {
        guard let context, let index = rows.firstIndex(where: { $0.id == rowID }) else { return }
        if case .saving = rows[index].state { return }
        guard rows[index].canSave else { return }
        rows[index].state = .saving
        await performCommit(rowID: rowID, context: context)
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

    func requestDelete(rowID: UUID) {
        guard let row = rows.first(where: { $0.id == rowID }),
              row.persistedID != nil
        else { return }
        pendingDelete = PendingDelete(id: rowID)
    }

    func confirmDelete() async {
        guard let context, let pending = pendingDelete else { return }
        pendingDelete = nil
        await performDelete(rowID: pending.id, context: context)
    }

    func cancelDelete() {
        pendingDelete = nil
    }

    func applySubject(rowID: UUID, subjectID: String, vocabulary: CitationComposerVocabulary) {
        mutate(rowID) { row in
            row.subjectID = subjectID
            let allowed = Set(vocabulary.propertyOptions(forSubjectID: subjectID).map(\.id))
            if !allowed.contains(row.propertyID) {
                row.propertyID = ""
                clearValue(&row)
                row.propertyError = String(localized: L10n.CitationComposer.propertyRequiredAfterSubject)
            } else {
                row.propertyError = nil
            }
            row.deriveState()
        }
    }

    func applyProperty(rowID: UUID, propertyID: String) {
        mutate(rowID) { row in
            row.propertyID = propertyID
            clearValue(&row)
            row.propertyError = propertyID.isEmpty
                ? String(localized: L10n.CitationComposer.dialogPropertyRequired)
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

    func applyDialogValues(rowID: UUID, draft: CitationComposerModel.ObservationDialogState) {
        mutate(rowID) { row in
            row.valueText = draft.valueText
            row.valueIntegerText = draft.valueIntegerText
            row.valueTermID = draft.valueTermID
            row.dateDraft = draft.dateDraft
            row.nameDraft = draft.nameDraft
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

    private func performCommit(rowID: UUID, context: CitationComposerContext) async {
        guard let index = rows.firstIndex(where: { $0.id == rowID }) else { return }
        let row = rows[index]
        guard let property = context.vocabulary.property(id: row.propertyID) else {
            markError(rowID: rowID, message: String(localized: L10n.CitationComposer.missingPropertyError))
            return
        }
        let (draftOrNil, draftError) = row.asDraft(property: property)
        if let draftError {
            markError(rowID: rowID, message: draftError)
            return
        }
        guard let draft = draftOrNil else { return }
        let values = context.citationValues()
        do {
            if context.citationID == nil {
                guard let artifactID = context.artifactID else {
                    markError(rowID: rowID, message: String(localized: L10n.CitationComposer.needArtifact))
                    return
                }
                let created = try await context.store.createCitationWithObservations(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    artifactID: artifactID,
                    locatorJSON: values.locator.encodeJSON(),
                    transcription: values.transcription,
                    description: values.description,
                    transcriptionUncertain: values.transcriptionUncertain,
                    transcriptionNote: values.transcriptionNote,
                    citationNotes: [],
                    observations: [draft]
                )
                context.citationID = created.0.id
                context.captureCitationBaseline()
                if let written = created.1.first {
                    markSaved(rowID: rowID, id: written.id, ref: written.ref)
                } else {
                    markError(rowID: rowID, message: String(localized: L10n.CitationComposer.rowStateError))
                }
                context.applySavedCitation()
                await context.reloadListedCitations()
            } else if row.persistedID == nil, let citationID = context.citationID {
                let written = try await context.store.addObservationsToCitation(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    citationID: citationID,
                    observations: [draft]
                )
                if let written = written.first {
                    markSaved(rowID: rowID, id: written.id, ref: written.ref)
                } else {
                    markError(rowID: rowID, message: String(localized: L10n.CitationComposer.rowStateError))
                }
                context.applySavedCitation()
            } else if let citationID = context.citationID {
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
                        observation: observation
                    )
                    markSaved(rowID: rowID, id: written.id, ref: written.ref)
                    context.applySavedCitation()
                } else {
                    markError(rowID: rowID, message: String(localized: L10n.CitationComposer.rowStateError))
                }
            }
        } catch {
            markError(rowID: rowID, message: L10n.Errors.message(for: error))
        }
    }

    private func performDelete(rowID: UUID, context: CitationComposerContext) async {
        guard let row = rows.first(where: { $0.id == rowID }),
              let persistedID = row.persistedID
        else { return }
        do {
            try await context.store.deleteObservation(
                projectDir: context.projectDir,
                userID: context.userID,
                observationID: persistedID
            )
            remove(rowID: rowID)
            context.applySavedCitation()
        } catch {
            markError(rowID: rowID, message: L10n.Errors.message(for: error))
        }
    }

    private func mutate(_ rowID: UUID, _ body: (inout ObservationRow) -> Void) {
        guard let index = rows.firstIndex(where: { $0.id == rowID }) else { return }
        body(&rows[index])
        focusedID = rowID
    }

    private func clearValue(_ row: inout ObservationRow) {
        row.valueText = ""
        row.valueIntegerText = ""
        row.valueTermID = ""
        row.valueSubjectID = ""
        row.dateDraft = .empty()
        row.nameDraft = .empty()
    }
}
