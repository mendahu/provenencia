import SwiftUI

/// The **Source types** workspace destination (S2-03 board / S2-16 PR):
/// browse the project's `source_types` vocabulary, edit or delete
/// a type, and assign or remove the metadata fields it suggests.
struct SourceTypesView: View {
    @Environment(WorkspaceSession.self) private var session
    @State private var model: SourceTypesModel

    private let detailPaneWidth: CGFloat = 380

    init(
        session: WorkspaceSession,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        _model = State(
            initialValue: SourceTypesModel(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        )
    }

    var body: some View {
        Group {
            if let typesHandle: QueryHandle<[CatalogSourceType]> = session.queryHandle(
                SourceTypesModel.typesListKey(for: session)
            ) {
                SourceTypesContent(
                    typesHandle: typesHandle,
                    model: model,
                    detailPaneWidth: detailPaneWidth
                )
            } else {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .task {
            model.warmListQueries()
        }
    }
}

private struct SourceTypesContent: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Bindable var typesHandle: QueryHandle<[CatalogSourceType]>
    @Bindable var model: SourceTypesModel
    let detailPaneWidth: CGFloat

    var body: some View {
        VStack(spacing: 0) {
            header
            PVDivider()
            if let loadError = model.loadError, model.types.isEmpty {
                PVCallout(tone: .danger, message: L10n.Errors.message(for: loadError))
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.top, PVSpacing.space8)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
                    .accessibilityIdentifier("sourceTypes.loadError")
            } else {
                HStack(spacing: 0) {
                    SourceTypesListPane(model: model)
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                    PVDivider(axis: .vertical)
                    SourceTypesDetailPane(model: model)
                        .frame(width: detailPaneWidth)
                        .frame(maxHeight: .infinity)
                }
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .vocabularyToastOverlay($model.toast, identifier: "sourceTypes.toast")
        .pvConfirm(
            item: pendingDelete,
            copy: deleteCopy(for:),
            isRunning: model.isDeleting,
            accessibilityIdentifierPrefix: "sourceTypes.delete",
            onConfirm: {
                Task {
                    if await model.confirmDelete() {
                        navigation.fallbackToSectionRoot()
                    }
                }
            }
        ) { type in
            deleteDetail(for: type)
        }
        .onChange(of: navigation.currentLocation) { _, location in
            reconcileSelection(for: location)
        }
        .onChange(of: typesHandle.status) { _, _ in
            reconcileSelection(for: navigation.currentLocation)
        }
        .onChange(of: typesHandle.value) { _, _ in
            reconcileSelection(for: navigation.currentLocation)
        }
        .onAppear {
            reconcileSelection(for: navigation.currentLocation)
        }
        .accessibilityIdentifier("sourceTypes")
    }

    private func reconcileSelection(for location: WorkspaceLocation) {
        guard typesHandle.status == .ready || !(typesHandle.value ?? []).isEmpty else { return }
        let outcome = model.syncSelection(from: location)
        if outcome == .missingDeepId {
            navigation.fallbackToSectionRoot()
        }
    }

    private var pendingDelete: Binding<CatalogSourceType?> {
        Binding(
            get: { model.pendingDeleteType },
            set: { if $0 == nil { model.cancelDelete() } }
        )
    }

    private func deleteCopy(for type: CatalogSourceType) -> PVConfirmCopy {
        PVConfirmCopy(
            title: L10n.SourceTypes.deleteConfirmTitle(label: type.label),
            message: String(localized: L10n.SourceTypes.deleteConfirmMessage),
            confirm: L10n.SourceTypes.deleteType,
            cancel: L10n.SourceTypes.deleteKeep
        )
    }

    @ViewBuilder
    private func deleteDetail(for type: CatalogSourceType) -> some View {
        PVConfirmKeyChip(label: L10n.SourceTypes.deleteKeyReleased, value: type.key)
            .accessibilityIdentifier("sourceTypes.delete.key")
        if let deleteError = model.deleteError {
            PVCallout(tone: .danger, message: deleteError)
        }
    }

    private var header: some View {
        VocabularyHeader(
            title: L10n.Workspace.sourceTypesTitle,
            description: L10n.SourceTypes.description,
            countLine: model.countLine,
            addLabel: L10n.SourceTypes.addType,
            isAddDisabled: model.isAdding,
            identifierPrefix: "sourceTypes",
            onAdd: { model.openAdd() }
        )
    }
}

#if DEBUG
#Preview {
    let store = FakeStore()
    let projectDir = "/tmp/preview.provenencia"
    store.fieldsByProject[projectDir] = [
        CatalogMetadataField(id: "f1", key: "author", origin: "provenencia", label: "Author", dataType: "text", description: "Person or body responsible for the content."),
        CatalogMetadataField(id: "f2", key: "publication-date", origin: "provenencia", label: "Publication date", dataType: "date", description: "Date the source was published."),
        CatalogMetadataField(id: "f3", key: "publisher", origin: "provenencia", label: "Publisher", dataType: "text", description: "Who issued it."),
        CatalogMetadataField(id: "f4", key: "grandmas-album-code", origin: "user", label: "Grandma's album code", dataType: "text", description: "Pencil code on the back of prints."),
    ]
    store.sourceTypesByProject[projectDir] = [
        CatalogSourceType(id: "t1", key: "photograph", origin: "provenencia", label: "Photograph", description: "A photographic image of people, places or objects.", usedBy: 41),
        CatalogSourceType(id: "t2", key: "book", origin: "provenencia", label: "Book", description: "A published monograph.", usedBy: 18),
        CatalogSourceType(id: "t3", key: "letter", origin: "provenencia", label: "Letter", description: "Private correspondence.", usedBy: 0),
        CatalogSourceType(id: "t4", key: "family-scrapbook", origin: "user", label: "Family scrapbook", description: "Nan's albums.", usedBy: 3),
        CatalogSourceType(id: "t5", key: "grave-memorial", origin: "plugin:findagrave", label: "Grave memorial", description: "A memorial page.", usedBy: 5),
    ]
    store.suggestionsByType["t2"] = [
        CatalogTypeSuggestion(field: store.fieldsByProject[projectDir]![0], sortOrder: 0),
        CatalogTypeSuggestion(field: store.fieldsByProject[projectDir]![2], sortOrder: 1),
        CatalogTypeSuggestion(field: store.fieldsByProject[projectDir]![1], sortOrder: 2),
    ]
    store.suggestionsByType["t4"] = [
        CatalogTypeSuggestion(field: store.fieldsByProject[projectDir]![3], sortOrder: 0),
    ]
    let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
    return SourceTypesView(
        session: session,
        userID: "00000000-0000-7000-8000-000000000001",
        store: store
    )
    .environment(WorkspaceNavigation())
    .environment(session)
    .frame(width: 1180, height: 760)
}
#endif
