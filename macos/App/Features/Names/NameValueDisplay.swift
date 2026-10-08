import Foundation

/// Compact NameValue summary for list rows and host previews.
enum NameValueDisplay {
    static func string(for draft: NameValueDraft) -> String {
        draft.trimmedForm
    }

    static func string(form: String) -> String {
        form.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    /// Display text for a structured NameValue. Parts win, joined in the
    /// order they were reconciled (stored `idx`). That order is temporary:
    /// name format profiles will take over cultural ordering
    /// (`structured-name-model` §4). A name with no parts falls back to
    /// `form`. Empty when neither carries text.
    static func string(for name: CatalogNameValue) -> String {
        let parts = name.parts
            .map { $0.value.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
            .joined(separator: " ")
        if !parts.isEmpty { return parts }
        return string(form: name.form)
    }
}
