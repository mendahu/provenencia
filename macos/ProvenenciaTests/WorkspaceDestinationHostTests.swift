import Foundation
import Testing
@testable import Provenencia

@Suite
struct WorkspaceDestinationHostTests {
    private let projectDir = "/tmp/workspace-destination-host.provenencia"

    private func presentation(for location: WorkspaceLocation) -> WorkspacePresentationID {
        WorkspaceDestinationHost.presentation(for: location, projectDir: projectDir)
    }

    @Test func presentationSourcesRoot() {
        #expect(presentation(for: .sectionRoot(.sources)) == .sourcesList)
    }

    @Test func presentationSourceDetail() {
        #expect(
            presentation(for: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Deed"))
                == .sourcePage
        )
    }

    @Test func presentationSourceGraph() {
        #expect(
            presentation(
                for: WorkspaceLocation(section: .sources, sourceId: "src-1", sourceSurface: .graph)
            ) == .sourceGraph
        )
    }

    @Test func presentationCitationComposer() {
        #expect(
            presentation(
                for: WorkspaceLocation(
                    section: .sources,
                    sourceId: "src-1",
                    subjectId: "sub-1",
                    sourceSurface: .citationComposer
                )
            ) == .sourceCitationComposer
        )
    }

    @Test func presentationSourceFieldsRootAndRow() {
        #expect(presentation(for: .sectionRoot(.sourceFields)) == .sourceFields)
        #expect(
            presentation(for: WorkspaceLocation(section: .sourceFields, fieldId: "fld-1", title: "Author"))
                == .sourceFields
        )
    }

    @Test func presentationSourceTypesRootAndRow() {
        #expect(presentation(for: .sectionRoot(.sourceTypes)) == .sourceTypes)
        #expect(
            presentation(for: WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Book"))
                == .sourceTypes
        )
    }

    @Test func presentationSubjectFields() {
        #expect(presentation(for: .sectionRoot(.subjectFields)) == .subjectFields)
    }

    @Test func sourceFamilyPresentationsUseSourcesDestination() {
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourcesList) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourcePage) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourceGraph) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourceCitationComposer) == .sources)
    }

    @Test func subjectFieldsPresentationHasOwnKind() {
        #expect(WorkspaceDestinationHost.destinationKind(for: .subjectFields) == .subjectFields)
    }

    @Test func registryPresentationsAreKnown() {
        let project = ProjectKey(projectDir: projectDir)
        let registry = PlaceRegistry.standard
        let locations: [WorkspaceLocation] = [
            .sectionRoot(.sources),
            WorkspaceLocation(section: .sources, sourceId: "s1"),
            WorkspaceLocation(section: .sources, sourceId: "s1", sourceSurface: .graph),
            WorkspaceLocation(
                section: .sources,
                sourceId: "s1",
                subjectId: "sub-1",
                sourceSurface: .citationComposer
            ),
            .sectionRoot(.sourceFields),
            WorkspaceLocation(section: .sourceFields, fieldId: "f1"),
            .sectionRoot(.sourceTypes),
            WorkspaceLocation(section: .sourceTypes, typeId: "t1"),
            .sectionRoot(.subjectFields),
        ]
        let known = Set(WorkspacePresentationID.allCases)
        for location in locations {
            guard let place = registry.resolve(location, project: project) else {
                Issue.record("Expected place for \(location)")
                continue
            }
            #expect(known.contains(place.presentation))
        }
        #expect(known.count == 7)
    }
}