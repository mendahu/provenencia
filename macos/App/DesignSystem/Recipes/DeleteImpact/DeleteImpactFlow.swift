import Foundation
import Observation

/// Shared GetDeleteImpact → confirm/notice state for the DeleteImpact recipe.
@MainActor
@Observable
final class DeleteImpactFlow {
    var request: PVDeleteImpactRequest?
    private(set) var isRunning = false
    private(set) var error: String?

    private var refetch: (() async throws -> CatalogDeleteImpact)?

    func applyRequest(_ value: PVDeleteImpactRequest?) {
        if value == nil {
            cancel()
        } else {
            request = value
        }
    }

    func ask(
        kind: String,
        id: String,
        ref: String,
        title: String,
        fetch: @escaping () async throws -> CatalogDeleteImpact
    ) async {
        error = nil
        refetch = fetch
        do {
            let report = try await fetch()
            request = PVDeleteImpactRequest(
                target: PVDeleteImpactTarget(kind: kind, id: id, ref: ref, title: title),
                report: report
            )
        } catch {
            self.error = L10n.Errors.message(for: error)
        }
    }

    func cancel() {
        guard !isRunning else { return }
        request = nil
        error = nil
        refetch = nil
    }

    @discardableResult
    func confirm(perform: (PVDeleteImpactTarget) async throws -> Void) async -> Bool {
        guard let current = request, current.report.allowed, !isRunning else {
            return false
        }
        isRunning = true
        error = nil
        defer { isRunning = false }
        do {
            try await perform(current.target)
            request = nil
            refetch = nil
            return true
        } catch {
            self.error = L10n.Errors.message(for: error)
            if Self.isInUse(error) {
                await refreshAfterInUse(target: current.target)
            }
            return false
        }
    }

    private func refreshAfterInUse(target: PVDeleteImpactTarget) async {
        guard let refetch else { return }
        do {
            let report = try await refetch()
            request = PVDeleteImpactRequest(target: target, report: report)
            if !report.allowed {
                error = nil
            }
        } catch {
            self.error = L10n.Errors.message(for: error)
        }
    }

    static func isInUse(_ error: Error) -> Bool {
        if case .coded(_, let code, _, _) = error as? CoreInvokeError {
            return code.hasSuffix(".in_use")
        }
        return false
    }
}
