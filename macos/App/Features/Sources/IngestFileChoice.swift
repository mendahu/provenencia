import Foundation

/// A file chosen to be ingested into an Artifact: picked, dropped, or refused.
/// The Artifact sheet's File row and the Add file dialog share it, so both
/// validate a file and map the engine's refusal to a Callout the same way.
struct IngestFileChoice: Equatable {
    /// Absolute path, set only when the app accepts the file; nil when refused or cleared.
    var filePath: String?
    var fileName: String?
    var reject: L10n.Errors.IngestCallout?

    /// Takes dropped files: more than one is refused (naming the first), one is chosen.
    mutating func choose(_ urls: [URL]) {
        if urls.count > 1 {
            filePath = nil
            fileName = urls.first?.lastPathComponent
            reject = L10n.Errors.ingestCallout(reason: .multiFile)
            return
        }
        guard let url = urls.first else { return }
        choose(url)
    }

    /// Takes one file: accepted, or refused with the policy's Callout and no engine call.
    mutating func choose(_ url: URL) {
        let name = url.lastPathComponent
        switch IngestMediaPolicy.validate(url: url) {
        case .ok:
            filePath = url.path
            fileName = name
            reject = nil
        case .reject(let reason):
            let size = reason == .tooLarge
                ? IngestMediaPolicy.formatByteSize(IngestMediaPolicy.byteSize(of: url))
                : nil
            filePath = nil
            fileName = name
            reject = L10n.Errors.ingestCallout(reason: reason, sizeLabel: size)
        }
    }

    /// Asks for a file in the open panel and chooses it; nothing changes when the panel is cancelled.
    @MainActor
    mutating func pick() {
        let path = ProjectFiles.pickFileForIngest(
            prompt: L10n.string(L10n.Sources.filePickPrompt),
            message: L10n.string(L10n.Sources.filePickMessage)
        )
        guard let path else { return }
        choose(URL(fileURLWithPath: path))
    }

    /// Keeps the file's name and shows the engine's refusal when `error` is an
    /// ingest refusal; true when it was one, else nothing changes.
    mutating func refuse(ifIngestRefusal error: Error) -> Bool {
        guard let coded = error as? CoreInvokeError, let code = coded.code,
              let callout = L10n.Errors.ingestCallout(code: code, params: coded.params)
        else { return false }
        filePath = nil
        reject = callout
        return true
    }

    /// True when the file was refused and no path is left to ingest: Save waits for another file.
    var isRefused: Bool { reject != nil && filePath == nil }

    /// The Artifact was made but the file did not attach: keeps the path and the name, so a
    /// retry ingests it again, and shows the engine's message under a title that says so.
    mutating func failAfterCreate(message: String) {
        reject = L10n.Errors.IngestCallout(
            title: L10n.Sources.artifactFileNotAttachedTitle,
            message: L10n.Sources.artifactFileNotAttachedBody(message)
        )
    }

    mutating func clear() {
        filePath = nil
        fileName = nil
        reject = nil
    }
}
