import Foundation
@testable import Provenencia

/// An Artifact row for tests. `ref` defaults to `ART-` plus the `id`, and `label`
/// to the `id`; a file-bearing Artifact passes `fileID` and `file` together.
func catalogArtifact(
    _ id: String,
    ref: String? = nil,
    sourceID: String = "s1",
    label: String? = nil,
    description: String = "",
    fileID: String = "",
    file: CatalogFileRef? = nil
) -> CatalogArtifact {
    CatalogArtifact(
        id: id, ref: ref ?? "ART-\(id)", sourceID: sourceID, fileID: fileID,
        label: label ?? id, description: description, file: file
    )
}

/// A file-bearing image Artifact for tests: ref `ART-<ID>` uppercased, file id
/// `file-<id>` unless `fileID` says otherwise, at `objects/<id>.jpg` unless
/// `relPath` does.
func catalogFileArtifact(
    _ id: String,
    ref: String? = nil,
    sourceID: String = "s1",
    label: String? = nil,
    fileID: String? = nil,
    relPath: String? = nil,
    mediaType: String = "image/jpeg"
) -> CatalogArtifact {
    let fileID = fileID ?? "file-\(id)"
    return catalogArtifact(
        id, ref: ref ?? "ART-\(id.uppercased())", sourceID: sourceID, label: label, fileID: fileID,
        file: CatalogFileRef(
            id: fileID, relPath: relPath ?? "objects/\(id).jpg", originalFilename: "\(id).jpg",
            mediaType: mediaType, byteSize: 4_200_000
        )
    )
}

/// A text metadata field for tests; `key` defaults to the lowercased `label`.
func catalogMetadataField(_ id: String, label: String, key: String? = nil) -> CatalogMetadataField {
    CatalogMetadataField(
        id: id, key: key ?? label.lowercased(), origin: "user", label: label, dataType: "text", description: ""
    )
}

/// A metadata row for tests; it holds a value exactly when `value` is not empty.
func catalogMetadataEntry(
    _ field: CatalogMetadataField, _ value: String = "", suggested: Bool = false, order: Int32 = 0
) -> CatalogMetadataEntry {
    CatalogMetadataEntry(
        field: field, valueText: value, hasValue: !value.isEmpty, suggested: suggested, sortOrder: order
    )
}
