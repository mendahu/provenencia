import AppKit
import SwiftUI

/// What a copy does and how long it shows as done, kept apart from the view so a
/// test can prove the pasteboard write and the timing without the real pasteboard
/// or a real wait.
@MainActor
@Observable
final class PVClipboardCopyState {
    /// How long the check shows before the copy glyph returns.
    static let copiedInterval: Duration = .milliseconds(1600)

    private(set) var copied = false

    @ObservationIgnored private let write: (String) -> Void
    @ObservationIgnored private let sleep: @MainActor (Duration) async -> Void
    @ObservationIgnored private var reset: Task<Void, Never>?

    init(
        write: @escaping (String) -> Void = PVClipboardCopyState.writeToGeneralPasteboard,
        sleep: @escaping @MainActor (Duration) async -> Void = { try? await Task.sleep(for: $0) }
    ) {
        self.write = write
        self.sleep = sleep
    }

    /// Writes `text`, shows the check, and restarts the timer that clears it.
    func copy(_ text: String) {
        write(text)
        copied = true
        reset?.cancel()
        reset = Task { [sleep] in
            await sleep(Self.copiedInterval)
            guard !Task.isCancelled else { return }
            copied = false
        }
    }

    /// Stops the timer when the button goes away.
    func cancel() {
        reset?.cancel()
        reset = nil
        copied = false
    }

    nonisolated static func writeToGeneralPasteboard(_ text: String) {
        let pasteboard = NSPasteboard.general
        pasteboard.clearContents()
        pasteboard.setString(text, forType: .string)
    }
}

/// Provenencia copy-to-clipboard icon button: a `PVIconButton` showing the copy
/// glyph that writes `text` to the pasteboard, turns to a green check for 1.6
/// seconds, and tells VoiceOver `copiedAnnouncement` politely. A click during the
/// check copies again and restarts the timer. The tooltip does not change (the
/// check and the announcement say it was copied).
///
/// `label` names what is copied ("Copy citation"); it is the tooltip and the
/// spoken label. Named for the pasteboard because ``PVCopy`` already means a
/// component's text.
struct PVClipboardButton: View {
    private let text: String
    private let label: any PVCopy
    private let copiedAnnouncement: String
    private let size: PVControlSize

    @State private var state: PVClipboardCopyState

    init(
        text: String,
        label: some PVCopy,
        copiedAnnouncement: String,
        size: PVControlSize = .sm
    ) {
        self.text = text
        self.label = label
        self.copiedAnnouncement = copiedAnnouncement
        self.size = size
        _state = State(initialValue: PVClipboardCopyState())
    }

    var body: some View {
        PVIconButton(
            state.copied ? .check : .copy,
            label: label,
            size: size,
            tone: state.copied ? .success : .neutral
        ) {
            state.copy(text)
            AccessibilityAnnouncement.post(copiedAnnouncement, priority: .low)
        }
        .onDisappear { state.cancel() }
    }
}

#Preview {
    HStack(spacing: PVSpacing.space5) {
        PVClipboardButton(text: "Parish register, p. 12", label: "Copy citation", copiedAnnouncement: "Citation copied")
        PVClipboardButton(text: "Parish register, p. 12", label: "Copy citation", copiedAnnouncement: "Citation copied", size: .md)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
