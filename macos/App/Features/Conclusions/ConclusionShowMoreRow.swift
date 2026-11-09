import SwiftUI

/// The "Show N more" row that ends a folded list on a Conclusion page: a plain
/// borderless button that words the next page (`fold.nextPageCount`) and, after a
/// key press, moves the keyboard to the first item the page reveals. Return and
/// Space activate it; a pointer click leaves focus where it is.
struct ConclusionShowMoreRow<Focus: Hashable>: View {
    /// How many items the fold leaves out now (the total, not the page).
    let hidden: Int
    let fold: ConclusionFold
    /// The section's focus state, where `revealed` is a value.
    let focus: FocusState<Focus?>.Binding
    /// The focus id of the first item the page adds, read from the content that
    /// was drawn; nil when there is nothing to move focus to.
    let revealed: Focus?
    let expand: () -> Void

    @FocusState private var hasKeyboard: Bool

    var body: some View {
        Button(action: press) {
            Text(verbatim: L10n.Conclusions.showMore(fold.nextPageCount(hidden: hidden)))
        }
        .buttonStyle(.pv(.ghost, size: .sm))
        .focused($hasKeyboard)
        .pvActivatesOnReturn(when: hasKeyboard, press)
    }

    private func press() {
        let claims = hasKeyboard
        expand()
        if claims, let revealed { FocusClaim.claim(focus, revealed) }
    }
}

/// The one fold rule of a Conclusion page's lists: the first `limit` of a list behind a
/// "Show N more" row, then `pageSize` more at a time. A fold that would leave fewer than
/// `minimumHidden` items hidden shows them all. Each content declares its folds once.
struct ConclusionFold: Hashable {
    /// How many items one "Show more" reveals.
    static let pageSize = 50

    let limit: Int
    let minimumHidden: Int

    /// How many items show after `pages` pages: all of them when the rest would be
    /// fewer than `minimumHidden`.
    func shownCount(count: Int, pages: Int) -> Int {
        let target = limit + max(pages, 0) * Self.pageSize
        return count - target < minimumHidden ? count : target
    }

    /// The first items of `list` at `pages` pages; `hidden` is how many the fold leaves out.
    func shown<T>(_ list: [T], pages: Int) -> (shown: [T], hidden: Int) {
        let count = shownCount(count: list.count, pages: pages)
        return (Array(list.prefix(count)), list.count - count)
    }

    /// How many items the next "Show more" adds: a whole page, or what is left when
    /// folding the rest would hide fewer than `minimumHidden`.
    func nextPageCount(hidden: Int) -> Int {
        hidden - Self.pageSize < minimumHidden ? hidden : Self.pageSize
    }

    /// The first item the next page adds, or nil when the fold hides nothing.
    func firstRevealed<T>(_ list: [T], pages: Int) -> T? {
        let count = shownCount(count: list.count, pages: pages)
        return count < list.count ? list[count] : nil
    }

    /// The fewest pages p >= 0 that show `list[index]` of a list of `count`.
    func pages(toShow index: Int, count: Int) -> Int {
        var pages = 0
        while index >= shownCount(count: count, pages: pages) && shownCount(count: count, pages: pages) < count {
            pages += 1
        }
        return pages
    }
}

/// How many pages of each fold a section has revealed. Reset when the page is another
/// entity. `Hashable` so a derived value can key on it.
struct ConclusionFoldState<Fold: Hashable>: Hashable {
    private var revealed: [Fold: Int] = [:]

    /// Pages revealed; 0 when never.
    func pages(_ fold: Fold) -> Int { revealed[fold] ?? 0 }

    mutating func revealPage(_ fold: Fold) { revealed[fold, default: 0] += 1 }

    /// Reveals through `pages`, keeping the larger of that and what is revealed.
    mutating func reveal(_ fold: Fold, pages: Int) {
        guard pages > self.pages(fold) else { return }
        revealed[fold] = pages
    }
}
