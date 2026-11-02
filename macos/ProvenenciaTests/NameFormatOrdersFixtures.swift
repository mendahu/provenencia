import Foundation
@testable import Provenencia

extension NameFormatOrders {
    /// Given then surname, naturally; surname, then given, sorted.
    static let givenSurname = NameFormatOrders(
        natural: [
            NameFormatPart(partType: "given", trailingSeparator: ""),
            NameFormatPart(partType: "surname", trailingSeparator: ""),
        ],
        sorted: [
            NameFormatPart(partType: "surname", trailingSeparator: ", "),
            NameFormatPart(partType: "given", trailingSeparator: ""),
        ]
    )

    /// The seeded Western orders: prefix, given, nickname, former surname, surname prefix,
    /// surname, suffix; sorted with the surname first, then a comma.
    static let western = NameFormatOrders(
        natural: [
            NameFormatPart(partType: "prefix", trailingSeparator: ""),
            NameFormatPart(partType: "given", trailingSeparator: ""),
            NameFormatPart(partType: "nick", trailingSeparator: ""),
            NameFormatPart(partType: "former_surname", trailingSeparator: ""),
            NameFormatPart(partType: "surname_prefix", trailingSeparator: ""),
            NameFormatPart(partType: "surname", trailingSeparator: ""),
            NameFormatPart(partType: "suffix", trailingSeparator: ""),
        ],
        sorted: [
            NameFormatPart(partType: "surname_prefix", trailingSeparator: ""),
            NameFormatPart(partType: "surname", trailingSeparator: ", "),
            NameFormatPart(partType: "prefix", trailingSeparator: ""),
            NameFormatPart(partType: "given", trailingSeparator: ""),
            NameFormatPart(partType: "nick", trailingSeparator: ""),
            NameFormatPart(partType: "former_surname", trailingSeparator: ""),
            NameFormatPart(partType: "suffix", trailingSeparator: ""),
        ]
    )
}
