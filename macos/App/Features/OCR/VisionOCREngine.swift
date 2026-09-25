import CoreGraphics
import Foundation
import Vision

/// Vision adapter. Tests inject a fake ``OCREngine`` and never construct this.
struct VisionOCREngine: OCREngine {
    func recognizeText(in image: CGImage) async throws -> String {
        try await Task.detached {
            let request = VNRecognizeTextRequest()
            request.recognitionLevel = .accurate
            let handler = VNImageRequestHandler(cgImage: image, options: [:])
            try handler.perform([request])
            let lines = (request.results ?? []).compactMap { observation in
                observation.topCandidates(1).first?.string
            }
            return lines.joined(separator: "\n")
        }.value
    }
}
