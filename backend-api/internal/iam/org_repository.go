package iam

import (
	"context"
	"database/sql"
	"errors"

	"backend-api/pkg/terrors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

type defaultOrganizationRepository struct {
	db bun.IDB
}

func newOrganizationRepository(db bun.IDB) organizationRepository {
	return &defaultOrganizationRepository{db: db}
}

func (r *defaultOrganizationRepository) WithTx(tx bun.Tx) organizationRepository {
	return &defaultOrganizationRepository{db: tx}
}

func (r *defaultOrganizationRepository) Create(ctx context.Context, org *Organization) error {
	_, err := r.db.NewInsert().Model(org).Exec(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return terrors.RecordAlreadyExists("organization already exists")
		}

		log.Ctx(ctx).Error().Err(err).Str("org_id", org.ID).Msg("failed to insert organization")
		return terrors.OperationFailed("failed to create organization")
	}
	return nil
}

func (r *defaultOrganizationRepository) Update(ctx context.Context, org *Organization) error {
	res, err := r.db.NewUpdate().Model(org).Where("id = ?", org.ID).Exec(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return terrors.RecordAlreadyExists("organization already exists")
		}

		log.Ctx(ctx).Error().Err(err).Str("org_id", org.ID).Msg("failed to update organization")
		return terrors.OperationFailed("failed to update organization")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return terrors.RecordNotFound("organization not found")
	}

	return nil
}

func (r *defaultOrganizationRepository) FindOneByID(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	err := r.db.NewSelect().Model(&org).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, terrors.RecordNotFound("organization not found")
		}
		log.Ctx(ctx).Error().Err(err).Str("org_id", id).Msg("failed to query organization by id")
		return nil, terrors.OperationFailed("failed to retrieve organization")
	}
	return &org, nil
}

func (r *defaultOrganizationRepository) FindOneBySlug(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	err := r.db.NewSelect().Model(&org).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, terrors.RecordNotFound("organization not found")
		}
		log.Ctx(ctx).Error().Err(err).Str("slug", slug).Msg("failed to query organization by slug")
		return nil, terrors.OperationFailed("failed to retrieve organization")
	}
	return &org, nil
}

func (r *defaultOrganizationRepository) FindAll(ctx context.Context) ([]Organization, error) {
	var orgs []Organization
	err := r.db.NewSelect().Model(&orgs).
		Limit(100).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to query organizations")
		return nil, terrors.OperationFailed("failed to retrieve organizations")
	}

	return orgs, nil
}

func (r *defaultOrganizationRepository) FindAllMembershipsByOrganizationID(ctx context.Context, orgID string) ([]OrganizationMembershipView, error) {
	var memberships []OrganizationMembershipView
	err := r.db.NewSelect().
		Model(&memberships).
		Where("organization_id = ?", orgID).
		Limit(100).
		Order("membership_created_at DESC").
		Scan(ctx)

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("org_id", orgID).Msg("failed to query memberships by organization id")
		return nil, terrors.OperationFailed("failed to retrieve organization memberships")
	}
	return memberships, nil
}

func (r *defaultOrganizationRepository) FindAllMembershipsByIdentityID(ctx context.Context, identityID string) ([]OrganizationMembershipView, error) {
	var memberships []OrganizationMembershipView
	err := r.db.NewSelect().
		Model(&memberships).
		Where("identity_id = ?", identityID).
		Limit(10).
		Order("membership_created_at DESC").
		Scan(ctx)

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("identity_id", identityID).Msg("failed to query memberships by identity id")
		return nil, terrors.OperationFailed("failed to retrieve identity memberships")
	}
	return memberships, nil
}

func (r *defaultOrganizationRepository) FindOldestMembershipsByIdentityID(ctx context.Context, identityID string) (*OrganizationMembershipView, error) {
	var membership OrganizationMembershipView
	err := r.db.NewSelect().
		Model(&membership).
		Where("identity_id = ?", identityID).
		Order("membership_created_at ASC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, terrors.RecordNotFound("membership not found")
		}
		log.Ctx(ctx).Error().Err(err).Str("identity_id", identityID).Msg("failed to query oldest membership by identity id")
		return nil, terrors.OperationFailed("failed to retrieve oldest membership")
	}
	return &membership, nil
}
