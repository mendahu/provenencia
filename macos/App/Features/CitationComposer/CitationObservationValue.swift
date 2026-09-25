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
    }

    enum Failure: Equatable {
        case invalidInteger
        case invalidDate
        case unsupported
    }

    static func fields(from row: ObservationRow) -> Fields {
        Fields(
            valueText: row.valueText,
            valueIntegerText: row.valueIntegerText,
            valueTermID: row.valueTermID,
            valueSubjectID: row.valueSubjectID,
            dateDraft: row.dateDraft,
            nameDraft: row.nameDraft
        )
    }

    static func fields(from draft: CitationComposerModel.ObservationDialogState) -> Fields {
        Fields(
            valueText: draft.valueText,
            valueIntegerText: draft.valueIntegerText,
            valueTermID: draft.valueTermID,
            valueSubjectID: "",
            dateDraft: draft.dateDraft,
            nameDraft: draft.nameDraft
        )
    }

    static func summary(valueType: String, fields: Fields, termLabel: String?) -> String {
        switch valueType {
        case PropertyValueType.text.rawValue:
            return fields.valueText
        case PropertyValueType.integer.rawValue:
            return fields.valueIntegerText
        case PropertyValueType.term.rawValue:
            return termLabel ?? fields.valueTermID
        case PropertyValueType.date.rawValue:
            return dateSummary(fields.dateDraft)
        case PropertyValueType.name.rawValue:
            return NameValueDisplay.string(for: fields.nameDraft)
        case PropertyValueType.subject.rawValue:
            let trimmed = fields.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
            return trimmed.isEmpty ? fields.valueSubjectID : trimmed
        default:
            return ""
        }
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
        default:
            return .unsupported
        }
        return nil
    }

    static func message(for failure: Failure) -> String {
        switch failure {
        case .invalidInteger:
            return String(localized: L10n.CitationComposer.invalidIntegerError)
        case .invalidDate:
            return String(localized: L10n.CitationComposer.invalidDateError)
        case .unsupported:
            return String(localized: L10n.CitationComposer.unsupportedValueTypeError)
        }
    }

    private static func dateSummary(_ draft: DateValueDraft) -> String {
        let formatted = DateValueDisplay.string(for: draft)
        if formatted.isEmpty {
            return String(localized: L10n.CitationComposer.dateUnset)
        }
        return formatted
    }
}
