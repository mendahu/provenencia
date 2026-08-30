import Foundation
import SwiftProtobuf

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
    static let ping = Int32(Provenencia_Engine_V1_Method.ping.rawValue)
    static let getVersion = Int32(Provenencia_Engine_V1_Method.getVersion.rawValue)
    static let getInstallIdentity = Int32(Provenencia_Engine_V1_Method.getInstallIdentity.rawValue)
    static let completeOnboarding = Int32(Provenencia_Engine_V1_Method.completeOnboarding.rawValue)
    static let removeInstallIdentity = Int32(Provenencia_Engine_V1_Method.removeInstallIdentity.rawValue)
    static let getActiveProject = Int32(Provenencia_Engine_V1_Method.getActiveProject.rawValue)
    static let openProject = Int32(Provenencia_Engine_V1_Method.openProject.rawValue)
    static let removeActiveProject = Int32(Provenencia_Engine_V1_Method.removeActiveProject.rawValue)
    static let listProjectUsers = Int32(Provenencia_Engine_V1_Method.listProjectUsers.rawValue)
    static let signOut = Int32(Provenencia_Engine_V1_Method.signOut.rawValue)
}

func provenenciaInvoke(method: Int32, request: Data) throws -> Data {
    var outPtr: UnsafeMutablePointer<UInt8>?
    var outLen: Int = 0
    let status: Int32 = request.withUnsafeBytes { raw in
        let inPtr = raw.bindMemory(to: UInt8.self).baseAddress
        return provenencia_call(method, inPtr, request.count, &outPtr, &outLen)
    }
    defer {
        if let outPtr {
            provenencia_free(UnsafeMutableRawPointer(outPtr))
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

func provenenciaCall<Request: Message & Sendable, Response: Message & Sendable>(
    method: Int32,
    request: Request
) async throws -> Response {
    try await Task.detached {
        let out = try provenenciaInvoke(method: method, request: try request.serializedData())
        return try Response(serializedBytes: out)
    }.value
}
