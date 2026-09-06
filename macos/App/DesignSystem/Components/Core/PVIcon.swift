import SwiftUI

/// Compile-checked SF Symbol names used by the design system.
enum PVSymbol: String {
    case folderPlus = "folder.badge.plus"
    case folderOpen = "folder"
    case check = "checkmark"
    case info = "info.circle"
    case chevronDown = "chevron.down"
    case success = "checkmark.circle.fill"
    case warning = "exclamationmark.triangle.fill"
    case danger = "exclamationmark.octagon.fill"
    case dismiss = "xmark"
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
