import Foundation
import Observation

/// Thin citation composer (S7-08): Artifact pick + citation fields + Observations → submit.
@MainActor
@Observable
final class CitationComposerModel {
    enum Phase: Equatable {
        case loading
        case noArtifacts
        case pickArtifact
        case compose
        case subjectMissing
    }

    struct ObservationRow: Identifiable, Equatable {
        var id: UUID
        var propertyID: String
        var polarity: String
        var valueText: String
        var valueIntegerText: String
        var valueTermID: String
        var dateDraft: DateValueDraft

        static func empty() -> ObservationRow {
            ObservationRow(
                id: UUID(),
                propertyID: "",
                polarity: "positive",
                valueText: "",
                valueIntegerText: "",
                valueTermID: "",
                dateDraft: .empty()
            )
        }
    }

    /// Value types the thin composer can edit (name / subject wait for later PRs).
    static let supportedValueTypes: Set<String> = ["text", "integer", "date", "term"]

    /// Valid locator until S7-07 draws real selectors. Engine refuses empty `selectors`.
    static let placeholderLocatorJSON =
        #"{"version":1,"selectors":[{"type":"page","artifact_page":1}]}"#

    let sourceID: String
    let subjectID: String
    private let userID: String
    private let store: any GenealogyStore
    let session: WorkspaceSession

    private(set) var phase: Phase = .loading
    private(set) var subjectLabel: String = ""
    private(set) var subjectTypeKey: String = ""
    private(set) var subjectTypeID: String = ""
    private(set) var artifacts: [CatalogArtifact] = []
    private(set) var selectedArtifactID: String?
    private(set) var availableProperties: [CatalogProperty] = []
    private(set) var termsByPropertyID: [String: [CatalogPropertyTerm]] = [:]
    private(set) var isSubmitting = false
    private(set) var didSubmit = false

    var transcription = ""
    var transcriptionUncertain = false
    var transcriptionNote = ""
    var citationDescription = ""
    var observations: [ObservationRow] = []
    var formError: String?
    var submitAttempted = false

    /// When true, host should `go(to:)` the Evidence graph (subject gone).
    private(set) var shouldFallbackToGraph = false

    init(
        sourceID: String,
        subjectID: String,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.sourceID = sourceID
        self.subjectID = subjectID
        self.session = session
        self.store = store
        self.userID = userID
    }

    var graphKey: CatalogQueryKey {
        .sourceGraph(project: session.projectKey, sourceId: sourceID)
    }

    var workspaceKey: CatalogQueryKey {
        .sourceWorkspace(project: session.projectKey, sourceId: sourceID)
    }

    var selectedArtifact: CatalogArtifact? {
        guard let id = selectedArtifactID else { return nil }
        return artifacts.first { $0.id == id }
    }

    var propertyOptions: [PVComboBoxOption] {
        availableProperties.map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    func catalogProperty(id: String) -> CatalogProperty? {
        availableProperties.first { $0.id == id }
    }

    func termOptions(for propertyID: String) -> [PVComboBoxOption] {
        (termsByPropertyID[propertyID] ?? []).map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    /// Connect-edge Property keys for a bridge type (S7-D4 §2.3).
    static func excludedEdgePropertyKeys(
        typeKey: String,
        rules: [CatalogConnectRule]
    ) -> Set<String> {
        Set(
            rules
                .filter { $0.bridgeTypeKey == typeKey && !$0.refuse }
                .flatMap(\.edgePropertyKeys)
        )
    }

    func warmQueries() {
        let _: QueryHandle<SourceGraphSnapshot> = session.query(graphKey)
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
    }

    func prepare() async {
        phase = .loading
        formError = nil
        shouldFallbackToGraph = false
        warmQueries()

        do {
            let projectDir = session.projectKey.projectDir
            let subjects = try await store.listSubjects(
                projectDir: projectDir,
                sourceID: sourceID
            )
            let positions = try await store.listSubjectPositions(
                projectDir: projectDir,
                sourceID: sourceID
            )
            async let workspaceLoad = store.getSourceWorkspace(
                projectDir: projectDir,
                sourceID: sourceID
            )
            async let rulesLoad = store.listConnectRules()
            async let typesLoad = store.listSubjectTypes(projectDir: projectDir)
            async let observationsLoad = store.listObservationsBySource(
                projectDir: projectDir,
                sourceID: sourceID
            )

            let types = try await typesLoad
            let observations = try await observationsLoad
            let snapshot = SourceGraphSnapshot.build(
                sourceId: sourceID,
                subjects: subjects,
                positions: positions,
                types: types,
                observations: observations
            )
            let workspace = try await workspaceLoad
            let rules = try await rulesLoad

            guard let resolved = Self.resolveSubject(
                id: subjectID,
                in: snapshot,
                types: types
            ) else {
                phase = .subjectMissing
                shouldFallbackToGraph = true
                return
            }

            subjectLabel = resolved.label
            subjectTypeKey = resolved.typeKey
            subjectTypeID = resolved.typeID
            artifacts = workspace.artifacts

            let fields = try await store.listSubjectTypeFields(
                projectDir: projectDir,
                subjectTypeID: resolved.typeID
            )
            let excluded = Self.excludedEdgePropertyKeys(
                typeKey: resolved.typeKey,
                rules: rules
            )
            availableProperties = fields
                .map(\.property)
                .filter {
                    Self.supportedValueTypes.contains($0.valueType)
                        && !excluded.contains($0.key)
                }
                .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }

            for property in availableProperties where property.valueType == "term" {
                termsByPropertyID[property.id] = try await store.listPropertyTerms(
                    projectDir: projectDir,
                    propertyID: property.id
                )
            }

            if artifacts.isEmpty {
                phase = .noArtifacts
                return
            }
            if artifacts.count == 1 {
                selectedArtifactID = artifacts[0].id
                ensureStarterObservation()
                phase = .compose
                return
            }
            if selectedArtifactID == nil {
                phase = .pickArtifact
            } else {
                ensureStarterObservation()
                phase = .compose
            }
        } catch {
            formError = L10n.Errors.message(for: error)
            phase = .noArtifacts
        }
    }

    func selectArtifact(_ id: String) {
        selectedArtifactID = id
        ensureStarterObservation()
        phase = .compose
        formError = nil
        submitAttempted = false
    }

    func changeArtifact() {
        guard artifacts.count > 1 else { return }
        phase = .pickArtifact
        formError = nil
        submitAttempted = false
    }

    func addObservation() {
        observations.append(.empty())
        formError = nil
    }

    func removeObservation(id: UUID) {
        observations.removeAll { $0.id == id }
        formError = nil
    }

    func updateObservation(_ row: ObservationRow) {
        guard let index = observations.firstIndex(where: { $0.id == row.id }) else { return }
        var next = row
        let previousID = observations[index].propertyID
        if previousID != row.propertyID,
           let previousType = catalogProperty(id: previousID)?.valueType,
           previousType != catalogProperty(id: row.propertyID)?.valueType
        {
            next.valueText = ""
            next.valueIntegerText = ""
            next.valueTermID = ""
            next.dateDraft = .empty()
        }
        observations[index] = next
        formError = nil
    }

    func createCustomTerm(propertyID: String, label: String) async -> CatalogPropertyTerm? {
        let trimmed = label.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }
        do {
            let term = try await store.createPropertyTerm(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                propertyID: propertyID,
                label: trimmed,
                description: ""
            )
            var list = termsByPropertyID[propertyID] ?? []
            list.append(term)
            list.sort { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
            termsByPropertyID[propertyID] = list
            return term
        } catch {
            formError = L10n.Errors.message(for: error)
            return nil
        }
    }

    /// Returns graph location after success; nil when validation/submit failed.
    func submit() async -> WorkspaceLocation? {
        submitAttempted = true
        formError = nil
        guard let artifactID = selectedArtifactID else {
            formError = String(localized: L10n.CitationComposer.needArtifact)
            return nil
        }
        guard let drafts = buildObservationDrafts() else { return nil }

        isSubmitting = true
        defer { isSubmitting = false }
        do {
            _ = try await store.createCitationWithObservations(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                artifactID: artifactID,
                locatorJSON: Self.placeholderLocatorJSON,
                transcription: transcription,
                description: citationDescription,
                transcriptionUncertain: transcriptionUncertain,
                transcriptionNote: transcriptionNote,
                citationNotes: [],
                observations: drafts
            )
            session.apply(.createdCitation(sourceId: sourceID))
            didSubmit = true
            return WorkspaceLocation(
                section: .sources,
                sourceId: sourceID,
                sourceSurface: .graph,
                title: nil
            )
        } catch {
            formError = L10n.Errors.message(for: error)
            return nil
        }
    }

    func graphLocation() -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: sourceID,
            sourceSurface: .graph
        )
    }

    // MARK: - Validation

    private func ensureStarterObservation() {
        if observations.isEmpty {
            observations = [.empty()]
        }
    }

    private func buildObservationDrafts() -> [CatalogObservationDraft]? {
        if observations.isEmpty {
            formError = String(localized: L10n.CitationComposer.noObservationsError)
            return nil
        }
        var drafts: [CatalogObservationDraft] = []
        drafts.reserveCapacity(observations.count)
        for row in observations {
            guard let property = catalogProperty(id: row.propertyID) else {
                formError = String(localized: L10n.CitationComposer.missingPropertyError)
                return nil
            }
            var draft = CatalogObservationDraft(
                subjectID: subjectID,
                propertyID: property.id,
                polarity: row.polarity == "negative" ? "negative" : "positive"
            )
            switch property.valueType {
            case "text":
                let text = row.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
                guard !text.isEmpty else {
                    formError = String(localized: L10n.CitationComposer.missingValueError)
                    return nil
                }
                draft.valueText = text
            case "integer":
                let trimmed = row.valueIntegerText.trimmingCharacters(in: .whitespacesAndNewlines)
                guard let value = Int64(trimmed) else {
                    formError = String(localized: L10n.CitationComposer.invalidIntegerError)
                    return nil
                }
                draft.valueInteger = value
            case "term":
                guard !row.valueTermID.isEmpty else {
                    formError = String(localized: L10n.CitationComposer.missingValueError)
                    return nil
                }
                draft.valueTermID = row.valueTermID
            case "date":
                guard row.dateDraft.isValid else {
                    formError = String(localized: L10n.CitationComposer.invalidDateError)
                    return nil
                }
                draft.date = row.dateDraft.toInput()
            default:
                formError = String(localized: L10n.CitationComposer.unsupportedValueTypeError)
                return nil
            }
            drafts.append(draft)
        }
        return drafts
    }

    private struct ResolvedSubject {
        var label: String
        var typeKey: String
        var typeID: String
    }

    private static func resolveSubject(
        id: String,
        in snapshot: SourceGraphSnapshot,
        types: [CatalogSubjectType]
    ) -> ResolvedSubject? {
        if let primary = snapshot.subjects.first(where: { $0.id == id }) {
            return ResolvedSubject(
                label: primary.subject.label,
                typeKey: primary.kind.rawValue,
                typeID: primary.subject.subjectTypeID
            )
        }
        if let bridge = snapshot.bridges.first(where: { $0.id == id }) {
            return ResolvedSubject(
                label: bridge.subject.label,
                typeKey: bridge.kind.rawValue,
                typeID: bridge.subject.subjectTypeID
            )
        }
        // Unplaced / type source — still allow cite if present in types+subjects via store path.
        // Snapshot omits unplaced; treat as missing for thin composer entry from graph cards.
        _ = types
        return nil
    }
}
