import Foundation
import Testing
@testable import Provenencia

/// The metadata board and its page-level dialogs, driven against a test host so
/// the contract is the `MetadataBoardHost` protocol, not any store.
@Suite
@MainActor
struct MetadataBoardModelTests {
    private let invalid = catalogMetadataRefusal()
    private let boom = CoreInvokeError.failed(status: 2)

    @Test func saveFillsSuggestionThenReportsWrite() async {
        let host = TestMetadataBoardHost(entries: [catalogMetadataEntry(catalogMetadataField("f-author", label: "Author"), suggested: true)])
        let saved = catalogMetadataEntry(catalogMetadataField("f-author", label: "Author"), "Mary Robins", suggested: true)
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
        let a = catalogMetadataEntry(catalogMetadataField("f-a", label: "A"), "1", order: 0)
        let b = catalogMetadataEntry(catalogMetadataField("f-b", label: "B"), "2", order: 1)
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
        let author = catalogMetadataEntry(catalogMetadataField("f-author", label: "Author"), suggested: true)
        let repo = catalogMetadataEntry(catalogMetadataField("f-repo", label: "Repository"), "NRO")
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
        let saved = catalogMetadataEntry(catalogMetadataField("f-author", label: "Author"), "Mary", suggested: true)
        let host = TestMetadataBoardHost(entries: [saved])
        let emptied = catalogMetadataEntry(catalogMetadataField("f-author", label: "Author"), suggested: true)
        host.clearResult = [emptied]
        let board = MetadataBoardModel(host: host)

        let cleared = await board.clear(fieldID: "f-author")

        #expect(cleared)
        #expect(host.entries == [emptied])
        #expect(board.drafts["f-author"] == "")
        #expect(host.calls.filter { $0 == .didWrite }.count == 1)
    }

    @Test func clearConfirmTargetsItsBoard() async {
        let saved = catalogMetadataEntry(catalogMetadataField("f-author", label: "Author"), "Mary")
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

    @Test func liveBoardAsksBeforeClearing() async {
        let saved = catalogMetadataEntry(catalogMetadataField("f-author", label: "Author"), "Mary")
        let host = TestMetadataBoardHost(entries: [saved])
        let board = MetadataBoardModel(host: host)

        #expect(await board.requestClear(fieldID: "f-author"))
        #expect(!host.calls.contains(.clear("f-author")))

        host.confirmsClear = false
        #expect(await board.requestClear(fieldID: "f-author") == false)
        #expect(host.calls.contains(.clear("f-author")))
    }

    @Test func addRowOnLiveBoard() async {
        let host = TestMetadataBoardHost(entries: [])
        host.fields = [catalogMetadataField("film", label: "Film number")]
        let board = MetadataBoardModel(host: host)
        board.openAddRow()
        board.addRow?.fieldID = "film"
        board.addRow?.value = "F-7"

        host.setError = invalid
        host.failureText = "Bad"
        await board.confirmAddRow()
        #expect(board.addRow?.error == "Bad")
        #expect(board.addRow?.fieldID == "film" && board.addRow?.value == "F-7")
        #expect(!board.isSavingAddRow)

        host.setError = boom
        await board.confirmAddRow()
        #expect(host.calls.contains(.showFailure("Bad")))
        #expect(board.addRow?.error == nil)
        #expect(board.addRow != nil)
    }

    @Test func addRowSurvivesReset() {
        let host = TestMetadataBoardHost(entries: [])
        host.fields = [catalogMetadataField("film", label: "Film number")]
        let board = MetadataBoardModel(host: host)
        board.openAddRow()
        board.addRow?.fieldID = "film"
        board.addRow?.value = "F-7"

        board.resetDrafts()
        #expect(board.addRow == MetadataBoardAddRow(fieldID: "film", value: "F-7"))

        host.entries = [catalogMetadataEntry(catalogMetadataField("film", label: "Film number"), "F-1")]
        board.resetDrafts()
        #expect(board.addRow == MetadataBoardAddRow(fieldID: "", value: "F-7", takenFieldLabel: "Film number"))
    }

    @Test func addRowTakenFieldSaysSo() {
        let host = TestMetadataBoardHost(entries: [])
        let board = MetadataBoardModel(host: host)
        board.openAddRow()
        board.pickAddRowField("film")
        host.entries = [catalogMetadataEntry(catalogMetadataField("film", label: "Film number"), "F-1")]
        board.resetDrafts()
        #expect(board.addRow?.fieldID == "")
        #expect(board.addRow?.takenFieldLabel == "Film number")
    }

    @Test func pickingAFieldClearsTheNotice() {
        let host = TestMetadataBoardHost(entries: [])
        let board = MetadataBoardModel(host: host)
        board.openAddRow()
        board.addRow?.takenFieldLabel = "Film number"
        board.addRow?.error = "Bad"
        board.pickAddRowField("roll")
        #expect(board.addRow?.takenFieldLabel == nil)
        #expect(board.addRow?.error == nil)
        board.addRow?.error = "Bad"
        board.setAddRowValue("x")
        #expect(board.addRow?.error == nil)
    }

    @Test func unconfirmedInputReadsOpenRows() {
        let host = TestMetadataBoardHost(entries: [])
        let board = MetadataBoardModel(host: host)
        #expect(!board.hasUnconfirmedInput)
        board.openAddRow()
        #expect(!board.hasUnconfirmedInput)
        board.setAddRowValue("F-7")
        #expect(board.hasUnconfirmedInput)
        board.closeAddRow()
        #expect(!board.hasUnconfirmedInput)
        board.setTyped(true, fieldID: "provider")
        #expect(board.hasUnconfirmedInput)
        #expect(board.hasUnsavedChanges)
    }

    @Test func oneInnerEditAtATime() {
        let saved = catalogMetadataEntry(catalogMetadataField("f-repo", label: "Repository"), "NRO")
        let host = TestMetadataBoardHost(entries: [saved])
        let board = MetadataBoardModel(host: host)
        board.openAddRow()
        board.beginEdit(fieldID: "f-repo")
        #expect(board.addRow == nil)
        #expect(board.editingFieldID == "f-repo")
        board.openAddRow()
        #expect(board.editingFieldID == nil)
        #expect(board.addRow != nil)
    }

    @Test func dismissingATypedSuggestionClearsItsFlag() async {
        let image = catalogMetadataEntry(catalogMetadataField("image", label: "Image number"), suggested: true)
        let host = TestMetadataBoardHost(entries: [image])
        let board = MetadataBoardModel(host: host)
        board.setTyped(true, fieldID: "image")
        #expect(board.hasUnconfirmedInput)
        await board.dismissSuggestion(fieldID: "image")
        #expect(!board.hasUnconfirmedInput)
    }

    @Test func refusalWithTypedTextElsewhereAnswersFalse() {
        let provider = catalogMetadataEntry(catalogMetadataField("provider", label: "Provider"), "A")
        let url = catalogMetadataEntry(catalogMetadataField("url", label: "URL"), "https://x.org")
        let host = TestMetadataBoardHost(entries: [provider, url])
        host.failureText = "Bad"
        let board = MetadataBoardModel(host: host)
        board.setTyped(true, fieldID: "provider")
        let refusal = catalogMetadataRefusal(params: ["url"])
        #expect(board.showRefusal(refusal) == false)
        #expect(board.editingFieldID == nil)
        board.setTyped(false, fieldID: "provider")
        #expect(board.showRefusal(refusal))
        #expect(board.editingFieldID == "url")
    }

    @Test func invalidValueStaysOnTheField() async {
        let repo = catalogMetadataEntry(catalogMetadataField("f-repo", label: "Repository"), "NRO")
        let host = TestMetadataBoardHost(entries: [repo])
        host.setError = invalid
        host.failureText = "Bad"
        let board = MetadataBoardModel(host: host)

        board.beginEdit(fieldID: "f-repo")
        board.drafts["f-repo"] = "file:/tmp"
        await board.save(fieldID: "f-repo")
        #expect(board.fieldError == "Bad")
        #expect(board.fieldErrorID == "f-repo")

        board.openAddRow()
        board.pickAddRowField("f-new")
        board.setAddRowValue("javascript:alert(1)")
        await board.confirmAddRow()
        #expect(board.addRow?.error == "Bad")
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
    var confirmsClear = true
    var isChanged = false
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
