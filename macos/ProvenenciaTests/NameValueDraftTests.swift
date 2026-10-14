import Foundation
import Testing
@testable import Provenencia

@Suite
struct NameValueDraftTests {
    @Test func emptyIsInvalid() {
        #expect(!NameValueDraft.empty().isValid)
    }

    @Test func formOnlyIsValid() {
        var draft = NameValueDraft.empty()
        draft.form = "Tatanka Iyotake"
        #expect(draft.isValid)
        #expect(draft.toInput().form == "Tatanka Iyotake")
        #expect(draft.toInput().parts.isEmpty)
    }

    @Test func formIsTrimmed() {
        var draft = NameValueDraft.empty()
        draft.form = "  John W. Alderwick  "
        #expect(draft.isValid)
        #expect(draft.toInput().form == "John W. Alderwick")
    }

    @Test func emptyPartValueIsInvalid() {
        var draft = NameValueDraft.empty()
        draft.form = "Ada"
        draft.parts = [.init(id: UUID(), value: "", type: "given")]
        #expect(!draft.isValid)
        #expect(draft.partValueError(at: 0) != nil)
    }

    @Test func unknownPartTypeIsInvalid() {
        var draft = NameValueDraft.empty()
        draft.form = "Ada Lovelace"
        draft.parts = [.init(id: UUID(), value: "Ada", type: "not-a-type")]
        #expect(!draft.isValid)
    }

    @Test func untypedPartIsValid() {
        var draft = NameValueDraft.empty()
        draft.form = "Guðrún Jónsdóttir"
        draft.parts = [
            .init(id: UUID(), value: "Guðrún", type: "given"),
            .init(id: UUID(), value: "Jónsdóttir", type: ""),
        ]
        #expect(draft.isValid)
        let parts = draft.toInput().parts
        #expect(parts.count == 2)
        #expect(parts[1].type.isEmpty)
    }

    @Test func typedPartsRoundTrip() {
        var draft = NameValueDraft.empty()
        draft.form = "John W. Alderwick"
        draft.parts = [
            .init(id: UUID(), value: "John", type: "given"),
            .init(id: UUID(), value: "W.", type: "given"),
            .init(id: UUID(), value: "Alderwick", type: "surname"),
        ]
        #expect(draft.isValid)
        let parts = draft.toInput().parts
        #expect(parts.map(\.type) == ["given", "given", "surname"])
        #expect(NameValueDisplay.string(for: draft) == "John W. Alderwick")
    }

    @Test func addPartAppendsEmptyRow() {
        var draft = NameValueDraft.empty()
        let id = draft.addPart()
        #expect(draft.parts.count == 1)
        #expect(draft.parts[0].id == id)
        #expect(draft.parts[0].value.isEmpty)
        #expect(draft.parts[0].type.isEmpty)
    }

    @Test func movePartSwapsNeighbors() {
        var draft = NameValueDraft.empty()
        draft.form = "Wang Xiuying"
        let surname = NameValueDraft.Part(id: UUID(), value: "Wang", type: "surname")
        let given = NameValueDraft.Part(id: UUID(), value: "Xiuying", type: "given")
        draft.parts = [given, surname]
        draft.movePart(at: 1, by: -1)
        #expect(draft.parts.map(\.type) == ["surname", "given"])
    }

    @Test func formErrorRequiresTouch() {
        var draft = NameValueDraft.empty()
        #expect(draft.formError == nil)
        draft.formTouched = true
        #expect(draft.formError != nil)
    }

    @Test func registryRejectsUnknownKeys() {
        #expect(NamePartType.isAllowed(""))
        #expect(NamePartType.isAllowed("given"))
        #expect(NamePartType.isAllowed("surname_prefix"))
        #expect(!NamePartType.isAllowed("maiden"))
    }

    @Test func storedAsMatchesBoard() {
        var draft = NameValueDraft.empty()
        #expect(draft.storedFormLine == "—")
        #expect(draft.storedPartsLine == "[ ]")

        draft.form = "John W. Alderwick"
        draft.parts = [
            .init(id: UUID(), value: "John", type: "given"),
            .init(id: UUID(), value: "W.", type: "given"),
            .init(id: UUID(), value: "", type: ""),
        ]
        #expect(draft.storedFormLine == "“John W. Alderwick”")
        #expect(draft.storedPartsLine == "1 given:John · 2 given:W. · 3 \"\":…")
    }
}

@Suite
struct NameFormSplitTests {
    @Test func commaPutsTheSurnameFirst() {
        #expect(split("Smith, John Henry") == [
            Piece(value: "Smith", type: "surname"),
            Piece(value: "John", type: "given"),
            Piece(value: "Henry", type: "given"),
        ])
        #expect(split("De la Cruz, Maria Elena") == [
            Piece(value: "De la Cruz", type: "surname"),
            Piece(value: "Maria", type: "given"),
            Piece(value: "Elena", type: "given"),
        ])
        #expect(split("Smith,") == [Piece(value: "Smith", type: "surname")])
        #expect(split("Smith Jr, John") == [
            Piece(value: "Smith Jr", type: "surname"),
            Piece(value: "John", type: "given"),
        ])
        #expect(split("Smith, John Jr") == [
            Piece(value: "Smith", type: "surname"),
            Piece(value: "John", type: "given"),
            Piece(value: "Jr", type: "suffix"),
        ])
        #expect(split("Smith, John, Jr.") == [
            Piece(value: "Smith", type: "surname"),
            Piece(value: "John", type: "given"),
            Piece(value: "Jr.", type: "suffix"),
        ])
        #expect(split("John Smith,") == [Piece(value: "John Smith", type: "surname")])
    }

    @Test func lastWordIsTheSurnameWhenSeveralRemain() {
        #expect(split("John Henry Smith") == [
            Piece(value: "John", type: "given"),
            Piece(value: "Henry", type: "given"),
            Piece(value: "Smith", type: "surname"),
        ])
        #expect(split("  John   Henry  Smith  ") == [
            Piece(value: "John", type: "given"),
            Piece(value: "Henry", type: "given"),
            Piece(value: "Smith", type: "surname"),
        ])
    }

    @Test func oneNameWordIsGiven() {
        #expect(split("John") == [Piece(value: "John", type: "given")])
        #expect(split("Dr.") == [Piece(value: "Dr.", type: "given")])
        #expect(split("Jr") == [Piece(value: "Jr", type: "given")])
        #expect(split("   ").isEmpty)
        #expect(split("").isEmpty)
    }

    @Test func titlesPeelOnlyWhenAnotherWordRemains() {
        #expect(split("Dr. John Smith Jr.") == [
            Piece(value: "Dr.", type: "prefix"),
            Piece(value: "John", type: "given"),
            Piece(value: "Smith", type: "surname"),
            Piece(value: "Jr.", type: "suffix"),
        ])
        #expect(split("Dr. John") == [
            Piece(value: "Dr.", type: "prefix"),
            Piece(value: "John", type: "given"),
        ])
        #expect(split("John Jr") == [
            Piece(value: "John", type: "given"),
            Piece(value: "Jr", type: "suffix"),
        ])
        #expect(split("Rev. Dr. John Smith Jr. III") == [
            Piece(value: "Rev.", type: "prefix"),
            Piece(value: "Dr.", type: "prefix"),
            Piece(value: "John", type: "given"),
            Piece(value: "Smith", type: "surname"),
            Piece(value: "Jr.", type: "suffix"),
            Piece(value: "III", type: "suffix"),
        ])
        #expect(split("Sr. Mary") == [
            Piece(value: "Sr.", type: "prefix"),
            Piece(value: "Mary", type: "given"),
        ])
        #expect(split("John Sr") == [
            Piece(value: "John", type: "given"),
            Piece(value: "Sr", type: "suffix"),
        ])
        #expect(split("van der Berg") == [
            Piece(value: "van", type: "given"),
            Piece(value: "der", type: "given"),
            Piece(value: "Berg", type: "surname"),
        ])
    }

    @Test func registryIgnoresCaseAndATrailingPeriod() {
        #expect(NameTitleRegistry.isPrefix("Dr."))
        #expect(NameTitleRegistry.isPrefix("dr"))
        #expect(NameTitleRegistry.isSuffix("JR."))
        #expect(!NameTitleRegistry.isPrefix("van"))
        #expect(!NameTitleRegistry.isSuffix("Smith"))
    }

    @Test func blurFillsOnlyAnEmptyList() {
        var draft = NameValueDraft.empty()
        draft.form = "John Henry Smith"
        draft.fillEmptyPartsFromForm()
        #expect(draft.parts.map(\.type) == ["given", "given", "surname"])
        #expect(!draft.partsHandEdited)

        draft.setPartValue(id: draft.parts[0].id, value: "Jon")
        #expect(draft.partsHandEdited)
        draft.form = "Ada Lovelace"
        draft.fillEmptyPartsFromForm()
        #expect(draft.parts[0].value == "Jon")
    }

    @Test func splitRefillsUntilAPartIsHandEdited() {
        var draft = NameValueDraft.empty()
        draft.form = "John Smith"
        draft.fillEmptyPartsFromForm()
        draft.form = "John Henry Smith"
        let refilled = draft.splitFromForm()
        #expect(refilled == .applied)
        #expect(draft.parts.map(\.value) == ["John", "Henry", "Smith"])
        #expect(!draft.partsHandEdited)

        draft.setPartValue(id: draft.parts[0].id, value: "Jon")
        let asked = draft.splitFromForm()
        guard case .needsConfirm(let pending) = asked else {
            Issue.record("Split should ask before replacing a hand-edited part")
            return
        }
        #expect(pending.map(\.value) == ["John", "Henry", "Smith"])
        #expect(draft.parts[0].value == "Jon")
        draft.applySplit(pending)
        #expect(draft.parts.map(\.value) == ["John", "Henry", "Smith"])
        #expect(!draft.partsHandEdited)
    }

    @Test func loadedPartsAskBeforeSplit() {
        var draft = NameValueDraft(
            form: "John Smith",
            parts: [
                CatalogNameValuePart(value: "John", type: "given"),
                CatalogNameValuePart(value: "Smith", type: "surname"),
            ]
        )
        #expect(draft.partsHandEdited)
        guard case .needsConfirm = draft.splitFromForm() else {
            Issue.record("Parts loaded from the catalog count as hand-edited")
            return
        }
        #expect(draft.parts.map(\.value) == ["John", "Smith"])
    }

    @Test func addingAPartCountsAsHandEdited() {
        var draft = NameValueDraft.empty()
        draft.form = "John Smith"
        draft.fillEmptyPartsFromForm()
        _ = draft.addPart()
        guard case .needsConfirm = draft.splitFromForm() else {
            Issue.record("Add part should make the next Split ask")
            return
        }
    }

    @Test func removingAPartCountsAsHandEdited() {
        var draft = splitFilled("John Smith")
        draft.removePart(id: draft.parts[0].id)
        #expect(draft.parts.map(\.value) == ["Smith"])
        guard case .needsConfirm(let pending) = draft.splitFromForm() else {
            Issue.record("Remove part should make the next Split ask")
            return
        }
        #expect(draft.parts.map(\.value) == ["Smith"])
        #expect(pending.map(\.value) == ["John", "Smith"])
        draft.applySplit(pending)
        #expect(draft.parts.map(\.value) == ["John", "Smith"])
        #expect(!draft.partsHandEdited)
    }

    @Test func reorderingAPartCountsAsHandEdited() {
        var draft = splitFilled("John Smith")
        draft.movePart(at: 0, by: 1)
        #expect(draft.parts.map(\.value) == ["Smith", "John"])
        guard case .needsConfirm(let pending) = draft.splitFromForm() else {
            Issue.record("Reorder should make the next Split ask")
            return
        }
        #expect(draft.parts.map(\.value) == ["Smith", "John"])
        draft.applySplit(pending)
        #expect(draft.parts.map(\.value) == ["John", "Smith"])
        #expect(!draft.partsHandEdited)
    }

    @Test func changingAPartTypeCountsAsHandEdited() {
        var draft = splitFilled("John Smith")
        draft.setPartType(id: draft.parts[0].id, type: "nick")
        #expect(draft.parts[0].type == "nick")
        guard case .needsConfirm(let pending) = draft.splitFromForm() else {
            Issue.record("Type change should make the next Split ask")
            return
        }
        #expect(draft.parts[0].type == "nick")
        #expect(pending.map(\.type) == ["given", "surname"])
        draft.applySplit(pending)
        #expect(draft.parts.map(\.type) == ["given", "surname"])
        #expect(!draft.partsHandEdited)
    }

    private func splitFilled(_ form: String) -> NameValueDraft {
        var draft = NameValueDraft.empty()
        draft.form = form
        draft.fillEmptyPartsFromForm()
        return draft
    }

    private struct Piece: Equatable {
        var value: String
        var type: String
    }

    private func split(_ form: String) -> [Piece] {
        NameFormSplit.parts(from: form).map { Piece(value: $0.value, type: $0.type) }
    }
}
