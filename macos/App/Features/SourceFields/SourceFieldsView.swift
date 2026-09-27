import SwiftUI

/// The **Source fields** workspace destination (S2-02 board / S2-15 PR):
/// browse and create/edit the project's `source_metadata_fields`
/// vocabulary. Mounts inside the existing S2-01 workspace content host.
struct SourceFieldsView: View {
    @Environment(WorkspaceSession.self) private var session
    @State private var model: SourceFieldsModel

    private let detailPaneWidth: CGFloat = 380

    init(
        session: WorkspaceSession,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        _model = State(
            initialValue: SourceFieldsModel(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        )
    }

    var body: some View {
        Group {
            if let fieldsHandle: QueryHandle<[CatalogMetadataField]> = session.queryHandle(
                SourceFieldsModel.fieldsListKey(for: session)
            ) {
                SourceFieldsContent(
                    fieldsHandle: fieldsHandle,
                    model: model,
                    detailPaneWidth: detailPaneWidth
                )
            } else {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .task {
            model.warmFieldsQuery()
        }
    }
}

private struct SourceFieldsContent: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Bindable var fieldsHandle: QueryHandle<[CatalogMetadataField]>
    @Bindable var model: SourceFieldsModel
    let detailPaneWidth: CGFloat

    var body: some View {
        VStack(spacing: 0) {
            header
            PVDivider()
            if let loadError = model.loadError, model.fields.isEmpty {
                PVCallout(tone: .danger, message: L10n.Errors.message(for: loadError))
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.top, PVSpacing.space8)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
                    .accessibilityIdentifier("sourceFields.loadError")
            } else {
                HStack(spacing: 0) {
                    SourceFieldsListPane(model: model)
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                    PVDivider(axis: .vertical)
                    SourceFieldsDetailPane(model: model)
                        .frame(width: detailPaneWidth)
                        .frame(maxHeight: .infinity)
                }
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .vocabularyToastOverlay($model.toast, identifier: "sourceFields.toast")
        .pvDeleteImpact(
            item: Binding(
                get: { model.pendingImpact },
                set: { newValue in
                    model.pendingImpact = newValue
                    if newValue == nil { model.cancelDelete() }
                }
            ),
            isRunning: model.isDeleting,
            accessibilityIdentifierPrefix: "sourceFields.deleteImpact",
            onConfirm: {
                Task {
                    if await model.confirmPendingImpact() {
                        navigation.fallbackToSectionRoot()
                    }
                }
            },
            onNavigate: { location in
                navigation.go(to: location)
            }
        )
        .onChange(of: navigation.currentLocation) { _, location in
            reconcileSelection(for: location)
        }
        .onChange(of: fieldsHandle.status) { _, status in
            if status == .ready { model.syncCatalogCounts() }
            reconcileSelection(for: navigation.currentLocation)
        }
        .onChange(of: fieldsHandle.value) { _, _ in
            reconcileSelection(for: navigation.currentLocation)
        }
        .onAppear {
            reconcileSelection(for: navigation.currentLocation)
            if fieldsHandle.status == .ready {
                model.syncCatalogCounts()
            }
        }
        .accessibilityIdentifier("sourceFields")
    }

    private func reconcileSelection(for location: WorkspaceLocation) {
        guard fieldsHandle.status == .ready || !(fieldsHandle.value ?? []).isEmpty else { return }
        let outcome = model.syncSelection(from: location)
        if outcome == .missingDeepId {
            navigation.fallbackToSectionRoot()
        }
    }

    private var header: some View {
        VocabularyHeader(
            title: L10n.Workspace.sourceFieldsTitle,
            description: L10n.SourceFields.description,
            countLine: model.countLine,
            addLabel: L10n.SourceFields.addField,
            isAddDisabled: model.isAdding,
            identifierPrefix: "sourceFields",
            onAdd: { model.openAdd() }
        )
    }
}

#if DEBUG
#Preview {
    let store = FakeStore()
    let projectDir = "/tmp/preview.provenencia"
    store.fieldsByProject[projectDir] = [
        CatalogMetadataField(id: "1", key: "author", origin: "provenencia", label: "Author", dataType: "text", description: "Person or body responsible for the content of the source."),
        CatalogMetadataField(id: "2", key: "publication-date", origin: "provenencia", label: "Publication date", dataType: "date", description: "Date the source was published."),
        CatalogMetadataField(id: "3", key: "grandmas-album-code", origin: "user", label: "Grandma's album code", dataType: "text", description: "Pencil code on the back of prints."),
        CatalogMetadataField(id: "4", key: "memorial-id", origin: "plugin:findagrave", label: "Memorial id", dataType: "text", description: "Numeric memorial identifier."),
    ]
    let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
    return SourceFieldsView(
        session: session,
        userID: "00000000-0000-7000-8000-000000000001",
        store: store
    )
    .environment(WorkspaceNavigation())
    .environment(session)
    .frame(width: 1180, height: 760)
}
#endif
