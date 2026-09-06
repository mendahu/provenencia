package handlers

import (
	"errors"
	"fmt"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/identity"
	"google.golang.org/protobuf/proto"
)

func GetActiveProject(in []byte) ([]byte, error) {
	var req engine.GetActiveProjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("get_active_project: unmarshal: %w", err)
	}
	a, err := identity.LoadActive(req.GetIdentityDir())
	if errors.Is(err, identity.ErrActiveNotFound) {
		return proto.Marshal(&engine.GetActiveProjectResponse{Found: false})
	}
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.GetActiveProjectResponse{
		Found:      true,
		ProjectDir: a.ProjectDir,
	})
}

func RemoveActiveProject(in []byte) ([]byte, error) {
	var req engine.RemoveActiveProjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("remove_active_project: unmarshal: %w", err)
	}
	if err := identity.RemoveActive(req.GetIdentityDir()); err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.RemoveActiveProjectResponse{})
}

func SignOut(in []byte) ([]byte, error) {
	var req engine.SignOutRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("sign_out: unmarshal: %w", err)
	}
	dir := req.GetIdentityDir()
	if err := identity.RemoveActive(dir); err != nil {
		return nil, err
	}
	if err := identity.Remove(dir); err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.SignOutResponse{})
}
