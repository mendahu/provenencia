import Foundation
import Observation

struct ConnectionRow: Identifiable, Equatable {
    var id: UUID
    var isPending: Bool
    var fromSubjectID: String
    var toSubjectID: String
    var fromLabel: String
    var toLabel: String
    var bridgeTypeKey: String
    var bridgeID: String? = nil
    var bridgeRef: String? = nil
    var sentence: String
    var termProperty: CatalogProperty? = nil
    var rolePersistedID: String? = nil
    var rolePersistedRef: String? = nil
    var roleTermID: String
    var roleBaselineTermID: String? = nil
    var termTouched: Bool
    var isSaving: Bool
    var error: String? = nil

    var isLocation: Bool { termProperty == nil }

    var isTouched: Bool {
        if termProperty == nil { return false }
        if isPending { return termTouched }
        return roleTermID != (roleBaselineTermID ?? "")
    }

    var canSave: Bool {
        if isSaving || !isPending { return false }
        if termProperty == nil { return true }
        return !roleTermID.isEmpty
    }

    var canCommitRole: Bool {
        !isPending && termProperty != nil && !isSaving && isTouched && !roleTermID.isEmpty
    }

    var termFieldLabel: String {
        switch bridgeTypeKey {
        case "relationship":
            return String(localized: L10n.CitationComposer.connectionRelationship)
        case "participation":
            return String(localized: L10n.CitationComposer.connectionRole)
        default:
            return String(localized: L10n.CitationComposer.connectionNoRole)
        }
    }

    var accessibilityLabel: String {
        if isLocation {
            return L10n.CitationComposer.connectionAccessibilityLocation(status: statusSpoken)
        }
        let term = roleTermID.isEmpty
            ? String(localized: L10n.CitationComposer.connectionTermNotChosen)
            : termFieldLabel
        return L10n.CitationComposer.connectionAccessibility(
            sentence: sentence,
            termLabel: termFieldLabel,
            term: term,
            status: statusSpoken
        )
    }

    private var statusSpoken: String {
        if isPending {
            return String(localized: L10n.CitationComposer.rowStateNew)
        }
        if isTouched {
            return String(localized: L10n.CitationComposer.rowStateEdited)
        }
        if let bridgeRef, !bridgeRef.isEmpty {
            return L10n.CitationComposer.connectionSavedStatus(ref: bridgeRef)
        }
        return String(localized: L10n.CitationComposer.rowStateSaved)
    }
}

/// Pending Connect row plus saved connection rows grouped from edge Observations.
@MainActor
@Observable
final class CitationConnections {
    var rows: [ConnectionRow] = []

    private weak var context: CitationComposerContext?
    private var pendingSeed: ConnectionRow?

    func attach(_ context: CitationComposerContext) {
        self.context = context
    }

    func seedPending(
        fromSubjectID: String,
        toSubjectID: String,
        fromLabel: String,
        toLabel: String,
        bridgeTypeKey: String,
        termProperty: CatalogProperty?,
        sentence: String
    ) {
        let pending = ConnectionRow(
            id: UUID(),
            isPending: true,
            fromSubjectID: fromSubjectID,
            toSubjectID: toSubjectID,
            fromLabel: fromLabel,
            toLabel: toLabel,
            bridgeTypeKey: bridgeTypeKey,
            bridgeID: nil,
            bridgeRef: nil,
            sentence: sentence,
            termProperty: termProperty,
            rolePersistedID: nil,
            rolePersistedRef: nil,
            roleTermID: "",
            roleBaselineTermID: nil,
            termTouched: false,
            isSaving: false,
            error: nil
        )
        pendingSeed = pending
        if !rows.contains(where: \.isPending) {
            rows.insert(pending, at: 0)
        }
    }

    /// Rebuild saved rows from the Citation. Keeps the pending row if present.
    func replaceSaved(_ saved: [ConnectionRow]) {
        let pending = rows.first(where: \.isPending) ?? pendingSeed
        rows = saved
        if let pending {
            rows.insert(pending, at: 0)
        }
    }

    func clearSavedKeepingPending() {
        rows = rows.filter(\.isPending)
    }

    func applyTerm(connectionID: UUID, termID: String) {
        guard let index = rows.firstIndex(where: { $0.id == connectionID }) else { return }
        guard rows[index].termProperty != nil else { return }
        rows[index].roleTermID = termID
        rows[index].termTouched = true
        rows[index].error = nil
    }

    func saveConnection() async {
        guard let context, let index = rows.firstIndex(where: \.isPending) else { return }
        if rows[index].isSaving || !rows[index].canSave { return }
        rows[index].isSaving = true
        rows[index].error = nil
        await performSave(context: context)
        if let current = rows.firstIndex(where: \.isPending) {
            rows[current].isSaving = false
        }
    }

    func discard() {
        rows.removeAll(where: \.isPending)
        pendingSeed = nil
    }

    func commitRole(connectionID: UUID) async {
        guard let context, let index = rows.firstIndex(where: { $0.id == connectionID }) else { return }
        if rows[index].isSaving || !rows[index].canCommitRole { return }
        rows[index].isSaving = true
        rows[index].error = nil
        await performCommitRole(connectionID: connectionID, context: context)
        if let current = rows.firstIndex(where: { $0.id == connectionID }) {
            rows[current].isSaving = false
        }
    }

    func revertRole(connectionID: UUID) {
        guard let index = rows.firstIndex(where: { $0.id == connectionID }) else { return }
        rows[index].roleTermID = rows[index].roleBaselineTermID ?? ""
        rows[index].termTouched = false
        rows[index].error = nil
    }

    func markPendingSaved(_ saved: ConnectionRow) {
        if let index = rows.firstIndex(where: \.isPending) {
            rows[index] = saved
        } else {
            rows.insert(saved, at: 0)
        }
        pendingSeed = nil
    }

    func markError(connectionID: UUID, message: String) {
        guard let index = rows.firstIndex(where: { $0.id == connectionID }) else { return }
        rows[index].error = message
        rows[index].isSaving = false
    }

    func updateRole(connectionID: UUID, persistedID: String, persistedRef: String, termID: String) {
        guard let index = rows.firstIndex(where: { $0.id == connectionID }) else { return }
        rows[index].rolePersistedID = persistedID
        rows[index].rolePersistedRef = persistedRef
        rows[index].roleTermID = termID
        rows[index].roleBaselineTermID = termID
        rows[index].termTouched = false
        rows[index].error = nil
    }

    var hasTouchedWork: Bool {
        rows.contains { $0.isTouched && !$0.isSaving }
    }

    static func groupSaved(
        observations: [CatalogObservation],
        vocabulary: CitationComposerVocabulary,
        snapshot: SourceGraphSnapshot
    ) -> (connections: [ConnectionRow], leftover: [CatalogObservation]) {
        var leftover = observations
        var used = Set<String>()
        var connections: [ConnectionRow] = []
        let edges = leftover.filter { vocabulary.isEdge(subjectID: $0.subjectID, propertyID: $0.propertyID) }
        let byBridge = Dictionary(grouping: edges, by: \.subjectID)
        for (bridgeID, group) in byBridge {
            group.forEach { used.insert($0.id) }
            let bridge = snapshot.bridges.first { $0.id == bridgeID }
            let typeKey = vocabulary.graphSubjects.first { $0.id == bridgeID }?.typeKey
                ?? bridge?.kind.rawValue
                ?? ""
            let rule = vocabulary.connectRule(bridgeTypeKey: typeKey)
            let sentence = bridge.map { EvidenceBridgeEdgeSummary.sentence(for: $0, in: snapshot) }
                ?? vocabulary.graphSubjects.first { $0.id == bridgeID }?.label
                ?? ""
            var termProperty: CatalogProperty?
            if let rule, rule.disambiguation != "none", !rule.disambiguation.isEmpty {
                termProperty = vocabulary.property(key: rule.disambiguation)
            }
            var role: CatalogObservation?
            if let termProperty {
                let roles = leftover
                    .filter {
                        $0.subjectID == bridgeID && $0.propertyID == termProperty.id && !used.contains($0.id)
                    }
                    .sorted { $0.ref.localizedCaseInsensitiveCompare($1.ref) == .orderedAscending }
                if let first = roles.first {
                    role = first
                    used.insert(first.id)
                }
            }
            connections.append(
                ConnectionRow(
                    id: UUID(),
                    isPending: false,
                    fromSubjectID: "",
                    toSubjectID: "",
                    fromLabel: "",
                    toLabel: "",
                    bridgeTypeKey: typeKey,
                    bridgeID: bridgeID,
                    bridgeRef: vocabulary.graphSubjects.first { $0.id == bridgeID }?.ref
                        ?? bridge?.subject.ref,
                    sentence: sentence,
                    termProperty: termProperty,
                    rolePersistedID: role?.id,
                    rolePersistedRef: role?.ref,
                    roleTermID: role?.valueTermID ?? "",
                    roleBaselineTermID: role?.valueTermID,
                    termTouched: false,
                    isSaving: false,
                    error: nil
                )
            )
        }
        leftover.removeAll { used.contains($0.id) }
        return (connections, leftover)
    }

    func edgeDrafts(for pending: ConnectionRow, vocabulary: CitationComposerVocabulary) -> [CatalogObservationDraft]? {
        guard let rule = vocabulary.connectRule(bridgeTypeKey: pending.bridgeTypeKey) else { return nil }
        let fromType = vocabulary.graphSubjects.first { $0.id == pending.fromSubjectID }?.typeKey ?? ""
        let toType = vocabulary.graphSubjects.first { $0.id == pending.toSubjectID }?.typeKey ?? ""
        guard let bindings = ConnectEndpointBinding.bind(
            rule: rule,
            fromID: pending.fromSubjectID,
            fromTypeKey: fromType,
            fromLabel: pending.fromLabel,
            toID: pending.toSubjectID,
            toTypeKey: toType,
            toLabel: pending.toLabel
        ) else { return nil }
        var drafts: [CatalogObservationDraft] = []
        for binding in bindings {
            guard let property = vocabulary.property(key: binding.propertyKey) else { return nil }
            drafts.append(
                CatalogObservationDraft(
                    subjectID: "",
                    propertyID: property.id,
                    polarity: ObservationPolarity.positive.rawValue,
                    valueSubjectID: binding.subjectID
                )
            )
        }
        if let termProperty = pending.termProperty {
            drafts.append(
                CatalogObservationDraft(
                    subjectID: "",
                    propertyID: termProperty.id,
                    polarity: ObservationPolarity.positive.rawValue,
                    valueTermID: pending.roleTermID
                )
            )
        }
        return drafts
    }

    private func performSave(context: CitationComposerContext) async {
        guard let pending = rows.first(where: \.isPending),
              let drafts = edgeDrafts(for: pending, vocabulary: context.vocabulary)
        else { return }
        let values = context.citationValues()
        do {
            let attachID = context.citationID
            let created = try await context.store.createCitedBridge(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                fromSubjectID: pending.fromSubjectID,
                toSubjectID: pending.toSubjectID,
                bridgeTypeKey: pending.bridgeTypeKey,
                description: "",
                artifactID: attachID == nil ? (context.artifactID ?? "") : "",
                locatorJSON: attachID == nil ? values.locator.encodeJSON() : "",
                transcription: attachID == nil ? values.transcription : "",
                citationDescription: attachID == nil ? values.description : "",
                transcriptionUncertain: attachID == nil && values.transcriptionUncertain,
                transcriptionNote: attachID == nil ? values.transcriptionNote : "",
                citationNotes: [],
                observations: drafts,
                citationID: attachID
            )
            if context.citationID == nil {
                context.citationID = created.1.id
                context.captureCitationBaseline()
            }
            context.applySavedCitation()
            await context.waitForGraph()
            await context.reloadListedCitations()
            let role = created.2.first { obs in
                pending.termProperty.map { obs.propertyID == $0.id } ?? false
            }
            markPendingSaved(
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
                    sentence: context.vocabulary.graphSubjects.first { $0.id == created.0.id }?.label
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
        } catch {
            markError(connectionID: pending.id, message: L10n.Errors.message(for: error))
        }
    }

    private func performCommitRole(connectionID: UUID, context: CitationComposerContext) async {
        guard let row = rows.first(where: { $0.id == connectionID }),
              let citationID = context.citationID,
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
                let written = try await context.store.updateObservation(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    observation: observation
                )
                updateRole(
                    connectionID: connectionID,
                    persistedID: written.id,
                    persistedRef: written.ref,
                    termID: written.valueTermID
                )
            } else {
                let written = try await context.store.addObservationsToCitation(
                    projectDir: context.projectDir,
                    userID: context.userID,
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
                    updateRole(
                        connectionID: connectionID,
                        persistedID: written.id,
                        persistedRef: written.ref,
                        termID: written.valueTermID
                    )
                } else {
                    markError(
                        connectionID: connectionID,
                        message: String(localized: L10n.CitationComposer.rowStateError)
                    )
                }
            }
            context.applySavedCitation()
        } catch {
            markError(connectionID: connectionID, message: L10n.Errors.message(for: error))
        }
    }
}
