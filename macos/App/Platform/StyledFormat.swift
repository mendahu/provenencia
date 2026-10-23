import Foundation

/// A localized format whose arguments carry their own styling, so a sentence like "Mary from the
/// *Census* belongs in…" is one translatable format with the names and the Source title styled
/// where the translation puts them.
enum StyledFormat {
    /// One argument: its text and the attributes it is drawn with.
    struct Argument {
        var text: String
        var attributes: AttributeContainer = AttributeContainer()
    }

    /// The resource's format with `arguments` filled in. `%1$@` names an argument by position;
    /// a plain `%@` takes the next one; `%%` is a percent sign. Text outside an argument has no
    /// attributes.
    static func attributed(_ resource: LocalizedStringResource, _ arguments: [Argument]) -> AttributedString {
        attributed(format: L10n.string(resource), arguments)
    }

    /// ``attributed(_:_:)`` for a format already resolved.
    static func attributed(format: String, _ arguments: [Argument]) -> AttributedString {
        var result = AttributedString()
        var literal = ""
        var next = 0
        func flush() {
            if !literal.isEmpty { result += AttributedString(literal) }
            literal = ""
        }
        func append(_ index: Int) {
            guard arguments.indices.contains(index) else { return }
            flush()
            var piece = AttributedString(arguments[index].text)
            piece.mergeAttributes(arguments[index].attributes)
            result += piece
        }
        var characters = format[...]
        while let char = characters.first {
            guard char == "%" else {
                literal.append(char)
                characters = characters.dropFirst()
                continue
            }
            let rest = characters.dropFirst()
            if rest.first == "%" {
                literal.append("%")
                characters = rest.dropFirst()
            } else if rest.first == "@" {
                append(next)
                next += 1
                characters = rest.dropFirst()
            } else if let digits = Optional(rest.prefix(while: \.isNumber)), !digits.isEmpty,
                      rest.dropFirst(digits.count).hasPrefix("$@"), let position = Int(digits) {
                append(position - 1)
                characters = rest.dropFirst(digits.count + 2)
            } else {
                literal.append(char)
                characters = rest
            }
        }
        flush()
        return result
    }
}
