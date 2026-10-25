import SwiftUI

/// The Source's Evidence graphs: one card per graph with Open and Delete.
struct SourcePageGraphsView: View {
    @Bindable var model: SourcePageModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVSectionHeader(
                title: L10n.Sources.graphsHeading,
                meta: model.graphs.rows.map { L10n.Sources.graphsCount($0.count) }
            )

            if let rows = model.graphs.rows {
                VStack(alignment: .leading, spacing: PVSpacing.space4) {
                    ForEach(rows, id: \.id) { graph in
                        SourcePageGraphRow(model: model, graph: graph)
                    }
                }
            } else if let error = model.graphs.loadError {
                PVCallout(tone: .danger, message: L10n.Errors.message(for: error))
                    .accessibilityIdentifier("sources.page.graphs.error")
            } else {
                loadingRow
            }
        }
        .padding(.top, PVSpacing.space11 - PVSpacing.space9)
        .accessibilityIdentifier("sources.page.graphs")
    }

    /// One card laid out like a row: name, ref and description bars; Open and Delete bars.
    private var loadingRow: some View {
        PVCard(elevated: false) {
            HStack(alignment: .top, spacing: PVSpacing.space6) {
                VStack(alignment: .leading, spacing: PVSpacing.space4) {
                    PVSkeletonBar(width: 320, height: 12)
                    PVSkeletonBar(width: 140, height: 10)
                    PVSkeletonBar(width: 520, height: 10)
                }
                Spacer(minLength: PVSpacing.space6)
                HStack(spacing: PVSpacing.space4) {
                    PVSkeletonBar(width: 72, height: 28)
                    PVSkeletonBar(width: 28, height: 28)
                }
            }
            .padding(PVSpacing.space6)
        }
        .accessibilityHidden(true)
        .accessibilityIdentifier("sources.page.graphs.loading")
    }
}

private struct SourcePageGraphRow: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Bindable var model: SourcePageModel
    let graph: CatalogEvidenceGraph

    var body: some View {
        let location = model.graphs.location(for: graph)
        let reason = model.graphs.blockedReason(for: graph)
        let description = graph.description.trimmingCharacters(in: .whitespacesAndNewlines)
        PVCard(elevated: false) {
            HStack(alignment: .top, spacing: PVSpacing.space6) {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(verbatim: graph.name)
                        .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(2)
                    Text(verbatim: graph.ref)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                    if !description.isEmpty {
                        Text(verbatim: description)
                            .font(PVFont.body(size: PVTypeScale.bodySmall))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(3)
                    }
                }
                .accessibilityElement(children: .combine)
                Spacer(minLength: PVSpacing.space6)
                HStack(spacing: PVSpacing.space4) {
                    PVButton(L10n.Sources.openEvidenceGraph, variant: .secondary, size: .sm, icon: .gitBranch) {
                        if let location {
                            navigation.go(to: location)
                        }
                    }
                    .disabled(location == nil)
                    .help(reason ?? L10n.string(L10n.Sources.openEvidenceGraph))
                    .accessibilityLabel(Text(L10n.Sources.openEvidenceGraphLabel(name: graph.name)))
                    .accessibilityHint(Text(reason ?? ""))
                    .accessibilityIdentifier("sources.page.graph.open.\(graph.ref)")

                    PVIconButton(
                        .trash,
                        label: L10n.Sources.deleteEvidenceGraphLabel(name: graph.name),
                        size: .sm,
                        tone: .danger
                    ) {
                        Task { await model.askDeleteGraph(id: graph.id) }
                    }
                    .disabled(model.source == nil)
                    .accessibilityIdentifier("sources.page.graph.delete.\(graph.ref)")
                }
                .fixedSize()
            }
            .padding(PVSpacing.space6)
        }
    }
}
