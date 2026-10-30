import SwiftUI

/// The sizes and thresholds of `PVPageColumns`, pure so they are tested
/// without a window. Not generic, so a caller can reach them without naming
/// the page's three view types.
enum PVPageColumnsMetrics {
    /// The id of the zero-height view at the very top of the scroll content.
    static let topAnchor = "pv.pageColumns.top"
    /// The scroll view's coordinate space, in which the top sentinel is measured.
    static let scrollSpace = "pv.pageColumns.scroll"
    /// A pinned header compacts only when the content overflows its viewport
    /// by at least this much: more than the full header is taller than the
    /// compact one, so compacting cannot shrink the bar enough to undo itself.
    static let compactMinOverflow: CGFloat = 120
    /// How far the content scrolls before a pinned header compacts.
    static let compactOffset: CGFloat = 8

    /// The primary column, the secondary column (nil when there is none) and
    /// the group that holds them, for a pane `pane` wide.
    ///
    /// Regular, or Wide with no secondary: one column up to
    /// `widthContentMax` inside `gutterPage`; a secondary stacks below at the
    /// same width. Wide with a secondary: the primary keeps `widthContentMax`
    /// while the secondary takes what is left, between `widthInspector` and
    /// `widthSecondaryMax`.
    static func columnWidths(
        pane: CGFloat,
        widthClass: PVWidthClass,
        hasSecondary: Bool
    ) -> (primary: CGFloat, secondary: CGFloat?, group: CGFloat) {
        if widthClass == .regular || !hasSecondary {
            let primary = max(0, min(PVSpacing.widthContentMax, pane - 2 * PVSpacing.gutterPage))
            return (primary, hasSecondary ? primary : nil, primary)
        }
        let available = pane - 2 * PVSpacing.gutterPage - PVSpacing.space9
        let primary = max(0, min(PVSpacing.widthContentMax, available - PVSpacing.widthInspector))
        let secondary = min(
            PVSpacing.widthSecondaryMax,
            max(PVSpacing.widthInspector, available - primary)
        )
        return (primary, secondary, primary + PVSpacing.space9 + secondary)
    }

    /// Whether a pinned header shows compact. Hysteresis: past
    /// `compactOffset` it compacts, back at the top it expands, and in
    /// between it keeps its state. Content that barely overflows never compacts.
    static func isScrolled(offset: CGFloat, overflow: CGFloat, wasScrolled: Bool) -> Bool {
        guard overflow >= compactMinOverflow else { return false }
        if offset > compactOffset { return true }
        if offset <= 0 { return false }
        return wasScrolled
    }
}

/// The scaffold of a record page: a header, a primary column and an optional
/// secondary column, centered in the pane (§ 9 of macos-client-patterns.md,
/// brief S14-D1). At Regular the columns stack in one column up to 1180pt; at
/// Wide a secondary sits beside the primary. It owns the scroll view and its
/// reset, reads the width class from `\.pvWidthClass` (measured once by
/// `WorkspaceDestinationHost`) and sets no background: the page owns that.
///
/// The header scrolls with the page by default. `headerPinned` holds it above
/// the scroll with a rule under it, and hands the builder `isScrolled` so it
/// can compact. `scrollResetKey` sends the scroll back to the top when it
/// changes (another section of the same page); a width-class change is not a
/// reset. The primary comes before the secondary in the view tree, so focus
/// moves primary then secondary.
struct PVPageColumns<Header: View, Primary: View, Secondary: View>: View {
    private let headerPinned: Bool
    private let scrollResetKey: AnyHashable?
    private let header: (_ isScrolled: Bool) -> Header
    private let primary: Primary
    private let secondary: Secondary

    @Environment(\.pvWidthClass) private var widthClass
    @State private var offset: CGFloat = 0
    @State private var contentHeight: CGFloat = 0
    @State private var viewportHeight: CGFloat = 0
    @State private var isScrolled = false

    init(
        headerPinned: Bool = false,
        scrollResetKey: AnyHashable? = nil,
        @ViewBuilder header: @escaping (_ isScrolled: Bool) -> Header,
        @ViewBuilder primary: () -> Primary,
        @ViewBuilder secondary: () -> Secondary
    ) {
        self.headerPinned = headerPinned
        self.scrollResetKey = scrollResetKey
        self.header = header
        self.primary = primary()
        self.secondary = secondary()
    }

    private var hasHeader: Bool { Header.self != EmptyView.self }
    private var hasSecondary: Bool { Secondary.self != EmptyView.self }

    var body: some View {
        if headerPinned && hasHeader {
            VStack(spacing: 0) {
                PVPageColumnsBandLayout(widthClass: widthClass, hasSecondary: hasSecondary) {
                    header(isScrolled)
                }
                PVColor.borderDefault.frame(height: 1)
                scroll(measures: true) {
                    columns.padding(.vertical, PVSpacing.space8)
                }
            }
        } else {
            scroll(measures: false) {
                VStack(spacing: 0) {
                    if hasHeader {
                        PVPageColumnsBandLayout(widthClass: widthClass, hasSecondary: hasSecondary) {
                            header(false)
                        }
                        Color.clear.frame(height: PVSpacing.space9)
                    }
                    columns
                }
                .padding(.vertical, PVSpacing.space8)
            }
        }
    }

    private var columns: some View {
        PVPageColumnsLayout(widthClass: widthClass) {
            VStack(alignment: .leading, spacing: 0) { primary }
            if hasSecondary {
                VStack(alignment: .leading, spacing: 0) { secondary }
            }
        }
    }

    private func scroll<Content: View>(
        measures: Bool,
        @ViewBuilder content: () -> Content
    ) -> some View {
        ScrollView {
            VStack(spacing: 0) {
                Color.clear
                    .frame(height: 0)
                    .id(PVPageColumnsMetrics.topAnchor)
                    .onGeometryChange(for: CGFloat.self) {
                        -$0.frame(in: .named(PVPageColumnsMetrics.scrollSpace)).minY
                    } action: {
                        guard measures else { return }
                        offset = $0
                        refreshScrolled()
                    }
                content()
            }
            .frame(maxWidth: .infinity)
            .onGeometryChange(for: CGFloat.self) { $0.size.height } action: {
                guard measures else { return }
                contentHeight = $0
                refreshScrolled()
            }
        }
        .coordinateSpace(name: PVPageColumnsMetrics.scrollSpace)
        .onGeometryChange(for: CGFloat.self) { $0.size.height } action: {
            guard measures else { return }
            viewportHeight = $0
            refreshScrolled()
        }
        .pvScrollsRowIntoView(
            Optional<String>.none,
            restartsAt: PVPageColumnsMetrics.topAnchor,
            whenChanged: scrollResetKey
        )
    }

    private func refreshScrolled() {
        let next = PVPageColumnsMetrics.isScrolled(
            offset: offset,
            overflow: contentHeight - viewportHeight,
            wasScrolled: isScrolled
        )
        if next != isScrolled { isScrolled = next }
    }
}

extension PVPageColumns where Secondary == EmptyView {
    init(
        headerPinned: Bool = false,
        scrollResetKey: AnyHashable? = nil,
        @ViewBuilder header: @escaping (_ isScrolled: Bool) -> Header,
        @ViewBuilder primary: () -> Primary
    ) {
        self.init(
            headerPinned: headerPinned,
            scrollResetKey: scrollResetKey,
            header: header,
            primary: primary,
            secondary: { EmptyView() }
        )
    }
}

extension PVPageColumns where Header == EmptyView, Secondary == EmptyView {
    init(
        scrollResetKey: AnyHashable? = nil,
        @ViewBuilder primary: () -> Primary
    ) {
        self.init(
            headerPinned: false,
            scrollResetKey: scrollResetKey,
            header: { _ in EmptyView() },
            primary: primary,
            secondary: { EmptyView() }
        )
    }
}

/// Places the primary and, when there is one, the secondary column: side by
/// side at Wide, stacked at Regular, the group centered in the pane. Sizes
/// itself from the proposed width alone, so it holds no state and crossing the
/// Wide breakpoint keeps the page's scroll and focus.
private struct PVPageColumnsLayout: Layout {
    let widthClass: PVWidthClass

    private func widths(_ proposal: ProposedViewSize, _ subviews: Subviews)
        -> (primary: CGFloat, secondary: CGFloat?, group: CGFloat, pane: CGFloat) {
        let proposed = proposal.width.flatMap { $0.isFinite ? $0 : nil }
        let probe = PVPageColumnsMetrics.columnWidths(
            pane: proposed ?? (PVSpacing.widthContentMax + 2 * PVSpacing.gutterPage),
            widthClass: widthClass,
            hasSecondary: subviews.count > 1
        )
        return (probe.primary, probe.secondary, probe.group, proposed ?? probe.group)
    }

    private func sideBySide(_ subviews: Subviews) -> Bool {
        widthClass == .wide && subviews.count > 1
    }

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        guard !subviews.isEmpty else { return .zero }
        let w = widths(proposal, subviews)
        let primary = subviews[0].sizeThatFits(ProposedViewSize(width: w.primary, height: nil))
        var height = primary.height
        if let secondaryWidth = w.secondary {
            let secondary = subviews[1].sizeThatFits(ProposedViewSize(width: secondaryWidth, height: nil))
            height = sideBySide(subviews)
                ? max(primary.height, secondary.height)
                : primary.height + PVSpacing.space9 + secondary.height
        }
        return CGSize(width: w.pane, height: height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        guard !subviews.isEmpty else { return }
        let w = widths(ProposedViewSize(width: bounds.width, height: nil), subviews)
        let x = bounds.minX + (bounds.width - w.group) / 2
        let primary = subviews[0].sizeThatFits(ProposedViewSize(width: w.primary, height: nil))
        subviews[0].place(
            at: CGPoint(x: x, y: bounds.minY),
            anchor: .topLeading,
            proposal: ProposedViewSize(width: w.primary, height: primary.height)
        )
        guard let secondaryWidth = w.secondary else { return }
        let secondary = subviews[1].sizeThatFits(ProposedViewSize(width: secondaryWidth, height: nil))
        let origin = sideBySide(subviews)
            ? CGPoint(x: x + w.primary + PVSpacing.space9, y: bounds.minY)
            : CGPoint(x: x, y: bounds.minY + primary.height + PVSpacing.space9)
        subviews[1].place(
            at: origin,
            anchor: .topLeading,
            proposal: ProposedViewSize(width: secondaryWidth, height: secondary.height)
        )
    }
}

/// Sizes the header to the column group's width and centers it, so the header
/// lines up with the columns below it.
private struct PVPageColumnsBandLayout: Layout {
    let widthClass: PVWidthClass
    let hasSecondary: Bool

    private func group(_ proposal: ProposedViewSize) -> (pane: CGFloat, group: CGFloat) {
        let proposed = proposal.width.flatMap { $0.isFinite ? $0 : nil }
        let probe = PVPageColumnsMetrics.columnWidths(
            pane: proposed ?? (PVSpacing.widthContentMax + 2 * PVSpacing.gutterPage),
            widthClass: widthClass,
            hasSecondary: hasSecondary
        )
        return (proposed ?? probe.group, probe.group)
    }

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        guard let child = subviews.first else { return .zero }
        let w = group(proposal)
        let size = child.sizeThatFits(ProposedViewSize(width: w.group, height: nil))
        return CGSize(width: w.pane, height: size.height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        guard let child = subviews.first else { return }
        let w = group(ProposedViewSize(width: bounds.width, height: nil))
        let size = child.sizeThatFits(ProposedViewSize(width: w.group, height: nil))
        child.place(
            at: CGPoint(x: bounds.minX + (bounds.width - w.group) / 2, y: bounds.minY),
            anchor: .topLeading,
            proposal: ProposedViewSize(width: w.group, height: size.height)
        )
    }
}

#Preview("Page columns, header scrolls") {
    PVPageColumns(
        header: { _ in Text(verbatim: "Header").font(PVFont.display(size: PVTypeScale.h2)) },
        primary: {
            VStack(alignment: .leading) {
                ForEach(0..<30, id: \.self) { Text(verbatim: "Row \($0)") }
            }
        }
    )
    .frame(width: 900, height: 500)
    .background(PVColor.surfacePage)
}

#Preview("Page columns, header pinned") {
    PVPageColumns(
        headerPinned: true,
        header: { isScrolled in
            Text(verbatim: isScrolled ? "Compact" : "Header")
                .padding(.vertical, isScrolled ? PVSpacing.space5 : PVSpacing.space9)
        },
        primary: {
            VStack(alignment: .leading) {
                ForEach(0..<30, id: \.self) { Text(verbatim: "Row \($0)") }
            }
        }
    )
    .frame(width: 900, height: 500)
    .background(PVColor.surfacePage)
}
