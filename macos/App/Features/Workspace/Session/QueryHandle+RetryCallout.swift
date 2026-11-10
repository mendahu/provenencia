import SwiftUI

extension PVRetryCallout {
    /// A Conclusions page's read-failure notice for one handle: Try again shows the handle's
    /// reading state and the notice speaks again for each new failure. `id` roots the Try again
    /// identifier (`"\(id).tryAgain"`); the caller puts `id` on the notice.
    init<Value>(
        title: some PVCopy,
        message: String,
        handle: QueryHandle<Value>,
        compact: Bool = false,
        id: String,
        onRetry: @escaping () -> Void
    ) {
        self.init(
            title: title,
            message: message,
            retry: L10n.Conclusions.readNoticeTryAgain,
            isRetrying: handle.isReading,
            compact: compact,
            retryIdentifier: "\(id).tryAgain",
            failureRevision: handle.failureRevision,
            onRetry: onRetry
        )
    }
}
