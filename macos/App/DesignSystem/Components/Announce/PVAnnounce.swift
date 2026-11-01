import AppKit
import SwiftUI

/// Speaks a text through VoiceOver when the view appears and again whenever the
/// text changes to a different non-nil value. Kit components that show a result
/// or a failure (`PVToast`, `PVCallout(announces:)`, `PVErrorText(announces:)`)
/// use it, so no call site posts its own announcement.
///
/// The same text again while the view stays mounted is not spoken again.
private struct PVAnnounceModifier: ViewModifier {
    let text: String?
    let priority: NSAccessibilityPriorityLevel

    func body(content: Content) -> some View {
        content
            .onAppear { post(text) }
            .onChange(of: text) { _, new in post(new) }
    }

    private func post(_ text: String?) {
        guard let text else { return }
        AccessibilityAnnouncement.post(text, priority: priority)
    }
}

extension View {
    /// Speaks `text` (nothing when nil) when the view appears or the text changes.
    func pvAnnounce(_ text: String?, priority: NSAccessibilityPriorityLevel = .medium) -> some View {
        modifier(PVAnnounceModifier(text: text, priority: priority))
    }
}
