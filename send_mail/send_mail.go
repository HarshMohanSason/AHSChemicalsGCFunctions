package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailMetaData defines the structure of the email payload expected in the request body.
type EmailMetaData struct {
	Recipients map[string]string `json:"recipients"` // Email address mapped to recipient name
	Data       map[string]any    `json:"data"`       // Dynamic template data for the email
	TemplateID string            `json:"template_id"`// SendGrid dynamic template ID
}

// Global variables to store the SendGrid API key and sender email address.
var (
	SENDGRID_API_KEY  string
	SENDGRID_FROM_MAIL string
)

// init initializes the function and loads configuration.
// In production mode (ENV != DEBUG), it retrieves secrets from Google Secret Manager.
// In debug mode, it falls back to environment variables for local testing.
func init() {
	if os.Getenv("ENV") != "DEBUG" {
		ctx := context.Background()
		projectID, err := metadata.ProjectIDWithContext(ctx)
		if err != nil{
			log.Fatalf("Failed to retrieve project ID from metadata: %v", err)
		}

		apiKeyPath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, "SENDGRID_API_KEY")
		fromMailPath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, "SENDGRID_FROM_MAIL")

		
		SENDGRID_API_KEY, err = shared.GetSecretFromGCP(apiKeyPath)
		if err != nil {
			log.Fatalf("Failed to fetch SENDGRID_API_KEY from Secret Manager: %v", err)
		}

		SENDGRID_FROM_MAIL, err = shared.GetSecretFromGCP(fromMailPath)
		if err != nil {
			log.Fatalf("Failed to fetch SENDGRID_FROM_MAIL from Secret Manager: %v", err)
		}

		functions.HTTP("send-mail", SendMail)
	} else {
		// Local development using environment variables
		SENDGRID_API_KEY = os.Getenv("SENDGRID_API_KEY")
		SENDGRID_FROM_MAIL = os.Getenv("SENDGRID_FROM_MAIL")
	}
}

// SendMail handles the HTTP POST request to send transactional emails using SendGrid.
// It validates the incoming request, constructs the email message, and sends it using the SendGrid API.

func SendMail(response http.ResponseWriter, request *http.Request) {
	// Enable CORS for allowed origins and handle preflight OPTIONS requests.
	if shared.CorsEnabledFunction(response, request) {
		return
	}

	// Ensure the request method is POST.
	if request.Method != http.MethodPost {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method, expected POST")
		return
	}

	// Ensure that the user is authorized 
	if err := shared.IsAuthorized(request); err != nil {
		shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
		return
	}
	
	// Decode the JSON body into EmailMetaData struct.
	var emailMetaData EmailMetaData
	if err := json.NewDecoder(request.Body).Decode(&emailMetaData); err != nil {
		log.Printf("Error decoding email metadata: %v", err)
		shared.WriteJSONError(response, http.StatusBadRequest, "Invalid request body for email metadata")
		return
	}

	// Construct the 'from' email.
	from := mail.NewEmail("AHSChemicals", SENDGRID_FROM_MAIL)

	// Build the list of recipients.
	var recipients []*mail.Email
	for email, name := range emailMetaData.Recipients {
		recipients = append(recipients, mail.NewEmail(name, email))
	}

	// Validate that at least one recipient is provided.
	if len(recipients) == 0 {
		shared.WriteJSONError(response, http.StatusBadRequest, "At least one recipient is required")
		return
	}

	// Validate that the SendGrid template ID is provided.
	if emailMetaData.TemplateID == "" {
		shared.WriteJSONError(response, http.StatusBadRequest, "Template ID is required")
		return
	}

	// Configure personalization for dynamic template data.
	p := mail.NewPersonalization()
	p.AddTos(recipients...)

	for key, value := range emailMetaData.Data {
		p.SetDynamicTemplateData(key, value)
	}

	// Create the SendGrid email message.
	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.AddPersonalizations(p)
	message.SetTemplateID(emailMetaData.TemplateID)

	// Send the email using SendGrid API.
	client := sendgrid.NewSendClient(SENDGRID_API_KEY)
	_, err := client.Send(message)
	if err != nil {
		log.Printf("Error sending email to recipients %v: %v", recipients, err)
		shared.WriteJSONError(response, http.StatusInternalServerError, "Failed to send email")
		return
	}

	//Only logging successfully sent emails. Not sending responses
	log.Printf("Email sent successfully to %v recipients using template %s", recipients, emailMetaData.TemplateID)
}