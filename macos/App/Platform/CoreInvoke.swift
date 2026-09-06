import Foundation
import SwiftProtobuf

enum CoreErrorKind: Equatable, Sendable {
    case unspecified
    case user
    case conflict
    case notFound
    case `internal`

    init(_ proto: Provenencia_Engine_V1_ErrorKind) {
        switch proto {
        case .user: self = .user
        case .conflict: self = .conflict
        case .notFound: self = .notFound
        case .internal: self = .internal
        case .unspecified, .UNRECOGNIZED:
            self = .unspecified
        }
    }
}

enum CoreInvokeError: Error, Equatable {
    /// FFI status 1 with a decoded protobuf Error.
    case coded(status: Int32, code: String, kind: CoreErrorKind, params: [String])
    /// Non-zero status that could not be decoded as Error (e.g. malloc failure).
    case failed(status: Int32)
}

extension CoreInvokeError: LocalizedError {
    var errorDescription: String? {
        switch self {
        case .coded(_, let code, _, let params):
            return L10n.Errors.message(code: code, params: params)
        case .failed:
            return String(localized: L10n.Errors.unknown)
        }
    }
}

enum CoreMethod {
    static let getInstallIdentity = Int32(Provenencia_Engine_V1_Method.getInstallIdentity.rawValue)
    static let completeOnboarding = Int32(Provenencia_Engine_V1_Method.completeOnboarding.rawValue)
    static let getActiveProject = Int32(Provenencia_Engine_V1_Method.getActiveProject.rawValue)
    static let openProject = Int32(Provenencia_Engine_V1_Method.openProject.rawValue)
    static let removeActiveProject = Int32(Provenencia_Engine_V1_Method.removeActiveProject.rawValue)
    static let listProjectUsers = Int32(Provenencia_Engine_V1_Method.listProjectUsers.rawValue)
    static let signOut = Int32(Provenencia_Engine_V1_Method.signOut.rawValue)
    static let getProjectInfo = Int32(Provenencia_Engine_V1_Method.getProjectInfo.rawValue)
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
        if status == 1, let outPtr, outLen > 0 {
            let data = Data(bytes: outPtr, count: outLen)
            if let decoded = try? Provenencia_Engine_V1_Error(serializedBytes: data), !decoded.code.isEmpty {
                throw CoreInvokeError.coded(
                    status: status,
                    code: decoded.code,
                    kind: CoreErrorKind(decoded.kind),
                    params: decoded.params
                )
            }
        }
        throw CoreInvokeError.failed(status: status)
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
