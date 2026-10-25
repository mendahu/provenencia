import Foundation
import Observation

enum ConnectionTermCopy {
    static func label(disambiguationKey: String) -> String {
        switch disambiguationKey {
        case SeededPropertyKey.relationshipType:
            return L10n.string(L10n.CitationComposer.connectionRelationship)
        case SeededPropertyKey.role:
            return L10n.string(L10n.CitationComposer.connectionRole)
        case SeededPropertyKey.placeRelationshipType:
            return L10n.string(L10n.CitationComposer.connectionPlaceRelationship)
        default:
            return L10n.string(L10n.CitationComposer.connectionNoRole)
        }
    }
}

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
        ConnectionTermCopy.label(disambiguationKey: termProperty?.key ?? "")
    }

    var accessibilityLabel: String {
        if isLocation {
            return L10n.CitationComposer.connectionAccessibilityLocation(status: statusSpoken)
        }
        let term = roleTermID.isEmpty
            ? L10n.string(L10n.CitationComposer.connectionTermNotChosen)
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
            return L10n.string(L10n.CitationComposer.rowStateNew)
        }
        if isTouched {
            return L10n.string(L10n.CitationComposer.rowStateEdited)
        }
        if let bridgeRef, !bridgeRef.isEmpty {
            return L10n.CitationComposer.connectionSavedStatus(ref: bridgeRef)
        }
        return L10n.string(L10n.CitationComposer.rowStateSaved)
    }
}

/// Pending Connect row plus saved connection rows grouped from edge Observations.
@MainActor
@Observable
final class CitationConnections {
    var rows: [ConnectionRow] = []

    private weak var context: CitationComposerContext?

    func attach(_ context: CitationComposerContext) {
        self.context = context
    }

    /// Rebuild saved rows from the Citation. Keeps the pending row if present.
    func replaceSaved(_ saved: [ConnectionRow]) {
        let pending = rows.first(where: \.isPending)
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

    func saveConnection(citation: CitationFieldsDraft.Values) async {
        guard let context, let index = rows.firstIndex(where: \.isPending) else { return }
        if rows[index].isSaving || !rows[index].canSave { return }
        rows[index].isSaving = true
        rows[index].error = nil
        await performSave(context: context, citation: citation)
        if let current = rows.firstIndex(where: \.isPending) {
            rows[current].isSaving = false
        }
    }

    func discard() {
        rows.removeAll(where: \.isPending)
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
        snapshot: EvidenceGraphSnapshot
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
            var fromSubjectID = bridge.flatMap(\.endpointAID) ?? ""
            var toSubjectID = bridge.flatMap(\.endpointBID) ?? ""
            if let rule, fromSubjectID.isEmpty || toSubjectID.isEmpty {
                for edge in rule.edges {
                    guard let observation = group.first(where: {
                        vocabulary.property(id: $0.propertyID)?.key == edge.propertyKey
                    }), !observation.valueSubjectID.isEmpty else { continue }
                    if edge.endpointTypeKey == rule.fromTypeKey, fromSubjectID.isEmpty {
                        fromSubjectID = observation.valueSubjectID
                    } else if edge.endpointTypeKey == rule.toTypeKey, toSubjectID.isEmpty {
                        toSubjectID = observation.valueSubjectID
                    }
                }
            }
            connections.append(
                ConnectionRow(
                    id: UUID(),
                    isPending: false,
                    fromSubjectID: fromSubjectID,
                    toSubjectID: toSubjectID,
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

    private func performSave(context: CitationComposerContext, citation: CitationFieldsDraft.Values) async {
        guard let pending = rows.first(where: \.isPending),
              let drafts = edgeDrafts(for: pending, vocabulary: context.vocabulary)
        else { return }
        do {
            let citationID = try await context.ensureCitationID(citation)
            let created = try await context.store.createCitedBridge(
                projectDir: context.projectDir,
                userID: context.userID,
                fromSubjectID: pending.fromSubjectID,
                toSubjectID: pending.toSubjectID,
                bridgeTypeKey: pending.bridgeTypeKey,
                description: "",
                artifactID: "",
                locatorJSON: "",
                transcription: "",
                citationDescription: "",
                transcriptionUncertain: false,
                transcriptionNote: "",
                citationNotes: [],
                observations: drafts,
                citationID: citationID
            )
            context.applyWrite(created.3)
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
        } catch is CitationComposerNeedArtifact {
            markError(
                connectionID: pending.id,
                message: L10n.string(L10n.CitationComposer.needArtifact)
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
            let effects: CatalogWriteEffects
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
                    persistedID: written.0.id,
                    persistedRef: written.0.ref,
                    termID: written.0.valueTermID
                )
                effects = written.1
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
                if let saved = written.0.first {
                    updateRole(
                        connectionID: connectionID,
                        persistedID: saved.id,
                        persistedRef: saved.ref,
                        termID: saved.valueTermID
                    )
                } else {
                    markError(
                        connectionID: connectionID,
                        message: L10n.string(L10n.CitationComposer.rowStateError)
                    )
                }
                effects = written.1
            }
            context.applyWrite(effects)
        } catch {
            markError(connectionID: connectionID, message: L10n.Errors.message(for: error))
        }
    }
}
