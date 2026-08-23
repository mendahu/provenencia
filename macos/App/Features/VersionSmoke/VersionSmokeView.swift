import SwiftUI

// TEMPORARY (Spike 1 PR3): placeholder window that prints the Go core version and SQLite probe.
// Not a product screen. Remove with VersionSmokeModel when onboarding (PR7) lands.
struct VersionSmokeView: View {
    @State private var model: VersionSmokeModel

    init(store: any GenealogyStore = GoStore()) {
        _model = State(initialValue: VersionSmokeModel(store: store))
    }

    var body: some View {
        VStack(spacing: 12) {
            Text("TEMPORARY: FFI probe")
                .font(.caption)
                .foregroundStyle(.secondary)
                .accessibilityIdentifier("versionSmoke.banner")
            Text(model.versionText.isEmpty ? "…" : model.versionText)
                .font(.title2)
                .accessibilityIdentifier("versionSmoke.version")
            Text(model.sqliteText.isEmpty ? "…" : model.sqliteText)
                .font(.body)
                .accessibilityIdentifier("versionSmoke.sqlite")
            if let errorText = model.errorText {
                Text(errorText)
                    .foregroundStyle(.red)
                    .accessibilityIdentifier("versionSmoke.error")
            }
        }
        .frame(minWidth: 640, minHeight: 420)
        .background(Color(nsColor: .windowBackgroundColor))
        .task {
            await model.loadVersion()
        }
    }
}

#Preview("TEMPORARY VersionSmoke") {
    VersionSmokeView(store: FakeStore())
}
