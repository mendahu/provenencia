import AppKit
import Testing
@testable import Provenencia

struct PVSymbolTests {
    /// Every S9-16 mark is a real SF Symbol on the deployment target.
    @Test(arguments: [
        PVSymbol.merge, .split, .stamp, .gitMerge, .scale, .signalLow, .ban, .circleMinus, .minus, .circleDashed,
    ])
    func symbolResolves(_ symbol: PVSymbol) {
        #expect(NSImage(systemSymbolName: symbol.rawValue, accessibilityDescription: nil) != nil, "\(symbol.rawValue)")
    }
}
