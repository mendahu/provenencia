import Foundation

/// Display, validation, and save-draft fill for one composer observation value.
enum CitationObservationValue {
    struct Fields: Equatable {
        var valueText: String = ""
        var valueIntegerText: String = ""
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
        switch valueType {
        case PropertyValueType.text.rawValue:
            return !fields.valueText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        case PropertyValueType.integer.rawValue:
            return Int64(fields.valueIntegerText.trimmingCharacters(in: .whitespacesAndNewlines)) != nil
        case PropertyValueType.term.rawValue:
            return !fields.valueTermID.isEmpty
        case PropertyValueType.date.rawValue:
            return fields.dateDraft.isValid
        case PropertyValueType.name.rawValue:
            return fields.nameDraft.isValid
        case PropertyValueType.depiction.rawValue:
            return fields.depictionDraft.hasRegion
        default:
            return false
        }
    }

    /// Writes the value onto `draft`. Returns a failure when the value cannot be saved.
    static func apply(
        valueType: String,
        fields: Fields,
        to draft: inout CatalogObservationDraft
    ) -> Failure? {
        switch valueType {
        case PropertyValueType.text.rawValue:
            draft.valueText = fields.valueText
        case PropertyValueType.integer.rawValue:
            guard let value = Int64(fields.valueIntegerText) else { return .invalidInteger }
            draft.valueInteger = value
        case PropertyValueType.term.rawValue:
            draft.valueTermID = fields.valueTermID
        case PropertyValueType.date.rawValue:
            guard fields.dateDraft.isValid else { return .invalidDate }
            draft.date = fields.dateDraft.toInput()
        case PropertyValueType.name.rawValue:
            guard fields.nameDraft.isValid else { return .unsupported }
            let input = fields.nameDraft.toInput()
            draft.nameForm = input.form
            draft.nameParts = input.parts
        case PropertyValueType.subject.rawValue:
            guard !fields.valueSubjectID.isEmpty else { return .unsupported }
            draft.valueSubjectID = fields.valueSubjectID
        case PropertyValueType.depiction.rawValue:
            // The model attaches the rendered image after this (it cannot render here).
            guard fields.depictionDraft.hasRegion else { return .unsupported }
        default:
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
