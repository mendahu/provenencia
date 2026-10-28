import SwiftUI

/// The Evidence graphs an Artifact belongs to, as a bordered checklist: one
/// kit checkbox per graph with its name and ref. State is the unchecked set
/// (`excluded`), so a graph made later starts checked. A `locked` graph (its
/// id mapped to its hint) is checked, shows the lock mark and cannot be
/// cleared. Used by the Add artifact dialog and an Artifact card in edit.
/// `accessibilityIdentifierPrefix` roots the ids: `<prefix>`, `.<graph id>`.
struct SourceArtifactGraphLinks: View {
    let graphs: [CatalogEvidenceGraph]
    @Binding var excluded: Set<String>
    var locked: [String: String] = [:]
    var isDisabled = false
    let accessibilityIdentifierPrefix: String

    var body: some View {
        PVChecklistFrame(items: graphs, id: \.id, row: row)
            .accessibilityIdentifier(accessibilityIdentifierPrefix)
    }

    private func row(_ graph: CatalogEvidenceGraph) -> some View {
        PVChecklistCheckbox(
            id: graph.id,
            title: graph.name,
            ref: graph.ref,
            description: graph.ref,
            lockReason: locked[graph.id],
            excluded: $excluded,
            isDisabled: isDisabled
        )
        .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(graph.id)")
        .padding(.vertical, PVSpacing.space2)
        .padding(.horizontal, PVSpacing.space5)
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

extension CatalogEvidenceGraphUnlinkBlock {
    /// The hint under a locked row: the Citations the engine says would be
    /// left or are using the Artifact, else the engine's own reason. One
    /// wording for the Artifact card and the graph form; no reason is
    /// decided here.
    var lockHint: String {
        citations.total > 0
            ? PVDeleteImpactCopy.heading(citations)
            : L10n.Errors.message(code: reason.wire)
    }
}
