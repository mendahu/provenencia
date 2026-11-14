import SwiftUI

/// Choose-file / drop row for the Artifact sheet's File field and the Add file dialog.
/// Bound to one `IngestFileChoice`, it draws the refused-file Callout under its caption.
struct IngestFileDropRow: View {
    let fileName: String?
    let reject: L10n.Errors.IngestCallout?
    let rejectAccessibilityIdentifier: String
    let isBusy: Bool
    let idleHint: LocalizedStringResource
    let onChoose: () -> Void
    let onClear: () -> Void
    let onDropURLs: ([URL]) -> Void

    init(
        file: Binding<IngestFileChoice>,
        isBusy: Bool,
        idleHint: LocalizedStringResource,
        rejectAccessibilityIdentifier: String
    ) {
        fileName = file.wrappedValue.fileName
        reject = file.wrappedValue.reject
        self.rejectAccessibilityIdentifier = rejectAccessibilityIdentifier
        self.isBusy = isBusy
        self.idleHint = idleHint
        onChoose = { file.wrappedValue.pick() }
        onClear = { file.wrappedValue.clear() }
        onDropURLs = { file.wrappedValue.choose($0) }
    }

    private var isRejected: Bool { reject != nil }

    @State private var isTargeted = false

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(spacing: PVSpacing.space4) {
                PVButton(
                    L10n.Sources.chooseFile,
                    variant: .secondary,
                    size: .sm,
                    icon: .fileUp
                ) {
                    onChoose()
                }
                .disabled(isBusy)
                .accessibilityIdentifier("sources.page.ingest.chooseFile")

                if isTargeted {
                    Text(L10n.Sources.ingestDropActive)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.accent)
                        .italic()
                } else if let fileName {
                    Text(fileName)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(isRejected ? PVColor.dangerForeground : PVColor.textPrimary)
                        .lineLimit(1)
                    PVButton(L10n.Sources.clearFile, variant: .ghost, size: .sm) {
                        onClear()
                    }
                    .disabled(isBusy)
                    .accessibilityIdentifier("sources.page.ingest.clearFile")
                } else {
                    Text(idleHint)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .italic()
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(PVSpacing.space5)
            .background(rowBackground)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .strokeBorder(rowBorder, lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
            .dropDestination(for: URL.self, action: { urls, _ in
                guard !isBusy else { return false }
                onDropURLs(urls)
                return true
            }, isTargeted: { isTargeted = $0 && !isBusy })

            Text(L10n.Sources.ingestAllowedTypesCaption)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)

            if let reject {
                PVCallout(tone: .danger, title: reject.title, message: reject.message, announces: true)
                    .accessibilityIdentifier(rejectAccessibilityIdentifier)
            }
        }
    }

    private var rowBackground: Color {
        if isTargeted { return PVColor.accentSoft }
        if isRejected { return PVColor.dangerSoft }
        return PVColor.surfaceSunken
    }

    private var rowBorder: Color {
        if isTargeted { return PVColor.accent }
        if isRejected { return PVColor.danger }
        return PVColor.borderSubtle
    }
}
