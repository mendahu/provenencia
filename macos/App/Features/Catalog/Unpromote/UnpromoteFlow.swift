import Foundation
import Observation
import SwiftUI

/// What the Unpromote confirm names: the Subject, the connection sentences that
/// come off with it, and what the preflight says the write deletes. Built fresh
/// on open, so `error` starts nil.
struct UnpromoteConfirmItem: Identifiable, Equatable {
    /// The Subject's id.
    let id: String
    let name: String
    let kind: EvidencePrimaryKind
    /// The Subject's own handle: the one its promotion filed it under.
    let handleID: String?
    /// One sentence per connection the engine says comes off with it.
    let bridges: [String]
    /// The connections could not be listed (no preflight, no graph rows, or none matched).
    var bridgesUnlisted = false
    /// Handles the preflight reports this Unpromote deletes; nil when every handle stays.
    var deletedHandles: CatalogDeleteImpactGroup?
    /// The impact check failed; confirming may still delete the Subject's handle.
    var impactCheckFailed = false
    var error: String?

    /// The write deletes exactly the Subject's own handle: the last record leaves it.
    var deletesOwnHandle: Bool {
        guard let group = deletedHandles, group.total == 1, let handleID else { return false }
        return group.listed.count == 1 && group.listed[0].id == handleID
    }

    var confirmTone: PVConfirmTone {
        if deletedHandles != nil || impactCheckFailed || error != nil {
            return .danger
        }
        return .irreversible
    }

    var message: String {
        if deletesOwnHandle, let ref = deletedHandles?.listed.first?.ref {
            return L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind: kind, ref: ref)
        }
        if let group = deletedHandles,
           let line = PVDeleteImpactCopy.cascadeLines(
               CatalogDeleteImpact(allowed: true, gate: .ok, groups: [], cascades: [group])
           ).first {
            return line
        }
        if impactCheckFailed {
            return L10n.EvidenceGraph.unpromoteConfirmPreflightFailed(kind: kind)
        }
        return L10n.string(L10n.EvidenceGraph.unpromoteConfirmMessage)
    }

    var content: PVUnpromoteConfirmContent {
        PVUnpromoteConfirmContent(
            id: id, name: name, message: message, tone: confirmTone, bridges: bridges,
            bridgesUnlisted: bridgesUnlisted, error: error, announcesError: impactCheckFailed
        )
    }
}

/// The Unpromote open, confirm and write, held by each page that offers it (the
/// Evidence graph's model, the Person page's Sources section) and bound with
/// `.unpromoteConfirm(flow:…)`, as `DeleteImpactFlow` is for deletes.
@MainActor
@Observable
final class UnpromoteFlow {
    private(set) var pending: UnpromoteConfirmItem?
    private(set) var isRunning = false
    /// Opens in progress. A count, so a dropped stale open cannot clear another's busy state.
    private var opening = 0
    @ObservationIgnored
    private var generation = 0

    /// An open is in progress or the write runs.
    var isBusy: Bool { opening > 0 || isRunning }

    /// Reads the preflight, words the connections it names from the Subject's
    /// graph, and shows the confirm. A later open drops an earlier one that is
    /// still reading.
    func open(
        subjectID: String,
        name: String,
        kind: EvidencePrimaryKind,
        handleID: String?,
        graphID: String,
        session: WorkspaceSession
    ) async {
        generation += 1
        let mine = generation
        opening += 1
        defer { opening -= 1 }
        var deletedHandles: CatalogDeleteImpactGroup?
        var impactCheckFailed = false
        var preflightError: String?
        var bridgeIDs: [String] = []
        do {
            let impact = try await session.store.getUnpromoteImpact(
                projectDir: session.projectKey.projectDir,
                subjectID: subjectID
            )
            if let group = impact.deletes, group.total > 0 {
                deletedHandles = group
            }
            bridgeIDs = impact.unpromotedBridgeSubjectIDs
        } catch {
            impactCheckFailed = true
            preflightError = L10n.Errors.message(for: error)
        }
        var sentences: [String] = []
        if !bridgeIDs.isEmpty, let snapshot = await session.evidenceGraphSnapshot(graphID: graphID) {
            let byID = EvidenceBridgeEdgeSummary.bridges(of: snapshot)
            let subjects = EvidenceBridgeEdgeSummary.subjects(of: snapshot)
            sentences = bridgeIDs.compactMap { id in
                byID[id].map {
                    EvidenceBridgeEdgeSummary.sentence(
                        for: $0, in: snapshot, orders: session.nameFormat, subjects: subjects, bridges: byID
                    )
                }
            }
        }
        guard generation == mine else { return }
        var item = UnpromoteConfirmItem(
            id: subjectID, name: name, kind: kind, handleID: handleID, bridges: sentences,
            bridgesUnlisted: impactCheckFailed || (!bridgeIDs.isEmpty && sentences.isEmpty)
        )
        item.deletedHandles = deletedHandles
        item.impactCheckFailed = impactCheckFailed
        item.error = preflightError
        pending = item
    }

    /// Writes the Unpromote. `beforeApply` sees the effects before the session
    /// applies them, so a page can leave an entity the write deleted. Success
    /// closes the confirm and returns the effects; a failure stays in the
    /// confirm as its error and returns nil.
    func confirm(
        session: WorkspaceSession,
        userID: String,
        beforeApply: @MainActor (CatalogWriteEffects) -> Void = { _ in }
    ) async -> CatalogWriteEffects? {
        guard let item = pending, !isRunning else { return nil }
        isRunning = true
        defer { isRunning = false }
        do {
            let effects = try await session.store.unpromoteSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: item.id
            )
            beforeApply(effects)
            session.apply(effects)
            pending = nil
            return effects
        } catch {
            pending?.error = L10n.Errors.message(for: error)
            return nil
        }
    }

    /// Closes the confirm without writing. Does nothing while the write runs.
    func cancel() {
        guard !isRunning else { return }
        pending = nil
    }
}

extension View {
    /// Binds the shared Unpromote confirm to `flow`.
    func unpromoteConfirm(
        flow: UnpromoteFlow,
        accessibilityIdentifierPrefix: String,
        onConfirm: @escaping () -> Void
    ) -> some View {
        pvUnpromoteConfirm(
            item: Binding(
                get: { flow.pending?.content },
                set: { if $0 == nil { flow.cancel() } }
            ),
            isRunning: flow.isRunning,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onConfirm: onConfirm
        )
    }
}
