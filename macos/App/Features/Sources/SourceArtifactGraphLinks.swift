import SwiftUI

/// The Evidence graphs an Artifact belongs to, as a bordered checklist: one
/// kit checkbox per graph with its name and ref. State is the unchecked set
/// (`excluded`), so a graph made later starts checked. A `locked` graph (its
/// id mapped to the reason) is checked, shows the lock mark and cannot be
/// cleared. Used by the Add artifact dialog and an Artifact card in edit.
/// `accessibilityIdentifierPrefix` roots the ids: `<prefix>`, `.<graph id>`.
struct SourceArtifactGraphLinks: View {
    let graphs: [CatalogEvidenceGraph]
    @Binding var excluded: Set<String>
    var locked: [String: String] = [:]
    let accessibilityIdentifierPrefix: String

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                LazyVStack(spacing: 0) {
                    ForEach(Array(graphs.enumerated()), id: \.element.id) { index, graph in
                        if index > 0 { PVDivider() }
                        row(graph)
                    }
                }
            }
            .frame(maxHeight: PVArtifactChecklist.maxRowsHeight)
        }
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .overlay {
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        }
        .accessibilityIdentifier(accessibilityIdentifierPrefix)
    }

    @ViewBuilder
    private func row(_ graph: CatalogEvidenceGraph) -> some View {
        Group {
            if let hint = locked[graph.id] {
                PVCheckbox(
                    verbatim: graph.name,
                    description: hint,
                    isChecked: .constant(true),
                    isLocked: true
                )
                .accessibilityValue(Text(verbatim: L10n.ArtifactChecklist.lockedValue(ref: graph.ref)))
            } else {
                PVCheckbox(
                    verbatim: graph.name,
                    description: graph.ref,
                    isChecked: Binding(
                        get: { !excluded.contains(graph.id) },
                        set: { on in
                            if on { excluded.remove(graph.id) } else { excluded.insert(graph.id) }
                        }
                    )
                )
            }
        }
        .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(graph.id)")
        .padding(.vertical, PVSpacing.space2)
        .padding(.horizontal, PVSpacing.space5)
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}
