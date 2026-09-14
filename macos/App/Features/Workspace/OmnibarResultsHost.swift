import AppKit
import SwiftUI

// MARK: - Omnibar results host (content-column overlay)

/// Renders the omnibar results panel over the workspace content column —
/// same overlay family as `HistoryJumpMenuHost` (near-square corners, not
/// a floating NSPanel / popover).
struct OmnibarResultsHost: View {
    var results: OmnibarResultsModel
    let projectDir: String
    let onActivate: (CatalogSearchHit) -> Void
    @State private var dismissMonitor: Any?

    var body: some View {
        Group {
            if results.isPresented, results.fieldFrame != .zero {
                OmnibarResultsPanel(
                    results: results,
                    projectDir: projectDir,
                    onActivate: onActivate
                )
                .offset(
                    x: results.fieldFrame.maxX - OmnibarResultsModel.panelWidth,
                    y: results.fieldFrame.maxY + 8
                )
            }
        }
        .onChange(of: results.isPresented) { _, open in
            if open {
                DispatchQueue.main.async { armDismissMonitor() }
            } else {
                removeDismissMonitor()
            }
        }
        .onDisappear {
            results.dismiss()
            removeDismissMonitor()
        }
    }

    private func armDismissMonitor() {
        removeDismissMonitor()
        if NSEvent.pressedMouseButtons != 0 {
            dismissMonitor = NSEvent.addLocalMonitorForEvents(
                matching: [.leftMouseUp, .rightMouseUp, .otherMouseUp]
            ) { event in
                DispatchQueue.main.async { armDismissMonitor() }
                return event
            }
            return
        }

        dismissMonitor = NSEvent.addLocalMonitorForEvents(
            matching: [.leftMouseUp, .keyDown]
        ) { event in
            if event.type == .keyDown {
                switch event.keyCode {
                case 53: // Escape
                    DispatchQueue.main.async { results.closePanel() }
                    return nil
                case 125: // Down
                    DispatchQueue.main.async { results.moveSelection(delta: 1) }
                    return nil
                case 126: // Up
                    DispatchQueue.main.async { results.moveSelection(delta: -1) }
                    return nil
                case 36, 76: // Return / keypad Enter
                    DispatchQueue.main.async {
                        if let hit = results.selectedHit() {
                            onActivate(hit)
                        }
                    }
                    return nil
                default:
                    return event
                }
            }
            DispatchQueue.main.async { results.closePanel() }
            return event
        }
    }

    private func removeDismissMonitor() {
        if let dismissMonitor {
            NSEvent.removeMonitor(dismissMonitor)
            self.dismissMonitor = nil
        }
    }
}

private struct OmnibarResultsPanel: View {
    var results: OmnibarResultsModel
    let projectDir: String
    let onActivate: (CatalogSearchHit) -> Void

    var body: some View {
        Group {
            if results.showsErrorState {
                errorState
            } else if results.showsEmptyState {
                emptyState
            } else if results.showsLoadingState {
                loadingState
            } else {
                // ScrollView expands to the proposed height unless we ask for
                // the ideal (content) height — clamp only when the list is tall.
                ScrollView {
                    VStack(alignment: .leading, spacing: 0) {
                        ForEach(Array(results.hits.enumerated()), id: \.element.id) { index, hit in
                            Button {
                                onActivate(hit)
                            } label: {
                                OmnibarHitRowView(
                                    hit: hit,
                                    projectDir: projectDir,
                                    selected: index == results.selectedIndex
                                )
                            }
                            .buttonStyle(OmnibarHitRowButtonStyle(
                                onHover: { hovering in
                                    if hovering { results.selectedIndex = index }
                                }
                            ))
                            .accessibilityLabel(
                                "\(hit.title), \(String(localized: OmnibarHitPresentation.kindLabel(for: hit.kind)))"
                            )
                            .accessibilityIdentifier("workspace.toolbar.omnibar.hit.\(hit.id)")
                        }
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                }
                .frame(maxHeight: 420)
                .fixedSize(horizontal: false, vertical: true)
            }
        }
        .padding(PVSpacing.space2)
        .frame(width: OmnibarResultsModel.panelWidth, alignment: .leading)
        .fixedSize(horizontal: false, vertical: true)
        .background(PVColor.surfaceCard)
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(PVColor.borderSubtle, lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .pvShadow(PVElevation.overlay)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.Workspace.omnibarResultsAccessibilityLabel))
        .accessibilityIdentifier("workspace.toolbar.omnibar.results")
    }

    private var loadingState: some View {
        HStack(spacing: PVSpacing.space3) {
            ProgressView()
                .controlSize(.small)
            Text(L10n.Workspace.omnibarSearching)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textMuted)
        }
        .padding(.horizontal, 9)
        .padding(.vertical, 12)
        .frame(maxWidth: .infinity, alignment: .leading)
        .accessibilityIdentifier("workspace.toolbar.omnibar.loading")
    }

    private var errorState: some View {
        PVCallout(
            tone: .danger,
            message: results.searchError ?? String(localized: L10n.Workspace.omnibarSearchFailed)
        )
        .padding(.horizontal, 9)
        .padding(.vertical, 8)
        .accessibilityIdentifier("workspace.toolbar.omnibar.error")
    }

    private var emptyState: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            Text(L10n.Workspace.omnibarNoMatchesTitle(query: results.trimmedQuery))
                .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textPrimary)
            Text(L10n.Workspace.omnibarNoMatchesHint)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
        }
        .padding(.horizontal, 9)
        .padding(.vertical, 12)
        .frame(maxWidth: .infinity, alignment: .leading)
        .accessibilityIdentifier("workspace.toolbar.omnibar.noMatches")
    }
}

/// Resolves lead thumbnail + fills `PVOmnibarHitRow` slots for one hit.
private struct OmnibarHitRowView: View {
    let hit: CatalogSearchHit
    let projectDir: String
    var selected: Bool

    var body: some View {
        PVOmnibarHitRow(
            lead: { leadView },
            title: hit.title,
            kindLabel: OmnibarHitPresentation.kindLabel(for: hit.kind),
            secondary: hit.subtitle,
            ref: hit.ref,
            refAccent: OmnibarHitPresentation.refAccent(for: hit),
            matchContext: OmnibarHitPresentation.showMatchContext(hit.matchReason) ? hit.matchReason : "",
            selected: selected
        )
    }

    @ViewBuilder
    private var leadView: some View {
        switch hit.kind {
        case "source":
            CachedThumbnail(
                projectDir: projectDir,
                relPath: hit.thumbnailRelPath,
                typeIconKey: hit.iconKey.isEmpty ? nil : hit.iconKey,
                size: 40,
                cornerRadius: PVRadius.sm
            )
        case "source_type":
            if hit.iconKey.isEmpty {
                PVThumbnail(PVThumbnail.Content(icon: .library), size: 40)
            } else {
                PVThumbnail(
                    PVThumbnail.Content(
                        evidenceIcon: PVEvidenceIconKey(catalogKey: hit.iconKey),
                        label: OmnibarHitPresentation.kindLabel(for: hit.kind)
                    ),
                    size: 40
                )
            }
        default:
            PVThumbnail(
                PVThumbnail.Content(icon: OmnibarHitPresentation.leadSymbol(for: hit.kind)),
                size: 40
            )
        }
    }
}

private struct OmnibarHitRowButtonStyle: ButtonStyle {
    var onHover: (Bool) -> Void

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .opacity(configuration.isPressed ? 0.92 : 1)
            .onHover { onHover($0) }
    }
}

struct OmnibarFieldAnchorKey: PreferenceKey {
    static let defaultValue: CGRect = .zero

    static func reduce(value: inout CGRect, nextValue: () -> CGRect) {
        let next = nextValue()
        if next != .zero {
            value = next
        }
    }
}
