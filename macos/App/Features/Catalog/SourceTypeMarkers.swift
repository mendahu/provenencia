import SwiftUI

/// Resting source-type label on identity chrome (Source page header, Person
/// page Sources rows): the kit's `PVTag` with the record colour as its dot.
/// Today every type shares the marriage/plum dot; per-type categorical
/// coloring can land when type keys are wired to record tones.
struct CatalogSourceTypePill: View {
    let label: String

    var body: some View {
        PVTag(text: label, dot: PVColor.recordMarriage)
    }
}

#Preview {
    CatalogSourceTypePill(label: "Civil registration")
        .padding(PVSpacing.space9)
        .background(PVColor.surfacePage)
}
