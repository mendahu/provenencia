import Foundation

/// The pieces a Place title is chosen from. Go returns the kept names;
/// this picks the row text. The chain line is formatted by PlaceChainDisplay.
struct PlaceTitleParts: Equatable, Sendable {
    var names: [String] = []
    var label: String = ""
    var ref: String = ""

    init(names: [String] = [], label: String = "", ref: String = "") {
        self.names = names
        self.label = label
        self.ref = ref
    }

    /// A Place read from its `toponym` field: the displayed values, as the
    /// reconciled-value display words them, without the empty ones.
    init(
        toponym: CatalogConclusionField?,
        entity: CatalogCanonicalEntity,
        locale: Locale,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let names = toponym?.displayedValues
            .map { ReconciledValueDisplay.string(for: $0.value, locale: locale, orders: orders) }
            .filter { !$0.isEmpty } ?? []
        self.init(names: names, label: entity.label, ref: entity.ref)
    }
}

enum PlaceTitleDisplay {
    static func title(_ header: CatalogPlaceHeader) -> String {
        titleSource(header.titleParts).text
    }

    /// First kept name, else the italic label, else the mono ref.
    static func titleSource(_ parts: PlaceTitleParts) -> ConclusionTitleSource {
        let names = parts.names.map(trimmed).filter { !$0.isEmpty }
        if let first = names.first {
            return .name(first)
        }
        let label = trimmed(parts.label)
        if !label.isEmpty {
            return .label(label)
        }
        return .ref(trimmed(parts.ref))
    }

    /// Kept names beyond the one shown, for a later +N.
    static func extraNameCount(_ parts: PlaceTitleParts) -> Int {
        max(0, parts.names.map(trimmed).filter { !$0.isEmpty }.count - 1)
    }

    private static func trimmed(_ value: String) -> String {
        value.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}

extension CatalogPlaceHeader {
    var titleParts: PlaceTitleParts {
        PlaceTitleParts(names: names, label: entity.label, ref: entity.ref)
    }

    var extraNameCount: Int { PlaceTitleDisplay.extraNameCount(titleParts) }
}

extension CatalogHeaderPlace {
    var titleParts: PlaceTitleParts {
        PlaceTitleParts(names: names, label: entity.label, ref: entity.ref)
    }
}
