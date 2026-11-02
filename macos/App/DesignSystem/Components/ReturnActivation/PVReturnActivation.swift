import SwiftUI

extension View {
    /// Makes Return activate this view when it has keyboard focus. The view
    /// owns its focus state; Space stays whatever the control already does (a
    /// `Button` activates on it). Content-agnostic.
    func pvActivatesOnReturn(_ action: @escaping () -> Void) -> some View {
        modifier(PVReturnActivation(action: action))
    }

    /// Return activates this view while `isFocused` is true, for a caller that
    /// owns its own focus binding (an enum `@FocusState`) and applies
    /// `.focused` itself.
    func pvActivatesOnReturn(when isFocused: Bool, _ action: @escaping () -> Void) -> some View {
        onKeyPress(.return) {
            guard isFocused else { return .ignored }
            action()
            return .handled
        }
    }
}

private struct PVReturnActivation: ViewModifier {
    let action: () -> Void

    @FocusState private var focused: Bool

    func body(content: Content) -> some View {
        content
            .focused($focused)
            .pvActivatesOnReturn(when: focused, action)
    }
}
