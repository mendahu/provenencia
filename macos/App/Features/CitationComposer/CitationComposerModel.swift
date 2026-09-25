import Foundation
import Observation

/// Citation composer (S8-10): Citation is the document; subjects live on each row.
@MainActor
@Observable
final class CitationComposerModel {
    enum Phase: Equatable {
        case loading
        case compose
        case subjectMissing
        case loadFailed
    }

    /// Confirm target when switching Artifact on a dirty saved Citation.
    struct ArtifactAbandon: Identifiable, Equatable {
        var id: String { targetArtifactID }
        var targetArtifactID: String
    }

    /// One Observation on the Citation. `subjectID` is a field on the row.
    struct ObservationRow: Identifiable, Equatable {
        var id: UUID
        var persistedID: String? = nil
        var subjectID: String = ""
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

    /// Live Citation on screen. Replaced as a unit; Entry never writes these fields directly.
    struct Document: Equatable {
        var artifactID: String?
        var citationID: String?
        var transcription: String = ""
        var transcriptionUncertain: Bool = false
        var transcriptionNote: String = ""
        var description: String = ""
        var locator: CitationLocatorDraft = .artifactOnly()
        var observations: [ObservationRow] = []

        static func blank(artifactID: String? = nil) -> Document {
            Document(artifactID: artifactID)
        }
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

    let entry: CitationComposerEntry
    private let userID: String
    private let store: any GenealogyStore
    let session: WorkspaceSession

    var sourceID: String { entry.sourceID }
    var subjectID: String { entry.subjectID }
    var citationID: String? { entry.citationID }
    var isConnectPrefill: Bool { entry.isConnect }

    private(set) var phase: Phase = .loading
    private(set) var subjectLabel: String = ""
    private(set) var subjectTypeKey: String = ""
    private(set) var subjectTypeID: String = ""
    var sourceTitle: String { workspace?.source.title ?? "" }
    /// Source-type `icon_key` for fileless Artifact picker thumbs (Frame 8).
    var sourceTypeIconKey: String {
        sourceTypes.first { $0.id == workspace?.source.sourceTypeID }?.iconKey ?? ""
    }
    var artifacts: [CatalogArtifact] { workspace?.artifacts ?? [] }
    private(set) var document = Document.blank()
    var listedCitations: [CatalogListedCitation] {
        guard let artifactID = selectedArtifactID else { return [] }
        let handle: QueryHandle<[CatalogListedCitation]>? = session.queryHandle(
            citationsKey(artifactID: artifactID)
        )
        return handle?.value ?? []
    }
    private(set) var focusedObservationID: UUID?
    var pendingArtifactAbandon: ArtifactAbandon?
    /// VoiceOver announcement after Artifact / Citation identity changes.
    private(set) var identityAnnouncement: String = ""
    /// Graph subjects offered on each Observation row (primaries + bridges).
    var graphSubjects: [GraphSubjectOption] {
        Self.graphSubjectOptions(in: graphSnapshot)
    }
    var propertiesBySubjectTypeID: [String: [CatalogProperty]] {
        Dictionary(
            uniqueKeysWithValues: fields.fieldsByTypeID.map { ($0.key, $0.value.map(\.property)) }
        )
    }
    var allPropertiesByID: [String: CatalogProperty] {
        var byID = Dictionary(uniqueKeysWithValues: fields.properties.map { ($0.id, $0) })
        for fieldsForType in fields.fieldsByTypeID.values {
            for field in fieldsForType {
                byID[field.property.id] = field.property
            }
        }
        return byID
    }
    var availableProperties: [CatalogProperty] {
        properties(
            forSubjectTypeID: subjectTypeID,
            typeKey: subjectTypeKey,
            excludingEdges: true
        )
    }
    /// Connect-edge Properties (excluded from Add observation) keyed by id.
    var connectEdgePropertiesByID: [String: CatalogProperty] {
        let excluded = Self.excludedEdgePropertyKeys(typeKey: subjectTypeKey, rules: connectRules)
        let props = propertiesBySubjectTypeID[subjectTypeID] ?? []
        return Dictionary(uniqueKeysWithValues: props.filter { excluded.contains($0.key) }.map { ($0.id, $0) })
    }
    /// Connect matrix used to lock edge rows for any Subject, not just entry type.
    var connectRules: [CatalogConnectRule] {
        let handle: QueryHandle<[CatalogConnectRule]>? = session.queryHandle(connectRulesKey)
        return handle?.value ?? []
    }
    /// Connect-new prefill; restored when identity returns to New.
    private var entryConnectPrefill: [ObservationRow] = []
    var termsByPropertyID: [String: [CatalogPropertyTerm]] {
        var result: [String: [CatalogPropertyTerm]] = [:]
        let ids = Set(
            allPropertiesByID.values.filter { $0.valueType == "term" }.map(\.id)
        )
        for id in ids {
            let handle: QueryHandle<[CatalogPropertyTerm]>? = session.queryHandle(
                termsKey(propertyID: id)
            )
            if let terms = handle?.value {
                result[id] = terms
            }
        }
        return result
    }
    private(set) var isSubmitting = false
    private(set) var didSubmit = false

    /// Isolated document viewer (S7-06). Composer owns locator JSON layers.
    let artifactViewer = ArtifactViewerModel()

    /// Armed region radio; overlay commits then this returns to nil.
    var armedRegionTool: ArtifactRegionTool?

    /// 1-based page from the Artifact viewer (PDF).
    var viewerPage: Int { artifactViewer.page }
    var viewerPageCount: Int { artifactViewer.pageCount }
    var locatorCapabilities: ArtifactLocatorCapabilities {
        artifactViewer.locatorCapabilities
    }

    var selectedArtifactID: String? { document.artifactID }
    var pendingArtifactID: String? {
        get { document.artifactID }
        set { document.artifactID = newValue }
    }
    var activeCitationID: String? { document.citationID }
    var locator: CitationLocatorDraft {
        get { document.locator }
        set { document.locator = newValue }
    }
    var transcription: String {
        get { document.transcription }
        set { document.transcription = newValue }
    }
    var transcriptionUncertain: Bool {
        get { document.transcriptionUncertain }
        set { document.transcriptionUncertain = newValue }
    }
    var transcriptionNote: String {
        get { document.transcriptionNote }
        set { document.transcriptionNote = newValue }
    }
    var citationDescription: String {
        get { document.description }
        set { document.description = newValue }
    }
    var observations: [ObservationRow] { document.observations }
    var observationDialog: ObservationDialogState?
    var formError: String?
    /// Catalog read failed before the form could open.
    private(set) var loadError: String?
    /// Custom-term dialog failure. Shown on that dialog, not the form footer.
    var termError: String?
    /// Row that the custom-term dialog writes back to.
    var pendingCustomTermRowID: UUID?
    var customTermLabel = ""
    var showCustomTermDialog = false
    var submitAttempted = false

    /// When true, host should `go(to:)` the Evidence graph (subject gone).
    private(set) var shouldFallbackToGraph = false

    var isEditingExisting: Bool { document.citationID != nil }

    var showsArtifactSwitcher: Bool { artifacts.count > 1 }

    var citationCountsByArtifact: [String: Int] {
        let handle: QueryHandle<[String: Int]>? = session.queryHandle(citationCountsKey)
        return handle?.value ?? [:]
    }

    func listedCount(for artifactID: String) -> Int {
        citationCountsByArtifact[artifactID, default: 0]
    }

    var activeCitationRef: String {
        if let id = activeCitationID,
           let listed = listedCitations.first(where: { $0.id == id })
        {
            return listed.citation.ref
        }
        return ""
    }

    var isDirty: Bool { currentSnapshot() != baseline }

    var subjectOptions: [PVComboBoxOption] {
        graphSubjects.map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.ref)
        }
    }

    var defaultObservationSubjectID: String {
        if !subjectID.isEmpty { return subjectID }
        return graphSubjects.first?.id ?? ""
    }

    init(
        entry: CitationComposerEntry,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.entry = entry
        self.session = session
        self.store = store
        self.userID = userID
        if case .edit(_, _, let citationID, let artifactID, _) = entry {
            document.citationID = citationID
            document.artifactID = artifactID
        } else if case .addProperty(_, _, let artifactID) = entry {
            document.artifactID = artifactID
        }
    }

    struct GraphSubjectOption: Identifiable, Equatable {
        var id: String
        var label: String
        var ref: String
        var typeID: String
        var typeKey: String
    }

    private struct FormSnapshot: Equatable {
        var artifactID: String?
        var citationID: String?
        var transcription: String
        var transcriptionUncertain: Bool
        var transcriptionNote: String
        var description: String
        var locator: CitationLocatorDraft
        var observations: [ObservationRow]
    }

    private var baseline = FormSnapshot(
        artifactID: nil,
        citationID: nil,
        transcription: "",
        transcriptionUncertain: false,
        transcriptionNote: "",
        description: "",
        locator: .artifactOnly(),
        observations: []
    )

    var graphKey: CatalogQueryKey {
        .sourceGraph(project: session.projectKey, sourceId: sourceID)
    }

    var workspaceKey: CatalogQueryKey {
        .sourceWorkspace(project: session.projectKey, sourceId: sourceID)
    }

    var citationCountsKey: CatalogQueryKey {
        .citationCounts(project: session.projectKey, sourceId: sourceID)
    }

    var fieldsKey: CatalogQueryKey {
        .subjectFieldsWorkspace(project: session.projectKey)
    }

    var connectRulesKey: CatalogQueryKey {
        .connectRules(project: session.projectKey)
    }

    private func citationsKey(artifactID: String) -> CatalogQueryKey {
        .citationsByArtifact(project: session.projectKey, artifactId: artifactID)
    }

    private func termsKey(propertyID: String) -> CatalogQueryKey {
        .propertyTerms(project: session.projectKey, propertyId: propertyID)
    }

    private var workspace: CatalogSourceWorkspace? {
        let handle: QueryHandle<CatalogSourceWorkspace>? = session.queryHandle(workspaceKey)
        return handle?.value
    }

    private var sourceTypes: [CatalogSourceType] {
        let handle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(sourceTypesKey)
        return handle?.value ?? []
    }

    private var fields: SubjectFieldsSnapshot {
        let handle: QueryHandle<SubjectFieldsSnapshot>? = session.queryHandle(fieldsKey)
        return handle?.value ?? .empty
    }

    private var graphSnapshot: SourceGraphSnapshot {
        let handle: QueryHandle<SourceGraphRows>? = session.queryHandle(graphKey)
        return SourceGraphSnapshot.build(
            rows: handle?.value ?? SourceGraphRows(sourceId: sourceID),
            types: fields.types
        )
    }

    private func properties(
        forSubjectTypeID typeID: String,
        typeKey: String,
        excludingEdges: Bool
    ) -> [CatalogProperty] {
        let excluded = excludingEdges
            ? Self.excludedEdgePropertyKeys(typeKey: typeKey, rules: connectRules)
            : []
        return (propertiesBySubjectTypeID[typeID] ?? [])
            .filter {
                Self.supportedValueTypes.contains($0.valueType) && !excluded.contains($0.key)
            }
            .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
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
        if let match = connectEdgePropertiesByID[id] {
            return match
        }
        return allPropertiesByID[id]
    }

    func propertyOptions(for subjectID: String) -> [PVComboBoxOption] {
        let match = graphSubjects.first(where: { $0.id == subjectID })
        let typeID = match?.typeID ?? subjectTypeID
        let typeKey = match?.typeKey ?? subjectTypeKey
        let excluded = Self.excludedEdgePropertyKeys(typeKey: typeKey, rules: connectRules)
        let props = propertiesBySubjectTypeID[typeID] ?? availableProperties
        return props
            .filter {
                Self.supportedValueTypes.contains($0.valueType) && !excluded.contains($0.key)
            }
            .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
            .map { PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key) }
    }

    func subjectLabel(for subjectID: String) -> String {
        graphSubjects.first(where: { $0.id == subjectID })?.label ?? subjectID
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
            try await warmCatalogHandles()
            let snapshot = graphSnapshot
            let types = fields.types

            let resolved: ResolvedSubject
            if let bridgeKey = entry.connectBridgeTypeKey,
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
            await warmTermHandles()

            let subjectLabelsByID = Self.subjectLabels(in: snapshot)
            let typeKeyBySubjectID = Self.typeKeyBySubjectID(in: snapshot)
            let allProperties = propertiesBySubjectTypeID[resolved.typeID] ?? []

            if let citationID = entry.citationID {
                let (citation, _, citationObservations) = try await store.getCitation(
                    projectDir: session.projectKey.projectDir,
                    citationID: citationID
                )
                replaceDocument(
                    persisted: citation,
                    observations: citationObservations,
                    subjectLabelsByID: subjectLabelsByID
                )
                applyEntryOverlay()
                phase = .compose
                await presentDocument(announce: false)
                return
            }

            document = .blank(artifactID: document.artifactID)
            if case let .connect(_, fromID, toID, _, termID, _, _) = entry,
               let typeA = typeKeyBySubjectID[fromID],
               let typeB = typeKeyBySubjectID[toID]
            {
                var rows = Self.connectEdgePrefillRows(
                    bridgeTypeKey: resolved.typeKey,
                    rules: connectRules,
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
                if let termID, !termID.isEmpty,
                   let termProperty = allProperties.first(where: { $0.key == connectRules
                       .first(where: { $0.bridgeTypeKey == resolved.typeKey && !$0.refuse })?
                       .disambiguation })
                {
                    await warmTerms(for: termProperty.id)
                    let terms = termsByPropertyID[termProperty.id] ?? []
                    rows.append(
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
                    observations: rows,
                    properties: allProperties,
                    subjectLabelsByID: subjectLabelsByID
                )
                entryConnectPrefill = rows
                document.observations = rows
            }

            if artifacts.isEmpty {
                document.artifactID = nil
                phase = .compose
                captureBaseline()
                return
            }
            if artifacts.count == 1 {
                await applySelectedArtifact(artifacts[0].id)
            } else if let workspace {
                pendingArtifactID = Self.defaultPendingArtifactID(
                    artifacts: artifacts,
                    citationCounts: citationCountsByArtifact,
                    selectedID: selectedArtifactID,
                    coverMode: workspace.source.coverMode,
                    primaryArtifactID: workspace.source.primaryArtifactID
                )
                if let id = pendingArtifactID {
                    await applySelectedArtifact(id)
                }
            }
            await reloadListedCitations()
            phase = .compose
            captureBaseline()
        } catch {
            loadError = L10n.Errors.message(for: error)
            phase = .loadFailed
        }
    }

    private func warmCatalogHandles() async throws {
        let _: QueryHandle<SourceGraphRows> = session.query(graphKey)
        let _: QueryHandle<SubjectFieldsSnapshot> = session.query(fieldsKey)
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
        let _: QueryHandle<[CatalogSourceType]> = session.query(sourceTypesKey)
        let _: QueryHandle<[String: Int]> = session.query(citationCountsKey)
        let _: QueryHandle<[CatalogConnectRule]> = session.query(connectRulesKey)

        let rows: SourceGraphRows? = await session.readyValue(graphKey)
        let fieldsReady: SubjectFieldsSnapshot? = await session.readyValue(fieldsKey)
        let workspaceReady: CatalogSourceWorkspace? = await session.readyValue(workspaceKey)
        _ = await session.readyValue(sourceTypesKey) as [CatalogSourceType]?
        _ = await session.readyValue(citationCountsKey) as [String: Int]?
        _ = await session.readyValue(connectRulesKey) as [CatalogConnectRule]?
        guard rows != nil, fieldsReady != nil, workspaceReady != nil else {
            struct CatalogHandleMissing: Error {}
            throw CatalogHandleMissing()
        }
    }

    private func warmTermHandles() async {
        for property in availableProperties where property.valueType == "term" {
            await warmTerms(for: property.id)
        }
    }

    private func warmTerms(for propertyID: String) async {
        let key = termsKey(propertyID: propertyID)
        let _: QueryHandle<[CatalogPropertyTerm]> = session.query(key)
        _ = await session.readyValue(key) as [CatalogPropertyTerm]?
    }

    func selectPendingArtifact(_ id: String) {
        pendingArtifactID = id
    }

    func requestSelectArtifact(_ id: String) {
        Task { await selectArtifactAndLoad(id) }
    }

    func selectArtifactAndLoad(_ id: String) async {
        guard artifacts.contains(where: { $0.id == id }), id != selectedArtifactID else { return }
        if isEditingExisting, isDirty {
            pendingArtifactAbandon = ArtifactAbandon(targetArtifactID: id)
            return
        }
        await applyArtifactSwitch(id)
    }

    func confirmAbandonArtifact() {
        guard let pending = pendingArtifactAbandon else { return }
        pendingArtifactAbandon = nil
        Task { await applyArtifactSwitch(pending.targetArtifactID) }
    }

    func confirmAbandonArtifactAndLoad() async {
        guard let pending = pendingArtifactAbandon else { return }
        pendingArtifactAbandon = nil
        await applyArtifactSwitch(pending.targetArtifactID)
    }

    func cancelAbandonArtifact() {
        pendingArtifactAbandon = nil
    }

    func selectCitation(_ id: String?) {
        Task { await selectCitationAndLoad(id) }
    }

    func selectNewCitation() async {
        await selectCitationAndLoad(nil)
    }

    func selectCitationAndLoad(_ id: String?) async {
        if id == nil || id?.isEmpty == true {
            resetToNewCitation()
            applyEntryOverlay()
            await presentDocument(announce: true)
            return
        }
        guard let id else { return }
        do {
            let (citation, _, citationObservations) = try await store.getCitation(
                projectDir: session.projectKey.projectDir,
                citationID: id
            )
            replaceDocument(
                persisted: citation,
                observations: citationObservations,
                subjectLabelsByID: Dictionary(uniqueKeysWithValues: graphSubjects.map { ($0.id, $0.label) })
            )
            applyEntryOverlay()
            await presentDocument(announce: true)
        } catch {
            formError = L10n.Errors.message(for: error)
        }
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
        appendDraftObservation(subjectID: defaultObservationSubjectID)
    }

    func updateObservationSubject(id: UUID, subjectID: String) {
        mutateObservation(id: id) { row in
            guard !row.isConnectFixed else { return }
            row.subjectID = subjectID
            let allowed = Set(propertyOptions(for: subjectID).map(\.value))
            if !allowed.contains(row.propertyID) {
                clearObservationValue(&row)
                row.propertyID = ""
            }
        }
        focusedObservationID = id
    }

    func updateObservationProperty(id: UUID, propertyID: String) {
        mutateObservation(id: id) { row in
            guard !row.isConnectFixed else { return }
            let previousType = catalogProperty(id: row.propertyID)?.valueType
            row.propertyID = propertyID
            let nextType = catalogProperty(id: propertyID)?.valueType
            if previousType != nextType {
                clearObservationValue(&row)
            }
        }
        focusedObservationID = id
    }

    func updateObservationText(id: UUID, text: String) {
        mutateObservation(id: id) { $0.valueText = text }
    }

    func updateObservationInteger(id: UUID, text: String) {
        mutateObservation(id: id) { $0.valueIntegerText = text }
    }

    func updateObservationTerm(id: UUID, termID: String) {
        mutateObservation(id: id) { $0.valueTermID = termID }
    }

    func toggleObservationPolarity(id: UUID) {
        mutateObservation(id: id) { row in
            guard !row.isConnectFixed else { return }
            row.polarity = row.polarity == "negative" ? "positive" : "negative"
        }
    }

    func beginEditObservation(_ row: ObservationRow) {
        guard !row.isConnectFixed else { return }
        let type = catalogProperty(id: row.propertyID)?.valueType
        guard type == "name" || type == "date" else { return }
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
            persistedID: observations.first(where: { $0.id == draft.editingID })?.persistedID,
            subjectID: observations.first(where: { $0.id == draft.editingID })?.subjectID
                ?? defaultObservationSubjectID,
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
           let index = document.observations.firstIndex(where: { $0.id == editingID })
        {
            document.observations[index] = row
        } else {
            document.observations.append(row)
        }
        observationDialog = nil
        formError = nil
    }

    func removeObservation(id: UUID) {
        document.observations.removeAll { $0.id == id && !$0.isConnectFixed }
        formError = nil
    }

    func beginAddCustomTerm(rowID: UUID) {
        termError = nil
        pendingCustomTermRowID = rowID
        customTermLabel = ""
        showCustomTermDialog = true
    }

    func cancelCustomTermDialog() {
        showCustomTermDialog = false
        pendingCustomTermRowID = nil
        customTermLabel = ""
        termError = nil
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
            session.apply(.createdPropertyTerm(propertyId: propertyID))
            let key = termsKey(propertyID: propertyID)
            let handle: QueryHandle<[CatalogPropertyTerm]>? = session.queryHandle(key)
            var list = handle?.value ?? []
            if !list.contains(where: { $0.id == term.id }) {
                list.append(term)
                list.sort { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
                session.setQueryValue(key, value: list)
            }
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
        if isConnectPrefill, !observations.contains(where: \.isConnectFixed) {
            formError = String(localized: L10n.CitationComposer.saveNeedsObservationError)
            return nil
        }
        let updates: [CatalogObservation]?
        let drafts: [CatalogObservationDraft]?
        if activeCitationID == nil {
            updates = nil
            drafts = buildObservationDrafts()
            guard drafts != nil else { return nil }
        } else {
            drafts = nil
            updates = buildObservationUpdates()
            guard updates != nil else { return nil }
        }

        isSubmitting = true
        defer { isSubmitting = false }
        do {
            if let citationID = activeCitationID, let updates {
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
                    observations: updates
                )
            } else if let drafts {
                if isConnectPrefill,
                   let fromID = entry.connectFromSubjectID,
                   let toID = entry.connectToSubjectID,
                   let bridgeKey = entry.connectBridgeTypeKey
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
                        gridX: entry.connectGridX,
                        gridY: entry.connectGridY,
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
            }
            session.apply(.savedCitation(sourceId: sourceID))
            session.noticeToast = VocabularyToast(
                title: String(localized: L10n.CitationComposer.saveToastTitle),
                body: String(localized: L10n.CitationComposer.saveToastBody),
                tone: .success
            )
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

    /// Hydrate a persisted Citation. Does not read Entry.
    private func replaceDocument(
        persisted citation: CatalogCitation,
        observations listed: [CatalogObservation],
        subjectLabelsByID: [String: String]
    ) {
        resetTransientChrome()
        var propertiesByID = allPropertiesByID
        for property in availableProperties + Array(connectEdgePropertiesByID.values) {
            propertiesByID[property.id] = property
        }
        document = Document(
            artifactID: citation.artifactID,
            citationID: citation.id,
            transcription: citation.transcription,
            transcriptionUncertain: citation.transcriptionUncertain,
            transcriptionNote: citation.transcriptionNote,
            description: citation.description,
            locator: CitationLocatorDraft.decode(citation.locatorJSON),
            observations: listed.compactMap {
                observationRow(
                    from: $0,
                    propertiesByID: propertiesByID,
                    subjectLabelsByID: subjectLabelsByID
                )
            }
        )
    }

    private func presentDocument(announce: Bool) async {
        if artifacts.isEmpty {
            document.artifactID = nil
            artifactViewer.unload()
        } else {
            await reloadArtifactViewer(preferredPage: document.locator.page)
            await reloadListedCitations()
        }
        if announce {
            announceIdentityChange()
        }
        captureBaseline()
    }

    /// The only place Entry is read after catalog load.
    private func applyEntryOverlay() {
        switch entry {
        case .edit(_, let subjectID, _, _, let observationID):
            if let observationID,
               let match = document.observations.first(where: { $0.persistedID == observationID })
            {
                focusedObservationID = match.id
            } else if let match = document.observations.first(where: { $0.subjectID == subjectID }) {
                focusedObservationID = match.id
            } else {
                focusedObservationID = document.observations.first?.id
            }
        case .addProperty(_, let subjectID, _):
            guard document.citationID != nil,
                  !document.observations.contains(where: { $0.subjectID == subjectID })
            else {
                focusedObservationID = document.observations.first(where: { $0.subjectID == subjectID })?.id
                    ?? document.observations.first?.id
                return
            }
            appendDraftObservation(subjectID: subjectID)
        case .connect:
            if document.citationID == nil {
                document.observations = entryConnectPrefill
                focusedObservationID = document.observations.first?.id
            } else {
                focusedObservationID = document.observations.first?.id
            }
        }
    }

    private func applySelectedArtifact(_ id: String) async {
        document.artifactID = id
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
            if row.propertyID.isEmpty { continue }
            guard let property = catalogProperty(id: row.propertyID) else {
                formError = String(localized: L10n.CitationComposer.missingPropertyError)
                return nil
            }
            let rowSubjectID = row.subjectID
            if !row.isConnectFixed, rowSubjectID.isEmpty {
                formError = String(localized: L10n.CitationComposer.dialogPropertyRequired)
                return nil
            }
            var draft = CatalogObservationDraft(
                subjectID: rowSubjectID,
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

    /// Observations for an update. Existing rows keep their id. A row with no persisted id is new.
    private func buildObservationUpdates() -> [CatalogObservation]? {
        var updates: [CatalogObservation] = []
        updates.reserveCapacity(observations.count)
        for row in observations {
            if row.propertyID.isEmpty { continue }
            guard let property = catalogProperty(id: row.propertyID) else {
                formError = String(localized: L10n.CitationComposer.missingPropertyError)
                return nil
            }
            let rowSubjectID = row.subjectID
            if !row.isConnectFixed, rowSubjectID.isEmpty {
                formError = String(localized: L10n.CitationComposer.dialogPropertyRequired)
                return nil
            }
            var filled = CatalogObservationDraft(
                subjectID: rowSubjectID,
                propertyID: property.id,
                polarity: row.polarity == "negative" ? "negative" : "positive"
            )
            if let failure = CitationObservationValue.apply(
                valueType: property.valueType,
                fields: CitationObservationValue.fields(from: row),
                to: &filled
            ) {
                formError = CitationObservationValue.message(for: failure)
                return nil
            }
            updates.append(CatalogObservation(
                id: row.persistedID ?? "",
                ref: "",
                citationID: activeCitationID ?? "",
                subjectID: rowSubjectID,
                propertyID: property.id,
                polarity: filled.polarity,
                valueText: filled.valueText,
                valueInteger: filled.valueInteger,
                valueDateID: filled.valueDateID,
                date: filled.date,
                valueNameID: filled.valueNameID,
                nameForm: filled.nameForm,
                nameParts: filled.nameParts,
                valueSubjectID: filled.valueSubjectID,
                valueTermID: filled.valueTermID,
                propertyKey: property.key,
                propertyLabel: property.label,
                propertyValueType: property.valueType
            ))
        }
        return updates
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

    private func observationRow(
        from observation: CatalogObservation,
        propertiesByID: [String: CatalogProperty],
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
        let propertyKey = propertiesByID[observation.propertyID]?.key ?? observation.propertyKey
        let isFixed = Self.isConnectEdgeProperty(
            propertyKey: propertyKey,
            subjectTypeKey: typeKey(forSubject: observation.subjectID),
            rules: connectRules
        )
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
            persistedID: observation.id,
            subjectID: observation.subjectID,
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

    private func typeKey(forSubject id: String) -> String {
        if let match = graphSubjects.first(where: { $0.id == id }) {
            return match.typeKey
        }
        if !id.isEmpty, id == subjectID {
            return subjectTypeKey
        }
        return ""
    }

    static func isConnectEdgeProperty(
        propertyKey: String,
        subjectTypeKey: String,
        rules: [CatalogConnectRule]
    ) -> Bool {
        excludedEdgePropertyKeys(typeKey: subjectTypeKey, rules: rules).contains(propertyKey)
    }

    private func applyArtifactSwitch(_ id: String) async {
        resetToNewCitation()
        await applySelectedArtifact(id)
        await presentDocument(announce: true)
    }

    private func resetToNewCitation() {
        resetTransientChrome()
        let artifactID = document.artifactID
        document = .blank(artifactID: artifactID)
        if isConnectPrefill {
            document.observations = entryConnectPrefill
        }
        focusedObservationID = document.observations.first?.id
    }

    private func resetTransientChrome() {
        observationDialog = nil
        cancelCustomTermDialog()
        focusedObservationID = nil
        formError = nil
        submitAttempted = false
        armedRegionTool = nil
        pendingArtifactAbandon = nil
    }

    private func appendDraftObservation(subjectID: String) {
        let row = ObservationRow(
            id: UUID(),
            subjectID: subjectID,
            propertyID: "",
            polarity: "positive",
            valueText: "",
            valueIntegerText: "",
            valueTermID: "",
            valueSubjectID: "",
            dateDraft: .empty(),
            isConnectFixed: false
        )
        document.observations.append(row)
        focusedObservationID = row.id
        formError = nil
    }

    private func mutateObservation(id: UUID, _ body: (inout ObservationRow) -> Void) {
        guard let index = document.observations.firstIndex(where: { $0.id == id }) else { return }
        body(&document.observations[index])
    }

    private func clearObservationValue(_ row: inout ObservationRow) {
        row.valueText = ""
        row.valueIntegerText = ""
        row.valueTermID = ""
        row.valueSubjectID = ""
        row.dateDraft = .empty()
        row.nameDraft = .empty()
    }

    private func reloadListedCitations() async {
        guard let artifactID = selectedArtifactID else { return }
        let key = citationsKey(artifactID: artifactID)
        let _: QueryHandle<[CatalogListedCitation]> = session.query(key)
        _ = await session.readyValue(key) as [CatalogListedCitation]?
    }

    private func currentSnapshot() -> FormSnapshot {
        FormSnapshot(
            artifactID: selectedArtifactID,
            citationID: activeCitationID,
            transcription: transcription,
            transcriptionUncertain: transcriptionUncertain,
            transcriptionNote: transcriptionNote,
            description: citationDescription,
            locator: locator,
            observations: observations
        )
    }

    private func captureBaseline() {
        baseline = currentSnapshot()
    }

    private func announceIdentityChange() {
        if let active = activeCitationID,
           let listed = listedCitations.first(where: { $0.id == active })
        {
            identityAnnouncement = L10n.CitationComposer.identityChangedCitation(
                ref: listed.citation.ref,
                count: listed.observationCount
            )
        } else if let artifact = selectedArtifact {
            identityAnnouncement = L10n.CitationComposer.identityChangedArtifact(title: artifact.label)
        } else {
            identityAnnouncement = String(localized: L10n.CitationComposer.identityChangedNew)
        }
    }

    private static func graphSubjectOptions(in snapshot: SourceGraphSnapshot) -> [GraphSubjectOption] {
        var options: [GraphSubjectOption] = []
        for placed in snapshot.subjects {
            options.append(
                GraphSubjectOption(
                    id: placed.id,
                    label: placed.subject.label,
                    ref: placed.subject.ref,
                    typeID: placed.subject.subjectTypeID,
                    typeKey: placed.kind.rawValue
                )
            )
        }
        for bridge in snapshot.bridges {
            options.append(
                GraphSubjectOption(
                    id: bridge.id,
                    label: bridge.subject.label,
                    ref: bridge.subject.ref,
                    typeID: bridge.subject.subjectTypeID,
                    typeKey: bridge.kind.rawValue
                )
            )
        }
        return options
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
