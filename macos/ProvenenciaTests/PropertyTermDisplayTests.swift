import Foundation
import Testing
@testable import Provenencia

@Suite
struct PropertyTermDisplayTests {
    @Test func productTermUsesL10n() {
        #expect(
            PropertyTermDisplay.name(key: "father", origin: CatalogOrigin.provenencia, propertyKey: "role", catalogLabel: "Dad")
                == L10n.string(L10n.PropertyTerm.roleFather.label)
        )
    }

    @Test func aUserTermKeyedLikeAProductTermKeepsItsOwnLabel() {
        #expect(
            PropertyTermDisplay.name(key: "subject", origin: "user", propertyKey: "role", catalogLabel: "Main person")
                == "Main person"
        )
        #expect(
            PropertyTermDisplay.name(key: "subject", origin: CatalogOrigin.provenencia, propertyKey: "role", catalogLabel: "Main person")
                == L10n.string(L10n.PropertyTerm.roleSubject.label)
        )
    }

    @Test func unknownTermFallsBackToCatalogLabel() {
        #expect(
            PropertyTermDisplay.name(key: "custom", origin: "user", propertyKey: "role", catalogLabel: "Godparent")
                == "Godparent"
        )
    }

    @Test func emptyCatalogLabelFallsBackToKey() {
        #expect(
            PropertyTermDisplay.name(key: "custom", origin: "user", propertyKey: "role", catalogLabel: "  ")
                == "custom"
        )
    }
}
