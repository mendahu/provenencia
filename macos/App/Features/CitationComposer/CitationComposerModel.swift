import AppKit
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
        var field: CitationSubjectField
        var error: String?
        var isSaving = false
    }

    struct PendingLeave: Identifiable, Equatable {
        var id = UUID()
        var citationDirty: Bool
        var observationCount: Int
        var connectionTouched: Bool
    }

    struct PendingTranscriptionConfirm: Identifiable, Equatable {
        var id = UUID()
        var existingText: String
        var includeWholePage: Bool
        var pastePage: Int? = nil
        var pasteLineCount: Int? = nil

        var isPaste: Bool { pastePage != nil }
    }

    struct CustomTermDraft: Equatable {
        var propertyID: String
        var label: String = ""
        var error: String?
        var isSaving = false
        var applyToEdit = false
    }

    enum ComposerModal: Equatable {
        case observation(ObservationDialogState)
        case customTerm(CustomTermDraft)
        case newSubject(NewSubjectDraft)
        case deleteImpact
        case leave(PendingLeave)
        case transcription(PendingTranscriptionConfirm)
    }

    enum CatalogWarmError: Error {
        case handleMissing
    }

    struct GraphSubjectOption: Identifiable, Equatable {
        var id: String
        var label: String
        var ref: String
        var typeID: String
        var typeKey: String
    }

    nonisolated static let newSubjectPrefix = "__new__."

    let entry: CitationComposerEntry
    let userID: String
    let store: any GenealogyStore
    let session: WorkspaceSession
    let context: CitationComposerContext
    @ObservationIgnored
    weak var navigation: WorkspaceNavigation?
    var modal: ComposerModal?
    var pendingLeave: PendingLeave? {
        get {
            if case .leave(let value) = modal { return value }
            return nil
        }
        set {
            if let newValue {
                present(.leave(newValue))
            } else if case .leave = modal {
                present(nil)
            }
        }
    }
    let deleteImpact = DeleteImpactFlow()
    var pendingImpact: PVDeleteImpactRequest? {
        get { deleteImpact.request }
        set { applyDeleteImpactRequest(newValue) }
    }
    var isDeletingResource: Bool { deleteImpact.isRunning }
    private var loadedCitationRef = ""

    let fields = CitationFieldsDraft()
    let observationRows = CitationObservationRows()
    let connections = CitationConnections()
    var authoring = CitationAuthoring.rest
    var authoringFocusNonce = 0
    var freshLineKey: String?
    var bridgeToast: (title: String, message: String)?
    var editSession: CitationEditSession?
    var editSheetSuspended = false
    var newCitationReturnIndex: Int?
    var stepContinuation: (() -> Void)?
    private var didApplyEntry = false
    let transcriptions: CitationTranscriptionModel
    var vocabulary: CitationComposerVocabulary { context.vocabulary }

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
    var artifactIDs: [String] { artifacts.map(\.id) }
    var listedCitations: [CatalogListedCitation] {
        guard let artifactID = selectedArtifactID else { return [] }
        let handle: QueryHandle<[CatalogListedCitation]>? = session.queryHandle(
            citationsKey(artifactID: artifactID)
        )
        return handle?.value ?? []
    }

    private(set) var identityAnnouncement: String = ""
    var landingColumn: Int64?
    var observationDialog: ObservationDialogState? {
        get {
            if case .observation(let value) = modal { return value }
            return nil
        }
        set {
            if let newValue {
                present(.observation(newValue))
            } else if case .observation = modal {
                observationDialogBaseline = nil
                present(nil)
            }
        }
    }
    var newSubjectDraft: NewSubjectDraft? {
        get {
            if case .newSubject(let value) = modal { return value }
            return nil
        }
        set {
            if let newValue {
                present(.newSubject(newValue))
            } else if case .newSubject = modal {
                present(nil)
            }
        }
    }
    var customTermDraft: CustomTermDraft? {
        get {
            if case .customTerm(let value) = modal { return value }
            return nil
        }
        set {
            if let newValue {
                present(.customTerm(newValue))
            } else if case .customTerm = modal {
                present(nil)
            }
        }
    }
    var pendingTranscriptionConfirm: PendingTranscriptionConfirm? {
        get {
            if case .transcription(let value) = modal { return value }
            return nil
        }
        set {
            if let newValue {
                present(.transcription(newValue))
            } else if case .transcription = modal {
                present(nil)
            }
        }
    }
    private(set) var loadError: String?
    private(set) var loadFailureIsMissingCitation = false
    var selectedArtifactID: String? {
        get { context.artifactID }
        set { context.artifactID = newValue }
    }
    private var identityTask: Task<Void, Never>?

    let artifactViewer = ArtifactViewerModel()
    var armedRegionTool: ArtifactRegionTool?
    var viewerPage: Int { artifactViewer.page }
    var viewerPageCount: Int { artifactViewer.pageCount }
    var locatorCapabilities: ArtifactLocatorCapabilities {
        artifactViewer.locatorCapabilities
    }
    var isTranscribing: Bool { transcriptions.isTranscribing }
    var transcriptionOCRMessage: String? { transcriptions.ocrMessage }

    init(
        entry: CitationComposerEntry,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String,
        ocrEngine: any OCREngine = VisionOCREngine()
    ) {
        self.entry = entry
        self.session = session
        self.store = store
        self.userID = userID
        self.context = CitationComposerContext(
            store: store,
            session: session,
            userID: userID,
            sourceID: entry.sourceID
        )
        if case .edit(_, _, _, let artifactID, _) = entry {
            context.artifactID = artifactID
        } else if case .addProperty(_, _, let artifactID) = entry {
            context.artifactID = artifactID
        }
        fields.attach(context)
        observationRows.attach(context)
        connections.attach(context)
        self.transcriptions = CitationTranscriptionModel(
            fields: fields,
            artifactViewer: artifactViewer,
            ocrEngine: ocrEngine
        )
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
        set {
            transcriptions.noteUserEditedTranscription()
            fields.transcription = newValue
        }
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
    var activeCitationID: String? {
        get { context.citationID }
        set { context.citationID = newValue }
    }
    var isEditingExisting: Bool { activeCitationID != nil }
    var showsArtifactSwitcher: Bool { artifacts.count > 1 }
    var hasNoArtifacts: Bool { phase == .compose && artifacts.isEmpty }
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
        return loadedCitationRef
    }

    var citationCountsByArtifact: [String: Int] {
        let handle: QueryHandle<[String: Int]>? = session.queryHandle(citationCountsKey)
        return handle?.value ?? [:]
    }

    func listedCount(for artifactID: String) -> Int {
        citationCountsByArtifact[artifactID, default: 0]
    }

    var identityMenusDisabled: Bool { hasUnsavedDocumentWork || isTranscribing }

    var hasUnsavedDocumentWork: Bool {
        fields.isDirty || authoring.holdsLeave || (editSession?.isDirty ?? false)
    }

    var shouldHoldLeave: Bool {
        hasUnsavedDocumentWork || isTranscribing
    }

    var imageRaster: NSImage? { transcriptions.imageRaster }

    var canAutoTranscribe: Bool { transcriptions.canAutoTranscribe }

    var canPasteTranscription: Bool { transcriptions.canPasteTranscription }

    var transcriptionActionHint: String { transcriptions.actionHint }

    var autoTranscribeHint: String { transcriptions.autoTranscribeHint }

    var needsWholePageWarning: Bool { transcriptions.needsWholePageWarning }

    var unsavedSummary: String {
        if authoring.holdsLeave {
            let detail = leaveDraftDetail
            let ref = activeCitationRef.isEmpty
                ? L10n.string(L10n.CitationComposer.newCitation)
                : activeCitationRef
            return L10n.CitationComposer.leaveDraft(detail: detail, ref: ref)
        }
        return L10n.CitationComposer.unsavedSummary(
            citationDirty: fields.isDirty,
            observationCount: (editSession?.isDirty ?? false) ? 1 : 0,
            connectionTouched: false
        )
    }

    private var leaveDraftDetail: String {
        switch authoring {
        case .bridge(let draft):
            if case .ready(let ready) = CitationComposerAuthoring.resolvePair(
                endA: draft.endA,
                endB: draft.endB,
                termID: draft.termID,
                vocabulary: vocabulary
            ), let reading = ready.reading {
                return reading
            }
            return L10n.string(L10n.CitationComposer.newBridgeTitle)
        case .property:
            return L10n.string(L10n.CitationComposer.newObservationTitle)
        case .rest:
            return L10n.string(L10n.CitationComposer.unsavedDraft)
        }
    }

    var graphHandle: QueryHandle<SourceGraphRows>? {
        session.queryHandle(graphKey)
    }

    var isGraphReloading: Bool {
        graphHandle?.status == .loading || (graphHandle?.isFetching ?? false)
    }

    var subjectOptions: [PVComboBoxOption] { vocabulary.subjectComboOptions }

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
            ? L10n.string(L10n.CitationComposer.dialogPropertyRequired)
            : nil
    }

    var dialogValueError: String? {
        guard let draft = observationDialog, draft.showValidation, dialogPropertyError == nil
        else { return nil }
        return canConfirmObservation ? nil : L10n.string(L10n.CitationComposer.dialogValueRequired)
    }

    func catalogProperty(id: String) -> CatalogProperty? { vocabulary.property(id: id) }

    func propertyOptions(for subjectID: String) -> [PVComboBoxOption] {
        vocabulary.propertyComboOptions(forSubjectID: subjectID)
    }

    func termOptions(for propertyID: String) -> [PVComboBoxOption] {
        vocabulary.termComboOptionsByPropertyID[propertyID] ?? []
    }

    func observationSummary(for row: ObservationRow) -> String {
        guard let property = catalogProperty(id: row.propertyID) else { return "" }
        return CitationObservationValue.summary(
            valueType: property.valueType,
            fields: row.valueFields(),
            termLabel: vocabulary.termLabel(termID: row.valueTermID, propertyID: property.id),
            orders: session.nameFormat
        )
    }

    var graphKey: CatalogQueryKey { .sourceGraph(project: session.projectKey, sourceId: sourceID) }
    var workspaceKey: CatalogQueryKey { .sourceWorkspace(project: session.projectKey, sourceId: sourceID) }
    var citationCountsKey: CatalogQueryKey { .citationCounts(project: session.projectKey, sourceId: sourceID) }
    var fieldsKey: CatalogQueryKey { .propertiesWorkspace(project: session.projectKey) }
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

    private var fieldsSnapshot: PropertiesSnapshot { context.fieldsSnapshot }

    var graphSnapshot: SourceGraphSnapshot { context.graphSnapshot }

    func prepare() async {
        phase = .loading
        loadError = nil
        loadFailureIsMissingCitation = false
        landingColumn = nil
        do {
            try await warmCatalogHandles()
            if entry.citationID == nil, entry.subjectID.isEmpty, !entry.isConnect {
                phase = .subjectMissing
                return
            }
            if !entry.subjectID.isEmpty,
               vocabulary.graphSubjects.first(where: { $0.id == entry.subjectID }) == nil,
               !entry.isConnect
            {
                phase = .subjectMissing
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
            if let citationID = entry.citationID {
                do {
                    try await loadCitation(citationID)
                } catch {
                    resetComposerIdentity(keepLocator: false)
                    loadFailureIsMissingCitation = true
                    loadError = L10n.string(L10n.CitationComposer.citationMissing)
                    phase = .loadFailed
                    return
                }
            } else {
                resetComposerIdentity(keepLocator: false)
            }
            if let selected = selectedArtifactID, !artifacts.contains(where: { $0.id == selected }) {
                resetComposerIdentity(keepLocator: false)
            }
            applyEntryFocus()
            if artifacts.isEmpty {
                selectedArtifactID = nil
                resetComposerIdentity(keepLocator: false)
            } else if activeCitationID != nil {
                // Loaded citation already has artifact + locator; do not reset to artifact-only.
                await reloadArtifactViewer(preferredPage: fields.locator.page)
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
        } catch is CatalogWarmError {
            loadFailureIsMissingCitation = false
            loadError = L10n.string(L10n.CitationComposer.catalogUnavailable)
            phase = .loadFailed
        } catch {
            loadFailureIsMissingCitation = false
            loadError = L10n.Errors.message(for: error)
            phase = .loadFailed
        }
    }

    func requestSelectArtifact(_ id: String) {
        requestStep { await self.selectArtifactAndLoad(id) }
    }

    func selectCitation(_ id: String?) {
        if id == nil || id?.isEmpty == true {
            beginNewCitation()
            return
        }
        requestStep { await self.selectCitationAndLoad(id) }
    }

    func beginNewCitation() {
        let ordered = CitationComposerAuthoring.stepperOrder(listedCitations)
        let index = ordered.firstIndex { $0.citation.id == activeCitationID }
        requestStep {
            self.newCitationReturnIndex = index
            await self.selectCitationAndLoad(nil)
        }
    }

    func stepArtifact(by delta: Int) {
        let ids = artifactIDs
        guard let current = selectedArtifactID, let index = ids.firstIndex(of: current) else { return }
        let next = index + delta
        guard ids.indices.contains(next) else { return }
        let id = ids[next]
        requestStep { await self.selectArtifactAndLoad(id) }
    }

    func stepCitation(by delta: Int) {
        let ordered = CitationComposerAuthoring.stepperOrder(listedCitations)
        if activeCitationID == nil {
            let index = min(max(newCitationReturnIndex ?? 0, 0), max(ordered.count - 1, 0))
            guard ordered.indices.contains(index) else { return }
            let id = ordered[index].citation.id
            requestStep { await self.selectCitationAndLoad(id) }
            return
        }
        guard let current = activeCitationID,
              let index = ordered.firstIndex(where: { $0.citation.id == current })
        else { return }
        let next = index + delta
        guard ordered.indices.contains(next) else { return }
        let id = ordered[next].citation.id
        requestStep { await self.selectCitationAndLoad(id) }
    }

    var canStepArtifactBackward: Bool {
        guard let current = selectedArtifactID, let index = artifactIDs.firstIndex(of: current) else { return false }
        return index > 0
    }

    var canStepArtifactForward: Bool {
        guard let current = selectedArtifactID, let index = artifactIDs.firstIndex(of: current) else { return false }
        return index + 1 < artifactIDs.count
    }

    var canStepCitationBackward: Bool {
        if activeCitationID == nil { return !listedCitations.isEmpty }
        let ordered = CitationComposerAuthoring.stepperOrder(listedCitations)
        guard let index = ordered.firstIndex(where: { $0.citation.id == activeCitationID }) else { return false }
        return index > 0
    }

    var canStepCitationForward: Bool {
        if activeCitationID == nil { return false }
        let ordered = CitationComposerAuthoring.stepperOrder(listedCitations)
        guard let index = ordered.firstIndex(where: { $0.citation.id == activeCitationID }) else { return false }
        return index + 1 < ordered.count
    }

    func requestStep(_ work: @escaping @MainActor () async -> Void) {
        guard !isTranscribing else { return }
        if hasUnsavedDocumentWork {
            stepContinuation = { self.startIdentitySwitch(work) }
            pendingLeave = PendingLeave(
                citationDirty: fields.isDirty,
                observationCount: (authoring.holdsLeave || (editSession?.isDirty ?? false)) ? 1 : 0,
                connectionTouched: false
            )
            return
        }
        startIdentitySwitch(work)
    }

    func refreshOpenCitation() async {
        if let id = activeCitationID {
            try? await loadCitation(id)
        }
        await presentAfterIdentityChange()
    }

    func askDeleteCitation() async {
        guard let id = activeCitationID else { return }
        let snippet = fields.transcription.trimmingCharacters(in: .whitespacesAndNewlines)
        await presentImpact(
            kind: "citation",
            id: id,
            ref: activeCitationRef.isEmpty ? id : activeCitationRef,
            title: snippet.isEmpty ? L10n.string(L10n.CitationComposer.untitledCitation) : snippet
        )
    }

    func askDeleteObservation(rowID: UUID) async {
        guard let row = observationRows.rows.first(where: { $0.id == rowID }),
              let persistedID = row.persistedID
        else { return }
        let ref = row.persistedRef ?? persistedID
        await presentImpact(kind: "observation", id: persistedID, ref: ref, title: ref)
    }

    func confirmPendingImpact() async {
        _ = await deleteImpact.confirm { target in
            if target.kind == "citation" {
                let effects = try await store.deleteCitation(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    citationID: target.id
                )
                context.applyWrite(effects)
                resetComposerIdentity(keepLocator: true)
                await reloadListedCitations()
                await presentAfterIdentityChange()
                return
            }
            if target.kind == "subject" {
                try await store.deleteSubject(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    subjectID: target.id
                )
                session.apply(.deletedSubject(sourceId: sourceID))
                if let id = activeCitationID {
                    try await loadCitation(id)
                }
                await presentAfterIdentityChange()
                return
            }
            if target.kind == "observation",
               let rowID = observationRows.rows.first(where: { $0.persistedID == target.id })?.id
            {
                await observationRows.deletePersisted(rowID: rowID)
                await refreshOpenCitation()
            }
        }
        if let message = deleteImpact.error {
            fields.error = message
        }
        if deleteImpact.request == nil, case .deleteImpact = modal {
            modal = nil
        }
    }

    func awaitIdentitySwitch() async {
        await identityTask?.value
    }

    func beginAddObservation() {
        openAuthoring { self.beginPropertyDraft(subjectID: "") }
    }

    func updateObservationSubject(id: UUID, subjectID: String) {
        if subjectID.hasPrefix(Self.newSubjectPrefix) {
            let key = String(subjectID.dropFirst(Self.newSubjectPrefix.count))
            beginNewSubject(typeKey: key, field: .propertySubject)
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
        var state = ObservationDialogState.editing(row)
        // The user's locale until a source carries its own language.
        if catalogProperty(id: row.propertyID)?.valueType == PropertyValueType.date.rawValue,
           state.dateDraft.isBlank,
           let match = DateValueTranscriptionMatch.first(in: fields.transcription, locale: .current) {
            state.dateDraft.startYear = match.year
            state.dateDraft.startMonth = match.month
            state.dateDraft.startDay = match.day
        }
        observationDialogBaseline = state
        observationDialog = state
    }

    /// Seed the dialog opened with, including a transcription prefill. Edits
    /// after that count; the prefill itself does not.
    var observationDialogBaseline: ObservationDialogState?

    var observationDialogIsDirty: Bool {
        guard let dialog = observationDialog else { return false }
        let baseline: ObservationDialogState
        if let observationDialogBaseline {
            baseline = observationDialogBaseline
        } else if let id = dialog.editingID,
                  let row = observationRows.rows.first(where: { $0.id == id }) {
            baseline = ObservationDialogState.editing(row)
        } else {
            return false
        }
        var current = dialog
        var opened = baseline
        current.showValidation = false
        opened.showValidation = false
        return current != opened
    }

    var customTermDialogIsDirty: Bool {
        let label = customTermDraft?.label.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        return !label.isEmpty
    }

    var newSubjectDialogIsDirty: Bool {
        guard let draft = newSubjectDraft else { return false }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let description = draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
        return !label.isEmpty || !description.isEmpty
    }

    func cancelObservationDialog() {
        observationDialog = nil
        if editSheetSuspended {
            editSheetSuspended = false
        }
    }

    func confirmObservationDialog() {
        guard let draft = observationDialog else { return }
        if editSheetSuspended {
            applyValueDialog(draft, toEdit: true)
            editSheetSuspended = false
            observationDialog = nil
            return
        }
        if case .property = authoring {
            applyValueDialog(draft, toEdit: false)
            observationDialog = nil
            return
        }
        if let rowID = draft.editingID {
            observationRows.applyDialogValues(rowID: rowID, draft: draft)
        }
        observationDialog = nil
    }

    private func applyValueDialog(_ draft: ObservationDialogState, toEdit: Bool) {
        func write(_ values: inout ObservationRowValues) {
            values.valueText = draft.valueText
            values.valueIntegerText = draft.valueIntegerText
            values.valueTermID = draft.valueTermID
            values.dateDraft = draft.dateDraft
            values.nameDraft = draft.nameDraft
        }
        if toEdit, var session = editSession, case .property(var values, let persistedID, let persistedRef) = session.kind {
            write(&values)
            session.kind = .property(values: values, persistedID: persistedID, persistedRef: persistedRef)
            editSession = session
            return
        }
        if case .property(var property) = authoring {
            write(&property.values)
            authoring = .property(property)
        }
    }

    func updateObservationDialog(_ draft: ObservationDialogState) {
        observationDialog = draft
    }

    func beginAddCustomTerm(propertyID: String, applyToEdit: Bool = false) {
        editSheetSuspended = true
        present(.customTerm(CustomTermDraft(propertyID: propertyID, applyToEdit: applyToEdit)))
    }

    func cancelCustomTermDialog() {
        if case .customTerm = modal { present(nil) }
        editSheetSuspended = false
    }

    func confirmCustomTerm() async {
        guard case .customTerm(var draft) = modal else { return }
        let propertyID = draft.propertyID
        guard !propertyID.isEmpty else {
            present(nil)
            return
        }
        let trimmed = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        draft.isSaving = true
        draft.error = nil
        present(.customTerm(draft))
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
            if draft.applyToEdit, var session = editSession {
                switch session.kind {
                case .property(var values, let persistedID, let persistedRef):
                    values.valueTermID = term.id
                    session.kind = .property(values: values, persistedID: persistedID, persistedRef: persistedRef)
                case .bridge(let endA, let endB, _, let bridgeID, let bridgeRef, let baseline):
                    session.kind = .bridge(
                        endA: endA,
                        endB: endB,
                        termID: term.id,
                        bridgeID: bridgeID,
                        bridgeRef: bridgeRef,
                        baselineTermID: baseline
                    )
                }
                editSession = session
            } else if case .property(var property) = authoring {
                property.values.valueTermID = term.id
                authoring = .property(property)
            } else if case .bridge(var bridge) = authoring {
                bridge.termID = term.id
                authoring = .bridge(bridge)
            }
            editSheetSuspended = false
            present(nil)
        } catch {
            draft.error = L10n.Errors.message(for: error)
            draft.isSaving = false
            present(.customTerm(draft))
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
        guard !isTranscribing else { return false }
        return locator.setPage(page, capabilities: locatorCapabilities)
    }

    @discardableResult
    func setRegion(_ draft: ArtifactRegionDraft) -> Bool {
        guard !isTranscribing else { return false }
        let ok = locator.setRegion(draft, capabilities: locatorCapabilities, viewerPage: artifactViewer.page)
        if ok { armedRegionTool = nil }
        return ok
    }

    func clearRegion() {
        guard !isTranscribing else { return }
        locator.clearRegion()
    }
    func resetToEntireArtifact() {
        guard !isTranscribing else { return }
        locator.resetToEntireArtifact()
        armedRegionTool = nil
    }
    func removePage() {
        guard !isTranscribing else { return }
        locator.removePage()
    }
    func removeRegion() {
        guard !isTranscribing else { return }
        locator.clearRegion()
    }
    func disarmRegionTool() { armedRegionTool = nil }
    func clearLocator() { resetToEntireArtifact() }

    func requestAutoTranscribe() {
        guard canAutoTranscribe else { return }
        if let pending = transcriptions.autoTranscribeConfirmIfNeeded() {
            present(.transcription(pending))
            return
        }
        Task { await transcriptions.runAutoTranscribe() }
    }

    func confirmAutoTranscribe() {
        present(nil)
        Task { await transcriptions.runAutoTranscribe() }
    }

    func cancelAutoTranscribeConfirm() {
        if case .transcription = modal { present(nil) }
    }

    func requestPasteTranscription() {
        guard canPasteTranscription else { return }
        if let pending = transcriptions.pasteConfirmIfNeeded() {
            present(.transcription(pending))
            return
        }
        transcriptions.applyPasteTranscription()
    }

    func confirmPasteTranscription() {
        present(nil)
        transcriptions.applyPasteTranscription()
    }

    func dismissTranscriptionOCRMessage() {
        transcriptions.dismissOCRMessage()
    }

    func runAutoTranscribe() async {
        await transcriptions.runAutoTranscribe()
    }

    func graphLocation() -> WorkspaceLocation {
        WorkspaceLocation(section: .sources, sourceId: sourceID, sourceSurface: .graph)
    }

    func sourcePageLocation() -> WorkspaceLocation {
        WorkspaceLocation(section: .sources, sourceId: sourceID, sourceSurface: .page)
    }

    func beginNewSubject(typeKey: String, field: CitationSubjectField) {
        guard !isGraphReloading,
              let type = fieldsSnapshot.types.first(where: { $0.key == typeKey })
        else { return }
        editSheetSuspended = true
        newSubjectDraft = NewSubjectDraft(typeKey: typeKey, typeID: type.id, field: field)
    }

    func cancelNewSubject() {
        newSubjectDraft = nil
        editSheetSuspended = false
    }

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
            await context.waitForGraph()
            assignCreatedSubject(created.id, field: draft.field)
            newSubjectDraft = nil
            editSheetSuspended = false
        } catch {
            draft.error = L10n.Errors.message(for: error)
            draft.isSaving = false
            newSubjectDraft = draft
        }
    }

    private func resetComposerIdentity(keepLocator: Bool) {
        activeCitationID = nil
        loadedCitationRef = ""
        fields.resetBlank(keepLocator: keepLocator)
        observationRows.replace([])
        connections.clearSavedKeepingPending()
        authoring = .rest
        editSession = nil
        editSheetSuspended = false
    }

    private func present(_ next: ComposerModal?) {
        let leavingDelete: Bool
        if case .deleteImpact = modal { leavingDelete = true } else { leavingDelete = false }
        let enteringDelete: Bool
        if case .deleteImpact = next { enteringDelete = true } else { enteringDelete = false }
        if leavingDelete && !enteringDelete {
            deleteImpact.cancel()
        }
        if next != nil && !enteringDelete {
            deleteImpact.cancel()
        }
        modal = next
    }

    func applyDeleteImpactRequest(_ value: PVDeleteImpactRequest?) {
        deleteImpact.applyRequest(value)
        if value == nil {
            if case .deleteImpact = modal { modal = nil }
        } else {
            modal = .deleteImpact
        }
    }

    func presentImpact(kind: String, id: String, ref: String, title: String) async {
        fields.error = nil
        present(nil)
        await deleteImpact.ask(kind: kind, id: id, ref: ref, title: title) {
            try await self.store.getDeleteImpact(
                projectDir: self.session.projectKey.projectDir,
                kind: kind,
                id: id
            )
        }
        if deleteImpact.request != nil {
            modal = .deleteImpact
        } else if let message = deleteImpact.error {
            fields.error = message
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
        authoring = .rest
        editSession = nil
        activeCitationID = nil
        fields.resetBlank()
        observationRows.replace([])
        connections.clearSavedKeepingPending()
        if Task.isCancelled { return }
        await applySelectedArtifact(id)
        if Task.isCancelled { return }
        await reloadListedCitations()
        let ordered = CitationComposerAuthoring.stepperOrder(listedCitations)
        if let first = ordered.first {
            try? await loadCitation(first.citation.id)
        }
        if Task.isCancelled { return }
        await presentAfterIdentityChange()
    }

    private func selectCitationAndLoad(_ id: String?) async {
        authoring = .rest
        editSession = nil
        if id == nil || id?.isEmpty == true {
            activeCitationID = nil
            fields.resetBlank(keepLocator: true)
            observationRows.replace([])
            connections.clearSavedKeepingPending()
            if Task.isCancelled { return }
            await presentAfterIdentityChange()
            return
        }
        guard let id else { return }
        do {
            try await loadCitation(id)
            if Task.isCancelled { return }
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
        await context.waitForGraph()
        if Task.isCancelled { return }
        let grouped = CitationConnections.groupSaved(
            observations: listed,
            vocabulary: vocabulary,
            snapshot: graphSnapshot
        )
        activeCitationID = citation.id
        loadedCitationRef = citation.ref
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
        let sentence: String
        if let rule {
            sentence = ConnectEndpointBinding.pendingSentence(
                rule: rule,
                fromID: fromID,
                fromTypeKey: from?.typeKey ?? "",
                fromLabel: from?.label ?? fromID,
                toID: toID,
                toTypeKey: to?.typeKey ?? "",
                toLabel: to?.label ?? toID
            )
        } else {
            sentence = ""
        }
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
        guard !didApplyEntry else { return }
        didApplyEntry = true
        switch entry {
        case .addProperty(_, let subjectID, _):
            guard activeCitationID != nil else { return }
            beginPropertyDraft(subjectID: subjectID)
        case .connect(_, let fromID, let toID, _):
            beginBridgeDraft(endA: fromID, endB: toID, termID: "", origin: .graph)
        case .edit:
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
        await context.reloadListedCitations()
    }

    private func warmCatalogHandles() async throws {
        let _: QueryHandle<SourceGraphRows> = session.query(graphKey)
        let _: QueryHandle<PropertiesSnapshot> = session.query(fieldsKey)
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
        let _: QueryHandle<[CatalogSourceType]> = session.query(sourceTypesKey)
        let _: QueryHandle<[String: Int]> = session.query(citationCountsKey)
        let _: QueryHandle<[CatalogConnectRule]> = session.query(connectRulesKey)
        let rows: SourceGraphRows? = await session.readyValue(graphKey)
        let fieldsReady: PropertiesSnapshot? = await session.readyValue(fieldsKey)
        let workspaceReady: CatalogSourceWorkspace? = await session.readyValue(workspaceKey)
        _ = await session.readyValue(sourceTypesKey) as [CatalogSourceType]?
        _ = await session.readyValue(citationCountsKey) as [String: Int]?
        _ = await session.readyValue(connectRulesKey) as [CatalogConnectRule]?
        guard rows != nil, fieldsReady != nil, workspaceReady != nil else {
            throw CatalogWarmError.handleMissing
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
            identityAnnouncement = L10n.string(L10n.CitationComposer.identityChangedNew)
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
}

extension CitationComposerModel: WorkspaceLeaveGuard {
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool {
        _ = pending
        guard shouldHoldLeave else { return false }
        pendingLeave = PendingLeave(
            citationDirty: fields.isDirty,
            observationCount: (authoring.holdsLeave || (editSession?.isDirty ?? false)) ? 1 : 0,
            connectionTouched: false
        )
        return true
    }

    func discardLeaveChanges() {
        fields.revert()
        authoring = .rest
        editSession = nil
        editSheetSuspended = false
        observationDialog = nil
        let next = stepContinuation
        stepContinuation = nil
        pendingLeave = nil
        if let next {
            next()
            return
        }
        navigation?.resumeHeldNavigation()
    }

    func keepEditingAfterLeave() {
        let stepping = stepContinuation != nil
        stepContinuation = nil
        pendingLeave = nil
        if !stepping {
            navigation?.cancelHeldNavigation()
        }
    }
}
