import Foundation

/// App-local A↔bridge↔B association until Citations/Observations exist (S6-04).
struct EvidenceProvisionalLink: Codable, Equatable, Sendable, Identifiable {
    var id: String { bridgeSubjectID }
    var bridgeSubjectID: String
    var endpointAID: String
    var endpointBID: String
}

struct EvidenceProvisionalLinkDocument: Codable, Equatable, Sendable {
    var v: Int
    var projectKey: String
    /// sourceID → links
    var linksBySource: [String: [EvidenceProvisionalLink]]

    static let formatVersion = 1
}

/// Protocol so tests can inject an in-memory double.
@MainActor
protocol EvidenceProvisionalLinkStoring: AnyObject {
    func links(for sourceID: String) -> [EvidenceProvisionalLink]
    func upsert(_ link: EvidenceProvisionalLink, sourceID: String)
    func remove(bridgeSubjectID: String, sourceID: String)
}

/// JSON under Application Support — same spirit as navigation history.
@MainActor
final class EvidenceProvisionalLinkStore: EvidenceProvisionalLinkStoring {
    private(set) var document: EvidenceProvisionalLinkDocument
    private let fileURL: URL
    private let fileManager: FileManager

    init(
        projectKey: String,
        fileURL: URL,
        fileManager: FileManager = .default
    ) {
        self.fileURL = fileURL
        self.fileManager = fileManager
        switch Self.load(from: fileURL, fileManager: fileManager) {
        case .missing:
            self.document = EvidenceProvisionalLinkDocument(
                v: EvidenceProvisionalLinkDocument.formatVersion,
                projectKey: projectKey,
                linksBySource: [:]
            )
            persist()
        case .decoded(let loaded) where loaded.projectKey == projectKey:
            var doc = loaded
            doc.v = EvidenceProvisionalLinkDocument.formatVersion
            self.document = doc
        case .decoded, .unreadable:
            self.document = EvidenceProvisionalLinkDocument(
                v: EvidenceProvisionalLinkDocument.formatVersion,
                projectKey: projectKey,
                linksBySource: [:]
            )
            persist()
        }
    }

    /// Convenience for the live app: file under Application Support.
    convenience init(projectDir: String, fileManager: FileManager = .default) throws {
        let url = try InstallPaths.evidenceLinksFile(
            projectDir: projectDir,
            fileManager: fileManager
        )
        self.init(projectKey: projectDir, fileURL: url, fileManager: fileManager)
    }

    func links(for sourceID: String) -> [EvidenceProvisionalLink] {
        document.linksBySource[sourceID] ?? []
    }

    func upsert(_ link: EvidenceProvisionalLink, sourceID: String) {
        var list = document.linksBySource[sourceID] ?? []
        if let index = list.firstIndex(where: { $0.bridgeSubjectID == link.bridgeSubjectID }) {
            list[index] = link
        } else {
            list.append(link)
        }
        document.linksBySource[sourceID] = list
        persist()
    }

    func remove(bridgeSubjectID: String, sourceID: String) {
        guard var list = document.linksBySource[sourceID] else { return }
        list.removeAll { $0.bridgeSubjectID == bridgeSubjectID }
        document.linksBySource[sourceID] = list
        persist()
    }

    private enum LoadOutcome {
        case missing
        case decoded(EvidenceProvisionalLinkDocument)
        case unreadable
    }

    private static func load(from url: URL, fileManager: FileManager) -> LoadOutcome {
        guard fileManager.fileExists(atPath: url.path) else { return .missing }
        guard let data = try? Data(contentsOf: url) else { return .unreadable }
        guard let doc = try? JSONDecoder().decode(EvidenceProvisionalLinkDocument.self, from: data)
        else { return .unreadable }
        return .decoded(doc)
    }

    private func persist() {
        do {
            try fileManager.createDirectory(
                at: fileURL.deletingLastPathComponent(),
                withIntermediateDirectories: true
            )
            let data = try JSONEncoder().encode(document)
            try data.write(to: fileURL, options: .atomic)
        } catch {
            // Non-blocking — canvas still works; links may not survive relaunch.
        }
    }
}

/// In-memory store for unit tests.
@MainActor
final class InMemoryEvidenceProvisionalLinkStore: EvidenceProvisionalLinkStoring {
    private var linksBySource: [String: [EvidenceProvisionalLink]] = [:]

    func links(for sourceID: String) -> [EvidenceProvisionalLink] {
        linksBySource[sourceID] ?? []
    }

    func upsert(_ link: EvidenceProvisionalLink, sourceID: String) {
        var list = linksBySource[sourceID] ?? []
        if let index = list.firstIndex(where: { $0.bridgeSubjectID == link.bridgeSubjectID }) {
            list[index] = link
        } else {
            list.append(link)
        }
        linksBySource[sourceID] = list
    }

    func remove(bridgeSubjectID: String, sourceID: String) {
        linksBySource[sourceID]?.removeAll { $0.bridgeSubjectID == bridgeSubjectID }
    }
}
