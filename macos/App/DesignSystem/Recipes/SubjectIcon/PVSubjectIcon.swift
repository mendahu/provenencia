import SwiftUI

/// Person / Event / Place marks for Evidence graph cards (S6-D1 board artwork).
///
/// Drawn as stroked paths — not SF Symbols — so kind stays legible at canvas zoom
/// and matches the Claude Design board. Template-style: callers set
/// `.foregroundStyle`.
enum PVSubjectIconKind: String, Sendable, CaseIterable {
    case person
    case event
    case place
    case relationship
    case participation
    case location
    case source
}

struct PVSubjectIcon: View {
    var kind: PVSubjectIconKind
    var size: CGFloat = 17

    var body: some View {
        Group {
            switch kind {
            case .person:
                PersonMark()
            case .event:
                EventMark()
            case .place:
                PlaceMark()
            case .relationship:
                RelationshipMark()
            case .participation:
                ParticipationMark()
            case .location:
                LocationMark()
            case .source:
                SourceMark()
            }
        }
        .frame(width: size, height: size)
        .accessibilityHidden(true)
    }
}

// MARK: - Paths (24×24 viewBox from S6-D1 board)

private struct PersonMark: View {
    var body: some View {
        Canvas { context, size in
            let scale = min(size.width, size.height) / 24
            let transform = CGAffineTransform(scaleX: scale, y: scale)
            let style = StrokeStyle(lineWidth: 1.4 * scale, lineCap: .round, lineJoin: .round)

            let head = Path(ellipseIn: CGRect(x: 12 - 3.3, y: 8.2 - 3.3, width: 6.6, height: 6.6))
            context.stroke(head.applying(transform), with: .foreground, style: style)

            // M5.2 19.4c0-3.5 3-5.9 6.8-5.9s6.8 2.4 6.8 5.9
            var shoulders = Path()
            shoulders.move(to: CGPoint(x: 5.2, y: 19.4))
            shoulders.addCurve(
                to: CGPoint(x: 12, y: 13.5),
                control1: CGPoint(x: 5.2, y: 15.9),
                control2: CGPoint(x: 8.2, y: 13.5)
            )
            shoulders.addCurve(
                to: CGPoint(x: 18.8, y: 19.4),
                control1: CGPoint(x: 15.8, y: 13.5),
                control2: CGPoint(x: 18.8, y: 15.9)
            )
            context.stroke(shoulders.applying(transform), with: .foreground, style: style)
        }
    }
}

private struct EventMark: View {
    var body: some View {
        Canvas { context, size in
            let scale = min(size.width, size.height) / 24
            let transform = CGAffineTransform(scaleX: scale, y: scale)
            let style = StrokeStyle(lineWidth: 1.4 * scale, lineCap: .round, lineJoin: .round)

            var top = Path()
            top.move(to: CGPoint(x: 6.6, y: 3.6))
            top.addLine(to: CGPoint(x: 17.4, y: 3.6))
            context.stroke(top.applying(transform), with: .foreground, style: style)

            var bottom = Path()
            bottom.move(to: CGPoint(x: 6.6, y: 20.4))
            bottom.addLine(to: CGPoint(x: 17.4, y: 20.4))
            context.stroke(bottom.applying(transform), with: .foreground, style: style)

            // M8.2 3.6v2.9c0 2.1 3.8 3.4 3.8 5.5s-3.8 3.4-3.8 5.5v2.9
            var left = Path()
            left.move(to: CGPoint(x: 8.2, y: 3.6))
            left.addLine(to: CGPoint(x: 8.2, y: 6.5))
            left.addCurve(
                to: CGPoint(x: 12, y: 12),
                control1: CGPoint(x: 8.2, y: 8.6),
                control2: CGPoint(x: 12, y: 9.9)
            )
            left.addCurve(
                to: CGPoint(x: 8.2, y: 17.5),
                control1: CGPoint(x: 12, y: 14.1),
                control2: CGPoint(x: 8.2, y: 15.4)
            )
            left.addLine(to: CGPoint(x: 8.2, y: 20.4))
            context.stroke(left.applying(transform), with: .foreground, style: style)

            // M15.8 3.6v2.9c0 2.1-3.8 3.4-3.8 5.5s3.8 3.4 3.8 5.5v2.9
            var right = Path()
            right.move(to: CGPoint(x: 15.8, y: 3.6))
            right.addLine(to: CGPoint(x: 15.8, y: 6.5))
            right.addCurve(
                to: CGPoint(x: 12, y: 12),
                control1: CGPoint(x: 15.8, y: 8.6),
                control2: CGPoint(x: 12, y: 9.9)
            )
            right.addCurve(
                to: CGPoint(x: 15.8, y: 17.5),
                control1: CGPoint(x: 12, y: 14.1),
                control2: CGPoint(x: 15.8, y: 15.4)
            )
            right.addLine(to: CGPoint(x: 15.8, y: 20.4))
            context.stroke(right.applying(transform), with: .foreground, style: style)
        }
    }
}

private struct PlaceMark: View {
    var body: some View {
        Canvas { context, size in
            let scale = min(size.width, size.height) / 24
            let transform = CGAffineTransform(scaleX: scale, y: scale)
            let style = StrokeStyle(lineWidth: 1.4 * scale, lineCap: .round, lineJoin: .round)

            var pole = Path()
            pole.move(to: CGPoint(x: 8.6, y: 21))
            pole.addLine(to: CGPoint(x: 8.6, y: 3.4))
            context.stroke(pole.applying(transform), with: .foreground, style: style)

            // M8.6 6.1h9.1l2.4 3-2.4 3H8.6z
            var flag = Path()
            flag.move(to: CGPoint(x: 8.6, y: 6.1))
            flag.addLine(to: CGPoint(x: 17.7, y: 6.1))
            flag.addLine(to: CGPoint(x: 20.1, y: 9.1))
            flag.addLine(to: CGPoint(x: 17.7, y: 12.1))
            flag.addLine(to: CGPoint(x: 8.6, y: 12.1))
            flag.closeSubpath()
            context.stroke(flag.applying(transform), with: .foreground, style: style)

            var base = Path()
            base.move(to: CGPoint(x: 5.4, y: 21))
            base.addLine(to: CGPoint(x: 11.8, y: 21))
            context.stroke(base.applying(transform), with: .foreground, style: style)
        }
    }
}

/// Quiet bridge marks (S6-04) — no pigment; same stroke language as primaries.
private struct RelationshipMark: View {
    var body: some View {
        Canvas { context, size in
            let scale = min(size.width, size.height) / 24
            let transform = CGAffineTransform(scaleX: scale, y: scale)
            let style = StrokeStyle(lineWidth: 1.4 * scale, lineCap: .round, lineJoin: .round)

            let left = Path(ellipseIn: CGRect(x: 4.5, y: 9, width: 6, height: 6))
            let right = Path(ellipseIn: CGRect(x: 13.5, y: 9, width: 6, height: 6))
            context.stroke(left.applying(transform), with: .foreground, style: style)
            context.stroke(right.applying(transform), with: .foreground, style: style)

            var link = Path()
            link.move(to: CGPoint(x: 10.5, y: 12))
            link.addLine(to: CGPoint(x: 13.5, y: 12))
            context.stroke(link.applying(transform), with: .foreground, style: style)
        }
    }
}

private struct ParticipationMark: View {
    var body: some View {
        Canvas { context, size in
            let scale = min(size.width, size.height) / 24
            let transform = CGAffineTransform(scaleX: scale, y: scale)
            let style = StrokeStyle(lineWidth: 1.4 * scale, lineCap: .round, lineJoin: .round)

            let person = Path(ellipseIn: CGRect(x: 5, y: 5.5, width: 5.5, height: 5.5))
            context.stroke(person.applying(transform), with: .foreground, style: style)

            var shoulders = Path()
            shoulders.move(to: CGPoint(x: 4, y: 18))
            shoulders.addQuadCurve(to: CGPoint(x: 11.5, y: 18), control: CGPoint(x: 7.75, y: 13.5))
            context.stroke(shoulders.applying(transform), with: .foreground, style: style)

            var event = Path()
            event.move(to: CGPoint(x: 14, y: 7))
            event.addLine(to: CGPoint(x: 20, y: 7))
            event.move(to: CGPoint(x: 14, y: 17))
            event.addLine(to: CGPoint(x: 20, y: 17))
            event.move(to: CGPoint(x: 17, y: 7))
            event.addLine(to: CGPoint(x: 17, y: 17))
            context.stroke(event.applying(transform), with: .foreground, style: style)
        }
    }
}

private struct LocationMark: View {
    var body: some View {
        Canvas { context, size in
            let scale = min(size.width, size.height) / 24
            let transform = CGAffineTransform(scaleX: scale, y: scale)
            let style = StrokeStyle(lineWidth: 1.4 * scale, lineCap: .round, lineJoin: .round)

            var pin = Path()
            pin.move(to: CGPoint(x: 12, y: 20))
            pin.addCurve(
                to: CGPoint(x: 12, y: 5),
                control1: CGPoint(x: 5.5, y: 14),
                control2: CGPoint(x: 5.5, y: 7)
            )
            pin.addCurve(
                to: CGPoint(x: 12, y: 20),
                control1: CGPoint(x: 18.5, y: 7),
                control2: CGPoint(x: 18.5, y: 14)
            )
            context.stroke(pin.applying(transform), with: .foreground, style: style)

            let hole = Path(ellipseIn: CGRect(x: 10, y: 8.5, width: 4, height: 4))
            context.stroke(hole.applying(transform), with: .foreground, style: style)
        }
    }
}

/// Open folio for Source reification (S7-D2 board — same 24×24 grid as S6 marks).
private struct SourceMark: View {
    var body: some View {
        Canvas { context, size in
            let scale = min(size.width, size.height) / 24
            let transform = CGAffineTransform(scaleX: scale, y: scale)
            let style = StrokeStyle(lineWidth: 1.5 * scale, lineCap: .round, lineJoin: .round)

            var left = Path()
            left.move(to: CGPoint(x: 12, y: 5.5))
            left.addLine(to: CGPoint(x: 5.5, y: 5.5))
            left.addLine(to: CGPoint(x: 5.5, y: 18.5))
            left.addLine(to: CGPoint(x: 12, y: 18.5))
            context.stroke(left.applying(transform), with: .foreground, style: style)

            var right = Path()
            right.move(to: CGPoint(x: 12, y: 5.5))
            right.addLine(to: CGPoint(x: 18.5, y: 5.5))
            right.addLine(to: CGPoint(x: 18.5, y: 18.5))
            right.addLine(to: CGPoint(x: 12, y: 18.5))
            context.stroke(right.applying(transform), with: .foreground, style: style)

            var spine = Path()
            spine.move(to: CGPoint(x: 12, y: 5.5))
            spine.addLine(to: CGPoint(x: 12, y: 18.5))
            context.stroke(spine.applying(transform), with: .foreground, style: style)
        }
    }
}
