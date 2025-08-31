package mail

import (
	"fmt"
	"os"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"

	"github.com/PDeXchange/pac/internal/pkg/notifier/client"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
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
	content := mail.NewContent("text", plainTextContent)
	m1.AddContent(content)

	personalization := mail.NewPersonalization()
	to := mail.NewEmail("", event.UserEmail)
	// TODO: Add BCC to all the admins or to the group alias when we have it
	bcc := mail.NewEmail("IBM® Power® Access Cloud", "PowerACL@ibm.com")
	personalization.AddTos(to)
	if event.NotifyAdmin {
		personalization.AddBCCs(bcc)
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
		l.Error("Error sending mail", zap.Error(err))
	}

	if response.StatusCode != 202 {
		l.Error("Error sending mail, response code is not 202", zap.Int("code", response.StatusCode))
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
