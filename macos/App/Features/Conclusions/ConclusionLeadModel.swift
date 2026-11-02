import Foundation
import Observation

/// Chooses the lead of any kind's gallery among its pictures, or goes back to the earliest (S10-39,
/// S14-02): a Person's portrait, an Event's or Place's thumbnail. A display choice, not evidence. The
/// write returns effects; the session refreshes the page and the lists from them. The copy of the
/// choice comes from the page's `gallery.choose`, read at each call.
@MainActor
@Observable
final class ConclusionLeadModel {
    private let session: WorkspaceSession
    private let userID: String
    private let entityId: String

    var toast: VocabularyToast?
    private(set) var isSaving = false

    init(session: WorkspaceSession, userID: String, entityId: String) {
        self.session = session
        self.userID = userID
        self.entityId = entityId
    }

    /// Makes `picture` the lead. Returns whether the save worked.
    @discardableResult
    func choose(_ picture: ConclusionDetailPicture, copy: ConclusionDetailGallery.LeadCopy) async -> Bool {
        await write(observationID: picture.id, copy: copy, body: picture.chosenToastBody)
    }

    /// Clears the choice, so the earliest picture leads.
    @discardableResult
    func useEarliest(copy: ConclusionDetailGallery.LeadCopy) async -> Bool {
        await write(observationID: "", copy: copy, body: copy.toastEarliestBody)
    }

    @discardableResult
    private func write(observationID: String, copy: ConclusionDetailGallery.LeadCopy, body: String) async -> Bool {
        guard !isSaving else { return false }
        isSaving = true
        defer { isSaving = false }
        do {
            let effects = try await session.store.setPreferredLikeness(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                entityID: entityId,
                observationID: observationID
            )
            session.apply(effects)
            toast = VocabularyToast(title: copy.toastTitle, body: body, tone: .success)
            return true
        } catch {
            let message = L10n.Errors.message(for: error)
            toast = VocabularyToast(title: message, body: "", tone: .danger)
            return false
        }
    }
}
