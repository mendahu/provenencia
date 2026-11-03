import SwiftUI

/// How a title draws itself. A resolved value is plain, a working label is
/// italic, and a ref is mono. Persons, Events, and Places share it. A name may
/// carry a styled copy whose emphasized runs draw italic.
enum PVTitleSource: Equatable {
    case name(String, styled: AttributedString? = nil, spoken: String? = nil)
    case label(String)
    case ref(String)

    var text: String {
        switch self {
        case .name(let text, _, _): text
        case .label(let text), .ref(let text): text
        }
    }

    /// What VoiceOver reads: the spoken copy of a decorated name, else `text`.
    var spokenText: String {
        if case .name(_, _, let spoken?) = self { return spoken }
        return text
    }
}

/// Applies a line limit only when the caller gave one.
private struct PVTitleLineLimit: ViewModifier {
    let lineLimit: Int?

    func body(content: Content) -> some View {
        if let lineLimit {
            content.lineLimit(lineLimit).truncationMode(.tail)
        } else {
            content
        }
    }
}

/// A title in one of two sizes: the list row's, or the detail page's display
/// size. Name plain, label italic, ref mono, at both. Draws a `+N` badge when
/// other names survived; the badge keeps its width and the title truncates.
struct PVTitleText: View {
    enum Style: Equatable {
        case row
        case display(compact: Bool)
    }

    let source: PVTitleSource
    let style: Style
    let extraCount: Int
    /// The most lines the title takes. Nil is the style's own: one line for `.row`,
    /// and no limit for `.display`.
    let lineLimit: Int?
    let emphasized: Bool

    init(
        _ source: PVTitleSource,
        style: Style = .row,
        extraCount: Int = 0,
        lineLimit: Int? = nil,
        emphasized: Bool = false
    ) {
        self.source = source
        self.style = style
        self.extraCount = extraCount
        self.lineLimit = lineLimit
        self.emphasized = emphasized
    }

    var body: some View {
        switch style {
        case .row:
            HStack(spacing: PVSpacing.space4) {
                rowText
                    .fontWeight(emphasized ? .semibold : nil)
                    .lineLimit(lineLimit ?? 1)
                    .truncationMode(.tail)
                if extraCount > 0 {
                    PVBadge.moreCount(extraCount)
                        .layoutPriority(1)
                }
            }
        case .display(let compact):
            displayText(compact: compact)
                .modifier(PVTitleLineLimit(lineLimit: lineLimit))
        }
    }

    @ViewBuilder
    private var rowText: some View {
        switch source {
        case .name(let text, let styled, _):
            if let styled {
                Text(Self.italicizingEmphasis(styled, font: PVListText.titleFont))
            } else {
                Text(verbatim: text)
            }
        case .label(let text):
            Text(verbatim: text).italic()
        case .ref(let text):
            Text(verbatim: text)
                .font(PVFont.mono(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textSecondary)
        }
    }

    @ViewBuilder
    private func displayText(compact: Bool) -> some View {
        let size = compact ? PVTypeScale.h3 : PVTypeScale.display3
        let font = PVFont.display(size: size, weight: PVFontWeight.medium)
        switch source {
        case .name(let text, let styled, let spoken):
            Group {
                if let styled {
                    Text(Self.italicizingEmphasis(styled, font: font))
                } else {
                    Text(verbatim: text)
                }
            }
            .font(font)
            .tracking(size * PVTracking.display)
            .foregroundStyle(PVColor.textDisplay)
            .accessibilityLabel(Text(verbatim: spoken ?? text))
            .accessibilityAddTraits(.isHeader)
        case .label(let text):
            Text(verbatim: text)
                .font(font.italic())
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        case .ref(let text):
            Text(verbatim: text)
                .font(PVFont.mono(size: compact ? PVTypeScale.h4 : PVTypeScale.h1, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        }
    }

    /// `styled` with `font` italicized on each emphasized run. The italic is set explicitly,
    /// so it does not depend on how SwiftUI draws `.emphasized` in the custom display font.
    nonisolated static func italicizingEmphasis(_ styled: AttributedString, font: Font) -> AttributedString {
        var out = styled
        var italic = AttributeContainer()
        italic[AttributeScopes.SwiftUIAttributes.FontAttribute.self] = font.italic()
        // Typed attribute subscripts, as in `PVComboBox`: the dynamic members warn under Swift 6.
        for run in styled.runs where run[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] == .emphasized {
            out[run.range].mergeAttributes(italic)
        }
        return out
    }
}
