import Foundation

/// Closed vs open policy for `PVSelect`. The view forwards keys and pointer
/// hit targets here so tests and chrome cannot drift.
struct PVSelectSession: Equatable {
    var options: [PVSelectOption]
    var selection: String
    var isDisabled: Bool
    var isOpen = false
    /// Highlight while the menu is open (`-1` = none).
    var highlightIndex = -1
    /// Committed value when the menu opened — Escape / click-away restore this.
    var selectionOnOpen = ""
    var typeSelect = PVTypeSelectMatcher()

    init(
        options: [PVSelectOption],
        selection: String,
        isDisabled: Bool = false
    ) {
        self.options = options
        self.selection = selection
        self.isDisabled = isDisabled
    }

    var canOpen: Bool {
        !isDisabled && !options.isEmpty
    }

    var selectedIndex: Int {
        options.firstIndex(where: { $0.id == selection }) ?? -1
    }

    var committedLabel: String? {
        options.first(where: { $0.id == selection })?.label
    }

    var highlightedLabel: String? {
        guard options.indices.contains(highlightIndex) else { return nil }
        return options[highlightIndex].label
    }

    var optionTitles: [String] {
        options.map(\.label)
    }

    mutating func syncOptions(_ options: [PVSelectOption], selection: String, isDisabled: Bool) {
        self.options = options
        self.isDisabled = isDisabled
        if self.selection != selection {
            self.selection = selection
        }
        if isOpen, !canOpen {
            dismissRestoring()
        }
    }

    // MARK: Keys

    mutating func handleKey(_ key: PVSelectKey, now: Date = Date()) {
        guard !isDisabled else { return }
        if isOpen {
            handleOpenKey(key, now: now)
        } else {
            handleClosedKey(key, now: now)
        }
    }

    mutating func increment() {
        handleKey(.down)
    }

    mutating func decrement() {
        handleKey(.up)
    }

    // MARK: Pointer

    mutating func handlePointer(_ event: PVSelectPointerEvent) {
        switch event {
        case .press(let target):
            press(target)
        case .drag(let target):
            drag(target)
        case .release(let target):
            release(target)
        case .clickAway:
            clickAway()
        }
    }

    // MARK: Closed

    private mutating func handleClosedKey(_ key: PVSelectKey, now: Date) {
        switch key {
        case .up:
            moveCommitted(delta: -1)
        case .down:
            moveCommitted(delta: 1)
        case .home:
            jumpCommitted(toFirst: true)
        case .end:
            jumpCommitted(toFirst: false)
        case .space, .return:
            open()
        case .escape:
            break
        case .character(let character):
            applyTypeSelect(character, now: now, commit: true)
        }
    }

    mutating func moveCommitted(delta: Int) {
        guard canOpen else { return }
        let next = PVFloatingMenuSelection.moveIndex(
            from: selectedIndex, delta: delta, count: options.count
        )
        commit(index: next, close: false)
    }

    mutating func jumpCommitted(toFirst: Bool) {
        guard canOpen else { return }
        commit(index: toFirst ? 0 : options.count - 1, close: false)
    }

    // MARK: Open

    private mutating func handleOpenKey(_ key: PVSelectKey, now: Date) {
        switch key {
        case .up:
            moveHighlight(delta: -1)
        case .down:
            moveHighlight(delta: 1)
        case .home:
            jumpHighlight(toFirst: true)
        case .end:
            jumpHighlight(toFirst: false)
        case .space, .return:
            commitHighlight()
        case .escape:
            dismissRestoring()
        case .character(let character):
            applyTypeSelect(character, now: now, commit: false)
        }
    }

    @discardableResult
    mutating func open() -> Bool {
        guard canOpen, !isOpen else { return false }
        isOpen = true
        selectionOnOpen = selection
        highlightIndex = selectedIndex >= 0 ? selectedIndex : 0
        typeSelect.reset()
        return true
    }

    mutating func dismissRestoring() {
        guard isOpen else { return }
        selection = selectionOnOpen
        isOpen = false
        highlightIndex = -1
        typeSelect.reset()
    }

    mutating func commitHighlight() {
        guard isOpen else { return }
        commit(index: highlightIndex, close: true)
    }

    mutating func commit(index: Int, close: Bool) {
        guard options.indices.contains(index) else { return }
        selection = options[index].id
        if close {
            isOpen = false
            highlightIndex = -1
            typeSelect.reset()
        }
    }

    mutating func moveHighlight(delta: Int) {
        guard isOpen, !options.isEmpty else { return }
        highlightIndex = PVFloatingMenuSelection.moveIndex(
            from: highlightIndex, delta: delta, count: options.count
        )
    }

    mutating func jumpHighlight(toFirst: Bool) {
        guard isOpen, !options.isEmpty else { return }
        highlightIndex = toFirst ? 0 : options.count - 1
    }

    mutating func setHighlight(_ index: Int) {
        guard isOpen else { return }
        if index < 0 {
            highlightIndex = -1
        } else if options.indices.contains(index) {
            highlightIndex = index
        }
    }

    /// Single-key cycle vs accumulating prefix, matching `PVContextMenuKeyboard`.
    @discardableResult
    mutating func applyTypeSelect(_ character: Character, now: Date = Date(), commit: Bool) -> Bool {
        guard canOpen else { return false }
        let buffer = typeSelect.append(character, now: now)
        let current = isOpen ? highlightIndex : selectedIndex
        let from = buffer.count == 1 ? current + 1 : max(0, current)
        let hit = PVTypeSelectMatcher.index(in: optionTitles, prefix: buffer, fromIndex: from)
        guard hit >= 0 else { return false }
        if commit {
            selection = options[hit].id
        } else if isOpen {
            highlightIndex = hit
        }
        return true
    }

    // MARK: Pointer transitions

    private mutating func press(_ target: PVSelectPointerTarget) {
        switch target {
        case .trigger:
            open()
        case .outside:
            clickAway()
        case .row:
            break
        }
    }

    private mutating func drag(_ target: PVSelectPointerTarget) {
        guard isOpen else { return }
        switch target {
        case .row(let index):
            setHighlight(index)
        case .trigger, .outside:
            highlightIndex = -1
        }
    }

    private mutating func release(_ target: PVSelectPointerTarget) {
        guard isOpen else { return }
        switch target {
        case .row(let index):
            commit(index: index, close: true)
        case .trigger:
            break
        case .outside:
            dismissRestoring()
        }
    }

    private mutating func clickAway() {
        if isOpen {
            dismissRestoring()
        }
    }
}

/// Keys the focused trigger forwards into `PVSelectSession`.
enum PVSelectKey: Equatable {
    case up
    case down
    case home
    case end
    case space
    case `return`
    case escape
    case character(Character)
}
