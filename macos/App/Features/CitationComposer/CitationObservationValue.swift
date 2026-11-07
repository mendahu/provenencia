import Foundation

/// Validation and save-draft fill for one composer observation value.
enum CitationObservationValue {
    struct Fields: Equatable {
        var valueText: String = ""
        var valueIntegerText: String = ""
        /// nil is "not chosen"; a yes/no is saved only once one is picked.
        var valueBoolean: Bool?
        var valueTermID: String = ""
        var valueSubjectID: String = ""
        var dateDraft: DateValueDraft = .empty()
        var nameDraft: NameValueDraft = .empty()
        var depictionDraft: DepictionValueDraft = .empty()
    }

    enum Failure: Equatable {
        case invalidInteger
        case invalidDate
        case unsupported
    }

    /// Dialog confirm. Subject values are saved from loaded rows, not picked in the dialog.
    static func isDialogValid(valueType: String, fields: Fields) -> Bool {
        switch PropertyValueType(rawValue: valueType) {
        case .text:
            return !fields.valueText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        case .integer:
            return Int64(fields.valueIntegerText.trimmingCharacters(in: .whitespacesAndNewlines)) != nil
        case .boolean:
            return fields.valueBoolean != nil
        case .term:
            return !fields.valueTermID.isEmpty
        case .date:
            return fields.dateDraft.isValid
        case .name:
            return fields.nameDraft.isValid
        case .depiction:
            return fields.depictionDraft.hasRegion
        case .subject, nil:
            return false
        }
    }

    /// Writes the value onto `draft`. Returns a failure when the value cannot be saved.
    static func apply(
        valueType: String,
        fields: Fields,
        to draft: inout CatalogObservationDraft
    ) -> Failure? {
        switch PropertyValueType(rawValue: valueType) {
        case .text:
            draft.valueText = fields.valueText
        case .integer:
            guard let value = Int64(fields.valueIntegerText) else { return .invalidInteger }
            draft.valueInteger = value
        case .boolean:
            guard let value = fields.valueBoolean else { return .unsupported }
            draft.valueBoolean = value
        case .term:
            draft.valueTermID = fields.valueTermID
        case .date:
            guard fields.dateDraft.isValid else { return .invalidDate }
            draft.date = fields.dateDraft.toInput()
        case .name:
            guard fields.nameDraft.isValid else { return .unsupported }
            let input = fields.nameDraft.toInput()
            draft.nameForm = input.form
            draft.nameParts = input.parts
        case .subject:
            guard !fields.valueSubjectID.isEmpty else { return .unsupported }
            draft.valueSubjectID = fields.valueSubjectID
        case .depiction:
            // The model attaches the rendered image after this (it cannot render here).
            guard fields.depictionDraft.hasRegion else { return .unsupported }
        case nil:
            return .unsupported
        }
        return nil
    }

    static func message(for failure: Failure) -> String {
        switch failure {
        case .invalidInteger:
            return L10n.string(L10n.CitationComposer.invalidIntegerError)
        case .invalidDate:
            return L10n.string(L10n.CitationComposer.invalidDateError)
        case .unsupported:
            return L10n.string(L10n.CitationComposer.unsupportedValueTypeError)
        }
    }
}
