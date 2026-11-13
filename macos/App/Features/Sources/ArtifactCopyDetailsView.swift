import SwiftUI

/// An Artifact row's saved copy details, read-only: the first four rows with a
/// value as field name and value, and "+n more" past them, which opens the
/// Artifact sheet at its copy details. An Artifact with none draws nothing inside its column.
struct ArtifactCopyDetailsView: View {
    static let shownCount = 4
    /// The field-name column.
    static let labelWidth: CGFloat = 96

    let rows: [CatalogMetadataEntry]
    let ref: String
    let artifactID: String
    let onMore: () -> Void

    /// The rows with a value, in saved order: the first four and how many follow.
    static func shown(_ rows: [CatalogMetadataEntry]) -> (shown: [CatalogMetadataEntry], more: Int) {
        let valued = rows.filter(\.hasValue)
        return (Array(valued.prefix(shownCount)), max(0, valued.count - shownCount))
    }

    var body: some View {
        let (shown, more) = Self.shown(rows)
        if !shown.isEmpty {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Grid(alignment: .topLeading, horizontalSpacing: PVSpacing.space5, verticalSpacing: PVSpacing.space1) {
                    ForEach(shown) { entry in
                        GridRow {
                            Text(verbatim: entry.field.label)
                                .font(PVFont.body(size: PVTypeScale.caption))
                                .foregroundStyle(PVColor.textMuted)
                                .lineLimit(1)
                                .truncationMode(.tail)
                                .help(entry.field.label)
                                .frame(width: Self.labelWidth, alignment: .leading)
                                .gridColumnAlignment(.leading)
                            MetadataValueText(
                                entry: entry,
                                accessibilityIdentifier: "sources.page.artifact.\(artifactID).copyDetails.\(entry.field.id)",
                                lineLimit: 1,
                                chrome: false
                            )
                            .frame(maxWidth: .infinity, alignment: .leading)
                        }
                        .accessibilityElement(children: .combine)
                        .accessibilityLabel(Text(verbatim: "\(entry.field.label), \(entry.valueText)"))
                    }
                }
                if more > 0 {
                    PVButton(L10n.Sources.artifactCopyDetailsMore(count: more), variant: .link, size: .sm) {
                        onMore()
                    }
                    .accessibilityLabel(L10n.Sources.artifactCopyDetailsMoreLabel(count: more, ref: ref))
                    .accessibilityIdentifier("sources.page.artifact.\(artifactID).copyDetails.more")
                }
            }
            .accessibilityElement(children: .contain)
            .accessibilityIdentifier("sources.page.artifact.\(artifactID).copyDetails")
        }
    }
}
