import Foundation
@testable import Provenencia

/// Families for the Family section's tests: the S13-D3 sample (Johann Weber)
/// as Go returns it.
enum FamilyFixtures {
    static func summary(_ id: String, _ name: String?, born: Int32? = nil) -> CatalogPersonSummary {
        var h = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: id, ref: "PER-\(id.uppercased())", subjectTypeID: "t", label: ""),
            name: name.map { CatalogNameValue(form: $0) },
            nameValueCount: name == nil ? 0 : 1
        )
        if let born { h.birth = CatalogLifeFacts(date: CatalogDateValueInput(kind: "point", startYear: born), dateCount: 1) }
        return h
    }

    static func member(_ id: String, _ name: String?, born: Int32? = nil, link: Bool = true) -> CatalogFamilyMember {
        CatalogFamilyMember(
            person: summary(id, name, born: born),
            link: link ? CatalogFamilyLink(recorded: true, associationIDs: [], implied: []) : nil
        )
    }

    static func unit(
        _ kind: CatalogFamilyUnit.Kind,
        partners: [CatalogFamilyMember] = [],
        children: [CatalogFamilyMember] = []
    ) -> CatalogFamilyUnit {
        CatalogFamilyUnit(kind: kind, partners: partners, children: children)
    }

    static func origin(_ member: CatalogFamilyMember, _ sets: [Int]) -> CatalogOriginSibling {
        CatalogOriginSibling(member: member, parentSets: sets)
    }

    static let johann = summary("johann", "Johann Weber", born: 1817)

    static var johannMember: CatalogFamilyMember { CatalogFamilyMember(person: johann, link: nil) }

    /// Frame 11a.
    static var full: CatalogPersonFamily {
        let margarethe = member("margarethe", "Margarethe Weber", born: 1815)
        let friedrich = member("friedrich", "Friedrich Weber", born: 1820)
        let wilhelm = member("wilhelm", "Wilhelm Weber")
        return CatalogPersonFamily(
            units: [
                unit(.parents, partners: [member("heinrich", "Heinrich Weber"), member("anna", "Anna Weber")],
                     children: [margarethe, friedrich, wilhelm]),
                unit(.partner, partners: [member("maria", "Maria Schulz")],
                     children: [member("karl", "Karl Weber"), member("luise", "Luise Weber"), member("emma", "Emma Weber")]),
                unit(.partner, partners: [member("sophie", "Sophie Brandt")], children: [member("otto", "Otto Weber")]),
                unit(.oneParent, children: [member("paul", "Paul Weber")]),
            ],
            person: johann,
            originSiblings: [
                origin(margarethe, [0]), origin(johannMember, [0]), origin(friedrich, [0]), origin(wilhelm, [0]),
            ]
        )
    }

    /// Frame 11c: a second parent set, Christian in it only.
    static var conflicting: CatalogPersonFamily {
        let margarethe = member("margarethe", "Margarethe Weber", born: 1815)
        let friedrich = member("friedrich", "Friedrich Weber", born: 1820)
        let christian = member("christian", "Christian Weber", born: 1823)
        let wilhelm = member("wilhelm", "Wilhelm Weber")
        return CatalogPersonFamily(
            units: [
                unit(.parents, partners: [member("heinrich", "Heinrich Weber"), member("anna", "Anna Weber")],
                     children: [margarethe, friedrich, wilhelm]),
                unit(.parents, partners: [member("heinrich", "Heinrich Weber"), member("katharina", "Katharina Weber")],
                     children: [friedrich, christian]),
            ],
            person: johann,
            originSiblings: [
                origin(johannMember, [0, 1]), origin(friedrich, [0, 1]), origin(christian, [1]),
                origin(margarethe, [0]), origin(wilhelm, [0]),
            ]
        )
    }

    /// Frame 11d: Clara and her recorded siblings only.
    static var siblingsOnly: CatalogPersonFamily {
        let clara = summary("clara", "Clara Weber", born: 1841)
        let ida = member("ida", "Ida Weber", born: 1838)
        let bernhard = member("bernhard", "Bernhard Weber", born: 1844)
        return CatalogPersonFamily(
            units: [unit(.siblings, children: [ida, bernhard])],
            person: clara,
            originSiblings: [
                origin(ida, []), origin(CatalogFamilyMember(person: clara, link: nil), []), origin(bernhard, []),
            ]
        )
    }

    /// Frame 11h: Anna, with Heinrich as partner.
    static var annaOnly: CatalogPersonFamily {
        let anna = summary("anna", "Anna Weber", born: 1790)
        return CatalogPersonFamily(
            units: [unit(.partner, partners: [member("heinrich", "Heinrich Weber")], children: [
                member("margarethe", "Margarethe Weber", born: 1815), member("johann", "Johann Weber", born: 1817),
                member("friedrich", "Friedrich Weber", born: 1820), member("wilhelm", "Wilhelm Weber"),
            ])],
            person: anna,
            originSiblings: [origin(CatalogFamilyMember(person: anna, link: nil), [])]
        )
    }
}
