import Foundation

/// A row in one of the project's catalog vocabularies (`source_types`,
/// `source_metadata_fields`). The two vocabulary destinations browse and
/// count their rows identically; this is the shape that sameness hangs off.
/// Find lives in the workspace omnibar (S3-10).
protocol CatalogVocabularyRow: Identifiable where ID == String {
    var id: String { get }
    var key: String { get }
    var origin: String { get }
    var label: String { get }
    var description: String { get }
    /// How many sources refer to this row — the delete gate.
    var usedBy: Int { get }
}

extension CatalogSourceType: CatalogVocabularyRow {}
extension CatalogMetadataField: CatalogVocabularyRow {}
extension CatalogProperty: CatalogVocabularyRow {}

extension Array where Element: CatalogVocabularyRow {
    // The header's count line splits the vocabulary by origin.

    var seededCount: Int { filter { $0.origin == CatalogOrigin.provenencia }.count }
    var userCount: Int { filter { $0.origin == CatalogOrigin.user }.count }
    var pluginCount: Int { filter { CatalogOrigin.isPlugin($0.origin) }.count }
}

/// Detail-pane mode shared by the vocabulary destinations: illegal
/// combinations of "adding" plus "a row is selected" are unrepresentable.
/// The models keep their `draft` set for `.adding` / `.editing` (form
/// bindings) and also for `.viewing`, so tearing down `Binding($model.draft)`
/// projections when leaving the form does not trap.
enum VocabularyPaneMode: Equatable {
    case empty
    case viewing(id: String)
    case adding(resumeID: String?)
    case editing(id: String)
}

/// A vocabulary destination's toast. Add / update read as success; detaching
/// a suggestion is not a win, just a change, so it carries `.info` — the
/// tone is the model's call, never hard-coded at the view.
struct VocabularyToast: Equatable, Hashable {
    var title: String
    var body: String
    var tone: PVToastTone
}
