import SwiftUI

/// Copy a component shows or speaks: a static catalog resource
/// (`L10n.NameValue.partsAdd`) or a `String` an `L10n` format function already
/// resolved and formatted (`L10n.NameValue.partRemove(position: 2)`).
///
/// Copy with arguments is always a format function returning `String`, never an
/// interpolated `LocalizedStringResource`, so components take either shape
/// through this one protocol and render it with ``pvText``.
protocol PVCopy {
    /// The copy as a `Text`, resolved from the catalog or shown verbatim.
    var pvText: Text { get }
}

extension LocalizedStringResource: PVCopy {
    var pvText: Text { Text(self) }
}

extension String: PVCopy {
    var pvText: Text { Text(verbatim: self) }
}
