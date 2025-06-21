package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	twilio "github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

var (
	TWILIO_RECIPIENTS_PHONE string
	TWILIO_ACCOUNT_SID      string
	TWILIO_AUTH_TOKEN       string
	TWILIO_FROM_PHONE       string
)

// ReceivedData represents the payload received to trigger the SMS message.
type ReceivedData struct {
	UID     string `json:"uid"`     // UID of the user triggering the notification
	Message string `json:"message"` // Message body to be sent
}

// init initializes configuration and HTTP handler for the function.
// In production, Twilio configuration is fetched securely from Google Secret Manager.
// In DEBUG mode, configuration is expected via environment variables.
func init() {
	if os.Getenv("ENV") != "DEBUG" {
		ctx := context.Background()
		projectID, err := metadata.ProjectIDWithContext(ctx)
		if err != nil {
			log.Fatalf("Failed to retrieve project ID from metadata: %v", err)
		}

		recipientsPhonePath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, "TWILIO_RECIPENTS_PHONE")
		accountSIDPath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, "TWILIO_ACCOUNT_SID")
		authTokenPath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, "TWILIO_AUTH_TOKEN")
		fromPhonePath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, "TWILIO_FROM_PHONE")

		TWILIO_RECIPIENTS_PHONE, err = shared.GetSecretFromGCP(recipientsPhonePath)
		if err != nil {
			log.Fatalf("Failed to fetch TWILIO_RECIPIENTS_PHONE from Secret Manager: %v", err)
		}
		TWILIO_ACCOUNT_SID, err = shared.GetSecretFromGCP(accountSIDPath)
		if err != nil {
			log.Fatalf("Failed to fetch TWILIO_ACCOUNT_SID from Secret Manager: %v", err)
		}
		TWILIO_AUTH_TOKEN, err = shared.GetSecretFromGCP(authTokenPath)
		if err != nil {
			log.Fatalf("Failed to fetch TWILIO_AUTH_TOKEN from Secret Manager: %v", err)
		}
		TWILIO_FROM_PHONE, err = shared.GetSecretFromGCP(fromPhonePath)
		if err != nil {
			log.Fatalf("Failed to fetch TWILIO_FROM_PHONE from Secret Manager: %v", err)
		}

		functions.HTTP("send-mobile-message", SendMobileMessage)
	} else {
		// Local environment configuration
		TWILIO_RECIPIENTS_PHONE = os.Getenv("TWILIO_RECIPENTS_PHONE")
		TWILIO_ACCOUNT_SID = os.Getenv("TWILIO_ACCOUNT_SID")
		TWILIO_AUTH_TOKEN = os.Getenv("TWILIO_AUTH_TOKEN")
		TWILIO_FROM_PHONE = os.Getenv("TWILIO_FROM_PHONE")
	}
}

// SendMobileMessage is the HTTP handler for sending SMS messages via Twilio.
// It expects a POST request with a JSON body conforming to the ReceivedData structure.
func SendMobileMessage(response http.ResponseWriter, request *http.Request) {
	// Handle CORS and preflight OPTIONS requests.
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
	
	// Parse and validate request body.
	var data ReceivedData
	if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
		log.Printf("Error decoding message body: %v", err)
		shared.WriteJSONError(response, http.StatusBadRequest, "Invalid request body for SMS message")
		return
	}

	if data.UID == "" {
		shared.WriteJSONError(response, http.StatusBadRequest, "UID is required")
		return
	}

	if data.Message == "" {
		shared.WriteJSONError(response, http.StatusBadRequest, "Message body is required")
		return
	}

	if TWILIO_RECIPIENTS_PHONE == "" || TWILIO_ACCOUNT_SID == "" || TWILIO_AUTH_TOKEN == "" || TWILIO_FROM_PHONE == "" {
		shared.WriteJSONError(response, http.StatusInternalServerError, "Twilio configuration incomplete")
		return
	}

	toPhones := strings.Split(TWILIO_RECIPIENTS_PHONE, ";")

	// Initialize Twilio client
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: TWILIO_ACCOUNT_SID,
		Password: TWILIO_AUTH_TOKEN,
	})

	var failedRecipients []string

	// Send message to each recipient
	for _, receiverPhone := range toPhones {
		params := &twilioApi.CreateMessageParams{}
		params.SetTo(receiverPhone)
		params.SetFrom(TWILIO_FROM_PHONE)
		params.SetBody(data.Message)

		_, err := client.Api.CreateMessage(params)
		if err != nil {
			log.Printf("Error sending message to %s: %v", receiverPhone, err)
			failedRecipients = append(failedRecipients, receiverPhone)
		}
	}

	// Log result
	if len(failedRecipients) > 0 {
		log.Printf("Failed to send mobile messages to some recipients %v", failedRecipients)
		shared.WriteJSONError(response, http.StatusInternalServerError, "Failed to send message to some recipients")
		return
	}

	log.Printf("Mobile message sent successfully to all recipients")
}
