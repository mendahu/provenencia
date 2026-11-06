#if DEBUG
import Foundation

/// Sources for the Sources section's `#Preview`s: the claims a Person's page reads, grouped by
/// Source in title order, each Source's records in graph name order.
enum PersonSourcesPreviewData {
    static let types = [
        CatalogSourceType(
            id: "type-census", key: "census", origin: "provenencia", label: "Census", description: "", iconKey: "type_census"
        ),
    ]

    private static func subject(_ id: String, ref: String, label: String) -> CatalogSubject {
        CatalogSubject(id: id, ref: ref, sourceID: "src", subjectTypeID: "type-person", label: label, description: "")
    }

    private static func claim(_ id: String, graph: String, ref: String, name: String, subject: CatalogSubject) -> CatalogPersonClaim {
        CatalogPersonClaim(
            claimID: id, graphID: graph, graphRef: ref, graphName: name, subject: subject, observations: []
        )
    }

    private static let johann = subject("s1", ref: "CPR-7Q2KD", label: "Johann Weber")

    /// A typed Source with two graphs, and one whose type is not in the list.
    static let regular = [
        CatalogPersonClaimGroup(
            sourceID: "src-1", sourceRef: "SRC-9TQ2M", title: "Hamburg census 1867", sourceTypeID: "type-census",
            claims: [
                claim("c1", graph: "g-214", ref: "EVG-1", name: "Household 214", subject: johann),
                claim("c2", graph: "g-388", ref: "EVG-5MD8R", name: "Household 388", subject: johann),
            ]
        ),
        CatalogPersonClaimGroup(
            sourceID: "src-2", sourceRef: "SRC-2K8XF", title: "Hamburger Adressbuch 1861", sourceTypeID: "gone",
            claims: [claim("c3", graph: "g-a", ref: "EVG-3", name: "Page 4", subject: johann)]
        ),
    ]

    /// A Source with no title (its ref stands in) and a graph with no name.
    static let untitled = [
        CatalogPersonClaimGroup(
            sourceID: "src-3", sourceRef: "SRC-4KQ7M", title: "", sourceTypeID: "type-census",
            claims: [claim("c4", graph: "g-l", ref: "EVG-7", name: "", subject: johann)]
        ),
    ]

    /// Seven records: five shown and "Show 2 more".
    static let folded = [
        CatalogPersonClaimGroup(
            sourceID: "src-4", sourceRef: "SRC-6P3VD", title: "Kirchenbuch St. Michaelis", sourceTypeID: "type-census",
            claims: (1...7).map {
                claim("f\($0)", graph: "g-f\($0)", ref: "EVG-\($0)", name: "Entry \($0)", subject: johann)
            }
        ),
    ]

    /// Sixty Sources: fifty shown and "Show 10 more".
    static let manySources = (1...60).map {
        CatalogPersonClaimGroup(
            sourceID: "src-m\($0)", sourceRef: "SRC-M\($0)", title: "Adressbuch \(1800 + $0)", sourceTypeID: "type-census",
            claims: [claim("m\($0)", graph: "g-m\($0)", ref: "EVG-M\($0)", name: "Page \($0)", subject: johann)]
        )
    }
}
#endif
