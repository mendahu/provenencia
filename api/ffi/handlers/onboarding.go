package handlers

import (
	"fmt"

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
		})
	}
	return proto.Marshal(resp)
}
