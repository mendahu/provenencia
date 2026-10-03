import SwiftUI

/// `CatalogMetadataField.dataType` namespace — mirrors
/// `core/database/metadatafields`'s `DataType…` constants (the FFI layer
/// carries these as plain strings, not an enum). Shared Catalog shelf for
/// Metadata, Source types, and the Source page — not owned by any one
/// destination.
enum CatalogFieldDataType {
    static let text = "text"
    static let date = "date"
    static let url = "url"

    /// Display label for a data-type value — shared by
    /// `CatalogFieldDataTypeBadge` and locked detail plain-text rendering.
    static func label(for dataType: String) -> LocalizedStringResource {
        switch dataType {
        case date:
            return L10n.Metadata.dataTypeDate
        case url:
            return L10n.Metadata.dataTypeUrl
        default:
            return L10n.Metadata.dataTypeText
        }
    }

    /// Icon for type microcaps / badges. Unknown values fall back to text.
    static func icon(for dataType: String) -> PVSymbol {
        switch dataType {
        case date:
            return .calendar
        case url:
            return .externalLink
        default:
            return .textType
        }
    }
}

/// The data-type → badge mapping in one place. Same open-vocabulary stance
/// as `OriginBadge`: anything the client doesn't know renders as text.
struct CatalogFieldDataTypeBadge: View {
    let dataType: String

    var body: some View {
        switch dataType {
        case CatalogFieldDataType.date:
            PVBadge(L10n.Metadata.dataTypeDate, tone: .info, icon: .calendar, subtle: true)
        case CatalogFieldDataType.url:
            PVBadge(L10n.Metadata.dataTypeUrl, tone: .info, icon: .externalLink, subtle: true)
        default:
            PVBadge(L10n.Metadata.dataTypeText, tone: .neutral, icon: .textType, subtle: true)
        }
    }
}

#Preview {
    VStack(alignment: .leading, spacing: PVSpacing.space5) {
        CatalogFieldDataTypeBadge(dataType: CatalogFieldDataType.text)
        CatalogFieldDataTypeBadge(dataType: CatalogFieldDataType.date)
        CatalogFieldDataTypeBadge(dataType: CatalogFieldDataType.url)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
