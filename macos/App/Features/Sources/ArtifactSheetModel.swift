import Foundation
import Observation

/// What the Artifact sheet's text fields hold.
struct ArtifactSheetDraft: Equatable {
    var label = ""
    var description = ""
}

/// The Artifact sheet: one Artifact's label, description, Evidence graph links
/// and copy details, edited together and written by one Save. Nothing is
/// written until Save; Cancel drops the model with everything in it.
@MainActor
@Observable
final class ArtifactSheetModel {
    enum Mode {
        case edit(CatalogArtifact)
    }

    /// Which part of the sheet its body opens at the top of.
    enum Section: Hashable {
        case top
        case graphs
        case copyDetails
    }

    let mode: Mode
    let opensAt: Section
    /// The label the sheet opened with, for the notice when the Artifact goes.
    let openedLabel: String

    var draft: ArtifactSheetDraft
    private let prefill: ArtifactSheetDraft
    var labelError: String?
    /// Why the last Save stopped, shown at the top of the sheet; nil otherwise.
    var saveError: String?
    private(set) var isSaving = false
    /// The section the scrolling body is asked to scroll to: where the sheet
    /// opened (nil at the top), then the top when a Save stops with a message.
    private(set) var scrollTarget: Section?
    /// Changes each time a Save asks the body to scroll to `scrollTarget` again.
    private(set) var scrollRequest = 0

    // MARK: Evidence graph links

    /// The Source's graphs when the sheet opened (or when the list first
    /// arrived), in the session's name order; nil while the session has none.
    private(set) var graphIDs: [String]?
    /// The unchecked graph ids, seeded with the graphs that do not link the Artifact.
    var excludedGraphIDs: Set<String> = []
    /// The engine's refusals to unlink this Artifact, from its preflight. Each
    /// locks its graph; Swift decides no lock.
    private let locks = UnlinkLocks()
    var linkLoadState: SourceFormLoadState { locks.state }
    /// The graph link read, so a test can await it.
    @ObservationIgnored var linkLoadTask: Task<Void, Never>?

    // MARK: Copy details

    let host: MetadataDraftHost
    let board: MetadataBoardModel

    private let context: SourcePageContext
    private let onSaved: @MainActor (CatalogArtifact, [CatalogMetadataEntry]?, CatalogWriteEffects) -> Void

    init(
        context: SourcePageContext,
        mode: Mode,
        opensAt: Section,
        onSaved: @escaping @MainActor (CatalogArtifact, [CatalogMetadataEntry]?, CatalogWriteEffects) -> Void
    ) {
        self.context = context
        self.mode = mode
        self.opensAt = opensAt
        scrollTarget = opensAt == .top ? nil : opensAt
        self.onSaved = onSaved
        switch mode {
        case .edit(let artifact):
            openedLabel = artifact.label
            prefill = ArtifactSheetDraft(label: artifact.label, description: artifact.description)
            let rows = context.workspace?.metadataRows(for: MetadataOwner(level: .artifact, id: artifact.id)) ?? []
            host = MetadataDraftHost(entries: rows, fields: context.fields)
        }
        draft = prefill
        board = MetadataBoardModel(host: host)
        if visibleGraphs != nil { graphsChanged() }
    }

    /// "SRC-4KD2P · title" under the sheet's title.
    var subtitle: String {
        guard let source = context.source else { return "" }
        return "\(source.ref) · \(source.title)"
    }

    var artifact: CatalogArtifact {
        switch mode {
        case .edit(let artifact): artifact
        }
    }

    // MARK: Graphs

    /// The Source's Evidence graphs when the session has them; nil otherwise,
    /// so no group shows.
    var visibleGraphs: [CatalogEvidenceGraph]? { context.evidenceGraphs }

    /// True while the graph list has not arrived, so the group shows its loading state.
    var awaitingGraphs: Bool { graphIDs == nil && context.evidenceGraphsLoading }

    /// The graph list changed: a sheet opened before it arrived gets its group
    /// now. Sets the graph choices and starts the lock read, once.
    func graphsChanged() {
        guard graphIDs == nil, let graphs = visibleGraphs else { return }
        graphIDs = graphs.map(\.id)
        excludedGraphIDs = Set(graphs.filter { !$0.links(artifact.id) }.map(\.id))
        linkLoadTask = Task { await loadLinkLocks() }
    }

    /// The graph ids whose session graph links the Artifact now.
    private var linkedGraphIDs: [String] {
        (graphIDs ?? []).filter {
            context.session.evidenceGraph(id: $0)?.links(artifact.id) == true
        }
    }

    /// The graph ids that are still in the session's graph list; a graph
    /// deleted since the sheet opened drops out of every count and write.
    private var liveGraphIDs: [String] {
        (graphIDs ?? []).filter { context.session.evidenceGraph(id: $0) != nil }
    }

    /// Graph id to the hint under its locked row: one lock per graph, from the
    /// first block that names it.
    var lockedGraphs: [String: String] {
        locks.locked(by: \.graphID)
    }

    /// The checked graph ids in list order; a locked graph is always checked.
    var checkedGraphIDs: [String] {
        guard graphIDs != nil else { return [] }
        return PVChecklistState.checkedIDs(
            ids: liveGraphIDs, excluded: excludedGraphIDs, locked: Set(lockedGraphs.keys)
        )
    }

    /// The graphs the sheet offers, read from the session in the order they
    /// were captured; a graph deleted since then drops out.
    var graphs: [CatalogEvidenceGraph] {
        (graphIDs ?? []).compactMap { context.session.evidenceGraph(id: $0) }
    }

    /// The graph boxes are read and none is checked: saving leaves the Artifact in no graph.
    var inNoGraph: Bool {
        graphIDs != nil && linkLoadState == .ready && checkedGraphIDs.isEmpty
    }

    /// A box was checked or unchecked since the sheet opened.
    var linksChanged: Bool {
        guard graphIDs != nil else { return false }
        return Set(checkedGraphIDs) != Set(linkedGraphIDs)
    }

    /// The full set of graph ids a links Save sends: the checked ones in list
    /// order, then the graphs made since the sheet opened that link the Artifact
    /// (it never showed them, so a Save must not drop them).
    private var savedGraphIDs: [String] {
        let shown = Set(graphIDs ?? [])
        let added = (context.session.evidenceGraphs(sourceID: context.sourceID) ?? [])
            .filter { !shown.contains($0.id) && $0.links(artifact.id) }
            .map(\.id)
        return checkedGraphIDs + added
    }

    /// Reads which graphs must keep the Artifact: one unlink preflight over a
    /// pair for every graph that links it. Every block the engine reports locks its graph.
    func loadLinkLocks() async {
        guard let graphIDs else { return }
        let artifactID = artifact.id
        let pairs = graphIDs.compactMap { graphID -> CatalogEvidenceGraphArtifactPair? in
            guard context.session.evidenceGraph(id: graphID)?.links(artifactID) == true else { return nil }
            return CatalogEvidenceGraphArtifactPair(graphID: graphID, artifactID: artifactID)
        }
        await locks.load(
            store: context.store,
            projectDir: context.projectDir,
            pairs: pairs,
            include: { $0.artifactID == artifactID },
            onError: { context.fail($0) }
        )
    }

    // MARK: Dirty and Save

    private var textChanged: Bool {
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let description = draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
        return label != prefill.label || description != prefill.description
    }

    /// Anything differs from what the sheet opened with, or text sits typed and
    /// unconfirmed in the board.
    var isDirty: Bool { textChanged || linksChanged || board.hasUnsavedChanges }

    /// Save is on unless a write runs or the graph lock read has not finished.
    /// The label is checked in `save()`, as the graph form does.
    var canSave: Bool {
        !isSaving && (graphIDs == nil || linkLoadState == .ready)
    }

    /// One Save: one `updateArtifact` with the label, the description, every
    /// link when a box changed and the copy details draft, so the engine applies
    /// all of it or none. True when the sheet may close; on false it stays open
    /// with every field, box and row as the researcher set it.
    func save() async -> Bool {
        guard !isSaving else { return false }
        saveError = nil
        labelError = nil
        if board.hasUnconfirmedInput {
            saveError = L10n.string(L10n.Sources.artifactSheetFinishOpenField)
            scrollToTop()
            return false
        }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let description = draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
        if label.isEmpty {
            labelError = L10n.string(L10n.Sources.artifactLabelRequired)
            scrollToTop()
            return false
        }
        if graphIDs != nil, linkLoadState != .ready { return false }
        let graphIDs: [String]? = linksChanged ? savedGraphIDs : nil
        let metadata = host.draft
        if !textChanged, graphIDs == nil, metadata == nil { return true }
        context.clearPageError()
        isSaving = true
        defer { isSaving = false }
        do {
            let (updated, rows, effects) = try await context.store.updateArtifact(
                projectDir: context.projectDir,
                userID: context.userID,
                artifactID: artifact.id,
                label: label,
                description: description,
                graphIDs: graphIDs,
                metadata: metadata
            )
            onSaved(updated, metadata == nil ? nil : rows, effects)
            return true
        } catch {
            guard let message = context.fail(error) else { return true }
            if board.showRefusal(error) { return false }
            // Read the locks again: a Citation may have taken the Artifact
            // since the sheet opened, and the engine's preflight says which graph keeps it.
            if graphIDs != nil { await loadLinkLocks() }
            saveError = saveErrorBody(message)
            scrollToTop()
            return false
        }
    }

    /// A message sits at the top of the body (the save error, the label's error): ask for it.
    private func scrollToTop() {
        scrollTarget = .top
        scrollRequest += 1
    }

    /// The save-error body. Nothing was written, so it never speaks of a partial
    /// save: it names the graph the engine says keeps the Artifact when the
    /// researcher unchecked it, else says the changes are kept.
    private func saveErrorBody(_ message: String) -> String {
        if linkLoadState == .ready {
            let locked = lockedGraphs
            let kept = (graphIDs ?? []).first { excludedGraphIDs.contains($0) && locked[$0] != nil }
            if let kept, let graph = context.session.evidenceGraph(id: kept) {
                return L10n.Sources.artifactCardSaveErrorGraphKeeps(graph: graph.name, error: message)
            }
        }
        return L10n.Sources.artifactCardSaveErrorKept(message)
    }
}
