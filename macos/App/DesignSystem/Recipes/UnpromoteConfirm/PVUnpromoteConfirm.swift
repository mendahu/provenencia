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
    /// One sentence per connection that comes off with the promotion.
    let bridges: [String]
    /// The connections could not be listed; the sentence stands in for the list.
    let bridgesUnlisted: Bool
    let error: String?
    /// The error is announced to VoiceOver (a preflight failure appears with the sheet).
    let announcesError: Bool
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
                    title: L10n.EvidenceGraph.unpromoteConfirmTitle(name: content.name),
                    message: content.message,
                    confirm: L10n.EvidenceGraph.unpromote,
                    cancel: L10n.EvidenceGraph.unpromoteKeep
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

    var body: some View {
        if content.bridgesUnlisted || !content.bridges.isEmpty || content.error != nil {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                if content.bridgesUnlisted {
                    Text(L10n.EvidenceGraph.unpromoteConfirmBridgesUnlisted)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                        .frame(maxWidth: .infinity, alignment: .leading)
                } else if !content.bridges.isEmpty {
                    Text(L10n.EvidenceGraph.unpromoteConfirmBridgesHeader)
                        .pvMicroCaps()
                        .foregroundStyle(PVColor.textMuted)
                    VStack(alignment: .leading, spacing: PVSpacing.space2) {
                        ForEach(Array(content.bridges.enumerated()), id: \.offset) { _, sentence in
                            Text(verbatim: sentence)
                                .font(PVFont.body(size: PVTypeScale.bodySmall))
                                .foregroundStyle(PVColor.textPrimary)
                                .frame(maxWidth: .infinity, alignment: .leading)
                        }
                    }
                    Text(L10n.EvidenceGraph.unpromoteConfirmBridgesNote)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                }
                if let error = content.error {
                    PVCallout(tone: .danger, message: error, announces: content.announcesError)
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
                bridges: ["J. Weber is the father of M. Weber."],
                bridgesUnlisted: false, error: nil, announcesError: false
            )),
            isRunning: false,
            accessibilityIdentifierPrefix: "preview.unpromote",
            onConfirm: {}
        )
}
