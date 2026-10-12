import AppKit
import SwiftUI

/// Subject and bridge cards drawn in AppKit.
///
/// The canvas hosts one of these instead of an `NSHostingView`. Text measurement
/// here is also the hit layout edges and the pointer read, so a wrapped title
/// or value moves its band with the paint. The placement ghost stays SwiftUI.
final class EvidenceAppKitCard: NSView {
    let layout: EvidenceCardLayout
    private let card: EvidenceCanvasCard

    init(card: EvidenceCanvasCard, actions: [NSAccessibilityCustomAction]) {
        self.card = card
        self.layout = Self.measure(card)
        let outset = GraphCanvasItemLayer.paintOutset
        super.init(frame: NSRect(
            x: 0, y: 0,
            width: layout.size.width + outset * 2,
            height: layout.size.height + outset * 2
        ))
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

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    override var isFlipped: Bool { true }
    override var isOpaque: Bool { false }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        syncBackingScale()
    }

    override func viewDidChangeBackingProperties() {
        super.viewDidChangeBackingProperties()
        syncBackingScale()
    }

    /// The first display can cache the card at 1x, before the window's scale
    /// is known. Selecting rebuilt the view and hid that; match the window
    /// scale as soon as the card is on screen.
    private func syncBackingScale() {
        guard let scale = window?.backingScaleFactor, scale > 0, let layer else { return }
        guard layer.contentsScale != scale else { return }
        layer.contentsScale = scale
        needsDisplay = true
    }

    override var intrinsicContentSize: NSSize {
        NSSize(
            width: layout.size.width + GraphCanvasItemLayer.paintOutset * 2,
            height: layout.size.height + GraphCanvasItemLayer.paintOutset * 2
        )
    }

    override func draw(_ dirtyRect: NSRect) {
        let alpha: CGFloat = {
            if card.state.isDragging { return 0.92 }
            if case .subject(let placed, _) = card.placed, !placed.isCited { return 0.76 }
            return 1
        }()
        NSGraphicsContext.saveGraphicsState()
        NSGraphicsContext.current?.cgContext.setAlpha(alpha)
        _ = Self.measure(card, painter: Painter(outset: GraphCanvasItemLayer.paintOutset))
        NSGraphicsContext.restoreGraphicsState()
    }

    private static func measure(_ card: EvidenceCanvasCard, painter: Painter? = nil) -> EvidenceCardLayout {
        switch card.placed {
        case .subject(let placed, let presentation):
            subject(placed, presentation: presentation, card: card, painter: painter)
        case .bridge(let placed):
            bridge(placed, card: card, painter: painter)
        }
    }
}

// MARK: - Painter

private final class Painter {
    let outset: CGFloat
    init(outset: CGFloat) { self.outset = outset }

    func place(_ rect: CGRect) -> CGRect { rect.offsetBy(dx: outset, dy: outset) }

    func fill(_ rect: CGRect, _ color: Color) {
        ns(color).setFill()
        place(rect).fill()
    }

    func fillRound(_ rect: CGRect, radius: CGFloat, _ color: Color, shadow: (CGFloat, CGFloat, CGFloat)? = nil) {
        let path = continuous(place(rect), radius: radius)
        if let shadow {
            NSGraphicsContext.saveGraphicsState()
            let shade = NSShadow()
            shade.shadowBlurRadius = shadow.0
            shade.shadowOffset = NSSize(width: 0, height: shadow.1)
            shade.shadowColor = NSColor.black.withAlphaComponent(shadow.2)
            shade.set()
            ns(color).setFill()
            path.fill()
            NSGraphicsContext.restoreGraphicsState()
        } else {
            ns(color).setFill()
            path.fill()
        }
    }

    func beginClip(_ rect: CGRect, radius: CGFloat) {
        NSGraphicsContext.saveGraphicsState()
        continuous(place(rect), radius: radius).setClip()
    }

    func endClip() {
        NSGraphicsContext.restoreGraphicsState()
    }

    func strokeRound(_ rect: CGRect, radius: CGFloat, _ color: Color, width: CGFloat, dash: [CGFloat] = []) {
        let path = continuous(place(rect), radius: radius)
        path.lineWidth = width
        if !dash.isEmpty { path.setLineDash(dash, count: dash.count, phase: 0) }
        ns(color).setStroke()
        path.stroke()
    }

    func text(
        _ string: String,
        in rect: CGRect,
        font: NSFont,
        color: Color,
        lines: Int,
        kern: CGFloat = 0,
        uppercased: Bool = false
    ) {
        let style = NSMutableParagraphStyle()
        style.lineBreakMode = .byWordWrapping
        let attributes: [NSAttributedString.Key: Any] = [
            .font: font,
            .foregroundColor: ns(color),
            .paragraphStyle: style,
            .kern: kern,
        ]
        let line = max(font.boundingRectForFont.height, 1)
        var box = place(rect)
        box.size.height = min(box.height, ceil(line) * CGFloat(max(lines, 1)))
        NSAttributedString(string: uppercased ? string.uppercased() : string, attributes: attributes)
            .draw(with: box, options: [.usesLineFragmentOrigin, .truncatesLastVisibleLine])
    }

    func symbol(_ name: String, in rect: CGRect, _ color: Color) {
        guard let base = NSImage(systemSymbolName: name, accessibilityDescription: nil) else { return }
        let image = base.withSymbolConfiguration(
            NSImage.SymbolConfiguration(pointSize: min(rect.width, rect.height), weight: .regular)
        ) ?? base
        drawTinted(image, in: place(rect), color: ns(color))
    }

    func mark(_ key: PVMarkKey, in rect: CGRect, _ color: Color) {
        guard let image = NSImage(named: key.assetName) else { return }
        drawTinted(image, in: place(rect), color: ns(color))
    }

    /// Recolors a template image in the view's current context.
    ///
    /// An intermediate `NSImage` rasterizes the SVG mark at one pixel per
    /// point. Panning and zooming then scale that bitmap, so the mark stays
    /// soft until some later redraw. A transparency layer tints the vector
    /// at the same scale as the rest of the card.
    private func drawTinted(_ image: NSImage, in rect: CGRect, color: NSColor) {
        guard let context = NSGraphicsContext.current else { return }
        context.saveGraphicsState()
        context.cgContext.beginTransparencyLayer(in: rect, auxiliaryInfo: nil)
        image.draw(
            in: rect,
            from: .zero,
            operation: .sourceOver,
            fraction: 1,
            respectFlipped: true,
            hints: [.interpolation: NSImageInterpolation.high]
        )
        color.setFill()
        rect.fill(using: .sourceAtop)
        context.cgContext.endTransparencyLayer()
        context.restoreGraphicsState()
    }

    func bracket(_ rect: CGRect) {
        let box = place(rect)
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
        ns(PVColor.warning).setStroke()
        path.stroke()
    }
}

/// Resolves a dynamic color to the appearance currently being drawn, so a
/// tint baked into an image does not stick on the default (black or accent).
private func ns(_ color: Color) -> NSColor {
    let dynamic = NSColor(color)
    return NSColor(cgColor: dynamic.cgColor) ?? dynamic
}

private func continuous(_ rect: CGRect, radius: CGFloat) -> NSBezierPath {
    let path = NSBezierPath()
    Path(roundedRect: rect, cornerRadius: radius, style: .continuous).forEach { element in
        switch element {
        case .move(let point): path.move(to: point)
        case .line(let point): path.line(to: point)
        case .quadCurve(let point, let control):
            path.curve(to: point, controlPoint1: control, controlPoint2: control)
        case .curve(let point, let control1, let control2):
            path.curve(to: point, controlPoint1: control1, controlPoint2: control2)
        case .closeSubpath: path.close()
        }
    }
    return path
}

// MARK: - Layout and paint

private extension EvidenceAppKitCard {
    static func subject(
        _ placed: SourceGraphPlacedSubject,
        presentation: CatalogSubjectTypePresentation?,
        card: EvidenceCanvasCard,
        painter: Painter?
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
            let line = lineHeight(mono(11))
            painter?.text(
                L10n.string(L10n.EvidenceGraph.connectingFrom),
                in: CGRect(x: EvidenceSubjectCard.shellPaddingX, y: y, width: width - 26, height: line),
                font: mono(11), color: PVColor.accent, lines: 1
            )
            y += line
        } else {
            let description = placed.subject.description.trimmingCharacters(in: .whitespacesAndNewlines)
            if !description.isEmpty {
                y += EvidenceSubjectCard.headerToBodyGap
                let bodyWidth = width - EvidenceSubjectCard.shellPaddingX * 2
                let font = PVFont.nsBody(size: 12.5, italic: true)
                let height = wrappedHeight(description, width: bodyWidth, font: font, maxLines: 3)
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
                let rowHeight = propertyRowHeight(observation, valueWidth: subjectValueWidth, valueFont: valueFont)
                let rect = CGRect(x: 0, y: y, width: width, height: rowHeight)
                rowRects.append((observation, rect))
                regions[EvidenceSubjectCard.editPropertyActionID(observationID: observation.id)] = rect
                y += rowHeight + EvidenceSubjectCard.stackHairline
            }
            addRect = CGRect(x: 0, y: y, width: width, height: addPropertyRowHeight)
            regions[EvidenceSubjectCard.addPropertyActionID] = addRect
            y += addPropertyRowHeight + EvidenceSubjectCard.stackHairline
            let footerID = placed.membership == nil
                ? EvidenceSubjectCard.promoteActionID : EvidenceSubjectCard.openHandleActionID
            footerRect = CGRect(x: 0, y: y, width: width, height: EvidenceSubjectCard.footerHeight)
            regions[footerID] = footerRect
            y += EvidenceSubjectCard.footerHeight
        } else {
            y += EvidenceSubjectCard.headerToBodyGap
            addRect = CGRect(x: 0, y: y, width: width, height: EvidenceSubjectCard.addPropertyQuietHeight)
            regions[EvidenceSubjectCard.addPropertyActionID] = addRect
            y += EvidenceSubjectCard.addPropertyQuietHeight
                + EvidenceSubjectCard.headerToBodyGap
                + EvidenceSubjectCard.uncitedFooterTopMargin
                + EvidenceSubjectCard.uncitedFooterRule
            let footerID = placed.membership == nil
                ? EvidenceSubjectCard.promoteActionID : EvidenceSubjectCard.openHandleActionID
            footerRect = CGRect(x: 0, y: y, width: width, height: EvidenceSubjectCard.footerHeight)
            regions[footerID] = footerRect
            y += EvidenceSubjectCard.footerHeight
        }

        paintSubjectShell(placed: placed, style: style, selected: selected, dragging: state.isDragging, height: y, painter: painter)
        let cardRect = CGRect(x: 0, y: 0, width: width, height: y)
        painter?.beginClip(cardRect, radius: PVRadius.md)
        paintHeader(
            title: displayLabel(placed), typeLine: typeLine(placed),
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
        paintRows(rowRects, tint: style.tint, chip: style.chip, ink: style.ink, hovered: state.hoveredActionID, width: width, valueWidth: subjectValueWidth, leading: EvidenceCitedPropertyMarks.leadingInset(shell: EvidenceSubjectCard.shellPaddingX), trailing: EvidenceSubjectCard.shellPaddingX, painter: painter)
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

    static func bridge(_ placed: SourceGraphPlacedBridge, card: EvidenceCanvasCard, painter: Painter?) -> EvidenceCardLayout {
        let width = EvidenceBridgeCard.width
        let state = card.state
        let selected = state.isSelected || state.isActivated || state.isDragging
        let body = placed.isCited ? (card.text.summary ?? "") : L10n.string(L10n.EvidenceGraph.bridgeHonestyBody)
        let bodyFont = placed.isCited
            ? PVFont.nsDisplay(size: 14.5, weight: PVFontWeight.medium)
            : PVFont.nsBody(size: 11.5, italic: true)
        let bodyHeight = wrappedHeight(body, width: bridgeBodyWidth, font: bodyFont, maxLines: 3)
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
                let rowHeight = propertyRowHeight(observation, valueWidth: bridgeValueWidth, valueFont: valueFont)
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

        let label = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let working = label.isEmpty ? placed.typeLabel : label
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
                PVSymbol.penLine.rawValue, in: pencil,
                state.hoveredActionID == EvidenceBridgeCard.editCitationActionID ? PVColor.accent : PVColor.textMuted
            )
        }
        if let first = rowRects.first, let last = rowRects.last {
            painter?.fill(CGRect(x: 0, y: first.1.minY - 1, width: width, height: last.1.maxY - first.1.minY + 1), PVColor.borderSubtle)
        }
        paintRows(rowRects, tint: PVColor.surfaceCard, chip: PVColor.surfaceHover, ink: PVColor.textMuted, hovered: state.hoveredActionID, width: width, valueWidth: bridgeValueWidth, leading: EvidenceCitedPropertyMarks.leadingInset(shell: EvidenceBridgeCard.shellPaddingX), trailing: EvidenceBridgeCard.shellPaddingX, painter: painter)
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
        placed: SourceGraphPlacedSubject,
        style: EvidenceSubjectKindStyle,
        selected: Bool,
        dragging: Bool,
        height: CGFloat,
        painter: Painter?
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
    ) -> (CGFloat, CGFloat, CGFloat)? {
        if dragging { return (10, 4, 0.1) }
        if selected || restingAlpha <= 0 { return nil }
        return (restingBlur, restingY, restingAlpha)
    }

    /// Inside stroke, drawn after the body so rows and the footer do not cover
    /// it. A centered stroke sat on the halo and only survived across the title.
    static func strokeCard(
        _ rect: CGRect,
        color: Color,
        width: CGFloat,
        dash: [CGFloat] = [],
        painter: Painter?
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
        painter: Painter?
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
                painter.symbol(PVSymbol.check.rawValue, in: badge.insetBy(dx: 2.5, dy: 2.5), chipFill)
            } else {
                painter.symbol(PVSymbol.circleDashed.rawValue, in: badge.insetBy(dx: 1, dy: 1), PVColor.evidenceUndocumented)
            }
        }
        let titleX = chipRect.maxX + 10
        let titleWidth = width - titleX - paddingX - 40
        if let title {
            let titleFont = PVFont.nsDisplay(size: titleSize, weight: PVFontWeight.medium)
            let titleHeight = wrappedHeight(title, width: titleWidth, font: titleFont, maxLines: 2)
            painter.text(title, in: CGRect(x: titleX, y: header.minY, width: titleWidth, height: titleHeight), font: titleFont, color: titleCited ? PVColor.textPrimary : PVColor.textSecondary, lines: 2)
            painter.text(typeLine, in: CGRect(x: titleX, y: header.minY + titleHeight + 2, width: titleWidth, height: 14), font: mono(typeSize), color: chipInk, lines: 1, kern: typeKern, uppercased: true)
        } else {
            painter.text(typeLine, in: CGRect(x: titleX, y: header.midY - 6, width: titleWidth, height: 14), font: mono(typeSize), color: PVColor.textMuted, lines: 1, kern: typeKern, uppercased: true)
        }
        let iconY = header.minY + 2
        let deleteX = width - paddingX - 12
        painter.symbol(PVSymbol.penLine.rawValue, in: CGRect(x: deleteX - 18, y: iconY, width: 12, height: 12), PVColor.textMuted)
        painter.symbol(PVSymbol.trash.rawValue, in: CGRect(x: deleteX, y: iconY, width: 12, height: 12), deleteHovered ? PVColor.danger : PVColor.textMuted)
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
        painter: Painter?
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
                font: mono(9), color: PVColor.textMuted, lines: 1, kern: 0.7, uppercased: true
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
            painter.text(
                ObservationValueDisplay.string(for: observation),
                in: CGRect(x: valueX, y: rect.minY + 22, width: max(valueWidth - (valueX - leading), 20), height: rect.height - 30),
                font: PVFont.nsBody(size: 13, italic: negated),
                color: negated ? PVColor.danger : PVColor.textPrimary, lines: 2
            )
            painter.symbol(PVSymbol.penLine.rawValue, in: CGRect(x: width - trailing - 12, y: rect.midY - 6, width: 12, height: 12), PVColor.textFaint)
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
        painter: Painter?
    ) {
        guard let painter else { return }
        let hot = hovered && canCite
        if !quiet { painter.fill(rect, hot ? chip : tint) }
        if hot { painter.fill(CGRect(x: 0, y: rect.minY, width: 2, height: rect.height), ink) }
        let color: Color = canCite ? (hovered ? PVColor.textPrimary : PVColor.textMuted) : PVColor.textFaint
        let y = quiet ? rect.minY + 4 : rect.minY + 9
        painter.symbol(PVSymbol.plus.rawValue, in: CGRect(x: paddingX, y: y, width: 12, height: 12), color)
        painter.text(
            L10n.string(L10n.EvidenceGraph.addProperty),
            in: CGRect(x: paddingX + 18, y: y - 1, width: rect.width - paddingX - 24, height: 16),
            font: PVFont.nsBody(size: 12), color: color, lines: 1
        )
    }

    static func paintFooter(
        _ rect: CGRect,
        placed: SourceGraphPlacedSubject,
        style: EvidenceSubjectKindStyle,
        hovered: String?,
        pressed: String?,
        orders: NameFormatOrders = NameFormatOrders(),
        painter: Painter?
    ) {
        guard let painter else { return }
        painter.fill(rect, style.band)
        if let membership = placed.membership {
            let down = pressed == EvidenceSubjectCard.openHandleActionID
            let hot = hovered == EvidenceSubjectCard.openHandleActionID
            if hot || down { painter.fill(rect, style.ink.opacity(down ? 0.18 : 0.10)) }
            painter.text(membership.entity.ref, in: CGRect(x: EvidenceSubjectCard.shellPaddingX, y: rect.midY - 7, width: 72, height: 14), font: mono(10), color: style.ink, lines: 1)
            let stored = EvidenceSubjectCard.membershipName(membership, orders: orders)
            let name = stored ?? L10n.string(L10n.EvidenceGraph.openHandlePage(kind: placed.kind))
            painter.text(
                name,
                in: CGRect(x: EvidenceSubjectCard.shellPaddingX + 78, y: rect.midY - 8, width: rect.width - 120, height: 16),
                font: PVFont.nsBody(size: 12.5, italic: stored == nil), color: style.ink, lines: 1
            )
            painter.symbol(PVSymbol.chevronForward.rawValue, in: CGRect(x: rect.maxX - 27, y: rect.midY - 7, width: 14, height: 14), style.ink)
        } else {
            let hot = hovered == EvidenceSubjectCard.promoteActionID
                || pressed == EvidenceSubjectCard.promoteActionID
            if hot {
                painter.fill(rect, style.chip)
                painter.fill(CGRect(x: 0, y: rect.minY, width: 2, height: rect.height), style.ink)
            }
            let font = PVFont.nsBody(size: 13, weight: PVFontWeight.medium)
            let line = lineHeight(font)
            let icon: CGFloat = 13
            let x = EvidenceSubjectCard.shellPaddingX
            painter.symbol(
                PVSymbol.arrowUpRight.rawValue,
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

private extension EvidenceAppKitCard {
    static var subjectValueWidth: CGFloat {
        EvidenceSubjectCard.width - EvidenceCitedPropertyMarks.leadingInset(shell: EvidenceSubjectCard.shellPaddingX) - EvidenceSubjectCard.shellPaddingX - 20
    }
    static var bridgeValueWidth: CGFloat {
        EvidenceBridgeCard.width - EvidenceCitedPropertyMarks.leadingInset(shell: EvidenceBridgeCard.shellPaddingX) - EvidenceBridgeCard.shellPaddingX - 20
    }
    static var bridgeBodyWidth: CGFloat {
        EvidenceBridgeCard.width - EvidenceBridgeCard.shellPaddingX * 2 - 20
    }
    static var addPropertyRowHeight: CGFloat { lineHeight(PVFont.nsBody(size: 12)) + 20 }

    static func displayLabel(_ placed: SourceGraphPlacedSubject) -> String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? placed.typeLabel : trimmed
    }
    static func typeLine(_ placed: SourceGraphPlacedSubject) -> String {
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        return ref.isEmpty ? placed.typeLabel : "\(placed.typeLabel) · \(ref)"
    }
    static func typeRef(_ placed: SourceGraphPlacedBridge) -> String {
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        return ref.isEmpty ? placed.typeLabel : "\(placed.typeLabel) · \(ref)"
    }
    static func titleBlockHeight(_ placed: SourceGraphPlacedSubject) -> CGFloat {
        let width = EvidenceSubjectCard.width - EvidenceSubjectCard.shellPaddingX * 2 - 78
        let title = wrappedHeight(displayLabel(placed), width: width, font: PVFont.nsDisplay(size: 16, weight: PVFontWeight.medium), maxLines: 2)
        return title + 2 + lineHeight(mono(10))
    }
    static func bridgeTitleBlock(_ placed: SourceGraphPlacedBridge) -> CGFloat {
        let label = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let title = label.isEmpty ? placed.typeLabel : label
        return wrappedHeight(title, width: bridgeBodyWidth, font: PVFont.nsDisplay(size: 14, weight: PVFontWeight.medium), maxLines: 2) + 1 + lineHeight(mono(9.5))
    }
    static func propertyRowHeight(_ observation: CatalogObservation, valueWidth: CGFloat, valueFont: NSFont) -> CGFloat {
        let valueHeight = wrappedHeight(ObservationValueDisplay.string(for: observation), width: valueWidth, font: valueFont, maxLines: 2)
        return max(EvidenceSubjectCard.propertyRowHeight, lineHeight(mono(9)) + 3 + valueHeight + 16)
    }
    static func mono(_ size: CGFloat) -> NSFont {
        let descriptor = NSFontDescriptor(fontAttributes: [
            .family: PVFontFamily.mono,
            .traits: [NSFontDescriptor.TraitKey.weight: NSFont.Weight.regular.rawValue],
        ])
        return NSFont(descriptor: descriptor, size: size) ?? NSFont.monospacedSystemFont(ofSize: size, weight: .regular)
    }
    static func lineHeight(_ font: NSFont) -> CGFloat { max(font.boundingRectForFont.height, 1) }
    static func wrappedHeight(_ text: String, width: CGFloat, font: NSFont, maxLines: Int) -> CGFloat {
        if text.isEmpty { return lineHeight(font) }
        let style = NSMutableParagraphStyle()
        style.lineBreakMode = .byWordWrapping
        let rect = (text as NSString).boundingRect(
            with: CGSize(width: max(width, 1), height: .greatestFiniteMagnitude),
            options: [.usesLineFragmentOrigin, .usesFontLeading],
            attributes: [.font: font, .paragraphStyle: style]
        )
        return min(ceil(rect.height), ceil(lineHeight(font)) * CGFloat(maxLines))
    }
}
