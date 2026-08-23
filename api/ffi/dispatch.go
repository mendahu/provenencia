package ffi

import (
	"fmt"

	"github.com/mendahu/provenance/api/ffi/handlers"
	"github.com/mendahu/provenance/api/proto/engine"
)

const (
	MethodUnspecified           = int32(engine.Method_METHOD_UNSPECIFIED)
	MethodPing                  = int32(engine.Method_METHOD_PING)
	MethodGetVersion            = int32(engine.Method_METHOD_GET_VERSION)
	MethodGetInstallIdentity    = int32(engine.Method_METHOD_GET_INSTALL_IDENTITY)
	MethodCompleteOnboarding    = int32(engine.Method_METHOD_COMPLETE_ONBOARDING)
	MethodRemoveInstallIdentity = int32(engine.Method_METHOD_REMOVE_INSTALL_IDENTITY)
	MethodGetActiveProject      = int32(engine.Method_METHOD_GET_ACTIVE_PROJECT)
	MethodOpenProject           = int32(engine.Method_METHOD_OPEN_PROJECT)
	MethodRemoveActiveProject   = int32(engine.Method_METHOD_REMOVE_ACTIVE_PROJECT)
	MethodListProjectUsers      = int32(engine.Method_METHOD_LIST_PROJECT_USERS)
	MethodSignOut               = int32(engine.Method_METHOD_SIGN_OUT)
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
	case MethodListProjectUsers:
		return handlers.ListProjectUsers(in)
	case MethodSignOut:
		return handlers.SignOut(in)
	default:
		return nil, fmt.Errorf("unknown method %d", method)
	}
}
