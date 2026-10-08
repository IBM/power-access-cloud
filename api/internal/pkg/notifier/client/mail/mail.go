package mail

import (
	"errors"
	"fmt"
	"os"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/IBM/power-access-cloud/api/internal/pkg/notifier/client"
	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
)

var _ client.Notifier = &Mail{}

var l = log.GetLogger()

type Mail struct {
	request           rest.Request
	from              *mail.Email
	envPrefixForEmail string
}

func (m *Mail) Notify(event models.Event) error {
	m1 := mail.NewV3Mail()
	m1.SetFrom(m.from)
	plainTextContent, err := event.ComposeMailBody()
	if err != nil {
		return err
	}
	// SendGrid requires "text/plain", not bare "text".
	content := mail.NewContent("text/plain", plainTextContent)
	m1.AddContent(content)

	personalization := mail.NewPersonalization()
	hasRecipients := false
	if event.Notify {
		if event.UserEmail == "" {
			return errors.New("user email is required when user notification is enabled")
		}
		personalization.AddTos(mail.NewEmail("", event.UserEmail))
		hasRecipients = true
	}
	if event.NotifyAdmin {
		if hasRecipients {
			// A user To is already present — add admin as BCC so they get a
			// copy without appearing in the user's To field.
			personalization.AddBCCs(mail.NewEmail("IBM® Power® Access Cloud", "PowerACL@ibm.com"))
		} else {
			// Admin is the sole recipient. SendGrid rejects messages that have
			// BCC but no To address (HTTP 400), so use To directly here.
			personalization.AddTos(mail.NewEmail("IBM® Power® Access Cloud", "PowerACL@ibm.com"))
		}
		hasRecipients = true
	}
	if !hasRecipients {
		return errors.New("no notification recipients configured for event")
	}
	personalization.Subject = fmt.Sprintf("IBM® Power® Access Cloud - %s", event.Type)
	if m.envPrefixForEmail != "" {
		personalization.Subject = fmt.Sprintf("%s %s", m.envPrefixForEmail, personalization.Subject)
	}
	m1.AddPersonalizations(personalization)

	req := m.request
	req.Body = mail.GetRequestBody(m1)
	response, err := sendgrid.API(req)
	if err != nil {
		return fmt.Errorf("sendgrid API call failed: %w", err)
	}
	if response.StatusCode != 202 {
		return fmt.Errorf("sendgrid returned non-202 status %d: %s", response.StatusCode, response.Body)
	}

	return nil
}

func New(emailSubjectPrefix string) client.Notifier {
	key := os.Getenv("SENDGRID_API_KEY")
	if key == "" {
		l.Fatal("SENDGRID_API_KEY not set")
	}
	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "")
	request.Method = "POST"
	return &Mail{
		request:           request,
		from:              mail.NewEmail("IBM® Power® Access Cloud", "PowerACL@ibm.com"),
		envPrefixForEmail: emailSubjectPrefix,
	}
}
