#if DEBUG
import SwiftUI

/// The whole Person page over a `FakeStore` that fails a read on purpose, so the notices of
/// S16-D1 can be seen in Xcode's canvas: a page that cannot be read, a page whose reload
/// failed behind a shown Timeline, and a Family that cannot be read.
private struct PersonDetailNoticePreview: View {
    static let entityId = "preview-person"

    enum Failure {
        case page
        case pageStale
        case family
    }

    let failure: Failure
    let section: ConclusionPageSection
    let width: CGFloat

    @State private var session: WorkspaceSession
    @State private var navigation: WorkspaceNavigation

    init(failure: Failure, section: ConclusionPageSection, width: CGFloat) {
        self.failure = failure
        self.section = section
        self.width = width
        let store = FakeStore()
        store.seededConclusionDetails[Self.entityId] = CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(
                id: Self.entityId, ref: "PER-7KD45", subjectTypeID: "type-person", label: "Johann Weber"
            ),
            fields: [],
            memberCount: 3
        )
        store.seededPersonEvents[Self.entityId] = CatalogPersonTimeline()
        store.seededPersonClaims[Self.entityId] = []
        switch failure {
        case .page: store.getConclusionDetailError = PersonSourcesPreviewData.previewFailure
        case .family: store.getPersonFamilyError = PersonSourcesPreviewData.previewFailure
        case .pageStale: break
        }
        // A page that shows sections is on a Person location; the default location offers none.
        let navigation = WorkspaceNavigation()
        navigation.go(to: WorkspaceLocation.personDetail(
            entityId: Self.entityId, ref: "PER-7KD45", title: nil
        ).withPageSection(section))
        _navigation = State(initialValue: navigation)
        _session = State(initialValue: WorkspaceSession(
            projectKey: ProjectKey(projectDir: "/tmp/preview.provenencia"), store: store
        ))
    }

    var body: some View {
        PersonDetailView(session: session, entityId: Self.entityId, userID: "preview-user")
            .environment(session)
            .environment(navigation)
            .environment(\.pvWidthClass, width >= PVSpacing.widthWideBreakpoint ? .wide : .regular)
            .frame(width: width, height: 900)
            .background(PVColor.surfacePage)
            .task { await failTheDetailAfterTheFirstAnswer() }
    }

    /// "Timeline, page stale": the page is read once, then the next read of it fails and is asked for.
    private func failTheDetailAfterTheFirstAnswer() async {
        guard failure == .pageStale else { return }
        let key = CatalogQueryKey.conclusionDetail(project: session.projectKey, entityId: Self.entityId)
        await PersonSourcesPreviewData.failAfterFirstAnswer(
            key, as: CatalogConclusionDetail.self, session: session
        ) { $0.getConclusionDetailError = PersonSourcesPreviewData.previewFailure }
    }
}

#Preview("Page failed") { PersonDetailNoticePreview(failure: .page, section: .timeline, width: 1102) }
#Preview("Page failed, Wide") { PersonDetailNoticePreview(failure: .page, section: .timeline, width: 2296) }
#Preview("Timeline, page stale") { PersonDetailNoticePreview(failure: .pageStale, section: .timeline, width: 1102) }
#Preview("Timeline, page stale, Wide") { PersonDetailNoticePreview(failure: .pageStale, section: .timeline, width: 2296) }
#Preview("Family failed") { PersonDetailNoticePreview(failure: .family, section: .family, width: 1102) }
#Preview("Family failed, Wide") { PersonDetailNoticePreview(failure: .family, section: .family, width: 2296) }
#endif
