import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct WorkspaceSessionTests {
    private let projectDir = "/tmp/workspace-session.provenencia"

    private func makeSession(store: FakeStore = FakeStore()) -> WorkspaceSession {
        WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
    }

    /// Waits until the handle is not actively refetching and has settled on `.ready` or `.error`.
    private func waitForFetchComplete<Value>(
        _ handle: QueryHandle<Value>,
        timeoutNanoseconds: UInt64 = 2_000_000_000
    ) async {
        var waited: UInt64 = 0
        let step: UInt64 = 10_000_000
        while waited < timeoutNanoseconds {
            if handle.isFetching {
                // stale-while-revalidate still in flight
            } else if handle.status == .ready || handle.status == .error {
                return
            }
            await Task.yield()
            try? await Task.sleep(nanoseconds: step)
            waited += step
        }
    }

    @Test func loadsAndCachesValue() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()

        let handle = session.ensureQuery(key) {
            counter.count += 1
            return ["alpha"]
        }
        await waitForFetchComplete(handle)
        #expect(handle.status == .ready)
        #expect(handle.value == ["alpha"])
        #expect(counter.count == 1)

        _ = session.ensureQuery(key) {
            counter.count += 1
            return ["beta"]
        }
        await Task.yield()
        #expect(counter.count == 1)
        #expect(handle.value == ["alpha"])
    }

    @Test func dedupesConcurrentLoads() async {
        let session = makeSession()
        let key = CatalogQueryKey.metadataFieldsList(project: session.projectKey)
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()
        let gate = Gate()

        await withTaskGroup(of: Void.self) { group in
            group.addTask { @MainActor in
                _ = session.ensureQuery(key) {
                    counter.count += 1
                    await gate.waitForOpen()
                    return ["shared"]
                }
            }
            group.addTask { @MainActor in
                _ = session.ensureQuery(key) {
                    counter.count += 1
                    await gate.waitForOpen()
                    return ["shared"]
                }
            }
            await gate.open()
            await group.waitForAll()
        }

        let handle: QueryHandle<[String]> = session.queryHandle(key)!
        await waitForFetchComplete(handle)
        #expect(counter.count == 1)
        #expect(handle.status == .ready)
        #expect(handle.value == ["shared"])
    }

    @Test func invalidateTriggersReload() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourceTypesList(project: session.projectKey)
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()

        let handle = session.ensureQuery(key) {
            counter.count += 1
            return ["first"]
        }
        await waitForFetchComplete(handle)
        #expect(counter.count == 1)

        session.invalidate(key)
        _ = session.ensureQuery(key) {
            counter.count += 1
            return ["second"]
        }
        await waitForFetchComplete(handle)
        #expect(counter.count == 2)
        #expect(handle.value == ["second"])
    }

    @Test func invalidateAllMatching() async {
        let session = makeSession()
        let otherProject = ProjectKey(projectDir: "/tmp/other.provenencia")
        let keyA = CatalogQueryKey.sourcesList(project: session.projectKey)
        let keyB = CatalogQueryKey.sourceTypesList(project: session.projectKey)
        let keyOther = CatalogQueryKey.sourcesList(project: otherProject)
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()

        let handleA = session.ensureQuery(keyA) {
            counter.count += 1
            return ["a"]
        }
        let handleB = session.ensureQuery(keyB) {
            counter.count += 1
            return ["b"]
        }
        let handleOther = session.ensureQuery(keyOther) {
            counter.count += 1
            return ["other"]
        }
        await waitForFetchComplete(handleA)
        await waitForFetchComplete(handleB)
        await waitForFetchComplete(handleOther)
        #expect(counter.count == 3)

        session.invalidateAll { $0.project == session.projectKey }

        _ = session.ensureQuery(keyA) {
            counter.count += 1
            return ["a2"]
        }
        _ = session.ensureQuery(keyB) {
            counter.count += 1
            return ["b2"]
        }
        _ = session.ensureQuery(keyOther) {
            counter.count += 1
            return ["other2"]
        }
        await waitForFetchComplete(handleA)
        await waitForFetchComplete(handleB)
        await waitForFetchComplete(handleOther)
        #expect(counter.count == 5)
        #expect(handleOther.value == ["other"])
    }

    @Test func staleWhileRevalidate() async {
        let session = makeSession()
        let key = CatalogQueryKey.typeSuggestions(project: session.projectKey, typeId: "t1")
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()
        let gate = Gate()

        let handle = session.ensureQuery(key) {
            counter.count += 1
            return ["first"]
        }
        await waitForFetchComplete(handle)

        session.invalidate(key)
        _ = session.ensureQuery(key) {
            counter.count += 1
            await gate.waitForOpen()
            return ["second"]
        }

        #expect(handle.value == ["first"])
        #expect(handle.isFetching)
        await gate.open()
        await waitForFetchComplete(handle)
        #expect(handle.value == ["second"])
        #expect(handle.isFetching == false)
        #expect(counter.count == 2)
    }

    @Test func propagatesLoaderError() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: "missing")
        struct TestFailure: Error {}

        let handle = session.ensureQuery(key) {
            throw TestFailure()
        }
        await waitForFetchComplete(handle)
        #expect(handle.status == .error)
        #expect(handle.error is TestFailure)
        #expect(handle.value == nil)
    }

    @Test func loadsSourcesListViaFakeStore() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "1", ref: "SRC-1", sourceTypeID: "", title: "A", description: ""),
        ]
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)

        let handle = session.ensureQuery(key) {
            try await store.listSources(projectDir: projectDir)
        }
        await waitForFetchComplete(handle)
        #expect(handle.status == .ready)
        #expect(handle.value?.count == 1)
        #expect(handle.value?.first?.ref == "SRC-1")
        #expect(store.heldCatalogProjectDir == projectDir)
    }
}

/// One-shot gate so concurrent loaders block until tests observe stale data.
@MainActor
private final class Gate {
    private var continuations: [CheckedContinuation<Void, Never>] = []
    private var isOpen = false

    func waitForOpen() async {
        if isOpen { return }
        await withCheckedContinuation { continuation in
            if isOpen {
                continuation.resume()
            } else {
                continuations.append(continuation)
            }
        }
    }

    func open() {
        isOpen = true
        for continuation in continuations {
            continuation.resume()
        }
        continuations.removeAll()
    }
}
