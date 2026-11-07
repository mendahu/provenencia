import SwiftUI

/// What the Unpromote confirm shows, as neutral props: the Subject's name, the
/// message, the tone, the connection sentences that come off with it, and a
/// write or preflight error. The caller maps its domain item onto this
/// (`UnpromoteConfirmItem.content`).
struct PVUnpromoteConfirmContent: Identifiable, Equatable {
    /// The Subject's id.
    let id: String
    let name: String
    let message: String
    let tone: PVConfirmTone
    let connections: PVUnpromoteConnections
    let error: String?
}

/// The connections that come off with the promotion.
enum PVUnpromoteConnections: Equatable {
    /// One sentence per connection; none draws nothing.
    case listed([String])
    /// The connections could not be listed; a sentence stands in for the list.
    case unlisted
}

extension View {
    /// The Unpromote confirm, shared by the Evidence graph and the Person page:
    /// an `.irreversible` or `.danger` `PVConfirm` over the connection list and
    /// any error. Dismissal stays on the caller's binding.
    func pvUnpromoteConfirm(
        item: Binding<PVUnpromoteConfirmContent?>,
        isRunning: Bool,
        accessibilityIdentifierPrefix: String,
        onConfirm: @escaping () -> Void
    ) -> some View {
        pvConfirm(
            item: item,
            copy: { content in
                PVConfirmCopy(
                    title: L10n.Unpromote.confirmTitle(name: content.name),
                    message: content.message,
                    confirm: L10n.Unpromote.button,
                    cancel: L10n.Unpromote.keep
                )
            },
            tone: item.wrappedValue?.tone ?? .irreversible,
            isRunning: isRunning,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onConfirm: onConfirm,
            detail: { PVUnpromoteConfirmDetail(content: $0) }
        )
    }
}

/// The connections that come off with the promotion, or the sentence that says
/// they could not be listed, then a write error. Nothing when there is none.
private struct PVUnpromoteConfirmDetail: View {
    let content: PVUnpromoteConfirmContent

    private var sentences: [String] {
        if case .listed(let sentences) = content.connections { return sentences }
        return []
    }

    private var isUnlisted: Bool { content.connections == .unlisted }

    var body: some View {
        if isUnlisted || !sentences.isEmpty || content.error != nil {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                if isUnlisted {
                    Text(L10n.Unpromote.confirmBridgesUnlisted)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                        .frame(maxWidth: .infinity, alignment: .leading)
                } else if !sentences.isEmpty {
                    Text(L10n.Unpromote.confirmBridgesHeader)
                        .pvMicroCaps()
                        .foregroundStyle(PVColor.textMuted)
                    ScrollView {
                        LazyVStack(alignment: .leading, spacing: PVSpacing.space2) {
                            ForEach(Array(sentences.enumerated()), id: \.offset) { _, sentence in
                                Text(verbatim: sentence)
                                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                                    .foregroundStyle(PVColor.textPrimary)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                            }
                        }
                    }
                    .frame(maxHeight: PVDeleteImpactLayout.listMaxHeight)
                    Text(L10n.Unpromote.confirmBridgesNote)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                }
                if let error = content.error {
                    PVCallout(tone: .danger, message: error, announces: true)
                }
            }
        }
    }
}

#Preview("Unpromote confirm") {
    Color.clear
        .frame(width: 520, height: 320)
        .pvUnpromoteConfirm(
            item: .constant(PVUnpromoteConfirmContent(
                id: "s1", name: "J. Weber",
                message: "This promotion is removed. The cards stay on the graph, and the handle stays.",
                tone: .irreversible,
                connections: .listed(["J. Weber is the father of M. Weber."]),
                error: nil
            )),
            isRunning: false,
            accessibilityIdentifierPrefix: "preview.unpromote",
            onConfirm: {}
        )
}

#Preview("Unpromote confirm, 200 connections") {
    Color.clear
        .frame(width: 520, height: 520)
        .pvUnpromoteConfirm(
            item: .constant(PVUnpromoteConfirmContent(
                id: "s1", name: "J. Weber",
                message: "This promotion is removed. The cards stay on the graph, and the handle stays.",
                tone: .irreversible,
                connections: .listed((1...200).map { "J. Weber is the father of M. Weber \($0)." }),
                error: nil
            )),
            isRunning: false,
            accessibilityIdentifierPrefix: "preview.unpromote.long",
            onConfirm: {}
        )
}
