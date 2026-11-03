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

    /// The lists the researcher has unfolded. Reset when the page is another person.
    @State private var expanded: Set<PersonTimelineContent.Fold> = []

    var body: some View {
        let orders = session.nameFormat
        let expanded = expanded
        ConclusionPageSectionHost<CatalogPersonTimeline, PersonTimelineContent, PersonTimelineList>(
            section: .timeline,
            entityId: entityId,
            errorTitle: L10n.Conclusions.timelineErrorTitle,
            emptyTitle: L10n.Conclusions.timelineEmptyTitle,
            emptyMessage: L10n.string(L10n.Conclusions.timelineEmptyMessage),
            make: { PersonTimelineContent(timeline: $0, expanded: expanded, locale: .autoupdatingCurrent, orders: orders) },
            isEmpty: \.isEmpty
        ) { content in
            PersonTimelineList(
                content: content,
                projectDir: session.projectKey.projectDir,
                open: { navigation.go(to: $0) },
                expand: { self.expanded.insert($0) }
            )
        }
        .onChange(of: entityId) { self.expanded = [] }
    }
}

/// The items beside the lifeline. A plain stack, because the line's length and
/// scale need every item's frame; folds bound it to about 60 items until unfolded.
struct PersonTimelineList: View {
    let content: PersonTimelineContent
    let projectDir: String
    let open: (WorkspaceLocation) -> Void
    let expand: (PersonTimelineContent.Fold) -> Void

    /// Each item's frame in the section's own space.
    @State private var itemFrames: [String: CGRect] = [:]
    @State private var lifeline = PersonTimelineLifeline(content: PersonTimelineContent(timeline: CatalogPersonTimeline()), itemFrames: [:])
    @State private var scrollTarget: String?
    @State private var scrollRequest = 0

    private nonisolated static let space = "persons.detail.timeline"

    /// Reads no scroll state: the page's scroll moves only `PersonTimelineTethers`.
    var body: some View {
        HStack(alignment: .top, spacing: 0) {
            Color.clear.frame(width: PersonTimelineLifeline.columnWidth)
            VStack(alignment: .leading, spacing: 0) {
                ForEach(content.items) { item in
                    PersonTimelineItemView(
                        item: item,
                        isLast: item.id == content.items.last?.id,
                        projectDir: projectDir,
                        open: open,
                        expand: expand
                    )
                    .id(item.id)
                    .accessibilityIdentifier("persons.detail.timeline.item.\(item.id)")
                    .onGeometryChange(for: CGRect.self) { $0.frame(in: .named(Self.space)) } action: { frame in
                        if itemFrames[item.id] != frame { itemFrames[item.id] = frame }
                    }
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .pvScrollsRowIntoView(scrollTarget, anchor: .top, whenRequested: scrollRequest)
        }
        .overlay(alignment: .topLeading) {
            ZStack(alignment: .topLeading) {
                PersonTimelineLine(lifeline: lifeline, select: select)
                PersonTimelineTethers(lifeline: lifeline, itemFrames: itemFrames)
            }
        }
        .coordinateSpace(name: Self.space)
        .accessibilityElement(children: .contain)
        .onChange(of: itemFrames) { rebuild() }
        .onChange(of: content) { rebuild() }
        .onAppear { rebuild() }
    }

    private func rebuild() {
        let ids = Set(content.items.map(\.id))
        let current = itemFrames.filter { ids.contains($0.key) }
        if current.count != itemFrames.count { itemFrames = current }
        let next = PersonTimelineLifeline(content: content, itemFrames: current)
        if next != lifeline { lifeline = next }
    }

    /// A click on a point or a cap: the list scrolls the item to the top, again on
    /// a second click on the same point. A folded row's point unfolds the list first.
    private func select(_ id: String) {
        if !content.items.contains(where: { $0.id == id }) { expand(.dated) }
        scrollTarget = id
        scrollRequest += 1
    }
}

/// One item of the list: an event row, a placeholder marker, the Undated heading
/// or a "Show N more" row.
private struct PersonTimelineItemView: View {
    let item: PersonTimelineContent.Item
    let isLast: Bool
    let projectDir: String
    let open: (WorkspaceLocation) -> Void
    let expand: (PersonTimelineContent.Fold) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            content
                .padding(.vertical, PVSpacing.space5)
            if !isLast { PVDivider() }
        }
    }

    @ViewBuilder
    private var content: some View {
        switch item {
        case .row(let row):
            PersonTimelineEventRow(row: row, projectDir: projectDir, open: open, expand: expand)
        case .marker(let marker):
            if let row = marker.row {
                PersonTimelineEventRow(row: row, projectDir: projectDir, open: open, expand: expand)
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
            ConclusionShowMoreRow(count: count) { expand(fold) }
        }
    }
}

/// The event row (D-16): picture, title and date line, the person's role Badges,
/// then the other people as compact cards. A Timeline snowflake: one caller.
private struct PersonTimelineEventRow: View {
    let row: PersonTimelineContent.Row
    let projectDir: String
    let open: (WorkspaceLocation) -> Void
    let expand: (PersonTimelineContent.Fold) -> Void

    private static let picture: CGFloat = 72
    private static let gap: CGFloat = 16

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(alignment: .top, spacing: Self.gap) {
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
                        PVTitleText(row.title, style: .display(compact: true), lineLimit: 2)
                            .multilineTextAlignment(.leading)
                    }
                    .buttonStyle(PVFocusRingButtonStyle())
                    .pvActivatesOnReturn { open(row.location) }
                    .accessibilityIdentifier("persons.detail.timeline.title.\(row.id)")
                    EventSecondaryLine(content: row.line, dateSize: PVTypeScale.body)
                }
                Spacer(minLength: Self.gap)
                HStack(spacing: PVSpacing.space2) {
                    ForEach(Array(row.roles.enumerated()), id: \.offset) { _, role in
                        PVBadge(text: role.text, tone: .neutral, subtle: role.isSubject)
                    }
                }
                .accessibilityHidden(true)
            }
            .frame(maxWidth: PVSpacing.widthContentMax - PersonTimelineLifeline.columnWidth, alignment: .leading)
            if !row.participants.isEmpty {
                VStack(alignment: .leading, spacing: PVSpacing.space3) {
                    LazyVGrid(
                        columns: [GridItem(.adaptive(minimum: PVTileCardSize.compact.minimumWidth), spacing: 8, alignment: .topLeading)],
                        alignment: .leading,
                        spacing: 8
                    ) {
                        ForEach(row.participants) { link in
                            VStack(alignment: .leading, spacing: 4) {
                                if let role = link.content.role {
                                    Text(verbatim: role)
                                        .pvMicroCaps()
                                        .foregroundStyle(PVColor.textMuted)
                                        .accessibilityHidden(true)
                                }
                                PVPersonCard(content: link.content, projectDir: projectDir, size: .compact) {
                                    if let location = link.location { open(location) }
                                }
                                .frame(height: PVTileCardSize.compact.height)
                                .accessibilityIdentifier("persons.detail.timeline.tile.\(link.id)")
                            }
                        }
                    }
                    if row.participantsHidden > 0 {
                        ConclusionShowMoreRow(count: row.participantsHidden) {
                            expand(.participants(eventId: row.id))
                        }
                        .accessibilityIdentifier("persons.detail.timeline.more.\(row.id)")
                    }
                }
                .padding(.leading, Self.picture + Self.gap)
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(row.spokenLabel)
    }
}

/// The lifeline's static drawing and its clickable points: the line, ticks and
/// labels, stubs, points and caps. It reads only the lifeline, so a scroll does
/// not redraw it. Hidden from VoiceOver and the keyboard; the rows carry the meaning.
private struct PersonTimelineLine: View {
    let lifeline: PersonTimelineLifeline
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
                height: lifeline.origin + lifeline.length + 2 * PersonTimelineDrawing.pad,
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
                    .offset(x: lineX - PersonTimelineDrawing.hit / 2, y: lifeline.origin + top(mark) - 6)
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
        var solidFrom: CGFloat = lifeline.origin
        let end = lifeline.origin + lifeline.length
        for stub in lifeline.stubs.sorted(by: { $0.lowerBound < $1.lowerBound }) {
            let lower = lifeline.origin + stub.lowerBound, upper = lifeline.origin + stub.upperBound
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
            let y = lifeline.origin + tick.y
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
            PersonTimelineDrawing.draw(mark, of: lifeline, color: PVColor.borderStrong, into: &context)
        }
    }
}

/// The part of the lifeline the page's scroll moves: the points of the rows in
/// view drawn in the accent, and a curve from each to its row. It alone reads
/// where the viewport sits over the section, so a scroll step redraws only this.
private struct PersonTimelineTethers: View {
    let lifeline: PersonTimelineLifeline
    let itemFrames: [String: CGRect]

    @Environment(\.pvPageViewportSize) private var viewportSize
    /// The section's frame in the page's scroll space: where the viewport sits over it.
    @State private var sectionFrame: CGRect = .zero

    var body: some View {
        let visible = CGRect(
            x: -sectionFrame.minX, y: -sectionFrame.minY, width: viewportSize.width, height: viewportSize.height
        )
        let tethers = lifeline.tethers(itemFrames: itemFrames, visible: visible)
        Color.clear
            .frame(width: PersonTimelineLifeline.columnWidth)
            .frame(maxHeight: .infinity)
            .onGeometryChange(for: CGRect.self) { $0.frame(in: .named(PVPageColumnsMetrics.scrollSpace)) } action: {
                sectionFrame = $0
            }
            .overlay(alignment: .topLeading) {
                Canvas { context, _ in
                    context.translateBy(x: 0, y: PersonTimelineDrawing.pad)
                    draw(tethers, into: &context)
                }
                .frame(
                    width: PersonTimelineLifeline.columnWidth + 4,
                    height: sectionFrame.height + 2 * PersonTimelineDrawing.pad,
                    alignment: .topLeading
                )
                .offset(y: -PersonTimelineDrawing.pad)
            }
            .allowsHitTesting(false)
            .accessibilityHidden(true)
    }

    private func draw(_ tethers: [PersonTimelineLifeline.Tether], into context: inout GraphicsContext) {
        for tether in tethers {
            if let mark = lifeline.mark(tether.id) {
                PersonTimelineDrawing.draw(mark, of: lifeline, color: PVColor.accent, into: &context)
            }
            // A curve from the point to a dot at the row's edge (a dashed dot alone for an undated row).
            let dot = CGRect(x: tether.to.x - 3, y: tether.to.y - 3, width: 6, height: 6)
            if tether.isDashed {
                context.stroke(Path(ellipseIn: dot), with: .color(PVColor.textMuted), style: StrokeStyle(lineWidth: 1, dash: [2, 2]))
                continue
            }
            var curve = Path()
            let midX = (tether.from.x + tether.to.x) / 2
            curve.move(to: tether.from)
            curve.addCurve(
                to: CGPoint(x: tether.to.x - 3, y: tether.to.y),
                control1: CGPoint(x: midX, y: tether.from.y),
                control2: CGPoint(x: midX, y: tether.to.y)
            )
            context.stroke(curve, with: .color(PVColor.accent), lineWidth: 1)
            context.fill(Path(ellipseIn: dot), with: .color(PVColor.accent))
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
        of lifeline: PersonTimelineLifeline,
        color: Color,
        into context: inout GraphicsContext
    ) {
        let lineX = PersonTimelineLifeline.lineX
        switch mark.shape {
        case .dot(let y):
            let c = CGPoint(x: lineX, y: lifeline.origin + y)
            context.fill(Path(ellipseIn: CGRect(x: c.x - 4, y: c.y - 4, width: 8, height: 8)), with: .color(color))
        case .bar(let from, let to, let openStart, let openEnd):
            let top = lifeline.origin + from, bottom = max(lifeline.origin + to, top + 6)
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
            let c = CGPoint(x: lineX, y: lifeline.origin + y)
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
    var expandedAtStart: Set<PersonTimelineContent.Fold> = []

    @State private var expanded: Set<PersonTimelineContent.Fold> = []

    var body: some View {
        PVPageColumns(primaryFill: true) {
            PersonTimelineList(
                content: PersonTimelineContent(timeline: timeline, expanded: expanded),
                projectDir: "/tmp/none",
                open: { _ in },
                expand: { expanded.insert($0) }
            )
        }
        .environment(\.pvWidthClass, width >= PVSpacing.widthWideBreakpoint ? .wide : .regular)
        .frame(width: width, height: 900)
        .background(PVColor.surfacePage)
        .onAppear { expanded = expandedAtStart }
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
    PersonTimelinePreview(timeline: PersonTimelinePreviewData.largeEvent, expandedAtStart: [.participants(eventId: "big")])
}
#Preview("Long life") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.lise) }
#Preview("Only undated") { PersonTimelinePreview(timeline: PersonTimelinePreviewData.ida) }
#endif
