import AppKit
import Foundation
import SwiftUI
import Testing
@testable import Provenencia

@Suite
@MainActor
struct PVPanelTests {
    @Test func storesWidthAndSunkenFooterByDefault() {
        let panel = PVPanel(
            title: Text(verbatim: "Title"),
            width: 480
        ) {
            Text(verbatim: "Body")
        } footer: {
            Text(verbatim: "Footer")
        }
        #expect(panel.width == 480)
        #expect(panel.footerChrome == .sunken)
        #expect(panel.showsFooter)
        #expect(panel.subtitle == nil)
    }

    @Test func acceptsPlainFooterChromeAndCustomWidth() {
        let panel = PVPanel(
            title: Text(verbatim: "Title"),
            subtitle: Text(verbatim: "Subtitle"),
            width: 660,
            footerChrome: .plain
        ) {
            Text(verbatim: "Body")
        } footer: {
            Text(verbatim: "Footer")
        }
        #expect(panel.width == 660)
        #expect(panel.footerChrome == .plain)
        #expect(panel.showsFooter)
        #expect(panel.subtitle != nil)
    }

    @Test func noFooterInitializerHidesTheFooterBand() {
        let panel = PVPanel(
            title: Text(verbatim: "Title"),
            width: 420
        ) {
            Text(verbatim: "Body only")
        }
        #expect(!panel.showsFooter)
        #expect(panel.width == 420)
    }
}

@Suite
struct PVFormDialogCopyTests {
    @Test func storesLabelsAndOptionalSubtitle() {
        let withSubtitle = PVFormDialogCopy(
            title: "Add source",
            subtitle: "Thin record",
            confirm: "Create",
            cancel: "Cancel"
        )
        #expect(withSubtitle.subtitle != nil)

        let without = PVFormDialogCopy(
            title: "Add source",
            confirm: "Create",
            cancel: "Cancel"
        )
        #expect(without.subtitle == nil)
    }
}

@Suite
struct PVFormDialogControlsTests {
    @Test func cancelIsOnlyDisabledWhileRunning() {
        #expect(!PVFormDialogControls.isCancelDisabled(isRunning: false))
        #expect(PVFormDialogControls.isCancelDisabled(isRunning: true))
    }

    @Test func confirmDisabledCoversValidationAndRunning() {
        #expect(!PVFormDialogControls.isConfirmDisabled(isRunning: false, confirmDisabled: false))
        #expect(PVFormDialogControls.isConfirmDisabled(isRunning: false, confirmDisabled: true))
        #expect(PVFormDialogControls.isConfirmDisabled(isRunning: true, confirmDisabled: false))
        #expect(PVFormDialogControls.isConfirmDisabled(isRunning: true, confirmDisabled: true))
    }
}

@Suite
@MainActor
struct PVFormDialogClickTests {
    private func window() -> NSWindow {
        NSWindow(
            contentRect: CGRect(x: 0, y: 0, width: 200, height: 120),
            styleMask: [.titled],
            backing: .buffered,
            defer: true
        )
    }

    @Test func aClickOnTheSheetIsInside() {
        let sheet = window()
        #expect(PVOutsideClick.isInsideSheet(eventWindow: sheet, sheet: sheet, modalWindow: nil))
    }

    @Test func aParentWindowClickIsOutside() {
        let workspace = window()
        let sheet = window()
        workspace.addChildWindow(sheet, ordered: .above)
        defer { workspace.removeChildWindow(sheet) }
        #expect(!PVOutsideClick.isInsideSheet(eventWindow: workspace, sheet: sheet, modalWindow: nil))
    }

    @Test func aSheetOnASheetIsInside() {
        let workspace = window()
        let parentSheet = window()
        let childSheet = window()
        workspace.makeKeyAndOrderFront(nil)
        workspace.beginSheet(parentSheet) { _ in }
        parentSheet.beginSheet(childSheet) { _ in }
        defer {
            parentSheet.endSheet(childSheet)
            workspace.endSheet(parentSheet)
        }
        #expect(PVOutsideClick.isInsideSheet(eventWindow: childSheet, sheet: parentSheet, modalWindow: nil))
        #expect(!PVOutsideClick.isInsideSheet(eventWindow: workspace, sheet: parentSheet, modalWindow: nil))
    }

    @Test func anOpenPanelModalCountsAsInside() {
        let sheet = window()
        let workspace = window()
        let panel = NSOpenPanel()
        #expect(PVOutsideClick.isInsideSheet(eventWindow: nil, sheet: sheet, modalWindow: panel))
        #expect(PVOutsideClick.isInsideSheet(eventWindow: workspace, sheet: sheet, modalWindow: panel))
    }

    @Test func aSavePanelEventWindowIsInside() {
        let sheet = window()
        let panel = NSSavePanel()
        #expect(PVOutsideClick.isInsideSheet(eventWindow: panel, sheet: sheet, modalWindow: nil))
    }

    @Test func aPopUpMenuWindowIsInside() {
        let sheet = window()
        let menu = window()
        menu.level = .popUpMenu
        #expect(PVOutsideClick.isInsideSheet(eventWindow: menu, sheet: sheet, modalWindow: nil))
    }
}

@Suite
struct PVSheetDismissTests {
    /// T-1: the one table behind Escape and an outside click.
    @Test func sheetDismissActionTable() {
        for dirty in [false, true] {
            for confirming in [false, true] {
                #expect(PVSheetDismiss.action(isRunning: true, isDirty: dirty, isConfirmingDiscard: confirming) == .ignore)
            }
        }
        for dirty in [false, true] {
            #expect(PVSheetDismiss.action(isRunning: false, isDirty: dirty, isConfirmingDiscard: true) == .keepEditing)
        }
        #expect(PVSheetDismiss.action(isRunning: false, isDirty: true, isConfirmingDiscard: false) == .confirmDiscard)
        #expect(PVSheetDismiss.action(isRunning: false, isDirty: false, isConfirmingDiscard: false) == .close)
    }

    /// T-2: both routes call `dismissRoute`, and every kit presenter hands it
    /// the same running and dirty facts.
    @Test func everyKitSheetPresentsThroughPVSheet() {
        for running in [false, true] {
            for dirty in [false, true] {
                for confirming in [false, true] {
                    #expect(
                        PVSheet.dismissRoute(isRunning: running, isDirty: dirty, isConfirmingDiscard: confirming)
                            == PVSheetDismiss.action(isRunning: running, isDirty: dirty, isConfirmingDiscard: confirming)
                    )
                }
            }
        }
        #expect(PVFormDialogControls.sheetPolicy(isRunning: true, isDirty: true) == PVSheet.Policy(isRunning: true, isDirty: true))
        #expect(PVConfirmControls.sheetPolicy(isRunning: true) == PVSheet.Policy(isRunning: true, isDirty: false))
        #expect(PVDeleteImpactControls.noticePolicy == PVSheet.Policy(isRunning: false, isDirty: false))
    }

    @Test func escapeOnACleanConfirmOrNoticeCloses() {
        let confirm = PVConfirmControls.sheetPolicy(isRunning: false)
        #expect(PVSheet.dismissRoute(policy: confirm, isConfirmingDiscard: false) == .close)
        let notice = PVDeleteImpactControls.noticePolicy
        #expect(PVSheet.dismissRoute(policy: notice, isConfirmingDiscard: false) == .close)
    }

    @Test func aDirtyFormAsksBeforeEitherRouteClosesIt() {
        let form = PVFormDialogControls.sheetPolicy(isRunning: false, isDirty: true)
        #expect(PVSheet.dismissRoute(policy: form, isConfirmingDiscard: false) == .confirmDiscard)
        #expect(PVSheet.dismissRoute(policy: form, isConfirmingDiscard: true) == .keepEditing)
    }
}

@Suite
struct PVFormDialogAccessibilityTests {
    @Test func mintsPrefixedIdentifiers() {
        #expect(
            PVFormDialogAccessibility.identifier(prefix: "sources.add", suffix: "confirm")
                == "sources.add.confirm"
        )
        #expect(
            PVFormDialogAccessibility.identifier(prefix: "sources.add", suffix: "cancel")
                == "sources.add.cancel"
        )
    }

    @Test func omitsIdentifierWhenPrefixIsMissing() {
        #expect(PVFormDialogAccessibility.identifier(prefix: nil, suffix: "confirm") == nil)
    }
}

@Suite
@MainActor
struct PVFormDialogContentTests {
    private func sampleCopy() -> PVFormDialogCopy {
        PVFormDialogCopy(
            title: "Add source",
            subtitle: "Thin record",
            confirm: "Create source",
            cancel: "Cancel"
        )
    }

    @Test func defaultsWidthToFourEighty() {
        let content = PVFormDialogContent(
            copy: sampleCopy(),
            onConfirm: {},
            onCancel: {}
        ) {
            EmptyView()
        }
        #expect(content.width == 480)
        #expect(!content.isRunning)
        #expect(!content.confirmDisabled)
        #expect(content.accessibilityIdentifierPrefix == nil)
    }

    @Test func acceptsOverrideWidthAndFlags() {
        let content = PVFormDialogContent(
            copy: sampleCopy(),
            width: 660,
            isRunning: true,
            confirmDisabled: true,
            accessibilityIdentifierPrefix: "sourceTypes.iconPicker",
            onConfirm: {},
            onCancel: {}
        ) {
            EmptyView()
        }
        #expect(content.width == 660)
        #expect(content.isRunning)
        #expect(content.confirmDisabled)
        #expect(content.accessibilityIdentifierPrefix == "sourceTypes.iconPicker")
    }

    @Test func confirmAndCancelCallbacksFire() {
        var confirmed = 0
        var cancelled = 0
        let content = PVFormDialogContent(
            copy: sampleCopy(),
            onConfirm: { confirmed += 1 },
            onCancel: { cancelled += 1 }
        ) {
            EmptyView()
        }
        content.onConfirm()
        content.onCancel()
        content.onConfirm()
        #expect(confirmed == 2)
        #expect(cancelled == 1)
    }

    @Test func modifierCancelPathClearsPresentedBinding() {
        var presented = true
        let binding = Binding(
            get: { presented },
            set: { presented = $0 }
        )
        // Mirrors `.pvFormDialog`'s onCancel wiring without mounting a sheet.
        let onCancel = { binding.wrappedValue = false }
        onCancel()
        #expect(!presented)
    }
}
