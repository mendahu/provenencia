import SwiftUI

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
                    .padding(.vertical, PVSpacing.space8)
                toolbar
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.bottom, PVSpacing.space8)
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
        .onAppear {
            if handle.status == .ready { model.syncCatalogCounts() }
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
        VocabularyHeader(
            title: L10n.Workspace.subjectFieldsTitle,
            description: L10n.SubjectFields.description,
            countLine: countLine,
            addLabel: L10n.SubjectFields.newProperty,
            isAddDisabled: false,
            identifierPrefix: "subjectFields",
            onAdd: { model.openCreate() }
        )
    }

    private var countLine: String {
        let summary = CatalogCountSummary.from(model.snapshot.properties)
        return L10n.SourceFields.countLine(
            total: summary.total,
            seeded: summary.seeded,
            user: summary.user
        )
    }

    private var typeStrip: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: PVSpacing.space4) {
                typeCard(
                    key: nil,
                    title: Text(L10n.SubjectFields.allProperties),
                    count: model.snapshot.properties.count,
                    selected: model.selectedTypeKey == nil
                )
                ForEach(model.types) { type in
                    typeCard(
                        key: type.key,
                        title: Text(verbatim: type.label),
                        count: model.snapshot.propertyCount(forTypeID: type.id),
                        selected: model.selectedTypeKey == type.key
                    )
                }
            }
        }
    }

    private func typeCard(
        key: String?,
        title: Text,
        count: Int,
        selected: Bool
    ) -> some View {
        Button {
            model.selectType(key)
        } label: {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                title
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                Text(verbatim: L10n.SubjectFields.stripFieldCount(count: count))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
            }
            .padding(PVSpacing.space5)
            .frame(minWidth: 120, alignment: .leading)
            .background(selected ? PVColor.surfaceSelected : PVColor.surfaceRaised)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md)
                    .strokeBorder(selected ? PVColor.accentLine : PVColor.borderSubtle, lineWidth: selected ? 2 : 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md))
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(selected ? .isSelected : [])
    }

    private var toolbar: some View {
        HStack(spacing: PVSpacing.space4) {
            TextField(
                text: $model.searchQuery,
                prompt: Text(L10n.SubjectFields.searchPlaceholder)
            ) { EmptyView() }
            .textFieldStyle(.roundedBorder)
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
                .buttonStyle(.borderedProminent)
                .keyboardShortcut("n", modifiers: .command)
            }
        }
    }

    private var propertyTable: some View {
        Group {
            if model.visibleProperties.isEmpty {
                Text(L10n.SubjectFields.emptySearch)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                    .padding(PVSpacing.gutterPage)
            } else {
                List(selection: Binding(
                    get: { model.selectedPropertyID },
                    set: { model.selectProperty($0) }
                )) {
                    Section {
                        ForEach(model.visibleProperties) { property in
                            propertyRow(property)
                                .tag(property.id)
                        }
                    } header: {
                        HStack {
                            Text(L10n.SubjectFields.columnProperty)
                                .frame(maxWidth: .infinity, alignment: .leading)
                            Text(L10n.SubjectFields.columnValueType)
                                .frame(width: 90, alignment: .leading)
                            Text(L10n.SubjectFields.columnOrigin)
                                .frame(width: 80, alignment: .leading)
                            Text(L10n.SubjectFields.columnBoundTo)
                                .frame(width: 140, alignment: .leading)
                        }
                        .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textMuted)
                        .textCase(nil)
                    }
                }
                .listStyle(.inset)
            }
        }
    }

    private func propertyRow(_ property: CatalogProperty) -> some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(verbatim: property.label)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                Text(verbatim: property.key)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            Text(SubjectPropertyValueType.label(property.valueType))
                .font(PVFont.body(size: PVTypeScale.caption))
                .frame(width: 90, alignment: .leading)
            OriginBadge(origin: property.origin)
                .frame(width: 80, alignment: .leading)
            Text(verbatim: boundSummary(for: property))
                .font(PVFont.body(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textSecondary)
                .frame(width: 140, alignment: .leading)
                .lineLimit(1)
        }
        .contentShape(Rectangle())
    }

    private func boundSummary(for property: CatalogProperty) -> String {
        let keys = model.snapshot.boundTypeKeys(for: property.id)
        if keys.isEmpty { return "—" }
        return keys.joined(separator: ", ")
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
                Text(L10n.SubjectFields.bindingsSection)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textMuted)
                    .textCase(.uppercase)
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

    private func bindingRow(property: CatalogProperty, type: CatalogSubjectType) -> some View {
        let bound = model.isBound(propertyID: property.id, typeID: type.id)
        let locked = model.bindingLocked(propertyID: property.id, typeID: type.id)
        return Button {
            Task { await model.toggleBinding(to: type) }
        } label: {
            HStack {
                Text(verbatim: type.label)
                    .foregroundStyle(PVColor.textPrimary)
                Spacer()
                if locked {
                    HStack(spacing: 4) {
                        Image(systemName: "lock.fill")
                        Text(L10n.SubjectFields.bindingLocked)
                    }
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(PVColor.surfaceSelected)
                    .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm))
                } else if bound {
                    Text(L10n.SubjectFields.bindingBound)
                        .font(PVFont.body(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.accentSoftForeground)
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
            Text(L10n.SubjectFields.deleteSeeded)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
        } else if property.usedBy > 0 {
            Text(L10n.SubjectFields.deleteInUse)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
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
        .frame(width: 420)
    }
}
