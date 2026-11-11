import Foundation

/// A layer a citation reads: the Source (the record), one of its Artifacts (the
/// copy used), one of its Evidence graphs (the item cited), or the Citation.
/// Mirrors `vocabkey.Level` in Go; `TestSwiftVocabularyMatchesGo` fails if the
/// two drift. The first three hold metadata values and type suggestions; the
/// Citation only built-ins.
enum Level: String, CaseIterable, Sendable {
    case source
    case artifact
    case graph
    case citation

    /// The levels that hold metadata; mirrors `vocabkey.MetadataLevels`.
    static let metadataLevels: [Level] = [
        .source,
        .artifact,
        .graph,
    ]

    var holdsMetadata: Bool { Self.metadataLevels.contains(self) }
}
