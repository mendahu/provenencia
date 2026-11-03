import SwiftUI

/// The "Show N more" row that ends a folded list on a Conclusion page: a plain
/// borderless button. Return and Space activate it.
struct ConclusionShowMoreRow: View {
    let count: Int
    let expand: () -> Void

    var body: some View {
        Button(action: expand) {
            Text(verbatim: L10n.Conclusions.showMore(count))
        }
        .buttonStyle(.pv(.ghost, size: .sm))
        .pvActivatesOnReturn(expand)
    }
}
