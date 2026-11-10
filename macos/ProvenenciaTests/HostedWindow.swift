import AppKit
import SwiftUI
import Testing

/// Hosts `view` in a borderless window so layout, `.task` and geometry
/// callbacks run. The hosting view is sized after it becomes the content view.
/// The window is never shown. Mark the suite `.hostedWindows`: the trait keeps
/// the window alive for the test and closes it when the test ends.
@MainActor
func hostInBorderlessWindow<V: View>(
    _ view: V,
    windowSize: CGSize,
    hostSize: CGSize? = nil
) -> (NSWindow, NSHostingView<V>) {
    let host = NSHostingView(rootView: view)
    let window = testWindow(size: windowSize)
    window.contentView = host
    host.frame = NSRect(origin: .zero, size: hostSize ?? windowSize)
    return (window, host)
}

/// Hosts an AppKit `view` as the content view of a borderless window of `size`.
/// Never shown; the `.hostedWindows` trait closes it.
@MainActor
func hostInBorderlessWindow(nsView view: NSView, size: CGSize) -> NSWindow {
    let window = testWindow(size: size)
    window.contentView = view
    return window
}

/// The one builder for test windows: never ordered in, tracked by the
/// `.hostedWindows` trait.
@MainActor
func testWindow(size: CGSize, styleMask: NSWindow.StyleMask = .borderless) -> NSWindow {
    trackTestWindow(
        NSWindow(
            contentRect: NSRect(origin: .zero, size: size),
            styleMask: styleMask,
            backing: .buffered,
            defer: false
        )
    )
}

/// Registers `window` with the running test's tracker so the trait closes it.
/// Without the `.hostedWindows` trait the test records an issue.
@MainActor
@discardableResult
func trackTestWindow<W: NSWindow>(_ window: W) -> W {
    if let windows = HostedWindowScope.current {
        windows.register(window)
    } else {
        Issue.record("Build test windows under the .hostedWindows trait.")
    }
    return window
}

/// The windows one test built; closed and dropped when the test ends.
@MainActor
final class HostedWindows {
    private(set) var windows: [NSWindow] = []

    func register(_ window: NSWindow) {
        window.isReleasedWhenClosed = false
        windows.append(window)
    }

    func closeAll() {
        windows.forEach { $0.close() }
        windows.removeAll()
    }
}

enum HostedWindowScope {
    @TaskLocal static var current: HostedWindows?
}

/// Gives each test case its own `HostedWindows` and closes it after the body,
/// passing or throwing.
struct HostedWindowsTrait: TestTrait, SuiteTrait, TestScoping {
    var isRecursive: Bool { true }

    func scopeProvider(for test: Test, testCase: Test.Case?) -> Self? {
        testCase == nil ? nil : self
    }

    func provideScope(
        for test: Test,
        testCase: Test.Case?,
        performing function: @Sendable () async throws -> Void
    ) async throws {
        let windows = HostedWindows()
        do {
            try await HostedWindowScope.$current.withValue(windows) { try await function() }
        } catch {
            await windows.closeAll()
            throw error
        }
        await windows.closeAll()
    }
}

extension Trait where Self == HostedWindowsTrait {
    static var hostedWindows: Self { Self() }
}
