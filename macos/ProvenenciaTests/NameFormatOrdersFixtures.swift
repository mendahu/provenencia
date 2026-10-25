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
}
