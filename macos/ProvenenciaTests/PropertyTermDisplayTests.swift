import Foundation
import Testing
@testable import Provenencia

@Suite
struct PropertyTermDisplayTests {
    @Test func productTermUsesL10n() {
        #expect(
            PropertyTermDisplay.name(key: "father", propertyKey: "role", catalogLabel: "Dad")
                == String(localized: L10n.PropertyTerm.roleFather)
        )
    }

    @Test func unknownTermFallsBackToCatalogLabel() {
        #expect(
            PropertyTermDisplay.name(key: "custom", propertyKey: "role", catalogLabel: "Godparent")
                == "Godparent"
        )
    }

    @Test func emptyCatalogLabelFallsBackToKey() {
        #expect(
            PropertyTermDisplay.name(key: "custom", propertyKey: "role", catalogLabel: "  ")
                == "custom"
        )
    }
}
