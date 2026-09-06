import Testing
@testable import Provenencia

@Suite
struct ProjectSlugTests {
    @Test func folderNameFixturesMatchGo() {
        for fixture in ProjectSlug.folderNameFixtures {
            #expect(
                ProjectSlug.folderName(from: fixture.label) == fixture.folder,
                "\(fixture.label) → \(fixture.folder)"
            )
        }
    }

    @Test func folderNameEmptyAlignsWithGoInvalid() {
        #expect(ProjectSlug.folderName(from: "  ") == "")
        #expect(ProjectSlug.folderName(from: "...") == "")
        #expect(ProjectSlug.folderName(from: "") == "")
    }

    @Test func labelFromFolder() {
        #expect(ProjectSlug.labelFromFolder("/tmp/Robins Family.provenencia") == "Robins Family")
        #expect(ProjectSlug.labelFromFolder("/tmp/robins-family.provenencia") == "robins-family")
    }
}
