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

/// A title in one of three sizes: the list row's, a heading's (an event row on
/// the Timeline), or the detail page's display size. Name plain, label italic,
/// ref mono, at each. Draws a `+N` badge when
/// other names survived; the badge keeps its width and the title truncates.
struct PVTitleText: View {
    enum Style: Equatable {
        case row
        /// The `.row` face for a title inside a link: the link owns the ink, so a ref
        /// keeps its mono font and sets no ink of its own.
        case rowLink
        case display
        /// The display face at the h2 size, for a row that heads a block.
        case heading
    }

    /// The sizes a display title takes: its name or label, and its ref.
    private struct DisplayMetrics {
        var titleSize: CGFloat
        var refSize: CGFloat
    }

    private var displayMetrics: DisplayMetrics {
        switch style {
        case .heading: DisplayMetrics(titleSize: PVTypeScale.h2, refSize: PVTypeScale.h3)
        case .row, .rowLink, .display: DisplayMetrics(titleSize: PVTypeScale.display3, refSize: PVTypeScale.h1)
        }
    }

    let source: PVTitleSource
    let style: Style
    let extraCount: Int
    /// The most lines the title takes. Nil is the style's own: one line for `.row`,
    /// and no limit for `.display`.
    let lineLimit: Int?
    let emphasized: Bool
    /// The font the caller set. An italic run of a row title takes it, so a compact card's
    /// caption-size name keeps one size; with none set, the list row's title font.
    @Environment(\.font) private var callerFont

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
        case .row, .rowLink:
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
        case .display, .heading:
            displayText(displayMetrics)
                .modifier(PVTitleLineLimit(lineLimit: lineLimit))
        }
    }

    @ViewBuilder
    private var rowText: some View {
        switch source {
        case .name(let text, let styled, _):
            if let styled {
                Text(Self.italicizingEmphasis(styled, font: callerFont ?? PVListText.titleFont))
            } else {
                Text(verbatim: text)
            }
        case .label(let text):
            Text(verbatim: text).italic()
        case .ref(let text):
            if style == .rowLink {
                Text(verbatim: text).font(PVFont.mono(size: PVTypeScale.bodySmall))
            } else {
                Text(verbatim: text)
                    .font(PVFont.mono(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
            }
        }
    }

    @ViewBuilder
    private func displayText(_ metrics: DisplayMetrics) -> some View {
        let size = metrics.titleSize
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
                .font(PVFont.mono(size: metrics.refSize, weight: PVFontWeight.medium))
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
