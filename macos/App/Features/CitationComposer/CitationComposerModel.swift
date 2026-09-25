import Foundation
import Observation

/// Citation composer coordinator: identity, leave guard, and child models.
@MainActor
@Observable
final class CitationComposerModel {
    enum Phase: Equatable {
        case loading
        case compose
        case subjectMissing
        case loadFailed
    }

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
                polarity: ObservationPolarity.positive.rawValue,
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

    struct NewSubjectDraft: Identifiable, Equatable {
        var id = UUID()
        var typeKey: String
        var typeID: String
        var label: String = ""
        var description: String = ""
        var rowID: UUID
        var error: String?
        var isSaving = false
    }

    struct PendingLeave: Identifiable, Equatable {
        var id = UUID()
        var citationDirty: Bool
        var observationCount: Int
        var connectionTouched: Bool
    }

    struct GraphSubjectOption: Identifiable, Equatable {
        var id: String
        var label: String
        var ref: String
        var typeID: String
        var typeKey: String
    }

    static let supportedValueTypes: Set<String> = [
        PropertyValueType.text.rawValue,
        PropertyValueType.integer.rawValue,
        PropertyValueType.date.rawValue,
        PropertyValueType.term.rawValue,
        PropertyValueType.name.rawValue,
    ]

    static let newSubjectPrefix = "__new__."

    let entry: CitationComposerEntry
    private let userID: String
    private let store: any GenealogyStore
    let session: WorkspaceSession
    @ObservationIgnored
    weak var navigation: WorkspaceNavigation?
    var pendingLeave: PendingLeave?

    let fields = CitationFieldsDraft()
    let observationRows = CitationObservationRows()
    let connections = CitationConnections()
    private(set) var vocabulary = CitationComposerVocabulary.empty

    var sourceID: String { entry.sourceID }
    var subjectID: String { entry.subjectID }

    private(set) var phase: Phase = .loading
    private(set) var subjectLabel: String = ""
    private(set) var subjectTypeKey: String = ""
    private(set) var subjectTypeID: String = ""
    var sourceTitle: String { workspace?.source.title ?? "" }
    var sourceTypeIconKey: String {
        sourceTypes.first { $0.id == workspace?.source.sourceTypeID }?.iconKey ?? ""
    }
    var artifacts: [CatalogArtifact] { workspace?.artifacts ?? [] }
    var listedCitations: [CatalogListedCitation] {
        guard let artifactID = selectedArtifactID else { return [] }
        let handle: QueryHandle<[CatalogListedCitation]>? = session.queryHandle(
            citationsKey(artifactID: artifactID)
        )
        return handle?.value ?? []
    }

    private(set) var identityAnnouncement: String = ""
    var landingColumn: Int64?
    var observationDialog: ObservationDialogState?
    var newSubjectDraft: NewSubjectDraft?
    var termError: String?
    var pendingCustomTermRowID: UUID?
    var customTermLabel = ""
    var showCustomTermDialog = false
    private(set) var loadError: String?
    private(set) var shouldFallbackToGraph = false
    var selectedArtifactID: String?
    private var identityTask: Task<Void, Never>?

    let artifactViewer = ArtifactViewerModel()
    var armedRegionTool: ArtifactRegionTool?
    var viewerPage: Int { artifactViewer.page }
    var viewerPageCount: Int { artifactViewer.pageCount }
    var locatorCapabilities: ArtifactLocatorCapabilities {
        artifactViewer.locatorCapabilities
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
        if case .edit(_, _, _, let artifactID, _) = entry {
            selectedArtifactID = artifactID
        } else if case .addProperty(_, _, let artifactID) = entry {
            selectedArtifactID = artifactID
        }
        fields.attach(self)
        observationRows.attach(self)
        connections.attach(self)
    }

    var graphSubjects: [GraphSubjectOption] { vocabulary.graphSubjects }
    var observations: [ObservationRow] { observationRows.rows }
    var focusedObservationID: UUID? { observationRows.focusedID }
    var locator: CitationLocatorDraft {
        get { fields.locator }
        set { fields.locator = newValue }
    }
    var transcription: String {
        get { fields.transcription }
        set { fields.transcription = newValue }
    }
    var transcriptionUncertain: Bool {
        get { fields.transcriptionUncertain }
        set { fields.transcriptionUncertain = newValue }
    }
    var transcriptionNote: String {
        get { fields.transcriptionNote }
        set { fields.transcriptionNote = newValue }
    }
    var citationDescription: String {
        get { fields.description }
        set { fields.description = newValue }
    }
    var activeCitationID: String?
    var isEditingExisting: Bool { activeCitationID != nil }
    var showsArtifactSwitcher: Bool { artifacts.count > 1 }
    var hasNoArtifacts: Bool { phase == .compose && artifacts.isEmpty }
    var formIsInert: Bool { hasNoArtifacts }
    var selectedArtifact: CatalogArtifact? {
        artifacts.first { $0.id == selectedArtifactID }
    }
    var selectedArtifactIndex: Int? {
        guard let id = selectedArtifactID else { return nil }
        return artifacts.firstIndex { $0.id == id }
    }
    var isPDFArtifact: Bool { artifactViewer.kind == .pdf }
    var isImageArtifact: Bool { artifactViewer.kind == .image }
    var hasLocator: Bool { selectedArtifactID != nil }
    var isLocatorArtifactOnly: Bool { locator.isArtifactOnly }
    var isLocatorPageSetOnViewer: Bool {
        guard let page = locator.page else { return false }
        return page == artifactViewer.page
    }
    var locatorPage: Int? { locator.page }
    var activeCitationRef: String {
        if let id = activeCitationID,
           let listed = listedCitations.first(where: { $0.id == id })
        {
            return listed.citation.ref
        }
        return ""
    }

    var citationCountsByArtifact: [String: Int] {
        let handle: QueryHandle<[String: Int]>? = session.queryHandle(citationCountsKey)
        return handle?.value ?? [:]
    }

    func listedCount(for artifactID: String) -> Int {
        citationCountsByArtifact[artifactID, default: 0]
    }

    var identityMenusDisabled: Bool { hasUnsavedDocumentWork }

    var hasUnsavedDocumentWork: Bool {
        fields.isDirty || observationRows.hasEditedOrSaving || connections.hasTouchedWork
    }

    var unsavedSummary: String {
        L10n.CitationComposer.unsavedSummary(
            citationDirty: fields.isDirty,
            observationCount: observationRows.rows.filter {
                if case .saved = $0.state { return false }
                return !$0.isEmptyDraft
            }.count,
            connectionTouched: connections.hasTouchedWork
        )
    }

    var graphHandle: QueryHandle<SourceGraphRows>? {
        session.queryHandle(graphKey)
    }

    var isGraphReloading: Bool {
        graphHandle?.status == .loading || (graphHandle?.isFetching ?? false)
    }

    var subjectOptions: [PVComboBoxOption] {
        let existing = vocabulary.graphSubjects.map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.ref)
        }
        let news: [PVComboBoxOption] = ["person", "event", "place"].compactMap { key in
            guard vocabulary.typesByID.values.contains(where: { $0.key == key }) else { return nil }
            return PVComboBoxOption(
                value: Self.newSubjectPrefix + key,
                label: L10n.CitationComposer.newSubject(typeKey: key)
            )
        }
        return existing + news
    }

    var defaultObservationSubjectID: String {
        if !subjectID.isEmpty { return subjectID }
        return vocabulary.graphSubjects.first?.id ?? ""
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
        return draft.propertyID.isEmpty
            ? String(localized: L10n.CitationComposer.dialogPropertyRequired)
            : nil
    }

    var dialogValueError: String? {
        guard let draft = observationDialog, draft.showValidation, dialogPropertyError == nil
        else { return nil }
        return canConfirmObservation ? nil : String(localized: L10n.CitationComposer.dialogValueRequired)
    }

    func catalogProperty(id: String) -> CatalogProperty? { vocabulary.property(id: id) }

    func propertyOptions(for subjectID: String) -> [PVComboBoxOption] {
        vocabulary.propertyOptions(forSubjectID: subjectID).map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    func termOptions(for propertyID: String) -> [PVComboBoxOption] {
        let propertyKey = catalogProperty(id: propertyID)?.key ?? ""
        return (vocabulary.termsByPropertyID[propertyID] ?? []).map { term in
            PVComboBoxOption(
                value: term.id,
                label: PropertyTermDisplay.name(term: term, propertyKey: propertyKey),
                subtext: term.key
            )
        }
    }

    func observationSummary(for row: ObservationRow) -> String {
        guard let property = catalogProperty(id: row.propertyID) else { return "" }
        return CitationObservationValue.summary(
            valueType: property.valueType,
            fields: row.valueFields(),
            termLabel: vocabulary.termLabel(termID: row.valueTermID, propertyID: property.id)
        )
    }

    var graphKey: CatalogQueryKey { .sourceGraph(project: session.projectKey, sourceId: sourceID) }
    var workspaceKey: CatalogQueryKey { .sourceWorkspace(project: session.projectKey, sourceId: sourceID) }
    var citationCountsKey: CatalogQueryKey { .citationCounts(project: session.projectKey, sourceId: sourceID) }
    var fieldsKey: CatalogQueryKey { .subjectFieldsWorkspace(project: session.projectKey) }
    var connectRulesKey: CatalogQueryKey { .connectRules(project: session.projectKey) }
    var sourceTypesKey: CatalogQueryKey { .sourceTypesList(project: session.projectKey) }

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

    private var fieldsSnapshot: SubjectFieldsSnapshot {
        let handle: QueryHandle<SubjectFieldsSnapshot>? = session.queryHandle(fieldsKey)
        return handle?.value ?? .empty
    }

    var graphSnapshot: SourceGraphSnapshot {
        let handle: QueryHandle<SourceGraphRows>? = session.queryHandle(graphKey)
        let rulesHandle: QueryHandle<[CatalogConnectRule]>? = session.queryHandle(connectRulesKey)
        return SourceGraphSnapshot.build(
            rows: handle?.value ?? SourceGraphRows(sourceId: sourceID),
            types: fieldsSnapshot.types,
            rules: rulesHandle?.value ?? []
        )
    }

    func prepare() async {
        phase = .loading
        loadError = nil
        shouldFallbackToGraph = false
        landingColumn = nil
        do {
            try await warmCatalogHandles()
            rebuildVocabulary()
            if entry.citationID == nil, entry.subjectID.isEmpty, !entry.isConnect {
                phase = .subjectMissing
                shouldFallbackToGraph = true
                return
            }
            if !entry.subjectID.isEmpty,
               vocabulary.graphSubjects.first(where: { $0.id == entry.subjectID }) == nil,
               !entry.isConnect
            {
                phase = .subjectMissing
                shouldFallbackToGraph = true
                return
            }
            if let existing = vocabulary.graphSubjects.first(where: { $0.id == entry.subjectID }) {
                subjectLabel = existing.label
                subjectTypeKey = existing.typeKey
                subjectTypeID = existing.typeID
            } else if let key = entry.connectBridgeTypeKey,
                      let type = fieldsSnapshot.types.first(where: { $0.key == key })
            {
                subjectTypeKey = key
                subjectTypeID = type.id
            }
            await warmTermHandles()
            rebuildVocabulary()
            seedPendingConnectionIfNeeded()
            if let citationID = entry.citationID {
                try await loadCitation(citationID)
            } else {
                activeCitationID = nil
                fields.resetBlank()
                observationRows.replace([])
                connections.clearSavedKeepingPending()
            }
            applyEntryFocus()
            if artifacts.isEmpty {
                selectedArtifactID = nil
            } else if artifacts.count == 1 {
                await applySelectedArtifact(artifacts[0].id)
            } else if let workspace {
                selectedArtifactID = Self.defaultPendingArtifactID(
                    artifacts: artifacts,
                    citationCounts: citationCountsByArtifact,
                    selectedID: selectedArtifactID,
                    coverMode: workspace.source.coverMode,
                    primaryArtifactID: workspace.source.primaryArtifactID
                )
                if let id = selectedArtifactID {
                    await applySelectedArtifact(id)
                }
            }
            await reloadListedCitations()
            phase = .compose
        } catch {
            loadError = L10n.Errors.message(for: error)
            phase = .loadFailed
        }
    }

    func requestSelectArtifact(_ id: String) {
        startIdentitySwitch { await self.selectArtifactAndLoad(id) }
    }

    func selectCitation(_ id: String?) {
        startIdentitySwitch { await self.selectCitationAndLoad(id) }
    }

    func awaitIdentitySwitch() async {
        await identityTask?.value
    }

    func beginAddObservation() {
        observationRows.addDraft(subjectID: defaultObservationSubjectID)
    }

    func updateObservationSubject(id: UUID, subjectID: String) {
        if subjectID.hasPrefix(Self.newSubjectPrefix) {
            let key = String(subjectID.dropFirst(Self.newSubjectPrefix.count))
            beginNewSubject(typeKey: key, rowID: id)
            return
        }
        observationRows.applySubject(rowID: id, subjectID: subjectID, vocabulary: vocabulary)
    }

    func updateObservationProperty(id: UUID, propertyID: String) {
        observationRows.applyProperty(rowID: id, propertyID: propertyID)
    }

    func updateObservationText(id: UUID, text: String) {
        observationRows.applyText(rowID: id, text: text)
    }

    func updateObservationInteger(id: UUID, text: String) {
        observationRows.applyInteger(rowID: id, text: text)
    }

    func updateObservationTerm(id: UUID, termID: String) {
        observationRows.applyTerm(rowID: id, termID: termID)
    }

    func toggleObservationPolarity(id: UUID) {
        observationRows.togglePolarity(rowID: id)
    }

    func beginEditObservation(_ row: ObservationRow) {
        observationDialog = .editing(row)
    }

    func cancelObservationDialog() {
        observationDialog = nil
    }

    func confirmObservationDialog() {
        guard let draft = observationDialog, let rowID = draft.editingID else { return }
        observationRows.applyDialogValues(rowID: rowID, draft: draft)
        observationDialog = nil
    }

    func updateObservationDialog(_ draft: ObservationDialogState) {
        observationDialog = draft
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
            await warmTerms(for: propertyID)
            rebuildVocabulary()
            termError = nil
            return term
        } catch {
            termError = L10n.Errors.message(for: error)
            return nil
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

    @discardableResult
    func setPageFromViewer() -> Bool { setPage(artifactViewer.page) }

    @discardableResult
    func setPage(_ page: Int) -> Bool {
        locator.setPage(page, capabilities: locatorCapabilities)
    }

    @discardableResult
    func setRegion(_ draft: ArtifactRegionDraft) -> Bool {
        let ok = locator.setRegion(draft, capabilities: locatorCapabilities, autoPage: artifactViewer.page)
        if ok { armedRegionTool = nil }
        return ok
    }

    func clearRegion() { locator.clearRegion() }
    func resetToEntireArtifact() {
        locator.resetToEntireArtifact()
        armedRegionTool = nil
    }
    func removePage() { locator.removePage() }
    func removeRegion() { locator.clearRegion() }
    func disarmRegionTool() { armedRegionTool = nil }
    func clearLocator() { resetToEntireArtifact() }

    func graphLocation() -> WorkspaceLocation {
        WorkspaceLocation(section: .sources, sourceId: sourceID, sourceSurface: .graph)
    }

    func sourcePageLocation() -> WorkspaceLocation {
        WorkspaceLocation(section: .sources, sourceId: sourceID, sourceSurface: .page)
    }

    func beginNewSubject(typeKey: String, rowID: UUID) {
        guard !isGraphReloading,
              let type = fieldsSnapshot.types.first(where: { $0.key == typeKey })
        else { return }
        newSubjectDraft = NewSubjectDraft(typeKey: typeKey, typeID: type.id, rowID: rowID)
    }

    func cancelNewSubject() { newSubjectDraft = nil }

    func confirmNewSubject() async {
        guard var draft = newSubjectDraft, !draft.isSaving else { return }
        draft.isSaving = true
        newSubjectDraft = draft
        let slot = EvidenceGraphPlacement.composerSlot(in: graphSnapshot, landingColumn: landingColumn)
        do {
            let created = try await store.createSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                sourceID: sourceID,
                subjectTypeID: draft.typeID,
                label: draft.label,
                description: draft.description,
                placement: slot.cell
            )
            landingColumn = slot.landingColumn
            session.apply(.mutatedSourceGraph(sourceId: sourceID))
            _ = await session.readyValue(graphKey) as SourceGraphRows?
            rebuildVocabulary()
            observationRows.applySubject(rowID: draft.rowID, subjectID: created.id, vocabulary: vocabulary)
            newSubjectDraft = nil
        } catch {
            draft.error = L10n.Errors.message(for: error)
            draft.isSaving = false
            newSubjectDraft = draft
        }
    }

    func performSaveCitation() async {
        guard let artifactID = selectedArtifactID else {
            fields.error = String(localized: L10n.CitationComposer.needArtifact)
            return
        }
        do {
            if let citationID = activeCitationID {
                _ = try await store.updateCitation(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    citationID: citationID,
                    locatorJSON: fields.locator.encodeJSON(),
                    transcription: fields.transcription,
                    description: fields.description,
                    transcriptionUncertain: fields.transcriptionUncertain,
                    transcriptionNote: fields.transcriptionNote
                )
            } else {
                let created = try await store.createCitationWithObservations(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    artifactID: artifactID,
                    locatorJSON: fields.locator.encodeJSON(),
                    transcription: fields.transcription,
                    description: fields.description,
                    transcriptionUncertain: fields.transcriptionUncertain,
                    transcriptionNote: fields.transcriptionNote,
                    citationNotes: [],
                    observations: []
                )
                activeCitationID = created.0.id
                await reloadListedCitations()
            }
            fields.captureBaseline()
            session.apply(.savedCitation(sourceId: sourceID))
        } catch {
            fields.error = L10n.Errors.message(for: error)
        }
    }

    func performCommitObservation(rowID: UUID) async {
        guard let index = observationRows.rows.firstIndex(where: { $0.id == rowID }) else { return }
        let row = observationRows.rows[index]
        guard let property = catalogProperty(id: row.propertyID) else {
            observationRows.markError(rowID: rowID, message: String(localized: L10n.CitationComposer.missingPropertyError))
            return
        }
        let (draftOrNil, draftError) = row.asDraft(property: property)
        if let draftError {
            observationRows.markError(rowID: rowID, message: draftError)
        } else if let draft = draftOrNil {
            do {
                if activeCitationID == nil {
                    guard let artifactID = selectedArtifactID else {
                        observationRows.markError(
                            rowID: rowID,
                            message: String(localized: L10n.CitationComposer.needArtifact)
                        )
                        return
                    }
                    let created = try await store.createCitationWithObservations(
                        projectDir: session.projectKey.projectDir,
                        userID: userID,
                        artifactID: artifactID,
                        locatorJSON: fields.locator.encodeJSON(),
                        transcription: fields.transcription,
                        description: fields.description,
                        transcriptionUncertain: fields.transcriptionUncertain,
                        transcriptionNote: fields.transcriptionNote,
                        citationNotes: [],
                        observations: [draft]
                    )
                    activeCitationID = created.0.id
                    fields.captureBaseline()
                    if let written = created.1.first {
                        observationRows.markSaved(rowID: rowID, id: written.id, ref: written.ref)
                    }
                    await reloadListedCitations()
                } else if row.persistedID == nil, let citationID = activeCitationID {
                    let written = try await store.addObservationsToCitation(
                        projectDir: session.projectKey.projectDir,
                        userID: userID,
                        citationID: citationID,
                        observations: [draft]
                    )
                    if let written = written.first {
                        observationRows.markSaved(rowID: rowID, id: written.id, ref: written.ref)
                    }
                } else if let citationID = activeCitationID {
                    let (observationOrNil, updateError) = row.asCatalogObservation(property: property)
                    if let updateError {
                        observationRows.markError(rowID: rowID, message: updateError)
                        return
                    }
                    if var observation = observationOrNil {
                        observation.citationID = citationID
                        let written = try await store.updateObservation(
                            projectDir: session.projectKey.projectDir,
                            userID: userID,
                            observation: observation
                        )
                        observationRows.markSaved(rowID: rowID, id: written.id, ref: written.ref)
                    }
                }
                session.apply(.savedCitation(sourceId: sourceID))
            } catch {
                observationRows.markError(rowID: rowID, message: L10n.Errors.message(for: error))
            }
        }
    }

    func performDeleteObservation(rowID: UUID) async {
        guard let row = observationRows.rows.first(where: { $0.id == rowID }),
              let persistedID = row.persistedID
        else { return }
        do {
            try await store.deleteObservation(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                observationID: persistedID
            )
            observationRows.remove(rowID: rowID)
            session.apply(.savedCitation(sourceId: sourceID))
        } catch {
            observationRows.markError(rowID: rowID, message: L10n.Errors.message(for: error))
        }
    }

    func performSaveConnection() async {
        guard let pending = connections.rows.first(where: \.isPending),
              let drafts = connections.edgeDrafts(for: pending, vocabulary: vocabulary)
        else { return }
        do {
            let attachID = activeCitationID
            let created = try await store.createCitedBridge(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                sourceID: sourceID,
                fromSubjectID: pending.fromSubjectID,
                toSubjectID: pending.toSubjectID,
                bridgeTypeKey: pending.bridgeTypeKey,
                description: "",
                artifactID: attachID == nil ? (selectedArtifactID ?? "") : "",
                locatorJSON: attachID == nil ? fields.locator.encodeJSON() : "",
                transcription: attachID == nil ? fields.transcription : "",
                citationDescription: attachID == nil ? fields.description : "",
                transcriptionUncertain: attachID == nil && fields.transcriptionUncertain,
                transcriptionNote: attachID == nil ? fields.transcriptionNote : "",
                citationNotes: [],
                observations: drafts,
                citationID: attachID
            )
            if activeCitationID == nil {
                activeCitationID = created.1.id
                fields.captureBaseline()
                await reloadListedCitations()
            }
            let role = created.2.first { obs in
                pending.termProperty.map { obs.propertyID == $0.id } ?? false
            }
            connections.markPendingSaved(
                ConnectionRow(
                    id: pending.id,
                    isPending: false,
                    fromSubjectID: pending.fromSubjectID,
                    toSubjectID: pending.toSubjectID,
                    fromLabel: pending.fromLabel,
                    toLabel: pending.toLabel,
                    bridgeTypeKey: pending.bridgeTypeKey,
                    bridgeID: created.0.id,
                    bridgeRef: created.0.ref,
                    sentence: vocabulary.graphSubjects.first { $0.id == created.0.id }?.label
                        ?? pending.sentence,
                    termProperty: pending.termProperty,
                    rolePersistedID: role?.id,
                    rolePersistedRef: role?.ref,
                    roleTermID: role?.valueTermID ?? pending.roleTermID,
                    roleBaselineTermID: role?.valueTermID ?? pending.roleTermID,
                    termTouched: false,
                    isSaving: false,
                    error: nil
                )
            )
            session.apply(.savedCitation(sourceId: sourceID))
            rebuildVocabulary()
        } catch {
            connections.markError(connectionID: pending.id, message: L10n.Errors.message(for: error))
        }
    }

    func performCommitRole(connectionID: UUID) async {
        guard let row = connections.rows.first(where: { $0.id == connectionID }),
              let citationID = activeCitationID,
              let termProperty = row.termProperty,
              let bridgeID = row.bridgeID
        else { return }
        do {
            if let roleID = row.rolePersistedID {
                let observation = CatalogObservation(
                    id: roleID,
                    ref: row.rolePersistedRef ?? "",
                    citationID: citationID,
                    subjectID: bridgeID,
                    propertyID: termProperty.id,
                    polarity: ObservationPolarity.positive.rawValue,
                    valueText: "",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: "",
                    valueTermID: row.roleTermID,
                    propertyKey: termProperty.key,
                    propertyLabel: termProperty.label,
                    propertyValueType: termProperty.valueType
                )
                let written = try await store.updateObservation(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    observation: observation
                )
                connections.updateRole(
                    connectionID: connectionID,
                    persistedID: written.id,
                    persistedRef: written.ref,
                    termID: written.valueTermID
                )
            } else {
                let written = try await store.addObservationsToCitation(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    citationID: citationID,
                    observations: [
                        CatalogObservationDraft(
                            subjectID: bridgeID,
                            propertyID: termProperty.id,
                            polarity: ObservationPolarity.positive.rawValue,
                            valueTermID: row.roleTermID
                        ),
                    ]
                )
                if let written = written.first {
                    connections.updateRole(
                        connectionID: connectionID,
                        persistedID: written.id,
                        persistedRef: written.ref,
                        termID: written.valueTermID
                    )
                }
            }
            session.apply(.savedCitation(sourceId: sourceID))
        } catch {
            connections.markError(connectionID: connectionID, message: L10n.Errors.message(for: error))
        }
    }

    private func startIdentitySwitch(_ work: @escaping @MainActor () async -> Void) {
        identityTask?.cancel()
        identityTask = Task { @MainActor in
            await work()
        }
    }

    private func selectArtifactAndLoad(_ id: String) async {
        guard artifacts.contains(where: { $0.id == id }), id != selectedArtifactID else { return }
        if Task.isCancelled { return }
        activeCitationID = nil
        fields.resetBlank()
        observationRows.replace([])
        connections.clearSavedKeepingPending()
        if Task.isCancelled { return }
        await applySelectedArtifact(id)
        if Task.isCancelled { return }
        await presentAfterIdentityChange()
    }

    private func selectCitationAndLoad(_ id: String?) async {
        if id == nil || id?.isEmpty == true {
            activeCitationID = nil
            fields.resetBlank(keepLocator: true)
            observationRows.replace([])
            connections.clearSavedKeepingPending()
            if Task.isCancelled { return }
            applyEntryFocus()
            await presentAfterIdentityChange()
            return
        }
        guard let id else { return }
        do {
            try await loadCitation(id)
            if Task.isCancelled { return }
            applyEntryFocus()
            await presentAfterIdentityChange()
        } catch {
            if Task.isCancelled { return }
            fields.error = L10n.Errors.message(for: error)
        }
    }

    private func loadCitation(_ citationID: String) async throws {
        let (citation, _, listed) = try await store.getCitation(
            projectDir: session.projectKey.projectDir,
            citationID: citationID
        )
        if Task.isCancelled { return }
        rebuildVocabulary()
        let grouped = CitationConnections.groupSaved(
            observations: listed,
            vocabulary: vocabulary,
            snapshot: graphSnapshot
        )
        activeCitationID = citation.id
        selectedArtifactID = citation.artifactID
        fields.replace(from: citation)
        connections.replaceSaved(grouped.connections)
        observationRows.replace(grouped.leftover.map(ObservationRow.loaded(from:)))
    }

    private func seedPendingConnectionIfNeeded() {
        guard case let .connect(_, fromID, toID, bridgeKey) = entry,
              connections.rows.first(where: \.isPending) == nil
        else { return }
        let from = vocabulary.graphSubjects.first { $0.id == fromID }
        let to = vocabulary.graphSubjects.first { $0.id == toID }
        let rule = vocabulary.connectRule(bridgeTypeKey: bridgeKey)
        let termProperty: CatalogProperty?
        if let rule, rule.disambiguation != "none", !rule.disambiguation.isEmpty {
            termProperty = vocabulary.property(key: rule.disambiguation)
        } else {
            termProperty = nil
        }
        let sentence = EvidenceBridgeEdgeSummary.sentence(
            kind: EvidenceBridgeKind(rawValue: bridgeKey) ?? .location,
            person: from?.typeKey == "person" ? from?.label : to?.typeKey == "person" ? to?.label : nil,
            related: bridgeKey == "relationship" ? to?.label : nil,
            event: from?.typeKey == "event" ? from?.label : to?.typeKey == "event" ? to?.label : nil,
            place: from?.typeKey == "place" ? from?.label : to?.typeKey == "place" ? to?.label : nil,
            term: nil
        )
        connections.seedPending(
            fromSubjectID: fromID,
            toSubjectID: toID,
            fromLabel: from?.label ?? fromID,
            toLabel: to?.label ?? toID,
            bridgeTypeKey: bridgeKey,
            termProperty: termProperty,
            sentence: sentence
        )
    }

    private func applyEntryFocus() {
        switch entry {
        case .edit(_, let subjectID, _, _, let observationID):
            if let observationID,
               let match = observationRows.rows.first(where: { $0.persistedID == observationID })
            {
                observationRows.focusedID = match.id
            } else {
                observationRows.focusedID = observationRows.rows.first { $0.subjectID == subjectID }?.id
                    ?? observationRows.rows.first?.id
            }
        case .addProperty(_, let subjectID, _):
            if activeCitationID != nil,
               !observationRows.rows.contains(where: { $0.subjectID == subjectID })
            {
                observationRows.addDraft(subjectID: subjectID)
            } else {
                observationRows.focusedID = observationRows.rows.first { $0.subjectID == subjectID }?.id
                    ?? observationRows.rows.first?.id
            }
        case .connect:
            break
        }
    }

    private func applySelectedArtifact(_ id: String) async {
        selectedArtifactID = id
        fields.locator = .artifactOnly()
        await reloadArtifactViewer(preferredPage: nil)
    }

    private func presentAfterIdentityChange() async {
        if artifacts.isEmpty {
            selectedArtifactID = nil
            artifactViewer.unload()
        } else {
            await reloadArtifactViewer(preferredPage: fields.locator.page)
            await reloadListedCitations()
        }
        announceIdentityChange()
    }

    private func reloadArtifactViewer(preferredPage: Int?) async {
        guard let artifact = selectedArtifact else {
            artifactViewer.unload()
            return
        }
        let relPath = (artifact.file?.relPath ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        let mediaType = artifact.file?.mediaType ?? ""
        await artifactViewer.load(
            ArtifactViewerSource(
                projectDir: session.projectKey.projectDir,
                relPath: relPath,
                mediaType: mediaType.isEmpty ? "application/octet-stream" : mediaType
            )
        )
        if let preferredPage, artifactViewer.supportsPages {
            artifactViewer.setPage(preferredPage)
        }
        locator.peelIllegalLayers(capabilities: locatorCapabilities)
        if !locatorCapabilities.supportsRegionLocator {
            armedRegionTool = nil
        }
    }

    private func reloadListedCitations() async {
        guard let artifactID = selectedArtifactID else { return }
        let key = citationsKey(artifactID: artifactID)
        let _: QueryHandle<[CatalogListedCitation]> = session.query(key)
        _ = await session.readyValue(key) as [CatalogListedCitation]?
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
        let ids = Set(
            vocabulary.propertiesByID.values
                .filter { $0.valueType == PropertyValueType.term.rawValue }
                .map(\.id)
        )
        for id in ids {
            await warmTerms(for: id)
        }
    }

    private func warmTerms(for propertyID: String) async {
        let key = termsKey(propertyID: propertyID)
        let _: QueryHandle<[CatalogPropertyTerm]> = session.query(key)
        _ = await session.readyValue(key) as [CatalogPropertyTerm]?
    }

    private func rebuildVocabulary() {
        var terms: [String: [CatalogPropertyTerm]] = [:]
        for property in fieldsSnapshot.properties
            where property.valueType == PropertyValueType.term.rawValue
        {
            let handle: QueryHandle<[CatalogPropertyTerm]>? = session.queryHandle(termsKey(propertyID: property.id))
            if let value = handle?.value {
                terms[property.id] = value
            }
        }
        let rulesHandle: QueryHandle<[CatalogConnectRule]>? = session.queryHandle(connectRulesKey)
        vocabulary = CitationComposerVocabulary.make(
            fields: fieldsSnapshot,
            snapshot: graphSnapshot,
            rules: rulesHandle?.value ?? [],
            termsByPropertyID: terms
        )
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
        if leaders.count == 1 { return leaders[0].id }
        if let thumbnailID, leaders.contains(where: { $0.id == thumbnailID }) {
            return thumbnailID
        }
        return leaders.first(where: { $0.file != nil && !$0.fileID.isEmpty })?.id
            ?? leaders.first?.id
    }

    static func isPDF(_ artifact: CatalogArtifact?) -> Bool {
        (artifact?.file?.mediaType ?? "").localizedCaseInsensitiveContains("pdf")
    }

    static func isImage(_ artifact: CatalogArtifact?) -> Bool {
        (artifact?.file?.mediaType ?? "").hasPrefix("image/")
    }
}

extension CitationComposerModel: WorkspaceLeaveGuard {
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool {
        _ = pending
        guard hasUnsavedDocumentWork else { return false }
        pendingLeave = PendingLeave(
            citationDirty: fields.isDirty,
            observationCount: observationRows.rows.filter {
                if case .saved = $0.state { return false }
                return !$0.isEmptyDraft
            }.count,
            connectionTouched: connections.hasTouchedWork
        )
        return true
    }

    func discardLeaveChanges() {
        pendingLeave = nil
        navigation?.resumeHeldNavigation()
    }

    func keepEditingAfterLeave() {
        pendingLeave = nil
        navigation?.cancelHeldNavigation()
    }
}
