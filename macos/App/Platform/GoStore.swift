import Foundation

struct GoStore: GenealogyStore {
    func ping(_ message: String) async throws -> String {
        try await Task.detached {
            var req = Provenance_Engine_V1_PingRequest()
            req.message = message
            let out = try provenanceInvoke(method: CoreMethod.ping, request: try req.serializedData())
            let resp = try Provenance_Engine_V1_PingResponse(serializedBytes: out)
            return resp.message
        }.value
    }

    func version() async throws -> String {
        try await Task.detached {
            let req = Provenance_Engine_V1_GetVersionRequest()
            let out = try provenanceInvoke(method: CoreMethod.getVersion, request: try req.serializedData())
            let resp = try Provenance_Engine_V1_GetVersionResponse(serializedBytes: out)
            return resp.version
        }.value
    }

    func sqliteProbe() async throws -> (ok: Bool, detail: String) {
        try await Task.detached {
            let req = Provenance_Engine_V1_SqliteProbeRequest()
            let out = try provenanceInvoke(method: CoreMethod.sqliteProbe, request: try req.serializedData())
            let resp = try Provenance_Engine_V1_SqliteProbeResponse(serializedBytes: out)
            return (resp.ok, resp.detail)
        }.value
    }
}
