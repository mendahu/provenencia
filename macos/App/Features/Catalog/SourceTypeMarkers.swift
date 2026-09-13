import SwiftUI

/// Resting source-type label on identity chrome (Source page header, later
/// list chips). Uses record-type soft tokens — not raw `PVPalette` at the
/// call site. Today every type shares the marriage/plum pair; per-type
/// categorical coloring can land when type keys are wired to record tones.
struct CatalogSourceTypePill: View {
    let label: String
    var iconKey: String? = nil

    var body: some View {
        HStack(spacing: PVSpacing.space2) {
            if let iconKey, !iconKey.isEmpty {
                PVEvidenceIcon(
                    PVEvidenceIconKey(catalogKey: iconKey),
                    size: .inline,
                    decorative: true
                )
            }
            Text(label)
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.recordMarriage)
        }
        .padding(.horizontal, PVSpacing.space4)
        .padding(.vertical, PVSpacing.space2)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.recordMarriageSoft)
        )
    }
}

#Preview {
    CatalogSourceTypePill(label: "Civil registration", iconKey: "type_certificate")
        .padding(PVSpacing.space9)
        .background(PVColor.surfacePage)
}
