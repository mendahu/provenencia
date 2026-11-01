import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct PlaceRegistryTests {
    private let project = ProjectKey(projectDir: "/tmp/place-registry.provenencia")
    private let registry = PlaceRegistry.standard

    private func resolve(_ location: WorkspaceLocation) -> ResolvedPlace? {
        registry.resolve(location, project: project)
    }

    @Test func resolvesSectionRoots() {
        let sources = resolve(.sectionRoot(.sources))
        #expect(sources?.placeID == .sourcesList)
        #expect(sources?.presentation == .sourcesList)
        #expect(sources?.queryKeys == [
            .sourcesList(project: project),
            .sourceTypesList(project: project),
            .evidenceGraphsList(project: project),
            .evidenceGraphProgress(project: project),
        ])

        let fields = resolve(.sectionRoot(.metadata))
        #expect(fields?.placeID == .metadata)
        #expect(fields?.presentation == .metadata)
        #expect(fields?.queryKeys == [.metadataFieldsList(project: project)])

        let types = resolve(.sectionRoot(.sourceTypes))
        #expect(types?.placeID == .sourceTypes)
        #expect(types?.presentation == .sourceTypes)
        #expect(types?.queryKeys == [
            .sourceTypesList(project: project),
            .metadataFieldsList(project: project),
        ])

        let properties = resolve(.sectionRoot(.properties))
        #expect(properties?.placeID == .properties)
        #expect(properties?.presentation == .properties)
        #expect(properties?.queryKeys == [
            .propertiesWorkspace(project: project),
        ])
    }

    @Test func resolvesSourceDetail() {
        let location = WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Deed")
        let place = resolve(location)
        #expect(place?.placeID == .sourceDetail)
        #expect(place?.presentation == .sourcePage)
        #expect(place?.queryKeys == [
            .sourceWorkspace(project: project, sourceId: "src-1"),
            .sourceTypesList(project: project),
            .metadataFieldsList(project: project),
            .credibilityGradesList(project: project),
            .evidenceGraphsList(project: project),
            .evidenceGraphProgress(project: project),
        ])
    }

    @Test func resolvesEvidenceGraphDistinctFromPage() {
        let page = WorkspaceLocation.sourcePage(sourceId: "src-1")
        let graph = WorkspaceLocation(section: .sources, sourceId: "src-1", graphId: "g-1", sourceSurface: .graph)
        #expect(page != graph)

        let pagePlace = resolve(page)
        #expect(pagePlace?.placeID == .sourceDetail)
        #expect(pagePlace?.presentation == .sourcePage)

        let graphPlace = resolve(graph)
        #expect(graphPlace?.placeID == .sourceGraph)
        #expect(graphPlace?.presentation == .sourceGraph)
        #expect(graphPlace?.queryKeys == [
            .evidenceGraph(project: project, graphId: "g-1"),
            .propertiesWorkspace(project: project),
            .sourcesList(project: project),
            .connectRules(project: project),
            .sourceWorkspace(project: project, sourceId: "src-1"),
            .citationCounts(project: project, graphId: "g-1"),
            .evidenceGraphsList(project: project),
        ])
    }

    @Test func resolvesSourceTypesWithSelection() {
        let location = WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Book")
        let place = resolve(location)
        #expect(place?.placeID == .sourceTypesDetail)
        #expect(place?.presentation == .sourceTypes)
        #expect(place?.queryKeys == [
            .sourceTypesList(project: project),
            .metadataFieldsList(project: project),
            .typeSuggestions(project: project, typeId: "typ-1"),
        ])
    }

    @Test func metadataRowSamePlaceAsRoot() {
        let root = resolve(.sectionRoot(.metadata))
        let row = resolve(WorkspaceLocation(section: .metadata, fieldId: "fld-1", title: "Author"))

        #expect(root?.placeID == .metadata)
        #expect(row?.placeID == .metadata)
        #expect(root?.presentation == row?.presentation)
        #expect(root?.queryKeys == row?.queryKeys)
    }

    @Test func ignoresCrossSectionDeepIds() {
        let location = WorkspaceLocation(
            section: .metadata,
            sourceId: "src-stray",
            fieldId: "fld-1"
        )
        let place = resolve(location)
        #expect(place?.placeID == .metadata)
        #expect(place?.queryKeys == [.metadataFieldsList(project: project)])
    }

    @Test func queryKeysMatchDesignTable() {
        let cases: [(WorkspaceLocation, PlaceID, [CatalogQueryKey])] = [
            (
                .promote(sourceId: "s1", graphId: "g-1", subjectId: "sub-1", kind: .person, ref: "CPR-1", title: "James", sourceTitle: nil),
                .sourcePromote,
                [
                    .evidenceGraph(project: project, graphId: "g-1"),
                    .sourcesList(project: project),
                    .sourceTypesList(project: project),
                    .propertiesWorkspace(project: project),
                    .connectRules(project: project),
                    .evidenceGraphsList(project: project),
                ]
            ),
            (
                .sectionRoot(.sources),
                .sourcesList,
                [
                    .sourcesList(project: project),
                    .sourceTypesList(project: project),
                    .evidenceGraphsList(project: project),
                    .evidenceGraphProgress(project: project),
                ]
            ),
            (
                WorkspaceLocation(section: .sources, sourceId: "s1"),
                .sourceDetail,
                [
                    .sourceWorkspace(project: project, sourceId: "s1"),
                    .sourceTypesList(project: project),
                    .metadataFieldsList(project: project),
                    .credibilityGradesList(project: project),
                    .evidenceGraphsList(project: project),
                    .evidenceGraphProgress(project: project),
                ]
            ),
            (
                WorkspaceLocation(section: .sources, sourceId: "s1", graphId: "g-1", sourceSurface: .graph),
                .sourceGraph,
                [
                    .evidenceGraph(project: project, graphId: "g-1"),
                    .propertiesWorkspace(project: project),
                    .sourcesList(project: project),
                    .connectRules(project: project),
                    .sourceWorkspace(project: project, sourceId: "s1"),
                    .citationCounts(project: project, graphId: "g-1"),
                    .evidenceGraphsList(project: project),
                ]
            ),
            (
                WorkspaceLocation(
                    section: .sources,
                    sourceId: "s1",
                    graphId: "g-1",
                    subjectId: "sub-1",
                    sourceSurface: .citationComposer
                ),
                .sourceCitationComposer,
                [
                    .evidenceGraph(project: project, graphId: "g-1"),
                    .propertiesWorkspace(project: project),
                    .sourcesList(project: project),
                    .connectRules(project: project),
                    .sourceWorkspace(project: project, sourceId: "s1"),
                    .sourceTypesList(project: project),
                    .citationCounts(project: project, graphId: "g-1"),
                    .evidenceGraphsList(project: project),
                ]
            ),
            (
                .sectionRoot(.metadata),
                .metadata,
                [.metadataFieldsList(project: project)]
            ),
            (
                .sectionRoot(.sourceTypes),
                .sourceTypes,
                [.sourceTypesList(project: project), .metadataFieldsList(project: project)]
            ),
            (
                WorkspaceLocation(section: .sourceTypes, typeId: "t1"),
                .sourceTypesDetail,
                [
                    .sourceTypesList(project: project),
                    .metadataFieldsList(project: project),
                    .typeSuggestions(project: project, typeId: "t1"),
                ]
            ),
            (
                .sectionRoot(.properties),
                .properties,
                [.propertiesWorkspace(project: project)]
            ),
        ]

        for (location, expectedID, expectedKeys) in cases {
            let place = resolve(location)
            #expect(place?.placeID == expectedID)
            #expect(place?.queryKeys == expectedKeys)
        }
    }

    @Test func registryCoversAllPlaceIDs() {
        let registered = Set(registry.allPlaceIDs())
        #expect(registered == Set(PlaceID.allCases))
        for id in PlaceID.allCases {
            let location: WorkspaceLocation = switch id {
            case .sourcesList: .sectionRoot(.sources)
            case .sourceDetail: WorkspaceLocation.sourcePage(sourceId: "s1")
            case .sourceGraph: WorkspaceLocation(section: .sources, sourceId: "s1", graphId: "g-1", sourceSurface: .graph)
            case .sourceCitationComposer:
                WorkspaceLocation(
                    section: .sources,
                    sourceId: "s1",
                    graphId: "g-1",
                    subjectId: "sub-1",
                    sourceSurface: .citationComposer
                )
            case .sourcePromote:
                .promote(sourceId: "s1", graphId: "g-1", subjectId: "sub-1", kind: .person, ref: "CPR-1", title: nil, sourceTitle: nil)
            case .metadata: .sectionRoot(.metadata)
            case .sourceTypes: .sectionRoot(.sourceTypes)
            case .sourceTypesDetail: WorkspaceLocation(section: .sourceTypes, typeId: "t1")
            case .properties: .sectionRoot(.properties)
            case .personsList: .sectionRoot(.persons)
            case .personDetail: .personDetail(entityId: "e1", ref: "PER-1", title: nil)
            case .eventsList: .sectionRoot(.events)
            case .eventDetail: .eventDetail(entityId: "e1", ref: "EVT-1", title: "Fire at York")
            case .placesList: .sectionRoot(.places)
            case .placeDetail: .placeDetail(entityId: "e1", ref: "PLC-1", title: "York")
            }
            let place = resolve(location)
            #expect(place?.placeID == id)
            #expect(place != nil)
        }
    }

    @Test func promoteWithoutSubjectFallsBackToTheSourcePage() {
        let location = WorkspaceLocation(section: .sources, sourceId: "s1", graphId: "g-1", sourceSurface: .promote)
        #expect(resolve(location)?.placeID != .sourcePromote)
    }

    @Test func eventsListWarmsTheSummaryRead() {
        let place = resolve(.sectionRoot(.events))
        #expect(place?.placeID == .eventsList)
        #expect(place?.queryKeys == [.eventsList(project: project)])
        let detail = resolve(.eventDetail(entityId: "evt-1", ref: "EVT-1", title: "Fire at York"))
        #expect(detail?.placeID == .eventDetail)
        #expect(detail?.queryKeys == [
            .conclusionDetail(project: project, entityId: "evt-1"),
        ])
        let persons = resolve(.sectionRoot(.persons))
        #expect(persons?.placeID == .personsList)
    }

    @Test func placesListWarmsTheSummaryRead() {
        let place = resolve(.sectionRoot(.places))
        #expect(place?.placeID == .placesList)
        #expect(place?.queryKeys == [.placesList(project: project)])
        let detail = resolve(.placeDetail(entityId: "plc-1", ref: "PLC-1", title: "York"))
        #expect(detail?.placeID == .placeDetail)
        #expect(detail?.queryKeys == [.conclusionDetail(project: project, entityId: "plc-1")])
    }

    @Test func viewIdentityResetsOnTheViewsOwnId() {
        func identity(_ location: WorkspaceLocation) -> String? {
            resolve(location)?.viewIdentity
        }
        let graphA = WorkspaceLocation.evidenceGraph(sourceId: "src-1", graphId: "g-a", title: "A")
        let graphB = WorkspaceLocation.evidenceGraph(sourceId: "src-1", graphId: "g-b", title: "B")
        #expect(identity(graphA) == "g-a")
        #expect(identity(graphB) == "g-b")

        #expect(identity(WorkspaceLocation(section: .sources, sourceId: "src-1", title: "A")) == "src-1")
        #expect(identity(WorkspaceLocation(section: .sources, sourceId: "src-2", title: "B")) == "src-2")

        let composer = WorkspaceLocation(
            section: .sources,
            sourceId: "s1",
            graphId: "g-1",
            citationId: "cit-1",
            artifactId: "art-0",
            sourceSurface: .citationComposer,
            ref: "CIT-AAAAA",
            title: "CIT-AAAAA"
        )
        #expect(identity(composer) != nil)
        #expect(identity(composer) == CitationComposerEntry(location: composer)?.identityKey)
        // Switching Artifacts rewrites the location in place and does not remount.
        var otherArtifact = composer
        otherArtifact.artifactId = "art-1"
        #expect(identity(otherArtifact) == identity(composer))

        let promote = WorkspaceLocation.promote(
            sourceId: "src-1", graphId: "g-1", subjectId: "sub-1", kind: .event, ref: "CEV-1", title: nil, sourceTitle: nil
        )
        #expect(identity(promote) != nil)
        #expect(identity(promote) == PromoteEntry(location: promote)?.identityKey)

        #expect(identity(.personDetail(entityId: "e1", ref: "PER-1", title: "Ann")) == "e1")
        #expect(identity(.eventDetail(entityId: "e1", ref: "EVT-1", title: "Fire")) == "e1")
        #expect(identity(.placeDetail(entityId: "e1", ref: "PLC-1", title: "York")) == "e1")

        #expect(identity(WorkspaceLocation(section: .metadata, fieldId: "fld-1", title: "Author")) == nil)
        #expect(identity(WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Book")) == nil)
        #expect(identity(.sectionRoot(.sources)) == nil)
        #expect(identity(.sectionRoot(.properties)) == nil)
    }

    @Test func citationJumpResolvesComposerWithoutSubjectId() {
        let location = WorkspaceLocation(
            section: .sources,
            sourceId: "s1",
            graphId: "g-1",
            citationId: "cit-1",
            artifactId: "art-0",
            sourceSurface: .citationComposer,
            ref: "CIT-AAAAA",
            title: "CIT-AAAAA"
        )
        let place = resolve(location)
        #expect(place?.placeID == .sourceCitationComposer)
        #expect(place?.presentation == .sourceCitationComposer)
        #expect(CitationComposerEntry(location: location)?.citationID == "cit-1")
        #expect(CitationComposerEntry(location: location)?.subjectID == "")
    }

    @Test func connectPrefillResolvesComposerWithoutSubjectId() {
        let location = WorkspaceLocation(
            section: .sources,
            sourceId: "s1",
            graphId: "g-1",
            connectFromSubjectId: "p1",
            connectToSubjectId: "e1",
            connectBridgeTypeKey: "participation",
            sourceSurface: .citationComposer,
            title: "Margt. participated as Witness at Birth"
        )
        #expect(location.isConnectPrefill)
        let place = resolve(location)
        #expect(place?.placeID == .sourceCitationComposer)
        #expect(place?.presentation == .sourceCitationComposer)

        let decoded = try? JSONDecoder().decode(
            WorkspaceLocation.self,
            from: JSONEncoder().encode(location)
        )
        #expect(decoded == location)
        #expect(decoded?.connectFromSubjectId == "p1")
    }

    @Test func pencilLocationRoundTripsArtifactAndObservation() {
        let location = WorkspaceLocation(
            section: .sources,
            sourceId: "s1",
            graphId: "g-1",
            subjectId: "sub-1",
            citationId: "cit-1",
            artifactId: "art-0",
            observationId: "obs-9",
            sourceSurface: .citationComposer
        )
        let decoded = try? JSONDecoder().decode(
            WorkspaceLocation.self,
            from: JSONEncoder().encode(location)
        )
        #expect(decoded == location)
        #expect(decoded?.artifactId == "art-0")
        #expect(decoded?.observationId == "obs-9")
        #expect(CitationComposerEntry(location: location)?.observationID == "obs-9")
        let place = resolve(location)
        #expect(place?.queryKeys.contains(
            .citationsByArtifact(project: project, graphId: "g-1", artifactId: "art-0")
        ) == true)
    }
}
