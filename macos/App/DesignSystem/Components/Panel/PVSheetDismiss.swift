/// What Escape or a click outside a sheet does. Running work is left alone. The
/// discard question, when it shows, steps back to the form. A dirty sheet asks
/// before it closes. A clean or read-only sheet closes. The visible Cancel
/// button is not this path: it closes at once.
enum PVSheetDismiss: Equatable {
    case ignore
    case keepEditing
    case confirmDiscard
    case close

    /// `innerEditOpen` is true while content reports an open inner edit (a row in
    /// edit, an add row): a click outside then does nothing, since the edit owns
    /// its own dismissal.
    static func action(
        isRunning: Bool,
        isDirty: Bool,
        isConfirmingDiscard: Bool,
        innerEditOpen: Bool = false
    ) -> PVSheetDismiss {
        if isRunning { return .ignore }
        if innerEditOpen && !isConfirmingDiscard { return .ignore }
        if isConfirmingDiscard { return .keepEditing }
        if isDirty { return .confirmDiscard }
        return .close
    }
}
