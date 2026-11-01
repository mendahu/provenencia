import Foundation
import Testing
@testable import Provenencia

struct PersonFamilyContentTests {
    private let en = Locale(identifier: "en_US")
    private func content(_ family: CatalogPersonFamily) -> PersonFamilyContent {
        PersonFamilyContent(family: family, locale: en)
    }
    private func names(_ tiles: [PersonFamilyContent.Tile]) -> [String] { tiles.map(\.content.title.text) }

    @Test func columnsFromFamily() {
        let full = content(FamilyFixtures.full)
        #expect(full.parentSets.map(names) == [["Heinrich Weber", "Anna Weber"]])
        #expect(full.siblings.map(\.tile.content.title.text) == ["Margarethe Weber", "Johann Weber", "Friedrich Weber", "Wilhelm Weber"])
        #expect(full.siblings.map(\.sets) == [[0], [0], [0], [0]])
        #expect(full.siblings.map(\.tile.content.isFocus) == [false, true, false, false])
        #expect(full.siblings[1].tile.location == nil && full.siblings[1].tile.content.role == nil)
        #expect(full.groups.map { $0.partner?.content.title.text } == ["Maria Schulz", "Sophie Brandt", nil])
        #expect(full.groups.map { names($0.children) } == [["Karl Weber", "Luise Weber", "Emma Weber"], ["Otto Weber"], ["Paul Weber"]])
        #expect(!full.siblingsCaption && !full.isEmpty)
        // The children of the parents unit are shown once, as siblings.
        let ids = full.parentSets.flatMap { $0 }.map(\.id) + full.siblings.map(\.tile.id)
            + full.groups.flatMap { ($0.partner.map { [$0] } ?? []) + $0.children }.map(\.id)
        #expect(Set(ids).count == ids.count)

        let conflicting = content(FamilyFixtures.conflicting)
        #expect(conflicting.parentSets.map(names) == [["Heinrich Weber", "Anna Weber"], ["Heinrich Weber", "Katharina Weber"]])
        #expect(conflicting.siblings.map(\.tile.content.title.text)
            == ["Johann Weber", "Friedrich Weber", "Christian Weber", "Margarethe Weber", "Wilhelm Weber"])
        #expect(conflicting.siblings.map(\.sets) == [[0, 1], [0, 1], [1], [0], [0]])
        let conflictingIDs = conflicting.parentSets.flatMap { $0 }.map(\.id) + conflicting.siblings.map(\.tile.id)
        #expect(Set(conflictingIDs).count == conflictingIDs.count)

        let only = content(FamilyFixtures.siblingsOnly)
        #expect(only.parentSets.isEmpty && only.siblingsCaption && only.groups.isEmpty)
        #expect(only.siblings.map(\.tile.content.title.text) == ["Ida Weber", "Clara Weber", "Bernhard Weber"])

        let anna = content(FamilyFixtures.annaOnly)
        #expect(anna.siblings.count == 1 && anna.siblings[0].tile.content.isFocus && !anna.siblingsCaption)
        #expect(anna.groups.count == 1 && anna.groups[0].children.count == 4)
    }

    @Test func aCoParentIsAPartnerAndNoPersonMeansNoFocus() {
        var family = FamilyFixtures.full
        family.units[1].partners[0].link = nil
        #expect(content(family).groups[0].partner?.content.title.text == "Maria Schulz")
        family.person = nil
        let noFocus = content(family)
        #expect(!noFocus.siblings.contains { $0.tile.content.isFocus })
    }

    @Test func tilesOpenTheRelativeOnFamily() throws {
        let anna = try #require(content(FamilyFixtures.full).parentSets[0].last)
        let summary = FamilyFixtures.summary("anna", "Anna Weber")
        #expect(anna.location == PersonsList.location(summary, orders: NameFormatOrders()).withPageSection(.family))
        #expect(anna.location?.pageSection == .family)
        #expect(anna.content.role == "Parent")
    }

    @Test func captionsAndGroupLabels() {
        let full = content(FamilyFixtures.full)
        #expect(full.parentCaptions == ["Parents"] && full.rightCaption == "Partners and children")
        #expect(full.siblingsLabel == "This person and siblings")
        #expect(full.groups.map(\.label) == ["Partner Maria Schulz and children", "Partner Sophie Brandt and children", "Children"])
        #expect(content(FamilyFixtures.conflicting).parentCaptions == ["Parents", "Other parents"])
        #expect(content(FamilyFixtures.annaOnly).rightCaption == "Partner and children")
        #expect(content(FamilyFixtures.annaOnly).siblingsLabel == "This person")

        let one = CatalogPersonFamily(
            units: [FamilyFixtures.unit(.parents, partners: [FamilyFixtures.member("dorothea", "Dorothea Weber")],
                                        children: [FamilyFixtures.member("heinrich", "Heinrich Weber")])],
            person: FamilyFixtures.johann,
            originSiblings: [FamilyFixtures.origin(FamilyFixtures.member("heinrich", "Heinrich Weber"), [0]),
                             FamilyFixtures.origin(FamilyFixtures.johannMember, [0])]
        )
        #expect(content(one).parentCaptions == ["Parent"])

        var partnerAlone = FamilyFixtures.annaOnly
        partnerAlone.units[0].children = []
        let alone = content(partnerAlone)
        #expect(alone.rightCaption == "Partner" && alone.groups[0].label == "Partner Heinrich Weber")

        let childrenOnly = CatalogPersonFamily(
            units: [FamilyFixtures.unit(.oneParent, children: [FamilyFixtures.member("kid", "Kid Weber")])],
            person: FamilyFixtures.johann,
            originSiblings: [FamilyFixtures.origin(FamilyFixtures.johannMember, [])]
        )
        #expect(content(childrenOnly).rightCaption == "Children")

        var twoPartners = FamilyFixtures.full
        twoPartners.units = [twoPartners.units[1], twoPartners.units[2]]
        twoPartners.units[0].children = []
        twoPartners.units[1].children = []
        #expect(content(twoPartners).rightCaption == "Partners")
    }

    @Test func fiftyMembersKeepOrder() {
        let kids = (0..<50).map { FamilyFixtures.member("kid\($0)", "Kid \($0)") }
        let family = CatalogPersonFamily(
            units: [FamilyFixtures.unit(.partner, partners: [FamilyFixtures.member("sam", "Sam Weber")], children: kids)],
            person: FamilyFixtures.johann,
            originSiblings: [FamilyFixtures.origin(FamilyFixtures.johannMember, [])]
        )
        let tiles = content(family).groups[0].children
        #expect(tiles.count == 50 && names(tiles) == (0..<50).map { "Kid \($0)" })
        #expect(Set(tiles.map(\.id)).count == 50)
    }

    @Test func emptyFamilyIsEmpty() {
        let empty = CatalogPersonFamily(
            units: [], person: FamilyFixtures.johann,
            originSiblings: [FamilyFixtures.origin(FamilyFixtures.johannMember, [])]
        )
        #expect(content(empty).isEmpty)
        #expect(content(CatalogPersonFamily(units: [])).isEmpty)
        #expect(!content(FamilyFixtures.full).isEmpty)
    }
}
