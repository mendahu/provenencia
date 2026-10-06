import SwiftUI

/// Placeholder page for a Conclude destination that hasn't shipped.
/// Places stay here until S9-26. The Events list and an Event's page are live.
struct ConclusionStubView: View {
    let section: WorkspaceSection

    var body: some View {
        PVEmptyState(
            icon: section.icon,
            title: section.label,
            message: Self.message(for: section)
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
