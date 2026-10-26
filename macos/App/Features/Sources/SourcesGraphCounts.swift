import Foundation

/// What an Evidence graph's counts show: not arrived yet, no Subjects yet, or
/// Subject and Observation counts. The token, its help, the Source page's count
/// line and VoiceOver all read this one value; `L10n.Sources.graphCounts` formats it.
enum SourcesGraphCounts: Equatable {
    case loading
    case notStarted
    case counts(subjects: Int, observations: Int)

    /// A graph missing from the progress map has zero counts (Go omits it).
    init(progress: EvidenceGraphProgress?, countsLoading: Bool) {
        if countsLoading && progress == nil {
            self = .loading
        } else if let progress, progress.subjectCount > 0 {
            self = .counts(subjects: progress.subjectCount, observations: progress.observationCount)
        } else {
            self = .notStarted
        }
    }
}
