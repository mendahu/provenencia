import SwiftUI

/// Properties workspace destination: type strip + dual cards
/// (property table + inspector). Chrome matches the board HTML.
struct PropertiesView: View {
    @Environment(WorkspaceSession.self) private var session
    @State private var model: PropertiesModel

    private let inspectorWidth: CGFloat = PVSpacing.widthInspector

    init(
        session: WorkspaceSession,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        _model = State(
            initialValue: PropertiesModel(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        )
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<PropertiesSnapshot> = session.queryHandle(
                PropertiesModel.workspaceKey(for: session)
            ) {
                PropertiesContent(handle: handle, model: model, inspectorWidth: inspectorWidth)
            } else {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .task {
            model.warmWorkspaceQuery()
        }
    }
}

private struct PropertiesContent: View {
    @Bindable var handle: QueryHandle<PropertiesSnapshot>
    @Bindable var model: PropertiesModel
    let inspectorWidth: CGFloat
    @Environment(WorkspaceNavigation.self) private var navigation
    @FocusState private var searchFocused: Bool

    var body: some View {
        VStack(spacing: 0) {
            header
            if let loadError = model.loadError, model.snapshot.properties.isEmpty {
                PVCallout(tone: .danger, message: L10n.Errors.message(for: loadError))
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.top, PVSpacing.space8)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            } else {
                typeStrip
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.bottom, PVSpacing.space6)
                HStack(alignment: .top, spacing: PVSpacing.space7) {
                    listCard
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                    inspectorCard
                        .frame(width: inspectorWidth)
                        .frame(maxHeight: .infinity)
                }
                .padding(.horizontal, PVSpacing.gutterPage)
                .padding(.bottom, PVSpacing.space8)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .vocabularyToastOverlay($model.toast, identifier: "properties.toast")
        .pvDeleteImpact(
            flow: model.deleteImpact,
            accessibilityIdentifierPrefix: "properties.deleteImpact",
            onConfirm: {
                Task {
                    if await model.confirmPendingImpact() {
                        navigation.fallbackToSectionRoot()
                    }
                }
            },
            onNavigate: { location in
                navigation.go(to: location)
            }
        )
        .onChange(of: navigation.currentLocation) { _, location in
            reconcileSelection(for: location)
        }
        .onChange(of: handle.status) { _, status in
            if status == .ready { model.syncCatalogCounts() }
            reconcileSelection(for: navigation.currentLocation)
        }
        .onChange(of: handle.value) { _, _ in
            reconcileSelection(for: navigation.currentLocation)
        }
        .onChange(of: model.searchFocused) { _, focused in
            if focused { searchFocused = true }
        }
        .onChange(of: searchFocused) { _, focused in
            model.searchFocused = focused
        }
        .onChange(of: model.addPropertySelection) { _, value in
            guard !value.isEmpty else { return }
            Task {
                if let location = await model.addPropertyFromCombo() {
                    navigation.go(to: location)
                }
            }
        }
        .onAppear {
            reconcileSelection(for: navigation.currentLocation)
            if handle.status == .ready { model.syncCatalogCounts() }
        }
        .background {
            PropertiesCreateHost(model: model)
            Button(action: { model.focusSearch() }) { EmptyView() }
                .keyboardShortcut("f", modifiers: .command)
                .opacity(0)
                .accessibilityHidden(true)
            Button(action: { model.openCreate() }) { EmptyView() }
                .keyboardShortcut("n", modifiers: .command)
                .opacity(0)
                .accessibilityHidden(true)
        }
        .onKeyPress(.space) {
            guard model.selectedType != nil, model.selectedProperty != nil else {
                return .ignored
            }
            Task { await model.toggleFocusedTypeBinding() }
            return .handled
        }
        .accessibilityIdentifier("properties")
    }

    private func reconcileSelection(for location: WorkspaceLocation) {
        guard handle.status == .ready || !(handle.value?.properties ?? []).isEmpty else { return }
        let outcome = model.syncSelection(from: location)
        if outcome == .missingDeepId {
            navigation.fallbackToSectionRoot()
        }
    }

    private var header: some View {
        VocabularyHeader(
            title: L10n.Workspace.propertiesTitle,
            description: L10n.Properties.description,
            addLabel: L10n.Properties.newProperty,
            isAddDisabled: false,
            identifierPrefix: "properties",
            descriptionItalic: true,
            addSize: .sm,
            onAdd: { model.openCreate() }
        )
    }

    private var typeStrip: some View {
        HStack(spacing: PVSpacing.space4) {
            typeCard(
                key: nil,
                title: Text(L10n.Properties.allProperties),
                count: model.snapshot.properties.count,
                selected: model.selectedTypeKey == nil,
                bridge: false,
                ink: PVColor.textPrimary,
                mark: nil
            )
            ForEach(model.types) { type in
                let presentation = model.snapshot.presentation(for: type)
                typeCard(
                    key: type.key,
                    title: Text(verbatim: type.label),
                    count: model.snapshot.propertyCount(forTypeID: type.id),
                    selected: model.selectedTypeKey == type.key,
                    bridge: PropertiesTypeChrome.showsBridgeLabel(typeKey: type.key),
                    ink: PropertiesTypeChrome.ink(typeKey: type.key, presentation: presentation),
                    mark: PropertiesTypeChrome.stripMarkKey(typeKey: type.key)
                )
            }
        }
        .frame(maxWidth: .infinity)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.Properties.typeStripAccessibility))
    }

    private func typeCard(
        key: String?,
        title: Text,
        count: Int,
        selected: Bool,
        bridge: Bool,
        ink: Color,
        mark: PVMarkKey?
    ) -> some View {
        Button {
            navigation.go(to: model.location(afterPressingType: key))
        } label: {
            VStack(alignment: .leading, spacing: 6) {
                HStack(alignment: .top) {
                    if let mark {
                        PVMark(mark, size: 16)
                            .foregroundStyle(ink)
                    } else {
                        Image(systemName: "square.grid.2x2")
                            .font(.system(size: 14, weight: .medium))
                            .foregroundStyle(PVColor.textPrimary)
                            .frame(width: 16, height: 16)
                    }
                    Spacer(minLength: 0)
                    if bridge {
                        Text(L10n.Properties.bridgeRole)
                            .font(PVFont.mono(size: 10))
                            .tracking(1)
                            .textCase(.uppercase)
                            .foregroundStyle(PVColor.textFaint)
                    }
                }
                title
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                    .minimumScaleFactor(0.85)
                Text(verbatim: L10n.Properties.stripFieldCount(count: count))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
            }
            .padding(.horizontal, 10)
            .padding(.vertical, 9)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(selected ? PVColor.surfaceSelected : PVColor.surfaceCard)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .strokeBorder(selected ? PVColor.accentLine : PVColor.borderSubtle, lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
            .pvShadow(selected ? PVElevation.sm : [])
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(selected ? [.isSelected, .isButton] : .isButton)
    }

    private var listCard: some View {
        PVCard {
            VStack(spacing: 0) {
                listToolbar
                PVDivider()
                propertyTable
            }
        }
    }

    private var listToolbar: some View {
        HStack(spacing: PVSpacing.space5) {
            PVInput(
                text: $model.searchQuery,
                size: .sm,
                prompt: L10n.Properties.searchPlaceholder,
                icon: .search,
                focused: $searchFocused
            )
            .frame(maxWidth: 300)
            Spacer(minLength: 0)
            if let type = model.selectedType {
                PVComboBox(
                    selection: $model.addPropertySelection,
                    options: model.addPropertyOptions,
                    size: .sm,
                    placeholder: L10n.Properties.addPropertyPlaceholder(typeLabel: type.label),
                    emptyLabel: L10n.Properties.addPropertyEmpty,
                    label: L10n.Properties.addPropertyPlaceholder(typeLabel: type.label),
                    accessibilityIdentifierPrefix: "properties.addCombo"
                )
                .frame(width: 330)
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
    }

    private var propertyTable: some View {
        Group {
            if model.visibleProperties.isEmpty {
                VStack(spacing: PVSpacing.space3) {
                    Text(verbatim: L10n.Properties.emptySearchTitle(query: model.searchQuery))
                        .font(PVFont.display(size: PVTypeScale.h4))
                        .foregroundStyle(PVColor.textSecondary)
                    Text(L10n.Properties.emptySearch)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .padding(PVSpacing.space9)
            } else {
                PVTable(
                    rows: model.visibleProperties,
                    columns: tableColumns,
                    selection: Binding(
                        get: { model.selectedPropertyID },
                        set: { id in
                            guard let id else { return }
                            navigation.go(to: model.location(selectingProperty: id))
                        }
                    ),
                    primaryText: { $0.label },
                    density: .compact,
                    label: L10n.Workspace.propertiesTitle,
                    rowAccessibilityIdentifier: { "properties.row.\($0.id)" },
                    content: { EmptyView() }
                )
            }
        }
    }

    private var tableColumns: [PVTableColumn<CatalogProperty>] {
        var columns: [PVTableColumn<CatalogProperty>] = []
        if model.selectedType != nil {
            columns.append(
                PVTableColumn(id: "on", title: L10n.Properties.columnOn, width: 36) { property in
                    onCell(for: property)
                }
            )
        }
        columns.append(contentsOf: [
            PVTableColumn(id: "property", title: L10n.Properties.columnProperty) { property in
                VStack(alignment: .leading, spacing: 1) {
                    Text(verbatim: property.label)
                        .font(PVFont.body(size: PVTypeScale.body))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    Text(verbatim: property.key)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textFaint)
                        .lineLimit(1)
                }
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(rowAccessibilityLabel(property))
            },
            PVTableColumn(id: "valueType", title: L10n.Properties.columnValueType, width: 96) { property in
                PropertiesValueTypePill(valueType: property.valueType)
            },
            PVTableColumn(id: "origin", title: L10n.Properties.columnOrigin, width: 76) { property in
                PropertiesOriginCell(origin: property.origin)
            },
            PVTableColumn(id: "boundTo", title: L10n.Properties.columnBoundTo, width: 226) { property in
                boundToCell(for: property)
            },
        ])
        return columns
    }

    @ViewBuilder
    private func onCell(for property: CatalogProperty) -> some View {
        if let type = model.selectedType {
            let locked = model.bindingLocked(propertyID: property.id, typeID: type.id)
            let bound = model.isBound(propertyID: property.id, typeID: type.id)
            Button {
                // Apply locally so the toggle sees the row before the history reconcile lands.
                navigation.go(to: model.location(selectingProperty: property.id))
                model.selectProperty(property.id)
                Task { await model.toggleBinding(to: type) }
            } label: {
                PropertiesBindBox(on: bound, locked: locked)
            }
            .buttonStyle(.plain)
            .disabled(locked)
            .accessibilityLabel(onAccessibilityLabel(property: property, type: type, bound: bound, locked: locked))
        }
    }

    private func onAccessibilityLabel(
        property: CatalogProperty,
        type: CatalogSubjectType,
        bound: Bool,
        locked: Bool
    ) -> String {
        var parts = [property.label]
        parts.append(bound
            ? L10n.string(L10n.Properties.bindingBound)
            : L10n.string(L10n.Properties.bindingNotBound))
        parts.append(type.label)
        if locked {
            parts.append(L10n.string(L10n.Properties.bindingLocked))
        }
        return parts.joined(separator: ", ")
    }

    private func boundToCell(for property: CatalogProperty) -> some View {
        let bound = model.snapshot.boundTypes(for: property.id)
        return Group {
            if bound.isEmpty {
                Text(verbatim: "—")
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textFaint)
            } else {
                HStack(spacing: 4) {
                    ForEach(bound.prefix(3)) { type in
                        PropertiesBoundChip(
                            label: type.label,
                            locked: model.bindingLocked(propertyID: property.id, typeID: type.id),
                            emphasized: model.selectedTypeKey == type.key
                        )
                    }
                    if bound.count > 3 {
                        Text(verbatim: L10n.Properties.boundOverflow(count: bound.count - 3))
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textFaint)
                    }
                }
            }
        }
    }

    private func rowAccessibilityLabel(_ property: CatalogProperty) -> String {
        let bound = model.snapshot.boundTypes(for: property.id)
        let valueType = L10n.string(SubjectPropertyValueType.label(property.valueType))
        let boundText = L10n.Properties.rowBoundAnnouncement(count: bound.count)
        return "\(property.label), \(valueType), \(property.origin), \(boundText)"
    }

    private var inspectorCard: some View {
        PVCard {
            Group {
                if let property = model.selectedProperty {
                    ScrollView {
                        inspectorBody(property)
                            .padding(16)
                    }
                } else {
                    Text(L10n.Properties.inspectorEmpty)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textMuted)
                        .padding(16)
                        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                }
            }
        }
        .accessibilityLabel(Text(L10n.Properties.inspectorAccessibility))
    }

    private func inspectorBody(_ property: CatalogProperty) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PropertiesIdentityEditor(property: property, model: model)
                .id(property.id)

            if model.deleteImpact.request == nil, let deleteError = model.deleteError {
                PVCallout(tone: .danger, message: deleteError)
                    .accessibilityIdentifier("properties.deleteError")
            }

            VStack(spacing: PVSpacing.space4) {
                inspectorMetaRow(label: L10n.Properties.inspectorValueType) {
                    HStack(spacing: 6) {
                        PropertiesValueTypePill(valueType: property.valueType)
                        Image(systemName: "lock.fill")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundStyle(PVColor.textFaint)
                            .accessibilityLabel(Text(L10n.Properties.valueTypeImmutable))
                    }
                }
                inspectorMetaRow(label: L10n.Properties.inspectorOrigin) {
                    Group {
                        switch property.origin {
                        case CatalogOrigin.user:
                            Text(L10n.Properties.inspectorOriginUser)
                        case CatalogOrigin.provenencia:
                            Text(L10n.Properties.inspectorOriginSeeded)
                        default:
                            Text(verbatim: property.origin)
                        }
                    }
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
                }
                inspectorMetaRow(label: L10n.Properties.inspectorUsedOn) {
                    Text(verbatim: L10n.Properties.usedOnCount(count: property.usedBy))
                        .font(PVFont.mono(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                }
            }

            if property.valueType == "term" {
                Text(L10n.Properties.termNote)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if let lockedCallout = model.lockedCallout {
                PVCallout(tone: .info, message: lockedCallout)
            }
            if let formError = model.formError, !model.isEditingIdentity {
                PVCallout(tone: .danger, message: formError)
            }

            PVDivider()

            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                HStack(alignment: .firstTextBaseline) {
                    Text(L10n.Properties.bindingsSection)
                        .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textMuted)
                        .tracking(PVTypeScale.micro * PVTracking.caps)
                        .textCase(.uppercase)
                    Spacer()
                    Text(verbatim: L10n.Properties.bindCount(
                        bound: model.snapshot.boundTypes(for: property.id).count,
                        total: model.types.count
                    ))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textFaint)
                }

                VStack(spacing: 2) {
                    ForEach(model.types) { type in
                        bindingRow(property: property, type: type)
                    }
                }
                .accessibilityElement(children: .contain)
                .accessibilityLabel(Text(L10n.Properties.bindingsAccessibility))
            }

        }
    }

    private func inspectorMetaRow<Content: View>(
        label: LocalizedStringResource,
        @ViewBuilder content: () -> Content
    ) -> some View {
        HStack(alignment: .firstTextBaseline) {
            Text(label)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
            Spacer(minLength: PVSpacing.space5)
            content()
        }
    }

    private func bindingRow(property: CatalogProperty, type: CatalogSubjectType) -> some View {
        let bound = model.isBound(propertyID: property.id, typeID: type.id)
        let locked = model.bindingLocked(propertyID: property.id, typeID: type.id)
        let presentation = model.snapshot.presentation(for: type)
        let ink = PropertiesTypeChrome.ink(typeKey: type.key, presentation: presentation)
        return Button {
            Task { await model.toggleBinding(to: type) }
        } label: {
            HStack(spacing: PVSpacing.space5) {
                PropertiesBindBox(on: bound, locked: locked)
                if let mark = PropertiesTypeChrome.stripMarkKey(typeKey: type.key) {
                    PVMark(mark, size: 14)
                        .foregroundStyle(ink)
                }
                Text(verbatim: type.label)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(locked ? PVColor.textSecondary : PVColor.textPrimary)
                    .frame(maxWidth: .infinity, alignment: .leading)
                if locked {
                    Text(L10n.Properties.bindingRegistry)
                        .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            .padding(.horizontal, 8)
            .frame(height: 30)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityLabel(
            "\(type.label), \(bound ? L10n.string(L10n.Properties.bindingBound) : L10n.string(L10n.Properties.bindingNotBound))"
            + (locked ? ", \(L10n.string(L10n.Properties.bindingLocked))" : "")
        )
    }
}

/// Local drafts so inspector keystrokes don't rebuild the type strip and table.
private struct PropertiesIdentityEditor: View {
    let property: CatalogProperty
    @Bindable var model: PropertiesModel
    @State private var labelDraft = ""
    @State private var descriptionDraft = ""

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            PVInlineEdit(
                isEditing: model.isEditingIdentity,
                isSaving: model.isSaving,
                error: model.isEditingIdentity ? model.formError : nil,
                saveLabel: L10n.Properties.editSave,
                cancelLabel: L10n.Properties.editCancel,
                editLabel: L10n.Properties.editAction,
                saveDisabled: !model.canSubmitEdit(label: labelDraft, description: descriptionDraft),
                showsEditControl: model.canEditSelected,
                axis: .horizontal,
                actionsStyle: .iconRow,
                accessibilityIdentifierPrefix: "properties.identity",
                onEdit: {
                    seedDrafts()
                    model.beginEdit()
                },
                onSave: {
                    Task { _ = await model.submitEdit(label: labelDraft, description: descriptionDraft) }
                },
                onCancel: { model.cancelEdit() },
                display: {
                    VStack(alignment: .leading, spacing: 3) {
                        Text(verbatim: property.label)
                            .font(PVFont.display(size: PVTypeScale.h3))
                            .foregroundStyle(PVColor.textDisplay)
                        Text(verbatim: property.key)
                            .font(PVFont.mono(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                    }
                },
                editor: {
                    VStack(alignment: .leading, spacing: 3) {
                        PVInput(
                            text: $labelDraft,
                            size: .sm,
                            isInvalid: labelDraft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
                            activateOnAppear: true
                        )
                        .accessibilityIdentifier("properties.identity.label")
                        Text(verbatim: property.key)
                            .font(PVFont.mono(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                    }
                },
                restingTrailing: {
                    if model.showsDelete {
                        PVIconButton(
                            .trash,
                            label: model.deleteTooltip,
                            size: .sm,
                            tone: .danger
                        ) {
                            Task { await model.askDelete() }
                        }
                        .disabled(!model.canDeleteSelected)
                        .accessibilityLabel(Text(verbatim: model.deleteAccessibilityLabel))
                        .accessibilityIdentifier("properties.delete")
                    }
                },
                editingTrailing: { EmptyView() }
            )
            if model.isEditingIdentity {
                PVTextArea(
                    text: $descriptionDraft,
                    lineLimit: 2...6,
                    prompt: L10n.Properties.editDescriptionPlaceholder
                )
                .accessibilityIdentifier("properties.identity.description")
                Text(L10n.Properties.editKeyStays(key: property.key))
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            } else if !property.description.isEmpty {
                Text(verbatim: property.description)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
        .onChange(of: model.isEditingIdentity) { _, editing in
            if editing { seedDrafts() }
        }
    }

    private func seedDrafts() {
        labelDraft = property.label
        descriptionDraft = property.description
    }
}

/// Local create drafts so dialog keystrokes don't rebuild the destination.
private struct PropertiesCreateHost: View {
    @Bindable var model: PropertiesModel
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var draft: PropertiesModel.Draft?

    var body: some View {
        Color.clear
            .frame(width: 0, height: 0)
            .accessibilityHidden(true)
            .pvFormDialog(
                isPresented: createOpenBinding,
                copy: PVFormDialogCopy(
                    title: L10n.Properties.createTitle,
                    subtitle: L10n.Properties.createOriginNote,
                    confirm: L10n.Properties.createSubmit,
                    cancel: L10n.Properties.createCancel
                ),
                isRunning: model.isSaving,
                confirmDisabled: !model.canSubmitCreate(draft),
                accessibilityIdentifierPrefix: "properties.create",
                onConfirm: {
                    Task {
                        if let location = await model.submitCreate(draft) {
                            navigation.go(to: location)
                        }
                    }
                }
            ) {
                createForm
            }
            .onAppear {
                if model.createOpen { draft = model.draft }
            }
            .onChange(of: model.createOpen) { _, open in
                draft = open ? model.draft : nil
            }
    }

    private var createOpenBinding: Binding<Bool> {
        Binding(
            get: { model.createOpen },
            set: { if !$0 { model.closeCreate() } }
        )
    }

    @ViewBuilder
    private var createForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let formError = model.formError {
                PVCallout(tone: .danger, message: formError)
                    .accessibilityIdentifier("properties.create.error")
            }

            PVField(
                label: L10n.Properties.createLabel,
                hint: L10n.Properties.createLabelHint,
                required: true
            ) {
                PVInput(
                    text: Binding(
                        get: { draft?.label ?? "" },
                        set: { draft?.label = $0 }
                    ),
                    prompt: L10n.Properties.createLabelPlaceholder
                )
                .accessibilityIdentifier("properties.create.label")
            }

            PVField(
                label: L10n.Properties.createKey,
                hint: L10n.Properties.createKeyHint
            ) {
                PVInput(
                    text: Binding(
                        get: { PropertiesModel.draftKey(label: draft?.label ?? "") },
                        set: { _ in }
                    ),
                    mono: true,
                    isReadOnly: true,
                    prompt: L10n.Properties.createKeyPlaceholder
                )
                .accessibilityIdentifier("properties.create.key")
            }

            PVField(label: L10n.Properties.createDescription) {
                PVTextArea(
                    text: Binding(
                        get: { draft?.description ?? "" },
                        set: { draft?.description = $0 }
                    ),
                    lineLimit: 2...4,
                    prompt: L10n.Properties.createDescriptionPlaceholder
                )
                .accessibilityIdentifier("properties.create.description")
            }

            PVField(
                label: L10n.Properties.createValueType,
                hint: L10n.Properties.createValueTypeHint,
                required: true
            ) {
                PVChipGroup(style: .segmented) {
                    ForEach(SubjectPropertyValueType.researcherCreatable, id: \.self) { vt in
                        PVChip(
                            text: vt,
                            isSelected: (draft?.valueType ?? "text") == vt,
                            expands: true,
                            selectionLift: true,
                            action: { draft?.valueType = vt }
                        )
                        .accessibilityIdentifier("properties.create.valueType.\(vt)")
                    }
                }
                .accessibilityLabel(Text(L10n.Properties.createValueType))
            }

            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                Text(L10n.Properties.createBindSection)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textMuted)
                    .tracking(PVTypeScale.micro * PVTracking.caps)
                    .textCase(.uppercase)

                LazyVGrid(
                    columns: [
                        GridItem(.flexible(), spacing: 2),
                        GridItem(.flexible(), spacing: 2),
                    ],
                    spacing: 2
                ) {
                    ForEach(model.types) { type in
                        createBindRow(type)
                    }
                }
                .accessibilityElement(children: .contain)
                .accessibilityLabel(Text(L10n.Properties.createBindSection))
            }
        }
    }

    private func createBindRow(_ type: CatalogSubjectType) -> some View {
        let presentation = model.snapshot.presentation(for: type)
        let ink = PropertiesTypeChrome.ink(typeKey: type.key, presentation: presentation)
        let on = draft?.bindTypeIDs.contains(type.id) == true
        return Button {
            toggleBind(typeID: type.id)
        } label: {
            HStack(spacing: PVSpacing.space5) {
                PropertiesBindBox(on: on, locked: false)
                if let mark = PropertiesTypeChrome.stripMarkKey(typeKey: type.key) {
                    PVMark(mark, size: 14)
                        .foregroundStyle(ink)
                }
                Text(verbatim: type.label)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            .padding(.horizontal, 8)
            .frame(height: 30)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityLabel(type.label)
        .accessibilityAddTraits(on ? [.isSelected] : [])
        .accessibilityIdentifier("properties.create.bind.\(type.key)")
    }

    private func toggleBind(typeID: String) {
        guard draft != nil else { return }
        if draft!.bindTypeIDs.contains(typeID) {
            draft!.bindTypeIDs.remove(typeID)
        } else {
            draft!.bindTypeIDs.insert(typeID)
        }
    }
}
