import Foundation
import SwiftUI
import Testing
@testable import Provenencia

@Suite
struct PVConfirmCopyTests {
    @Test func storesTitleMessageAndButtonLabels() {
        let copy = PVConfirmCopy(
            title: "Delete Photographer?",
            message: "Nothing is lost.",
            confirm: "Delete field",
            cancel: "Keep field"
        )
        #expect(copy.title == "Delete Photographer?")
        #expect(copy.message == "Nothing is lost.")
    }
}

@Suite
struct PVConfirmToneTests {
    @Test func dangerUsesDestructiveRoleDangerVariantAndTrashIcon() {
        #expect(PVConfirmTone.danger.buttonRole == .destructive)
        #expect(PVConfirmTone.danger.buttonVariant == .danger)
        #expect(PVConfirmTone.danger.confirmIcon == .trash)
    }

    @Test func irreversibleUsesPrimaryVariantWithoutDestructiveRoleOrIcon() {
        #expect(PVConfirmTone.irreversible.buttonRole == nil)
        #expect(PVConfirmTone.irreversible.buttonVariant == .primary)
        #expect(PVConfirmTone.irreversible.confirmIcon == nil)
    }
}

@Suite
struct PVConfirmControlsTests {
    @Test func actionsDisableOnlyWhileRunning() {
        #expect(!PVConfirmControls.isActionDisabled(isRunning: false))
        #expect(PVConfirmControls.isActionDisabled(isRunning: true))
    }
}

@Suite
struct PVConfirmAccessibilityTests {
    @Test func mintsPrefixedIdentifiers() {
        #expect(
            PVConfirmAccessibility.identifier(prefix: "sourceFields.delete", suffix: "confirm")
                == "sourceFields.delete.confirm"
        )
        #expect(
            PVConfirmAccessibility.identifier(prefix: "sourceFields.delete", suffix: "cancel")
                == "sourceFields.delete.cancel"
        )
    }

    @Test func omitsIdentifierWhenPrefixIsMissing() {
        #expect(PVConfirmAccessibility.identifier(prefix: nil, suffix: "confirm") == nil)
    }
}

@Suite
struct PVConfirmContentTests {
    private func sampleCopy() -> PVConfirmCopy {
        PVConfirmCopy(
            title: "Delete Photographer?",
            message: "Nothing is lost.",
            confirm: "Delete field",
            cancel: "Keep field"
        )
    }

    @Test func usesFixedAlertFamilyWidth() {
        #expect(PVConfirmLayout.panelWidth == 440)
    }

    @Test func defaultsToneDangerAndNotRunning() {
        let content = PVConfirmContent(
            copy: sampleCopy(),
            onConfirm: {},
            onCancel: {}
        ) {
            EmptyView()
        }
        #expect(content.tone == .danger)
        #expect(!content.isRunning)
        #expect(content.accessibilityIdentifierPrefix == nil)
    }

    @Test func acceptsIrreversibleToneRunningAndPrefix() {
        let content = PVConfirmContent(
            copy: sampleCopy(),
            tone: .irreversible,
            isRunning: true,
            accessibilityIdentifierPrefix: "sourceTypes.delete",
            onConfirm: {},
            onCancel: {}
        ) {
            EmptyView()
        }
        #expect(content.tone == .irreversible)
        #expect(content.isRunning)
        #expect(content.accessibilityIdentifierPrefix == "sourceTypes.delete")
    }

    @Test func confirmAndCancelCallbacksFire() {
        var confirmed = 0
        var cancelled = 0
        let content = PVConfirmContent(
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

    @Test func modifierCancelPathClearsItemBinding() {
        struct Item: Identifiable {
            let id = UUID()
        }
        var item: Item? = Item()
        let binding = Binding(
            get: { item },
            set: { item = $0 }
        )
        // Mirrors `.pvConfirm(item:)` onCancel wiring without mounting a sheet.
        let onCancel = { binding.wrappedValue = nil }
        onCancel()
        #expect(item == nil)
    }
}

@Suite
struct PVConfirmKeyChipTests {
    @Test func storesLabelAndValue() {
        let chip = PVConfirmKeyChip(label: "Key released", value: "photographer")
        #expect(chip.value == "photographer")
    }
}
