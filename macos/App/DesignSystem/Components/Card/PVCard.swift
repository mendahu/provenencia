import SwiftUI

/// Surface container chrome — mirrors `components/core/Card.jsx` (card fill,
/// hairline border, radius-md, optional shallow shadow). Title / actions /
/// footer header slots from the web kit aren't ported yet; call sites compose
/// section headers outside the card. `border: .dashed` is a Provenencia
/// addition for provisional rows (metadata suggestions), not on the web Card.
struct PVCard<Content: View>: View {
    enum Tone {
        case card, raised, sunken

        fileprivate var fill: Color {
            switch self {
            case .card: PVColor.surfaceCard
            case .raised: PVColor.surfaceRaised
            case .sunken: PVColor.surfaceSunken
            }
        }
    }

    enum Border {
        case solid
        case dashed
    }

    private let tone: Tone
    private let border: Border
    private let cornerRadius: CGFloat
    private let elevated: Bool
    private let padding: CGFloat?
    private let content: Content

    init(
        tone: Tone = .card,
        border: Border = .solid,
        cornerRadius: CGFloat = PVRadius.md,
        elevated: Bool = false,
        padding: CGFloat? = nil,
        @ViewBuilder content: () -> Content
    ) {
        self.tone = tone
        self.border = border
        self.cornerRadius = cornerRadius
        self.elevated = elevated
        self.padding = padding
        self.content = content()
    }

    var body: some View {
        let shape = RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
        content
            .padding(padding ?? 0)
            .background(tone.fill)
            .clipShape(shape)
            .overlay(borderOverlay(shape: shape))
            .pvShadow(elevated ? PVElevation.sm : [])
    }

    private func borderOverlay(shape: RoundedRectangle) -> some View {
        switch border {
        case .solid:
            shape.strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        case .dashed:
            shape.strokeBorder(
                PVColor.borderDefault,
                style: StrokeStyle(lineWidth: 1, dash: [4, 3])
            )
        }
    }
}

#Preview("Solid elevated") {
    PVCard(elevated: true, padding: PVSpacing.space8) {
        Text("Vital records")
            .font(PVFont.display(size: PVTypeScale.h3))
            .foregroundStyle(PVColor.textDisplay)
            .frame(maxWidth: .infinity, alignment: .leading)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}

#Preview("Dashed") {
    PVCard(border: .dashed, cornerRadius: PVRadius.sm, padding: PVSpacing.space5) {
        Text("Suggested field")
            .font(PVFont.body(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textSecondary)
            .frame(maxWidth: .infinity, alignment: .leading)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
