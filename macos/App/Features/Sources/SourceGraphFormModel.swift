import Foundation
import Observation

/// What the Evidence graph form lets the researcher choose. One value, so
/// "did anything change" is one comparison against the prefill. State is the
/// unchecked Artifacts, so an Artifact added while the form is open starts
/// checked.
struct SourceGraphFormDraft: Equatable {
    var name = ""
    var description = ""
    var excludedArtifactIDs: Set<String> = []
}

/// State of the Evidence graph form, for one new graph or one existing one:
/// the draft, the Artifacts the engine says the graph must keep linked (read
/// when an edit opens), and the one write that saves it. The Source page presents
/// it through `View.sourceGraphFormDialog`.
@MainActor
@Observable
final class SourceGraphFormModel {
    enum Mode {
        case create
        case edit(CatalogEvidenceGraph)
    }

    let mode: Mode
    var draft: SourceGraphFormDraft
    let prefill: SourceGraphFormDraft
    var nameError: String?
    var saveError: String?
    private(set) var isSaving = false
    private(set) var loadState: SourceFormLoadState
    /// The linked Artifacts the engine refuses to unlink, with why.
    private(set) var blocks: [CatalogEvidenceGraphUnlinkBlock] = []

    private let context: SourcePageContext

    init(context: SourcePageContext, mode: Mode) {
        self.context = context
        self.mode = mode
        let artifactIDs = (context.workspace?.artifacts ?? []).map(\.id)
        switch mode {
        case .create:
            // A new graph starts with the Artifacts in no graph checked; when
            // the session cannot say, or every Artifact is in a graph, all are.
            var excluded: Set<String> = []
            if let none = context.session.artifactIDsInNoGraph(
                sourceID: context.sourceID, artifactIDs: artifactIDs
            ), !none.isEmpty {
                excluded = Set(artifactIDs.filter { !none.contains($0) })
            }
            prefill = SourceGraphFormDraft(excludedArtifactIDs: excluded)
            loadState = .ready
        case .edit(let graph):
            let linked = Set(graph.artifactIDs)
            prefill = SourceGraphFormDraft(
                name: graph.name,
                description: graph.description,
                excludedArtifactIDs: Set(artifactIDs.filter { !linked.contains($0) })
            )
            loadState = graph.hasArtifact ? .loading : .ready
        }
        draft = prefill
    }

    // MARK: Reads

    var graph: CatalogEvidenceGraph? {
        if case .edit(let graph) = mode { return graph }
        return nil
    }

    /// The Source's Artifacts in page order.
    var artifacts: [CatalogArtifact] { context.workspace?.artifacts ?? [] }

    var projectDir: String { context.projectDir }

    /// The Source's type icon, for Artifact thumbnails that have no file.
    var sourceTypeIconKey: String? {
        context.sourceTypeIconKey
    }

    /// Artifact id to the hint under its locked row; see `lockHints(by:)` in
    /// `SourceArtifactGraphLinks.swift`.
    var locked: [String: String] { blocks.lockHints(by: \.artifactID) }

    /// The locked Artifacts that Citations need, with those Citations, in page
    /// order: the Kept linked group lists Citations, so a lock with none (the
    /// graph's last Artifact) shows only on its row.
    var keptLinked: [(artifact: CatalogArtifact, block: CatalogEvidenceGraphUnlinkBlock)] {
        artifacts.compactMap { artifact in
            blocks.first { $0.artifactID == artifact.id && $0.citations.total > 0 }
                .map { (artifact, $0) }
        }
    }

    /// The ids a save sends: the checked Artifacts in page order.
    var checkedArtifactIDs: [String] {
        PVChecklistState.checkedIDs(
            ids: artifacts.map(\.id),
            excluded: draft.excludedArtifactIDs,
            locked: Set(locked.keys)
        )
    }

    /// Every Artifact is unchecked on a Source that has some.
    var needsArtifact: Bool { !artifacts.isEmpty && checkedArtifactIDs.isEmpty }

    var canSave: Bool { loadState == .ready && !isSaving && !needsArtifact }

    var isDirty: Bool { draft != prefill }

    // MARK: Actions

    /// Asks the engine which of the graph's linked Artifacts it will not
    /// unlink. One call covers the whole form. A graph with no links has nothing
    /// to read. Read again calls this too.
    func loadKeptLinks() async {
        guard let graph, graph.hasArtifact else { return }
        loadState = .loading
        do {
            let read = try await context.store.getEvidenceGraphUnlinkImpact(
                projectDir: context.projectDir,
                pairs: graph.artifactIDs.map {
                    CatalogEvidenceGraphArtifactPair(graphID: graph.id, artifactID: $0)
                }
            )
            blocks = read
            loadState = .ready
        } catch {
            loadState = .failed(L10n.Errors.message(for: error))
        }
    }

    /// One write. On success returns true for the host to close the form; on
    /// failure the form keeps every choice.
    func save() async -> Bool {
        guard !isSaving else { return false }
        let name = draft.name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !name.isEmpty else {
            nameError = L10n.string(L10n.Sources.graphFormNameRequired)
            return false
        }
        guard canSave else { return false }

        isSaving = true
        defer { isSaving = false }
        saveError = nil
        let description = draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
        do {
            switch mode {
            case .create:
                let (graph, effects) = try await context.store.createEvidenceGraph(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    sourceID: context.sourceID,
                    name: name,
                    description: description,
                    artifactIDs: checkedArtifactIDs
                )
                context.session.apply(effects)
                context.toast = VocabularyToast(
                    title: L10n.string(L10n.Sources.toastGraphAddedTitle),
                    body: L10n.Sources.toastGraphAddedBody(ref: graph.ref),
                    tone: .success
                )
            case .edit(let graph):
                // Empty ids (a Source with no Artifacts) leave the links unchanged.
                let effects = try await context.store.updateEvidenceGraph(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    graphID: graph.id,
                    name: name,
                    description: description,
                    artifactIDs: checkedArtifactIDs
                )
                context.session.apply(effects)
            }
            return true
        } catch {
            let message = L10n.Sources.graphFormSaveErrorMessage(L10n.Errors.message(for: error))
            saveError = message
            // Any refusal of an edit: the engine's locks may have moved since
            // the form opened (a Citation took an Artifact), so read them again
            // and keep the refusal in view.
            if let graph, graph.hasArtifact {
                await loadKeptLinks()
                saveError = message
            }
            return false
        }
    }
}
