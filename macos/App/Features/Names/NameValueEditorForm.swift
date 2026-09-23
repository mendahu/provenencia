import AppKit
import SwiftUI

/// Shared NameValue create/edit form body (full form + optional ordered parts).
struct NameValueEditorForm: View {
    @Binding var draft: NameValueDraft
    var accessibilityIdentifierPrefix: String = "nameValue"
    var showStoredAs: Bool = true

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            formSection
            PVDivider()
            partsSection
            if showStoredAs {
                PVDivider()
                storedAsSection
            }
        }
    }

    private var formSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            PVField(
                label: L10n.NameValue.formLabel,
                hint: draft.formError == nil ? L10n.NameValue.formHint : nil,
                error: draft.formError,
                required: true
            ) {
                PVInput(text: formBinding, size: .sm, isInvalid: draft.formError != nil)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).form")
                    .accessibilityHint(draft.formError.map { Text(verbatim: $0) } ?? Text(L10n.NameValue.formHint))
            }
        }
    }

    private var partsSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            Text(L10n.NameValue.partsHeading)
                .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                .tracking(PVTypeScale.micro * PVTracking.caps)
                .textCase(.uppercase)
                .foregroundStyle(PVColor.textMuted)

            if draft.parts.isEmpty {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.NameValue.partsEmpty)
                        .font(PVFont.body(size: PVTypeScale.body))
                        .foregroundStyle(PVColor.textSecondary)
                    Text(L10n.NameValue.partsEmptyHint)
                        .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
            } else {
                PVReorderableList(items: draft.parts, onMove: { source, destination in
                    draft.movePart(from: source, to: destination)
                }) { part in
                    if let index = draft.parts.firstIndex(where: { $0.id == part.id }) {
                        partRow(at: index)
                    }
                }
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).parts")
                Text(L10n.NameValue.partsHint)
                    .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }

            PVButton(L10n.NameValue.partsAdd, variant: .ghost, size: .sm) {
                draft.addPart()
            }
            .accessibilityIdentifier("\(accessibilityIdentifierPrefix).parts.add")
        }
    }

    private func partRow(at index: Int) -> some View {
        let part = draft.parts[index]
        let count = draft.parts.count
        let typeLabel = String(localized: NamePartType.label(forRaw: part.type))
        let groupLabel = L10n.NameValue.partAccessibility(
            position: index + 1,
            of: count,
            typeLabel: typeLabel
        )
        return VStack(alignment: .leading, spacing: PVSpacing.space2) {
            HStack(alignment: .bottom, spacing: PVSpacing.space3) {
                PVReorderHandle()
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.NameValue.partValueLabel)
                        .font(PVFont.body(size: PVTypeScale.micro))
                        .foregroundStyle(draft.partValueError(at: index) == nil ? PVColor.textSecondary : PVColor.danger)
                    PVInput(
                        text: partValueBinding(index),
                        size: .sm,
                        isInvalid: draft.partValueError(at: index) != nil
                    )
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).value")
                }
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.NameValue.partTypeLabel)
                        .font(PVFont.body(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textSecondary)
                    Picker(selection: partTypeBinding(index)) {
                        Text(L10n.NameValue.partTypeNone).tag("")
                        ForEach(NamePartType.allCases, id: \.rawValue) { type in
                            Text(type.label).tag(type.rawValue)
                        }
                    } label: {
                        EmptyView()
                    }
                    .labelsHidden()
                    .frame(width: 168)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).type")
                }
                HStack(spacing: PVSpacing.space1) {
                    PVIconButton(
                        .sortAscending,
                        label: L10n.NameValue.partMoveUpLabel,
                        accessibilityLabel: L10n.NameValue.partMoveUp(position: index + 1),
                        size: .sm,
                        action: { moveAndAnnounce(at: index, by: -1) }
                    )
                    .keyboardShortcut(.upArrow, modifiers: .option)
                    .disabled(index == 0)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).moveUp")
                    PVIconButton(
                        .sortDescending,
                        label: L10n.NameValue.partMoveDownLabel,
                        accessibilityLabel: L10n.NameValue.partMoveDown(position: index + 1),
                        size: .sm,
                        action: { moveAndAnnounce(at: index, by: 1) }
                    )
                    .keyboardShortcut(.downArrow, modifiers: .option)
                    .disabled(index == count - 1)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).moveDown")
                    PVIconButton(
                        .trash,
                        label: L10n.NameValue.partRemoveLabel,
                        accessibilityLabel: L10n.NameValue.partRemove(position: index + 1),
                        size: .sm,
                        tone: .danger,
                        action: { draft.removePart(id: part.id) }
                    )
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).remove")
                }
            }
            if let error = draft.partValueError(at: index) {
                Text(verbatim: error)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.danger)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).error")
            }
        }
        .padding(.vertical, PVSpacing.space2)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: groupLabel))
    }

    private var storedAsSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(L10n.NameValue.storedAs)
                .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                .tracking(PVTypeScale.micro * PVTracking.caps)
                .textCase(.uppercase)
                .foregroundStyle(PVColor.textMuted)
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                storedLine(label: "form", value: draft.trimmedForm)
                storedLine(label: "parts", value: draft.storedPartsLine)
            }
            .padding(PVSpacing.space5)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(PVColor.surfaceSunken)
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(PVColor.borderSubtle, lineWidth: 1)
            )
            .accessibilityIdentifier("\(accessibilityIdentifierPrefix).storedAs")
        }
    }

    private func storedLine(label: String, value: String) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            Text(verbatim: label)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
                .frame(width: 40, alignment: .leading)
            if value.isEmpty {
                Text(verbatim: "—")
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textFaint)
            } else {
                Text(verbatim: value)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textSecondary)
            }
        }
    }

    private var formBinding: Binding<String> {
        Binding(
            get: { draft.form },
            set: { next in
                draft.form = next
                draft.formTouched = true
            }
        )
    }

    private func partValueBinding(_ index: Int) -> Binding<String> {
        Binding(
            get: { draft.parts.indices.contains(index) ? draft.parts[index].value : "" },
            set: { next in
                guard draft.parts.indices.contains(index) else { return }
                draft.parts[index].value = next
            }
        )
    }

    private func partTypeBinding(_ index: Int) -> Binding<String> {
        Binding(
            get: { draft.parts.indices.contains(index) ? draft.parts[index].type : "" },
            set: { next in
                guard draft.parts.indices.contains(index) else { return }
                draft.parts[index].type = next
            }
        )
    }

    private func moveAndAnnounce(at index: Int, by delta: Int) {
        draft.movePart(at: index, by: delta)
        let dest = index + delta
        guard draft.parts.indices.contains(dest) else { return }
        let message = L10n.NameValue.partMoved(position: dest + 1, of: draft.parts.count)
        NSAccessibility.post(
            element: NSApp as Any,
            notification: .announcementRequested,
            userInfo: [.announcement: message, .priority: NSAccessibilityPriorityLevel.medium]
        )
    }
}
