import Foundation

enum CoreInvokeError: Error {
    case failed(Int32, message: String)
}

extension CoreInvokeError: LocalizedError {
    var errorDescription: String? {
        switch self {
        case .failed(_, let message):
            return message
        }
    }
}

enum CoreMethod {
    static let ping = Int32(Provenance_Engine_V1_Method.ping.rawValue)
    static let getVersion = Int32(Provenance_Engine_V1_Method.getVersion.rawValue)
    static let getInstallIdentity = Int32(Provenance_Engine_V1_Method.getInstallIdentity.rawValue)
    static let completeOnboarding = Int32(Provenance_Engine_V1_Method.completeOnboarding.rawValue)
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
        var message = "core call failed (\(status))"
        if let outPtr, outLen > 0 {
            message = String(decoding: Data(bytes: outPtr, count: outLen), as: UTF8.self)
        }
        throw CoreInvokeError.failed(status, message: message)
    }
    guard let outPtr, outLen > 0 else {
        return Data()
    }
    return Data(bytes: outPtr, count: outLen)
}
