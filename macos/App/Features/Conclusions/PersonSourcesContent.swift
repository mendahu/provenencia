import Foundation

/// What a Person's Sources section shows (S13-16, board S13-D6): one row per Source the
/// person's records come from, each with one line per record on that Source's Evidence
/// graphs. Pure; the section draws it.
///
/// Go gives the Sources in title order and each Source's records in graph name order; this
/// never sorts. It holds formatted strings (the record names), so a name-format change is a
/// content change.
struct PersonSourcesContent: Equatable {
    /// Lines shown under a Source before "Show N more", and the fewest a fold must hide:
    /// six lines show whole; seven show five and "Show 2 more".
    static let lineFoldLimit = 5
    static let lineFoldMinimumHidden = 2
    /// Source rows shown before "Show N more": fifty show whole; fifty-one show fifty and "Show 1 more".
    static let rowFoldLimit = 50

    struct Row: Equatable, Identifiable {
        /// The Source id.
        var id: String
        var title: PVTitleSource
        /// The full title, for the tooltip.
        var helpTitle: String
        /// Nil when the type id is not in the types list.
        var typeLabel: String?
        var typeIconKey: String?
        var ref: String
        /// The row as one VoiceOver element: "Hamburg census 1867, Census, date not recorded, SRC-9TQ2M".
        var accessibilityLabel: String
        var location: WorkspaceLocation
        /// The lines shown, in the read's order.
        var lines: [Line]
        var hiddenCount: Int
        /// The line an unfold adds first, whether or not the row is unfolded; nil when the row does not fold.
        var firstRevealedLineID: String?
    }

    struct Line: Equatable, Identifiable {
        /// The claim id.
        var id: String
        var subjectID: String
        var graphID: String
        /// The graph's name, else its ref.
        var graphTitle: String
        /// Nil when the graph has no ref, or its ref is already the title.
        var graphRef: String?
        var subjectName: String
        var subjectRef: String
        /// "Open Household 214, record Johann Weber, CPR-7Q2KD".
        var accessibilityLabel: String
        /// "Unpromote Household 214, record Johann Weber, CPR-7Q2KD".
        var unpromoteAccessibilityLabel: String
        var location: WorkspaceLocation
    }

    /// The rows shown: every Source, or the first `rowFoldLimit` until unfolded.
    let rows: [Row]
    /// The Sources the row fold leaves out.
    let hiddenRowCount: Int
    /// The first line of the first Source an unfold adds, whether or not the rows are unfolded; nil when the rows do not fold.
    let firstRevealedRowLineID: String?

    var isEmpty: Bool { rows.isEmpty }

    /// `expanded` holds the ids of the Sources whose lines are unfolded; `rowsExpanded` unfolds the Sources past the fold.
    init(
        groups: [CatalogPersonClaimGroup],
        sourceTypes: [CatalogSourceType],
        orders: NameFormatOrders,
        expanded: Set<String>,
        rowsExpanded: Bool = false
    ) {
        let all = groups.map { group in
            let lines = group.claims.map { claim -> Line in
                let name = claim.graphName.trimmedNonEmpty
                let ref = claim.graphRef.trimmedNonEmpty
                let title = name ?? ref ?? ""
                let subjectName = EvidenceGraphPlacedSubject.displayName(
                    subject: claim.subject,
                    kind: .person,
                    observations: claim.observations,
                    eventTitle: nil,
                    orders: orders
                )
                return Line(
                    id: claim.claimID,
                    subjectID: claim.subject.id,
                    graphID: claim.graphID,
                    graphTitle: title,
                    graphRef: name == nil ? nil : ref,
                    subjectName: subjectName,
                    subjectRef: claim.subject.ref,
                    accessibilityLabel: L10n.Conclusions.sourcesOpenGraph(
                        graph: title, subject: subjectName, ref: claim.subject.ref
                    ),
                    unpromoteAccessibilityLabel: L10n.Conclusions.sourcesUnpromoteAccessibility(
                        graph: title, subject: subjectName, ref: claim.subject.ref
                    ),
                    location: SourcesListNavigation.canvasLocation(
                        sourceId: group.sourceID,
                        sourceRef: group.sourceRef,
                        sourceTitle: group.title,
                        graphId: claim.graphID,
                        subjectId: claim.subject.id
                    )
                )
            }
            let fold = ConclusionFold.shown(
                lines,
                limit: Self.lineFoldLimit,
                minimumHidden: Self.lineFoldMinimumHidden,
                expanded: expanded.contains(group.sourceID)
            )
            let title: PVTitleSource = group.title.trimmedNonEmpty.map { .name($0) } ?? .ref(group.sourceRef)
            let typeLabel = sourceTypes.label(forTypeID: group.sourceTypeID)
            let date = L10n.string(L10n.Conclusions.sourcesDateNotRecorded)
            return Row(
                id: group.sourceID,
                title: title,
                helpTitle: title.text,
                typeLabel: typeLabel,
                typeIconKey: sourceTypes.iconKey(forTypeID: group.sourceTypeID),
                ref: group.sourceRef,
                // An untitled Source's title is its ref, which is not spoken twice.
                accessibilityLabel: L10n.DesignSystem.spokenJoin([
                    title.text, typeLabel ?? "", date, title == .ref(group.sourceRef) ? "" : group.sourceRef,
                ]) ?? "",
                location: SourcesListNavigation.pageLocation(
                    sourceId: group.sourceID, ref: group.sourceRef, title: group.title
                ),
                lines: fold.shown,
                hiddenCount: fold.hidden,
                firstRevealedLineID: ConclusionFold.firstRevealed(
                    lines, limit: Self.lineFoldLimit, minimumHidden: Self.lineFoldMinimumHidden
                )?.id
            )
        }
        let fold = ConclusionFold.shown(all, limit: Self.rowFoldLimit, minimumHidden: 1, expanded: rowsExpanded)
        rows = fold.shown
        hiddenRowCount = fold.hidden
        firstRevealedRowLineID = ConclusionFold.firstRevealed(all, limit: Self.rowFoldLimit, minimumHidden: 1)?.lines.first?.id
    }

    /// The claim id of the line to focus once the line of Subject `subjectID` is gone: the next
    /// visible line in reading order (Sources in order, their visible lines in order), else the
    /// previous one, else nil. A line folded away is never returned. A Subject has one accepted
    /// claim, so its id names one line.
    func lineAfterRemoving(subjectID: String) -> String? {
        let lines = rows.flatMap(\.lines)
        guard let index = lines.firstIndex(where: { $0.subjectID == subjectID }) else { return nil }
        if index + 1 < lines.count { return lines[index + 1].id }
        return index > 0 ? lines[index - 1].id : nil
    }
}
