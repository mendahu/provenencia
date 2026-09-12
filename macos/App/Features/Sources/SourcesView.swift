import SwiftUI

/// The **Sources** workspace destination (S2-04 board / S2-17–18): evidence
/// list (not `PVTable`), Add Source dialog, and navigation to the Source page.
/// Mounts inside the S2-01 workspace content host.
struct SourcesView: View {
    @State private var model: SourcesModel
    private let sessionDisplayName: String

    init(
        projectDir: String,
        userID: String,
        sessionDisplayName: String = "",
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        self.sessionDisplayName = sessionDisplayName
        _model = State(
            initialValue: SourcesModel(
                projectDir: projectDir,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        )
    }

    var body: some View {
        Group {
            if let opened = model.openedSourceID {
                SourcePageView(
                    sourceID: opened,
                    projectDir: model.pageProjectDir,
                    userID: model.pageUserID,
                    sessionDisplayName: sessionDisplayName,
                    store: model.pageStore,
                    onBackToList: { model.closeSource() },
                    onSourceUpdated: { model.applyUpdatedSource($0) }
                )
                .id(opened)
            } else {
                listDestination
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .vocabularyToastOverlay($model.toast, identifier: "sources.toast")
        .pvDialog(
            isPresented: addPresented,
            copy: PVDialogCopy(
                title: L10n.Sources.addDialogTitle,
                subtitle: L10n.Sources.addDialogSubtitle,
                confirm: L10n.Sources.createAction,
                cancel: L10n.Sources.cancelAction
            ),
            isRunning: model.isSaving,
            accessibilityIdentifierPrefix: "sources.add",
            onConfirm: { Task { await model.create() } }
        ) {
            // Read options here so SourcesView observes `types` and the
            // sheet rebuilds when `refreshTypes` fills the pool.
            addForm(typeOptions: model.typeComboOptions)
        }
        .task { await model.load() }
        .task(id: model.isAdding) {
            guard model.isAdding else { return }
            await model.refreshTypes()
            model.selectSoleTypeIfNeeded()
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

    private var listDestination: some View {
        VStack(spacing: 0) {
            header
            PVDivider()
            if model.isLoading && model.sources.isEmpty && model.loadError == nil {
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
        // Same search chrome as `VocabularyListPane.searchBar` (Source fields /
        // Source types): default `PVInput` size, card strip, bottom hairline.
        HStack(spacing: PVSpacing.space5) {
            PVInput(
                text: $model.query,
                prompt: L10n.Sources.searchPlaceholder,
                icon: .search
            )
            .frame(maxWidth: 420)
            .accessibilityIdentifier("sources.search")

            if !model.query.isEmpty {
                PVButton(L10n.Sources.clearSearch, variant: .ghost, size: .sm) {
                    model.query = ""
                }
                .accessibilityIdentifier("sources.clearSearch")
            }

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
    }

    private var filterMenu: some View {
        // Inline `Picker` (same as `PVTable` column filters): AppKit menus drop
        // custom `HStack` button labels, which left these rows blank.
        Menu {
            Picker(String(localized: L10n.Sources.filterMenu), selection: $model.typeFilterID) {
                Text(L10n.Sources.filterAllTypes).tag("")
                ForEach(model.types) { type in
                    Text(type.label).tag(type.id)
                }
            }
            .pickerStyle(.inline)
            .labelsHidden()
        } label: {
            toolbarChip(icon: .filter, label: model.filterLabel)
        }
        .menuStyle(.borderlessButton)
        .accessibilityLabel(Text(L10n.Sources.filterMenu))
        .accessibilityIdentifier("sources.filter")
    }

    private var sortMenu: some View {
        Menu {
            Picker(String(localized: L10n.Sources.sortMenu), selection: $model.sort) {
                ForEach(SourcesModel.Sort.allCases) { option in
                    Text(option.label).tag(option)
                }
            }
            .pickerStyle(.inline)
            .labelsHidden()
        } label: {
            toolbarChip(icon: .sort, label: model.sortControlLabel)
        }
        .menuStyle(.borderlessButton)
        .accessibilityLabel(Text(L10n.Sources.sortMenu))
        .accessibilityIdentifier("sources.sort")
    }

    private func toolbarChip(icon: PVSymbol, label: String) -> some View {
        HStack(spacing: PVSpacing.space3) {
            PVIcon(icon, size: 12)
            Text(label)
                .lineLimit(1)
            PVIcon(.chevronDown, size: 11)
        }
        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
        .foregroundStyle(PVColor.textPrimary)
        .padding(.horizontal, PVSpacing.space4)
        .frame(height: PVSpacing.controlHeightMedium)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.surfaceRaised)
        )
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(PVColor.borderDefault, lineWidth: 1)
        )
    }

    @ViewBuilder
    private var listBody: some View {
        if model.visibleSources.isEmpty {
            PVEmptyState(
                icon: .searchEmpty,
                title: L10n.Sources.noMatchesTitle,
                message: L10n.Sources.noMatchesMessage(query: model.query),
                compact: true
            )
            .padding(.horizontal, PVSpacing.gutterPage)
            .padding(.top, PVSpacing.space6)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            .accessibilityIdentifier("sources.noMatches")
        } else {
            PVList(
                rows: model.visibleSources,
                label: L10n.Sources.listAccessibilityLabel,
                primary: { $0.title },
                secondary: { model.typeLabel(for: $0) },
                meta: { $0.ref },
                thumbnail: { source in
                    CachedThumbnail(
                        projectDir: model.pageProjectDir,
                        relPath: source.thumbnailRelPath,
                        mediaType: source.thumbnailMediaType,
                        originalFilename: source.thumbnailOriginalFilename
                    )
                },
                onActivate: { model.openSource(id: $0) },
                rowAccessibilityIdentifier: { "sources.row.\($0.id)" }
            )
            .accessibilityIdentifier("sources.list")
        }
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
                )
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
}
