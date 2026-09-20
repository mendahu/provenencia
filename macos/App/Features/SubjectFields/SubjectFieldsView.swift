import SwiftUI

/// Subject fields workspace destination (S7-D2 / S7-05): type strip + dual cards
/// (property table + inspector). Chrome matches the board HTML.
struct SubjectFieldsView: View {
    @Environment(WorkspaceSession.self) private var session
    @State private var model: SubjectFieldsModel

    private let inspectorWidth: CGFloat = PVSpacing.widthInspector

    init(
        session: WorkspaceSession,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        _model = State(
            initialValue: SubjectFieldsModel(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        )
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<SubjectFieldsSnapshot> = session.queryHandle(
                SubjectFieldsModel.workspaceKey(for: session)
            ) {
                SubjectFieldsContent(handle: handle, model: model, inspectorWidth: inspectorWidth)
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

private struct SubjectFieldsContent: View {
    @Bindable var handle: QueryHandle<SubjectFieldsSnapshot>
    @Bindable var model: SubjectFieldsModel
    let inspectorWidth: CGFloat
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
        .vocabularyToastOverlay($model.toast, identifier: "subjectFields.toast")
        .pvFormDialog(
            isPresented: createOpenBinding,
            copy: PVFormDialogCopy(
                title: L10n.SubjectFields.createTitle,
                subtitle: L10n.SubjectFields.createOriginNote,
                confirm: L10n.SubjectFields.createSubmit,
                cancel: L10n.SubjectFields.createCancel
            ),
            isRunning: model.isSaving,
            confirmDisabled: !model.canSubmitCreate,
            accessibilityIdentifierPrefix: "subjectFields.create",
            onConfirm: {
                Task { _ = await model.submitCreate() }
            }
        ) {
            createForm
        }
        .pvConfirm(
            item: pendingDelete,
            copy: { _ in
                PVConfirmCopy(
                    title: String(localized: L10n.SubjectFields.deleteConfirmTitle),
                    message: String(localized: L10n.SubjectFields.deleteConfirmMessage),
                    confirm: L10n.SubjectFields.deleteProperty,
                    cancel: L10n.SubjectFields.deleteKeep
                )
            },
            isRunning: model.isDeleting,
            accessibilityIdentifierPrefix: "subjectFields.delete",
            onConfirm: {
                Task { _ = await model.confirmDelete() }
            }
        ) { property in
            if let deleteError = model.deleteError {
                PVCallout(tone: .danger, message: deleteError)
            }
            Text(verbatim: property.key)
                .font(PVFont.mono())
                .foregroundStyle(PVColor.textMuted)
        }
        .onChange(of: handle.status) { _, status in
            if status == .ready { model.syncCatalogCounts() }
        }
        .onChange(of: model.searchFocused) { _, focused in
            if focused { searchFocused = true }
        }
        .onChange(of: searchFocused) { _, focused in
            model.searchFocused = focused
        }
        .onChange(of: model.addPropertySelection) { _, value in
            guard !value.isEmpty else { return }
            Task { await model.addPropertyFromCombo() }
        }
        .onAppear {
            if handle.status == .ready { model.syncCatalogCounts() }
        }
        .background {
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
        .accessibilityIdentifier("subjectFields")
    }

    private var createOpenBinding: Binding<Bool> {
        Binding(
            get: { model.createOpen },
            set: { if !$0 { model.closeCreate() } }
        )
    }

    private var pendingDelete: Binding<CatalogProperty?> {
        Binding(
            get: { model.pendingDeleteProperty },
            set: { if $0 == nil { model.cancelDelete() } }
        )
    }

    private var header: some View {
        HStack(alignment: .bottom, spacing: PVSpacing.space6) {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.Workspace.subjectFieldsTitle)
                    .font(PVFont.display(size: PVTypeScale.h1))
                    .foregroundStyle(PVColor.textDisplay)
                Text(L10n.SubjectFields.description)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: PVSpacing.measureProse, alignment: .leading)
            }
            Spacer(minLength: PVSpacing.space6)
            PVButton(L10n.SubjectFields.newProperty, variant: .primary, size: .sm, icon: .plus) {
                model.openCreate()
            }
            .accessibilityIdentifier("subjectFields.add")
        }
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.top, PVSpacing.space8)
        .padding(.bottom, PVSpacing.space6)
    }

    private var typeStrip: some View {
        HStack(spacing: PVSpacing.space4) {
            typeCard(
                key: nil,
                title: Text(L10n.SubjectFields.allProperties),
                count: model.snapshot.properties.count,
                selected: model.selectedTypeKey == nil,
                bridge: false,
                ink: PVColor.textPrimary,
                icon: nil
            )
            ForEach(model.types) { type in
                let presentation = model.snapshot.presentation(for: type)
                typeCard(
                    key: type.key,
                    title: Text(verbatim: type.label),
                    count: model.snapshot.propertyCount(forTypeID: type.id),
                    selected: model.selectedTypeKey == type.key,
                    bridge: SubjectFieldsTypeChrome.showsBridgeLabel(typeKey: type.key),
                    ink: SubjectFieldsTypeChrome.ink(typeKey: type.key, presentation: presentation),
                    icon: SubjectFieldsTypeChrome.stripIconKind(typeKey: type.key)
                )
            }
        }
        .frame(maxWidth: .infinity)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.SubjectFields.typeStripAccessibility))
    }

    private func typeCard(
        key: String?,
        title: Text,
        count: Int,
        selected: Bool,
        bridge: Bool,
        ink: Color,
        icon: PVSubjectIconKind?
    ) -> some View {
        Button {
            model.selectType(key)
        } label: {
            VStack(alignment: .leading, spacing: 6) {
                HStack(alignment: .top) {
                    if let icon {
                        PVSubjectIcon(kind: icon, size: 16)
                            .foregroundStyle(ink)
                    } else {
                        Image(systemName: "square.grid.2x2")
                            .font(.system(size: 14, weight: .medium))
                            .foregroundStyle(PVColor.textPrimary)
                            .frame(width: 16, height: 16)
                    }
                    Spacer(minLength: 0)
                    if bridge {
                        Text(L10n.SubjectFields.bridgeRole)
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
                Text(verbatim: L10n.SubjectFields.stripFieldCount(count: count))
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
                prompt: L10n.SubjectFields.searchPlaceholder,
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
                    placeholder: L10n.SubjectFields.addPropertyPlaceholder(typeLabel: type.label),
                    emptyLabel: L10n.SubjectFields.addPropertyEmpty,
                    label: L10n.SubjectFields.addPropertyPlaceholder(typeLabel: type.label),
                    accessibilityIdentifierPrefix: "subjectFields.addCombo"
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
                    Text(verbatim: L10n.SubjectFields.emptySearchTitle(query: model.searchQuery))
                        .font(PVFont.display(size: PVTypeScale.h4))
                        .foregroundStyle(PVColor.textSecondary)
                    Text(L10n.SubjectFields.emptySearch)
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
                        set: { model.selectProperty($0) }
                    ),
                    primaryText: { $0.label },
                    density: .compact,
                    label: L10n.Workspace.subjectFieldsTitle,
                    rowAccessibilityIdentifier: { "subjectFields.row.\($0.id)" },
                    content: { EmptyView() }
                )
            }
        }
    }

    private var tableColumns: [PVTableColumn<CatalogProperty>] {
        var columns: [PVTableColumn<CatalogProperty>] = []
        if model.selectedType != nil {
            columns.append(
                PVTableColumn(id: "on", title: L10n.SubjectFields.columnOn, width: 36) { property in
                    onCell(for: property)
                }
            )
        }
        columns.append(contentsOf: [
            PVTableColumn(id: "property", title: L10n.SubjectFields.columnProperty) { property in
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
            PVTableColumn(id: "valueType", title: L10n.SubjectFields.columnValueType, width: 96) { property in
                SubjectFieldsValueTypePill(valueType: property.valueType)
            },
            PVTableColumn(id: "origin", title: L10n.SubjectFields.columnOrigin, width: 76) { property in
                SubjectFieldsOriginCell(origin: property.origin)
            },
            PVTableColumn(id: "boundTo", title: L10n.SubjectFields.columnBoundTo, width: 226) { property in
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
                model.selectProperty(property.id)
                Task { await model.toggleBinding(to: type) }
            } label: {
                SubjectFieldsBindBox(on: bound, locked: locked)
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
            ? String(localized: L10n.SubjectFields.bindingBound)
            : String(localized: L10n.SubjectFields.bindingNotBound))
        parts.append(type.label)
        if locked {
            parts.append(String(localized: L10n.SubjectFields.bindingLocked))
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
                        SubjectFieldsBoundChip(
                            label: type.label,
                            locked: model.bindingLocked(propertyID: property.id, typeID: type.id),
                            emphasized: model.selectedTypeKey == type.key
                        )
                    }
                    if bound.count > 3 {
                        Text(verbatim: L10n.SubjectFields.boundOverflow(count: bound.count - 3))
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textFaint)
                    }
                }
            }
        }
    }

    private func rowAccessibilityLabel(_ property: CatalogProperty) -> String {
        let bound = model.snapshot.boundTypes(for: property.id)
        let valueType = String(localized: SubjectPropertyValueType.label(property.valueType))
        let boundText = L10n.SubjectFields.rowBoundAnnouncement(count: bound.count)
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
                    Text(L10n.SubjectFields.inspectorEmpty)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textMuted)
                        .padding(16)
                        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                }
            }
        }
        .accessibilityLabel(Text(L10n.SubjectFields.inspectorAccessibility))
    }

    private func inspectorBody(_ property: CatalogProperty) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            VStack(alignment: .leading, spacing: 3) {
                Text(verbatim: property.label)
                    .font(PVFont.display(size: PVTypeScale.h3))
                    .foregroundStyle(PVColor.textDisplay)
                Text(verbatim: property.key)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }

            if !property.description.isEmpty {
                Text(verbatim: property.description)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
                    .fixedSize(horizontal: false, vertical: true)
            }

            VStack(spacing: PVSpacing.space4) {
                inspectorMetaRow(label: L10n.SubjectFields.inspectorValueType) {
                    HStack(spacing: 6) {
                        SubjectFieldsValueTypePill(valueType: property.valueType)
                        Image(systemName: "lock.fill")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundStyle(PVColor.textFaint)
                            .accessibilityLabel(Text(L10n.SubjectFields.valueTypeImmutable))
                    }
                }
                inspectorMetaRow(label: L10n.SubjectFields.inspectorOrigin) {
                    Group {
                        switch property.origin {
                        case CatalogOrigin.user:
                            Text(L10n.SubjectFields.inspectorOriginUser)
                        case CatalogOrigin.provenencia:
                            Text(L10n.SubjectFields.inspectorOriginSeeded)
                        default:
                            Text(verbatim: property.origin)
                        }
                    }
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
                }
                inspectorMetaRow(label: L10n.SubjectFields.inspectorValuesRecorded) {
                    Text(verbatim: L10n.SubjectFields.valuesRecordedCount(count: property.usedBy))
                        .font(PVFont.mono(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                }
            }

            if property.valueType == "term" {
                Text(L10n.SubjectFields.termNote)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if let lockedCallout = model.lockedCallout {
                PVCallout(tone: .info, message: lockedCallout)
            }
            if let formError = model.formError {
                PVCallout(tone: .danger, message: formError)
            }

            PVDivider()

            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                HStack(alignment: .firstTextBaseline) {
                    Text(L10n.SubjectFields.bindingsSection)
                        .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textMuted)
                        .tracking(PVTypeScale.micro * PVTracking.caps)
                        .textCase(.uppercase)
                    Spacer()
                    Text(verbatim: L10n.SubjectFields.bindCount(
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
                .accessibilityLabel(Text(L10n.SubjectFields.bindingsAccessibility))
            }

            PVDivider()

            deleteControls(for: property)
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
        let ink = SubjectFieldsTypeChrome.ink(typeKey: type.key, presentation: presentation)
        return Button {
            Task { await model.toggleBinding(to: type) }
        } label: {
            HStack(spacing: PVSpacing.space5) {
                SubjectFieldsBindBox(on: bound, locked: locked)
                if let kind = SubjectFieldsTypeChrome.stripIconKind(typeKey: type.key) {
                    PVSubjectIcon(kind: kind, size: 14)
                        .foregroundStyle(ink)
                }
                Text(verbatim: type.label)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(locked ? PVColor.textSecondary : PVColor.textPrimary)
                    .frame(maxWidth: .infinity, alignment: .leading)
                if locked {
                    Text(L10n.SubjectFields.bindingRegistry)
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
            "\(type.label), \(bound ? String(localized: L10n.SubjectFields.bindingBound) : String(localized: L10n.SubjectFields.bindingNotBound))"
            + (locked ? ", \(String(localized: L10n.SubjectFields.bindingLocked))" : "")
        )
    }

    @ViewBuilder
    private func deleteControls(for property: CatalogProperty) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            PVButton(
                L10n.SubjectFields.deleteProperty,
                variant: .danger,
                size: .sm,
                icon: .trash
            ) {
                model.askDelete()
            }
            .disabled(!model.canDeleteSelected)
            .frame(maxWidth: .infinity)

            Text(deleteNote(for: property))
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    private func deleteNote(for property: CatalogProperty) -> LocalizedStringResource {
        if property.origin != CatalogOrigin.user {
            return L10n.SubjectFields.deleteSeeded
        }
        if property.usedBy > 0 {
            return L10n.SubjectFields.deleteInUse
        }
        return L10n.SubjectFields.deleteUnused
    }

    private var createForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let formError = model.formError {
                PVCallout(tone: .danger, message: formError)
                    .accessibilityIdentifier("subjectFields.create.error")
            }

            PVField(
                label: L10n.SubjectFields.createLabel,
                hint: L10n.SubjectFields.createLabelHint,
                required: true
            ) {
                PVInput(
                    text: Binding(
                        get: { model.draft?.label ?? "" },
                        set: { model.draft?.label = $0 }
                    ),
                    prompt: L10n.SubjectFields.createLabelPlaceholder
                )
                .accessibilityIdentifier("subjectFields.create.label")
            }

            PVField(
                label: L10n.SubjectFields.createKey,
                hint: L10n.SubjectFields.createKeyHint
            ) {
                PVInput(
                    text: Binding(
                        get: { model.draftKey },
                        set: { _ in }
                    ),
                    mono: true,
                    isReadOnly: true,
                    prompt: L10n.SubjectFields.createKeyPlaceholder
                )
                .accessibilityIdentifier("subjectFields.create.key")
            }

            PVField(label: L10n.SubjectFields.createDescription) {
                PVTextArea(
                    text: Binding(
                        get: { model.draft?.description ?? "" },
                        set: { model.draft?.description = $0 }
                    ),
                    lineLimit: 2...4,
                    prompt: L10n.SubjectFields.createDescriptionPlaceholder
                )
                .accessibilityIdentifier("subjectFields.create.description")
            }

            PVField(
                label: L10n.SubjectFields.createValueType,
                hint: L10n.SubjectFields.createValueTypeHint,
                required: true
            ) {
                PVChipGroup(style: .segmented) {
                    ForEach(SubjectPropertyValueType.researcherCreatable, id: \.self) { vt in
                        PVChip(
                            text: vt,
                            isSelected: (model.draft?.valueType ?? "text") == vt,
                            expands: true,
                            selectionLift: true,
                            action: { model.draft?.valueType = vt }
                        )
                        .accessibilityIdentifier("subjectFields.create.valueType.\(vt)")
                    }
                }
                .accessibilityLabel(Text(L10n.SubjectFields.createValueType))
            }

            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                Text(L10n.SubjectFields.createBindSection)
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
                .accessibilityLabel(Text(L10n.SubjectFields.createBindSection))
            }
        }
    }

    private func createBindRow(_ type: CatalogSubjectType) -> some View {
        let presentation = model.snapshot.presentation(for: type)
        let ink = SubjectFieldsTypeChrome.ink(typeKey: type.key, presentation: presentation)
        let on = model.draft?.bindTypeIDs.contains(type.id) == true
        return Button {
            model.toggleDraftBind(typeID: type.id)
        } label: {
            HStack(spacing: PVSpacing.space5) {
                SubjectFieldsBindBox(on: on, locked: false)
                if let kind = SubjectFieldsTypeChrome.stripIconKind(typeKey: type.key) {
                    PVSubjectIcon(kind: kind, size: 14)
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
        .accessibilityIdentifier("subjectFields.create.bind.\(type.key)")
    }
}
