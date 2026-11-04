import AppKit

/// Speaks a message through VoiceOver; silent when VoiceOver is off.
enum AccessibilityAnnouncement {
    /// What a title and a message say together: the non-empty parts, joined
    /// through the catalog; nil when there is nothing to say.
    static func text(title: String?, message: String?) -> String? {
        L10n.DesignSystem.spokenJoin([title, message].compactMap { $0 })
    }

    /// A failure interrupts; a result waits its turn.
    static func priority(danger: Bool) -> NSAccessibilityPriorityLevel {
        danger ? .high : .medium
    }

    @MainActor
    static func post(_ message: String, priority: NSAccessibilityPriorityLevel = .medium) {
        NSAccessibility.post(
            element: NSApp as Any,
            notification: .announcementRequested,
            userInfo: [.announcement: message, .priority: priority]
        )
    }
}
