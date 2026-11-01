import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct CatalogSearchModelTests {
    private let projectDir = "/tmp/catalog-search.provenencia"

    @Test func belowMinimumQueryIsTheTwoLetterFloor() {
        let model = CatalogSearchModel()
        for query in ["", "P", " P "] {
            model.query = query
            #expect(model.isBelowMinimumQuery)
        }
        model.query = "Po"
        #expect(model.isBelowMinimumQuery == false)
    }

    @Test func accessibilityLabelUsesTheGivenNameOrder() {
        let name = CatalogNameValue(form: "John Henry Smith", parts: [
            CatalogNameValuePart(value: "John Henry", type: "given"),
            CatalogNameValuePart(value: "Smith", type: "surname"),
        ])
        let summary = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "t", label: ""),
            name: name, nameValueCount: 1
        )
        let hit = CatalogSearchHit(
            kind: "person", id: "e1", ref: "PER-1", title: "John Henry Smith", subtitle: "", matchReason: "title",
            location: WorkspaceLocation(section: .persons, entityId: "e1", ref: "PER-1", title: "John Henry Smith"),
            summary: .person(summary)
        )
        let orders = NameFormatOrders(natural: [], sorted: [
            NameFormatPart(partType: "surname", trailingSeparator: ", "),
            NameFormatPart(partType: "given", trailingSeparator: ""),
        ])
        let title = CatalogSearchHitPresentation.title(for: hit, orders: orders)
        #expect(title == "Smith, John Henry")
        #expect(CatalogSearchHitPresentation.accessibilityLabel(for: hit, orders: orders).hasPrefix(title))
        #expect(CatalogSearchHitPresentation.accessibilityLabel(for: hit).hasPrefix(CatalogSearchHitPresentation.title(for: hit)))
    }

    @Test func exactRefHitIsTop() async {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-ZZ9K2", sourceTypeID: "", title: "Quiet", description: ""),
            CatalogSource(id: "s2", ref: "SRC-OTHER", sourceTypeID: "", title: "SRC-ZZ9K2 in title", description: ""),
        ]
        let model = CatalogSearchModel()
        model.query = "SRC-ZZ9K2"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        await waitForSearch(model)
        #expect(model.hits.first?.id == "s1")
        #expect(model.hits.first?.matchReason == "ref")
        if let top = model.hits.first {
            #expect(CatalogSearchHitPresentation.refAccent(for: top))
        }
    }

    @Test func emptyStateRequiresCompletedSearchWithoutError() {
        let model = CatalogSearchModel()
        model.query = "zz"
        model.hasSearched = true
        model.isLoading = false
        #expect(model.showsEmptyState)
        model.searchError = "boom"
        #expect(!model.showsEmptyState)
        #expect(model.showsErrorState)
    }

    @Test func moveSelectionClampsAndSelectedHit() {
        let model = CatalogSearchModel()
        model.hits = [
            CatalogSearchHit(
                kind: "source", id: "a", ref: "", title: "A", subtitle: "",
                matchReason: "title", location: .sectionRoot(.sources)
            ),
            CatalogSearchHit(
                kind: "source", id: "b", ref: "", title: "B", subtitle: "",
                matchReason: "title", location: .sectionRoot(.sources)
            ),
        ]
        #expect(model.selectedHit()?.id == "a")
        model.moveSelection(delta: 1)
        #expect(model.selectedIndex == 1)
        #expect(model.selectedHit()?.id == "b")
        model.moveSelection(delta: 5)
        #expect(model.selectedIndex == 1)
        model.moveSelection(delta: -10)
        #expect(model.selectedIndex == 0)
    }

    @Test func kindsRestrictHits() async throws {
        let (store, james) = try await robinsSearchStore(projectDir: projectDir)
        let model = CatalogSearchModel(kinds: ["person"])
        model.query = "Robins"
        model.scheduleSearch(projectDir: projectDir, location: nil, store: store)
        await waitForSearch(model)
        #expect(model.hits.map(\.kind) == ["person"])
        #expect(model.hits.first?.ref == james.entity.ref)
        #expect(model.hasSearched)
        #expect(model.searchError == nil)
    }

    @Test func noLocationGivesNoSectionNudge() async throws {
        let (store, _) = try await robinsSearchStore(projectDir: projectDir)
        let model = CatalogSearchModel()
        model.query = "Robins"
        model.scheduleSearch(projectDir: projectDir, location: nil, store: store)
        await waitForSearch(model)
        #expect(model.hits.map(\.kind) == ["person", "source"])
    }

    @Test func cancelKeepsResultsAndStopsLoading() {
        let model = CatalogSearchModel()
        model.query = "Ilminster"
        model.hasSearched = true
        model.hits = [
            CatalogSearchHit(
                kind: "source", id: "s1", ref: "SRC-AAAAA", title: "A", subtitle: "",
                matchReason: "title", location: .sectionRoot(.sources)
            ),
        ]
        model.isLoading = true
        model.cancel()
        #expect(!model.isLoading)
        #expect(model.query == "Ilminster")
        #expect(model.hits.count == 1)
        #expect(model.hasSearched)
    }

    @Test func showsResultsNeedsALongQueryAndASearch() {
        let model = CatalogSearchModel()
        model.query = "I"
        model.hasSearched = true
        #expect(!model.showsResults)
        model.query = "Il"
        model.hasSearched = false
        #expect(!model.showsResults)
        model.isLoading = true
        #expect(model.showsResults)
        model.isLoading = false
        model.hasSearched = true
        #expect(model.showsResults)
    }
}

@Suite
struct CatalogSearchHitPresentationTests {
    @Test func aSummaryIsTitledByItsKind() {
        func entity(_ ref: String) -> CatalogCanonicalEntity {
            CatalogCanonicalEntity(id: ref, ref: ref, subjectTypeID: "t", label: "")
        }
        let person = CatalogPersonSummary(
            entity: entity("PER-1"),
            name: CatalogNameValue(form: "John Smith", parts: [
                CatalogNameValuePart(value: "John", type: "given"),
                CatalogNameValuePart(value: "Smith", type: "surname"),
            ]),
            nameValueCount: 1
        )
        let event = CatalogEventSummary(
            entity: entity("EVT-1"),
            title: CatalogEventTitle(rule: .typeAtPlace, ref: "EVT-1", typeLabel: "Birth", place: "Gumpertz")
        )
        let place = CatalogPlaceSummary(entity: entity("PLC-1"), names: ["The farm"], parents: ["Ontario", "Canada"])

        #expect(ConclusionSummaryDisplay.title(.person(person), style: .sorted)
            == PersonTitleDisplay.title(person, style: .sorted))
        #expect(ConclusionSummaryDisplay.title(.event(event)) == EventTitleDisplay.title(event))
        #expect(ConclusionSummaryDisplay.title(.place(place)) == PlaceTitleDisplay.title(place))
        #expect(ConclusionSummaryDisplay.secondary(.person(person)) == PersonLifeDisplay.line(person).text)
        #expect(ConclusionSummaryDisplay.secondary(.event(event)) == EventSecondaryDisplay.line(event))
        #expect(ConclusionSummaryDisplay.secondary(.place(place))
            == PlaceChainDisplay.line(parents: ["Ontario", "Canada"], candidates: false))
        #expect(ConclusionSummaryDisplay.ref(.person(person)) == "PER-1")
        #expect(ConclusionSummaryDisplay.ref(.event(event)) == "EVT-1")
        #expect(ConclusionSummaryDisplay.ref(.place(place)) == "PLC-1")
    }

    @Test func accessibilityLabelComposesVisibleRowMetadata() {
        let hit = CatalogSearchHit(
            kind: "source",
            id: "s1",
            ref: "SRC-3K9M2",
            title: "Ilminster parish register",
            subtitle: "Parish register",
            matchReason: "notes",
            matchSnippet: "Zemblanity",
            location: .sectionRoot(.sources)
        )
        let label = CatalogSearchHitPresentation.accessibilityLabel(for: hit)
        #expect(label.contains("Ilminster parish register"))
        #expect(label.contains(L10n.string(L10n.Workspace.omnibarKindSource)))
        #expect(label.contains("Parish register"))
        #expect(label.contains("SRC-3K9M2"))
        #expect(label.contains(L10n.Workspace.omnibarMatchNote(snippet: "Zemblanity")))
    }

    @Test func accessibilityLabelOmitsEmptyOptionalParts() {
        let hit = CatalogSearchHit(
            kind: "metadata_field",
            id: "f1",
            ref: "",
            title: "Citation",
            subtitle: "",
            matchReason: "title",
            location: .sectionRoot(.metadata)
        )
        let label = CatalogSearchHitPresentation.accessibilityLabel(for: hit)
        #expect(label.contains("Citation"))
        #expect(label.contains(L10n.string(L10n.Workspace.omnibarKindMetadataField)))
        #expect(!label.contains("Note:"))
        #expect(!label.contains("Metadata:"))
    }

    @Test func matchContextLocalizesBodyFieldsAndHidesTitleRef() {
        #expect(CatalogSearchHitPresentation.matchContextText(field: "title", snippet: "") == "")
        #expect(CatalogSearchHitPresentation.matchContextText(field: "ref", snippet: "") == "")
        #expect(CatalogSearchHitPresentation.matchContextText(field: "fuzzy", snippet: "") == "")
        #expect(
            CatalogSearchHitPresentation.matchContextText(field: "notes", snippet: "Zemblanity")
                == L10n.Workspace.omnibarMatchNote(snippet: "Zemblanity")
        )
        #expect(
            CatalogSearchHitPresentation.matchContextText(field: "metadata", snippet: "author")
                == L10n.Workspace.omnibarMatchMetadata(snippet: "author")
        )
        #expect(
            CatalogSearchHitPresentation.matchContextText(field: "filename", snippet: "scan.png")
                == L10n.Workspace.omnibarMatchFilename(snippet: "scan.png")
        )
        #expect(
            CatalogSearchHitPresentation.matchContextText(field: "description", snippet: "")
                == L10n.string(L10n.Workspace.omnibarMatchDescription)
        )
        #expect(
            CatalogSearchHitPresentation.matchContextText(field: "name", snippet: "Jim Robins")
                == L10n.Workspace.omnibarMatchName(snippet: "Jim Robins")
        )
        #expect(
            CatalogSearchHitPresentation.matchContextText(field: "place", snippet: "Montreal")
                == L10n.Workspace.omnibarMatchPlace(snippet: "Montreal")
        )
        #expect(L10n.string(CatalogSearchHitPresentation.kindLabel(for: .person)) == L10n.string(L10n.Workspace.omnibarKindPerson))
        #expect(CatalogSearchHitPresentation.leadThumbnail(for: leadHit(kind: "event"))?.glyph == .mark(.subjectEvent))
        #expect(CatalogSearchHitPresentation.leadThumbnail(for: leadHit(kind: "event"))?.label == nil)
        let typeLead = CatalogSearchHitPresentation.leadThumbnail(for: leadHit(kind: "source_type", iconKey: "type_book"))
        #expect(typeLead?.glyph == .mark(.typeBook))
        #expect(typeLead?.label != nil)
        #expect(CatalogSearchHitPresentation.leadThumbnail(for: leadHit(kind: "source_type"))?.glyph == .symbol(.tag))
        #expect(CatalogSearchHitPresentation.leadThumbnail(for: leadHit(kind: "metadata_field"))?.glyph == .symbol(.list))
        #expect(CatalogSearchHitPresentation.leadThumbnail(for: leadHit(kind: "source")) == nil)
        #expect(CatalogSearchHitPresentation.refAccent(for: CatalogSearchHit(
            kind: "person", id: "p", ref: "PER-7KD45", title: "James Robins", subtitle: "",
            matchReason: "ref", location: .sectionRoot(.persons)
        )))
    }

    @Test func evidenceGraphHitPresentation() {
        let hit = CatalogSearchHit(
            kind: "evidence_graph", id: "g", ref: "GPH-J0NES", title: "Jones household",
            subtitle: "1881 Census, Ward 3", matchReason: "description", location: .sectionRoot(.sources)
        )
        #expect(hit.searchKind == .evidenceGraph)
        #expect(
            L10n.string(CatalogSearchHitPresentation.kindLabel(for: .evidenceGraph))
                == L10n.string(L10n.Workspace.omnibarKindEvidenceGraph)
        )
        #expect(CatalogSearchHitPresentation.leadThumbnail(for: hit)?.glyph == .symbol(.gitBranch))
        let expected = ListFormatter.localizedString(byJoining: [
            "Jones household", L10n.string(L10n.Workspace.omnibarKindEvidenceGraph),
            "1881 Census, Ward 3", "GPH-J0NES", L10n.string(L10n.Workspace.omnibarMatchDescription),
        ])
        #expect(CatalogSearchHitPresentation.accessibilityLabel(for: hit) == expected)
    }

    @Test func unknownKindKeepsTheFallbacks() {
        let hit = leadHit(kind: "widget")
        #expect(hit.searchKind == nil)
        #expect(L10n.string(CatalogSearchHitPresentation.kindLabel(for: hit.searchKind)) == L10n.string(L10n.Workspace.omnibarKindSource))
        #expect(CatalogSearchHitPresentation.leadThumbnail(for: hit)?.glyph == .symbol(.search))
        #expect(leadHit(kind: "source_type").searchKind == .sourceType)
    }

    private func leadHit(kind: String, iconKey: String = "") -> CatalogSearchHit {
        CatalogSearchHit(
            kind: kind, id: "h", ref: "REF-AAAAA", title: "Title", subtitle: "",
            matchReason: "title", location: .sectionRoot(.persons), iconKey: iconKey
        )
    }
}
