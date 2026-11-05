import SwiftUI

/// Which field of a `PVTabPair` a handle points at, in layout order.
enum PVTabPairField: Hashable {
    case first, second

    var next: PVTabPairField? { self == .first ? .second : nil }
    var previous: PVTabPairField? { self == .second ? .first : nil }
}

/// Owns focus for two fields in one form, in layout order: Tab in the first
/// moves to the second, Shift-Tab in the second moves back. The AppKit key
/// loop skips a text area that follows a text field, so a form pairing a
/// `PVInput` with a `PVTextArea` wraps both here and passes each its handle.
/// The body is the content unchanged, so the parent stack lays the fields out.
struct PVTabPair<Content: View>: View {
    @FocusState private var field: PVTabPairField?
    private let content: (PVFieldFocus, PVFieldFocus) -> Content

    init(@ViewBuilder content: @escaping (PVFieldFocus, PVFieldFocus) -> Content) {
        self.content = content
    }

    private func handle(_ f: PVTabPairField) -> PVFieldFocus {
        PVFieldFocus($field, equals: f, next: f.next, previous: f.previous)
    }

    var body: some View {
        content(handle(.first), handle(.second))
    }
}

#Preview {
    PVTabPair { first, second in
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVInput(text: .constant("Family album, Norfolk"), prompt: "Title", focus: first)
            PVTextArea(
                text: .constant("Held by Mary; consulted at the NRO."),
                prompt: "Description",
                focus: second
            )
        }
    }
    .padding(PVSpacing.space9)
    .frame(width: 420)
    .background(PVColor.surfacePage)
}
