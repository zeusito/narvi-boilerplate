package iam

import (
	"context"

	"github.com/uptrace/bun"
)

type defaultOrganizationRepository struct {
	db *bun.DB
}

func newOrganizationRepository(db *bun.DB) organizationRepository {
	return &defaultOrganizationRepository{db: db}
}

func (r *defaultOrganizationRepository) Create(ctx context.Context, org *Organization) error {
	_, err := r.db.NewInsert().Model(org).Exec(ctx)
	return err
}

func (r *defaultOrganizationRepository) FindOneByID(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	err := r.db.NewSelect().Model(&org).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *defaultOrganizationRepository) FindOneBySlug(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	err := r.db.NewSelect().Model(&org).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *defaultOrganizationRepository) FindMembershipsByOrganizationID(ctx context.Context, orgID string) ([]OrganizationMembershipView, error) {
	var memberships []OrganizationMembershipView
	err := r.db.NewSelect().
		Model(&memberships).
		Where("organization_id = ?", orgID).
		Limit(100).
		Order("membership_created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *defaultOrganizationRepository) FindMembershipsByIdentityID(ctx context.Context, identityID string) ([]OrganizationMembershipView, error) {
	var memberships []OrganizationMembershipView
	err := r.db.NewSelect().
		Model(&memberships).
		Where("identity_id = ?", identityID).
		Limit(10).
		Order("membership_created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return memberships, nil
}
