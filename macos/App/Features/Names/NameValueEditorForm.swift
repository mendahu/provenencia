import AppKit
import SwiftUI

/// Shared NameValue create/edit form body (S7-D5 board). Hosts embed this
/// on a sheet or inline — the body is the same either way.
struct NameValueEditorForm: View {
    @Binding var draft: NameValueDraft
    var accessibilityIdentifierPrefix: String = "nameValue"
    var showStoredAs: Bool = true
    /// Host focus for the form field. The editor does not take focus on its own.
    var formFocus: PVInputFocus? = nil
    @State private var partValueToFocus: UUID?
    @State private var pendingSplit: PendingNameSplit?

    private var typeOptions: [PVSelectOption] {
        NamePartType.selectOptions
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            formSection
            PVDivider()
            partsSection
            if showStoredAs {
                storedAsSection
            }
        }
        .pvConfirm(
            item: $pendingSplit,
            copy: { _ in
                PVConfirmCopy(
                    title: L10n.string(L10n.NameValue.splitConfirmTitle),
                    message: L10n.string(L10n.NameValue.splitConfirmMessage),
                    confirm: L10n.NameValue.splitConfirmReplace,
                    cancel: L10n.NameValue.splitConfirmKeep
                )
            },
            tone: .irreversible,
            accessibilityIdentifierPrefix: "\(accessibilityIdentifierPrefix).splitConfirm",
            onConfirm: applyPendingSplit,
            detail: splitConfirmDetail
        )
    }

    private var formSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                Text(L10n.NameValue.formLabel)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textSecondary)
                Text(L10n.NameValue.formRequired)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textFaint)
            }
            HStack(alignment: .center, spacing: PVSpacing.space4) {
                PVInput(
                    text: formBinding,
                    size: .md,
                    prompt: L10n.NameValue.formPlaceholder,
                    isInvalid: draft.formError != nil,
                    focus: formFocus,
                    onBlur: fillEmptyPartsFromForm
                )
                .frame(maxWidth: .infinity)
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).form")
                .accessibilityHint(draft.formError.map { Text(verbatim: $0) } ?? Text(L10n.NameValue.formHint))
                PVButton(
                    L10n.NameValue.split,
                    variant: .secondary,
                    size: .md,
                    action: splitFromForm
                )
                .disabled(!draft.canSplitForm)
                .fixedSize(horizontal: true, vertical: false)
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).split")
            }
            if let error = draft.formError {
                Text(verbatim: error)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.danger)
            }
            Text(L10n.NameValue.formHint)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    private var partsSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .firstTextBaseline) {
                Text(L10n.NameValue.partsHeading)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .tracking(PVTypeScale.micro * PVTracking.caps)
                    .textCase(.uppercase)
                    .foregroundStyle(PVColor.textMuted)
                Spacer()
                Text(verbatim: L10n.NameValue.partsCount(draft.parts.count))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textFaint)
            }

            if draft.parts.isEmpty {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.NameValue.partsEmpty)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                    Text(L10n.NameValue.partsEmptyHint)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                .padding(14)
                .frame(maxWidth: .infinity, alignment: .leading)
                .overlay(
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .strokeBorder(style: StrokeStyle(lineWidth: 1, dash: [4, 3]))
                        .foregroundStyle(PVColor.borderDefault)
                )
            } else {
                VStack(alignment: .leading, spacing: PVSpacing.space3) {
                    partHeader
                    ForEach(Array(draft.parts.enumerated()), id: \.element.id) { index, _ in
                        partRow(at: index)
                    }
                }
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).parts")
            }

            HStack(alignment: .center, spacing: PVSpacing.space5) {
                PVButton(
                    L10n.NameValue.partsAdd,
                    variant: .ghost,
                    size: .sm,
                    icon: .plus
                ) {
                    partValueToFocus = draft.addPart()
                }
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).parts.add")
                Text(L10n.NameValue.partsHint)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textFaint)
                    .frame(maxWidth: .infinity, alignment: .trailing)
            }
        }
    }

    private var partHeader: some View {
        HStack(spacing: PVSpacing.space4) {
            Color.clear.frame(width: 22)
            Text(L10n.NameValue.partValueLabel)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .frame(maxWidth: .infinity, alignment: .leading)
            Text(L10n.NameValue.partTypeLabel)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .frame(width: NamePartTypeColumn.width, alignment: .leading)
            Color.clear.frame(width: 88)
        }
    }

    private func partRow(at index: Int) -> some View {
        let part = draft.parts[index]
        let count = draft.parts.count
        let typeLabel = L10n.string(NamePartType.label(forRaw: part.type))
        let groupLabel = L10n.NameValue.partAccessibility(
            position: index + 1,
            of: count,
            typeLabel: typeLabel
        )
        return VStack(alignment: .leading, spacing: PVSpacing.space2) {
            HStack(spacing: PVSpacing.space4) {
                Text(verbatim: "\(index + 1)")
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textFaint)
                    .frame(width: 22, alignment: .trailing)
                PVInput(
                    text: partValueBinding(index),
                    size: .sm,
                    isInvalid: draft.partValueError(at: index) != nil,
                    activateOnAppear: part.id == partValueToFocus
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).value")
                PVSelect(
                    selection: partTypeBinding(index),
                    options: typeOptions,
                    size: .sm,
                    menuWidth: 190,
                    fillsWidth: true,
                    accessibilityLabel: L10n.NameValue.partTypeLabel,
                    accessibilityIdentifier: "\(accessibilityIdentifierPrefix).part.\(index).type"
                )
                .frame(width: NamePartTypeColumn.width)
                HStack(spacing: 2) {
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
                        .dismiss,
                        label: L10n.NameValue.partRemoveLabel,
                        accessibilityLabel: L10n.NameValue.partRemove(position: index + 1),
                        size: .sm,
                        action: { draft.removePart(id: part.id) }
                    )
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).remove")
                }
                .frame(width: 88)
            }
            if let error = draft.partValueError(at: index) {
                Text(verbatim: error)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.danger)
                    .padding(.leading, 30)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).part.\(index).error")
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: groupLabel))
    }

    private var storedAsSection: some View {
        HStack(alignment: .top, spacing: 14) {
            Text(L10n.NameValue.storedAs)
                .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                .tracking(PVTypeScale.micro * PVTracking.caps)
                .textCase(.uppercase)
                .foregroundStyle(PVColor.textMuted)
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                storedLine(label: "form", value: draft.storedFormLine, emphasis: true)
                storedLine(label: "parts", value: draft.storedPartsLine, emphasis: false)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .background(PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(PVColor.borderSubtle, lineWidth: 1)
        )
        .accessibilityIdentifier("\(accessibilityIdentifierPrefix).storedAs")
    }

    private func storedLine(label: String, value: String, emphasis: Bool) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            Text(verbatim: label)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(emphasis ? PVColor.textPrimary : PVColor.textMuted)
            Text(verbatim: value)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(emphasis ? PVColor.textPrimary : PVColor.textMuted)
                .frame(maxWidth: .infinity, alignment: .leading)
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
                let id = draft.parts[index].id
                var nextDraft = draft
                nextDraft.setPartValue(id: id, value: next)
                draft = nextDraft
            }
        )
    }

    private func partTypeBinding(_ index: Int) -> Binding<String> {
        Binding(
            get: { draft.parts.indices.contains(index) ? draft.parts[index].type : "" },
            set: { next in
                guard draft.parts.indices.contains(index) else { return }
                let id = draft.parts[index].id
                var nextDraft = draft
                nextDraft.setPartType(id: id, type: next)
                draft = nextDraft
            }
        )
    }

    private func fillEmptyPartsFromForm() {
        var next = draft
        next.fillEmptyPartsFromForm()
        draft = next
    }

    private func splitFromForm() {
        var next = draft
        switch next.splitFromForm() {
        case .none:
            break
        case .applied:
            draft = next
        case .needsConfirm(let parts):
            pendingSplit = PendingNameSplit(parts: parts)
        }
    }

    private func applyPendingSplit() {
        guard let pending = pendingSplit else { return }
        var next = draft
        next.applySplit(pending.parts)
        draft = next
        pendingSplit = nil
    }

    private func splitConfirmDetail(_ pending: PendingNameSplit) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            Text(L10n.NameValue.splitConfirmPending)
                .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                .tracking(PVTypeScale.micro * PVTracking.caps)
                .textCase(.uppercase)
                .foregroundStyle(PVColor.textMuted)
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                ForEach(Array(pending.parts.enumerated()), id: \.element.id) { index, part in
                    HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
                        Text(verbatim: "\(index + 1)")
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textFaint)
                            .frame(width: 22, alignment: .trailing)
                        Text(NamePartType.label(forRaw: part.type))
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                        Text(verbatim: part.value)
                            .font(PVFont.body(size: PVTypeScale.bodySmall))
                            .foregroundStyle(PVColor.textPrimary)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }
                }
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(PVColor.surfaceSunken)
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(PVColor.borderSubtle, lineWidth: 1)
            )
        }
    }

    private func moveAndAnnounce(at index: Int, by delta: Int) {
        draft.movePart(at: index, by: delta)
        let dest = index + delta
        guard draft.parts.indices.contains(dest) else { return }
        let message = L10n.NameValue.partMoved(position: dest + 1, of: draft.parts.count)
        AccessibilityAnnouncement.post(message)
    }
}

/// Snapshot of the parts Split would write, so the confirm stays filled while it dismisses.
private struct PendingNameSplit: Identifiable {
    let id = UUID()
    let parts: [NameValueDraft.Part]
}

/// Closed part-type select, wide enough for the longest label and stable across selections.
private enum NamePartTypeColumn {
    @MainActor
    static var width: CGFloat {
        let font = PVFont.nsBody(size: PVTypeScale.caption)
        let longest = NamePartType.selectOptions
            .map { ($0.label as NSString).size(withAttributes: [.font: font]).width }
            .max() ?? 0
        let trailing = PVInputChrome.horizontalInset
            + PVSelectFieldChrome.chevronSize
            + PVInputChrome.iconTextGap
        return ceil(longest + PVInputChrome.horizontalInset + trailing + 2)
    }
}
