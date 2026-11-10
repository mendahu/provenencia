import SwiftUI

/// A danger callout for a failed read with one small secondary retry button.
///
/// Pass `failureRevision` (a handle's `QueryHandle.failureRevision`) and the callout speaks
/// again for each new failure, and keeps the keyboard on Try again after a press that failed
/// again. It takes the keyboard only after its own button was pressed, so a failure that
/// arrives some other way does not move it. Without `failureRevision` it behaves as a plain
/// danger callout with a button.
struct PVRetryCallout: View {
    private let title: (any PVCopy)?
    private let message: String
    private let retry: any PVCopy
    private let isRetrying: Bool
    private let compact: Bool
    private let retryIdentifier: String
    private let onRetry: () -> Void
    private let retryFocused: FocusState<Bool>.Binding?
    private let failureRevision: Int?
    @State private var retryPressed = false
    @FocusState private var ownRetryFocus: Bool

    init(
        title: (any PVCopy)? = nil,
        message: String,
        retry: some PVCopy,
        isRetrying: Bool,
        compact: Bool = false,
        retryIdentifier: String,
        retryFocused: FocusState<Bool>.Binding? = nil,
        failureRevision: Int? = nil,
        onRetry: @escaping () -> Void
    ) {
        self.title = title
        self.message = message
        self.retry = retry
        self.isRetrying = isRetrying
        self.compact = compact
        self.retryIdentifier = retryIdentifier
        self.retryFocused = retryFocused
        self.failureRevision = failureRevision
        self.onRetry = onRetry
    }

    var body: some View {
        PVCallout(tone: .danger, title: title, message: message, compact: compact, announces: true) {
            retryButton
        }
        // A new failure remounts the callout so it is announced again.
        .id(failureRevision)
        .onAppear(perform: claimFocusAfterPressedRetry)
    }

    private var focusBinding: FocusState<Bool>.Binding { retryFocused ?? $ownRetryFocus }

    private var retryButton: some View {
        PVButton(retry, variant: .secondary, size: .sm, loading: isRetrying, action: pressRetry)
            .accessibilityIdentifier(retryIdentifier)
            .focused(focusBinding)
    }

    private func pressRetry() {
        if failureRevision != nil { retryPressed = true }
        onRetry()
    }

    /// The callout came back after a failure its own Try again caused: put the keyboard on Try again.
    private func claimFocusAfterPressedRetry() {
        guard retryPressed else { return }
        FocusClaim.claim(focusBinding, true) { retryPressed = false }
    }
}

#Preview {
    PVRetryCallout(
        title: L10n.EvidenceGraph.readErrorTitle,
        message: "Could not read the graph.",
        retry: L10n.EvidenceGraph.readAgain,
        isRetrying: false,
        retryIdentifier: "preview.readAgain",
        onRetry: {}
    )
    .padding()
}

#Preview("Speaks again per failure") {
    PVRetryCallout(
        title: L10n.EvidenceGraph.readErrorTitle,
        message: "Could not read the graph.",
        retry: L10n.EvidenceGraph.readAgain,
        isRetrying: false,
        retryIdentifier: "preview.readAgain.revision",
        failureRevision: 2,
        onRetry: {}
    )
    .padding()
}
