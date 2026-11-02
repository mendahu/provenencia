import SwiftUI

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

    /// The last-updated column: the history icon and the cell text, faint when
    /// no time is recorded. `now` is read as each row builds (no timer).
    static func updatedColumn<Kind: ConclusionListKind>(_: Kind.Type) -> PVListDetailColumn<Kind.Row> {
        PVListDetailColumn(width: PVLastUpdated.columnWidth) { row in
            PVLastUpdated.detail(Kind.updatedAt(row), in: .current)
        }
    }

    static func accessibilityLabel(_ source: PVTitleSource, ref: String) -> String {
        L10n.Workspace.personRowAccessibility(title: source.spokenText, ref: ref)
    }
}
