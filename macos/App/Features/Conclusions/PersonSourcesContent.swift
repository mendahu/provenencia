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
    static let lineFold = ConclusionFold(limit: 5, minimumHidden: 2)
    /// Source rows shown before "Show N more": fifty show whole; fifty-one show fifty and "Show 1 more".
    static let rowFold = ConclusionFold(limit: 50, minimumHidden: 1)

    /// A list that can be folded: the Source rows, or one Source's record lines.
    enum Fold: Hashable {
        case rows
        case lines(sourceID: String)
    }

    struct Row: Equatable, Identifiable {
        /// The Source id.
        var id: String
        var title: PVTitleSource
        /// The full title, for the tooltip.
        var helpTitle: String
        /// Nil when the type id is not in the types list.
        var typeLabel: String?
        var ref: String
        /// The row as one VoiceOver element: "Hamburg census 1867, Census, date not recorded, SRC-9TQ2M".
        var accessibilityLabel: String
        var location: WorkspaceLocation
        /// The lines shown, in the read's order.
        var lines: [Line]
        var hiddenCount: Int
    }

    struct Line: Equatable, Identifiable {
        /// The claim id.
        var id: String
        var subjectID: String
        var graphID: String
        /// The graph's name.
        var graphTitle: String
        /// Nil when the graph has no ref.
        var graphRef: String?
        var subjectName: String
        var subjectRef: String
        /// "Open Household 214, record Johann Weber, CPR-7Q2KD".
        var accessibilityLabel: String
        /// "Unpromote Household 214, record Johann Weber, CPR-7Q2KD".
        var unpromoteAccessibilityLabel: String
        var location: WorkspaceLocation
    }

    /// The rows shown: every Source, or the first `rowFold.limit` until unfolded, then a page at a time.
    let rows: [Row]
    /// The location of each graph's first line, folded or not, by graph id: where a node of the
    /// record links diagram opens.
    let graphLocations: [String: WorkspaceLocation]
    /// The graph of each claim, by claim id, folded or not.
    let claimGraphs: [String: String]
    /// The Sources the row fold leaves out.
    let hiddenRowCount: Int
    /// The claim id of the first line each fold's next page adds; absent when it hides nothing.
    private let revealed: [Fold: String]
    /// The graphs a fold hides all of: the ones a "Show more" stands for.
    private let foldedGraphs: [Fold: Set<String>]

    var isEmpty: Bool { rows.isEmpty }

    /// The claim id of the first line that the next "Show more" of `fold` adds, for the
    /// keyboard to land on; nil when the fold hides nothing.
    func firstRevealed(by fold: Fold) -> String? { revealed[fold] }

    /// The graphs whose record lines are all behind `fold` (a graph is on one Source): for one
    /// Source's lines, the graphs with a line behind its fold and none shown; for the rows, the
    /// graphs of the Sources the fold leaves out. The record links diagram highlights its Show
    /// more for them (S16-D2 FH-1, FH-2).
    func foldedGraphIDs(by fold: Fold) -> Set<String> { foldedGraphs[fold] ?? [] }

    /// `folds` holds the pages revealed of the Sources' lines and of the Sources past the fold.
    init(
        groups: [CatalogPersonClaimGroup],
        sourceTypes: [CatalogSourceType],
        orders: NameFormatOrders,
        folds: ConclusionFoldState<Fold> = ConclusionFoldState()
    ) {
        var graphLocations: [String: WorkspaceLocation] = [:]
        var claimGraphs: [String: String] = [:]
        var revealed: [Fold: String] = [:]
        var foldedGraphs: [Fold: Set<String>] = [:]
        /// Every graph of each Source, in `groups` order, for the rows fold.
        var sourceGraphs: [Set<String>] = []
        let all = groups.map { group in
            let lines = group.claims.map { claim -> Line in
                let title = claim.graphName
                let subjectName = EvidenceGraphPlacedSubject.displayName(
                    subject: claim.subject,
                    kind: .person,
                    observations: claim.observations,
                    eventTitle: nil,
                    orders: orders
                )
                let location = SourcesListNavigation.canvasLocation(
                    sourceId: group.sourceID,
                    sourceRef: group.sourceRef,
                    sourceTitle: group.title,
                    graphId: claim.graphID,
                    subjectId: claim.subject.id
                )
                if graphLocations[claim.graphID] == nil {
                    graphLocations[claim.graphID] = location
                }
                claimGraphs[claim.claimID] = claim.graphID
                return Line(
                    id: claim.claimID,
                    subjectID: claim.subject.id,
                    graphID: claim.graphID,
                    graphTitle: title,
                    graphRef: claim.graphRef.trimmedNonEmpty,
                    subjectName: subjectName,
                    subjectRef: claim.subject.ref,
                    accessibilityLabel: L10n.Conclusions.sourcesOpenGraph(
                        graph: title, subject: subjectName, ref: claim.subject.ref
                    ),
                    unpromoteAccessibilityLabel: L10n.Conclusions.sourcesUnpromoteAccessibility(
                        graph: title, subject: subjectName, ref: claim.subject.ref
                    ),
                    location: location
                )
            }
            let lineFold = Fold.lines(sourceID: group.sourceID)
            let fold = Self.lineFold.shown(lines, pages: folds.pages(lineFold))
            sourceGraphs.append(Set(lines.map(\.graphID)))
            if let first = Self.lineFold.firstRevealed(lines, pages: folds.pages(lineFold)) {
                revealed[lineFold] = first.id
                let shownGraphs = Set(fold.shown.map(\.graphID))
                foldedGraphs[lineFold] = Set(lines.dropFirst(fold.shown.count).map(\.graphID)).subtracting(shownGraphs)
            }
            let title = PVTitleSource.name(group.title)
            let typeLabel = sourceTypes.label(forTypeID: group.sourceTypeID)
            let date = L10n.string(L10n.Conclusions.sourcesDateNotRecorded)
            return Row(
                id: group.sourceID,
                title: title,
                helpTitle: title.text,
                typeLabel: typeLabel,
                ref: group.sourceRef,
                accessibilityLabel: L10n.DesignSystem.spokenJoin([
                    title.text, typeLabel ?? "", date, group.sourceRef,
                ]) ?? "",
                location: SourcesListNavigation.pageLocation(
                    sourceId: group.sourceID, ref: group.sourceRef, title: group.title
                ),
                lines: fold.shown,
                hiddenCount: fold.hidden
            )
        }
        let fold = Self.rowFold.shown(all, pages: folds.pages(.rows))
        rows = fold.shown
        if fold.hidden > 0 {
            let behind = all.dropFirst(fold.shown.count)
            if let id = behind.first?.lines.first?.id { revealed[.rows] = id }
            foldedGraphs[.rows] = sourceGraphs.dropFirst(fold.shown.count).reduce(into: Set<String>()) { $0.formUnion($1) }
        }
        self.revealed = revealed
        self.foldedGraphs = foldedGraphs
        self.graphLocations = graphLocations
        self.claimGraphs = claimGraphs
        hiddenRowCount = fold.hidden
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
