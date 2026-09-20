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
    case plus = "plus"
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
}

/// Renders a design-system icon via SF Symbols.
struct PVIcon: View {
    private let symbol: PVSymbol
    private let size: CGFloat

    init(_ symbol: PVSymbol, size: CGFloat = 16) {
        self.symbol = symbol
        self.size = size
    }

    var body: some View {
        Image(systemName: symbol.rawValue)
            .resizable()
            .scaledToFit()
            .frame(width: size, height: size)
            .accessibilityHidden(true)
    }
}

#Preview {
    HStack(spacing: PVSpacing.space6) {
        PVIcon(.folderPlus)
        PVIcon(.folderOpen)
        PVIcon(.check)
        PVIcon(.info)
        PVIcon(.chevronDown)
    }
    .foregroundStyle(PVColor.accent)
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
