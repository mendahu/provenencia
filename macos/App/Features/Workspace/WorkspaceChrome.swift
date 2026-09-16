import SwiftUI

/// Layout constants genuinely shared across files: `WorkspaceSidebar`'s
/// brand row and `WorkspaceToolbar` so their hairline rules read as one
/// continuous line across the whole window width (S2-01 Frame 7 decision),
/// and `TrafficLights.swift` so the system-drawn stoplights align to that
/// same row instead of the row aligning to them.
///
/// Constants that only one file actually reads — the sidebar's own
/// traffic-light clearance padding, the collapsed rail width, the
/// stoplights' own horizontal shift — live as private constants next to
/// their one consumer instead of here, so this file stays a true
/// cross-file contract rather than a general dumping ground.
enum WorkspaceChrome {
    /// Row height for both headers. Taller than a plain 46pt hairline row
    /// to give the fixed-position, system-drawn traffic lights some
    /// breathing room instead of crowding them — closer to the unified
    /// toolbar height apps like Xcode and Mail use for a sidebar this
    /// wide, not the source board's flat 46px mockup value (its "browser
    /// chrome" traffic lights are drawn dots, not real hit targets).
    static let headerHeight: CGFloat = 52
    /// Offset from this row's true vertical center (`headerHeight / 2`)
    /// for both the header text and — via `TrafficLights` — the stoplight
    /// buttons themselves, so every element on the row (lights, logo,
    /// title, "Sources", project identity) shares one baseline. `0` means
    /// dead center between the window's top edge and the row's bottom
    /// hairline; adjust here to bias the whole row up or down.
    static let verticalNudge: CGFloat = 0
}

extension View {
    /// Applies `WorkspaceChrome`'s shared header-row recipe — the fixed
    /// row height, the vertical nudge aligning this row's content with
    /// `TrafficLights`' stoplights, and the bottom hairline that makes the
    /// sidebar and content headers read as one continuous row across the
    /// window (S2-01 Frame 7).
    ///
    /// - Parameter windowDrag: Where the row's `WindowDragRegion` sits.
    ///   `.overlay` (default) covers the whole row and suits chrome
    ///   **without** interactive controls (sidebar brand). `WorkspaceToolbar`
    ///   passes `.behindContent` so the full 52pt row drags while
    ///   Back/Forward, breadcrumbs, and the omnibar keep their own clicks.
    func pvWorkspaceHeaderRow(windowDrag: PVWindowDragPlacement = .overlay) -> some View {
        offset(y: WorkspaceChrome.verticalNudge)
            .frame(height: WorkspaceChrome.headerHeight)
            .overlay(alignment: .bottom) {
                PVDivider()
            }
            .pvWindowDragRegion(windowDrag)
    }
}
