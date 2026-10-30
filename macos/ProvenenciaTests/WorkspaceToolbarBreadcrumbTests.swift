import Foundation
import Testing
@testable import Provenencia

@MainActor @Suite
struct WorkspaceToolbarBreadcrumbTests {
    @Test func conclusionDetailCrumbPrefersTheTitle() {
        var wentTo: WorkspaceLocation?
        let items = WorkspaceToolbar.breadcrumbItems(
            for: .eventDetail(entityId: "evt-1", ref: "EVT-1", title: "Fire at York"),
            goTo: { wentTo = $0 }
        )
        #expect(items.count == 2)
        #expect(items[0].label == L10n.string(L10n.Workspace.eventsTitle))
        #expect(items[1].label == "Fire at York")
        items[0].action?()
        #expect(wentTo == .sectionRoot(.events))
    }

    @Test func sectionRootIsSingleNonNavigatingCrumb() {
        let items = WorkspaceToolbar.breadcrumbItems(
            for: .sectionRoot(.sourceTypes),
            goTo: { _ in }
        )
        #expect(items.count == 1)
        #expect(items[0].action == nil)
        #expect(items[0].label == L10n.string(L10n.Workspace.sourceTypesTitle))
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
            for: WorkspaceLocation(section: .metadata, fieldId: "f1", title: "Author"),
            goTo: { _ in }
        )
        #expect(titled.last?.label == "Author")

        let bare = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(section: .metadata, fieldId: "f1"),
            goTo: { _ in }
        )
        #expect(bare.last?.label == "…")
    }

    @Test func propertiesCategoryAndPropertyAreDeep() {
        var wentTo: WorkspaceLocation?
        let category = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(section: .properties, subjectTypeKey: "person", title: "Person"),
            goTo: { wentTo = $0 }
        )
        #expect(category.count == 2)
        #expect(category.last?.label == "Person")
        category[0].action?()
        #expect(wentTo == .sectionRoot(.properties))

        let property = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .properties,
                subjectTypeKey: "person",
                propertyId: "prop-1",
                title: "Name"
            ),
            goTo: { _ in }
        )
        #expect(property.count == 2)
        #expect(property.last?.label == "Name")
    }

    private let names = WorkspaceToolbar.BreadcrumbNames(
        sourceTitle: { $0 == "src-1" ? "1851 census" : nil },
        graphName: { $0 == "g-1" ? "Jones household" : nil }
    )

    @Test func evidenceGraphCrumbsNameSourceAndGraph() {
        var wentTo: WorkspaceLocation?
        let items = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                graphId: "g-1",
                sourceSurface: .graph,
                ref: "SRC-AAAA",
                title: "Deed"
            ),
            names: names,
            goTo: { wentTo = $0 }
        )
        // The names come from the lists, not from the title the location carries.
        #expect(items.map(\.label) == [L10n.string(L10n.Workspace.sourcesTitle), "1851 census", "Jones household"])
        #expect(items[1].action != nil)
        #expect(items[2].action == nil)
        items[1].action?()
        #expect(wentTo == .sourcePage(sourceId: "src-1"))
        #expect(wentTo?.sourceSurface == .page)
        #expect(wentTo?.title == "1851 census")
    }

    @Test func citationComposerCrumbsNameSourceAndGraph() {
        var wentTo: WorkspaceLocation?
        let items = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                graphId: "g-1",
                subjectId: "sub-1",
                sourceSurface: .citationComposer,
                title: "John Jones",
                sourceTitle: "Old title"
            ),
            names: names,
            goTo: { wentTo = $0 }
        )
        #expect(items.count == 4)
        #expect(items[0].label == L10n.string(L10n.Workspace.sourcesTitle))
        #expect(items[1].label == "1851 census")
        #expect(items[2].label == "Jones household")
        #expect(items[2].action != nil)
        #expect(items[3].label == L10n.CitationComposer.breadcrumbCitationFor(scope: "John Jones"))
        #expect(items[3].action == nil)
        items[2].action?()
        #expect(wentTo?.sourceId == "src-1")
        #expect(wentTo?.graphId == "g-1")
        #expect(wentTo?.sourceSurface == .graph)
        #expect(wentTo?.title == "1851 census")
    }

    @Test func promoteCrumbsNameSourceAndGraph() {
        var wentTo: WorkspaceLocation?
        let items = WorkspaceToolbar.breadcrumbItems(
            for: .promote(
                sourceId: "src-1", graphId: "g-1", subjectId: "sub-1", kind: .person,
                ref: "SUB-AAAAA", title: "James Robins", sourceTitle: "1851 census"
            ),
            names: names,
            goTo: { wentTo = $0 }
        )
        #expect(items.count == 4)
        #expect(items[1].label == "1851 census")
        #expect(items[2].label == "Jones household")
        #expect(items[3].label == L10n.Promote.breadcrumb(ref: "SUB-AAAAA"))
        #expect(items[3].action == nil)
        items[1].action?()
        #expect(wentTo?.sourceSurface == .page)
        #expect(wentTo?.sourceId == "src-1")
        items[2].action?()
        #expect(wentTo?.sourceSurface == .graph)
        #expect(wentTo?.sourceId == "src-1")
        #expect(wentTo?.graphId == "g-1")
    }

    @Test func graphCrumbsFallBackWithoutNames() {
        let canvas = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .sources, sourceId: "src-1", graphId: "g-1",
                sourceSurface: .graph, ref: "SRC-AAAA", title: "1851 census"
            ),
            goTo: { _ in }
        )
        #expect(canvas.map(\.label) == [
            L10n.string(L10n.Workspace.sourcesTitle),
            "1851 census",
            L10n.string(L10n.Workspace.evidenceGraphTitle),
        ])

        let composer = WorkspaceToolbar.breadcrumbItems(
            for: WorkspaceLocation(
                section: .sources, sourceId: "src-1", graphId: "g-1", subjectId: "sub-1",
                sourceSurface: .citationComposer, title: "John Jones"
            ),
            goTo: { _ in }
        )
        #expect(composer.map(\.label) == [
            L10n.string(L10n.Workspace.sourcesTitle),
            L10n.string(L10n.Workspace.evidenceGraphTitle),
            L10n.CitationComposer.breadcrumbCitationFor(scope: "John Jones"),
        ])
    }

    @Test func pageSectionAddsACrumb() {
        var wentTo: WorkspaceLocation?
        let gallery = WorkspaceLocation.personDetail(
            entityId: "per-1", ref: "PER-1", title: "Jane Doe", pageSection: .gallery
        )
        let items = WorkspaceToolbar.breadcrumbItems(for: gallery, goTo: { wentTo = $0 })
        #expect(items.map(\.label) == [
            L10n.string(L10n.Workspace.personsTitle), "Jane Doe", L10n.string(L10n.Conclusions.pageSectionGallery),
        ])
        #expect(items[2].action == nil)
        items[1].action?()
        #expect(wentTo == .personDetail(entityId: "per-1", ref: "PER-1", title: "Jane Doe"))
        #expect(wentTo?.pageSection == nil)

        let vitals = WorkspaceToolbar.breadcrumbItems(
            for: .personDetail(entityId: "per-1", ref: "PER-1", title: "Jane Doe"),
            goTo: { _ in }
        )
        #expect(vitals.map(\.label) == [L10n.string(L10n.Workspace.personsTitle), "Jane Doe"])
        #expect(vitals[1].action == nil)
    }

    @Test func locationLabelsNameTheSection() {
        let family = WorkspaceToolbar.locationLabels(
            for: .personDetail(entityId: "per-1", ref: "PER-1", title: "Jane Doe", pageSection: .family)
        )
        #expect(family.leaf == "Jane Doe")
        #expect(family.section == L10n.string(L10n.Conclusions.pageSectionFamily))

        let vitals = WorkspaceToolbar.locationLabels(
            for: .personDetail(entityId: "per-1", ref: "PER-1", title: "Jane Doe")
        )
        #expect(vitals.leaf == "Jane Doe")
        #expect(vitals.section == nil)

        let bare = WorkspaceToolbar.locationLabels(
            for: .personDetail(entityId: "per-1", ref: "PER-1", title: nil)
        )
        #expect(bare.leaf == "PER-1")

        let source = WorkspaceToolbar.locationLabels(
            for: .sourcePage(sourceId: "src-1", ref: "SRC-AAAA", title: "Deed")
        )
        #expect(source.leaf == "Deed")
        #expect(source.section == nil)

        let root = WorkspaceToolbar.locationLabels(for: .sectionRoot(.metadata))
        #expect(root.leaf == nil)
        #expect(root.section == nil)
    }
}
