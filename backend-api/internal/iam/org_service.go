package iam

import (
	"backend-api/pkg/terrors"
	"context"
)

type defaultOrgService struct {
	orgRepo organizationRepository
}

func NewOrgService(orgRepo organizationRepository) orgService {
	return &defaultOrgService{
		orgRepo: orgRepo,
	}
}

func (s *defaultOrgService) Create(ctx context.Context, req *CreateOrganizationRequest) (*OrganizationResponse, error) {
	return nil, terrors.Forbidden("Not implemented")
}

func (s *defaultOrgService) GetById(ctx context.Context, id string) (*OrganizationResponse, error) {
	return nil, terrors.Forbidden("Not implemented")
}

func (s *defaultOrgService) GetAll(ctx context.Context) *OrganizationListResponse {
	return nil
}
