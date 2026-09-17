import Foundation
import Testing
@testable import Provenencia

@Suite
struct WorkspaceToolbarBreadcrumbTests {
    @Test func sectionRootIsSingleNonNavigatingCrumb() {
        let items = WorkspaceToolbar.breadcrumbItems(
            for: .sectionRoot(.sourceTypes),
            goToSectionRoot: { _ in }
        )
        #expect(items.count == 1)
        #expect(items[0].action == nil)
        #expect(items[0].label == String(localized: L10n.Workspace.sourceTypesTitle))
    }

    @Test func deepLocationHasNavigableSectionAndLeaf() {
        var wentTo: WorkspaceSection?
        let items = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                ref: "SRC-AAAA",
                title: "Deed"
            ),
            goToSectionRoot: { wentTo = $0 }
        )
        #expect(items.count == 2)
        #expect(items[0].action != nil)
        #expect(items[1].action == nil)
        #expect(items[1].label == "SRC-AAAA")
        items[0].action?()
        #expect(wentTo == .sources)
    }

    @Test func deepLeafFallsBackToTitleThenEllipsis() {
        let titled = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(section: .sourceFields, fieldId: "f1", title: "Author"),
            goToSectionRoot: { _ in }
        )
        #expect(titled.last?.label == "Author")

        let bare = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(section: .sourceFields, fieldId: "f1"),
            goToSectionRoot: { _ in }
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
            goToSectionRoot: { _ in }
        )
        #expect(items.count == 2)
        #expect(items[1].label == String(localized: L10n.Workspace.evidenceGraphTitle))
    }
}
