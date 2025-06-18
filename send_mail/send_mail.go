package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailMetaData struct {
	Recipients map[string]string `json:"recipients"`
	Data       map[string]any    `json:"data"`
	TemplateID string            `json:"template_id"`
}

func init() {
	if os.Getenv("ENV") != "DEBUG" {
		functions.HTTP("send-mail", SendMail)
	}
}

func SendMail(response http.ResponseWriter, request *http.Request) {
	if shared.CorsEnabledFunction(response, request) {
		return
	}

	if request.Method != http.MethodPost {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method, expected POST")
		return
	}

	var emailMetaData EmailMetaData
	if err := json.NewDecoder(request.Body).Decode(&emailMetaData); err != nil {
		log.Printf("Error decoding email metadata: %v", err)
		shared.WriteJSONError(response, http.StatusBadRequest, "Invalid request body for email metadata")
		return
	}

	from := mail.NewEmail("AHSChemicals", os.Getenv("SENDGRID_FROM_MAIL"))

	var recipients []*mail.Email
	for email, name := range emailMetaData.Recipients {
		recipients = append(recipients, mail.NewEmail(name, email))
	}

	if len(recipients) == 0 {
		shared.WriteJSONError(response, http.StatusBadRequest, "At least one recipient is required")
		return
	}

	if emailMetaData.TemplateID == "" {
		shared.WriteJSONError(response, http.StatusBadRequest, "Template ID is required")
		return
	}

	p := mail.NewPersonalization()
	p.AddTos(recipients...)

	for key, value := range emailMetaData.Data {
		p.SetDynamicTemplateData(key, value)
	}

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.AddPersonalizations(p)
	message.SetTemplateID(emailMetaData.TemplateID)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	if err != nil {
		log.Printf("Error sending email to recipients %v: %v", recipients, err)
		shared.WriteJSONError(response, http.StatusInternalServerError, "Failed to send email")
		return
	}

	log.Printf("Email sent successfully to %v recipients using template %s", recipients, emailMetaData.TemplateID)
}