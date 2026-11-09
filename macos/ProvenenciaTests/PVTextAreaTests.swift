import AppKit
import Testing
@testable import Provenencia

@Suite
@MainActor
struct PVTextAreaTests {
    @Test func tabAndShiftTabLeaveTheField() {
        #expect(PVTextAreaTabRouting.movesFocus(#selector(NSResponder.insertTab(_:))))
        #expect(PVTextAreaTabRouting.movesFocus(#selector(NSResponder.insertBacktab(_:))))
        #expect(!PVTextAreaTabRouting.movesFocus(#selector(NSResponder.insertTabIgnoringFieldEditor(_:))))
        #expect(!PVTextAreaTabRouting.movesFocus(#selector(NSResponder.insertNewline(_:))))
    }

    @Test func routerMovesForTabAndForwardsEverythingElse() {
        let fixture = RouterFixture()

        let tab = fixture.send(#selector(NSResponder.insertTab(_:)))
        let backtab = fixture.send(#selector(NSResponder.insertBacktab(_:)))
        let optionTab = fixture.send(#selector(NSResponder.insertTabIgnoringFieldEditor(_:)))
        let newline = fixture.send(#selector(NSResponder.insertNewline(_:)))

        #expect(tab == true)
        #expect(backtab == true)
        #expect(optionTab == false)
        #expect(newline == false)
        #expect(fixture.spy.commands == [
            #selector(NSResponder.insertTabIgnoringFieldEditor(_:)),
            #selector(NSResponder.insertNewline(_:)),
        ])
    }

    @Test func tabAndBacktabCallTheNeighbourWhenSet() {
        let fixture = RouterFixture()
        var tabs = 0
        var backtabs = 0
        fixture.router.onTab = { tabs += 1 }
        fixture.router.onBacktab = { backtabs += 1 }

        #expect(fixture.send(#selector(NSResponder.insertTab(_:))) == true)
        #expect(tabs == 1)
        #expect(backtabs == 0)

        #expect(fixture.send(#selector(NSResponder.insertBacktab(_:))) == true)
        #expect(tabs == 1)
        #expect(backtabs == 1)
        #expect(fixture.spy.commands.isEmpty)
    }

    @Test func noNeighbourKeepsTodaysRouting() {
        let fixture = RouterFixture()
        var backtabs = 0
        fixture.router.onBacktab = { backtabs += 1 }

        let tab = fixture.send(#selector(NSResponder.insertTab(_:)))
        let optionTab = fixture.send(#selector(NSResponder.insertTabIgnoringFieldEditor(_:)))
        let newline = fixture.send(#selector(NSResponder.insertNewline(_:)))

        #expect(tab == true)
        #expect(backtabs == 0)
        #expect(optionTab == false)
        #expect(newline == false)
        #expect(fixture.spy.commands == [
            #selector(NSResponder.insertTabIgnoringFieldEditor(_:)),
            #selector(NSResponder.insertNewline(_:)),
        ])
    }

    @Test func textViewSearchSkipsAFieldEditor() {
        let fieldEditor = NSTextView()
        fieldEditor.isFieldEditor = true
        let plain = NSTextView()
        let both = NSView()
        both.frame = NSRect(x: 0, y: 0, width: 100, height: 100)
        fieldEditor.frame = both.bounds
        plain.frame = both.bounds
        both.addSubview(fieldEditor)
        both.addSubview(plain)
        let point = NSPoint(x: 50, y: 50)
        #expect(both.descendantTextView(containing: point) === plain)

        let only = NSView()
        let editor = NSTextView()
        editor.isFieldEditor = true
        only.frame = NSRect(x: 0, y: 0, width: 100, height: 100)
        editor.frame = only.bounds
        only.addSubview(editor)
        #expect(only.descendantTextView(containing: point) == nil)
    }

    @Test func twoRoutersOnOneTextViewDoNotLoop() {
        let fixture = RouterFixture()
        let other = PVTextAreaTabRouter()
        other.attach(to: fixture.textView)
        fixture.router.attach(to: fixture.textView)

        let selector = #selector(NSTextDelegate.textDidChange(_:))
        #expect(fixture.textView.delegate === fixture.router)
        #expect(fixture.textView.delegate?.responds(to: selector) == true)
        #expect(fixture.send(#selector(NSResponder.insertTab(_:))) == true)
        let notification = Notification(name: NSText.didChangeNotification, object: fixture.textView)
        _ = fixture.textView.delegate?.perform(selector, with: notification)
        #expect(fixture.spy.changeCount == 1)
    }

    @Test func textViewSearchPicksTheTextViewUnderTheArea() {
        let container = NSView(frame: NSRect(x: 0, y: 0, width: 200, height: 200))
        let first = NSTextView(frame: NSRect(x: 0, y: 0, width: 200, height: 80))
        let second = NSTextView(frame: NSRect(x: 0, y: 0, width: 200, height: 80))
        let firstScroll = NSScrollView(frame: NSRect(x: 0, y: 0, width: 200, height: 80))
        let secondScroll = NSScrollView(frame: NSRect(x: 0, y: 100, width: 200, height: 80))
        firstScroll.documentView = first
        secondScroll.documentView = second
        container.addSubview(firstScroll)
        container.addSubview(secondScroll)
        let marker = NSView(frame: NSRect(x: 0, y: 100, width: 200, height: 80))
        container.addSubview(marker)

        let midpoint = marker.convert(NSPoint(x: marker.bounds.midX, y: marker.bounds.midY), to: nil)
        #expect(container.descendantTextView(containing: midpoint) === second)
    }

    @Test func routerStillForwardsEditsToTheOriginalDelegate() {
        let fixture = RouterFixture()

        let selector = #selector(NSTextDelegate.textDidChange(_:))
        let notification = Notification(name: NSText.didChangeNotification, object: fixture.textView)
        #expect(fixture.textView.delegate?.responds(to: selector) == true)
        _ = fixture.textView.delegate?.perform(selector, with: notification)
        #expect(fixture.spy.changeCount == 1)
    }
}

/// A text view whose delegate is a spy with the router attached in front of it.
@MainActor
private struct RouterFixture {
    let textView = NSTextView()
    let spy = TextViewCommandSpy()
    let router = PVTextAreaTabRouter()

    init() {
        textView.delegate = spy
        router.attach(to: textView)
    }

    func send(_ selector: Selector) -> Bool? {
        textView.delegate?.textView?(textView, doCommandBy: selector)
    }
}

private final class TextViewCommandSpy: NSObject, NSTextViewDelegate {
    var commands: [Selector] = []
    var changeCount = 0

    func textDidChange(_ notification: Notification) {
        changeCount += 1
    }

    func textView(_ textView: NSTextView, doCommandBy commandSelector: Selector) -> Bool {
        commands.append(commandSelector)
        return false
    }
}
