import Foundation
import Observation

/// Thin citation composer (S7-08): board-aligned viewer chrome | form, dialog Observations.
@MainActor
@Observable
final class CitationComposerModel {
    enum Phase: Equatable {
        case loading
        case pickArtifact
        case compose
        case subjectMissing
    }

    /// Committed Observation summary row (edited via dialog, not inline).
    struct ObservationRow: Identifiable, Equatable {
        var id: UUID
        var propertyID: String
        var polarity: String
        var valueText: String
        var valueIntegerText: String
        var valueTermID: String
        var valueSubjectID: String
        var dateDraft: DateValueDraft
        /// Connect-edge endpoint rows (Frame 10) — system-owned, not editable/removable.
        var isConnectFixed: Bool
    }

    /// Draft for the Add / Edit observation `PVFormDialog` (Frames 3–6, 13).
    struct ObservationDialogState: Equatable {
        var editingID: UUID?
        var propertyID: String
        var polarity: String
        var valueText: String
        var valueIntegerText: String
        var valueTermID: String
        var dateDraft: DateValueDraft
        var showValidation: Bool

        static func fresh() -> ObservationDialogState {
            ObservationDialogState(
                editingID: nil,
                propertyID: "",
                polarity: "positive",
                valueText: "",
                valueIntegerText: "",
                valueTermID: "",
                dateDraft: .empty(),
                showValidation: false
            )
        }

        static func editing(_ row: ObservationRow) -> ObservationDialogState {
            ObservationDialogState(
                editingID: row.id,
                propertyID: row.propertyID,
                polarity: row.polarity,
                valueText: row.valueText,
                valueIntegerText: row.valueIntegerText,
                valueTermID: row.valueTermID,
                dateDraft: row.dateDraft,
                showValidation: false
            )
        }
    }

    /// Value types the thin composer can edit (name / subject wait for later PRs).
    static let supportedValueTypes: Set<String> = ["text", "integer", "date", "term"]

    let sourceID: String
    let subjectID: String
    /// When set, prepare/save load and update this Citation instead of creating.
    let citationID: String?
    private let userID: String
    private let store: any GenealogyStore
    let session: WorkspaceSession
    private let linkStore: any EvidenceProvisionalLinkStoring

    private(set) var phase: Phase = .loading
    private(set) var subjectLabel: String = ""
    private(set) var subjectTypeKey: String = ""
    private(set) var subjectTypeID: String = ""
    private(set) var sourceTitle: String = ""
    /// Source-type `icon_key` for fileless Artifact picker thumbs (Frame 8).
    private(set) var sourceTypeIconKey: String = ""
    private(set) var artifacts: [CatalogArtifact] = []
    private(set) var selectedArtifactID: String?
    /// Frame 7 selection before Continue.
    var pendingArtifactID: String?
    private(set) var availableProperties: [CatalogProperty] = []
    /// Connect-edge Properties (excluded from Add observation) keyed by id.
    private(set) var connectEdgePropertiesByID: [String: CatalogProperty] = [:]
    private(set) var termsByPropertyID: [String: [CatalogPropertyTerm]] = [:]
    private(set) var isSubmitting = false
    private(set) var didSubmit = false

    /// Isolated document viewer (S7-06). Composer syncs page ↔ locator only.
    let artifactViewer = ArtifactViewerModel()

    /// Committed locator page; nil until the researcher sets one (Frame 12).
    private(set) var locatorPage: Int?
    /// Thin image path used Draw region to commit a whole-image page selector.
    private(set) var locatorIsWholeImage = false

    /// 1-based page from the Artifact viewer (PDF).
    var viewerPage: Int { artifactViewer.page }
    var viewerPageCount: Int { artifactViewer.pageCount }

    var transcription = ""
    var transcriptionUncertain = false
    var transcriptionNote = ""
    var citationDescription = ""
    private(set) var observations: [ObservationRow] = []
    var observationDialog: ObservationDialogState?
    var formError: String?
    var locatorError: String?
    var submitAttempted = false

    /// When true, host should `go(to:)` the Evidence graph (subject gone).
    private(set) var shouldFallbackToGraph = false

    var isEditingExisting: Bool { citationID != nil }

    init(
        sourceID: String,
        subjectID: String,
        citationID: String? = nil,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String,
        linkStore: (any EvidenceProvisionalLinkStoring)? = nil
    ) {
        self.sourceID = sourceID
        self.subjectID = subjectID
        self.citationID = citationID
        self.session = session
        self.store = store
        self.userID = userID
        if let linkStore {
            self.linkStore = linkStore
        } else if let live = try? EvidenceProvisionalLinkStore(
            projectDir: session.projectKey.projectDir
        ) {
            self.linkStore = live
        } else {
            self.linkStore = InMemoryEvidenceProvisionalLinkStore()
        }
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

    var hasNoArtifacts: Bool {
        phase == .compose && artifacts.isEmpty
    }

    var formIsInert: Bool {
        hasNoArtifacts
    }

    var selectedArtifactIndex: Int? {
        guard let id = selectedArtifactID else { return nil }
        return artifacts.firstIndex(where: { $0.id == id })
    }

    var isPDFArtifact: Bool {
        Self.isPDF(selectedArtifact)
    }

    var isImageArtifact: Bool {
        Self.isImage(selectedArtifact)
    }

    var hasLocator: Bool {
        locatorPage != nil
    }

    var propertyOptions: [PVComboBoxOption] {
        availableProperties.map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    var canConfirmObservation: Bool {
        guard let draft = observationDialog else { return false }
        return Self.observationValueIsValid(draft: draft, property: catalogProperty(id: draft.propertyID))
    }

    var dialogPropertyError: String? {
        guard let draft = observationDialog, draft.showValidation else { return nil }
        if draft.propertyID.isEmpty {
            return String(localized: L10n.CitationComposer.dialogPropertyRequired)
        }
        return nil
    }

    var dialogValueError: String? {
        guard let draft = observationDialog, draft.showValidation else { return nil }
        if draft.propertyID.isEmpty {
            return String(localized: L10n.CitationComposer.dialogValueRequired)
        }
        guard let property = catalogProperty(id: draft.propertyID) else {
            return String(localized: L10n.CitationComposer.dialogValueRequired)
        }
        if Self.observationValueIsValid(draft: draft, property: property) { return nil }
        return String(localized: L10n.CitationComposer.dialogValueRequired)
    }

    func catalogProperty(id: String) -> CatalogProperty? {
        if let match = availableProperties.first(where: { $0.id == id }) {
            return match
        }
        return connectEdgePropertiesByID[id]
    }

    func termOptions(for propertyID: String) -> [PVComboBoxOption] {
        (termsByPropertyID[propertyID] ?? []).map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    func termLabel(for termID: String, propertyID: String) -> String? {
        termsByPropertyID[propertyID]?.first(where: { $0.id == termID })?.label
    }

    func observationSummary(for row: ObservationRow) -> String {
        guard let property = catalogProperty(id: row.propertyID) else { return "" }
        switch property.valueType {
        case "text":
            return row.valueText
        case "integer":
            return row.valueIntegerText
        case "term":
            return termLabel(for: row.valueTermID, propertyID: property.id) ?? row.valueTermID
        case "date":
            return Self.dateSummary(row.dateDraft)
        case "subject":
            let trimmed = row.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
            return trimmed.isEmpty ? row.valueSubjectID : trimmed
        default:
            return ""
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

    /// Ordered edge Property keys for the first matching connect rule.
    static func edgePropertyKeys(
        typeKey: String,
        rules: [CatalogConnectRule]
    ) -> [String] {
        rules
            .first { $0.bridgeTypeKey == typeKey && !$0.refuse }?
            .edgePropertyKeys ?? []
    }

    /// Maps a connect-edge Property key to the endpoint Subject type it binds.
    static func endpointTypeKey(forEdgePropertyKey key: String) -> String? {
        switch key {
        case "person", "related_to": return "person"
        case "event": return "event"
        case "place": return "place"
        default: return nil
        }
    }

    /// Builds Frame-10 fixed rows from provisional endpoints + the connect matrix.
    static func connectEdgePrefillRows(
        bridgeTypeKey: String,
        rules: [CatalogConnectRule],
        properties: [CatalogProperty],
        endpointA: (id: String, label: String, typeKey: String),
        endpointB: (id: String, label: String, typeKey: String)
    ) -> [ObservationRow] {
        let keys = edgePropertyKeys(typeKey: bridgeTypeKey, rules: rules)
        guard !keys.isEmpty else { return [] }
        let propertyByKey = Dictionary(uniqueKeysWithValues: properties.map { ($0.key, $0) })
        var remaining = [endpointA, endpointB]
        var rows: [ObservationRow] = []
        rows.reserveCapacity(keys.count)
        for key in keys {
            guard let property = propertyByKey[key],
                  let wantedType = endpointTypeKey(forEdgePropertyKey: key),
                  let index = remaining.firstIndex(where: { $0.typeKey == wantedType })
            else { continue }
            let endpoint = remaining.remove(at: index)
            rows.append(
                ObservationRow(
                    id: UUID(),
                    propertyID: property.id,
                    polarity: "positive",
                    valueText: endpoint.label,
                    valueIntegerText: "",
                    valueTermID: "",
                    valueSubjectID: endpoint.id,
                    dateDraft: .empty(),
                    isConnectFixed: true
                )
            )
        }
        return rows
    }

    static func isPDF(_ artifact: CatalogArtifact?) -> Bool {
        (artifact?.file?.mediaType ?? "").localizedCaseInsensitiveContains("pdf")
    }

    static func isImage(_ artifact: CatalogArtifact?) -> Bool {
        (artifact?.file?.mediaType ?? "").hasPrefix("image/")
    }

    func warmQueries() {
        let _: QueryHandle<SourceGraphSnapshot> = session.query(graphKey)
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
    }

    func prepare() async {
        phase = .loading
        formError = nil
        locatorError = nil
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
            // Sequential FakeStore access: bags are not synchronized, and
            // concurrent `async let` here has aborted CI under suite fan-out.
            let workspace = try await store.getSourceWorkspace(
                projectDir: projectDir,
                sourceID: sourceID
            )
            let rules = try await store.listConnectRules()
            let types = try await store.listSubjectTypes(projectDir: projectDir)
            let listedObservations = try await store.listObservationsBySource(
                projectDir: projectDir,
                sourceID: sourceID
            )
            let sourceTypes = try await store.listSourceTypes(projectDir: projectDir)

            let snapshot = SourceGraphSnapshot.build(
                sourceId: sourceID,
                subjects: subjects,
                positions: positions,
                types: types,
                observations: listedObservations
            )

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
            sourceTitle = workspace.source.title
            sourceTypeIconKey = sourceTypes
                .first { $0.id == workspace.source.sourceTypeID }?
                .iconKey ?? ""
            artifacts = workspace.artifacts

            let fields = try await store.listSubjectTypeFields(
                projectDir: projectDir,
                subjectTypeID: resolved.typeID
            )
            let excluded = Self.excludedEdgePropertyKeys(
                typeKey: resolved.typeKey,
                rules: rules
            )
            let allProperties = fields.map(\.property)
            connectEdgePropertiesByID = Dictionary(
                uniqueKeysWithValues: allProperties
                    .filter { excluded.contains($0.key) }
                    .map { ($0.id, $0) }
            )
            availableProperties = allProperties
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

            let subjectLabelsByID = Dictionary(
                uniqueKeysWithValues: subjects.map { ($0.id, $0.label) }
            )
            let typeKeyBySubjectID = Dictionary(
                uniqueKeysWithValues: subjects.compactMap { subject -> (String, String)? in
                    guard let type = types.first(where: { $0.id == subject.subjectTypeID })
                    else { return nil }
                    return (subject.id, type.key)
                }
            )

            if let citationID {
                let (citation, _, citationObservations) = try await store.getCitation(
                    projectDir: projectDir,
                    citationID: citationID
                )
                applyLoadedCitation(
                    citation,
                    observations: citationObservations,
                    subjectLabelsByID: subjectLabelsByID
                )
                if artifacts.isEmpty {
                    selectedArtifactID = nil
                    pendingArtifactID = nil
                    phase = .compose
                    return
                }
                phase = .compose
                await reloadArtifactViewer(preferredPage: locatorPage)
                return
            }

            observations = []
            if EvidenceBridgeKind(rawValue: resolved.typeKey) != nil,
               let link = linkStore.links(for: sourceID)
                .first(where: { $0.bridgeSubjectID == subjectID }),
               let typeA = typeKeyBySubjectID[link.endpointAID],
               let typeB = typeKeyBySubjectID[link.endpointBID]
            {
                observations = Self.connectEdgePrefillRows(
                    bridgeTypeKey: resolved.typeKey,
                    rules: rules,
                    properties: allProperties,
                    endpointA: (
                        id: link.endpointAID,
                        label: subjectLabelsByID[link.endpointAID] ?? link.endpointAID,
                        typeKey: typeA
                    ),
                    endpointB: (
                        id: link.endpointBID,
                        label: subjectLabelsByID[link.endpointBID] ?? link.endpointBID,
                        typeKey: typeB
                    )
                )
            }

            if artifacts.isEmpty {
                selectedArtifactID = nil
                pendingArtifactID = nil
                phase = .compose
                return
            }
            if artifacts.count == 1 {
                await applySelectedArtifact(artifacts[0].id)
                phase = .compose
                return
            }
            pendingArtifactID = selectedArtifactID ?? artifacts.first?.id
            phase = .pickArtifact
        } catch {
            formError = L10n.Errors.message(for: error)
            artifacts = []
            phase = .compose
        }
    }

    func selectPendingArtifact(_ id: String) {
        pendingArtifactID = id
    }

    func confirmArtifactSelection() {
        guard let id = pendingArtifactID, artifacts.contains(where: { $0.id == id }) else { return }
        selectedArtifactID = id
        pendingArtifactID = id
        clearLocator()
        phase = .compose
        formError = nil
        locatorError = nil
        submitAttempted = false
        Task { await reloadArtifactViewer(preferredPage: nil) }
    }

    func changeArtifact() {
        guard artifacts.count > 1 else { return }
        pendingArtifactID = selectedArtifactID
        phase = .pickArtifact
        formError = nil
        locatorError = nil
        submitAttempted = false
    }

    func goToPreviousPage() {
        guard artifactViewer.supportsPages else { return }
        artifactViewer.goToPreviousPage()
        syncLocatorFromViewerPage()
    }

    func goToNextPage() {
        guard artifactViewer.supportsPages else { return }
        artifactViewer.goToNextPage()
        syncLocatorFromViewerPage()
    }

    /// Called when the Artifact viewer page changes (chevrons / page field).
    func syncLocatorFromViewerPage() {
        guard artifactViewer.supportsPages else { return }
        commitLocatorPage(artifactViewer.page)
    }

    /// Thin stand-in for Draw region (no polygon UI until S7-07).
    func markWholeImageLocator() {
        if isPDFArtifact {
            locatorIsWholeImage = false
            commitLocatorPage(artifactViewer.page)
            return
        }
        locatorIsWholeImage = true
        commitLocatorPage(1)
    }

    func clearLocator() {
        locatorPage = nil
        locatorIsWholeImage = false
        locatorError = nil
    }

    func beginAddObservation() {
        observationDialog = .fresh()
    }

    func beginEditObservation(_ row: ObservationRow) {
        guard !row.isConnectFixed else { return }
        observationDialog = .editing(row)
    }

    func cancelObservationDialog() {
        observationDialog = nil
    }

    func updateObservationDialog(_ draft: ObservationDialogState) {
        var next = draft
        if let current = observationDialog, current.propertyID != draft.propertyID {
            let previousType = catalogProperty(id: current.propertyID)?.valueType
            let nextType = catalogProperty(id: draft.propertyID)?.valueType
            if previousType != nil, previousType != nextType {
                next.valueText = ""
                next.valueIntegerText = ""
                next.valueTermID = ""
                next.dateDraft = .empty()
            }
        }
        observationDialog = next
    }

    func confirmObservationDialog() {
        guard var draft = observationDialog else { return }
        draft.showValidation = true
        observationDialog = draft
        guard let property = catalogProperty(id: draft.propertyID),
              Self.observationValueIsValid(draft: draft, property: property)
        else { return }

        let row = ObservationRow(
            id: draft.editingID ?? UUID(),
            propertyID: draft.propertyID,
            polarity: draft.polarity == "negative" ? "negative" : "positive",
            valueText: draft.valueText.trimmingCharacters(in: .whitespacesAndNewlines),
            valueIntegerText: draft.valueIntegerText.trimmingCharacters(in: .whitespacesAndNewlines),
            valueTermID: draft.valueTermID,
            valueSubjectID: "",
            dateDraft: draft.dateDraft,
            isConnectFixed: false
        )
        if let editingID = draft.editingID,
           let index = observations.firstIndex(where: { $0.id == editingID })
        {
            observations[index] = row
        } else {
            observations.append(row)
        }
        observationDialog = nil
        formError = nil
    }

    func removeObservation(id: UUID) {
        observations.removeAll { $0.id == id && !$0.isConnectFixed }
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
        locatorError = nil

        if hasNoArtifacts {
            formError = String(localized: L10n.CitationComposer.needArtifact)
            return nil
        }
        guard let artifactID = selectedArtifactID else {
            formError = String(localized: L10n.CitationComposer.needArtifact)
            return nil
        }
        guard let page = locatorPage, let locatorJSON = Self.locatorJSON(page: page) else {
            locatorError = String(localized: L10n.CitationComposer.locatorUnsetError)
            formError = String(localized: L10n.CitationComposer.locatorSaveError)
            return nil
        }
        guard !observations.isEmpty else {
            formError = String(localized: L10n.CitationComposer.saveNeedsObservationError)
            return nil
        }
        guard let drafts = buildObservationDrafts() else { return nil }

        isSubmitting = true
        defer { isSubmitting = false }
        do {
            if let citationID {
                _ = try await store.updateCitationWithObservations(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    citationID: citationID,
                    artifactID: artifactID,
                    locatorJSON: locatorJSON,
                    transcription: transcription,
                    description: citationDescription,
                    transcriptionUncertain: transcriptionUncertain,
                    transcriptionNote: transcriptionNote,
                    citationNotes: [],
                    observations: drafts
                )
            } else {
                _ = try await store.createCitationWithObservations(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    artifactID: artifactID,
                    locatorJSON: locatorJSON,
                    transcription: transcription,
                    description: citationDescription,
                    transcriptionUncertain: transcriptionUncertain,
                    transcriptionNote: transcriptionNote,
                    citationNotes: [],
                    observations: drafts
                )
            }
            session.apply(.createdCitation(sourceId: sourceID))
            didSubmit = true
            return graphLocation()
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

    func sourcePageLocation() -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: sourceID,
            sourceSurface: .page
        )
    }

    // MARK: - Private

    private func applyLoadedCitation(
        _ citation: CatalogCitation,
        observations listed: [CatalogObservation],
        subjectLabelsByID: [String: String]
    ) {
        selectedArtifactID = citation.artifactID
        pendingArtifactID = citation.artifactID
        transcription = citation.transcription
        transcriptionUncertain = citation.transcriptionUncertain
        transcriptionNote = citation.transcriptionNote
        citationDescription = citation.description
        if let page = Self.pageFromLocatorJSON(citation.locatorJSON) {
            locatorPage = page
        }
        let fixedIDs = Set(connectEdgePropertiesByID.keys)
        observations = listed
            .filter { $0.subjectID == subjectID }
            .compactMap {
                Self.observationRow(
                    from: $0,
                    propertiesByID: Dictionary(
                        uniqueKeysWithValues: (availableProperties + Array(connectEdgePropertiesByID.values))
                            .map { ($0.id, $0) }
                    ),
                    fixedPropertyIDs: fixedIDs,
                    subjectLabelsByID: subjectLabelsByID
                )
            }
    }

    private func applySelectedArtifact(_ id: String) async {
        selectedArtifactID = id
        pendingArtifactID = id
        clearLocator()
        await reloadArtifactViewer(preferredPage: nil)
    }

    private func reloadArtifactViewer(preferredPage: Int?) async {
        guard let artifact = selectedArtifact else {
            artifactViewer.unload()
            return
        }
        let relPath = (artifact.file?.relPath ?? "")
            .trimmingCharacters(in: .whitespacesAndNewlines)
        let mediaType = artifact.file?.mediaType ?? ""
        guard !relPath.isEmpty else {
            await artifactViewer.load(
                ArtifactViewerSource(
                    projectDir: session.projectKey.projectDir,
                    relPath: "",
                    mediaType: mediaType.isEmpty ? "application/octet-stream" : mediaType
                )
            )
            return
        }
        await artifactViewer.load(
            ArtifactViewerSource(
                projectDir: session.projectKey.projectDir,
                relPath: relPath,
                mediaType: mediaType
            )
        )
        if let preferredPage, artifactViewer.supportsPages {
            artifactViewer.setPage(preferredPage)
        }
    }

    private func commitLocatorPage(_ page: Int) {
        locatorPage = max(1, page)
        locatorError = nil
    }

    private func buildObservationDrafts() -> [CatalogObservationDraft]? {
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
                draft.valueText = row.valueText
            case "integer":
                guard let value = Int64(row.valueIntegerText) else {
                    formError = String(localized: L10n.CitationComposer.invalidIntegerError)
                    return nil
                }
                draft.valueInteger = value
            case "term":
                draft.valueTermID = row.valueTermID
            case "date":
                guard row.dateDraft.isValid else {
                    formError = String(localized: L10n.CitationComposer.invalidDateError)
                    return nil
                }
                draft.date = row.dateDraft.toInput()
            case "subject":
                guard !row.valueSubjectID.isEmpty else {
                    formError = String(localized: L10n.CitationComposer.unsupportedValueTypeError)
                    return nil
                }
                draft.valueSubjectID = row.valueSubjectID
            default:
                formError = String(localized: L10n.CitationComposer.unsupportedValueTypeError)
                return nil
            }
            drafts.append(draft)
        }
        return drafts
    }

    private static func observationValueIsValid(
        draft: ObservationDialogState,
        property: CatalogProperty?
    ) -> Bool {
        guard let property else { return false }
        switch property.valueType {
        case "text":
            return !draft.valueText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        case "integer":
            return Int64(draft.valueIntegerText.trimmingCharacters(in: .whitespacesAndNewlines)) != nil
        case "term":
            return !draft.valueTermID.isEmpty
        case "date":
            return draft.dateDraft.isValid
        default:
            return false
        }
    }

    static func locatorJSON(page: Int) -> String? {
        guard page >= 1 else { return nil }
        return #"{"version":1,"selectors":[{"type":"page","artifact_page":\#(page)}]}"#
    }

    static func pageFromLocatorJSON(_ json: String) -> Int? {
        struct LocatorDoc: Decodable {
            struct Selector: Decodable {
                var type: String
                var artifact_page: Int?
            }
            var selectors: [Selector]?
        }
        guard let data = json.data(using: .utf8),
              let doc = try? JSONDecoder().decode(LocatorDoc.self, from: data),
              let page = doc.selectors?.first(where: { $0.type == "page" })?.artifact_page,
              page >= 1
        else { return nil }
        return page
    }

    private static func observationRow(
        from observation: CatalogObservation,
        propertiesByID: [String: CatalogProperty],
        fixedPropertyIDs: Set<String>,
        subjectLabelsByID: [String: String]
    ) -> ObservationRow? {
        guard propertiesByID[observation.propertyID] != nil else {
            return nil
        }
        var integerText = ""
        if let value = observation.valueInteger {
            integerText = String(value)
        }
        let dateDraft: DateValueDraft
        if let date = observation.date {
            dateDraft = DateValueDraft(from: date)
        } else {
            dateDraft = .empty()
        }
        let isFixed = fixedPropertyIDs.contains(observation.propertyID)
        let subjectLabel = subjectLabelsByID[observation.valueSubjectID] ?? ""
        let valueText: String
        if !observation.valueText.isEmpty {
            valueText = observation.valueText
        } else if !subjectLabel.isEmpty {
            valueText = subjectLabel
        } else {
            valueText = ""
        }
        return ObservationRow(
            id: UUID(),
            propertyID: observation.propertyID,
            polarity: observation.polarity.isEmpty ? "positive" : observation.polarity,
            valueText: valueText,
            valueIntegerText: integerText,
            valueTermID: observation.valueTermID,
            valueSubjectID: observation.valueSubjectID,
            dateDraft: dateDraft,
            isConnectFixed: isFixed
        )
    }

    static func dateSummary(_ draft: DateValueDraft) -> String {
        let formatted = DateValueDisplay.string(for: draft)
        if formatted.isEmpty {
            return String(localized: L10n.CitationComposer.dateUnset)
        }
        return formatted
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
        _ = types
        return nil
    }
}
