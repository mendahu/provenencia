import Foundation
import Observation
import os
import Testing
@testable import Provenencia

/// The metadata board over `MetadataDraftHost`: every action changes the draft and writes nothing.
@Suite
@MainActor
struct MetadataDraftHostTests {
    private let provider = catalogMetadataField("provider", label: "Provider")
    private let url = catalogMetadataField("url", label: "URL")
    private let roll = catalogMetadataField("roll", label: "Roll")
    private let film = catalogMetadataField("film", label: "Film number")

    private func makeHost() -> MetadataDraftHost {
        var urlRow = catalogMetadataEntry(url, "https://example.org/a", suggested: true, order: 1)
        urlRow.openHref = "https://example.org/a"
        return MetadataDraftHost(
            entries: [
                catalogMetadataEntry(provider, suggested: true, order: 0),
                urlRow,
                catalogMetadataEntry(roll, "12", order: 2),
            ],
            fields: [provider, url, roll, film]
        )
    }

    private func ids(_ rows: [CatalogMetadataEntry]) -> [String] { rows.map(\.field.id) }

    /// T-9: a view over a draft board redraws after each kind of change.
    @Test func draftHostIsObservable() async {
        func fires(_ change: (MetadataDraftHost) -> Void) -> Bool {
            let host = makeHost()
            let fired = OSAllocatedUnfairLock(initialState: false)
            withObservationTracking { _ = host.entries } onChange: { fired.withLock { $0 = true } }
            change(host)
            return fired.withLock { $0 }
        }
        #expect(fires { _ = $0.clear(fieldID: "roll") })
        #expect(fires { _ = $0.dismissSuggestion(fieldID: "provider") })
        #expect(fires { _ = $0.reorder(fieldIDs: ["roll", "url", "provider"]) })
        #expect(fires { _ = $0.set(fieldID: "provider", valueText: "Ancestry") })
    }

    @Test func boardActionsBuildTheDraft() async throws {
        let host = makeHost()
        let board = MetadataBoardModel(host: host)
        let store = FakeStore()
        store.artifactsBySource["s1"] = [catalogArtifact("a1", label: "Page 1")]

        board.drafts["provider"] = "Ancestry"
        await board.save(fieldID: "provider")
        #expect(host.entries.first { $0.field.id == "provider" }?.valueText == "Ancestry")
        #expect(host.entries.first { $0.field.id == "provider" }?.hasValue == true)
        _ = await board.add(fieldID: "film", valueText: "F-7")
        let filmRow = try #require(host.entries.last)
        #expect(filmRow.field.id == "film" && filmRow.hasValue && !filmRow.suggested)
        #expect(store.updateArtifactCalls == 0)

        let sent = host.draft
        _ = try await store.updateArtifact(
            projectDir: "/tmp/p", userID: "u1", artifactID: "a1", label: "Page 1", description: "",
            graphIDs: nil, metadata: sent
        )
        #expect(store.updateArtifactMetadata == [sent])
        let draft = try #require(sent)
        #expect(draft.values == [
            .init(fieldID: "provider", valueText: "Ancestry"),
            .init(fieldID: "url", valueText: "https://example.org/a"),
            .init(fieldID: "roll", valueText: "12"),
            .init(fieldID: "film", valueText: "F-7"),
        ])
        #expect(draft.layout == [
            .init(fieldID: "provider", sortOrder: 0, dismissed: false),
            .init(fieldID: "url", sortOrder: 1, dismissed: false),
            .init(fieldID: "roll", sortOrder: 2, dismissed: false),
            .init(fieldID: "film", sortOrder: 3, dismissed: false),
        ])
    }

    @Test func clearKeepsSuggestionsAndDropsExtras() async {
        let host = makeHost()
        let board = MetadataBoardModel(host: host)
        _ = await board.clear(fieldID: "url")
        _ = await board.clear(fieldID: "roll")
        _ = await board.clear(fieldID: "film")
        let urlRow = host.entries.first { $0.field.id == "url" }
        #expect(urlRow?.suggested == true && urlRow?.hasValue == false && urlRow?.openHref == "")
        #expect(!ids(host.entries).contains("roll"))
        #expect(!(host.draft?.values.contains { $0.fieldID == "url" || $0.fieldID == "roll" } ?? false))
    }

    @Test func dismissOnlyEmptySuggestions() async throws {
        let host = makeHost()
        let board = MetadataBoardModel(host: host)
        let before = host.entries
        await board.dismissSuggestion(fieldID: "url")
        await board.dismissSuggestion(fieldID: "roll")
        #expect(host.entries == before)
        #expect(!host.isChanged)

        await board.dismissSuggestion(fieldID: "provider")
        #expect(!ids(host.entries).contains("provider"))
        let last = try #require(host.draft?.layout.last)
        #expect(last == .init(fieldID: "provider", sortOrder: 2, dismissed: true))
    }

    @Test func readdingADismissedFieldListsItOnce() async throws {
        let host = makeHost()
        let board = MetadataBoardModel(host: host)
        await board.dismissSuggestion(fieldID: "provider")
        _ = await board.add(fieldID: "provider", valueText: "Ancestry")
        let row = try #require(host.entries.first { $0.field.id == "provider" })
        #expect(row.hasValue && !row.suggested)
        let listed = try #require(host.draft).layout.filter { $0.fieldID == "provider" }
        #expect(listed == [.init(fieldID: "provider", sortOrder: Int32(host.entries.count - 1), dismissed: false)])
    }

    @Test func draftClearsWithoutConfirm() async {
        let host = makeHost()
        let board = MetadataBoardModel(host: host)
        #expect(await board.requestClear(fieldID: "roll") == false)
        #expect(!ids(host.entries).contains("roll"))
    }

    @Test func revertedDraftIsUnchanged() async throws {
        let host = makeHost()
        let board = MetadataBoardModel(host: host)

        board.drafts["provider"] = "Ancestry"
        await board.save(fieldID: "provider")
        _ = await board.clear(fieldID: "provider")
        #expect(!host.isChanged)
        #expect(host.draft == nil)

        board.beginEdit(fieldID: "url")
        board.drafts["url"] = "https://example.org/b"
        await board.save(fieldID: "url")
        #expect(host.isChanged)
        board.beginEdit(fieldID: "url")
        board.drafts["url"] = "https://example.org/a"
        await board.save(fieldID: "url")
        #expect(!host.isChanged)
        #expect(host.entries.first { $0.field.id == "url" }?.openHref == "https://example.org/a")

        await board.moveSaved(from: IndexSet(integer: 1), to: 0)
        #expect(ids(host.entries) == ["roll", "url", "provider"])
        #expect(try #require(host.draft).layout.map(\.sortOrder) == [0, 1, 2])
        await board.moveSaved(from: IndexSet(integer: 1), to: 0)
        #expect(ids(host.entries) == ["url", "roll", "provider"])
        #expect(!host.isChanged)
        #expect(host.draft == nil)

        let a = catalogMetadataEntry(provider, "1", order: 5)
        let b = catalogMetadataEntry(roll, "2", order: 6)
        let reordered = [a, b].reordered(by: ["roll", "x", "roll"])
        #expect(ids(reordered) == ["roll", "provider"])
        #expect(reordered.map(\.sortOrder) == [0, 1])
    }

    @Test func unsavedRowsAreTheTypesSuggestions() {
        let rows = MetadataDraftHost.unsavedRows(from: [
            CatalogTypeSuggestion(field: provider, sortOrder: 0),
            CatalogTypeSuggestion(field: url, sortOrder: 1),
        ])
        #expect(ids(rows) == ["provider", "url"])
        #expect(rows.allSatisfy { $0.suggested && !$0.hasValue && $0.valueText.isEmpty })
        #expect(MetadataDraftHost.unsavedRows(from: []).isEmpty)
    }

    @Test func addRowConfirmsOnlyWhenFilled() async {
        let host = makeHost()
        let board = MetadataBoardModel(host: host)
        #expect(!board.canConfirmAddRow)
        board.openAddRow()
        #expect(!board.canConfirmAddRow)
        let before = host.entries.count
        board.addRow?.value = "F-7"
        #expect(!board.canConfirmAddRow)
        await board.confirmAddRow()
        #expect(board.addRow != nil)
        #expect(host.entries.count == before)
        board.addRow?.value = ""
        board.addRow?.fieldID = "film"
        board.addRow?.value = "  "
        #expect(!board.canConfirmAddRow)
        await board.confirmAddRow()
        #expect(board.addRow != nil)
        #expect(!ids(host.entries).contains("film"))
        board.addRow?.value = " F-7 "
        #expect(board.canConfirmAddRow)
        await board.confirmAddRow()
        #expect(host.entries.last?.field.id == "film")
        #expect(host.entries.last?.valueText == "F-7")
        #expect(board.addRow == nil)
    }

    @Test func refusalLandsOnItsRow() {
        let host = makeHost()
        let board = MetadataBoardModel(host: host)
        func invalid(_ params: [String]) -> CoreInvokeError { catalogMetadataRefusal(params: params) }
        let notMetadata = CoreInvokeError.coded(status: 1, code: "artifacts.invalid", kind: .user, params: ["url"])

        board.beginEdit(fieldID: "roll")
        let error = invalid(["url"])
        #expect(board.showRefusal(error))
        #expect(board.editingFieldID == "url")
        #expect(board.drafts["roll"] == "12")
        #expect(board.drafts["url"] == "https://example.org/a")
        #expect(board.fieldErrorID == "url")
        #expect(board.fieldError == L10n.Errors.message(for: error))

        let typedHere = MetadataBoardModel(host: makeHost())
        typedHere.beginEdit(fieldID: "roll")
        typedHere.drafts["roll"] = "13"
        typedHere.setTyped(true, fieldID: "roll")
        #expect(typedHere.showRefusal(invalid(["roll"])))
        #expect(typedHere.drafts["roll"] == "13")
        #expect(typedHere.editingFieldID == "roll")
        #expect(typedHere.fieldErrorID == "roll")

        let typedElsewhere = MetadataBoardModel(host: makeHost())
        typedElsewhere.beginEdit(fieldID: "url")
        typedElsewhere.drafts["url"] = "https://example.org/typed"
        typedElsewhere.setTyped(true, fieldID: "url")
        #expect(!typedElsewhere.showRefusal(invalid(["roll"])))
        #expect(typedElsewhere.drafts["url"] == "https://example.org/typed")
        #expect(typedElsewhere.editingFieldID == "url")
        #expect(typedElsewhere.fieldError == nil && typedElsewhere.fieldErrorID == nil)

        let fresh = MetadataBoardModel(host: makeHost())
        for refused in [invalid([]), invalid([""]), invalid(["gone"]), notMetadata] {
            #expect(!fresh.showRefusal(refused))
        }
        #expect(fresh.editingFieldID == nil && fresh.fieldError == nil && fresh.fieldErrorID == nil)
    }
}
