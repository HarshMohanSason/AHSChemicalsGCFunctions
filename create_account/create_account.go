package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"firebase.google.com/go/v4/auth"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

// CreateAccountRequest represents the expected JSON structure in the request body
type CreateAccountRequest struct {
	Name        string              `json:"name"`         // Full name of the user (required)
	PhoneNumber string              `json:"phone_number"` // User's phone number (required)
	Email       string              `json:"email"`        // User's email (required)
	Password    string              `json:"password"`     // User's password (required)
	Properties  []map[string]string `json:"properties"`   // List of address properties (required)
	Brands      []string            `json:"brands"`       // List of brands associated with the user (required)
}

func init() {
	// Initialize Firebase and register the Cloud Function for production environment only
	if os.Getenv("ENV") != "DEBUG" {
		shared.InitFirebaseProd(nil)
		functions.HTTP("create-account", CreateAccount)
	}
}

// CreateAccount handles the creation of a new user in Firebase Authentication
// and stores additional user data (properties, brands) in Firestore.
//
// Authorization: Requires a valid Bearer token with 'admin' custom claim set to true.
// Method: POST
// Request Body: JSON matching the CreateAccountRequest struct
// Success Response: 200 OK with success message
// Error Response: Appropriate HTTP status codes with descriptive error messages
func CreateAccount(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	// Handle CORS (OPTIONS) requests and setup CORS headers
	if shared.CorsEnabledFunction(response, request) {
		return
	}

	// Only allow POST method for creating accounts
	if request.Method != http.MethodPost {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method")
		return
	}

	// Ensure that the user is authorized (admin privileges)
	if err := shared.IsAuthorizedAndAdmin(request); err != nil {
		shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
		return
	}

	defer request.Body.Close()

	// Decode the JSON request body
	var createAccountRequest CreateAccountRequest
	if err := json.NewDecoder(request.Body).Decode(&createAccountRequest); err != nil {
		log.Printf("JSON decode error: %v", err)
		shared.WriteJSONError(response, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if len(createAccountRequest.Properties) == 0 {
		shared.WriteJSONError(response, http.StatusBadRequest, "At least one property is required")
		return
	}

	// Validate each provided property to ensure all required fields are present
	for _, property := range createAccountRequest.Properties {
		if property["street"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "No street selected in one of the properties")
			return
		}
		if property["city"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "No city selected in one of the properties")
			return
		}
		if property["county"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "No county selected in one of the properties")
			return
		}
		if property["state"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "No state selected in one of the properties")
			return
		}
		if property["postal"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "No postal found in one of the properties")
			return
		}
	}

	if len(createAccountRequest.Brands) == 0 {
		shared.WriteJSONError(response, http.StatusBadRequest, "At least one brand is required")
		return
	}

	// Create the user in Firebase Authentication
	userToCreate := (&auth.UserToCreate{}).
		Email(createAccountRequest.Email).
		Password(createAccountRequest.Password).
		DisplayName(createAccountRequest.Name).
		PhoneNumber(createAccountRequest.PhoneNumber)

	createdUser, err := shared.AuthClient.CreateUser(ctx, userToCreate)
	if err != nil {
		log.Printf("Error occurred when creating the user: %v", err)

		// Attempt to extract more readable Firebase error, if available
		firebaseError := shared.ExtractFirebaseErrorFromResponse(err)
		if firebaseError != nil {
			shared.WriteJSONError(response, http.StatusInternalServerError, firebaseError.Error.Message)
		} else {
			shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Prepare the Firestore document with the user's properties and brands
	data := map[string]any{
		"properties": createAccountRequest.Properties,
		"brands":     createAccountRequest.Brands,
	}

	// Store the additional user data in Firestore
	if _, err := shared.FirestoreClient.Collection("users").Doc(createdUser.UID).Set(ctx, data); err != nil {
		log.Printf("Firestore write error: %v", err)
		shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
		return
	}

	// Respond with success
	shared.WriteJSONSuccess(response, http.StatusOK, "Account Created Successfully", nil)
}