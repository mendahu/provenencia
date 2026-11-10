import SwiftUI

/// What a Conclusion page shows for its handle's state. A stale page stays
/// readable while it refreshes. Person, Event, and Place each supply
/// their own content; the page does not know which kind it is.
enum ConclusionDetailPresentation<Content: Equatable>: Equatable {
    case loading
    /// The handle no longer exists (deleted or merged), not a failure that may pass; the page leaves.
    case gone
    case failed(String)
    /// `refreshFailure` is the message of a reload that failed behind the content, nil otherwise.
    case content(Content, refreshing: Bool, refreshFailure: String?)

    init<Value>(
        _ reading: QueryReading<Value>,
        isFetching: Bool,
        content: (Value) -> Content
    ) {
        switch reading {
        case .loaded(let value, let refreshFailure):
            if let refreshFailure, Self.isGone(refreshFailure) {
                self = .gone
            } else {
                self = .content(
                    content(value),
                    refreshing: isFetching,
                    refreshFailure: refreshFailure.map { L10n.Errors.message(for: $0) }
                )
            }
        case .failed(let error):
            if Self.isGone(error) {
                self = .gone
            } else {
                self = .failed(L10n.Errors.message(for: error))
            }
        case .loading:
            self = .loading
        }
    }

    private static func isGone(_ error: Error) -> Bool {
        (error as? CoreInvokeError)?.engineCode == .conclusionNotFound
    }
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
        leadLabel: (String) -> String
    ) -> ConclusionDetailThumbnail {
        let lead = gallery.lead
        return ConclusionDetailThumbnail(
            mark: gallery.fallbackMark,
            label: lead == nil ? noImage : withImage,
            imageRelPath: lead?.imageRelPath ?? "",
            imageLabel: lead.map { leadLabel($0.citation.title) },
            colors: gallery.fallbackColors
        )
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
    var title: PVTitleSource { get }
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
    var openPageSection: ConclusionPageSection? = nil
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
                    key: key,
                    session: session,
                    entityId: entityId,
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
        // The viewer draws the Toast itself while it is open; one Toast has one host.
        .vocabularyToastOverlay(lead.toastInViewer ? .constant(nil) : $lead.toast, identifier: "\(pageIdentifier).toast")
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
            ConclusionGallerySection(
                look: .small,
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

private struct ConclusionDetailLoaded<Content: ConclusionDetailBody, Secondary: View, PageSections: View>: View {
    @Bindable var handle: QueryHandle<CatalogConclusionDetail>
    let key: CatalogQueryKey
    let session: WorkspaceSession
    let entityId: String
    let projectDir: String
    let pageIdentifier: String
    let errorIdentifier: String
    let make: (CatalogConclusionDetail) -> Content
    let secondary: (Content) -> Secondary
    let lead: ConclusionLeadModel
    let pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections

    @Environment(WorkspaceNavigation.self) private var navigation

    private var presentation: ConclusionDetailPresentation<Content> {
        ConclusionDetailPresentation(handle.reading, isFetching: handle.isFetching, content: make)
    }

    /// The page sections this kind of page offers; none for Event and Place.
    private var offeredPageSections: [ConclusionPageSection] {
        ConclusionPageSection.offered(by: navigation.currentLocation.section)
    }

    /// The open page section: the location's, else the kind's first.
    private var openPageSection: ConclusionPageSection? {
        navigation.currentLocation.openPageSection
    }

    /// The open section is reloading behind a value it already shows. Its first load
    /// is the section's own spinner and does not count.
    private var sectionRefreshing: Bool {
        guard let openPageSection else { return false }
        return openPageSection.reads(project: session.projectKey, entityId: entityId)
            .filter(\.role.showsRefreshing)
            .contains { session.isRefreshing($0.key) }
    }

    /// Reads the page's detail again; the notice's Try again.
    private func readAgain() {
        let _: QueryHandle<CatalogConclusionDetail> = session.reread(key)
    }

    var body: some View {
        let presentation = presentation
        Group {
            switch presentation {
            case .loading, .gone:
                // A handle that is gone leaves for its list; the Callout never paints.
                ConclusionDetailPage<Content, Secondary, PageSections>.loadingPage
            case .failed(let message):
                PVPageColumns(primary: {
                    PVRetryCallout(
                        title: L10n.Conclusions.pageReadErrorTitle,
                        message: message,
                        handle: handle,
                        id: errorIdentifier,
                        onRetry: readAgain
                    )
                    .accessibilityIdentifier(errorIdentifier)
                })
            case .content(let content, let refreshing, let refreshFailure):
                PVPageColumns(
                    headerPinned: !offeredPageSections.isEmpty,
                    scrollResetKey: openPageSection,
                    primaryFill: openPageSection?.fillsWidth ?? false,
                    header: {
                        ConclusionDetailHeaderBand(
                            content: content,
                            projectDir: projectDir,
                            secondary: secondary,
                            lead: lead,
                            pageIdentifier: pageIdentifier,
                            offeredPageSections: offeredPageSections,
                            openPageSection: openPageSection,
                            refreshing: refreshing || sectionRefreshing,
                            refreshFailure: refreshFailure,
                            handle: handle,
                            retry: readAgain,
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
                                    openPageSection: openPageSection
                                )
                            )
                        }
                    }
                )
            }
        }
        .leavesForSectionRoot(
            when: presentation == .gone,
            notice: presentation == .gone
                ? L10n.string(L10n.Conclusions.goneNotice(section: navigation.currentLocation.section))
                : nil
        )
    }
}

/// The header and, on a page with sections, the section switcher under it. Pinned pages
/// supply the band's own padding; an unpinned page has none.
private struct ConclusionDetailHeaderBand<Content: ConclusionDetailBody, Secondary: View>: View {
    let content: Content
    let projectDir: String
    let secondary: (Content) -> Secondary
    let lead: ConclusionLeadModel
    let pageIdentifier: String
    let offeredPageSections: [ConclusionPageSection]
    let openPageSection: ConclusionPageSection?
    let refreshing: Bool
    /// The message of a page re-read that failed behind the shown page, nil when none did.
    let refreshFailure: String?
    let handle: QueryHandle<CatalogConclusionDetail>
    let retry: () -> Void
    let select: (ConclusionPageSection) -> Void

    /// A page with page sections pins its header; one without scrolls it away.
    private var pinned: Bool { openPageSection != nil }

    private var bar: CGFloat { PVSpacing.space7 }

    var body: some View {
        VStack(alignment: .leading, spacing: bar) {
            ConclusionDetailHeader(
                content: content,
                projectDir: projectDir,
                secondary: secondary,
                lead: lead,
                pageIdentifier: pageIdentifier
            )
            if let openPageSection {
                switcher(selection: openPageSection)
            }
            if let refreshFailure {
                PVRetryCallout(
                    title: L10n.Conclusions.pageReadErrorTitle,
                    message: L10n.Conclusions.readNoticeStaleMessage(refreshFailure),
                    handle: handle,
                    id: "\(pageIdentifier).refreshError",
                    onRetry: retry
                )
                .accessibilityIdentifier("\(pageIdentifier).refreshError")
            }
        }
        .padding(.top, pinned ? PVSpacing.space9 : 0)
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

/// Thumbnail slot · badge, title and the kind's secondary · ref and member count.
private struct ConclusionDetailHeader<Content: ConclusionDetailBody, Secondary: View>: View {
    let content: Content
    let projectDir: String
    let secondary: (Content) -> Secondary
    let lead: ConclusionLeadModel
    let pageIdentifier: String

    /// PDP-11: the menu exists only while the lead is not the earliest.
    private var showsLeadMenu: Bool { content.gallery.offersEarliest }

    var body: some View {
        HStack(alignment: .top, spacing: PVSpacing.space7) {
            thumbnail
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                if let badge = content.badge {
                    PVBadge(text: badge.text, tone: .neutral, subtle: false, foreground: badge.ink)
                        .accessibilityLabel(badge.accessibilityLabel)
                }
                PVTitleText(content.title, style: .display)
                secondary(content)
            }
            .padding(.top, PVSpacing.space3)
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
            .padding(.top, PVSpacing.space6)
        }
    }

    @ViewBuilder
    private var thumbnail: some View {
        let thumb = CachedThumbnail(
            projectDir: projectDir,
            relPath: content.thumbnail.imageRelPath,
            size: 120,
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
}
