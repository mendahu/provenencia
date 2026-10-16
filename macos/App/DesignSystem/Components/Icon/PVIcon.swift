import SwiftUI

/// Compile-checked SF Symbol names used by the design system.
enum PVSymbol: String {
    case folderPlus = "folder.badge.plus"
    case folderOpen = "folder"
    case check = "checkmark"
    case info = "info.circle"
    case chevronDown = "chevron.down"
    case chevronUp = "chevron.up"
    case chevronBack = "chevron.left"
    case chevronForward = "chevron.right"
    /// Lucide `arrow-right` — trailing jump on the Evidence graph header.
    case arrowRight = "arrow.right"
    /// Lucide `arrow-left` — leading Back to the previous step (Promote footer).
    case arrowLeft = "arrow.left"
    /// Lucide `git-branch` — Evidence graph jump on the Source identity header.
    case gitBranch = "arrow.triangle.branch"
    /// Lucide `merge` — Merged state badge (S9-16).
    case merge = "arrow.merge"
    /// Lucide `split` — Mixed state badge (S9-16).
    case split = "arrow.branch"
    /// Lucide `stamp` — Concluded state badge (S9-16).
    case stamp = "seal"
    /// Lucide `git-merge` — a record folded into a fuller value (S9-16).
    case gitMerge = "arrow.triangle.merge"
    /// Lucide `scale` — a record outvoted by more Sources (S9-16).
    case scale = "scalemass"
    /// Lucide `signal-low` — a record dropped as weak evidence (S9-16).
    case signalLow = "gauge.with.dots.needle.0percent"
    /// Lucide `ban` — a record denied by a stronger negative (S9-16).
    case ban = "nosign"
    /// Lucide `circle-minus` — a negative record counting against (S9-16).
    case circleMinus = "minus.circle"
    /// Lucide `minus` — a record with no usable value (S9-16).
    case minus = "minus"
    case photo = "photo"
    case scrollText = "doc.text"
    /// Lucide `file` — locator floor (entire artifact).
    case file = "doc"
    /// Lucide `quote` — citation menu in the composer header.
    case quote = "quote.opening"
    /// Lucide `crop` — region summary in the locator chip.
    case crop = "crop"
    /// Lucide `list-plus` — empty observation list.
    case listPlus = "list.bullet.badge.plus"
    /// Lucide `signature` — opens the name editor.
    case signature = "signature"
    /// Lucide `bookmark` — locator page row.
    case bookmark = "bookmark"
    case sort = "arrow.up.arrow.down"
    case success = "checkmark.circle.fill"
    case warning = "exclamationmark.triangle.fill"
    case danger = "exclamationmark.octagon.fill"
    case dismiss = "xmark"
    case library = "books.vertical"
    case tag = "tag"
    /// Lucide `layout-grid` — the Properties All card.
    case grid = "square.grid.2x2"
    case list = "list.bullet"
    /// Subject types — kinds of thing a document can talk about (S5-D3 `shapes`).
    case shapes = "square.on.circle"
    /// Properties — properties hanging off a subject kind (S5-D3 `list-tree`).
    case listTree = "list.bullet.indent"
    case account = "person.crop.circle"
    /// Lucide `pin` — confirmed matches pinned to a claim (Promote summary).
    case pin = "pin"
    case sidebarToggle = "sidebar.left"
    case search = "magnifyingglass"
    case searchEmpty = "text.magnifyingglass"
    /// Design-system `user-search`: no person found yet.
    case userSearch = "person.crop.circle.badge.questionmark"
    case zoomOut = "minus.magnifyingglass"
    case zoomIn = "plus.magnifyingglass"
    case plus = "plus"
    case ellipsis = "ellipsis"
    case sortAscending = "arrow.up"
    case sortDescending = "arrow.down"
    case sortUnsorted = "chevron.up.chevron.down"
    case filter = "line.3.horizontal.decrease"
    /// Lucide `network` — Evidence graph zone glyph (S5-D2 / S5-08).
    case network = "point.3.connected.trianglepath.dotted"
    /// Lucide `layers` — group rows on the Promote page.
    case layers = "square.3.layers.3d.down.right"
    /// Lucide `refresh-cw` — a suggested Promote row that moved.
    case refresh = "arrow.clockwise"
    /// Lucide `history` — last updated on a list row.
    case history = "clock.arrow.circlepath"
    case shieldCheck = "checkmark.shield.fill"
    case calendar = "calendar"
    case textType = "textformat"
    case lock = "lock.fill"
    case trash = "trash"
    case plug = "powerplug.fill"
    case externalLink = "arrow.up.right.square"
    /// Promote: file a Subject on a handle.
    case promote = "arrow.turn.right.up"
    /// Unpromote: take a filing back (S10-44).
    case unpromote = "arrow.uturn.down"
    /// Lucide `arrow-up-right` — jump out of a blocker row.
    case arrowUpRight = "arrow.up.right"
    case fileUp = "doc.badge.plus"
    case penLine = "pencil.line"
    /// Lucide `image-up` — a picture with an upward arrow. Stroked in
    /// ``PVIcon``; `square.and.arrow.up` is the share glyph, a different mark.
    case imageUp = "pv.image.up"
    case circleDashed = "circle.dashed"
    case regionRectangle = "rectangle"
    /// Custom six-sided L glyphs — not SF Symbols (see ``PVIcon``).
    case regionLTopRight = "pv.region.l.topRight"
    case regionLTopLeft = "pv.region.l.topLeft"
    case regionLBottomRight = "pv.region.l.bottomRight"
    case regionLBottomLeft = "pv.region.l.bottomLeft"
    case regionCircle = "circle"
    case regionFreeform = "lasso"
    /// Lucide `scan-text` — Auto transcribe on a transcription field.
    case scanText = "doc.text.viewfinder"
    /// Lucide `clipboard-paste` — paste PDF selection into transcription.
    case clipboardPaste = "document.on.clipboard"
    /// Lucide `link-2` — start a bridge.
    case link = "link"
    /// Lucide `copy-plus` — duplicate a saved line into the authoring box.
    case duplicate = "plus.square.on.square"
    /// Lucide `arrow-left-right` — swap the two ends of a bridge.
    case swap = "arrow.left.arrow.right"

    fileprivate var regionLOpen: PVLOpenCorner? {
        switch self {
        case .regionLTopRight: .topRight
        case .regionLTopLeft: .topLeft
        case .regionLBottomRight: .bottomRight
        case .regionLBottomLeft: .bottomLeft
        default: nil
        }
    }
}

fileprivate enum PVLOpenCorner {
    case topRight, topLeft, bottomRight, bottomLeft
}

/// Renders a design-system icon via SF Symbols, or a stroked L hexagon.
struct PVIcon: View {
    private let symbol: PVSymbol
    private let size: CGFloat

    init(_ symbol: PVSymbol, size: CGFloat = 16) {
        self.symbol = symbol
        self.size = size
    }

    var body: some View {
        Group {
            if let open = symbol.regionLOpen {
                PVRegionLShape(open: open)
                    .stroke(style: StrokeStyle(lineWidth: max(1.15, size * 0.09), lineJoin: .miter))
            } else if symbol == .imageUp {
                PVImageUpShape()
                    .stroke(style: StrokeStyle(
                        lineWidth: max(1, size * 2 / 24),
                        lineCap: .round,
                        lineJoin: .round
                    ))
            } else {
                Image(systemName: symbol.rawValue)
                    .resizable()
                    .scaledToFit()
            }
        }
        .frame(width: size, height: size)
        .accessibilityHidden(true)
    }
}

/// Six-sided L polygon in a unit square (inset so the stroke is not clipped).
private struct PVRegionLShape: Shape {
    var open: PVLOpenCorner

    func path(in rect: CGRect) -> Path {
        let inset = min(rect.width, rect.height) * 0.14
        let box = rect.insetBy(dx: inset, dy: inset)
        let midX = box.midX
        let midY = box.midY
        let points: [CGPoint]
        switch open {
        case .topRight:
            points = [
                CGPoint(x: box.minX, y: box.minY),
                CGPoint(x: midX, y: box.minY),
                CGPoint(x: midX, y: midY),
                CGPoint(x: box.maxX, y: midY),
                CGPoint(x: box.maxX, y: box.maxY),
                CGPoint(x: box.minX, y: box.maxY),
            ]
        case .topLeft:
            points = [
                CGPoint(x: midX, y: box.minY),
                CGPoint(x: box.maxX, y: box.minY),
                CGPoint(x: box.maxX, y: box.maxY),
                CGPoint(x: box.minX, y: box.maxY),
                CGPoint(x: box.minX, y: midY),
                CGPoint(x: midX, y: midY),
            ]
        case .bottomRight:
            points = [
                CGPoint(x: box.minX, y: box.minY),
                CGPoint(x: box.maxX, y: box.minY),
                CGPoint(x: box.maxX, y: midY),
                CGPoint(x: midX, y: midY),
                CGPoint(x: midX, y: box.maxY),
                CGPoint(x: box.minX, y: box.maxY),
            ]
        case .bottomLeft:
            points = [
                CGPoint(x: box.minX, y: box.minY),
                CGPoint(x: box.maxX, y: box.minY),
                CGPoint(x: box.maxX, y: box.maxY),
                CGPoint(x: midX, y: box.maxY),
                CGPoint(x: midX, y: midY),
                CGPoint(x: box.minX, y: midY),
            ]
        }
        var path = Path()
        path.move(to: points[0])
        for point in points.dropFirst() {
            path.addLine(to: point)
        }
        path.closeSubpath()
        return path
    }
}

/// Lucide `image-up` in a 24-unit box: a picture frame and an arrow pointing up.
private struct PVImageUpShape: Shape {
    func path(in rect: CGRect) -> Path {
        let side = min(rect.width, rect.height)
        let origin = CGPoint(x: rect.midX - side / 2, y: rect.midY - side / 2)
        func pt(_ x: CGFloat, _ y: CGFloat) -> CGPoint {
            CGPoint(x: origin.x + side * x / 24, y: origin.y + side * y / 24)
        }

        var path = Path()
        let frame = CGRect(x: pt(2, 4).x, y: pt(2, 4).y, width: side * 13 / 24, height: side * 16 / 24)
        path.addPath(RoundedRectangle(cornerRadius: side * 2 / 24, style: .continuous).path(in: frame))
        path.move(to: pt(18.5, 20))
        path.addLine(to: pt(18.5, 6))
        path.move(to: pt(15, 9.5))
        path.addLine(to: pt(18.5, 6))
        path.addLine(to: pt(22, 9.5))
        return path
    }
}

#Preview {
    HStack(spacing: PVSpacing.space6) {
        PVIcon(.folderPlus)
        PVIcon(.imageUp)
        PVIcon(.regionRectangle)
        PVIcon(.regionLTopRight)
        PVIcon(.regionLTopLeft)
        PVIcon(.regionLBottomRight)
        PVIcon(.regionLBottomLeft)
        PVIcon(.regionCircle)
        PVIcon(.regionFreeform)
    }
    .foregroundStyle(PVColor.accent)
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
