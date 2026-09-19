import SwiftUI

/// Workspace destination for a Source’s Evidence graph.
///
/// Product place under Sources (`sourceSurface: .graph`). Composes the
/// reusable [`GraphCanvas`](../GraphCanvas/) shell with primary subject cards
/// (S6-02). Palette / create / drag land in S6-03.
struct EvidenceGraphView: View {
    /// Large enough to pan; Source-scoped graphs stay small (design note §7.1).
    private static let contentSize = CGSize(width: 4_000, height: 4_000)

    let sourceID: String
    let session: WorkspaceSession

    @State private var selectedSubjectID: String?
    @State private var sourceTitle: String = ""

    private var graphKey: CatalogQueryKey {
        CatalogQueryKey.sourceGraph(project: session.projectKey, sourceId: sourceID)
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<SourceGraphSnapshot> = session.queryHandle(graphKey) {
                EvidenceGraphContent(
                    handle: handle,
                    sourceTitle: sourceTitle,
                    contentSize: Self.contentSize,
                    selectedSubjectID: $selectedSubjectID
                )
            } else {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .background(PVColor.surfacePage)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .accessibilityIdentifier("workspace.destination.evidenceGraph")
        .task(id: sourceID) {
            let _: QueryHandle<SourceGraphSnapshot> = session.query(graphKey)
            await refreshSourceTitle()
        }
    }

    private func refreshSourceTitle() async {
        let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
        if let handle: QueryHandle<[CatalogSource]> = session.queryHandle(listKey),
           let sources = handle.value,
           let match = sources.first(where: { $0.id == sourceID })
        {
            sourceTitle = match.title
            return
        }
        let handle: QueryHandle<[CatalogSource]> = session.query(listKey)
        // Give the list a moment if it was just warmed.
        var waited = 0
        while handle.value == nil && waited < 40 {
            try? await Task.sleep(nanoseconds: 25_000_000)
            waited += 1
        }
        if let match = handle.value?.first(where: { $0.id == sourceID }) {
            sourceTitle = match.title
        }
    }
}

// MARK: - Content

private struct EvidenceGraphContent: View {
    @Bindable var handle: QueryHandle<SourceGraphSnapshot>
    let sourceTitle: String
    let contentSize: CGSize
    @Binding var selectedSubjectID: String?

    private var snapshot: SourceGraphSnapshot {
        handle.value ?? SourceGraphSnapshot(sourceId: "")
    }

    private var subjects: [SourceGraphPlacedSubject] {
        snapshot.subjects
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            canvas
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .background(PVColor.surfacePage)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.Workspace.evidenceGraphTitle))
        .accessibilityRepresentation {
            accessibilityList
        }
        .accessibilityRotor(String(localized: L10n.EvidenceGraph.subjectsRotor)) {
            ForEach(subjects) { placed in
                AccessibilityRotorEntry(
                    EvidenceSubjectCard.accessibilityLabel(for: placed),
                    id: placed.id
                ) {
                    selectedSubjectID = placed.id
                }
            }
        }
    }

    private var header: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
            Text(L10n.Workspace.evidenceGraphTitle)
                .font(PVFont.display(size: PVTypeScale.h1, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
            if !sourceTitle.isEmpty {
                Text(verbatim: sourceTitle)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
            }
            Spacer(minLength: 0)
            Text(L10n.EvidenceGraph.subjectCount(count: subjects.count))
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textFaint)
        }
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space5)
        .accessibilityElement(children: .combine)
    }

    private var canvas: some View {
        GraphCanvasScrollView(contentSize: contentSize) {
            ZStack(alignment: .topLeading) {
                GraphCanvasGridView(contentSize: contentSize)
                    .accessibilityHidden(true)

                ForEach(subjects) { placed in
                    EvidenceSubjectCard(
                        placed: placed,
                        isSelected: selectedSubjectID == placed.id,
                        onSelect: { selectedSubjectID = placed.id }
                    )
                    .position(
                        GraphCanvasGridMapping.contentPoint(
                            gridX: placed.gridX,
                            gridY: placed.gridY
                        )
                    )
                }

                if subjects.isEmpty, handle.status == .ready {
                    emptyOverlay
                }
            }
            .frame(width: contentSize.width, height: contentSize.height)
        }
    }

    private var emptyOverlay: some View {
        PVEmptyState(
            icon: .shapes,
            title: L10n.EvidenceGraph.emptyTitle,
            message: String(localized: L10n.EvidenceGraph.emptyMessage)
        )
        .frame(width: 420)
        .position(x: contentSize.width / 2, y: contentSize.height / 2)
        .allowsHitTesting(false)
    }

    private var accessibilityList: some View {
        List(selection: $selectedSubjectID) {
            ForEach(subjects) { placed in
                Text(verbatim: EvidenceSubjectCard.accessibilityLabel(for: placed))
                    .tag(placed.id)
                    .accessibilityIdentifier("evidenceGraph.subject.\(placed.id)")
                    .accessibilityAddTraits(
                        selectedSubjectID == placed.id ? [.isSelected] : []
                    )
            }
        }
    }
}
