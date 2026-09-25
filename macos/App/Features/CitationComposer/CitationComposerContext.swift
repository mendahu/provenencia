import Foundation
import Observation

/// Shared write surface and memoized catalog snapshot for composer children.
@MainActor
@Observable
final class CitationComposerContext {
    private struct MemoKey: Equatable {
        var fields: SubjectFieldsSnapshot
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
    var fieldsKey: CatalogQueryKey { .subjectFieldsWorkspace(project: session.projectKey) }
    var connectRulesKey: CatalogQueryKey { .connectRules(project: session.projectKey) }

    var fieldsSnapshot: SubjectFieldsSnapshot {
        let handle: QueryHandle<SubjectFieldsSnapshot>? = session.queryHandle(fieldsKey)
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
