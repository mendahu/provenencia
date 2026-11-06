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

/// The one fold rule of a Conclusion page's lists: the first `limit` of a list behind a
/// "Show N more" row, or all of it once expanded.
enum ConclusionFold {
    /// All of `list` when `expanded` or when folding would hide fewer than `minimumHidden`
    /// items, else its first `limit`; `hidden` is how many the fold leaves out.
    static func shown<T>(
        _ list: [T],
        limit: Int,
        minimumHidden: Int,
        expanded: Bool
    ) -> (shown: [T], hidden: Int) {
        if expanded || list.count - limit < minimumHidden { return (list, 0) }
        return (Array(list.prefix(limit)), list.count - limit)
    }

    /// The first item an unfold adds: `list[limit]` when folding would hide at least
    /// `minimumHidden`, expanded or not; else nil.
    static func firstRevealed<T>(_ list: [T], limit: Int, minimumHidden: Int) -> T? {
        list.count - limit >= minimumHidden ? list[limit] : nil
    }
}
