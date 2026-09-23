import SwiftUI

/// Connect disambiguation body (role / relationship_type) for `.pvFormDialog`.
struct EvidenceConnectDisambiguationForm: View {
    @Bindable var model: EvidenceGraphModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let line = model.disambiguationSubtitle() {
                Text(verbatim: line)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textSecondary)
            }
            PVComboBox(
                selection: termBinding,
                options: model.disambiguationTermOptions,
                size: .sm,
                placeholder: L10n.EvidenceGraph.connectDisambiguationPlaceholder,
                emptyLabel: L10n.EvidenceGraph.connectDisambiguationEmpty,
                label: L10n.EvidenceGraph.connectDisambiguationTermLabel,
                accessibilityIdentifierPrefix: "evidenceGraph.connect.disambiguation.term"
            )
        }
    }

    private var termBinding: Binding<String> {
        Binding(
            get: { model.pendingDisambiguation?.selectedTermID ?? "" },
            set: { model.selectDisambiguationTerm($0) }
        )
    }
}
