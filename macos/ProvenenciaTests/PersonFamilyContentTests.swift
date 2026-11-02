import Foundation
import Testing
@testable import Provenencia

struct PersonFamilyContentTests {
    private let en = Locale(identifier: "en_US")
    private func content(_ family: CatalogPersonFamily, expanded: Set<PersonFamilyContent.Fold> = []) -> PersonFamilyContent {
        PersonFamilyContent(family: family, expanded: expanded, locale: en)
    }
    private func names(_ tiles: [PersonFamilyContent.Tile]) -> [String] { tiles.map(\.content.title.text) }

    @Test func columnsFromFamily() {
        let full = content(PersonFamilyPreviewData.full)
        #expect(full.parentSets.map(names) == [["Heinrich Weber", "Anna Weber"]])
        #expect(full.siblings.map(\.tile.content.title.text) == ["Margarethe Weber", "Johann Weber", "Friedrich Weber", "Wilhelm Weber"])
        #expect(full.siblings.map(\.sets) == [[0], [0], [0], [0]])
        #expect(full.siblings.map(\.tile.content.isFocus) == [false, true, false, false])
        #expect(full.siblings[1].tile.location == nil && full.siblings[1].tile.content.role == nil)
        #expect(full.groups.map { $0.partner?.content.title.text } == ["Maria Schulz", "Sophie Brandt", nil])
        #expect(full.groups.map { names($0.children) } == [["Karl Weber", "Luise Weber", "Emma Weber"], ["Otto Weber"], ["Paul Weber"]])
        #expect(full.siblingsCaption == nil && !full.isEmpty)
        // The children of the parents unit are shown once, as siblings.
        let ids = full.parentSets.flatMap { $0 }.map(\.id) + full.siblings.map(\.tile.id)
            + full.groups.flatMap { ($0.partner.map { [$0] } ?? []) + $0.children }.map(\.id)
        #expect(Set(ids).count == ids.count)

        let conflicting = content(PersonFamilyPreviewData.conflicting)
        #expect(conflicting.parentSets.map(names) == [["Heinrich Weber", "Anna Weber"], ["Heinrich Weber", "Katharina Weber"]])
        #expect(conflicting.siblings.map(\.tile.content.title.text)
            == ["Johann Weber", "Friedrich Weber", "Christian Weber", "Margarethe Weber", "Wilhelm Weber"])
        #expect(conflicting.siblings.map(\.sets) == [[0, 1], [0, 1], [1], [0], [0]])
        let conflictingIDs = conflicting.parentSets.flatMap { $0 }.map(\.id) + conflicting.siblings.map(\.tile.id)
        #expect(Set(conflictingIDs).count == conflictingIDs.count)

        let only = content(PersonFamilyPreviewData.siblingsOnly)
        #expect(only.parentSets.isEmpty && only.siblingsCaption == "Siblings" && only.groups.isEmpty)
        #expect(only.siblings.map(\.tile.content.title.text) == ["Ida Weber", "Clara Weber", "Bernhard Weber"])

        let anna = content(PersonFamilyPreviewData.annaOnly)
        #expect(anna.siblings.count == 1 && anna.siblings[0].tile.content.isFocus && anna.siblingsCaption == nil)
        #expect(anna.groups.count == 1 && anna.groups[0].children.count == 4)
    }

    @Test func aCoParentIsAPartnerAndNoPersonMeansNoFocus() {
        var family = PersonFamilyPreviewData.full
        family.units[1].partners[0].link = nil
        #expect(content(family).groups[0].partner?.content.title.text == "Maria Schulz")
        family.person = nil
        let noFocus = content(family)
        #expect(!noFocus.siblings.contains { $0.tile.content.isFocus })
    }

    @Test func tilesOpenTheRelativeOnFamily() throws {
        let anna = try #require(content(PersonFamilyPreviewData.full).parentSets[0].last)
        let summary = PersonFamilyPreviewData.summary("anna", "Anna Weber")
        #expect(anna.location == PersonsList.location(summary, orders: NameFormatOrders()).withPageSection(.family))
        #expect(anna.location?.pageSection == .family)
        #expect(anna.content.role == "Parent")
    }

    @Test func captionsAndGroupLabels() {
        let full = content(PersonFamilyPreviewData.full)
        #expect(full.parentCaptions == ["Parents"] && full.rightCaption == "Partners and children")
        #expect(full.siblingsLabel == "This person and siblings")
        #expect(full.groups.map(\.label) == ["Partner Maria Schulz and children", "Partner Sophie Brandt and children", "Children"])
        #expect(content(PersonFamilyPreviewData.conflicting).parentCaptions == ["Parents", "Other parents"])
        #expect(content(PersonFamilyPreviewData.annaOnly).rightCaption == "Partner and children")
        #expect(content(PersonFamilyPreviewData.annaOnly).siblingsLabel == "This person")

        let one = CatalogPersonFamily(
            units: [PersonFamilyPreviewData.unit(.parents, partners: [PersonFamilyPreviewData.member("dorothea", "Dorothea Weber")],
                                        children: [PersonFamilyPreviewData.member("heinrich", "Heinrich Weber")])],
            person: PersonFamilyPreviewData.johann,
            originSiblings: [PersonFamilyPreviewData.origin(PersonFamilyPreviewData.member("heinrich", "Heinrich Weber"), [0]),
                             PersonFamilyPreviewData.origin(PersonFamilyPreviewData.johannMember, [0])]
        )
        #expect(content(one).parentCaptions == ["Parent"])

        var partnerAlone = PersonFamilyPreviewData.annaOnly
        partnerAlone.units[0].children = []
        let alone = content(partnerAlone)
        #expect(alone.rightCaption == "Partner" && alone.groups[0].label == "Partner Heinrich Weber")

        let childrenOnly = CatalogPersonFamily(
            units: [PersonFamilyPreviewData.unit(.oneParent, children: [PersonFamilyPreviewData.member("kid", "Kid Weber")])],
            person: PersonFamilyPreviewData.johann,
            originSiblings: [PersonFamilyPreviewData.origin(PersonFamilyPreviewData.johannMember, [])]
        )
        #expect(content(childrenOnly).rightCaption == "Children")

        var twoPartners = PersonFamilyPreviewData.full
        twoPartners.units = [twoPartners.units[1], twoPartners.units[2]]
        twoPartners.units[0].children = []
        twoPartners.units[1].children = []
        #expect(content(twoPartners).rightCaption == "Partners")
    }

    @Test func fiftyMembersKeepOrder() {
        let kids = (0..<50).map { PersonFamilyPreviewData.member("kid\($0)", "Kid \($0)") }
        let family = CatalogPersonFamily(
            units: [PersonFamilyPreviewData.unit(.partner, partners: [PersonFamilyPreviewData.member("sam", "Sam Weber")], children: kids)],
            person: PersonFamilyPreviewData.johann,
            originSiblings: [PersonFamilyPreviewData.origin(PersonFamilyPreviewData.johannMember, [])]
        )
        let tiles = content(family, expanded: [.children(0)]).groups[0].children
        #expect(tiles.count == 50 && names(tiles) == (0..<50).map { "Kid \($0)" })
        #expect(Set(tiles.map(\.id)).count == 50)
    }

    private typealias Data = PersonFamilyPreviewData

    @Test func foldsEachListAtThirty() {
        #expect(PersonFamilyContent.foldLimit == 30)
        let kids = (0..<50).map { Data.member("kid\($0)", "Kid \($0)") }
        let children = content(CatalogPersonFamily(
            units: [Data.unit(.partner, partners: [Data.member("sam", "Sam Weber")], children: kids)],
            person: Data.johann, originSiblings: [Data.origin(Data.johannMember, [])]
        ))
        #expect(children.groups[0].children.count == 30 && children.groups[0].childrenHidden == 20)
        #expect(names(children.groups[0].children) == (0..<30).map { "Kid \($0)" })

        let parents = (0..<40).map { Data.unit(.parents, partners: [Data.member("p\($0)", "Parent \($0)")]) }
        let sets = content(CatalogPersonFamily(
            units: parents, person: Data.johann,
            originSiblings: [Data.origin(Data.johannMember, Array(0..<40))]
        ))
        #expect(sets.parentSets.count == 30 && sets.parentCaptions.count == 30 && sets.parentSetsHidden == 10)
        #expect(sets.siblings[0].sets == Array(0..<30))

        let siblings = (0..<40).map { Data.member("sib\($0)", "Sib \($0)") }
        let sibs = content(CatalogPersonFamily(
            units: [Data.unit(.siblings, children: siblings)], person: Data.johann,
            originSiblings: [Data.origin(Data.johannMember, [])] + siblings.dropLast().map { Data.origin($0, []) }
        ))
        #expect(sibs.siblings.count == 30 && sibs.siblingsHidden == 10)

        let partners = (0..<40).map { Data.unit(.partner, partners: [Data.member("mate\($0)", "Mate \($0)")]) }
        let groups = content(CatalogPersonFamily(
            units: partners, person: Data.johann, originSiblings: [Data.origin(Data.johannMember, [])]
        ))
        #expect(groups.groups.count == 30 && groups.groupsHidden == 10)

        #expect(content(Data.full).parentSetsHidden == 0 && content(Data.full).siblingsHidden == 0)
        #expect(content(Data.full).groupsHidden == 0 && content(Data.full).groups.allSatisfy { $0.childrenHidden == 0 })
    }

    @Test func expandedShowsAll() {
        let kids = (0..<50).map { Data.member("kid\($0)", "Kid \($0)") }
        let parents = (0..<40).map { Data.unit(.parents, partners: [Data.member("p\($0)", "Parent \($0)")]) }
        let mates = (0..<40).map { Data.unit(.partner, partners: [Data.member("mate\($0)", "Mate \($0)")], children: kids) }
        let siblings = (0..<40).map { Data.member("sib\($0)", "Sib \($0)") }
        let family = CatalogPersonFamily(
            units: parents + [Data.unit(.siblings, children: siblings)] + mates,
            person: Data.johann,
            originSiblings: [Data.origin(Data.johannMember, Array(0..<40))] + siblings.map { Data.origin($0, []) }
        )
        let folded = content(family)
        #expect(folded.parentSetsHidden == 10 && folded.siblingsHidden == 11 && folded.groupsHidden == 10)
        #expect(folded.groups[0].childrenHidden == 20)

        var everything: Set<PersonFamilyContent.Fold> = [.parentSets, .siblings, .partners]
        for index in 0..<40 { everything.insert(.children(index)) }
        let open = content(family, expanded: everything)
        #expect(open.parentSets.count == 40 && open.siblings.count == 41 && open.groups.count == 40)
        #expect(open.parentSetsHidden == 0 && open.siblingsHidden == 0 && open.groupsHidden == 0)
        #expect(open.groups.allSatisfy { $0.children.count == 50 && $0.childrenHidden == 0 })
        #expect(open.allTiles.count == 40 + 41 + 40 + 40 * 50)
    }

    @Test func foldKeepsTheFocus() {
        let siblings = (0..<49).map { Data.member("sib\($0)", "Sib \($0)") }
        let family = CatalogPersonFamily(
            units: [Data.unit(.siblings, children: siblings)],
            person: Data.johann,
            originSiblings: siblings[..<45].map { Data.origin($0, []) } + [Data.origin(Data.johannMember, [])]
                + siblings[45...].map { Data.origin($0, []) }
        )
        let folded = content(family)
        #expect(folded.siblings.count == 31 && folded.siblingsHidden == 19)
        #expect(names(folded.siblings.map(\.tile)) == (0..<30).map { "Sib \($0)" } + ["Johann Weber"])
        #expect(folded.siblings.last?.tile.content.isFocus == true)
    }

    @Test func foldDropsHiddenSetIndexes() {
        let parents = (0..<31).map { Data.unit(.parents, partners: [Data.member("p\($0)", "Parent \($0)")]) }
        let late = Data.member("late", "Late Weber")
        let family = CatalogPersonFamily(
            units: parents, person: Data.johann,
            originSiblings: [Data.origin(Data.johannMember, Array(0..<31)), Data.origin(late, [30])]
        )
        let folded = content(family)
        #expect(folded.parentSets.count == 30 && folded.parentSetsHidden == 1)
        #expect(folded.siblings.map(\.sets) == [Array(0..<30), []])
    }

    @Test func emptyFamilyIsEmpty() {
        let empty = CatalogPersonFamily(
            units: [], person: PersonFamilyPreviewData.johann,
            originSiblings: [PersonFamilyPreviewData.origin(PersonFamilyPreviewData.johannMember, [])]
        )
        #expect(content(empty).isEmpty)
        #expect(content(CatalogPersonFamily(units: [])).isEmpty)
        #expect(!content(PersonFamilyPreviewData.full).isEmpty)
    }
}
