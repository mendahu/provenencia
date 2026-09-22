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
        var dateDraft: DateValueDraft
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
    private let userID: String
    private let store: any GenealogyStore
    let session: WorkspaceSession

    private(set) var phase: Phase = .loading
    private(set) var subjectLabel: String = ""
    private(set) var subjectTypeKey: String = ""
    private(set) var subjectTypeID: String = ""
    private(set) var sourceTitle: String = ""
    private(set) var artifacts: [CatalogArtifact] = []
    private(set) var selectedArtifactID: String?
    /// Frame 7 selection before Continue.
    var pendingArtifactID: String?
    private(set) var availableProperties: [CatalogProperty] = []
    private(set) var termsByPropertyID: [String: [CatalogPropertyTerm]] = [:]
    private(set) var isSubmitting = false
    private(set) var didSubmit = false

    /// Viewer page index (1-based). Stub count until S7-06.
    private(set) var viewerPage: Int = 1
    private(set) var viewerPageCount: Int = 1
    /// Committed locator page; nil until the researcher sets one (Frame 12).
    private(set) var locatorPage: Int?
    /// Thin image path used Draw region to commit a whole-image page selector.
    private(set) var locatorIsWholeImage = false

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
        availableProperties.first { $0.id == id }
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
            let listedObservations = try await observationsLoad
            let snapshot = SourceGraphSnapshot.build(
                sourceId: sourceID,
                subjects: subjects,
                positions: positions,
                types: types,
                observations: listedObservations
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
            sourceTitle = workspace.source.title
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
                selectedArtifactID = nil
                pendingArtifactID = nil
                phase = .compose
                return
            }
            if artifacts.count == 1 {
                applySelectedArtifact(artifacts[0].id)
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
        applySelectedArtifact(id)
        phase = .compose
        formError = nil
        locatorError = nil
        submitAttempted = false
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
        guard isPDFArtifact, viewerPage > 1 else { return }
        viewerPage -= 1
        commitLocatorPage(viewerPage)
    }

    func goToNextPage() {
        guard isPDFArtifact, viewerPage < viewerPageCount else { return }
        viewerPage += 1
        commitLocatorPage(viewerPage)
    }

    /// Thin stand-in for Draw region (no polygon UI until S7-07).
    func markWholeImageLocator() {
        if isPDFArtifact {
            locatorIsWholeImage = false
            commitLocatorPage(viewerPage)
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
            dateDraft: draft.dateDraft
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
        observations.removeAll { $0.id == id }
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

    private func applySelectedArtifact(_ id: String) {
        selectedArtifactID = id
        pendingArtifactID = id
        viewerPage = 1
        viewerPageCount = 1
        clearLocator()
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
