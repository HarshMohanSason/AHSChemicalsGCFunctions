package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"reflect"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

// UpdateUserRequest defines the structure of the incoming JSON request
type UpdateUserRequest struct {
	UID        string                `json:"uid"`        // Firebase User ID (required)
	Brands     []string              `json:"brands"`     // List of brand names associated with the user (required)
	Properties []map[string]string   `json:"properties"` // List of property objects with fields: street, city, county, state, postal (required)
}

func init() {
	// Register the Cloud Function only in production environments
	if os.Getenv("ENV") != "DEBUG" {
		shared.InitFirebaseProd(nil)
		functions.HTTP("update-account", UpdateAccount)
	}
}

// UpdateAccount is a Google Cloud Function that updates a user's brands and properties
// in Firestore if they differ from the existing records.
//
// Authorization: Requires Firebase ID token with 'admin' custom claim set to true
// Method: PUT
// Request Body: JSON matching UpdateUserRequest structure
// Response: Success or error message with appropriate HTTP status code
func UpdateAccount(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	// Handle CORS and preflight requests
	if shared.CorsEnabledFunction(response, request) {
		return
	}

	// Only allow PUT method for this function
	if request.Method != http.MethodPut {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method, expected PUT")
		return
	}

	// Ensure that the user is authorized (admin privileges)
	if err := shared.IsAuthorizedAndAdmin(request); err != nil {
		shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
		return
	}

	defer request.Body.Close()

	// Decode the incoming JSON payload
	var user UpdateUserRequest
	if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request body: %v", err)
		shared.WriteJSONError(response, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if user.UID == "" {
		shared.WriteJSONError(response, http.StatusBadRequest, "UID is required")
		return
	}
	if len(user.Properties) == 0 {
		shared.WriteJSONError(response, http.StatusBadRequest, "User properties cannot be empty")
		return
	}
	for _, property := range user.Properties {
		if property["city"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "City is required in one of the properties")
			return
		}
		if property["county"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "County is required in one of the properties")
			return
		}
		if property["state"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "State is required in one of the properties")
			return
		}
		if property["postal"] == "" {
			shared.WriteJSONError(response, http.StatusBadRequest, "Postal code is required in one of the properties")
			return
		}
	}
	if len(user.Brands) == 0 {
		shared.WriteJSONError(response, http.StatusBadRequest, "User brands cannot be empty")
		return
	}

	// Fetch current user document from Firestore
	docSnapshot, err := shared.FirestoreClient.Collection("users").Doc(user.UID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching document for uid %v: %v", user.UID, err)
		shared.WriteJSONError(response, http.StatusBadRequest, "No user found with the given UID")
		return
	}

	// Prepare update operations by comparing incoming data with current Firestore document
	currentData := docSnapshot.Data()
	updates := []firestore.Update{}

	var firestoreUser UpdateUserRequest
	// Convert Firestore data to JSON → back to struct for comparison
	dataBytes, _ := json.Marshal(currentData)
	json.Unmarshal(dataBytes, &firestoreUser)

	if !reflect.DeepEqual(firestoreUser.Brands, user.Brands) {
		updates = append(updates, firestore.Update{FieldPath: []string{"brands"}, Value: user.Brands})
	}
	if !reflect.DeepEqual(firestoreUser.Properties, user.Properties) {
		updates = append(updates, firestore.Update{FieldPath: []string{"properties"}, Value: user.Properties})
	}

	// If no changes detected, respond with success and exit
	if len(updates) == 0 {
		shared.WriteJSONSuccess(response, http.StatusOK, "No changes detected", nil)
		return
	}

	// Apply updates to Firestore
	_, err = shared.FirestoreClient.Collection("users").Doc(user.UID).Update(ctx, updates)
	if err != nil {
		log.Printf("Firestore update error: %v", err)
		shared.WriteJSONError(response, http.StatusInternalServerError, "Error occurred while updating the user data")
		return
	}

	shared.WriteJSONSuccess(response, http.StatusOK, "Information updated successfully", nil)
}