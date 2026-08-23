package ffi

import (
	"fmt"

	"github.com/mendahu/provenance/api/ffi/handlers"
)

// Method values match provenance.engine.v1.Method in engine.proto.
const (
	MethodUnspecified           int32 = 0
	MethodPing                  int32 = 1
	MethodGetVersion            int32 = 2
	MethodGetInstallIdentity    int32 = 3
	MethodCompleteOnboarding    int32 = 4
	MethodRemoveInstallIdentity int32 = 5
	MethodGetActiveProject      int32 = 6
	MethodOpenProject           int32 = 7
	MethodRemoveActiveProject   int32 = 8
)

// Call routes one coarse FFI operation to api/ffi/handlers.
func Call(method int32, in []byte) ([]byte, error) {
	switch method {
	case MethodPing:
		return handlers.Ping(in)
	case MethodGetVersion:
		return handlers.GetVersion(in)
	case MethodGetInstallIdentity:
		return handlers.GetInstallIdentity(in)
	case MethodCompleteOnboarding:
		return handlers.CompleteOnboarding(in)
	case MethodRemoveInstallIdentity:
		return handlers.RemoveInstallIdentity(in)
	case MethodGetActiveProject:
		return handlers.GetActiveProject(in)
	case MethodOpenProject:
		return handlers.OpenProject(in)
	case MethodRemoveActiveProject:
		return handlers.RemoveActiveProject(in)
	default:
		return nil, fmt.Errorf("unknown method %d", method)
	}
}
