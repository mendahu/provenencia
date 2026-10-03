import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct WorkspaceSidebarSectionsTests {
    private let projectDir = "/tmp/sidebar.provenencia"

    private func counts(_ store: FakeStore = FakeStore()) -> CatalogCounts {
        CatalogCounts(projectDir: projectDir, store: store)
    }

    @Test func sectionsAreSourceConcludeThenConfigure() {
        let sections = WorkspaceSidebarSections.make(counts: counts())
        #expect(sections.map(\.id) == ["source", "conclude", "configure"])
        #expect(sections.map(\.placement) == [.top, .top, .bottom])
        #expect(sections.map(\.title) == [
            L10n.Workspace.sidebarSourceTitle,
            L10n.Workspace.sidebarConcludeTitle,
            L10n.Workspace.sidebarConfigureTitle,
        ])
        #expect(!sections.map(\.id).contains("narrate"))
    }

    @Test func destinationsAreTopLevelRowsInOrder() {
        let sections = WorkspaceSidebarSections.make(counts: counts())
        #expect(sections.map { $0.items.map(\.id) } == [
            [WorkspaceSection.sources.rawValue],
            [WorkspaceSection.persons, .events, .places].map(\.rawValue),
            [WorkspaceSection.sourceTypes, .metadata, .properties].map(\.rawValue),
        ])
        #expect(sections.flatMap(\.items).allSatisfy { $0.children.isEmpty })
        #expect(sections.flatMap(\.items).allSatisfy {
            $0.accessibilityIdentifier == "workspace.nav.\($0.id)"
        })
    }

    @Test func researchRowsCountAndConfigureRowsDoNot() async throws {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
        ]
        store.subjectsBySource["s1"] = [
            CatalogSubject(id: "sub-1", ref: "CPR-1", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
        ]
        _ = try await store.promoteSubject(projectDir: projectDir, userID: "u", subjectID: "sub-1")
        let catalogCounts = counts(store)
        await catalogCounts.refreshAll()

        let sections = WorkspaceSidebarSections.make(counts: catalogCounts)
        let byID = Dictionary(uniqueKeysWithValues: sections.flatMap(\.items).map { ($0.id, $0.count) })
        #expect(byID[WorkspaceSection.sources.rawValue] == 0)
        #expect(byID[WorkspaceSection.persons.rawValue] == 1)
        // A fresh project's zero is shown, not hidden (SB-4).
        #expect(byID[WorkspaceSection.events.rawValue] == 0)
        #expect(byID[WorkspaceSection.places.rawValue] == 0)
        for section in [WorkspaceSection.sourceTypes, .metadata, .properties] {
            #expect(byID[section.rawValue] == .some(nil))
        }
    }

    @Test func countsAreAbsentUntilLoaded() {
        let sections = WorkspaceSidebarSections.make(counts: counts())
        #expect(sections.flatMap(\.items).allSatisfy { $0.count == nil })
    }
}
