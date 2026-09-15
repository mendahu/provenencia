import Foundation
import SwiftUI

/// Top-level workspace sidebar destinations (W-13, W-16).
/// Raw values match kebab-case section ids in navigation history JSON.
enum WorkspaceSection: String, Sendable, CaseIterable, Codable {
    case sources
    case sourceTypes = "source-types"
    case sourceFields = "source-fields"

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        let raw = try container.decode(String.self)
        // Descoped Files destination — restore old history entries to Sources.
        if raw == "files" {
            self = .sources
        } else if let section = WorkspaceSection(rawValue: raw) {
            self = section
        } else {
            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Unknown workspace section: \(raw)"
            )
        }
    }

    var label: LocalizedStringResource {
        switch self {
        case .sources: L10n.Workspace.sourcesTitle
        case .sourceTypes: L10n.Workspace.sourceTypesTitle
        case .sourceFields: L10n.Workspace.sourceFieldsTitle
        }
    }

    var icon: PVSymbol {
        switch self {
        case .sources: .library
        case .sourceTypes: .tag
        case .sourceFields: .list
        }
    }
}
