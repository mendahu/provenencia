import SwiftUI

/// How a Conclusion row titles itself. A resolved value is plain, a working
/// label is italic, and a ref is mono. Persons, Events, and Places share it.
enum ConclusionTitleSource: Equatable {
    case name(String)
    case label(String)
    case ref(String)

    var text: String {
        switch self {
        case .name(let text), .label(let text), .ref(let text): text
        }
    }
}

/// The S9-D2 row anatomy for a Conclusion list, supplying `PVList`'s slots.
/// Shared by Persons (S9-09), Events (S9-23), and Places (S9-26). The
/// secondary line is per kind: birth and death for Persons, the date and
/// place for Events, the parent chain for Places.
///
/// - **thumbnail:** always reserved; the kind's `subject_*` mark, never a photo.
/// - **title:** auto-reconciled value → *italic* working label → mono ref.
/// - **meta:** the ref, always shown — even when it is also the title — so the
///   column scans.
/// - No *mixed* marker in rows (S9-D2 decision): a mixed value shows its
///   top-ranked value; disagreement is surfaced on the detail page.
enum ConclusionListRow {
    static func thumbnail(mark: PVMarkKey) -> PVThumbnail.Content {
        PVThumbnail.Content(mark: mark, label: nil)
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
                PVBadge(text: "+\(extraCount)", tone: .neutral, subtle: true)
                    .layoutPriority(1)
                    .accessibilityHidden(true)
            }
        }
    }

    @ViewBuilder
    private static func titleText(_ source: ConclusionTitleSource) -> some View {
        switch source {
        case .name(let text):
            Text(verbatim: text)
        case .label(let text):
            Text(verbatim: text).italic()
        case .ref(let text):
            Text(verbatim: text)
                .font(PVFont.mono(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textSecondary)
        }
    }

    static func accessibilityLabel(_ source: ConclusionTitleSource, ref: String) -> String {
        L10n.Workspace.personRowAccessibility(title: source.text, ref: ref)
    }
}
