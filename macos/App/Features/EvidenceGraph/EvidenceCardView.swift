import AppKit
import SwiftUI

/// Subject and bridge cards drawn in AppKit.
///
/// The canvas hosts one of these per card. Text measurement
/// here is also the hit layout edges and the pointer read, so a wrapped title
/// or value moves its band with the paint. The placement preview is this view built with ``init(ghostOf:)``.
final class EvidenceCardView: GraphCanvasItemView {
    let layout: EvidenceCardLayout
    private let card: EvidenceCanvasCard
    private let isGhost: Bool
    private var loads: [String: Task<Void, Never>] = [:]

    convenience init(card: EvidenceCanvasCard, actions: [NSAccessibilityCustomAction]) {
        self.init(card: card, actions: actions, isGhost: false)
    }

    /// The placement ghost: the same paint at 62% opacity, with no VoiceOver
    /// element and no actions.
    convenience init(ghostOf card: EvidenceCanvasCard) {
        self.init(card: card, actions: [], isGhost: true)
    }

    private init(card: EvidenceCanvasCard, actions: [NSAccessibilityCustomAction], isGhost: Bool) {
        self.card = card
        self.isGhost = isGhost
        self.layout = Self.measure(card)
        super.init(contentSize: layout.size)
        if isGhost {
            setAccessibilityElement(false)
        } else {
            setAccessibilityElement(true)
            setAccessibilityRole(.group)
            setAccessibilityLabel(card.text.accessibilityLabel)
            let prefix = card.isBridge ? "evidenceGraph.bridge." : "evidenceGraph.subject."
            setAccessibilityIdentifier(prefix + card.id)
            setAccessibilityCustomActions(actions)
            if card.state.isSelected || card.state.isActivated || card.state.isConnectingFrom {
                setAccessibilitySelected(true)
            }
        }
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        if window == nil {
            loads.values.forEach { $0.cancel() }
            loads = [:]
        } else {
            loadMissingLikenesses()
        }
    }

    /// One decode per missing likeness image, calling `invalidatePaint()` when
    /// it lands. A ghost, and a card with no project folder, paint the
    /// placeholder only.
    private func loadMissingLikenesses() {
        guard !isGhost, !card.projectDir.isEmpty else { return }
        let projectDir = card.projectDir
        for relPath in Self.likenessRelPaths(card) where loads[relPath] == nil {
            guard ThumbnailCache.shared.cachedImage(projectDir: projectDir, relPath: relPath) == nil else { continue }
            loads[relPath] = Task { @MainActor [weak self] in
                _ = await ThumbnailCache.shared.image(projectDir: projectDir, relPath: relPath)
                guard !Task.isCancelled else { return }
                self?.invalidatePaint()
            }
        }
    }

    override func paint() {
        let alpha: CGFloat = {
            if isGhost { return 0.62 }
            if card.state.isDragging { return 0.92 }
            if case .subject(let placed, _) = card.placed, !placed.isCited { return 0.76 }
            return 1
        }()
        NSGraphicsContext.saveGraphicsState()
        NSGraphicsContext.current?.cgContext.setAlpha(alpha)
        _ = Self.measure(card, painter: PVPainter())
        NSGraphicsContext.restoreGraphicsState()
    }

    /// The images of the card's likeness rows, in row order.
    private static func likenessRelPaths(_ card: EvidenceCanvasCard) -> [String] {
        let observations: [CatalogObservation] = switch card.placed {
        case .subject(let placed, _): placed.observations
        case .bridge(let placed): EvidenceBridgeCard.extraObservations(for: placed)
        }
        return observations.compactMap(likenessRelPath)
    }

    private static func measure(_ card: EvidenceCanvasCard, painter: PVPainter? = nil) -> EvidenceCardLayout {
        switch card.placed {
        case .subject(let placed, let presentation):
            subject(placed, presentation: presentation, card: card, painter: painter)
        case .bridge(let placed):
            bridge(placed, card: card, painter: painter)
        }
    }
}

// MARK: - Painter

extension PVPainter {
    /// The conflict bracket beside a run of rows that disagree.
    func bracket(_ rect: CGRect) {
        let box = rect
        let x = box.minX + 8
        let top = box.minY + 12
        let bottom = box.minY + max(box.height - 12, 14)
        let path = NSBezierPath()
        path.move(to: NSPoint(x: x + 5, y: top))
        path.line(to: NSPoint(x: x, y: top))
        path.line(to: NSPoint(x: x, y: bottom))
        path.line(to: NSPoint(x: x + 5, y: bottom))
        path.lineWidth = 1.5
        path.lineCapStyle = .round
        path.lineJoinStyle = .round
        Self.color(PVColor.warning).setStroke()
        path.stroke()
    }
}

// MARK: - Layout and paint

private extension EvidenceCardView {
    static func subject(
        _ placed: EvidenceGraphPlacedSubject,
        presentation: CatalogSubjectTypePresentation?,
        card: EvidenceCanvasCard,
        painter: PVPainter?
    ) -> EvidenceCardLayout {
        let style = EvidenceSubjectKindStyle.resolve(typeKey: placed.kind.rawValue, presentation: presentation)
        let width = EvidenceSubjectCard.width
        let state = card.state
        let selected = state.isSelected || state.isActivated || state.isConnectingFrom || state.isDragging
        var regions: [String: CGRect] = [:]
        var y = EvidenceSubjectCard.shellPaddingTop
        let headerHeight = max(EvidenceSubjectCard.headerHeight, titleBlockHeight(placed))
        let headerRect = CGRect(x: 0, y: y, width: width, height: headerHeight)
        y += headerHeight

        if state.isConnectingFrom {
            y += EvidenceSubjectCard.headerToBodyGap
            let line = PVTextMetrics.fontLineHeight(PVFont.nsMono(size: 11))
            painter?.text(
                L10n.string(L10n.EvidenceGraph.connectingFrom),
                in: CGRect(x: EvidenceSubjectCard.shellPaddingX, y: y, width: width - 26, height: line),
                font: PVFont.nsMono(size: 11), color: PVColor.accent, lines: 1
            )
            y += line
        } else {
            let description = placed.subject.description.trimmingCharacters(in: .whitespacesAndNewlines)
            if !description.isEmpty {
                y += EvidenceSubjectCard.headerToBodyGap
                let bodyWidth = width - EvidenceSubjectCard.shellPaddingX * 2
                let font = PVFont.nsBody(size: 12.5, italic: true)
                let height = PVTextMetrics.textHeight(description, width: bodyWidth, font: font, maxLines: 3)
                painter?.text(
                    description,
                    in: CGRect(x: EvidenceSubjectCard.shellPaddingX, y: y, width: bodyWidth, height: height),
                    font: font, color: PVColor.textSecondary, lines: 3
                )
                y += height
            }
        }

        var rowRects: [(CatalogObservation, CGRect)] = []
        let addRect: CGRect
        let footerRect: CGRect
        if placed.isCited {
            y += EvidenceSubjectCard.headerToBodyGap + EvidenceSubjectCard.stackHairline
            let valueFont = PVFont.nsBody(size: 13)
            for observation in placed.observations {
                let rowHeight = propertyRowHeight(observation, valueWidth: subjectValueWidth, valueFont: valueFont, orders: card.nameOrders)
                let rect = CGRect(x: 0, y: y, width: width, height: rowHeight)
                rowRects.append((observation, rect))
                regions[EvidenceSubjectCard.editPropertyActionID(observationID: observation.id)] = rect
                y += rowHeight + EvidenceSubjectCard.stackHairline
            }
            addRect = CGRect(x: 0, y: y, width: width, height: addPropertyRowHeight)
            regions[EvidenceSubjectCard.addPropertyActionID] = addRect
            y += addPropertyRowHeight + EvidenceSubjectCard.stackHairline
            footerRect = recordFooter(placed: placed, y: &y, width: width, regions: &regions)
        } else {
            y += EvidenceSubjectCard.headerToBodyGap
            addRect = CGRect(x: 0, y: y, width: width, height: EvidenceSubjectCard.addPropertyQuietHeight)
            regions[EvidenceSubjectCard.addPropertyActionID] = addRect
            y += EvidenceSubjectCard.addPropertyQuietHeight
                + EvidenceSubjectCard.headerToBodyGap
                + EvidenceSubjectCard.uncitedFooterTopMargin
                + EvidenceSubjectCard.uncitedFooterRule
            footerRect = recordFooter(placed: placed, y: &y, width: width, regions: &regions)
        }

        paintSubjectShell(placed: placed, style: style, selected: selected, dragging: state.isDragging, height: y, painter: painter)
        let cardRect = CGRect(x: 0, y: 0, width: width, height: y)
        painter?.beginClip(cardRect, radius: PVRadius.md)
        paintHeader(
            title: placed.titleOrTypeLabel, typeLine: typeLine(placed),
            chipFill: style.chip, chipInk: style.ink,
            chipLine: placed.isCited ? style.line : PVColor.borderDefault,
            mark: placed.kind.markKey, chipSize: 28, markSize: 15, citedChip: placed.isCited,
            header: headerRect, width: width, paddingX: EvidenceSubjectCard.shellPaddingX,
            titleCited: placed.isCited,
            deleteHovered: state.hoveredActionID == EvidenceSubjectCard.deleteActionID,
            painter: painter
        )
        if placed.isCited, let first = rowRects.first, let last = rowRects.last {
            painter?.fill(
                CGRect(x: 0, y: first.1.minY - 1, width: width, height: last.1.maxY - first.1.minY + 1),
                style.line
            )
        }
        paintRows(rowRects, tint: style.tint, chip: style.chip, ink: style.ink, hovered: state.hoveredActionID, width: width, valueWidth: subjectValueWidth, leading: EvidenceCitedPropertyMarks.leadingInset(shell: EvidenceSubjectCard.shellPaddingX), trailing: EvidenceSubjectCard.shellPaddingX, orders: card.nameOrders, projectDir: card.projectDir, painter: painter)
        paintAdd(addRect, quiet: !placed.isCited, canCite: state.canCite, hovered: state.hoveredActionID == EvidenceSubjectCard.addPropertyActionID, tint: style.tint, chip: style.chip, ink: style.ink, paddingX: EvidenceSubjectCard.shellPaddingX, painter: painter)
        if !placed.isCited {
            painter?.fill(
                CGRect(x: 0, y: footerRect.minY - EvidenceSubjectCard.uncitedFooterRule, width: width, height: EvidenceSubjectCard.uncitedFooterRule),
                style.line
            )
        }
        paintFooter(footerRect, placed: placed, style: style, hovered: state.hoveredActionID, pressed: state.pressedActionID, orders: card.nameOrders, painter: painter)
        strokeCard(
            cardRect,
            color: selected ? PVColor.accent : (placed.isCited ? style.line : PVColor.borderDefault),
            width: selected ? 1.5 : 1,
            dash: (placed.isCited || selected) ? [] : [4, 3],
            painter: painter
        )
        painter?.endClip()
        return EvidenceCardLayout(size: CGSize(width: width, height: y), regions: regions)
    }

    static func bridge(_ placed: EvidenceGraphPlacedBridge, card: EvidenceCanvasCard, painter: PVPainter?) -> EvidenceCardLayout {
        let width = EvidenceBridgeCard.width
        let state = card.state
        let selected = state.isSelected || state.isActivated || state.isDragging
        let body = placed.isCited ? (card.text.summary ?? "") : L10n.string(L10n.EvidenceGraph.bridgeHonestyBody)
        let bodyFont = placed.isCited
            ? PVFont.nsDisplay(size: 14.5, weight: PVFontWeight.medium)
            : PVFont.nsBody(size: 11.5, italic: true)
        let bodyHeight = PVTextMetrics.textHeight(body, width: bridgeBodyWidth, font: bodyFont, maxLines: 3)
        let headerHeight = placed.isCited
            ? EvidenceBridgeCard.headerContentHeightCited
            : max(EvidenceBridgeCard.headerContentHeightUncited, bridgeTitleBlock(placed))
        var y = EvidenceBridgeCard.shellPaddingTop
        let headerRect = CGRect(x: 0, y: y, width: width, height: headerHeight)
        y += headerHeight + EvidenceBridgeCard.headerToBodySpacing
        let bodyRect = CGRect(x: EvidenceBridgeCard.shellPaddingX, y: y, width: bridgeBodyWidth, height: bodyHeight)
        var regions: [String: CGRect] = [:]
        if state.canCite {
            let iconX = width - EvidenceBridgeCard.shellPaddingX - 12 - (placed.isCited ? 0 : 18)
            regions[EvidenceBridgeCard.editCitationActionID] = CGRect(x: iconX, y: y + 2, width: 12, height: 12)
        }
        y += bodyHeight + EvidenceBridgeCard.headerToBodySpacing
        var rowRects: [(CatalogObservation, CGRect)] = []
        let addRect: CGRect
        if placed.isCited {
            y += EvidenceSubjectCard.stackHairline
            let valueFont = PVFont.nsBody(size: 13)
            for observation in EvidenceBridgeCard.extraObservations(for: placed) {
                let rowHeight = propertyRowHeight(observation, valueWidth: bridgeValueWidth, valueFont: valueFont, orders: card.nameOrders)
                let rect = CGRect(x: 0, y: y, width: width, height: rowHeight)
                rowRects.append((observation, rect))
                regions[EvidenceSubjectCard.editPropertyActionID(observationID: observation.id)] = rect
                y += rowHeight + EvidenceSubjectCard.stackHairline
            }
            addRect = CGRect(x: 0, y: y, width: width, height: addPropertyRowHeight)
            regions[EvidenceSubjectCard.addPropertyActionID] = addRect
            y += addPropertyRowHeight
        } else {
            addRect = CGRect(x: 0, y: y, width: width, height: 26)
            regions[EvidenceSubjectCard.addPropertyActionID] = addRect
            y += 26 + EvidenceBridgeCard.shellPaddingTop
        }

        let rect = CGRect(x: 0, y: 0, width: width, height: y)
        if selected { painter?.fillRound(rect.insetBy(dx: -3, dy: -3), radius: PVRadius.md + 2, PVColor.graphRing) }
        painter?.fillRound(
            rect, radius: PVRadius.md, PVColor.surfaceCard,
            shadow: cardShadow(selected: selected, dragging: state.isDragging, restingAlpha: placed.isCited ? 0.08 : 0.06, restingBlur: 4, restingY: 1)
        )
        painter?.beginClip(rect, radius: PVRadius.md)

        let working = placed.titleOrTypeLabel
        if placed.isCited {
            paintHeader(
                title: nil, typeLine: typeRef(placed),
                chipFill: PVColor.surfaceSunken, chipInk: PVColor.textMuted, chipLine: PVColor.borderSubtle,
                mark: placed.kind.markKey, chipSize: 22, markSize: 12, citedChip: true,
                header: headerRect, width: width, paddingX: EvidenceBridgeCard.shellPaddingX,
                titleCited: true, deleteHovered: state.hoveredActionID == EvidenceBridgeCard.deleteActionID,
                titleSize: 14, typeSize: 9.5, typeKern: 0.6, painter: painter
            )
        } else {
            paintHeader(
                title: working, typeLine: placed.typeLabel,
                chipFill: PVColor.surfaceSunken, chipInk: PVColor.textMuted, chipLine: PVColor.borderSubtle,
                mark: placed.kind.markKey, chipSize: 22, markSize: 12, citedChip: true,
                header: headerRect, width: width, paddingX: EvidenceBridgeCard.shellPaddingX,
                titleCited: false, deleteHovered: state.hoveredActionID == EvidenceBridgeCard.deleteActionID,
                titleSize: 14, typeSize: 9.5, typeKern: 1, painter: painter
            )
        }
        painter?.text(body, in: bodyRect, font: bodyFont, color: placed.isCited ? PVColor.textPrimary : PVColor.textMuted, lines: 3)
        if let pencil = regions[EvidenceBridgeCard.editCitationActionID] {
            painter?.symbol(
                PVSymbol.penLine, in: pencil,
                state.hoveredActionID == EvidenceBridgeCard.editCitationActionID ? PVColor.accent : PVColor.textMuted
            )
        }
        if let first = rowRects.first, let last = rowRects.last {
            painter?.fill(CGRect(x: 0, y: first.1.minY - 1, width: width, height: last.1.maxY - first.1.minY + 1), PVColor.borderSubtle)
        }
        paintRows(rowRects, tint: PVColor.surfaceCard, chip: PVColor.surfaceHover, ink: PVColor.textMuted, hovered: state.hoveredActionID, width: width, valueWidth: bridgeValueWidth, leading: EvidenceCitedPropertyMarks.leadingInset(shell: EvidenceBridgeCard.shellPaddingX), trailing: EvidenceBridgeCard.shellPaddingX, orders: card.nameOrders, projectDir: card.projectDir, painter: painter)
        paintAdd(addRect, quiet: !placed.isCited, canCite: state.canCite, hovered: state.hoveredActionID == EvidenceSubjectCard.addPropertyActionID, tint: PVColor.surfaceCard, chip: PVColor.surfaceHover, ink: PVColor.textMuted, paddingX: EvidenceBridgeCard.shellPaddingX, painter: painter)
        strokeCard(
            rect,
            color: selected ? PVColor.accent : PVColor.borderSubtle,
            width: selected ? 1.5 : 1,
            painter: painter
        )
        painter?.endClip()
        return EvidenceCardLayout(size: CGSize(width: width, height: y), regions: regions)
    }

    static func paintSubjectShell(
        placed: EvidenceGraphPlacedSubject,
        style: EvidenceSubjectKindStyle,
        selected: Bool,
        dragging: Bool,
        height: CGFloat,
        painter: PVPainter?
    ) {
        let rect = CGRect(x: 0, y: 0, width: EvidenceSubjectCard.width, height: height)
        if selected {
            painter?.fillRound(rect.insetBy(dx: -3, dy: -3), radius: PVRadius.md + 2, PVColor.graphRing)
        }
        painter?.fillRound(
            rect, radius: PVRadius.md, placed.isCited ? style.tint : PVColor.surfaceCard,
            shadow: cardShadow(selected: selected, dragging: dragging, restingAlpha: placed.isCited ? 0.06 : 0, restingBlur: 6, restingY: 2)
        )
        if !placed.isCited { painter?.fillRound(rect, radius: PVRadius.md, style.tint.opacity(0.38)) }
    }

    /// The selection halo is the selected card's edge. A drop shadow on top of
    /// it draws a dark ring, which on a light grid reads as a black outline.
    static func cardShadow(
        selected: Bool,
        dragging: Bool,
        restingAlpha: CGFloat,
        restingBlur: CGFloat,
        restingY: CGFloat
    ) -> PVShadowLayer? {
        if dragging { return PVShadowLayer(color: Color.black.opacity(0.1), radius: 10, x: 0, y: 4) }
        if selected || restingAlpha <= 0 { return nil }
        return PVShadowLayer(color: Color.black.opacity(restingAlpha), radius: restingBlur, x: 0, y: restingY)
    }

    /// Inside stroke, drawn after the body so rows and the footer do not cover
    /// it. A centered stroke sat on the halo and only survived across the title.
    static func strokeCard(
        _ rect: CGRect,
        color: Color,
        width: CGFloat,
        dash: [CGFloat] = [],
        painter: PVPainter?
    ) {
        let inset = width / 2
        painter?.strokeRound(
            rect.insetBy(dx: inset, dy: inset),
            radius: max(PVRadius.md - inset, 0),
            color,
            width: width,
            dash: dash
        )
    }

    static func paintHeader(
        title: String?,
        typeLine: String,
        chipFill: Color,
        chipInk: Color,
        chipLine: Color,
        mark: PVMarkKey,
        chipSize: CGFloat,
        markSize: CGFloat,
        citedChip: Bool,
        header: CGRect,
        width: CGFloat,
        paddingX: CGFloat,
        titleCited: Bool,
        deleteHovered: Bool,
        titleSize: CGFloat = 16,
        typeSize: CGFloat = 10,
        typeKern: CGFloat = 0.6,
        painter: PVPainter?
    ) {
        guard let painter else { return }
        let chipRect = CGRect(x: paddingX, y: header.minY, width: chipSize, height: chipSize)
        let radius: CGFloat = chipSize > 24 ? 4 : 3
        painter.fillRound(chipRect, radius: radius, chipFill)
        painter.strokeRound(chipRect, radius: radius, chipLine, width: 1, dash: citedChip ? [] : [3, 2])
        painter.mark(mark, in: CGRect(x: chipRect.midX - markSize / 2, y: chipRect.midY - markSize / 2, width: markSize, height: markSize), chipInk)
        if chipSize >= 28 {
            let badge = CGRect(x: chipRect.minX - 5, y: chipRect.maxY - 9, width: 14, height: 14)
            painter.fillRound(badge, radius: 7, citedChip ? chipInk : PVColor.surfaceCard)
            painter.strokeRound(badge, radius: 7, PVColor.surfaceCard, width: 1.5)
            if citedChip {
                painter.symbol(PVSymbol.check, in: badge.insetBy(dx: 2.5, dy: 2.5), chipFill)
            } else {
                painter.symbol(PVSymbol.circleDashed, in: badge.insetBy(dx: 1, dy: 1), PVColor.evidenceUndocumented)
            }
        }
        let titleX = chipRect.maxX + 10
        let titleWidth = width - titleX - paddingX - 40
        if let title {
            let titleFont = PVFont.nsDisplay(size: titleSize, weight: PVFontWeight.medium)
            let titleHeight = PVTextMetrics.textHeight(title, width: titleWidth, font: titleFont, maxLines: 2)
            painter.text(title, in: CGRect(x: titleX, y: header.minY, width: titleWidth, height: titleHeight), font: titleFont, color: titleCited ? PVColor.textPrimary : PVColor.textSecondary, lines: 2)
            painter.text(typeLine, in: CGRect(x: titleX, y: header.minY + titleHeight + 2, width: titleWidth, height: 14), font: PVFont.nsMono(size: typeSize), color: chipInk, lines: 1, kern: typeKern, uppercased: true)
        } else {
            painter.text(typeLine, in: CGRect(x: titleX, y: header.midY - 6, width: titleWidth, height: 14), font: PVFont.nsMono(size: typeSize), color: PVColor.textMuted, lines: 1, kern: typeKern, uppercased: true)
        }
        let iconY = header.minY + 2
        let deleteX = width - paddingX - 12
        painter.symbol(PVSymbol.penLine, in: CGRect(x: deleteX - 18, y: iconY, width: 12, height: 12), PVColor.textMuted)
        painter.symbol(PVSymbol.trash, in: CGRect(x: deleteX, y: iconY, width: 12, height: 12), deleteHovered ? PVColor.danger : PVColor.textMuted)
    }

    static func paintRows(
        _ rows: [(CatalogObservation, CGRect)],
        tint: Color,
        chip: Color,
        ink: Color,
        hovered: String?,
        width: CGFloat,
        valueWidth: CGFloat,
        leading: CGFloat,
        trailing: CGFloat,
        orders: NameFormatOrders,
        projectDir: String,
        painter: PVPainter?
    ) {
        guard let painter else { return }
        for (observation, rect) in rows {
            let id = EvidenceSubjectCard.editPropertyActionID(observationID: observation.id)
            let hot = hovered == id
            painter.fill(rect, hot ? chip : tint)
            if hot { painter.fill(CGRect(x: 0, y: rect.minY, width: 2, height: rect.height), ink) }
            let label = observation.propertyLabel.trimmingCharacters(in: .whitespacesAndNewlines)
            painter.text(
                label.isEmpty ? observation.propertyKey : label,
                in: CGRect(x: leading, y: rect.minY + 8, width: valueWidth, height: 12),
                font: PVFont.nsMono(size: 9), color: PVColor.textMuted, lines: 1, kern: 0.7, uppercased: true
            )
            let negated = EvidenceCitedPropertyMarks.isNegated(observation)
            var valueX = leading
            if negated {
                painter.text(
                    L10n.string(L10n.EvidenceGraph.negatedPrefix),
                    in: CGRect(x: leading, y: rect.minY + 22, width: 28, height: 14),
                    font: PVFont.nsBody(size: PVTypeScale.micro, weight: PVFontWeight.semibold),
                    color: PVColor.danger, lines: 1, kern: PVTypeScale.micro * PVTracking.caps, uppercased: true
                )
                valueX += 30
            }
            if let relPath = likenessRelPath(observation) {
                let image = ThumbnailCache.shared.cachedImage(projectDir: projectDir, relPath: relPath)
                painter.tile(image, in: CGRect(x: valueX, y: rect.minY + 22, width: likenessTile, height: likenessTile), radius: 4)
            } else {
                painter.text(
                    ObservationValueDisplay.string(for: observation, orders: orders),
                    in: CGRect(x: valueX, y: rect.minY + 22, width: max(valueWidth - (valueX - leading), 20), height: rect.height - 30),
                    font: PVFont.nsBody(size: 13, italic: negated),
                    color: negated ? PVColor.danger : PVColor.textPrimary, lines: 2
                )
            }
            painter.symbol(PVSymbol.penLine, in: CGRect(x: width - trailing - 12, y: rect.midY - 6, width: 12, height: 12), PVColor.textFaint)
        }
        let observations = rows.map(\.0)
        for run in EvidenceCitedPropertyMarks.conflictRuns(in: observations) {
            guard rows.indices.contains(run.lowerBound), rows.indices.contains(run.upperBound - 1) else { continue }
            let start = rows[run.lowerBound].1
            let end = rows[run.upperBound - 1].1
            painter.bracket(CGRect(x: 0, y: start.minY, width: 20, height: end.maxY - start.minY))
        }
    }

    static func paintAdd(
        _ rect: CGRect,
        quiet: Bool,
        canCite: Bool,
        hovered: Bool,
        tint: Color,
        chip: Color,
        ink: Color,
        paddingX: CGFloat,
        painter: PVPainter?
    ) {
        guard let painter else { return }
        let hot = hovered && canCite
        if !quiet { painter.fill(rect, hot ? chip : tint) }
        if hot { painter.fill(CGRect(x: 0, y: rect.minY, width: 2, height: rect.height), ink) }
        let color: Color = canCite ? (hovered ? PVColor.textPrimary : PVColor.textMuted) : PVColor.textFaint
        let y = quiet ? rect.minY + 4 : rect.minY + 9
        painter.symbol(PVSymbol.plus, in: CGRect(x: paddingX, y: y, width: 12, height: 12), color)
        painter.text(
            L10n.string(L10n.EvidenceGraph.addProperty),
            in: CGRect(x: paddingX + 18, y: y - 1, width: rect.width - paddingX - 24, height: 16),
            font: PVFont.nsBody(size: 12), color: color, lines: 1
        )
    }

    /// Records the footer's hit regions at `y` and advances `y` past it. An
    /// unpromoted card records Promote over the whole footer; a promoted one
    /// records the handle row and the Unpromote square from the same parts the
    /// paint reads. Returns the whole footer rect.
    private static func recordFooter(
        placed: EvidenceGraphPlacedSubject,
        y: inout CGFloat,
        width: CGFloat,
        regions: inout [String: CGRect]
    ) -> CGRect {
        let footer = CGRect(x: 0, y: y, width: width, height: EvidenceSubjectCard.footerHeight)
        y += EvidenceSubjectCard.footerHeight
        if placed.membership == nil {
            regions[EvidenceSubjectCard.promoteActionID] = footer
        } else {
            let parts = EvidenceSubjectCard.promotedFooterParts(footer: footer)
            regions[EvidenceSubjectCard.openHandleActionID] = parts.handle
            regions[EvidenceSubjectCard.unpromoteActionID] = parts.slot
        }
        return footer
    }

    static func paintFooter(
        _ rect: CGRect,
        placed: EvidenceGraphPlacedSubject,
        style: EvidenceSubjectKindStyle,
        hovered: String?,
        pressed: String?,
        orders: NameFormatOrders,
        painter: PVPainter?
    ) {
        guard let painter else { return }
        painter.fill(rect, style.band)
        if let membership = placed.membership {
            let parts = EvidenceSubjectCard.promotedFooterParts(footer: rect)
            let down = pressed == EvidenceSubjectCard.openHandleActionID
            let hot = hovered == EvidenceSubjectCard.openHandleActionID
            if hot || down { painter.fill(parts.handle, style.ink.opacity(down ? 0.18 : 0.10)) }
            painter.text(membership.entity.ref, in: CGRect(x: EvidenceSubjectCard.shellPaddingX, y: rect.midY - 7, width: 72, height: 14), font: PVFont.nsMono(size: 10), color: style.ink, lines: 1)
            let stored = EvidenceSubjectCard.membershipName(membership, orders: orders)
            let name = stored ?? L10n.string(L10n.EvidenceGraph.openHandlePage(kind: placed.kind))
            painter.text(
                name,
                in: CGRect(x: EvidenceSubjectCard.shellPaddingX + 78, y: rect.midY - 8, width: parts.handle.width - 120, height: 16),
                font: PVFont.nsBody(size: 12.5, italic: stored == nil), color: style.ink, lines: 1
            )
            painter.symbol(PVSymbol.chevronForward, in: CGRect(x: parts.handle.maxX - 27, y: rect.midY - 7, width: 14, height: 14), style.ink)
            painter.fill(parts.hairline, style.line)
            let unpromoteDown = pressed == EvidenceSubjectCard.unpromoteActionID
            if unpromoteDown || hovered == EvidenceSubjectCard.unpromoteActionID {
                let side = PVControlSize.sm.height
                painter.fillRound(
                    CGRect(x: parts.slot.midX - side / 2, y: parts.slot.midY - side / 2, width: side, height: side),
                    radius: PVRadius.sm, style.ink.opacity(unpromoteDown ? 0.18 : 0.10)
                )
            }
            let glyph = PVControlSize.sm.iconGlyphSize
            painter.symbol(
                PVSymbol.unpromote,
                in: CGRect(x: parts.slot.midX - glyph / 2, y: parts.slot.midY - glyph / 2, width: glyph, height: glyph),
                style.ink
            )
        } else {
            let hot = hovered == EvidenceSubjectCard.promoteActionID
                || pressed == EvidenceSubjectCard.promoteActionID
            if hot {
                painter.fill(rect, style.chip)
                painter.fill(CGRect(x: 0, y: rect.minY, width: 2, height: rect.height), style.ink)
            }
            let font = PVFont.nsBody(size: 13, weight: PVFontWeight.medium)
            let line = PVTextMetrics.fontLineHeight(font)
            let icon: CGFloat = 13
            let x = EvidenceSubjectCard.shellPaddingX
            painter.symbol(
                PVSymbol.promote,
                in: CGRect(x: x, y: rect.midY - icon / 2, width: icon, height: icon),
                PVColor.textSecondary
            )
            painter.text(
                L10n.string(L10n.EvidenceGraph.promote),
                in: CGRect(x: x + icon + 6, y: rect.midY - line / 2, width: rect.width - x - icon - 20, height: line),
                font: font, color: PVColor.textSecondary, lines: 1
            )
        }
    }
}

// MARK: - Text measurement

private extension EvidenceCardView {
    static var subjectValueWidth: CGFloat {
        EvidenceSubjectCard.width - EvidenceCitedPropertyMarks.leadingInset(shell: EvidenceSubjectCard.shellPaddingX) - EvidenceSubjectCard.shellPaddingX - 20
    }
    static var bridgeValueWidth: CGFloat {
        EvidenceBridgeCard.width - EvidenceCitedPropertyMarks.leadingInset(shell: EvidenceBridgeCard.shellPaddingX) - EvidenceBridgeCard.shellPaddingX - 20
    }
    static var bridgeBodyWidth: CGFloat {
        EvidenceBridgeCard.width - EvidenceBridgeCard.shellPaddingX * 2 - 20
    }
    static var addPropertyRowHeight: CGFloat { PVTextMetrics.fontLineHeight(PVFont.nsBody(size: 12)) + 20 }

    static func typeLine(_ placed: EvidenceGraphPlacedSubject) -> String {
        typeAndRef(typeLabel: placed.typeLabel, ref: placed.subject.ref)
    }
    static func typeRef(_ placed: EvidenceGraphPlacedBridge) -> String {
        typeAndRef(typeLabel: placed.typeLabel, ref: placed.subject.ref)
    }
    private static func typeAndRef(typeLabel: String, ref: String) -> String {
        let ref = ref.trimmingCharacters(in: .whitespacesAndNewlines)
        return ref.isEmpty ? typeLabel : "\(typeLabel) · \(ref)"
    }
    static func titleBlockHeight(_ placed: EvidenceGraphPlacedSubject) -> CGFloat {
        let width = EvidenceSubjectCard.width - EvidenceSubjectCard.shellPaddingX * 2 - 78
        let title = PVTextMetrics.textHeight(placed.titleOrTypeLabel, width: width, font: PVFont.nsDisplay(size: 16, weight: PVFontWeight.medium), maxLines: 2)
        return title + 2 + PVTextMetrics.fontLineHeight(PVFont.nsMono(size: 10))
    }
    static func bridgeTitleBlock(_ placed: EvidenceGraphPlacedBridge) -> CGFloat {
        PVTextMetrics.textHeight(placed.titleOrTypeLabel, width: bridgeBodyWidth, font: PVFont.nsDisplay(size: 14, weight: PVFontWeight.medium), maxLines: 2) + 1 + PVTextMetrics.fontLineHeight(PVFont.nsMono(size: 9.5))
    }
    /// Side of a likeness row's image tile.
    static var likenessTile: CGFloat { 32 }

    /// The image path when the row holds a depiction, else nil.
    static func likenessRelPath(_ observation: CatalogObservation) -> String? {
        guard let path = observation.depiction?.imageRelPath.trimmingCharacters(in: .whitespacesAndNewlines),
              !path.isEmpty else { return nil }
        return path
    }

    static func propertyRowHeight(
        _ observation: CatalogObservation,
        valueWidth: CGFloat,
        valueFont: NSFont,
        orders: NameFormatOrders
    ) -> CGFloat {
        if likenessRelPath(observation) != nil {
            return max(EvidenceSubjectCard.propertyRowHeight, PVTextMetrics.fontLineHeight(PVFont.nsMono(size: 9)) + 3 + likenessTile + 16)
        }
        let valueHeight = PVTextMetrics.textHeight(
            ObservationValueDisplay.string(for: observation, orders: orders),
            width: valueWidth,
            font: valueFont,
            maxLines: 2
        )
        return max(EvidenceSubjectCard.propertyRowHeight, PVTextMetrics.fontLineHeight(PVFont.nsMono(size: 9)) + 3 + valueHeight + 16)
    }
}
