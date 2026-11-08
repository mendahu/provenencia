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

    // MARK: Johann Weber: seven records on five Sources (S13-D6 § 5) and how they connect (S13-D10 § 5)

    private static func record(_ claim: String, _ graph: String, _ ref: String, _ name: String, _ label: String, _ subjectRef: String) -> CatalogPersonClaim {
        CatalogPersonClaim(
            claimID: claim, graphID: graph, graphRef: ref, graphName: name,
            subject: subject("s-\(claim)", ref: subjectRef, label: label), observations: []
        )
    }

    static let johannClaims = [
        CatalogPersonClaimGroup(
            sourceID: "src-census", sourceRef: "SRC-9TQ2M", title: "Hamburg census 1867", sourceTypeID: "type-census",
            claims: [
                record("c-214", "g-214", "EVG-2K7PQ", "Household 214", "Johann Weber", "CPR-7Q2KD"),
                record("c-388", "g-388", "EVG-5MD8R", "Household 388", "J. Weber", "CPR-2M8XF"),
            ]
        ),
        CatalogPersonClaimGroup(
            sourceID: "src-adress", sourceRef: "SRC-3HW6B", title: "Hamburger Adressbuch 1861", sourceTypeID: "type-census",
            claims: [record("c-vol3", "g-vol3", "EVG-8TZ4N", "Volume 3", "Weber, Joh., Tischler", "CPR-9PA4D")]
        ),
        CatalogPersonClaimGroup(
            sourceID: "src-news", sourceRef: "SRC-7KP2Q", title: "Hamburger Nachrichten, obituary 1880", sourceTypeID: "type-census",
            claims: [record("c-221", "g-221", "EVG-1QR9D", "Issue 221", "Johann Heinrich Weber", "CPR-4TZ6K")]
        ),
        CatalogPersonClaimGroup(
            sourceID: "src-konf", sourceRef: "SRC-2DN5X", title: "Konfirmandenbuch St. Ansgarii 1831", sourceTypeID: "type-census",
            claims: [
                record("c-reg2", "g-reg2", "EVG-6WB3M", "Register 2", "Johann Weber", "CPR-6HQ3R"),
                record("c-reg3", "g-reg3", "EVG-4PL7T", "Register 3", "Joh. Weber", "CPR-1XG8K"),
            ]
        ),
        CatalogPersonClaimGroup(
            sourceID: "src-verein", sourceRef: "SRC-5RJ4W",
            title: "Verzeichniß der Mitglieder des Vereins für Hamburgische Geschichte, nebst einem Anhang über die Statuten",
            sourceTypeID: "type-census",
            claims: [record("c-p7", "g-p7", "EVG-9CN2H", "Page 7", "Weber, J. H.", "CPR-8WK2N")]
        ),
    ]

    private static func links(
        graphs claims: [CatalogPersonClaimGroup], pairs: [CatalogRecordLinksPair]
    ) -> CatalogPersonRecordLinks {
        CatalogPersonRecordLinks(
            graphs: claims.flatMap { group in
                group.claims.map {
                    CatalogRecordLinksGraph(
                        graphID: $0.graphID, graphRef: $0.graphRef, graphName: $0.graphName,
                        sourceID: group.sourceID, sourceTitle: group.title, sourceTypeID: group.sourceTypeID
                    )
                }
            },
            pairs: pairs
        )
    }

    private static func pair(
        _ a: String, _ b: String, _ lines: [(String, CatalogRecordLinkStrength)]
    ) -> CatalogRecordLinksPair {
        CatalogRecordLinksPair(
            graphAID: a, graphBID: b, lines: lines.map { CatalogRecordLink(entityID: $0.0, strength: $0.1) }
        )
    }

    /// Eight lines on five pairs; the Volume 3 and Page 7 graphs share no one.
    static let johannRecordLinks = links(graphs: johannClaims, pairs: [
        pair("g-214", "g-388", [("ent-anna", .strong), ("ent-karl", .weak)]),
        pair("g-214", "g-221", [("ent-anna", .strong), ("ent-karl", .strong)]),
        pair("g-388", "g-221", [("ent-anna", .weak)]),
        pair("g-214", "g-reg2", [("ent-heinrich", .strong)]),
        pair("g-reg2", "g-reg3", [("ent-heinrich", .weak), ("ent-maria", .weak)]),
    ])

    /// `graphs` records on fourteen Sources (the first holds nine), at most the fifty the map's
    /// Scale names; graph `g-m<i>`.
    static func manyClaims(graphs: Int) -> [CatalogPersonClaimGroup] {
        let sizes = [9, 6, 5, 4, 4, 3, 3, 3, 3, 3, 2, 2, 2, 1]
        var next = 1
        var groups: [CatalogPersonClaimGroup] = []
        for (index, size) in sizes.enumerated() {
            let ids = Array(next..<min(next + size, graphs + 1))
            next += size
            guard !ids.isEmpty else { break }
            groups.append(CatalogPersonClaimGroup(
                sourceID: "src-n\(index + 1)", sourceRef: "SRC-N\(index + 1)",
                title: index == 0 ? "Hamburger Adressbuch 1855 – 1880" : "Jahrbuch \(1850 + index)",
                sourceTypeID: "type-census",
                claims: ids.map {
                    record("cm\($0)", "g-m\($0)", "EVG-M\($0)", "Volume \($0)", "Johann Weber", "CPR-M\($0)")
                }
            ))
        }
        return groups
    }

    /// Graph i shares a person with i + 1 and i + 2, strongly when i is even; the last three
    /// graphs stand apart.
    static func manyRecordLinks(graphs: Int) -> CatalogPersonRecordLinks {
        let apart = Set((graphs - 2)...graphs)
        var pairs: [CatalogRecordLinksPair] = []
        for i in 1...max(1, graphs) {
            for j in [i + 1, i + 2] where j <= graphs && !apart.contains(i) && !apart.contains(j) {
                pairs.append(pair("g-m\(i)", "g-m\(j)", [("ent-\(i)-\(j)", i % 2 == 0 ? .strong : .weak)]))
            }
        }
        return links(graphs: manyClaims(graphs: graphs), pairs: pairs)
    }

    /// A session over a `FakeStore` that serves the record links of `entityId`: the given
    /// links, a failure, or (with neither) an answer an hour late, so the column shows its
    /// spinner.
    @MainActor
    static func previewSession(
        entityId: String, recordLinks: CatalogPersonRecordLinks?, failure: Bool
    ) -> WorkspaceSession {
        let store = FakeStore()
        if failure {
            store.getPersonRecordLinksError = CoreInvokeError.coded(
                status: 1, code: "catalog.already_open", kind: .user, params: []
            )
        } else if let recordLinks {
            store.seededPersonRecordLinks[entityId] = recordLinks
        } else {
            store.personRecordLinksDelayNanoseconds = 3_600_000_000_000
        }
        return WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/preview.provenencia"), store: store)
    }
}
#endif
