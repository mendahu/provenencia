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
        let textView = NSTextView()
        let spy = TextViewCommandSpy()
        textView.delegate = spy
        let router = PVTextAreaTabRouter()
        router.attach(to: textView)

        let tab = textView.delegate?.textView?(textView, doCommandBy: #selector(NSResponder.insertTab(_:)))
        let backtab = textView.delegate?.textView?(textView, doCommandBy: #selector(NSResponder.insertBacktab(_:)))
        let optionTab = textView.delegate?.textView?(
            textView,
            doCommandBy: #selector(NSResponder.insertTabIgnoringFieldEditor(_:))
        )
        let newline = textView.delegate?.textView?(textView, doCommandBy: #selector(NSResponder.insertNewline(_:)))

        #expect(tab == true)
        #expect(backtab == true)
        #expect(optionTab == false)
        #expect(newline == false)
        #expect(spy.commands == [
            #selector(NSResponder.insertTabIgnoringFieldEditor(_:)),
            #selector(NSResponder.insertNewline(_:)),
        ])
    }

    @Test func tabAndBacktabCallTheNeighbourWhenSet() {
        let textView = NSTextView()
        let spy = TextViewCommandSpy()
        textView.delegate = spy
        let router = PVTextAreaTabRouter()
        var tabs = 0
        var backtabs = 0
        router.onTab = { tabs += 1 }
        router.onBacktab = { backtabs += 1 }
        router.attach(to: textView)

        let tab = textView.delegate?.textView?(textView, doCommandBy: #selector(NSResponder.insertTab(_:)))
        #expect(tab == true)
        #expect(tabs == 1)
        #expect(backtabs == 0)

        let backtab = textView.delegate?.textView?(textView, doCommandBy: #selector(NSResponder.insertBacktab(_:)))
        #expect(backtab == true)
        #expect(tabs == 1)
        #expect(backtabs == 1)
        #expect(spy.commands.isEmpty)
    }

    @Test func noNeighbourKeepsTodaysRouting() {
        let textView = NSTextView()
        let spy = TextViewCommandSpy()
        textView.delegate = spy
        let router = PVTextAreaTabRouter()
        var backtabs = 0
        router.onBacktab = { backtabs += 1 }
        router.attach(to: textView)

        let tab = textView.delegate?.textView?(textView, doCommandBy: #selector(NSResponder.insertTab(_:)))
        let optionTab = textView.delegate?.textView?(
            textView,
            doCommandBy: #selector(NSResponder.insertTabIgnoringFieldEditor(_:))
        )
        let newline = textView.delegate?.textView?(textView, doCommandBy: #selector(NSResponder.insertNewline(_:)))

        #expect(tab == true)
        #expect(backtabs == 0)
        #expect(optionTab == false)
        #expect(newline == false)
        #expect(spy.commands == [
            #selector(NSResponder.insertTabIgnoringFieldEditor(_:)),
            #selector(NSResponder.insertNewline(_:)),
        ])
    }

    @Test func routerStillForwardsEditsToTheOriginalDelegate() {
        let textView = NSTextView()
        let spy = TextViewCommandSpy()
        textView.delegate = spy
        PVTextAreaTabRouter().attach(to: textView)

        let selector = #selector(NSTextDelegate.textDidChange(_:))
        let notification = Notification(name: NSText.didChangeNotification, object: textView)
        #expect(textView.delegate?.responds(to: selector) == true)
        _ = textView.delegate?.perform(selector, with: notification)
        #expect(spy.changeCount == 1)
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
