import SwiftUI

/// How a Conclusion row titles itself. A resolved value is plain, a working
/// label is italic, and a ref is mono. Persons, Events, and Places share it. A name may carry
/// a styled copy whose emphasized runs draw italic.
enum ConclusionTitleSource: Equatable {
    case name(String, styled: AttributedString? = nil)
    case label(String)
    case ref(String)

    var text: String {
        switch self {
        case .name(let text, _): text
        case .label(let text), .ref(let text): text
        }
    }
}

/// The S9-D2 row anatomy for a Conclusion list, supplying `PVList`'s slots.
/// Shared by Persons (S9-09), Events (S9-23), and Places (S9-26). The
/// secondary line is per kind: birth and death for Persons, the date and
/// place for Events, the parent chain for Places.
///
/// - **thumbnail:** always reserved; the kind's `subject_*` mark, or the picture a kind leads with.
/// - **title:** auto-reconciled value → *italic* working label → mono ref.
/// - **meta:** the ref, always shown — even when it is also the title — so the
///   column scans.
/// - No *mixed* marker in rows (S9-D2 decision): a mixed value shows its
///   top-ranked value; disagreement is surfaced on the detail page.
enum ConclusionListRow {
    static func thumbnail(mark: PVMarkKey) -> PVThumbnail.Content {
        PVThumbnail.Content(glyph: .mark(mark), label: nil)
    }

    /// The title, with a `+N` badge when other names survived. The badge keeps
    /// its width; the title truncates. Main actor: `PVBadge` builds its label there.
    @MainActor
    @ViewBuilder
    static func title(_ source: ConclusionTitleSource, extraCount: Int = 0) -> some View {
        HStack(spacing: PVSpacing.space4) {
            titleText(source)
                .lineLimit(1)
                .truncationMode(.tail)
            if extraCount > 0 {
                PVBadge(text: L10n.Conclusions.moreCount(extraCount), tone: .neutral, subtle: true)
                    .layoutPriority(1)
                    .accessibilityHidden(true)
            }
        }
    }

    @ViewBuilder
    private static func titleText(_ source: ConclusionTitleSource) -> some View {
        switch source {
        case .name(let text, let styled):
            if let styled {
                Text(italicizingEmphasis(styled, font: PVFont.body(size: PVTypeScale.body)))
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

    /// `styled` with `font` italicized on each emphasized run. The italic is set explicitly,
    /// so it does not depend on how SwiftUI draws `.emphasized` in the custom display font.
    static func italicizingEmphasis(_ styled: AttributedString, font: Font) -> AttributedString {
        var out = styled
        var italic = AttributeContainer()
        italic[AttributeScopes.SwiftUIAttributes.FontAttribute.self] = font.italic()
        // Typed attribute subscripts, as in `PVComboBox`: the dynamic members warn under Swift 6.
        for run in styled.runs where run[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] == .emphasized {
            out[run.range].mergeAttributes(italic)
        }
        return out
    }

    /// The last-updated column: the history icon and the cell text, faint when
    /// no time is recorded. `now` is read as each row builds (no timer).
    static func updatedColumn<Kind: ConclusionListKind>(_: Kind.Type) -> PVListDetailColumn<Kind.Row> {
        PVListDetailColumn(width: PVLastUpdated.columnWidth) { row in
            PVLastUpdated.detail(Kind.updatedAt(row), in: .current)
        }
    }

    static func accessibilityLabel(_ source: ConclusionTitleSource, ref: String) -> String {
        L10n.Workspace.personRowAccessibility(title: source.text, ref: ref)
    }
}
