import CoreGraphics
import Foundation

/// Layered citation locator: always an `artifact` floor, optional page / region,
/// plus unknown extra selectors preserved losslessly on edit.
struct CitationLocatorDraft: Equatable, Sendable {
    var page: Int?
    var region: ArtifactRegionDraft?
    /// Raw JSON objects for selector types this client does not model (e.g. text_quote).
    var extras: [String]

    static func artifactOnly() -> CitationLocatorDraft {
        CitationLocatorDraft(page: nil, region: nil, extras: [])
    }

    var isArtifactOnly: Bool {
        page == nil && region == nil
    }

    var hasPage: Bool { page != nil }
    var hasRegion: Bool { region != nil }

    mutating func setPage(_ value: Int, capabilities: ArtifactLocatorCapabilities) -> Bool {
        guard capabilities.supportsPageLocator, value >= 1 else { return false }
        page = value
        return true
    }

    mutating func setRegion(
        _ draft: ArtifactRegionDraft,
        capabilities: ArtifactLocatorCapabilities,
        viewerPage: Int?
    ) -> Bool {
        guard capabilities.supportsRegionLocator, draft.isValid else { return false }
        // A region is ink on the current viewer page. Restamp even when a
        // leftover page locator exists (clear-region-then-redraw on another page).
        if capabilities.supportsPageLocator {
            guard let viewerPage, viewerPage >= 1 else { return false }
            page = viewerPage
        }
        region = draft
        return true
    }

    mutating func clearRegion() {
        region = nil
    }

    mutating func removePage() {
        page = nil
        region = nil
    }

    mutating func resetToEntireArtifact() {
        page = nil
        region = nil
    }

    mutating func peelIllegalLayers(capabilities: ArtifactLocatorCapabilities) {
        if !capabilities.supportsPageLocator {
            page = nil
        }
        if !capabilities.supportsRegionLocator {
            region = nil
        }
    }

    func encodeJSON() -> String {
        var selectors: [[String: Any]] = [["type": "artifact"]]
        if let page, page >= 1 {
            selectors.append(["type": "page", "artifact_page": page])
        }
        if let region {
            let points = region.points.map { ["x": $0.x, "y": $0.y] }
            selectors.append([
                "type": "region",
                "unit": "normalized",
                "points": points,
            ])
        }
        for extra in extras {
            if let data = extra.data(using: .utf8),
               let object = try? JSONSerialization.jsonObject(with: data) as? [String: Any]
            {
                selectors.append(object)
            }
        }
        let doc: [String: Any] = ["version": 1, "selectors": selectors]
        guard let data = try? JSONSerialization.data(withJSONObject: doc, options: []),
              let json = String(data: data, encoding: .utf8)
        else {
            return #"{"version":1,"selectors":[{"type":"artifact"}]}"#
        }
        return json
    }

    static func decode(_ json: String) -> CitationLocatorDraft {
        guard let data = json.data(using: .utf8),
              let object = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
              let selectors = object["selectors"] as? [[String: Any]]
        else {
            return .artifactOnly()
        }

        var draft = CitationLocatorDraft.artifactOnly()
        for selector in selectors {
            let type = (selector["type"] as? String ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
            switch type {
            case "artifact":
                continue
            case "page":
                if let page = intValue(selector["artifact_page"]), page >= 1 {
                    draft.page = page
                }
            case "region":
                if let region = decodeRegion(selector) {
                    draft.region = region
                }
            default:
                if type.isEmpty { continue }
                if let extra = encodeObject(selector) {
                    draft.extras.append(extra)
                }
            }
        }
        return draft
    }

    private static func decodeRegion(_ selector: [String: Any]) -> ArtifactRegionDraft? {
        let unit = (selector["unit"] as? String ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        guard unit == "normalized" else { return nil }
        guard let rawPoints = selector["points"] as? [[String: Any]] else { return nil }
        let points = rawPoints.compactMap { item -> CGPoint? in
            guard let x = doubleValue(item["x"]), let y = doubleValue(item["y"]) else { return nil }
            return ArtifactRegionGeometry.clampNormalized(CGPoint(x: x, y: y))
        }
        guard ArtifactRegionGeometry.isValidPolygon(points) else { return nil }
        return ArtifactRegionDraft(kind: ArtifactRegionKind.infer(fromNormalized: points), points: points)
    }

    private static func encodeObject(_ object: [String: Any]) -> String? {
        guard JSONSerialization.isValidJSONObject(object),
              let data = try? JSONSerialization.data(withJSONObject: object, options: []),
              let json = String(data: data, encoding: .utf8)
        else { return nil }
        return json
    }

    private static func intValue(_ value: Any?) -> Int? {
        if let int = value as? Int { return int }
        if let number = value as? NSNumber { return number.intValue }
        return nil
    }

    private static func doubleValue(_ value: Any?) -> Double? {
        if let double = value as? Double { return double }
        if let int = value as? Int { return Double(int) }
        if let number = value as? NSNumber { return number.doubleValue }
        return nil
    }
}
