import SwiftUI

/// A danger callout for a failed read with one small secondary retry button.
struct PVRetryCallout: View {
    private let title: (any PVCopy)?
    private let message: String
    private let retry: any PVCopy
    private let isRetrying: Bool
    private let compact: Bool
    private let retryIdentifier: String
    private let onRetry: () -> Void
    private let retryFocused: FocusState<Bool>.Binding?

    init(
        title: (any PVCopy)? = nil,
        message: String,
        retry: some PVCopy,
        isRetrying: Bool,
        compact: Bool = false,
        retryIdentifier: String,
        retryFocused: FocusState<Bool>.Binding? = nil,
        onRetry: @escaping () -> Void
    ) {
        self.title = title
        self.message = message
        self.retry = retry
        self.isRetrying = isRetrying
        self.compact = compact
        self.retryIdentifier = retryIdentifier
        self.retryFocused = retryFocused
        self.onRetry = onRetry
    }

    var body: some View {
        PVCallout(tone: .danger, title: title, message: message, compact: compact, announces: true) {
            retryButton
        }
    }

    @ViewBuilder private var retryButton: some View {
        let button = PVButton(retry, variant: .secondary, size: .sm, loading: isRetrying, action: onRetry)
            .accessibilityIdentifier(retryIdentifier)
        if let retryFocused {
            button.focused(retryFocused)
        } else {
            button
        }
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
