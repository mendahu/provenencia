import SwiftUI

/// What a Conclusion page shows for its handle's state. A stale page stays
/// readable while it refreshes. Person, Event, and Place each supply
/// their own content; the page does not know which kind it is.
enum ConclusionDetailPresentation<Content: Equatable>: Equatable {
    case loading
    case failed(String)
    case content(Content, refreshing: Bool)

    init(
        _ read: QueryRead<CatalogConclusionDetail>,
        isFetching: Bool,
        content: (CatalogConclusionDetail) -> Content
    ) {
        switch read {
        case .loaded(let value):
            self = .content(content(value), refreshing: isFetching)
        case .failed(let error):
            self = .failed(L10n.Errors.message(for: error))
        case .loading:
            self = .loading
        }
    }
}

/// One related-place row inside a Place detail section (S9-40).
struct ConclusionDetailRelationship: Equatable, Identifiable {
    var relationship: CatalogPlaceRelationship
    var role: String? = nil

    var id: String {
        if let role, !role.isEmpty {
            return "\(relationship.id)-\(role)"
        }
        return relationship.id
    }
}

/// One Place hierarchy section (Part of / Contains / Succession). Carries relationship
/// rows when Go filled them; otherwise the empty sentence.
struct ConclusionDetailSection: Equatable, Identifiable {
    var id: String
    var title: LocalizedStringResource
    var aside: String
    var emptyText: String
    var relationships: [ConclusionDetailRelationship] = []
    /// Contains remainder when the list is truncated (PLD-5).
    var omittedCount: Int = 0
}

/// What the header thumb takes from a kind. One slot, so a later PR adds to
/// this struct instead of adding another protocol member (S10-23, S10-35).
struct ConclusionDetailThumbnail: Equatable {
    /// The kind's mark and its spoken label, shown when there is no image.
    var mark: PVMarkKey
    var label: LocalizedStringResource
    /// The `objects/…` path of the image to show; "" for none.
    var imageRelPath: String = ""
    /// Spoken label while the image shows.
    var imageLabel: String? = nil
    var colors: PVThumbnail.Colors? = nil

    /// The header thumb of a kind that leads with its gallery's lead picture. The gallery is the
    /// one source for the lead: the engine sets the summary path and every picture path from one
    /// lead rule, so the page reads no summary path beside it (S14-02 D-5).
    static func leading(
        gallery: ConclusionDetailGallery,
        noImage: LocalizedStringResource,
        withImage: LocalizedStringResource,
        leadLabel: (String) -> String,
        colors: PVThumbnail.Colors? = nil
    ) -> ConclusionDetailThumbnail {
        let lead = gallery.lead
        return ConclusionDetailThumbnail(
            mark: gallery.fallbackMark,
            label: lead == nil ? noImage : withImage,
            imageRelPath: lead?.imageRelPath ?? "",
            imageLabel: lead.map { leadLabel($0.sourceTitle) },
            colors: colors
        )
    }
}

/// One picture tile in a page's gallery: the image, its Source's title, the composer location its
/// Citation opens at, and the Strings the section and the lead model speak for it (S10-35, S14-02).
struct ConclusionDetailPicture: Equatable, Identifiable {
    var id: String
    var imageRelPath: String
    var sourceTitle: String
    var location: WorkspaceLocation
    /// The tile the header leads with wears the lead Badge; exactly one tile.
    var isLead: Bool
    /// Spoken label of the tile.
    var accessibilityLabel: String
    /// Spoken label of the button that makes this tile the lead.
    var useAsLabel: String
    /// Toast body after this tile is chosen as the lead.
    var chosenToastBody: String
}

/// A page's gallery, one slot on the body for every kind: the section's words, the mark a missing
/// image falls back to, the tiles (earliest recorded first), and the copy of choosing a lead
/// (S14-02 D-1, D-2).
struct ConclusionDetailGallery: Equatable {
    var title: LocalizedStringResource
    var openTip: LocalizedStringResource
    var fallbackMark: PVMarkKey
    var tiles: [ConclusionDetailPicture]
    var choose: LeadCopy

    /// The tile the header leads with.
    var lead: ConclusionDetailPicture? { tiles.first(where: \.isLead) }
    /// The header leads with a picture other than the earliest, so its menu offers the earliest
    /// (S10-39, user 2026-10-17).
    var offersEarliest: Bool { tiles.first.map { !$0.isLead } ?? false }

    /// The static copy of choosing a lead; the per-tile Strings live on the tile.
    struct LeadCopy: Equatable {
        var badge: LocalizedStringResource
        var useAs: LocalizedStringResource
        var useEarliest: LocalizedStringResource
        var menuTitle: LocalizedStringResource
        var menuHint: LocalizedStringResource
        var toastTitle: String
        var toastEarliestBody: String

        static let personPortrait = LeadCopy(
            badge: L10n.Conclusions.portraitBadge,
            useAs: L10n.Conclusions.useAsPortrait,
            useEarliest: L10n.Conclusions.useEarliestLikeness,
            menuTitle: L10n.Conclusions.portraitMenuTitle,
            menuHint: L10n.Conclusions.portraitMenuHint,
            toastTitle: L10n.Conclusions.toastPortraitUpdatedTitle,
            toastEarliestBody: L10n.Conclusions.toastPortraitEarliestBody
        )
        static let eventThumbnail = LeadCopy(
            badge: L10n.Conclusions.thumbnailBadge,
            useAs: L10n.Conclusions.useAsThumbnail,
            useEarliest: L10n.Conclusions.useEarliestPicture,
            menuTitle: L10n.Conclusions.eventThumbnailMenuTitle,
            menuHint: L10n.Conclusions.thumbnailMenuHint,
            toastTitle: L10n.Conclusions.toastThumbnailUpdatedTitle,
            toastEarliestBody: L10n.Conclusions.toastThumbnailEarliestBody
        )
        /// A Place's thumbnail copy is an Event's with its own menu title.
        static let placeThumbnail: LeadCopy = {
            var copy = eventThumbnail
            copy.menuTitle = L10n.Conclusions.placeThumbnailMenuTitle
            return copy
        }()
    }

    /// The gallery of an Event or a Place: the shared section words and tile Strings, with the
    /// kind's fallback mark and lead copy.
    static func pictures(
        from detail: CatalogConclusionDetail,
        fallbackMark: PVMarkKey,
        choose: LeadCopy
    ) -> ConclusionDetailGallery {
        ConclusionDetailGallery(
            title: L10n.Conclusions.picturesSection,
            openTip: L10n.Conclusions.pictureOpenTip,
            fallbackMark: fallbackMark,
            tiles: tiles(
                from: detail,
                tileLabel: L10n.Conclusions.pictureTileLabel,
                useAsLabel: L10n.Conclusions.useAsThumbnailLabel,
                chosenBody: L10n.Conclusions.toastThumbnailChosenBody
            ),
            choose: choose
        )
    }

    /// The tiles of `detail.pictures`, in wire order. Each function takes the picture's Source title
    /// (the Observation ref when the title is empty) and returns the kind's wording.
    static func tiles(
        from detail: CatalogConclusionDetail,
        tileLabel: (String) -> String,
        useAsLabel: (String) -> String,
        chosenBody: (String) -> String
    ) -> [ConclusionDetailPicture] {
        detail.pictures.map {
            let source = ReconciliationRecord.sourceTitle(for: $0.record)
            return ConclusionDetailPicture(
                id: $0.record.observationID,
                imageRelPath: $0.imageRelPath,
                sourceTitle: source,
                location: ReconciliationRecord.composerLocation(for: $0.record),
                isLead: $0.leads,
                accessibilityLabel: tileLabel(source),
                useAsLabel: useAsLabel(source),
                chosenToastBody: chosenBody(source)
            )
        }
    }
}

/// The Badge a kind may draw above the title: data, so Event and Place stay
/// free of view generics.
struct ConclusionDetailBadge: Equatable {
    var text: String
    var ink: Color
    var accessibilityLabel: String
}

/// The fields a Conclusion detail page lays out. Each kind owns how it
/// fills them (mark, title, the line under the title, and the rows). The
/// page places the header and then the kind-supplied `pageSections` slot,
/// so a kind does not get its own page type. `sections` here is the Place
/// hierarchy data, not the page's sections.
protocol ConclusionDetailBody: Equatable {
    var title: ConclusionTitleSource { get }
    var ref: String { get }
    var showsRef: Bool { get }
    var members: String { get }
    var rows: [ReconciledValueRowModel] { get }
    var sections: [ConclusionDetailSection] { get }
    var thumbnail: ConclusionDetailThumbnail { get }
    var badge: ConclusionDetailBadge? { get }
    var gallery: ConclusionDetailGallery { get }
}

extension ConclusionDetailBody {
    var sections: [ConclusionDetailSection] { [] }
    var badge: ConclusionDetailBadge? { nil }
}

/// What the `pageSections` slot receives beside the content.
struct ConclusionDetailPageSectionsContext {
    var projectDir: String
    var refreshing: Bool
    var pageIdentifier: String
    var lead: ConclusionLeadModel
    /// The section of the page that is open; nil for a kind that offers none.
    var pageSection: ConclusionPageSection? = nil
}

private struct ConclusionDetailHeaderCompactKey: EnvironmentKey {
    static let defaultValue = false
}

extension EnvironmentValues {
    /// True while a pinned page header is compact (the page scrolled). The header sets it
    /// for the kind's `secondary`, which may lay itself out more tightly.
    var conclusionDetailHeaderCompact: Bool {
        get { self[ConclusionDetailHeaderCompactKey.self] }
        set { self[ConclusionDetailHeaderCompactKey.self] = newValue }
    }
}

/// Shared Conclusion detail chrome (S9-24). A view, not a place: the caller
/// names the entity, and the place registry warms `conclusionDetail` for
/// that id. Person, Event, and Place stay separate history entries.
struct ConclusionDetailPage<Content: ConclusionDetailBody, Secondary: View, PageSections: View>: View {
    let session: WorkspaceSession
    let entityId: String
    let pageIdentifier: String
    let errorIdentifier: String
    let make: (CatalogConclusionDetail) -> Content
    let secondary: (Content) -> Secondary
    /// Everything below the header. The convenience initializer supplies `ConclusionDetailDefaultPageSections`.
    let pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections
    /// Chooses the lead of any kind's gallery. The page owns it, so the page's toast is its own and the
    /// model resets with the page, which is keyed by entity id (S14-02 D-7).
    @State private var lead: ConclusionLeadModel

    init(
        session: WorkspaceSession,
        entityId: String,
        pageIdentifier: String,
        errorIdentifier: String,
        userID: String,
        make: @escaping (CatalogConclusionDetail) -> Content,
        secondary: @escaping (Content) -> Secondary,
        pageSections: @escaping (Content, ConclusionDetailPageSectionsContext) -> PageSections
    ) {
        self.session = session
        self.entityId = entityId
        self.pageIdentifier = pageIdentifier
        self.errorIdentifier = errorIdentifier
        self.make = make
        self.secondary = secondary
        self.pageSections = pageSections
        _lead = State(initialValue: ConclusionLeadModel(session: session, userID: userID, entityId: entityId))
    }

    private var key: CatalogQueryKey {
        .conclusionDetail(project: session.projectKey, entityId: entityId)
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<CatalogConclusionDetail> = session.queryHandle(key) {
                ConclusionDetailLoaded(
                    handle: handle,
                    projectDir: session.projectKey.projectDir,
                    pageIdentifier: pageIdentifier,
                    errorIdentifier: errorIdentifier,
                    make: make,
                    secondary: secondary,
                    lead: lead,
                    pageSections: pageSections
                )
            } else {
                Self.loadingPage
            }
        }
        .background(PVColor.surfacePage)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier(pageIdentifier)
        .task {
            let _: QueryHandle<CatalogConclusionDetail> = session.query(key)
        }
        .vocabularyToastOverlay($lead.toast, identifier: "\(pageIdentifier).toast")
    }

    /// The header-less page shown while the detail loads, before and after its handle exists.
    static var loadingPage: some View {
        PVPageColumns(primary: { loading })
    }

    static var loading: some View {
        ProgressView()
            .controlSize(.small)
            .frame(maxWidth: .infinity, alignment: .leading)
    }
}

extension ConclusionDetailPage where PageSections == ConclusionDetailDefaultPageSections<Content> {
    init(
        session: WorkspaceSession,
        entityId: String,
        pageIdentifier: String,
        errorIdentifier: String,
        userID: String,
        make: @escaping (CatalogConclusionDetail) -> Content,
        secondary: @escaping (Content) -> Secondary
    ) {
        self.init(
            session: session,
            entityId: entityId,
            pageIdentifier: pageIdentifier,
            errorIdentifier: errorIdentifier,
            userID: userID,
            make: make,
            secondary: secondary,
            pageSections: { ConclusionDetailDefaultPageSections(content: $0, context: $1) }
        )
    }
}

/// Today's below-header fill: Details, the gallery, then the Place relationship sections.
/// The body returns sibling views that take their spacing from the page's `VStack`, so a custom
/// fill returns siblings too; a modifier on the whole fill would join them into one block.
struct ConclusionDetailDefaultPageSections<Content: ConclusionDetailBody>: View {
    let content: Content
    let context: ConclusionDetailPageSectionsContext

    var body: some View {
        ConclusionDetailRowsSection(
            rows: content.rows,
            refreshing: context.refreshing,
            title: L10n.Conclusions.detailsSection
        )
        if !content.gallery.tiles.isEmpty {
            ConclusionDetailGallerySection(
                projectDir: context.projectDir,
                pageIdentifier: context.pageIdentifier,
                gallery: content.gallery,
                lead: context.lead
            )
        }
        ForEach(content.sections) { section in
            ConclusionDetailRelationshipSection(section: section)
        }
    }
}

/// A page's field rows, under a heading with the refreshing spinner when `title` is given.
/// A page that draws neither (the Person page's Vitals, which shows refreshing in its
/// section switcher) passes nil.
struct ConclusionDetailRowsSection: View {
    let rows: [ReconciledValueRowModel]
    let refreshing: Bool
    var title: LocalizedStringResource? = nil
    var provenance: ReconciledValueRow.Provenance = .why

    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(\.pvWidthClass) private var widthClass
    /// The row whose provenance sheet is open (`.sheet` mode); one presenter for all rows.
    @State private var provenanceItem: ProvenanceSheetItem?

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if let title {
                PVSectionHeader(title: title, aside: {
                    if refreshing {
                        ProgressView().controlSize(.small)
                    }
                })
            }
            ForEach(rows) { row in
                ReconciledValueRow(
                    model: row, provenance: provenance,
                    openProvenance: { provenanceItem = ProvenanceSheetItem(id: row.id) }
                )
            }
        }
        .pvSheet(item: $provenanceItem) { item in
            if let row = item.row(in: rows) {
                ProvenanceSheet(row: row, widthClass: widthClass, close: { provenanceItem = nil })
                    .environment(navigation)
            }
        }
        // A row that leaves the page closes its sheet.
        .onChange(of: rows) {
            if let item = provenanceItem, item.row(in: rows) == nil {
                provenanceItem = nil
            }
        }
    }
}

private struct ConclusionDetailLoaded<Content: ConclusionDetailBody, Secondary: View, PageSections: View>: View {
    @Bindable var handle: QueryHandle<CatalogConclusionDetail>
    let projectDir: String
    let pageIdentifier: String
    let errorIdentifier: String
    let make: (CatalogConclusionDetail) -> Content
    let secondary: (Content) -> Secondary
    let lead: ConclusionLeadModel
    let pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections

    @Environment(WorkspaceNavigation.self) private var navigation

    private var presentation: ConclusionDetailPresentation<Content> {
        ConclusionDetailPresentation(handle.read, isFetching: handle.isFetching, content: make)
    }

    /// The page sections this kind of page offers; none for Event and Place.
    private var offeredPageSections: [ConclusionPageSection] {
        ConclusionPageSection.offered(by: navigation.currentLocation.section)
    }

    /// The open page section: the location's, else the kind's first.
    private var pageSection: ConclusionPageSection? {
        offeredPageSections.isEmpty ? nil : (navigation.currentLocation.pageSection ?? offeredPageSections[0])
    }

    var body: some View {
        Group {
            switch presentation {
            case .loading:
                ConclusionDetailPage<Content, Secondary, PageSections>.loadingPage
            case .failed(let message):
                PVPageColumns(primary: {
                    PVCallout(tone: .danger, message: message)
                        .accessibilityIdentifier(errorIdentifier)
                })
            case .content(let content, let refreshing):
                PVPageColumns(
                    headerPinned: !offeredPageSections.isEmpty,
                    scrollResetKey: pageSection,
                    header: { isScrolled in
                        ConclusionDetailHeaderBand(
                            content: content,
                            projectDir: projectDir,
                            secondary: secondary,
                            lead: lead,
                            pageIdentifier: pageIdentifier,
                            offeredPageSections: offeredPageSections,
                            pageSection: pageSection,
                            refreshing: refreshing,
                            isScrolled: isScrolled,
                            select: { navigation.go(to: navigation.currentLocation.withPageSection($0)) }
                        )
                    },
                    primary: {
                        VStack(alignment: .leading, spacing: PVSpacing.space9) {
                            pageSections(
                                content,
                                ConclusionDetailPageSectionsContext(
                                    projectDir: projectDir,
                                    refreshing: refreshing,
                                    pageIdentifier: pageIdentifier,
                                    lead: lead,
                                    pageSection: pageSection
                                )
                            )
                        }
                    }
                )
            }
        }
    }
}

/// The header and, on a page with sections, the section switcher under it. Pinned pages
/// supply the band's own padding, full or compact; an unpinned page has none.
private struct ConclusionDetailHeaderBand<Content: ConclusionDetailBody, Secondary: View>: View {
    let content: Content
    let projectDir: String
    let secondary: (Content) -> Secondary
    let lead: ConclusionLeadModel
    let pageIdentifier: String
    let offeredPageSections: [ConclusionPageSection]
    let pageSection: ConclusionPageSection?
    let refreshing: Bool
    let isScrolled: Bool
    let select: (ConclusionPageSection) -> Void

    /// A page with page sections pins its header; one without scrolls it away.
    private var pinned: Bool { pageSection != nil }

    private var bar: CGFloat { isScrolled ? PVSpacing.space5 : PVSpacing.space7 }

    var body: some View {
        VStack(alignment: .leading, spacing: bar) {
            ConclusionDetailHeader(
                content: content,
                projectDir: projectDir,
                secondary: secondary,
                lead: lead,
                pageIdentifier: pageIdentifier,
                compact: isScrolled
            )
            if let pageSection {
                switcher(selection: pageSection)
            }
        }
        .padding(.top, pinned ? (isScrolled ? PVSpacing.space5 : PVSpacing.space9) : 0)
        .padding(.bottom, pinned ? bar : 0)
    }

    private func switcher(selection: ConclusionPageSection) -> some View {
        HStack(spacing: PVSpacing.space6) {
            PVSegmentedControl(
                L10n.Conclusions.pageSectionsLabel,
                items: offeredPageSections.enumerated().map { index, item in
                    .init(
                        item.label,
                        value: item,
                        glyph: item.glyph,
                        shortcut: KeyEquivalent(Character(String(index + 1))),
                        accessibilityIdentifier: "\(pageIdentifier).section.\(item.rawValue)"
                    )
                },
                selection: selection,
                size: .lg,
                select: select
            )
            if refreshing {
                HStack(spacing: PVSpacing.space3) {
                    ProgressView().controlSize(.small)
                    Text(L10n.Conclusions.detailRefreshing)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                .fixedSize()
            }
        }
    }
}

/// A page's gallery: a header with the count, then one 96pt crop per picture in a wrapping grid,
/// each captioned with its Source. A tile opens the Citation that holds it and offers the lead
/// choice (S10-35, S10-D12 PDP-7, PDP-8; S14-02). The section is kind-blind: its words come from
/// `gallery`, its ids from the page's identifier.
struct ConclusionDetailGallerySection: View {
    let projectDir: String
    let pageIdentifier: String
    let gallery: ConclusionDetailGallery
    let lead: ConclusionLeadModel

    @Environment(WorkspaceNavigation.self) private var navigation
    @FocusState private var focusedID: String?

    private static let column: CGFloat = 120

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: gallery.title, meta: String(gallery.tiles.count))
            LazyVGrid(
                columns: [GridItem(
                    .adaptive(minimum: Self.column, maximum: Self.column),
                    spacing: PVSpacing.space6,
                    alignment: .top
                )],
                alignment: .leading,
                spacing: PVSpacing.space7
            ) {
                ForEach(gallery.tiles) { tile in
                    tileView(tile)
                }
            }
            .padding(.top, PVSpacing.space6)
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.string(gallery.title))
    }

    private func tileView(_ tile: ConclusionDetailPicture) -> some View {
        let tileID = "\(pageIdentifier).picture.\(tile.id)"
        return VStack(alignment: .leading, spacing: PVSpacing.space3) {
            GalleryTileView(
                projectDir: projectDir,
                relPath: tile.imageRelPath,
                fallbackMark: gallery.fallbackMark,
                caption: tile.sourceTitle,
                captionWidth: Self.column
            ) {
                Button {
                    navigation.go(to: tile.location)
                } label: {
                    Color.clear
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                        .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .focused($focusedID, equals: tile.id)
                .pvFocusRing(focusedID == tile.id, cornerRadius: PVRadius.sm)
                .help(Text(gallery.openTip))
                .accessibilityLabel(tile.accessibilityLabel)
                .accessibilityIdentifier(tileID)
            }
            if tile.isLead {
                PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)
                    .accessibilityIdentifier("\(tileID).leadBadge")
            } else {
                PVButton(gallery.choose.useAs, variant: .ghost, size: .sm) {
                    Task {
                        if await lead.choose(tile, copy: gallery.choose) { focusedID = tile.id }
                    }
                }
                .disabled(lead.isSaving)
                .fixedSize(horizontal: true, vertical: false)
                .accessibilityLabel(tile.useAsLabel)
                .accessibilityIdentifier("\(tileID).useAsLead")
            }
        }
        .frame(width: Self.column, alignment: .leading)
    }
}

/// A relationship section: header, rows when present, otherwise the empty sentence.
private struct ConclusionDetailRelationshipSection: View {
    let section: ConclusionDetailSection

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: section.title, aside: {
                Text(verbatim: section.aside)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(2)
            })
            if section.relationships.isEmpty {
                Text(verbatim: section.emptyText)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.vertical, PVSpacing.space5)
            } else {
                ForEach(section.relationships) { row in
                    PlaceRelationshipRow(relationship: row.relationship, role: row.role)
                }
                if section.omittedCount > 0 {
                    Text(verbatim: L10n.Conclusions.placeContainsMore(section.omittedCount))
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                        .padding(.vertical, PVSpacing.space3)
                }
            }
        }
        .accessibilityElement(children: section.relationships.isEmpty ? .combine : .contain)
        .accessibilityLabel(
            section.relationships.isEmpty
                ? L10n.DesignSystem.spokenJoin(L10n.string(section.title), section.emptyText)
                : L10n.string(section.title)
        )
    }
}

/// Thumbnail slot · badge, title and the kind's secondary · ref and member count. Compact
/// (a pinned page that has scrolled): a 40pt thumb, no Badge, the title one size down.
private struct ConclusionDetailHeader<Content: ConclusionDetailBody, Secondary: View>: View {
    let content: Content
    let projectDir: String
    let secondary: (Content) -> Secondary
    let lead: ConclusionLeadModel
    let pageIdentifier: String
    var compact = false

    /// PDP-11: the menu exists only while the lead is not the earliest.
    private var showsLeadMenu: Bool { content.gallery.offersEarliest }

    var body: some View {
        HStack(alignment: compact ? .center : .top, spacing: compact ? PVSpacing.space5 : PVSpacing.space7) {
            thumbnail
            VStack(alignment: .leading, spacing: compact ? PVSpacing.space1 : PVSpacing.space4) {
                if !compact, let badge = content.badge {
                    PVBadge(text: badge.text, tone: .neutral, subtle: false, foreground: badge.ink)
                        .accessibilityLabel(badge.accessibilityLabel)
                }
                title
                secondary(content)
                    .environment(\.conclusionDetailHeaderCompact, compact)
            }
            .padding(.top, compact ? 0 : PVSpacing.space3)
            .frame(maxWidth: .infinity, alignment: .leading)
            VStack(alignment: .trailing, spacing: PVSpacing.space2) {
                if content.showsRef {
                    Text(verbatim: content.ref)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                        .accessibilityLabel(L10n.Conclusions.a11yRef(content.ref))
                }
                Text(verbatim: content.members)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
            .padding(.top, compact ? 0 : PVSpacing.space6)
        }
    }

    @ViewBuilder
    private var thumbnail: some View {
        let thumb = CachedThumbnail(
            projectDir: projectDir,
            relPath: content.thumbnail.imageRelPath,
            size: compact ? 40 : 120,
            fallbackMark: content.thumbnail.mark,
            fallbackLabel: content.thumbnail.label,
            imageLabel: content.thumbnail.imageLabel,
            colors: content.thumbnail.colors
        )
        if showsLeadMenu {
            let copy = content.gallery.choose
            thumb
                .pvRightClickMenu(
                    title: copy.menuTitle,
                    accessibilityIdentifier: "\(pageIdentifier).lead.menu",
                    items: [
                        PVActionMenuItem(
                            title: copy.useEarliest,
                            accessibilityIdentifier: "\(pageIdentifier).lead.useEarliest"
                        ) {
                            Task { await lead.useEarliest(copy: copy) }
                        },
                    ]
                )
                .help(L10n.string(copy.menuHint))
                .accessibilityHint(L10n.string(copy.menuHint))
                .accessibilityIdentifier("\(pageIdentifier).lead")
        } else {
            thumb
        }
    }

    private func nameText(_ text: String, styled: AttributedString?) -> Text {
        guard let styled else { return Text(verbatim: text) }
        return Text(ConclusionListRow.italicizingEmphasis(styled, font: titleFont))
    }

    private var titleFont: Font { PVFont.display(size: titleSize, weight: PVFontWeight.medium) }

    private var titleSize: CGFloat { compact ? PVTypeScale.h3 : PVTypeScale.display3 }

    @ViewBuilder
    private var title: some View {
        switch content.title {
        case .name(let text, let styled, let spoken):
            nameText(text, styled: styled)
                .font(titleFont)
                .tracking(titleSize * PVTracking.display)
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityLabel(Text(verbatim: spoken ?? text))
                .accessibilityAddTraits(.isHeader)
        case .label(let text):
            Text(verbatim: text)
                .font(titleFont.italic())
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        case .ref(let text):
            Text(verbatim: text)
                .font(PVFont.mono(size: compact ? PVTypeScale.h4 : PVTypeScale.h1, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        }
    }
}

/// One picture: its crop over a muted caption of up to two lines. The gallery passes the button that
/// opens the Citation as `overlay` (drawn over the crop only).
private struct GalleryTileView<Overlay: View>: View {
    let projectDir: String
    let relPath: String
    let fallbackMark: PVMarkKey
    let caption: String
    let captionWidth: CGFloat
    private let overlay: Overlay

    init(
        projectDir: String,
        relPath: String,
        fallbackMark: PVMarkKey,
        caption: String,
        captionWidth: CGFloat,
        @ViewBuilder overlay: () -> Overlay
    ) {
        self.projectDir = projectDir
        self.relPath = relPath
        self.fallbackMark = fallbackMark
        self.caption = caption
        self.captionWidth = captionWidth
        self.overlay = overlay()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: relPath,
                size: CachedThumbnail.pictureCrop,
                fallbackMark: fallbackMark
            )
            .accessibilityHidden(true)
            .overlay { overlay }
            Text(verbatim: caption)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(2)
                .truncationMode(.tail)
                .frame(maxWidth: captionWidth, alignment: .leading)
                .help(Text(verbatim: caption))
                .accessibilityHidden(true)
        }
    }
}
