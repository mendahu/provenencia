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
    /// The Evidence graph the composer works on: its rows, the new Subjects and
    /// the saved Citations all read it.
    let graphID: String

    var citationID: String?
    var artifactID: String?

    @ObservationIgnored
    private var memoToken: MemoToken?
    @ObservationIgnored
    private var memoSnapshot: EvidenceGraphSnapshot
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
        sourceID: String,
        graphID: String
    ) {
        self.store = store
        self.session = session
        self.userID = userID
        self.sourceID = sourceID
        self.graphID = graphID
        memoSnapshot = EvidenceGraphSnapshot(graphId: graphID)
    }

    var projectDir: String { session.projectKey.projectDir }
    var graphKey: CatalogQueryKey { .evidenceGraph(project: session.projectKey, graphId: graphID) }
    var fieldsKey: CatalogQueryKey { .propertiesWorkspace(project: session.projectKey) }
    var connectRulesKey: CatalogQueryKey { .connectRules(project: session.projectKey) }

    var fieldsSnapshot: PropertiesSnapshot {
        let handle: QueryHandle<PropertiesSnapshot>? = session.queryHandle(fieldsKey)
        return handle?.value ?? .empty
    }

    var graphSnapshot: EvidenceGraphSnapshot {
        refreshMemo()
        return memoSnapshot
    }

    var vocabulary: CitationComposerVocabulary {
        refreshMemo()
        return memoVocabulary
    }

    /// The one funnel for composer writes. The location names the selected
    /// Artifact, so its citations key is visible and `apply` revalidates it
    /// from the write's `citations` read; nothing queries afterward.
    func applyWrite(_ effects: CatalogWriteEffects) {
        session.apply(effects)
    }

    /// The Citations of one Artifact as this graph sees them.
    func citationsKey(artifactID: String) -> CatalogQueryKey {
        .citationsByArtifact(project: session.projectKey, graphId: graphID, artifactId: artifactID)
    }

    /// `nil` follows the citation open in the composer.
    func ensureCitationID(_ values: CitationFieldsDraft.Values) async throws -> String {
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
            let id = try await createCitationIfNeeded(values)
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

    private func createCitationIfNeeded(_ values: CitationFieldsDraft.Values) async throws -> String {
        if let citationID, !citationID.isEmpty {
            return citationID
        }
        guard let artifactID, !artifactID.isEmpty else {
            throw CitationComposerNeedArtifact()
        }
        let created = try await store.createCitationWithObservations(
            projectDir: projectDir,
            userID: userID,
            graphID: graphID,
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
        applyWrite(created.2)
        await reloadListedCitations()
        return created.0.id
    }

    func reloadListedCitations() async {
        guard let artifactID else { return }
        let key = citationsKey(artifactID: artifactID)
        let _: QueryHandle<[CatalogListedCitation]> = session.query(key)
        _ = await session.readyValue(key) as [CatalogListedCitation]?
    }

    func waitForGraph() async {
        let _: QueryHandle<EvidenceGraphRows> = session.query(graphKey)
        _ = await session.readyValue(graphKey) as EvidenceGraphRows?
    }

    private func termsKey(propertyID: String) -> CatalogQueryKey {
        .propertyTerms(project: session.projectKey, propertyId: propertyID)
    }

    private func refreshMemo() {
        let fieldsHandle: QueryHandle<PropertiesSnapshot>? = session.queryHandle(fieldsKey)
        let rowsHandle: QueryHandle<EvidenceGraphRows>? = session.queryHandle(graphKey)
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
        let rows = rowsHandle?.value ?? EvidenceGraphRows(graphId: graphID)
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
        memoSnapshot = EvidenceGraphSnapshot.build(rows: rows, types: fields.types, rules: rules)
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
