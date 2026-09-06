package iam

import "context"

type defaultSessionService struct {
	sessionRepo sessionRepository
}

func newDefaultSessionService(sessionRepo sessionRepository) sessionService {
	return &defaultSessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *defaultSessionService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionRepo.RemoveBySessionID(ctx, sessionID)
}
