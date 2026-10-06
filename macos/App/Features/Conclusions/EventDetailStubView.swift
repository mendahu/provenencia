import SwiftUI

/// Placeholder for one Event's page until S9-24 draws it. The list already
/// opened this place; the body does not read the detail.
struct EventDetailStubView: View {
    var body: some View {
        PVEmptyState(
            icon: .calendar,
            title: L10n.Workspace.eventDetailStubTitle,
            message: L10n.string(L10n.Workspace.eventsStubMessage)
        )
        .frame(maxWidth: 480)
        .padding(PVSpacing.space9)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .accessibilityIdentifier("events.detail.stub")
    }
}
