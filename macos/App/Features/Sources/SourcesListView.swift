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
        projectUuid: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        _model = State(
            initialValue: SourcesModel(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts,
                settingsStore: SourcesListSettingsStore(projectUuid: projectUuid, userID: userID)
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
                    progressHandle: session.queryHandle(
                        SourcesModel.sourceGraphProgressKey(for: session)
                    ),
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
    var progressHandle: QueryHandle<[String: SourceGraphProgress]>?
    @Bindable var model: SourcesModel

    var body: some View {
        let rows = model.visibleSources
        return VStack(spacing: 0) {
            header(visibleCount: rows.count)
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
                listBody(rows: rows)
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
            isDirty: model.addSheetIsDirty,
            confirmDisabled: !model.canSubmitAdd,
            accessibilityIdentifierPrefix: "sources.add",
            onConfirm: {
                Task {
                    if let created = await model.create() {
                        navigation.go(to: SourcesListNavigation.pageLocation(for: created))
                    }
                }
            }
        ) {
            addForm
        }
        .sourceCopyDialog($model.copying)
        .task(id: model.isAdding) {
            guard model.isAdding else { return }
            model.refreshTypes()
            model.selectSoleTypeIfNeeded()
        }
        .onSettled(sourcesHandle) { model.syncCatalogCounts() }
        .onChange(of: typesHandle.status) { _, _ in
            if model.isAdding { model.selectSoleTypeIfNeeded() }
        }
        .onChange(of: model.settings) { _, _ in
            navigation.go(to: navigation.currentLocation.withListScrollAnchor(nil))
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

    private func header(visibleCount: Int) -> some View {
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
                    Text(model.countLine(visibleCount: visibleCount))
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
            searchField
            filterMenu
            statusMenu
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

    private var searchField: some View {
        PVInput(
            text: $model.settings.query,
            prompt: L10n.Sources.searchPrompt,
            icon: .search
        )
        .frame(minWidth: 160, maxWidth: 280)
        .accessibilityIdentifier("sources.search")
    }

    private var statusMenu: some View {
        let options = [PVSelectOption(
            value: "",
            label: L10n.string(L10n.Sources.statusAll),
            accessibilityIdentifier: "sources.status.all"
        )] + SourcesModel.EvidenceStatus.allCases.map {
            PVSelectOption(
                value: $0.rawValue,
                label: L10n.string($0.label),
                accessibilityIdentifier: "sources.status.\($0.rawValue)"
            )
        }
        return PVSelect(
            selection: Binding(
                get: { model.settings.status?.rawValue ?? "" },
                set: { model.settings.status = SourcesModel.EvidenceStatus(rawValue: $0) }
            ),
            options: options,
            icon: .network,
            menuWidth: 220,
            fillsWidth: false,
            accessibilityLabel: L10n.Sources.statusMenu,
            accessibilityIdentifier: "sources.status"
        )
    }

    private var filterMenu: some View {
        let options = [PVSelectOption(
            value: "",
            label: L10n.string(L10n.Sources.filterAllTypes),
            accessibilityIdentifier: "sources.filter.all"
        )] + model.types.map {
            PVSelectOption(
                value: $0.id,
                label: $0.label,
                accessibilityIdentifier: "sources.filter.\($0.id)"
            )
        }
        return PVSelect(
            selection: Binding(
                get: { model.effectiveTypeFilterID },
                set: { model.settings.typeFilterID = $0 }
            ),
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
                label: L10n.string($0.label),
                accessibilityIdentifier: "sources.sort.\($0.rawValue)"
            )
        }
        return PVSelect(
            selection: Binding(
                get: { model.settings.sort.rawValue },
                set: { if let sort = SourcesModel.Sort(rawValue: $0) { model.settings.sort = sort } }
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
    private func listBody(rows: [CatalogSource]) -> some View {
        if let statusError = model.statusCountsError {
            PVCallout(tone: .danger, message: L10n.Errors.message(for: statusError))
                .padding(.horizontal, PVSpacing.gutterPage)
                .padding(.top, PVSpacing.space8)
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
                .accessibilityIdentifier("sources.statusError")
        } else if model.isAwaitingFilterData {
            ProgressView()
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .center)
                .padding(.top, PVSpacing.space9)
        } else if rows.isEmpty {
            PVEmptyState(
                icon: .filter,
                title: L10n.Sources.noFilterMatchesTitle,
                message: L10n.string(L10n.Sources.noFilterMatchesMessage),
                compact: true
            ) {
                PVButton(L10n.Sources.clearFilters, variant: .secondary) {
                    model.clearFilters()
                }
                .accessibilityIdentifier("sources.clearFilters")
            }
            .padding(.horizontal, PVSpacing.gutterPage)
            .padding(.top, PVSpacing.space6)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            .accessibilityIdentifier("sources.filterEmpty")
        } else {
            ScrollView {
                LazyVStack(spacing: 0, pinnedViews: [.sectionHeaders]) {
                    Section {
                        ForEach(rows) { source in
                            if let progressHandle {
                                SourcesSplitRowProgress(source: source, model: model, handle: progressHandle)
                                    .id(source.id)
                            } else {
                                SourcesSplitRow(
                                    source: source,
                                    typeLabel: model.typeLabel(for: source),
                                    typeIconKey: model.typeIconKey(for: source),
                                    projectDir: model.pageProjectDir,
                                    countsLoading: true,
                                    onOpenPage: {
                                        navigation.go(to: SourcesListNavigation.pageLocation(for: source))
                                    },
                                    onOpenGraph: {
                                        if let location = SourcesListNavigation.graphLocation(for: source) {
                                            navigation.go(to: location)
                                        }
                                    },
                                    onCopy: { model.openCopy(source) }
                                )
                                .id(source.id)
                            }
                        }
                    } header: {
                        captionBand
                    }
                }
                .scrollTargetLayout()
            }
            .modifier(ListScrollAnchorModifier(
                anchor: navigation.currentLocation.listScrollAnchor,
                rowIDs: rows.map(\.id),
                isLoaded: true
            ))
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
                // Padding inside the fixed graph-zone track (not outside — that skewed the hairline).
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
            message: L10n.string(L10n.Sources.emptyMessage)
        )
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.top, PVSpacing.space8)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .accessibilityIdentifier("sources.empty")
    }

    private var addForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let createError = model.createError {
                PVCallout(tone: .danger, message: createError)
                    .accessibilityIdentifier("sources.add.error")
            } else if let loadError = model.loadError, model.types.isEmpty {
                PVCallout(
                    tone: .danger,
                    message: L10n.Errors.message(for: loadError)
                )
                .accessibilityIdentifier("sources.add.loadError")
            }
            SourceTitleTypeFields(
                draft: $model.draft,
                titleError: $model.titleError,
                typeError: $model.typeError,
                types: model.types,
                accessibilityIdentifierPrefix: "sources.add"
            )
            PVField(
                label: L10n.Sources.formDescription,
                hint: L10n.Sources.formDescriptionHint
            ) {
                PVInput(text: $model.draft.description)
                    .accessibilityIdentifier("sources.add.description")
            }
        }
    }
}

/// Observes the progress map so one Source's Get-one merge repaints that row.
private struct SourcesSplitRowProgress: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    let source: CatalogSource
    let model: SourcesModel
    @Bindable var handle: QueryHandle<[String: SourceGraphProgress]>

    var body: some View {
        SourcesSplitRow(
            source: source,
            typeLabel: model.typeLabel(for: source),
            typeIconKey: model.typeIconKey(for: source),
            projectDir: model.pageProjectDir,
            progress: handle.value?[source.id],
            countsLoading: handle.value == nil && handle.status != .ready,
            onOpenPage: {
                navigation.go(to: SourcesListNavigation.pageLocation(for: source))
            },
            onOpenGraph: {
                if let location = SourcesListNavigation.graphLocation(for: source) {
                    navigation.go(to: location)
                }
            },
            onCopy: { model.openCopy(source) }
        )
    }
}
