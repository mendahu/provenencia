import Foundation

/// What an Evidence graph's counts show: not arrived yet, failed to arrive, no
/// Subjects yet, or Subject and Observation counts. The graph button, its help, the
/// Source page's count line and VoiceOver all read this one value;
/// `L10n.Sources.graphCounts` formats it.
enum SourcesGraphCounts: Equatable {
    case loading
    /// The progress map failed to load and left no value; `message` says why.
    case unavailable(message: String)
    case notStarted
    case counts(subjects: Int, observations: Int, filed: Int)

    /// The graph's status: Not started with no Subjects, Promoted once any is
    /// filed, else In progress. Nil while the counts load or when they failed.
    var status: SourcesModel.EvidenceStatus? {
        switch self {
        case .loading, .unavailable: nil
        case .notStarted: .empty
        case .counts(_, _, let filed): filed > 0 ? .promoted : .inProgress
        }
    }

    /// A graph missing from the progress map has zero counts (Go omits it). A
    /// stale progress value wins over a failure, as `QueryHandle.read` does.
    init(progress: EvidenceGraphProgress?, countsLoading: Bool, countsError: Error?) {
        if countsLoading && progress == nil {
            self = .loading
        } else if let progress, progress.subjectCount > 0 {
            self = .counts(
                subjects: progress.subjectCount,
                observations: progress.observationCount,
                filed: progress.filedCount
            )
        } else if progress == nil, let countsError {
            self = .unavailable(message: L10n.Errors.message(for: countsError))
        } else {
            self = .notStarted
        }
    }
}
