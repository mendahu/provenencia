import Foundation

/// What a metadata board needs from the place that holds one owner's rows: the
/// rows, the owner's writes, the field vocabulary, and where its failures show.
/// A page instances one host per owner; the board never learns which page it is on.
@MainActor
protocol MetadataBoardHost: AnyObject {
    /// The owner's whole board in display order. The board patches it after a write.
    var entries: [CatalogMetadataEntry] { get set }
    /// Every metadata field the researcher can add.
    var fields: [CatalogMetadataField] { get }

    /// Saves a value; answers the owner's whole board.
    func set(fieldID: String, valueText: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)
    /// Clears a value; answers the owner's whole board.
    func clear(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)
    /// Permanently hides an unfilled suggestion; answers the owner's whole board.
    func dismissSuggestion(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)
    /// Saves the display order; answers the owner's whole board.
    func reorder(fieldIDs: [String]) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)

    /// Applies a write's effects once the board has patched `entries`.
    func didWrite(_ effects: CatalogWriteEffects)
    /// Clears the failure a previous write showed, before a new write.
    func clearFailure()
    /// The message for a failed write; nil means the owner is gone and its host is leaving, so no state is written.
    func failureMessage(for error: Error) -> String?
    /// Shows a failure message where this host's failures show.
    func showFailure(_ message: String)
}

/// A host over the Source page: one owner's rows inside the page's workspace,
/// its writes through the page's store, and its failures in the page banner.
@MainActor
final class SourcePageMetadataHost: MetadataBoardHost {
    private let context: SourcePageContext
    private let owner: MetadataOwner
    private let rows: (CatalogSourceWorkspace) -> [CatalogMetadataEntry]
    private let setRows: (inout CatalogSourceWorkspace, [CatalogMetadataEntry]) -> Void

    init(
        context: SourcePageContext,
        owner: MetadataOwner,
        rows: @escaping (CatalogSourceWorkspace) -> [CatalogMetadataEntry],
        setRows: @escaping (inout CatalogSourceWorkspace, [CatalogMetadataEntry]) -> Void
    ) {
        self.context = context
        self.owner = owner
        self.rows = rows
        self.setRows = setRows
    }

    /// The host for the Source's own board.
    static func source(context: SourcePageContext) -> SourcePageMetadataHost {
        SourcePageMetadataHost(
            context: context,
            owner: MetadataOwner(level: .source, id: context.sourceID),
            rows: { $0.metadata },
            setRows: { $0.metadata = $1 }
        )
    }

    var entries: [CatalogMetadataEntry] {
        get { context.workspace.map(rows) ?? [] }
        set {
            guard var workspace = context.workspace else { return }
            setRows(&workspace, newValue)
            context.workspace = workspace
        }
    }

    var fields: [CatalogMetadataField] { context.fields }

    func set(fieldID: String, valueText: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        try await context.store.setMetadata(
            projectDir: context.projectDir, userID: context.userID, owner: owner,
            fieldID: fieldID, valueText: valueText
        )
    }

    func clear(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        try await context.store.clearMetadata(
            projectDir: context.projectDir, userID: context.userID, owner: owner, fieldID: fieldID
        )
    }

    func dismissSuggestion(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        try await context.store.dismissMetadataSuggestion(
            projectDir: context.projectDir, userID: context.userID, owner: owner, fieldID: fieldID
        )
    }

    func reorder(fieldIDs: [String]) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        try await context.store.reorderMetadata(
            projectDir: context.projectDir, userID: context.userID, owner: owner, fieldIDs: fieldIDs
        )
    }

    func didWrite(_ effects: CatalogWriteEffects) { context.applyWrite(effects) }
    func clearFailure() { context.clearPageError() }
    func failureMessage(for error: Error) -> String? { context.fail(error) }
    func showFailure(_ message: String) { context.showPageError(message) }
}
