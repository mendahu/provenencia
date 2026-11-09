import SwiftUI

/// Keeps the keyboard on a stop that a view lets go: the Regular/Wide line rebuilds a lazy
/// grid and folds cards away (Wide rule 5), a lazy grid drops a focused tile, a reload takes
/// a card. The view holds one in `@State`, applies `focusKeeper(...)`, and ends a viewer or
/// sheet with `FocusKeeper.hand`. `FocusClaim` stays the writer.
///
/// The state is plain data so the rules test without a view; the statics are the one place
/// a stop is claimed.
struct FocusKeeper<Stop: Hashable>: Hashable {
    /// The last stop that had the keyboard, kept to the next main-actor turn after focus
    /// goes, so a width change in the same update still sees it while a click elsewhere
    /// clears it.
    private(set) var last: Stop?
    /// Where the keyboard goes when `last` is gone (a card's row), if anywhere.
    private(set) var lastFallback: Stop?
    /// A stop to give the keyboard back once its view exists again, until its expiry.
    private(set) var pending: Stop?

    enum Lost: Equatable {
        case none
        /// A pending stop's view let it go: claim it again.
        case reclaim
        /// The last stop is no longer shown: the keyboard goes to this one.
        case fallback(Stop)
    }

    /// Focus moved to `stop`. The pending stop ends when the keyboard went to another stop
    /// (not the pending stop or its `revealedPending` form); the pending stop itself stays
    /// until its expiry, so its view's `onAppear` can still claim it.
    mutating func took(_ stop: Stop, fallback: Stop?, revealedPending: Stop?) {
        last = stop
        lastFallback = fallback
        if pending != nil, pending != stop, revealedPending != stop { pending = nil }
    }

    /// Focus went nil.
    mutating func lost(isShown: (Stop) -> Bool) -> Lost {
        if pending != nil { return .reclaim }
        if let last, !isShown(last), let lastFallback { return .fallback(lastFallback) }
        return .none
    }

    /// Focus stayed nil: a click elsewhere, nothing is pulled back.
    mutating func forgetLast() {
        last = nil
        lastFallback = nil
    }

    /// The stop to hand back on a width change, or nil when nothing had the keyboard: the
    /// focused stop, else the last, replaced by its fallback when it is no longer shown.
    func widthStop(current: Stop?, isShown: (Stop) -> Bool) -> Stop? {
        guard let stop = current ?? last else { return nil }
        return isShown(stop) ? stop : lastFallback
    }

    mutating func give(_ stop: Stop) { pending = stop }

    mutating func expire(_ stop: Stop) {
        if pending == stop { pending = nil }
    }

    /// The stop a claim of the pending stop writes: `scroll` brings the pending stop into view,
    /// then `map` names the stop that can take the keyboard. Nil when nothing is pending. Only
    /// a claim scrolls; recording a focus move never does.
    func claimed(map: (Stop) -> Stop, scroll: (Stop) -> Void) -> Stop? {
        guard let pending else { return nil }
        scroll(pending)
        return map(pending)
    }

    /// Claims the pending stop: `scroll` brings it into view, `map` names the stop that can take
    /// the keyboard, then `FocusClaim` writes it.
    @MainActor
    static func reclaim(
        keeper: Binding<FocusKeeper>,
        focus: FocusState<Stop?>.Binding,
        map: @escaping (Stop) -> Stop,
        scroll: @escaping (Stop) -> Void
    ) {
        guard let stop = keeper.wrappedValue.claimed(map: map, scroll: scroll) else { return }
        FocusClaim.claim(focus, Optional(stop))
    }

    /// Gives `stop` the keyboard once its view exists, and ends the pending stop a moment later
    /// (a landing does not end it), so a view that never dropped does not take the keyboard
    /// back after a click elsewhere.
    @MainActor
    static func hand(
        _ stop: Stop,
        keeper: Binding<FocusKeeper>,
        focus: FocusState<Stop?>.Binding,
        map: @escaping (Stop) -> Stop,
        scroll: @escaping (Stop) -> Void
    ) {
        keeper.wrappedValue.give(stop)
        reclaim(keeper: keeper, focus: focus, map: map, scroll: scroll)
        DispatchQueue.main.asyncAfter(deadline: .now() + .milliseconds(400)) {
            keeper.wrappedValue.expire(stop)
        }
    }
}

extension View {
    /// Keeps the keyboard across a width change, a lazy view letting a stop go and a reload
    /// that removes the focused stop (see `FocusKeeper`). `isShown` says whether a stop is
    /// drawn now; `fallback` is where the keyboard goes when a stop is folded away; `map` names
    /// the stop that can take the keyboard in place of a stop (pure); `scroll` brings a stop into
    /// view, and only a claim of a pending stop calls it.
    func focusKeeper<Stop: Hashable>(
        _ keeper: Binding<FocusKeeper<Stop>>,
        focus: FocusState<Stop?>.Binding,
        isShown: @escaping (Stop) -> Bool,
        fallback: @escaping (Stop) -> Stop?,
        map: @escaping (Stop) -> Stop,
        scroll: @escaping (Stop) -> Void
    ) -> some View {
        modifier(FocusKeeping(keeper: keeper, focus: focus, isShown: isShown, fallback: fallback, map: map, scroll: scroll))
    }
}

private struct FocusKeeping<Stop: Hashable>: ViewModifier {
    let keeper: Binding<FocusKeeper<Stop>>
    let focus: FocusState<Stop?>.Binding
    let isShown: (Stop) -> Bool
    let fallback: (Stop) -> Stop?
    let map: (Stop) -> Stop
    let scroll: (Stop) -> Void

    @Environment(\.pvWidthClass) private var widthClass

    func body(content: Content) -> some View {
        content
            .onChange(of: focus.wrappedValue) { _, stop in
                if let stop {
                    keeper.wrappedValue.took(
                        stop,
                        fallback: fallback(stop),
                        revealedPending: keeper.wrappedValue.pending.map(map)
                    )
                } else {
                    switch keeper.wrappedValue.lost(isShown: isShown) {
                    case .reclaim:
                        FocusKeeper.reclaim(keeper: keeper, focus: focus, map: map, scroll: scroll)
                    case .fallback(let stop):
                        FocusClaim.claim(focus, Optional(stop))
                    case .none:
                        break
                    }
                    DispatchQueue.main.async {
                        if focus.wrappedValue == nil { keeper.wrappedValue.forgetLast() }
                    }
                }
            }
            .onChange(of: widthClass) {
                if let stop = keeper.wrappedValue.widthStop(current: focus.wrappedValue, isShown: isShown) {
                    FocusKeeper.hand(stop, keeper: keeper, focus: focus, map: map, scroll: scroll)
                }
            }
    }
}
