import SwiftUI

/// A bare button's look: the label as it is, a hit shape, and the focus ring
/// at `cornerRadius`; no other chrome. For a card, a tile or a preview that is
/// itself the button (hover lift is `.pvLiftHover`).
struct PVFocusRingButtonStyle: ButtonStyle {
    var cornerRadius: CGFloat = PVRadius.sm

    func makeBody(configuration: Configuration) -> some View {
        PVFocusRingButtonBody(configuration: configuration, cornerRadius: cornerRadius)
    }
}

private struct PVFocusRingButtonBody: View {
    let configuration: ButtonStyleConfiguration
    let cornerRadius: CGFloat

    /// Read here, in the button's focus scope (see `PVButtonBody`).
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        configuration.label
            .contentShape(Rectangle())
            .pvFocusRing(isFocused, cornerRadius: cornerRadius)
    }
}
