import SwiftUI

/// The width class of the workspace pane: Regular below
/// `PVSpacing.widthWideBreakpoint`, Wide at or above it. Measured once by
/// `pvMeasuresWidthClass()` at `WorkspaceDestinationHost`; pages read
/// `@Environment(\.pvWidthClass)`. See
/// [§ 9 of macos-client-patterns.md](../../../../../docs/macos-client-patterns.md#9-window-widths-regular-and-wide).
enum PVWidthClass: Equatable, Sendable {
    case regular
    case wide

    init(paneWidth: CGFloat) {
        self = paneWidth >= PVSpacing.widthWideBreakpoint ? .wide : .regular
    }
}

private struct PVWidthClassKey: EnvironmentKey {
    static let defaultValue: PVWidthClass = .regular
}

extension EnvironmentValues {
    var pvWidthClass: PVWidthClass {
        get { self[PVWidthClassKey.self] }
        set { self[PVWidthClassKey.self] = newValue }
    }
}

private struct PVMeasuresWidthClass: ViewModifier {
    @State private var widthClass: PVWidthClass = .regular

    func body(content: Content) -> some View {
        content
            .environment(\.pvWidthClass, widthClass)
            .onGeometryChange(for: PVWidthClass.self) {
                PVWidthClass(paneWidth: $0.size.width)
            } action: {
                widthClass = $0
            }
    }
}

extension View {
    /// Measures the view it is applied to and gives that width class to the
    /// view's content, so the view must fill the space to be measured.
    func pvMeasuresWidthClass() -> some View {
        modifier(PVMeasuresWidthClass())
    }
}
