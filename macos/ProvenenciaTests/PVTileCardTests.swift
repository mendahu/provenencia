import Testing
@testable import Provenencia

struct PVTileCardTests {
    @Test func sizes() {
        let regular = PVTileCardSize.regular
        #expect(regular.height == 92)
        #expect(regular.thumbnail == 70)
        #expect(regular.titleLineLimit == 2)
        #expect(regular.linesLineLimit == 2)
        #expect(regular.minimumWidth == 120)
        #expect(regular.padding == 10)

        let compact = PVTileCardSize.compact
        #expect(compact.height == 52)
        #expect(compact.thumbnail == 40)
        #expect(compact.titleLineLimit == 1)
        #expect(compact.linesLineLimit == 1)
        #expect(compact.minimumWidth == 232)
        #expect(compact.padding == 6)
    }
}
