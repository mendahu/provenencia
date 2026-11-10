import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@Suite(.hostedWindows) @MainActor
struct HostedWindowTests {
    @Test func hostedWindowIsNeverShownAndIsTracked() {
        let (window, _) = hostInBorderlessWindow(Text(verbatim: "x"), windowSize: CGSize(width: 200, height: 100))
        #expect(!window.isVisible)
        #expect(HostedWindowScope.current?.windows.count == 1)
    }

    @Test func closeAllClosesAndForgets() {
        let fresh = HostedWindows()
        let closes = CloseCounter()
        HostedWindowScope.$current.withValue(fresh) {
            _ = testWindow(size: CGSize(width: 10, height: 10))
            _ = testWindow(size: CGSize(width: 10, height: 10))
        }
        let observers = fresh.windows.map { window in
            NotificationCenter.default.addObserver(
                forName: NSWindow.willCloseNotification, object: window, queue: nil
            ) { _ in closes.count += 1 }
        }
        defer { observers.forEach(NotificationCenter.default.removeObserver) }
        fresh.closeAll()
        #expect(closes.count == 2)
        #expect(fresh.windows.isEmpty)
    }

    @Test func buildingWithoutTheTraitRecordsAnIssue() {
        var window: NSWindow?
        withKnownIssue {
            window = HostedWindowScope.$current.withValue(nil) {
                testWindow(size: CGSize(width: 10, height: 10))
            }
        }
        window?.isReleasedWhenClosed = false
        window?.close()
    }
}

private final class CloseCounter: @unchecked Sendable {
    var count = 0
}
