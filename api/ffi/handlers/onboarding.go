package handlers

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/onboarding"
	"google.golang.org/protobuf/proto"
)

func CompleteOnboarding(in []byte) ([]byte, error) {
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
		Ref:         res.Identity.Ref,
		Project:     projectInfoProto(res.ProjectDir, res.Project),
	})
}

func OpenProject(in []byte) ([]byte, error) {
	var req engine.OpenProjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("open_project: unmarshal: %w", err)
	}
	res, err := onboarding.Open(req.GetIdentityDir(), req.GetProjectDir(), req.GetDisplayName(), req.GetAdoptUserId())
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.OpenProjectResponse{
		ProjectDir:  res.ProjectDir,
		UserId:      res.Identity.UserID.String(),
		DisplayName: res.Identity.DisplayName,
		Ref:         res.Identity.Ref,
		Project:     projectInfoProto(res.ProjectDir, res.Project),
	})
}

func ListProjectUsers(in []byte) ([]byte, error) {
	var req engine.ListProjectUsersRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("list_project_users: unmarshal: %w", err)
	}
	ids, err := onboarding.ListContributors(req.GetProjectDir())
	if err != nil {
		return nil, err
	}
	resp := &engine.ListProjectUsersResponse{}
	for _, u := range ids {
		id, err := uuid.FromBytes(u.ID)
		if err != nil {
			return nil, err
		}
		resp.Users = append(resp.Users, &engine.ProjectUser{
			UserId:      id.String(),
			DisplayName: u.DisplayName,
			Ref:         u.Ref,
		})
	}
	return proto.Marshal(resp)
}

func GetProjectInfo(in []byte) ([]byte, error) {
	var req engine.GetProjectInfoRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("get_project_info: unmarshal: %w", err)
	}
	info, err := onboarding.ProjectInfo(req.GetProjectDir())
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.GetProjectInfoResponse{
		Project: projectInfoProto(req.GetProjectDir(), info),
	})
}

func projectInfoProto(projectDir string, info onboarding.ResolvedInfo) *engine.ProjectInfo {
	return &engine.ProjectInfo{
		Label:                info.Info.Label,
		FolderName:           filepath.Base(projectDir),
		CreatedAt:            info.Info.CreatedAt,
		UpdatedAt:            info.Info.UpdatedAt,
		UpdatedByUserId:      info.UpdatedByUserID,
		UpdatedByDisplayName: info.UpdatedByDisplayName,
		UpdatedByRef:         info.UpdatedByRef,
	}
}
