import Foundation
import SwiftUI

/// Spike 2's top-level workspace destinations (W-13, W-16), plus Files.
/// Raw values match the design board's kebab-case section ids.
enum WorkspaceSection: String, CaseIterable, Codable, Sendable {
    case sources
    case sourceTypes = "source-types"
    case sourceFields = "source-fields"
    case files

    var label: LocalizedStringResource {
        switch self {
        case .sources: L10n.Workspace.sourcesTitle
        case .sourceTypes: L10n.Workspace.sourceTypesTitle
        case .sourceFields: L10n.Workspace.sourceFieldsTitle
        case .files: L10n.Workspace.filesTitle
        }
    }

    var placeholderNote: LocalizedStringResource {
        // Only Files still uses the workspace placeholder host; Sources /
        // types / fields mount real destinations.
        L10n.Workspace.filesPlaceholderNote
    }

    var icon: PVSymbol {
        switch self {
        case .sources: .library
        case .sourceTypes: .tag
        case .sourceFields: .list
        case .files: .folderOpen
        }
    }
}
