package email

import (
	"context"
	"fmt"
	"os"

	"github.com/resend/resend-go/v3"
)

type EmailService interface {
	SendMagicLink(ctx context.Context, to string, link string) error
	SendAndroidInvite(ctx context.Context, to string, playStoreLink string) error
	SendIOSInvite(ctx context.Context, to string, testFlightLink string) error
}

type Service struct {
	Client    *resend.Client
	FromEmail string
}

func NewService() *Service {
	apiKey := os.Getenv("RESEND_API_KEY")
	return &Service{
		Client:    resend.NewClient(apiKey),
		FromEmail: os.Getenv("FROM_EMAIL"),
	}
}

func (s *Service) SendMagicLink(ctx context.Context, to string, link string) error {
	params := &resend.SendEmailRequest{
		From:    s.FromEmail,
		To:      []string{to},
		Subject: "Tester admin dashboard Access link",
		Html:    buildMagicLinkHTML(link),
	}
	return s.send(ctx, params)
}

func (s *Service) SendAndroidInvite(ctx context.Context, to string, playStoreLink string) error {
	params := &resend.SendEmailRequest{
		From:    s.FromEmail,
		To:      []string{to},
		Subject: "You've been approved to test ThePrice on Android",
		Html:    buildAndroidInviteHTML(playStoreLink),
	}
	return s.send(ctx, params)
}

func (s *Service) SendIOSInvite(ctx context.Context, to string, testFlightLink string) error {
	params := &resend.SendEmailRequest{
		From:    s.FromEmail,
		To:      []string{to},
		Subject: "You've been approved to test ThePrice on iOS",
		Html:    buildIOSInviteHTML(testFlightLink),
	}
	return s.send(ctx, params)
}

func (s *Service) send(ctx context.Context, params *resend.SendEmailRequest) error {
	_, err := s.Client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to email to %s: %w", params.To, err)
	}
	return nil
}

func buildMagicLinkHTML(link string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; justify-content: center; align-items: center;">
    <div style="background-color: #f4f4f4; padding: 20px; border-radius: 5px;">
        <h2 style="color: #2c3e50;">ThePrice Admin Access</h2>
        <p>Click the link below to access your admin dashboard:</p>
        <p style="margin: 30px 0;">
            <a href="%s" style="display: inline-block; padding: 12px 30px; background-color: #3498db; color: #ffffff; text-decoration: none; border-radius: 5px; font-weight: bold;">Login to ThePrice Admin</a>
        </p>
        <p style="color: #e74c3c; font-weight: bold;">This link expires in 2 minutes.</p>
        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
        <p style="font-size: 12px; color: #7f8c8d;">
            If you didn't request this link, please ignore this email.
        </p>
    </div>
</body>
</html>
    `, link)
}

func buildAndroidInviteHTML(playLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f4f4f4; padding: 20px; border-radius: 5px;">
        <h2 style="color: #27ae60;">Congratulations! 🎉</h2>
        <p>You've been approved to test <strong>ThePrice</strong> on Android!</p>

        <h3 style="color: #2c3e50;">How to join the test:</h3>
        <ol style="line-height: 2;">
            <li>Click the button below using your <strong>Android device</strong></li>
            <li>Make sure you're signed in with the <strong>same Google account</strong> you used to sign up</li>
            <li>Accept the invitation to join the beta test</li>
            <li>Download and install ThePrice from the Play Store</li>
        </ol>

        <p style="margin: 30px 0;">
            <a href="%s" style="display: inline-block; padding: 12px 30px; background-color: #27ae60; color: #ffffff; text-decoration: none; border-radius: 5px; font-weight: bold;">Join Android Beta Test</a>
        </p>

        <div style="background-color: #fff3cd; padding: 15px; border-left: 4px solid #ffc107; margin: 20px 0;">
            <p style="margin: 0;"><strong>⚠️ Important:</strong></p>
            <ul style="margin: 10px 0;">
                <li>Use the same Google account as your signup</li>
                <li>If the test doesn't appear immediately, wait a few minutes and try again</li>
            </ul>
        </div>

        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
        <p style="font-size: 12px; color: #7f8c8d;">
            Need help? Reply to this email and we'll assist you.
        </p>
    </div>
</body>
</html>
    `, playLink)
}

func buildIOSInviteHTML(testFlightLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f4f4f4; padding: 20px; border-radius: 5px;">
        <h2 style="color: #5856d6;">Congratulations! 🎉</h2>
        <p>You've been approved to test <strong>ThePrice</strong> on iOS!</p>

        <h3 style="color: #2c3e50;">How to join the test:</h3>
        <ol style="line-height: 2;">
            <li>Make sure you have <strong>TestFlight</strong> installed on your iPhone/iPad<br>
                <span style="font-size: 12px; color: #7f8c8d;">Download from the App Store if you don't have it</span>
            </li>
            <li>Click the button below using your <strong>iOS device</strong></li>
            <li>Accept the invitation in TestFlight</li>
            <li>Install and test ThePrice</li>
        </ol>

        <p style="margin: 30px 0;">
            <a href="%s" style="display: inline-block; padding: 12px 30px; background-color: #5856d6; color: #ffffff; text-decoration: none; border-radius: 5px; font-weight: bold;">Join iOS Beta Test</a>
        </p>

        <div style="background-color: #e3f2fd; padding: 15px; border-left: 4px solid #2196f3; margin: 20px 0;">
            <p style="margin: 0;"><strong>💡 Tip:</strong></p>
            <ul style="margin: 10px 0;">
                <li>Download TestFlight from the App Store first: <a href="https://apps.apple.com/app/testflight/id899247664" style="color: #2196f3;">Get TestFlight</a></li>
                <li>Ensure you have the latest version of TestFlight installed</li>
                <li>You may need to accept the invite within TestFlight after clicking the link</li>
            </ul>
        </div>

        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
        <p style="font-size: 12px; color: #7f8c8d;">
            Need help? Reply to this email and we'll assist you.
        </p>
    </div>
</body>
</html>
    `, testFlightLink)
}
