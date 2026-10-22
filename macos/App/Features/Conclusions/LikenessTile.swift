import SwiftUI

/// One likeness: its 96 pt crop over a one-line muted caption. The Person
/// page's gallery passes the button that opens the Citation as `overlay`
/// (drawn over the crop only); the Promote evidence sheet passes none.
struct LikenessTile<Overlay: View>: View {
    static var crop: CGFloat { 96 }

    let projectDir: String
    let relPath: String
    let caption: String
    let captionWidth: CGFloat
    private let overlay: Overlay

    init(
        projectDir: String,
        relPath: String,
        caption: String,
        captionWidth: CGFloat,
        @ViewBuilder overlay: () -> Overlay
    ) {
        self.projectDir = projectDir
        self.relPath = relPath
        self.caption = caption
        self.captionWidth = captionWidth
        self.overlay = overlay()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: relPath,
                size: Self.crop,
                fallbackMark: .subjectPerson
            )
            .accessibilityHidden(true)
            .overlay { overlay }
            Text(verbatim: caption)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
                .truncationMode(.tail)
                .frame(maxWidth: captionWidth, alignment: .leading)
                .help(Text(verbatim: caption))
                .accessibilityHidden(true)
        }
    }
}

extension LikenessTile where Overlay == EmptyView {
    init(projectDir: String, relPath: String, caption: String, captionWidth: CGFloat) {
        self.init(
            projectDir: projectDir,
            relPath: relPath,
            caption: caption,
            captionWidth: captionWidth,
            overlay: { EmptyView() }
        )
    }
}
