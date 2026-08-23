import SwiftUI

struct ContentView: View {
    @State private var model: RootModel

    init(store: any GenealogyStore = GoStore()) {
        _model = State(initialValue: RootModel(store: store))
    }

    var body: some View {
        VStack(spacing: 12) {
            Text(model.versionText.isEmpty ? "…" : model.versionText)
                .font(.title2)
                .accessibilityIdentifier("root.version")
            if let errorText = model.errorText {
                Text(errorText)
                    .foregroundStyle(.red)
                    .accessibilityIdentifier("root.error")
            }
        }
        .frame(minWidth: 640, minHeight: 420)
        .background(Color(nsColor: .windowBackgroundColor))
        .task {
            await model.loadVersion()
        }
    }
}

#Preview {
    ContentView(store: FakeStore())
}
