import SwiftUI

/// The **Source types** workspace destination (S2-03 board / S2-16 PR):
/// browse the project's `source_types` vocabulary, edit or delete
/// a type, and assign or remove the metadata fields it suggests. Mounts
/// inside the existing S2-01 workspace content host — see
/// `WorkspaceContent` — not a second window chrome.
///
/// Navigation choice (S2-03 T-13 leaves this open): a master–detail split,
/// not expand-in-list or a pushed detail. A type carries a description plus
/// a variable-length suggestions list, so expanding a row in place would
/// push the rest of the vocabulary off screen; and assigning fields is a
/// back-and-forth task that a push would make tedious. It is also what
/// Source fields already does, so the two vocabulary destinations behave
/// alike.
struct SourceTypesView: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: SourceTypesModel

    /// Detail pane width — between the web board's `minmax(340px, 400px)`
    /// column and `PVSpacing.widthInspector` (340pt); no shared token covers
    /// this exact range, so it's a local literal like `PVToast`'s own
    /// `frame(maxWidth: 360)`.
    private let detailPaneWidth: CGFloat = 380

    init(
        projectDir: String,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        _model = State(
            initialValue: SourceTypesModel(
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
        .pvConfirmSheet(
            item: pendingDelete,
            copy: deleteCopy(for:),
            isRunning: model.isDeleting,
            accessibilityIdentifierPrefix: "sourceTypes.delete",
            onConfirm: { Task { await model.confirmDelete() } }
        ) { type in
            deleteDetail(for: type)
        }
        .task {
            await model.load()
            applyWorkspaceLocation()
        }
        .onAppear { applyWorkspaceLocation() }
        .onChange(of: navigation.currentLocation) { _, _ in applyWorkspaceLocation() }
        .accessibilityIdentifier("sourceTypes")
    }

    private func applyWorkspaceLocation() {
        let location = navigation.currentLocation
        guard location.section == .sourceTypes else { return }
        if model.isAdding { return }
        if let typeId = location.typeId {
            if model.types.contains(where: { $0.id == typeId }) {
                model.select(typeId)
            } else if model.hasCompletedInitialLoad {
                navigation.fallbackToSectionRoot()
            }
        } else {
            model.clearHistorySelection()
        }
    }

    /// Dismissal is driven by the model, not by the sheet: a successful delete
    /// clears `pendingDeleteID`, and a failed one keeps the sheet up with the
    /// reason (see `pvConfirmSheet`). The sheet renders its copy and detail
    /// from the type it receives — a snapshot `pvConfirmSheet` holds through
    /// the dismiss animation — never from `model.pendingDeleteType` live.
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

    /// The consequence that earns this a sheet rather than a plain alert: the
    /// mono-set key the delete releases, plus the reason if it failed.
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
        CatalogSourceType(id: "t2", key: "book", origin: "provenencia", label: "Book", description: "A published monograph — county history, compiled genealogy, printed transcript.", usedBy: 18),
        CatalogSourceType(id: "t3", key: "letter", origin: "provenencia", label: "Letter", description: "Private correspondence, held as the original or as a later transcript.", usedBy: 0),
        CatalogSourceType(id: "t4", key: "family-scrapbook", origin: "user", label: "Family scrapbook", description: "Nan's albums — pasted prints, clippings and pencil captions.", usedBy: 3),
        CatalogSourceType(id: "t5", key: "grave-memorial", origin: "plugin:findagrave", label: "Grave memorial", description: "A memorial page assembled from headstone photographs.", usedBy: 5),
    ]
    store.suggestionsByType["t2"] = [
        CatalogTypeSuggestion(field: store.fieldsByProject[projectDir]![0], sortOrder: 0),
        CatalogTypeSuggestion(field: store.fieldsByProject[projectDir]![2], sortOrder: 1),
        CatalogTypeSuggestion(field: store.fieldsByProject[projectDir]![1], sortOrder: 2),
    ]
    store.suggestionsByType["t4"] = [
        CatalogTypeSuggestion(field: store.fieldsByProject[projectDir]![3], sortOrder: 0),
    ]
    return SourceTypesView(
        projectDir: projectDir,
        userID: "00000000-0000-7000-8000-000000000001",
        store: store
    )
    .environment(WorkspaceNavigation())
    .frame(width: 1180, height: 760)
}
#endif
