import SwiftUI

/// The `origin` namespace shared by every catalog vocabulary row
/// (`source_types`, `source_metadata_fields`) — mirrors the `Origin…`
/// constants in `core/database/sourcetypes` and `…/metadatafields`. The FFI
/// layer carries origin as a plain string, not an enum, and the vocabulary
/// is open: anything that is neither seeded nor researcher-authored belongs
/// to a plugin.
enum CatalogOrigin {
    static let provenencia = "provenencia"
    static let user = "user"

    private static let pluginPrefix = "plugin:"

    /// Anything that is neither seeded nor researcher-authored is owned by a
    /// plugin — the same open-vocabulary stance the badges take.
    static func isPlugin(_ origin: String) -> Bool {
        origin != provenencia && origin != user
    }

    /// The id after `plugin:` (e.g. `"plugin:findagrave"` → `"findagrave"`),
    /// or the raw origin unchanged when it has no such prefix.
    static func pluginID(from origin: String) -> String {
        origin.hasPrefix(pluginPrefix) ? String(origin.dropFirst(pluginPrefix.count)) : origin
    }
}

/// The full origin badge shown on a detail panel, for metadata fields and
/// source types alike. The client styles only the two values it knows and
/// passes anything else (a `plugin:…` id) through as raw text.
struct OriginBadge: View {
    let origin: String

    var body: some View {
        switch origin {
        case CatalogOrigin.provenencia:
            PVBadge(L10n.Origin.provenencia, tone: .accent)
        case CatalogOrigin.user:
            PVBadge(L10n.Origin.user, tone: .warning)
        default:
            PVBadge(text: origin, tone: .info)
        }
    }
}

/// The glyph pill that marks a row's origin inline after its label in a
/// vocabulary list. `user` rows show nothing — the researcher's own rows are
/// the norm — so the column reads as an exception marker rather than a third
/// origin column. The full vocabulary still renders as `OriginBadge` on the
/// detail panel.
struct OriginPill: View {
    let origin: String

    var body: some View {
        if origin == CatalogOrigin.provenencia {
            PVBadge(icon: .shieldCheck, label: L10n.Origin.seededPill, tone: .accent, subtle: true)
        } else if CatalogOrigin.isPlugin(origin) {
            PVBadge(
                icon: .plug,
                label: L10n.Origin.pluginPill(pluginID: CatalogOrigin.pluginID(from: origin)),
                tone: .info,
                subtle: true
            )
        }
    }
}

#Preview {
    VStack(alignment: .leading, spacing: PVSpacing.space5) {
        OriginBadge(origin: CatalogOrigin.provenencia)
        OriginBadge(origin: CatalogOrigin.user)
        OriginBadge(origin: "plugin:findagrave")
        OriginPill(origin: CatalogOrigin.provenencia)
        OriginPill(origin: CatalogOrigin.user)
        OriginPill(origin: "plugin:findagrave")
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
