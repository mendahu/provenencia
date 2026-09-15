import Foundation

/// Shared apply paths for destination views: turn `WorkspaceNavigation.currentLocation`
/// into model open/select/clear, or prune missing deep ids after load.
/// See `docs/deployment-plan/spike-3/navigation-history.md` and
/// `.cursor/skills/add-workspace-location`.
enum WorkspaceLocationApply {
    /// Omnibar Return / click: commit the hit and clear the results chrome.
    @MainActor
    static func activateOmnibarHit(
        _ hit: CatalogSearchHit,
        navigation: WorkspaceNavigation,
        results: OmnibarResultsModel
    ) {
        navigation.go(to: hit.location)
        results.clearAfterNavigate()
    }

    @MainActor
    static func applySources(
        location: WorkspaceLocation,
        model: SourcesModel,
        navigation: WorkspaceNavigation
    ) {
        guard location.section == .sources else { return }
        if let sourceId = location.sourceId {
            if model.sources.contains(where: { $0.id == sourceId }) {
                model.openSource(id: sourceId)
            } else {
                // Missing or deleted — only called after first load completes.
                navigation.fallbackToSectionRoot()
            }
        } else {
            model.closeSource()
        }
    }

    @MainActor
    static func applySourceFields(
        location: WorkspaceLocation,
        model: SourceFieldsModel,
        navigation: WorkspaceNavigation
    ) {
        guard location.section == .sourceFields else { return }
        if model.isAdding { return }
        if let fieldId = location.fieldId {
            if model.fields.contains(where: { $0.id == fieldId }) {
                model.select(fieldId)
            } else {
                // Missing or deleted — only called after first load completes.
                navigation.fallbackToSectionRoot()
            }
        } else {
            model.clearHistorySelection()
        }
    }

    @MainActor
    static func applySourceTypes(
        location: WorkspaceLocation,
        model: SourceTypesModel,
        navigation: WorkspaceNavigation
    ) {
        guard location.section == .sourceTypes else { return }
        if model.isAdding { return }
        if let typeId = location.typeId {
            if model.types.contains(where: { $0.id == typeId }) {
                model.select(typeId)
            } else {
                // Missing or deleted — only called after first load completes.
                navigation.fallbackToSectionRoot()
            }
        } else {
            model.clearHistorySelection()
        }
    }
}
