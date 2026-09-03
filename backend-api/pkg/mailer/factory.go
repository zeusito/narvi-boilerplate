package mailer

import (
	"context"
)

type Mailer interface {
	SendOTPCode(ctx context.Context, email, code string) error
	SendInvitation(ctx context.Context, email, kind string) error
}

func NewFakeMailer() Mailer {
	return &FakeMailer{}
}
