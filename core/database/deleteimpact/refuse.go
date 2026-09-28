package deleteimpact

// Codes maps Impact gates onto domain sentinels.
type Codes struct {
	InUse        error
	OriginLocked error
	EdgeLocked   error
	NotFound     error
	Infra        error
}

// Refuse returns nil when the report allows erase. Otherwise it returns the
// matching domain sentinel (falling back to InUse, then ErrInvalid).
func Refuse(report Report, codes Codes) error {
	if report.Allowed {
		return nil
	}
	switch report.Gate {
	case GateInbound:
		return firstErr(codes.InUse)
	case GateOriginLocked:
		return firstErr(codes.OriginLocked, codes.InUse)
	case GateEdgeLocked:
		return firstErr(codes.EdgeLocked, codes.InUse)
	case GateNotFound:
		return firstErr(codes.NotFound, ErrInvalid)
	case GateInfra:
		return firstErr(codes.Infra, codes.NotFound, ErrInvalid)
	default:
		return firstErr(codes.InUse, ErrInvalid)
	}
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return ErrInvalid
}
