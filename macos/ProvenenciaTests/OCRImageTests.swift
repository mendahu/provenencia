import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

struct OCRImageTests {
    @Test func isOversizedUsesLongestEdgeAndPixelCount() {
        #expect(!OCRImage.isOversized(CGSize(width: 800, height: 600)))
        #expect(OCRImage.isOversized(CGSize(width: 4000, height: 10)))
        #expect(OCRImage.isOversized(CGSize(width: 10, height: 4000)))
        #expect(OCRImage.isOversized(CGSize(width: 6200, height: 8400)))
        #expect(!OCRImage.isOversized(.zero))
    }

    @Test func flowingTranscriptionReplacesLineBreaksWithSpaces() {
        #expect(OCRTranscription.flowing(["Ada", "Lovelace"]) == "Ada Lovelace")
        #expect(OCRTranscription.flowing(["Ada\n", "", "Lovelace"]) == "Ada Lovelace")
        #expect(OCRTranscription.flowing(["Wm.\r\nRobins"]) == "Wm. Robins")
    }
}
