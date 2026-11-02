import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

struct PersonFamilyLayoutTests {
    private let en = Locale(identifier: "en_US")

    private func layout(_ family: CatalogPersonFamily, width: CGFloat = 1038, _ widthClass: PVWidthClass = .regular) -> (PersonFamilyContent, PersonFamilyLayout) {
        let content = PersonFamilyContent(family: family, locale: en)
        return (content, PersonFamilyLayout(content: content, width: width, widthClass: widthClass))
    }

    private func frame(_ layout: PersonFamilyLayout, _ content: PersonFamilyContent, _ name: String) -> CGRect? {
        let tiles = content.parentSets.flatMap { $0 } + content.siblings.map(\.tile)
            + content.groups.flatMap { ($0.partner.map { [$0] } ?? []) + $0.children }
        guard let tile = tiles.first(where: { $0.content.title.text == name }) else { return nil }
        return layout.frame(ofTile: tile.id)
    }

    private func verticals(_ layout: PersonFamilyLayout, at x: CGFloat) -> [CGRect] {
        layout.segments.filter { $0.minX == x && $0.width == 1 && $0.height > 1 }
    }

    private func horizontals(_ layout: PersonFamilyLayout, at y: CGFloat) -> [CGRect] {
        layout.segments.filter { $0.minY == y && $0.height == 1 && $0.width > 1 }
    }

    @Test func regularAndWideMeasures() throws {
        let (content, l) = layout(PersonFamilyPreviewData.full)
        #expect(l.columnWidth == 485 && l.rightX == 553 && l.busX == 521)
        let margarethe = try #require(frame(l, content, "Margarethe Weber"))
        #expect(margarethe.minX == 52 && margarethe.width == 453 && margarethe.minY == 242)
        let maria = try #require(frame(l, content, "Maria Schulz"))
        #expect(maria.minX == 553 && maria.minY == 26 && maria.midY == 72 && maria.width == 485)
        let paul = try #require(frame(l, content, "Paul Weber"))
        #expect(paul.minX == 585 && paul.midY == 728)
        let johann = try #require(frame(l, content, "Johann Weber"))
        // The focus's horizontal ends at the bus; the bus spans Maria's centre to Paul's.
        #expect(horizontals(l, at: johann.midY).contains { $0.minX == johann.maxX && $0.maxX == 521 })
        let bus = try #require(verticals(l, at: 521).first)
        #expect(bus.minY == 72 && bus.maxY == 729)
        // The rail of the one parent set.
        let rail = try #require(verticals(l, at: 10).first)
        #expect(rail.minY == 72 && rail.maxY == 601)
        #expect(horizontals(l, at: 72).contains { $0.minX == 10 && $0.maxX == 20 })
        // Tiles are listed left column, then right column.
        #expect(l.groups.map(\.label) == [
            "Parents", "This person and siblings",
            "Partner Maria Schulz and children", "Partner Sophie Brandt and children", "Children",
        ])
        #expect(l.captions.map(\.text) == ["Parents", "Partners and children"])

        let (_, regular) = layout(PersonFamilyPreviewData.full, width: 1180, .regular)
        #expect(regular.columnWidth == 556 && PersonFamilyLayout.gutter(.regular) == 48)
        let (_, wide) = layout(PersonFamilyPreviewData.full, width: 1180, .wide)
        #expect(wide.columnWidth == 532 && PersonFamilyLayout.gutter(.wide) == 96 && wide.rightX == 20 + 532 + 96)
    }

    @Test func parentSetRails() throws {
        let (content, l) = layout(PersonFamilyPreviewData.conflicting)
        let rail0 = try #require(verticals(l, at: 6).first)
        let rail1 = try #require(verticals(l, at: 14).first)
        let wilhelm = try #require(frame(l, content, "Wilhelm Weber"))
        let christian = try #require(frame(l, content, "Christian Weber"))
        #expect(rail0.maxY == wilhelm.midY + 1 && rail1.maxY == christian.midY + 1)
        // The second rail draws a horizontal only to Christian.
        #expect(horizontals(l, at: christian.midY).contains { $0.minX == 14 && $0.maxX == 52 })
        let friedrich = try #require(frame(l, content, "Friedrich Weber"))
        #expect(!horizontals(l, at: friedrich.midY).contains { $0.minX == 14 })

        #expect([0, 1, 2].map { PersonFamilyLayout.railX($0, of: 3) } == [4, 10, 16])
        #expect(PersonFamilyLayout.railX(0, of: 1) == 10)
    }

    @Test func siblingsOnlyAndTheFocusAlone() throws {
        let (content, l) = layout(PersonFamilyPreviewData.siblingsOnly)
        let ida = try #require(frame(l, content, "Ida Weber"))
        #expect(ida.minX == 20 && ida.minY == 26 && !verticals(l, at: 10).isEmpty)
        #expect(l.captions.map(\.text) == ["Siblings"])
        let (annaContent, anna) = layout(PersonFamilyPreviewData.annaOnly)
        let focus = try #require(frame(anna, annaContent, "Anna Weber"))
        #expect(focus.origin == CGPoint(x: 20, y: 26))
        #expect(anna.captions.map(\.text) == ["Partner and children"])
    }

    @Test func fiftySiblingsStack() {
        let kids = (0..<49).map { PersonFamilyPreviewData.member("kid\($0)", "Kid \($0)") }
        let family = CatalogPersonFamily(
            units: [PersonFamilyPreviewData.unit(.parents, partners: [PersonFamilyPreviewData.member("a", "A"), PersonFamilyPreviewData.member("b", "B")], children: kids)],
            person: PersonFamilyPreviewData.johann,
            originSiblings: kids.map { PersonFamilyPreviewData.origin($0, [0]) } + [PersonFamilyPreviewData.origin(PersonFamilyPreviewData.johannMember, [0])]
        )
        let (content, l) = layout(family)
        let ids = content.parentSets.flatMap { $0 }.map(\.id) + content.siblings.map(\.tile.id)
        let frames = ids.compactMap { l.frame(ofTile: $0) }
        #expect(frames.count == 52)
        let captionAndParents: CGFloat = 26 + 2 * 104 - 12
        let siblings: CGFloat = 50 * 104 - 12
        #expect(l.height == captionAndParents + 20 + siblings)
        for (a, b) in zip(frames, frames.dropFirst()) { #expect(!a.intersects(b)) }
    }
}
