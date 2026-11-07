import SwiftUI

/// The host of a Person page section that reads its own query
/// (`ConclusionPageSection.queryKey`): it gets the section's handle, starts
/// the read, and draws the spinner, the error Callout and the empty state
/// around the section's `content`. Only the content differs between sections.
///
/// The id root is `persons.detail.<section>`: `.error`, `.empty`, and the root
/// itself on the content.
struct ConclusionPageSectionHost<Value, Content: Equatable, Body: View>: View {
    let section: ConclusionPageSection
    let entityId: String
    let errorTitle: LocalizedStringResource
    let emptyTitle: LocalizedStringResource
    let emptyMessage: String
    let make: (Value) -> Content
    let isEmpty: (Content) -> Bool
    let content: (Content) -> Body

    @Environment(WorkspaceSession.self) private var session

    init(
        section: ConclusionPageSection,
        entityId: String,
        errorTitle: LocalizedStringResource,
        emptyTitle: LocalizedStringResource,
        emptyMessage: String,
        make: @escaping (Value) -> Content,
        isEmpty: @escaping (Content) -> Bool,
        @ViewBuilder body: @escaping (Content) -> Body
    ) {
        self.section = section
        self.entityId = entityId
        self.errorTitle = errorTitle
        self.emptyTitle = emptyTitle
        self.emptyMessage = emptyMessage
        self.make = make
        self.isEmpty = isEmpty
        self.content = body
    }

    private var key: CatalogQueryKey {
        guard let key = section.queryKey(project: session.projectKey, entityId: entityId) else {
            preconditionFailure("\(section.rawValue) reads no query of its own")
        }
        return key
    }

    private var idRoot: String { "persons.detail.\(section.rawValue)" }

    var body: some View {
        Group {
            if let handle: QueryHandle<Value> = session.queryHandle(key) {
                loaded(handle)
            } else {
                loading
            }
        }
        .task {
            let _: QueryHandle<Value> = session.query(key)
        }
    }

    private var loading: some View {
        ProgressView()
            .frame(maxWidth: .infinity)
            .pvKeepsPageBand()
            .padding(.top, 96)
    }

    private func error(_ message: String) -> some View {
        PVCallout(tone: .danger, title: errorTitle, message: message)
            .pvKeepsPageBand()
            .accessibilityIdentifier("\(idRoot).error")
    }

    @ViewBuilder
    private func loaded(_ handle: QueryHandle<Value>) -> some View {
        let presentation = ConclusionDetailPresentation<Content>(
            handle.read,
            isFetching: handle.isFetching,
            content: make
        )
        switch presentation {
        case .loading:
            loading
        case .gone:
            // The page above leaves when its own read reports the handle gone; a section whose
            // read fails first, under a page that still holds its value, shows why.
            error(L10n.string(L10n.Errors.conclusionDetailsNotFound))
        case .failed(let message):
            error(message)
        case .content(let value, _):
            if isEmpty(value) {
                PVEmptyState(glyph: section.glyph, title: emptyTitle, message: emptyMessage)
                    .pvKeepsPageBand()
                    .accessibilityIdentifier("\(idRoot).empty")
            } else {
                content(value)
                    .accessibilityIdentifier(idRoot)
            }
        }
    }
}
