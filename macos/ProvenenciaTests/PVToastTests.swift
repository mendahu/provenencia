import Testing
@testable import Provenencia

@MainActor
struct PVToastTests {
    /// A toast replaced before its time cancels its wait; the cancelled wait must not
    /// dismiss the toast that replaced it.
    @Test func aCancelledWaitDismissesNothing() async {
        var dismissed = false
        let wait = Task { await PVToast.autoDismiss(after: 60, onDismiss: { dismissed = true }) }
        wait.cancel()
        await wait.value
        #expect(!dismissed)
    }

    @Test func aWaitThatRunsOutDismisses() async {
        var dismissed = false
        await PVToast.autoDismiss(after: 0.01, onDismiss: { dismissed = true })
        #expect(dismissed)
    }
}
