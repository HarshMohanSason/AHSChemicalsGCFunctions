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

type UpdateUserRequest struct {
	UID        		string   			`json:"uid"`
	Brands     		[]string 		 	`json:"brands"`
	Properties 		[]map[string]any 	`json:"properties"`
}

func init() {
	if os.Getenv("ENV") != "DEBUG"{
		shared.InitFirebaseProd(nil)
		functions.HTTP("update-account", UpdateAccount)
	}
}
func UpdateAccount(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if shared.CorsEnabledFunction(response, request) {
		return
	}

	if err := shared.IsAuthorized(request); err != nil {
		shared.WriteJSONError(response, http.StatusUnauthorized, "Not authorized to do this action")
		return
	}

	if request.Method != http.MethodPut {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method, expected PUT")
		return
	}

	defer request.Body.Close()

	var user UpdateUserRequest
	if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request body: %v", err)
		shared.WriteJSONError(response, http.StatusBadRequest, "Invalid request body")
		return
	}

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

	docSnapshot, err := shared.FirestoreClient.Collection("users").Doc(user.UID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching document for uid %v: %v", user.UID, err)
		shared.WriteJSONError(response, http.StatusBadRequest, "No user found with the given UID")
		return
	}

	currentData := docSnapshot.Data()
	updates := []firestore.Update{}

	if !reflect.DeepEqual(currentData["brands"], user.Brands) {
		updates = append(updates, firestore.Update{FieldPath: []string{"brands"}, Value: user.Brands})
	}
	if !reflect.DeepEqual(currentData["properties"], user.Properties) {
		updates = append(updates, firestore.Update{FieldPath: []string{"properties"}, Value: user.Properties})
	}

	if len(updates) == 0 {
		shared.WriteJSONSuccess(response, http.StatusOK, "No changes detected", nil)
		return
	}

	_, err = shared.FirestoreClient.Collection("users").Doc(user.UID).Update(ctx, updates)
	if err != nil {
		log.Printf("Firestore update error: %v", err)
		shared.WriteJSONError(response, http.StatusInternalServerError, "Error occurred while updating the user data")
		return
	}

	shared.WriteJSONSuccess(response, http.StatusOK, "Information updated successfully", nil)
}