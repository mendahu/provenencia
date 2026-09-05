---
name: add-localized-string
description: >-
  Adds or updates Provenencia macOS UI copy via typed L10n and String Catalogs.
  Use when adding or changing user-facing SwiftUI/AppKit strings, L10n.swift,
  Localizable.xcstrings, InfoPlist.xcstrings, localization, i18n, l10n, or when
  replacing hard-coded English in macos/App. Also use when mapping a new FFI
  error code under L10n.Errors.
---

# Add or update a localized string

Mac UI copy is **never** hard-coded at call sites. Typed keys live in [`macos/App/Platform/L10n.swift`](macos/App/Platform/L10n.swift). English + future locales live in [`macos/App/Resources/Localizable.xcstrings`](macos/App/Resources/Localizable.xcstrings) (UI) or [`InfoPlist.xcstrings`](macos/App/Resources/InfoPlist.xcstrings) (Info.plist keys).

Authoritative notes: [`docs/macos-client-patterns.md`](docs/macos-client-patterns.md) § Localization.

## Checklist (keep in sync)

Copy and track:

```
- [ ] L10n member added/updated under the right feature nest
- [ ] Matching key + en value (+ comment) in the String Catalog
- [ ] Call site uses L10n.* only (no raw English / raw key literals)
- [ ] Unused L10n members and catalog keys pruned together
- [ ] Tests that asserted old English compare via String(localized: L10n.…)
```

Do **not** edit only `L10n.swift` or only the catalog.

## Steps

1. Choose a **semantic dotted key**: `<feature>.<screenOrArea>.<name>`  
   Examples: `onboarding.chooseFile.welcomeTitle`, `onboarding.common.continue`.  
   Not English-as-key. Brand `Provenencia` stays untranslated in values.
2. Nest under `enum L10n { enum <Feature> { … } }` (e.g. `L10n.Onboarding`). Add a new nest for a new feature.
3. Prefer `LocalizedStringResource` for SwiftUI (`Text` / `Button` / `TextField` / `Picker`):

```swift
static let welcomeTitle = LocalizedStringResource(
    "onboarding.chooseFile.welcomeTitle",
    defaultValue: "Welcome to Provenencia",
    comment: "Onboarding choose-file headline"
)
```

4. For **one argument** (`%@`), keep a static catalog key and format in the helper (do not interpolate into the `LocalizedStringResource` key — that requires `StaticString`):

```swift
static func bodySignedIn(displayName: String) -> String {
    let format = String(localized: LocalizedStringResource(
        "onboarding.chooseFile.bodySignedIn",
        defaultValue: "You're signed in as %@. Open a project you already have, or create a new one.",
        comment: "…; argument is display name"
    ))
    return String(format: format, locale: .current, displayName)
}
```

Catalog `value` must use the same `%@` placeholders.

5. Add/update the key in `Localizable.xcstrings` (`extractionState: "manual"`, `en` `state: "translated"`, same comment/value as `defaultValue`). For Info.plist TCC/copy keys, use `InfoPlist.xcstrings` instead (e.g. `NSDocumentsFolderUsageDescription`) — do not put those only in `INFOPLIST_KEY_*` build settings.
6. Call sites:

```swift
Text(L10n.Onboarding.welcomeTitle)
Button(String(localized: L10n.Onboarding.continueAction)) { … }
TextField(String(localized: L10n.Onboarding.researcherName), text: $name, prompt: Text(L10n.Onboarding.researcherNamePrompt))
panel.prompt = String(localized: L10n.Onboarding.openPanelPrompt)
errorText = String(localized: L10n.Onboarding.missingProject)
```

`Text(…)` accepts `LocalizedStringResource`. `Button` / `TextField` / `Picker` title inits on our macOS 14 deployment target need `StringProtocol` — wrap with `String(localized:)`.
7. Accessibility: keep `.accessibilityIdentifier("dotted.name")`. Do not UI-test by localized title.
8. Go/FFI failures arrive as protobuf `Error` codes. Add `L10n.Errors` + catalog keys (`error.<domain>.<name>`) when introducing a new user-visible code; map via `L10n.Errors.message`.

## Do not

- Put user-facing string literals in views/models (except non-copy data: names, paths, IDs)
- Add SwiftGen or a localization Run Script
- Invent a second string table outside `L10n` + the catalogs
- Bump `VERSION` for localization-wiring-only changes
