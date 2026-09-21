import SwiftUI

/// Citation composer place stub (S7-09). Full form/viewer ships in S7-08 / S7-D4.
struct CitationComposerStubView: View {
    let sourceID: String
    let subjectID: String
    let session: WorkspaceSession
    let store: any GenealogyStore

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            Text(L10n.EvidenceGraph.composerStubTitle)
                .font(PVFont.display(size: PVTypeScale.h1, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
            Text(L10n.EvidenceGraph.composerStubMessage)
                .font(PVFont.body(size: PVTypeScale.body))
                .foregroundStyle(PVColor.textSecondary)
                .fixedSize(horizontal: false, vertical: true)
            Text(verbatim: subjectID)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textFaint)
            Spacer(minLength: 0)
        }
        .padding(PVSpacing.space9)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(PVColor.surfacePage)
        .accessibilityIdentifier("workspace.destination.citationComposer")
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.EvidenceGraph.composerStubTitle))
        // Keep unused params referenced for the eventual S7-08 host signature.
        .onAppear {
            _ = sourceID
            _ = session
            _ = store
        }
    }
}
