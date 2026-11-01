import SwiftUI

/// The Person page's Family section (S13-12, board S13-D3): the person's
/// parents, siblings, partners and children as person tiles joined by lines.
/// It reads its own query (`ConclusionPageSection.queryKey`), keeps the diagram
/// drawn while it reloads, and opens a relative on their Family.
struct PersonFamilySection: View {
    let entityId: String

    @Environment(WorkspaceSession.self) private var session

    private var key: CatalogQueryKey {
        .personFamily(project: session.projectKey, entityId: entityId)
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<CatalogPersonFamily> = session.queryHandle(key) {
                PersonFamilyLoaded(handle: handle, projectDir: session.projectKey.projectDir, orders: session.nameFormat)
            } else {
                PersonFamilyLoading()
            }
        }
        .task {
            let _: QueryHandle<CatalogPersonFamily> = session.query(key)
        }
    }
}

private struct PersonFamilyLoading: View {
    var body: some View {
        ProgressView()
            .frame(maxWidth: .infinity)
            .padding(.top, 96)
    }
}

private struct PersonFamilyLoaded: View {
    @Bindable var handle: QueryHandle<CatalogPersonFamily>
    let projectDir: String
    let orders: NameFormatOrders

    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(\.pvWidthClass) private var widthClass

    var body: some View {
        let presentation = ConclusionDetailPresentation<PersonFamilyContent>(
            handle.read,
            isFetching: handle.isFetching
        ) { PersonFamilyContent(family: $0, locale: .autoupdatingCurrent, orders: orders) }
        switch presentation {
        case .loading:
            PersonFamilyLoading()
        case .failed(let message):
            PVCallout(tone: .danger, title: L10n.Conclusions.familyErrorTitle, message: message)
                .accessibilityIdentifier("persons.detail.family.error")
        case .content(let content, _):
            if content.isEmpty {
                PVEmptyState(
                    glyph: ConclusionPageSection.family.glyph,
                    title: L10n.Conclusions.familyEmptyTitle,
                    message: L10n.string(L10n.Conclusions.familyEmptyMessage)
                )
                .accessibilityIdentifier("persons.detail.family.empty")
            } else {
                PersonFamilyDiagram(
                    content: content,
                    projectDir: projectDir,
                    widthClass: widthClass,
                    open: { navigation.go(to: $0) }
                )
                .accessibilityIdentifier("persons.detail.family")
            }
        }
    }
}

/// The two columns laid out by `PersonFamilyLayout`: lines, captions, then one
/// container per group in reading order, so Tab runs the left column top to
/// bottom, then the right column group by group.
struct PersonFamilyDiagram: View {
    let content: PersonFamilyContent
    let projectDir: String
    let widthClass: PVWidthClass
    let open: (WorkspaceLocation) -> Void

    @State private var width: CGFloat = 0

    var body: some View {
        Group {
            if width > 0 {
                diagram(PersonFamilyLayout(content: content, width: width, widthClass: widthClass))
            } else {
                Color.clear.frame(height: 1)
            }
        }
        .frame(maxWidth: .infinity, alignment: .topLeading)
        .onGeometryChange(for: CGFloat.self) { $0.size.width } action: { width = $0 }
    }

    private func diagram(_ layout: PersonFamilyLayout) -> some View {
        let tiles = Dictionary(uniqueKeysWithValues: content.allTiles.map { ($0.id, $0) })
        return ZStack(alignment: .topLeading) {
            Path { path in
                for segment in layout.segments { path.addRect(segment) }
            }
            .fill(PVColor.borderStrong)
            .accessibilityHidden(true)
            ForEach(Array(layout.captions.enumerated()), id: \.offset) { _, caption in
                Text(verbatim: caption.text)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textMuted)
                    .fixedSize()
                    .offset(x: caption.origin.x, y: caption.origin.y)
                    .accessibilityHidden(true)
            }
            ForEach(Array(layout.groups.enumerated()), id: \.offset) { _, group in
                ZStack(alignment: .topLeading) {
                    ForEach(group.tiles, id: \.id) { placed in
                        if let tile = tiles[placed.id] {
                            PersonTile(content: tile.content, projectDir: projectDir) {
                                if let location = tile.location { open(location) }
                            }
                            .frame(width: placed.frame.width, height: placed.frame.height)
                            .offset(x: placed.frame.minX, y: placed.frame.minY)
                            .accessibilityIdentifier("persons.detail.family.tile.\(placed.id)")
                        }
                    }
                }
                .frame(width: group.frame.width, height: group.frame.height, alignment: .topLeading)
                .offset(x: group.frame.minX, y: group.frame.minY)
                .accessibilityElement(children: .contain)
                .accessibilityLabel(group.label)
            }
        }
        .frame(width: width, height: layout.height, alignment: .topLeading)
    }
}

#if DEBUG
private enum FamilyPreviewData {
    static func person(_ id: String, _ name: String, born: Int32? = nil) -> CatalogPersonSummary {
        var h = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: id, ref: "PER-\(id.uppercased())", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: name),
            nameValueCount: 1
        )
        if let born { h.birth = CatalogLifeFacts(date: CatalogDateValueInput(kind: "point", startYear: born), dateCount: 1) }
        return h
    }

    static func member(_ id: String, _ name: String, born: Int32? = nil) -> CatalogFamilyMember {
        CatalogFamilyMember(person: person(id, name, born: born), link: nil)
    }

    static func unit(_ kind: CatalogFamilyUnit.Kind, _ partners: [CatalogFamilyMember], _ children: [CatalogFamilyMember]) -> CatalogFamilyUnit {
        CatalogFamilyUnit(kind: kind, partners: partners, children: children)
    }

    static let johann = member("johann", "Johann Weber", born: 1817)
    static let margarethe = member("margarethe", "Margarethe Weber", born: 1815)
    static let friedrich = member("friedrich", "Friedrich Weber", born: 1820)
    static let wilhelm = member("wilhelm", "Wilhelm Weber")
    static let christian = member("christian", "Christian Weber", born: 1823)
    static let heinrich = member("heinrich", "Heinrich Weber")

    static func origin(_ m: CatalogFamilyMember, _ sets: [Int]) -> CatalogOriginSibling {
        CatalogOriginSibling(member: m, parentSets: sets)
    }

    /// Frame 11a.
    static var full: CatalogPersonFamily {
        CatalogPersonFamily(
            units: [
                unit(.parents, [heinrich, member("anna", "Anna Weber")], [margarethe, friedrich, wilhelm]),
                unit(.partner, [member("maria", "Maria Schulz")], [member("karl", "Karl Weber"), member("luise", "Luise Weber"), member("emma", "Emma Weber")]),
                unit(.partner, [member("sophie", "Sophie Brandt")], [member("otto", "Otto Weber")]),
                unit(.oneParent, [], [member("paul", "Paul Weber")]),
            ],
            person: johann.person,
            originSiblings: [origin(margarethe, [0]), origin(johann, [0]), origin(friedrich, [0]), origin(wilhelm, [0])]
        )
    }

    /// Frame 11c.
    static var conflicting: CatalogPersonFamily {
        CatalogPersonFamily(
            units: [
                unit(.parents, [heinrich, member("anna", "Anna Weber")], [margarethe, friedrich, wilhelm]),
                unit(.parents, [heinrich, member("katharina", "Katharina Weber")], [friedrich, christian]),
            ],
            person: johann.person,
            originSiblings: [
                origin(johann, [0, 1]), origin(friedrich, [0, 1]), origin(christian, [1]),
                origin(margarethe, [0]), origin(wilhelm, [0]),
            ]
        )
    }

    /// Frame 11d.
    static var siblingsOnly: CatalogPersonFamily {
        let clara = member("clara", "Clara Weber", born: 1841)
        return CatalogPersonFamily(
            units: [unit(.siblings, [], [member("ida", "Ida Weber", born: 1838), member("bernhard", "Bernhard Weber", born: 1844)])],
            person: clara.person,
            originSiblings: [
                origin(member("ida", "Ida Weber", born: 1838), []), origin(clara, []),
                origin(member("bernhard", "Bernhard Weber", born: 1844), []),
            ]
        )
    }

    /// Frame 11h.
    static var anna: CatalogPersonFamily {
        let anna = member("anna", "Anna Weber", born: 1790)
        return CatalogPersonFamily(
            units: [unit(.partner, [heinrich], [margarethe, johann, friedrich, wilhelm])],
            person: anna.person,
            originSiblings: [origin(anna, [])]
        )
    }

    @MainActor
    static func diagram(_ family: CatalogPersonFamily, width: CGFloat = 1038) -> some View {
        ScrollView {
            PersonFamilyDiagram(
                content: PersonFamilyContent(family: family),
                projectDir: "/tmp/preview.provenencia",
                widthClass: .regular,
                open: { _ in }
            )
        }
        .frame(width: width, height: 900)
        .background(PVColor.surfacePage)
    }
}

#Preview("11a Full family") { FamilyPreviewData.diagram(FamilyPreviewData.full) }
#Preview("11c Conflicting parents") { FamilyPreviewData.diagram(FamilyPreviewData.conflicting) }
#Preview("11d Recorded siblings only") { FamilyPreviewData.diagram(FamilyPreviewData.siblingsOnly) }
#Preview("11h After clicking Anna") { FamilyPreviewData.diagram(FamilyPreviewData.anna) }
#endif
