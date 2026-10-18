import AppKit

/// Speaks a message through VoiceOver; silent when VoiceOver is off.
enum AccessibilityAnnouncement {
    @MainActor
    static func post(_ message: String, priority: NSAccessibilityPriorityLevel = .medium) {
        NSAccessibility.post(
            element: NSApp as Any,
            notification: .announcementRequested,
            userInfo: [.announcement: message, .priority: priority]
        )
    }
}
