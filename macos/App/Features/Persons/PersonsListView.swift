import SwiftUI

/// Persons destination (Conclude). Stub until S9-09 builds the designed list
/// (S9-D2) on the `.personsList` query key; until then it reads nothing.
struct PersonsListView: View {
    var body: some View {
        PVEmptyState(
            icon: .person,
            title: L10n.Workspace.personsTitle,
            message: String(localized: L10n.Workspace.personsStubMessage)
        )
        .frame(maxWidth: 480)
        .padding(PVSpacing.space9)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

#Preview {
    PersonsListView()
}
