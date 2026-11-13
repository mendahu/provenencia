import SwiftUI

/// The measures of a `PVTileCard`: one place for every size. The shell applies
/// `height`, `thumbnail`, `padding` and `contentSpacing`; the title and lines slots apply
/// `titleLineLimit` and `linesLineLimit`.
enum PVTileCardSize: Equatable {
    case regular, compact

    nonisolated var height: CGFloat {
        switch self {
        case .regular: 92
        case .compact: 46
        }
    }

    /// The side of the square thumbnail slot, which is also the height the text beside it
    /// gets. A compact card's thumbnail fills its height inside the padding.
    nonisolated var thumbnail: CGFloat {
        switch self {
        case .regular: 70
        case .compact: 38
        }
    }

    /// The space between the title and the lines under it.
    nonisolated var lineSpacing: CGFloat { 3 }

    nonisolated var titleLineLimit: Int {
        switch self {
        case .regular: 2
        case .compact: 1
        }
    }

    nonisolated var linesLineLimit: Int {
        switch self {
        case .regular: 2
        case .compact: 1
        }
    }

    /// The narrowest a layout should make the card.
    nonisolated var minimumWidth: CGFloat {
        switch self {
        case .regular: 120
        case .compact: 300
        }
    }

    nonisolated var padding: EdgeInsets {
        switch self {
        case .regular: EdgeInsets(top: 10, leading: 10, bottom: 10, trailing: 10)
        case .compact: EdgeInsets(top: 4, leading: 4, bottom: 4, trailing: 8)
        }
    }

    /// The space between the thumbnail and the text.
    nonisolated var contentSpacing: CGFloat {
        switch self {
        case .regular: 10
        case .compact: 8
        }
    }
}

/// A card with a thumbnail, a title, lines under it and an optional trailing
/// view beside the title. Content-agnostic: slots and tokens only. With an
/// `action` it is one button (focus ring, hover lift, Return); without one it
/// is a static element that VoiceOver reads as its label, so a static card's
/// slots hold no control. `isRaised` draws the raised, elevated surface.
struct PVTileCard<Thumbnail: View, Title: View, Lines: View, Trailing: View>: View {
    let size: PVTileCardSize
    let isRaised: Bool
    let accessibilityLabel: String
    let help: String
    let action: (() -> Void)?
    let thumbnail: Thumbnail
    let title: Title
    let lines: Lines
    let trailing: Trailing
    private let hasTrailing: Bool

    init(
        size: PVTileCardSize,
        isRaised: Bool = false,
        accessibilityLabel: String,
        help: String,
        action: (() -> Void)? = nil,
        @ViewBuilder thumbnail: () -> Thumbnail,
        @ViewBuilder title: () -> Title,
        @ViewBuilder lines: () -> Lines,
        @ViewBuilder trailing: () -> Trailing
    ) {
        self.size = size
        self.isRaised = isRaised
        self.accessibilityLabel = accessibilityLabel
        self.help = help
        self.action = action
        self.thumbnail = thumbnail()
        self.title = title()
        self.lines = lines()
        self.trailing = trailing()
        hasTrailing = !(Trailing.self is EmptyView.Type)
    }

    var body: some View {
        if let action {
            Button(action: action) { card }
                .buttonStyle(PVFocusRingButtonStyle())
                .pvActivatesOnReturn(action)
                .pvLiftHover(scale: 1.025)
                .accessibilityLabel(accessibilityLabel)
                .help(help)
        } else {
            card
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(accessibilityLabel)
                .help(help)
        }
    }

    private var card: some View {
        PVCard(
            tone: isRaised ? .raised : .card,
            cornerRadius: PVRadius.sm,
            elevated: isRaised,
            insets: size.padding
        ) {
            HStack(alignment: .center, spacing: size.contentSpacing) {
                thumbnail
                VStack(alignment: .leading, spacing: size.lineSpacing) {
                    if hasTrailing {
                        HStack(spacing: PVSpacing.space4) {
                            title
                            trailing.layoutPriority(1)
                        }
                    } else {
                        title
                    }
                    lines
                }
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .frame(maxWidth: .infinity, minHeight: size.thumbnail, maxHeight: size.thumbnail, alignment: .leading)
        }
        .frame(height: size.height)
        .clipped()
    }
}

extension PVTileCard where Trailing == EmptyView {
    init(
        size: PVTileCardSize,
        isRaised: Bool = false,
        accessibilityLabel: String,
        help: String,
        action: (() -> Void)? = nil,
        @ViewBuilder thumbnail: () -> Thumbnail,
        @ViewBuilder title: () -> Title,
        @ViewBuilder lines: () -> Lines
    ) {
        self.init(
            size: size,
            isRaised: isRaised,
            accessibilityLabel: accessibilityLabel,
            help: help,
            action: action,
            thumbnail: thumbnail,
            title: title,
            lines: lines,
            trailing: { EmptyView() }
        )
    }
}

#if DEBUG
private struct PVTileCardPreviewThumbnail: View {
    let side: CGFloat

    var body: some View {
        RoundedRectangle(cornerRadius: PVRadius.sm)
            .fill(PVColor.surfaceSunken)
            .frame(width: side, height: side)
    }
}

@MainActor private func previewCard(_ size: PVTileCardSize, raised: Bool = false, action: (() -> Void)? = nil) -> some View {
    PVTileCard(
        size: size,
        isRaised: raised,
        accessibilityLabel: "Anna Weber",
        help: "Anna Weber",
        action: action,
        thumbnail: { PVTileCardPreviewThumbnail(side: size.thumbnail) },
        title: {
            Text(verbatim: "Anna Weber")
                .font(PVListText.titleFont)
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(size.titleLineLimit)
        },
        lines: {
            Text(verbatim: "b. 3 May 1790 · Bremen")
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(size.linesLineLimit)
        },
        trailing: { PVBadge("Badge") }
    )
    .frame(width: max(size.minimumWidth, 260))
}

#Preview("Regular: static, raised, actionable") {
    VStack(spacing: 12) {
        previewCard(.regular)
        previewCard(.regular, raised: true)
        previewCard(.regular, action: {})
    }
    .padding()
}

#Preview("Compact: static, raised, actionable") {
    VStack(spacing: 12) {
        previewCard(.compact)
        previewCard(.compact, raised: true)
        previewCard(.compact, action: {})
    }
    .padding()
}
#endif
