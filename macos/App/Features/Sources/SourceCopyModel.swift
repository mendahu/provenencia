import Foundation
import Observation

/// What the Copy Source form lets the researcher choose. One value, so "did
/// anything change" is one comparison against the prefill.
struct SourceCopyChoices: Equatable {
    var draft: SourcesModel.Draft
    var includeDescription = true
    var includeMetadata = true
    var includeCredibility = true
    var excludedArtifactIDs: Set<String> = []
}

/// State of the Copy Source form: prefill from the original, which parts the
/// original has, what the researcher kept, and the copy itself. Both hosts (the
/// Source page and, later, the Sources list) present the same model through
/// `View.sourceCopyDialog`.
@MainActor
@Observable
final class SourceCopyModel {
    let original: CatalogSource
    var choices: SourceCopyChoices
    let prefill: SourceCopyChoices
    var titleError: String?
    var typeError: String?
    var copyError: String?
    private(set) var isSaving = false

    private let session: WorkspaceSession
    private let userID: String
    private let store: any GenealogyStore
    private let workspaceKey: CatalogQueryKey
    private let workspaceHandle: QueryHandle<CatalogSourceWorkspace>
    private let typesHandle: QueryHandle<[CatalogSourceType]>

    init(
        original: CatalogSource,
        session: WorkspaceSession,
        userID: String,
        store: any GenealogyStore
    ) {
        self.original = original
        self.session = session
        self.userID = userID
        self.store = store
        let key = CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: original.id)
        workspaceKey = key
        workspaceHandle = session.query(key)
        typesHandle = session.query(.sourceTypesList(project: session.projectKey))
        let prefill = SourceCopyChoices(
            draft: SourcesModel.Draft(
                sourceTypeID: original.sourceTypeID,
                title: L10n.Sources.copySourceTitlePrefill(original.title),
                description: ""
            )
        )
        self.prefill = prefill
        choices = prefill
    }

    // MARK: Reads

    var workspace: CatalogSourceWorkspace? { workspaceHandle.value }

    /// Project folder for resolving `objects/…` thumbnail paths.
    var projectDir: String { session.projectKey.projectDir }

    var types: [CatalogSourceType] { typesHandle.value ?? [] }

    /// The original's type icon, for Artifact thumbnails that have no file.
    var sourceTypeIconKey: String? {
        types.first { $0.id == original.sourceTypeID }?.iconKey
    }

    /// A value wins: a failed refetch keeps the stale parts and stays `.ready`.
    var loadState: SourceFormLoadState {
        if workspaceHandle.value != nil { return .ready }
        if workspaceHandle.status == .error {
            let message = workspaceHandle.error.map { L10n.Errors.message(for: $0) }
            return .failed(message ?? L10n.string(L10n.Errors.unknown))
        }
        return .loading
    }

    var descriptionAvailable: Bool {
        guard let workspace else { return false }
        return !workspace.source.description.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    var metadataAvailable: Bool {
        guard let workspace else { return false }
        return !workspace.metadata.isEmpty
    }

    var credibilityAvailable: Bool { workspace?.credibility != nil }

    /// The original's Artifacts in workspace order, minus the unchecked ones. A
    /// deleted Artifact drops out and a new one starts checked.
    var artifactIDsToCopy: [String] {
        PVArtifactChecklistState.checkedIDs(
            ids: (workspace?.artifacts ?? []).map(\.id),
            excluded: choices.excludedArtifactIDs,
            locked: []
        )
    }

    var canCopy: Bool {
        loadState == .ready && !isSaving && choices.draft.validation.isValid
    }

    var isDirty: Bool { choices != prefill }

    // MARK: Actions

    /// Reads the original's parts again, in every state.
    func reloadWorkspace() {
        session.invalidate(workspaceKey)
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
    }

    /// On success returns the copy for the host to open. On failure the form
    /// keeps every choice.
    func copy() async -> CatalogSource? {
        guard !isSaving else { return nil }
        let validation = choices.draft.validation
        titleError = validation.titleError
        typeError = validation.typeError
        guard validation.isValid else { return nil }

        isSaving = true
        defer { isSaving = false }
        copyError = nil
        do {
            let (copy, effects) = try await store.copySource(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                sourceID: original.id,
                title: choices.draft.trimmedTitle,
                sourceTypeID: choices.draft.sourceTypeID,
                copyDescription: choices.includeDescription && descriptionAvailable,
                copyMetadata: choices.includeMetadata && metadataAvailable,
                copyCredibility: choices.includeCredibility && credibilityAvailable,
                artifactIDs: artifactIDsToCopy
            )
            session.apply(effects, .createdSource)
            return copy
        } catch {
            copyError = L10n.Errors.message(for: error)
            return nil
        }
    }
}
