import SwiftUI

/// A pill with a colour dot — mirrors `components/PVTag.jsx` in the design
/// system bundle (raised surface, subtle border, a 7pt dot, secondary caption
/// text). Deliberately not ported: `onRemove` and `interactive` (no caller
/// yet; add them with the first one), and a built-in fixed size — a Tag shows
/// data of any length, so it takes the width it is offered and truncates its
/// text; a caller that must keep it whole adds `.fixedSize()`. The 3, 9 and 7pt
/// measures are the web kit's own, so they stay literals.
struct PVTag: View {
    let text: String
    var dot: Color = PVColor.textMuted

    var body: some View {
        HStack(spacing: PVSpacing.space3) {
            Circle()
                .fill(dot)
                .frame(width: 7, height: 7)
            Text(text)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .lineLimit(1)
                .truncationMode(.tail)
        }
        .padding(.top, 3)
        .padding(.bottom, 3)
        .padding(.leading, PVSpacing.space4)
        .padding(.trailing, 9)
        .background(Capsule().fill(PVColor.surfaceRaised))
        .overlay(Capsule().strokeBorder(PVColor.borderSubtle, lineWidth: 1))
        .accessibilityElement(children: .combine)
    }
}

#Preview {
    VStack(alignment: .leading, spacing: PVSpacing.space5) {
        PVTag(text: "Census")
        PVTag(text: "Civil registration", dot: PVColor.recordMarriage)
        PVTag(text: "Church register (parish copy)", dot: PVColor.recordMarriage)
            .frame(width: 120)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
