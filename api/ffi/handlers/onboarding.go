package handlers

import (
	"fmt"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core/onboarding"
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
	res, err := onboarding.Open(req.GetIdentityDir(), req.GetProjectDir(), req.GetDisplayName())
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.OpenProjectResponse{
		ProjectDir:  res.ProjectDir,
		UserId:      res.Identity.UserID.String(),
		DisplayName: res.Identity.DisplayName,
	})
}
