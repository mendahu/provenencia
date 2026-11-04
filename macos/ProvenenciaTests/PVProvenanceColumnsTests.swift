import Testing
@testable import Provenencia

struct PVProvenanceColumnsTests {
    @Test func columnsFollowTheWidthClass() {
        let regular = PVProvenanceColumns()
        #expect(regular.readAsWidth == 140 && regular.outcomeWidth == 190 && regular.panelWidth == 880)
        let wide = PVProvenanceColumns(widthClass: .wide)
        #expect(wide.readAsWidth == 220 && wide.outcomeWidth == 280 && wide.panelWidth == 1280)
        #expect(PVProvenanceColumns.recordWidth == 112 && PVProvenanceColumns.actionWidth == 88)
        #expect(regular.showsReadAs && !PVProvenanceColumns(showsReadAs: false).showsReadAs)
    }
}
