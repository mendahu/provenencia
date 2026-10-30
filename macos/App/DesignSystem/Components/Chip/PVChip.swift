import SwiftUI

/// Meaning-free selectable pill — UI Lego for exclusive options (source
/// credibility, date kind/qualifier, later pickers). Product semantics
/// (evidence grades, credibility keys) live in feature helpers that *compose*
/// this control; do not put domain vocabulary here.
///
/// Not a direct port of a web-kit component — extracted from Provenencia
/// Source-page / DateValue chrome during structural cleanup. Sibling to
/// display-only `PVBadge`. Call sites own `.accessibilityIdentifier`.
struct PVChip: View {
    /// Selected fill / stroke palette. `.neutral` is paper/card chrome
    /// (segmented kind toggle); tonal cases match `PVBadgeTone` soft fills.
    enum Tone {
        case neutral, accent, success, warning, danger, info

        fileprivate var colors: (fill: Color, foreground: Color, stroke: Color) {
            switch self {
            case .neutral:
                (PVColor.surfaceCard, PVColor.textPrimary, PVColor.borderDefault)
            case .accent:
                (PVColor.accentSoft, PVColor.accentSoftForeground, PVColor.accentLine)
            case .success:
                (PVColor.successSoft, PVColor.successForeground, PVColor.success)
            case .warning:
                (PVColor.warningSoft, PVColor.warningForeground, PVColor.warning)
            case .danger:
                (PVColor.dangerSoft, PVColor.dangerForeground, PVColor.danger)
            case .info:
                (PVColor.infoSoft, PVColor.infoForeground, PVColor.info)
            }
        }
    }

    private let label: Text
    private let isSelected: Bool
    /// Dashed border — e.g. “standard” credibility before an assessment exists.
    private let isDashed: Bool
    private let tone: Tone
    private let size: PVControlSize
    private let glyph: PVGlyph?
    /// Stretch to fill (segmented track).
    private let expands: Bool
    /// Drop a small elevation when selected (segmented kind toggle).
    private let selectionLift: Bool
    private let action: () -> Void

    /// Fixed UI copy (e.g. date kind / qualifier labels).
    init(
        _ titleKey: LocalizedStringResource,
        isSelected: Bool,
        tone: Tone = .neutral,
        size: PVControlSize = .md,
        glyph: PVGlyph? = nil,
        isDashed: Bool = false,
        expands: Bool = false,
        selectionLift: Bool = false,
        action: @escaping () -> Void
    ) {
        label = Text(titleKey)
        self.isSelected = isSelected
        self.isDashed = isDashed
        self.tone = tone
        self.size = size
        self.glyph = glyph
        self.expands = expands
        self.selectionLift = selectionLift
        self.action = action
    }

    /// Data label (e.g. catalog credibility grade label).
    init(
        text: String,
        isSelected: Bool,
        tone: Tone = .neutral,
        size: PVControlSize = .md,
        glyph: PVGlyph? = nil,
        isDashed: Bool = false,
        expands: Bool = false,
        selectionLift: Bool = false,
        action: @escaping () -> Void
    ) {
        label = Text(text)
        self.isSelected = isSelected
        self.isDashed = isDashed
        self.tone = tone
        self.size = size
        self.glyph = glyph
        self.expands = expands
        self.selectionLift = selectionLift
        self.action = action
    }

    var body: some View {
        let colors = tone.colors
        Button(action: action) {
            content
                .font(PVFont.body(
                    size: size == .lg ? PVTypeScale.bodySmall : PVTypeScale.caption,
                    weight: isSelected ? PVFontWeight.semibold : PVFontWeight.regular
                ))
                .foregroundStyle(foreground(colors: colors))
                .padding(.horizontal, expands ? PVSpacing.space4 : PVSpacing.space5)
                .frame(maxWidth: expands ? .infinity : nil)
                .frame(height: size.height)
                .background(background(colors: colors))
                .overlay(border(colors: colors))
                .clipShape(RoundedRectangle(cornerRadius: chipRadius, style: .continuous))
                .contentShape(Rectangle())
                .pvShadow(selectionLift && isSelected ? PVElevation.sm : [])
        }
        .buttonStyle(PVChipStyle(cornerRadius: chipRadius))
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }

    /// The label, led by the glyph when there is one. A stretched chip in a
    /// narrow column drops the glyph before the label would truncate.
    @ViewBuilder
    private var content: some View {
        if let glyph {
            let withGlyph = HStack(spacing: PVSpacing.space3) {
                PVGlyphView(glyph, size: size.chipGlyphSize)
                label
            }
            if expands {
                ViewThatFits(in: .horizontal) {
                    withGlyph
                    label
                }
            } else {
                withGlyph
            }
        } else {
            label
        }
    }

    private var chipRadius: CGFloat {
        expands ? 3 : PVRadius.sm
    }

    private func foreground(colors: (fill: Color, foreground: Color, stroke: Color)) -> Color {
        if isSelected {
            return colors.foreground
        }
        return expands ? PVColor.textMuted : PVColor.textSecondary
    }

    @ViewBuilder
    private func background(colors: (fill: Color, foreground: Color, stroke: Color)) -> some View {
        let shape = RoundedRectangle(cornerRadius: chipRadius, style: .continuous)
        if isSelected {
            shape.fill(colors.fill)
        } else if expands {
            shape.fill(Color.clear)
        } else {
            shape.fill(PVColor.surfaceRaised)
        }
    }

    private func border(colors: (fill: Color, foreground: Color, stroke: Color)) -> some View {
        RoundedRectangle(cornerRadius: chipRadius, style: .continuous)
            .strokeBorder(
                borderColor(colors: colors),
                style: StrokeStyle(lineWidth: expands ? 0 : 1, dash: isDashed ? PVStroke.dashPattern : [])
            )
    }

    private func borderColor(colors: (fill: Color, foreground: Color, stroke: Color)) -> Color {
        if expands { return .clear }
        if isSelected { return colors.stroke }
        return PVColor.borderDefault
    }
}

private extension PVControlSize {
    /// Glyph beside a chip's label.
    var chipGlyphSize: CGFloat {
        switch self {
        case .sm: return 12
        case .md: return 13
        case .lg: return 14
        }
    }
}

/// Draws keyboard focus for a chip, as `PVButtonBody` does for a button. A
/// custom `ButtonStyle` gets no focus indication of its own.
private struct PVChipStyle: ButtonStyle {
    let cornerRadius: CGFloat

    func makeBody(configuration: Configuration) -> some View {
        PVChipStyleBody(configuration: configuration, cornerRadius: cornerRadius)
    }
}

private struct PVChipStyleBody: View {
    let configuration: ButtonStyleConfiguration
    let cornerRadius: CGFloat
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        configuration.label
            .pvFocusRing(isFocused, cornerRadius: cornerRadius)
    }
}

/// Layout wrapper for a row of `PVChip`s.
struct PVChipGroup<Content: View>: View {
    enum Style {
        /// Spaced chips (credibility, date qualifier).
        case loose
        /// Packed equal-width chips in a sunken track (date kind).
        case segmented
    }

    var style: Style = .loose
    var spacing: CGFloat = PVSpacing.space3
    @ViewBuilder var content: () -> Content

    var body: some View {
        switch style {
        case .loose:
            HStack(spacing: spacing) {
                content()
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        case .segmented:
            HStack(spacing: 0) {
                content()
            }
            .padding(2)
            .background(PVColor.surfaceSunken)
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(PVColor.borderSubtle, lineWidth: 1)
            )
        }
    }
}

#Preview("Loose tones") {
    PVChipGroup(style: .loose, spacing: PVSpacing.space4) {
        PVChip(text: "Low trust", isSelected: false, tone: .danger, action: {})
        PVChip(text: "Standard", isSelected: true, tone: .accent, isDashed: true, action: {})
        PVChip(text: "High trust", isSelected: false, tone: .success, action: {})
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}

#Preview("Segmented") {
    PVChipGroup(style: .segmented) {
        PVChip("Point", isSelected: true, expands: true, selectionLift: true, action: {})
        PVChip("Range", isSelected: false, expands: true, selectionLift: true, action: {})
    }
    .frame(width: 280)
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
