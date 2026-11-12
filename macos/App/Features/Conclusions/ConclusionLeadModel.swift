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
    /// The likeness viewer is open and draws the Toast itself, so the page does not draw it a second
    /// time (S13-15 D-10). Only ``viewerOpened()`` and ``viewerClosed()`` write it.
    private(set) var toastInViewer = false
    /// The picture whose save is running ("" for going back to the earliest), nil when none is.
    private(set) var savingID: String?
    var isSaving: Bool { savingID != nil }

    init(session: WorkspaceSession, userID: String, entityId: String) {
        self.session = session
        self.userID = userID
        self.entityId = entityId
    }

    /// The viewer is on screen: it draws the Toast, and any the page was showing goes.
    func viewerOpened() {
        toastInViewer = true
        toast = nil
    }

    /// The viewer is gone, however it went: the page draws the Toast again, and any the viewer was showing goes.
    /// With no viewer open (the Gallery section left, or an Event or Place page) it does nothing, so a
    /// Toast the page is showing stays.
    func viewerClosed() {
        guard toastInViewer else { return }
        toastInViewer = false
        toast = nil
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
        savingID = observationID
        defer { savingID = nil }
        do {
            let effects = try await session.store.setPreferredLikeness(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                entityID: entityId,
                observationID: observationID
            )
            session.apply(effects)
            // The clicked button spins until the page shows the new lead, in place of the page's
            // Refreshing beside the section picker.
            let _: CatalogConclusionDetail? = await session.readyValue(
                .conclusionDetail(project: session.projectKey, entityId: entityId)
            )
            toast = VocabularyToast(title: copy.toastTitle, body: body, tone: .success)
            return true
        } catch {
            let message = L10n.Errors.message(for: error)
            toast = VocabularyToast(title: message, body: "", tone: .danger)
            return false
        }
    }
}
