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

/// One section under Details. Place hierarchy sections carry relationship
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
    /// The header leads with a likeness other than the earliest, so its menu
    /// offers the earliest (S10-39, user 2026-10-17).
    var offersEarliest: Bool = false
}

/// One likeness tile under Details: the image, its Source's title, and the
/// composer location its Citation opens at (S10-35).
struct ConclusionDetailLikeness: Equatable, Identifiable {
    var id: String
    var imageRelPath: String
    var sourceTitle: String
    var location: WorkspaceLocation
    /// The tile the header leads with wears the Portrait Badge (S10-39, user
    /// 2026-10-17); exactly one tile.
    var isPortrait: Bool = false
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
    var likenesses: [ConclusionDetailLikeness] { get }
}

extension ConclusionDetailBody {
    var sections: [ConclusionDetailSection] { [] }
    var badge: ConclusionDetailBadge? { nil }
    var likenesses: [ConclusionDetailLikeness] { [] }
}

/// What the `pageSections` slot receives beside the content.
struct ConclusionDetailPageSectionsContext {
    var projectDir: String
    var refreshing: Bool
    var portrait: ConclusionPortraitModel?
}

/// Shared Conclusion detail chrome (S9-24). A view, not a place: the caller
/// names the entity, and the place registry warms `conclusionDetail` for
/// that id. Person, Event, and Place stay separate history entries.
struct ConclusionDetailPage<Content: ConclusionDetailBody, Summary: View, PageSections: View>: View {
    let session: WorkspaceSession
    let entityId: String
    let pageIdentifier: String
    let errorIdentifier: String
    let make: (CatalogConclusionDetail) -> Content
    let summary: (Content) -> Summary
    /// Handed only by a kind whose likenesses can be chosen as the portrait (the Person page).
    var portrait: ConclusionPortraitModel? = nil
    /// Everything below the header. The convenience initializer supplies `ConclusionDetailDefaultPageSections`.
    let pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections

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
                    summary: summary,
                    portrait: portrait,
                    pageSections: pageSections
                )
            } else {
                Self.page(identifier: pageIdentifier) { Self.loading }
            }
        }
        .task {
            let _: QueryHandle<CatalogConclusionDetail> = session.query(key)
        }
    }

    static var loading: some View {
        ProgressView()
            .controlSize(.small)
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    /// Page chrome shared by every state, in the content column.
    static func page<Body: View>(identifier: String, @ViewBuilder body: () -> Body) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: PVSpacing.space9) {
                body()
            }
            .frame(maxWidth: PVSpacing.widthContentMax, alignment: .leading)
            .padding(.horizontal, PVSpacing.gutterPage)
            .padding(.vertical, PVSpacing.space8)
            .frame(maxWidth: .infinity)
        }
        .background(PVColor.surfacePage)
        .accessibilityIdentifier(identifier)
    }
}

extension ConclusionDetailPage where PageSections == ConclusionDetailDefaultPageSections<Content> {
    init(
        session: WorkspaceSession,
        entityId: String,
        pageIdentifier: String,
        errorIdentifier: String,
        make: @escaping (CatalogConclusionDetail) -> Content,
        summary: @escaping (Content) -> Summary,
        portrait: ConclusionPortraitModel? = nil
    ) {
        self.init(
            session: session,
            entityId: entityId,
            pageIdentifier: pageIdentifier,
            errorIdentifier: errorIdentifier,
            make: make,
            summary: summary,
            portrait: portrait,
            pageSections: { ConclusionDetailDefaultPageSections(content: $0, context: $1) }
        )
    }
}

/// Today's below-header fill: Details, the likeness gallery, then the Place relationship sections.
struct ConclusionDetailDefaultPageSections<Content: ConclusionDetailBody>: View {
    let content: Content
    let context: ConclusionDetailPageSectionsContext

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: L10n.Conclusions.detailsSection, aside: {
                if context.refreshing {
                    ProgressView().controlSize(.small)
                }
            })
            ForEach(content.rows) { row in
                ReconciledValueRow(model: row)
            }
        }
        if !content.likenesses.isEmpty {
            ConclusionDetailLikenessSection(
                projectDir: context.projectDir,
                likenesses: content.likenesses,
                portrait: context.portrait
            )
        }
        ForEach(content.sections) { section in
            ConclusionDetailRelationshipSection(section: section)
        }
    }
}

private struct ConclusionDetailLoaded<Content: ConclusionDetailBody, Summary: View, PageSections: View>: View {
    @Bindable var handle: QueryHandle<CatalogConclusionDetail>
    let projectDir: String
    let pageIdentifier: String
    let errorIdentifier: String
    let make: (CatalogConclusionDetail) -> Content
    let summary: (Content) -> Summary
    let portrait: ConclusionPortraitModel?
    let pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections

    private var presentation: ConclusionDetailPresentation<Content> {
        ConclusionDetailPresentation(handle.read, isFetching: handle.isFetching, content: make)
    }

    var body: some View {
        ConclusionDetailPage<Content, Summary, PageSections>.page(identifier: pageIdentifier) {
            switch presentation {
            case .loading:
                ConclusionDetailPage<Content, Summary, PageSections>.loading
            case .failed(let message):
                PVCallout(tone: .danger, message: message)
                    .accessibilityIdentifier(errorIdentifier)
            case .content(let content, let refreshing):
                ConclusionDetailHeader(
                    content: content,
                    projectDir: projectDir,
                    summary: summary,
                    portrait: portrait
                )
                pageSections(
                    content,
                    ConclusionDetailPageSectionsContext(
                        projectDir: projectDir,
                        refreshing: refreshing,
                        portrait: portrait
                    )
                )
            }
        }
    }
}

/// The Person page's likenesses: a header with the count, then one 96pt crop
/// per likeness in a wrapping grid, each captioned with its Source. A tile
/// opens the Citation that holds it (S10-35, S10-D12 PDP-7, PDP-8).
private struct ConclusionDetailLikenessSection: View {
    let projectDir: String
    let likenesses: [ConclusionDetailLikeness]
    let portrait: ConclusionPortraitModel?

    @Environment(WorkspaceNavigation.self) private var navigation
    @FocusState private var focusedID: String?

    private static let column: CGFloat = 120

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: L10n.Conclusions.likenessesSection, meta: String(likenesses.count))
            LazyVGrid(
                columns: [GridItem(
                    .adaptive(minimum: Self.column, maximum: Self.column),
                    spacing: PVSpacing.space6,
                    alignment: .top
                )],
                alignment: .leading,
                spacing: PVSpacing.space7
            ) {
                ForEach(likenesses) { tile in
                    tileView(tile)
                }
            }
            .padding(.top, PVSpacing.space6)
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.string(L10n.Conclusions.likenessesSection))
    }

    private func tileView(_ tile: ConclusionDetailLikeness) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            LikenessTileView(
                projectDir: projectDir,
                relPath: tile.imageRelPath,
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
                .help(Text(L10n.Conclusions.likenessOpenTip))
                .accessibilityLabel(L10n.Conclusions.likenessTileLabel(tile.sourceTitle))
                .accessibilityIdentifier("persons.detail.likeness.\(tile.id)")
            }
            if let portrait {
                if tile.isPortrait {
                    PVBadge(L10n.Conclusions.portraitBadge, tone: .accent, icon: .photo)
                        .accessibilityIdentifier("persons.detail.likeness.\(tile.id).portraitBadge")
                } else {
                    PVButton(L10n.Conclusions.useAsPortrait, variant: .ghost, size: .sm) {
                        Task { await portrait.choose(tile) }
                    }
                    .disabled(portrait.isSaving)
                    .fixedSize(horizontal: true, vertical: false)
                    .accessibilityLabel(L10n.Conclusions.useAsPortraitLabel(tile.sourceTitle))
                    .accessibilityIdentifier("persons.detail.likeness.\(tile.id).usePortrait")
                }
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
                ? L10n.Conclusions.a11yList(L10n.string(section.title), rest: section.emptyText)
                : L10n.string(section.title)
        )
    }
}

/// Thumbnail slot · badge, title and the kind's summary · ref and member count.
private struct ConclusionDetailHeader<Content: ConclusionDetailBody, Summary: View>: View {
    let content: Content
    let projectDir: String
    let summary: (Content) -> Summary
    let portrait: ConclusionPortraitModel?

    /// PDP-11: the menu exists only while the lead is not the earliest.
    private var showsPortraitMenu: Bool { portrait != nil && content.thumbnail.offersEarliest }

    var body: some View {
        HStack(alignment: .top, spacing: PVSpacing.space7) {
            thumbnail
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                if let badge = content.badge {
                    PVBadge(text: badge.text, tone: .neutral, subtle: false, foreground: badge.ink)
                        .accessibilityLabel(badge.accessibilityLabel)
                }
                title
                summary(content)
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
        if showsPortraitMenu {
            thumb
                .pvRightClickMenu(
                    title: L10n.Conclusions.portraitMenuTitle,
                    accessibilityIdentifier: "persons.detail.portrait.menu",
                    items: [
                        PVActionMenuItem(
                            title: L10n.Conclusions.useEarliestLikeness,
                            accessibilityIdentifier: "persons.detail.portrait.useEarliest"
                        ) {
                            Task { await portrait?.useEarliest() }
                        },
                    ]
                )
                .help(L10n.string(L10n.Conclusions.portraitMenuHint))
                .accessibilityHint(L10n.string(L10n.Conclusions.portraitMenuHint))
                .accessibilityIdentifier("persons.detail.portrait")
        } else {
            thumb
        }
    }

    @ViewBuilder
    private var title: some View {
        switch content.title {
        case .name(let text):
            Text(verbatim: text)
                .font(PVFont.display(size: PVTypeScale.display3, weight: PVFontWeight.medium))
                .tracking(PVTypeScale.display3 * PVTracking.display)
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        case .label(let text):
            Text(verbatim: text)
                .font(PVFont.display(size: PVTypeScale.display3, weight: PVFontWeight.medium).italic())
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        case .ref(let text):
            Text(verbatim: text)
                .font(PVFont.mono(size: PVTypeScale.h1, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        }
    }
}

/// One likeness: its crop over a one-line muted caption. The gallery passes the button that opens
/// the Citation as `overlay` (drawn over the crop only).
private struct LikenessTileView<Overlay: View>: View {
    let projectDir: String
    let relPath: String
    let caption: String
    let captionWidth: CGFloat
    private let overlay: Overlay

    init(
        projectDir: String,
        relPath: String,
        caption: String,
        captionWidth: CGFloat,
        @ViewBuilder overlay: () -> Overlay
    ) {
        self.projectDir = projectDir
        self.relPath = relPath
        self.caption = caption
        self.captionWidth = captionWidth
        self.overlay = overlay()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: relPath,
                size: CachedThumbnail.likenessCrop,
                fallbackMark: .subjectPerson
            )
            .accessibilityHidden(true)
            .overlay { overlay }
            Text(verbatim: caption)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
                .truncationMode(.tail)
                .frame(maxWidth: captionWidth, alignment: .leading)
                .help(Text(verbatim: caption))
                .accessibilityHidden(true)
        }
    }
}
