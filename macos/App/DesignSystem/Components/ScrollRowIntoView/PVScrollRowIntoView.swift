import SwiftUI

/// Keeps the keyboard's row in view in a scrolling list. When `row` changes,
/// the enclosing `ScrollView` scrolls to the row tagged `.id(row)`; with no
/// `anchor` it moves only as far as it must, so a row already on screen (the
/// one the pointer hovers, say) does not move the list.
///
/// Apply it to the `ScrollView` or to the rows inside one. Each row carries
/// `.id(_:)` with the value the caller passes as `row`. `PVTable`, `PVList`,
/// `PVSelect`, `PVComboBox` and the omnibar results all scroll through it.
struct PVScrollRowIntoView<ID: Hashable>: ViewModifier {
    let row: ID?
    let anchor: UnitPoint?
    let animated: Bool
    let scrollsOnAppear: Bool
    let rows: [ID]?
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    func body(content: Content) -> some View {
        ScrollViewReader { proxy in
            content
                .onChange(of: row, initial: scrollsOnAppear) { _, row in
                    guard let row else { return }
                    withAnimation(animated && !reduceMotion ? PVMotion.fastStandard : nil) {
                        proxy.scrollTo(row, anchor: anchor)
                    }
                }
                .onChange(of: rows) { _, rows in
                    // A new list starts at the top. Without this the scroller
                    // keeps the offset the last list left, hiding its first row.
                    guard let first = rows?.first else { return }
                    proxy.scrollTo(first, anchor: .top)
                }
        }
    }
}

extension View {
    /// Scrolls `row` into view whenever it changes. See `PVScrollRowIntoView`.
    ///
    /// - Parameters:
    ///   - row: The id of the highlighted or selected row, or nil for none.
    ///   - anchor: Where the row lands; nil scrolls only as far as needed.
    ///   - animated: Animates the scroll with `PVMotion.fastStandard`, never
    ///     under Reduce Motion.
    ///   - scrollsOnAppear: Also scrolls when the list first appears, for a
    ///     menu that opens on its current choice.
    ///   - rows: When given, a change in these ids (a new query's results)
    ///     scrolls back to the first row.
    func pvScrollsRowIntoView<ID: Hashable>(
        _ row: ID?,
        anchor: UnitPoint? = nil,
        animated: Bool = true,
        scrollsOnAppear: Bool = false,
        startsAtTopWhenRowsChange rows: [ID]? = nil
    ) -> some View {
        modifier(PVScrollRowIntoView(
            row: row,
            anchor: anchor,
            animated: animated,
            scrollsOnAppear: scrollsOnAppear,
            rows: rows
        ))
    }
}

private struct PVScrollRowIntoViewPreview: View {
    @State private var highlighted = 0
    private let rows = Array(0..<30)

    var body: some View {
        VStack(spacing: PVSpacing.space2) {
            HStack {
                Button(action: { highlighted = max(0, highlighted - 1) }) { Text(verbatim: "↑") }
                Button(action: { highlighted = min(rows.count - 1, highlighted + 1) }) { Text(verbatim: "↓") }
            }
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(rows, id: \.self) { row in
                        Text(verbatim: "Row \(row)")
                            .padding(PVSpacing.space2)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .background(row == highlighted ? PVColor.surfaceSelected : .clear)
                            .id(row)
                    }
                }
            }
            .frame(height: 160)
            .pvScrollsRowIntoView(highlighted)
        }
        .padding()
        .frame(width: 240)
    }
}

#Preview("Scroll row into view") {
    PVScrollRowIntoViewPreview()
}
