import AppKit

/// Text measurement for code that draws with AppKit instead of SwiftUI: a
/// card painted into a canvas item must size its text with the rule it draws
/// with, so layout and paint cannot disagree.
enum PVTextMetrics {
    /// One line of `font`, never below a point.
    static func fontLineHeight(_ font: NSFont) -> CGFloat {
        max(font.boundingRectForFont.height, 1)
    }

    /// The height `text` wraps to at `width`, capped at `maxLines` lines. Empty
    /// text measures one line.
    static func textHeight(_ text: String, width: CGFloat, font: NSFont, maxLines: Int) -> CGFloat {
        if text.isEmpty { return fontLineHeight(font) }
        let style = NSMutableParagraphStyle()
        style.lineBreakMode = .byWordWrapping
        let rect = (text as NSString).boundingRect(
            with: CGSize(width: max(width, 1), height: .greatestFiniteMagnitude),
            options: [.usesLineFragmentOrigin, .usesFontLeading],
            attributes: [.font: font, .paragraphStyle: style]
        )
        return min(ceil(rect.height), ceil(fontLineHeight(font)) * CGFloat(maxLines))
    }
}
