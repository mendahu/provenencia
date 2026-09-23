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

    static func fields(from row: CitationComposerModel.ObservationRow) -> Fields {
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
        case "text":
            return fields.valueText
        case "integer":
            return fields.valueIntegerText
        case "term":
            return termLabel ?? fields.valueTermID
        case "date":
            return dateSummary(fields.dateDraft)
        case "name":
            return NameValueDisplay.string(for: fields.nameDraft)
        case "subject":
            let trimmed = fields.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
            return trimmed.isEmpty ? fields.valueSubjectID : trimmed
        default:
            return ""
        }
    }

    /// Dialog confirm. Subject values are saved from loaded rows, not picked in the dialog.
    static func isDialogValid(valueType: String, fields: Fields) -> Bool {
        switch valueType {
        case "text":
            return !fields.valueText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        case "integer":
            return Int64(fields.valueIntegerText.trimmingCharacters(in: .whitespacesAndNewlines)) != nil
        case "term":
            return !fields.valueTermID.isEmpty
        case "date":
            return fields.dateDraft.isValid
        case "name":
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
        case "text":
            draft.valueText = fields.valueText
        case "integer":
            guard let value = Int64(fields.valueIntegerText) else { return .invalidInteger }
            draft.valueInteger = value
        case "term":
            draft.valueTermID = fields.valueTermID
        case "date":
            guard fields.dateDraft.isValid else { return .invalidDate }
            draft.date = fields.dateDraft.toInput()
        case "name":
            guard fields.nameDraft.isValid else { return .unsupported }
            let input = fields.nameDraft.toInput()
            draft.nameForm = input.form
            draft.nameParts = input.parts
        case "subject":
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
