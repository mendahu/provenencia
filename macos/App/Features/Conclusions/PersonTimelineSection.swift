import SwiftUI

/// The Person page's Timeline section (S13-13, board S13-D4): the person's
/// events earliest first beside a lifeline drawn to scale from birth to death.
/// It reads its own query (`ConclusionPageSection.queryKey`), keeps the list and
/// the line drawn while it reloads, opens an event on its page and a
/// participant on their Timeline.
struct PersonTimelineSection: View {
    let entityId: String

    @Environment(WorkspaceSession.self) private var session
    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(\.pvWidthClass) private var widthClass

    /// The pages of each list the researcher has unfolded. Reset when the page is another person.
    @State private var folds = ConclusionFoldState<PersonTimelineContent.Fold>()

    var body: some View {
        let orders = session.nameFormat
        let folds = folds
        let widthClass = widthClass
        ConclusionPageSectionHost<CatalogPersonTimeline, PersonTimelineContent, PersonTimelineList>(
            section: .timeline,
            entityId: entityId,
            errorTitle: L10n.Conclusions.timelineErrorTitle,
            emptyTitle: L10n.Conclusions.timelineEmptyTitle,
            emptyMessage: L10n.string(L10n.Conclusions.timelineEmptyMessage),
            make: { PersonTimelineContent(timeline: $0, folds: folds, locale: .autoupdatingCurrent, orders: orders, widthClass: widthClass) },
            isEmpty: \.isEmpty
        ) { content in
            PersonTimelineList(
                content: content,
                projectDir: session.projectKey.projectDir,
                open: { navigation.go(to: $0) },
                expand: { self.folds.revealPage($0) },
                expandThrough: { self.folds.reveal($0, pages: $1) }
            )
        }
    }
}

/// The items beside the lifeline, each placed as near its event's point as the
/// rows above allow (`PersonTimelinePlacement`). Folds bound it to about 60
/// items until unfolded.
struct PersonTimelineList: View {
    let content: PersonTimelineContent
    let projectDir: String
    let open: (WorkspaceLocation) -> Void
    let expand: (PersonTimelineContent.Fold) -> Void
    /// Unfolds a list through a page (the fewest that show a row a click on the lifeline names).
    let expandThrough: (PersonTimelineContent.Fold, Int) -> Void

    @Environment(\.pvPageViewportSize) private var viewportSize
    @Environment(\.pvWidthClass) private var widthClass
    /// Each item's whole height, measured.
    @State private var heights: [String: CGFloat] = [:]
    @State private var placement = PersonTimelinePlacement(content: PersonTimelineContent(timeline: CatalogPersonTimeline()), heights: [:], viewportHeight: 0)
    /// The scale once every item is measured, kept while the items and the
    /// viewport stay as they were, so an unfold moves only the rows below it.
    @State private var chosenScale: ChosenScale?
    /// The first item an unfold added, to scroll into view once it is placed.
    @State private var pendingReveal: String?
    /// Keeps the keyboard on a card across a width change or a reload that takes it.
    @State private var keeper = FocusKeeper<String>()
    /// The items' measured width, after the line column.
    @State private var itemsWidth: CGFloat = 0
    @State private var scrollTarget: String?
    @State private var scrollRequest = 0
    /// The title button or card with the keyboard: a row's event id or a card's id.
    @FocusState private var focused: String?

    private struct ChosenScale: Equatable {
        var ids: [String]
        var viewportHeight: CGFloat
        var pointsPerYear: CGFloat
    }

    /// Reads no scroll state: the page's scroll moves only `PersonTimelineTethers`.
    var body: some View {
        let items = content.items
        HStack(alignment: .top, spacing: 0) {
            Color.clear.frame(width: PersonTimelineLifeline.columnWidth + PersonTimelineLifeline.listGap)
            PersonTimelineStack(
                tops: items.map { placement.tops[$0.id] ?? 0 },
                height: placement.height
            ) {
                ForEach(Array(items.enumerated()), id: \.element.id) { index, item in
                    PersonTimelineItemView(
                        item: item,
                        previous: index > 0 ? items[index - 1] : nil,
                        projectDir: projectDir,
                        stacksPeople: PersonTimelinePlacement.stacksPeople(itemsWidth: itemsWidth, widthClass: widthClass),
                        open: open,
                        revealed: content.firstRevealed(by:),
                        expand: unfold,
                        focused: $focused
                    )
                    .id(item.id)
                    .accessibilityIdentifier("persons.detail.timeline.item.\(item.id)")
                    .onGeometryChange(for: CGFloat.self) { $0.size.height } action: { height in
                        if abs((heights[item.id] ?? -1) - height) > 0.01 { heights[item.id] = height }
                    }
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .onGeometryChange(for: CGFloat.self) { $0.size.width } action: { itemsWidth = $0 }
            .pvScrollsRowIntoView(scrollTarget, anchor: .top, whenRequested: scrollRequest)
        }
        .overlay(alignment: .topLeading) {
            if let lifeline = placement.lifeline {
                ZStack(alignment: .topLeading) {
                    PersonTimelineLine(lifeline: lifeline, height: placement.height, labels: content.pointLabels, select: select)
                    PersonTimelineTethers(lifeline: lifeline, anchors: placement.anchors, spans: placement.spans, height: placement.height)
                }
            }
        }
        .accessibilityElement(children: .contain)
        .focusKeeper(
            $keeper,
            focus: $focused,
            isShown: { content.isShown($0) },
            fallback: { content.row(holdingCard: $0) },
            reveal: { $0 }
        )
        .onChange(of: heights) { rebuild() }
        .onChange(of: content) { rebuild() }
        .onChange(of: viewportSize.height) { rebuild() }
        .onAppear { rebuild() }
    }

    private func rebuild() {
        let ids = content.items.map(\.id)
        let idSet = Set(ids)
        let current = heights.filter { idSet.contains($0.key) }
        if current.count != heights.count { heights = current }
        let fixed = chosenScale.flatMap { $0.ids == ids && $0.viewportHeight == viewportSize.height ? $0.pointsPerYear : nil }
        let next = PersonTimelinePlacement(content: content, heights: current, viewportHeight: viewportSize.height, pointsPerYear: fixed)
        if next != placement { placement = next }
        // The scale is kept only once every item is measured.
        let scale = current.count == ids.count && !ids.isEmpty
            ? ChosenScale(ids: ids, viewportHeight: viewportSize.height, pointsPerYear: next.pointsPerYear) : nil
        if scale != chosenScale { chosenScale = scale }
        if let target = pendingReveal, current[target] != nil {
            pendingReveal = nil
            scrollTarget = target
            scrollRequest += 1
        }
    }

    /// "Show N more": the Show more row moves the keyboard to the first item it adds; a
    /// fold of rows also scrolls it into view once placed.
    private func unfold(_ fold: PersonTimelineContent.Fold) {
        if case .participants = fold {} else { pendingReveal = content.firstRevealed(by: fold) }
        expand(fold)
    }

    /// A click on a point or a cap: the list scrolls the item to the top, again on
    /// a second click on the same point. A folded row's point unfolds the dated list
    /// through its page first.
    private func select(_ id: String) {
        if let pages = content.pages(toShow: id) { expandThrough(.dated, pages) }
        scrollTarget = id
        scrollRequest += 1
    }
}

/// Places each subview at the top `PersonTimelinePlacement` gave it, full width.
private struct PersonTimelineStack: Layout {
    /// In subview order.
    var tops: [CGFloat]
    var height: CGFloat

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        CGSize(width: proposal.width ?? subviews.map { $0.sizeThatFits(.unspecified).width }.max() ?? 0, height: height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let proposed = ProposedViewSize(width: bounds.width, height: nil)
        for (index, subview) in subviews.enumerated() {
            let top = index < tops.count ? tops[index] : 0
            subview.place(at: CGPoint(x: bounds.minX, y: bounds.minY + top), anchor: .topLeading, proposal: proposed)
        }
    }
}

/// One item of the list: an event row, a placeholder marker, the Undated heading
/// or a "Show N more" row. It owns the divider above and the paddings, which
/// `PersonTimelinePlacement` also reads for its anchors.
private struct PersonTimelineItemView: View {
    let item: PersonTimelineContent.Item
    let previous: PersonTimelineContent.Item?
    let projectDir: String
    let stacksPeople: Bool
    let open: (WorkspaceLocation) -> Void
    /// The id of the first item the next page of a fold adds, from the content that was drawn.
    let revealed: (PersonTimelineContent.Fold) -> String?
    let expand: (PersonTimelineContent.Fold) -> Void
    let focused: FocusState<String?>.Binding

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if PersonTimelinePlacement.hasDivider(item, after: previous) { PVDivider() }
            content
                .padding(.top, PersonTimelinePlacement.topPadding(item, after: previous))
                .padding(.bottom, PersonTimelinePlacement.bottomPadding)
        }
    }

    @ViewBuilder
    private var content: some View {
        switch item {
        case .row(let row):
            PersonTimelineEventRow(row: row, isUndated: row.span == nil, projectDir: projectDir, stacksPeople: stacksPeople, open: open, revealed: revealed, expand: expand, focused: focused)
        case .marker(let marker):
            if let row = marker.row {
                PersonTimelineEventRow(row: row, isUndated: false, projectDir: projectDir, stacksPeople: stacksPeople, open: open, revealed: revealed, expand: expand, focused: focused)
            } else {
                Text(verbatim: marker.spokenLabel)
                    .font(PVFont.body(size: PVTypeScale.body, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        case .undatedHeader:
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(L10n.Conclusions.timelineUndated)
                    .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .accessibilityAddTraits(.isHeader)
                Text(L10n.Conclusions.timelineUndatedHint)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        case .more(let fold, let count):
            ConclusionShowMoreRow(
                hidden: count,
                fold: PersonTimelineContent.eventFold,
                focus: focused,
                revealed: revealed(fold)
            ) { expand(fold) }
        }
    }
}

/// The event row (D-6): the age flag's gutter, the picture, the title and date
/// line with the person's role Badges, and at the right the other people as
/// compact cards with their roles inside. A Timeline snowflake: one caller.
private struct PersonTimelineEventRow: View {
    let row: PersonTimelineContent.Row
    /// A row with no date: its flag is dashed.
    let isUndated: Bool
    let projectDir: String
    /// The list is too narrow for the people beside the text: they go under it.
    let stacksPeople: Bool
    let open: (WorkspaceLocation) -> Void
    let revealed: (PersonTimelineContent.Fold) -> String?
    let expand: (PersonTimelineContent.Fold) -> Void
    let focused: FocusState<String?>.Binding

    @Environment(\.pvWidthClass) private var widthClass

    private static let picture = PersonTimelinePlacement.rowPicture
    private static let gap = PersonTimelinePlacement.rowGap
    var body: some View {
        HStack(alignment: .top, spacing: Self.gap) {
            Color.clear
                .frame(width: PersonTimelinePlacement.gutterWidth, height: PersonTimelinePlacement.flagInset + PersonTimelinePlacement.flagHeight)
                .overlay(alignment: .topLeading) {
                    PersonTimelineAgeFlag(row: row, isDashed: isUndated)
                        .padding(.leading, PersonTimelinePlacement.flagInset)
                        .padding(.top, PersonTimelinePlacement.flagInset)
                }
            CachedThumbnail(
                projectDir: projectDir,
                relPath: row.imageRelPath,
                size: Self.picture,
                cornerRadius: PVRadius.sm,
                fallbackMark: .subjectEvent,
                fallbackLabel: L10n.Conclusions.eventPictureThumb,
                colors: nil
            )
            .accessibilityHidden(true)
            VStack(alignment: .leading, spacing: 4) {
                Button { open(row.location) } label: {
                    PVTitleText(row.title, style: .heading, lineLimit: 2)
                        .multilineTextAlignment(.leading)
                }
                .buttonStyle(PVFocusRingButtonStyle())
                .focused(focused, equals: row.id)
                .pvActivatesOnReturn(when: focused.wrappedValue == row.id) { open(row.location) }
                .accessibilityIdentifier("persons.detail.timeline.title.\(row.id)")
                HStack(spacing: 8) {
                    ForEach(Array(row.roles.enumerated()), id: \.offset) { _, role in
                        PVBadge(text: role.text, tone: .neutral, subtle: role.isSubject)
                    }
                    .accessibilityHidden(true)
                    EventSecondaryLine(content: row.line, dateSize: PVTypeScale.body)
                }
                if stacksPeople && !row.participants.isEmpty { people.padding(.top, 4) }
            }
            .frame(maxWidth: PersonTimelinePlacement.textCap, alignment: .leading)
            Spacer(minLength: 0)
            if !stacksPeople && !row.participants.isEmpty { people }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(row.spokenLabel)
    }

    private var people: some View {
        VStack(alignment: .leading, spacing: 4) {
            if widthClass == .wide {
                LazyVGrid(
                    columns: Array(repeating: GridItem(.flexible(), spacing: PersonTimelinePlacement.cardColumnGap, alignment: .topLeading), count: 3),
                    alignment: .leading,
                    spacing: 4
                ) {
                    ForEach(row.participants) { card($0) }
                }
            } else {
                ForEach(row.participants) { card($0) }
            }
            if row.participantsHidden > 0 {
                ConclusionShowMoreRow(
                    hidden: row.participantsHidden,
                    fold: PersonTimelineContent.participantFold(widthClass),
                    focus: focused,
                    revealed: revealed(.participants(eventId: row.id))
                ) {
                    expand(.participants(eventId: row.id))
                }
                .accessibilityIdentifier("persons.detail.timeline.more.\(row.id)")
            }
        }
        .frame(width: PersonTimelinePlacement.peopleWidth(widthClass), alignment: .leading)
    }

    private func card(_ link: PersonCardLink) -> some View {
        PVPersonCard(content: link.content, projectDir: projectDir, size: .compact) {
            if let location = link.location { open(location) }
        }
        .frame(height: PVTileCardSize.compact.height)
        .focused(focused, equals: link.id)
        .accessibilityIdentifier("persons.detail.timeline.tile.\(link.id)")
    }
}

/// The age at the event (TL-5c) in a small flag at the end of the row's curve; a
/// dashed one for a row with no date. A Timeline snowflake: one caller.
private struct PersonTimelineAgeFlag: View {
    let row: PersonTimelineContent.Row
    let isDashed: Bool

    var body: some View {
        let shape = RoundedRectangle(cornerRadius: 3)
        Text(verbatim: row.ageText)
            .font(PVFont.mono(size: PVTypeScale.micro))
            .foregroundStyle(PVColor.textSecondary)
            .frame(width: PersonTimelinePlacement.flagWidth, height: PersonTimelinePlacement.flagHeight)
            .background(shape.fill(PVColor.surfaceCard))
            .overlay(shape.strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: isDashed ? [3, 3] : [])))
            .help(row.ageHelp)
            .accessibilityHidden(true)
    }
}

/// The lifeline's static drawing and its clickable points: the line, ticks and
/// labels, stubs, points and caps. It reads only the lifeline, so a scroll does
/// not redraw it. Hidden from VoiceOver and the keyboard; the rows carry the meaning.
private struct PersonTimelineLine: View {
    let lifeline: PersonTimelineLifeline
    let height: CGFloat
    /// Each point's name, shown on hover.
    let labels: [String: String]
    let select: (String) -> Void

    private var lineX: CGFloat { PersonTimelineLifeline.lineX }

    var body: some View {
        ZStack(alignment: .topLeading) {
            Canvas { context, _ in
                context.translateBy(x: 0, y: PersonTimelineDrawing.pad)
                draw(into: &context)
            }
            .frame(
                width: PersonTimelineLifeline.columnWidth + 4,
                height: height + 2 * PersonTimelineDrawing.pad,
                alignment: .topLeading
            )
            .offset(y: -PersonTimelineDrawing.pad)
            .allowsHitTesting(false)
            ForEach(lifeline.marks, id: \.id) { mark in
                Button { select(mark.id) } label: { Color.clear }
                    .buttonStyle(.plain)
                    .focusable(false)
                    .frame(width: PersonTimelineDrawing.hit, height: hitHeight(mark))
                    .contentShape(Rectangle())
                    .offset(x: lineX - PersonTimelineDrawing.hit / 2, y: top(mark) - 6)
                    .help(labels[mark.id] ?? "")
                    .accessibilityIdentifier("persons.detail.timeline.point.\(mark.id)")
            }
        }
        .frame(width: PersonTimelineLifeline.columnWidth, alignment: .topLeading)
        .accessibilityHidden(true)
        .accessibilityIdentifier("persons.detail.timeline.line")
    }

    private func top(_ mark: PersonTimelineLifeline.Mark) -> CGFloat {
        switch mark.shape {
        case .dot(let y), .cap(let y, _): y
        case .bar(let from, _, _, _): from
        }
    }

    private func hitHeight(_ mark: PersonTimelineLifeline.Mark) -> CGFloat {
        switch mark.shape {
        case .bar(let from, let to, _, _): max(12, to - from + 12)
        case .dot, .cap: 12
        }
    }

    private func draw(into context: inout GraphicsContext) {
        let line = PVColor.borderStrong
        let dash = StrokeStyle(lineWidth: 1.5, dash: [3, 3])
        // The line, solid except over the stubs.
        var solidFrom = PersonTimelineLifeline.pad
        let end = lifeline.length - PersonTimelineLifeline.pad
        for stub in lifeline.stubs.sorted(by: { $0.lowerBound < $1.lowerBound }) {
            let lower = stub.lowerBound, upper = stub.upperBound
            if lower > solidFrom {
                context.fill(Path(CGRect(x: lineX - 0.75, y: solidFrom, width: 1.5, height: lower - solidFrom)), with: .color(line))
            }
            var dashed = Path()
            dashed.move(to: CGPoint(x: lineX, y: lower))
            dashed.addLine(to: CGPoint(x: lineX, y: upper))
            context.stroke(dashed, with: .color(line), style: dash)
            solidFrom = max(solidFrom, upper)
        }
        if end > solidFrom {
            context.fill(Path(CGRect(x: lineX - 0.75, y: solidFrom, width: 1.5, height: end - solidFrom)), with: .color(line))
        }
        // Ticks and year labels.
        for tick in lifeline.ticks {
            let y = tick.y
            let length: CGFloat = tick.isQuarter ? 4 : 8
            context.fill(Path(CGRect(x: lineX - length, y: y - 0.5, width: length, height: 1)), with: .color(line))
            if let year = tick.year {
                let label = Text(verbatim: String(year))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                context.draw(context.resolve(label), at: CGPoint(x: lineX - 8 - 4, y: y), anchor: .trailing)
            }
        }
        // Points and caps, all at rest; the rows in view are redrawn in the accent above.
        for mark in lifeline.marks {
            PersonTimelineDrawing.draw(mark, color: PVColor.borderStrong, into: &context)
        }
    }
}

/// The part of the lifeline the page's scroll moves: the points of the rows in
/// view drawn in the accent, and a curve from each to its row's flag. It alone
/// reads where the viewport sits over the section, so a scroll step redraws only
/// this.
private struct PersonTimelineTethers: View {
    let lifeline: PersonTimelineLifeline
    let anchors: [String: CGFloat]
    let spans: [String: ClosedRange<CGFloat>]
    let height: CGFloat

    var body: some View {
        PVPageVisibleRect { visible, _ in
            tethersBody(visible: visible)
        }
    }

    private func tethersBody(visible: CGRect) -> some View {
        let tethers = lifeline.tethers(anchors: anchors, spans: spans, visible: visible)
        return Color.clear
            .frame(width: PersonTimelineLifeline.columnWidth)
            .frame(height: height)
            .overlay(alignment: .topLeading) {
                Canvas { context, _ in
                    context.translateBy(x: 0, y: PersonTimelineDrawing.pad)
                    draw(tethers, into: &context)
                }
                .frame(
                    width: PersonTimelineTethers.canvasWidth,
                    height: height + 2 * PersonTimelineDrawing.pad,
                    alignment: .topLeading
                )
                .offset(y: -PersonTimelineDrawing.pad)
            }
            .allowsHitTesting(false)
            .accessibilityHidden(true)
    }

    /// Wide enough for the curve to reach the flag.
    private static let canvasWidth = PersonTimelineLifeline.flagX

    private func draw(_ tethers: [PersonTimelineLifeline.Tether], into context: inout GraphicsContext) {
        for tether in tethers {
            if let mark = lifeline.mark(tether.id) {
                PersonTimelineDrawing.draw(mark, color: PVColor.accent, into: &context)
            }
            // A curve from the point to the edge of the row's flag.
            var curve = Path()
            let midX = (tether.from.x + tether.to.x) / 2
            curve.move(to: tether.from)
            curve.addCurve(
                to: tether.to,
                control1: CGPoint(x: midX, y: tether.from.y),
                control2: CGPoint(x: midX, y: tether.to.y)
            )
            context.stroke(curve, with: .color(PVColor.accent), lineWidth: 1)
        }
    }
}

/// What the line's two layers share: their sizes and how one point or cap is drawn.
private enum PersonTimelineDrawing {
    /// Room above and below the line for a cap's radius.
    static let pad: CGFloat = 8
    static let hit: CGFloat = 24

    static func draw(
        _ mark: PersonTimelineLifeline.Mark,
        color: Color,
        into context: inout GraphicsContext
    ) {
        let lineX = PersonTimelineLifeline.lineX
        switch mark.shape {
        case .dot(let y):
            let c = CGPoint(x: lineX, y: y)
            context.fill(Path(ellipseIn: CGRect(x: c.x - 4, y: c.y - 4, width: 8, height: 8)), with: .color(color))
        case .bar(let from, let to, let openStart, let openEnd):
            let top = from, bottom = max(to, top + 6)
            let rect = CGRect(x: lineX - 3, y: top, width: 6, height: bottom - top)
            let fade = min(16, rect.height)
            let shading: GraphicsContext.Shading = .linearGradient(
                Gradient(stops: [
                    .init(color: openStart ? color.opacity(0) : color, location: 0),
                    .init(color: color, location: openStart ? fade / rect.height : 0),
                    .init(color: color, location: openEnd ? 1 - fade / rect.height : 1),
                    .init(color: openEnd ? color.opacity(0) : color, location: 1),
                ]),
                startPoint: CGPoint(x: lineX, y: rect.minY),
                endPoint: CGPoint(x: lineX, y: rect.maxY)
            )
            context.fill(Path(roundedRect: rect, cornerRadius: 3), with: shading)
        case .cap(let y, let isPlaceholder):
            let c = CGPoint(x: lineX, y: y)
            let circle = Path(ellipseIn: CGRect(x: c.x - 6, y: c.y - 6, width: 12, height: 12))
            if isPlaceholder {
                context.fill(circle, with: .color(PVColor.surfacePage))
                context.stroke(circle, with: .color(PVColor.textMuted), style: StrokeStyle(lineWidth: 1, dash: [3, 3]))
                let mark = Text(verbatim: "?")
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                context.draw(context.resolve(mark), at: c, anchor: .center)
            } else {
                context.fill(circle, with: .color(color))
            }
        }
    }
}

#if DEBUG
/// A Person page's Timeline over a fixture, at a pane width, with its folds working.
private struct PersonTimelinePreview: View {
    let timeline: CatalogPersonTimeline
    var width: CGFloat = 1102
    var foldsAtStart = ConclusionFoldState<PersonTimelineContent.Fold>()

    @State private var folds = ConclusionFoldState<PersonTimelineContent.Fold>()

    var body: some View {
        PVPageColumns(primaryFill: true) {
            PersonTimelineList(
                content: PersonTimelineContent(timeline: timeline, folds: folds, widthClass: width >= PVSpacing.widthWideBreakpoint ? .wide : .regular),
                projectDir: "/tmp/none",
                open: { _ in },
                expand: { folds.revealPage($0) },
                expandThrough: { folds.reveal($0, pages: $1) }
            )
        }
        .environment(\.pvWidthClass, width >= PVSpacing.widthWideBreakpoint ? .wide : .regular)
        .frame(width: width, height: 900)
        .background(PVColor.surfacePage)
        .onAppear { folds = foldsAtStart }
    }
}

#Preview("Full timeline") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.johann) }
#Preview("Full timeline, Wide") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.johann, width: 2296) }
#Preview("Unknown ends") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.unknownEnds) }
#Preview("Short life") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.ernst) }
#Preview("Row sheet") {
    PersonTimelinePreview(timeline: {
        var timeline = PersonTimelinePreviewData.johann
        let sheet = Set(["4", "5", "8", "11", "12", "13"])
        timeline.events = timeline.events.filter { sheet.contains($0.id) }
        timeline.birth = CatalogLifeMarker(besideEventId: "4")
        timeline.death = CatalogLifeMarker(besideEventId: "11")
        return timeline
    }())
}
#Preview("Large event") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.largeEvent) }
#Preview("Large event, unfolded") {
    PersonTimelinePreview(timeline: PersonTimelinePreviewData.largeEvent, foldsAtStart: {
        var folds = ConclusionFoldState<PersonTimelineContent.Fold>()
        folds.revealPage(.participants(eventId: "big"))
        return folds
    }())
}
#Preview("Long life") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.lise) }
#Preview("Only undated") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.ida) }
#endif
