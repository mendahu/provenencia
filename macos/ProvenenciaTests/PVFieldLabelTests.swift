import Foundation
import Testing
@testable import Provenencia

@Suite
struct PVFieldLabelTests {
    @Test func fieldLabelFillsIn() {
        #expect(PVFieldLabel.resolve(own: nil, field: "Name") == "Name")
    }

    @Test func ownLabelWins() {
        #expect(PVFieldLabel.resolve(own: "Filter", field: "Name") == "Filter")
    }

    @Test func noLabelStaysNil() {
        #expect(PVFieldLabel.resolve(own: nil, field: nil) == nil)
    }
}
