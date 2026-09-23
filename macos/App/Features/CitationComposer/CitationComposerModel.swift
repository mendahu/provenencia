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
        case loadFailed
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
        var nameDraft: NameValueDraft = .empty()
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
        var nameDraft: NameValueDraft = .empty()
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
                nameDraft: .empty(),
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
                nameDraft: row.nameDraft,
                showValidation: false
            )
        }
    }

    /// Value types the composer can edit (subject waits for a later PR).
    static let supportedValueTypes: Set<String> = ["text", "integer", "date", "term", "name"]

    let sourceID: String
    let subjectID: String
    /// When set, prepare/save load and update this Citation instead of creating.
    let citationID: String?
    let connectFromSubjectID: String?
    let connectToSubjectID: String?
    let connectBridgeTypeKey: String?
    let connectDisambiguationTermID: String?
    let connectGridX: Int64
    let connectGridY: Int64
    private let userID: String
    private let store: any GenealogyStore
    let session: WorkspaceSession

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

    /// Isolated document viewer (S7-06). Composer owns locator JSON layers.
    let artifactViewer = ArtifactViewerModel()

    /// Always includes the `artifact` floor after compose starts.
    private(set) var locator = CitationLocatorDraft.artifactOnly()
    /// Armed region radio; overlay commits then this returns to nil.
    var armedRegionTool: ArtifactRegionTool?

    /// 1-based page from the Artifact viewer (PDF).
    var viewerPage: Int { artifactViewer.page }
    var viewerPageCount: Int { artifactViewer.pageCount }
    var locatorCapabilities: ArtifactLocatorCapabilities {
        artifactViewer.locatorCapabilities
    }

    var transcription = ""
    var transcriptionUncertain = false
    var transcriptionNote = ""
    var citationDescription = ""
    private(set) var observations: [ObservationRow] = []
    var observationDialog: ObservationDialogState?
    var formError: String?
    /// Catalog read failed before the form could open.
    private(set) var loadError: String?
    /// Custom-term dialog failure. Shown on that dialog, not the form footer.
    var termError: String?
    var submitAttempted = false

    /// When true, host should `go(to:)` the Evidence graph (subject gone).
    private(set) var shouldFallbackToGraph = false

    var isEditingExisting: Bool { citationID != nil }

    var isConnectPrefill: Bool {
        subjectID.isEmpty
            && connectFromSubjectID != nil
            && connectToSubjectID != nil
            && connectBridgeTypeKey != nil
    }

    init(
        sourceID: String,
        subjectID: String,
        citationID: String? = nil,
        connectFromSubjectID: String? = nil,
        connectToSubjectID: String? = nil,
        connectBridgeTypeKey: String? = nil,
        connectDisambiguationTermID: String? = nil,
        connectGridX: Int64 = 0,
        connectGridY: Int64 = 0,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.sourceID = sourceID
        self.subjectID = subjectID
        self.citationID = citationID
        self.connectFromSubjectID = connectFromSubjectID
        self.connectToSubjectID = connectToSubjectID
        self.connectBridgeTypeKey = connectBridgeTypeKey
        self.connectDisambiguationTermID = connectDisambiguationTermID
        self.connectGridX = connectGridX
        self.connectGridY = connectGridY
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

    var citationCountsKey: CatalogQueryKey {
        .citationCounts(project: session.projectKey, sourceId: sourceID)
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
        artifactViewer.kind == .pdf
    }

    var isImageArtifact: Bool {
        artifactViewer.kind == .image
    }

    var hasLocator: Bool {
        selectedArtifactID != nil
    }

    var isLocatorArtifactOnly: Bool {
        locator.isArtifactOnly
    }

    var isLocatorPageSetOnViewer: Bool {
        guard let page = locator.page else { return false }
        return page == artifactViewer.page
    }

    var locatorPage: Int? { locator.page }

    var propertyOptions: [PVComboBoxOption] {
        availableProperties.map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    var canConfirmObservation: Bool {
        guard let draft = observationDialog else { return false }
        guard let property = catalogProperty(id: draft.propertyID) else { return false }
        return CitationObservationValue.isDialogValid(
            valueType: property.valueType,
            fields: CitationObservationValue.fields(from: draft)
        )
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
        if CitationObservationValue.isDialogValid(
            valueType: property.valueType,
            fields: CitationObservationValue.fields(from: draft)
        ) { return nil }
        return String(localized: L10n.CitationComposer.dialogValueRequired)
    }

    func catalogProperty(id: String) -> CatalogProperty? {
        if let match = availableProperties.first(where: { $0.id == id }) {
            return match
        }
        return connectEdgePropertiesByID[id]
    }

    func termOptions(for propertyID: String) -> [PVComboBoxOption] {
        let propertyKey = catalogProperty(id: propertyID)?.key ?? ""
        return (termsByPropertyID[propertyID] ?? []).map { term in
            PVComboBoxOption(
                value: term.id,
                label: PropertyTermDisplay.name(term: term, propertyKey: propertyKey),
                subtext: term.key
            )
        }
    }

    func termLabel(for termID: String, propertyID: String) -> String? {
        guard let term = termsByPropertyID[propertyID]?.first(where: { $0.id == termID }) else {
            return nil
        }
        let propertyKey = catalogProperty(id: propertyID)?.key ?? ""
        return PropertyTermDisplay.name(term: term, propertyKey: propertyKey)
    }

    func observationSummary(for row: ObservationRow) -> String {
        guard let property = catalogProperty(id: row.propertyID) else { return "" }
        return CitationObservationValue.summary(
            valueType: property.valueType,
            fields: CitationObservationValue.fields(from: row),
            termLabel: termLabel(for: row.valueTermID, propertyID: property.id)
        )
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

    static func hasAttachedFile(_ artifact: CatalogArtifact) -> Bool {
        artifact.file != nil && !artifact.fileID.isEmpty
    }

    /// Picker default for a multi-artifact Source.
    /// 1. Unique most-cited artifact; citation ties prefer the Source thumbnail.
    /// 2. First artifact with a file when the cover is the type-icon fallback
    ///    (or the thumbnail is not among the citation leaders).
    /// 3. First remaining artifact in list order.
    static func defaultPendingArtifactID(
        artifacts: [CatalogArtifact],
        citationCounts: [String: Int],
        selectedID: String?,
        coverMode: String,
        primaryArtifactID: String
    ) -> String? {
        if let selectedID, artifacts.contains(where: { $0.id == selectedID }) {
            return selectedID
        }
        guard !artifacts.isEmpty else { return nil }

        let maxCount = artifacts.map { citationCounts[$0.id, default: 0] }.max() ?? 0
        let leaders = artifacts.filter { citationCounts[$0.id, default: 0] == maxCount }
        let thumbnailID = (coverMode == "artifact" && !primaryArtifactID.isEmpty)
            ? primaryArtifactID : nil

        if leaders.count == 1 {
            return leaders[0].id
        }
        if let thumbnailID, leaders.contains(where: { $0.id == thumbnailID }) {
            return thumbnailID
        }
        return leaders.first(where: hasAttachedFile)?.id ?? leaders.first?.id
    }

    var sourceTypesKey: CatalogQueryKey {
        .sourceTypesList(project: session.projectKey)
    }

    func prepare() async {
        phase = .loading
        formError = nil
        loadError = nil
        termError = nil
        shouldFallbackToGraph = false

        do {
            let projectDir = session.projectKey.projectDir
            // Prefer handles PlaceRegistry already warmed. Cold opens fill the
            // same keys sequentially — FakeStore bags are not safe to hit concurrently.
            let snapshot = try await cached(graphKey) {
                let subjects = try await self.store.listSubjects(
                    projectDir: projectDir,
                    sourceID: self.sourceID
                )
                let positions = try await self.store.listSubjectPositions(
                    projectDir: projectDir,
                    sourceID: self.sourceID
                )
                let types = try await self.store.listSubjectTypes(projectDir: projectDir)
                let observations = try await self.store.listObservationsBySource(
                    projectDir: projectDir,
                    sourceID: self.sourceID
                )
                return SourceGraphSnapshot.build(
                    sourceId: self.sourceID,
                    subjects: subjects,
                    positions: positions,
                    types: types,
                    observations: observations
                )
            }
            let workspace = try await cached(workspaceKey) {
                try await self.store.getSourceWorkspace(
                    projectDir: projectDir,
                    sourceID: self.sourceID
                )
            }
            let sourceTypes = try await cached(sourceTypesKey) {
                try await self.store.listSourceTypes(projectDir: projectDir)
            }
            let rules = try await store.listConnectRules()
            let types = try await store.listSubjectTypes(projectDir: projectDir)

            let resolved: ResolvedSubject
            if isConnectPrefill, let bridgeKey = connectBridgeTypeKey,
               let type = types.first(where: { $0.key == bridgeKey })
            {
                resolved = ResolvedSubject(label: "", typeKey: bridgeKey, typeID: type.id)
            } else if let existing = Self.resolveSubject(id: subjectID, in: snapshot) {
                resolved = existing
            } else {
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

            let subjectLabelsByID = Self.subjectLabels(in: snapshot)
            let typeKeyBySubjectID = Self.typeKeyBySubjectID(in: snapshot)

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
                await reloadArtifactViewer(preferredPage: locator.page)
                return
            }

            observations = []
            if isConnectPrefill,
               let fromID = connectFromSubjectID,
               let toID = connectToSubjectID,
               let typeA = typeKeyBySubjectID[fromID],
               let typeB = typeKeyBySubjectID[toID]
            {
                observations = Self.connectEdgePrefillRows(
                    bridgeTypeKey: resolved.typeKey,
                    rules: rules,
                    properties: allProperties,
                    endpointA: (
                        id: fromID,
                        label: subjectLabelsByID[fromID] ?? fromID,
                        typeKey: typeA
                    ),
                    endpointB: (
                        id: toID,
                        label: subjectLabelsByID[toID] ?? toID,
                        typeKey: typeB
                    )
                )
                if let termID = connectDisambiguationTermID, !termID.isEmpty,
                   let termProperty = allProperties.first(where: { $0.key == rules
                       .first(where: { $0.bridgeTypeKey == resolved.typeKey && !$0.refuse })?
                       .disambiguation })
                {
                    let terms = try await store.listPropertyTerms(
                        projectDir: projectDir,
                        propertyID: termProperty.id
                    )
                    termsByPropertyID[termProperty.id] = terms
                    observations.append(
                        ObservationRow(
                            id: UUID(),
                            propertyID: termProperty.id,
                            polarity: "positive",
                            valueText: terms.first(where: { $0.id == termID })?.label ?? "",
                            valueIntegerText: "",
                            valueTermID: termID,
                            valueSubjectID: "",
                            dateDraft: .empty(),
                            isConnectFixed: false
                        )
                    )
                }
                subjectLabel = bridgeSentence(
                    typeKey: resolved.typeKey,
                    observations: observations,
                    properties: allProperties,
                    subjectLabelsByID: subjectLabelsByID
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
            let counts = try await cached(citationCountsKey) {
                try await self.store.citationCountsBySource(
                    projectDir: projectDir,
                    sourceID: self.sourceID
                )
            }
            pendingArtifactID = Self.defaultPendingArtifactID(
                artifacts: artifacts,
                citationCounts: counts,
                selectedID: selectedArtifactID,
                coverMode: workspace.source.coverMode,
                primaryArtifactID: workspace.source.primaryArtifactID
            )
            phase = .pickArtifact
        } catch {
            loadError = L10n.Errors.message(for: error)
            artifacts = []
            phase = .loadFailed
        }
    }

    /// Uses a warmed session value when one exists. Otherwise loads sequentially
    /// and publishes into the same key so the cache stays the owner.
    private func cached<Value>(
        _ key: CatalogQueryKey,
        load: () async throws -> Value
    ) async throws -> Value {
        if let ready: Value = await session.readyValue(key) {
            return ready
        }
        let value = try await load()
        session.setQueryValue(key, value: value)
        return value
    }

    func selectPendingArtifact(_ id: String) {
        pendingArtifactID = id
    }

    func confirmArtifactSelection() {
        Task { await confirmArtifactSelectionAndLoad() }
    }

    func confirmArtifactSelectionAndLoad() async {
        guard let id = pendingArtifactID, artifacts.contains(where: { $0.id == id }) else { return }
        selectedArtifactID = id
        pendingArtifactID = id
        resetToEntireArtifact()
        phase = .compose
        formError = nil
        submitAttempted = false
        await reloadArtifactViewer(preferredPage: nil)
    }

    func changeArtifact() {
        guard artifacts.count > 1 else { return }
        pendingArtifactID = selectedArtifactID
        phase = .pickArtifact
        formError = nil
        submitAttempted = false
    }

    func goToPreviousPage() {
        guard artifactViewer.supportsPages else { return }
        artifactViewer.goToPreviousPage()
    }

    func goToNextPage() {
        guard artifactViewer.supportsPages else { return }
        artifactViewer.goToNextPage()
    }

    /// Set Page — writes the current viewer page; browsing chevrons do not.
    @discardableResult
    func setPageFromViewer() -> Bool {
        setPage(artifactViewer.page)
    }

    @discardableResult
    func setPage(_ page: Int) -> Bool {
        locator.setPage(page, capabilities: locatorCapabilities)
    }

    @discardableResult
    func setRegion(_ draft: ArtifactRegionDraft) -> Bool {
        let ok = locator.setRegion(
            draft,
            capabilities: locatorCapabilities,
            autoPage: artifactViewer.page
        )
        if ok {
            armedRegionTool = nil
        }
        return ok
    }

    func clearRegion() {
        locator.clearRegion()
    }

    func resetToEntireArtifact() {
        locator.resetToEntireArtifact()
        armedRegionTool = nil
    }

    func removePage() {
        locator.removePage()
    }

    func removeRegion() {
        locator.clearRegion()
    }

    func disarmRegionTool() {
        armedRegionTool = nil
    }

    func clearLocator() {
        resetToEntireArtifact()
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
                next.nameDraft = .empty()
            }
        }
        observationDialog = next
    }

    func confirmObservationDialog() {
        guard var draft = observationDialog else { return }
        draft.showValidation = true
        observationDialog = draft
        guard let property = catalogProperty(id: draft.propertyID),
              CitationObservationValue.isDialogValid(
                valueType: property.valueType,
                fields: CitationObservationValue.fields(from: draft)
              )
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
            nameDraft: draft.nameDraft,
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
            termError = nil
            return term
        } catch {
            termError = L10n.Errors.message(for: error)
            return nil
        }
    }

    /// Returns graph location after success; nil when validation/submit failed.
    func submit() async -> WorkspaceLocation? {
        submitAttempted = true
        formError = nil
        if hasNoArtifacts {
            formError = String(localized: L10n.CitationComposer.needArtifact)
            return nil
        }
        guard let artifactID = selectedArtifactID else {
            formError = String(localized: L10n.CitationComposer.needArtifact)
            return nil
        }
        let locatorJSON = locator.encodeJSON()
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
            } else if isConnectPrefill,
                      let fromID = connectFromSubjectID,
                      let toID = connectToSubjectID,
                      let bridgeKey = connectBridgeTypeKey
            {
                _ = try await store.createCitedBridge(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    sourceID: sourceID,
                    fromSubjectID: fromID,
                    toSubjectID: toID,
                    bridgeTypeKey: bridgeKey,
                    label: subjectLabel,
                    description: "",
                    gridX: connectGridX,
                    gridY: connectGridY,
                    artifactID: artifactID,
                    locatorJSON: locatorJSON,
                    transcription: transcription,
                    citationDescription: citationDescription,
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
        locator = CitationLocatorDraft.decode(citation.locatorJSON)
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
        peelIllegalLocatorLayers()
    }

    private func peelIllegalLocatorLayers() {
        locator.peelIllegalLayers(capabilities: locatorCapabilities)
        if !locatorCapabilities.supportsRegionLocator {
            armedRegionTool = nil
        }
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
            if let failure = CitationObservationValue.apply(
                valueType: property.valueType,
                fields: CitationObservationValue.fields(from: row),
                to: &draft
            ) {
                formError = CitationObservationValue.message(for: failure)
                return nil
            }
            drafts.append(draft)
        }
        return drafts
    }

    static func locatorJSON(page: Int) -> String? {
        var draft = CitationLocatorDraft.artifactOnly()
        guard draft.setPage(page, capabilities: ArtifactViewerKind.pdf.locatorCapabilities) else {
            return nil
        }
        return draft.encodeJSON()
    }

    static func pageFromLocatorJSON(_ json: String) -> Int? {
        CitationLocatorDraft.decode(json).page
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
            nameDraft: NameValueDraft(from: observation),
            isConnectFixed: isFixed
        )
    }

    private struct ResolvedSubject {
        var label: String
        var typeKey: String
        var typeID: String
    }

    private static func resolveSubject(
        id: String,
        in snapshot: SourceGraphSnapshot
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
        return nil
    }

    private static func subjectLabels(in snapshot: SourceGraphSnapshot) -> [String: String] {
        var labels: [String: String] = [:]
        for placed in snapshot.subjects {
            labels[placed.id] = placed.subject.label
        }
        for bridge in snapshot.bridges {
            labels[bridge.id] = bridge.subject.label
        }
        return labels
    }

    private static func typeKeyBySubjectID(in snapshot: SourceGraphSnapshot) -> [String: String] {
        var keys: [String: String] = [:]
        for placed in snapshot.subjects {
            keys[placed.id] = placed.kind.rawValue
        }
        for bridge in snapshot.bridges {
            keys[bridge.id] = bridge.kind.rawValue
        }
        return keys
    }

    /// Maps draft rows onto the graph's string sentence. The graph type does not
    /// know about composer rows.
    private func bridgeSentence(
        typeKey: String,
        observations: [ObservationRow],
        properties: [CatalogProperty],
        subjectLabelsByID: [String: String]
    ) -> String {
        guard let kind = EvidenceBridgeKind(rawValue: typeKey) else { return "" }
        let propertyByID = Dictionary(uniqueKeysWithValues: properties.map { ($0.id, $0) })
        func subject(for key: String) -> String? {
            guard let row = observations.first(where: { propertyByID[$0.propertyID]?.key == key })
            else { return nil }
            let labeled = subjectLabelsByID[row.valueSubjectID]
                ?? row.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
            return labeled.isEmpty ? nil : labeled
        }
        func term(for key: String) -> String? {
            guard let row = observations.first(where: { propertyByID[$0.propertyID]?.key == key }),
                  !row.valueTermID.isEmpty
            else { return nil }
            if let resolved = termLabel(for: row.valueTermID, propertyID: row.propertyID), !resolved.isEmpty {
                return resolved
            }
            let fallback = row.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
            return fallback.isEmpty ? nil : fallback
        }
        return EvidenceBridgeEdgeSummary.sentence(
            kind: kind,
            person: subject(for: "person"),
            related: subject(for: "related_to"),
            event: subject(for: "event"),
            place: subject(for: "place"),
            term: term(for: kind == .relationship ? "relationship_type" : "role")
        )
    }
}
