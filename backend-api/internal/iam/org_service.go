package iam

import (
	"backend-api/pkg/terrors"
	"backend-api/pkg/toolbox"
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type defaultOrgService struct {
	orgRepo organizationRepository
}

func newOrgService(orgRepo organizationRepository) orgService {
	return &defaultOrgService{
		orgRepo: orgRepo,
	}
}

func (s *defaultOrgService) Create(ctx context.Context, req *CreateOrganizationRequest) (*OrganizationResponse, error) {
	log.Ctx(ctx).Info().Msgf("Creating organization with name %s", req.Name)

	now := time.Now().UTC()

	record := &Organization{
		ID:        toolbox.GenerateTypeId(toolbox.TypeIdPrefixOrganization),
		Name:      req.Name,
		Kind:      req.Kind,
		Slug:      toolbox.GenerateTypeId(toolbox.TypeIdPrefixOrgSlug),
		State:     OrganizationStateActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.orgRepo.Create(ctx, record)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to create organization")
		return nil, terrors.OperationFailed("failed to create record")
	}

	log.Ctx(ctx).Info().Msg("Organization created successfully")

	return &OrganizationResponse{
		ID:        record.ID,
		Name:      record.Name,
		Slug:      record.Slug,
		State:     record.State,
		CreatedAt: record.CreatedAt.Format(time.RFC3339),
		UpdatedAt: record.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *defaultOrgService) GetById(ctx context.Context, id string) (*OrganizationResponse, error) {
	log.Ctx(ctx).Info().Msgf("Getting organization with id %s", id)

	record, err := s.orgRepo.FindOneByID(ctx, id)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get organization")
		return nil, terrors.RecordNotFound("organization not found")
	}

	return &OrganizationResponse{
		ID:        record.ID,
		Name:      record.Name,
		Slug:      record.Slug,
		State:     record.State,
		CreatedAt: record.CreatedAt.Format(time.RFC3339),
		UpdatedAt: record.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *defaultOrgService) GetAll(ctx context.Context) *OrganizationListResponse {
	log.Ctx(ctx).Info().Msg("Getting all organizations")

	records, err := s.orgRepo.FindAll(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get organizations")
		return &OrganizationListResponse{Count: 0}
	}

	data := make([]OrganizationResponse, 0, len(records))
	for _, record := range records {
		data = append(data, OrganizationResponse{
			ID:        record.ID,
			Name:      record.Name,
			Slug:      record.Slug,
			State:     record.State,
			CreatedAt: record.CreatedAt.Format(time.RFC3339),
			UpdatedAt: record.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &OrganizationListResponse{
		Count: len(data),
		Data:  data,
	}
}
