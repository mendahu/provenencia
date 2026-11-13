import SwiftUI

/// What a row's right-click menu offers; S15-25 adds a `.move` case between edit and delete.
enum ArtifactRowAction: CaseIterable {
    case edit, setCover, delete

    /// The actions a row offers: Set cover only when the Artifact can be the cover.
    static func menuActions(canSetCover: Bool) -> [ArtifactRowAction] {
        allCases.filter { $0 != .setCover || canSetCover }
    }

    var title: LocalizedStringResource {
        switch self {
        case .edit: L10n.Sources.artifactMenuEdit
        case .setCover: L10n.Sources.useAsThumbnail
        case .delete: L10n.Sources.artifactMenuDelete
        }
    }

    var idSuffix: String {
        switch self {
        case .edit: "edit"
        case .setCover: "setCover"
        case .delete: "delete"
        }
    }
}

/// One Artifact on the Source page: a list row (S15-D14). A Source page snowflake.
struct SourcePageArtifactRow: View {
    static let thumbnailSize: CGFloat = 64

    /// The copy-details column: 320pt at Regular, 440pt at Wide.
    static func copyDetailsWidth(_ widthClass: PVWidthClass) -> CGFloat {
        switch widthClass {
        case .regular: 320
        case .wide: 440
        }
    }

    var body: some View {
        EmptyView()
    }
}
