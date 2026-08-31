package handlers

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/users"
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
	pi := projectInfoProto(
		res.ProjectDir,
		res.Project,
		res.Identity.UserID.String(),
		res.Identity.DisplayName,
		res.Identity.Ref,
	)
	return proto.Marshal(&engine.CompleteOnboardingResponse{
		ProjectDir:  res.ProjectDir,
		UserId:      res.Identity.UserID.String(),
		DisplayName: res.Identity.DisplayName,
		Ref:         res.Identity.Ref,
		Project:     pi,
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
	pi, err := enrichProjectInfo(res.ProjectDir, res.Project)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.OpenProjectResponse{
		ProjectDir:  res.ProjectDir,
		UserId:      res.Identity.UserID.String(),
		DisplayName: res.Identity.DisplayName,
		Ref:         res.Identity.Ref,
		Project:     pi,
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
	for _, id := range ids {
		resp.Users = append(resp.Users, &engine.ProjectUser{
			UserId:      id.UserID.String(),
			DisplayName: id.DisplayName,
			Ref:         id.Ref,
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
	pi, err := enrichProjectInfo(req.GetProjectDir(), info)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.GetProjectInfoResponse{Project: pi})
}

func enrichProjectInfo(projectDir string, info project.Info) (*engine.ProjectInfo, error) {
	byID, byName, byRef := "", "", ""
	if len(info.UpdatedBy) == 16 {
		if uid, err := uuid.FromBytes(info.UpdatedBy); err == nil {
			byID = uid.String()
		}
		proj, err := database.Open(projectDir)
		if err != nil {
			return nil, err
		}
		u, err := users.Lookup(proj, info.UpdatedBy)
		_ = proj.Close()
		if err != nil {
			return nil, err
		}
		byName = u.DisplayName
		byRef = u.Ref
	}
	return projectInfoProto(projectDir, info, byID, byName, byRef), nil
}

func projectInfoProto(projectDir string, info project.Info, byID, byName, byRef string) *engine.ProjectInfo {
	return &engine.ProjectInfo{
		Label:                info.Label,
		FolderName:           filepath.Base(projectDir),
		CreatedAt:            info.CreatedAt,
		UpdatedAt:            info.UpdatedAt,
		UpdatedByUserId:      byID,
		UpdatedByDisplayName: byName,
		UpdatedByRef:         byRef,
	}
}
