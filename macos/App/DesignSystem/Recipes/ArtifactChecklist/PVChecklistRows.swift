import SwiftUI

/// The bordered, divided, height-capped list that every id checklist draws its
/// rows in: an optional `header` row above the scrolling rows (the All row),
/// then one row per item with a divider between rows. Owns the 264pt cap, so
/// a Source with many Artifacts or graphs scrolls inside the form.
struct PVChecklistFrame<Item, Header: View, Row: View>: View {
    private static var maxRowsHeight: CGFloat { 264 }

    let items: [Item]
    let id: KeyPath<Item, String>
    private let showsHeader: Bool
    private let header: Header
    private let row: (Item) -> Row

    init(
        items: [Item],
        id: KeyPath<Item, String>,
        @ViewBuilder header: () -> Header,
        @ViewBuilder row: @escaping (Item) -> Row
    ) {
        self.items = items
        self.id = id
        self.showsHeader = true
        self.header = header()
        self.row = row
    }

    var body: some View {
        VStack(spacing: 0) {
            if showsHeader { header }
            ScrollView {
                LazyVStack(spacing: 0) {
                    ForEach(items, id: id) { item in
                        if showsHeader || item[keyPath: id] != items.first?[keyPath: id] { PVDivider() }
                        row(item)
                    }
                }
            }
            .frame(maxHeight: Self.maxRowsHeight)
        }
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .overlay {
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        }
    }
}

extension PVChecklistFrame where Header == EmptyView {
    /// A frame with no header row: the first row has no divider above it.
    init(items: [Item], id: KeyPath<Item, String>, @ViewBuilder row: @escaping (Item) -> Row) {
        self.items = items
        self.id = id
        self.showsHeader = false
        self.header = EmptyView()
        self.row = row
    }
}

/// One checklist row's checkbox for an id in the `excluded` set. A row with a
/// `lockReason` is checked, shows the lock mark with the reason under its
/// title, speaks "Locked, <ref>" and cannot be cleared. Otherwise it toggles
/// its id in `excluded`. `showsTitle` false draws the box alone and speaks the
/// title (an Artifact with no label of its own); `description` is a caption
/// under the title; `speaksRef` adds the ref as the spoken value.
struct PVChecklistCheckbox: View {
    let id: String
    let title: String
    let ref: String
    var showsTitle = true
    var description: String?
    var speaksRef = false
    var lockReason: String?
    @Binding var excluded: Set<String>
    var isDisabled = false

    var body: some View {
        if let lockReason {
            // Always the labelled form, so the reason is never alone.
            PVCheckbox(
                verbatim: title,
                description: lockReason,
                isChecked: .constant(true),
                isLocked: true
            )
            .accessibilityValue(Text(verbatim: L10n.Checklist.lockedValue(ref: ref)))
        } else {
            let box = unlockedBox
            if speaksRef {
                box.accessibilityValue(Text(verbatim: ref))
            } else {
                box
            }
        }
    }

    @ViewBuilder
    private var unlockedBox: some View {
        if showsTitle {
            PVCheckbox(verbatim: title, description: description, isChecked: isChecked, isDisabled: isDisabled)
        } else {
            PVCheckbox(isChecked: isChecked, isDisabled: isDisabled, accessibilitySpokenLabel: title)
        }
    }

    private var isChecked: Binding<Bool> {
        Binding(
            get: { !excluded.contains(id) },
            set: { PVChecklistState.setChecked($0, id: id, in: &excluded) }
        )
    }
}
