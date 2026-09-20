import SwiftUI

/// Mirrors `components/data/Table.jsx` (and the design system's own
/// `swift/ProvenenciaTable.swift` port).
///
/// The Provenencia browse table: micro-caps column header, hairline row
/// rules, an accent bar on the selected row, badge cells. Custom chrome on
/// purpose — visual fidelity to the design system wins over SwiftUI `Table`
/// / `NSTableView` chrome — but with the keyboard and VoiceOver behaviour of
/// a native macOS single-selection data table. See `DesignSystem/README.md`
/// for that tradeoff and its honest limits.
///
/// Contract:
/// - **The caller owns the data.** `rows` arrive already filtered and sorted;
///   the table renders sort/filter *controls* and reports intent through
///   `onSortChange` / `PVTableColumnFilter.onChange`.
/// - **Column definitions are the single source of truth for width** — the
///   header and the rows share them, so widths are never restated at the
///   call site.
/// - Search bars and result footers are feature chrome, not table chrome;
///   they stay in the pane, above and below the table.
///
/// Deliberately not implemented, same scope line as the web component:
/// multi-select, column resize/reorder, drag-and-drop, inline editing.

// MARK: - Pure helpers (unit-testable without UI)

enum PVTableSelection {
    /// Clamped selection movement — never wraps, matching a native macOS table.
    static func moveIndex(_ index: Int, delta: Int, count: Int) -> Int {
        guard count > 0 else { return -1 }
        if index < 0 { return delta > 0 ? 0 : count - 1 }
        return min(count - 1, max(0, index + delta))
    }
}

/// Type-to-select buffering, split out so prefix matching is testable without
/// mounting a view. Mirrors `typeSelectIndex` + `TYPE_SELECT_RESET_MS`.
struct PVTableTypeSelectMatcher {
    /// `TYPE_SELECT_RESET_MS` — the buffer starts over after this long a pause.
    static let resetInterval: TimeInterval = 0.8

    private var buffer = ""
    private var lastKeystroke = Date.distantPast

    init() {}

    /// Appends a character, resetting the buffer first if the user paused.
    mutating func append(_ character: Character, now: Date = Date()) -> String {
        buffer = now.timeIntervalSince(lastKeystroke) > Self.resetInterval ? String(character) : buffer + String(character)
        lastKeystroke = now
        return buffer
    }

    mutating func reset() {
        buffer = ""
        lastKeystroke = .distantPast
    }

    /// First index whose value starts with `prefix`, searching forward from
    /// `fromIndex` and wrapping, so repeated single keystrokes cycle matches.
    static func index(in values: [String], prefix: String, fromIndex: Int = 0) -> Int {
        let query = prefix.lowercased()
        guard !query.isEmpty, !values.isEmpty else { return -1 }
        for offset in 0..<values.count {
            let idx = (max(0, fromIndex) + offset) % values.count
            if values[idx].lowercased().hasPrefix(query) { return idx }
        }
        return -1
    }
}

// MARK: - Column definition

/// A menu of mutually exclusive values a column can be narrowed to. The table
/// only renders the control — the caller applies the filter to `rows`.
struct PVTableColumnFilter {
    struct Option: Identifiable {
        let value: String
        let label: LocalizedStringResource
        let count: Int?
        var id: String { value }

        init(value: String, label: LocalizedStringResource, count: Int? = nil) {
            self.value = value
            self.label = label
            self.count = count
        }
    }

    let value: String
    let active: Bool
    let options: [Option]
    let onChange: (String) -> Void

    init(value: String, active: Bool = false, options: [Option], onChange: @escaping (String) -> Void) {
        self.value = value
        self.active = active
        self.options = options
        self.onChange = onChange
    }
}

/// One column. `width` is the single source of truth for that column's width
/// in both the header and every row; `nil` means flexible (fills the rest).
struct PVTableColumn<Row: Identifiable>: Identifiable {
    let id: String
    let title: LocalizedStringResource
    let width: CGFloat?
    let sortable: Bool
    let filter: PVTableColumnFilter?
    let cell: (Row) -> AnyView

    init<Cell: View>(
        id: String,
        title: LocalizedStringResource,
        width: CGFloat? = nil,
        sortable: Bool = false,
        filter: PVTableColumnFilter? = nil,
        @ViewBuilder cell: @escaping (Row) -> Cell
    ) {
        self.id = id
        self.title = title
        self.width = width
        self.sortable = sortable
        self.filter = filter
        self.cell = { AnyView(cell($0)) }
    }
}

struct PVTableSort: Equatable {
    enum Direction {
        case ascending, descending
    }

    var columnID: String
    var direction: Direction

    init(columnID: String, direction: Direction = .ascending) {
        self.columnID = columnID
        self.direction = direction
    }
}

/// Row rhythm. `comfortable` is the design system's 12pt table row padding;
/// `compact` tightens it for dense lists.
enum PVTableDensity {
    case comfortable, compact

    var rowPadding: CGFloat { self == .compact ? 7 : 12 }
}

// MARK: - Table

struct PVTable<Row: Identifiable, Content: View>: View {
    /// Already filtered and sorted by the caller — the table never sorts itself.
    let rows: [Row]
    let columns: [PVTableColumn<Row>]
    @Binding var selection: Row.ID?
    /// What type-to-select matches against — the row's primary text.
    let primaryText: (Row) -> String
    let sort: PVTableSort?
    let onSortChange: ((String) -> Void)?
    let density: PVTableDensity
    let gutter: CGFloat
    let label: LocalizedStringResource
    let rowAccessibilityIdentifier: ((Row) -> String)?
    let sortAccessibilityIdentifier: ((String) -> String)?
    /// Rendered inside the scroller after the rows — an empty state, a
    /// pending/unsaved row. Mirrors the web component's `children`.
    let content: Content

    @FocusState private var isFocused: Bool
    @State private var matcher = PVTableTypeSelectMatcher()
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    init(
        rows: [Row],
        columns: [PVTableColumn<Row>],
        selection: Binding<Row.ID?>,
        primaryText: @escaping (Row) -> String,
        sort: PVTableSort? = nil,
        onSortChange: ((String) -> Void)? = nil,
        density: PVTableDensity = .comfortable,
        gutter: CGFloat = PVSpacing.gutterPage,
        label: LocalizedStringResource,
        rowAccessibilityIdentifier: ((Row) -> String)? = nil,
        sortAccessibilityIdentifier: ((String) -> String)? = nil,
        @ViewBuilder content: () -> Content = { EmptyView() }
    ) {
        self.rows = rows
        self.columns = columns
        _selection = selection
        self.primaryText = primaryText
        self.sort = sort
        self.onSortChange = onSortChange
        self.density = density
        self.gutter = gutter
        self.label = label
        self.rowAccessibilityIdentifier = rowAccessibilityIdentifier
        self.sortAccessibilityIdentifier = sortAccessibilityIdentifier
        self.content = content()
    }

    private var selectedIndex: Int {
        rows.firstIndex { $0.id == selection } ?? -1
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            rowList
        }
    }

    /// The scrolling row list: one focus stop, list-like semantics.
    private var rowList: some View {
        ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(spacing: 0) {
                    ForEach(rows) { row in
                        rowView(row).id(row.id)
                    }
                    content
                }
            }
            .onChange(of: selection) { _, new in
                guard let new else { return }
                withAnimation(reduceMotion ? nil : PVMotion.fastStandard) { proxy.scrollTo(new) }
            }
        }
        .focusable()
        .focusEffectDisabled()
        .focused($isFocused)
        .onKeyPress(action: handleKey)
        // Never swap view *structure* on focus (same rule as `PVInput`) —
        // only the ring and the selection tint change.
        // Tables keep square corners (`--radius-none`).
        .pvFocusRing(isFocused, cornerRadius: PVRadius.none)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(label))
    }

    // MARK: Header

    private var header: some View {
        HStack(spacing: PVSpacing.space6) {
            ForEach(columns) { column in
                HStack(spacing: PVSpacing.space1) {
                    if column.sortable, let onSortChange {
                        Button { onSortChange(column.id) } label: { headerLabel(column) }
                            .buttonStyle(.plain)
                            .accessibilityLabel(Text(column.title))
                            .accessibilityValue(Text(sortAccessibilityValue(column)))
                            .accessibilityIdentifier(sortAccessibilityIdentifier?(column.id) ?? "")
                    } else {
                        headerLabel(column)
                    }
                    if let filter = column.filter {
                        filterMenu(column, filter)
                    }
                }
                .frame(maxWidth: column.width == nil ? .infinity : nil, alignment: .leading)
                .frame(width: column.width, alignment: .leading)
            }
        }
        .padding(.horizontal, gutter)
        .padding(.vertical, PVSpacing.space3)
        .background(PVColor.surfaceSunken)
        .overlay(alignment: .bottom) {
            PVDivider(color: PVColor.borderDefault)
        }
    }

    private func headerLabel(_ column: PVTableColumn<Row>) -> some View {
        let isSorted = sort?.columnID == column.id
        return HStack(spacing: PVSpacing.space3) {
            Text(column.title)
                .pvMicroCaps()
                .foregroundStyle(isSorted ? PVColor.textPrimary : PVColor.textMuted)
                .lineLimit(1)
            if column.sortable {
                PVIcon(sortIcon(isSorted: isSorted), size: 11)
                    .foregroundStyle(isSorted ? PVColor.accent : PVColor.textFaint)
            }
        }
        .contentShape(Rectangle())
    }

    private func sortIcon(isSorted: Bool) -> PVSymbol {
        guard isSorted, let sort else { return .sortUnsorted }
        return sort.direction == .ascending ? .sortAscending : .sortDescending
    }

    /// The chevron is decorative (`PVIcon` is `accessibilityHidden`), so the
    /// direction and what activating will do are spoken here instead.
    private func sortAccessibilityValue(_ column: PVTableColumn<Row>) -> LocalizedStringResource {
        guard sort?.columnID == column.id, let sort else { return L10n.DesignSystem.tableSortNone }
        return sort.direction == .ascending ? L10n.DesignSystem.tableSortAscending : L10n.DesignSystem.tableSortDescending
    }

    private func filterMenu(_ column: PVTableColumn<Row>, _ filter: PVTableColumnFilter) -> some View {
        Menu {
            Picker(String(localized: column.title), selection: filterBinding(filter)) {
                ForEach(filter.options) { option in
                    if let count = option.count {
                        Text(L10n.DesignSystem.tableFilterOptionCount(label: String(localized: option.label), count: count))
                            .tag(option.value)
                    } else {
                        Text(option.label).tag(option.value)
                    }
                }
            }
            .pickerStyle(.inline)
            .labelsHidden()
        } label: {
            PVIcon(filter.active ? .filter : .chevronDown, size: 11)
                .foregroundStyle(filter.active ? PVColor.accent : PVColor.textFaint)
                .frame(width: 22, height: 22)
                .background(filter.active ? PVColor.surfaceSelected : Color.clear)
                .clipShape(RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous))
        }
        .menuStyle(.borderlessButton)
        .menuIndicator(.hidden)
        .fixedSize()
        .accessibilityLabel(Text(L10n.DesignSystem.tableFilterColumn(column: String(localized: column.title))))
    }

    private func filterBinding(_ filter: PVTableColumnFilter) -> Binding<String> {
        Binding(get: { filter.value }, set: { filter.onChange($0) })
    }

    // MARK: Rows

    private func rowView(_ row: Row) -> some View {
        let isSelected = row.id == selection
        return PVHoverEffect(isPressed: false, hoverAnimation: PVMotion.fastStandard) { showHover in
            HStack(spacing: PVSpacing.space6) {
                ForEach(columns) { column in
                    column.cell(row)
                        .frame(maxWidth: column.width == nil ? .infinity : nil, alignment: .leading)
                        .frame(width: column.width, alignment: .leading)
                }
            }
            .padding(.horizontal, gutter)
            .padding(.vertical, density.rowPadding)
            .frame(maxWidth: .infinity, alignment: .leading)
            .contentShape(Rectangle())
            .background(rowBackground(isSelected: isSelected, showHover: showHover))
            .overlay(alignment: .leading) {
                // The selected row's leading accent bar, dimmed like a native
                // table's when the table does not have focus.
                Rectangle()
                    .fill(isSelected ? (isFocused ? PVColor.accent : PVColor.borderStrong) : .clear)
                    .frame(width: 2)
            }
            .overlay(alignment: .bottom) {
                PVDivider()
            }
        }
        .onTapGesture { selection = row.id }
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(isSelected ? [.isButton, .isSelected] : .isButton)
        .accessibilityIdentifier(rowAccessibilityIdentifier?(row) ?? "")
    }

    private func rowBackground(isSelected: Bool, showHover: Bool) -> Color {
        if isSelected {
            return isFocused ? PVColor.surfaceSelected : PVColor.surfaceSelectedInactive
        }
        return showHover ? PVColor.surfaceHover : .clear
    }

    // MARK: Keyboard

    private func select(at index: Int) {
        guard rows.indices.contains(index) else { return }
        selection = rows[index].id
    }

    private func handleKey(_ press: KeyPress) -> KeyPress.Result {
        guard !rows.isEmpty else { return .ignored }
        let index = selectedIndex

        switch press.key {
        case .upArrow, .downArrow:
            let isDown = press.key == .downArrow
            if press.modifiers.contains(.option) {
                select(at: isDown ? rows.count - 1 : 0)
            } else {
                select(at: PVTableSelection.moveIndex(index, delta: isDown ? 1 : -1, count: rows.count))
            }
            return .handled
        case .home:
            select(at: 0)
            return .handled
        case .end:
            select(at: rows.count - 1)
            return .handled
        case .pageDown, .pageUp:
            select(at: PVTableSelection.moveIndex(index, delta: press.key == .pageDown ? 10 : -10, count: rows.count))
            return .handled
        case .escape:
            matcher.reset()
            return .handled
        default:
            guard press.characters.count == 1,
                  let character = press.characters.first,
                  character.isLetter || character.isNumber,
                  !press.modifiers.contains(.command),
                  !press.modifiers.contains(.control)
            else { return .ignored }

            let buffer = matcher.append(character)
            // A single keystroke advances past the current row, so pressing
            // the same letter repeatedly cycles through its matches.
            let from = buffer.count == 1 ? index + 1 : max(0, index)
            let hit = PVTableTypeSelectMatcher.index(in: rows.map(primaryText), prefix: buffer, fromIndex: from)
            if hit >= 0 { select(at: hit) }
            return .handled
        }
    }
}

// MARK: - Preview

private struct PVTablePreviewRow: Identifiable {
    let id: String
    let label: String
    let key: String
    let dataType: String
    let seeded: Bool
}

#Preview {
    struct Harness: View {
        @State private var selection: String? = "3"
        @State private var sort = PVTableSort(columnID: "label")

        private let rows = [
            PVTablePreviewRow(id: "1", label: "Author", key: "author", dataType: "text", seeded: true),
            PVTablePreviewRow(id: "2", label: "Certificate number", key: "certificate-number", dataType: "text", seeded: true),
            PVTablePreviewRow(id: "3", label: "Grandma's album code", key: "grandmas-album-code", dataType: "text", seeded: false),
            PVTablePreviewRow(id: "4", label: "Interview date", key: "interview-date", dataType: "date", seeded: false),
            PVTablePreviewRow(id: "5", label: "Memorial id", key: "memorial-id", dataType: "text", seeded: false),
        ]

        var body: some View {
            let ordered = rows.sorted {
                sort.direction == .ascending ? $0.label < $1.label : $0.label > $1.label
            }
            return PVTable(
                rows: ordered,
                columns: [
                    PVTableColumn(id: "label", title: "Label", sortable: true) { row in
                        HStack(spacing: PVSpacing.space2) {
                            Text(row.label)
                                .font(PVFont.body(size: PVTypeScale.bodySmall))
                                .foregroundStyle(PVColor.textPrimary)
                                .lineLimit(1)
                            if row.seeded {
                                PVBadge(icon: .shieldCheck, label: "Seeded by Provenencia", tone: .accent, subtle: true)
                            }
                        }
                    },
                    PVTableColumn(id: "key", title: "Key", width: 160) { row in
                        Text(row.key)
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                    },
                    PVTableColumn(id: "dataType", title: "Data type", width: 116) { row in
                        PVBadge(
                            text: row.dataType,
                            tone: row.dataType == "date" ? .info : .neutral,
                            icon: row.dataType == "date" ? .calendar : .textType,
                            subtle: true
                        )
                    },
                ],
                selection: $selection,
                primaryText: \.label,
                sort: sort,
                onSortChange: { id in
                    sort = PVTableSort(
                        columnID: id,
                        direction: sort.columnID == id && sort.direction == .ascending ? .descending : .ascending
                    )
                },
                label: "Source fields",
                rowAccessibilityIdentifier: { "pvTable.row.\($0.id)" },
                sortAccessibilityIdentifier: { "pvTable.sortBy.\($0)" }
            )
            .frame(width: 720, height: 300)
            .background(PVColor.surfaceCard)
        }
    }
    return Harness()
}
