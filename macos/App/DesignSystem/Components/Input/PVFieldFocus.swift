import SwiftUI

/// The one focus handle every kit field takes (`PVInput`, `PVTextArea`,
/// `PVComboBox`, `PVSelect`, `PVSegmentedControl`). The field applies `.focused($binding, equals:)`
/// itself, so a write to the host's `FocusState` lands in the field. A handle
/// may also carry a Tab neighbour: the field to move to on Tab (`next`) and on
/// Shift-Tab (`previous`). With no neighbour a field leaves Tab to the key loop.
struct PVFieldFocus {
    let isActive: () -> Bool
    let focus: () -> Void
    let resign: () -> Void
    let apply: (AnyView) -> AnyView
    let next: (@MainActor () -> Void)?
    let previous: (@MainActor () -> Void)?

    init(
        isActive: @escaping () -> Bool,
        focus: @escaping () -> Void,
        resign: @escaping () -> Void,
        apply: @escaping (AnyView) -> AnyView,
        next: (@MainActor () -> Void)? = nil,
        previous: (@MainActor () -> Void)? = nil
    ) {
        self.isActive = isActive
        self.focus = focus
        self.resign = resign
        self.apply = apply
        self.next = next
        self.previous = previous
    }

    @MainActor
    init<Value: Hashable>(
        _ binding: FocusState<Value?>.Binding,
        equals value: Value,
        next: Value? = nil,
        previous: Value? = nil
    ) {
        let neighbours = Self.neighbours(next: next, previous: previous) { neighbour in
            FocusClaim.claim(binding, neighbour)
        }
        self.init(
            isActive: { binding.wrappedValue == value },
            focus: { binding.wrappedValue = value },
            resign: {
                if binding.wrappedValue == value { binding.wrappedValue = nil }
            },
            apply: { view in AnyView(view.focused(binding, equals: value)) },
            next: neighbours.next,
            previous: neighbours.previous
        )
    }

    /// The claim a Tab press takes: `previous` for Shift-Tab, else `next`.
    /// Nil leaves the key to the key loop.
    func tabClaim(shift: Bool) -> (@MainActor () -> Void)? {
        shift ? previous : next
    }

    /// True when Tab or Shift-Tab has somewhere to go on this handle.
    var hasNeighbour: Bool { next != nil || previous != nil }

    /// Takes a Tab press for the neighbour it names: runs the claim and reports
    /// the key handled. With no neighbour the key is left to the key loop.
    /// `KeyPress` has no test initialiser, so the caller passes the Shift state.
    @MainActor
    func claimTab(shift: Bool) -> KeyPress.Result {
        guard let claim = tabClaim(shift: shift) else { return .ignored }
        claim()
        return .handled
    }

    /// Turns a neighbour value into the claim that moves focus to it; a nil
    /// neighbour stays nil.
    @MainActor
    static func neighbours<Value>(
        next: Value?,
        previous: Value?,
        claim: @escaping @MainActor (Value) -> Void
    ) -> (next: (@MainActor () -> Void)?, previous: (@MainActor () -> Void)?) {
        var nextClaim: (@MainActor () -> Void)?
        var previousClaim: (@MainActor () -> Void)?
        if let next { nextClaim = { claim(next) } }
        if let previous { previousClaim = { claim(previous) } }
        return (next: nextClaim, previous: previousClaim)
    }
}

extension KeyEquivalent {
    /// AppKit reports Shift-Tab as this character (U+0019), not as `.tab` with
    /// Shift, so `.onKeyPress(.tab)` never sees it.
    static let backTab = KeyEquivalent("\u{19}")
}

/// Applies a host's `PVFieldFocus` to a field, or the field's own focus when
/// it has no host.
struct PVFieldFocusModifier: ViewModifier {
    let host: PVFieldFocus?
    var own: FocusState<Bool>.Binding

    @ViewBuilder
    func body(content: Content) -> some View {
        if let host {
            host.apply(AnyView(content))
        } else {
            content.focused(own)
        }
    }
}

/// The one resolution of "the host's handle, else the field's own focus", so
/// no field repeats it.
extension Optional where Wrapped == PVFieldFocus {
    func isActive(own: Bool) -> Bool {
        self?.isActive() ?? own
    }

    func focus(own: FocusState<Bool>.Binding) {
        if let self { self.focus() } else { own.wrappedValue = true }
    }

    func resign(own: FocusState<Bool>.Binding) {
        if let self { self.resign() } else { own.wrappedValue = false }
    }
}
