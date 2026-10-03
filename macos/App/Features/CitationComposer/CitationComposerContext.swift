import Foundation
import Observation

/// Shared write surface and memoized catalog snapshot for composer children.
@MainActor
@Observable
final class CitationComposerContext {
    private struct MemoKey: Equatable {
        var fields: PropertiesSnapshot
        var rows: SourceGraphRows
        var rules: [CatalogConnectRule]
        var terms: [String: [CatalogPropertyTerm]]
    }

    let store: any GenealogyStore
    let session: WorkspaceSession
    let userID: String
    let sourceID: String

    var citationID: String?
    var artifactID: String?
    weak var fields: CitationFieldsDraft?

    @ObservationIgnored
    private var memoKey: MemoKey?
    @ObservationIgnored
    private var memoSnapshot: SourceGraphSnapshot
    @ObservationIgnored
    private var memoVocabulary = CitationComposerVocabulary.empty
    @ObservationIgnored
    private var isCreatingCitation = false
    @ObservationIgnored
    private var citationCreateWaiters: [CheckedContinuation<String, Error>] = []

    init(
        store: any GenealogyStore,
        session: WorkspaceSession,
        userID: String,
        sourceID: String
    ) {
        self.store = store
        self.session = session
        self.userID = userID
        self.sourceID = sourceID
        memoSnapshot = SourceGraphSnapshot(sourceId: sourceID)
    }

    var projectDir: String { session.projectKey.projectDir }
    var graphKey: CatalogQueryKey { .sourceGraph(project: session.projectKey, sourceId: sourceID) }
    var fieldsKey: CatalogQueryKey { .propertiesWorkspace(project: session.projectKey) }
    var connectRulesKey: CatalogQueryKey { .connectRules(project: session.projectKey) }

    var fieldsSnapshot: PropertiesSnapshot {
        let handle: QueryHandle<PropertiesSnapshot>? = session.queryHandle(fieldsKey)
        return handle?.value ?? .empty
    }

    var graphSnapshot: SourceGraphSnapshot {
        refreshMemo()
        return memoSnapshot
    }

    var vocabulary: CitationComposerVocabulary {
        refreshMemo()
        return memoVocabulary
    }

    func applySavedCitation() {
        session.apply(.savedCitation(sourceId: sourceID))
    }

    /// One in-flight create so citation Save, Observation commit, and Connect
    /// cannot mint two citations from a blank document.
    func ensureCitationID() async throws -> String {
        if let citationID, !citationID.isEmpty {
            return citationID
        }
        if isCreatingCitation {
            return try await withCheckedThrowingContinuation { continuation in
                citationCreateWaiters.append(continuation)
            }
        }
        isCreatingCitation = true
        do {
            let id = try await createCitationIfNeeded()
            isCreatingCitation = false
            let waiters = citationCreateWaiters
            citationCreateWaiters = []
            for waiter in waiters {
                waiter.resume(returning: id)
            }
            return id
        } catch {
            isCreatingCitation = false
            let waiters = citationCreateWaiters
            citationCreateWaiters = []
            for waiter in waiters {
                waiter.resume(throwing: error)
            }
            throw error
        }
    }

    private func createCitationIfNeeded() async throws -> String {
        if let citationID, !citationID.isEmpty {
            return citationID
        }
        guard let artifactID, !artifactID.isEmpty else {
            throw CitationComposerNeedArtifact()
        }
        let values = citationValues()
        let created = try await store.createCitationWithObservations(
            projectDir: projectDir,
            userID: userID,
            artifactID: artifactID,
            locatorJSON: values.locator.encodeJSON(),
            transcription: values.transcription,
            description: values.description,
            transcriptionUncertain: values.transcriptionUncertain,
            transcriptionNote: values.transcriptionNote,
            citationNotes: [],
            observations: []
        )
        citationID = created.0.id
        captureCitationBaseline()
        applySavedCitation()
        await reloadListedCitations()
        return created.0.id
    }

    func reloadListedCitations() async {
        guard let artifactID else { return }
        let key = CatalogQueryKey.citationsByArtifact(
            project: session.projectKey,
            artifactId: artifactID
        )
        let _: QueryHandle<[CatalogListedCitation]> = session.query(key)
        _ = await session.readyValue(key) as [CatalogListedCitation]?
    }

    func waitForGraph() async {
        let _: QueryHandle<SourceGraphRows> = session.query(graphKey)
        _ = await session.readyValue(graphKey) as SourceGraphRows?
    }

    func citationValues() -> CitationFieldsDraft.Values {
        fields?.values ?? CitationFieldsDraft.Values()
    }

    func captureCitationBaseline() {
        fields?.captureBaseline()
    }

    private func termsKey(propertyID: String) -> CatalogQueryKey {
        .propertyTerms(project: session.projectKey, propertyId: propertyID)
    }

    private func refreshMemo() {
        let fields = fieldsSnapshot
        let rowsHandle: QueryHandle<SourceGraphRows>? = session.queryHandle(graphKey)
        let rows = rowsHandle?.value ?? SourceGraphRows(sourceId: sourceID)
        let rulesHandle: QueryHandle<[CatalogConnectRule]>? = session.queryHandle(connectRulesKey)
        let rules = rulesHandle?.value ?? []
        var terms: [String: [CatalogPropertyTerm]] = [:]
        for property in fields.properties
            where property.valueType == PropertyValueType.term.rawValue
        {
            let handle: QueryHandle<[CatalogPropertyTerm]>? = session.queryHandle(
                termsKey(propertyID: property.id)
            )
            if let value = handle?.value {
                terms[property.id] = value
            }
        }
        let key = MemoKey(fields: fields, rows: rows, rules: rules, terms: terms)
        if key == memoKey { return }
        memoKey = key
        memoSnapshot = SourceGraphSnapshot.build(rows: rows, types: fields.types, rules: rules)
        memoVocabulary = CitationComposerVocabulary.make(
            fields: fields,
            snapshot: memoSnapshot,
            rules: rules,
            termsByPropertyID: terms
        )
    }
}

struct CitationComposerNeedArtifact: Error {}
