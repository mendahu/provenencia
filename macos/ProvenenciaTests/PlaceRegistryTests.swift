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
    }

    @Test func resolvesSourceDetail() {
        let location = WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Deed")
        let place = resolve(location)
        #expect(place?.placeID == .sourceDetail)
        #expect(place?.presentation == .sourcePage)
        #expect(place?.deepId == "src-1")
        // Page payload + the three shared vocabulary lists it reads.
        #expect(place?.queryKeys == [
            .sourceWorkspace(project: project, sourceId: "src-1"),
            .sourceTypesList(project: project),
            .metadataFieldsList(project: project),
            .credibilityGradesList(project: project),
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
            case .sourceDetail: WorkspaceLocation(section: .sources, sourceId: "s1")
            case .sourceFields: .sectionRoot(.sourceFields)
            case .sourceTypes: .sectionRoot(.sourceTypes)
            case .sourceTypesDetail: WorkspaceLocation(section: .sourceTypes, typeId: "t1")
            }
            let place = resolve(location)
            #expect(place?.placeID == id)
            #expect(place != nil)
        }
    }
}
