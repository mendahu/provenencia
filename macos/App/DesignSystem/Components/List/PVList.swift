import SwiftUI

/// The browse list for navigation targets — mirrors the design system's
/// `components/PVList.jsx`. Homogeneous rows: an optional thumbnail slot, a
/// title, an optional secondary line, a trailing mono reference, and a
/// chevron. Rows are **buttons, not selection**: no selection model, no
/// columns, no sort — only hover, a pressed tint, and keyboard focus. The
/// counterpart to `PVTable` (single-selection browse with columns).
///
/// The list does not scroll itself; put it in the page's `ScrollView`. Rows
/// are lazy, so hundreds stay cheap. Keyboard: the list is one focus stop;
/// ↑ / ↓ move the focused row (⌥ jumps to the ends), Home / End, Page Up /
/// Page Down move by ten, Return / Space activate.
struct PVList<Item: Identifiable, Primary: View, Secondary: View>: View {
    private let items: [Item]
    private let thumbnail: ((Item) -> PVThumbnail.Content)?
    private let thumbnailSize: CGFloat
    private let primary: (Item) -> Primary
    private let secondary: (Item) -> Secondary
    private let meta: (Item) -> String?
    private let chevron: Bool
    private let density: PVListDensity
    private let label: LocalizedStringResource
    private let itemAccessibilityLabel: (Item) -> String
    private let itemAccessibilityIdentifier: ((Item) -> String)?
    private let onActivate: (Item) -> Void

    @FocusState private var isFocused: Bool
    @State private var focusIndex = 0

    init(
        items: [Item],
        thumbnail: ((Item) -> PVThumbnail.Content)? = nil,
        thumbnailSize: CGFloat = 44,
        meta: @escaping (Item) -> String? = { _ in nil },
        chevron: Bool = true,
        density: PVListDensity = .comfortable,
        label: LocalizedStringResource,
        itemAccessibilityLabel: @escaping (Item) -> String,
        itemAccessibilityIdentifier: ((Item) -> String)? = nil,
        onActivate: @escaping (Item) -> Void,
        @ViewBuilder primary: @escaping (Item) -> Primary,
        @ViewBuilder secondary: @escaping (Item) -> Secondary
    ) {
        self.items = items
        self.thumbnail = thumbnail
        self.thumbnailSize = thumbnailSize
        self.primary = primary
        self.secondary = secondary
        self.meta = meta
        self.chevron = chevron
        self.density = density
        self.label = label
        self.itemAccessibilityLabel = itemAccessibilityLabel
        self.itemAccessibilityIdentifier = itemAccessibilityIdentifier
        self.onActivate = onActivate
    }

    private var currentIndex: Int {
        min(focusIndex, max(0, items.count - 1))
    }

    var body: some View {
        ScrollViewReader { proxy in
            LazyVStack(spacing: 0) {
                ForEach(Array(items.enumerated()), id: \.element.id) { index, item in
                    row(item, index: index)
                        .id(item.id)
                }
            }
            .onChange(of: focusIndex) { _, index in
                guard items.indices.contains(index) else { return }
                proxy.scrollTo(items[index].id)
            }
        }
        .pvListChrome()
        .focusable()
        .focusEffectDisabled()
        .focused($isFocused)
        .onKeyPress { press in
            let action = PVListKeyboard.action(
                for: press.key,
                option: press.modifiers.contains(.option),
                index: currentIndex,
                count: items.count
            )
            switch action {
            case .move(let index):
                focusIndex = index
                return .handled
            case .activate(let index):
                onActivate(items[index])
                return .handled
            case .ignore:
                return .ignored
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(label))
    }

    private func row(_ item: Item, index: Int) -> some View {
        Button {
            focusIndex = index
            onActivate(item)
        } label: {
            HStack(spacing: PVSpacing.space6) {
                if let thumbnail {
                    PVThumbnail(thumbnail(item), size: thumbnailSize)
                }
                VStack(alignment: .leading, spacing: 3) {
                    primary(item)
                        .font(PVFont.body(size: PVTypeScale.body))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                        .truncationMode(.tail)
                    secondary(item)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                HStack(spacing: PVSpacing.space7) {
                    if let meta = meta(item) {
                        Text(verbatim: meta)
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                            .fixedSize()
                    }
                    if chevron {
                        PVIcon(.chevronForward, size: 15)
                            .foregroundStyle(PVColor.textFaint)
                    }
                }
                .layoutPriority(1)
            }
            .padding(.vertical, density.vertical)
            .padding(.horizontal, density.horizontal)
        }
        .buttonStyle(PVListRowStyle(focused: isFocused && index == currentIndex))
        .overlay(alignment: .bottom) {
            if index < items.count - 1 {
                PVDivider()
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(Text(verbatim: itemAccessibilityLabel(item)))
        .accessibilityAddTraits(.isButton)
        .accessibilityIdentifier(itemAccessibilityIdentifier?(item) ?? "")
    }
}

extension PVList where Secondary == EmptyView {
    /// A list with no secondary line.
    init(
        items: [Item],
        thumbnail: ((Item) -> PVThumbnail.Content)? = nil,
        thumbnailSize: CGFloat = 44,
        meta: @escaping (Item) -> String? = { _ in nil },
        chevron: Bool = true,
        density: PVListDensity = .comfortable,
        label: LocalizedStringResource,
        itemAccessibilityLabel: @escaping (Item) -> String,
        itemAccessibilityIdentifier: ((Item) -> String)? = nil,
        onActivate: @escaping (Item) -> Void,
        @ViewBuilder primary: @escaping (Item) -> Primary
    ) {
        self.init(
            items: items, thumbnail: thumbnail, thumbnailSize: thumbnailSize, meta: meta,
            chevron: chevron, density: density, label: label,
            itemAccessibilityLabel: itemAccessibilityLabel,
            itemAccessibilityIdentifier: itemAccessibilityIdentifier,
            onActivate: onActivate, primary: primary, secondary: { _ in EmptyView() }
        )
    }
}

/// Row padding: comfortable 11 / 16, compact 7 / 12 (the kit's `PAD`).
enum PVListDensity: Sendable {
    case comfortable
    case compact

    var vertical: CGFloat { self == .comfortable ? 11 : 7 }
    var horizontal: CGFloat { self == .comfortable ? PVSpacing.space6 : PVSpacing.space5 }
}

/// What a key does to the list — pure, so it is testable without a view.
enum PVListKeyAction: Equatable {
    case move(to: Int)
    case activate(Int)
    case ignore
}

enum PVListKeyboard {
    static let pageStep = 10

    static func action(for key: KeyEquivalent, option: Bool, index: Int, count: Int) -> PVListKeyAction {
        guard count > 0 else { return .ignore }
        switch key {
        case .downArrow:
            return .move(to: option ? count - 1 : PVFloatingMenuSelection.moveIndex(from: index, delta: 1, count: count))
        case .upArrow:
            return .move(to: option ? 0 : PVFloatingMenuSelection.moveIndex(from: index, delta: -1, count: count))
        case .home:
            return .move(to: 0)
        case .end:
            return .move(to: count - 1)
        case .pageDown:
            return .move(to: PVFloatingMenuSelection.moveIndex(from: index, delta: pageStep, count: count))
        case .pageUp:
            return .move(to: PVFloatingMenuSelection.moveIndex(from: index, delta: -pageStep, count: count))
        case .return, .space:
            return .activate(min(max(index, 0), count - 1))
        default:
            return .ignore
        }
    }
}

/// Row chrome: card at rest, hover tint, pressed tint with the shared press
/// scale (via `PVHoverEffect`), and an inset focus ring on the focused row.
private struct PVListRowStyle: ButtonStyle {
    let focused: Bool

    func makeBody(configuration: Configuration) -> some View {
        PVHoverEffect(isPressed: configuration.isPressed) { hovering in
            configuration.label
                .background(
                    configuration.isPressed ? PVColor.surfaceActive
                        : hovering ? PVColor.surfaceHover : PVColor.surfaceCard
                )
                .pvFocusRing(focused, cornerRadius: PVRadius.sm)
        }
    }
}

/// First-load placeholder in row geometry: static rule bars, a loading tile,
/// no spinner, inert. Pair it with no count in the section header.
struct PVListSkeleton: View {
    var rows: Int = 5
    var thumbnailSize: CGFloat = 44
    var density: PVListDensity = .comfortable

    private static let titleWidths: [CGFloat] = [0.46, 0.32, 0.54, 0.38, 0.42]

    var body: some View {
        VStack(spacing: 0) {
            ForEach(0 ..< rows, id: \.self) { index in
                HStack(spacing: PVSpacing.space6) {
                    PVThumbnail(PVThumbnail.Content(loading: true), size: thumbnailSize)
                    GeometryReader { geometry in
                        VStack(alignment: .leading, spacing: PVSpacing.space3) {
                            bar(width: geometry.size.width * Self.titleWidths[index % Self.titleWidths.count], height: 12)
                            bar(width: geometry.size.width * 0.62, height: 10)
                        }
                        .frame(maxHeight: .infinity, alignment: .center)
                    }
                    bar(width: 72, height: 10)
                }
                .frame(height: thumbnailSize)
                .padding(.vertical, density.vertical)
                .padding(.horizontal, density.horizontal)
                .overlay(alignment: .bottom) {
                    if index < rows - 1 { PVDivider() }
                }
            }
        }
        .pvListChrome()
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(Text(L10n.DesignSystem.listLoading))
    }

    private func bar(width: CGFloat, height: CGFloat) -> some View {
        RoundedRectangle(cornerRadius: 2, style: .continuous)
            .fill(PVColor.borderSubtle)
            .frame(width: width, height: height)
    }
}

private extension View {
    /// Card fill, hairline border, medium radius — shared by the list and its skeleton.
    func pvListChrome() -> some View {
        background(PVColor.surfaceCard)
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
            )
    }
}

#Preview {
    struct Row: Identifiable {
        let id: String
        let title: String
        let subtitle: String?
    }
    let rows = [
        Row(id: "PER-7KD45", title: "James Robins", subtitle: "b. 14 May 1817"),
        Row(id: "PER-2HT08", title: "Eliza Robins", subtitle: nil),
        Row(id: "PER-9ZZ02", title: "PER-9ZZ02", subtitle: nil),
    ]
    return VStack(spacing: PVSpacing.space8) {
        PVList(
            items: rows,
            thumbnail: { _ in PVThumbnail.Content(mark: .subjectPerson) },
            meta: { $0.id },
            label: "Persons",
            itemAccessibilityLabel: { "\($0.title), \($0.id)" },
            onActivate: { _ in },
            primary: { Text(verbatim: $0.title) },
            secondary: { row in
                if let subtitle = row.subtitle { Text(verbatim: subtitle) }
            }
        )
        PVListSkeleton(rows: 3)
    }
    .padding(PVSpacing.space9)
    .frame(width: 640)
    .background(PVColor.surfacePage)
}
