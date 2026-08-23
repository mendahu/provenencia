package ffi

import (
	"errors"
	"fmt"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core"
	"github.com/mendahu/provenance/core/identity"
	"github.com/mendahu/provenance/core/onboarding"
	"google.golang.org/protobuf/proto"
)

// Method values match provenance.engine.v1.Method in engine.proto.
const (
	MethodUnspecified        int32 = 0
	MethodPing               int32 = 1
	MethodGetVersion         int32 = 2
	MethodGetInstallIdentity int32 = 3
	MethodCompleteOnboarding int32 = 4
)

// Call dispatches one coarse FFI operation. in/out are protobuf bytes.
func Call(method int32, in []byte) ([]byte, error) {
	switch method {
	case MethodPing:
		var req engine.PingRequest
		if err := proto.Unmarshal(in, &req); err != nil {
			return nil, fmt.Errorf("ping: unmarshal: %w", err)
		}
		return proto.Marshal(&engine.PingResponse{Message: req.GetMessage()})
	case MethodGetVersion:
		var req engine.GetVersionRequest
		if len(in) > 0 {
			if err := proto.Unmarshal(in, &req); err != nil {
				return nil, fmt.Errorf("get_version: unmarshal: %w", err)
			}
		}
		return proto.Marshal(&engine.GetVersionResponse{Version: core.Version})
	case MethodGetInstallIdentity:
		var req engine.GetInstallIdentityRequest
		if err := proto.Unmarshal(in, &req); err != nil {
			return nil, fmt.Errorf("get_install_identity: unmarshal: %w", err)
		}
		id, err := identity.Load(req.GetIdentityDir())
		if errors.Is(err, identity.ErrNotFound) {
			return proto.Marshal(&engine.GetInstallIdentityResponse{Found: false})
		}
		if err != nil {
			return nil, err
		}
		return proto.Marshal(&engine.GetInstallIdentityResponse{
			Found:       true,
			UserId:      id.UserID.String(),
			DisplayName: id.DisplayName,
		})
	case MethodCompleteOnboarding:
		var req engine.CompleteOnboardingRequest
		if err := proto.Unmarshal(in, &req); err != nil {
			return nil, fmt.Errorf("complete_onboarding: unmarshal: %w", err)
		}
		res, err := onboarding.Complete(req.GetIdentityDir(), req.GetParentDir(), req.GetDisplayName(), req.GetFamilyName())
		if err != nil {
			return nil, err
		}
		return proto.Marshal(&engine.CompleteOnboardingResponse{
			ProjectDir:  res.ProjectDir,
			UserId:      res.Identity.UserID.String(),
			DisplayName: res.Identity.DisplayName,
		})
	default:
		return nil, fmt.Errorf("unknown method %d", method)
	}
}
