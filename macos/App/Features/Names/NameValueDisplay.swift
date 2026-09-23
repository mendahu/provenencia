import Foundation

/// Compact NameValue summary for list rows and host previews.
enum NameValueDisplay {
    static func string(for draft: NameValueDraft) -> String {
        draft.trimmedForm
    }

    static func string(form: String) -> String {
        form.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}
