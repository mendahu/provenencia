import SwiftUI

/// The sizes and thresholds of `PVPageColumns`, pure so they are tested
/// without a window. Not generic, so a caller can reach them without naming
/// the page's three view types.
enum PVPageColumnsMetrics {
    /// The id of the zero-height view at the very top of the scroll content.
    static let topAnchor = "pv.pageColumns.top"
    /// The scroll view's coordinate space, for a section that measures its rows in it.
    static let scrollSpace = "pv.pageColumns.scroll"
    /// The padding above and below the page's columns, and the distance a pinned
    /// column keeps from the top of the visible area.
    static let verticalPadding: CGFloat = PVSpacing.space8

    /// How far a side column is pushed down so it stays `inset` below the top of the
    /// visible area: 0 when it is taller than the visible area less the two insets (it
    /// then scrolls with the page), else the visible top plus `inset`, held so the column's
    /// bottom never passes the section's. All in the section's own coordinates.
    static func pinOffset(
        visibleMinY: CGFloat,
        visibleHeight: CGFloat,
        sectionHeight: CGFloat,
        sideHeight: CGFloat,
        inset: CGFloat
    ) -> CGFloat {
        guard sideHeight <= visibleHeight - 2 * inset else { return 0 }
        return max(0, min(visibleMinY + inset, sectionHeight - sideHeight))
    }

    /// The primary column, the secondary column (nil when there is none), the
    /// group that holds them and the band the header takes, for a pane `pane`
    /// wide.
    ///
    /// Regular, or Wide with no secondary: one column up to
    /// `widthContentMax` inside `gutterPage`; a secondary stacks below at the
    /// same width. Wide with a secondary: the primary keeps `widthContentMax`
    /// while the secondary takes what is left, between `widthInspector` and
    /// `widthSecondaryMax`. The band is the group, except in full-width mode.
    ///
    /// `primaryFill` (full-width mode) lets the primary fill the pane between
    /// the gutters at Wide when there is no secondary; the header band keeps
    /// `widthContentMax`. Ignored at Regular and with a secondary.
    static func columnWidths(
        pane: CGFloat,
        widthClass: PVWidthClass,
        hasSecondary: Bool,
        primaryFill: Bool = false
    ) -> (primary: CGFloat, secondary: CGFloat?, group: CGFloat, band: CGFloat) {
        if primaryFill && widthClass == .wide && !hasSecondary {
            let primary = max(0, pane - 2 * PVSpacing.gutterPage)
            return (primary, nil, primary, min(PVSpacing.widthContentMax, primary))
        }
        if widthClass == .regular || !hasSecondary {
            let primary = max(0, min(PVSpacing.widthContentMax, pane - 2 * PVSpacing.gutterPage))
            return (primary, hasSecondary ? primary : nil, primary, primary)
        }
        let split = split(available: pane - 2 * PVSpacing.gutterPage, sideMax: PVSpacing.widthSecondaryMax)
        return (split.main, split.side, split.group, split.group)
    }

    /// A main column beside a side column in `available` width: the main keeps up to
    /// `widthContentMax` while the side takes what is left, between `widthInspector`
    /// and `sideMax`. `group` is the two and the gap between them.
    static func split(available: CGFloat, sideMax: CGFloat) -> (main: CGFloat, side: CGFloat, group: CGFloat) {
        let room = available - PVSpacing.space9
        let main = max(0, min(PVSpacing.widthContentMax, room - PVSpacing.widthInspector))
        let side = min(sideMax, max(PVSpacing.widthInspector, room - main))
        return (main, side, main + PVSpacing.space9 + side)
    }

    /// The widths of a main column and a side column inside `available` (the section's
    /// width between the gutters), at any width class. Beside each other while the main
    /// column keeps `mainMin` and the side column keeps `widthInspector`; stacked, the side
    /// column under the main at its own width up to `widthSideColumnMax`, when it does not.
    static func sideColumnWidths(available: CGFloat, mainMin: CGFloat) -> PVPageSideColumnWidths {
        if available < mainMin + PVSpacing.space9 + PVSpacing.widthInspector {
            return PVPageSideColumnWidths(
                main: available,
                side: min(available, PVSpacing.widthSideColumnMax),
                group: available,
                stacked: true
            )
        }
        let split = split(available: available, sideMax: PVSpacing.widthSideColumnMax)
        return PVPageSideColumnWidths(main: split.main, side: split.side, group: split.group, stacked: false)
    }

    /// The pane and the column widths for a layout proposed `proposal`. An
    /// unspecified or infinite width is sized as if the pane were as wide as
    /// the content and its gutters; the pane is then the group itself, and
    /// `primaryFill` is ignored.
    static func columnWidths(
        proposal: ProposedViewSize,
        widthClass: PVWidthClass,
        hasSecondary: Bool,
        primaryFill: Bool = false
    ) -> (pane: CGFloat, primary: CGFloat, secondary: CGFloat?, group: CGFloat, band: CGFloat) {
        let proposed = proposal.width.flatMap { $0.isFinite ? $0 : nil }
        let widths = columnWidths(
            pane: proposed ?? (PVSpacing.widthContentMax + 2 * PVSpacing.gutterPage),
            widthClass: widthClass,
            hasSecondary: hasSecondary,
            primaryFill: primaryFill && proposed != nil
        )
        return (proposed ?? widths.group, widths.primary, widths.secondary, widths.group, widths.band)
    }
}

/// The widths of a main column and its side column, and whether the side column
/// sits under the main instead of beside it.
struct PVPageSideColumnWidths: Equatable {
    let main: CGFloat
    /// Nil for a page with no side column.
    let side: CGFloat?
    let group: CGFloat
    let stacked: Bool

    init(main: CGFloat, side: CGFloat?, group: CGFloat, stacked: Bool) {
        self.main = main
        self.side = side
        self.group = group
        self.stacked = stacked
    }
}

private struct PVPageViewportSizeKey: EnvironmentKey {
    static let defaultValue: CGSize = .zero
}

extension EnvironmentValues {
    /// The size of the page's scroll view, published by `PVPageColumns` for a
    /// section that needs to know what is in view. `.zero` outside a
    /// `PVPageColumns`; it changes on resize, not on scroll.
    var pvPageViewportSize: CGSize {
        get { self[PVPageViewportSizeKey.self] }
        set { self[PVPageViewportSizeKey.self] = newValue }
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
/// the scroll with a rule under it, in its full form at every scroll
/// position. `scrollResetKey` sends the scroll back to the top when it
/// changes (another section of the same page); a width-class change is not a
/// reset. The primary comes before the secondary in the view tree, so focus
/// moves primary then secondary.
///
/// `primaryFill` is an opt-in full-width mode: at Wide, with no secondary, the
/// primary fills the pane between the gutters while the header keeps its
/// `widthContentMax` band. The section inside keeps its own text measure.
struct PVPageColumns<Header: View, Primary: View, Secondary: View>: View {
    private let headerPinned: Bool
    private let scrollResetKey: AnyHashable?
    private let primaryFill: Bool
    private let header: () -> Header
    private let primary: Primary
    private let secondary: Secondary

    @Environment(\.pvWidthClass) private var widthClass
    @State private var viewportSize: CGSize = .zero

    init(
        headerPinned: Bool = false,
        scrollResetKey: AnyHashable? = nil,
        primaryFill: Bool = false,
        @ViewBuilder header: @escaping () -> Header,
        @ViewBuilder primary: () -> Primary,
        @ViewBuilder secondary: () -> Secondary
    ) {
        self.headerPinned = headerPinned
        self.scrollResetKey = scrollResetKey
        self.primaryFill = primaryFill
        self.header = header
        self.primary = primary()
        self.secondary = secondary()
    }

    private var hasHeader: Bool { Header.self != EmptyView.self }
    private var hasSecondary: Bool { Secondary.self != EmptyView.self }

    var body: some View {
        if headerPinned && hasHeader {
            VStack(spacing: 0) {
                PVPageColumnsBandLayout(widthClass: widthClass, hasSecondary: hasSecondary, primaryFill: primaryFill) {
                    header()
                }
                PVColor.borderDefault.frame(height: 1)
                scroll {
                    columns.padding(.vertical, PVPageColumnsMetrics.verticalPadding)
                }
            }
        } else {
            scroll {
                VStack(spacing: 0) {
                    if hasHeader {
                        PVPageColumnsBandLayout(widthClass: widthClass, hasSecondary: hasSecondary, primaryFill: primaryFill) {
                            header()
                        }
                        Color.clear.frame(height: PVSpacing.space9)
                    }
                    columns
                }
                .padding(.vertical, PVPageColumnsMetrics.verticalPadding)
            }
        }
    }

    private var columns: some View {
        let (widthClass, hasSecondary, primaryFill) = (widthClass, hasSecondary, primaryFill)
        return PVPageColumnsLayout(widths: { proposal in
            let w = PVPageColumnsMetrics.columnWidths(
                proposal: proposal, widthClass: widthClass, hasSecondary: hasSecondary, primaryFill: primaryFill
            )
            return PVPageSideColumnWidths(
                main: w.primary,
                side: w.secondary,
                group: w.group,
                stacked: hasSecondary && widthClass != .wide
            )
        }) {
            VStack(alignment: .leading, spacing: 0) { primary }
            if hasSecondary {
                VStack(alignment: .leading, spacing: 0) { secondary }
            }
        }
    }

    private func scroll<Content: View>(
        @ViewBuilder content: () -> Content
    ) -> some View {
        ScrollView {
            VStack(spacing: 0) {
                Color.clear
                    .frame(height: 0)
                    .id(PVPageColumnsMetrics.topAnchor)
                content()
            }
            .frame(maxWidth: .infinity)
        }
        .coordinateSpace(name: PVPageColumnsMetrics.scrollSpace)
        .onGeometryChange(for: CGSize.self) { $0.size } action: {
            viewportSize = $0
        }
        .environment(\.pvPageViewportSize, viewportSize)
        .pvScrollsRowIntoView(
            Optional<String>.none,
            restartsAt: PVPageColumnsMetrics.topAnchor,
            whenChanged: scrollResetKey
        )
    }
}

extension PVPageColumns where Secondary == EmptyView {
    init(
        headerPinned: Bool = false,
        scrollResetKey: AnyHashable? = nil,
        primaryFill: Bool = false,
        @ViewBuilder header: @escaping () -> Header,
        @ViewBuilder primary: () -> Primary
    ) {
        self.init(
            headerPinned: headerPinned,
            scrollResetKey: scrollResetKey,
            primaryFill: primaryFill,
            header: header,
            primary: primary,
            secondary: { EmptyView() }
        )
    }
}

extension PVPageColumns where Header == EmptyView, Secondary == EmptyView {
    init(
        scrollResetKey: AnyHashable? = nil,
        primaryFill: Bool = false,
        @ViewBuilder primary: () -> Primary
    ) {
        self.init(
            headerPinned: false,
            scrollResetKey: scrollResetKey,
            primaryFill: primaryFill,
            header: { EmptyView() },
            primary: primary,
            secondary: { EmptyView() }
        )
    }
}

/// Hands its content the part of the page's scroll view that is in view, in the
/// content's own coordinates (its top-left is the origin), and the content's size. It
/// holds the reading itself, so a section that needs to know what is in view reads it
/// here and nowhere else. Outside a `PVPageColumns` the visible rectangle is empty.
struct PVPageVisibleRect<Content: View>: View {
    private let content: (_ visible: CGRect, _ size: CGSize) -> Content

    @Environment(\.pvPageViewportSize) private var viewportSize
    /// The content's frame in the page's scroll space: where the viewport sits over it.
    @State private var frame: CGRect = .zero

    init(@ViewBuilder content: @escaping (_ visible: CGRect, _ size: CGSize) -> Content) {
        self.content = content
    }

    var body: some View {
        let visible = CGRect(
            x: -frame.minX, y: -frame.minY, width: viewportSize.width, height: viewportSize.height
        )
        content(visible, frame.size)
            .onGeometryChange(for: CGRect.self) { $0.frame(in: .named(PVPageColumnsMetrics.scrollSpace)) } action: {
                frame = $0
            }
    }
}

/// Places a main column and, when there is one, a side column: side by side, the
/// group centered, or stacked, the side under the main. Sizes itself from the
/// proposed width alone, so it holds no state and crossing a breakpoint keeps the
/// page's scroll and focus. `widths` decides the columns for the proposal: the
/// page's primary and secondary (`columnWidths`), or a section's main and side column
/// (`sideColumnWidths`).
struct PVPageColumnsLayout: Layout {
    let widths: @Sendable (ProposedViewSize) -> PVPageSideColumnWidths

    private func pane(_ proposal: ProposedViewSize, _ w: PVPageSideColumnWidths) -> CGFloat {
        proposal.width.flatMap { $0.isFinite ? $0 : nil } ?? w.group
    }

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        guard !subviews.isEmpty else { return .zero }
        let w = widths(proposal)
        let main = subviews[0].sizeThatFits(ProposedViewSize(width: w.main, height: nil))
        var height = main.height
        if let sideWidth = w.side {
            let side = subviews[1].sizeThatFits(ProposedViewSize(width: sideWidth, height: nil))
            height = w.stacked
                ? main.height + PVSpacing.space9 + side.height
                : max(main.height, side.height)
        }
        return CGSize(width: pane(proposal, w), height: height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        guard !subviews.isEmpty else { return }
        let w = widths(ProposedViewSize(width: bounds.width, height: nil))
        let x = bounds.minX + (bounds.width - w.group) / 2
        let main = subviews[0].sizeThatFits(ProposedViewSize(width: w.main, height: nil))
        subviews[0].place(
            at: CGPoint(x: x, y: bounds.minY),
            anchor: .topLeading,
            proposal: ProposedViewSize(width: w.main, height: main.height)
        )
        guard let sideWidth = w.side else { return }
        let side = subviews[1].sizeThatFits(ProposedViewSize(width: sideWidth, height: nil))
        let origin = w.stacked
            ? CGPoint(x: x, y: bounds.minY + main.height + PVSpacing.space9)
            : CGPoint(x: x + w.main + PVSpacing.space9, y: bounds.minY)
        subviews[1].place(
            at: origin,
            anchor: .topLeading,
            proposal: ProposedViewSize(width: sideWidth, height: side.height)
        )
    }
}

/// Sizes the header to the band's width and centers it, so the header lines
/// up with the columns below it (in full-width mode it keeps the Regular
/// measure while the columns widen).
private struct PVPageColumnsBandLayout: Layout {
    let widthClass: PVWidthClass
    let hasSecondary: Bool
    let primaryFill: Bool

    private func group(_ proposal: ProposedViewSize) -> (pane: CGFloat, group: CGFloat) {
        let widths = PVPageColumnsMetrics.columnWidths(
            proposal: proposal, widthClass: widthClass, hasSecondary: hasSecondary, primaryFill: primaryFill
        )
        return (widths.pane, widths.band)
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
        header: { Text(verbatim: "Header").font(PVFont.display(size: PVTypeScale.h2)) },
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
        header: {
            Text(verbatim: "Header").padding(.vertical, PVSpacing.space9)
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

extension View {
    /// Keeps a full-width section's empty, loading or error state in the page band: centered at
    /// `widthContentMax`, as the Regular primary column is, instead of stretching across a Wide
    /// pane. A no-op at Regular.
    func pvKeepsPageBand() -> some View {
        frame(maxWidth: PVSpacing.widthContentMax)
            .frame(maxWidth: .infinity)
    }
}
