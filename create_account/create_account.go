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

func init() {
	if os.Getenv("ENV") != "DEBUG" {
		shared.InitFirebaseProd(nil)
		functions.HTTP("create-account", CreateAccount)
	}
}

type CreateAccountRequest struct {
	Name        string              `json:"name"`
	PhoneNumber string              `json:"phone_number"`
	Email       string              `json:"email"`
	Password    string              `json:"password"`
	Properties  []map[string]string `json:"properties"`
	Brands      []string            `json:"brands"`
}

func CreateAccount(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if shared.CorsEnabledFunction(response, request) {
		return
	}

	if request.Method != http.MethodPost {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong http method")
		return
	}

	if err := shared.IsAuthorized(request); err != nil {
		shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
		return
	}

	defer request.Body.Close()

	var createAccountRequest CreateAccountRequest
	if err := json.NewDecoder(request.Body).Decode(&createAccountRequest); err != nil {
		log.Printf("JSON decode error: %v", err)
		shared.WriteJSONError(response, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(createAccountRequest.Properties) == 0 {
		shared.WriteJSONError(response, http.StatusBadRequest, "At least one property is required")
		return
	}

	// Check if the properties entered are valid
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

	// Make sure at least one brand is selected
	if len(createAccountRequest.Brands) == 0 {
		shared.WriteJSONError(response, http.StatusBadRequest, "At least one brand is required")
		return
	}

	userToCreate := (&auth.UserToCreate{}).
		Email(createAccountRequest.Email).
		Password(createAccountRequest.Password).
		DisplayName(createAccountRequest.Name).
		PhoneNumber(createAccountRequest.PhoneNumber)

	createdUser, err := shared.AuthClient.CreateUser(ctx, userToCreate)
	
	if err != nil {
		log.Printf("Error ocurred when creating the user: %v", err)
		
		firebaseError := shared.ExtractFirebaseErrorFromResponse(err)
		if firebaseError != nil{
			shared.WriteJSONError(response, http.StatusInternalServerError, firebaseError.Error.Message)
		}else{
			shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	data := map[string]any{
		"properties": createAccountRequest.Properties,
		"brands":     createAccountRequest.Brands,
	}

	if _, err := shared.FirestoreClient.Collection("users").Doc(createdUser.UID).Set(ctx, data); err != nil {
		log.Printf("Firestore write error: %v", err)
		shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
		return
	}

	shared.WriteJSONSuccess(response, http.StatusOK, "Account Created Successfully", nil)
}
