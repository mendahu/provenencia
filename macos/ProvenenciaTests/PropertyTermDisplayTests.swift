import Foundation
import Testing
@testable import Provenencia

@Suite
struct PropertyTermDisplayTests {
    @Test func productTermUsesL10n() {
        #expect(
            PropertyTermDisplay.name(key: "father", origin: CatalogOrigin.provenencia, propertyKey: "role", catalogLabel: "Dad")
                == L10n.string(L10n.PropertyTerm.roleFather)
        )
    }

    @Test func aUserTermKeyedLikeAProductTermKeepsItsOwnLabel() {
        #expect(
            PropertyTermDisplay.name(key: "subject", origin: "user", propertyKey: "role", catalogLabel: "Main person")
                == "Main person"
        )
        #expect(
            PropertyTermDisplay.name(key: "subject", origin: CatalogOrigin.provenencia, propertyKey: "role", catalogLabel: "Main person")
                == L10n.string(L10n.PropertyTerm.roleSubject)
        )
    }

    @Test func unknownTermFallsBackToCatalogLabel() {
        #expect(
            PropertyTermDisplay.name(key: "custom", origin: "user", propertyKey: "role", catalogLabel: "Godparent")
                == "Godparent"
        )
    }

    @Test func nameOfAValue() {
        func term(_ key: String, _ label: String, _ origin: String) -> CatalogConclusionValue {
            .term(id: "t", key: key, label: label, origin: origin)
        }
        #expect(PropertyTermDisplay.name(term("witness", "Witness at the ceremony", CatalogOrigin.provenencia), propertyKey: SeededPropertyKey.role) == "Witness")
        #expect(PropertyTermDisplay.name(term("godparent", "Godparent", "user"), propertyKey: SeededPropertyKey.role) == "Godparent")
        #expect(PropertyTermDisplay.name(term("godparent", "", "user"), propertyKey: SeededPropertyKey.role) == "godparent")
        #expect(PropertyTermDisplay.name(term("woman", "", CatalogOrigin.provenencia), propertyKey: SeededPropertyKey.gender) == "Woman")
        #expect(PropertyTermDisplay.name(term("", "", "user"), propertyKey: SeededPropertyKey.role) == nil)
        #expect(PropertyTermDisplay.name(.text("x"), propertyKey: SeededPropertyKey.role) == nil)
    }

    @Test func emptyCatalogLabelFallsBackToKey() {
        #expect(
            PropertyTermDisplay.name(key: "custom", origin: "user", propertyKey: "role", catalogLabel: "  ")
                == "custom"
        )
    }
}
