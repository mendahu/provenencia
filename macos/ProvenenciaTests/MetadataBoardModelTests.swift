import Foundation
import Testing
@testable import Provenencia

/// The metadata board and its page-level dialogs, driven against a test host so
/// the contract is the `MetadataBoardHost` protocol, not any store.
@Suite
@MainActor
struct MetadataBoardModelTests {
    private let invalid = CoreInvokeError.coded(
        status: 1, code: "sourcemetadata.invalid", kind: .user, params: []
    )
    private let boom = CoreInvokeError.failed(status: 2)

    private func entry(
        _ id: String, label: String, value: String = "", suggested: Bool = false, order: Int32 = 0
    ) -> CatalogMetadataEntry {
        CatalogMetadataEntry(
            field: CatalogMetadataField(
                id: id, key: label.lowercased(), origin: "user", label: label, dataType: "text", description: ""
            ),
            valueText: value, hasValue: !value.isEmpty, suggested: suggested, sortOrder: order
        )
    }

    @Test func saveFillsSuggestionThenReportsWrite() async {
        let host = TestMetadataBoardHost(entries: [entry("f-author", label: "Author", suggested: true)])
        let saved = entry("f-author", label: "Author", value: "Mary Robins", suggested: true)
        host.setResult = [saved]
        let board = MetadataBoardModel(host: host)
        board.drafts["f-author"] = "Mary Robins"

        await board.save(fieldID: "f-author")

        #expect(host.calls == [.clearFailure, .set("f-author", "Mary Robins"), .didWrite])
        #expect(host.entries == [saved])
        #expect(host.entriesAtDidWrite == [saved])
        #expect(board.drafts["f-author"] == "Mary Robins")
        #expect(board.suggested.isEmpty)
        #expect(board.saved == [saved])
    }

    @Test func reorderFailureRevertsAndShowsFailure() async {
        let a = entry("f-a", label: "A", value: "1", order: 0)
        let b = entry("f-b", label: "B", value: "2", order: 1)
        let host = TestMetadataBoardHost(entries: [a, b])
        host.reorderError = boom
        host.failureText = "Boom"
        let board = MetadataBoardModel(host: host)

        await board.moveSaved(from: IndexSet(integer: 0), to: 2)

        #expect(host.entries == [a, b])
        #expect(host.calls.filter { $0 == .showFailure("Boom") }.count == 1)
        #expect(!host.calls.contains(.didWrite))
    }

    @Test func goneOwnerWritesNoState() async {
        let author = entry("f-author", label: "Author", suggested: true)
        let repo = entry("f-repo", label: "Repository", value: "NRO")
        let host = TestMetadataBoardHost(entries: [author, repo])
        host.failureText = nil
        let board = MetadataBoardModel(host: host)

        host.setError = invalid
        board.drafts["f-author"] = "Mary"
        await board.save(fieldID: "f-author")
        #expect(board.fieldError == nil)
        #expect(host.entries == [author, repo])

        host.clearError = boom
        #expect(await board.clear(fieldID: "f-repo") == false)

        host.reorderError = boom
        await board.moveSaved(from: IndexSet(integer: 0), to: 1)
        #expect(host.entries == [author, repo])
        #expect(!host.calls.contains { if case .showFailure = $0 { true } else { false } })
        #expect(!host.calls.contains(.didWrite))
    }

    @Test func clearTakesTheHostsRows() async {
        let saved = entry("f-author", label: "Author", value: "Mary", suggested: true)
        let host = TestMetadataBoardHost(entries: [saved])
        let emptied = entry("f-author", label: "Author", suggested: true)
        host.clearResult = [emptied]
        let board = MetadataBoardModel(host: host)

        let cleared = await board.clear(fieldID: "f-author")

        #expect(cleared)
        #expect(host.entries == [emptied])
        #expect(board.drafts["f-author"] == "")
        #expect(host.calls.filter { $0 == .didWrite }.count == 1)
    }

    @Test func addDialogWritesToItsTargetBoard() async {
        let hostA = TestMetadataBoardHost(entries: [])
        let hostB = TestMetadataBoardHost(entries: [])
        hostB.setResult = [entry("f-author", label: "Author", value: "Ancestry")]
        let boardA = MetadataBoardModel(host: hostA)
        let boardB = MetadataBoardModel(host: hostB)
        let dialogs = MetadataBoardDialogs()

        dialogs.openAdd(on: boardB)
        #expect(dialogs.isAdding)
        dialogs.addFieldID = "f-author"
        dialogs.addValue = "Ancestry"
        await dialogs.createFromAdd()

        #expect(hostB.calls.contains(.set("f-author", "Ancestry")))
        #expect(hostA.calls.isEmpty)
        #expect(boardA.entries.isEmpty)
        #expect(!dialogs.isAdding)
    }

    @Test func clearConfirmTargetsItsBoard() async {
        let saved = entry("f-author", label: "Author", value: "Mary")
        let hostA = TestMetadataBoardHost(entries: [saved])
        let hostB = TestMetadataBoardHost(entries: [saved])
        let boardA = MetadataBoardModel(host: hostA)
        _ = MetadataBoardModel(host: hostB)
        let dialogs = MetadataBoardDialogs()

        dialogs.askClear(on: boardA, fieldID: "f-author")
        #expect(dialogs.pendingDelete?.label == "Author")
        await dialogs.confirmClear()

        #expect(hostA.calls.contains(.clear("f-author")))
        #expect(hostB.calls.isEmpty)
        #expect(dialogs.pendingDelete == nil)

        hostA.entries = [saved]
        hostA.clearError = boom
        hostA.failureText = "Boom"
        dialogs.askClear(on: boardA, fieldID: "f-author")
        await dialogs.confirmClear()
        #expect(dialogs.pendingDelete != nil)
    }

    @Test func invalidValueStaysOnTheFieldOrDialog() async {
        let repo = entry("f-repo", label: "Repository", value: "NRO")
        let host = TestMetadataBoardHost(entries: [repo])
        host.setError = invalid
        host.failureText = "Bad"
        let board = MetadataBoardModel(host: host)

        board.beginEdit(fieldID: "f-repo")
        board.drafts["f-repo"] = "file:/tmp"
        await board.save(fieldID: "f-repo")
        #expect(board.fieldError == "Bad")
        #expect(board.fieldErrorID == "f-repo")

        let dialogs = MetadataBoardDialogs()
        dialogs.openAdd(on: board)
        dialogs.addFieldID = "f-new"
        dialogs.addValue = "javascript:alert(1)"
        await dialogs.createFromAdd()
        #expect(dialogs.addValueError == "Bad")
        #expect(dialogs.isAdding)
        #expect(!host.calls.contains { if case .showFailure = $0 { true } else { false } })
    }
}

/// A host that records every call in order and answers with scripted rows or errors.
@MainActor
private final class TestMetadataBoardHost: MetadataBoardHost {
    enum Call: Equatable {
        case clearFailure
        case set(String, String)
        case clear(String)
        case dismiss(String)
        case reorder([String])
        case didWrite
        case showFailure(String)
    }

    var entries: [CatalogMetadataEntry]
    var fields: [CatalogMetadataField] = []
    private(set) var calls: [Call] = []
    /// What `entries` held when `didWrite` ran.
    private(set) var entriesAtDidWrite: [CatalogMetadataEntry] = []
    var setResult: [CatalogMetadataEntry] = []
    var clearResult: [CatalogMetadataEntry] = []
    var dismissResult: [CatalogMetadataEntry] = []
    var setError: Error?
    var clearError: Error?
    var reorderError: Error?
    var failureText: String? = "Boom"

    init(entries: [CatalogMetadataEntry]) {
        self.entries = entries
    }

    func set(fieldID: String, valueText: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        calls.append(.set(fieldID, valueText))
        if let setError { throw setError }
        return (setResult, CatalogWriteEffects())
    }

    func clear(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        calls.append(.clear(fieldID))
        if let clearError { throw clearError }
        return (clearResult, CatalogWriteEffects())
    }

    func dismissSuggestion(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        calls.append(.dismiss(fieldID))
        return (dismissResult, CatalogWriteEffects())
    }

    func reorder(fieldIDs: [String]) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        calls.append(.reorder(fieldIDs))
        if let reorderError { throw reorderError }
        return (entries, CatalogWriteEffects())
    }

    func didWrite(_: CatalogWriteEffects) {
        calls.append(.didWrite)
        entriesAtDidWrite = entries
    }

    func clearFailure() { calls.append(.clearFailure) }
    func failureMessage(for _: Error) -> String? { failureText }
    func showFailure(_ message: String) { calls.append(.showFailure(message)) }
}
