import Foundation

/// Editable draft of a genealogical NameValue (form + optional ordered parts).
/// Validation mirrors `core/database/namevalues`: form required; empty part
/// values refused; unknown non-empty types refused; empty type is untyped.
struct NameValueDraft: Equatable, Sendable {
    struct Part: Identifiable, Equatable, Sendable {
        var id: UUID
        var value: String
        var type: String

        static func empty() -> Part {
            Part(id: UUID(), value: "", type: "")
        }
    }

    var form: String
    var parts: [Part]
    /// Show the missing-form error after the field has been edited.
    var formTouched: Bool

    static func empty() -> NameValueDraft {
        NameValueDraft(form: "", parts: [], formTouched: false)
    }

    var trimmedForm: String {
        form.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    var isValid: Bool {
        guard !trimmedForm.isEmpty else { return false }
        for part in parts {
            let value = part.value.trimmingCharacters(in: .whitespacesAndNewlines)
            if value.isEmpty { return false }
            if !NamePartType.isAllowed(part.type) { return false }
        }
        return true
    }

    var formError: String? {
        guard trimmedForm.isEmpty, formTouched else { return nil }
        return String(localized: L10n.NameValue.formErrorMissing)
    }

    func partValueError(at index: Int) -> String? {
        guard parts.indices.contains(index) else { return nil }
        let value = parts[index].value.trimmingCharacters(in: .whitespacesAndNewlines)
        guard value.isEmpty else { return nil }
        return L10n.NameValue.partEmptyValue(position: index + 1)
    }

    @discardableResult
    mutating func addPart() -> UUID {
        let part = Part.empty()
        parts.append(part)
        return part.id
    }

    mutating func removePart(id: UUID) {
        parts.removeAll { $0.id == id }
    }

    mutating func movePart(from source: IndexSet, to destination: Int) {
        parts.move(fromOffsets: source, toOffset: destination)
    }

    mutating func movePart(at index: Int, by delta: Int) {
        let next = index + delta
        guard parts.indices.contains(index), parts.indices.contains(next) else { return }
        parts.swapAt(index, next)
    }

    func toInput() -> (form: String, parts: [CatalogNameValuePart]) {
        (
            form: trimmedForm,
            parts: parts.map {
                CatalogNameValuePart(
                    value: $0.value.trimmingCharacters(in: .whitespacesAndNewlines),
                    type: $0.type.trimmingCharacters(in: .whitespacesAndNewlines)
                )
            }
        )
    }

    /// Board “Stored as” form line (`“Ada Lovelace”` or `—`).
    var storedFormLine: String {
        trimmedForm.isEmpty ? "—" : "“\(trimmedForm)”"
    }

    /// Board “Stored as” parts line (`1 given:Ada · 2 surname:Lovelace` or `[ ]`).
    var storedPartsLine: String {
        guard !parts.isEmpty else { return "[ ]" }
        return parts.enumerated().map { index, part in
            let value = part.value.trimmingCharacters(in: .whitespacesAndNewlines)
            let type = part.type.trimmingCharacters(in: .whitespacesAndNewlines)
            let typeToken = type.isEmpty ? "\"\"" : type
            let valueToken = value.isEmpty ? "…" : value
            return "\(index + 1) \(typeToken):\(valueToken)"
        }.joined(separator: " · ")
    }
}

extension NameValueDraft {
    init(form: String, parts: [CatalogNameValuePart]) {
        self.init(
            form: form,
            parts: parts.map { Part(id: UUID(), value: $0.value, type: $0.type) },
            formTouched: !form.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        )
    }

    init(from observation: CatalogObservation) {
        self.init(form: observation.nameForm, parts: observation.nameParts)
    }
}
