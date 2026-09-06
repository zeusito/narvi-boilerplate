package mailer

import (
	"context"
	"sync"
)

// SpyMailer is a mailer that is used for testing purposes. Don't use it in production.
type SpyMailer struct {
	mu       sync.Mutex
	lastCode string
}

func NewSpyMailer() *SpyMailer {
	return &SpyMailer{}
}

func (s *SpyMailer) SendOTPCode(ctx context.Context, email, code string) error {
	s.mu.Lock()
	s.lastCode = code
	s.mu.Unlock()
	return nil
}

func (s *SpyMailer) SendInvitation(ctx context.Context, email, kind string) error {
	return nil
}

func (s *SpyMailer) LastCode() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastCode
}
