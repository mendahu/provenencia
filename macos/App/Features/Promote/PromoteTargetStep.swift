import SwiftUI

/// Step 1 of Promote: where the subject goes — a new handle, or an existing
/// one of the same kind, found by search or picked from suggestions (S9-D9
/// frames 01–04).
struct PromoteTargetStep: View {
    @Bindable var model: PromoteModel
    let suggestions: QueryHandle<[CatalogPromoteTargetSuggestion]>?

    private var kind: EvidencePrimaryKind { model.kind }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            PVSectionHeader(title: L10n.Promote.chooseStep(kind), meta: model.stepText)

            VStack(alignment: .leading, spacing: 14) {
                PVRadio(
                    verbatim: L10n.string(L10n.Promote.newOption(kind)),
                    description: L10n.Promote.newDescription(kind, name: model.entry.subjectName),
                    isSelected: model.choice == .new
                ) {
                    model.choose(.new)
                }
                .accessibilityIdentifier("promote.choice.new")
                PVRadio(
                    verbatim: L10n.string(L10n.Promote.existingOption(kind)),
                    description: L10n.string(L10n.Promote.existingDescription(kind)),
                    isSelected: model.choice == .existing
                ) {
                    model.choose(.existing)
                }
                .accessibilityIdentifier("promote.choice.existing")
            }
            .accessibilityElement(children: .contain)
            .accessibilityLabel(Text(verbatim: L10n.Promote.choiceLabel(name: model.entry.subjectName)))

            if model.choice == .existing {
                VStack(alignment: .leading, spacing: 18) {
                    search
                    if let suggestions {
                        PromoteSuggestions(handle: suggestions, model: model)
                    } else {
                        PromoteSuggestionsNote(text: L10n.string(L10n.Promote.suggestedLoading))
                    }
                }
                .padding(.leading, 28)
            }

            if let error = model.saveError {
                PVCallout(tone: .danger, message: error)
                    .accessibilityIdentifier("promote.error")
            }
        }
    }

    private var search: some View {
        PVComboBox(
            selection: Binding(
                get: { model.searchSelection },
                set: { model.selectSearchResult($0) }
            ),
            options: model.searchOptions,
            placeholder: L10n.Promote.searchPlaceholder(kind),
            label: L10n.Promote.searchLabel(kind),
            accessibilityIdentifierPrefix: "promote.search",
            onQueryChange: { model.updateQuery($0) },
            row: { option, query in
                PromoteSearchRow(option: option, query: query)
            },
            empty: { query in
                if query.isEmpty {
                    Text(L10n.Promote.searchPrompt(kind))
                } else {
                    Text(verbatim: L10n.Promote.searchNoMatch(kind, query: query))
                }
            }
        )
    }
}

/// A search result in the Promote picker: the handle's name, with its ref on
/// the line under it. Life years and places join the name line with S9-32.
private struct PromoteSearchRow: View {
    let option: PVComboBoxOption
    let query: String

    var body: some View {
        VStack(alignment: .leading, spacing: 1) {
            Text(PVComboBoxHighlight.attributed(option.label, query: query))
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(1)
            Text(verbatim: option.subtext)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

/// The Suggested group: header with count and how it is ranked, then the
/// candidates as radio rows, or the no-suggestions state (frame 04).
private struct PromoteSuggestions: View {
    @Bindable var handle: QueryHandle<[CatalogPromoteTargetSuggestion]>
    let model: PromoteModel

    var body: some View {
        if let rows = handle.value {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                PVSectionHeader(title: L10n.Promote.suggested, meta: "\(rows.count)") {
                    if !rows.isEmpty {
                        Text(L10n.Promote.suggestedAside)
                            .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                            .foregroundStyle(PVColor.textMuted)
                    }
                }
                if rows.isEmpty {
                    PVEmptyState(
                        icon: .userSearch,
                        verbatimTitle: L10n.Promote.noSuggestionsTitle(model.kind, name: model.entry.subjectName),
                        message: L10n.string(L10n.Promote.noSuggestionsMessage(model.kind)),
                        compact: true
                    )
                    .accessibilityIdentifier("promote.suggestions.empty")
                } else {
                    VStack(spacing: 0) {
                        ForEach(Array(rows.enumerated()), id: \.element.id) { index, suggestion in
                            if index > 0 {
                                Rectangle().fill(PVColor.borderSubtle).frame(height: 1)
                            }
                            PromoteCandidateRow(
                                suggestion: suggestion,
                                kind: model.kind,
                                isSelected: model.target?.entityID == suggestion.entity.id
                            ) {
                                model.select(PromoteModel.target(for: suggestion))
                            }
                        }
                    }
                    .background(PVColor.surfaceCard)
                    .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
                    .overlay(
                        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                            .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                    )
                    .accessibilityElement(children: .contain)
                    .accessibilityLabel(Text(L10n.Promote.suggested))
                }
            }
        } else if handle.error != nil {
            PromoteSuggestionsNote(text: L10n.string(L10n.Promote.suggestedFailed))
        } else {
            PromoteSuggestionsNote(text: L10n.string(L10n.Promote.suggestedLoading))
        }
    }
}

private struct PromoteSuggestionsNote: View {
    let text: String

    var body: some View {
        Text(verbatim: text)
            .font(PVFont.body(size: PVTypeScale.caption, italic: true))
            .foregroundStyle(PVColor.textMuted)
    }
}

/// One candidate: the D2 list row (tile, title, members line, trailing ref)
/// with a leading radio. Life years and places fill in with S9-32.
struct PromoteCandidateRow: View {
    let suggestion: CatalogPromoteTargetSuggestion
    let kind: EvidencePrimaryKind
    let isSelected: Bool
    let action: () -> Void

    @State private var isHovered = false

    private var members: String { PromoteModel.members(suggestion.memberCount) }

    var body: some View {
        Button(action: action) {
            HStack(spacing: PVSpacing.space5) {
                PVRadioMark(isSelected: isSelected)
                PromoteKindTile(
                    kind: kind,
                    size: 28,
                    markSize: 15,
                    background: EvidenceSubjectKindStyle.forKind(kind).tint
                )
                VStack(alignment: .leading, spacing: 3) {
                    title
                        .font(PVFont.display(size: 15, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    Text(verbatim: members)
                        .font(PVFont.body(size: 12.5, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                Spacer(minLength: PVSpacing.space5)
                Text(verbatim: suggestion.entity.ref)
                    .font(PVFont.mono(size: 11.5))
                    .foregroundStyle(PVColor.textMuted)
            }
            .padding(.vertical, 10)
            .padding(.horizontal, 14)
            .background(isSelected ? PVColor.surfaceActive : (isHovered ? PVColor.surfaceHover : .clear))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .onHover { isHovered = $0 }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(Text(verbatim: L10n.Promote.rowAccessibility(
            title: PromoteModel.title(for: suggestion),
            members: members,
            ref: suggestion.entity.ref
        )))
        .accessibilityAddTraits(isSelected ? [.isButton, .isSelected] : .isButton)
        .accessibilityIdentifier("promote.candidate.\(suggestion.entity.ref)")
    }

    @ViewBuilder
    private var title: some View {
        if let person = suggestion.person {
            ConclusionListRow.title(PersonHeaderDisplay.titleSource(person))
        } else {
            Text(verbatim: PromoteModel.title(for: suggestion))
        }
    }
}
