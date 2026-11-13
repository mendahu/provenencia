import Foundation
import Testing
@testable import Provenencia

/// The draft board against the engine's rules: the scenarios in
/// `core/database/sourcemetadata/testdata/draft_board_scenarios.json`, which Go's
/// `TestDraftBoardScenarios` runs through the live writes and `Apply`. Each runs here through
/// `MetadataBoardModel` over `MetadataDraftHost` and must give the file's board and draft.
@Suite
@MainActor
struct MetadataDraftHostParityTests {
    private struct File: Decodable {
        var suggestions: [String]
        var extras: [String]
        var scenarios: [Scenario]
    }

    private struct Scenario: Decodable {
        var name: String
        var knownDivergence: String?
        var start: Board
        var actions: [Action]
        var board: Board
        var draft: Draft?

        enum CodingKeys: String, CodingKey {
            case name, start, actions, board, draft
            case knownDivergence = "known_divergence"
        }
    }

    private struct Board: Decodable, Equatable {
        var saved: [Saved]
        var empty: [String]
    }

    private struct Saved: Decodable, Equatable {
        var field: String
        var value: String
        var suggested: Bool
    }

    private struct Action: Decodable {
        var op: String
        var field: String
        var value: String?
        var before: String?
    }

    private struct Draft: Decodable, Equatable {
        var values: [Value]
        var layout: [Layout]

        struct Value: Decodable, Equatable {
            var field: String
            var value: String
        }

        struct Layout: Decodable, Equatable {
            var field: String
            var sortOrder: Int
            var dismissed: Bool
        }
    }

    /// The file, found from this source file up to the repo root.
    private static func load() throws -> File {
        let root = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent().deletingLastPathComponent().deletingLastPathComponent()
        let url = root.appendingPathComponent("core/database/sourcemetadata/testdata/draft_board_scenarios.json")
        return try JSONDecoder().decode(File.self, from: Data(contentsOf: url))
    }

    private static func board(of host: MetadataDraftHost) -> Board {
        Board(
            saved: host.entries.filter(\.hasValue).map {
                Saved(field: $0.field.id, value: $0.valueText, suggested: $0.suggested)
            },
            empty: host.entries.filter { !$0.hasValue && $0.suggested }.map(\.field.id)
        )
    }

    private static func draft(of host: MetadataDraftHost) -> Draft? {
        host.draft.map { sent in
            Draft(
                values: sent.values.map { .init(field: $0.fieldID, value: $0.valueText) },
                layout: sent.layout.map { .init(field: $0.fieldID, sortOrder: Int($0.sortOrder), dismissed: $0.dismissed) }
            )
        }
    }

    /// The board with what the scenario's known divergence says the draft host differs on left out.
    private static func compared(_ board: Board, divergence: String?) -> Board {
        var board = board
        switch divergence {
        case "order": board.saved.sort { $0.field < $1.field }
        case "suggested": board.saved = board.saved.map { Saved(field: $0.field, value: $0.value, suggested: false) }
        default: break
        }
        return board
    }

    @Test func scenariosMatchTheEngine() async throws {
        let file = try Self.load()
        #expect(!file.scenarios.isEmpty)
        let fields = (file.suggestions + file.extras).map { catalogMetadataField($0, label: $0) }
        for scenario in file.scenarios {
            var rows: [CatalogMetadataEntry] = []
            for saved in scenario.start.saved {
                rows.append(catalogMetadataEntry(
                    catalogMetadataField(saved.field, label: saved.field), saved.value,
                    suggested: saved.suggested, order: Int32(rows.count)
                ))
            }
            for id in scenario.start.empty {
                rows.append(catalogMetadataEntry(
                    catalogMetadataField(id, label: id), suggested: true, order: Int32(rows.count)
                ))
            }
            let host = MetadataDraftHost(entries: rows, fields: fields)
            let board = MetadataBoardModel(host: host)
            #expect(Self.board(of: host) == scenario.start, "\(scenario.name): start")

            for action in scenario.actions {
                switch action.op {
                case "set":
                    let value = action.value ?? ""
                    if host.entries.contains(where: { $0.field.id == action.field }) {
                        board.drafts[action.field] = value
                        await board.save(fieldID: action.field)
                    } else {
                        _ = await board.add(fieldID: action.field, valueText: value)
                    }
                case "clear": _ = await board.clear(fieldID: action.field)
                case "dismiss": await board.dismissSuggestion(fieldID: action.field)
                case "move":
                    let saved = board.saved.map(\.field.id)
                    let from = try #require(saved.firstIndex(of: action.field))
                    let to = action.before.flatMap { saved.firstIndex(of: $0) } ?? saved.count
                    await board.moveSaved(from: IndexSet(integer: from), to: to)
                default:
                    Issue.record("\(scenario.name): unknown op \(action.op)")
                }
            }

            let divergence = scenario.knownDivergence
            #expect(
                Self.compared(Self.board(of: host), divergence: divergence)
                    == Self.compared(scenario.board, divergence: divergence),
                "\(scenario.name): board"
            )
            #expect(Self.draft(of: host) == scenario.draft, "\(scenario.name): draft")
        }
    }
}
