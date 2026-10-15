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
        /// A citation switch under an open draft. Discard replaces the citation
        /// text and leaves the draft in place.
        var keepDraft = false
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

    enum ComposerBase: Equatable {
        case none
        case authoring
        case edit
    }

    /// The one dialog on screen. A child remembers the dialog underneath it.
    /// Leave, transcription, and delete restore the exact dialog they replaced,
    /// including a custom-term or new-subject draft.
    enum ComposerPresented: Equatable {
        case authoring
        case edit
        case customTerm(CustomTermDraft, over: ComposerBase)
        case newSubject(NewSubjectDraft, over: ComposerBase)
        indirect case leave(PendingLeave, returnTo: ComposerPresented?)
        indirect case transcription(PendingTranscriptionConfirm, returnTo: ComposerPresented?)
        indirect case deleteImpact(returnTo: ComposerPresented?)
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

    var pendingLeave: PendingLeave? {
        get {
            if case .leave(let value, _) = presented { return value }
            return nil
        }
        set {
            if let newValue {
                show(.leave(newValue, returnTo: suspendedDialog()))
            } else if case .leave(_, let returnTo) = presented {
                restorePresented(returnTo)
            }
        }
    }
    let deleteImpact = DeleteImpactFlow()
    var pendingImpact: PVDeleteImpactRequest? {
        get { deleteImpact.request }
        set { applyDeleteImpactRequest(newValue) }
    }
    private var loadedCitationRef = ""

    let fields = CitationFieldsDraft()
    let observationRows = CitationObservationRows()
    let connections = CitationConnections()
    var authoring = CitationAuthoring.rest
    var freshLineKey: String?
    var bridgeToast: (title: String, message: String)?
    var editSession: CitationEditSession?
    private(set) var presented: ComposerPresented?
    var stepContinuation: (() -> Void)?
    private var didApplyEntry = false
    private var switchGeneration = 0
    var isSwitchingCitation = false
    let transcriptions: CitationTranscriptionModel
    var vocabulary: CitationComposerVocabulary { context.vocabulary }

    var sourceID: String { entry.sourceID }
    var subjectID: String { entry.subjectID }

    private(set) var phase: Phase = .loading
    private(set) var subjectLabel: String = ""
    private(set) var subjectTypeKey: String = ""
    private(set) var subjectTypeID: String = ""
    var sourceTitle: String { workspace?.source.title ?? "" }
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
    var newSubjectDraft: NewSubjectDraft? {
        get {
            if case .newSubject(let value, _) = presented { return value }
            return nil
        }
        set {
            if let newValue {
                let over: ComposerBase
                if case .newSubject(_, let existing) = presented {
                    over = existing
                } else {
                    over = dialogBase()
                }
                show(.newSubject(newValue, over: over))
            } else if case .newSubject(_, let over) = presented {
                restore(over)
            }
        }
    }
    var customTermDraft: CustomTermDraft? {
        get {
            if case .customTerm(let value, _) = presented { return value }
            return nil
        }
        set {
            if let newValue {
                let over: ComposerBase
                if case .customTerm(_, let existing) = presented {
                    over = existing
                } else {
                    over = dialogBase()
                }
                show(.customTerm(newValue, over: over))
            } else if case .customTerm(_, let over) = presented {
                restore(over)
            }
        }
    }
    var pendingTranscriptionConfirm: PendingTranscriptionConfirm? {
        get {
            if case .transcription(let value, _) = presented { return value }
            return nil
        }
        set {
            if let newValue {
                show(.transcription(newValue, returnTo: suspendedDialog()))
            } else if case .transcription(_, let returnTo) = presented {
                restorePresented(returnTo)
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
    var hasNoArtifacts: Bool { phase == .compose && artifacts.isEmpty }
    var selectedArtifact: CatalogArtifact? {
        artifacts.first { $0.id == selectedArtifactID }
    }
    var isPDFArtifact: Bool { artifactViewer.kind == .pdf }
    var isImageArtifact: Bool { artifactViewer.kind == .image }
    var isLocatorArtifactOnly: Bool { locator.isArtifactOnly }
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

    func catalogProperty(id: String) -> CatalogProperty? { vocabulary.property(id: id) }

    func propertyOptions(for subjectID: String) -> [PVComboBoxOption] {
        vocabulary.propertyComboOptions(forSubjectID: subjectID)
    }

    func termOptions(for propertyID: String) -> [PVComboBoxOption] {
        vocabulary.termComboOptionsByPropertyID[propertyID] ?? []
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
        if preservesAuthoringOnCitationChange {
            if fields.isDirty {
                guard !isTranscribing else { return }
                let target = id
                stepContinuation = { self.switchOpenCitation(to: target) }
                pendingLeave = PendingLeave(keepDraft: true)
                return
            }
            switchOpenCitation(to: id)
            return
        }
        if id == nil || id?.isEmpty == true {
            beginNewCitation()
            return
        }
        requestStep { await self.selectCitationAndLoad(id) }
    }

    func beginNewCitation() {
        requestStep { await self.selectCitationAndLoad(nil) }
    }

    func requestStep(_ work: @escaping @MainActor () async -> Void) {
        guard !isTranscribing else { return }
        if hasUnsavedDocumentWork {
            stepContinuation = { self.startIdentitySwitch(work) }
            pendingLeave = PendingLeave()
            return
        }
        startIdentitySwitch(work)
    }

    /// Reloads observations and the citation list for the open citation.
    /// Citation text, label, and the file viewer stay as they are.
    func refreshCitationLines() async {
        guard let id = activeCitationID else { return }
        if let loaded = try? await fetchOpenCitation(id) {
            if Task.isCancelled { return }
            applyGroupedLines(loaded.observations)
        }
        await reloadListedCitations()
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
                let effects = try await store.deleteSubject(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    subjectID: target.id
                )
                session.apply(effects, .deletedSubject(sourceId: sourceID))
                await refreshCitationLines()
                return
            }
            if target.kind == "observation",
               let rowID = observationRows.rows.first(where: { $0.persistedID == target.id })?.id
            {
                await observationRows.deletePersisted(rowID: rowID)
                await refreshCitationLines()
            }
        }
        if let message = deleteImpact.error {
            fields.error = message
        }
        if deleteImpact.request == nil, case .deleteImpact(let returnTo) = presented {
            restorePresented(returnTo)
        }
    }

    func awaitIdentitySwitch() async {
        await identityTask?.value
    }

    func beginAddObservation() {
        openAuthoring { self.beginPropertyDraft(subjectID: "") }
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

    func beginAddCustomTerm(propertyID: String, applyToEdit: Bool = false) {
        let over = dialogBase()
        show(.customTerm(
            CustomTermDraft(propertyID: propertyID, applyToEdit: applyToEdit),
            over: over
        ))
    }

    func cancelCustomTermDialog() {
        if case .customTerm(_, let over) = presented {
            restore(over)
        }
    }

    func confirmCustomTerm() async {
        guard case .customTerm(var draft, let over) = presented else { return }
        let propertyID = draft.propertyID
        guard !propertyID.isEmpty else {
            restore(over)
            return
        }
        let trimmed = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        draft.isSaving = true
        draft.error = nil
        show(.customTerm(draft, over: over))
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
            restore(over)
        } catch {
            draft.error = L10n.Errors.message(for: error)
            draft.isSaving = false
            show(.customTerm(draft, over: over))
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

    func requestAutoTranscribe() {
        guard canAutoTranscribe else { return }
        if let pending = transcriptions.autoTranscribeConfirmIfNeeded() {
            pendingTranscriptionConfirm = pending
            return
        }
        Task { await transcriptions.runAutoTranscribe() }
    }

    func confirmAutoTranscribe() {
        pendingTranscriptionConfirm = nil
        Task { await transcriptions.runAutoTranscribe() }
    }

    func cancelAutoTranscribeConfirm() {
        if case .transcription = presented { pendingTranscriptionConfirm = nil }
    }

    func requestPasteTranscription() {
        guard canPasteTranscription else { return }
        if let pending = transcriptions.pasteConfirmIfNeeded() {
            pendingTranscriptionConfirm = pending
            return
        }
        transcriptions.applyPasteTranscription()
    }

    func confirmPasteTranscription() {
        pendingTranscriptionConfirm = nil
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
        newSubjectDraft = NewSubjectDraft(typeKey: typeKey, typeID: type.id, field: field)
    }

    func cancelNewSubject() {
        newSubjectDraft = nil
    }

    func confirmNewSubject() async {
        guard var draft = newSubjectDraft, !draft.isSaving else { return }
        draft.isSaving = true
        newSubjectDraft = draft
        let slot = EvidenceGraphPlacement.composerSlot(in: graphSnapshot, landingColumn: landingColumn)
        do {
            let (created, effects) = try await store.createSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                sourceID: sourceID,
                subjectTypeID: draft.typeID,
                label: draft.label,
                description: draft.description,
                placement: slot.cell
            )
            landingColumn = slot.landingColumn
            session.apply(effects, .mutatedSourceGraph(sourceId: sourceID))
            await context.waitForGraph()
            assignCreatedSubject(created.id, field: draft.field)
            newSubjectDraft = nil
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
        show(nil)
    }

    /// The sheet underneath a custom-term or new-subject dialog.
    func dialogBase() -> ComposerBase {
        switch presented {
        case nil:
            if editSession != nil { return .edit }
            if case .rest = authoring { return .none }
            return .authoring
        default:
            return base(of: presented)
        }
    }

    private func base(of presented: ComposerPresented?) -> ComposerBase {
        switch presented {
        case .authoring:
            return .authoring
        case .edit:
            return .edit
        case .customTerm(_, let over), .newSubject(_, let over):
            return over
        case .leave(_, let returnTo), .transcription(_, let returnTo), .deleteImpact(let returnTo):
            return base(of: returnTo)
        case nil:
            return .none
        }
    }

    /// The dialog a leave, transcription, or delete prompt should come back to,
    /// including a child draft that was on screen.
    func suspendedDialog() -> ComposerPresented? {
        switch presented {
        case .authoring, .edit, .customTerm, .newSubject:
            return presented
        case .leave(_, let returnTo), .transcription(_, let returnTo), .deleteImpact(let returnTo):
            return returnTo
        case nil:
            return nil
        }
    }

    func restore(_ base: ComposerBase) {
        switch base {
        case .none:
            show(nil)
        case .authoring:
            if case .rest = authoring { show(nil) } else { show(.authoring) }
        case .edit:
            show(editSession == nil ? nil : .edit)
        }
    }

    func restorePresented(_ next: ComposerPresented?) {
        switch next {
        case .authoring:
            restore(.authoring)
        case .edit:
            restore(.edit)
        case .customTerm, .newSubject:
            show(next)
        case .leave, .transcription, .deleteImpact, nil:
            show(nil)
        }
    }

    func showAuthoringDialog() {
        if case .rest = authoring { return }
        show(.authoring)
    }

    func showEditDialog() {
        guard editSession != nil else { return }
        show(.edit)
    }

    func show(_ next: ComposerPresented?) {
        let leavingDelete: Bool
        if case .deleteImpact = presented { leavingDelete = true } else { leavingDelete = false }
        let enteringDelete: Bool
        if case .deleteImpact = next { enteringDelete = true } else { enteringDelete = false }
        if leavingDelete && !enteringDelete {
            deleteImpact.cancel()
        }
        if next != nil && !enteringDelete {
            deleteImpact.cancel()
        }
        presented = next
    }

    func applyDeleteImpactRequest(_ value: PVDeleteImpactRequest?) {
        deleteImpact.applyRequest(value)
        if value == nil {
            if case .deleteImpact(let returnTo) = presented {
                restorePresented(returnTo)
            }
        } else {
            show(.deleteImpact(returnTo: suspendedDialog()))
        }
    }

    func presentImpact(kind: String, id: String, ref: String, title: String) async {
        fields.error = nil
        let returnTo = suspendedDialog()
        await deleteImpact.ask(kind: kind, id: id, ref: ref, title: title) {
            try await self.store.getDeleteImpact(
                projectDir: self.session.projectKey.projectDir,
                kind: kind,
                id: id
            )
        }
        if deleteImpact.request != nil {
            show(.deleteImpact(returnTo: returnTo))
        } else if let message = deleteImpact.error {
            fields.error = message
        }
    }

    private func startIdentitySwitch(_ work: @escaping @MainActor () async -> Void) {
        switchGeneration += 1
        let generation = switchGeneration
        isSwitchingCitation = true
        identityTask?.cancel()
        identityTask = Task { @MainActor in
            await work()
            if generation == self.switchGeneration {
                self.isSwitchingCitation = false
            }
        }
    }

    private func selectArtifactAndLoad(_ id: String) async {
        guard artifacts.contains(where: { $0.id == id }), id != selectedArtifactID else { return }
        if Task.isCancelled { return }
        authoring = .rest
        editSession = nil
        show(nil)
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

    /// The add-observation dialog stays open while the citation underneath it changes.
    private var preservesAuthoringOnCitationChange: Bool {
        if editSession != nil { return false }
        if case .rest = authoring { return false }
        return true
    }

    private func switchOpenCitation(to id: String?) {
        let target = (id?.isEmpty == false) ? id : nil
        guard target != activeCitationID else { return }
        startIdentitySwitch { await self.loadCitationPreservingAuthoring(target) }
    }

    private func loadCitationPreservingAuthoring(_ id: String?) async {
        if id == nil || id?.isEmpty == true {
            openBlankCitation()
        } else if let id {
            do {
                try await loadCitation(id)
            } catch {
                if Task.isCancelled { return }
                fields.error = L10n.Errors.message(for: error)
                return
            }
        }
        if Task.isCancelled { return }
        let blank = id == nil || id?.isEmpty == true
        await presentAfterIdentityChange(keepReadingPage: blank)
    }

    /// Drops the open citation. The new one is the whole artifact; it does not
    /// inherit the previous page or region.
    private func openBlankCitation() {
        activeCitationID = nil
        loadedCitationRef = ""
        fields.resetBlank()
        observationRows.replace([])
        connections.clearSavedKeepingPending()
    }

    private func selectCitationAndLoad(_ id: String?) async {
        authoring = .rest
        editSession = nil
        if id == nil || id?.isEmpty == true {
            openBlankCitation()
            if Task.isCancelled { return }
            await presentAfterIdentityChange(keepReadingPage: true)
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

    private struct FetchedCitation {
        var citation: CatalogCitation
        var observations: [CatalogObservation]
    }

    /// Nil when this task was cancelled before the citation was read.
    private func fetchOpenCitation(_ citationID: String) async throws -> FetchedCitation? {
        await context.waitForGraph()
        if Task.isCancelled { return nil }
        if let cached = await cachedCitation(id: citationID),
           let graphHandle: QueryHandle<SourceGraphRows> = session.queryHandle(graphKey),
           let graphRows = graphHandle.value
        {
            return FetchedCitation(
                citation: cached,
                observations: graphRows.observations.filter { $0.citationID == cached.id }
            )
        }
        // Not on a file this source still has. One direct read; prepare
        // then drops a citation whose artifact is gone.
        let (fetched, _, listed) = try await store.getCitation(
            projectDir: session.projectKey.projectDir,
            citationID: citationID
        )
        if Task.isCancelled { return nil }
        return FetchedCitation(citation: fetched, observations: listed)
    }

    private func applyGroupedLines(_ observations: [CatalogObservation]) {
        let grouped = CitationConnections.groupSaved(
            observations: observations,
            vocabulary: vocabulary,
            snapshot: graphSnapshot
        )
        connections.replaceSaved(grouped.connections)
        observationRows.replace(grouped.leftover.map(ObservationRow.loaded(from:)))
    }

    private func loadCitation(_ citationID: String) async throws {
        guard let loaded = try await fetchOpenCitation(citationID) else { return }
        activeCitationID = loaded.citation.id
        loadedCitationRef = loaded.citation.ref
        selectedArtifactID = loaded.citation.artifactID
        fields.replace(from: loaded.citation)
        applyGroupedLines(loaded.observations)
    }

    /// Citation text lives on the artifact list. Observations live on the source graph.
    private func cachedCitation(id: String) async -> CatalogCitation? {
        var artifactIDs: [String] = []
        if let selectedArtifactID, !selectedArtifactID.isEmpty {
            artifactIDs.append(selectedArtifactID)
        }
        for artifact in artifacts where !artifactIDs.contains(artifact.id) {
            artifactIDs.append(artifact.id)
        }
        for artifactID in artifactIDs {
            let key = citationsKey(artifactID: artifactID)
            let _: QueryHandle<[CatalogListedCitation]> = session.query(key)
            let rows = await session.readyValue(key) as [CatalogListedCitation]?
            if let match = rows?.first(where: { $0.citation.id == id }) {
                return match.citation
            }
        }
        return nil
    }

    private func applyEntryFocus() {
        guard !didApplyEntry else { return }
        didApplyEntry = true
        switch entry {
        case .addProperty(_, let subjectID, _):
            beginPropertyDraft(subjectID: subjectID)
        case .connect(_, let fromID, let toID, _):
            beginBridgeDraft(endA: fromID, endB: toID, termID: "", origin: .graph)
        case .edit(_, let subjectID, _, _, let observationID):
            guard let observationID, !observationID.isEmpty else { break }
            beginEdit(observationID: observationID, subjectID: subjectID)
        }
    }

    private func applySelectedArtifact(_ id: String) async {
        selectedArtifactID = id
        fields.locator = .artifactOnly()
        await reloadArtifactViewer(preferredPage: nil)
    }

    private func presentAfterIdentityChange(keepReadingPage: Bool = false) async {
        if artifacts.isEmpty {
            selectedArtifactID = nil
            artifactViewer.unload()
        } else {
            let page = fields.locator.page ?? (keepReadingPage ? currentReadingPage : nil)
            await reloadArtifactViewer(preferredPage: page)
            await reloadListedCitations()
        }
        announceIdentityChange()
    }

    /// Page to leave the viewer on when a new citation has no page of its own.
    private var currentReadingPage: Int? {
        guard artifactViewer.supportsPages else { return nil }
        let page = artifactViewer.page
        return page >= 1 ? page : nil
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
        pendingLeave = PendingLeave()
        return true
    }

    func discardLeaveChanges() {
        let keepDraft = pendingLeave?.keepDraft == true
        let next = stepContinuation
        stepContinuation = nil
        if keepDraft {
            pendingLeave = nil
        } else {
            fields.revert()
            authoring = .rest
            editSession = nil
            show(nil)
        }
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
