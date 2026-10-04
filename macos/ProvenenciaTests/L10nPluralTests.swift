import Foundation
import Testing
@testable import Provenencia

@Suite
struct L10nPluralTests {
    /// Every plural key in the compiled catalog resolves through `L10n.format`
    /// to a different form for one and for several, with no placeholder left over.
    @Test func everyCatalogPluralPicksAFormPerCount() throws {
        let url = try #require(Bundle.main.url(forResource: "Localizable", withExtension: "stringsdict"))
        let table = try #require(NSDictionary(contentsOf: url) as? [String: [String: Any]])
        #expect(table.count > 30)

        for (key, entry) in table.sorted(by: { $0.key < $1.key }) {
            let shape = try #require(PluralShape(entry: entry), "unreadable plural entry \(key)")
            let resource = LocalizedStringResource(String.LocalizationValue(key))
            let one = L10n.format(resource, in: .main, locale: L10n.formattingLocale, arguments: shape.arguments(count: 1))
            let several = L10n.format(resource, in: .main, locale: L10n.formattingLocale, arguments: shape.arguments(count: 2))

            #expect(one != several, "\(key) reads the same for 1 and 2: \(one)")
            #expect(one.contains("1"), "\(key) for 1: \(one)")
            #expect(several.contains("2"), "\(key) for 2: \(several)")
            for text in [one, several] {
                #expect(!text.contains("%") && !text.contains("#@"), "\(key) left a placeholder: \(text)")
            }
        }
    }

    /// A language with more plural forms than English, resolved through the same
    /// path as app copy. The user's region is US English; the app resolved
    /// Russian, so `formattingLocale` must format in Russian for the rules to apply.
    @Test func secondLanguagePluralsFollowTheResolvedLocalization() throws {
        let bundle = try Self.russianBundle()
        let resolved = try #require(bundle.preferredLocalizations.first)
        #expect(resolved == "ru")

        let locale = L10n.formattingLocale(current: Locale(identifier: "en_US"), resolvedLocalization: resolved)
        let resource = LocalizedStringResource("test.persons.count", defaultValue: "%lld persons")
        func persons(_ count: Int) -> String {
            L10n.format(resource, in: bundle, locale: locale, arguments: [count])
        }

        #expect(persons(1) == "1 человек (one)")
        #expect(persons(2) == "2 человека (few)")
        #expect(persons(5) == "5 человек (many)")
        #expect(persons(21) == "21 человек (one)")
    }

    // MARK: - Helpers

    /// Builds a bundle whose only localization is Russian, with one plural key.
    private static func russianBundle() throws -> Bundle {
        let root = FileManager.default.temporaryDirectory
            .appendingPathComponent("L10nPluralTests-\(UUID().uuidString).bundle")
        let lproj = root.appendingPathComponent("Contents/Resources/ru.lproj")
        try FileManager.default.createDirectory(at: lproj, withIntermediateDirectories: true)
        let info: NSDictionary = ["CFBundleIdentifier": "test.l10n.plural", "CFBundleDevelopmentRegion": "ru"]
        info.write(to: root.appendingPathComponent("Contents/Info.plist"), atomically: true)
        let plural: NSDictionary = [
            "test.persons.count": [
                "NSStringLocalizedFormatKey": "%#@count@",
                "count": [
                    "NSStringFormatSpecTypeKey": "NSStringPluralRuleType",
                    "NSStringFormatValueTypeKey": "lld",
                    "one": "%lld человек (one)",
                    "few": "%lld человека (few)",
                    "many": "%lld человек (many)",
                    "other": "%lld (other)",
                ],
            ],
        ]
        plural.write(to: lproj.appendingPathComponent("Localizable.stringsdict"), atomically: true)
        return try #require(Bundle(url: root))
    }
}

/// The arguments a compiled plural entry expects: which position is the count,
/// and whether each other position takes a number or a string.
private struct PluralShape {
    private var countPositions: Set<Int> = []
    private var stringPositions: Set<Int> = []
    private var numberPositions: Set<Int> = []

    init?(entry: [String: Any]) {
        guard let format = entry["NSStringLocalizedFormatKey"] as? String else { return nil }
        var texts = [format]
        for (name, value) in entry where name != "NSStringLocalizedFormatKey" {
            guard let rule = value as? [String: Any] else { return nil }
            texts += rule.compactMap { $0.key.hasPrefix("NS") ? nil : $0.value as? String }
            // The variable's own argument: its explicit position, or the first one.
            countPositions.insert(Self.position(ofVariable: name, in: format) ?? 1)
        }
        for text in texts {
            for (position, isString) in Self.specifiers(in: text) where !countPositions.contains(position) {
                if isString { stringPositions.insert(position) } else { numberPositions.insert(position) }
            }
        }
    }

    func arguments(count: Int) -> [any CVarArg] {
        let last = (countPositions.union(stringPositions).union(numberPositions)).max() ?? 1
        return (1...last).map { position -> any CVarArg in
            if countPositions.contains(position) { return count }
            if stringPositions.contains(position) { return "X" }
            return 7
        }
    }

    private static func position(ofVariable name: String, in format: String) -> Int? {
        guard let range = format.range(of: #"%(\d+)\$#@"# + name + "@", options: .regularExpression) else {
            return nil
        }
        return Int(format[range].dropFirst().prefix { $0.isNumber })
    }

    /// `%2$@` / `%lld` specifiers in `text`, numbered left to right when not positional.
    private static func specifiers(in text: String) -> [(Int, Bool)] {
        let pattern = #"%(?:(\d+)\$)?(#@|ll[du]|l[du]|[du@])"#
        let regex = try! NSRegularExpression(pattern: pattern)
        var next = 1
        return regex.matches(in: text, range: NSRange(text.startIndex..., in: text)).compactMap { match in
            let kind = String(text[Range(match.range(at: 2), in: text)!])
            guard kind != "#@" else { return nil }
            let position = Range(match.range(at: 1), in: text).flatMap { Int(text[$0]) } ?? next
            next = position + 1
            return (position, kind == "@")
        }
    }
}
