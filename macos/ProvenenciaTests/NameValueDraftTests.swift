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
            .init(id: UUID(), value: "W.", type: "initial"),
            .init(id: UUID(), value: "Alderwick", type: "surname"),
        ]
        #expect(draft.isValid)
        let parts = draft.toInput().parts
        #expect(parts.map(\.type) == ["given", "initial", "surname"])
        #expect(NameValueDisplay.string(for: draft) == "John W. Alderwick")
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
            .init(id: UUID(), value: "W.", type: "initial"),
            .init(id: UUID(), value: "", type: ""),
        ]
        #expect(draft.storedFormLine == "“John W. Alderwick”")
        #expect(draft.storedPartsLine == "1 given:John · 2 initial:W. · 3 \"\":…")
    }
}
