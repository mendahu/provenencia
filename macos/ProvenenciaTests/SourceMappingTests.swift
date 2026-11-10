import Foundation
import Testing
@testable import Provenencia

@Suite
struct SourceMappingTests {
    @Test func mapsDateSortKey() {
        var proto = Provenencia_Engine_V1_Source()
        proto.dateSortKey = 42
        #expect(GoStore.mapSource(proto).dateSortKey == 42)
        #expect(GoStore.mapSource(Provenencia_Engine_V1_Source()).dateSortKey == nil)
    }
}
