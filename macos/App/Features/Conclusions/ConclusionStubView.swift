import SwiftUI

/// Placeholder page for a Conclude destination whose list hasn't shipped:
/// Persons until S9-09, Events until S9-23, Places until S9-26. Reads nothing.
struct ConclusionStubView: View {
    let section: WorkspaceSection

    var body: some View {
        PVEmptyState(icon: section.icon, title: section.label, message: Self.message(for: section))
            .frame(maxWidth: 480)
            .padding(PVSpacing.space9)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    static func message(for section: WorkspaceSection) -> String {
        switch section {
        case .events: String(localized: L10n.Workspace.eventsStubMessage)
        case .places: String(localized: L10n.Workspace.placesStubMessage)
        default: String(localized: L10n.Workspace.personsStubMessage)
        }
    }
}

#Preview {
    ConclusionStubView(section: .events)
}
