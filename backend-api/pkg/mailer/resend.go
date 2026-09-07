package mailer

import (
	"backend-api/pkg/configurer"
	"context"

	"github.com/resend/resend-go/v4"
)

type ResendMailer struct {
	enabled        bool
	from           string
	client         *resend.Client
	otpTemplate    string
	inviteTemplate string
}

func NewResendMailer(emailConfigs configurer.EmailConfigurations) *ResendMailer {
	client := resend.NewClient(emailConfigs.ApiKey)

	return &ResendMailer{
		enabled:        emailConfigs.Enabled,
		from:           emailConfigs.From,
		client:         client,
		otpTemplate:    emailConfigs.OtpTemplate,
		inviteTemplate: emailConfigs.InviteTemplate,
	}
}

func (r *ResendMailer) SendOTPCode(ctx context.Context, email, code string) error {
	if !r.enabled {
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      []string{email},
		Subject: "Your Sign-In Code",
		Template: &resend.EmailTemplate{
			Id:        r.otpTemplate,
			Variables: map[string]any{},
		},
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResendMailer) SendInvitation(ctx context.Context, email, kind string) error {
	if !r.enabled {
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      []string{email},
		Subject: "You have been invited",
		Template: &resend.EmailTemplate{
			Id:        r.inviteTemplate,
			Variables: map[string]any{},
		},
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}
