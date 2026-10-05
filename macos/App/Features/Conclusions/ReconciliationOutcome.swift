import SwiftUI

/// An outcome as one mark and one phrase (board S9-D5). The mark is
/// decorative; the phrase carries the meaning, so it reads without colour.
struct ReconciliationOutcome: View {
    let mark: PVSymbol
    let phrase: String

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            PVIcon(mark, size: 14)
                .alignmentGuide(.firstTextBaseline) { $0[VerticalAlignment.center] + 4 }
            Text(phrase)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
        }
        .foregroundStyle(PVColor.textSecondary)
    }
}
