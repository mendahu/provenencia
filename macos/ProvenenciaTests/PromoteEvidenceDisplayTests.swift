import Foundation
import Testing
@testable import Provenencia

/// The evidence sheet's choices and words, tested without a view.
@Suite
struct PromoteEvidenceDisplayTests {
    @Test func aPartlyCheckedBoxSaysSo() {
        #expect(L10n.string(L10n.DesignSystem.checkboxPartlyChecked) == "partly checked")
    }

    @Test func anOptionalFieldSaysSo() {
        #expect(L10n.string(L10n.DesignSystem.optionalMarker) == "(optional)")
    }
}
