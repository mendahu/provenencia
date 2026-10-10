import Foundation
import Observation

/// Shared write surface and memoized catalog snapshot for composer children.
@MainActor
@Observable
final class CitationComposerContext {
    /// Revisions of the queries the memo was built from. Comparing these
    /// avoids walking the source graph on every vocabulary read.
    private struct MemoToken: Equatable {
        var graph: Int
        var fields: Int
        var rules: Int
        var terms: [String: Int]
    }

    let store: any GenealogyStore
    let session: WorkspaceSession
    let userID: String
    let sourceID: String

    var citationID: String?
    var artifactID: String?
    weak var fields: CitationFieldsDraft?

    @ObservationIgnored
    private var memoToken: MemoToken?
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
        let fieldsHandle: QueryHandle<PropertiesSnapshot>? = session.queryHandle(fieldsKey)
        let rowsHandle: QueryHandle<SourceGraphRows>? = session.queryHandle(graphKey)
        let rulesHandle: QueryHandle<[CatalogConnectRule]>? = session.queryHandle(connectRulesKey)
        let graph = rowsHandle?.revision ?? 0
        let fieldsRevision = fieldsHandle?.revision ?? 0
        let rulesRevision = rulesHandle?.revision ?? 0
        if let memoToken,
            memoToken.graph == graph,
            memoToken.fields == fieldsRevision,
            memoToken.rules == rulesRevision,
            termRevisionsMatch(memoToken.terms)
        {
            return
        }
        let fields = fieldsHandle?.value ?? .empty
        let terms = termRevisions(in: fields)
        memoToken = MemoToken(
            graph: graph,
            fields: fieldsRevision,
            rules: rulesRevision,
            terms: terms
        )
        let rows = rowsHandle?.value ?? SourceGraphRows(sourceId: sourceID)
        let rules = rulesHandle?.value ?? []
        var termLists: [String: [CatalogPropertyTerm]] = [:]
        for propertyID in terms.keys {
            let handle: QueryHandle<[CatalogPropertyTerm]>? = session.queryHandle(
                termsKey(propertyID: propertyID)
            )
            if let value = handle?.value {
                termLists[propertyID] = value
            }
        }
        memoSnapshot = SourceGraphSnapshot.build(rows: rows, types: fields.types, rules: rules)
        memoVocabulary = CitationComposerVocabulary.make(
            fields: fields,
            snapshot: memoSnapshot,
            rules: rules,
            termsByPropertyID: termLists
        )
    }

    private func termRevisionsMatch(_ stored: [String: Int]) -> Bool {
        for (propertyID, revision) in stored {
            let handle: QueryHandle<[CatalogPropertyTerm]>? = session.queryHandle(
                termsKey(propertyID: propertyID)
            )
            if (handle?.revision ?? 0) != revision { return false }
        }
        return true
    }

    private func termRevisions(in fields: PropertiesSnapshot) -> [String: Int] {
        var ids = Set(
            fields.properties
                .filter { $0.valueType == PropertyValueType.term.rawValue }
                .map(\.id)
        )
        for typeFields in fields.propertiesByTypeID.values {
            for field in typeFields where field.property.valueType == PropertyValueType.term.rawValue {
                ids.insert(field.property.id)
            }
        }
        var revisions: [String: Int] = [:]
        for propertyID in ids {
            let handle: QueryHandle<[CatalogPropertyTerm]>? = session.queryHandle(
                termsKey(propertyID: propertyID)
            )
            revisions[propertyID] = handle?.revision ?? 0
        }
        return revisions
    }
}

struct CitationComposerNeedArtifact: Error {}
