import Testing
@testable import Provenencia

@Suite
struct ProjectSlugTests {
    @Test func folderNameFromLabel() {
        #expect(ProjectSlug.folderName(from: "Robins Family") == "robins-family.provenencia")
        #expect(ProjectSlug.folderName(from: "A/B") == "a-b.provenencia")
        #expect(ProjectSlug.folderName(from: "  Robins   Family  ") == "robins-family.provenencia")
        #expect(ProjectSlug.folderName(from: "Robins Family.provenencia") == "robins-family.provenencia")
    }

    @Test func labelFromFolder() {
        #expect(ProjectSlug.labelFromFolder("/tmp/Robins Family.provenencia") == "Robins Family")
        #expect(ProjectSlug.labelFromFolder("/tmp/robins-family.provenencia") == "robins-family")
    }
}
