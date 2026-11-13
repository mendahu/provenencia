import SwiftUI
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
        #expect(regular.padding == EdgeInsets(top: 10, leading: 10, bottom: 10, trailing: 10))
        #expect(regular.contentSpacing == 10)

        let compact = PVTileCardSize.compact
        #expect(compact.height == 46)
        #expect(compact.thumbnail == 38)
        #expect(compact.titleLineLimit == 1)
        #expect(compact.linesLineLimit == 1)
        #expect(compact.minimumWidth == 300)
        #expect(compact.padding == EdgeInsets(top: 4, leading: 4, bottom: 4, trailing: 8))
        #expect(compact.contentSpacing == 8)
    }

    /// A compact card's name line, the gap and its life-span line fit the height the card
    /// gives its text, so neither line runs into the padding or the card's edge.
    @Test @MainActor func compactTextFitsItsSlot() {
        let size = PVTileCardSize.compact
        let layout = NSLayoutManager()
        let text = layout.defaultLineHeight(for: PVPersonCard.nsTitleFont(size))
            + size.lineSpacing
            + layout.defaultLineHeight(for: PVPersonCard.nsCompactYearsFont)
        #expect(text <= size.thumbnail)
    }

    /// A compact card's thumbnail fills the card's height inside the padding, with the
    /// same inset above, below and to its left.
    @Test func compactThumbnailFillsItsEdgeEvenly() {
        let size = PVTileCardSize.compact
        #expect(size.thumbnail + size.padding.top + size.padding.bottom == size.height)
        #expect(size.padding.leading == size.padding.top)
        #expect(size.padding.top == size.padding.bottom)
    }
}
