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
        ])

        let fields = resolve(.sectionRoot(.sourceFields))
        #expect(fields?.placeID == .sourceFields)
        #expect(fields?.presentation == .sourceFields)
        #expect(fields?.queryKeys == [.metadataFieldsList(project: project)])

        let types = resolve(.sectionRoot(.sourceTypes))
        #expect(types?.placeID == .sourceTypes)
        #expect(types?.presentation == .sourceTypes)
        #expect(types?.queryKeys == [
            .sourceTypesList(project: project),
            .metadataFieldsList(project: project),
        ])

        let subjectFields = resolve(.sectionRoot(.subjectFields))
        #expect(subjectFields?.placeID == .subjectFields)
        #expect(subjectFields?.presentation == .subjectFields)
        #expect(subjectFields?.queryKeys == [
            .subjectFieldsWorkspace(project: project),
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
        #expect(graphPlace?.queryKeys == [.sourceGraph(project: project, sourceId: "src-1")])
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

    @Test func sourceFieldsRowSamePlaceAsRoot() {
        let root = resolve(.sectionRoot(.sourceFields))
        let row = resolve(WorkspaceLocation(section: .sourceFields, fieldId: "fld-1", title: "Author"))

        #expect(root?.placeID == .sourceFields)
        #expect(row?.placeID == .sourceFields)
        #expect(root?.presentation == row?.presentation)
        #expect(root?.queryKeys == row?.queryKeys)
        #expect(root?.deepId == nil)
        #expect(row?.deepId == "fld-1")
    }

    @Test func ignoresCrossSectionDeepIds() {
        let location = WorkspaceLocation(
            section: .sourceFields,
            sourceId: "src-stray",
            fieldId: "fld-1"
        )
        let place = resolve(location)
        #expect(place?.placeID == .sourceFields)
        #expect(place?.deepId == "fld-1")
        #expect(place?.queryKeys == [.metadataFieldsList(project: project)])
    }

    @Test func queryKeysMatchDesignTable() {
        let cases: [(WorkspaceLocation, PlaceID, [CatalogQueryKey])] = [
            (
                .sectionRoot(.sources),
                .sourcesList,
                [.sourcesList(project: project), .sourceTypesList(project: project)]
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
                [.sourceGraph(project: project, sourceId: "s1")]
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
                    .sourceWorkspace(project: project, sourceId: "s1"),
                ]
            ),
            (
                .sectionRoot(.sourceFields),
                .sourceFields,
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
                .sectionRoot(.subjectFields),
                .subjectFields,
                [.subjectFieldsWorkspace(project: project)]
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
            case .sourceFields: .sectionRoot(.sourceFields)
            case .sourceTypes: .sectionRoot(.sourceTypes)
            case .sourceTypesDetail: WorkspaceLocation(section: .sourceTypes, typeId: "t1")
            case .subjectFields: .sectionRoot(.subjectFields)
            }
            let place = resolve(location)
            #expect(place?.placeID == id)
            #expect(place != nil)
        }
    }
}
