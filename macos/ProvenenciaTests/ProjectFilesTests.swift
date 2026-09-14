import Foundation
import Testing
@testable import Provenencia

@Suite
struct ProjectFilesTests {
    @Test func objectURLJoinsNormalContentAddressedPath() {
        let url = ProjectFiles.objectURL(
            projectDir: "/tmp/proj.provenencia",
            relPath: "objects/ab/cd/abcd"
        )
        #expect(url?.path == "/tmp/proj.provenencia/objects/ab/cd/abcd")
    }

    @Test func objectURLJoinsPathWithExtension() {
        let url = ProjectFiles.objectURL(
            projectDir: "/tmp/proj.provenencia",
            relPath: "objects/ab/cd/abcd.jpg"
        )
        #expect(url?.path == "/tmp/proj.provenencia/objects/ab/cd/abcd.jpg")
    }

    @Test func objectURLRejectsEmptyRelPath() {
        #expect(ProjectFiles.objectURL(projectDir: "/tmp/proj.provenencia", relPath: "") == nil)
        #expect(ProjectFiles.objectURL(projectDir: "/tmp/proj.provenencia", relPath: "   ") == nil)
    }

    @Test func objectURLRejectsParentDirectorySegments() {
        #expect(
            ProjectFiles.objectURL(
                projectDir: "/tmp/proj.provenencia",
                relPath: "../etc/passwd"
            ) == nil
        )
        #expect(
            ProjectFiles.objectURL(
                projectDir: "/tmp/proj.provenencia",
                relPath: "objects/../../etc/passwd"
            ) == nil
        )
        #expect(
            ProjectFiles.objectURL(
                projectDir: "/tmp/proj.provenencia",
                relPath: "objects/ab/../../../etc/passwd"
            ) == nil
        )
    }

    @Test func objectURLRejectsDotSegments() {
        #expect(
            ProjectFiles.objectURL(
                projectDir: "/tmp/proj.provenencia",
                relPath: "objects/./ab/cd"
            ) == nil
        )
    }

    @Test func objectURLRejectsHomeRelativeSegment() {
        #expect(
            ProjectFiles.objectURL(
                projectDir: "/tmp/proj.provenencia",
                relPath: "~/secret"
            ) == nil
        )
    }

    @Test func openObjectReturnsFalseForRejectedRelPath() {
        #expect(
            ProjectFiles.openObject(
                projectDir: "/tmp/proj.provenencia",
                relPath: "../etc/passwd"
            ) == false
        )
    }
}
