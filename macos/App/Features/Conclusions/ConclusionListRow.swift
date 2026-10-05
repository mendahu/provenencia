import SwiftUI

/// The S9-D2 row anatomy for a Conclusion list, supplying `PVList`'s slots.
/// Shared by Persons (S9-09), Events (S9-23), and Places (S9-26); only the
/// secondary line's content differs per kind, and it arrives with S9-32.
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

    @ViewBuilder
    static func title(_ source: PersonHeaderDisplay.TitleSource) -> some View {
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

    static func accessibilityLabel(_ source: PersonHeaderDisplay.TitleSource, ref: String) -> String {
        L10n.Workspace.personRowAccessibility(title: source.text, ref: ref)
    }
}
