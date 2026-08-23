import Foundation

enum CoreInvokeError: Error {
    case failed(Int32)
}

enum CoreMethod {
    static let ping = Int32(Provenance_Engine_V1_Method.ping.rawValue)
    static let getVersion = Int32(Provenance_Engine_V1_Method.getVersion.rawValue)
    // TEMPORARY (Spike 1 PR3): dylib SQLite probe. Remove when project open/create exists.
    static let sqliteProbe = Int32(Provenance_Engine_V1_Method.sqliteProbe.rawValue)
}

func provenanceInvoke(method: Int32, request: Data) throws -> Data {
    var outPtr: UnsafeMutablePointer<UInt8>?
    var outLen: Int = 0
    let status: Int32 = request.withUnsafeBytes { raw in
        let inPtr = raw.bindMemory(to: UInt8.self).baseAddress.map {
            UnsafeMutablePointer(mutating: $0)
        }
        return provenance_call(method, inPtr, request.count, &outPtr, &outLen)
    }
    defer {
        if let outPtr {
            provenance_free(UnsafeMutableRawPointer(outPtr))
        }
    }
    if status != 0 {
        throw CoreInvokeError.failed(status)
    }
    guard let outPtr, outLen > 0 else {
        return Data()
    }
    return Data(bytes: outPtr, count: outLen)
}
