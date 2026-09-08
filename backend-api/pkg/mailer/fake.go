package mailer

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

type FakeMailer struct {
	mu sync.RWMutex
}

func NewFakeMailer() Mailer {
	return &FakeMailer{}
}

func (m *FakeMailer) SendOTPCode(ctx context.Context, email, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info().Msgf("FakeMailer: OTP code email sent to %s: %s", email, code)

	return nil
}

func (m *FakeMailer) SendInvitation(ctx context.Context, email, kind string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info().Msgf("FakeMailer: Invitation email sent to %s of kind %s", email, kind)

	return nil
}
