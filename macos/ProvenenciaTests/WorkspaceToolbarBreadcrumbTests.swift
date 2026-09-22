import Foundation
import Testing
@testable import Provenencia

@Suite
struct WorkspaceToolbarBreadcrumbTests {
    @Test func sectionRootIsSingleNonNavigatingCrumb() {
        let items = WorkspaceToolbar.breadcrumbItems(
            for: .sectionRoot(.sourceTypes),
            goTo: { _ in }
        )
        #expect(items.count == 1)
        #expect(items[0].action == nil)
        #expect(items[0].label == String(localized: L10n.Workspace.sourceTypesTitle))
    }

    @Test func deepLocationHasNavigableSectionAndLeaf() {
        var wentTo: WorkspaceLocation?
        let items = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                ref: "SRC-AAAA",
                title: "Deed"
            ),
            goTo: { wentTo = $0 }
        )
        #expect(items.count == 2)
        #expect(items[0].action != nil)
        #expect(items[1].action == nil)
        #expect(items[1].label == "SRC-AAAA")
        items[0].action?()
        #expect(wentTo == .sectionRoot(.sources))
    }

    @Test func deepLeafFallsBackToTitleThenEllipsis() {
        let titled = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(section: .sourceFields, fieldId: "f1", title: "Author"),
            goTo: { _ in }
        )
        #expect(titled.last?.label == "Author")

        let bare = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(section: .sourceFields, fieldId: "f1"),
            goTo: { _ in }
        )
        #expect(bare.last?.label == "…")
    }

    @Test func evidenceGraphLeafUsesGraphTitle() {
        let items = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                sourceSurface: .graph,
                ref: "SRC-AAAA",
                title: "Deed"
            ),
            goTo: { _ in }
        )
        #expect(items.count == 2)
        #expect(items[1].label == String(localized: L10n.Workspace.evidenceGraphTitle))
    }

    @Test func citationComposerIncludesNavigableEvidenceGraph() {
        var wentTo: WorkspaceLocation?
        let items = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                subjectId: "sub-1",
                sourceSurface: .citationComposer,
                title: "Margt. Alderwick",
                sourceTitle: "Alderwick family bible"
            ),
            goTo: { wentTo = $0 }
        )
        #expect(items.count == 3)
        #expect(items[0].label == String(localized: L10n.Workspace.sourcesTitle))
        #expect(
            items[1].label
                == L10n.Workspace.evidenceGraphFor(sourceTitle: "Alderwick family bible")
        )
        #expect(items[1].action != nil)
        #expect(
            items[2].label
                == L10n.CitationComposer.breadcrumbCitationFor(scope: "Margt. Alderwick")
        )
        #expect(items[2].action == nil)
        items[1].action?()
        #expect(wentTo?.sourceId == "src-1")
        #expect(wentTo?.sourceSurface == .graph)
        #expect(wentTo?.title == "Alderwick family bible")
    }
}
