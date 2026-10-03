import Foundation

/// Compact NameValue summary for list rows and host previews.
enum NameValueDisplay {
    static func string(for draft: NameValueDraft) -> String {
        draft.trimmedForm
    }

    static func string(form: String) -> String {
        form.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    /// Display text for a structured NameValue from the engine. There is no
    /// name format profile yet (Spike 9 R2), so the full-form reading wins;
    /// a blank form falls back to the parts in order. Empty when neither
    /// carries text.
    static func string(for name: CatalogNameValue) -> String {
        let form = string(form: name.form)
        if !form.isEmpty { return form }
        return name.parts
            .map { $0.value.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
            .joined(separator: " ")
    }
}
