import Foundation
import Observation

/// Chooses a Person's portrait among its gallery likenesses, or goes back to
/// the earliest (S10-39). A display choice, not evidence. The write returns
/// effects; the session refreshes the page and the Persons list from them.
@MainActor
@Observable
final class ConclusionPortraitModel {
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

    /// Makes `likeness` the portrait.
    func choose(_ likeness: ConclusionDetailLikeness) async {
        await write(
            observationID: likeness.id,
            body: L10n.Conclusions.toastPortraitChosenBody(likeness.sourceTitle)
        )
    }

    /// Clears the choice, so the earliest likeness leads.
    func useEarliest() async {
        await write(observationID: "", body: L10n.Conclusions.toastPortraitEarliestBody)
    }

    private func write(observationID: String, body: String) async {
        guard !isSaving else { return }
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
            toast = VocabularyToast(
                title: L10n.Conclusions.toastPortraitUpdatedTitle,
                body: body,
                tone: .success
            )
        } catch {
            toast = VocabularyToast(title: L10n.Errors.message(for: error), body: "", tone: .danger)
        }
    }
}
