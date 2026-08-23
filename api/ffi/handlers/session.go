package handlers

import (
	"errors"
	"fmt"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core/session"
	"google.golang.org/protobuf/proto"
)

func GetActiveProject(in []byte) ([]byte, error) {
	var req engine.GetActiveProjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("get_active_project: unmarshal: %w", err)
	}
	a, err := session.Load(req.GetIdentityDir())
	if errors.Is(err, session.ErrNotFound) {
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
	if err := session.Remove(req.GetIdentityDir()); err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.RemoveActiveProjectResponse{})
}
