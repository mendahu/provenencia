package handlers

import (
	"errors"
	"fmt"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/identity"
	"google.golang.org/protobuf/proto"
)

func GetInstallIdentity(in []byte) ([]byte, error) {
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
}

func RemoveInstallIdentity(in []byte) ([]byte, error) {
	var req engine.RemoveInstallIdentityRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("remove_install_identity: unmarshal: %w", err)
	}
	if err := identity.Remove(req.GetIdentityDir()); err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.RemoveInstallIdentityResponse{})
}
