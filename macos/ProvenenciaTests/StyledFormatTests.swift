import Foundation
import Testing
@testable import Provenencia

/// `StyledFormat`: one translatable format whose arguments carry their own styling.
@Suite
struct StyledFormatTests {

    @Test func aFormatPlacesStyledArgumentsWhereTheTranslationPutsThem() {
        var bold = AttributeContainer()
        bold[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] = .stronglyEmphasized
        let styled = StyledFormat.attributed(format: "%2$@ of %1$@, 100%% sure, %@", [
            .init(text: "A", attributes: bold), .init(text: "B"),
        ])
        #expect(String(styled.characters) == "B of A, 100% sure, A")
        let a = styled.runs.first { String(styled.characters[$0.range]) == "A" }
        #expect(a?[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] == .stronglyEmphasized)
        let b = styled.runs.first { String(styled.characters[$0.range]).hasPrefix("B of") }
        #expect(b?[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] == nil)
    }
}
