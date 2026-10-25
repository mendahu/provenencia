import SwiftUI

/// The arithmetic behind ``PVArtifactChecklist``, apart from the view so it
/// is testable. State is the set of unchecked ids (`excluded`), so an Artifact
/// that appears later starts checked. A `locked` id is always checked: it is
/// never counted as unchecked, and the All row never clears it.
enum PVArtifactChecklistState {
    /// The unchecked ids that are still on the list and not locked; a deleted
    /// Artifact's id may linger in `excluded` and must not change the count.
    static func liveExcluded(ids: [String], excluded: Set<String>, locked: Set<String>) -> Set<String> {
        excluded.intersection(ids).subtracting(locked)
    }

    /// The checked ids in list order.
    static func checkedIDs(ids: [String], excluded: Set<String>, locked: Set<String>) -> [String] {
        let live = liveExcluded(ids: ids, excluded: excluded, locked: locked)
        return ids.filter { !live.contains($0) }
    }

    static func checkedCount(ids: [String], excluded: Set<String>, locked: Set<String>) -> Int {
        checkedIDs(ids: ids, excluded: excluded, locked: locked).count
    }

    static func isAllChecked(ids: [String], excluded: Set<String>, locked: Set<String>) -> Bool {
        liveExcluded(ids: ids, excluded: excluded, locked: locked).isEmpty
    }

    /// Some checked, some not: the All row draws a minus.
    static func isAllIndeterminate(ids: [String], excluded: Set<String>, locked: Set<String>) -> Bool {
        !isAllChecked(ids: ids, excluded: excluded, locked: locked)
            && checkedCount(ids: ids, excluded: excluded, locked: locked) > 0
    }

    /// What the All row sets `excluded` to: off clears every id that is not locked.
    static func toggledAll(on: Bool, ids: [String], locked: Set<String>) -> Set<String> {
        on ? [] : Set(ids).subtracting(locked)
    }
}

/// A Source's Artifacts as a checklist: an "All Artifacts" row and one row per
/// Artifact, each with its thumbnail, label and ref. Used by the Copy Source
/// form and the Evidence graph form.
///
/// `locked` maps an Artifact id to the hint shown under its label; that row is
/// checked, shows the lock mark and cannot be cleared, by its own box or by the
/// All row. `description` is a caption line under the heading.
/// `accessibilityIdentifierPrefix` roots the ids: `<prefix>`, `.all`, `.empty`,
/// `.<artifact id>`.
struct PVArtifactChecklist: View {
    let artifacts: [CatalogArtifact]
    @Binding var excluded: Set<String>
    let projectDir: String
    let sourceTypeIconKey: String?
    let isDisabled: Bool
    var locked: [String: String] = [:]
    var description: LocalizedStringResource?
    let accessibilityIdentifierPrefix: String

    private static let thumbnailSize: CGFloat = 32
    private static let maxRowsHeight: CGFloat = 264

    private var ids: [String] { artifacts.map(\.id) }
    private var lockedIDs: Set<String> { Set(locked.keys) }

    var body: some View {
        if artifacts.isEmpty {
            Text(L10n.ArtifactChecklist.empty)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).empty")
        } else {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                header
                VStack(spacing: 0) {
                    allRow
                    ScrollView {
                        LazyVStack(spacing: 0) {
                            ForEach(artifacts, id: \.id) { artifact in
                                PVDivider()
                                row(artifact)
                            }
                        }
                    }
                    .frame(maxHeight: Self.maxRowsHeight)
                }
                .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
                .overlay {
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                }
            }
            .accessibilityIdentifier(accessibilityIdentifierPrefix)
        }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            HStack {
                Text(L10n.ArtifactChecklist.heading)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                Spacer(minLength: 0)
                Text(verbatim: L10n.ArtifactChecklist.count(
                    PVArtifactChecklistState.checkedCount(ids: ids, excluded: excluded, locked: lockedIDs),
                    artifacts.count
                ))
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
            }
            if let description {
                Text(description)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
    }

    private var allRow: some View {
        PVCheckbox(
            L10n.ArtifactChecklist.all,
            isChecked: Binding(
                get: { PVArtifactChecklistState.isAllChecked(ids: ids, excluded: excluded, locked: lockedIDs) },
                set: { on in
                    excluded = PVArtifactChecklistState.toggledAll(on: on, ids: ids, locked: lockedIDs)
                }
            ),
            isIndeterminate: PVArtifactChecklistState.isAllIndeterminate(
                ids: ids, excluded: excluded, locked: lockedIDs
            ),
            isDisabled: isDisabled
        )
        .accessibilityIdentifier("\(accessibilityIdentifierPrefix).all")
        .padding(.vertical, PVSpacing.space2)
        .padding(.trailing, PVSpacing.space5)
        // Row padding + thumbnail + gap, so this box lines up with the rows' boxes.
        .padding(.leading, PVSpacing.space5 + Self.thumbnailSize + PVSpacing.space5)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceSunken)
    }

    private func row(_ artifact: CatalogArtifact) -> some View {
        HStack(spacing: PVSpacing.space5) {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: artifact.thumbnailRelPath,
                mediaType: artifact.file?.mediaType ?? "",
                originalFilename: artifact.file?.originalFilename ?? "",
                typeIconKey: artifact.thumbnailTypeIconKey(sourceTypeIconKey: sourceTypeIconKey),
                size: Self.thumbnailSize
            )
            .accessibilityHidden(true)
            rowCheckbox(artifact)
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(artifact.id)")
            Spacer(minLength: 0)
            Text(verbatim: artifact.ref)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
                .accessibilityHidden(true)
        }
        .padding(.vertical, PVSpacing.space2)
        .padding(.horizontal, PVSpacing.space5)
    }

    @ViewBuilder
    private func rowCheckbox(_ artifact: CatalogArtifact) -> some View {
        if let hint = locked[artifact.id] {
            // Always the labelled form: an unlabelled Artifact shows its ref as
            // the label, so the hint is never alone.
            PVCheckbox(
                verbatim: artifact.displayTitle,
                description: hint,
                isChecked: .constant(true),
                isLocked: true
            )
            .accessibilityValue(Text(verbatim: L10n.ArtifactChecklist.lockedValue(ref: artifact.ref)))
        } else {
            let isChecked = Binding(
                get: { !excluded.contains(artifact.id) },
                set: { on in
                    if on { excluded.remove(artifact.id) } else { excluded.insert(artifact.id) }
                }
            )
            if artifact.hasDisplayLabel {
                PVCheckbox(verbatim: artifact.displayTitle, isChecked: isChecked, isDisabled: isDisabled)
                    .accessibilityValue(Text(verbatim: artifact.ref))
            } else {
                PVCheckbox(
                    isChecked: isChecked,
                    isDisabled: isDisabled,
                    accessibilitySpokenLabel: artifact.displayTitle
                )
                .accessibilityValue(Text(verbatim: artifact.ref))
            }
        }
    }
}

#Preview("PVArtifactChecklist") {
    PVArtifactChecklistPreview()
        .frame(width: 440)
        .padding(PVSpacing.space8)
        .background(PVColor.surfacePage)
}

private struct PVArtifactChecklistPreview: View {
    @State private var excluded: Set<String> = ["a1"]

    private let artifacts: [CatalogArtifact] = (1...4).map { number in
        CatalogArtifact(
            id: "a\(number)",
            ref: "ART-0000\(number)",
            sourceID: "src",
            fileID: "",
            label: "Page \(number)",
            description: ""
        )
    }

    var body: some View {
        PVArtifactChecklist(
            artifacts: artifacts,
            excluded: $excluded,
            projectDir: "",
            sourceTypeIconKey: nil,
            isDisabled: false,
            locked: ["a3": "2 citations in this graph use this artifact"],
            accessibilityIdentifierPrefix: "preview.artifacts"
        )
    }
}
