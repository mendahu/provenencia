import Foundation

/// Names read from the session's lists, so a rename shows at once. `.none`
/// knows no name; labels then fall back to what the location carries.
struct LocationNames {
    var sourceTitle: @MainActor (String) -> String?
    var graphName: @MainActor (String) -> String?

    static var none: LocationNames {
        LocationNames(sourceTitle: { _ in nil }, graphName: { _ in nil })
    }

    init(
        sourceTitle: @escaping @MainActor (String) -> String?,
        graphName: @escaping @MainActor (String) -> String?
    ) {
        self.sourceTitle = sourceTitle
        self.graphName = graphName
    }

    @MainActor
    init(session: WorkspaceSession) {
        self.init(
            sourceTitle: { session.source(id: $0)?.title },
            graphName: { session.evidenceGraph(id: $0)?.name }
        )
    }
}

/// What a location is called in the crumbs, the Back / Forward menus and the
/// canvas's VoiceOver label: one home for all three.
@MainActor
enum WorkspaceLocationLabels {
    /// The record (`title ?? ref`) and, on a section of its page, the section's
    /// name. On an Evidence graph, its composer or its Promote page the leaf is
    /// the graph's name and the section is the surface.
    static func labels(
        for location: WorkspaceLocation,
        names: LocationNames = .none
    ) -> (leaf: String?, section: String?) {
        if let graphID = location.graphSurfaceID {
            return (graphLabel(graphID: graphID, names: names), surfaceLabel(for: location))
        }
        return (
            location.title?.trimmedNonEmpty ?? location.ref?.trimmedNonEmpty,
            location.pageSection.map { L10n.string($0.label) }
        )
    }

    /// The graph's name, when the session knows one.
    static func graphName(graphID: String, names: LocationNames) -> String? {
        names.graphName(graphID)?.trimmedNonEmpty
    }

    /// The graph's name, else "Evidence graph".
    static func graphLabel(graphID: String, names: LocationNames) -> String {
        graphName(graphID: graphID, names: names) ?? L10n.string(L10n.Workspace.evidenceGraphTitle)
    }

    /// The surface on a graph: the composer's Citation, the Promote page. The
    /// canvas and every other location have none.
    static func surfaceLabel(for location: WorkspaceLocation) -> String? {
        switch location.sourceSurface {
        case .promote:
            return location.promoteAll
                ? L10n.string(L10n.EvidenceGraph.promoteAll)
                : L10n.Promote.breadcrumb(ref: location.ref?.trimmedNonEmpty ?? scope(of: location))
        case .citationComposer:
            return L10n.CitationComposer.breadcrumbCitationFor(scope: scope(of: location))
        default:
            return nil
        }
    }

    /// What the surface is about: the title, else the ref, else an ellipsis.
    private static func scope(of location: WorkspaceLocation) -> String {
        location.title?.trimmedNonEmpty ?? location.ref?.trimmedNonEmpty ?? "…"
    }
}

extension String {
    var trimmedNonEmpty: String? {
        let trimmed = trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }
}

extension Optional where Wrapped == String {
    var trimmedNonEmpty: String? {
        self?.trimmedNonEmpty
    }
}
