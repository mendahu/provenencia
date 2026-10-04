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

    @Test func presentationPromote() {
        let location = WorkspaceLocation.promote(
            sourceId: "src-1", subjectId: "sub-1", kind: .event, ref: "CEV-1", title: nil, sourceTitle: nil
        )
        #expect(presentation(for: location) == .sourcePromote)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourcePromote) == .sources)
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

    @Test func presentationCitationJumpComposer() {
        #expect(
            presentation(
                for: WorkspaceLocation(
                    section: .sources,
                    sourceId: "src-1",
                    citationId: "cit-1",
                    artifactId: "art-0",
                    sourceSurface: .citationComposer
                )
            ) == .sourceCitationComposer
        )
    }

    @Test func presentationConnectPrefillComposer() {
        #expect(
            presentation(
                for: WorkspaceLocation(
                    section: .sources,
                    sourceId: "src-1",
                    connectFromSubjectId: "p1",
                    connectToSubjectId: "e1",
                    connectBridgeTypeKey: "participation",
                    sourceSurface: .citationComposer
                )
            ) == .sourceCitationComposer
        )
    }

    @Test func presentationMetadataRootAndRow() {
        #expect(presentation(for: .sectionRoot(.metadata)) == .metadata)
        #expect(
            presentation(for: WorkspaceLocation(section: .metadata, fieldId: "fld-1", title: "Author"))
                == .metadata
        )
    }

    @Test func presentationSourceTypesRootAndRow() {
        #expect(presentation(for: .sectionRoot(.sourceTypes)) == .sourceTypes)
        #expect(
            presentation(for: WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Book"))
                == .sourceTypes
        )
    }

    @Test func presentationProperties() {
        #expect(presentation(for: .sectionRoot(.properties)) == .properties)
    }

    @Test func sourceFamilyPresentationsUseSourcesDestination() {
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourcesList) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourcePage) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourceGraph) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourceCitationComposer) == .sources)
    }

    @Test func conclusionPresentationsHaveOwnKinds() {
        #expect(presentation(for: .sectionRoot(.persons)) == .personsList)
        #expect(presentation(for: .sectionRoot(.events)) == .eventsList)
        #expect(presentation(for: .sectionRoot(.places)) == .placesList)
        #expect(WorkspaceDestinationHost.destinationKind(for: .personsList) == .persons)
        #expect(presentation(for: .personDetail(entityId: "e1", ref: "PER-1", title: nil)) == .personDetail)
        #expect(WorkspaceDestinationHost.destinationKind(for: .personDetail) == .persons)
        #expect(WorkspaceDestinationHost.destinationKind(for: .eventsList) == .events)
        #expect(WorkspaceDestinationHost.destinationKind(for: .placesList) == .places)
    }

    @Test func propertiesPresentationHasOwnKind() {
        #expect(WorkspaceDestinationHost.destinationKind(for: .properties) == .properties)
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
            .sectionRoot(.metadata),
            WorkspaceLocation(section: .metadata, fieldId: "f1"),
            .sectionRoot(.sourceTypes),
            WorkspaceLocation(section: .sourceTypes, typeId: "t1"),
            .sectionRoot(.properties),
            .sectionRoot(.persons),
            .sectionRoot(.events),
            .sectionRoot(.places),
            .personDetail(entityId: "e1", ref: "PER-1", title: nil),
        ]
        let known = Set(WorkspacePresentationID.allCases)
        for location in locations {
            guard let place = registry.resolve(location, project: project) else {
                Issue.record("Expected place for \(location)")
                continue
            }
            #expect(known.contains(place.presentation))
        }
        #expect(known.count == 12)
    }
}