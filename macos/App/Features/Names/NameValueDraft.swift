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
    /// True once a person changed a part, or the parts were loaded from the catalog.
    /// A fill from the full form clears it.
    var partsHandEdited: Bool = false

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
        return L10n.string(L10n.NameValue.formErrorMissing)
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
        partsHandEdited = true
        return part.id
    }

    mutating func removePart(id: UUID) {
        let before = parts.count
        parts.removeAll { $0.id == id }
        if parts.count != before { partsHandEdited = true }
    }

    mutating func movePart(at index: Int, by delta: Int) {
        let next = index + delta
        guard parts.indices.contains(index), parts.indices.contains(next) else { return }
        parts.swapAt(index, next)
        partsHandEdited = true
    }

    mutating func setPartValue(id: UUID, value: String) {
        guard let index = parts.firstIndex(where: { $0.id == id }) else { return }
        guard parts[index].value != value else { return }
        parts[index].value = value
        partsHandEdited = true
    }

    mutating func setPartType(id: UUID, type: String) {
        guard let index = parts.firstIndex(where: { $0.id == id }) else { return }
        guard parts[index].type != type else { return }
        parts[index].type = type
        partsHandEdited = true
    }

    /// Leaving the full-form field fills parts only when the list is empty.
    mutating func fillEmptyPartsFromForm() {
        guard parts.isEmpty else { return }
        let next = NameFormSplit.parts(from: form)
        guard !next.isEmpty else { return }
        parts = next
        partsHandEdited = false
    }

    /// Split refills from the full form. Hand-edited parts come back as a snapshot
    /// for the confirm; the draft stays put until that snapshot is applied.
    mutating func splitFromForm() -> NameSplitApply {
        let next = NameFormSplit.parts(from: form)
        guard !next.isEmpty else { return .none }
        if partsHandEdited, !parts.isEmpty {
            return .needsConfirm(next)
        }
        parts = next
        partsHandEdited = false
        return .applied
    }

    mutating func applySplit(_ next: [Part]) {
        parts = next
        partsHandEdited = false
    }

    var canSplitForm: Bool {
        !NameFormSplit.parts(from: form).isEmpty
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
        let mapped = parts.map { Part(id: UUID(), value: $0.value, type: $0.type) }
        self.init(
            form: form,
            parts: mapped,
            formTouched: !form.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            partsHandEdited: !mapped.isEmpty
        )
    }

    init(from observation: CatalogObservation) {
        self.init(form: observation.nameForm, parts: observation.nameParts)
    }
}

/// What Split does with the current full form.
enum NameSplitApply: Equatable {
    case none
    case applied
    case needsConfirm([NameValueDraft.Part])
}

/// Western split of a full-form line. Titles come from `NameTitleRegistry`.
enum NameFormSplit {
    static func parts(from form: String) -> [NameValueDraft.Part] {
        let collapsed = form
            .split(whereSeparator: \.isWhitespace)
            .map(String.init)
            .joined(separator: " ")
        guard !collapsed.isEmpty else { return [] }
        guard let comma = collapsed.firstIndex(of: ",") else {
            return classify(words(in: collapsed), allowSurname: true)
        }
        let left = collapsed[..<comma].trimmingCharacters(in: .whitespaces)
        let rightIndex = collapsed.index(after: comma)
        let right = rightIndex < collapsed.endIndex ? String(collapsed[rightIndex...]) : ""
        var result: [NameValueDraft.Part] = []
        if !left.isEmpty {
            result.append(part(String(left), .surname))
        }
        result.append(contentsOf: classify(words(in: right), allowSurname: false))
        return result
    }

    /// A comma stuck to a word (`John,`) is a separator, not part of the name.
    private static func words(in text: String) -> [String] {
        text.split(whereSeparator: \.isWhitespace).compactMap { token in
            let stripped = token.trimmingCharacters(in: CharacterSet(charactersIn: ","))
            return stripped.isEmpty ? nil : String(stripped)
        }
    }

    /// Peel titles only while another word remains. One leftover word is given.
    /// With a surname allowed, the last of two or more leftovers is the surname.
    private static func classify(_ tokens: [String], allowSurname: Bool) -> [NameValueDraft.Part] {
        var words = tokens
        var prefixes: [String] = []
        var suffixes: [String] = []
        while words.count > 1, NameTitleRegistry.isPrefix(words[0]) {
            prefixes.append(words.removeFirst())
        }
        while words.count > 1, NameTitleRegistry.isSuffix(words[words.count - 1]) {
            suffixes.insert(words.removeLast(), at: 0)
        }
        var result = prefixes.map { part($0, .prefix) }
        if words.count <= 1 || !allowSurname {
            result.append(contentsOf: words.map { part($0, .given) })
        } else {
            let surname = words.removeLast()
            result.append(contentsOf: words.map { part($0, .given) })
            result.append(part(surname, .surname))
        }
        result.append(contentsOf: suffixes.map { part($0, .suffix) })
        return result
    }

    private static func part(_ value: String, _ type: NamePartType) -> NameValueDraft.Part {
        NameValueDraft.Part(id: UUID(), value: value, type: type.rawValue)
    }
}
