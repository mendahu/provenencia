import SwiftUI

// MARK: - Shared chrome (S7-D2)

/// Filled lock box — never a disabled checkbox (board Locked bindings).
struct SubjectFieldsLockBox: View {
    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: "lock.fill")
            Text(L10n.SubjectFields.bindingLocked)
        }
        .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
        .foregroundStyle(PVColor.textPrimary)
        .padding(.horizontal, 8)
        .padding(.vertical, 5)
        .background(PVColor.surfaceInset)
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm)
                .strokeBorder(PVColor.borderStrong, lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm))
        .accessibilityLabel(L10n.SubjectFields.bindingLocked)
    }
}

struct SubjectFieldsBoundChip: View {
    let type: CatalogSubjectType
    let presentation: CatalogSubjectTypePresentation?
    let locked: Bool

    var body: some View {
        HStack(spacing: 4) {
            if let kind = SubjectFieldsTypeChrome.iconKind(
                symbol: presentation?.iconSymbol,
                typeKey: type.key
            ) {
                PVSubjectIcon(kind: kind, size: 11)
                    .foregroundStyle(SubjectFieldsTypeChrome.ink(for: presentation))
            }
            Text(verbatim: type.label)
                .font(PVFont.body(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textSecondary)
                .lineLimit(1)
            if locked {
                Image(systemName: "lock.fill")
                    .font(.system(size: 9, weight: .semibold))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .padding(.horizontal, 7)
        .padding(.vertical, 3)
        .background(PVColor.surfaceInset)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm))
    }
}

struct SubjectFieldsValueTypePill: View {
    let valueType: String

    var body: some View {
        PVBadge(SubjectPropertyValueType.label(valueType), tone: .neutral, subtle: true)
    }
}

// MARK: - Destination

/// Subject fields workspace destination (S7-D2 / S7-05): type strip + property table + inspector.
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
            PVDivider()
            if let loadError = model.loadError, model.snapshot.properties.isEmpty {
                PVCallout(tone: .danger, message: L10n.Errors.message(for: loadError))
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.top, PVSpacing.space8)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            } else {
                typeStrip
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.vertical, PVSpacing.space6)
                toolbar
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.bottom, PVSpacing.space6)
                PVDivider()
                HStack(spacing: 0) {
                    propertyTable
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                    PVDivider(axis: .vertical)
                    inspector
                        .frame(width: inspectorWidth)
                        .frame(maxHeight: .infinity)
                }
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .vocabularyToastOverlay($model.toast, identifier: "subjectFields.toast")
        .sheet(isPresented: createOpenBinding) {
            createSheet
        }
        .pvConfirmSheet(
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
        HStack(alignment: .bottom, spacing: PVSpacing.space8) {
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
            HStack(spacing: PVSpacing.space6) {
                Text(verbatim: model.countLine)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("subjectFields.countLine")
                PVButton(L10n.SubjectFields.newProperty, variant: .primary, icon: .plus) {
                    model.openCreate()
                }
                .accessibilityIdentifier("subjectFields.add")
            }
        }
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.top, PVSpacing.space8)
        .padding(.bottom, PVSpacing.space6)
    }

    private var typeStrip: some View {
        HStack(spacing: PVSpacing.space3) {
            typeCard(
                key: nil,
                title: Text(L10n.SubjectFields.allProperties),
                count: model.snapshot.properties.count,
                presentation: nil,
                selected: model.selectedTypeKey == nil
            )
            ForEach(model.types) { type in
                typeCard(
                    key: type.key,
                    title: Text(verbatim: type.label),
                    count: model.snapshot.propertyCount(forTypeID: type.id),
                    presentation: model.snapshot.presentation(for: type),
                    selected: model.selectedTypeKey == type.key
                )
            }
        }
        .frame(maxWidth: .infinity)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.SubjectFields.typeStripAccessibility)
    }

    private func typeCard(
        key: String?,
        title: Text,
        count: Int,
        presentation: CatalogSubjectTypePresentation?,
        selected: Bool
    ) -> some View {
        let ink = SubjectFieldsTypeChrome.ink(for: presentation)
        let bridge = SubjectFieldsTypeChrome.isBridge(role: presentation?.role)
        return Button {
            model.selectType(key)
        } label: {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                HStack(alignment: .top) {
                    if let kind = SubjectFieldsTypeChrome.iconKind(
                        symbol: presentation?.iconSymbol,
                        typeKey: key ?? ""
                    ) {
                        PVSubjectIcon(kind: kind, size: 22)
                            .foregroundStyle(ink)
                    } else if key == nil {
                        Image(systemName: "square.grid.2x2")
                            .font(.system(size: 16, weight: .medium))
                            .foregroundStyle(PVColor.textPrimary)
                            .frame(width: 22, height: 22)
                    }
                    Spacer(minLength: 0)
                    if bridge {
                        Text(L10n.SubjectFields.bridgeRole)
                            .font(PVFont.body(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textMuted)
                    }
                }
                title
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                    .minimumScaleFactor(0.85)
                Text(verbatim: L10n.SubjectFields.stripFieldCount(count: count))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
            }
            .padding(PVSpacing.space5)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(selected ? PVColor.surfaceSelected : PVColor.surfaceRaised)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md)
                    .strokeBorder(selected ? PVColor.accentLine : PVColor.borderSubtle, lineWidth: selected ? 2 : 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md))
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(selected ? [.isSelected, .isButton] : .isButton)
    }

    private var toolbar: some View {
        HStack(spacing: PVSpacing.space4) {
            PVInput(
                text: $model.searchQuery,
                size: .sm,
                prompt: L10n.SubjectFields.searchPlaceholder,
                icon: .search,
                focused: $searchFocused
            )
            .frame(maxWidth: 280)

            Picker(selection: Binding(
                get: { model.originFilter },
                set: { model.setOriginFilter($0) }
            )) {
                ForEach(SubjectFieldsOriginFilter.allCases) { filter in
                    Text(filter.label).tag(filter)
                }
            } label: { EmptyView() }
            .pickerStyle(.segmented)
            .frame(maxWidth: 240)

            Picker(selection: Binding(
                get: { model.valueTypeFilter ?? "" },
                set: { model.setValueTypeFilter($0.isEmpty ? nil : $0) }
            )) {
                Text(L10n.SubjectFields.valueTypeAny).tag("")
                ForEach(SubjectPropertyValueType.researcherCreatable + ["term"], id: \.self) { vt in
                    Text(SubjectPropertyValueType.label(vt)).tag(vt)
                }
            } label: { EmptyView() }
            .frame(maxWidth: 160)

            Spacer(minLength: 0)

            if let type = model.selectedType {
                Button {
                    model.openCreate()
                } label: {
                    Text(verbatim: L10n.SubjectFields.addToType(typeLabel: type.label))
                }
                .buttonStyle(.bordered)
            }
        }
    }

    private var propertyTable: some View {
        Group {
            if model.visibleProperties.isEmpty {
                VStack(alignment: .leading, spacing: PVSpacing.space3) {
                    Text(L10n.SubjectFields.emptySearchTitle)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                    Text(L10n.SubjectFields.emptySearch)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                .padding(PVSpacing.gutterPage)
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
        [
            PVTableColumn(id: "on", title: L10n.SubjectFields.columnOn, width: 72) { property in
                onCell(for: property)
            },
            PVTableColumn(id: "property", title: L10n.SubjectFields.columnProperty) { property in
                VStack(alignment: .leading, spacing: 2) {
                    Text(verbatim: property.label)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    Text(verbatim: property.key)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(rowAccessibilityLabel(property))
            },
            PVTableColumn(id: "valueType", title: L10n.SubjectFields.columnValueType, width: 100) { property in
                SubjectFieldsValueTypePill(valueType: property.valueType)
            },
            PVTableColumn(id: "origin", title: L10n.SubjectFields.columnOrigin, width: 88) { property in
                if property.origin == CatalogOrigin.provenencia {
                    OriginPill(origin: property.origin)
                } else if property.origin == CatalogOrigin.user {
                    Text(L10n.SubjectFields.originUserShort)
                        .font(PVFont.body(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                } else {
                    OriginPill(origin: property.origin)
                }
            },
            PVTableColumn(id: "boundTo", title: L10n.SubjectFields.columnBoundTo, width: 200) { property in
                boundToCell(for: property)
            },
        ]
    }

    @ViewBuilder
    private func onCell(for property: CatalogProperty) -> some View {
        if let type = model.selectedType {
            let locked = model.bindingLocked(propertyID: property.id, typeID: type.id)
            let bound = model.isBound(propertyID: property.id, typeID: type.id)
            if locked {
                Button {
                    model.selectProperty(property.id)
                    Task { await model.toggleBinding(to: type) }
                } label: {
                    SubjectFieldsLockBox()
                }
                .buttonStyle(.plain)
            } else {
                Toggle(
                    isOn: Binding(
                        get: { bound },
                        set: { _ in
                            model.selectProperty(property.id)
                            Task { await model.toggleBinding(to: type) }
                        }
                    )
                ) { EmptyView() }
                .toggleStyle(.checkbox)
                .labelsHidden()
            }
        } else {
            Text(verbatim: "—")
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textFaint)
        }
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
                            type: type,
                            presentation: model.snapshot.presentation(for: type),
                            locked: model.bindingLocked(propertyID: property.id, typeID: type.id)
                        )
                    }
                    if bound.count > 3 {
                        Text(verbatim: L10n.SubjectFields.boundOverflow(count: bound.count - 3))
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textMuted)
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

    private var inspector: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            if let property = model.selectedProperty {
                Text(L10n.SubjectFields.inspectorTitle)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textMuted)
                    .textCase(.uppercase)
                Text(verbatim: property.label)
                    .font(PVFont.display(size: PVTypeScale.h3))
                Text(verbatim: property.key)
                    .font(PVFont.mono())
                    .foregroundStyle(PVColor.textMuted)
                if !property.description.isEmpty {
                    Text(verbatim: property.description)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                }

                inspectorMetaRow(label: L10n.SubjectFields.inspectorValueType) {
                    SubjectFieldsValueTypePill(valueType: property.valueType)
                }
                inspectorMetaRow(label: L10n.SubjectFields.inspectorOrigin) {
                    OriginBadge(origin: property.origin)
                }
                inspectorMetaRow(label: L10n.SubjectFields.inspectorValuesRecorded) {
                    Text(verbatim: L10n.SubjectFields.valuesRecordedCount(count: property.usedBy))
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                }

                if property.valueType == "term" {
                    Text(L10n.SubjectFields.termNote)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                if let lockedCallout = model.lockedCallout {
                    PVCallout(tone: .warning, message: lockedCallout)
                }
                if let formError = model.formError {
                    PVCallout(tone: .danger, message: formError)
                }

                HStack {
                    Text(L10n.SubjectFields.bindingsSection)
                        .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textMuted)
                        .textCase(.uppercase)
                    Spacer()
                    Text(verbatim: L10n.SubjectFields.bindCount(count: model.snapshot.boundTypes(for: property.id).count))
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                }
                VStack(spacing: 2) {
                    ForEach(model.types) { type in
                        bindingRow(property: property, type: type)
                    }
                }
                Spacer(minLength: 0)
                deleteControls(for: property)
            } else {
                Text(L10n.SubjectFields.inspectorEmpty)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            }
        }
        .padding(PVSpacing.gutterPage)
        .background(PVColor.surfaceRaised)
    }

    private func inspectorMetaRow<Content: View>(
        label: LocalizedStringResource,
        @ViewBuilder content: () -> Content
    ) -> some View {
        HStack(alignment: .firstTextBaseline) {
            Text(label)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .frame(width: 110, alignment: .leading)
            content()
            Spacer(minLength: 0)
        }
    }

    private func bindingRow(property: CatalogProperty, type: CatalogSubjectType) -> some View {
        let bound = model.isBound(propertyID: property.id, typeID: type.id)
        let locked = model.bindingLocked(propertyID: property.id, typeID: type.id)
        let presentation = model.snapshot.presentation(for: type)
        return Button {
            Task { await model.toggleBinding(to: type) }
        } label: {
            HStack(spacing: PVSpacing.space3) {
                if let kind = SubjectFieldsTypeChrome.iconKind(
                    symbol: presentation?.iconSymbol,
                    typeKey: type.key
                ) {
                    PVSubjectIcon(kind: kind, size: 14)
                        .foregroundStyle(SubjectFieldsTypeChrome.ink(for: presentation))
                }
                VStack(alignment: .leading, spacing: 1) {
                    Text(verbatim: type.label)
                        .foregroundStyle(PVColor.textPrimary)
                    if SubjectFieldsTypeChrome.isBridge(role: presentation?.role) {
                        Text(L10n.SubjectFields.bridgeRole)
                            .font(PVFont.body(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textMuted)
                    }
                }
                Spacer()
                if locked {
                    SubjectFieldsLockBox()
                } else if bound {
                    Image(systemName: "checkmark")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundStyle(PVColor.accentSoftForeground)
                        .accessibilityLabel(L10n.SubjectFields.bindingBound)
                }
            }
            .padding(.vertical, 6)
            .padding(.horizontal, 8)
            .background(bound ? PVColor.surfaceSelected.opacity(0.5) : Color.clear)
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm))
        }
        .buttonStyle(.plain)
    }

    @ViewBuilder
    private func deleteControls(for property: CatalogProperty) -> some View {
        if property.origin != CatalogOrigin.user {
            PVCallout(tone: .info, message: String(localized: L10n.SubjectFields.deleteSeeded))
        } else if property.usedBy > 0 {
            PVCallout(tone: .warning, message: String(localized: L10n.SubjectFields.deleteInUse))
        } else {
            Button(role: .destructive) {
                model.askDelete()
            } label: {
                Text(L10n.SubjectFields.deleteProperty)
            }
        }
    }

    private var createSheet: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            HStack {
                Text(L10n.SubjectFields.createTitle)
                    .font(PVFont.display(size: PVTypeScale.h3))
                Spacer()
                Button {
                    model.closeCreate()
                } label: {
                    Image(systemName: "xmark")
                }
                .buttonStyle(.borderless)
            }
            Text(L10n.SubjectFields.createOriginNote)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.SubjectFields.createLabel)
                TextField(text: Binding(
                    get: { model.draft?.label ?? "" },
                    set: { model.draft?.label = $0 }
                ), prompt: Text(L10n.SubjectFields.createLabelHint)) { EmptyView() }
                .textFieldStyle(.roundedBorder)
                Text(verbatim: model.draftKey)
                    .font(PVFont.mono())
                    .foregroundStyle(PVColor.textMuted)
            }
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.SubjectFields.createValueType)
                Picker(selection: Binding(
                    get: { model.draft?.valueType ?? "text" },
                    set: { model.draft?.valueType = $0 }
                )) {
                    ForEach(SubjectPropertyValueType.researcherCreatable, id: \.self) { vt in
                        Text(SubjectPropertyValueType.label(vt)).tag(vt)
                    }
                } label: { EmptyView() }
            }
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.SubjectFields.createDescription)
                TextField(text: Binding(
                    get: { model.draft?.description ?? "" },
                    set: { model.draft?.description = $0 }
                ), prompt: Text(L10n.SubjectFields.createDescription), axis: .vertical) { EmptyView() }
                .lineLimit(3...6)
                .textFieldStyle(.roundedBorder)
            }
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                Text(L10n.SubjectFields.createBindSection)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textMuted)
                    .textCase(.uppercase)
                ForEach(model.types) { type in
                    let presentation = model.snapshot.presentation(for: type)
                    Toggle(isOn: Binding(
                        get: { model.draft?.bindTypeIDs.contains(type.id) == true },
                        set: { _ in model.toggleDraftBind(typeID: type.id) }
                    )) {
                        HStack(spacing: 8) {
                            if let kind = SubjectFieldsTypeChrome.iconKind(
                                symbol: presentation?.iconSymbol,
                                typeKey: type.key
                            ) {
                                PVSubjectIcon(kind: kind, size: 14)
                                    .foregroundStyle(SubjectFieldsTypeChrome.ink(for: presentation))
                            }
                            Text(verbatim: type.label)
                            if SubjectFieldsTypeChrome.isBridge(role: presentation?.role) {
                                Text(L10n.SubjectFields.bridgeRole)
                                    .font(PVFont.body(size: PVTypeScale.micro))
                                    .foregroundStyle(PVColor.textMuted)
                            }
                        }
                    }
                    .toggleStyle(.checkbox)
                }
            }
            if let formError = model.formError {
                PVCallout(tone: .danger, message: formError)
            }
            HStack {
                Spacer()
                Button(L10n.SubjectFields.createCancel) { model.closeCreate() }
                Button(L10n.SubjectFields.createSubmit) {
                    Task { _ = await model.submitCreate() }
                }
                .buttonStyle(.borderedProminent)
                .disabled(!model.canSubmitCreate || model.isSaving)
            }
        }
        .padding(PVSpacing.gutterPage)
        .frame(width: 440)
    }
}
