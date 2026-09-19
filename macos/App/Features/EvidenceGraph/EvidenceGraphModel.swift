import Foundation
import Observation

/// Create / place / drag logic for the Evidence graph (S6-03).
@MainActor
@Observable
final class EvidenceGraphModel {
    struct CreateDraft: Equatable {
        var label: String = ""
        var description: String = ""
    }

    let sourceID: String
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    var armedKind: EvidencePrimaryKind?
    var isCreating = false
    var pendingGridX: Int64 = 0
    var pendingGridY: Int64 = 0
    var draft = CreateDraft()
    var isSaving = false
    var labelError: String?
    var createError: String?
    var hoverGridX: Int64?
    var hoverGridY: Int64?
    var draggingSubjectID: String?
    var dragTranslation: CGSize = .zero

    /// Seeded type ids keyed by `EvidencePrimaryKind.rawValue`.
    private(set) var typeIDByKind: [String: String] = [:]
    private(set) var typeLabelByKind: [String: String] = [:]

    private var graphKey: CatalogQueryKey {
        CatalogQueryKey.sourceGraph(project: session.projectKey, sourceId: sourceID)
    }

    init(
        sourceID: String,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.sourceID = sourceID
        self.session = session
        self.store = store
        self.userID = userID
    }

    func prepare() async {
        do {
            let types = try await store.listSubjectTypes(projectDir: session.projectKey.projectDir)
            var ids: [String: String] = [:]
            var labels: [String: String] = [:]
            for type in types {
                guard EvidencePrimaryKind(rawValue: type.key) != nil else { continue }
                ids[type.key] = type.id
                labels[type.key] = type.label
            }
            typeIDByKind = ids
            typeLabelByKind = labels
        } catch {
            typeIDByKind = [:]
            typeLabelByKind = [:]
        }
    }

    func toggleArm(_ kind: EvidencePrimaryKind) {
        if armedKind == kind {
            disarm()
        } else {
            armedKind = kind
            hoverGridX = nil
            hoverGridY = nil
            draggingSubjectID = nil
            dragTranslation = .zero
        }
    }

    func disarm() {
        armedKind = nil
        hoverGridX = nil
        hoverGridY = nil
    }

    func updateHover(contentPoint: CGPoint) {
        guard armedKind != nil, !isCreating else { return }
        let cell = GraphCanvasGridMapping.gridCell(contentPoint: contentPoint)
        hoverGridX = cell.gridX
        hoverGridY = cell.gridY
    }

    func clearHover() {
        hoverGridX = nil
        hoverGridY = nil
    }

    /// Opens the create dialog at the snapped cell (tool must be armed).
    func beginCreate(at contentPoint: CGPoint) {
        guard let kind = armedKind, !isCreating, !isSaving else { return }
        let cell = GraphCanvasGridMapping.gridCell(contentPoint: contentPoint)
        pendingGridX = cell.gridX
        pendingGridY = cell.gridY
        draft = CreateDraft(
            label: defaultLabel(for: kind),
            description: ""
        )
        labelError = nil
        createError = nil
        isCreating = true
        clearHover()
    }

    func cancelCreate() {
        guard !isSaving else { return }
        isCreating = false
        labelError = nil
        createError = nil
        // Board: cancel leaves the tool armed.
    }

    /// Creates the subject + position. On success disarms and returns the new id.
    @discardableResult
    func confirmCreate() async -> String? {
        guard let kind = armedKind, !isSaving else { return nil }
        let trimmed = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            labelError = String(localized: L10n.EvidenceGraph.labelRequired)
            return nil
        }
        guard let typeID = typeIDByKind[kind.rawValue] else {
            createError = String(localized: L10n.EvidenceGraph.typesUnavailable)
            return nil
        }

        isSaving = true
        defer { isSaving = false }
        labelError = nil
        createError = nil
        do {
            let created = try await store.createSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                sourceID: sourceID,
                subjectTypeID: typeID,
                label: trimmed,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            _ = try await store.setSubjectPosition(
                projectDir: session.projectKey.projectDir,
                subjectID: created.id,
                gridX: pendingGridX,
                gridY: pendingGridY
            )
            session.apply(.createdSubject(sourceId: sourceID))
            isCreating = false
            disarm()
            return created.id
        } catch {
            createError = L10n.Errors.message(for: error)
            return nil
        }
    }

    func beginDrag(subjectID: String) {
        guard armedKind == nil, !isCreating else { return }
        draggingSubjectID = subjectID
        dragTranslation = .zero
    }

    func updateDrag(translation: CGSize) {
        guard draggingSubjectID != nil else { return }
        dragTranslation = translation
    }

    /// Snaps and persists; returns the new cell if the move succeeded.
    @discardableResult
    func endDrag(
        subjectID: String,
        originGridX: Int64,
        originGridY: Int64
    ) async -> (gridX: Int64, gridY: Int64)? {
        defer {
            draggingSubjectID = nil
            dragTranslation = .zero
        }
        let origin = GraphCanvasGridMapping.contentPoint(gridX: originGridX, gridY: originGridY)
        let dropped = CGPoint(
            x: origin.x + dragTranslation.width,
            y: origin.y + dragTranslation.height
        )
        let cell = GraphCanvasGridMapping.gridCell(contentPoint: dropped)
        return await persistPosition(subjectID: subjectID, gridX: cell.gridX, gridY: cell.gridY)
    }

    /// Moves a selected subject one grid cell (arrow keys).
    @discardableResult
    func moveSubject(
        subjectID: String,
        fromGridX: Int64,
        fromGridY: Int64,
        deltaX: Int64,
        deltaY: Int64
    ) async -> (gridX: Int64, gridY: Int64)? {
        guard armedKind == nil, !isCreating else { return nil }
        let nextX = fromGridX + deltaX
        let nextY = fromGridY + deltaY
        return await persistPosition(subjectID: subjectID, gridX: nextX, gridY: nextY)
    }

    func ghostPlacedSubject() -> SourceGraphPlacedSubject? {
        guard let kind = armedKind,
              let x = hoverGridX,
              let y = hoverGridY,
              !isCreating
        else { return nil }
        let label = typeLabelByKind[kind.rawValue] ?? kind.rawValue
        return SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "ghost",
                ref: "",
                sourceID: sourceID,
                subjectTypeID: typeIDByKind[kind.rawValue] ?? "",
                label: defaultLabel(for: kind),
                description: ""
            ),
            kind: kind,
            typeLabel: label,
            gridX: x,
            gridY: y,
            isCited: false
        )
    }

    func createDialogTitle(for kind: EvidencePrimaryKind) -> LocalizedStringResource {
        switch kind {
        case .person: L10n.EvidenceGraph.addPersonTitle
        case .event: L10n.EvidenceGraph.addEventTitle
        case .place: L10n.EvidenceGraph.addPlaceTitle
        }
    }

    func toolAccessibilityLabel(for kind: EvidencePrimaryKind, armed: Bool) -> String {
        let name = String(localized: toolName(for: kind))
        let state = armed
            ? String(localized: L10n.EvidenceGraph.toolOn)
            : String(localized: L10n.EvidenceGraph.toolOff)
        return "\(name), \(String(localized: L10n.EvidenceGraph.toolRole)), \(state)"
    }

    func toolName(for kind: EvidencePrimaryKind) -> LocalizedStringResource {
        switch kind {
        case .person: L10n.EvidenceGraph.toolPerson
        case .event: L10n.EvidenceGraph.toolEvent
        case .place: L10n.EvidenceGraph.toolPlace
        }
    }

    func armedHint(for kind: EvidencePrimaryKind) -> LocalizedStringResource {
        switch kind {
        case .person: L10n.EvidenceGraph.armedHintPerson
        case .event: L10n.EvidenceGraph.armedHintEvent
        case .place: L10n.EvidenceGraph.armedHintPlace
        }
    }

    private func defaultLabel(for kind: EvidencePrimaryKind) -> String {
        switch kind {
        case .person: String(localized: L10n.EvidenceGraph.defaultLabelPerson)
        case .event: String(localized: L10n.EvidenceGraph.defaultLabelEvent)
        case .place: String(localized: L10n.EvidenceGraph.defaultLabelPlace)
        }
    }

    private func persistPosition(
        subjectID: String,
        gridX: Int64,
        gridY: Int64
    ) async -> (gridX: Int64, gridY: Int64)? {
        do {
            _ = try await store.setSubjectPosition(
                projectDir: session.projectKey.projectDir,
                subjectID: subjectID,
                gridX: gridX,
                gridY: gridY
            )
            if let handle: QueryHandle<SourceGraphSnapshot> = session.queryHandle(graphKey),
               let snapshot = handle.value
            {
                session.setQueryValue(
                    graphKey,
                    value: snapshot.updatingPosition(subjectID: subjectID, gridX: gridX, gridY: gridY)
                )
            }
            return (gridX, gridY)
        } catch {
            return nil
        }
    }
}
