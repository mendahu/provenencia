import SwiftUI

/// Display + pencil → editor + Save/Cancel.
///
/// Extracted from Provenencia Source-page chrome during structural cleanup
/// (not a direct port of a web-kit component). Unifies the repeated
/// resting/editing patterns that previously differed only in Save/Cancel
/// placement and cancel chrome (text button vs dismiss icon).
///
/// Accessibility identifiers are left to the call site
/// (`docs/macos-client-patterns.md` §5), except when
/// `accessibilityIdentifierPrefix` is provided for the built-in controls.
struct PVInlineEdit<Display: View, Editor: View, RestingTrailing: View, EditingTrailing: View>: View {
    /// How content and action controls are arranged while editing.
    enum Axis {
        /// Content and Save/Cancel share one row (title, compact metadata).
        case horizontal
        /// Editor stacks above the action row (description, notes).
        case vertical
    }

    let isEditing: Bool
    let isSaving: Bool
    var error: String? = nil
    var saveLabel: LocalizedStringResource
    var cancelLabel: LocalizedStringResource
    var editLabel: LocalizedStringResource
    var saveDisabled: Bool = false
    /// When false, the resting pencil is omitted (caller hosts Edit elsewhere).
    var showsEditControl: Bool = true
    /// When true (default), the display/editor expands and pushes trailing
    /// actions to the far edge. When false, content hugs its text so the
    /// pencil / Save sit immediately after (Source title).
    var expandsContent: Bool = true
    var axis: Axis = .vertical
    /// Labeled text buttons (default) vs stacked check / dismiss icons
    /// (metadata board).
    var actionsStyle: PVInlineEditActions.Style = .labeled
    var accessibilityIdentifierPrefix: String? = nil
    let onEdit: () -> Void
    let onSave: () -> Void
    let onCancel: () -> Void
    @ViewBuilder var display: () -> Display
    @ViewBuilder var editor: () -> Editor
    @ViewBuilder var restingTrailing: () -> RestingTrailing
    @ViewBuilder var editingTrailing: () -> EditingTrailing

    var body: some View {
        Group {
            if isEditing {
                editingBody
            } else {
                restingBody
            }
        }
    }

    @ViewBuilder
    private var restingBody: some View {
        HStack(alignment: .top, spacing: PVSpacing.space3) {
            display()
                .modifier(PVInlineEditContentFlex(expands: expandsContent))
            if showsEditControl {
                PVIconButton(.penLine, label: editLabel, size: .sm, action: onEdit)
                    .accessibilityIdentifier(prefixed("edit"))
            }
            restingTrailing()
        }
    }

    @ViewBuilder
    private var editingBody: some View {
        switch axis {
        case .horizontal:
            HStack(alignment: .top, spacing: PVSpacing.space3) {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    editor()
                    errorCaption
                }
                .modifier(PVInlineEditContentFlex(expands: expandsContent))
                actions
                editingTrailing()
            }
        case .vertical:
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                editor()
                errorCaption
                HStack(spacing: PVSpacing.space4) {
                    actions
                    Spacer(minLength: 0)
                    editingTrailing()
                }
            }
        }
    }

    @ViewBuilder
    private var errorCaption: some View {
        if let error, !error.isEmpty {
            Text(error)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.danger)
        }
    }

    private var actions: some View {
        PVInlineEditActions(
            isSaving: isSaving,
            saveLabel: saveLabel,
            cancelLabel: cancelLabel,
            saveDisabled: saveDisabled,
            showsCancel: true,
            style: actionsStyle,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onSave: onSave,
            onCancel: onCancel
        )
    }

    private func prefixed(_ suffix: String) -> String {
        if let accessibilityIdentifierPrefix, !accessibilityIdentifierPrefix.isEmpty {
            return "\(accessibilityIdentifierPrefix).\(suffix)"
        }
        return suffix
    }
}

extension PVInlineEdit where RestingTrailing == EmptyView, EditingTrailing == EmptyView {
    init(
        isEditing: Bool,
        isSaving: Bool,
        error: String? = nil,
        saveLabel: LocalizedStringResource,
        cancelLabel: LocalizedStringResource,
        editLabel: LocalizedStringResource,
        saveDisabled: Bool = false,
        showsEditControl: Bool = true,
        expandsContent: Bool = true,
        axis: Axis = .vertical,
        actionsStyle: PVInlineEditActions.Style = .labeled,
        accessibilityIdentifierPrefix: String? = nil,
        onEdit: @escaping () -> Void,
        onSave: @escaping () -> Void,
        onCancel: @escaping () -> Void,
        @ViewBuilder display: @escaping () -> Display,
        @ViewBuilder editor: @escaping () -> Editor
    ) {
        self.init(
            isEditing: isEditing,
            isSaving: isSaving,
            error: error,
            saveLabel: saveLabel,
            cancelLabel: cancelLabel,
            editLabel: editLabel,
            saveDisabled: saveDisabled,
            showsEditControl: showsEditControl,
            expandsContent: expandsContent,
            axis: axis,
            actionsStyle: actionsStyle,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onEdit: onEdit,
            onSave: onSave,
            onCancel: onCancel,
            display: display,
            editor: editor,
            restingTrailing: { EmptyView() },
            editingTrailing: { EmptyView() }
        )
    }
}

/// Expands to fill the row (metadata) or hugs text so trailing actions sit
/// immediately after the content (title).
private struct PVInlineEditContentFlex: ViewModifier {
    var expands: Bool

    func body(content: Content) -> some View {
        if expands {
            content.frame(maxWidth: .infinity, alignment: .leading)
        } else {
            content.fixedSize(horizontal: false, vertical: true)
        }
    }
}

/// Shared Save / Cancel chrome used by `PVInlineEdit` and always-editable
/// surfaces (artifact fields, credibility) that only need the action row.
struct PVInlineEditActions: View {
    /// How Save / Cancel are rendered.
    enum Style {
        /// Primary + ghost text buttons in a horizontal row.
        case labeled
        /// Stacked icon buttons: filled check (Save) over dismiss (Cancel).
        case iconStack
    }

    let isSaving: Bool
    var saveLabel: LocalizedStringResource
    var cancelLabel: LocalizedStringResource
    var saveDisabled: Bool = false
    var showsCancel: Bool = true
    var style: Style = .labeled
    var accessibilityIdentifierPrefix: String? = nil
    let onSave: () -> Void
    let onCancel: () -> Void

    var body: some View {
        switch style {
        case .labeled:
            labeledBody
        case .iconStack:
            iconStackBody
        }
    }

    private var labeledBody: some View {
        HStack(spacing: PVSpacing.space4) {
            PVButton(
                saveLabel,
                variant: .primary,
                size: .sm,
                loading: isSaving,
                action: onSave
            )
            .disabled(saveDisabled && !isSaving)
            .accessibilityIdentifier(prefixed("save"))

            if showsCancel {
                PVButton(cancelLabel, variant: .ghost, size: .sm, action: onCancel)
                    .disabled(isSaving)
                    .accessibilityIdentifier(prefixed("cancel"))
            }
        }
    }

    private var iconStackBody: some View {
        VStack(spacing: PVSpacing.space2) {
            if isSaving {
                ProgressView()
                    .controlSize(.small)
                    .frame(width: PVControlSize.sm.height, height: PVControlSize.sm.height)
                    .accessibilityIdentifier(prefixed("save"))
            } else {
                PVIconButton(.check, label: saveLabel, size: .sm, tone: .accent, action: onSave)
                    .disabled(saveDisabled)
                    .accessibilityIdentifier(prefixed("save"))
            }

            if showsCancel {
                PVIconButton(.dismiss, label: cancelLabel, size: .sm, action: onCancel)
                    .disabled(isSaving)
                    .accessibilityIdentifier(prefixed("cancel"))
            }
        }
    }

    private func prefixed(_ suffix: String) -> String {
        if let accessibilityIdentifierPrefix, !accessibilityIdentifierPrefix.isEmpty {
            return "\(accessibilityIdentifierPrefix).\(suffix)"
        }
        return suffix
    }
}

#Preview("Resting") {
    PVInlineEdit(
        isEditing: false,
        isSaving: false,
        saveLabel: LocalizedStringResource("Save"),
        cancelLabel: LocalizedStringResource("Cancel"),
        editLabel: LocalizedStringResource("Edit"),
        onEdit: {},
        onSave: {},
        onCancel: {}
    ) {
        Text("A note about the parish register")
            .font(PVFont.body(size: PVTypeScale.bodySmall))
            .foregroundStyle(PVColor.textSecondary)
    } editor: {
        Text("editor")
    }
    .padding(PVSpacing.space6)
    .frame(width: 420)
}

#Preview("Editing vertical") {
    PVInlineEdit(
        isEditing: true,
        isSaving: false,
        error: nil,
        saveLabel: LocalizedStringResource("Save"),
        cancelLabel: LocalizedStringResource("Cancel"),
        editLabel: LocalizedStringResource("Edit"),
        axis: .vertical,
        onEdit: {},
        onSave: {},
        onCancel: {}
    ) {
        Text("display")
    } editor: {
        Text("Draft body…")
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(PVSpacing.space4)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(PVColor.borderDefault, lineWidth: 1)
            )
    }
    .padding(PVSpacing.space6)
    .frame(width: 420)
}

#Preview("Editing icon stack") {
    PVInlineEdit(
        isEditing: true,
        isSaving: false,
        saveLabel: LocalizedStringResource("Save value"),
        cancelLabel: LocalizedStringResource("Cancel"),
        editLabel: LocalizedStringResource("Edit value"),
        axis: .horizontal,
        actionsStyle: .iconStack,
        onEdit: {},
        onSave: {},
        onCancel: {}
    ) {
        Text("display")
    } editor: {
        Text("Norfolk Record Office")
            .font(PVFont.mono(size: PVTypeScale.caption))
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(PVSpacing.space3)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(PVColor.borderDefault, lineWidth: 1)
            )
    }
    .padding(PVSpacing.space6)
    .frame(width: 420)
}
