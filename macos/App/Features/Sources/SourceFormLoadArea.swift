import SwiftUI

/// Which body a Source-page form's loaded area shows.
enum SourceFormLoadState: Equatable {
    case loading
    case failed(String)
    case ready
}

/// The area of a Source-page form that waits on a read (the Copy form's parts,
/// the graph form's kept links, an Artifact card's graph links): skeleton bars
/// with a caption while loading (three, or one half-width bar when `compact`,
/// for a card), a compact danger callout with **Read again** when the read
/// failed, and `ready` once it has a value. One body for every form.
struct SourceFormLoadArea<Ready: View>: View {
    let state: SourceFormLoadState
    let loadingText: LocalizedStringResource
    let errorTitle: LocalizedStringResource
    let errorMessage: (String) -> String
    let onReadAgain: () -> Void
    /// One half-width skeleton bar instead of three, for a small host.
    var compact = false
    /// Roots the ids: `.loading`, `.loadError`, `.readAgain`.
    let accessibilityIdentifierPrefix: String
    @ViewBuilder let ready: () -> Ready

    var body: some View {
        switch state {
        case .ready:
            ready()
        case .loading:
            loading
        case .failed(let message):
            PVCallout(
                tone: .danger,
                title: errorTitle,
                message: errorMessage(message),
                compact: true
            ) {
                PVButton(L10n.Sources.formReadAgain, variant: .secondary, size: .sm, action: onReadAgain)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).readAgain")
            }
            .accessibilityIdentifier("\(accessibilityIdentifierPrefix).loadError")
        }
    }

    private var loading: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            GeometryReader { geometry in
                VStack(alignment: .leading, spacing: PVSpacing.space5) {
                    ForEach(compact ? [0.48] : [0.42, 0.36, 0.48], id: \.self) { fraction in
                        PVSkeletonBar(width: geometry.size.width * fraction, height: 12)
                    }
                }
            }
            .frame(height: compact ? 12 : 12 * 3 + PVSpacing.space5 * 2)
            Text(loadingText)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(Text(loadingText))
        .accessibilityIdentifier("\(accessibilityIdentifierPrefix).loading")
    }
}
