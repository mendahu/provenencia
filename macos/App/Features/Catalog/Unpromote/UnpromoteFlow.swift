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
    /// The connections the engine says come off with it, or that they could not be listed
    /// (no preflight, no graph rows, or none matched).
    let connections: PVUnpromoteConnections
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

    /// The ref the preflight lists for the Subject's own handle, nil when it does not list it.
    var ownHandleDeletedRef: String? {
        deletedHandles?.listed.first { $0.id == handleID }?.ref
    }

    var message: String {
        if deletesOwnHandle, let ref = ownHandleDeletedRef {
            return L10n.Unpromote.confirmMessageDeletes(kind: kind, ref: ref)
        }
        if let group = deletedHandles,
           let line = PVDeleteImpactCopy.cascadeLines(
               CatalogDeleteImpact(allowed: true, gate: .ok, groups: [], cascades: [group])
           ).first {
            return line
        }
        if impactCheckFailed {
            return L10n.Unpromote.confirmPreflightFailed(kind: kind)
        }
        return L10n.string(L10n.Unpromote.confirmMessage)
    }

    var content: PVUnpromoteConfirmContent {
        PVUnpromoteConfirmContent(
            id: id, name: name, message: message, tone: confirmTone, connections: connections, error: error
        )
    }
}

/// What a page that shows the Subject's own handle does once the Unpromote has written: leave
/// when the write deleted that handle, else stay. Each carries what VoiceOver says.
enum UnpromotePageOutcome: Equatable {
    /// The write deleted the page's handle; the text names the deleted ref, nil when the
    /// preflight did not list it.
    case leave(String?)
    /// The page's handle stays; the text names the record that left.
    case stay(String)

    static func after(_ effects: CatalogWriteEffects, item: UnpromoteConfirmItem) -> UnpromotePageOutcome {
        if let handleID = item.handleID, effects.deletedHandles.contains(handleID) {
            return .leave(item.ownHandleDeletedRef.map { L10n.Unpromote.deletedAnnouncement(ref: $0) })
        }
        return .stay(L10n.Unpromote.unpromotedAnnouncement(name: item.name))
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
            id: subjectID, name: name, kind: kind, handleID: handleID,
            connections: impactCheckFailed || (!bridgeIDs.isEmpty && sentences.isEmpty) ? .unlisted : .listed(sentences)
        )
        item.deletedHandles = deletedHandles
        item.impactCheckFailed = impactCheckFailed
        item.error = preflightError
        pending = item
    }

    /// Writes the Unpromote for a page that shows no handle of its own (the Evidence graph) and
    /// applies its effects. Success closes the confirm and returns true; a failure stays in the
    /// confirm as its error and returns false.
    @discardableResult
    func confirm(session: WorkspaceSession, userID: String) async -> Bool {
        await write(session: session, userID: userID) { session.apply($0) }
    }

    /// Writes the Unpromote for a page that shows the Subject's own handle, and applies its
    /// effects as `UnpromotePageOutcome` says: `leavingPage` when the write deleted that handle,
    /// else a plain apply; `announce` then speaks the outcome's text. A failed write calls
    /// neither; the confirm keeps the error.
    @discardableResult
    func confirm(
        session: WorkspaceSession,
        userID: String,
        leavingPage leave: @escaping @MainActor () -> Void,
        announce: @escaping @MainActor (String) -> Void = { AccessibilityAnnouncement.post($0) }
    ) async -> Bool {
        guard let item = pending else { return false }
        return await write(session: session, userID: userID) { effects in
            switch UnpromotePageOutcome.after(effects, item: item) {
            case .leave(let text):
                session.apply(effects, leavingPage: leave)
                if let text { announce(text) }
            case .stay(let text):
                session.apply(effects)
                announce(text)
            }
        }
    }

    private func write(
        session: WorkspaceSession,
        userID: String,
        apply: @MainActor (CatalogWriteEffects) -> Void
    ) async -> Bool {
        guard let item = pending, !isRunning else { return false }
        isRunning = true
        defer { isRunning = false }
        pending?.error = nil
        do {
            let effects = try await session.store.unpromoteSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: item.id
            )
            apply(effects)
            pending = nil
            return true
        } catch {
            pending?.error = L10n.Errors.message(for: error)
            return false
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
