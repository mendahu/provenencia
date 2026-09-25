import CoreGraphics
import Foundation

/// On-device text recognition. Callers supply a `CGImage`; this type does not
/// know about citations, locators, or media kinds.
protocol OCREngine: Sendable {
    func recognizeText(in image: CGImage) async throws -> String
}
