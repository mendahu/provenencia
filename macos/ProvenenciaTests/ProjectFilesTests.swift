import Foundation
import Testing
@testable import Provenencia

@Suite
struct ProjectFilesTests {
    @Test func objectURLJoinsNormalContentAddressedPath() throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/abcd")

        let url = ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/abcd")
        #expect(url?.path == project + "/objects/ab/cd/abcd")
    }

    @Test func objectURLJoinsPathWithExtension() throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/abcd.jpg")

        let url = ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/abcd.jpg")
        #expect(url?.path.hasSuffix("objects/ab/cd/abcd.jpg") == true)
    }

    @Test func objectURLRejectsEmptyRelPath() {
        #expect(ProjectFiles.objectURL(projectDir: "/tmp/proj.provenencia", relPath: "") == nil)
        #expect(ProjectFiles.objectURL(projectDir: "/tmp/proj.provenencia", relPath: "   ") == nil)
    }

    @Test func objectURLRejectsParentDirectorySegments() throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/x")

        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/../ab/cd/x") == nil)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x") != nil)
    }

    @Test func objectURLRejectsDotSegments() throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/x")

        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/./ab/cd/x") == nil)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x") != nil)
    }

    @Test func objectURLRejectsHomeRelativeSegment() throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/x")
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/~/x")

        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/~/x") == nil)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x") != nil)
    }

    @Test func objectURLRefusesSymlinkedObjectsFolder() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        let outside = try ArtifactFileFixtures.linkToOutsideFolder(
            at: project + "/objects",
            holding: "ab/cd/x.png"
        )
        defer {
            try? fm.removeItem(atPath: project)
            try? fm.removeItem(atPath: outside)
        }
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x.png") == nil)

        try fm.removeItem(atPath: project + "/objects")
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/x.png")
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x.png") != nil)
    }

    @Test func objectURLRefusesSymlinkedShardFolder() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        let outside = try ArtifactFileFixtures.linkToOutsideFolder(
            at: project + "/objects/ab",
            holding: "cd/x.png"
        )
        defer {
            try? fm.removeItem(atPath: project)
            try? fm.removeItem(atPath: outside)
        }
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x.png") == nil)

        try fm.removeItem(atPath: project + "/objects/ab")
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/x.png")
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x.png") != nil)
    }

    @Test func objectURLRefusesSymlinkedObject() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        let rel = "objects/ab/cd/x.png"
        let outside = try ArtifactFileFixtures.linkToOutsidePNG(at: project + "/" + rel)
        defer {
            try? fm.removeItem(atPath: project)
            try? fm.removeItem(atPath: outside)
        }
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: rel) == nil)

        try fm.removeItem(atPath: project + "/" + rel)
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: rel)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: rel) != nil)
    }

    @Test func objectURLRefusesAliasObject() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        let rel = "objects/ab/cd/" + String(repeating: "a", count: 64) + ".png"
        let outside = try ArtifactFileFixtures.aliasToOutsidePNG(at: project + "/" + rel)
        defer {
            try? fm.removeItem(atPath: project)
            try? fm.removeItem(atPath: outside)
        }
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: rel) == nil)

        try fm.removeItem(atPath: project + "/" + rel)
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: rel)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: rel) != nil)
    }

    @Test(.timeLimit(.minutes(1)))
    func objectURLRefusesNonRegularObject() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? fm.removeItem(atPath: project) }
        let rel = "objects/ab/cd/x.png"

        try fm.createDirectory(atPath: project + "/" + rel, withIntermediateDirectories: true)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: rel) == nil)

        try fm.removeItem(atPath: project + "/" + rel)
        #expect(mkfifo(project + "/" + rel, 0o644) == 0)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: rel) == nil)

        try fm.removeItem(atPath: project + "/" + rel)
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: rel)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: rel) != nil)
    }

    @Test func objectURLRefusesMissingObject() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? fm.removeItem(atPath: project) }
        try fm.createDirectory(atPath: project + "/objects/ab/cd", withIntermediateDirectories: true)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x.png") == nil)

        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/x.png")
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x.png") != nil)
    }

    @Test func objectURLRefusesPathOutsideObjects() throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "provenencia.sqlite")
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "derivatives/x.jpg")
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/x.png")

        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "provenencia.sqlite") == nil)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "derivatives/x.jpg") == nil)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects") == nil)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: "objects/ab/cd/x.png") != nil)
    }

    @Test func objectURLAllowsSymlinkedProjectFolder() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        let link = ArtifactFileFixtures.uniqueProjectDir()
        defer {
            try? fm.removeItem(atPath: project)
            try? fm.removeItem(atPath: link)
        }
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/ab/cd/x.png")
        try fm.createSymbolicLink(atPath: link, withDestinationPath: project)

        let url = ProjectFiles.objectURL(projectDir: link, relPath: "objects/ab/cd/x.png")
        #expect(url?.path == link + "/objects/ab/cd/x.png")
    }

    /// A file system that cannot hold a permission bit (exFAT, FAT) reports every
    /// file as executable, so the bit alone never refuses an object.
    @Test func objectURLAcceptsExecuteBit() throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        let rel = "objects/ab/cd/x.png"
        try ArtifactFileFixtures.write(Data("x".utf8), projectDir: project, relPath: rel, mode: 0o755)
        #expect(ProjectFiles.objectURL(projectDir: project, relPath: rel) != nil)
    }

    @Test(arguments: [
        ("objects/aa/bb/aabb", mode_t(0o644)),
        ("objects/aa/bb/aabb.sh", mode_t(0o644)),
        ("objects/aa/bb/aabb.command", mode_t(0o644)),
        ("objects/aa/bb/aabb.app", mode_t(0o644)),
    ])
    func openableObjectURLRefusesUnsafeObjects(relPath: String, mode: mode_t) throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        try ArtifactFileFixtures.write(Data("x".utf8), projectDir: project, relPath: relPath, mode: mode)

        #expect(ProjectFiles.openableObjectURL(projectDir: project, relPath: relPath) == .failure(.unsafe))
    }

    @Test func openableObjectURLAcceptsAStoredPDF() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        let rel = "objects/aa/bb/aabb.pdf"
        defer { try? fm.removeItem(atPath: project) }
        #expect(ProjectFiles.openableObjectURL(projectDir: project, relPath: rel) == .failure(.missing))

        try ArtifactFileFixtures.writeBlankPDF(projectDir: project, relPath: rel, pageCount: 1)
        let opened = try? ProjectFiles.openableObjectURL(projectDir: project, relPath: rel).get()
        #expect(opened?.path == project + "/" + rel)

        // The execute bit alone is not a reason to refuse (exFAT and FAT report it on every file).
        #expect(chmod(project + "/" + rel, 0o755) == 0)
        let withBit = try? ProjectFiles.openableObjectURL(projectDir: project, relPath: rel).get()
        #expect(withBit?.path == project + "/" + rel)

        try fm.removeItem(atPath: project + "/" + rel)
        let outside = try ArtifactFileFixtures.linkToOutsidePNG(at: project + "/" + rel)
        defer { try? fm.removeItem(atPath: outside) }
        #expect(ProjectFiles.openableObjectURL(projectDir: project, relPath: rel) == .failure(.missing))
    }

    @Test func openObjectReportsMissingForSymlinkedObject() throws {
        let fm = FileManager.default
        let project = ArtifactFileFixtures.uniqueProjectDir()
        let rel = "objects/ab/cd/x.png"
        let outside = try ArtifactFileFixtures.linkToOutsidePNG(at: project + "/" + rel)
        defer {
            try? fm.removeItem(atPath: project)
            try? fm.removeItem(atPath: outside)
        }
        #expect(ProjectFiles.openObject(projectDir: project, relPath: rel) == .missing)
    }

    @Test func openObjectReportsMissingForRejectedRelPath() {
        #expect(
            ProjectFiles.openObject(
                projectDir: "/tmp/proj.provenencia",
                relPath: "../etc/passwd"
            ) == .missing
        )
    }
}
