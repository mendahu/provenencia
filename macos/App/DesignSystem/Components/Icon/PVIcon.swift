import SwiftUI

/// Compile-checked SF Symbol names used by the design system.
enum PVSymbol: String {
    case folderPlus = "folder.badge.plus"
    case folderOpen = "folder"
    case check = "checkmark"
    case info = "info.circle"
    case chevronDown = "chevron.down"
    case chevronBack = "chevron.left"
    case chevronForward = "chevron.right"
    case photo = "photo"
    case scrollText = "doc.text"
    /// Lucide `file` — locator floor (entire artifact).
    case file = "doc"
    /// Lucide `bookmark` — locator page row.
    case bookmark = "bookmark"
    case sort = "arrow.up.arrow.down"
    case success = "checkmark.circle.fill"
    case warning = "exclamationmark.triangle.fill"
    case danger = "exclamationmark.octagon.fill"
    case dismiss = "xmark"
    case library = "books.vertical"
    case tag = "tag"
    case list = "list.bullet"
    /// Subject types — kinds of thing a document can talk about (S5-D3 `shapes`).
    case shapes = "square.on.circle"
    /// Subject fields — properties hanging off a subject kind (S5-D3 `list-tree`).
    case listTree = "list.bullet.indent"
    case account = "person.crop.circle"
    case sidebarToggle = "sidebar.left"
    case search = "magnifyingglass"
    case searchEmpty = "text.magnifyingglass"
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
    case shieldCheck = "checkmark.shield.fill"
    case calendar = "calendar"
    case textType = "textformat"
    case lock = "lock.fill"
    case trash = "trash"
    case plug = "powerplug.fill"
    case externalLink = "arrow.up.right.square"
    case fileUp = "doc.badge.plus"
    case penLine = "pencil.line"
    case imageUp = "square.and.arrow.up"
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

#Preview {
    HStack(spacing: PVSpacing.space6) {
        PVIcon(.folderPlus)
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
