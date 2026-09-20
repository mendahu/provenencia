import Foundation
import SwiftUI
import Testing
@testable import Provenencia

@Suite
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
