import SwiftUI

/// The **Sources** list workspace destination (S2-04 / S4-06): evidence list,
/// Add Source dialog, and row navigation to the Source page via history.
struct SourcesListView: View {
    @Environment(WorkspaceSession.self) private var session
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: SourcesModel

    init(
        session: WorkspaceSession,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        _model = State(
            initialValue: SourcesModel(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        )
    }

    var body: some View {
        Group {
            if let sourcesHandle: QueryHandle<[CatalogSource]> = session.queryHandle(
                SourcesModel.sourcesListKey(for: session)
            ),
            let typesHandle: QueryHandle<[CatalogSourceType]> = session.queryHandle(
                SourcesModel.sourceTypesListKey(for: session)
            ) {
                SourcesListContent(
                    sourcesHandle: sourcesHandle,
                    typesHandle: typesHandle,
                    model: model
                )
            } else {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .center)
            }
        }
        .task {
            model.warmListQueries()
        }
    }
}

/// Observes query handles directly so list loads repaint when cache settles.
private struct SourcesListContent: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Bindable var sourcesHandle: QueryHandle<[CatalogSource]>
    @Bindable var typesHandle: QueryHandle<[CatalogSourceType]>
    @Bindable var model: SourcesModel

    var body: some View {
        VStack(spacing: 0) {
            header
            PVDivider()
            if sourcesHandle.status == .loading && model.sources.isEmpty && model.loadError == nil {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .center)
                    .padding(.top, PVSpacing.space9)
            } else if let loadError = model.loadError, model.sources.isEmpty {
                PVCallout(tone: .danger, message: L10n.Errors.message(for: loadError))
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.top, PVSpacing.space8)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
                    .accessibilityIdentifier("sources.loadError")
            } else if model.isCatalogEmpty {
                emptyState
            } else {
                toolbar
                listBody
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .vocabularyToastOverlay($model.toast, identifier: "sources.toast")
        .pvFormDialog(
            isPresented: addPresented,
            copy: PVFormDialogCopy(
                title: L10n.Sources.addDialogTitle,
                subtitle: L10n.Sources.addDialogSubtitle,
                confirm: L10n.Sources.createAction,
                cancel: L10n.Sources.cancelAction
            ),
            isRunning: model.isSaving,
            accessibilityIdentifierPrefix: "sources.add",
            onConfirm: {
                Task {
                    if let created = await model.create() {
                        navigation.go(to: WorkspaceLocation(
                            section: .sources,
                            sourceId: created.id,
                            ref: created.ref,
                            title: created.title
                        ))
                    }
                }
            }
        ) {
            addForm(typeOptions: model.typeComboOptions)
        }
        .task(id: model.isAdding) {
            guard model.isAdding else { return }
            model.refreshTypes()
            model.selectSoleTypeIfNeeded()
        }
        .onChange(of: sourcesHandle.status) { _, status in
            if status == .ready { model.syncCatalogCounts() }
        }
        .onChange(of: typesHandle.status) { _, _ in
            if model.isAdding { model.selectSoleTypeIfNeeded() }
        }
        .onAppear {
            if sourcesHandle.status == .ready {
                model.syncCatalogCounts()
            }
        }
        .accessibilityIdentifier("sources")
    }

    private var addPresented: Binding<Bool> {
        Binding(
            get: { model.isAdding },
            set: { presented in
                if presented {
                    model.openAdd()
                } else {
                    model.cancelAdd()
                }
            }
        )
    }

    private var header: some View {
        HStack(alignment: .bottom, spacing: PVSpacing.space8) {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.Workspace.sourcesTitle)
                    .font(PVFont.display(size: PVTypeScale.h1))
                    .foregroundStyle(PVColor.textDisplay)
                Text(L10n.Sources.description)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: PVSpacing.measureProse, alignment: .leading)
            }
            Spacer(minLength: PVSpacing.space6)
            HStack(spacing: PVSpacing.space6) {
                if !model.isCatalogEmpty {
                    Text(model.countLine)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                        .accessibilityIdentifier("sources.countLine")
                }
                PVButton(L10n.Sources.addSource, variant: .primary, icon: .plus) {
                    model.openAdd()
                }
                .accessibilityIdentifier("sources.add")
            }
        }
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.top, PVSpacing.space8)
        .padding(.bottom, PVSpacing.space7)
    }

    private var toolbar: some View {
        HStack(spacing: PVSpacing.space5) {
            filterMenu
            sortMenu
            Spacer(minLength: 0)
        }
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.vertical, PVSpacing.space6)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .bottom) {
            PVDivider()
        }
        .zIndex(1)
    }

    private var filterMenu: some View {
        let options = [PVSelectOption(
            value: "",
            label: String(localized: L10n.Sources.filterAllTypes),
            accessibilityIdentifier: "sources.filter.all"
        )] + model.types.map {
            PVSelectOption(
                value: $0.id,
                label: $0.label,
                accessibilityIdentifier: "sources.filter.\($0.id)"
            )
        }
        return PVSelect(
            selection: $model.typeFilterID,
            options: options,
            icon: .filter,
            menuWidth: 220,
            fillsWidth: false,
            accessibilityLabel: L10n.Sources.filterMenu,
            accessibilityIdentifier: "sources.filter"
        )
    }

    private var sortMenu: some View {
        let options = SourcesModel.Sort.allCases.map {
            PVSelectOption(
                value: $0.rawValue,
                label: String(localized: $0.label),
                accessibilityIdentifier: "sources.sort.\($0.rawValue)"
            )
        }
        return PVSelect(
            selection: Binding(
                get: { model.sort.rawValue },
                set: { if let sort = SourcesModel.Sort(rawValue: $0) { model.sort = sort } }
            ),
            options: options,
            icon: .sort,
            displayLabel: model.sortControlLabel,
            menuWidth: 220,
            fillsWidth: false,
            accessibilityLabel: L10n.Sources.sortMenu,
            accessibilityIdentifier: "sources.sort"
        )
    }

    @ViewBuilder
    private var listBody: some View {
        if model.visibleSources.isEmpty {
            PVEmptyState(
                icon: .filter,
                title: L10n.Sources.noFilterMatchesTitle,
                message: String(localized: L10n.Sources.noFilterMatchesMessage),
                compact: true
            )
            .padding(.horizontal, PVSpacing.gutterPage)
            .padding(.top, PVSpacing.space6)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            .accessibilityIdentifier("sources.filterEmpty")
        } else {
            ScrollView {
                LazyVStack(spacing: 0, pinnedViews: [.sectionHeaders]) {
                    Section {
                        ForEach(model.visibleSources) { source in
                            SourcesSplitRow(
                                source: source,
                                typeLabel: model.typeLabel(for: source),
                                typeIconKey: model.typeIconKey(for: source),
                                projectDir: model.pageProjectDir,
                                onOpenPage: {
                                    navigation.go(to: SourcesListNavigation.pageLocation(for: source))
                                },
                                onOpenGraph: {
                                    if let location = SourcesListNavigation.graphLocation(for: source) {
                                        navigation.go(to: location)
                                    }
                                }
                            )
                        }
                    } header: {
                        captionBand
                    }
                }
            }
            .accessibilityElement(children: .contain)
            .accessibilityLabel(Text(L10n.Sources.listAccessibilityLabel))
            .accessibilityIdentifier("sources.list")
        }
    }

    private var captionBand: some View {
        SourcesSplitLayout.columns {
            Text(L10n.Sources.columnSource)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.leading, PVSpacing.gutterPage)
                .padding(.trailing, PVSpacing.space6)
        } graph: {
            Text(L10n.Sources.columnEvidenceGraph)
                .frame(maxWidth: .infinity, alignment: .leading)
                // Padding inside the fixed 210pt track (not outside — that skewed the hairline).
                .padding(.horizontal, PVSpacing.space6)
        }
        .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
        .tracking(PVTypeScale.micro * PVTracking.caps)
        .textCase(.uppercase)
        .foregroundStyle(PVColor.textMuted)
        .padding(.vertical, PVSpacing.space3)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .bottom) {
            PVDivider()
        }
        .accessibilityIdentifier("sources.captionBand")
    }

    private var emptyState: some View {
        PVEmptyState(
            icon: .scrollText,
            title: L10n.Sources.emptyTitle,
            message: String(localized: L10n.Sources.emptyMessage)
        )
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.top, PVSpacing.space8)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .accessibilityIdentifier("sources.empty")
    }

    private func addForm(typeOptions: [PVComboBoxOption]) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let createError = model.createError {
                PVCallout(tone: .danger, message: createError)
                    .accessibilityIdentifier("sources.add.error")
            } else if let loadError = model.loadError, typeOptions.isEmpty {
                PVCallout(
                    tone: .danger,
                    message: L10n.Errors.message(for: loadError)
                )
                .accessibilityIdentifier("sources.add.loadError")
            }
            PVField(
                label: L10n.Sources.formType,
                error: model.typeError,
                required: true
            ) {
                PVComboBox(
                    selection: $model.draft.sourceTypeID,
                    options: typeOptions,
                    placeholder: L10n.Sources.typePlaceholder,
                    emptyLabel: typeOptions.isEmpty
                        ? L10n.Sources.typePoolEmpty
                        : L10n.Sources.typeNoMatch,
                    isInvalid: model.typeError != nil,
                    label: L10n.Sources.formType,
                    accessibilityIdentifierPrefix: "sources.add.type"
                ) { option, query in
                    sourceTypeComboRow(option: option, query: query)
                }
                .onChange(of: model.draft.sourceTypeID) { _, newValue in
                    if !newValue.isEmpty { model.typeError = nil }
                }
            }
            PVField(
                label: L10n.Sources.formTitle,
                error: model.titleError,
                required: true
            ) {
                PVInput(text: $model.draft.title, isInvalid: model.titleError != nil)
                    .accessibilityIdentifier("sources.add.title")
            }
            PVField(
                label: L10n.Sources.formDescription,
                hint: L10n.Sources.formDescriptionHint
            ) {
                PVInput(text: $model.draft.description)
                    .accessibilityIdentifier("sources.add.description")
            }
        }
    }

    private func sourceTypeComboRow(option: PVComboBoxOption, query: String) -> some View {
        HStack(spacing: PVSpacing.space3) {
            if let type = model.types.first(where: { $0.id == option.value }) {
                PVMark(
                    PVMarkKey(catalogKey: type.iconKey),
                    size: .row,
                    decorative: true
                )
            }
            PVComboBoxPlainRow(option: option, query: query)
        }
    }
}
