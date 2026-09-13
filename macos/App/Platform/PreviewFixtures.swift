#if DEBUG
import Foundation

/// Shared `#Preview` fixture data. `OnboardingView`'s previews and
/// `WorkspaceView.previewModel()` each used to build their own,
/// near-identical `InstallIdentity`/`ProjectInfo` literals; centralizing
/// them here means every preview reflects the same install/project
/// unless a specific preview deliberately overrides one. Matches
/// `FakeStore.lastResult`'s own built-in default persona (Jake Robins /
/// Robins Family), rather than introducing a third one.
enum PreviewFixture {
    static let identity = InstallIdentity(
        userID: "00000000-0000-7000-8000-000000000001",
        displayName: "Jake Robins",
        ref: "USR-F4N2P"
    )

    static let project = ProjectInfo(
        label: "Robins Family",
        folderName: "robins-family.provenencia",
        createdAt: "",
        updatedAt: "",
        updatedByUserID: "",
        updatedByDisplayName: "",
        updatedByRef: "",
        uuid: "00000000-0000-7000-8000-0000000000aa"
    )
}
#endif
