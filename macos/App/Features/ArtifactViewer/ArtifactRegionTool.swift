import CoreGraphics
import Foundation

/// Seven region radios on the Artifact viewer strip (S7-D9 / S7-07).
enum ArtifactRegionTool: String, Equatable, Sendable, CaseIterable {
    case rectangle
    case lOpenTopRight
    case lOpenTopLeft
    case lOpenBottomRight
    case lOpenBottomLeft
    case circle
    case freeform

    var symbol: PVSymbol {
        switch self {
        case .rectangle: .regionRectangle
        case .lOpenTopRight: .regionLTopRight
        case .lOpenTopLeft: .regionLTopLeft
        case .lOpenBottomRight: .regionLBottomRight
        case .lOpenBottomLeft: .regionLBottomLeft
        case .circle: .regionCircle
        case .freeform: .regionFreeform
        }
    }

    var regionKind: ArtifactRegionKind {
        switch self {
        case .rectangle: .rectangle
        case .lOpenTopRight: .lOpenTopRight
        case .lOpenTopLeft: .lOpenTopLeft
        case .lOpenBottomRight: .lOpenBottomRight
        case .lOpenBottomLeft: .lOpenBottomLeft
        case .circle: .circle
        case .freeform: .freeform
        }
    }
}

/// Persisted region noun + normalized polygon (unit square, y down).
enum ArtifactRegionKind: String, Equatable, Sendable {
    case rectangle
    case lOpenTopRight
    case lOpenTopLeft
    case lOpenBottomRight
    case lOpenBottomLeft
    case circle
    case freeform

    var isLShape: Bool {
        switch self {
        case .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft:
            return true
        default:
            return false
        }
    }

    static func infer(fromNormalized points: [CGPoint]) -> ArtifactRegionKind {
        if points.count == 4, ArtifactRegionGeometry.isAxisAlignedRectangle(points) {
            return .rectangle
        }
        if points.count == ArtifactRegionGeometry.circleRingCount,
           ArtifactRegionGeometry.looksLikeCircleRing(points)
        {
            return .circle
        }
        if points.count == 6, let open = ArtifactRegionGeometry.inferredLOpen(from: points) {
            return open
        }
        return .freeform
    }
}

/// One committed region layer (normalized points, no closing duplicate).
struct ArtifactRegionDraft: Equatable, Sendable {
    var kind: ArtifactRegionKind
    var points: [CGPoint]

    var isValid: Bool {
        ArtifactRegionGeometry.isValidPolygon(points)
    }
}
