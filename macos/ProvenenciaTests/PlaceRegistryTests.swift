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
        #expect(sources?.deepId == nil)
        #expect(sources?.queryKeys == [
            .sourcesList(project: project),
            .sourceTypesList(project: project),
            .sourceGraphProgress(project: project),
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
        #expect(place?.deepId == "src-1")
        #expect(place?.queryKeys == [
            .sourceWorkspace(project: project, sourceId: "src-1"),
            .sourceTypesList(project: project),
            .metadataFieldsList(project: project),
            .credibilityGradesList(project: project),
        ])
    }

    @Test func resolvesSourceGraphDistinctFromPage() {
        let page = WorkspaceLocation(section: .sources, sourceId: "src-1", sourceSurface: .page)
        let graph = WorkspaceLocation(section: .sources, sourceId: "src-1", sourceSurface: .graph)
        #expect(page != graph)

        let pagePlace = resolve(page)
        #expect(pagePlace?.placeID == .sourceDetail)
        #expect(pagePlace?.presentation == .sourcePage)

        let graphPlace = resolve(graph)
        #expect(graphPlace?.placeID == .sourceGraph)
        #expect(graphPlace?.presentation == .sourceGraph)
        #expect(graphPlace?.deepId == "src-1")
        #expect(graphPlace?.queryKeys == [
            .sourceGraph(project: project, sourceId: "src-1"),
            .propertiesWorkspace(project: project),
            .sourcesList(project: project),
            .connectRules(project: project),
        ])
    }

    @Test func resolvesSourceTypesWithSelection() {
        let location = WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Book")
        let place = resolve(location)
        #expect(place?.placeID == .sourceTypesDetail)
        #expect(place?.presentation == .sourceTypes)
        #expect(place?.deepId == "typ-1")
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
        #expect(root?.deepId == nil)
        #expect(row?.deepId == "fld-1")
    }

    @Test func ignoresCrossSectionDeepIds() {
        let location = WorkspaceLocation(
            section: .metadata,
            sourceId: "src-stray",
            fieldId: "fld-1"
        )
        let place = resolve(location)
        #expect(place?.placeID == .metadata)
        #expect(place?.deepId == "fld-1")
        #expect(place?.queryKeys == [.metadataFieldsList(project: project)])
    }

    @Test func queryKeysMatchDesignTable() {
        let cases: [(WorkspaceLocation, PlaceID, [CatalogQueryKey])] = [
            (
                .promote(sourceId: "s1", subjectId: "sub-1", kind: .person, ref: "CPR-1", title: "James", sourceTitle: nil),
                .sourcePromote,
                [
                    .sourceGraph(project: project, sourceId: "s1"),
                    .sourcesList(project: project),
                    .promoteTargets(project: project, subjectId: "sub-1"),
                    .confidenceGradesList(project: project),
                    .propertiesWorkspace(project: project),
                ]
            ),
            (
                .sectionRoot(.sources),
                .sourcesList,
                [
                    .sourcesList(project: project),
                    .sourceTypesList(project: project),
                    .sourceGraphProgress(project: project),
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
                ]
            ),
            (
                WorkspaceLocation(section: .sources, sourceId: "s1", sourceSurface: .graph),
                .sourceGraph,
                [
                    .sourceGraph(project: project, sourceId: "s1"),
                    .propertiesWorkspace(project: project),
                    .sourcesList(project: project),
                    .connectRules(project: project),
                ]
            ),
            (
                WorkspaceLocation(
                    section: .sources,
                    sourceId: "s1",
                    subjectId: "sub-1",
                    sourceSurface: .citationComposer
                ),
                .sourceCitationComposer,
                [
                    .sourceGraph(project: project, sourceId: "s1"),
                    .propertiesWorkspace(project: project),
                    .sourcesList(project: project),
                    .connectRules(project: project),
                    .sourceWorkspace(project: project, sourceId: "s1"),
                    .sourceTypesList(project: project),
                    .citationCounts(project: project, sourceId: "s1"),
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
            case .sourceDetail: WorkspaceLocation(section: .sources, sourceId: "s1", sourceSurface: .page)
            case .sourceGraph: WorkspaceLocation(section: .sources, sourceId: "s1", sourceSurface: .graph)
            case .sourceCitationComposer:
                WorkspaceLocation(
                    section: .sources,
                    sourceId: "s1",
                    subjectId: "sub-1",
                    sourceSurface: .citationComposer
                )
            case .sourcePromote:
                .promote(sourceId: "s1", subjectId: "sub-1", kind: .person, ref: "CPR-1", title: nil, sourceTitle: nil)
            case .metadata: .sectionRoot(.metadata)
            case .sourceTypes: .sectionRoot(.sourceTypes)
            case .sourceTypesDetail: WorkspaceLocation(section: .sourceTypes, typeId: "t1")
            case .properties: .sectionRoot(.properties)
            case .personsList: .sectionRoot(.persons)
            case .personDetail: .personDetail(entityId: "e1", ref: "PER-1", title: nil)
            case .eventsList: .sectionRoot(.events)
            case .placesList: .sectionRoot(.places)
            }
            let place = resolve(location)
            #expect(place?.placeID == id)
            #expect(place != nil)
        }
    }

    @Test func promoteWithoutSubjectFallsBackToTheSourcePage() {
        let location = WorkspaceLocation(section: .sources, sourceId: "s1", sourceSurface: .promote)
        #expect(resolve(location)?.placeID != .sourcePromote)
    }

    @Test func conclusionStubsResolveWithNoQueryKeys() {
        for (section, id) in [(WorkspaceSection.events, PlaceID.eventsList), (.places, .placesList)] {
            let place = resolve(.sectionRoot(section))
            #expect(place?.placeID == id)
            #expect(place?.queryKeys == [])
        }
    }

    @Test func citationJumpResolvesComposerWithoutSubjectId() {
        let location = WorkspaceLocation(
            section: .sources,
            sourceId: "s1",
            citationId: "cit-1",
            artifactId: "art-0",
            sourceSurface: .citationComposer,
            ref: "CIT-AAAAA",
            title: "CIT-AAAAA"
        )
        let place = resolve(location)
        #expect(place?.placeID == .sourceCitationComposer)
        #expect(place?.presentation == .sourceCitationComposer)
        #expect(place?.deepId == "cit-1")
        #expect(CitationComposerEntry(location: location)?.citationID == "cit-1")
        #expect(CitationComposerEntry(location: location)?.subjectID == "")
    }

    @Test func connectPrefillResolvesComposerWithoutSubjectId() {
        let location = WorkspaceLocation(
            section: .sources,
            sourceId: "s1",
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
            .citationsByArtifact(project: project, artifactId: "art-0")
        ) == true)
    }
}
