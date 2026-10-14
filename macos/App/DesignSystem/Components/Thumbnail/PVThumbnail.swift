import SwiftUI

/// Fixed preview tile for evidence / media browse rows — mirrors the design
/// system's `components/core/Thumbnail.jsx` (and the S2-04 board's
/// `PVThumbnail` extraction). Sources list rows use it now; Artifacts and
/// Files reuse it later.
///
/// States: **image**, **mark** (`file_*` / `type_*` / `subject_*`), **SF glyph**,
/// **empty** (em dash), **loading** (skeleton bar). A missing image is the empty
/// placeholder or a typed MIME stand-in — never an error UI.
struct PVThumbnail: View {
    /// How a list row asks for a thumbnail — the list owns framing; the
    /// feature decides which state each row needs.
    struct Content {
        var image: Image?
        var glyph: PVGlyph?
        var loading: Bool
        var label: LocalizedStringResource?

        init(
            image: Image? = nil,
            glyph: PVGlyph? = nil,
            loading: Bool = false,
            label: LocalizedStringResource? = nil
        ) {
            self.image = image
            self.glyph = glyph
            self.loading = loading
            self.label = label
        }

        static let empty = Content()
    }

    private enum State {
        case image
        case mark(PVMarkKey)
        case symbol(PVSymbol)
        case empty
        case loading
    }

    private let content: Content
    private let size: CGFloat
    private let cornerRadius: CGFloat

    init(
        _ content: Content = .empty,
        size: CGFloat = 44,
        cornerRadius: CGFloat = PVRadius.sm
    ) {
        self.content = content
        self.size = size
        self.cornerRadius = cornerRadius
    }

    private var state: State {
        if content.loading { return .loading }
        if content.image != nil { return .image }
        switch content.glyph {
        case .mark(let key)?: return .mark(key)
        case .symbol(let icon)?: return .symbol(icon)
        case nil: return .empty
        }
    }

    private var isFilled: Bool {
        switch state {
        case .image, .mark, .symbol: true
        case .empty, .loading: false
        }
    }

    var body: some View {
        ZStack {
            switch state {
            case .image:
                if let image = content.image {
                    image
                        .resizable()
                        .scaledToFill()
                        .frame(width: size, height: size)
                        .clipped()
                }
            case .mark(let key):
                PVMark(key, size: size * 0.82, decorative: content.label == nil)
                    .foregroundStyle(PVColor.textSecondary)
            case .symbol(let icon):
                PVIcon(icon, size: max(12, (size * 0.41).rounded()))
                    .foregroundStyle(PVColor.textFaint)
            case .empty:
                Text(verbatim: "—")
                    .font(PVFont.mono(size: max(11, (size * 0.3).rounded())))
                    .foregroundStyle(PVColor.textFaint)
                    .accessibilityHidden(true)
            case .loading:
                Capsule()
                    .fill(PVColor.borderDefault)
                    .frame(height: 2)
                    .padding(.horizontal, size * 0.18)
                    .frame(maxHeight: .infinity, alignment: .bottom)
                    .padding(.bottom, size * 0.18)
                    .opacity(0.55)
                    .accessibilityHidden(true)
            }
        }
        .frame(width: size, height: size)
        .background(isFilled ? PVColor.surfaceInset : PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: cornerRadius, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                .strokeBorder(isFilled ? PVColor.borderDefault : PVColor.borderSubtle, lineWidth: 1)
        )
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(Text(accessibilityLabel))
    }

    private var accessibilityLabel: LocalizedStringResource {
        if let label = content.label { return label }
        switch state {
        case .mark(let key): return key.accessibilityName
        case .loading: return L10n.DesignSystem.thumbnailLoading
        case .empty, .symbol, .image: return L10n.DesignSystem.thumbnailEmpty
        }
    }
}

#Preview {
    HStack(spacing: PVSpacing.space6) {
        PVThumbnail(.empty)
        PVThumbnail(PVThumbnail.Content(glyph: .symbol(.photo)))
        PVThumbnail(PVThumbnail.Content(glyph: .mark(.filePDF)))
        PVThumbnail(PVThumbnail.Content(loading: true))
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
