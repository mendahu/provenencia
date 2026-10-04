import SwiftUI

/// Placeholder page for a Conclude destination that hasn't shipped: the
/// Events list until S9-23, Places until S9-26, and Person detail until
/// S9-16. Reads nothing.
struct ConclusionStubView: View {
    let section: WorkspaceSection
    /// Set for a detail page stub (one handle), nil for a list stub.
    var detailRef: String?

    var body: some View {
        PVEmptyState(
            icon: section.icon,
            title: section.label,
            message: detailRef.map { L10n.Workspace.personDetailStubMessage(ref: $0) } ?? Self.message(for: section)
        )
            .frame(maxWidth: 480)
            .padding(PVSpacing.space9)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    static func message(for section: WorkspaceSection) -> String {
        switch section {
        case .events: L10n.string(L10n.Workspace.eventsStubMessage)
        case .places: L10n.string(L10n.Workspace.placesStubMessage)
        default: L10n.string(L10n.Workspace.personsStubMessage)
        }
    }
}

#Preview {
    ConclusionStubView(section: .events)
}
