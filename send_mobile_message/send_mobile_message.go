package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	twilio "github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type ReceivedData struct {
	UID     string `json:"uid"`     // This UID is of the user who placed the order
	Message string `json:"message"` // Message to be sent to admins
}

func init() {
	if os.Getenv("ENV") != "DEBUG" {
		functions.HTTP("send-mobile-message", SendMobileMessage)
	}
}

func SendMobileMessage(response http.ResponseWriter, request *http.Request) {
	if shared.CorsEnabledFunction(response, request) {
		return
	}

	if request.Method != http.MethodPost {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method, expected POST")
		return
	}

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

	recipients := os.Getenv("TWILIO_RECIPENTS_PHONE")
	if recipients == "" {
		shared.WriteJSONError(response, http.StatusInternalServerError, "No recipients configured")
		return
	}
	toPhones := strings.Split(recipients, ";")

	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromPhone := os.Getenv("TWILIO_FROM_PHONE")

	if accountSid == "" || authToken == "" || fromPhone == "" {
		shared.WriteJSONError(response, http.StatusInternalServerError, "Twilio configuration incomplete")
		return
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	var failedRecipients []string
	for _, receiverPhone := range toPhones {
		params := &twilioApi.CreateMessageParams{}
		params.SetTo(receiverPhone)
		params.SetFrom(fromPhone)
		params.SetBody(data.Message)

		_, err := client.Api.CreateMessage(params)
		if err != nil {
			log.Printf("Error sending message to %s: %v", receiverPhone, err)
			failedRecipients = append(failedRecipients, receiverPhone)
		}
	}

	if len(failedRecipients) > 0 {
		log.Printf("Failed to send mobile messages to some recipients %v", failedRecipients)
		return
	}
}