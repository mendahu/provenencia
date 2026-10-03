import Foundation
import SwiftUI

/// Top-level workspace sidebar destinations (W-13, W-16).
/// Raw values match kebab-case section ids in navigation history JSON.
enum WorkspaceSection: String, Sendable, CaseIterable, Codable {
    case sources
    case sourceTypes = "source-types"
    case metadata
    case properties
    /// Conclude: canonical handles. Persons since S9-07; Events and Places
    /// are stub pages until S9-23 / S9-26.
    case persons
    case events
    case places

    /// Retired section ids still found in saved navigation history (and
    /// accepted from any caller) mapped to their current section.
    static let legacyIDs: [String: WorkspaceSection] = [
        "files": .sources, // descoped Files destination
        "source-fields": .metadata, // renamed in S9-07b
        "subject-fields": .properties, // renamed in S9-07b
    ]

    /// Parses a section id from history, the engine, or the sidebar. Use this
    /// rather than `init(rawValue:)` so retired ids keep resolving.
    init?(id: String) {
        if let section = WorkspaceSection(rawValue: id) ?? Self.legacyIDs[id] {
            self = section
        } else {
            return nil
        }
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        let raw = try container.decode(String.self)
        guard let section = WorkspaceSection(id: raw) else {
            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Unknown workspace section: \(raw)"
            )
        }
        self = section
    }

    var label: LocalizedStringResource {
        switch self {
        case .sources: L10n.Workspace.sourcesTitle
        case .sourceTypes: L10n.Workspace.sourceTypesTitle
        case .metadata: L10n.Workspace.metadataTitle
        case .properties: L10n.Workspace.propertiesTitle
        case .persons: L10n.Workspace.personsTitle
        case .events: L10n.Workspace.eventsTitle
        case .places: L10n.Workspace.placesTitle
        }
    }

    var icon: PVSymbol {
        switch self {
        case .sources: .library
        case .sourceTypes: .tag
        case .metadata: .list
        case .properties: .listTree
        case .persons: .person
        case .events: .calendar
        case .places: .mapPin
        }
    }
}
