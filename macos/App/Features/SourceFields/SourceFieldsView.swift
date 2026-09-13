import SwiftUI

/// The **Source fields** workspace destination (S2-02 board / S2-15 PR):
/// browse, search, and create/edit the project's `source_metadata_fields`
/// vocabulary. Mounts inside the existing S2-01 workspace content host —
/// see `WorkspaceContent` — not a second window chrome.
///
/// Chrome choice for add/edit (S2-02 F-18 leaves this open): a nested
/// detail pane beside the list, not a modal/sheet. It keeps the
/// researcher's place in the list while filling out the form and matches
/// the density of the rest of the workspace.
struct SourceFieldsView: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: SourceFieldsModel

    /// Detail pane width — between the web board's `minmax(360px, 420px)`
    /// column and `PVSpacing.widthInspector` (340pt); no shared token
    /// covers this exact range, so it's a local literal like
    /// `PVToast`'s own `frame(maxWidth: 360)`.
    private let detailPaneWidth: CGFloat = 380

    init(
        projectDir: String,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        _model = State(
            initialValue: SourceFieldsModel(
                projectDir: projectDir,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        )
    }

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
        .pvConfirmSheet(
            item: pendingDelete,
            copy: deleteCopy(for:),
            isRunning: model.isDeleting,
            accessibilityIdentifierPrefix: "sourceFields.delete",
            onConfirm: { Task { await model.confirmDelete() } }
        ) { field in
            deleteDetail(for: field)
        }
        .task {
            await model.load()
            applyWorkspaceLocation()
        }
        .onAppear { applyWorkspaceLocation() }
        .onChange(of: navigation.currentLocation) { _, _ in applyWorkspaceLocation() }
        .accessibilityIdentifier("sourceFields")
    }

    private func applyWorkspaceLocation() {
        let location = navigation.currentLocation
        guard location.section == .sourceFields else { return }
        if model.isAdding { return }
        if let fieldId = location.fieldId {
            if model.fields.contains(where: { $0.id == fieldId }) {
                model.select(fieldId)
            } else if !model.isLoading {
                navigation.fallbackToSectionRoot()
            }
        } else {
            model.clearHistorySelection()
        }
    }

    /// Dismissal is driven by the model, not by the sheet: a successful delete
    /// clears `pendingDeleteID`, and a failed one keeps the sheet up with the
    /// reason (see `pvConfirmSheet`). The sheet renders its copy and detail
    /// from the field it receives — a snapshot `pvConfirmSheet` holds through
    /// the dismiss animation — never from `model.pendingDeleteField` live.
    private var pendingDelete: Binding<CatalogMetadataField?> {
        Binding(
            get: { model.pendingDeleteField },
            set: { if $0 == nil { model.cancelDelete() } }
        )
    }

    private func deleteCopy(for field: CatalogMetadataField) -> PVConfirmCopy {
        PVConfirmCopy(
            title: L10n.SourceFields.deleteConfirmTitle(label: field.label),
            message: String(localized: L10n.SourceFields.deleteConfirmMessage),
            confirm: L10n.SourceFields.deleteField,
            cancel: L10n.SourceFields.deleteKeep
        )
    }

    /// The consequence that earns this a sheet rather than a plain alert: the
    /// mono-set key the delete releases, plus the reason if it failed.
    @ViewBuilder
    private func deleteDetail(for field: CatalogMetadataField) -> some View {
        PVConfirmKeyChip(label: L10n.SourceFields.deleteKeyReleased, value: field.key)
            .accessibilityIdentifier("sourceFields.delete.key")
        if let deleteError = model.deleteError {
            PVCallout(tone: .danger, message: deleteError)
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
    return SourceFieldsView(
        projectDir: projectDir,
        userID: "00000000-0000-7000-8000-000000000001",
        store: store
    )
    .environment(WorkspaceNavigation())
    .frame(width: 1180, height: 760)
}
#endif
